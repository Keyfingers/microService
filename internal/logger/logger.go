package logger

import (
	"fmt"
	"os"

	"github.com/zhang/microservice/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志实例
var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// Init 初始化日志系统
// 参数:
//
//	cfg: 日志配置
//
// 返回:
//
//	error: 错误信息
func Init(cfg config.LoggerConfig) error {
	// 设置日志级别
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置输出路径
	var cores []zapcore.Core

	// 普通日志输出
	for _, path := range cfg.OutputPaths {
		writer, err := getWriter(path)
		if err != nil {
			return fmt.Errorf("创建日志输出失败: %w", err)
		}
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			level,
		)
		cores = append(cores, core)
	}

	// 错误日志输出
	for _, path := range cfg.ErrorOutputPaths {
		writer, err := getWriter(path)
		if err != nil {
			return fmt.Errorf("创建错误日志输出失败: %w", err)
		}
		core := zapcore.NewCore(
			encoder,
			zapcore.AddSync(writer),
			zapcore.ErrorLevel,
		)
		cores = append(cores, core)
	}

	// 创建 logger
	options := []zap.Option{}
	if cfg.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if cfg.EnableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	Logger = zap.New(zapcore.NewTee(cores...), options...)
	Sugar = Logger.Sugar()

	return nil
}

// getWriter 获取日志输出 Writer
// 参数:
//
//	path: 输出路径
//
// 返回:
//
//	zapcore.WriteSyncer: 日志写入器
//	error: 错误信息
func getWriter(path string) (zapcore.WriteSyncer, error) {
	if path == "stdout" {
		return zapcore.AddSync(os.Stdout), nil
	}
	if path == "stderr" {
		return zapcore.AddSync(os.Stderr), nil
	}

	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, err
	}

	// 打开文件
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return zapcore.AddSync(file), nil
}

// Sync 刷新日志缓冲区
// 在程序退出前应该调用此方法
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Debug 记录 Debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info 记录 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn 记录 Warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 记录 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 记录 Fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// WithRequestID 创建带有请求 ID 的日志记录器
// 参数:
//
//	requestID: 请求 ID
//
// 返回:
//
//	*zap.Logger: 带有请求 ID 的日志记录器
func WithRequestID(requestID string) *zap.Logger {
	return Logger.With(zap.String("request_id", requestID))
}
