package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

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
	startBlock    int64
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
	},
	{
		"inputs": [
			{"name": "title", "type": "string"},
			{"name": "description", "type": "string"},
			{"name": "targetAmount", "type": "uint256"},
			{"name": "startTime", "type": "uint256"},
			{"name": "endTime", "type": "uint256"},
			{"name": "creator", "type": "address"}
		],
		"name": "createProject",
		"outputs": [
			{"name": "projectId", "type": "uint256"}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

func Init(cfg config.EthereumConfig) (*Client, error) {
	// 验证RPC URL
	if cfg.RpcUrl == "" {
		return nil, fmt.Errorf("ethereum RPC URL is required")
	}

	// 连接以太坊客户端
	client, err := ethclient.Dial(cfg.RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network ID from ethereum client: %w", err)
	}

	fmt.Printf("Connected to Ethereum network with chainID: %s , RpcUrl: %s\n", chainID.String(), cfg.RpcUrl)

	// 验证私钥格式
	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required")
	}

	// 解析私钥
	privateKeyStr := strings.TrimPrefix(cfg.PrivateKey, "0x")
	if len(privateKeyStr) != 64 {
		return nil, fmt.Errorf("invalid private key length: expected 64 characters (32 bytes), got %d", len(privateKeyStr))
	}

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
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
func (c *Client) GetLatestBlock() (int64, error) {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return int64(header.Number.Uint64()), nil
}

// GetBlockByNumber 根据区块号获取区块
func (c *Client) GetBlockByNumber(blockNumber int64) (*types.Block, error) {
	return c.client.BlockByNumber(context.Background(), new(big.Int).SetInt64(blockNumber))
}

// GetLogs 获取指定区块范围内的日志
func (c *Client) GetLogs(fromBlock, toBlock int64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetInt64(fromBlock),
		ToBlock:   new(big.Int).SetInt64(toBlock),
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

	return int64(latestBlock) >= int64(receipt.BlockNumber.Uint64())+int64(c.confirmations), nil
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

// CreateProject 在智能合约中创建项目
func (c *Client) CreateProject(title, description string, targetAmount float64, startTime, endTime time.Time, creator common.Address) (common.Hash, error) {
	// 获取交易授权
	auth := c.GetAuth()

	// 设置gas限制
	auth.GasLimit = 300000

	// 将金额转换为Wei (假设使用ETH，1 ETH = 10^18 Wei)
	targetAmountWei := new(big.Int)
	targetAmountWei.SetString(fmt.Sprintf("%.0f", targetAmount*1e18), 10)

	// 将时间转换为Unix时间戳
	startTimestamp := big.NewInt(startTime.Unix())
	endTimestamp := big.NewInt(endTime.Unix())

	// 准备合约调用数据
	packedData, err := c.contractABI.Pack("createProject",
		title,
		description,
		targetAmountWei,
		startTimestamp,
		endTimestamp,
		creator,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack contract data: %w", err)
	}

	// 创建交易
	tx := types.NewTransaction(
		auth.Nonce.Uint64(),
		c.ContractAddr,
		big.NewInt(0), // 不发送ETH
		auth.GasLimit,
		auth.GasPrice,
		packedData,
	)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(1)), c.privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// 发送交易
	err = c.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

// GetStartBlock 获取起始区块号
func (c *Client) GetStartBlock() int64 {
	return c.startBlock
}

// GetCurrentBlockNumber 获取当前区块号
func (c *Client) GetCurrentBlockNumber() (int64, error) {
	return c.GetLatestBlock()
}

// GetBlockLogs 获取指定区块的日志
func (c *Client) GetBlockLogs(blockNum int64) ([]types.Log, error) {
	return c.GetLogs(blockNum, blockNum)
}

// GetBlockTransactions 获取指定区块的所有交易
func (c *Client) GetBlockTransactions(blockNumber int64) ([]*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	return block.Transactions(), nil
}

// GetBlockTransactionCount 获取指定区块的交易数量
func (c *Client) GetBlockTransactionCount(blockNumber int64) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return 0, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	return len(block.Transactions()), nil
}

// GetTransactionByHash 根据交易哈希获取交易详情
func (c *Client) GetTransactionByHash(txHash common.Hash) (*types.Transaction, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, isPending, err := c.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get transaction %s: %w", txHash.Hex(), err)
	}

	return tx, isPending, nil
}

// GetTransactionInBlock 获取区块中指定索引的交易
func (c *Client) GetTransactionInBlock(blockNumber int64, txIndex int) (*types.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", blockNumber, err)
	}

	transactions := block.Transactions()
	if txIndex >= len(transactions) {
		return nil, fmt.Errorf("transaction index %d out of range for block %d (has %d transactions)", txIndex, blockNumber, len(transactions))
	}

	return transactions[txIndex], nil
}

// GetBlockWithTransactions 获取包含交易详情的完整区块信息
func (c *Client) GetBlockWithTransactions(blockNumber int64) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return nil, fmt.Errorf("failed to get block %d with transactions: %w", blockNumber, err)
	}

	return block, nil
}

// GetBlockWithTransactionsByHash 根据区块哈希获取包含交易详情的完整区块信息
func (c *Client) GetBlockWithTransactionsByHash(blockHash common.Hash) (*types.Block, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	block, err := c.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s with transactions: %w", blockHash.Hex(), err)
	}

	return block, nil
}

// GetTransactionSender 获取交易的发送者地址
func (c *Client) GetTransactionSender(tx *types.Transaction) (common.Address, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chainID, err := c.client.NetworkID(ctx)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get network ID: %w", err)
	}

	sender, err := types.Sender(types.NewEIP155Signer(chainID), tx)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to get transaction sender: %w", err)
	}

	return sender, nil
}

// GetTransactionValue 获取交易的价值（ETH数量）
func (c *Client) GetTransactionValue(tx *types.Transaction) *big.Int {
	return tx.Value()
}

// GetTransactionGasPrice 获取交易的Gas价格
func (c *Client) GetTransactionGasPrice(tx *types.Transaction) *big.Int {
	return tx.GasPrice()
}

// GetTransactionGasLimit 获取交易的Gas限制
func (c *Client) GetTransactionGasLimit(tx *types.Transaction) uint64 {
	return tx.Gas()
}
