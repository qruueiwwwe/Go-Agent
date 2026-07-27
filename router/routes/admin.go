package routes

import (
	"agent/webapi/controllers"
)

// RegisterAdminRoutes 注册管理员路由
func RegisterAdminRoutes(group *GroupRouter, adminCtrl *controllers.AdminController, authMw Middleware) {
	admin := group.Group("/api/admin", authMw, AdminMiddleware)
	admin.GET("/users", Wrap(adminCtrl.GetUsers))
	admin.POST("/users/{id}/role", Wrap(adminCtrl.UpdateUserRole))
	admin.POST("/users/{id}/status", Wrap(adminCtrl.UpdateUserStatus))
}
