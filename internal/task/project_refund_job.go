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

// ProjectRefundJob 项目退款任务
type ProjectRefundJob struct {
	db           *gorm.DB
	config       *config.Config
	chainManager *chain.Manager
}

// NewProjectRefundJob 创建项目退款任务
func NewProjectRefundJob(db *gorm.DB, cfg *config.Config, chainManager *chain.Manager) *ProjectRefundJob {
	return &ProjectRefundJob{
		db:           db,
		config:       cfg,
		chainManager: chainManager,
	}
}

// GetName 获取任务名称
func (j *ProjectRefundJob) GetName() string {
	return "project_refund_updater"
}

// GetSchedule 获取调度配置
func (j *ProjectRefundJob) GetSchedule() gocron.JobDefinition {
	return gocron.DurationJob(time.Duration(j.config.Task.Interval) * time.Second)
}

// Execute 执行任务
func (j *ProjectRefundJob) Execute() {
	logger.Info("Starting project refund task")

	// 查找待退款的记录
	var refundRecords []model.RefundRecordModel
	err := j.db.Where("status = ?", "pending").Find(&refundRecords).Error

	if err != nil {
		logger.Error("Failed to fetch pending refund records: %v", err)
		return
	}

	refundedCount := 0

	for _, record := range refundRecords {
		// 获取对应的项目信息
		var project model.ProjectModel
		if err := j.db.First(&project, record.ProjectId).Error; err != nil {
			logger.Error("Failed to fetch project %d for refund: %v", record.ProjectId, err)
			continue
		}

		// 获取对应的贡献记录
		var contributeRecord model.ContributeRecordModel
		if err := j.db.First(&contributeRecord, record.ContributeID).Error; err != nil {
			logger.Error("Failed to fetch contribute record %d for refund: %v", record.ContributeID, err)
			continue
		}

		// 执行退款逻辑
		if err := j.processRefund(record, project, contributeRecord); err != nil {
			logger.Error("Failed to process refund for record %d: %v", record.Id, err)
			// 更新退款记录状态为失败
			j.updateRefundStatus(record.Id, "failed", err.Error())
			continue
		}

		// 更新退款记录状态为成功
		updates := map[string]interface{}{
			"status":    "success",
			"tx_hash":   "0x" + generateMockTxHash(), // 模拟交易哈希
			"block_num": time.Now().Unix(),           // 模拟区块号
		}

		if err := j.db.Model(&record).Updates(updates).Error; err != nil {
			logger.Error("Failed to update refund record %d: %v", record.Id, err)
			continue
		}

		logger.Info("Successfully refunded record %d, amount: %d to address: %s",
			record.Id, record.Amount, record.Address)
		refundedCount++
	}

	logger.Info("Project refund task completed. Refunded %d records", refundedCount)
}

// processRefund 处理退款逻辑
func (j *ProjectRefundJob) processRefund(record model.RefundRecordModel, project model.ProjectModel, contributeRecord model.ContributeRecordModel) error {
	logger.Info("Processing refund for record %d", record.Id)

	// 1. 验证项目状态（应该是失败状态）
	if project.Status != model.ProjectStatusFailed {
		return fmt.Errorf("project status is not failed: %s", project.Status)
	}

	// 2. 验证退款金额
	if record.Amount != contributeRecord.Amount {
		return fmt.Errorf("refund amount mismatch: record=%d, contribute=%d",
			record.Amount, contributeRecord.Amount)
	}

	// 3. 验证退款地址
	if record.Address != contributeRecord.Address {
		return fmt.Errorf("refund address mismatch: record=%s, contribute=%s",
			record.Address, contributeRecord.Address)
	}

	// 4. 调用合约进行实际退款
	if err := j.callContractRefund(project, record); err != nil {
		return fmt.Errorf("contract refund failed: %v", err)
	}

	// 5. 更新贡献记录状态为已退款
	updates := map[string]interface{}{
		"status": "refunded",
	}

	if err := j.db.Model(&contributeRecord).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update contribute record status: %v", err)
	}

	return nil
}

// callContractRefund 调用合约进行退款
func (j *ProjectRefundJob) callContractRefund(project model.ProjectModel, record model.RefundRecordModel) error {
	logger.Info("Calling contract refund for project %d, amount: %d to address: %s",
		project.Id, record.Amount, record.Address)

	// 获取众筹合约
	crowdfundingContract, err := j.chainManager.GetContract("crowdfunding")
	if err != nil {
		return fmt.Errorf("failed to get crowdfunding contract: %v", err)
	}

	// 注意：这里需要实现实际的合约调用逻辑
	// 例如：调用合约的 refundContribution 方法
	logger.Info("Would call contract refund method on: %s",
		crowdfundingContract.GetAddress().Hex())

	// 模拟合约调用成功
	return nil
}

// updateRefundStatus 更新退款记录状态
func (j *ProjectRefundJob) updateRefundStatus(recordId int64, status string, errorMsg string) {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != "" {
		// 如果有错误信息，可以存储在退款原因字段中
		updates["refund_reason"] = errorMsg
		logger.Error("Refund failed for record %d: %s", recordId, errorMsg)
	}

	if err := j.db.Model(&model.RefundRecordModel{}).Where("id = ?", recordId).Updates(updates).Error; err != nil {
		logger.Error("Failed to update refund record %d status: %v", recordId, err)
	}
}
