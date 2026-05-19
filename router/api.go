package router

import (
	"net/http"

	"agent/library/log"
	"agent/webapi/controllers"
)

// RegisterAPIRoutes 注册前端 API 路由（/api/*）
// 后续可添加校验中间件（token验证、频率限制等）
func RegisterAPIRoutes(mux *http.ServeMux, r *Router) {
	// 聊天接口（需要认证）
	if r.authMiddleware != nil {
		mux.HandleFunc("/api/chat", r.authMiddleware(r.handleChat))
	} else {
		mux.HandleFunc("/api/chat", r.handleChat)
	}

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

	// 后台管理接口（需要管理员权限）
	if r.adminCtrl != nil && r.authMiddleware != nil {
		// 获取用户列表
		mux.HandleFunc("/api/admin/users", r.authMiddleware(AdminMiddleware(Wrap(r.adminCtrl.GetUsers))))
		// 更新用户角色
		mux.HandleFunc("/api/admin/users/", r.authMiddleware(AdminMiddleware(r.handleAdminUserUpdate)))
	}
}

// handleAdminUserUpdate 处理管理员用户更新请求（角色/状态）
func (r *Router) handleAdminUserUpdate(w http.ResponseWriter, rq *http.Request) {
	ctx := rq.Context()
	path := rq.URL.Path
	logid := log.GenerateLogIDWithUser(controllers.GetUserFromContext(ctx))

	// 根据 URL 后缀判断是更新角色还是状态
	if len(path) > len("/api/admin/users/") {
		suffix := path[len("/api/admin/users/"):]
		// 提取用户ID
		// 格式: /api/admin/users/{id}/role 或 /api/admin/users/{id}/status
		// suffix 格式为 "{id}/role" 或 "{id}/status"

		// 根据 method 和路径调用对应处理函数
		if rq.Method == http.MethodPost {
			// role: 最短 "{id}/role" = "1/role" = 6 字符
			if len(suffix) >= 6 && suffix[len(suffix)-4:] == "role" {
				r.adminCtrl.UpdateUserRole(ctx, w, rq)
				return
			}
			// status: 最短 "{id}/status" = "1/status" = 8 字符
			if len(suffix) >= 8 && suffix[len(suffix)-6:] == "status" {
				r.adminCtrl.UpdateUserStatus(ctx, w, rq)
				return
			}
		}
	}

	controllers.Reply(w, 404, "接口不存在", nil, logid)
}
