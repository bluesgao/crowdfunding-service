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
	retryCount       int           // 重试次数
	lastRetryTime    time.Time     // 上次重试时间
	backoffDuration  time.Duration // 退避时间
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
		retryCount:       0,
		backoffDuration:  time.Minute, // 初始退避时间1分钟
	}
}

// Start 开始监控链上事件
func (m *EventMonitor) Start() error {
	// 获取起始区块号
	m.lastBlock = m.getStartBlockNum()

	logger.Info("Starting blockchain monitor from block %d", m.lastBlock)

	// 启动监控循环
	go m.monitorLoop()
	return nil
}

// Stop 停止监控
func (m *EventMonitor) Stop() {
	m.cancel()
}

// monitorLoop 监控循环
func (m *EventMonitor) monitorLoop() {
	baseInterval := 1 * time.Minute // 基础检查间隔
	ticker := time.NewTicker(baseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logger.Info("Monitor stopped")
			return
		case <-ticker.C:
			// 检查是否需要退避
			if m.shouldBackoff() {
				logger.Info("Still in backoff period, skipping this cycle")
				continue
			}

			// 处理下一个区块
			if err := m.processBlock(m.lastBlock); err != nil {
				logger.Error("Error processing block %d: %v", m.lastBlock+1, err)
				// 如果是API限制错误，启动退避策略
				if strings.Contains(err.Error(), "Too Many Requests") {
					m.handleRateLimit()
				} else {
					// 其他错误，重置退避状态
					m.resetBackoff()
				}
			} else {
				// 成功处理，更新lastBlock并重置退避状态
				m.lastBlock++
				m.resetBackoff()
			}
		}
	}
}

