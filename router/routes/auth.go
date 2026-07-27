package routes

import (
	"agent/webapi/controllers"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(group *GroupRouter, authCtrl *controllers.AuthController, authMw Middleware) {
	// 公开接口（无需认证）
	pub := group.Group("/api/auth")
	pub.POST("/register", Wrap(authCtrl.Register))
	pub.POST("/login", Wrap(authCtrl.Login))
	pub.POST("/sms/send", Wrap(authCtrl.SendSMSCode))
	pub.POST("/register-by-phone", Wrap(authCtrl.RegisterByPhone))
	pub.POST("/password/reset", Wrap(authCtrl.ResetPassword))
	pub.POST("/email/send", Wrap(authCtrl.SendEmailCode))
	pub.POST("/register-by-email", Wrap(authCtrl.RegisterByEmail))
	pub.POST("/password/reset-by-email", Wrap(authCtrl.ResetPasswordByEmail))

	// 需要认证的接口
	authed := group.Group("/api/auth", authMw)
	authed.GET("/me", Wrap(authCtrl.GetMe))
}
