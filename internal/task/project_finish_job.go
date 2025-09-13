package task

import (
	"time"

	"github.com/blues/cfs/internal/chain"
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// ProjectFinishJob 项目完成任务
type ProjectFinishJob struct {
	db           *gorm.DB
	config       *config.Config
	chainManager *chain.Manager
}

// NewProjectFinishJob 创建项目完成任务
func NewProjectFinishJob(db *gorm.DB, cfg *config.Config, chainManager *chain.Manager) *ProjectFinishJob {
	return &ProjectFinishJob{
		db:           db,
		config:       cfg,
		chainManager: chainManager,
	}
}

// GetName 获取任务名称
func (j *ProjectFinishJob) GetName() string {
	return "project_finish_updater"
}

// GetSchedule 获取调度配置
func (j *ProjectFinishJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Task.Interval) * time.Second)
}

// Execute 执行任务
func (j *ProjectFinishJob) Execute() {
	logger.Info("Starting project finish task")

	now := time.Now()

	// 查找需要完成的项目：状态为进行中且结束时间小于等于当前时间
	var projects []model.ProjectModel
	err := j.db.Where("status = ? AND end_time <= ?",
		model.ProjectStatusActive, now).Find(&projects).Error

	if err != nil {
		logger.Error("Failed to fetch projects for finishing: %v", err)
		return
	}

	finishedCount := 0

	for _, project := range projects {
		// 判断项目是否成功（达到目标金额）
		var newStatus model.ProjectStatus
		if project.CurrentAmount >= project.TargetAmount {
			newStatus = model.ProjectStatusSuccess
			logger.Info("Project %d reached target amount: %d/%d",
				project.Id, project.CurrentAmount, project.TargetAmount)
		} else {
			newStatus = model.ProjectStatusFailed
			logger.Info("Project %d failed to reach target amount: %d/%d",
				project.Id, project.CurrentAmount, project.TargetAmount)
		}

		// 更新项目状态
		updates := map[string]interface{}{
			"status": newStatus,
		}

		if err := j.db.Model(&project).Updates(updates).Error; err != nil {
			logger.Error("Failed to update project %d status to %s: %v",
				project.Id, newStatus, err)
			continue
		}

		// 如果项目成功，可能需要触发合约上的结算逻辑
		if newStatus == model.ProjectStatusSuccess {
			j.handleSuccessfulProject(project)
		} else {
			// 如果项目失败，可能需要触发退款逻辑
			j.handleFailedProject(project)
		}

		logger.Info("Successfully finished project %d with status: %s",
			project.Id, newStatus)
		finishedCount++
	}

	logger.Info("Project finish task completed. Finished %d projects", finishedCount)
}

// handleSuccessfulProject 处理成功的项目
func (j *ProjectFinishJob) handleSuccessfulProject(project model.ProjectModel) {
	logger.Info("Handling successful project %d", project.Id)

	// 这里可以添加成功项目的后续处理逻辑，比如：
	// 1. 调用合约的结算方法
	// 2. 创建结算记录
	// 3. 通知项目创建者

	// 示例：创建结算记录
	settlementRecord := model.SettlementRecordModel{
		ProjectId:      project.Id,
		TotalAmount:    project.CurrentAmount,
		SettledAmount:  0,                     // 初始已结算金额为0
		PlatformFee:    0,                     // 平台手续费，可以根据配置计算
		CreatorAmount:  project.CurrentAmount, // 创建者获得金额，暂时等于总金额
		Status:         "pending",             // 待结算
		SettlementType: "success",             // 成功结算
	}

	if err := j.db.Create(&settlementRecord).Error; err != nil {
		logger.Error("Failed to create settlement record for project %d: %v",
			project.Id, err)
	} else {
		logger.Info("Created settlement record for successful project %d", project.Id)
	}
}

// handleFailedProject 处理失败的项目
func (j *ProjectFinishJob) handleFailedProject(project model.ProjectModel) {
	logger.Info("Handling failed project %d", project.Id)

	// 这里可以添加失败项目的后续处理逻辑，比如：
	// 1. 调用合约的退款方法
	// 2. 创建退款记录
	// 3. 通知所有贡献者

	// 示例：创建退款记录
	refundRecord := model.RefundRecordModel{
		ProjectId: project.Id,
		Amount:    project.CurrentAmount,
		Status:    "pending", // 待退款
		CreatedAt: time.Now(),
	}

	if err := j.db.Create(&refundRecord).Error; err != nil {
		logger.Error("Failed to create refund record for project %d: %v",
			project.Id, err)
	} else {
		logger.Info("Created refund record for failed project %d", project.Id)
	}
}
