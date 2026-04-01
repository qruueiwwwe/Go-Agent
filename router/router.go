package router

import (
	"context"
	"net/http"

	"agent/webapi/controllers"
)

// Router 路由结构
type Router struct {
	chatCtrl   *controllers.ChatController
	healthCtrl *controllers.HealthController
	fileCtrl   *controllers.FileUploadController
}

// NewRouter 创建路由
func NewRouter(chatCtrl *controllers.ChatController) *Router {
	return &Router{
		chatCtrl:   chatCtrl,
		healthCtrl: controllers.NewHealthController(),
		fileCtrl:   controllers.NewFileUploadController(),
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	// 静态文件
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// API 路由
	mux.HandleFunc("/api/chat", r.handleChat)
	mux.HandleFunc("/api/health", r.handleHealth)
	
	// 文件上传相关路由
	mux.HandleFunc("/api/upload", r.handleUpload)
	mux.HandleFunc("/api/files", r.handleListFiles)
	mux.HandleFunc("/api/file/delete", r.handleDeleteFile)
}

// handleChat 处理聊天请求
func (r *Router) handleChat(w http.ResponseWriter, rq *http.Request) {
	ctx := context.Background()
	r.chatCtrl.Chat(ctx, w, rq)
}

// handleHealth 处理健康检查
func (r *Router) handleHealth(w http.ResponseWriter, rq *http.Request) {
	ctx := context.Background()
	r.healthCtrl.Health(ctx, w, rq)
}

// handleUpload 处理文件上传
func (r *Router) handleUpload(w http.ResponseWriter, rq *http.Request) {
	ctx := context.Background()
	r.fileCtrl.Upload(ctx, w, rq)
}

// handleListFiles 处理获取文件列表
func (r *Router) handleListFiles(w http.ResponseWriter, rq *http.Request) {
	ctx := context.Background()
	r.fileCtrl.ListFiles(ctx, w, rq)
}

// handleDeleteFile 处理删除文件
func (r *Router) handleDeleteFile(w http.ResponseWriter, rq *http.Request) {
	ctx := context.Background()
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
