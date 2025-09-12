package model

import (
	"time"
)

// ProjectTeamModel 项目团队
type ProjectTeamModel struct {
	Id        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ProjectId  int64     `json:"project_id" gorm:"not null"`
	MemberName string    `json:"member_name" gorm:"not null"`
	MemberRole string    `json:"member_role" gorm:"not null"` // creator, developer, designer, marketer, advisor
	Address    string    `json:"address" gorm:"not null"`     // 成员钱包地址
	Email      string    `json:"email"`
	Bio        string    `json:"bio" gorm:"type:text"`
	AvatarURL  string    `json:"avatar_url"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	JoinTime   time.Time `json:"join_time"`
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

// TableName 自定义表名
func (ProjectTeamModel) TableName() string {
	return "project_team"
}
