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
	if err := p.db.Preload("ProjectTeam").
		Preload("ProjectMilestone").
		First(&project, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, fmt.Errorf("获取项目详情失败: %w", err)
	}

	return &project, nil
}

// GetProjectStats 获取项目统计信息
func (p *ProjectLogic) GetProjectStats(id int64) (map[string]interface{}, error) {
	// 使用一个 SQL 查询获取所有统计信息
	var stats struct {
		ProjectId         int64     `json:"project_id"`
		CurrentAmount     int64     `json:"current_amount"`
		TargetAmount      int64     `json:"target_amount"`
		Status            string    `json:"status"`
		StartTime         time.Time `json:"start_time"`
		EndTime           time.Time `json:"end_time"`
		ContributorCount  int64     `json:"contributor_count"`
		ContributionCount int64     `json:"contribution_count"`
	}

	// 使用子查询和 JOIN 来获取所有统计信息
	err := p.db.Raw(`
		SELECT 
			p.id as project_id,
			p.current_amount,
			p.target_amount,
			p.status,
			p.start_time,
			p.end_time,
			COALESCE(contributor_stats.contributor_count, 0) as contributor_count,
			COALESCE(contribution_stats.contribution_count, 0) as contribution_count
		FROM project p
		LEFT JOIN (
			SELECT 
				project_id,
				COUNT(DISTINCT address) as contributor_count
			FROM contribute_record 
			WHERE project_id = ?
			GROUP BY project_id
		) contributor_stats ON p.id = contributor_stats.project_id
		LEFT JOIN (
			SELECT 
				project_id,
				COUNT(*) as contribution_count
			FROM contribute_record 
			WHERE project_id = ?
			GROUP BY project_id
		) contribution_stats ON p.id = contribution_stats.project_id
		WHERE p.id = ?
	`, id, id, id).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("获取项目统计信息失败: %w", err)
	}

	// 检查项目是否存在
	if stats.ProjectId == 0 {
		return nil, errors.New("项目不存在")
	}

	// 计算完成百分比
	completionPercentage := float64(0)
	if stats.TargetAmount > 0 {
		completionPercentage = float64(stats.CurrentAmount) / float64(stats.TargetAmount) * 100
	}

	// 计算剩余时间
	remainingTime := time.Duration(0)
	if stats.Status == string(model.ProjectStatusActive) && time.Now().Before(stats.EndTime) {
		remainingTime = time.Until(stats.EndTime)
	}

	return map[string]interface{}{
		"project_id":            stats.ProjectId,
		"current_amount":        stats.CurrentAmount,
		"target_amount":         stats.TargetAmount,
		"completion_percentage": completionPercentage,
		"contributor_count":     stats.ContributorCount,
		"contribution_count":    stats.ContributionCount,
		"remaining_time":        remainingTime.String(),
		"status":                stats.Status,
	}, nil
}

// GetAllProjectStats 获取所有项目的统计信息
func (p *ProjectLogic) GetAllProjectStats() (map[string]interface{}, error) {
	// 统计项目总数
	var totalProjects int64
	p.db.Model(&model.ProjectModel{}).Count(&totalProjects)

	// 统计各状态项目数量
	var pendingProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusPending).
		Count(&pendingProjects)

	var deployingProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusDeploying).
		Count(&deployingProjects)

	var activeProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusActive).
		Count(&activeProjects)

	var successProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusSuccess).
		Count(&successProjects)

	var failedProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusFailed).
		Count(&failedProjects)

	var cancelledProjects int64
	p.db.Model(&model.ProjectModel{}).
		Where("status = ?", model.ProjectStatusCancelled).
		Count(&cancelledProjects)

	// 统计总当前金额
	var totalCurrentAmount int64
	p.db.Model(&model.ProjectModel{}).
		Select("SUM(current_amount)").
		Scan(&totalCurrentAmount)

	// 统计总贡献者数量（去重）
	var totalContributors int64
	p.db.Model(&model.ContributeRecordModel{}).
		Distinct("address").
		Count(&totalContributors)

	return map[string]interface{}{
		"totalProjects":     totalProjects,
		"pendingProjects":   pendingProjects,
		"deployingProjects": deployingProjects,
		"activeProjects":    activeProjects,
		"successProjects":   successProjects,
		"failedProjects":    failedProjects,
		"cancelledProjects": cancelledProjects,
		"totalRaised":       fmt.Sprintf("%d", totalCurrentAmount),
		"totalInvestors":    totalContributors,
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
