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
}

// NewRouter 创建路由
func NewRouter(chatCtrl *controllers.ChatController) *Router {
	return &Router{
		chatCtrl:   chatCtrl,
		healthCtrl: controllers.NewHealthController(),
	}
}

// RegisterRoutes 注册路由
func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	// 静态文件
	mux.Handle("/", http.FileServer(http.Dir("./static")))

	// API 路由
	mux.HandleFunc("/api/chat", r.handleChat)
	mux.HandleFunc("/api/health", r.handleHealth)
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

// HandlerFunc 包装控制器方法为 http.HandlerFunc
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// Wrap 包装为 http.HandlerFunc
func Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fn(ctx, w, r)
	}
}
