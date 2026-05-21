package router

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"agent/library/log"
	"agent/models/service/auth"
	"agent/webapi/controllers"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(authSvc *auth.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logid := log.GenerateLogID()

		// 从 Header 获取 token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondAuthError(w, 401, "未登录", logid)
			return
		}

		// Bearer token 格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			respondAuthError(w, 401, "无效的认证格式", logid)
			return
		}

		// 验证 token
		claims, err := authSvc.ParseToken(tokenString)
		if err != nil {
			respondAuthError(w, 401, "登录已过期", logid)
			return
		}

		// 校验用户是否仍然存在且未删除/禁用
		user, err := authSvc.ValidateToken(r.Context(), tokenString)
		if err != nil || user == nil || user.Status != 1 {
			respondAuthError(w, 401, "用户未注册，需注册后才可体验", logid)
			return
		}

		// 将用户信息与logid存入 context
		ctx := context.WithValue(r.Context(), controllers.UserKey, claims)
		ctx = log.WithLogID(ctx, logid)
		next(w, r.WithContext(ctx))
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := controllers.GetUserFromContext(r.Context())
		logid := log.GenerateLogIDWithUser(claims)

		if claims == nil {
			respondAuthError(w, 401, "未登录", logid)
			return
		}

		if claims.Role != "admin" {
			respondAuthError(w, 403, "无权限访问", logid)
			return
		}

		ctx := log.WithLogID(r.Context(), logid)
		next(w, r.WithContext(ctx))
	}
}

// VIPMiddleware VIP或管理员权限中间件
func VIPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := controllers.GetUserFromContext(r.Context())
		logid := log.GenerateLogIDWithUser(claims)

		if claims == nil {
			respondAuthError(w, 401, "未登录", logid)
			return
		}

		if claims.Role != "vip" && claims.Role != "admin" {
			respondAuthError(w, 403, "该功能仅限VIP用户使用，请升级VIP", logid)
			return
		}

		ctx := log.WithLogID(r.Context(), logid)
		next(w, r.WithContext(ctx))
	}
}

// respondAuthError 返回认证错误响应
func respondAuthError(w http.ResponseWriter, errno int, errmsg string, logid string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK) // 统一返回 200
	json.NewEncoder(w).Encode(controllers.Response{
		Data:   nil,
		Errmsg: errmsg,
		Errno:  errno,
		Logid:  logid,
	})
}
