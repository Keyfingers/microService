# 金融级系统升级方案

## 概述

将微服务系统升级到金融级标准，需要在安全性、可靠性、合规性、高可用性等多个维度进行全面优化。

---

## 一、安全性增强 🔐

### 1.1 数据加密

#### 传输层加密 (TLS/SSL)
```go
// internal/security/tls.go
package security

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

// LoadTLSConfig 加载 TLS 配置
func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// 加载 CA 证书
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13, // 强制使用 TLS 1.3
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}, nil
}
```

#### 数据库字段加密
```go
// internal/security/encryption.go
package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
func NewEncryptor(key string) *Encryptor {
	return &Encryptor{key: []byte(key)}
}

// Encrypt 加密敏感数据
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密敏感数据
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
```

### 1.2 身份认证与授权

#### JWT 认证中间件
```go
// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("your-secret-key") // 应从配置读取

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		// 提取 token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		// 解析 token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole 角色权限检查
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
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

		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		c.Abort()
	}
}

// GenerateToken 生成 JWT token
func GenerateToken(userID int64, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
```

### 1.3 审计日志

```go
// internal/audit/audit.go
package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zhang/microservice/internal/database"
	"go.uber.org/zap"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`     // 操作类型
	Resource   string    `json:"resource"`   // 资源类型
	ResourceID string    `json:"resource_id"` // 资源ID
	Method     string    `json:"method"`     // HTTP方法
	Path       string    `json:"path"`       // 请求路径
	IP         string    `json:"ip"`         // 客户端IP
	UserAgent  string    `json:"user_agent"` // 用户代理
	RequestBody  string  `json:"request_body"`  // 请求体
	ResponseCode int     `json:"response_code"` // 响应码
	Status     string    `json:"status"`     // success/failed
	ErrorMsg   string    `json:"error_msg"`  // 错误信息
	Duration   int64     `json:"duration"`   // 耗时(ms)
	CreatedAt  time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// Logger 审计日志记录器
type Logger struct {
	logger *zap.Logger
}

// NewLogger 创建审计日志记录器
func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{logger: logger}
}

// Log 记录审计日志
func (l *Logger) Log(ctx context.Context, log *AuditLog) error {
	log.CreatedAt = time.Now()

	// 写入数据库
	if err := database.DB.WithContext(ctx).Create(log).Error; err != nil {
		l.logger.Error("写入审计日志失败",
			zap.Error(err),
			zap.Any("audit_log", log),
		)
		return err
	}

	// 记录到日志文件
	logJSON, _ := json.Marshal(log)
	l.logger.Info("审计日志", zap.String("data", string(logJSON)))

	return nil
}

// QueryLogs 查询审计日志
func (l *Logger) QueryLogs(ctx context.Context, userID int64, startTime, endTime time.Time) ([]*AuditLog, error) {
	var logs []*AuditLog
	query := database.DB.WithContext(ctx).Model(&AuditLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	err := query.Order("created_at DESC").Limit(1000).Find(&logs).Error
	return logs, err
}
```

### 1.4 审计中间件

```go
// internal/middleware/audit.go
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/audit"
)

// AuditMiddleware 审计中间件
func AuditMiddleware(auditLogger *audit.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			// 恢复请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 记录审计日志
		duration := time.Since(startTime).Milliseconds()
		
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		log := &audit.AuditLog{
			UserID:       getUserID(userID),
			Username:     getUsername(username),
			Action:       getAction(c),
			Resource:     getResource(c),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  maskSensitiveData(requestBody),
			ResponseCode: c.Writer.Status(),
			Status:       getStatus(c.Writer.Status()),
			Duration:     duration,
		}

		if len(c.Errors) > 0 {
			log.ErrorMsg = c.Errors.String()
		}

		_ = auditLogger.Log(c.Request.Context(), log)
	}
}

func getUserID(val interface{}) int64 {
	if val == nil {
		return 0
	}
	return val.(int64)
}

