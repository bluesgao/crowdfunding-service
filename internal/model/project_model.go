package model

import (
	"time"
)

// ProjectModel 众筹项目模型
type ProjectModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 基本信息
	Title       string `json:"title" gorm:"not null" binding:"required"`
	Description string `json:"description" gorm:"type:text"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`

	// 众筹信息
	TargetAmount  int64 `json:"target_amount" gorm:"not null" binding:"required,min=0"`
	CurrentAmount int64 `json:"current_amount" gorm:"default:0"`
	MinAmount     int64 `json:"min_amount" gorm:"default:0"`
	MaxAmount     int64 `json:"max_amount" gorm:"default:0"`

	// 时间信息
	StartTime time.Time `json:"start_time" gorm:"not null"`
	EndTime   time.Time `json:"end_time" gorm:"not null"`

	// 状态
	Status ProjectStatus `json:"status" gorm:"default:'pending'"`

	// 创建者信息
	CreatorAddress string `json:"creator_address" gorm:"not null"`
	CreatorName    string `json:"creator_name"`

	// 区块链信息
	ContractAddress string `json:"contract_address"`
	TransactionHash string `json:"transaction_hash"`
}

// ProjectStatus 项目状态
type ProjectStatus string

const (
	ProjectStatusPending   ProjectStatus = "pending"   // 待开始
	ProjectStatusDeploying ProjectStatus = "deploying" // 待上链
	ProjectStatusActive    ProjectStatus = "active"    // 进行中
	ProjectStatusSuccess   ProjectStatus = "success"   // 成功
	ProjectStatusFailed    ProjectStatus = "failed"    // 失败
	ProjectStatusCancelled ProjectStatus = "cancelled" // 已取消
)

// TableName 自定义表名
func (ProjectModel) TableName() string {
	return "project"
}
