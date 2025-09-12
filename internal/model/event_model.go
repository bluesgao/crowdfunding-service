package model

import (
	"time"
)

// EventModel 链上事件记录
type EventModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ContractAddress string `json:"contract_address" gorm:"not null"`
	ContractName    string `json:"contract_name" gorm:"not null"`
	EventType       string `json:"event_type" gorm:"not null"`
	TxHash          string `json:"tx_hash" gorm:"not null"`
	BlockNum        int64  `json:"block_num" gorm:"not null"`
	LogIndex        int64  `json:"log_index"`
	Data            string `json:"data" gorm:"type:text"`
	Processed       bool   `json:"processed" gorm:"default:false"`
}

// TableName 自定义表名
func (EventModel) TableName() string {
	return "event"
}
