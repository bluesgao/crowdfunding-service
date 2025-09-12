package model

import (
	"time"
)

// SettlementRecordModel 结算记录
type SettlementRecordModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectId      int64      `json:"project_id" gorm:"not null"`
	TotalAmount    int64      `json:"total_amount" gorm:"not null"`   // 总金额
	SettledAmount  int64      `json:"settled_amount" gorm:"not null"` // 已结算金额
	PlatformFee    int64      `json:"platform_fee" gorm:"default:0"`  // 平台手续费
	CreatorAmount  int64      `json:"creator_amount" gorm:"not null"` // 创建者获得金额
	TxHash         string     `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum       int64      `json:"block_num"`
	Status         string     `json:"status" gorm:"default:'pending'"` // pending, success, failed
	SettlementType string     `json:"settlement_type" gorm:"not null"` // success, failed, partial
	SettlementTime *time.Time `json:"settlement_time"`
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

// TableName 自定义表名
func (SettlementRecordModel) TableName() string {
	return "settlement_record"
}
