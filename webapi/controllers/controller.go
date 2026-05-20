package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
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
	w.WriteHeader(http.StatusOK)
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
	agentSvc         *agent.AgentService
	rateLimiter      *authService.RateLimiter
	sessionMessages  map[string][]api.Message
	mu               sync.RWMutex
	testBypassAPIKey string
}

// NewChatController 创建聊天控制器
func NewChatController(agentSvc *agent.AgentService, rateLimiter *authService.RateLimiter) *ChatController {
	bypassKey := os.Getenv("CHAT_BYPASS_KEY")
	return &ChatController{
		agentSvc:         agentSvc,
		rateLimiter:      rateLimiter,
		sessionMessages:  make(map[string][]api.Message),
		testBypassAPIKey: bypassKey,
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
func (c *ChatController) Chat(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if ctx == nil || r == nil {
		log.Error(context.Background(), "params is nil")
		ReplyError(w, 404, "参数错误", log.GenerateLogID())
		return
	}

	claims := GetUserFromContext(ctx)
	logid := log.GetLogID(ctx)
	logCtx := log.WithLogID(ctx, logid)
	start := time.Now()

	if r.Method != http.MethodPost {
		log.Warn(logCtx, "收到非POST请求: method=%s", r.Method)
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	bypass := r.Header.Get("key") == c.testBypassAPIKey && c.testBypassAPIKey != ""
	if !bypass && claims == nil {
		log.Warn(logCtx, "Chat: 拒绝未注册请求")
		ReplyError(w, 401, "用户未注册，需注册后才可体验", logid)
		return
	}

	if c.rateLimiter != nil && claims != nil {
		allowed, waitSeconds, err := c.rateLimiter.CheckLimit(ctx, claims.UserID, claims.Role)
		if err != nil {
			log.Error(logCtx, "频率限制检查失败: %v", err)
		} else if !allowed {
			log.Warn(logCtx, "Chat: 频率限制 userID=%d role=%s waitSeconds=%d", claims.UserID, claims.Role, waitSeconds)
			ReplyError(w, 429, fmt.Sprintf("请求过于频繁，请等待 %d 秒后再试", waitSeconds), logid)
			return
		}
		if err := c.rateLimiter.RecordRequest(ctx, claims.UserID); err != nil {
			log.Error(logCtx, "记录请求失败: %v", err)
		}
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(logCtx, "解析请求失败: %v", err)
		ReplyError(w, 404, "解析请求失败", logid)
		return
	}
	if req.Message == "" {
		log.Warn(logCtx, "收到空消息")
		ReplyError(w, 404, "消息不能为空", logid)
		return
	}

	sessionKey := c.buildSessionKey(claims, r, bypass)
	history := c.getSessionHistory(sessionKey)

	log.Info(logCtx, "收到用户消息: %s", req.Message)
	result := c.agentSvc.Process(logCtx, req.Message, history)
	c.appendSessionHistory(sessionKey,
		api.Message{Role: "user", Content: req.Message},
		api.Message{Role: "assistant", Content: result},
	)

	log.Info(logCtx, "处理完成，耗时: %v", time.Since(start))
	ReplySuccess(w, ChatResponse{Result: result}, logid)
}

func (c *ChatController) buildSessionKey(claims *authService.Claims, r *http.Request, bypass bool) string {
	if claims != nil {
		return "user:" + strconv.FormatInt(claims.UserID, 10)
	}
	if bypass {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && host != "" {
			return "guest:" + host
		}
		if r.RemoteAddr != "" {
			return "guest:" + r.RemoteAddr
		}
	}
	return "guest:unknown"
}

func (c *ChatController) getSessionHistory(sessionKey string) []api.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	history := c.sessionMessages[sessionKey]
	copied := make([]api.Message, len(history))
	copy(copied, history)
	return copied
}

func (c *ChatController) appendSessionHistory(sessionKey string, msgs ...api.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionMessages[sessionKey] = append(c.sessionMessages[sessionKey], msgs...)
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
