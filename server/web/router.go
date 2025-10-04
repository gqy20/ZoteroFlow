package web

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// 获取当前工作目录和静态文件路径
	wd, _ := os.Getwd()
	staticPath := filepath.Join(wd, "web", "static")

	// 静态文件服务 - 统一设置避免路由冲突
	r.Static("/static", staticPath)

	// 主页 - 返回index.html
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(filepath.Join(staticPath, "index.html"))
	})

	// API路由组
	api := r.Group("/api")
	{
		api.POST("/ask", HandleAsk)
		api.GET("/status", HandleStatus)
		api.GET("/config", HandleStaticConfig)
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "zoteroflow2-web",
		})
	})

	return r
}

// corsMiddleware CORS中间件
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
