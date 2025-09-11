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
func (p *ProjectLogic) CreateProject(project *model.Project) error {
	// 验证项目数据
	if err := p.validateProject(project); err != nil {
		return err
	}

	// 设置默认值
	project.Status = model.ProjectStatusPending
	project.CurrentAmount = 0

	// 创建项目
	if err := p.db.Create(project).Error; err != nil {
		return err
	}

	return nil
}

// UpdateProject 更新项目
func (p *ProjectLogic) UpdateProject(id uint, updates map[string]interface{}) error {
	// 检查项目是否存在
	var project model.Project
	if err := p.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("项目不存在")
		}
		return err
	}

	// 只允许更新特定字段
	allowedFields := []string{"title", "description", "image_url", "category"}
	for key := range updates {
		if !contains(allowedFields, key) {
			delete(updates, key)
		}
	}

	if len(updates) == 0 {
		return errors.New("没有要更新的字段")
	}

	// 更新项目
	if err := p.db.Model(&project).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// CancelProject 取消项目
func (p *ProjectLogic) CancelProject(id uint) error {
	var project model.Project
	if err := p.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("项目不存在")
		}
		return err
	}

	// 检查项目状态
	if project.Status != model.ProjectStatusPending && project.Status != model.ProjectStatusActive {
		return errors.New("只有待开始或进行中的项目才能取消")
	}

	// 更新状态
	if err := p.db.Model(&project).Update("status", model.ProjectStatusCancelled).Error; err != nil {
		return err
	}

	return nil
}

// GetProjects 获取项目列表
func (p *ProjectLogic) GetProjects(status string, category string, creator string, page, pageSize int) ([]model.Project, int64, error) {
	var projects []model.Project
	var total int64

	query := p.db.Model(&model.Project{})

	// 添加过滤条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if creator != "" {
		query = query.Where("creator = ?", creator)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取项目总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("获取项目列表失败: %w", err)
	}

	return projects, total, nil
}

// GetProject 获取项目详情
func (p *ProjectLogic) GetProject(id uint) (*model.Project, error) {
	var project model.Project
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

// UpdateProjectStatus 更新项目状态
func (p *ProjectLogic) UpdateProjectStatus(id uint, status model.ProjectStatus) error {
	var project model.Project
	if err := p.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("项目不存在")
		}
		return fmt.Errorf("获取项目失败: %w", err)
	}

	// 更新项目状态
	if err := p.db.Model(&project).Update("status", status).Error; err != nil {
		return fmt.Errorf("更新项目状态失败: %w", err)
	}

	return nil
}

// GetProjectContributions 获取项目贡献记录
func (p *ProjectLogic) GetProjectContributions(projectID uint, page, pageSize int) ([]model.ContributeRecord, int64, error) {
	var contributions []model.ContributeRecord
	var total int64

	// 获取总数
	if err := p.db.Model(&model.ContributeRecord{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取贡献记录总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := p.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&contributions).Error; err != nil {
		return nil, 0, fmt.Errorf("获取贡献记录失败: %w", err)
	}

	return contributions, total, nil
}

// GetProjectStats 获取项目统计信息
func (p *ProjectLogic) GetProjectStats(id uint) (map[string]interface{}, error) {
	var project model.Project
	if err := p.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("项目不存在")
		}
		return nil, err
	}

	// 统计贡献者数量
	var contributorCount int64
	p.db.Model(&model.ContributeRecord{}).
		Where("project_id = ?", id).
		Distinct("address").
		Count(&contributorCount)

	// 统计贡献记录数量
	var contributionCount int64
	p.db.Model(&model.ContributeRecord{}).
		Where("project_id = ?", id).
		Count(&contributionCount)

	// 计算完成百分比
	completionPercentage := float64(0)
	if project.TargetAmount > 0 {
		completionPercentage = (project.CurrentAmount / project.TargetAmount) * 100
	}

	// 计算剩余时间
	remainingTime := time.Duration(0)
	if project.Status == model.ProjectStatusActive && time.Now().Before(project.EndTime) {
		remainingTime = project.EndTime.Sub(time.Now())
	}

	return map[string]interface{}{
		"project_id":            project.ID,
		"current_amount":        project.CurrentAmount,
		"target_amount":         project.TargetAmount,
		"completion_percentage": completionPercentage,
		"contributor_count":     contributorCount,
		"contribution_count":    contributionCount,
		"remaining_time":        remainingTime.String(),
		"status":                project.Status,
	}, nil
}

// validateProject 验证项目数据
func (p *ProjectLogic) validateProject(project *model.Project) error {
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
