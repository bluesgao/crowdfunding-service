package model

import (
	"time"

	"gorm.io/gorm"
)

// ProjectTeam 项目团队
type ProjectTeam struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	ProjectID  uint      `json:"project_id" gorm:"not null"`
	MemberName string    `json:"member_name" gorm:"not null"`
	MemberRole string    `json:"member_role" gorm:"not null"` // creator, developer, designer, marketer, advisor
	Address    string    `json:"address" gorm:"not null"`     // 成员钱包地址
	Email      string    `json:"email"`
	Bio        string    `json:"bio" gorm:"type:text"`
	AvatarURL  string    `json:"avatar_url"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	JoinTime   time.Time `json:"join_time"`

	// 关联
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// TeamRole 团队角色
type TeamRole string

const (
	TeamRoleCreator   TeamRole = "creator"   // 创建者
	TeamRoleDeveloper TeamRole = "developer" // 开发者
	TeamRoleDesigner  TeamRole = "designer"  // 设计师
	TeamRoleMarketer  TeamRole = "marketer"  // 市场推广
	TeamRoleAdvisor   TeamRole = "advisor"   // 顾问
)
