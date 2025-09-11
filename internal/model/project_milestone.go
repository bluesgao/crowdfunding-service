package model

import (
	"time"

	"gorm.io/gorm"
)

// ProjectMilestone 项目里程碑
type ProjectMilestone struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID     uint       `json:"project_id" gorm:"not null"`
	Title         string     `json:"title" gorm:"not null"`
	Description   string     `json:"description" gorm:"type:text"`
	TargetDate    time.Time  `json:"target_date" gorm:"not null"`
	CompletedDate *time.Time `json:"completed_date"`
	Status        string     `json:"status" gorm:"default:'pending'"`  // pending, in_progress, completed, delayed
	Progress      int        `json:"progress" gorm:"default:0"`        // 进度百分比 0-100
	Priority      string     `json:"priority" gorm:"default:'medium'"` // low, medium, high, critical
	IsPublic      bool       `json:"is_public" gorm:"default:true"`    // 是否公开显示

	// 关联
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// MilestoneStatus 里程碑状态
type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"     // 待开始
	MilestoneStatusInProgress MilestoneStatus = "in_progress" // 进行中
	MilestoneStatusCompleted  MilestoneStatus = "completed"   // 已完成
	MilestoneStatusDelayed    MilestoneStatus = "delayed"     // 延期
)

// MilestonePriority 里程碑优先级
type MilestonePriority string

const (
	MilestonePriorityLow      MilestonePriority = "low"      // 低优先级
	MilestonePriorityMedium   MilestonePriority = "medium"   // 中等优先级
	MilestonePriorityHigh     MilestonePriority = "high"     // 高优先级
	MilestonePriorityCritical MilestonePriority = "critical" // 关键优先级
)
