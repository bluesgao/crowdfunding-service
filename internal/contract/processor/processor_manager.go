package processor

import (
	"sync"

	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// ProcessorManager 事件处理器管理器
type ProcessorManager struct {
	mu         sync.RWMutex
	processors map[string]EventProcessor
}

// EventProcessor 事件处理器接口
type EventProcessor interface {
	Process(event *model.EventModel, eventData map[string]interface{}) error
	GetEventType() string
}

// NewProcessorManager 创建处理器管理器
func NewProcessorManager(db *gorm.DB) *ProcessorManager {
	manager := &ProcessorManager{
		processors: make(map[string]EventProcessor),
	}

	// 注册所有处理器
	manager.RegisterProcessor(NewProjectProcessor(db))

	logger.Info("ProcessorManager initialized with %d processors", len(manager.processors))
	return manager
}

// RegisterProcessor 注册事件处理器
func (pm *ProcessorManager) RegisterProcessor(processor EventProcessor) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	eventType := processor.GetEventType()
	pm.processors[eventType] = processor
	logger.Info("Registered processor for event type: %s", eventType)
}

// GetProcessor 获取指定事件类型的处理器
func (pm *ProcessorManager) GetProcessor(eventType string) (EventProcessor, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	processor, exists := pm.processors[eventType]
	return processor, exists
}

// ProcessEvent 处理事件
func (pm *ProcessorManager) ProcessEvent(event *model.EventModel, eventData map[string]interface{}) error {
	processor, exists := pm.GetProcessor(event.EventType)
	if !exists {
		logger.Warn("No processor found for event type: %s", event.EventType)
		return nil // 跳过未知事件类型
	}

	return processor.Process(event, eventData)
}

// GetAllProcessors 获取所有处理器
func (pm *ProcessorManager) GetAllProcessors() map[string]EventProcessor {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]EventProcessor)
	for eventType, processor := range pm.processors {
		result[eventType] = processor
	}
	return result
}

// GetSupportedEventTypes 获取支持的事件类型列表
func (pm *ProcessorManager) GetSupportedEventTypes() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	eventTypes := make([]string, 0, len(pm.processors))
	for eventType := range pm.processors {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}
