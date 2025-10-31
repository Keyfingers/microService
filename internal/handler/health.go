package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/cache"
	"github.com/zhang/microservice/internal/database"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]ServiceInfo `json:"services,omitempty"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// HealthCheck 健康检查处理器
// 用途: 检查服务及其依赖的健康状态
// 返回:
//
//	gin.HandlerFunc: Gin 处理器函数
func HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status:    "ok",
			Timestamp: time.Now().Format(time.RFC3339),
		}

		c.JSON(http.StatusOK, response)
	}
}

// DetailedHealthCheck 详细健康检查处理器
// 用途: 检查服务及其所有依赖（数据库、Redis等）的健康状态
// 返回:
//
//	gin.HandlerFunc: Gin 处理器函数
func DetailedHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		services := make(map[string]ServiceInfo)
		overallStatus := "ok"

		// 检查数据库
		if err := database.HealthCheck(); err != nil {
			services["database"] = ServiceInfo{
				Status:  "error",
				Message: err.Error(),
			}
			overallStatus = "degraded"
		} else {
			services["database"] = ServiceInfo{
				Status: "ok",
			}
		}

		// 检查 Redis
		if err := cache.HealthCheck(); err != nil {
			services["redis"] = ServiceInfo{
				Status:  "error",
				Message: err.Error(),
			}
			overallStatus = "degraded"
		} else {
			services["redis"] = ServiceInfo{
				Status: "ok",
			}
		}

		response := HealthResponse{
			Status:    overallStatus,
			Timestamp: time.Now().Format(time.RFC3339),
			Services:  services,
		}

		// 根据整体状态返回相应的 HTTP 状态码
		statusCode := http.StatusOK
		if overallStatus == "degraded" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, response)
	}
}
