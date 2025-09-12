package event

import (
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/model"
)

// ProjectProcessor 项目事件处理器
type ProjectProcessor struct {
	projectLogic *logic.ProjectLogic
}

// NewProjectProcessor 创建项目事件处理器
func NewProjectProcessor(projectLogic *logic.ProjectLogic) *ProjectProcessor {
	return &ProjectProcessor{
		projectLogic: projectLogic,
	}
}

// Process 处理项目相关事件
func (p *ProjectProcessor) Process(event *model.EventModel, eventData map[string]interface{}) error {
	// 根据事件类型处理不同的事件
	switch event.EventType {
	case "ProjectCreated":
		return p.processProjectCreated(event, eventData)
	case "ProjectStatusChanged":
		return p.processProjectStatusChanged(event, eventData)
	default:
		logger.Warn("Unknown project event type: %s", event.EventType)
		return nil
	}
}

// processProjectCreated 处理项目创建事件
func (p *ProjectProcessor) processProjectCreated(event *model.EventModel, eventData map[string]interface{}) error {
	// 这里可以根据需要处理项目创建事件
	// 例如：更新项目的合约地址等

	logger.Info("Processed project creation event for project %d", event.ProjectId)
	return nil
}

// processProjectStatusChanged 处理项目状态变更事件
func (p *ProjectProcessor) processProjectStatusChanged(event *model.EventModel, eventData map[string]interface{}) error {
	// 获取项目状态
	status := eventData["status"].(int64)

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

	// 通过logic层更新项目状态
	if err := p.projectLogic.UpdateProjectStatus(event.ProjectId, projectStatus); err != nil {
		logger.Error("Failed to update project status: %v", err)
		return err
	}

	logger.Info("Updated project %d status to %s", event.ProjectId, projectStatus)
	return nil
}
