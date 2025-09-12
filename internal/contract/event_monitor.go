package contract

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blues/cfs/internal/contract/processor"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type EventMonitor struct {
	contractManager  *ContractManager
	db               *gorm.DB
	processorManager *processor.ProcessorManager
	lastBlock        int64
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewEventMonitor(
	contractManager *ContractManager,
	db *gorm.DB,
) *EventMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建处理器管理器
	processorManager := processor.NewProcessorManager(db)

	return &EventMonitor{
		contractManager:  contractManager,
		db:               db,
		processorManager: processorManager,
		lastBlock:        0,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start 开始监控链上事件
func (m *EventMonitor) Start() error {
	// 获取最后处理的区块号
	if err := m.loadLastBlock(); err != nil {
		logger.Warn("Failed to load last block, starting from config: %v", err)
		// 获取第一个启用的合约作为默认起始区块
		contracts := m.contractManager.GetAllContracts()
		if len(contracts) > 0 {
			m.lastBlock = 0 // 从0开始
		}
	}

	logger.Info("Starting blockchain monitor from block %d", m.lastBlock)

	// 启动监控循环
	go m.monitorLoop()
	return nil
}

// Stop 停止监控
func (m *EventMonitor) Stop() {
	m.cancel()
}

// loadLastBlock 从数据库加载最后处理的区块号
func (m *EventMonitor) loadLastBlock() error {
	var lastEvent model.EventModel
	err := m.db.Order("block_num DESC").First(&lastEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			m.lastBlock = 0 // 没有事件记录，从0开始
			logger.Info("No previous events found, starting from block 0")
			return nil
		}
		return err
	}
	m.lastBlock = lastEvent.BlockNum
	logger.Info("Loaded last block %d", lastEvent.BlockNum)
	return nil
}

// monitorLoop 监控循环
func (m *EventMonitor) monitorLoop() {
	ticker := time.NewTicker(1 * time.Minute) // 每5分钟检查一次，避免触发API限制
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logger.Info("Monitor stopped")
			return
		case <-ticker.C:
			if err := m.processNewBlocks(); err != nil {
				logger.Error("Error processing blocks: %v", err)
				// 如果是API限制错误，等待更长时间
				if strings.Contains(err.Error(), "Too Many Requests") {
					logger.Warn("API rate limit hit, waiting 10 minutes before retry")
					time.Sleep(1 * time.Minute)
				}
			}
		}
	}
}

// processNewBlocks 处理新区块
func (m *EventMonitor) processNewBlocks() error {
	logger.Info("Processing new blocks")

	// 获取第一个启用的合约来获取当前区块号
	contracts := m.contractManager.GetAllContracts()
	if len(contracts) == 0 {
		return fmt.Errorf("no contracts available")
	}

	var currentContract *Contract
	for _, contract := range contracts {
		currentContract = contract
		break
	}

	currentBlock, err := currentContract.GetCurrentBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get current block number: %w", err)
	}

	// 处理从lastBlock到currentBlock的所有区块
	for blockNum := m.lastBlock; blockNum <= currentBlock; blockNum++ {
		if err := m.processBlock(blockNum); err != nil {
			logger.Error("Error processing block %d: %v", blockNum, err)
			continue
		}
		m.lastBlock = blockNum
	}

	return nil
}

// processBlock 处理单个区块
func (m *EventMonitor) processBlock(blockNum int64) error {
	logger.Info("Processing block %d", blockNum)

	// 获取所有合约的日志
	contracts := m.contractManager.GetAllContracts()
	for name, contract := range contracts {
		logger.Debug("Processing block %d for contract %s", blockNum, name)

		// 获取区块日志
		logs, err := contract.GetBlockLogs(blockNum)
		if err != nil {
			logger.Error("Failed to get block logs for contract %s: %v", name, err)
			continue
		}

		// 处理每个日志
		for _, l := range logs {
			logger.Info("Processing log: %v", l)
			if err := m.processLog(l, contract); err != nil {
				logger.Error("Error processing log: %v", err)
				continue
			}
		}
	}

	return nil
}

// processLog 处理单个日志
func (m *EventMonitor) processLog(l types.Log, contract *Contract) error {
	// 解析事件数据
	eventData, err := contract.ParseEvent(l)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	// 检查事件是否已存在
	var count int64
	if err := m.db.Model(&model.EventModel{}).Where("tx_hash = ? AND log_index = ?", l.TxHash.Hex(), int(l.Index)).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check if event exists: %w", err)
	}
	if count > 0 {
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
		ContractAddress: contract.GetAddress().Hex(),
		ContractName:    contract.GetName(),
		EventType:       eventData["eventType"].(string),
		TxHash:          l.TxHash.Hex(),
		BlockNum:        int64(l.BlockNumber),
		LogIndex:        int64(l.Index),
		Data:            string(dataJSON),
		Processed:       false,
	}

	// 直接保存到数据库
	if err := m.db.Create(&event).Error; err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	logger.Info("Saved event: %s in block %d", event.EventType, l.BlockNumber)

	// 处理事件
	return m.handleEvent(&event, eventData)
}

// handleEvent 处理事件
func (m *EventMonitor) handleEvent(event *model.EventModel, eventData map[string]interface{}) error {
	// 使用处理器管理器处理事件
	if err := m.processorManager.ProcessEvent(event, eventData); err != nil {
		logger.Error("Failed to process event %s from contract %s: %v", event.EventType, event.ContractName, err)
		return err
	}

	// 标记事件为已处理
	if err := m.db.Model(&model.EventModel{}).Where("id = ?", event.Id).Update("processed", true).Error; err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	logger.Info("Successfully processed event: %s from contract %s", event.EventType, event.ContractName)
	return nil
}
