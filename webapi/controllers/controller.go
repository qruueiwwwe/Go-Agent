package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"agent/library/log"
	"agent/models/dao"
	"agent/models/entity"
	"agent/models/service/agent"
	authService "agent/models/service/auth"

	"github.com/google/uuid"
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
	chatDAO          *dao.ChatDAO
	testBypassAPIKey string
}

// NewChatController 创建聊天控制器
func NewChatController(agentSvc *agent.AgentService, rateLimiter *authService.RateLimiter, chatDAO *dao.ChatDAO) *ChatController {
	bypassKey := os.Getenv("CHAT_BYPASS_KEY")
	return &ChatController{
		agentSvc:         agentSvc,
		rateLimiter:      rateLimiter,
		chatDAO:          chatDAO,
		testBypassAPIKey: bypassKey,
	}
}

// ========== 普通模式 /api/chat ==========

// Chat 处理普通聊天请求（非流式，持久化到 MySQL）
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

	var req entity.ChatRequest
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

	// 决定用户 ID：无 claims 且 bypass 时用 0 表示访客（不入库）
	var userID int64
	if claims != nil {
		userID = claims.UserID
	}

	// 会话处理
	sessionID, session, err := c.ensureSession(logCtx, req.SessionID, userID, entity.ChatModeNormal)
	if err != nil {
		log.Error(logCtx, "ensureSession 失败: %v", err)
		ReplyError(w, 500, "创建会话失败", logid)
		return
	}

	// 拉取历史
	history := c.loadHistory(logCtx, sessionID)

	log.Info(logCtx, "收到用户消息: %s (session=%s)", req.Message, sessionID)
	result := c.agentSvc.Process(logCtx, req.Message, history)

	// 持久化：user 消息 + assistant 消息
	if c.chatDAO != nil && userID > 0 {
		_ = c.chatDAO.InsertMessage(logCtx, &entity.ChatMessage{
			SessionID: sessionID, Role: "user", Content: req.Message,
			Mode: entity.ChatModeNormal, Status: entity.MsgStatusSuccess,
		})
		_ = c.chatDAO.InsertMessage(logCtx, &entity.ChatMessage{
			SessionID: sessionID, Role: "assistant", Content: result,
			Mode: entity.ChatModeNormal, Status: entity.MsgStatusSuccess,
		})
		_ = c.chatDAO.UpdateSessionLastMessageAt(logCtx, sessionID, time.Now())
	}

	// 异步生成标题（首次对话）
	title := session.Title
	if title == "" && c.chatDAO != nil && userID > 0 {
		go c.asyncGenerateTitle(sessionID, req.Message, result)
	}

	log.Info(logCtx, "处理完成，耗时: %v", time.Since(start))
	ReplySuccess(w, entity.ChatData{
		Result:    result,
		SessionID: sessionID,
		Title:     title,
	}, logid)
}

// ensureSession 确保存在会话；若 sessionID 为空则创建新会话
func (c *ChatController) ensureSession(ctx context.Context, sessionID string, userID int64, mode string) (string, *entity.ChatSession, error) {
	if c.chatDAO == nil || userID <= 0 {
		// DAO 不可用或访客模式：返回临时 sessionID，不入库
		if sessionID == "" {
			sessionID = uuid.New().String()
		}
		return sessionID, &entity.ChatSession{SessionID: sessionID, UserID: userID, Mode: mode}, nil
	}

	if sessionID != "" {
		s, err := c.chatDAO.GetSessionByID(ctx, sessionID)
		if err != nil {
			return "", nil, err
		}
		if s != nil {
			if s.UserID != userID {
				return "", nil, fmt.Errorf("session 不属于当前用户")
			}
			return s.SessionID, s, nil
		}
	}

	newSession := &entity.ChatSession{
		SessionID:     uuid.New().String(),
		UserID:        userID,
		Mode:          mode,
		LastMessageAt: time.Now(),
	}
	if err := c.chatDAO.CreateSession(ctx, newSession); err != nil {
		return "", nil, err
	}
	return newSession.SessionID, newSession, nil
}

