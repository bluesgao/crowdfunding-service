package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/blues/cfs/internal/chain"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/panjf2000/ants/v2"
	"gorm.io/gorm"
)

// EventMonitor 区块链事件监控器
type EventMonitor struct {
	chainManager    *chain.Manager
	db              *gorm.DB
	eventProcessor  *EventProcessor
	pool            *ants.Pool // 协程池
	startBlockNum   int64
	ctx             context.Context
	cancel          context.CancelFunc
	retryCount      int           // 重试次数
	lastRetryTime   time.Time     // 上次重试时间
	backoffDuration time.Duration // 退避时间
	mu              sync.RWMutex  // 保护 startBlockNum 的并发访问
}

// NewEventMonitor 创建事件监控器
func NewEventMonitor(
	chainManager *chain.Manager,
	db *gorm.DB,
	poolSize int,
) *EventMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建协程池
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		logger.Error("Failed to create goroutine pool: %v", err)
		pool = nil
	}

	// 创建事件处理器
	eventProcessor := NewEventProcessor(db)

	return &EventMonitor{
		chainManager:    chainManager,
		db:              db,
		eventProcessor:  eventProcessor,
		pool:            pool,
		startBlockNum:   0,
		ctx:             ctx,
		cancel:          cancel,
		retryCount:      0,
		lastRetryTime:   time.Now(),
		backoffDuration: time.Second * 5, // 初始退避时间5秒
	}
}

// Start 启动监控
func (m *EventMonitor) Start() error {
	logger.Info("Starting blockchain event monitor")

	// 检查是否有合约
	contracts := m.chainManager.GetContracts()
	if len(contracts) == 0 {
		return fmt.Errorf("no contracts available for monitoring")
	}
	logger.Info("Found %d contracts to monitor", len(contracts))

	// 检查 RPC 连接
	client := m.chainManager.GetClient()
	if client == nil {
		return fmt.Errorf("chain client not available")
	}

	// 测试 RPC 连接
	currentBlock, err := m.getCurrentBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain: %w", err)
	}
	logger.Info("Connected to blockchain, current block: %d", currentBlock)

	// 获取起始区块号
	startBlock := m.getStartBlockNum()
	if startBlock == 0 {
		return fmt.Errorf("failed to determine start block number")
	}

	// 设置起始区块号
	m.mu.Lock()
	m.startBlockNum = startBlock
	m.mu.Unlock()

	logger.Info("Starting monitor from block %d", startBlock)

	// 启动监控循环
	go m.loop()

	return nil
}

// Stop 停止监控
func (m *EventMonitor) Stop() {
	logger.Info("Stopping blockchain event monitor")
	m.cancel()

	// 等待协程池关闭
	if m.pool != nil {
		m.pool.Release()
	}
}

// loop 监控循环
func (m *EventMonitor) loop() {
	ticker := time.NewTicker(time.Second * 10) // 每10秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logger.Info("Monitor stopped")
			return
		case <-ticker.C:
			// 获取当前区块号
			currentBlock, err := m.getCurrentBlockNumber()
			if err != nil {
				logger.Error("Failed to get current block number: %v", err)
				m.handleError(err)
				continue
			}

			logger.Debug("Current block number: %d", currentBlock)

			// 获取起始区块号
			startBlock := m.getStartBlockNum()
			if startBlock == 0 {
				logger.Error("Failed to determine start block number")
				continue
			}

			// 处理每个区块
			logger.Debug("Processing blocks from %d to %d", startBlock, currentBlock)
			for blockNum := startBlock; blockNum <= currentBlock; blockNum++ {
				logger.Debug("Processing block %d", blockNum)
				if err := m.processBlock(blockNum); err != nil {
					// 如果是API限制错误，直接返回错误
					if strings.Contains(err.Error(), "Too Many Requests") {
						logger.Error("API rate limit hit while processing block %d: %v", blockNum, err)
						m.handleError(err)
						break
					}
					logger.Error("Error processing block %d: %v", blockNum, err)
					continue
				}

				// 更新起始区块号
				m.updateStartBlockNum(blockNum + 1)
			}
		}
	}
}

// processBlock 处理单个区块
func (m *EventMonitor) processBlock(blockNum int64) error {
	// 检查是否已经处理过这个区块
	if m.isBlockProcessed(blockNum) {
		logger.Debug("Block %d already processed, skipping", blockNum)
		return nil
	}

	// 获取所有合约
	contracts := m.chainManager.GetContracts()
	if len(contracts) == 0 {
		logger.Debug("No contracts found")
		return nil
	}

	logger.Debug("Found %d contracts to process for block %d", len(contracts), blockNum)

	// 获取区块信息
	block := chain.NewBlock()
	client := m.chainManager.GetClient()

	// 处理每个合约
	for contractName, contract := range contracts {
		// 检查合约是否已部署
		if blockNum < contract.GetBlockNum() {
			logger.Debug("Skipping block %d for contract %s (deployed at block %d)",
				blockNum, contractName, contract.GetBlockNum())
			continue
		}

		// 获取该合约在该区块的日志
		logs, err := block.GetBlockLogs(client, contract.GetAddress(), blockNum)
		if err != nil {
			// 如果是API限制错误，直接返回错误
			if strings.Contains(err.Error(), "Too Many Requests") {
				return fmt.Errorf("API rate limit hit while getting logs for contract %s: %w", contractName, err)
			}
			logger.Error("Error getting logs for block %d, contract %s: %v", blockNum, contractName, err)
			continue
		}

		// 如果没有日志，跳过
		if len(logs) == 0 {
			continue
		}

		logger.Debug("Found %d logs for contract %s in block %d", len(logs), contractName, blockNum)

		// 使用协程池并发处理日志
		if m.pool != nil {
			err := m.pool.Submit(func() {
				m.processLogs(contract, logs, blockNum)
			})
			if err != nil {
				return err
			}
		} else {
			// 如果没有协程池，直接处理
			m.processLogs(contract, logs, blockNum)
		}
	}

	// 标记区块已处理
	m.markBlockProcessed(blockNum)

	return nil
}

