package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/logger"
	"github.com/zhang/microservice/internal/storage"
	"go.uber.org/zap"
)

// UploadRequest 上传请求
type UploadRequest struct {
	File interface{} `form:"file" binding:"required"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// UploadFile 文件上传处理器
// 用途: 处理文件上传到 S3
// 返回:
//
//	gin.HandlerFunc: Gin 处理器函数
func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get("request_id")

		// 获取上传的文件
		file, err := c.FormFile("file")
		if err != nil {
			logger.Error("获取上传文件失败",
				zap.String("request_id", requestID.(string)),
				zap.Error(err),
			)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请上传文件",
			})
			return
		}

		// 打开文件
		src, err := file.Open()
		if err != nil {
			logger.Error("打开上传文件失败",
				zap.String("request_id", requestID.(string)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "处理文件失败",
			})
			return
		}
		defer src.Close()

		// 上传到 S3
		url, key, err := storage.S3Storage.Upload(file.Filename, src, file.Header.Get("Content-Type"))
		if err != nil {
			logger.Error("上传文件到 S3 失败",
				zap.String("request_id", requestID.(string)),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "上传文件失败",
			})
			return
		}

		c.JSON(http.StatusOK, UploadResponse{
			URL: url,
			Key: key,
		})
	}
}

// GetPresignedURL 获取预签名 URL 处理器
// 用途: 生成文件的临时访问 URL
// 返回:
//
//	gin.HandlerFunc: Gin 处理器函数
func GetPresignedURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		key := c.Query("key")

		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请提供文件 key",
			})
			return
		}

		// 生成预签名 URL
		url, err := storage.S3Storage.GetPresignedURL(key)
		if err != nil {
			logger.Error("生成预签名 URL 失败",
				zap.String("request_id", requestID.(string)),
				zap.String("key", key),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "生成访问链接失败",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"url": url,
		})
	}
}
