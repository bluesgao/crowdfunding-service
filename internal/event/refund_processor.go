package event

import (
	"math/big"

	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
)

// RefundProcessor 退款事件处理器
type RefundProcessor struct {
	refundLogic *logic.RefundRecordLogic
}

// NewRefundProcessor 创建退款事件处理器
func NewRefundProcessor(refundLogic *logic.RefundRecordLogic) *RefundProcessor {
	return &RefundProcessor{
		refundLogic: refundLogic,
	}
}

// Process 处理退款事件
func (p *RefundProcessor) Process(event *model.EventModel, eventData map[string]interface{}) error {
	// 创建退款记录
	refundee := eventData["refundee"].(string)
	amount := eventData["amount"].(*big.Int)
	reason := eventData["reason"].(string)

	refundRecord := model.RefundRecordModel{
		ProjectId:    event.ProjectId,
		Amount:       amount.Int64(), // 保持wei单位
		Address:      refundee,
		TxHash:       event.TxHash,
		BlockNum:     event.BlockNum,
		Status:       string(model.RefundStatusSuccess),
		RefundReason: reason,
	}

	// 通过logic层创建退款记录
	if err := p.refundLogic.CreateRefundRecord(&refundRecord); err != nil {
		logger.Error("Failed to create refund record: %v", err)
		return err
	}

	logger.Info("Processed refund: %f ETH to %s for project %d",
		refundRecord.Amount, refundee, event.ProjectId)

	return nil
}
