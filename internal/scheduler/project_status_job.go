package scheduler

import (
	"log"
	"time"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/model"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// ProjectStatusJob 项目状态更新任务
type ProjectStatusJob struct {
	db     *gorm.DB
	config *config.Config
}

// NewProjectStatusJob 创建项目状态更新任务
func NewProjectStatusJob(db *gorm.DB, cfg *config.Config) *ProjectStatusJob {
	return &ProjectStatusJob{
		db:     db,
		config: cfg,
	}
}

// GetName 获取任务名称
func (j *ProjectStatusJob) GetName() string {
	return "project_status_updater"
}

// GetSchedule 获取调度配置
func (j *ProjectStatusJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Scheduler.Interval) * time.Second)
}

// Execute 执行任务
func (j *ProjectStatusJob) Execute() {
	log.Println("Starting project status update task")

	now := time.Now()

	// 查找需要更新状态的项目
	var projects []model.Project
	err := j.db.Where("status IN ?", []model.ProjectStatus{
		model.ProjectStatusPending,
		model.ProjectStatusActive,
	}).Find(&projects).Error

	if err != nil {
		log.Printf("Failed to fetch projects: %v", err)
		return
	}

	updatedCount := 0

	for _, project := range projects {
		var newStatus model.ProjectStatus
		shouldUpdate := false

		switch project.Status {
		case model.ProjectStatusPending:
			// 检查是否到了开始时间
			if now.After(project.StartTime) {
				newStatus = model.ProjectStatusActive
				shouldUpdate = true
			}

		case model.ProjectStatusActive:
			// 检查是否到了结束时间或达到目标金额
			if now.After(project.EndTime) {
				if project.CurrentAmount >= project.TargetAmount {
					newStatus = model.ProjectStatusSuccess
				} else {
					newStatus = model.ProjectStatusFailed
				}
				shouldUpdate = true
			} else if project.CurrentAmount >= project.TargetAmount {
				newStatus = model.ProjectStatusSuccess
				shouldUpdate = true
			}
		}

		if shouldUpdate {
			if err := j.db.Model(&project).Update("status", newStatus).Error; err != nil {
				log.Printf("Failed to update project %d status: %v", project.ID, err)
				continue
			}

			log.Printf("Updated project %d status from %s to %s",
				project.ID, project.Status, newStatus)
			updatedCount++
		}
	}

	log.Printf("Project status update completed. Updated %d projects", updatedCount)
}
