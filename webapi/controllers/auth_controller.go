package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"agent/models/service/auth"
)

// UserKey 用户信息上下文键
const UserKey contextKey = "user"

type contextKey string

// AuthController 认证控制器
type AuthController struct {
	authSvc  *auth.AuthService
	smsSvc   *auth.SMSService
	emailSvc *auth.EmailService
}

// NewAuthController 创建认证控制器
func NewAuthController(authSvc *auth.AuthService, smsSvc *auth.SMSService, emailSvc *auth.EmailService) *AuthController {
	return &AuthController{authSvc: authSvc, smsSvc: smsSvc, emailSvc: emailSvc}
}

// Register 用户注册
func (c *AuthController) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}

	if req.Username == "" {
		respondJSON(w, http.StatusBadRequest, "用户名不能为空", nil)
		return
	}

	if req.Password == "" {
		respondJSON(w, http.StatusBadRequest, "密码不能为空", nil)
		return
	}

	// 如果昵称为空，使用用户名作为昵称
	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	user, err := c.authSvc.Register(ctx, req.Username, req.Password, req.Nickname)
	if err != nil {
		if err == auth.ErrUserExists {
			respondJSON(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
		if err == auth.ErrPasswordTooShort {
			respondJSON(w, http.StatusBadRequest, err.Error(), nil)
			return
		}
		respondJSON(w, http.StatusInternalServerError, "服务器错误", nil)
		return
	}

	respondJSON(w, http.StatusOK, "注册成功", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	})
}

// Login 用户登录
func (c *AuthController) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}

	if req.Username == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, "用户名和密码不能为空", nil)
		return
	}

	token, user, err := c.authSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials || err == auth.ErrUserDisabled {
			respondJSON(w, http.StatusUnauthorized, err.Error(), nil)
			return
		}
		respondJSON(w, http.StatusInternalServerError, "服务器错误", nil)
		return
	}

	respondJSON(w, http.StatusOK, "登录成功", map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"role":     user.Role,
		},
	})
}

// GetMe 获取当前用户信息
func (c *AuthController) GetMe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// 从 context 获取用户信息（由中间件注入）
	claims := GetUserFromContext(ctx)
	if claims == nil {
		respondJSON(w, http.StatusUnauthorized, "未登录", nil)
		return
	}

	respondJSON(w, http.StatusOK, "success", map[string]interface{}{
		"id":       claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// respondJSON 返回 JSON 响应
// SendSMSCode 发送验证码
func (c *AuthController) SendSMSCode(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
		Scene string `json:"scene"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if err := c.smsSvc.SendCode(ctx, req.Phone, req.Scene); err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "验证码发送成功", nil)
}

// RegisterByPhone 手机号验证码注册
func (c *AuthController) RegisterByPhone(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if req.Username == "" || req.Password == "" || req.Phone == "" || req.Code == "" {
		respondJSON(w, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	user, err := c.authSvc.RegisterByPhoneCode(ctx, req.Phone, req.Code, req.Username, req.Password, req.Nickname)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "注册成功", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"phone":    user.Phone,
	})
}

// ResetPassword 重置密码
func (c *AuthController) ResetPassword(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone       string `json:"phone"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if req.Phone == "" || req.Code == "" || req.NewPassword == "" {
		respondJSON(w, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}
	if err := c.authSvc.ResetPasswordByPhoneCode(ctx, req.Phone, req.Code, req.NewPassword); err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "密码重置成功", nil)
}

// SendEmailCode 发送邮箱验证码接口。
// 场景 scene 支持：register / reset_password。
func (c *AuthController) SendEmailCode(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Scene string `json:"scene"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if err := c.emailSvc.SendCode(ctx, req.Email, req.Scene); err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "验证码发送成功", nil)
}

// RegisterByEmail 邮箱验证码注册接口。
// phone 为可选参数，不影响邮箱主流程。
func (c *AuthController) RegisterByEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if req.Email == "" || req.Code == "" || req.Username == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	user, err := c.authSvc.RegisterByEmailCode(ctx, req.Email, req.Code, req.Username, req.Password, req.Nickname, req.Phone)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "注册成功", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"phone":    user.Phone,
		"email":    user.Email,
	})
}

// ResetPasswordByEmail 邮箱验证码重置密码接口。
func (c *AuthController) ResetPasswordByEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, "参数错误", nil)
		return
	}
	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		respondJSON(w, http.StatusBadRequest, "缺少必要参数", nil)
		return
	}
	if err := c.authSvc.ResetPasswordByEmailCode(ctx, req.Email, req.Code, req.NewPassword); err != nil {
		respondJSON(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	respondJSON(w, http.StatusOK, "密码重置成功", nil)
}

func respondJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	})
}
