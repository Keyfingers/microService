package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/zhang/microservice/internal/cache"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/database"
	"github.com/zhang/microservice/internal/logger"
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

	logger.Info("定时任务服务启动中...")

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

	// 检查是否启用定时任务
	if !config.GlobalConfig.Cron.Enable {
		logger.Info("定时任务未启用")
		return
	}

	// 创建定时任务调度器
	c := cron.New(cron.WithSeconds())

	// 注册定时任务
	for _, job := range config.GlobalConfig.Cron.Jobs {
		if !job.Enabled {
			logger.Info("跳过未启用的任务", zap.String("任务", job.Name))
			continue
		}

		// 复制变量避免闭包问题
		jobName := job.Name
		jobSpec := job.Spec

		// 添加任务
		_, err := c.AddFunc(jobSpec, func() {
			executeJob(jobName)
		})
		if err != nil {
			logger.Error("注册定时任务失败",
				zap.String("任务", jobName),
				zap.Error(err),
			)
			continue
		}

		logger.Info("注册定时任务成功",
			zap.String("任务", jobName),
			zap.String("表达式", jobSpec),
		)
	}

	// 启动调度器
	c.Start()
	logger.Info("定时任务服务启动成功")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭定时任务服务...")

	// 停止调度器
	ctx := c.Stop()
	<-ctx.Done()

	logger.Info("定时任务服务已关闭")
}

// executeJob 执行定时任务
// 使用分布式锁确保任务不会重复执行
// 参数:
//
//	jobName: 任务名称
func executeJob(jobName string) {
	ctx := context.Background()
	lockKey := fmt.Sprintf("cron:lock:%s", jobName)

	// 尝试获取分布式锁（5分钟过期）
	locked, err := cache.Lock(ctx, lockKey, 5*time.Minute)
	if err != nil {
		logger.Error("获取任务锁失败",
			zap.String("任务", jobName),
			zap.Error(err),
		)
		return
	}

	if !locked {
		logger.Warn("任务正在执行中，跳过本次执行",
			zap.String("任务", jobName),
		)
		return
	}

	// 确保释放锁
	defer func() {
		if err := cache.Unlock(ctx, lockKey); err != nil {
			logger.Error("释放任务锁失败",
				zap.String("任务", jobName),
				zap.Error(err),
			)
		}
	}()

	logger.Info("开始执行定时任务", zap.String("任务", jobName))
	startTime := time.Now()

	// 根据任务名称执行相应的任务
	switch jobName {
	case "clean_expired_data":
		cleanExpiredData()
	case "daily_statistics":
		dailyStatistics()
	case "health_check":
		healthCheck()
	default:
		logger.Warn("未知的任务", zap.String("任务", jobName))
	}

	duration := time.Since(startTime)
	logger.Info("定时任务执行完成",
		zap.String("任务", jobName),
		zap.Duration("耗时", duration),
	)
}

// cleanExpiredData 清理过期数据任务
func cleanExpiredData() {
	logger.Info("执行清理过期数据任务")
	// TODO: 实现具体的清理逻辑
	// 例如：删除过期的缓存、日志、临时文件等
}

// dailyStatistics 每日统计任务
func dailyStatistics() {
	logger.Info("执行每日统计任务")
	// TODO: 实现具体的统计逻辑
	// 例如：统计用户数、订单数、收入等
}

// healthCheck 健康检查任务
func healthCheck() {
	logger.Debug("执行健康检查任务")

	// 检查数据库
	if err := database.HealthCheck(); err != nil {
		logger.Error("数据库健康检查失败", zap.Error(err))
	} else {
		logger.Debug("数据库健康检查通过")
	}

	// 检查 Redis
	if err := cache.HealthCheck(); err != nil {
		logger.Error("Redis 健康检查失败", zap.Error(err))
	} else {
		logger.Debug("Redis 健康检查通过")
	}
}
