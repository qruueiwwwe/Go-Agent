package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agent/library/log"
	"agent/models/entity"
)

// ChatDAO 通用对话（session + message）数据访问
type ChatDAO struct {
	mysql *MySQL
}

// NewChatDAO 创建 ChatDAO 实例
func NewChatDAO(mysql *MySQL) *ChatDAO {
	return &ChatDAO{mysql: mysql}
}

// CreateSession 创建新会话
func (d *ChatDAO) CreateSession(ctx context.Context, s *entity.ChatSession) error {
	if s.Mode == "" {
		s.Mode = entity.ChatModeNormal
	}
	if s.Status == 0 {
		s.Status = 1
	}
	now := time.Now()
	if s.LastMessageAt.IsZero() {
		s.LastMessageAt = now
	}
	query := `INSERT INTO chat_sessions (session_id, user_id, title, mode, last_message_at, status)
	          VALUES (?, ?, ?, ?, ?, ?)`
	result, err := d.mysql.db.ExecContext(ctx, query,
		s.SessionID, s.UserID, s.Title, s.Mode, s.LastMessageAt, s.Status)
	if err != nil {
		log.Error(ctx, "ChatDAO.CreateSession: 失败 sessionID=%s, err=%v", s.SessionID, err)
		return err
	}
	s.ID, _ = result.LastInsertId()
	return nil
}

// UpdateSessionTitle 更新会话标题
func (d *ChatDAO) UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	query := `UPDATE chat_sessions SET title = ? WHERE session_id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, title, sessionID)
	if err != nil {
		log.Error(ctx, "ChatDAO.UpdateSessionTitle: 失败 sessionID=%s, err=%v", sessionID, err)
	}
	return err
}

// UpdateSessionLastMessageAt 更新会话最新消息时间
func (d *ChatDAO) UpdateSessionLastMessageAt(ctx context.Context, sessionID string, t time.Time) error {
	query := `UPDATE chat_sessions SET last_message_at = ? WHERE session_id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, t, sessionID)
	if err != nil {
		log.Error(ctx, "ChatDAO.UpdateSessionLastMessageAt: 失败 sessionID=%s, err=%v", sessionID, err)
	}
	return err
}

// GetSessionByID 根据 session_id 获取会话
func (d *ChatDAO) GetSessionByID(ctx context.Context, sessionID string) (*entity.ChatSession, error) {
	query := `SELECT id, session_id, user_id, title, mode, last_message_at, status, created_at, updated_at
	          FROM chat_sessions WHERE session_id = ? AND status >= 0`
	s := &entity.ChatSession{}
	err := d.mysql.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID, &s.SessionID, &s.UserID, &s.Title, &s.Mode,
		&s.LastMessageAt, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "ChatDAO.GetSessionByID: 失败 sessionID=%s, err=%v", sessionID, err)
		return nil, err
	}
	return s, nil
}

// ListSessionsByUser 分页列出用户会话
func (d *ChatDAO) ListSessionsByUser(ctx context.Context, userID int64, page, size int) ([]entity.ChatSession, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	var total int64
	if err := d.mysql.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chat_sessions WHERE user_id = ? AND status = 1`, userID).Scan(&total); err != nil {
		log.Error(ctx, "ChatDAO.ListSessionsByUser: count 失败 userID=%d, err=%v", userID, err)
		return nil, 0, err
	}

	rows, err := d.mysql.db.QueryContext(ctx,
		`SELECT id, session_id, user_id, title, mode, last_message_at, status, created_at, updated_at
		 FROM chat_sessions WHERE user_id = ? AND status = 1
		 ORDER BY last_message_at DESC LIMIT ? OFFSET ?`, userID, size, offset)
	if err != nil {
		log.Error(ctx, "ChatDAO.ListSessionsByUser: 查询失败 userID=%d, err=%v", userID, err)
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]entity.ChatSession, 0, size)
	for rows.Next() {
		s := entity.ChatSession{}
		if err := rows.Scan(&s.ID, &s.SessionID, &s.UserID, &s.Title, &s.Mode,
			&s.LastMessageAt, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			log.Error(ctx, "ChatDAO.ListSessionsByUser: 扫描失败 err=%v", err)
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

// SoftDeleteSession 软删除会话
func (d *ChatDAO) SoftDeleteSession(ctx context.Context, sessionID string, userID int64) error {
	query := `UPDATE chat_sessions SET status = -1 WHERE session_id = ? AND user_id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		log.Error(ctx, "ChatDAO.SoftDeleteSession: 失败 sessionID=%s, err=%v", sessionID, err)
	}
	return err
}

