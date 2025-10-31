package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/config"
)

// CORS 跨域中间件
// 参数:
//
//	cfg: CORS 配置
//
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enable {
			c.Next()
			return
		}

		// 设置允许的源
		if len(cfg.AllowOrigins) > 0 {
			origin := c.Request.Header.Get("Origin")
			for _, allowOrigin := range cfg.AllowOrigins {
				if allowOrigin == "*" || allowOrigin == origin {
					c.Header("Access-Control-Allow-Origin", allowOrigin)
					break
				}
			}
		}

		// 设置允许的方法
		if len(cfg.AllowMethods) > 0 {
			methods := ""
			for i, method := range cfg.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)
		}

		// 设置允许的头
		if len(cfg.AllowHeaders) > 0 {
			headers := ""
			for i, header := range cfg.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
		}

		// 设置暴露的头
		if len(cfg.ExposeHeaders) > 0 {
			headers := ""
			for i, header := range cfg.ExposeHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Expose-Headers", headers)
		}

		// 设置是否允许凭证
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求缓存时间
		if cfg.MaxAge > 0 {
			maxAge := time.Duration(cfg.MaxAge) * time.Hour
			c.Header("Access-Control-Max-Age", string(rune(maxAge.Seconds())))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
