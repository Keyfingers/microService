package database

import (
	"context"
	"fmt"
	"time"

	"github.com/zhang/microservice/internal/config"
	zapLogger "github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库实例
var DB *gorm.DB

// Init 初始化数据库连接
// 参数:
//
//	cfg: 数据库配置
//
// 返回:
//
//	error: 错误信息
func Init(cfg config.DatabaseConfig) error {
	var err error

	// 配置 GORM
	gormConfig := &gorm.Config{}

	// 配置日志
	if cfg.LogMode {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	} else {
		gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	}

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层的 sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	zapLogger.Info("数据库连接成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	return nil
}

// Close 关闭数据库连接
// 返回:
//
//	error: 错误信息
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
// 返回:
//
//	*gorm.DB: 数据库实例
func GetDB() *gorm.DB {
	return DB
}

// Transaction 执行事务
// 参数:
//
//	fn: 事务处理函数
//
// 返回:
//
//	error: 错误信息
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// HealthCheck 健康检查
// 返回:
//
//	error: 错误信息
func HealthCheck() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	ctx, cancel := timeoutContext(5 * time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}

// timeoutContext 创建超时上下文
func timeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