// InsertMessage 写入一条消息
func (d *ChatDAO) InsertMessage(ctx context.Context, m *entity.ChatMessage) error {
	if m.Mode == "" {
		m.Mode = entity.ChatModeNormal
	}
	if m.Status == 0 {
		m.Status = entity.MsgStatusSuccess
	}
	query := `INSERT INTO chat_messages (session_id, role, content, thought_content, tool_calls, mode, status, token_count)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := d.mysql.db.ExecContext(ctx, query,
		m.SessionID, m.Role, m.Content, m.ThoughtContent, m.ToolCalls, m.Mode, m.Status, m.TokenCount)
	if err != nil {
		log.Error(ctx, "ChatDAO.InsertMessage: 失败 sessionID=%s, role=%s, err=%v", m.SessionID, m.Role, err)
		return err
	}
	m.ID, _ = result.LastInsertId()
	return nil
}

// ListMessagesBySession 按创建顺序拉取消息（用于历史展示）
func (d *ChatDAO) ListMessagesBySession(ctx context.Context, sessionID string, limit int) ([]entity.ChatMessage, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := d.mysql.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, IFNULL(thought_content,''), IFNULL(tool_calls,''),
		        IFNULL(mode,'normal'), IFNULL(status,1), IFNULL(token_count,0), created_at
		 FROM chat_messages WHERE session_id = ? ORDER BY id ASC LIMIT ?`, sessionID, limit)
	if err != nil {
		log.Error(ctx, "ChatDAO.ListMessagesBySession: 失败 sessionID=%s, err=%v", sessionID, err)
		return nil, err
	}
	defer rows.Close()

	list := make([]entity.ChatMessage, 0)
	for rows.Next() {
		m := entity.ChatMessage{}
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content,
			&m.ThoughtContent, &m.ToolCalls, &m.Mode, &m.Status, &m.TokenCount, &m.CreatedAt); err != nil {
			log.Error(ctx, "ChatDAO.ListMessagesBySession: 扫描失败 err=%v", err)
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

// ListRecentForContext 拉取最近 N 条消息（按时间升序返回，用于组装 LLM 上下文）
func (d *ChatDAO) ListRecentForContext(ctx context.Context, sessionID string, n int) ([]entity.ChatMessage, error) {
	if n <= 0 {
		n = 12
	}
	rows, err := d.mysql.db.QueryContext(ctx,
		`SELECT id, session_id, role, content, IFNULL(thought_content,''), IFNULL(tool_calls,''),
		        IFNULL(mode,'normal'), IFNULL(status,1), IFNULL(token_count,0), created_at
		 FROM chat_messages WHERE session_id = ? ORDER BY id DESC LIMIT ?`, sessionID, n)
	if err != nil {
		log.Error(ctx, "ChatDAO.ListRecentForContext: 失败 sessionID=%s, err=%v", sessionID, err)
		return nil, err
	}
	defer rows.Close()

	list := make([]entity.ChatMessage, 0, n)
	for rows.Next() {
		m := entity.ChatMessage{}
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content,
			&m.ThoughtContent, &m.ToolCalls, &m.Mode, &m.Status, &m.TokenCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	// 反转为时间升序
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}
