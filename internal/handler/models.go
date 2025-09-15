package handler

import (
	"time"

	"github.com/blues/cfs/internal/model"
)

// 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 分页信息结构
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"pageSize"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"totalPage"`
}

// 项目相关响应模型

// ProjectTeamResponse 项目团队响应模型
type ProjectTeamResponse struct {
	ID         uint   `json:"id"`
	ProjectID  uint   `json:"projectId"`
	MemberName string `json:"memberName"`
	MemberRole string `json:"memberRole"`
	Address    string `json:"address"`
	Email      string `json:"email"`
	Bio        string `json:"bio"`
	AvatarURL  string `json:"avatarUrl"`
	IsActive   bool   `json:"isActive"`
}

// ProjectMilestoneResponse 项目里程碑响应模型
type ProjectMilestoneResponse struct {
	ID            uint       `json:"id"`
	ProjectID     uint       `json:"projectId"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	TargetDate    time.Time  `json:"targetDate"`
	CompletedDate *time.Time `json:"completedDate"`
	Status        string     `json:"status"`
	Progress      int        `json:"progress"`
	Priority      string     `json:"priority"`
}

// ProjectResponse 项目响应模型
type ProjectResponse struct {
	ID               uint                       `json:"id"`
	Title            string                     `json:"title"`
	Description      string                     `json:"description"`
	ImageURL         string                     `json:"imageUrl"`
	Category         string                     `json:"category"`
	Creator          string                     `json:"creator"`
	TargetAmount     int64                      `json:"targetAmount"`
	CurrentAmount    int64                      `json:"currentAmount"`
	MinAmount        int64                      `json:"minAmount"`
	MaxAmount        int64                      `json:"maxAmount"`
	Status           string                     `json:"status"`
	StartTime        time.Time                  `json:"startTime"`
	EndTime          time.Time                  `json:"endTime"`
	CreatedAt        time.Time                  `json:"createdAt"`
	UpdatedAt        time.Time                  `json:"updatedAt"`
	ProjectTeam      []ProjectTeamResponse      `json:"projectTeam"`
	ProjectMilestone []ProjectMilestoneResponse `json:"projectMilestone"`
}

// CreateProjectResponse 创建项目响应
type CreateProjectResponse struct {
	Project ProjectResponse `json:"project"`
}

// GetProjectsResponse 获取项目列表响应
type GetProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

// GetProjectResponse 获取项目详情响应
type GetProjectResponse struct {
	Project ProjectResponse `json:"project"`
}

// GetProjectStatsResponse 获取项目统计响应
type GetProjectStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// AllProjectStatsResponse 所有项目统计响应
type AllProjectStatsResponse struct {
	TotalProjects     int64  `json:"totalProjects"`
	PendingProjects   int64  `json:"pendingProjects"`
	DeployingProjects int64  `json:"deployingProjects"`
	ActiveProjects    int64  `json:"activeProjects"`
	SuccessProjects   int64  `json:"successProjects"`
	FailedProjects    int64  `json:"failedProjects"`
	CancelledProjects int64  `json:"cancelledProjects"`
	TotalRaised       string `json:"totalRaised"`
	TotalInvestors    int64  `json:"totalInvestors"`
}

// GetAllProjectStatsResponse 获取所有项目统计响应
type GetAllProjectStatsResponse struct {
	Stats AllProjectStatsResponse `json:"stats"`
}

// 贡献记录相关响应模型

