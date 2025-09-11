package event

import (
	"log"

	"github.com/blues/cfs/internal/model"
)

// ProjectCreatedProcessor 项目创建事件处理器
type ProjectCreatedProcessor struct {
}

// NewProjectCreatedProcessor 创建项目创建事件处理器
func NewProjectCreatedProcessor() *ProjectCreatedProcessor {
	return &ProjectCreatedProcessor{}
}

// Process 处理项目创建事件
func (p *ProjectCreatedProcessor) Process(event *model.Event, eventData map[string]interface{}) error {
	// 这里可以根据需要处理项目创建事件
	// 例如：更新项目的合约地址等

	log.Printf("Processed project creation event for project %d", event.ProjectID)
	return nil
}
