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
	"github.com/ethereum/go-ethereum/common"
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
) *EventMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	// 创建事件处理器
	eventProcessor := NewEventProcessor(db)

	return &EventMonitor{
		chainManager:    chainManager,
		db:              db,
		eventProcessor:  eventProcessor,
		pool:            nil, // 不再使用全局协程池
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
}

// loop 监控循环
func (m *EventMonitor) loop() {
	ticker := time.NewTicker(time.Second * 60) // 每30秒检查一次，减少API调用频率
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

			// 获取所有合约
			contracts := m.chainManager.GetContracts()
			if len(contracts) == 0 {
				logger.Debug("No contracts found")
				continue
			}

			// 批量处理区块
			if err := m.processBlocksInBatches(contracts, m.startBlockNum, currentBlock); err != nil {
				logger.Error("Error processing blocks: %v", err)
				m.handleError(err)
			}
		}
	}
}

// processBlocksInBatches 分批处理区块
func (m *EventMonitor) processBlocksInBatches(contracts map[string]*chain.Contract, fromBlock, toBlock int64) error {
	logger.Debug("Processing blocks from %d to %d", fromBlock, toBlock)
	batchSize := int64(500) // 减小批量大小，避免API限制

	for currentFrom := fromBlock; currentFrom <= toBlock; currentFrom += batchSize {
		currentTo := currentFrom + batchSize - 1
		if currentTo > toBlock {
			currentTo = toBlock
		}

		logger.Debug("Processing batch blocks %d to %d", currentFrom, currentTo)
		if err := m.processBatchBlocks(contracts, currentFrom, currentTo); err != nil {
			if m.isAPIRateLimitError(err) {
				logger.Error("API rate limit hit while processing blocks %d-%d: %v", currentFrom, currentTo, err)
				return err // 返回错误，让上层处理
			}
			logger.Error("Error processing blocks %d-%d: %v", currentFrom, currentTo, err)
			continue // 继续处理下一批
		}

		// 更新起始区块号
		m.updateStartBlockNum(currentTo + 1)

		// 添加延迟，避免API限制
		time.Sleep(time.Millisecond * 500)
	}

	return nil
}

// processBatchBlocks 批量处理区块
func (m *EventMonitor) processBatchBlocks(contracts map[string]*chain.Contract, fromBlock, toBlock int64) error {
	logger.Debug("Processing blocks %d to %d with %d contracts", fromBlock, toBlock, len(contracts))

	// 获取区块信息
	block := chain.NewBlock()
	client := m.chainManager.GetClient()

	// 获取已部署的合约地址和映射
	contractAddresses, contractMap := m.getDeployedContracts(contracts, fromBlock, toBlock)
	if len(contractAddresses) == 0 {
		logger.Debug("No deployed contracts for blocks %d-%d", fromBlock, toBlock)
		m.markBatchBlocksProcessed(fromBlock, toBlock)
		return nil
	}

	// 批量获取所有合约的日志
	logs, err := block.GetBatchBlockLogs(client, contractAddresses, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("error getting logs for blocks %d-%d: %w", fromBlock, toBlock, err)
	}

	// 如果没有日志，直接标记区块已处理
	if len(logs) == 0 {
		logger.Debug("No logs found for blocks %d-%d", fromBlock, toBlock)
		m.markBatchBlocksProcessed(fromBlock, toBlock)
		return nil
	}

	logger.Debug("Found %d logs for blocks %d-%d", len(logs), fromBlock, toBlock)

	// 按合约地址分组日志
	logsByContract := m.groupLogsByContract(logs)
	groupCount := len(logsByContract)

	if groupCount == 0 {
		logger.Debug("No contract groups to process")
		return nil
	}

	logger.Debug("Processing %d contract groups", groupCount)

	// 创建临时协程池，大小等于分组数量
	tempPool, err := ants.NewPool(groupCount)
	if err != nil {
		return fmt.Errorf("failed to create temporary pool for %d groups: %w", groupCount, err)
	}
	defer tempPool.Release()

	// 使用临时协程池并发处理每个合约的日志
	for address, contractLogs := range logsByContract {
		contract := contractMap[address]
		if contract == nil {
			logger.Warn("Unknown contract address: %s", address.Hex())
			continue
		}

		logger.Debug("Processing %d logs for contract %s", len(contractLogs), contract.GetName())

		err := tempPool.Submit(func() {
			m.processContractLogs(contract, contractLogs)
		})
		if err != nil {
			logger.Error("Failed to submit task to pool: %v", err)
			continue
		}
	}

	// 等待所有协程完成（通过defer tempPool.Release()自动等待）

	// 标记区块已处理
	m.markBatchBlocksProcessed(fromBlock, toBlock)

	return nil
}

