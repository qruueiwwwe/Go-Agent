package router

import (
	"context"
	"net/http"

	"agent/webapi/controllers"
)

// InternalRouter 内部路由处理器
type InternalRouter struct {
	toolCtrl *controllers.ToolController
}

// NewInternalRouter 创建内部路由
func NewInternalRouter(tc *controllers.ToolController) *InternalRouter {
	return &InternalRouter{
		toolCtrl: tc,
	}
}

// RegisterInternalRoutes 注册后端内部路由（/internal/*）
// 这些接口面向内部服务，无需校验
func RegisterInternalRoutes(mux *http.ServeMux, ir *InternalRouter) {
	// 工具列表
	mux.HandleFunc("/internal/tools/list", ir.handleListTools)

	// 单独工具接口
	mux.HandleFunc("/internal/tools/weather", ir.handleWeather)
	mux.HandleFunc("/internal/tools/calculator", ir.handleCalculator)
	mux.HandleFunc("/internal/tools/file", ir.handleFile)
	mux.HandleFunc("/internal/tools/nbnhhsh", ir.handleNbnhhsh)

	// 统一工具执行接口（动态工具名）
	mux.HandleFunc("/internal/tools/execute/", ir.handleExecuteTool)
}

func (ir *InternalRouter) handleListTools(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.ListTools(ctx, w, r)
}

func (ir *InternalRouter) handleWeather(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.Weather(ctx, w, r)
}

func (ir *InternalRouter) handleCalculator(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.Calculator(ctx, w, r)
}

func (ir *InternalRouter) handleFile(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.File(ctx, w, r)
}

func (ir *InternalRouter) handleNbnhhsh(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.Nbnhhsh(ctx, w, r)
}

func (ir *InternalRouter) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	ir.toolCtrl.ExecuteTool(ctx, w, r)
}
