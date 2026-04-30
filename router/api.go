package router

import (
	"net/http"
)

// RegisterAPIRoutes 注册前端 API 路由（/api/*）
// 后续可添加校验中间件（token验证、频率限制等）
func RegisterAPIRoutes(mux *http.ServeMux, r *Router) {
	// 聊天接口
	mux.HandleFunc("/api/chat", r.handleChat)
	
	// 健康检查
	mux.HandleFunc("/api/health", r.handleHealth)
	
	// 文件管理
	mux.HandleFunc("/api/upload", r.handleUpload)
	mux.HandleFunc("/api/files", r.handleListFiles)
	mux.HandleFunc("/api/file/delete", r.handleDeleteFile)
}