package model

import (
	"time"
)

// RefundRecordModel 退款记录
type RefundRecordModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectId    int64  `json:"project_id" gorm:"not null"`
	ContributeID int64  `json:"contribute_id" gorm:"not null"`
	Amount       int64  `json:"amount" gorm:"not null"`
	Address      string `json:"address" gorm:"not null"`
	TxHash       string `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum     int64  `json:"block_num"`
	Status       string `json:"status" gorm:"default:'pending'"` // pending, success, failed
	RefundReason string `json:"refund_reason" gorm:"type:text"`
}

// RefundStatus 退款状态
type RefundStatus string

const (
	RefundStatusPending RefundStatus = "pending" // 待处理
	RefundStatusSuccess RefundStatus = "success" // 成功
	RefundStatusFailed  RefundStatus = "failed"  // 失败
)

// TableName 自定义表名
func (RefundRecordModel) TableName() string {
	return "refund_record"
}