// loadHistory 从 DAO 拉取最近的历史消息，转换为 api.Message 格式
func (c *ChatController) loadHistory(ctx context.Context, sessionID string) []api.Message {
	if c.chatDAO == nil {
		return nil
	}
	msgs, err := c.chatDAO.ListRecentForContext(ctx, sessionID, 12)
	if err != nil || len(msgs) == 0 {
		return nil
	}
	out := make([]api.Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, api.Message{Role: m.Role, Content: m.Content})
	}
	return out
}

// asyncGenerateTitle 异步生成会话标题
func (c *ChatController) asyncGenerateTitle(sessionID, userMsg, aiAnswer string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ctx = log.WithContext(ctx)
	title := c.agentSvc.GenerateTitle(ctx, userMsg, aiAnswer)
	if title == "" {
		return
	}
	if err := c.chatDAO.UpdateSessionTitle(ctx, sessionID, title); err != nil {
		log.Warn(ctx, "asyncGenerateTitle: 更新标题失败 sessionID=%s err=%v", sessionID, err)
		return
	}
	log.Info(ctx, "asyncGenerateTitle: 标题已生成 sessionID=%s title=%s", sessionID, title)
}

// ========== 流式模式 /api/chat/stream ==========

// ChatStream 流式对话处理
func (c *ChatController) ChatStream(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GetLogID(ctx)
	logCtx := log.WithLogID(ctx, logid)

	if claims == nil {
		ReplyError(w, 401, "未登录", logid)
		return
	}
	if r.Method != http.MethodPost {
		ReplyError(w, 404, "只支持 POST 方法", logid)
		return
	}

	// 频率限制
	if c.rateLimiter != nil {
		allowed, waitSeconds, err := c.rateLimiter.CheckLimit(ctx, claims.UserID, claims.Role)
		if err == nil && !allowed {
			ReplyError(w, 429, fmt.Sprintf("请求过于频繁，请等待 %d 秒后再试", waitSeconds), logid)
			return
		}
		_ = c.rateLimiter.RecordRequest(ctx, claims.UserID)
	}

	// 解析请求
	var req entity.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ReplyError(w, 404, "解析请求失败", logid)
		return
	}
	if req.Message == "" {
		ReplyError(w, 404, "消息不能为空", logid)
		return
	}

	// 模式
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = entity.ChatModeThinking
	}

	// SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 会话
	sessionID, session, err := c.ensureSession(logCtx, req.SessionID, claims.UserID, mode)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"content": "会话创建失败"})
		return
	}

	// 立即推送 session 事件
	writeSSE(w, flusher, "session", map[string]string{"session_id": sessionID, "title": session.Title})

	// 保存 user 消息（status=进行中，回答完成后一起更新）
	if c.chatDAO != nil {
		_ = c.chatDAO.InsertMessage(logCtx, &entity.ChatMessage{
			SessionID: sessionID, Role: "user", Content: req.Message,
			Mode: mode, Status: entity.MsgStatusSuccess,
		})
	}

	// 组装历史
	history := c.loadHistory(logCtx, sessionID)
	// 去掉最新一条（当前 user 消息），因为 Process 会自己加
	if len(history) > 0 && history[len(history)-1].Role == "user" {
		history = history[:len(history)-1]
	}

	// 启动流式生成
	chunkCh := make(chan agent.StreamChunk, 64)
	doneCh := make(chan struct{})
	var thoughtOut, answerOut string
	go func() {
		streamMode := agent.StreamModeThinking
		if mode == entity.ChatModeAuto {
			streamMode = agent.StreamModeAuto
		}
		thoughtOut, answerOut = c.agentSvc.ProcessStream(logCtx, req.Message, history, streamMode, chunkCh)
		close(chunkCh)
		close(doneCh)
	}()

	// 消费 chunk 并推送 SSE
	streamStatus := entity.MsgStatusSuccess