func getUsername(val interface{}) string {
	if val == nil {
		return "anonymous"
	}
	return val.(string)
}

func getAction(c *gin.Context) string {
	return c.Request.Method + " " + c.Request.URL.Path
}

func getResource(c *gin.Context) string {
	// 从路径提取资源类型
	return c.Param("resource")
}

func getStatus(code int) string {
	if code >= 200 && code < 300 {
		return "success"
	}
	return "failed"
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string) string {
	// TODO: 实现敏感字段脱敏逻辑
	// 例如：密码、银行卡号、身份证号等
	return data
}
```

---

## 二、分布式事务处理 💎

### 2.1 Saga 模式实现

```go
// internal/transaction/saga.go
package transaction

import (
	"context"
	"fmt"
)

// Step 事务步骤
type Step struct {
	Name       string
	Do         func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// Saga 分布式事务协调器
type Saga struct {
	steps         []Step
	executedSteps []int
}

// NewSaga 创建 Saga
func NewSaga() *Saga {
	return &Saga{
		steps:         make([]Step, 0),
		executedSteps: make([]int, 0),
	}
}

// AddStep 添加步骤
func (s *Saga) AddStep(step Step) *Saga {
	s.steps = append(s.steps, step)
	return s
}

// Execute 执行 Saga
func (s *Saga) Execute(ctx context.Context) error {
	// 执行所有步骤
	for i, step := range s.steps {
		if err := step.Do(ctx); err != nil {
			// 执行失败，触发补偿
			s.compensate(ctx)
			return fmt.Errorf("步骤 %s 执行失败: %w", step.Name, err)
		}
		s.executedSteps = append(s.executedSteps, i)
	}
	return nil
}

// compensate 执行补偿操作
func (s *Saga) compensate(ctx context.Context) {
	// 反向执行补偿
	for i := len(s.executedSteps) - 1; i >= 0; i-- {
		stepIndex := s.executedSteps[i]
		step := s.steps[stepIndex]
		
		if err := step.Compensate(ctx); err != nil {
			// 记录补偿失败（需要人工介入）
			fmt.Printf("补偿步骤 %s 失败: %v\n", step.Name, err)
		}
	}
}
```

### 2.2 使用示例：转账业务

```go
// internal/service/transfer.go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/zhang/microservice/internal/transaction"
)

// TransferService 转账服务
type TransferService struct{}

// Transfer 转账（使用 Saga 模式）
func (s *TransferService) Transfer(ctx context.Context, fromAccount, toAccount string, amount float64) error {
	saga := transaction.NewSaga()

	// 步骤1: 扣减源账户
	saga.AddStep(transaction.Step{
		Name: "扣减源账户",
		Do: func(ctx context.Context) error {
			return s.debitAccount(ctx, fromAccount, amount)
		},
		Compensate: func(ctx context.Context) error {
			return s.creditAccount(ctx, fromAccount, amount)
		},
	})

	// 步骤2: 增加目标账户
	saga.AddStep(transaction.Step{
		Name: "增加目标账户",
		Do: func(ctx context.Context) error {
			return s.creditAccount(ctx, toAccount, amount)
		},
		Compensate: func(ctx context.Context) error {
			return s.debitAccount(ctx, toAccount, amount)
		},
	})

	// 步骤3: 记录转账日志
	saga.AddStep(transaction.Step{
		Name: "记录转账日志",
		Do: func(ctx context.Context) error {
			return s.logTransfer(ctx, fromAccount, toAccount, amount)
		},
		Compensate: func(ctx context.Context) error {
			return s.deleteTransferLog(ctx, fromAccount, toAccount, amount)
		},
	})

	return saga.Execute(ctx)
}

func (s *TransferService) debitAccount(ctx context.Context, account string, amount float64) error {
	// 实现扣款逻辑
	return nil
}

func (s *TransferService) creditAccount(ctx context.Context, account string, amount float64) error {
	// 实现入账逻辑
	return nil
}

