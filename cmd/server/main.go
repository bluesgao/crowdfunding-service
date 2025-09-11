package main

import (
	"log"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/database"
	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/router"
	"github.com/blues/cfs/internal/scheduler"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化以太坊客户端
	ethClient, err := ethereum.Init(cfg.Ethereum)
	if err != nil {
		log.Fatalf("Failed to initialize ethereum client: %v", err)
	}

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化路由
	r := router.Setup(db, ethClient, cfg)

	// 启动定时任务
	scheduler.Start(db, ethClient, cfg)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
