package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/blues/cfs/internal/config"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	client        *ethclient.Client
	privateKey    *ecdsa.PrivateKey
	ContractAddr  common.Address
	startBlock    uint64
	confirmations int
	contractABI   abi.ABI
}

// 众筹合约ABI定义（简化版）
const contractABI = `[
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "projectId", "type": "uint256"},
			{"indexed": true, "name": "contributor", "type": "address"},
			{"indexed": false, "name": "amount", "type": "uint256"}
		],
		"name": "ContributionMade",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "projectId", "type": "uint256"},
			{"indexed": false, "name": "status", "type": "uint8"}
		],
		"name": "ProjectStatusChanged",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{"indexed": true, "name": "projectId", "type": "uint256"},
			{"indexed": false, "name": "title", "type": "string"},
			{"indexed": false, "name": "targetAmount", "type": "uint256"},
			{"indexed": false, "name": "creator", "type": "address"}
		],
		"name": "ProjectCreated",
		"type": "event"
	}
]`

func Init(cfg config.EthereumConfig) (*Client, error) {
	// 连接以太坊客户端
	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	// 解析私钥
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(cfg.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// 解析合约地址
	contractAddr := common.HexToAddress(cfg.ContractAddr)

	// 解析ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &Client{
		client:        client,
		privateKey:    privateKey,
		startBlock:    cfg.StartBlock,
		ContractAddr:  contractAddr,
		confirmations: cfg.Confirmations,
		contractABI:   parsedABI,
	}, nil
}

// GetLatestBlock 获取最新区块号
func (c *Client) GetLatestBlock() (uint64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return header.Number.Uint64(), nil
}

// GetBlockByNumber 根据区块号获取区块
func (c *Client) GetBlockByNumber(blockNumber uint64) (*types.Block, error) {
	return c.client.BlockByNumber(context.Background(), new(big.Int).SetUint64(blockNumber))
}

// GetLogs 获取指定区块范围内的日志
func (c *Client) GetLogs(fromBlock, toBlock uint64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		ToBlock:   new(big.Int).SetUint64(toBlock),
		Addresses: []common.Address{c.ContractAddr},
	}

	return c.client.FilterLogs(context.Background(), query)
}

// ParseEvent 解析事件日志
func (c *Client) ParseEvent(log types.Log) (map[string]interface{}, error) {
	// 检查事件签名
	eventSignature := log.Topics[0].Hex()

	switch eventSignature {
	case c.contractABI.Events["ContributionMade"].ID.Hex():
		return c.parseContributionMadeEvent(log)
	case c.contractABI.Events["ProjectStatusChanged"].ID.Hex():
		return c.parseProjectStatusChangedEvent(log)
	case c.contractABI.Events["ProjectCreated"].ID.Hex():
		return c.parseProjectCreatedEvent(log)
	default:
		return nil, fmt.Errorf("unknown event signature: %s", eventSignature)
	}
}

// parseContributionMadeEvent 解析贡献事件
func (c *Client) parseContributionMadeEvent(log types.Log) (map[string]interface{}, error) {
	event := make(map[string]interface{})

	// 解析索引参数
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("invalid ContributionMade event: insufficient topics")
	}

	event["type"] = "ContributionMade"
	event["projectId"] = new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64()
	event["contributor"] = common.BytesToAddress(log.Topics[2].Bytes()).Hex()

	// 解析非索引参数
	if len(log.Data) > 0 {
		amount := new(big.Int).SetBytes(log.Data)
		event["amount"] = amount
	}

	event["txHash"] = log.TxHash.Hex()
	event["blockNumber"] = log.BlockNumber
	event["logIndex"] = log.Index

	return event, nil
}

// parseProjectStatusChangedEvent 解析项目状态变更事件
func (c *Client) parseProjectStatusChangedEvent(log types.Log) (map[string]interface{}, error) {
	event := make(map[string]interface{})

	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("invalid ProjectStatusChanged event: insufficient topics")
	}

	event["type"] = "ProjectStatusChanged"
	event["projectId"] = new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64()

	if len(log.Data) > 0 {
		status := new(big.Int).SetBytes(log.Data).Uint64()
		event["status"] = status
	}

	event["txHash"] = log.TxHash.Hex()
	event["blockNumber"] = log.BlockNumber
	event["logIndex"] = log.Index

	return event, nil
}

// parseProjectCreatedEvent 解析项目创建事件
func (c *Client) parseProjectCreatedEvent(log types.Log) (map[string]interface{}, error) {
	event := make(map[string]interface{})

	if len(log.Topics) < 2 {
		return nil, fmt.Errorf("invalid ProjectCreated event: insufficient topics")
	}

	event["type"] = "ProjectCreated"
	event["projectId"] = new(big.Int).SetBytes(log.Topics[1].Bytes()).Uint64()

	// 解析非索引参数（title, targetAmount, creator）
	// 这里需要根据实际ABI进行解析
	event["txHash"] = log.TxHash.Hex()
	event["blockNumber"] = log.BlockNumber
	event["logIndex"] = log.Index

	return event, nil
}

// GetTransactionReceipt 获取交易回执
func (c *Client) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	return c.client.TransactionReceipt(context.Background(), txHash)
}

// IsTransactionConfirmed 检查交易是否已确认
func (c *Client) IsTransactionConfirmed(txHash common.Hash) (bool, error) {
	receipt, err := c.GetTransactionReceipt(txHash)
	if err != nil {
		return false, err
	}

	if receipt == nil {
		return false, nil
	}

	latestBlock, err := c.GetLatestBlock()
	if err != nil {
		return false, err
	}

	return latestBlock >= receipt.BlockNumber.Uint64()+uint64(c.confirmations), nil
}

// GetAccountAddress 获取账户地址
func (c *Client) GetAccountAddress() common.Address {
	return crypto.PubkeyToAddress(c.privateKey.PublicKey)
}

// GetAuth 获取交易授权
func (c *Client) GetAuth() *bind.TransactOpts {
	auth, _ := bind.NewKeyedTransactorWithChainID(c.privateKey, big.NewInt(1)) // 主网
	return auth
}