func (s *TransferService) logTransfer(ctx context.Context, from, to string, amount float64) error {
	// 记录转账日志
	return nil
}

func (s *TransferService) deleteTransferLog(ctx context.Context, from, to string, amount float64) error {
	// 删除转账日志
	return nil
}
```

---

## 三、幂等性保证 🔄

### 3.1 幂等性中间件

```go
// internal/middleware/idempotent.go
package middleware

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/cache"
)

// IdempotentMiddleware 幂等性中间件
func IdempotentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对写操作进行幂等性检查
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		// 获取幂等性令牌
		idempotentKey := c.GetHeader("X-Idempotent-Key")
		if idempotentKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "缺少幂等性令牌",
			})
			c.Abort()
			return
		}

		// 生成 Redis key
		redisKey := fmt.Sprintf("idempotent:%s", idempotentKey)
		ctx := context.Background()

		// 检查是否已执行
		exists, _ := cache.Exists(ctx, redisKey)
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "请求已处理，请勿重复提交",
			})
			c.Abort()
			return
		}

		// 使用分布式锁防止并发
		lockKey := fmt.Sprintf("lock:%s", idempotentKey)
		locked, err := cache.Lock(ctx, lockKey, 30*time.Second)
		if err != nil || !locked {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求处理中，请稍后",
			})
			c.Abort()
			return
		}
		defer cache.Unlock(ctx, lockKey)

		// 处理请求
		c.Next()

		// 如果成功，标记为已执行（保留24小时）
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			_ = cache.Set(ctx, redisKey, "1", 24*time.Hour)
		}
	}
}

// GenerateIdempotentKey 生成幂等性令牌
func GenerateIdempotentKey(userID int64, action string, params ...string) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%d:%s:%v", userID, action, params))
	return hex.EncodeToString(h.Sum(nil))
}
```

---

## 四、限流与熔断 ⚡

### 4.1 令牌桶限流

```go
// internal/ratelimit/token_bucket.go
package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/zhang/microservice/internal/cache"
)

