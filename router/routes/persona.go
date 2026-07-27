package routes

import (
	"agent/webapi/controllers"
)

// RegisterPersonaRoutes 注册角色卡路由
func RegisterPersonaRoutes(group *GroupRouter, personaCtrl *controllers.PersonaController, authMw Middleware) {
	persona := group.Group("/api/persona", authMw, VIPMiddleware)
	persona.GET("/list", Wrap(personaCtrl.PersonaList))
	persona.GET("/detail", Wrap(personaCtrl.PersonaDetail))
	persona.POST("/create", Wrap(personaCtrl.PersonaCreate))
	persona.POST("/update", Wrap(personaCtrl.PersonaUpdate))
	persona.POST("/delete", Wrap(personaCtrl.PersonaDelete))
	persona.POST("/chat", Wrap(personaCtrl.PersonaChat))
	persona.GET("/sessions", Wrap(personaCtrl.PersonaSessions))
	persona.GET("/history", Wrap(personaCtrl.PersonaHistory))
}
