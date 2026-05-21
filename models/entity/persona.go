package entity

import "time"

// Persona 角色卡实体
type Persona struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar,omitempty"`
	Tagline      string    `json:"tagline,omitempty"`
	Personality  string    `json:"personality,omitempty"`
	SystemPrompt string    `json:"system_prompt"`
	CreatorID    *int64    `json:"creator_id,omitempty"` // NULL 表示系统预置
	IsPublic     bool      `json:"is_public"`
	UsageCount   int       `json:"usage_count"`
	Status       int       `json:"status"` // 1-正常 0-禁用 -1-删除
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联数据（不存数据库）
	Examples []Example `json:"examples,omitempty"`
}

// Example 角色卡示例对话
type Example struct {
	ID         int64  `json:"id,omitempty"`
	PersonaID  int64  `json:"persona_id,omitempty"`
	UserInput  string `json:"user_input"`
	AIResponse string `json:"ai_response"`
	SortOrder  int    `json:"sort_order,omitempty"`
}

// PersonaChat 角色卡对话记录
type PersonaChat struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	PersonaID int64     `json:"persona_id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"` // user/assistant/system
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// PersonaCreateRequest 创建角色卡请求
type PersonaCreateRequest struct {
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar,omitempty"`
	Tagline      string    `json:"tagline,omitempty"`
	Personality  string    `json:"personality,omitempty"`
	SystemPrompt string    `json:"system_prompt"`
	IsPublic     bool      `json:"is_public"`
	Examples     []Example `json:"examples,omitempty"`
}

// PersonaChatRequest 角色卡对话请求
type PersonaChatRequest struct {
	PersonaID int64  `json:"persona_id"`
	SessionID string `json:"session_id,omitempty"` // 可选，不传则创建新会话
	Message   string `json:"message"`
}

// PersonaChatResponse 角色卡对话响应
type PersonaChatResponse struct {
	SessionID   string `json:"session_id"`
	Response    string `json:"response"`
	PersonaID   int64  `json:"persona_id"`
	PersonaName string `json:"persona_name"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID   string    `json:"session_id"`
	PersonaID   int64     `json:"persona_id"`
	PersonaName string    `json:"persona_name"`
	LastMessage string    `json:"last_message"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SessionHistory 会话历史
type SessionHistory struct {
	SessionID   string        `json:"session_id"`
	PersonaID   int64         `json:"persona_id"`
	PersonaName string        `json:"persona_name"`
	Messages    []PersonaChat `json:"messages"`
}

// PersonaListResponse 角色卡列表响应
type PersonaListResponse struct {
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
	List  []Persona `json:"list"`
}
