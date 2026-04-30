package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"agent/global"
	"agent/library/log"
	"agent/models/dao"
	"agent/models/service/agent"
	"agent/models/service/calculator"
	"agent/models/service/nbnhhsh"
	"agent/models/service/weather"
	"agent/router"
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
	cfg.Server.Port = "25565"

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

	// 初始化 Ollama 服务
	ollamaSvc := agent.NewOllamaService(client, cfg.Ollama.Model, cfg.Ollama.Temperature)
	log.Info(ctx, "Ollama 服务初始化成功")

	// 初始化 MySQL
	mysql, err := dao.NewMySQL(cfg.Database)
	if err != nil {
		log.Error(ctx, "初始化 MySQL 失败: %v", err)
		// MySQL 失败不阻止服务启动，只是nbnhhsh功能不可用
	} else {
		log.Info(ctx, "MySQL 初始化成功")
	}

	// 初始化工具管理器
	toolManager := agent.NewToolManager()
	toolManager.Register(calculator.NewCalculator())
	toolManager.Register(weather.NewWeather(cfg.WeatherAPI))
	toolManager.Register(agent.NewFileTool(ollamaSvc, "./data"))

	// 注册 nbnhhsh 工具（如果 MySQL 可用）
	if mysql != nil {
		nbnhhshDAO := dao.NewNbnhhshDAO(mysql)
		toolManager.Register(nbnhhsh.NewCanYouSay(nbnhhshDAO))
		log.Info(ctx, "工具注册完成: calculator, weather, file, nbnhhsh")
	} else {
		log.Info(ctx, "工具注册完成: calculator, weather, file (nbnhhsh不可用)")
	}

	// 初始化 Agent 服务
	agentSvc := agent.NewAgentService(ollamaSvc, toolManager)
	log.Info(ctx, "Agent 服务初始化完成")

	// 初始化控制器
	chatCtrl := controllers.NewChatController(agentSvc)
	toolCtrl := controllers.NewToolController(toolManager)
	log.Info(ctx, "控制器初始化完成")

	// 初始化路由
	r := router.NewRouter(chatCtrl, toolCtrl)
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
