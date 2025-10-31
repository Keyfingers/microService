package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
)

// Claims JWT 声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     []byte
	ExpireTime time.Duration
}

var defaultJWTConfig = &JWTConfig{
	Secret:     []byte("your-secret-key-change-in-production"),
	ExpireTime: 24 * time.Hour,
}

// SetJWTConfig 设置 JWT 配置
func SetJWTConfig(config *JWTConfig) {
	defaultJWTConfig = config
}

// JWTAuth JWT 认证中间件
// 用途: 验证请求中的 JWT token，并将用户信息存入上下文
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("未提供认证令牌",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供认证令牌",
				"code":  "AUTH_TOKEN_MISSING",
			})
			c.Abort()
			return
		}

		// 提取 token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			logger.Warn("认证令牌格式错误",
				zap.String("header", authHeader),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌格式错误",
				"code":  "AUTH_TOKEN_INVALID_FORMAT",
			})
			c.Abort()
			return
		}

		// 解析 token
		tokenString := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return defaultJWTConfig.Secret, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("认证令牌无效",
				zap.Error(err),
				zap.String("token", tokenString[:10]+"..."),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "认证令牌无效或已过期",
				"code":  "AUTH_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		logger.Debug("用户认证成功",
			zap.Int64("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// OptionalJWTAuth 可选的 JWT 认证
// 用途: 如果提供了 token 则验证，未提供则继续处理
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
				return defaultJWTConfig.Secret, nil
			})

			if err == nil && token.Valid {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				c.Set("role", claims.Role)
			}
		}

		c.Next()
	}
}

// RequireRole 角色权限检查中间件
// 用途: 检查用户是否具有指定角色
// 参数:
//
//	roles: 允许的角色列表
//
// 返回:
//
//	gin.HandlerFunc: Gin 中间件函数
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			logger.Warn("未找到用户角色信息",
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusForbidden, gin.H{
				"error": "权限不足",
				"code":  "PERMISSION_DENIED",
			})
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		logger.Warn("用户角色不匹配",
			zap.String("user_role", roleStr),
			zap.Strings("required_roles", roles),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"error": "权限不足",
			"code":  "PERMISSION_DENIED",
		})
		c.Abort()
	}
}

// GenerateToken 生成 JWT token
// 用途: 为用户生成认证令牌
// 参数:
//
//	userID: 用户ID
//	username: 用户名
//	role: 角色
//
// 返回:
//
//	string: JWT token
//	error: 错误信息
func GenerateToken(userID int64, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(defaultJWTConfig.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "microservice",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(defaultJWTConfig.Secret)
	if err != nil {
		logger.Error("生成token失败",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return "", err
	}

	return tokenString, nil
}

// RefreshToken 刷新 token
// 用途: 基于旧 token 生成新 token
// 参数:
//
//	oldToken: 旧的 JWT token
//
// 返回:
//
//	string: 新的 JWT token
//	error: 错误信息
func RefreshToken(oldToken string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(oldToken, claims, func(token *jwt.Token) (interface{}, error) {
		return defaultJWTConfig.Secret, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	// 生成新 token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// GetUserID 从上下文获取用户ID
// 用途: 获取当前请求的用户ID
// 参数:
//
//	c: Gin 上下文
//
// 返回:
//
//	int64: 用户ID
//	bool: 是否存在
func GetUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return val.(int64), true
}

// GetUsername 从上下文获取用户名
// 用途: 获取当前请求的用户名
// 参数:
//
//	c: Gin 上下文
//
// 返回:
//
//	string: 用户名
//	bool: 是否存在
func GetUsername(c *gin.Context) (string, bool) {
	val, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return val.(string), true
}

// GetUserRole 从上下文获取用户角色
// 用途: 获取当前请求的用户角色
// 参数:
//
//	c: Gin 上下文
//
// 返回:
//
//	string: 角色
//	bool: 是否存在
func GetUserRole(c *gin.Context) (string, bool) {
	val, exists := c.Get("role")
	if !exists {
		return "", false
	}
	return val.(string), true
}
