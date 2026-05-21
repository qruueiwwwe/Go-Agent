package dao

import (
	"context"
	"database/sql"
	"errors"

	"agent/library/log"
	"agent/models/entity"
)

// PersonaChatDAO 角色卡对话历史数据访问对象
type PersonaChatDAO struct {
	mysql *MySQL
}

// NewPersonaChatDAO 创建 PersonaChatDAO 实例
func NewPersonaChatDAO(mysql *MySQL) *PersonaChatDAO {
	return &PersonaChatDAO{mysql: mysql}
}

// Create 创建对话记录
func (d *PersonaChatDAO) Create(ctx context.Context, chat *entity.PersonaChat) error {
	query := `INSERT INTO persona_chats (user_id, persona_id, session_id, role, content) VALUES (?, ?, ?, ?, ?)`

	result, err := d.mysql.db.ExecContext(ctx, query, chat.UserID, chat.PersonaID, chat.SessionID, chat.Role, chat.Content)
	if err != nil {
		log.Error(ctx, "PersonaChatDAO.Create: 创建对话记录失败 userID=%d, err=%v", chat.UserID, err)
		return err
	}

	chat.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

// FindBySession 根据会话ID查询对话历史
func (d *PersonaChatDAO) FindBySession(ctx context.Context, userID int64, sessionID string) ([]entity.PersonaChat, error) {
	query := `SELECT id, user_id, persona_id, session_id, role, content, created_at 
	          FROM persona_chats WHERE user_id = ? AND session_id = ? AND (is_deleted = 0 OR is_deleted IS NULL) ORDER BY created_at ASC`

	rows, err := d.mysql.db.QueryContext(ctx, query, userID, sessionID)
	if err != nil {
		log.Error(ctx, "PersonaChatDAO.FindBySession: 查询失败 sessionID=%s, err=%v", sessionID, err)
		return nil, err
	}
	defer rows.Close()

	chats := make([]entity.PersonaChat, 0)
	for rows.Next() {
		chat := entity.PersonaChat{}
		err := rows.Scan(&chat.ID, &chat.UserID, &chat.PersonaID, &chat.SessionID, &chat.Role, &chat.Content, &chat.CreatedAt)
		if err != nil {
			log.Error(ctx, "PersonaChatDAO.FindBySession: 扫描失败 err=%v", err)
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// FindSessions 查询用户的会话列表
func (d *PersonaChatDAO) FindSessions(ctx context.Context, userID int64, personaID int64) ([]entity.SessionInfo, error) {
	var query string
	var args []interface{}

	if personaID > 0 {
		query = `
			SELECT pc.session_id, pc.persona_id, p.name as persona_name, 
			       (SELECT content FROM persona_chats WHERE session_id = pc.session_id AND (is_deleted = 0 OR is_deleted IS NULL) ORDER BY created_at DESC LIMIT 1) as last_message,
			       MIN(pc.created_at) as created_at, MAX(pc.created_at) as updated_at
			FROM persona_chats pc
			LEFT JOIN personas p ON pc.persona_id = p.id
			WHERE pc.user_id = ? AND pc.persona_id = ? AND (pc.is_deleted = 0 OR pc.is_deleted IS NULL)
			GROUP BY pc.session_id, pc.persona_id, p.name
			ORDER BY updated_at DESC
		`
		args = []interface{}{userID, personaID}
	} else {
		query = `
			SELECT pc.session_id, pc.persona_id, p.name as persona_name, 
			       (SELECT content FROM persona_chats WHERE session_id = pc.session_id AND (is_deleted = 0 OR is_deleted IS NULL) ORDER BY created_at DESC LIMIT 1) as last_message,
			       MIN(pc.created_at) as created_at, MAX(pc.created_at) as updated_at
			FROM persona_chats pc
			LEFT JOIN personas p ON pc.persona_id = p.id
			WHERE pc.user_id = ? AND (pc.is_deleted = 0 OR pc.is_deleted IS NULL)
			GROUP BY pc.session_id, pc.persona_id, p.name
			ORDER BY updated_at DESC
		`
		args = []interface{}{userID}
	}

	rows, err := d.mysql.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Error(ctx, "PersonaChatDAO.FindSessions: 查询失败 userID=%d, err=%v", userID, err)
		return nil, err
	}
	defer rows.Close()

	sessions := make([]entity.SessionInfo, 0)
	for rows.Next() {
		si := entity.SessionInfo{}
		var personaName sql.NullString
		var lastMessage sql.NullString

		err := rows.Scan(&si.SessionID, &si.PersonaID, &personaName, &lastMessage, &si.CreatedAt, &si.UpdatedAt)
		if err != nil {
			log.Error(ctx, "PersonaChatDAO.FindSessions: 扫描失败 err=%v", err)
			return nil, err
		}

		if personaName.Valid {
			si.PersonaName = personaName.String
		}
		if lastMessage.Valid {
			si.LastMessage = lastMessage.String
			// 截取前100字符
			if len(si.LastMessage) > 100 {
				si.LastMessage = si.LastMessage[:100] + "..."
			}
		}

		sessions = append(sessions, si)
	}

	return sessions, nil
}

// DeleteSession 软删除会话的所有对话记录
func (d *PersonaChatDAO) DeleteSession(ctx context.Context, userID int64, sessionID string) error {
	query := `UPDATE persona_chats SET is_deleted = 1 WHERE user_id = ? AND session_id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, userID, sessionID)
	if err != nil {
		log.Error(ctx, "PersonaChatDAO.DeleteSession: 删除失败 sessionID=%s, err=%v", sessionID, err)
		return err
	}
	log.Info(ctx, "PersonaChatDAO.DeleteSession: 删除会话成功 sessionID=%s", sessionID)
	return nil
}

// FindByID 根据ID查询对话记录
func (d *PersonaChatDAO) FindByID(ctx context.Context, id int64) (*entity.PersonaChat, error) {
	query := `SELECT id, user_id, persona_id, session_id, role, content, created_at FROM persona_chats WHERE id = ? AND (is_deleted = 0 OR is_deleted IS NULL)`

	chat := &entity.PersonaChat{}
	err := d.mysql.db.QueryRowContext(ctx, query, id).Scan(
		&chat.ID, &chat.UserID, &chat.PersonaID, &chat.SessionID, &chat.Role, &chat.Content, &chat.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "PersonaChatDAO.FindByID: 查询失败 id=%d, err=%v", id, err)
		return nil, err
	}

	return chat, nil
}
