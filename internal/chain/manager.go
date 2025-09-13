package chain

import (
	"context"
	"fmt"
	"sync"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Manager 单链管理器
type Manager struct {
	mu        sync.RWMutex
	contracts map[string]*Contract // 合约映射: "contractName" -> Contract
	client    *ethclient.Client    // 链客户端
	config    config.ChainConfig   // 存储链配置
}

// NewManager 创建单链管理器
func NewManager(cfg config.ChainConfig) (*Manager, error) {
	manager := &Manager{
		contracts: make(map[string]*Contract),
		client:    nil, // 将在初始化时创建
		config:    cfg,
	}

	// 初始化客户端
	if err := manager.initClient(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	// 初始化所有启用的合约
	if err := manager.initContracts(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize contracts: %w", err)
	}

	return manager, nil
}

// initClient 初始化客户端
func (m *Manager) initClient(cfg config.ChainConfig) error {
	logger.Info("Initializing chain client (type: %s, id: %d)", cfg.ChainType, cfg.ChainId)

	// 创建客户端
	client, err := m.createChainClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// 保存客户端
	m.client = client
	logger.Info("Successfully initialized client")

	return nil
}

// initContracts 初始化所有合约
func (m *Manager) initContracts(cfg config.ChainConfig) error {
	var initErrors []error

	// 遍历所有合约
	for contractName, contractCfg := range cfg.Contracts {
		if !contractCfg.Enabled {
			logger.Info("Skipping disabled contract: %s", contractName)
			continue
		}

		logger.Info("Initializing contract: %s (address: %s)", contractName, contractCfg.Address)

		// 创建合约实例
		contract, err := NewContract(m.client, contractName, contractCfg, cfg)
		if err != nil {
			logger.Error("Failed to create contract %s: %v", contractName, err)
			initErrors = append(initErrors, fmt.Errorf("failed to create contract %s: %w", contractName, err))
			continue
		}

		// 存储合约
		m.contracts[contractName] = contract
		logger.Info("Successfully initialized contract: %s", contractName)
	}

	// 如果有错误，返回第一个错误
	if len(initErrors) > 0 {
		return initErrors[0]
	}

	logger.Info("Successfully initialized %d contracts", len(m.contracts))
	return nil
}

// createChainClient 创建链客户端
func (m *Manager) createChainClient(cfg config.ChainConfig) (*ethclient.Client, error) {
	rpcUrl := cfg.RpcUrl
	if rpcUrl == "" {
		return nil, fmt.Errorf("no RPC URL configured")
	}

	// 验证链类型
	supportedTypes := []string{"ethereum", "polygon", "bsc", "arbitrum", "optimism"}
	isSupported := false
	for _, supportedType := range supportedTypes {
		if cfg.ChainType == supportedType {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return nil, fmt.Errorf("unsupported chain type %s, supported types: ethereum, polygon, bsc, arbitrum, optimism", cfg.ChainType)
	}

	// 根据链类型创建客户端
	logger.Info("Creating %s client connection (RPC: %s)", cfg.ChainType, rpcUrl)
	client, err := m.createChainTypeClient(cfg.ChainType, rpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", cfg.ChainType, err)
	}

	// 测试连接
	if err := m.testClientConnection(client); err != nil {
		client.Close()
		return nil, fmt.Errorf("client connection test failed (%s): %w", cfg.ChainType, err)
	}

	logger.Info("Successfully created %s client", cfg.ChainType)
	return client, nil
}

// createChainTypeClient 根据链类型创建客户端
func (m *Manager) createChainTypeClient(chainType, rpcUrl string) (*ethclient.Client, error) {
	switch chainType {
	case "ethereum":
		return ethclient.Dial(rpcUrl)
	case "polygon":
		return ethclient.Dial(rpcUrl)
	case "bsc":
		return ethclient.Dial(rpcUrl)
	case "arbitrum":
		return ethclient.Dial(rpcUrl)
	case "optimism":
		return ethclient.Dial(rpcUrl)
	default:
		return nil, fmt.Errorf("unsupported chain type: %s", chainType)
	}
}

// testClientConnection 测试客户端连接
func (m *Manager) testClientConnection(client *ethclient.Client) error {
	// 尝试获取最新区块号
	_, err := client.BlockNumber(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get block number: %w", err)
	}
	return nil
}

// GetClient 获取客户端
func (m *Manager) GetClient() *ethclient.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

// GetContract 获取指定合约
func (m *Manager) GetContract(contractName string) (*Contract, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contract, exists := m.contracts[contractName]
	if !exists {
		return nil, fmt.Errorf("contract %s not found", contractName)
	}

	return contract, nil
}

// GetContracts 获取所有合约
func (m *Manager) GetContracts() map[string]*Contract {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本以避免并发修改
	contracts := make(map[string]*Contract)
	for name, contract := range m.contracts {
		contracts[name] = contract
	}

	return contracts
}

// GetContractNames 获取所有合约名称
func (m *Manager) GetContractNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.contracts))
	for name := range m.contracts {
		names = append(names, name)
	}

	return names
}

// GetConfig 获取链配置
func (m *Manager) GetConfig() config.ChainConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetChainId 获取链ID
func (m *Manager) GetChainId() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.ChainId
}

// GetChainType 获取链类型
func (m *Manager) GetChainType() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.ChainType
}

// IsContractRegistered 检查合约是否已注册
func (m *Manager) IsContractRegistered(contractName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.contracts[contractName]
	return exists
}

// GetHealthStatus 获取健康状态
func (m *Manager) GetHealthStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := map[string]interface{}{
		"chain_type":    m.config.ChainType,
		"chain_id":      m.config.ChainId,
		"client_status": "connected",
		"contracts":     make(map[string]interface{}),
	}

	// 检查客户端连接状态
	if m.client != nil {
		if _, err := m.client.BlockNumber(context.TODO()); err != nil {
			health["client_status"] = "disconnected"
		}
	} else {
		health["client_status"] = "not_initialized"
	}

	// 检查每个合约的状态
	for contractName, contract := range m.contracts {
		contractHealth := map[string]interface{}{
			"enabled":   true, // 工具类合约默认启用
			"address":   contract.GetAddress().Hex(),
			"chain_id":  contract.GetChainId(),
			"block_num": contract.GetBlockNum(),
		}
		health["contracts"].(map[string]interface{})[contractName] = contractHealth
	}

	return health
}

// Close 关闭管理器
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		m.client.Close()
	}

	logger.Info("Chain manager closed")
	return nil
}
