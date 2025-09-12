package event

import (
	"math/big"

	"github.com/blues/cfs/internal/logger"
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
func (p *ContributeProcessor) Process(event *model.EventModel, eventData map[string]interface{}) error {
	// 创建贡献记录
	contributor := eventData["contributor"].(string)
	amount := eventData["amount"].(*big.Int)

	contribution := model.ContributeRecordModel{
		ProjectId: event.ProjectId,
		Amount:    amount.Int64(), // 保持wei单位
		Address:   contributor,
		TxHash:    event.TxHash,
		BlockNum:  event.BlockNum,
	}

	// 通过logic层创建贡献记录
	if err := p.contributeLogic.CreateContributeRecord(&contribution); err != nil {
		logger.Error("Failed to create contribution record: %v", err)
		return err
	}

	logger.Info("Processed contribution: %f ETH from %s to project %d",
		contribution.Amount, contributor, event.ProjectId)

	return nil
}
