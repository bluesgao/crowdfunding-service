package event

import (
	"log"
	"math/big"

	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
)

// ContributeProcessor 贡献事件处理器
type ContributeProcessor struct {
	contributeLogic *logic.ContributeRecordLogic
}

// NewContributeProcessor 创建贡献事件处理器
func NewContributeProcessor(contributeLogic *logic.ContributeRecordLogic) *ContributeProcessor {
	return &ContributeProcessor{
		contributeLogic: contributeLogic,
	}
}

// Process 处理贡献事件
func (p *ContributeProcessor) Process(event *model.Event, eventData map[string]interface{}) error {
	// 创建贡献记录
	contributor := eventData["contributor"].(string)
	amount := eventData["amount"].(*big.Int)

	contribution := model.ContributeRecord{
		ProjectID: event.ProjectID,
		Amount:    float64(amount.Int64()) / 1e18, // 转换为ETH
		Address:   contributor,
		TxHash:    event.TxHash,
		BlockNum:  event.BlockNum,
	}

	// 通过logic层创建贡献记录
	if err := p.contributeLogic.CreateContributeRecord(&contribution); err != nil {
		log.Printf("Failed to create contribution record: %v", err)
		return err
	}

	log.Printf("Processed contribution: %f ETH from %s to project %d",
		contribution.Amount, contributor, event.ProjectID)

	return nil
}
