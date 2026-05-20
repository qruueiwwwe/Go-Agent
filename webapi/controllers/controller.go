package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agent/library/log"
	"agent/models/service/agent"
	authService "agent/models/service/auth"

	"github.com/ollama/ollama/api"
)

// ========== 响应结构 ==========

// Response 统一响应结构（新格式）
type Response struct {
	Data   interface{} `json:"data,omitempty"`
	Errmsg string      `json:"errmsg,omitempty"`
	Errno  int         `json:"errno"`
	Logid  string      `json:"logid"`
}

// Reply 统一回复（所有请求都返回 200）
func Reply(w http.ResponseWriter, errno int, errmsg string, data interface{}, logid string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 统一返回 200
	json.NewEncoder(w).Encode(Response{
		Data:   data,
		Errmsg: errmsg,
		Errno:  errno,
		Logid:  logid,
	})
}

// ReplySuccess 成功回复
func ReplySuccess(w http.ResponseWriter, data interface{}, logid string) {
	Reply(w, 0, "", data, logid)
}

// ReplyError 错误回复
func ReplyError(w http.ResponseWriter, errno int, errmsg string, logid string) {
	Reply(w, errno, errmsg, nil, logid)
}

// ========== ChatController ==========

// ChatController 聊天控制器
type ChatController struct {
	agentSvc    *agent.AgentService
	rateLimiter *authService.RateLimiter
	messages    []api.Message
}

// NewChatController 创建聊天控制器
func NewChatController(agentSvc *agent.AgentService, rateLimiter *authService.RateLimiter) *ChatController {
	return &ChatController{
		agentSvc:    agentSvc,
		rateLimiter: rateLimiter,
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message string `json:"message" required:"true"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Result string `json:"result"`
}

// Chat 处理聊天请求
// @Summary 聊天接口
// @Tags chat
// @Accept json
// @Param message body ChatRequest true "消息内容"
// @Success 200 {object} Response
// @Router /api/chat [post]
func (c *ChatController) Chat(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// 校验参数
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, 404, "参数错误", log.GenerateLogID())
		return
	}

	// 获取用户信息用于生成 logid
	claims := GetUserFromContext(ctx)
	logid := log.GetLogID(ctx)

	// 添加 logid（不要覆盖原 ctx，用新的变量）
	logCtx := log.WithLogID(ctx, logid)
	start := time.Now()

	// 只支持 POST
	if r.Method != http.MethodPost {
		log.Warn(logCtx, "收到非POST请求: method=%s", r.Method)
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	// 频率限制检查（需要登录）- 使用原始 ctx 获取用户信息
	if c.rateLimiter != nil {
		if claims == nil {
			log.Warn(logCtx, "Chat: 用户未登录")
			ReplyError(w, 401, "请先登录", logid)
			return
		}

		allowed, waitSeconds, err := c.rateLimiter.CheckLimit(ctx, claims.UserID, claims.Role)
		if err != nil {
			log.Error(logCtx, "频率限制检查失败: %v", err)
		} else if !allowed {
			log.Warn(logCtx, "Chat: 频率限制 userID=%d role=%s waitSeconds=%d", claims.UserID, claims.Role, waitSeconds)
			ReplyError(w, 429, fmt.Sprintf("请求过于频繁，请等待 %d 秒后再试", waitSeconds), logid)
			return
		}
		// 记录请求
		if err := c.rateLimiter.RecordRequest(ctx, claims.UserID); err != nil {
			log.Error(logCtx, "记录请求失败: %v", err)
		}
		log.Info(logCtx, "Chat: 频率检查通过 userID=%d role=%s", claims.UserID, claims.Role)
	}

	// 解析请求
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(logCtx, "解析请求失败: %v", err)
		ReplyError(w, 404, "解析请求失败", logid)
		return
	}

	// 校验必填字段
	if req.Message == "" {
		log.Warn(logCtx, "收到空消息")
		ReplyError(w, 404, "消息不能为空", logid)
		return
	}

	log.Info(logCtx, "收到用户消息: %s", req.Message)

	// 调用服务
	result := c.agentSvc.Process(logCtx, req.Message, c.messages)

	// 保存到历史
	c.messages = append(c.messages, api.Message{Role: "user", Content: req.Message})
	c.messages = append(c.messages, api.Message{Role: "assistant", Content: result})

	log.Info(logCtx, "处理完成，耗时: %v", time.Since(start))

	// 返回结果
	ReplySuccess(w, ChatResponse{Result: result}, logid)
}

// ========== 其他控制器占位 ==========

// HealthController 健康检查控制器
type HealthController struct{}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Health 健康检查
func (h *HealthController) Health(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	logid := log.GetLogID(ctx)
	ReplySuccess(w, map[string]string{
		"status": "ok",
	}, logid)
}
