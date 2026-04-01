package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"agent/global"
	"agent/library/log"
	"agent/models/service/agent"

	"github.com/ollama/ollama/api"
)

// ========== 响应结构 ==========

// ReplyResponse 统一响应结构
type ReplyResponse struct {
	Code    global.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`
}

// Reply 统一回复
func Reply(w http.ResponseWriter, code global.ErrorCode, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReplyResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// ReplySuccess 成功回复
func ReplySuccess(w http.ResponseWriter, data interface{}) {
	Reply(w, global.ErrCodeSuccess, global.GetErrorMsg(global.ErrCodeSuccess), data)
}

// ReplyError 错误回复
func ReplyError(w http.ResponseWriter, code global.ErrorCode, message string) {
	Reply(w, code, message, nil)
}

// ========== ChatController ==========

// ChatController 聊天控制器
type ChatController struct {
	agentSvc *agent.AgentService
	messages []api.Message
}

// NewChatController 创建聊天控制器
func NewChatController(agentSvc *agent.AgentService) *ChatController {
	return &ChatController{
		agentSvc: agentSvc,
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
// @Success 200 {object} ReplyResponse
// @Router /api/chat [post]
func (c *ChatController) Chat(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// 校验参数
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, global.ErrCodeInvalidParams, global.GetErrorMsg(global.ErrCodeInvalidParams))
		return
	}

	// 添加 logid
	ctx = log.WithContext(ctx)
	start := time.Now()

	// 只支持 POST
	if r.Method != http.MethodPost {
		log.Warn(ctx, "收到非POST请求: method=%s", r.Method)
		ReplyError(w, global.ErrCodeInvalidParams, "只支持 POST 方法")
		return
	}

	// 解析请求
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(ctx, "解析请求失败: %v", err)
		ReplyError(w, global.ErrCodeInvalidParams, "解析请求失败")
		return
	}

	// 校验必填字段
	if req.Message == "" {
		log.Warn(ctx, "收到空消息")
		ReplyError(w, global.ErrCodeInvalidParams, "消息不能为空")
		return
	}

	log.Info(ctx, "收到用户消息: %s", req.Message)

	// 调用服务
	result := c.agentSvc.Process(ctx, req.Message, c.messages)

	// 保存到历史
	c.messages = append(c.messages, api.Message{Role: "user", Content: req.Message})
	c.messages = append(c.messages, api.Message{Role: "assistant", Content: result})

	log.Info(ctx, "处理完成，耗时: %v", time.Since(start))

	// 返回结果
	ReplySuccess(w, ChatResponse{Result: result})
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
	ReplySuccess(w, map[string]string{
		"status": "ok",
	})
}