// processBlock 处理单个区块
func (m *EventMonitor) processBlock(blockNum int64) error {
	logger.Debug("Processing block %d", blockNum)

	// 获取所有合约的日志
	contracts := m.contractManager.GetAllContracts()
	totalLogs := 0
	processedLogs := 0

	for name, contract := range contracts {
		// 检查合约是否已部署（如果blockNum为0，说明还在异步获取中）
		if contract.GetBlockNum() > 0 && blockNum < contract.GetBlockNum() {
			logger.Debug("Skipping block %d for contract %s (deployed at block %d)",
				blockNum, name, contract.GetBlockNum())
			continue
		}

		// 获取区块日志
		logs, err := contract.GetBlockLogs(blockNum)
		if err != nil {
			// 如果是API限制错误，直接返回错误
			if strings.Contains(err.Error(), "Too Many Requests") {
				return fmt.Errorf("API rate limit hit while getting logs for contract %s: %w", name, err)
			}
			logger.Error("Failed to get block logs for contract %s: %v", name, err)
			continue
		}

		totalLogs += len(logs)
		logger.Debug("Found %d logs in block %d for contract %s", len(logs), blockNum, name)

		// 处理每个日志
		for _, l := range logs {
			if err := m.processLog(l, contract); err != nil {
				logger.Error("Error processing log in block %d: %v", blockNum, err)
				continue
			}
			processedLogs++
		}
	}

	if totalLogs > 0 {
		logger.Info("Processed block %d: %d/%d logs processed", blockNum, processedLogs, totalLogs)
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
	logger.Info("Parsed event data: %v", eventData)

	// 创建事件记录
	event, err := m.createEventRecord(l, contract, eventData)
	if err != nil {
		return err
	}

	// 使用 upsert 操作保存事件到数据库
	if err := m.upsertEventRecord(event); err != nil {
		return err
	}

	logger.Info("Saved event: %s in block %d", event.EventName, l.BlockNumber)

	// 处理事件
	return m.handleEvent(event, eventData)
}

// handleEvent 处理事件
func (m *EventMonitor) handleEvent(event *model.EventModel, eventData map[string]interface{}) error {
	// 使用处理器管理器处理事件
	if err := m.processorManager.ProcessEvent(event, eventData); err != nil {
		logger.Error("Failed to process event %s from contract %s: %v", event.EventName, event.ContractName, err)
		return err
	}

	// 标记事件为已处理
	if err := m.db.Model(&model.EventModel{}).Where("id = ?", event.Id).Update("processed", true).Error; err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	logger.Info("Successfully processed event: %s from contract %s", event.EventName, event.ContractName)
	return nil
}

// createEventRecord 创建事件记录
func (m *EventMonitor) createEventRecord(l types.Log, contract *Contract, eventData map[string]interface{}) (*model.EventModel, error) {
	// 序列化事件数据
	dataJSON, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 创建事件记录
	event := &model.EventModel{
		ContractAddress: contract.GetAddress().Hex(),
		ContractName:    contract.GetName(),
		EventName:       eventData["eventName"].(string),
		TxHash:          l.TxHash.Hex(),
		BlockNum:        int64(l.BlockNumber),
		LogIndex:        int64(l.Index),
		Data:            string(dataJSON),
		Processed:       false,
	}

	return event, nil
}

// upsertEventRecord 使用 upsert 操作保存事件记录
func (m *EventMonitor) upsertEventRecord(event *model.EventModel) error {
	// 使用 GORM 的 Save 方法进行 upsert 操作
	// 如果记录存在则更新，不存在则创建
	if err := m.db.Save(event).Error; err != nil {
		return fmt.Errorf("failed to upsert event: %w", err)
	}
	return nil
}

// getContractStartBlock 获取合约的起始区块号
func (m *EventMonitor) getContractStartBlock() int64 {
	contracts := m.contractManager.GetAllContracts()
	if len(contracts) == 0 {
		logger.Warn("No contracts available, starting from block 0")
		return 0
	}

	// 找到最早的合约部署区块号
	var earliestBlock int64 = -1
	for name, contract := range contracts {
		blockNum := contract.GetBlockNum()
		if blockNum > 0 {
			if earliestBlock == -1 || blockNum < earliestBlock {
				earliestBlock = blockNum
			}
			logger.Info("Contract %s deployed at block %d", name, blockNum)
		}
	}

	if earliestBlock == -1 {
		logger.Warn("No valid deployment blocks found, starting from block 0")
		return 0
	}

	logger.Info("Using earliest contract deployment block %d as start block", earliestBlock)
	return earliestBlock
}

// shouldBackoff 检查是否应该退避
func (m *EventMonitor) shouldBackoff() bool {
	if m.retryCount == 0 {
		return false
	}
	return time.Since(m.lastRetryTime) < m.backoffDuration
}

// handleRateLimit 处理API限制
func (m *EventMonitor) handleRateLimit() {
	m.retryCount++
	m.lastRetryTime = time.Now()

	// 指数退避：1分钟 -> 2分钟 -> 4分钟 -> 8分钟 -> 最大10分钟
	m.backoffDuration = time.Duration(1<<uint(m.retryCount-1)) * time.Minute
	if m.backoffDuration > 10*time.Minute {
		m.backoffDuration = 10 * time.Minute
	}

	logger.Warn("API rate limit hit, retry count: %d, backoff duration: %v",
		m.retryCount, m.backoffDuration)
}

// resetBackoff 重置退避状态
func (m *EventMonitor) resetBackoff() {
	if m.retryCount > 0 {
		logger.Info("Resetting backoff state after successful processing")
		m.retryCount = 0
		m.backoffDuration = time.Minute
	}
}

// getStartBlockNum 获取监控的起始区块号
func (m *EventMonitor) getStartBlockNum() int64 {
	// 1. 首先尝试从数据库加载已经处理的最大块号
	maxProcessedBlock := m.getMaxProcessedBlockNum()
	if maxProcessedBlock > 0 {
		return maxProcessedBlock
	}

	logger.Info("No previous events found in database, determining start block")

	// 2. 优先使用合约的部署区块号
	contractStartBlock := m.getContractStartBlock()
	if contractStartBlock > 0 {
		logger.Info("Using contract deployment block %d as start block", contractStartBlock)
		return contractStartBlock
	}

	// 3. 如果没有找到有效的部署区块，从当前最新区块开始
	currentBlock, err := m.getCurrentBlockNumber()
	if err != nil {
		logger.Error("Failed to get current block number: %v", err)
		logger.Fatal("Falling back to start from block 0")
	}

	return currentBlock
}

// getMaxProcessedBlockNum 获取已经处理的最大块号
func (m *EventMonitor) getMaxProcessedBlockNum() int64 {
	var lastEvent model.EventModel
	err := m.db.Order("block_num DESC").First(&lastEvent).Error
	if err == nil {
		// 成功加载到数据库中的最后区块号
		logger.Info("Loaded max processed block %d from database", lastEvent.BlockNum)
		return lastEvent.BlockNum
	}

	// 如果数据库中没有记录或查询失败
	if err != gorm.ErrRecordNotFound {
		logger.Warn("Failed to load max processed block from database: %v", err)
	}

	return 0 // 没有找到已处理的区块
}

// getCurrentBlockNumber 获取当前最新区块号
func (m *EventMonitor) getCurrentBlockNumber() (int64, error) {
	contracts := m.contractManager.GetAllContracts()
	if len(contracts) == 0 {
		return 0, fmt.Errorf("no contracts available")
	}

	// 使用第一个合约获取当前区块号
	for _, contract := range contracts {
		return contract.GetCurrentBlockNumber()
	}

	// 这行代码永远不会执行，因为上面已经检查了 contracts 长度
	panic("unreachable code")
}
