package logic

import (
	"errors"
	"fmt"

	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// ContributeRecordLogic 贡献记录业务逻辑
type ContributeRecordLogic struct {
	db *gorm.DB
}

// NewContributeRecordLogic 创建贡献记录业务逻辑
func NewContributeRecordLogic(db *gorm.DB) *ContributeRecordLogic {
	return &ContributeRecordLogic{db: db}
}

// CreateContributeRecord 创建贡献记录
func (c *ContributeRecordLogic) CreateContributeRecord(contributeRecord *model.ContributeRecord) error {
	// 验证贡献数据
	if err := c.validateContributeRecord(contributeRecord); err != nil {
		return err
	}

	// 检查项目是否存在且状态正确
	var project model.Project
	if err := c.db.First(&project, contributeRecord.ProjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("项目不存在")
		}
		return err
	}

	if project.Status != model.ProjectStatusActive {
		return errors.New("项目不在进行中，无法接受贡献")
	}

	// 检查是否超过最大贡献限制
	if project.MaxContribution > 0 && contributeRecord.Amount > project.MaxContribution {
		return errors.New("贡献金额超过最大限制")
	}

	// 检查是否低于最小贡献限制
	if project.MinContribution > 0 && contributeRecord.Amount < project.MinContribution {
		return errors.New("贡献金额低于最小限制")
	}

	// 开始事务
	tx := c.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建贡献记录
	if err := tx.Create(contributeRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新项目当前金额
	if err := tx.Model(&project).Update("current_amount", gorm.Expr("current_amount + ?", contributeRecord.Amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 检查是否达到目标金额
	if project.CurrentAmount+contributeRecord.Amount >= project.TargetAmount {
		if err := tx.Model(&project).Update("status", model.ProjectStatusSuccess).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetProjectContributeRecords 获取项目贡献记录
func (c *ContributeRecordLogic) GetProjectContributeRecords(projectID uint, page, pageSize int) ([]model.ContributeRecord, int64, error) {
	var contributions []model.ContributeRecord
	var total int64

	// 获取总数
	if err := c.db.Model(&model.ContributeRecord{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (page - 1) * pageSize
	if err := c.db.Where("project_id = ?", projectID).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&contributions).Error; err != nil {
		return nil, 0, err
	}

	return contributions, total, nil
}

// GetUserContributeRecords 获取用户贡献记录
func (c *ContributeRecordLogic) GetUserContributeRecords(address string, page, pageSize int) ([]model.ContributeRecord, int64, error) {
	var contributions []model.ContributeRecord
	var total int64

	// 获取总数
	if err := c.db.Model(&model.ContributeRecord{}).Where("address = ?", address).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取数据
	offset := (page - 1) * pageSize
	if err := c.db.Where("address = ?", address).
		Preload("Project").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&contributions).Error; err != nil {
		return nil, 0, err
	}

	return contributions, total, nil
}

// validateContributeRecord 验证贡献数据
func (c *ContributeRecordLogic) validateContributeRecord(contributeRecord *model.ContributeRecord) error {
	if contributeRecord.ProjectID == 0 {
		return errors.New("项目ID不能为空")
	}
	if contributeRecord.Amount <= 0 {
		return errors.New("贡献金额必须大于0")
	}
	if contributeRecord.Address == "" {
		return errors.New("贡献者地址不能为空")
	}
	if contributeRecord.TxHash == "" {
		return errors.New("交易哈希不能为空")
	}
	return nil
}

// GetContributeRecordByTxHash 根据交易哈希获取贡献记录
func (c *ContributeRecordLogic) GetContributeRecordByTxHash(txHash string) (*model.ContributeRecord, error) {
	var record model.ContributeRecord
	if err := c.db.Where("tx_hash = ?", txHash).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("贡献记录不存在")
		}
		return nil, fmt.Errorf("获取贡献记录失败: %w", err)
	}

	return &record, nil
}

// GetContributeStatistics 获取贡献统计信息
func (c *ContributeRecordLogic) GetContributeStatistics(projectID uint) (map[string]interface{}, error) {
	var stats struct {
		TotalContributions int64   `json:"total_contributions"`
		TotalAmount        float64 `json:"total_amount"`
		UniqueContributors int64   `json:"unique_contributors"`
		AverageAmount      float64 `json:"average_amount"`
	}

	// 总贡献记录数
	if err := c.db.Model(&model.ContributeRecord{}).Where("project_id = ?", projectID).Count(&stats.TotalContributions).Error; err != nil {
		return nil, fmt.Errorf("获取总贡献记录数失败: %w", err)
	}

	// 总贡献金额
	if err := c.db.Model(&model.ContributeRecord{}).Where("project_id = ?", projectID).Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount).Error; err != nil {
		return nil, fmt.Errorf("获取总贡献金额失败: %w", err)
	}

	// 唯一贡献者数量
	if err := c.db.Model(&model.ContributeRecord{}).Where("project_id = ?", projectID).Select("COUNT(DISTINCT address)").Scan(&stats.UniqueContributors).Error; err != nil {
		return nil, fmt.Errorf("获取唯一贡献者数量失败: %w", err)
	}

	// 平均贡献金额
	if stats.TotalContributions > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.TotalContributions)
	}

	return map[string]interface{}{
		"total_contributions": stats.TotalContributions,
		"total_amount":        stats.TotalAmount,
		"unique_contributors": stats.UniqueContributors,
		"average_amount":      stats.AverageAmount,
	}, nil
}
