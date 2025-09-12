package router

import (
	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/handler"
	"github.com/gin-gonic/gin"
)

func Setup(
	projectHandler *handler.ProjectHandler,
	contributeHandler *handler.ContributeHandler,
	refundHandler *handler.RefundHandler,
	cfg *config.Config,
) *gin.Engine {
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
		projects := v1.Group("/project")
		{
			projects.POST("", projectHandler.CreateProject)
			projects.GET("", projectHandler.GetProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.PUT("/:id", projectHandler.UpdateProject)
			projects.DELETE("/:id", projectHandler.CancelProject)
			projects.GET("/:id/contributions", projectHandler.GetProjectContributions)
			projects.GET("/:id/stats", projectHandler.GetProjectStats)
		}

		// 记录相关路由组
		recordGroup := v1.Group("/record")
		{
			// 贡献记录路由
			contribute := recordGroup.Group("/contribute")
			{
				contribute.GET("/project/:id", contributeHandler.GetProjectContributeRecords)
				contribute.GET("/user/:address", contributeHandler.GetUserContributeRecords)
				contribute.GET("/tx/:hash", contributeHandler.GetContributeRecordByTxHash)
				contribute.GET("/statistics", contributeHandler.GetContributeStatistics)
			}

			// 退款记录路由
			refund := recordGroup.Group("/refund")
			{
				refund.GET("/project/:id", refundHandler.GetProjectRefunds)
				refund.GET("/user/:address", refundHandler.GetUserRefunds)
				refund.GET("/tx/:hash", refundHandler.GetRefundByTxHash)
				refund.PUT("/:id/status", refundHandler.UpdateRefundStatus)
				refund.GET("/statistics", refundHandler.GetRefundStatistics)
			}
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
