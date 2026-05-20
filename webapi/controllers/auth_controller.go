package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"agent/library/log"
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
	logid := log.GenerateLogID()
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}

	if req.Username == "" {
		Reply(w, 404, "用户名不能为空", nil, logid)
		return
	}

	if req.Password == "" {
		Reply(w, 404, "密码不能为空", nil, logid)
		return
	}

	// 如果昵称为空，使用用户名作为昵称
	if req.Nickname == "" {
		req.Nickname = req.Username
	}

	user, err := c.authSvc.Register(ctx, req.Username, req.Password, req.Nickname)
	if err != nil {
		if err == auth.ErrUserExists {
			Reply(w, 404, err.Error(), nil, logid)
			return
		}
		if err == auth.ErrPasswordTooShort {
			Reply(w, 404, err.Error(), nil, logid)
			return
		}
		Reply(w, 500, "服务器错误", nil, logid)
		return
	}

	Reply(w, 0, "", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
	}, logid)
}

// Login 用户登录
func (c *AuthController) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}

	if req.Username == "" || req.Password == "" {
		Reply(w, 404, "用户名和密码不能为空", nil, logid)
		return
	}

	token, user, err := c.authSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials || err == auth.ErrUserDisabled || err == auth.ErrUserDeleted {
			Reply(w, 401, err.Error(), nil, logid)
			return
		}
		Reply(w, 500, "服务器错误", nil, logid)
		return
	}

	Reply(w, 0, "", map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"role":     user.Role,
		},
	}, logid)
}

// GetMe 获取当前用户信息
func (c *AuthController) GetMe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GenerateLogIDWithUser(claims)

	if claims == nil {
		Reply(w, 401, "未登录", nil, logid)
		return
	}

	Reply(w, 0, "", map[string]interface{}{
		"id":       claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	}, logid)
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) *auth.Claims {
	if claims, ok := ctx.Value(UserKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// SendSMSCode 发送验证码
func (c *AuthController) SendSMSCode(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Phone string `json:"phone"`
		Scene string `json:"scene"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	if err := c.smsSvc.SendCode(ctx, req.Phone, req.Scene); err != nil {
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	Reply(w, 0, "", nil, logid)
}

// RegisterByPhone 手机号验证码注册
func (c *AuthController) RegisterByPhone(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Phone    string `json:"phone"`
		Code     string `json:"code"`
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	if req.Username == "" || req.Password == "" || req.Phone == "" || req.Code == "" {
		Reply(w, 404, "缺少必要参数", nil, logid)
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	user, err := c.authSvc.RegisterByPhoneCode(ctx, req.Phone, req.Code, req.Username, req.Password, req.Nickname)
	if err != nil {
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	Reply(w, 0, "", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"phone":    user.Phone,
	}, logid)
}

// ResetPassword 重置密码
func (c *AuthController) ResetPassword(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Phone       string `json:"phone"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	if req.Phone == "" || req.Code == "" || req.NewPassword == "" {
		Reply(w, 404, "缺少必要参数", nil, logid)
		return
	}
	if err := c.authSvc.ResetPasswordByPhoneCode(ctx, req.Phone, req.Code, req.NewPassword); err != nil {
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	Reply(w, 0, "", nil, logid)
}

// SendEmailCode 发送邮箱验证码接口。
// 场景 scene 支持：register / reset_password。
func (c *AuthController) SendEmailCode(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Email string `json:"email"`
		Scene string `json:"scene"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	if err := c.emailSvc.SendCode(ctx, req.Email, req.Scene); err != nil {
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	Reply(w, 0, "", nil, logid)
}

// RegisterByEmail 邮箱验证码注册接口。
// phone 为可选参数，不影响邮箱主流程。
func (c *AuthController) RegisterByEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Email    string `json:"email"`
		Code     string `json:"code"`
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx, "RegisterByEmail: 参数解析失败 err=%v", err)
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	log.Info(ctx, "RegisterByEmail: 收到注册请求 email=%s code=%s username=%s passwordLen=%s nickname=%s phone=%s",
		req.Email, req.Code, req.Username, req.Password, req.Nickname, req.Phone)
	if req.Email == "" || req.Code == "" || req.Username == "" || req.Password == "" {
		Reply(w, 404, "缺少必要参数", nil, logid)
		return
	}
	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	user, err := c.authSvc.RegisterByEmailCode(ctx, req.Email, req.Code, req.Username, req.Password, req.Nickname, req.Phone)
	if err != nil {
		log.Error(ctx, "RegisterByEmail: 注册失败 email=%s err=%v", req.Email, err)
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	log.Info(ctx, "RegisterByEmail: 注册成功 id=%d username=%s", user.ID, user.Username)
	Reply(w, 0, "", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"nickname": user.Nickname,
		"phone":    user.Phone,
		"email":    user.Email,
	}, logid)
}

// ResetPasswordByEmail 邮箱验证码重置密码接口。
func (c *AuthController) ResetPasswordByEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GenerateLogID()
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Reply(w, 404, "参数错误", nil, logid)
		return
	}
	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		Reply(w, 404, "缺少必要参数", nil, logid)
		return
	}
	if err := c.authSvc.ResetPasswordByEmailCode(ctx, req.Email, req.Code, req.NewPassword); err != nil {
		Reply(w, 404, err.Error(), nil, logid)
		return
	}
	Reply(w, 0, "", nil, logid)
}
