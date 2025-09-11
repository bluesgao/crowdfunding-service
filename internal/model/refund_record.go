package model

import (
	"time"

	"gorm.io/gorm"
)

// RefundRecord 退款记录
type RefundRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID    uint    `json:"project_id" gorm:"not null"`
	ContributeID uint    `json:"contribute_id" gorm:"not null"`
	Amount       float64 `json:"amount" gorm:"not null"`
	Address      string  `json:"address" gorm:"not null"`
	TxHash       string  `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum     uint64  `json:"block_num"`
	Status       string  `json:"status" gorm:"default:'pending'"` // pending, success, failed
	RefundReason string  `json:"refund_reason" gorm:"type:text"`

	// 关联
	Project    Project          `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Contribute ContributeRecord `json:"contribute,omitempty" gorm:"foreignKey:ContributeID"`
}

// RefundStatus 退款状态
type RefundStatus string

const (
	RefundStatusPending RefundStatus = "pending" // 待处理
	RefundStatusSuccess RefundStatus = "success" // 成功
	RefundStatusFailed  RefundStatus = "failed"  // 失败
)
