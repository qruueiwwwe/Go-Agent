package dao

import (
	"context"
	"database/sql"
	"errors"

	"agent/library/log"
	"agent/models/entity"
)

// PersonaDAO 角色卡数据访问对象
type PersonaDAO struct {
	mysql *MySQL
}

// NewPersonaDAO 创建 PersonaDAO 实例
func NewPersonaDAO(mysql *MySQL) *PersonaDAO {
	return &PersonaDAO{mysql: mysql}
}

// Create 创建角色卡
func (d *PersonaDAO) Create(ctx context.Context, persona *entity.Persona) error {
	query := `INSERT INTO personas (name, avatar, tagline, personality, system_prompt, creator_id, is_public, usage_count, status) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var creatorID interface{}
	if persona.CreatorID != nil {
		creatorID = *persona.CreatorID
	}

	result, err := d.mysql.db.ExecContext(ctx, query,
		persona.Name, nullable(persona.Avatar), nullable(persona.Tagline), nullable(persona.Personality),
		persona.SystemPrompt, creatorID, persona.IsPublic, persona.UsageCount, persona.Status)
	if err != nil {
		log.Error(ctx, "PersonaDAO.Create: 创建角色卡失败 name=%s, err=%v", persona.Name, err)
		return err
	}

	persona.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	log.Info(ctx, "PersonaDAO.Create: 角色卡创建成功 id=%d, name=%s", persona.ID, persona.Name)
	return nil
}

// FindByID 根据ID查询角色卡
func (d *PersonaDAO) FindByID(ctx context.Context, id int64) (*entity.Persona, error) {
	query := `SELECT id, name, avatar, tagline, personality, system_prompt, creator_id, is_public, usage_count, status, created_at, updated_at 
	          FROM personas WHERE id = ? AND status != -1`

	persona := &entity.Persona{}
	var avatar, tagline, personality sql.NullString
	var creatorID sql.NullInt64

	err := d.mysql.db.QueryRowContext(ctx, query, id).Scan(
		&persona.ID, &persona.Name, &avatar, &tagline, &personality,
		&persona.SystemPrompt, &creatorID, &persona.IsPublic, &persona.UsageCount, &persona.Status,
		&persona.CreatedAt, &persona.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "PersonaDAO.FindByID: 查询失败 id=%d, err=%v", id, err)
		return nil, err
	}

	persona.Avatar = nullStringValue(avatar)
	persona.Tagline = nullStringValue(tagline)
	persona.Personality = nullStringValue(personality)
	if creatorID.Valid {
		pid := creatorID.Int64
		persona.CreatorID = &pid
	}

	return persona, nil
}

// FindPublic 查询公开角色卡列表
func (d *PersonaDAO) FindPublic(ctx context.Context, page, size int) ([]*entity.Persona, int64, error) {
	// 查询总数
	var total int64
	err := d.mysql.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM personas WHERE is_public = true AND status = 1`).Scan(&total)
	if err != nil {
		log.Error(ctx, "PersonaDAO.FindPublic: 查询总数失败 err=%v", err)
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * size
	query := `SELECT id, name, avatar, tagline, personality, system_prompt, creator_id, is_public, usage_count, status, created_at, updated_at 
	          FROM personas WHERE is_public = true AND status = 1 
	          ORDER BY usage_count DESC, created_at DESC 
	          LIMIT ? OFFSET ?`

	rows, err := d.mysql.db.QueryContext(ctx, query, size, offset)
	if err != nil {
		log.Error(ctx, "PersonaDAO.FindPublic: 查询列表失败 err=%v", err)
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*entity.Persona, 0, size)
	for rows.Next() {
		persona := &entity.Persona{}
		var avatar, tagline, personality sql.NullString
		var creatorID sql.NullInt64

		err := rows.Scan(
			&persona.ID, &persona.Name, &avatar, &tagline, &personality,
			&persona.SystemPrompt, &creatorID, &persona.IsPublic, &persona.UsageCount, &persona.Status,
			&persona.CreatedAt, &persona.UpdatedAt,
		)
		if err != nil {
			log.Error(ctx, "PersonaDAO.FindPublic: 扫描失败 err=%v", err)
			return nil, 0, err
		}

		persona.Avatar = nullStringValue(avatar)
		persona.Tagline = nullStringValue(tagline)
		persona.Personality = nullStringValue(personality)
		if creatorID.Valid {
			pid := creatorID.Int64
			persona.CreatorID = &pid
		}
		list = append(list, persona)
	}

	return list, total, nil
}

