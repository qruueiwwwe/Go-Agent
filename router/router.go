package router

import (
	"io/fs"
	"net/http"

	"agent/models/service/auth"
	"agent/router/routes"
	"agent/webapi/controllers"
)

// RoutesDeps 路由注册所需的依赖
type RoutesDeps struct {
	ChatCtrl    *controllers.ChatController
	ToolCtrl    *controllers.ToolController
	AuthCtrl    *controllers.AuthController
	AdminCtrl   *controllers.AdminController
	PersonaCtrl *controllers.PersonaController
	HealthCtrl  *controllers.HealthController
	FileCtrl    *controllers.FileUploadController
	StaticFS    fs.FS
}

// Router 路由管理器
type Router struct {
	authMw routes.Middleware
}

// NewRouter 创建路由管理器
func NewRouter() *Router {
	return &Router{}
}

// SetAuth 设置认证中间件
func (r *Router) SetAuth(authSvc *auth.AuthService) {
	r.authMw = routes.NewAuthMiddleware(authSvc)
}

// RegisterRoutes 注册所有路由到 mux
func (r *Router) RegisterRoutes(mux *http.ServeMux, deps RoutesDeps) {
	// 静态文件
	if deps.StaticFS != nil {
		mux.Handle("/", http.FileServer(http.FS(deps.StaticFS)))
	} else {
		mux.Handle("/", http.FileServer(http.Dir("./static")))
	}

	group := routes.NewGroupRouter()

	// 通用 API 路由（聊天、健康检查、文件）
	if deps.ChatCtrl != nil {
		healthCtrl := deps.HealthCtrl
		if healthCtrl == nil {
			healthCtrl = controllers.NewHealthController()
		}
		fileCtrl := deps.FileCtrl
		if fileCtrl == nil {
			fileCtrl = controllers.NewFileUploadController()
		}
		routes.RegisterChatRoutes(group, deps.ChatCtrl, healthCtrl, fileCtrl, r.authMw)
	}

	// 认证路由
	if deps.AuthCtrl != nil && r.authMw != nil {
		routes.RegisterAuthRoutes(group, deps.AuthCtrl, r.authMw)
	}

	// 管理员路由
	if deps.AdminCtrl != nil && r.authMw != nil {
		routes.RegisterAdminRoutes(group, deps.AdminCtrl, r.authMw)
	}

	// 角色卡路由
	if deps.PersonaCtrl != nil && r.authMw != nil {
		routes.RegisterPersonaRoutes(group, deps.PersonaCtrl, r.authMw)
	}

	// 内部工具路由
	if deps.ToolCtrl != nil {
		routes.RegisterInternalRoutes(group, deps.ToolCtrl)
	}

	// 将所有路由注册到 mux
	group.RegisterTo(mux)
}