// processLogs 处理日志
func (m *EventMonitor) processLogs(contract *chain.Contract, logs []types.Log, blockNum int64) {
	for _, log := range logs {
		// 解析事件
		eventData, err := contract.ParseEvent(log)
		if err != nil {
			logger.Error("Error parsing event for contract %s: %v", contract.GetName(), err)
			continue
		}

		// 创建事件模型
		event := &model.EventModel{
			ContractAddress: contract.GetAddress().Hex(),
			ContractName:    contract.GetName(),
			BlockNum:        blockNum,
			TxHash:          log.TxHash.Hex(),
			LogIndex:        int64(log.Index),
			EventName:       eventData["eventName"].(string),
			Data:            fmt.Sprintf("%v", eventData),
		}

		// 处理事件
		if err := m.eventProcessor.ProcessEvent(event, eventData); err != nil {
			logger.Error("Error processing event for contract %s: %v", contract.GetName(), err)
			continue
		}

		logger.Debug("Processed event for contract %s at block %d", contract.GetName(), blockNum)
	}
}

// getCurrentBlockNumber 获取当前最新区块号
func (m *EventMonitor) getCurrentBlockNumber() (int64, error) {
	block := chain.NewBlock()
	client := m.chainManager.GetClient()
	return block.GetCurrentBlockNumber(client)
}

// getStartBlockNum 获取起始区块号
func (m *EventMonitor) getStartBlockNum() int64 {
	m.mu.RLock()
	startBlock := m.startBlockNum
	m.mu.RUnlock()

	// 如果已经设置了起始区块号，直接返回
	if startBlock > 0 {
		return startBlock
	}

	// 1. 从配置文件中获取合约部署的最小区块号
	contracts := m.chainManager.GetContracts()
	if len(contracts) == 0 {
		logger.Error("No contracts found in configuration")
		return 0
	}

	minDeployBlock := int64(0)
	first := true
	for _, contract := range contracts {
		if first {
			minDeployBlock = contract.GetBlockNum()
			first = false
		} else if contract.GetBlockNum() < minDeployBlock {
			minDeployBlock = contract.GetBlockNum()
		}
	}
	logger.Debug("Minimum deploy block from config: %d", minDeployBlock)

	// 2. 从数据库获取已处理的最大区块号
	var maxProcessedBlock int64
	err := m.db.Model(&model.EventModel{}).
		Select("COALESCE(MAX(block_num), 0)").
		Scan(&maxProcessedBlock).Error

	if err != nil {
		logger.Error("Failed to get max processed block number from database: %v", err)
		// 如果数据库查询失败，使用配置中的最小部署区块号
		return minDeployBlock
	}
	logger.Debug("Max processed block from database: %d", maxProcessedBlock)

	// 3. 取两个值中的较大者作为起始点
	finalStartBlock := minDeployBlock
	if maxProcessedBlock > minDeployBlock {
		finalStartBlock = maxProcessedBlock + 1 // 从下一个区块开始处理
	}

	logger.Info("Final start block: %d (config: %d, db: %d)", finalStartBlock, minDeployBlock, maxProcessedBlock)
	return finalStartBlock
}

// updateStartBlockNum 更新起始区块号
func (m *EventMonitor) updateStartBlockNum(blockNum int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startBlockNum = blockNum
}

// isBlockProcessed 检查区块是否已处理
func (m *EventMonitor) isBlockProcessed(blockNum int64) bool {
	var count int64
	err := m.db.Model(&model.EventModel{}).
		Where("block_num = ?", blockNum).
		Count(&count).Error

	if err != nil {
		logger.Error("Error checking if block %d is processed: %v", blockNum, err)
		return false
	}

	return count > 0
}

// markBlockProcessed 标记区块已处理
func (m *EventMonitor) markBlockProcessed(blockNum int64) {
	// 这里可以添加标记逻辑，比如记录到数据库
	logger.Debug("Marked block %d as processed", blockNum)
}

// handleError 处理错误
func (m *EventMonitor) handleError(err error) {
	m.retryCount++
	m.lastRetryTime = time.Now()

	// 指数退避
	if m.retryCount > 5 {
		m.backoffDuration = time.Minute * 5 // 最大退避时间5分钟
	} else {
		m.backoffDuration = time.Duration(m.retryCount) * time.Second * 10
	}

	logger.Error("Monitor encountered error (retry %d): %v", m.retryCount, err)
}

// GetStatus 获取监控状态
func (m *EventMonitor) GetStatus() map[string]interface{} {
	contracts := m.chainManager.GetContracts()

	status := map[string]interface{}{
		"start_block":    m.getStartBlockNum(),
		"contract_count": len(contracts),
		"pool_status":    m.getPoolStatus(),
		"chain_info":     m.chainManager.GetHealthStatus(),
	}

	return status
}

// getPoolStatus 获取协程池状态
func (m *EventMonitor) getPoolStatus() map[string]interface{} {
	if m.pool == nil {
		return map[string]interface{}{
			"running": 0,
			"free":    0,
			"cap":     0,
		}
	}

	return map[string]interface{}{
		"running": m.pool.Running(),
		"free":    m.pool.Free(),
		"cap":     m.pool.Cap(),
	}
}

// GetStatusJSON 获取监控状态的JSON格式
func (m *EventMonitor) GetStatusJSON() (string, error) {
	status := m.GetStatus()
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal monitor status: %w", err)
	}
	return string(jsonData), nil
}
