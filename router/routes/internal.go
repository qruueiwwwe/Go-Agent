package routes

import (
	"agent/webapi/controllers"
)

// RegisterInternalRoutes 注册内部工具路由（无需认证）
func RegisterInternalRoutes(group *GroupRouter, toolCtrl *controllers.ToolController) {
	internal := group.Group("/internal/tools")
	internal.GET("/list", Wrap(toolCtrl.ListTools))
	internal.POST("/weather", Wrap(toolCtrl.Weather))
	internal.POST("/calculator", Wrap(toolCtrl.Calculator))
	internal.POST("/file", Wrap(toolCtrl.File))
	internal.POST("/nbnhhsh", Wrap(toolCtrl.Nbnhhsh))
	internal.POST("/execute/{toolName}", Wrap(toolCtrl.ExecuteTool))
}
