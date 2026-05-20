package router

import (
	"context"
	"io/fs"
	"net/http"

	"agent/models/service/auth"
	"agent/webapi/controllers"
)

// Router 路由结构
type Router struct {
	chatCtrl       *controllers.ChatController
	healthCtrl     *controllers.HealthController
	fileCtrl       *controllers.FileUploadController
	internal       *InternalRouter
	staticFS       fs.FS // 静态文件系统（可选，用于嵌入模式）
	authCtrl       *controllers.AuthController
	adminCtrl      *controllers.AdminController
	authMiddleware func(http.HandlerFunc) http.HandlerFunc
}

// NewRouter 创建路由
func NewRouter(chatCtrl *controllers.ChatController, toolCtrl *controllers.ToolController) *Router {
	return &Router{
		chatCtrl:   chatCtrl,
		healthCtrl: controllers.NewHealthController(),
		fileCtrl:   controllers.NewFileUploadController(),
		internal:   NewInternalRouter(toolCtrl),
	}
}

// SetAuth 设置认证控制器和中间件
func (r *Router) SetAuth(authCtrl *controllers.AuthController, authSvc *auth.AuthService) {
	r.authCtrl = authCtrl
	r.authMiddleware = func(next http.HandlerFunc) http.HandlerFunc {
		return AuthMiddleware(authSvc, next)
	}
}

// SetAdmin 设置后台管理控制器
func (r *Router) SetAdmin(adminCtrl *controllers.AdminController) {
	r.adminCtrl = adminCtrl
}

// SetStaticFS 设置静态文件系统（用于嵌入模式）
func (r *Router) SetStaticFS(fsys fs.FS) {
	r.staticFS = fsys
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	// 静态文件
	if r.staticFS != nil {
		// 使用嵌入的静态文件
		mux.Handle("/", http.FileServer(http.FS(r.staticFS)))
	} else {
		// 使用外部静态文件（开发模式）
		mux.Handle("/", http.FileServer(http.Dir("./static")))
	}

	// 前端 API（/api/*）
	RegisterAPIRoutes(mux, r)

	// 后端内部接口（/internal/*）
	RegisterInternalRoutes(mux, r.internal)
}

// handleChat 处理聊天请求
func (r *Router) handleChat(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context() // 使用请求的 context，包含中间件存入的用户信息
	r.chatCtrl.Chat(ctx, w, rq)
}

// handleHealth 处理健康检查
func (r *Router) handleHealth(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context()
	r.healthCtrl.Health(ctx, w, rq)
}

// handleUpload 处理文件上传
func (r *Router) handleUpload(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context()
	r.fileCtrl.Upload(ctx, w, rq)
}

// handleListFiles 处理获取文件列表
func (r *Router) handleListFiles(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context()
	r.fileCtrl.ListFiles(ctx, w, rq)
}

// handleDeleteFile 处理删除文件
func (r *Router) handleDeleteFile(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context()
	r.fileCtrl.DeleteFile(ctx, w, rq)
}

// HandlerFunc 包装控制器方法为 http.HandlerFunc
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// Wrap 包装为 http.HandlerFunc
func Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fn(ctx, w, r)
	}
}
