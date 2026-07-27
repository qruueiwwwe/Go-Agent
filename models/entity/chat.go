package entity

import "time"

// ChatMode 对话模式
const (
	ChatModeNormal   = "normal"   // 普通模式：非流式
	ChatModeThinking = "thinking" // 深度思考模式：流式 + 思考链
	ChatModeAuto     = "auto"     // Auto 模式：流式无思考（管理员）
)

// ChatSession 通用对话会话（区别于 PersonaChat 的角色卡会话）
type ChatSession struct {
	ID            int64     `json:"-"`
	SessionID     string    `json:"session_id"`
	UserID        int64     `json:"user_id"`
	Title         string    `json:"title"`
	Mode          string    `json:"mode"`
	LastMessageAt time.Time `json:"last_message_at"`
	Status        int       `json:"status"` // 1-正常 0-归档 -1-软删
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ChatMessage 通用对话消息
type ChatMessage struct {
	ID             int64     `json:"id"`
	SessionID      string    `json:"session_id"`
	Role           string    `json:"role"` // user / assistant / system
	Content        string    `json:"content"`
	ThoughtContent string    `json:"thought_content,omitempty"`
	ToolCalls      string    `json:"tool_calls,omitempty"`
	Mode           string    `json:"mode,omitempty"`   // normal / thinking / auto
	Status         int       `json:"status,omitempty"` // 1-成功 2-进行中 3-失败
	TokenCount     int       `json:"token_count,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// 消息状态常量
const (
	MsgStatusSuccess    = 1
	MsgStatusInProgress = 2
	MsgStatusFailed     = 3
)

// ChatRequest 通用对话请求（普通模式与流式模式共用）
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"` // 空则新建会话
	Mode      string `json:"mode,omitempty"`       // normal/thinking/auto
}

// ChatData 普通模式响应 data 部分
type ChatData struct {
	Result    string `json:"result"`
	SessionID string `json:"session_id"`
	Title     string `json:"title,omitempty"`
}

// ChatListRequest 获取会话列表请求
type ChatListRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// ChatListResponse 会话列表响应
type ChatListResponse struct {
	Total int64         `json:"total"`
	List  []ChatSession `json:"list"`
}

// ChatHistoryRequest 获取会话消息请求
type ChatHistoryRequest struct {
	SessionID string `json:"session_id"`
	Limit     int    `json:"limit"`
}

// ChatHistoryResponse 会话消息响应
type ChatHistoryResponse struct {
	Session  *ChatSession  `json:"session"`
	Messages []ChatMessage `json:"messages"`
}
