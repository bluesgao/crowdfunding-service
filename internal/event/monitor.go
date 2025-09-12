package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
	"github.com/ethereum/go-ethereum/core/types"
)

type Monitor struct {
	ethClient           *ethereum.Client
	eventLogic          *logic.EventLogic
	contributeProcessor *ContributeProcessor
	refundProcessor     *RefundProcessor
	projectProcessor    *ProjectProcessor
	lastBlock           int64
	ctx                 context.Context
	cancel              context.CancelFunc
}

func NewMonitor(
	ethClient *ethereum.Client,
	eventLogic *logic.EventLogic,
	projectLogic *logic.ProjectLogic,
	contributeLogic *logic.ContributeRecordLogic,
	refundLogic *logic.RefundRecordLogic,
) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建事件处理器
	contributeProcessor := NewContributeProcessor(contributeLogic)
	refundProcessor := NewRefundProcessor(refundLogic)
	projectProcessor := NewProjectProcessor(projectLogic)

	return &Monitor{
		ethClient:           ethClient,
		eventLogic:          eventLogic,
		contributeProcessor: contributeProcessor,
		refundProcessor:     refundProcessor,
		projectProcessor:    projectProcessor,
		ctx:                 ctx,
		cancel:              cancel,
	}
}

// Start 开始监控链上事件
func (m *Monitor) Start() error {
	// 获取最后处理的区块号
	if err := m.loadLastBlock(); err != nil {
		logger.Warn("Failed to load last block, starting from config: %v", err)
		m.lastBlock = m.ethClient.GetStartBlock()
	}

	logger.Info("Starting blockchain monitor from block %d", m.lastBlock)

	// 启动监控循环
	go m.monitorLoop()
	return nil
}

// Stop 停止监控
func (m *Monitor) Stop() {
	m.cancel()
}

// loadLastBlock 从数据库加载最后处理的区块号
func (m *Monitor) loadLastBlock() error {
	lastBlock, err := m.eventLogic.GetLastProcessedBlock()
	if err != nil {
		return err
	}

	if lastBlock == 0 {
		m.lastBlock = m.ethClient.GetStartBlock()
	} else {
		m.lastBlock = int64(lastBlock)
	}
	logger.Info("Loaded last block %d", lastBlock)
	return nil
}

// monitorLoop 监控循环
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(60 * time.Second) // 每60秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logger.Info("Monitor stopped")
			return
		case <-ticker.C:
			if err := m.processNewBlocks(); err != nil {
				logger.Error("Error processing blocks: %v", err)
			}
		}
	}
}

// processNewBlocks 处理新区块
func (m *Monitor) processNewBlocks() error {
	// 获取当前区块号
	currentBlock, err := m.ethClient.GetCurrentBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// 处理从lastBlock到currentBlock的所有区块
	for blockNum := m.lastBlock + 1; blockNum <= currentBlock; blockNum++ {
		if err := m.processBlock(blockNum); err != nil {
			logger.Error("Error processing block %d: %v", blockNum, err)
			continue
		}
		m.lastBlock = blockNum
	}

	return nil
}

// processBlock 处理单个区块
func (m *Monitor) processBlock(blockNum int64) error {
	logger.Debug("Processing block %d", blockNum)

	// 获取区块交易
	transactions, err := m.ethClient.GetBlockTransactions(blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block transactions: %w", err)
	}

	for _, tx := range transactions {
		logger.Debug("Processing transaction: %v", tx.Hash().Hex())
	}

	// 获取区块日志
	logs, err := m.ethClient.GetBlockLogs(blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block logs: %w", err)
	}

	// 处理每个日志
	for _, l := range logs {
		if err := m.processLog(l); err != nil {
			logger.Error("Error processing log: %v", err)
			continue
		}
	}

	return nil
}

// processLog 处理单个日志
func (m *Monitor) processLog(l types.Log) error {
	// 解析事件数据
	eventData, err := m.ethClient.ParseEvent(l)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	// 检查事件是否已存在
	exists, err := m.eventLogic.CheckEventExists(l.TxHash.Hex(), int(l.Index))
	if err != nil {
		return fmt.Errorf("failed to check if event exists: %w", err)
	}
	if exists {
		// 事件已存在，跳过
		return nil
	}

	// 序列化事件数据
	dataJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 创建事件记录
	event := model.EventModel{
		EventType: eventData["eventType"].(string),
		TxHash:    l.TxHash.Hex(),
		BlockNum:  int64(l.BlockNumber),
		LogIndex:  int64(l.Index),
		Data:      string(dataJSON),
		Processed: false,
	}

	// 如果是项目相关事件，设置项目ID
	if projectId, ok := eventData["projectId"]; ok {
		event.ProjectId = projectId.(int64)
	}

	// 通过event_logic保存到数据库
	if err := m.eventLogic.CreateEvent(&event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	logger.Info("Saved event: %s in block %d", event.EventType, l.BlockNumber)

	// 处理事件
	return m.handleEvent(&event, eventData)
}

// handleEvent 处理事件
func (m *Monitor) handleEvent(event *model.EventModel, eventData map[string]interface{}) error {
	var err error

	switch event.EventType {
	case "ContributionMade":
		err = m.contributeProcessor.Process(event, eventData)
	case "RefundProcessed":
		err = m.refundProcessor.Process(event, eventData)
	case "ProjectStatusChanged", "ProjectStatus", "ProjectCreated":
		err = m.projectProcessor.Process(event, eventData)
	default:
		logger.Warn("Unknown event type: %s", event.EventType)
		return nil
	}

	if err != nil {
		return err
	}

	// 标记事件为已处理
	if err := m.eventLogic.UpdateEventProcessed(event.Id, true); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}
