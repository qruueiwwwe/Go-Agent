package routes

import (
	"context"
	"net/http"
)

// Middleware 中间件函数类型
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Route 路由条目
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}

// GroupRouter 路由分组，支持前缀和中间件链
type GroupRouter struct {
	prefix      string
	middlewares []Middleware
	routes      []Route
	children    []*GroupRouter
}

// NewGroupRouter 创建根路由组
func NewGroupRouter() *GroupRouter {
	return &GroupRouter{}
}

// Group 创建子分组，继承父分组中间件
func (g *GroupRouter) Group(prefix string, middlewares ...Middleware) *GroupRouter {
	child := &GroupRouter{
		prefix:      g.prefix + prefix,
		middlewares: append(g.cloneMiddlewares(), middlewares...),
	}
	g.children = append(g.children, child)
	return child
}

// Use 注册中间件（对该分组后续注册的路由生效）
func (g *GroupRouter) Use(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// GET 注册 GET 路由
func (g *GroupRouter) GET(path string, handler http.HandlerFunc) {
	g.Handle(http.MethodGet, path, handler)
}

// POST 注册 POST 路由
func (g *GroupRouter) POST(path string, handler http.HandlerFunc) {
	g.Handle(http.MethodPost, path, handler)
}

// ANY 注册所有方法路由
func (g *GroupRouter) ANY(path string, handler http.HandlerFunc) {
	fullPath := g.prefix + path
	h := g.applyMiddlewares(handler)
	g.routes = append(g.routes, Route{Method: "", Path: fullPath, Handler: h})
}

// Handle 注册指定方法的路由
func (g *GroupRouter) Handle(method string, path string, handler http.HandlerFunc) {
	fullPath := g.prefix + path
	h := g.applyMiddlewares(handler)
	g.routes = append(g.routes, Route{Method: method, Path: fullPath, Handler: h})
}

// RegisterTo 将所有路由注册到 http.ServeMux（Go 1.22+ 支持 "METHOD /path"）
func (g *GroupRouter) RegisterTo(mux *http.ServeMux) {
	for _, route := range g.routes {
		if route.Method == "" {
			mux.HandleFunc(route.Path, route.Handler)
		} else {
			pattern := route.Method + " " + route.Path
			mux.HandleFunc(pattern, route.Handler)
		}
	}
	for _, child := range g.children {
		child.RegisterTo(mux)
	}
}

// applyMiddlewares 按注册顺序包装中间件
func (g *GroupRouter) applyMiddlewares(h http.HandlerFunc) http.HandlerFunc {
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		h = g.middlewares[i](h)
	}
	return h
}

// cloneMiddlewares 复制中间件切片
func (g *GroupRouter) cloneMiddlewares() []Middleware {
	if len(g.middlewares) == 0 {
		return nil
	}
	ms := make([]Middleware, len(g.middlewares))
	copy(ms, g.middlewares)
	return ms
}

// HandlerFunc 控制器方法签名
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// Wrap 将 HandlerFunc 包装为 http.HandlerFunc
func Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		fn(ctx, w, r)
	}
}
