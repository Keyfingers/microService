package storage

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/zhang/microservice/internal/config"
	"github.com/zhang/microservice/internal/logger"
	"go.uber.org/zap"
)

// S3Client S3 客户端
type S3Client struct {
	client *s3.S3
	bucket string
	prefix string
	expire time.Duration
}

// S3Storage 全局 S3 存储实例
var S3Storage *S3Client

// Init 初始化 S3 客户端
// 参数:
//
//	cfg: AWS 配置
//
// 返回:
//
//	error: 错误信息
func Init(cfg config.AWSConfig) error {
	// 创建 AWS 会话
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	})
	if err != nil {
		return fmt.Errorf("创建 AWS 会话失败: %w", err)
	}

	// 创建 S3 客户端
	S3Storage = &S3Client{
		client: s3.New(sess),
		bucket: cfg.S3.Bucket,
		prefix: cfg.S3.UploadPrefix,
		expire: cfg.S3.GetPresignedExpire(),
	}

	logger.Info("S3 客户端初始化成功",
		zap.String("region", cfg.Region),
		zap.String("bucket", cfg.S3.Bucket),
	)

	return nil
}

// Upload 上传文件到 S3
// 参数:
//
//	filename: 文件名
//	content: 文件内容
//	contentType: 文件类型
//
// 返回:
//
//	string: 文件 URL
//	string: 文件 Key
//	error: 错误信息
func (s *S3Client) Upload(filename string, content io.Reader, contentType string) (string, string, error) {
	// 生成文件 key
	key := s.generateKey(filename)

	// 读取文件内容
	buf := new(bytes.Buffer)
	size, err := buf.ReadFrom(content)
	if err != nil {
		return "", "", fmt.Errorf("读取文件内容失败: %w", err)
	}

	// 上传到 S3
	_, err = s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", "", fmt.Errorf("上传文件到 S3 失败: %w", err)
	}

	// 生成文件 URL
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, key)

	logger.Info("文件上传成功",
		zap.String("key", key),
		zap.String("url", url),
		zap.Int64("size", size),
	)

	return url, key, nil
}

// Download 从 S3 下载文件
// 参数:
//
//	key: 文件 Key
//
// 返回:
//
//	io.ReadCloser: 文件内容读取器
//	error: 错误信息
func (s *S3Client) Download(key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("从 S3 下载文件失败: %w", err)
	}

	return result.Body, nil
}

// Delete 从 S3 删除文件
// 参数:
//
//	key: 文件 Key
//
// 返回:
//
//	error: 错误信息
func (s *S3Client) Delete(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("从 S3 删除文件失败: %w", err)
	}

	logger.Info("文件删除成功", zap.String("key", key))
	return nil
}

// GetPresignedURL 生成预签名 URL
// 参数:
//
//	key: 文件 Key
//
// 返回:
//
//	string: 预签名 URL
//	error: 错误信息
func (s *S3Client) GetPresignedURL(key string) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(s.expire)
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	return url, nil
}

// ListFiles 列出文件
// 参数:
//
//	prefix: 文件前缀
//
// 返回:
//
//	[]string: 文件 Key 列表
//	error: 错误信息
func (s *S3Client) ListFiles(prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("列出 S3 文件失败: %w", err)
	}

	var files []string
	for _, item := range result.Contents {
		files = append(files, *item.Key)
	}

	return files, nil
}

// generateKey 生成文件存储 key
// 参数:
//
//	filename: 原始文件名
//
// 返回:
//
//	string: 生成的 Key
func (s *S3Client) generateKey(filename string) string {
	// 使用时间戳避免文件名冲突
	timestamp := time.Now().Format("20060102150405")
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]

	return fmt.Sprintf("%s%s_%s%s", s.prefix, name, timestamp, ext)
}
