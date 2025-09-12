package main

import (
	"os"
	"path/filepath"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/database"
	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/event"
	"github.com/blues/cfs/internal/handler"
	"github.com/blues/cfs/internal/logger"
	"github.com/blues/cfs/internal/logic"
	"github.com/blues/cfs/internal/router"
	"github.com/blues/cfs/internal/task"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化日志器
	initLogger(cfg.Log)

	// 初始化数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化以太坊客户端
	ethClient, err := ethereum.Init(cfg.Ethereum)
	if err != nil {
		logger.Fatalf("Failed to initialize ethereum client: %v", err)
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
	eventLogic := logic.NewEventLogic(db)
	monitor := event.NewMonitor(ethClient, eventLogic, projectLogic, contributeRecordLogic, refundRecordLogic)
	if err := monitor.Start(); err != nil {
		logger.Fatalf("Failed to start blockchain monitor: %v", err)
	}
	logger.Info("Blockchain monitor started successfully")

	// 启动定时任务
	task.Start(db, ethClient, cfg)

	// 启动服务器
	logger.Info("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

// initLogger 初始化日志器
func initLogger(logCfg config.LogConfig) {
	// 设置日志级别
	level := logger.ParseLogLevel(logCfg.Level)
	logger.SetLevel(level)

	// 设置输出目标
	switch logCfg.Output {
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		// 确保日志目录存在
		logDir := filepath.Dir(logCfg.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logger.Fatalf("Failed to create log directory: %v", err)
		}

		// 打开日志文件
		logFile, err := os.OpenFile(logCfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalf("Failed to open log file: %v", err)
		}

		logger.SetOutput(logFile)
	default: // stdout
		logger.SetOutput(os.Stdout)
	}

	logger.Info("Logger initialized with level: %s, output: %s", logCfg.Level, logCfg.Output)
}
