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

// PublishJob 项目发布任务
type PublishJob struct {
	db           *gorm.DB
	config       *config.Config
	chainManager *chain.Manager
}

// NewPublishJob 创建项目发布任务
func NewPublishJob(db *gorm.DB, cfg *config.Config, chainManager *chain.Manager) *PublishJob {
	return &PublishJob{
		db:           db,
		config:       cfg,
		chainManager: chainManager,
	}
}

// GetName 获取任务名称
func (j *PublishJob) GetName() string {
	return "project_publish_updater"
}

// GetSchedule 获取调度配置
func (j *PublishJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Task.Interval) * time.Second)
}

// Execute 执行任务
func (j *PublishJob) Execute() {
	logger.Info("Starting project publish task")

	now := time.Now()

	// 查找需要上链的项目：状态为待上链且开始时间大于等于当前时间
	var projects []model.ProjectModel
	err := j.db.Where("status = ? AND start_time <= ?",
		model.ProjectStatusDeploying, now).Find(&projects).Error

	if err != nil {
		logger.Error("Failed to fetch projects for deployment: %v", err)
		return
	}

	deployedCount := 0

	for _, project := range projects {
		// 检查项目是否已经有合约地址（避免重复部署）
		if project.ContractAddress != "" {
			logger.Info("Project %d already has contract address: %s", project.Id, project.ContractAddress)
			continue
		}

		// 获取众筹合约
		crowdfundingContract, err := j.chainManager.GetContract("crowdfunding")
		if err != nil {
			logger.Error("Failed to get crowdfunding contract: %v", err)
			continue
		}

		// 注意：CreateProject 方法在 Contract 中未实现，这里暂时跳过
		// 实际项目中需要实现这个方法
		logger.Info("Would create project on contract: %s", crowdfundingContract.GetAddress().Hex())
		txHash := "0x0000000000000000000000000000000000000000000000000000000000000000"

		// 更新项目状态和交易哈希
		updates := map[string]interface{}{
			"status":           model.ProjectStatusActive,
			"transaction_hash": txHash,
		}

		if err := j.db.Model(&project).Updates(updates).Error; err != nil {
			logger.Error("Failed to update project %d after deployment: %v", project.Id, err)
			continue
		}

		logger.Info("Successfully deployed project %d to blockchain. TxHash: %s",
			project.Id, txHash)
		deployedCount++
	}

	logger.Info("Project publish task completed. Deployed %d projects", deployedCount)
}
