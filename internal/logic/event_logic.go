package logic

import (
	"errors"
	"fmt"
	"time"

	"github.com/blues/cfs/internal/model"
	"gorm.io/gorm"
)

// EventLogic 事件业务逻辑
type EventLogic struct {
	db *gorm.DB
}

// NewEventLogic 创建事件业务逻辑
func NewEventLogic(db *gorm.DB) *EventLogic {
	return &EventLogic{db: db}
}

// CreateEvent 创建事件记录
func (e *EventLogic) CreateEvent(event *model.EventModel) error {
	// 验证事件数据
	if err := e.validateEvent(event); err != nil {
		return err
	}

	// 检查事件是否已存在
	var existingEvent model.EventModel
	if err := e.db.Where("tx_hash = ? AND log_index = ?", event.TxHash, event.LogIndex).First(&existingEvent).Error; err == nil {
		return errors.New("事件已存在")
	}

	// 创建事件记录
	if err := e.db.Create(event).Error; err != nil {
		return fmt.Errorf("创建事件记录失败: %w", err)
	}

	return nil
}

// GetEvents 获取事件列表
func (e *EventLogic) GetEvents(projectId int64, eventName string, page, pageSize int) ([]model.EventModel, int64, error) {
	var events []model.EventModel
	var total int64

	// 构建查询条件
	query := e.db.Model(&model.EventModel{})
	if projectId > 0 {
		query = query.Where("project_id = ?", projectId)
	}
	if eventName != "" {
		query = query.Where("event_name = ?", eventName)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件列表失败: %w", err)
	}

	return events, total, nil
}

// GetEvent 获取单个事件
func (e *EventLogic) GetEvent(id int64) (*model.EventModel, error) {
	var event model.EventModel
	if err := e.db.First(&event, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("事件不存在")
		}
		return nil, fmt.Errorf("获取事件失败: %w", err)
	}

	return &event, nil
}

// GetEventByTxHash 根据交易哈希获取事件
func (e *EventLogic) GetEventByTxHash(txHash string) (*model.EventModel, error) {
	var event model.EventModel
	if err := e.db.Where("tx_hash = ?", txHash).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("事件不存在")
		}
		return nil, fmt.Errorf("获取事件失败: %w", err)
	}

	return &event, nil
}

// UpdateEventProcessed 更新事件处理状态
func (e *EventLogic) UpdateEventProcessed(id int64, processed bool) error {
	if err := e.db.Model(&model.EventModel{}).Where("id = ?", id).Update("processed", processed).Error; err != nil {
		return fmt.Errorf("更新事件处理状态失败: %w", err)
	}

	return nil
}

// GetUnprocessedEvents 获取未处理的事件
func (e *EventLogic) GetUnprocessedEvents(limit int) ([]model.EventModel, error) {
	var events []model.EventModel
	if err := e.db.Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, fmt.Errorf("获取未处理事件失败: %w", err)
	}

	return events, nil
}

// GetEventStatistics 获取事件统计信息
func (e *EventLogic) GetEventStatistics(projectId int64) (map[string]interface{}, error) {
	var stats struct {
		TotalEvents     int64 `json:"total_events"`
		ProcessedEvents int64 `json:"processed_events"`
		PendingEvents   int64 `json:"pending_events"`
	}

	// 构建查询条件
	query := e.db.Model(&model.EventModel{})
	if projectId > 0 {
		query = query.Where("project_id = ?", projectId)
	}

	// 总事件数
	if err := query.Count(&stats.TotalEvents).Error; err != nil {
		return nil, fmt.Errorf("获取总事件数失败: %w", err)
	}

	// 已处理事件数
	if err := query.Where("processed = ?", true).Count(&stats.ProcessedEvents).Error; err != nil {
		return nil, fmt.Errorf("获取已处理事件数失败: %w", err)
	}

	// 待处理事件数
	if err := query.Where("processed = ?", false).Count(&stats.PendingEvents).Error; err != nil {
		return nil, fmt.Errorf("获取待处理事件数失败: %w", err)
	}

	return map[string]interface{}{
		"total_events":     stats.TotalEvents,
		"processed_events": stats.ProcessedEvents,
		"pending_events":   stats.PendingEvents,
	}, nil
}

// GetEventsByType 根据事件类型获取事件
func (e *EventLogic) GetEventsByType(eventName string, page, pageSize int) ([]model.EventModel, int64, error) {
	var events []model.EventModel
	var total int64

	// 获取总数
	if err := e.db.Model(&model.EventModel{}).Where("event_name = ?", eventName).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := e.db.Where("event_name = ?", eventName).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件列表失败: %w", err)
	}

	return events, total, nil
}

// GetEventsByTimeRange 根据时间范围获取事件
func (e *EventLogic) GetEventsByTimeRange(startTime, endTime time.Time, page, pageSize int) ([]model.EventModel, int64, error) {
	var events []model.EventModel
	var total int64

	// 获取总数
	if err := e.db.Model(&model.EventModel{}).Where("created_at BETWEEN ? AND ?", startTime, endTime).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := e.db.Where("created_at BETWEEN ? AND ?", startTime, endTime).
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, 0, fmt.Errorf("获取事件列表失败: %w", err)
	}

	return events, total, nil
}

// validateEvent 验证事件数据
func (e *EventLogic) validateEvent(event *model.EventModel) error {
	if event.ContractAddress == "" {
		return errors.New("合约地址不能为空")
	}
	if event.ContractName == "" {
		return errors.New("合约名称不能为空")
	}
	if event.EventName == "" {
		return errors.New("事件名称不能为空")
	}
	if event.TxHash == "" {
		return errors.New("交易哈希不能为空")
	}
	if event.BlockNum == 0 {
		return errors.New("区块号不能为空")
	}

	return nil
}

// GetLastProcessedBlock 获取最后处理的区块号
func (e *EventLogic) GetLastProcessedBlock() (uint64, error) {
	var lastEvent model.EventModel
	err := e.db.Order("block_num DESC").First(&lastEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil // 没有事件记录，返回0
		}
		return 0, fmt.Errorf("获取最后处理区块号失败: %w", err)
	}
	return uint64(lastEvent.BlockNum), nil
}

// CheckEventExists 检查事件是否已存在
func (e *EventLogic) CheckEventExists(txHash string, logIndex int) (bool, error) {
	var count int64
	err := e.db.Model(&model.EventModel{}).Where("tx_hash = ? AND log_index = ?", txHash, logIndex).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("检查事件是否存在失败: %w", err)
	}
	return count > 0, nil
}
