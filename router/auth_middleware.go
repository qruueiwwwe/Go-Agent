package router

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"agent/models/service/auth"
	"agent/webapi/controllers"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(authSvc *auth.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 Header 获取 token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondAuthError(w, http.StatusUnauthorized, "未登录")
			return
		}

		// Bearer token 格式
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			respondAuthError(w, http.StatusUnauthorized, "无效的认证格式")
			return
		}

		// 验证 token
		claims, err := authSvc.ParseToken(tokenString)
		if err != nil {
			respondAuthError(w, http.StatusUnauthorized, "登录已过期")
			return
		}

		// 将用户信息存入 context
		ctx := context.WithValue(r.Context(), controllers.UserKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// respondAuthError 返回认证错误响应
func respondAuthError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
