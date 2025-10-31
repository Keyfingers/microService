package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/logger"
	"github.com/zhang/microservice/internal/queue"
	"go.uber.org/zap"
)

// MessageRequest 消息请求
type MessageRequest struct {
	Queue   string      `json:"queue" binding:"required"`
	Message interface{} `json:"message" binding:"required"`
}

// PublishMessage 发布消息处理器
// 用途: 发送消息到消息队列
// 返回:
//
//	gin.HandlerFunc: Gin 处理器函数
func PublishMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get("request_id")

		var req MessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("解析请求失败",
				zap.String("request_id", requestID.(string)),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请求参数错误",
			})
			return
		}

		// 将消息序列化为 JSON
		messageBody, err := json.Marshal(req.Message)
		if err != nil {
			logger.Error("序列化消息失败",
				zap.String("request_id", requestID.(string)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "处理消息失败",
			})
			return
		}

		// 发布消息到队列
		routingKey := req.Queue + ".*"
		if err := queue.MQClient.Publish(routingKey, messageBody); err != nil {
			logger.Error("发布消息失败",
				zap.String("request_id", requestID.(string)),
				zap.String("queue", req.Queue),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "发送消息失败",
			})
			return
		}

		logger.Info("消息发布成功",
			zap.String("request_id", requestID.(string)),
			zap.String("queue", req.Queue),
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "消息发送成功",
		})
	}
}
