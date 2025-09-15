package router

import (
	"time"

	"github.com/blues/cfs/internal/config"
	"github.com/blues/cfs/internal/handler"
	"github.com/blues/cfs/internal/logger"
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
	r.Use(customLoggerMiddleware())
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
			projects.GET("/:id/stats", projectHandler.GetProjectStats)
			projects.GET("/stats", projectHandler.GetAllProjectStats)
		}

		// 记录相关路由组
		recordGroup := v1.Group("/record")
		{
			// 贡献记录路由
			contribute := recordGroup.Group("/contribute")
			{
				contribute.GET("/project/:id", contributeHandler.GetProjectContributeRecords)
				contribute.GET("/stats", contributeHandler.GetContributeStats)
			}

			// 退款记录路由
			refund := recordGroup.Group("/refund")
			{
				refund.GET("/project/:id", refundHandler.GetProjectRefunds)
				refund.GET("/stats", refundHandler.GetRefundStats)
			}
		}
	}

	return r
}

// customLoggerMiddleware 自定义日志中间件，使用我们的logger系统
func customLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// 根据状态码选择日志级别
		switch {
		case statusCode >= 500:
			logger.Error("HTTP %s %s %d %v %s %d", method, path, statusCode, latency, clientIP, bodySize)
		case statusCode >= 400:
			logger.Warn("HTTP %s %s %d %v %s %d", method, path, statusCode, latency, clientIP, bodySize)
		default:
			logger.Info("HTTP %s %s %d %v %s %d", method, path, statusCode, latency, clientIP, bodySize)
		}
	}
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
