package model

import (
	"time"

	"gorm.io/gorm"
)

// Event 链上事件记录
type Event struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID uint   `json:"project_id"`
	EventType string `json:"event_type" gorm:"not null"`
	TxHash    string `json:"tx_hash" gorm:"not null"`
	BlockNum  uint64 `json:"block_num" gorm:"not null"`
	LogIndex  uint   `json:"log_index"`
	Data      string `json:"data" gorm:"type:text"`
	Processed bool   `json:"processed" gorm:"default:false"`

	// 关联
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}
