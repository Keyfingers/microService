package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	AWS        AWSConfig        `mapstructure:"aws"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Cron       CronConfig       `mapstructure:"cron"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
	GRPC       GRPCConfig       `mapstructure:"grpc"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	GatewayPort     int    `mapstructure:"gateway_port"`
	GRPCPort        int    `mapstructure:"grpc_port"`
	Mode            string `mapstructure:"mode"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	LogMode         bool   `mapstructure:"log_mode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// RabbitMQConfig RabbitMQ 配置
type RabbitMQConfig struct {
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	User     string         `mapstructure:"user"`
	Password string         `mapstructure:"password"`
	Vhost    string         `mapstructure:"vhost"`
	Exchange ExchangeConfig `mapstructure:"exchange"`
	Queues   []QueueConfig  `mapstructure:"queues"`
}

// ExchangeConfig 交换机配置
type ExchangeConfig struct {
	Name    string `mapstructure:"name"`
	Type    string `mapstructure:"type"`
	Durable bool   `mapstructure:"durable"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	Name       string `mapstructure:"name"`
	RoutingKey string `mapstructure:"routing_key"`
	Durable    bool   `mapstructure:"durable"`
}

// AWSConfig AWS 配置
type AWSConfig struct {
	Region    string   `mapstructure:"region"`
	AccessKey string   `mapstructure:"access_key"`
	SecretKey string   `mapstructure:"secret_key"`
	S3        S3Config `mapstructure:"s3"`
}

// S3Config S3 配置
type S3Config struct {
	Bucket          string `mapstructure:"bucket"`
	UploadPrefix    string `mapstructure:"upload_prefix"`
	PresignedExpire int    `mapstructure:"presigned_expire"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level            string   `mapstructure:"level"`
	Format           string   `mapstructure:"format"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	EnableCaller     bool     `mapstructure:"enable_caller"`
	EnableStacktrace bool     `mapstructure:"enable_stacktrace"`
}

// CronConfig 定时任务配置
type CronConfig struct {
	Enable bool        `mapstructure:"enable"`
	Jobs   []JobConfig `mapstructure:"jobs"`
}

// JobConfig 任务配置
type JobConfig struct {
	Name    string `mapstructure:"name"`
	Spec    string `mapstructure:"spec"`
	Enabled bool   `mapstructure:"enabled"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	RequestLog RequestLogConfig `mapstructure:"request_log"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	Enable           bool     `mapstructure:"enable"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enable            bool `mapstructure:"enable"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// RequestLogConfig 请求日志配置
type RequestLogConfig struct {
	Enable          bool `mapstructure:"enable"`
	LogRequestBody  bool `mapstructure:"log_request_body"`
	LogResponseBody bool `mapstructure:"log_response_body"`
}

// GRPCConfig gRPC 配置
type GRPCConfig struct {
	MaxRecvMsgSize    int `mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize    int `mapstructure:"max_send_msg_size"`
	ConnectionTimeout int `mapstructure:"connection_timeout"`
	KeepaliveTime     int `mapstructure:"keepalive_time"`
	KeepaliveTimeout  int `mapstructure:"keepalive_timeout"`
}

// 全局配置实例
var GlobalConfig *Config

// Load 加载配置文件
// 参数:
//
//	configPath: 配置文件路径
//
// 返回:
//
//	error: 错误信息
func Load(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 支持环境变量覆盖配置
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	return nil
}

// GetDatabaseDSN 获取数据库连接字符串
// 返回:
//
//	string: PostgreSQL 连接字符串
func (c *DatabaseConfig) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
}

// GetRedisAddr 获取 Redis 地址
// 返回:
//
//	string: Redis 地址 (host:port)
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetRabbitMQURL 获取 RabbitMQ 连接地址
// 返回:
//
//	string: RabbitMQ AMQP URL
func (c *RabbitMQConfig) GetRabbitMQURL() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d%s",
		c.User, c.Password, c.Host, c.Port, c.Vhost,
	)
}

// GetConnMaxLifetime 获取连接最大生命周期时间
// 返回:
//
//	time.Duration: 连接最大生命周期
func (c *DatabaseConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Minute
}

// GetShutdownTimeout 获取优雅关闭超时时间
// 返回:
//
//	time.Duration: 超时时间
func (c *ServerConfig) GetShutdownTimeout() time.Duration {
	return time.Duration(c.ShutdownTimeout) * time.Second
}

// GetPresignedExpire 获取预签名 URL 过期时间
// 返回:
//
//	time.Duration: 过期时间
func (c *S3Config) GetPresignedExpire() time.Duration {
	return time.Duration(c.PresignedExpire) * time.Minute
}
