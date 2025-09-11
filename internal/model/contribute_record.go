package model

import (
	"time"

	"gorm.io/gorm"
)

// ContributeRecord 贡献记录
type ContributeRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID uint    `json:"project_id" gorm:"not null"`
	Amount    float64 `json:"amount" gorm:"not null"`
	Address   string  `json:"address" gorm:"not null"`
	TxHash    string  `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum  uint64  `json:"block_num"`

	// 关联
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}
