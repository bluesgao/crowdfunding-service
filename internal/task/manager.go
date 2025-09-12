package task

import (
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/logger"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// Manager 任务管理器
type Manager struct {
	scheduler gocron.Scheduler
	db        *gorm.DB
	ethClient *ethereum.Client
	config    *config.Config
}

// NewManager 创建新的任务管理器
func NewManager(db *gorm.DB, ethClient *ethereum.Client, cfg *config.Config) *Manager {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("Failed to create scheduler: %v", err)
	}

	return &Manager{
		scheduler: s,
		db:        db,
		ethClient: ethClient,
		config:    cfg,
	}
}

// Start 启动任务管理器
func Start(db *gorm.DB, ethClient *ethereum.Client, cfg *config.Config) {
	manager := NewManager(db, ethClient, cfg)

	// 注册所有任务
	manager.RegisterJobs()

	// 启动调度器
	manager.scheduler.Start()

	logger.Info("Task manager started successfully")
}

// RegisterJobs 注册所有任务
func (m *Manager) RegisterJobs() {
	// 注册项目部署任务
	m.RegisterProjectDeployJob()
}

// RegisterProjectDeployJob 注册项目部署任务
func (m *Manager) RegisterProjectDeployJob() {
	job := NewProjectDeployJob(m.db, m.config, m.ethClient)

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
