package routes

import (
	"net/http"

	"agent/webapi/controllers"
)

// RegisterChatRoutes 注册聊天、健康检查、文件管理路由
func RegisterChatRoutes(group *GroupRouter, chatCtrl *controllers.ChatController, healthCtrl *controllers.HealthController, fileCtrl *controllers.FileUploadController, authMw Middleware) {
	// 健康检查（公开）
	group.GET("/api/health", Wrap(healthCtrl.Health))

	// 聊天接口（需认证）
	chat := group.Group("/api", authMw)
	chat.POST("/chat", func(w http.ResponseWriter, r *http.Request) {
		chatCtrl.Chat(r.Context(), w, r)
	})
	// 流式深度思考/Auto 模式（认证 + 模式权限校验）
	chat.POST("/chat/stream", StreamModeMiddleware(func(w http.ResponseWriter, r *http.Request) {
		chatCtrl.ChatStream(r.Context(), w, r)
	}))
	// 历史会话
	chat.POST("/getchatlist", func(w http.ResponseWriter, r *http.Request) {
		chatCtrl.GetChatList(r.Context(), w, r)
	})
	chat.POST("/getchathistory", func(w http.ResponseWriter, r *http.Request) {
		chatCtrl.GetChatHistory(r.Context(), w, r)
	})

	// 文件管理（公开，保持原行为）
	group.POST("/api/upload", Wrap(fileCtrl.Upload))
	group.GET("/api/files", Wrap(fileCtrl.ListFiles))
	group.POST("/api/file/delete", Wrap(fileCtrl.DeleteFile))
}