loop:
	for {
		select {
		case <-ctx.Done():
			log.Info(logCtx, "ChatStream: 客户端断开")
			streamStatus = entity.MsgStatusFailed
			break loop
		case chunk, ok := <-chunkCh:
			if !ok {
				break loop
			}
			payload := map[string]interface{}{"content": chunk.Content}
			if chunk.Tool != "" {
				payload["tool"] = chunk.Tool
			}
			if chunk.ToolInput != "" {
				payload["tool_input"] = chunk.ToolInput
			}
			writeSSE(w, flusher, string(chunk.Type), payload)
			if chunk.Type == agent.ChunkError {
				streamStatus = entity.MsgStatusFailed
				break loop
			}
		}
	}
	<-doneCh

	// 保存 assistant 消息
	if c.chatDAO != nil && answerOut != "" {
		_ = c.chatDAO.InsertMessage(logCtx, &entity.ChatMessage{
			SessionID:      sessionID,
			Role:           "assistant",
			Content:        answerOut,
			ThoughtContent: thoughtOut,
			Mode:           mode,
			Status:         streamStatus,
		})
		_ = c.chatDAO.UpdateSessionLastMessageAt(logCtx, sessionID, time.Now())
	}

	// 生成标题（同步，用于紧接推 title 事件）
	if session.Title == "" && c.chatDAO != nil && answerOut != "" {
		titleCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		title := c.agentSvc.GenerateTitle(titleCtx, req.Message, answerOut)
		if title != "" {
			_ = c.chatDAO.UpdateSessionTitle(titleCtx, sessionID, title)
			writeSSE(w, flusher, "title", map[string]string{"session_id": sessionID, "title": title})
		}
	}

	// done 事件
	writeSSE(w, flusher, "done", map[string]string{"session_id": sessionID})
}

// writeSSE 写入一帧 SSE
func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	payload, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}

// ========== 历史查询接口 ==========

// GetChatList 获取当前用户的会话列表
func (c *ChatController) GetChatList(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GetLogID(ctx)
	if claims == nil {
		ReplyError(w, 401, "未登录", logid)
		return
	}
	if c.chatDAO == nil {
		ReplyError(w, 500, "会话服务未启用", logid)
		return
	}

	var req entity.ChatListRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	list, total, err := c.chatDAO.ListSessionsByUser(ctx, claims.UserID, req.Page, req.Size)
	if err != nil {
		ReplyError(w, 500, "查询失败", logid)
		return
	}
	ReplySuccess(w, entity.ChatListResponse{Total: total, List: list}, logid)
}

// GetChatHistory 获取指定会话的所有消息
func (c *ChatController) GetChatHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	claims := GetUserFromContext(ctx)
	logid := log.GetLogID(ctx)
	if claims == nil {
		ReplyError(w, 401, "未登录", logid)
		return
	}
	if c.chatDAO == nil {
		ReplyError(w, 500, "会话服务未启用", logid)
		return
	}

	var req entity.ChatHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		ReplyError(w, 400, "session_id 不能为空", logid)
		return
	}

	session, err := c.chatDAO.GetSessionByID(ctx, req.SessionID)
	if err != nil {
		ReplyError(w, 500, "查询失败", logid)
		return
	}
	if session == nil {
		ReplyError(w, 404, "会话不存在", logid)
		return
	}
	if session.UserID != claims.UserID && claims.Role != "admin" {
		ReplyError(w, 403, "无权访问该会话", logid)
		return
	}

	messages, err := c.chatDAO.ListMessagesBySession(ctx, req.SessionID, req.Limit)
	if err != nil {
		ReplyError(w, 500, "查询消息失败", logid)
		return
	}
	ReplySuccess(w, entity.ChatHistoryResponse{Session: session, Messages: messages}, logid)
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
