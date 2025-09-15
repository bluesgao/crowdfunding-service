package logic

import (
	"errors"
	"fmt"
	"time"

	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// ProjectLogic 项目业务逻辑
type ProjectLogic struct {
	db *gorm.DB
}

// NewProjectLogic 创建项目业务逻辑
func NewProjectLogic(db *gorm.DB) *ProjectLogic {
	return &ProjectLogic{db: db}
}

// CreateProject 创建项目
func (p *ProjectLogic) CreateProject(project *model.ProjectModel) error {
	// 验证项目数据
	if err := p.validateProject(project); err != nil {
		return err
	}

	// 设置默认值
	project.Status = model.ProjectStatusDeploying
	project.CurrentAmount = 0

	// 创建项目
	if err := p.db.Create(project).Error; err != nil {
		return err
	}

	return nil
}

// GetProjects 获取项目列表
func (p *ProjectLogic) GetProjects() ([]model.ProjectModel, error) {
	var projects []model.ProjectModel

	// 获取所有项目
	if err := p.db.Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("获取项目列表失败: %w", err)
	}

	return projects, nil
}

// GetProject 获取项目详情
func (p *ProjectLogic) GetProject(id int64) (*model.ProjectModel, error) {
	var project model.ProjectModel
	if err := p.db.Preload("Contributions").
		Preload("Events").
		First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("项目不存在")
		}
		return nil, fmt.Errorf("获取项目详情失败: %w", err)
	}

	return &project, nil
}

// GetProjectStats 获取项目统计信息
func (p *ProjectLogic) GetProjectStats(id int64) (map[string]interface{}, error) {
	var project model.ProjectModel
	if err := p.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 统计贡献者数量
	var contributorCount int64
	p.db.Model(&model.ContributeRecordModel{}).
		Where("project_id = ?", id).
		Distinct("address").
		Count(&contributorCount)

	// 统计贡献记录数量
	var contributionCount int64
	p.db.Model(&model.ContributeRecordModel{}).
		Where("project_id = ?", id).
		Count(&contributionCount)

	// 计算完成百分比
	completionPercentage := float64(0)
	if project.TargetAmount > 0 {
		completionPercentage = float64(project.CurrentAmount) / float64(project.TargetAmount) * 100
	}

	// 计算剩余时间
	remainingTime := time.Duration(0)
	if project.Status == model.ProjectStatusActive && time.Now().Before(project.EndTime) {
		remainingTime = time.Until(project.EndTime)
	}

	return map[string]interface{}{
		"project_id":            project.Id,
		"current_amount":        project.CurrentAmount,
		"target_amount":         project.TargetAmount,
		"completion_percentage": completionPercentage,
		"contributor_count":     contributorCount,
		"contribution_count":    contributionCount,
		"remaining_time":        remainingTime.String(),
		"status":                project.Status,
	}, nil
}

// GetAllProjectStats 获取所有项目的统计信息
func (p *ProjectLogic) GetAllProjectStats() (map[string]interface{}, error) {
	// 统计项目总数
	var totalProjects int64
	p.db.Model(&model.ProjectModel{}).Count(&totalProjects)

	// 按状态统计项目数量
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	p.db.Model(&model.ProjectModel{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusStats)

	// 统计总目标金额和当前金额
	var totalStats struct {
		TotalTargetAmount  int64 `json:"total_target_amount"`
		TotalCurrentAmount int64 `json:"total_current_amount"`
	}
	p.db.Model(&model.ProjectModel{}).
		Select("SUM(target_amount) as total_target_amount, SUM(current_amount) as total_current_amount").
		Scan(&totalStats)

	// 统计总贡献者数量（去重）
	var totalContributors int64
	p.db.Model(&model.ContributeRecordModel{}).
		Distinct("address").
		Count(&totalContributors)

	// 统计总贡献记录数量
	var totalContributions int64
	p.db.Model(&model.ContributeRecordModel{}).Count(&totalContributions)

	// 统计活跃项目数量（进行中的项目）
	var activeProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusActive).
		Count(&activeProjects)

	// 统计成功项目数量
	var successProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusSuccess).
		Count(&successProjects)

	// 统计失败项目数量
	var failedProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusFailed).
		Count(&failedProjects)

	// 统计最近7天创建的项目数量
	var recentProjects int64
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	p.db.Model(&model.ProjectModel{}).
		Where("created_at >= ?", sevenDaysAgo).
		Count(&recentProjects)

	// 计算成功项目数量（包括成功和失败的项目）
	completedProjects := successProjects + failedProjects

	// 计算成功率
	successRate := float64(0)
	if completedProjects > 0 {
		successRate = float64(successProjects) / float64(completedProjects) * 100
	}

	// 计算平均投资金额
	avgInvestment := float64(0)
	if totalContributions > 0 {
		avgInvestment = float64(totalStats.TotalCurrentAmount) / float64(totalContributions)
	}

	return map[string]interface{}{
		"totalProjects":     totalProjects,
		"activeProjects":    activeProjects,
		"completedProjects": completedProjects,
		"totalRaised":       fmt.Sprintf("%d", totalStats.TotalCurrentAmount),
		"totalInvestors":    totalContributors,
		"totalGoal":         fmt.Sprintf("%d", totalStats.TotalTargetAmount),
		"successRate":       fmt.Sprintf("%.2f", successRate),
		"averageInvestment": fmt.Sprintf("%.0f", avgInvestment),
	}, nil
}

// validateProject 验证项目数据
func (p *ProjectLogic) validateProject(project *model.ProjectModel) error {
	if project.Title == "" {
		return errors.New("项目标题不能为空")
	}
	if project.TargetAmount <= 0 {
		return errors.New("目标金额必须大于0")
	}
	if project.StartTime.After(project.EndTime) {
		return errors.New("开始时间不能晚于结束时间")
	}
	if project.StartTime.Before(time.Now()) {
		return errors.New("开始时间不能早于当前时间")
	}
	return nil
}