// ContributeRecordResponse 贡献记录响应模型
type ContributeRecordResponse struct {
	ID        uint      `json:"id"`
	ProjectID uint      `json:"projectId"`
	Address   string    `json:"address"`
	Amount    int64     `json:"amount"`
	TxHash    string    `json:"txHash"`
	BlockNum  int64     `json:"blockNum"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetProjectContributeRecordsResponse 获取项目贡献记录响应
type GetProjectContributeRecordsResponse struct {
	Records    []ContributeRecordResponse `json:"records"`
	Pagination Pagination                 `json:"pagination"`
}

// GetContributeStatsResponse 获取贡献统计响应
type GetContributeStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// 退款记录相关响应模型

// RefundRecordResponse 退款记录响应模型
type RefundRecordResponse struct {
	ID        uint      `json:"id"`
	ProjectID uint      `json:"projectId"`
	Address   string    `json:"address"`
	Amount    int64     `json:"amount"`
	TxHash    string    `json:"txHash"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// GetProjectRefundsResponse 获取项目退款记录响应
type GetProjectRefundsResponse struct {
	Refunds    []RefundRecordResponse `json:"refunds"`
	Pagination Pagination             `json:"pagination"`
}

// GetRefundStatsResponse 获取退款统计响应
type GetRefundStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

// 转换函数

// ToProjectTeamResponse 将项目团队数据库模型转换为响应模型
func ToProjectTeamResponse(team *model.ProjectTeamModel) ProjectTeamResponse {
	return ProjectTeamResponse{
		ID:         uint(team.Id),
		ProjectID:  uint(team.ProjectId),
		MemberName: team.MemberName,
		MemberRole: team.MemberRole,
		Address:    team.Address,
		Email:      team.Email,
		Bio:        team.Bio,
		AvatarURL:  team.AvatarURL,
		IsActive:   team.IsActive,
	}
}

// ToProjectTeamResponseList 将项目团队数据库模型列表转换为响应模型列表
func ToProjectTeamResponseList(teams []model.ProjectTeamModel) []ProjectTeamResponse {
	result := make([]ProjectTeamResponse, len(teams))
	for i, team := range teams {
		result[i] = ToProjectTeamResponse(&team)
	}
	return result
}

// ToProjectMilestoneResponse 将项目里程碑数据库模型转换为响应模型
func ToProjectMilestoneResponse(milestone *model.ProjectMilestoneModel) ProjectMilestoneResponse {
	return ProjectMilestoneResponse{
		ID:            uint(milestone.Id),
		ProjectID:     uint(milestone.ProjectId),
		Title:         milestone.Title,
		Description:   milestone.Description,
		TargetDate:    milestone.TargetDate,
		CompletedDate: milestone.CompletedDate,
		Status:        milestone.Status,
		Progress:      milestone.Progress,
		Priority:      milestone.Priority,
	}
}

// ToProjectMilestoneResponseList 将项目里程碑数据库模型列表转换为响应模型列表
func ToProjectMilestoneResponseList(milestones []model.ProjectMilestoneModel) []ProjectMilestoneResponse {
	result := make([]ProjectMilestoneResponse, len(milestones))
	for i, milestone := range milestones {
		result[i] = ToProjectMilestoneResponse(&milestone)
	}
	return result
}

// ToProjectResponse 将数据库模型转换为响应模型
func ToProjectResponse(project *model.ProjectModel) ProjectResponse {
	return ProjectResponse{
		ID:               uint(project.Id),
		Title:            project.Title,
		Description:      project.Description,
		ImageURL:         project.ImageURL,
		Category:         project.Category,
		Creator:          project.CreatorAddress,
		TargetAmount:     project.TargetAmount,
		CurrentAmount:    project.CurrentAmount,
		MinAmount:        project.MinAmount,
		MaxAmount:        project.MaxAmount,
		Status:           string(project.Status),
		StartTime:        project.StartTime,
		EndTime:          project.EndTime,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
		ProjectTeam:      ToProjectTeamResponseList(project.ProjectTeam),
		ProjectMilestone: ToProjectMilestoneResponseList(project.ProjectMilestone),
	}
}

// ToProjectResponseList 将数据库模型列表转换为响应模型列表
func ToProjectResponseList(projects []model.ProjectModel) []ProjectResponse {
	result := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		result[i] = ToProjectResponse(&project)
	}
	return result
}

// ToContributeRecordResponse 将贡献记录数据库模型转换为响应模型
func ToContributeRecordResponse(record *model.ContributeRecordModel) ContributeRecordResponse {
	return ContributeRecordResponse{
		ID:        uint(record.Id),
		ProjectID: uint(record.ProjectId),
		Address:   record.Address,
		Amount:    record.Amount,
		TxHash:    record.TxHash,
		BlockNum:  record.BlockNum,
		CreatedAt: record.CreatedAt,
	}
}

// ToContributeRecordResponseList 将贡献记录数据库模型列表转换为响应模型列表
func ToContributeRecordResponseList(records []model.ContributeRecordModel) []ContributeRecordResponse {
	result := make([]ContributeRecordResponse, len(records))
	for i, record := range records {
		result[i] = ToContributeRecordResponse(&record)
	}
	return result
}

// ToRefundRecordResponse 将退款记录数据库模型转换为响应模型
func ToRefundRecordResponse(record *model.RefundRecordModel) RefundRecordResponse {
	return RefundRecordResponse{
		ID:        uint(record.Id),
		ProjectID: uint(record.ProjectId),
		Address:   record.Address,
		Amount:    record.Amount,
		TxHash:    record.TxHash,
		Status:    string(record.Status),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

// ToRefundRecordResponseList 将退款记录数据库模型列表转换为响应模型列表
func ToRefundRecordResponseList(records []model.RefundRecordModel) []RefundRecordResponse {
	result := make([]RefundRecordResponse, len(records))
	for i, record := range records {
		result[i] = ToRefundRecordResponse(&record)
	}
	return result
}

// ToAllProjectStatsResponse 将logic层返回的map转换为AllProjectStatsResponse
func ToAllProjectStatsResponse(stats map[string]interface{}) AllProjectStatsResponse {
	return AllProjectStatsResponse{
		TotalProjects:     stats["totalProjects"].(int64),
		PendingProjects:   stats["pendingProjects"].(int64),
		DeployingProjects: stats["deployingProjects"].(int64),
		ActiveProjects:    stats["activeProjects"].(int64),
		SuccessProjects:   stats["successProjects"].(int64),
		FailedProjects:    stats["failedProjects"].(int64),
		CancelledProjects: stats["cancelledProjects"].(int64),
		TotalRaised:       stats["totalRaised"].(string),
		TotalInvestors:    stats["totalInvestors"].(int64),
	}
}
