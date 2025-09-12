package main

import (
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
	level := logger.ParseLogLevel(logCfg.Level)

	// 根据配置选择日志输出方式
	if logCfg.Output == "file" {
		// 使用文件轮转日志
		config := logger.LumberjackConfig{
			Filename:   logCfg.File,
			MaxSize:    100,  // 100MB
			MaxBackups: 5,    // 保留5个备份
			MaxAge:     30,   // 保留30天
			Compress:   true, // 压缩旧文件
		}

		fileLogger, err := logger.NewWithLumberjackConfig(level, config)
		if err != nil {
			panic("Failed to initialize file logger: " + err.Error())
		}

		// 替换默认日志器
		logger.SetDefaultLogger(fileLogger)
		logger.Info("Logger initialized with level: %s, output: %s, file: %s", logCfg.Level, logCfg.Output, logCfg.File)
	} else {
		// 使用标准输出
		logger.SetLevel(level)
		logger.Info("Logger initialized with level: %s, output: %s", logCfg.Level, logCfg.Output)
	}
}
