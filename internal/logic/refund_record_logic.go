package logic

import (
	"errors"
	"fmt"

	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// RefundRecordLogic 退款记录业务逻辑
type RefundRecordLogic struct {
	db *gorm.DB
}

// NewRefundRecordLogic 创建退款记录业务逻辑
func NewRefundRecordLogic(db *gorm.DB) *RefundRecordLogic {
	return &RefundRecordLogic{db: db}
}

// CreateRefundRecord 创建退款记录
func (r *RefundRecordLogic) CreateRefundRecord(refundRecord *model.RefundRecord) error {
	// 验证退款数据
	if err := r.validateRefundRecord(refundRecord); err != nil {
		return err
	}

	// 检查项目是否存在
	var project model.Project
	if err := r.db.First(&project, refundRecord.ProjectID).Error; err != nil {
		return errors.New("项目不存在")
	}

	// 检查项目状态是否允许退款
	if project.Status != model.ProjectStatusFailed && project.Status != model.ProjectStatusCancelled {
		return errors.New("项目状态不允许退款")
	}

	// 检查贡献记录是否存在
	var contributeRecord model.ContributeRecord
	if err := r.db.Where("project_id = ? AND address = ?", refundRecord.ProjectID, refundRecord.Address).First(&contributeRecord).Error; err != nil {
		return errors.New("未找到对应的贡献记录")
	}

	// 检查是否已经退款
	var existingRefund model.RefundRecord
	if err := r.db.Where("project_id = ? AND address = ?", refundRecord.ProjectID, refundRecord.Address).First(&existingRefund).Error; err == nil {
		return errors.New("该地址已经退款")
	}

	// 检查交易哈希是否已存在
	if err := r.db.Where("tx_hash = ?", refundRecord.TxHash).First(&existingRefund).Error; err == nil {
		return errors.New("交易哈希已存在")
	}

	// 设置退款金额为贡献金额
	refundRecord.Amount = contributeRecord.Amount

	// 创建退款记录
	if err := r.db.Create(refundRecord).Error; err != nil {
		return fmt.Errorf("创建退款记录失败: %w", err)
	}

	return nil
}

// GetProjectRefunds 获取项目退款记录
func (r *RefundRecordLogic) GetProjectRefunds(projectID uint, page, pageSize int) ([]model.RefundRecord, int64, error) {
	var refunds []model.RefundRecord
	var total int64

	// 获取总数
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ?", projectID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取退款记录总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&refunds).Error; err != nil {
		return nil, 0, fmt.Errorf("获取退款记录失败: %w", err)
	}

	return refunds, total, nil
}

// GetUserRefunds 获取用户退款记录
func (r *RefundRecordLogic) GetUserRefunds(address string, page, pageSize int) ([]model.RefundRecord, int64, error) {
	var refunds []model.RefundRecord
	var total int64

	// 获取总数
	if err := r.db.Model(&model.RefundRecord{}).Where("address = ?", address).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户退款记录总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.Where("address = ?", address).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&refunds).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户退款记录失败: %w", err)
	}

	return refunds, total, nil
}

// GetRefundByTxHash 根据交易哈希获取退款记录
func (r *RefundRecordLogic) GetRefundByTxHash(txHash string) (*model.RefundRecord, error) {
	var refund model.RefundRecord
	if err := r.db.Where("tx_hash = ?", txHash).First(&refund).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("退款记录不存在")
		}
		return nil, fmt.Errorf("获取退款记录失败: %w", err)
	}

	return &refund, nil
}

// UpdateRefundStatus 更新退款状态
func (r *RefundRecordLogic) UpdateRefundStatus(id uint, status model.RefundStatus) error {
	if err := r.db.Model(&model.RefundRecord{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("更新退款状态失败: %w", err)
	}

	return nil
}

// GetRefundStatistics 获取退款统计信息
func (r *RefundRecordLogic) GetRefundStatistics(projectID uint) (map[string]interface{}, error) {
	var stats struct {
		TotalRefunds     int64   `json:"total_refunds"`
		TotalAmount      float64 `json:"total_amount"`
		PendingRefunds   int64   `json:"pending_refunds"`
		CompletedRefunds int64   `json:"completed_refunds"`
		FailedRefunds    int64   `json:"failed_refunds"`
	}

	// 总退款记录数
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ?", projectID).Count(&stats.TotalRefunds).Error; err != nil {
		return nil, fmt.Errorf("获取总退款记录数失败: %w", err)
	}

	// 总退款金额
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ?", projectID).Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount).Error; err != nil {
		return nil, fmt.Errorf("获取总退款金额失败: %w", err)
	}

	// 待处理退款数
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ? AND status = ?", projectID, model.RefundStatusPending).Count(&stats.PendingRefunds).Error; err != nil {
		return nil, fmt.Errorf("获取待处理退款数失败: %w", err)
	}

	// 已完成退款数
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ? AND status = ?", projectID, model.RefundStatusSuccess).Count(&stats.CompletedRefunds).Error; err != nil {
		return nil, fmt.Errorf("获取已完成退款数失败: %w", err)
	}

	// 失败退款数
	if err := r.db.Model(&model.RefundRecord{}).Where("project_id = ? AND status = ?", projectID, model.RefundStatusFailed).Count(&stats.FailedRefunds).Error; err != nil {
		return nil, fmt.Errorf("获取失败退款数失败: %w", err)
	}

	return map[string]interface{}{
		"total_refunds":     stats.TotalRefunds,
		"total_amount":      stats.TotalAmount,
		"pending_refunds":   stats.PendingRefunds,
		"completed_refunds": stats.CompletedRefunds,
		"failed_refunds":    stats.FailedRefunds,
	}, nil
}

// validateRefundRecord 验证退款数据
func (r *RefundRecordLogic) validateRefundRecord(refundRecord *model.RefundRecord) error {
	if refundRecord.ProjectID == 0 {
		return errors.New("项目ID不能为空")
	}
	if refundRecord.Address == "" {
		return errors.New("退款地址不能为空")
	}
	if refundRecord.TxHash == "" {
		return errors.New("交易哈希不能为空")
	}
	if refundRecord.Amount <= 0 {
		return errors.New("退款金额必须大于0")
	}

	return nil
}
