package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/cache"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/database"
	"github.com/zhang/microservice/internal/handler"
	"github.com/zhang/microservice/internal/logger"
	"github.com/zhang/microservice/internal/middleware"
	"github.com/zhang/microservice/internal/queue"
	"github.com/zhang/microservice/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := config.Load("config/config.yaml"); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(config.GlobalConfig.Logger); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("网关服务启动中...")

	// 初始化数据库
	if err := database.Init(config.GlobalConfig.Database); err != nil {
		logger.Fatal("初始化数据库失败", zap.Error(err))
	}
	defer database.Close()

	// 初始化 Redis
	if err := cache.Init(config.GlobalConfig.Redis); err != nil {
		logger.Fatal("初始化 Redis 失败", zap.Error(err))
	}
	defer cache.Close()

	// 初始化消息队列
	if err := queue.Init(config.GlobalConfig.RabbitMQ); err != nil {
		logger.Fatal("初始化消息队列失败", zap.Error(err))
	}
	defer queue.Close()

	// 初始化 S3 存储
	if err := storage.Init(config.GlobalConfig.AWS); err != nil {
		logger.Fatal("初始化 S3 存储失败", zap.Error(err))
	}

	// 设置 Gin 模式
	gin.SetMode(config.GlobalConfig.Server.Mode)

	// 创建路由
	router := setupRouter()

	// 创建 HTTP 服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.GatewayPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动服务器
	go func() {
		logger.Info("网关服务启动成功",
			zap.String("地址", addr),
			zap.String("模式", config.GlobalConfig.Server.Mode),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.GlobalConfig.Server.GetShutdownTimeout(),
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已关闭")
}

// setupRouter 设置路由
// 返回:
//
//	*gin.Engine: Gin 路由引擎
func setupRouter() *gin.Engine {
	router := gin.New()

	// 使用中间件
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(config.GlobalConfig.Middleware.CORS))
	router.Use(middleware.RateLimit(config.GlobalConfig.Middleware.RateLimit))

	// 健康检查
	router.GET("/health", handler.HealthCheck())
	router.GET("/health/detail", handler.DetailedHealthCheck())

	// API 路由组
	v1 := router.Group("/api/v1")
	{
		// 文件上传
		v1.POST("/upload", handler.UploadFile())
		v1.GET("/presigned-url", handler.GetPresignedURL())

		// 消息队列
		v1.POST("/message", handler.PublishMessage())
	}

	return router
}
