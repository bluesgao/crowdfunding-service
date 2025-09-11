package router

import (
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/ethereum"
	"github.com/blues/cfs/internal/handler"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, ethClient *ethereum.Client, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "crowdfunding-service",
		})
	})

	// API版本组
	v1 := r.Group("/api/v1")
	{
		// 项目相关路由
		projectHandler := handler.NewProjectHandler(db)
		projects := v1.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.CancelProject)
			projects.GET("/:id/contributions", projectHandler.GetProjectContributions)
			projects.GET("/:id/stats", projectHandler.GetProjectStats)
		}

	}

	return r
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
