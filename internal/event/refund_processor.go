package event

import (
	"log"
	"math/big"

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
func (p *RefundProcessor) Process(event *model.Event, eventData map[string]interface{}) error {
	// 创建退款记录
	refundee := eventData["refundee"].(string)
	amount := eventData["amount"].(*big.Int)
	reason := eventData["reason"].(string)

	refundRecord := model.RefundRecord{
		ProjectID:    event.ProjectID,
		Amount:       float64(amount.Int64()) / 1e18, // 转换为ETH
		Address:      refundee,
		TxHash:       event.TxHash,
		BlockNum:     event.BlockNum,
		Status:       string(model.RefundStatusSuccess),
		RefundReason: reason,
	}

	// 通过logic层创建退款记录
	if err := p.refundLogic.CreateRefundRecord(&refundRecord); err != nil {
		log.Printf("Failed to create refund record: %v", err)
		return err
	}

	log.Printf("Processed refund: %f ETH to %s for project %d",
		refundRecord.Amount, refundee, event.ProjectID)

	return nil
}
