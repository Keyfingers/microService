package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
)

// RedisClient 全局 Redis 客户端实例
var RedisClient *redis.Client

// Init 初始化 Redis 连接
// 参数:
//
//	cfg: Redis 配置
//
// 返回:
//
//	error: 错误信息
func Init(cfg config.RedisConfig) error {
	// 创建 Redis 客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	logger.Info("Redis 连接成功",
		zap.String("addr", cfg.GetRedisAddr()),
		zap.Int("db", cfg.DB),
	)

	return nil
}

// Close 关闭 Redis 连接
// 返回:
//
//	error: 错误信息
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// Get 获取键值
// 参数:
//
//	ctx: 上下文
//	key: 键名
//
// 返回:
//
//	string: 值
//	error: 错误信息
func Get(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// Set 设置键值
// 参数:
//
//	ctx: 上下文
//	key: 键名
//	value: 值
//	expiration: 过期时间（0表示永不过期）
//
// 返回:
//
//	error: 错误信息
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// Delete 删除键
// 参数:
//
//	ctx: 上下文
//	keys: 键名列表
//
// 返回:
//
//	error: 错误信息
func Delete(ctx context.Context, keys ...string) error {
	return RedisClient.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
// 参数:
//
//	ctx: 上下文
//	keys: 键名列表
//
// 返回:
//
//	int64: 存在的键数量
//	error: 错误信息
func Exists(ctx context.Context, keys ...string) (int64, error) {
	return RedisClient.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
// 参数:
//
//	ctx: 上下文
//	key: 键名
//	expiration: 过期时间
//
// 返回:
//
//	error: 错误信息
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	return RedisClient.Expire(ctx, key, expiration).Err()
}

// Incr 键值自增
// 参数:
//
//	ctx: 上下文
//	key: 键名
//
// 返回:
//
//	int64: 自增后的值
//	error: 错误信息
func Incr(ctx context.Context, key string) (int64, error) {
	return RedisClient.Incr(ctx, key).Result()
}

// Decr 键值自减
// 参数:
//
//	ctx: 上下文
//	key: 键名
//
// 返回:
//
//	int64: 自减后的值
//	error: 错误信息
func Decr(ctx context.Context, key string) (int64, error) {
	return RedisClient.Decr(ctx, key).Result()
}

// Lock 获取分布式锁
// 参数:
//
//	ctx: 上下文
//	key: 锁的键名
//	expiration: 锁的过期时间
//
// 返回:
//
//	bool: 是否成功获取锁
//	error: 错误信息
func Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// 使用 SET NX EX 命令实现分布式锁
	return RedisClient.SetNX(ctx, key, "locked", expiration).Result()
}

// Unlock 释放分布式锁
// 参数:
//
//	ctx: 上下文
//	key: 锁的键名
//
// 返回:
//
//	error: 错误信息
func Unlock(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// HGet 获取哈希字段值
// 参数:
//
//	ctx: 上下文
//	key: 哈希键名
//	field: 字段名
//
// 返回:
//
//	string: 字段值
//	error: 错误信息
func HGet(ctx context.Context, key, field string) (string, error) {
	return RedisClient.HGet(ctx, key, field).Result()
}

// HSet 设置哈希字段值
// 参数:
//
//	ctx: 上下文
//	key: 哈希键名
//	field: 字段名
//	value: 字段值
//
// 返回:
//
//	error: 错误信息
func HSet(ctx context.Context, key, field string, value interface{}) error {
	return RedisClient.HSet(ctx, key, field, value).Err()
}

// HGetAll 获取哈希所有字段
// 参数:
//
//	ctx: 上下文
//	key: 哈希键名
//
// 返回:
//
//	map[string]string: 所有字段和值
//	error: 错误信息
func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return RedisClient.HGetAll(ctx, key).Result()
}

// HealthCheck Redis 健康检查
// 返回:
//
//	error: 错误信息
func HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return RedisClient.Ping(ctx).Err()
}
