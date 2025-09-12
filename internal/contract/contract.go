package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract 统一合约包装器
type Contract struct {
	client   *ethclient.Client   // 以太坊客户端
	contract *bind.BoundContract // 合约绑定
	address  common.Address      // 合约地址
	abi      abi.ABI             // 合约ABI
	name     string              // 合约名称
	blockNum int64               // 合约部署的区块号
	mu       sync.RWMutex        // 读写锁，保护blockNum
}

// NewContract 创建合约实例
func NewContract(client *ethclient.Client, name string, cfg config.ContractConfig) (*Contract, error) {
	// 加载ABI
	abiData, err := os.ReadFile(cfg.ABIPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ABI from %s: %w", cfg.ABIPath, err)
	}

	// 尝试解析为完整的编译输出文件
	var compiledOutput struct {
		ABI json.RawMessage `json:"abi"`
	}

	var parsedABI abi.ABI

	// 首先尝试解析为完整编译输出
	if err := json.Unmarshal(abiData, &compiledOutput); err == nil && compiledOutput.ABI != nil {
		// 从编译输出中提取ABI
		parsedABI, err = abi.JSON(bytes.NewReader(compiledOutput.ABI))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI from compiled output: %w", err)
		}
	} else {
		// 如果不是完整编译输出，尝试直接解析为ABI数组
		parsedABI, err = abi.JSON(bytes.NewReader(abiData))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI: %w", err)
		}
	}

	// 解析合约地址
	contractAddr := common.HexToAddress(cfg.Address)

	// 创建合约实例
	contract := &Contract{
		client:   client,
		contract: bind.NewBoundContract(contractAddr, parsedABI, client, client, client),
		address:  contractAddr,
		abi:      parsedABI,
		name:     name,
		blockNum: 0, // 默认值，将在后台异步获取
	}

	// 异步获取合约部署区块号，避免启动时的API限制
	go func() {
		if deploymentBlock, err := getContractDeploymentBlock(client, contractAddr); err == nil {
			contract.mu.Lock()
			contract.blockNum = deploymentBlock
			contract.mu.Unlock()
			logger.Info("Successfully retrieved deployment block %d for contract %s", deploymentBlock, name)
		} else {
			logger.Warn("Failed to get deployment block for contract %s: %v", name, err)
		}
	}()

	return contract, nil
}

// GetAddress 获取合约地址
func (c *Contract) GetAddress() common.Address {
	return c.address
}

// GetABI 获取合约ABI
func (c *Contract) GetABI() abi.ABI {
	return c.abi
}

// GetName 获取合约名称
func (c *Contract) GetName() string {
	return c.name
}

// ParseEvent 解析事件日志
func (c *Contract) ParseEvent(log types.Log) (map[string]interface{}, error) {
	eventSignature := log.Topics[0].Hex()

	// 遍历ABI中的事件
	for eventName, event := range c.abi.Events {
		if event.ID.Hex() == eventSignature {
			return c.parseEvent(eventName, log, event)
		}
	}

	// 未知事件
	logger.Warn("Unknown event signature: %s in contract %s", eventSignature, c.name)
	return map[string]interface{}{
		"eventName":   "Unknown",
		"signature":   eventSignature,
		"contract":    c.name,
		"txHash":      log.TxHash.Hex(),
		"blockNumber": log.BlockNumber,
		"logIndex":    log.Index,
	}, nil
}

// parseEvent 解析事件
func (c *Contract) parseEvent(eventName string, log types.Log, event abi.Event) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["eventName"] = eventName
	result["contract"] = c.name
	result["txHash"] = log.TxHash.Hex()
	result["blockNumber"] = log.BlockNumber
	result["logIndex"] = log.Index

	// 解析索引参数
	if len(log.Topics) > 1 {
		for i, input := range event.Inputs {
			if input.Indexed && i+1 < len(log.Topics) {
				value, err := c.parseTopicValue(log.Topics[i+1], input.Type)
				if err != nil {
					logger.Warn("Failed to parse indexed parameter %s: %v", input.Name, err)
					continue
				}
				result[input.Name] = value
			}
		}
	}

	// 解析非索引参数
	if len(log.Data) > 0 {
		nonIndexedInputs := make([]abi.Argument, 0)
		for _, input := range event.Inputs {
			if !input.Indexed {
				nonIndexedInputs = append(nonIndexedInputs, input)
			}
		}

		if len(nonIndexedInputs) > 0 {
			values, err := c.abi.Unpack(eventName, log.Data)
			if err != nil {
				logger.Warn("Failed to unpack non-indexed parameters: %v", err)
			} else {
				for i, input := range nonIndexedInputs {
					if i < len(values) {
						result[input.Name] = values[i]
					}
				}
			}
		}
	}

	return result, nil
}

// parseTopicValue 解析主题值
func (c *Contract) parseTopicValue(topic common.Hash, t abi.Type) (interface{}, error) {
	switch t.T {
	case abi.UintTy:
		return new(big.Int).SetBytes(topic.Bytes()), nil
	case abi.IntTy:
		return new(big.Int).SetBytes(topic.Bytes()), nil
	case abi.AddressTy:
		return common.BytesToAddress(topic.Bytes()), nil
	case abi.BoolTy:
		return new(big.Int).SetBytes(topic.Bytes()).Cmp(big.NewInt(0)) > 0, nil
	case abi.BytesTy:
		return topic.Bytes(), nil
	default:
		return topic.Hex(), nil
	}
}

// GetBlockLogs 获取区块日志
func (c *Contract) GetBlockLogs(blockNum int64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []common.Address{c.address},
	}

	return c.client.FilterLogs(context.Background(), query)
}

// GetCurrentBlockNumber 获取当前区块号
func (c *Contract) GetCurrentBlockNumber() (int64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

// GetBlockNum 获取合约部署区块号
func (c *Contract) GetBlockNum() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blockNum
}

// getContractDeploymentBlock 获取合约部署区块号
func getContractDeploymentBlock(client *ethclient.Client, address common.Address) (int64, error) {
	// 通过获取合约代码来确定部署区块
	// 从最新区块开始向前搜索，直到找到合约代码
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}

	currentBlock := header.Number.Int64()

	// 向前搜索最多1000个区块
	maxSearchBlocks := int64(1000)
	startBlock := currentBlock - maxSearchBlocks
	if startBlock < 0 {
		startBlock = 0
	}

	for blockNum := currentBlock; blockNum >= startBlock; blockNum-- {
		code, err := client.CodeAt(context.Background(), address, big.NewInt(blockNum))
		if err != nil {
			continue
		}

		// 如果代码不为空，说明合约在这个区块存在
		if len(code) > 0 {
			return blockNum, nil
		}
	}

	return 0, fmt.Errorf("could not find deployment block for contract %s", address.Hex())
}
