package dao

import (
	"context"
	"database/sql"
	"errors"

	"agent/library/log"
	"agent/models/entity"
)

// PersonaExampleDAO 角色卡示例对话数据访问对象
type PersonaExampleDAO struct {
	mysql *MySQL
}

// NewPersonaExampleDAO 创建 PersonaExampleDAO 实例
func NewPersonaExampleDAO(mysql *MySQL) *PersonaExampleDAO {
	return &PersonaExampleDAO{mysql: mysql}
}

// BatchCreate 批量创建示例对话
func (d *PersonaExampleDAO) BatchCreate(ctx context.Context, personaID int64, examples []entity.Example) error {
	if len(examples) == 0 {
		return nil
	}

	query := `INSERT INTO persona_examples (persona_id, user_input, ai_response, sort_order) VALUES (?, ?, ?, ?)`
	for i, ex := range examples {
		_, err := d.mysql.db.ExecContext(ctx, query, personaID, ex.UserInput, ex.AIResponse, i)
		if err != nil {
			log.Error(ctx, "PersonaExampleDAO.BatchCreate: 创建示例失败 personaID=%d, err=%v", personaID, err)
			return err
		}
	}

	log.Info(ctx, "PersonaExampleDAO.BatchCreate: 批量创建示例成功 personaID=%d, count=%d", personaID, len(examples))
	return nil
}

// FindByPersonaID 根据角色卡ID查询示例对话
func (d *PersonaExampleDAO) FindByPersonaID(ctx context.Context, personaID int64) ([]entity.Example, error) {
	query := `SELECT id, persona_id, user_input, ai_response, sort_order 
	          FROM persona_examples WHERE persona_id = ? AND (is_deleted = 0 OR is_deleted IS NULL) ORDER BY sort_order ASC`

	rows, err := d.mysql.db.QueryContext(ctx, query, personaID)
	if err != nil {
		log.Error(ctx, "PersonaExampleDAO.FindByPersonaID: 查询失败 personaID=%d, err=%v", personaID, err)
		return nil, err
	}
	defer rows.Close()

	examples := make([]entity.Example, 0)
	for rows.Next() {
		ex := entity.Example{}
		err := rows.Scan(&ex.ID, &ex.PersonaID, &ex.UserInput, &ex.AIResponse, &ex.SortOrder)
		if err != nil {
			log.Error(ctx, "PersonaExampleDAO.FindByPersonaID: 扫描失败 err=%v", err)
			return nil, err
		}
		examples = append(examples, ex)
	}

	return examples, nil
}

// DeleteByPersonaID 软删除角色卡的所有示例对话
func (d *PersonaExampleDAO) DeleteByPersonaID(ctx context.Context, personaID int64) error {
	query := `UPDATE persona_examples SET is_deleted = 1 WHERE persona_id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, personaID)
	if err != nil {
		log.Error(ctx, "PersonaExampleDAO.DeleteByPersonaID: 删除失败 personaID=%d, err=%v", personaID, err)
		return err
	}
	log.Info(ctx, "PersonaExampleDAO.DeleteByPersonaID: 删除示例成功 personaID=%d", personaID)
	return nil
}

// FindByID 根据ID查款示例对话
func (d *PersonaExampleDAO) FindByID(ctx context.Context, id int64) (*entity.Example, error) {
	query := `SELECT id, persona_id, user_input, ai_response, sort_order FROM persona_examples WHERE id = ? AND (is_deleted = 0 OR is_deleted IS NULL)`

	ex := &entity.Example{}
	err := d.mysql.db.QueryRowContext(ctx, query, id).Scan(&ex.ID, &ex.PersonaID, &ex.UserInput, &ex.AIResponse, &ex.SortOrder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "PersonaExampleDAO.FindByID: 查询失败 id=%d, err=%v", id, err)
		return nil, err
	}

	return ex, nil
}
