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

	// 认证接口
	if r.authCtrl != nil {
		// 公开接口（无需认证）
		mux.HandleFunc("/api/auth/register", Wrap(r.authCtrl.Register))
		mux.HandleFunc("/api/auth/login", Wrap(r.authCtrl.Login))
		mux.HandleFunc("/api/auth/sms/send", Wrap(r.authCtrl.SendSMSCode))
		mux.HandleFunc("/api/auth/register-by-phone", Wrap(r.authCtrl.RegisterByPhone))
		mux.HandleFunc("/api/auth/password/reset", Wrap(r.authCtrl.ResetPassword))
		mux.HandleFunc("/api/auth/email/send", Wrap(r.authCtrl.SendEmailCode))
		mux.HandleFunc("/api/auth/register-by-email", Wrap(r.authCtrl.RegisterByEmail))
		mux.HandleFunc("/api/auth/password/reset-by-email", Wrap(r.authCtrl.ResetPasswordByEmail))

		// 需要认证的接口
		if r.authMiddleware != nil {
			mux.HandleFunc("/api/auth/me", r.authMiddleware(Wrap(r.authCtrl.GetMe)))
		}
	}
}
