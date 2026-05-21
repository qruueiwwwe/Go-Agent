package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"agent/global"
	"agent/library/log"
	"agent/models/dao"
	"agent/models/service/agent"
	authService "agent/models/service/auth"
	"agent/models/service/blocker"
	"agent/models/service/calculator"
	"agent/models/service/nbnhhsh"
	"agent/models/service/persona"
	"agent/models/service/weather"
	"agent/router"
	"agent/webapi/controllers"

	"github.com/joho/godotenv"
	"github.com/ollama/ollama/api"
)

// 版本信息，通过 ldflags 注入
var (
	Version       = "dev"
	BuildTime     = "unknown"
	GitCommit     = "unknown"
	ZhipuAPIKey   = "" // 智谱清言API Key
	WeatherAPIID  = "" // 接口盒子天气API ID
	WeatherAPIKey = "" // 接口盒子天气API Key
	AmapAPIKey    = "" // 高德天气API Key
)

func main() {
	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		log.InfoOld(".env 文件不存在，使用系统环境变量或编译时配置")
	}

	// 显示版本信息
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Agent Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		os.Exit(0)
	}

	// 显示帮助信息
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("Agent - AI 智能助手")
		fmt.Println()
		fmt.Println("用法: agent [命令]")
		fmt.Println()
		fmt.Println("命令:")
		fmt.Println("  version  显示版本信息")
		fmt.Println("  -h, --help  显示帮助信息")
		fmt.Println()
		fmt.Println("环境变量:")
		fmt.Println("  ZHIPU_API_KEY     智谱清言API密钥（必填）")
		fmt.Println("  WEATHER_API_ID    接口盒子天气API ID")
		fmt.Println("  WEATHER_API_KEY   接口盒子天气API Key")
		fmt.Println("  AMAP_API_KEY      高德天气API Key")
		fmt.Println("  DATABASE_PASSWORD MySQL数据库密码")
		os.Exit(0)
	}

	// 创建带logid的上下文
	ctx := log.WithContext(context.Background())
	log.Info(ctx, "启动 Agent 服务...")
	log.Info(ctx, "Version: %s, Build Time: %s, Git Commit: %s", Version, BuildTime, GitCommit)

	// 加载配置
	cfg := global.DefaultConfig
	cfg.Server.Port = "25565"

	// 从环境变量或编译时注入的值覆盖配置
	// 天气API ID
	if weatherID := os.Getenv("WEATHER_API_ID"); weatherID != "" {
		cfg.WeatherAPI.ID = weatherID
	} else if WeatherAPIID != "" {
		cfg.WeatherAPI.ID = WeatherAPIID
	}
	// 天气API Key
	if weatherKey := os.Getenv("WEATHER_API_KEY"); weatherKey != "" {
		cfg.WeatherAPI.Key = weatherKey
	} else if WeatherAPIKey != "" {
		cfg.WeatherAPI.Key = WeatherAPIKey
	}
	// 高德天气API Key
	if amapKey := os.Getenv("AMAP_API_KEY"); amapKey != "" {
		cfg.WeatherAPI.AmapKey = amapKey
	} else if AmapAPIKey != "" {
		cfg.WeatherAPI.AmapKey = AmapAPIKey
	}
	// 数据库密码
	if dbPassword := os.Getenv("DATABASE_PASSWORD"); dbPassword != "" {
		cfg.Database.Password = dbPassword
	}
	// 屏蔽词配置
	if blockedWords := os.Getenv("BLOCKED_WORDS"); blockedWords != "" {
		cfg.Blocker.Words = strings.Split(blockedWords, ",")
	}

	// 初始化日志
	log.Init(cfg.Log)
	log.Info(ctx, "日志系统初始化完成")

	// 初始化 Ollama 客户端（可选）
	var client *api.Client
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Info(ctx, "Ollama 客户端创建失败: %v，将使用智谱清言服务", err)
	} else {
		log.Info(ctx, "Ollama 客户端已创建（实际连接在首次调用时验证）")
	}

	// 初始化 Ollama 服务（如果客户端可用）
	var ollamaSvc *agent.OllamaService
	if client != nil {
		ollamaSvc = agent.NewOllamaService(client, cfg.Ollama.Model, cfg.Ollama.Temperature)
		log.Info(ctx, "Ollama 服务已配置")
	} else {
		log.Info(ctx, "Ollama 服务未配置，将使用智谱清言服务")
	}

	// 初始化智谱清言后备服务
	var zhipuSvc *agent.ZhipuService
	zhipuAPIKey := os.Getenv("ZHIPU_API_KEY")

	// 如果环境变量没有设置，使用编译时注入的值
	if zhipuAPIKey == "" {
		zhipuAPIKey = ZhipuAPIKey
	}

	if zhipuAPIKey != "" && cfg.Zhipu.Enable {
		zhipuCfg := cfg.Zhipu
		zhipuCfg.APIKey = zhipuAPIKey
		zhipuSvc = agent.NewZhipuService(zhipuCfg)
		log.Info(ctx, "智谱清言后备服务初始化成功")
	} else {
		log.Warn(ctx, "智谱清言后备服务未启用（未配置ZHIPU_API_KEY）")
		if ollamaSvc == nil {
			log.Error(ctx, "警告：Ollama 和智谱清言都不可用，服务将无法处理对话请求！")
		}
	}

	// 初始化 MySQL
	mysql, err := dao.NewMySQL(cfg.Database)
	if err != nil {
		log.Error(ctx, "初始化 MySQL 失败: %v", err)
		// MySQL 失败不阻止服务启动，只是nbnhhsh功能不可用
	} else {
		log.Info(ctx, "MySQL 初始化成功")

		// 执行数据库迁移
		if err := mysql.AutoMigrate(ctx); err != nil {
			log.Error(ctx, "数据库迁移失败: %v", err)
		}

		// 执行角色卡表迁移
		if err := mysql.AutoMigratePersonaTables(ctx); err != nil {
			log.Error(ctx, "角色卡表迁移失败: %v", err)
		}
	}

	// 初始化认证服务
	var authCtrl *controllers.AuthController
	var authSvc *authService.AuthService
	var rateLimiter *authService.RateLimiter
	var adminCtrl *controllers.AdminController
	var userDAO *dao.UserDAO
	if mysql != nil {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default-secret-key-please-change-in-production"
			log.Warn(ctx, "JWT_SECRET 未设置，使用默认密钥（不安全！）")
		}
		userDAO = dao.NewUserDAO(mysql)
		smsCodeDAO := dao.NewSMSCodeDAO(mysql)
		smsSvc := authService.NewSMSService(smsCodeDAO)
		emailCodeDAO := dao.NewEmailCodeDAO(mysql)
		emailSvc := authService.NewEmailService(emailCodeDAO)
		authSvc = authService.NewAuthService(userDAO, smsCodeDAO, smsSvc, emailCodeDAO, emailSvc, jwtSecret, 24*time.Hour)
		authCtrl = controllers.NewAuthController(authSvc, smsSvc, emailSvc)

		// 初始化频率限制服务
		rateLimitDAO := dao.NewRateLimitDAO(mysql)
		rateLimiter = authService.NewRateLimiter(rateLimitDAO)

		// 初始化后台管理控制器
		adminCtrl = controllers.NewAdminController(userDAO)

		log.Info(ctx, "认证服务初始化完成")
	} else {
		log.Warn(ctx, "MySQL 不可用，认证服务未启用")
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
	contextBudget := os.Getenv("CONTEXT_BUDGET")
	agentSvc := agent.NewAgentService(ollamaSvc, zhipuSvc, toolManager, contextBudget)
	log.Info(ctx, "Agent 服务初始化完成，CONTEXT_BUDGET=%s", contextBudget)

	// 初始化控制器
	chatCtrl := controllers.NewChatController(agentSvc, rateLimiter)
	toolCtrl := controllers.NewToolController(toolManager)
	log.Info(ctx, "控制器初始化完成")

	// 初始化角色卡服务（需要 MySQL）
	var personaCtrl *controllers.PersonaController
	if mysql != nil {
		personaDAO := dao.NewPersonaDAO(mysql)
		personaExampleDAO := dao.NewPersonaExampleDAO(mysql)
		personaChatDAO := dao.NewPersonaChatDAO(mysql)

		// 初始化屏蔽词服务
		blockerSvc := blocker.NewService(cfg.Blocker.Words)
		log.Info(ctx, "屏蔽词服务初始化完成，屏蔽词数量: %d", len(cfg.Blocker.Words))

		personaSvc := persona.NewPersonaService(personaDAO, personaExampleDAO, blockerSvc)

		// 创建智谱适配器函数
		var zhipuChatFunc func(ctx context.Context, msgs []api.Message) (string, error)
		if zhipuSvc != nil {
			zhipuChatFunc = zhipuSvc.ChatWithAPIMessages
		}
		chatSvc := persona.NewChatService(personaDAO, personaChatDAO, personaExampleDAO, ollamaSvc, zhipuChatFunc)

		personaCtrl = controllers.NewPersonaController(personaSvc, chatSvc)
		log.Info(ctx, "角色卡服务初始化完成")
	} else {
		log.Warn(ctx, "MySQL 不可用，角色卡服务未启用")
	}

	// 初始化路由
	r := router.NewRouter(chatCtrl, toolCtrl)
	r.SetStaticFS(StaticFS()) // 使用嵌入的静态文件
	if authCtrl != nil && authSvc != nil {
		r.SetAuth(authCtrl, authSvc)
	}
	if adminCtrl != nil {
		r.SetAdmin(adminCtrl)
	}
	if personaCtrl != nil {
		r.SetPersona(personaCtrl)
	}
	mux := http.NewServeMux()
	r.RegisterRoutes(mux)

	// 优雅退出
	go func() {
		addr := "0.0.0.0:" + cfg.Server.Port // 监听所有网络接口
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			log.Error(ctx, "服务启动失败: %v", err)
		}
	}()

	log.Info(ctx, "服务启动成功，监听端口 %s", cfg.Server.Port)
	fmt.Println("=== Agent 服务已启动 ===")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("本地地址：http://localhost:" + cfg.Server.Port)
	// 获取本机IP
	if localIP := getLocalIP(); localIP != "" {
		fmt.Println("局域网地址：http://" + localIP + ":" + cfg.Server.Port)
		fmt.Println("API 地址：http://" + localIP + ":" + cfg.Server.Port + "/api/chat")
	}

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info(ctx, "收到退出信号，正在关闭服务...")
	os.Exit(0)
}

// getLocalIP 获取本机局域网IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
