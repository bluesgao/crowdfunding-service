package contract

import (
	"fmt"
	"sync"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Manager 合约管理器
type ContractManager struct {
	mu        sync.RWMutex
	contracts map[string]*Contract
	client    *ethclient.Client
}

// NewManager 创建合约管理器
func NewContractManager(cfg config.EthereumConfig) (*ContractManager, error) {
	// 连接以太坊客户端
	client, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	manager := &ContractManager{
		contracts: make(map[string]*Contract),
		client:    client,
	}

	// 初始化所有启用的合约
	if err := manager.initContracts(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize contracts: %w", err)
	}

	return manager, nil
}

// initContracts 初始化所有合约
func (m *ContractManager) initContracts(cfg config.EthereumConfig) error {
	for name, contractCfg := range cfg.Contracts {
		if !contractCfg.Enabled {
			logger.Info("Contract %s is disabled, skipping", name)
			continue
		}

		logger.Info("Initializing contract: %s at %s", name, contractCfg.Address)

		// 创建合约实例
		contract, err := NewContract(m.client, name, contractCfg)
		if err != nil {
			logger.Error("Failed to initialize contract %s: %v", name, err)
			continue
		}

		// 注册合约
		m.contracts[name] = contract
		logger.Info("Successfully registered contract: %s", name)
	}

	return nil
}

// GetContract 获取指定名称的合约
func (m *ContractManager) GetContract(name string) (*Contract, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contract, exists := m.contracts[name]
	if !exists {
		return nil, fmt.Errorf("contract %s not found", name)
	}

	return contract, nil
}

// GetAllContracts 获取所有合约
func (m *ContractManager) GetAllContracts() map[string]*Contract {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Contract)
	for name, contract := range m.contracts {
		result[name] = contract
	}

	return result
}