// TokenBucket 令牌桶
type TokenBucket struct {
	key      string
	capacity int64 // 桶容量
	rate     int64 // 每秒生成令牌数
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(key string, capacity, rate int64) *TokenBucket {
	return &TokenBucket{
		key:      key,
		capacity: capacity,
		rate:     rate,
	}
}

// Allow 尝试获取令牌
func (tb *TokenBucket) Allow(ctx context.Context) (bool, error) {
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_time')
		local tokens = tonumber(bucket[1]) or capacity
		local last_time = tonumber(bucket[2]) or now
		
		-- 计算新增令牌
		local delta = now - last_time
		local new_tokens = math.min(capacity, tokens + delta * rate)
		
		if new_tokens >= 1 then
			new_tokens = new_tokens - 1
			redis.call('HMSET', key, 'tokens', new_tokens, 'last_time', now)
			redis.call('EXPIRE', key, 3600)
			return 1
		else
			redis.call('HMSET', key, 'tokens', new_tokens, 'last_time', now)
			return 0
		end
	`

	now := time.Now().Unix()
	result, err := cache.RedisClient.Eval(ctx, script, []string{tb.key}, 
		tb.capacity, tb.rate, now).Int()
	
	if err != nil {
		return false, err
	}

	return result == 1, nil
}
```

### 4.2 熔断器

```go
// internal/circuitbreaker/breaker.go
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State 熔断器状态
type State int

const (
	StateClosed State = iota // 关闭状态（正常）
	StateOpen                // 开启状态（熔断）
	StateHalfOpen            // 半开状态（试探）
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	successCount    int
	failureThreshold int           // 失败阈值
	timeout         time.Duration  // 熔断超时时间
	openTime        time.Time      // 熔断开始时间
}

var ErrCircuitOpen = errors.New("熔断器已开启")

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		failureThreshold: failureThreshold,
		timeout:         timeout,
	}
}

// Call 执行调用
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.allow() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.record(err)
	return err
}

// allow 是否允许请求
func (cb *CircuitBreaker) allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否超时，可以尝试恢复
		if time.Since(cb.openTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// record 记录结果
func (cb *CircuitBreaker) record(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.successCount = 0

		// 达到失败阈值，开启熔断
		if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			cb.openTime = time.Now()
		} else if cb.state == StateHalfOpen {
			// 半开状态下失败，重新熔断
			cb.state = StateOpen
			cb.openTime = time.Now()
		}
	} else {
		cb.successCount++
		cb.failureCount = 0

		// 半开状态下成功，关闭熔断
		if cb.state == StateHalfOpen && cb.successCount >= 2 {
			cb.state = StateClosed
		}
	}
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
```

---

## 五、数据一致性保证 📊

### 5.1 双写一致性（缓存）

```go
// internal/consistency/cache.go
package consistency

import (
	"context"
	"fmt"
	"time"

	"github.com/zhang/microservice/internal/cache"
	"github.com/zhang/microservice/internal/database"
)

// CacheAside Cache-Aside 模式
type CacheAside struct{}

// Get 读取数据
func (ca *CacheAside) Get(ctx context.Context, key string, loader func() (interface{}, error)) (interface{}, error) {
	// 1. 先查缓存
	val, err := cache.Get(ctx, key)
	if err == nil && val != "" {
		return val, nil
	}

	// 2. 缓存未命中，使用分布式锁
	lockKey := fmt.Sprintf("lock:%s", key)
	locked, _ := cache.Lock(ctx, lockKey, 10*time.Second)
	if !locked {
		// 未获取到锁，稍后重试
		time.Sleep(100 * time.Millisecond)
		return ca.Get(ctx, key, loader)
	}
	defer cache.Unlock(ctx, lockKey)

	// 3. 双重检查
	val, err = cache.Get(ctx, key)
	if err == nil && val != "" {
		return val, nil
	}

	// 4. 加载数据
	data, err := loader()
	if err != nil {
		return nil, err
	}

	// 5. 写入缓存（设置随机过期时间，防止缓存雪崩）
	expiration := time.Duration(300+randInt(300)) * time.Second
	_ = cache.Set(ctx, key, data, expiration)

	return data, nil
}

// Update 更新数据
func (ca *CacheAside) Update(ctx context.Context, key string, updater func() error) error {
	// 1. 更新数据库
	if err := updater(); err != nil {
		return err
	}

	// 2. 删除缓存（而不是更新）
	_ = cache.Delete(ctx, key)

	return nil
}
```

### 5.2 延迟双删策略

```go
// DelayedDoubleDelete 延迟双删
func DelayedDoubleDelete(ctx context.Context, key string, updater func() error) error {
	// 1. 第一次删除缓存
	_ = cache.Delete(ctx, key)

	// 2. 更新数据库
	if err := updater(); err != nil {
		return err
	}

	// 3. 延迟后再次删除缓存
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = cache.Delete(context.Background(), key)
	}()

	return nil
}
```

---

## 六、监控告警 📈

### 6.1 Prometheus 指标

```go
// internal/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求总数
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP 请求延迟
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 数据库查询延迟
	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// 缓存命中率
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)

	// 业务指标
	TransactionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transaction_total",
			Help: "Total number of transactions",
		},
		[]string{"type", "status"},
	)

	TransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "transaction_amount",
			Help: "Transaction amount distribution",
		},
		[]string{"type"},
	)
)
```

### 6.2 监控中间件

```go
// internal/middleware/metrics.go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhang/microservice/internal/metrics"
)

// MetricsMiddleware 监控中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		// 记录请求总数
		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			fmt.Sprintf("%d", status),
		).Inc()

		// 记录请求延迟
		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}
```

---

## 七、配置更新

### 7.1 金融级配置文件

```yaml
# config/config.financial.yaml

# 服务配置
server:
  gateway_port: 8443  # HTTPS 端口
  grpc_port: 50051
  mode: release
  shutdown_timeout: 60
  tls:
    enabled: true
    cert_file: /certs/server.crt
    key_file: /certs/server.key
    ca_file: /certs/ca.crt

# 安全配置
security:
  jwt:
    secret_key: ${JWT_SECRET_KEY}
    expire_hours: 2
    refresh_expire_hours: 168
  encryption:
    key: ${ENCRYPTION_KEY}
    algorithm: AES-256-GCM
  rate_limit:
    enabled: true
    requests_per_second: 100
    burst: 200

# 数据库配置（主从）
database:
  master:
    host: master-db.internal
    port: 5432
    user: ${DB_USER}
    password: ${DB_PASSWORD}
    dbname: financial_prod
    ssl_mode: require
    max_open_conns: 50
    max_idle_conns: 10
  slave:
    host: slave-db.internal
    port: 5432
    user: ${DB_USER}
    password: ${DB_PASSWORD}
    dbname: financial_prod
    ssl_mode: require
    max_open_conns: 100
    max_idle_conns: 20

# 事务配置
transaction:
  saga:
    enabled: true
    timeout: 30s
  idempotent:
    enabled: true
    expire_hours: 24

# 熔断配置
circuit_breaker:
  failure_threshold: 5
  timeout_seconds: 60
  half_open_requests: 3

# 审计配置
audit:
  enabled: true
  retention_days: 2555  # 7年
  sensitive_fields:
    - password
    - card_number
    - id_number
    - bank_account

# 监控配置
monitoring:
  prometheus:
    enabled: true
    port: 9090
  alert:
    enabled: true
    webhook: ${ALERT_WEBHOOK}
```

---

## 八、实施计划

### 阶段一（1-2周）
1. ✅ 实现 JWT 认证和授权
2. ✅ 添加审计日志系统
3. ✅ 实现幂等性保证
4. ✅ 添加数据加密

### 阶段二（2-3周）
1. ✅ 实现 Saga 分布式事务
2. ✅ 添加限流和熔断
3. ✅ 实现缓存一致性
4. ✅ 配置 TLS/SSL

### 阶段三（3-4周）
1. ✅ 集成 Prometheus 监控
2. ✅ 配置告警系统
3. ✅ 完善测试覆盖
4. ✅ 性能压测和优化

### 阶段四（1周）
1. ✅ 安全审计
2. ✅ 压力测试
3. ✅ 灾备演练
4. ✅ 文档完善

---

## 九、测试要求

### 9.1 单元测试
- 代码覆盖率 > 85%
- 关键业务逻辑 100% 覆盖

### 9.2 集成测试
- 所有 API 接口测试
- 异常场景测试
- 并发场景测试

### 9.3 压力测试
- TPS > 10000
- 响应时间 P99 < 100ms
- 错误率 < 0.01%

### 9.4 安全测试
- SQL 注入测试
- XSS 攻击测试
- 权限绕过测试
- 敏感信息泄露测试

---

## 十、部署要求

### 10.1 高可用部署
- 至少 3 个可用区
- 每个服务 >= 3 个实例
- 数据库主从 + 读写分离
- Redis 集群模式

### 10.2 容灾备份
- 数据库定时备份（每小时）
- 增量备份 + 全量备份
- 异地容灾（RPO < 5分钟）
- 定期演练

### 10.3 监控告警
- 服务可用性监控
- 性能指标监控
- 业务指标监控
- 7x24 告警响应

---

## 总结

升级到金融级系统需要：

**安全性**：数据加密、身份认证、审计日志
**可靠性**：分布式事务、幂等性、熔断限流
**一致性**：缓存策略、数据同步
**监控性**：指标采集、告警通知
**合规性**：审计追踪、数据保留

预计投入：**1-2 个月开发 + 1 个月测试**
团队规模：**3-5 人**
成本增加：**服务器 +50%，运维 +30%**

这是一个系统性工程，需要循序渐进，不可操之过急。

