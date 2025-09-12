package logic

import (
	"errors"
	"time"

	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// MilestoneLogic 里程碑业务逻辑
type MilestoneLogic struct {
	db *gorm.DB
}

// NewMilestoneLogic 创建里程碑业务逻辑
func NewMilestoneLogic(db *gorm.DB) *MilestoneLogic {
	return &MilestoneLogic{db: db}
}

// CreateMilestone 创建里程碑
func (m *MilestoneLogic) CreateMilestone(milestone *model.ProjectMilestoneModel) error {
	// 验证里程碑数据
	if err := m.validateMilestone(milestone); err != nil {
		return err
	}

	// 检查项目是否存在
	var project model.ProjectModel
	if err := m.db.First(&project, milestone.ProjectId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("项目不存在")
		}
		return err
	}

	// 创建里程碑
	if err := m.db.Create(milestone).Error; err != nil {
		return err
	}

	return nil
}

// UpdateMilestone 更新里程碑
func (m *MilestoneLogic) UpdateMilestone(id int64, updates map[string]interface{}) error {
	// 检查里程碑是否存在
	var milestone model.ProjectMilestoneModel
	if err := m.db.First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("里程碑不存在")
		}
		return err
	}

	// 只允许更新特定字段
	allowedFields := []string{"title", "description", "target_date", "status", "progress", "priority", "is_public"}
	for key := range updates {
		if !contains(allowedFields, key) {
			delete(updates, key)
		}
	}

	if len(updates) == 0 {
		return errors.New("没有要更新的字段")
	}

	// 如果状态更新为已完成，设置完成时间
	if status, ok := updates["status"]; ok && status == model.MilestoneStatusCompleted {
		now := time.Now()
		updates["completed_date"] = &now
	}

	// 更新里程碑
	if err := m.db.Model(&milestone).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetProjectMilestones 获取项目里程碑
func (m *MilestoneLogic) GetProjectMilestones(projectId int64, isPublic bool) ([]model.ProjectMilestoneModel, error) {
	var milestones []model.ProjectMilestoneModel

	query := m.db.Where("project_id = ?", projectId)
	if isPublic {
		query = query.Where("is_public = ?", true)
	}

	if err := query.Order("target_date ASC").Find(&milestones).Error; err != nil {
		return nil, err
	}

	return milestones, nil
}

// UpdateMilestoneProgress 更新里程碑进度
func (m *MilestoneLogic) UpdateMilestoneProgress(id int64, progress int) error {
	if progress < 0 || progress > 100 {
		return errors.New("进度必须在0-100之间")
	}

	var milestone model.ProjectMilestoneModel
	if err := m.db.First(&milestone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("里程碑不存在")
		}
		return err
	}

	updates := map[string]interface{}{
		"progress": progress,
	}

	// 根据进度更新状态
	if progress == 0 {
		updates["status"] = model.MilestoneStatusPending
	} else if progress == 100 {
		updates["status"] = model.MilestoneStatusCompleted
		now := time.Now()
		updates["completed_date"] = &now
	} else {
		updates["status"] = model.MilestoneStatusInProgress
	}

	if err := m.db.Model(&milestone).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// validateMilestone 验证里程碑数据
func (m *MilestoneLogic) validateMilestone(milestone *model.ProjectMilestoneModel) error {
	if milestone.ProjectId == 0 {
		return errors.New("项目ID不能为空")
	}
	if milestone.Title == "" {
		return errors.New("里程碑标题不能为空")
	}
	if milestone.TargetDate.IsZero() {
		return errors.New("目标日期不能为空")
	}
	if milestone.Progress < 0 || milestone.Progress > 100 {
		return errors.New("进度必须在0-100之间")
	}
	return nil
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