// processContractLogs 处理合约的所有日志
func (m *EventMonitor) processContractLogs(contract *chain.Contract, logs []types.Log) {
	for _, log := range logs {
		// 解析事件
		eventData, err := contract.ParseEvent(log)
		if err != nil {
			logger.Error("Error parsing event for contract %s: %v", contract.GetName(), err)
			continue
		}

		// 将事件数据转换为JSON
		eventDataJSON, err := json.Marshal(eventData)
		if err != nil {
			logger.Error("Failed to marshal event data to JSON: %v", err)
			continue
		}

		// 创建事件模型
		event := &model.EventModel{
			ContractAddress: contract.GetAddress().Hex(),
			ContractName:    contract.GetName(),
			BlockNum:        int64(log.BlockNumber),
			TxHash:          log.TxHash.Hex(),
			LogIndex:        int64(log.Index),
			EventName:       eventData["eventName"].(string),
			Data:            string(eventDataJSON),
		}

		// 处理事件
		if err := m.eventProcessor.ProcessEvent(event, eventData); err != nil {
			logger.Error("Error processing event for contract %s: %v", contract.GetName(), err)
			continue
		}

		logger.Debug("Processed event for contract %s at block %d", contract.GetName(), log.BlockNumber)
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

// markBatchBlocksProcessed 标记批量区块已处理
func (m *EventMonitor) markBatchBlocksProcessed(fromBlock, toBlock int64) {
	// 这里可以添加标记逻辑，比如记录到数据库
	logger.Debug("Marked blocks %d-%d as processed", fromBlock, toBlock)
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
	// 现在使用临时协程池，不再有全局协程池状态
	return map[string]interface{}{
		"type":    "temporary_pools",
		"running": 0,
		"free":    0,
		"cap":     0,
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

// getDeployedContracts 获取已部署的合约地址和映射
func (m *EventMonitor) getDeployedContracts(contracts map[string]*chain.Contract, fromBlock, toBlock int64) ([]common.Address, map[common.Address]*chain.Contract) {
	var contractAddresses []common.Address
	contractMap := make(map[common.Address]*chain.Contract)

	for contractName, contract := range contracts {
		// 检查合约是否已部署
		if toBlock < contract.GetBlockNum() {
			logger.Debug("Skipping blocks %d-%d for contract %s (deployed at block %d)",
				fromBlock, toBlock, contractName, contract.GetBlockNum())
			continue
		}

		address := contract.GetAddress()
		contractAddresses = append(contractAddresses, address)
		contractMap[address] = contract
	}

	return contractAddresses, contractMap
}

// isAPIRateLimitError 检查是否为API限制错误
func (m *EventMonitor) isAPIRateLimitError(err error) bool {
	return strings.Contains(err.Error(), "Too Many Requests")
}

// groupLogsByContract 按合约地址分组日志
func (m *EventMonitor) groupLogsByContract(logs []types.Log) map[common.Address][]types.Log {
	logsByContract := make(map[common.Address][]types.Log)

	for _, log := range logs {
		logsByContract[log.Address] = append(logsByContract[log.Address], log)
	}

	return logsByContract
}
