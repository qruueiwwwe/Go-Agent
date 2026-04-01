package log

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"agent/global"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	once        sync.Once
	config      global.LogConfig
	logFile     *os.File
)

const (
	// LogIDKey 日志ID在Context中的Key
	LogIDKey = "log_id"
	// CallerKey 调用者信息在Context中的Key
	CallerKey = "caller"
)

// Context 日志上下文
type Context map[string]interface{}

// Init 初始化日志系统
func Init(cfg global.LogConfig) {
	once.Do(func() {
		config = cfg
		initLoggers()
	})
}

// initLoggers 初始化日志记录器
func initLoggers() {
	// 确保日志目录存在
	if config.Path != "" {
		_ = os.MkdirAll(config.Path, 0755)
	}

	// 创建日志文件
	logFilePath := getLogFileName()
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// 如果无法创建文件，只输出到控制台
		infoLogger = log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
		errorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
		return
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	infoLogger = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// getLogFileName 获取日志文件名
func getLogFileName() string {
	now := time.Now()
	return filepath.Join(config.Path, fmt.Sprintf("agent_%s.log", now.Format("2006-01-02")))
}

// GenerateLogID 生成日志ID
func GenerateLogID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// WithContext 创建带日志ID的上下文
func WithContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	// 如果上下文中没有log_id，生成一个新的
	if _, ok := ctx.Value(LogIDKey).(string); !ok {
		ctx = context.WithValue(ctx, LogIDKey, GenerateLogID())
	}
	return ctx
}

// WithLogID 创建带指定日志ID的上下文
func WithLogID(ctx context.Context, logID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, LogIDKey, logID)
}

// GetLogID 从上下文获取日志ID
func GetLogID(ctx context.Context) string {
	if ctx == nil {
		return GenerateLogID()
	}
	if logID, ok := ctx.Value(LogIDKey).(string); ok && logID != "" {
		return logID
	}
	return GenerateLogID()
}

// formatWithCtx 格式化带上下文的日志
func formatWithCtx(ctx context.Context, format string) string {
	logID := GetLogID(ctx)
	return fmt.Sprintf("[%s] %s", logID, format)
}

// Info 记录Info级别日志
func Info(ctx context.Context, format string, v ...interface{}) {
	msg := formatWithCtx(ctx, format)
	infoLogger.Output(2, fmt.Sprintf(msg, v...))
}

// Infof 记录Info级别日志（格式化）
func Infof(ctx context.Context, format string, v ...interface{}) {
	Info(ctx, format, v...)
}

// Error 记录Error级别日志
func Error(ctx context.Context, format string, v ...interface{}) {
	msg := formatWithCtx(ctx, format)
	errorLogger.Output(2, fmt.Sprintf(msg, v...))
}

// Errorf 记录Error级别日志（格式化）
func Errorf(ctx context.Context, format string, v ...interface{}) {
	Error(ctx, format, v...)
}

// Warn 记录Warn级别日志
func Warn(ctx context.Context, format string, v ...interface{}) {
	msg := formatWithCtx(ctx, "[WARN] "+format)
	infoLogger.Output(2, fmt.Sprintf(msg, v...))
}

// Debug 记录Debug级别日志
func Debug(ctx context.Context, format string, v ...interface{}) {
	if config.Level == "debug" {
		msg := formatWithCtx(ctx, "[DEBUG] "+format)
		infoLogger.Output(2, fmt.Sprintf(msg, v...))
	}
}

// WithFields 创建带额外字段的日志上下文
func WithFields(ctx context.Context, fields Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "fields", fields)
}

// WithField 添加单个字段到上下文
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, key, value)
}

// RequestLog 记录HTTP请求日志
func RequestLog(ctx context.Context, method, path, ip string, status int, duration time.Duration) {
	Info(ctx, "HTTP请求: %s %s from %s status=%d duration=%v",
		method, path, ip, status, duration)
}

// ToolLog 记录工具调用日志
func ToolLog(ctx context.Context, toolName, params, result string, duration time.Duration, err error) {
	if err != nil {
		Error(ctx, "工具调用失败: tool=%s params=%s error=%v duration=%v",
			toolName, params, err, duration)
	} else {
		Info(ctx, "工具调用成功: tool=%s params=%s duration=%v result=%s",
			toolName, params, duration, result)
	}
}

// Old compatibility functions (without ctx)
func init() {
	config = global.DefaultConfig.Log
	initLoggers()
}

// InfoOld 兼容旧的Info调用
func InfoOld(format string, v ...interface{}) {
	infoLogger.Output(2, fmt.Sprintf(format, v...))
}

// ErrorOld 兼容旧的Error调用
func ErrorOld(format string, v ...interface{}) {
	errorLogger.Output(2, fmt.Sprintf(format, v...))
}
