package model

import (
	"time"

	"gorm.io/gorm"
)

// SettlementRecord 结算记录
type SettlementRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID      uint       `json:"project_id" gorm:"not null"`
	TotalAmount    float64    `json:"total_amount" gorm:"not null"`   // 总金额
	SettledAmount  float64    `json:"settled_amount" gorm:"not null"` // 已结算金额
	PlatformFee    float64    `json:"platform_fee" gorm:"default:0"`  // 平台手续费
	CreatorAmount  float64    `json:"creator_amount" gorm:"not null"` // 创建者获得金额
	TxHash         string     `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum       uint64     `json:"block_num"`
	Status         string     `json:"status" gorm:"default:'pending'"` // pending, success, failed
	SettlementType string     `json:"settlement_type" gorm:"not null"` // success, failed, partial
	SettlementTime *time.Time `json:"settlement_time"`

	// 关联
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// SettlementStatus 结算状态
type SettlementStatus string

const (
	SettlementStatusPending SettlementStatus = "pending" // 待处理
	SettlementStatusSuccess SettlementStatus = "success" // 成功
	SettlementStatusFailed  SettlementStatus = "failed"  // 失败
)

// SettlementType 结算类型
type SettlementType string

const (
	SettlementTypeSuccess SettlementType = "success" // 成功结算
	SettlementTypeFailed  SettlementType = "failed"  // 失败结算
	SettlementTypePartial SettlementType = "partial" // 部分结算
)
