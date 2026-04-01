package app

import (
	"context"
	"fmt"
	"sync"

	"agent/global"
	"agent/log"
)

var (
	startOnce sync.Once
	stopOnce  sync.Once
)

// Service 服务接口
type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// services 存放所有服务
var services = make(map[string]Service)

// RegisterService 注册服务
func RegisterService(svc Service) {
	services[svc.Name()] = svc
	log.Info(context.Background(), "注册服务: %s", svc.Name())
}

// StartAll 启动所有服务
func StartAll(ctx context.Context) error {
	log.Info(ctx, "开始启动所有服务...")
	
	for name, svc := range services {
		log.Info(ctx, "启动服务: %s", name)
		if err := svc.Start(ctx); err != nil {
			log.Error(ctx, "启动服务失败: %s error=%v", name, err)
			return fmt.Errorf("启动服务 %s 失败: %w", name, err)
		}
		log.Info(ctx, "服务启动成功: %s", name)
	}
	
	log.Info(ctx, "所有服务启动完成")
	return nil
}

// StopAll 停止所有服务
func StopAll(ctx context.Context) error {
	log.Info(ctx, "开始停止所有服务...")

	// 反向停止
	for name, svc := range services {
		log.Info(ctx, "停止服务: %s", name)
		if err := svc.Stop(ctx); err != nil {
			log.Error(ctx, "停止服务失败: %s error=%v", name, err)
			continue
		}
		log.Info(ctx, "服务停止成功: %s", name)
	}

	log.Info(ctx, "所有服务已停止")
	return nil
}

// MySQLService MySQL服务（暂未实现）
type MySQLService struct{}

func (m *MySQLService) Name() string        { return "mysql" }
func (m *MySQLService) Start(ctx context.Context) error {
	log.Info(ctx, "MySQL 服务启动（占位）")
	return nil
}
func (m *MySQLService) Stop(ctx context.Context) error {
	log.Info(ctx, "MySQL 服务停止（占位）")
	return nil
}

// RedisService Redis服务（暂未实现）
type RedisService struct{}

func (r *RedisService) Name() string        { return "redis" }
func (r *RedisService) Start(ctx context.Context) error {
	log.Info(ctx, "Redis 服务启动（占位）")
	return nil
}
func (r *RedisService) Stop(ctx context.Context) error {
	log.Info(ctx, "Redis 服务停止（占位）")
	return nil
}

// Register 注册默认服务
func init() {
	RegisterService(&MySQLService{})
	RegisterService(&RedisService{})
}

// Config 应用配置
type Config struct {
	Global global.Config
}

// Run 运行应用
func Run(ctx context.Context, cfg Config) error {
	startOnce.Do(func() {
		// 启动所有服务
		if err := StartAll(ctx); err != nil {
			log.Error(ctx, "启动服务失败: %v", err)
			panic(err)
		}
	})

	// 等待退出信号
	<-ctx.Done()

	log.Info(ctx, "接收到退出信号，正在停止服务...")
	stopOnce.Do(func() {
		if err := StopAll(ctx); err != nil {
			log.Error(ctx, "停止服务失败: %v", err)
		}
	})

	return nil
}

// NewConfig 创建默认配置
func NewConfig() Config {
	return Config{
		Global: global.DefaultConfig,
	}
}