// FindByCreator 查询用户创建的角色卡
func (d *PersonaDAO) FindByCreator(ctx context.Context, creatorID int64, page, size int) ([]*entity.Persona, int64, error) {
	// 查询总数
	var total int64
	err := d.mysql.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM personas WHERE creator_id = ? AND status != -1`, creatorID).Scan(&total)
	if err != nil {
		log.Error(ctx, "PersonaDAO.FindByCreator: 查询总数失败 creatorID=%d, err=%v", creatorID, err)
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * size
	query := `SELECT id, name, avatar, tagline, personality, system_prompt, creator_id, is_public, usage_count, status, created_at, updated_at 
	          FROM personas WHERE creator_id = ? AND status != -1 
	          ORDER BY created_at DESC 
	          LIMIT ? OFFSET ?`

	rows, err := d.mysql.db.QueryContext(ctx, query, creatorID, size, offset)
	if err != nil {
		log.Error(ctx, "PersonaDAO.FindByCreator: 查询列表失败 creatorID=%d, err=%v", creatorID, err)
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]*entity.Persona, 0, size)
	for rows.Next() {
		persona := &entity.Persona{}
		var avatar, tagline, personality sql.NullString
		var dbCreatorID sql.NullInt64

		err := rows.Scan(
			&persona.ID, &persona.Name, &avatar, &tagline, &personality,
			&persona.SystemPrompt, &dbCreatorID, &persona.IsPublic, &persona.UsageCount, &persona.Status,
			&persona.CreatedAt, &persona.UpdatedAt,
		)
		if err != nil {
			log.Error(ctx, "PersonaDAO.FindByCreator: 扫描失败 err=%v", err)
			return nil, 0, err
		}

		persona.Avatar = nullStringValue(avatar)
		persona.Tagline = nullStringValue(tagline)
		persona.Personality = nullStringValue(personality)
		if dbCreatorID.Valid {
			pid := dbCreatorID.Int64
			persona.CreatorID = &pid
		}
		list = append(list, persona)
	}

	return list, total, nil
}

// Update 更新角色卡
func (d *PersonaDAO) Update(ctx context.Context, persona *entity.Persona) error {
	query := `UPDATE personas SET name = ?, avatar = ?, tagline = ?, personality = ?, system_prompt = ?, is_public = ?, updated_at = NOW() 
	          WHERE id = ?`

	_, err := d.mysql.db.ExecContext(ctx, query,
		persona.Name, nullable(persona.Avatar), nullable(persona.Tagline), nullable(persona.Personality),
		persona.SystemPrompt, persona.IsPublic, persona.ID)
	if err != nil {
		log.Error(ctx, "PersonaDAO.Update: 更新失败 id=%d, err=%v", persona.ID, err)
		return err
	}

	log.Info(ctx, "PersonaDAO.Update: 角色卡更新成功 id=%d", persona.ID)
	return nil
}

// Delete 软删除角色卡
func (d *PersonaDAO) Delete(ctx context.Context, id int64) error {
	query := `UPDATE personas SET status = -1, updated_at = NOW() WHERE id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error(ctx, "PersonaDAO.Delete: 删除失败 id=%d, err=%v", id, err)
		return err
	}
	log.Info(ctx, "PersonaDAO.Delete: 角色卡删除成功 id=%d", id)
	return nil
}

// IncrementUsage 增加使用次数
func (d *PersonaDAO) IncrementUsage(ctx context.Context, id int64) error {
	query := `UPDATE personas SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error(ctx, "PersonaDAO.IncrementUsage: 增加使用次数失败 id=%d, err=%v", id, err)
		return err
	}
	return nil
}

// FindByIDAndCreator 根据ID和创建者查询（用于权限校验）
func (d *PersonaDAO) FindByIDAndCreator(ctx context.Context, id, creatorID int64) (*entity.Persona, error) {
	query := `SELECT id, name, avatar, tagline, personality, system_prompt, creator_id, is_public, usage_count, status, created_at, updated_at 
	          FROM personas WHERE id = ? AND creator_id = ? AND status != -1`

	persona := &entity.Persona{}
	var avatar, tagline, personality sql.NullString
	var dbCreatorID sql.NullInt64

	err := d.mysql.db.QueryRowContext(ctx, query, id, creatorID).Scan(
		&persona.ID, &persona.Name, &avatar, &tagline, &personality,
		&persona.SystemPrompt, &dbCreatorID, &persona.IsPublic, &persona.UsageCount, &persona.Status,
		&persona.CreatedAt, &persona.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "PersonaDAO.FindByIDAndCreator: 查询失败 id=%d, creatorID=%d, err=%v", id, creatorID, err)
		return nil, err
	}

	persona.Avatar = nullStringValue(avatar)
	persona.Tagline = nullStringValue(tagline)
	persona.Personality = nullStringValue(personality)
	if dbCreatorID.Valid {
		pid := dbCreatorID.Int64
		persona.CreatorID = &pid
	}

	return persona, nil
}
