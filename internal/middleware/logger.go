package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
)

// Logger 日志中间件
// 记录每个 HTTP 请求的详细信息
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求 ID（使用时间戳+随机数）
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		// 记录请求开始时间
		startTime := time.Now()

		// 记录请求信息
		logger.Info("HTTP 请求开始",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(startTime)

		// 记录响应信息
		logger.Info("HTTP 请求完成",
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		)

		// 如果有错误，记录错误日志
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("请求处理错误",
					zap.String("request_id", requestID),
					zap.Error(err),
				)
			}
		}
	}
}

// Recovery 恢复中间件
// 捕获 panic 并记录错误日志
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				logger.Error("发生 panic",
					zap.String("request_id", requestID.(string)),
					zap.Any("error", err),
					zap.Stack("stacktrace"),
				)

				c.JSON(500, gin.H{
					"error": "内ductservererror",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// generateRequestID 生成请求 ID
// 返回:
//
//	string: 请求 ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
