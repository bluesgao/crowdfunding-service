package scheduler

import (
	"log"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/ethereum"
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
		log.Fatalf("Failed to create scheduler: %v", err)
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

	log.Println("Task manager started successfully")
}

// RegisterJobs 注册所有任务
func (m *Manager) RegisterJobs() {
	// 注册项目状态更新任务
	m.RegisterProjectStatusJob()
}

// RegisterProjectStatusJob 注册项目状态更新任务
func (m *Manager) RegisterProjectStatusJob() {
	job := NewProjectStatusJob(m.db, m.config)
	m.registerProjectStatusJob(job)
}

// registerProjectStatusJob 注册项目状态任务
func (m *Manager) registerProjectStatusJob(job *ProjectStatusJob) {
	_, err := m.scheduler.NewJob(
		job.GetSchedule(),
		gocron.NewTask(job.Execute),
		gocron.WithName(job.GetName()),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		log.Printf("Failed to register job %s: %v", job.GetName(), err)
	}
}

// Stop 停止任务管理器
func (m *Manager) Stop() {
	if err := m.scheduler.Shutdown(); err != nil {
		log.Printf("Failed to shutdown scheduler: %v", err)
	}
	log.Println("Task manager stopped")
}
