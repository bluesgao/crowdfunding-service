package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type Monitor struct {
	client                  *ethereum.Client
	eventLogic              *logic.EventLogic
	contributeProcessor     *ContributeProcessor
	refundProcessor         *RefundProcessor
	projectStatusProcessor  *ProjectStatusProcessor
	projectCreatedProcessor *ProjectCreatedProcessor
	lastBlock               uint64
	ctx                     context.Context
	cancel                  context.CancelFunc
}

func NewMonitor(client *ethereum.Client, db *gorm.DB) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建logic层实例
	eventLogic := logic.NewEventLogic(db)
	projectLogic := logic.NewProjectLogic(db)
	contributeLogic := logic.NewContributeRecordLogic(db)
	refundLogic := logic.NewRefundRecordLogic(db)

	// 创建事件处理器
	contributeProcessor := NewContributeProcessor(contributeLogic)
	refundProcessor := NewRefundProcessor(refundLogic)
	projectStatusProcessor := NewProjectStatusProcessor(projectLogic)
	projectCreatedProcessor := NewProjectCreatedProcessor()

	return &Monitor{
		client:                  client,
		eventLogic:              eventLogic,
		contributeProcessor:     contributeProcessor,
		refundProcessor:         refundProcessor,
		projectStatusProcessor:  projectStatusProcessor,
		projectCreatedProcessor: projectCreatedProcessor,
		ctx:                     ctx,
		cancel:                  cancel,
	}
}

// Start 开始监控链上事件
func (m *Monitor) Start() error {
	// 获取最后处理的区块号
	if err := m.loadLastBlock(); err != nil {
		log.Printf("Failed to load last block, starting from config: %v", err)
		m.lastBlock = m.client.StartBlock
	}

	log.Printf("Starting blockchain monitor from block %d", m.lastBlock)

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
		m.lastBlock = m.client.StartBlock
	} else {
		m.lastBlock = lastBlock
	}
	return nil
}

// monitorLoop 监控循环
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			log.Println("Monitor stopped")
			return
		case <-ticker.C:
			if err := m.processNewBlocks(); err != nil {
				log.Printf("Error processing blocks: %v", err)
			}
		}
	}
}

// processNewBlocks 处理新区块
func (m *Monitor) processNewBlocks() error {
	// 获取当前区块号
	currentBlock, err := m.client.GetCurrentBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// 处理从lastBlock到currentBlock的所有区块
	for blockNum := m.lastBlock + 1; blockNum <= currentBlock; blockNum++ {
		if err := m.processBlock(blockNum); err != nil {
			log.Printf("Error processing block %d: %v", blockNum, err)
			continue
		}
		m.lastBlock = blockNum
	}

	return nil
}

// processBlock 处理单个区块
func (m *Monitor) processBlock(blockNum uint64) error {
	// 获取区块日志
	logs, err := m.client.GetBlockLogs(blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block logs: %w", err)
	}

	// 处理每个日志
	for _, l := range logs {
		if err := m.processLog(l); err != nil {
			log.Printf("Error processing log: %v", err)
			continue
		}
	}

	return nil
}

// processLog 处理单个日志
func (m *Monitor) processLog(l types.Log) error {
	// 解析事件数据
	eventData, err := m.client.ParseEvent(l)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	// 检查事件是否已存在
	exists, err := m.eventLogic.CheckEventExists(l.TxHash.Hex(), l.Index)
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
	event := model.Event{
		EventType: eventData["eventType"].(string),
		TxHash:    l.TxHash.Hex(),
		BlockNum:  l.BlockNumber,
		LogIndex:  l.Index,
		Data:      string(dataJSON),
		Processed: false,
	}

	// 如果是项目相关事件，设置项目ID
	if projectID, ok := eventData["projectId"]; ok {
		event.ProjectID = uint(projectID.(uint64))
	}

	// 通过event_logic保存到数据库
	if err := m.eventLogic.CreateEvent(&event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	log.Printf("Saved event: %s in block %d", event.EventType, l.BlockNumber)

	// 处理事件
	return m.handleEvent(&event, eventData)
}

// handleEvent 处理事件
func (m *Monitor) handleEvent(event *model.Event, eventData map[string]interface{}) error {
	var err error

	switch event.EventType {
	case "ContributionMade":
		err = m.contributeProcessor.Process(event, eventData)
	case "RefundProcessed":
		err = m.refundProcessor.Process(event, eventData)
	case "ProjectStatusChanged", "ProjectStatus":
		err = m.projectStatusProcessor.Process(event, eventData)
	case "ProjectCreated":
		err = m.projectCreatedProcessor.Process(event, eventData)
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return nil
	}

	if err != nil {
		return err
	}

	// 标记事件为已处理
	if err := m.eventLogic.UpdateEventProcessed(event.ID, true); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}
