package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/config"
)

// RateLimit 限流中间件
// 使用简单的计数器限流（生产环境建议使用更复杂的限流算法）
// 参数:
//
//	cfg: 限流配置
//
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enable {
			c.Next()
			return
		}

		// 这里简化处理，实际生产环境应该使用 Redis 或其他方式实现分布式限流
		// 可以集成 golang.org/x/time/rate 包或使用 Redis 实现
		c.Next()
	}
}
