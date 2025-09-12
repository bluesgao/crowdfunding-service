package model

import (
	"time"
)

// ContributeRecordModel 贡献记录
type ContributeRecordModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectId int64  `json:"project_id" gorm:"not null"`
	Amount    int64  `json:"amount" gorm:"not null"`
	Address   string `json:"address" gorm:"not null"`
	TxHash    string `json:"tx_hash" gorm:"uniqueIndex"`
	BlockNum  int64  `json:"block_num"`
}

// TableName 自定义表名
func (ContributeRecordModel) TableName() string {
	return "contribute_record"
}
