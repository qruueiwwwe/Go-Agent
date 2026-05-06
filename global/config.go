package global

import (
	"time"
)

// Config 全局配置
type Config struct {
	Server     ServerConfig
	Ollama     OllamaConfig
	Log        LogConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	WeatherAPI WeatherAPIConfig
	Zhipu      ZhipuConfig
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
	Host        string        // 地址
	Model       string        // 模型名称
	Timeout     time.Duration // 超时时间
	Temperature float64       // 温度参数（0-1，越低越稳定）
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

// WeatherAPIConfig 天气API配置
type WeatherAPIConfig struct {
	ID      string        // 接口盒子开发者ID
	Key     string        // 接口盒子开发者KEY
	AmapKey string        // 高德天气API Key
	BaseURL string        // API地址
	Timeout time.Duration // 超时时间
}

// ZhipuConfig 智谱清言配置
type ZhipuConfig struct {
	APIKey      string        // API密钥（从环境变量读取）
	Model       string        // 模型名称
	BaseURL     string        // API地址
	Timeout     time.Duration // 超时时间
	Temperature float64       // 温度参数
	Enable      bool          // 是否启用后备
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Server: ServerConfig{
		Port:         "25565",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	},
	Ollama: OllamaConfig{
		Host:        "localhost:11434",
		Model:       "gemma3:4b",
		Timeout:     120 * time.Second,
		Temperature: 0.3,
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
		Password: "", // 从环境变量 DATABASE_PASSWORD 读取
		DBName:   "goagent",
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
	WeatherAPI: WeatherAPIConfig{
		ID:      "", // 从环境变量 WEATHER_API_ID 读取
		Key:     "", // 从环境变量 WEATHER_API_KEY 读取
		AmapKey: "", // 从环境变量 AMAP_API_KEY 读取
		BaseURL: "https://cn.apihz.cn/api/tianqi/tqyb.php",
		Timeout: 10 * time.Second,
	},
	Zhipu: ZhipuConfig{
		APIKey:      "", // 从环境变量 ZHIPU_API_KEY 读取
		Model:       "glm-4-flash",
		BaseURL:     "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		Timeout:     60 * time.Second,
		Temperature: 0.3,
		Enable:      true,
	},
}
