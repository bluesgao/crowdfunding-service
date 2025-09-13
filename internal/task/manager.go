package task

import (
	"github.com/blues/cfs/internal/chain"
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// Manager 任务管理器
type Manager struct {
	scheduler    gocron.Scheduler
	db           *gorm.DB
	chainManager *chain.Manager
	config       *config.Config
}

// NewManager 创建新的任务管理器
func NewManager(db *gorm.DB, chainManager *chain.Manager, cfg *config.Config) *Manager {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("Failed to create scheduler: %v", err)
	}

	return &Manager{
		scheduler:    s,
		db:           db,
		chainManager: chainManager,
		config:       cfg,
	}
}

// Start 启动任务管理器
func Start(db *gorm.DB, chainManager *chain.Manager, cfg *config.Config) {
	manager := NewManager(db, chainManager, cfg)

	// 注册所有任务
	manager.RegisterJobs()

	// 启动调度器
	manager.scheduler.Start()

	logger.Info("Task manager started successfully")
}

// RegisterJobs 注册所有任务
func (m *Manager) RegisterJobs() {
	// 注册项目发布任务
	m.RegisterPublishJob()
	// 注册项目完成任务
	m.RegisterFinishJob()
	// 注册项目结算任务
	m.RegisterSettlementJob()
	// 注册项目退款任务
	m.RegisterRefundJob()
}

// RegisterPublishJob 注册项目发布任务
func (m *Manager) RegisterPublishJob() {
	job := NewPublishJob(m.db, m.config, m.chainManager)

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

// RegisterFinishJob 注册项目完成任务
func (m *Manager) RegisterFinishJob() {
	job := NewFinishJob(m.db, m.config, m.chainManager)

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

// RegisterSettlementJob 注册项目结算任务
func (m *Manager) RegisterSettlementJob() {
	job := NewSettlementJob(m.db, m.config, m.chainManager)

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

// RegisterRefundJob 注册项目退款任务
func (m *Manager) RegisterRefundJob() {
	job := NewRefundJob(m.db, m.config, m.chainManager)

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
func (m *Manager) Stop() {
	if err := m.scheduler.Shutdown(); err != nil {
		logger.Error("Failed to shutdown scheduler: %v", err)
	}
	logger.Info("Task manager stopped")
}
