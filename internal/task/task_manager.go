package task

import (
	"github.com/blues/cfs/internal/chain"
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// TaskManager 任务管理器
type TaskManager struct {
	scheduler    gocron.Scheduler
	db           *gorm.DB
	chainManager *chain.Manager
	config       *config.Config
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(db *gorm.DB, chainManager *chain.Manager, cfg *config.Config) *TaskManager {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("Failed to create scheduler: %v", err)
	}

	return &TaskManager{
		scheduler:    s,
		db:           db,
		chainManager: chainManager,
		config:       cfg,
	}
}

// Start 启动任务管理器
func Start(db *gorm.DB, chainManager *chain.Manager, cfg *config.Config) {
	manager := NewTaskManager(db, chainManager, cfg)

	// 注册所有任务
	manager.RegisterJobs()

	// 启动调度器
	manager.scheduler.Start()

	logger.Info("Task manager started successfully")
}

// RegisterJobs 注册所有任务
func (m *TaskManager) RegisterJobs() {
	// 注册项目发布任务
	m.RegisterProjectPublishJob()
	// 注册项目完成任务
	m.RegisterProjectFinishJob()
	// 注册项目结算任务
	m.RegisterProjectSettlementJob()
	// 注册项目退款任务
	m.RegisterProjectRefundJob()
}

// RegisterProjectPublishJob 注册项目发布任务
func (m *TaskManager) RegisterProjectPublishJob() {
	job := NewProjectPublishJob(m.db, m.config, m.chainManager)

	_, err := m.scheduler.NewJob(
		job.GetSchedule(),
		gocron.NewTask(job.Execute),
		gocron.WithName(job.GetName()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logger.Error("Failed to register job %s: %v", job.GetName(), err)
	}
}

// RegisterProjectFinishJob 注册项目完成任务
func (m *TaskManager) RegisterProjectFinishJob() {
	job := NewProjectFinishJob(m.db, m.config, m.chainManager)

	_, err := m.scheduler.NewJob(
		job.GetSchedule(),
		gocron.NewTask(job.Execute),
		gocron.WithName(job.GetName()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logger.Error("Failed to register job %s: %v", job.GetName(), err)
	}
}

// RegisterProjectSettlementJob 注册项目结算任务
func (m *TaskManager) RegisterProjectSettlementJob() {
	job := NewProjectSettlementJob(m.db, m.config, m.chainManager)

	_, err := m.scheduler.NewJob(
		job.GetSchedule(),
		gocron.NewTask(job.Execute),
		gocron.WithName(job.GetName()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logger.Error("Failed to register job %s: %v", job.GetName(), err)
	}
}

// RegisterProjectRefundJob 注册项目退款任务
func (m *TaskManager) RegisterProjectRefundJob() {
	job := NewProjectRefundJob(m.db, m.config, m.chainManager)

	_, err := m.scheduler.NewJob(
		job.GetSchedule(),
		gocron.NewTask(job.Execute),
		gocron.WithName(job.GetName()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logger.Error("Failed to register job %s: %v", job.GetName(), err)
	}
}

// Stop 停止任务管理器
func (m *TaskManager) Stop() {
	if err := m.scheduler.Shutdown(); err != nil {
		logger.Error("Failed to shutdown scheduler: %v", err)
	}
	logger.Info("Task manager stopped")
}
