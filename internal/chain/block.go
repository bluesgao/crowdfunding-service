package chain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Block 区块操作工具类
type Block struct{}

// NewBlock 创建区块工具类实例
func NewBlock() *Block {
	return &Block{}
}

// GetBlockLogs 获取指定区块的日志
func (b *Block) GetBlockLogs(client *ethclient.Client, contractAddress common.Address, blockNum int64) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []common.Address{contractAddress},
	}

	return client.FilterLogs(context.Background(), query)
}

// GetCurrentBlockNumber 获取当前最新区块号
func (b *Block) GetCurrentBlockNumber(client *ethclient.Client) (int64, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}
