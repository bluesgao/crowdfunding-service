package main

import (
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/contract"
	"github.com/blues/cfs/internal/handler"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/repository"
	"github.com/blues/cfs/internal/router"
	"github.com/blues/cfs/internal/task"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志器
	logger.Init(cfg.Log)

	// 初始化数据库
	db, err := repository.Init(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化合约管理器
	contractManager, err := contract.NewContractManager(cfg.Ethereum)
	if err != nil {
		logger.Fatalf("Failed to initialize contract manager: %v", err)
	}

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 依赖装配 - Logic层
	projectLogic := logic.NewProjectLogic(db)
	contributeRecordLogic := logic.NewContributeRecordLogic(db)
	refundRecordLogic := logic.NewRefundRecordLogic(db)
	// milestoneLogic := logic.NewMilestoneLogic(db) // 暂未使用

	// 依赖装配 - Handler层
	projectHandler := handler.NewProjectHandler(projectLogic)
	contributeHandler := handler.NewContributeHandler(contributeRecordLogic)
	refundHandler := handler.NewRefundHandler(refundRecordLogic)

	// 初始化路由
	r := router.Setup(projectHandler, contributeHandler, refundHandler, cfg)

	// 启动区块链事件监控
	monitor := contract.NewEventMonitor(contractManager, db)
	if err := monitor.Start(); err != nil {
		logger.Fatalf("Failed to start blockchain monitor: %v", err)
	}
	logger.Info("Blockchain monitor started successfully")

	// 启动定时任务
	task.Start(db, contractManager, cfg)

	// 启动服务器
	logger.Info("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
