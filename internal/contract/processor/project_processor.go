package processor

import (
	"math/big"

	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// ProjectProcessor 项目事件处理器
type ProjectProcessor struct {
	db *gorm.DB
}

// NewProjectProcessor 创建项目事件处理器
func NewProjectProcessor(db *gorm.DB) *ProjectProcessor {
	return &ProjectProcessor{
		db: db,
	}
}

// Process 处理所有事件类型
func (p *ProjectProcessor) Process(event *model.EventModel, eventData map[string]interface{}) error {
	// 根据事件类型处理不同的事件
	switch event.EventType {
	case "ProjectCreated":
		return p.processProjectCreated(event, eventData)
	case "ProjectStatusChanged":
		return p.processProjectStatusChanged(event, eventData)
	case "ContributionMade":
		return p.processContributionMade(event, eventData)
	case "RefundProcessed":
		return p.processRefundProcessed(event, eventData)
	default:
		logger.Warn("Unknown event type: %s", event.EventType)
		return nil
	}
}

// processProjectCreated 处理项目创建事件
func (p *ProjectProcessor) processProjectCreated(event *model.EventModel, eventData map[string]interface{}) error {
	// 这里可以根据需要处理项目创建事件
	// 例如：更新项目的合约地址等
	projectId := eventData["projectId"].(int64)

	logger.Info("Processed project creation event for project %d", projectId)
	return nil
}

// processProjectStatusChanged 处理项目状态变更事件
func (p *ProjectProcessor) processProjectStatusChanged(event *model.EventModel, eventData map[string]interface{}) error {
	// 获取项目状态
	status := eventData["status"].(int64)
	projectId := eventData["projectId"].(int64)

	// 映射状态
	var projectStatus model.ProjectStatus
	switch status {
	case 0:
		projectStatus = model.ProjectStatusPending
	case 1:
		projectStatus = model.ProjectStatusActive
	case 2:
		projectStatus = model.ProjectStatusSuccess
	case 3:
		projectStatus = model.ProjectStatusFailed
	case 4:
		projectStatus = model.ProjectStatusCancelled
	default:
		logger.Warn("Unknown project status: %d", status)
		return nil
	}

	// 直接通过数据库更新项目状态
	if err := p.db.Model(&model.ProjectModel{}).Where("id = ?", projectId).Update("status", projectStatus).Error; err != nil {
		logger.Error("Failed to update project status: %v", err)
		return err
	}

	logger.Info("Updated project %d status to %s", projectId, projectStatus)
	return nil
}

// processContributionMade 处理贡献事件
func (p *ProjectProcessor) processContributionMade(event *model.EventModel, eventData map[string]interface{}) error {
	// 创建贡献记录
	contributor := eventData["contributor"].(string)
	amount := eventData["amount"].(*big.Int)
	projectId := eventData["projectId"].(int64)

	contribution := model.ContributeRecordModel{
		ProjectId: projectId,
		Amount:    amount.Int64(), // 保持wei单位
		Address:   contributor,
		TxHash:    event.TxHash,
		BlockNum:  event.BlockNum,
	}

	// 直接通过数据库创建贡献记录
	if err := p.db.Create(&contribution).Error; err != nil {
		logger.Error("Failed to create contribution record: %v", err)
		return err
	}

	logger.Info("Processed contribution: %d wei from %s to project %d",
		contribution.Amount, contributor, projectId)

	return nil
}

// processRefundProcessed 处理退款事件
func (p *ProjectProcessor) processRefundProcessed(event *model.EventModel, eventData map[string]interface{}) error {
	// 创建退款记录
	refundee := eventData["refundee"].(string)
	amount := eventData["amount"].(*big.Int)
	reason := eventData["reason"].(string)
	projectId := eventData["projectId"].(int64)

	refundRecord := model.RefundRecordModel{
		ProjectId:    projectId,
		Amount:       amount.Int64(), // 保持wei单位
		Address:      refundee,
		TxHash:       event.TxHash,
		BlockNum:     event.BlockNum,
		Status:       string(model.RefundStatusSuccess),
		RefundReason: reason,
	}

	// 直接通过数据库创建退款记录
	if err := p.db.Create(&refundRecord).Error; err != nil {
		logger.Error("Failed to create refund record: %v", err)
		return err
	}

	logger.Info("Processed refund: %d wei to %s for project %d",
		refundRecord.Amount, refundee, projectId)

	return nil
}

// GetEventType 获取支持的事件类型
func (p *ProjectProcessor) GetEventType() string {
	return "ProjectCreated" // 主要处理项目创建事件，其他事件通过内部方法处理
}
