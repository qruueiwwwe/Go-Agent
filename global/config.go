package global

import (
	"time"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig
	Ollama   OllamaConfig
	Log      LogConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port         string        // 端口
	ReadTimeout  time.Duration // 读取超时
	WriteTimeout time.Duration // 写入超时
	IdleTimeout  time.Duration // 空闲超时
}

// OllamaConfig Ollama 配置
type OllamaConfig struct {
	Host      string // 地址
	Model     string // 模型名称
	Timeout   time.Duration // 超时时间
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string // 日志级别
	Path       string // 日志路径
	MaxSize    int64  // 单文件最大大小(MB)
	MaxBackups int    // 最多保留文件数
	MaxAge     int    // 保留天数
	Compress   bool   // 是否压缩
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	MaxOpen  int
	MaxIdle  int
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Server: ServerConfig{
		Port:         "8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	},
	Ollama: OllamaConfig{
		Host:    "localhost:11434",
		Model:   "qwen:7b",
		Timeout: 120 * time.Second,
	},
	Log: LogConfig{
		Level:      "info",
		Path:       "./logs",
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     7,
		Compress:   true,
	},
	Database: DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		DBName:   "agent",
		MaxOpen:  10,
		MaxIdle:  5,
	},
	Redis: RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	},
}
