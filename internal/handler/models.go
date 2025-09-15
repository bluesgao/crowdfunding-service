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

// ProjectResponse 项目响应模型
type ProjectResponse struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"imageUrl"`
	Category      string    `json:"category"`
	Creator       string    `json:"creator"`
	TargetAmount  int64     `json:"targetAmount"`
	CurrentAmount int64     `json:"currentAmount"`
	MinAmount     int64     `json:"minAmount"`
	MaxAmount     int64     `json:"maxAmount"`
	Status        string    `json:"status"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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
	ActiveProjects    int64  `json:"activeProjects"`
	CompletedProjects int64  `json:"completedProjects"`
	TotalRaised       string `json:"totalRaised"`
	TotalInvestors    int64  `json:"totalInvestors"`
	TotalGoal         string `json:"totalGoal"`
	SuccessRate       string `json:"successRate"`
	AverageInvestment string `json:"averageInvestment"`
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

// ToProjectResponse 将数据库模型转换为响应模型
func ToProjectResponse(project *model.ProjectModel) ProjectResponse {
	return ProjectResponse{
		ID:            uint(project.Id),
		Title:         project.Title,
		Description:   project.Description,
		ImageURL:      project.ImageURL,
		Category:      project.Category,
		Creator:       project.CreatorAddress,
		TargetAmount:  project.TargetAmount,
		CurrentAmount: project.CurrentAmount,
		MinAmount:     project.MinAmount,
		MaxAmount:     project.MaxAmount,
		Status:        string(project.Status),
		StartTime:     project.StartTime,
		EndTime:       project.EndTime,
		CreatedAt:     project.CreatedAt,
		UpdatedAt:     project.UpdatedAt,
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
		ActiveProjects:    stats["activeProjects"].(int64),
		CompletedProjects: stats["completedProjects"].(int64),
		TotalRaised:       stats["totalRaised"].(string),
		TotalInvestors:    stats["totalInvestors"].(int64),
		TotalGoal:         stats["totalGoal"].(string),
		SuccessRate:       stats["successRate"].(string),
		AverageInvestment: stats["averageInvestment"].(string),
	}
}
