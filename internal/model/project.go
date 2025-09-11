package model

import (
	"time"

	"gorm.io/gorm"
)

// Project 众筹项目模型
type Project struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 基本信息
	Title       string `json:"title" gorm:"not null" binding:"required"`
	Description string `json:"description" gorm:"type:text"`
	ImageURL    string `json:"image_url"`
	Category    string `json:"category"`

	// 众筹信息
	TargetAmount    float64 `json:"target_amount" gorm:"not null" binding:"required,min=0"`
	CurrentAmount   float64 `json:"current_amount" gorm:"default:0"`
	MinContribution float64 `json:"min_contribution" gorm:"default:0"`
	MaxContribution float64 `json:"max_contribution" gorm:"default:0"`

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

	// 关联
	Contributions []ContributeRecord `json:"contributions,omitempty" gorm:"foreignKey:ProjectID"`
	Events        []Event            `json:"events,omitempty" gorm:"foreignKey:ProjectID"`
}

// ProjectStatus 项目状态
type ProjectStatus string

const (
	ProjectStatusPending   ProjectStatus = "pending"   // 待开始
	ProjectStatusActive    ProjectStatus = "active"    // 进行中
	ProjectStatusSuccess   ProjectStatus = "success"   // 成功
	ProjectStatusFailed    ProjectStatus = "failed"    // 失败
	ProjectStatusCancelled ProjectStatus = "cancelled" // 已取消
)
