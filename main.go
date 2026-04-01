package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"agent/global"
	log "agent/library/log/logger"
	"agent/router"
	"agent/service/agent"
	"agent/service/calculator"
	"agent/service/weather"
	"agent/webapi/controllers"

	"github.com/ollama/ollama/api"
)

// 版本信息，通过 ldflags 注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// 显示版本信息
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Agent Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 创建带logid的上下文
	ctx := log.WithContext(context.Background())
	log.Info(ctx, "启动 Agent 服务...")
	log.Info(ctx, "Version: %s, Build Time: %s, Git Commit: %s", Version, BuildTime, GitCommit)

	// 加载配置
	cfg := global.DefaultConfig
	cfg.Server.Port = "8080"

	// 初始化日志
	log.Init(cfg.Log)
	log.Info(ctx, "日志系统初始化完成")

	// 初始化 Ollama 客户端
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Error(ctx, "初始化 Ollama 客户端失败: %v", err)
		// 使用旧的日志方式兼容
		log.InfoOld("初始化 Ollama 客户端失败: %v", err)
		panic(err)
	}
	log.Info(ctx, "Ollama 客户端初始化成功")

	// 初始化工具管理器
	toolManager := agent.NewToolManager()
	toolManager.Register(calculator.NewCalculator())
	toolManager.Register(weather.NewWeather())
	log.Info(ctx, "工具注册完成: calculator, weather")

	// 初始化服务
	ollamaSvc := agent.NewOllamaService(client, cfg.Ollama.Model)
	agentSvc := agent.NewAgentService(ollamaSvc, toolManager)
	log.Info(ctx, "Agent 服务初始化完成")

	// 初始化控制器
	chatCtrl := controllers.NewChatController(agentSvc)
	log.Info(ctx, "控制器初始化完成")

	// 初始化路由
	r := router.NewRouter(chatCtrl)
	mux := http.NewServeMux()
	r.RegisterRoutes(mux)

	log.Info(ctx, "服务启动成功，监听端口 %s", cfg.Server.Port)
	fmt.Println("=== Agent 服务已启动 ===")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("前端地址：http://localhost:" + cfg.Server.Port)
	fmt.Println("API 地址：http://localhost:" + cfg.Server.Port + "/api/chat")

	// 优雅退出
	go func() {
		if err := http.ListenAndServe(":"+cfg.Server.Port, mux); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "服务启动失败: %v", err)
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(ctx, "收到退出信号，正在关闭服务...")
	os.Exit(0)
}
