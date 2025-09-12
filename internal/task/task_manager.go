package task

import (
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/contract"
	"github.com/blues/cfs/internal/logger"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// TaskManager 任务管理器
type TaskManager struct {
	scheduler       gocron.Scheduler
	db              *gorm.DB
	contractManager *contract.ContractManager
	config          *config.Config
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager(db *gorm.DB, contractManager *contract.ContractManager, cfg *config.Config) *TaskManager {
	s, err := gocron.NewScheduler()
	if err != nil {
		logger.Fatal("Failed to create scheduler: %v", err)
	}

	return &TaskManager{
		scheduler:       s,
		db:              db,
		contractManager: contractManager,
		config:          cfg,
	}
}

// Start 启动任务管理器
func Start(db *gorm.DB, contractManager *contract.ContractManager, cfg *config.Config) {
	manager := NewTaskManager(db, contractManager, cfg)

	// 注册所有任务
	manager.RegisterJobs()

	// 启动调度器
	manager.scheduler.Start()

	logger.Info("Task manager started successfully")
}

// RegisterJobs 注册所有任务
func (m *TaskManager) RegisterJobs() {
	// 注册项目部署任务
	m.RegisterProjectDeployJob()
}

// RegisterProjectDeployJob 注册项目部署任务
func (m *TaskManager) RegisterProjectDeployJob() {
	job := NewProjectDeployJob(m.db, m.config, m.contractManager)

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
