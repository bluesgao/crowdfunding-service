package task

import (
	"log"
	"time"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/model"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// ProjectDeployJob 项目部署任务
type ProjectDeployJob struct {
	db        *gorm.DB
	config    *config.Config
	ethClient *ethereum.Client
}

// NewProjectDeployJob 创建项目部署任务
func NewProjectDeployJob(db *gorm.DB, cfg *config.Config, ethClient *ethereum.Client) *ProjectDeployJob {
	return &ProjectDeployJob{
		db:        db,
		config:    cfg,
		ethClient: ethClient,
	}
}

// GetName 获取任务名称
func (j *ProjectDeployJob) GetName() string {
	return "project_deploy_updater"
}

// GetSchedule 获取调度配置
func (j *ProjectDeployJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Task.Interval) * time.Second)
}

// Execute 执行任务
func (j *ProjectDeployJob) Execute() {
	log.Println("Starting project deploy task")

	now := time.Now()

	// 查找需要上链的项目：状态为待上链且开始时间大于等于当前时间
	var projects []model.ProjectModel
	err := j.db.Where("status = ? AND start_time <= ?",
		model.ProjectStatusDeploying, now).Find(&projects).Error

	if err != nil {
		log.Printf("Failed to fetch projects for deployment: %v", err)
		return
	}

	deployedCount := 0

	for _, project := range projects {
		// 检查项目是否已经有合约地址（避免重复部署）
		if project.ContractAddress != "" {
			log.Printf("Project %d already has contract address: %s", project.Id, project.ContractAddress)
			continue
		}

		// 调用智能合约创建项目
		creatorAddress := common.HexToAddress(project.CreatorAddress)
		txHash, err := j.ethClient.CreateProject(
			project.Title,
			project.Description,
			float64(project.TargetAmount),
			project.StartTime,
			project.EndTime,
			creatorAddress,
		)

		if err != nil {
			log.Printf("Failed to deploy project %d to blockchain: %v", project.Id, err)
			continue
		}

		// 更新项目状态和交易哈希
		updates := map[string]interface{}{
			"status":           model.ProjectStatusActive,
			"transaction_hash": txHash.Hex(),
		}

		if err := j.db.Model(&project).Updates(updates).Error; err != nil {
			log.Printf("Failed to update project %d after deployment: %v", project.Id, err)
			continue
		}

		log.Printf("Successfully deployed project %d to blockchain. TxHash: %s",
			project.Id, txHash.Hex())
		deployedCount++
	}

	log.Printf("Project deploy task completed. Deployed %d projects", deployedCount)
}
