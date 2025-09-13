package task

import (
	"fmt"
	"time"

	"github.com/blues/cfs/internal/chain"
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/model"
	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

// ProjectSettlementJob 项目结算任务
type ProjectSettlementJob struct {
	db           *gorm.DB
	config       *config.Config
	chainManager *chain.Manager
}

// NewProjectSettlementJob 创建项目结算任务
func NewProjectSettlementJob(db *gorm.DB, cfg *config.Config, chainManager *chain.Manager) *ProjectSettlementJob {
	return &ProjectSettlementJob{
		db:           db,
		config:       cfg,
		chainManager: chainManager,
	}
}

// GetName 获取任务名称
func (j *ProjectSettlementJob) GetName() string {
	return "project_settlement_updater"
}

// GetSchedule 获取调度配置
func (j *ProjectSettlementJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Task.Interval) * time.Second)
}

// Execute 执行任务
func (j *ProjectSettlementJob) Execute() {
	logger.Info("Starting project settlement task")

	// 查找待结算的记录
	var settlementRecords []model.SettlementRecordModel
	err := j.db.Where("status = ?", "pending").Find(&settlementRecords).Error

	if err != nil {
		logger.Error("Failed to fetch pending settlement records: %v", err)
		return
	}

	settledCount := 0

	for _, record := range settlementRecords {
		// 获取对应的项目信息
		var project model.ProjectModel
		if err := j.db.First(&project, record.ProjectId).Error; err != nil {
			logger.Error("Failed to fetch project %d for settlement: %v", record.ProjectId, err)
			continue
		}

		// 执行结算逻辑
		if err := j.processSettlement(record, project); err != nil {
			logger.Error("Failed to process settlement for project %d: %v", record.ProjectId, err)
			// 更新结算记录状态为失败
			j.updateSettlementStatus(record.Id, "failed", err.Error())
			continue
		}

		// 更新结算记录状态为成功
		now := time.Now()
		updates := map[string]interface{}{
			"status":          "success",
			"settlement_time": &now,
			"tx_hash":         "0x" + generateMockTxHash(), // 模拟交易哈希
			"block_num":       time.Now().Unix(),           // 模拟区块号
		}

		if err := j.db.Model(&record).Updates(updates).Error; err != nil {
			logger.Error("Failed to update settlement record %d: %v", record.Id, err)
			continue
		}

		logger.Info("Successfully settled project %d, amount: %d",
			record.ProjectId, record.TotalAmount)
		settledCount++
	}

	logger.Info("Project settlement task completed. Settled %d projects", settledCount)
}

// processSettlement 处理结算逻辑
func (j *ProjectSettlementJob) processSettlement(record model.SettlementRecordModel, project model.ProjectModel) error {
	logger.Info("Processing settlement for project %d", project.Id)

	// 1. 验证项目状态
	if project.Status != model.ProjectStatusSuccess {
		return fmt.Errorf("project status is not success: %s", project.Status)
	}

	// 2. 验证结算金额
	if record.TotalAmount != project.CurrentAmount {
		return fmt.Errorf("settlement amount mismatch: record=%d, project=%d",
			record.TotalAmount, project.CurrentAmount)
	}

	// 3. 计算平台手续费（示例：5%）
	platformFeeRate := 0.05
	platformFee := int64(float64(record.TotalAmount) * platformFeeRate)
	creatorAmount := record.TotalAmount - platformFee

	// 4. 更新结算记录的计算字段
	updates := map[string]interface{}{
		"platform_fee":   platformFee,
		"creator_amount": creatorAmount,
		"settled_amount": record.TotalAmount,
	}

	if err := j.db.Model(&record).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update settlement calculation: %v", err)
	}

	// 5. 调用合约进行实际结算（这里暂时模拟）
	if err := j.callContractSettlement(project, creatorAmount); err != nil {
		return fmt.Errorf("contract settlement failed: %v", err)
	}

	// 6. 记录平台手续费收入（可以创建平台收入记录）
	if err := j.recordPlatformFee(project.Id, platformFee); err != nil {
		logger.Warn("Failed to record platform fee for project %d: %v", project.Id, err)
		// 不返回错误，因为主要结算已完成
	}

	return nil
}

// callContractSettlement 调用合约进行结算
func (j *ProjectSettlementJob) callContractSettlement(project model.ProjectModel, amount int64) error {
	logger.Info("Calling contract settlement for project %d, amount: %d", project.Id, amount)

	// 获取众筹合约
	crowdfundingContract, err := j.chainManager.GetContract("crowdfunding")
	if err != nil {
		return fmt.Errorf("failed to get crowdfunding contract: %v", err)
	}

	// 注意：这里需要实现实际的合约调用逻辑
	// 例如：调用合约的 settleProject 方法
	logger.Info("Would call contract settlement method on: %s",
		crowdfundingContract.GetAddress().Hex())

	// 模拟合约调用成功
	return nil
}

// recordPlatformFee 记录平台手续费收入
func (j *ProjectSettlementJob) recordPlatformFee(projectId int64, fee int64) error {
	// 这里可以创建平台收入记录表，或者更新平台收入统计
	logger.Info("Recording platform fee for project %d: %d", projectId, fee)

	// 示例：可以创建一个平台收入记录
	// platformIncome := model.PlatformIncomeModel{
	//     ProjectId: projectId,
	//     Amount:    fee,
	//     Type:      "settlement_fee",
	//     CreatedAt: time.Now(),
	// }
	// return j.db.Create(&platformIncome).Error

	return nil
}

// updateSettlementStatus 更新结算记录状态
func (j *ProjectSettlementJob) updateSettlementStatus(recordId int64, status string, errorMsg string) {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		// 如果有错误信息，可以存储在某个字段中
		logger.Error("Settlement failed for record %d: %s", recordId, errorMsg)
	}

	if err := j.db.Model(&model.SettlementRecordModel{}).Where("id = ?", recordId).Updates(updates).Error; err != nil {
		logger.Error("Failed to update settlement record %d status: %v", recordId, err)
	}
}

// generateMockTxHash 生成模拟交易哈希
func generateMockTxHash() string {
	// 这里应该生成真实的交易哈希，暂时使用模拟值
	return "1234567890abcdef1234567890abcdef12345678"
}
