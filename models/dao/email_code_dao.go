package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agent/library/log"
)

// EmailCode 邮箱验证码记录
type EmailCode struct {
	ID        int64
	Email     string
	Scene     string
	Code      string
	ExpiredAt time.Time
	Used      int
	Attempts  int
	CreatedAt time.Time
}

// EmailCodeDAO 邮箱验证码 DAO
type EmailCodeDAO struct {
	mysql *MySQL
}

// NewEmailCodeDAO 创建邮箱验证码 DAO。
func NewEmailCodeDAO(mysql *MySQL) *EmailCodeDAO {
	return &EmailCodeDAO{mysql: mysql}
}

// CreateCode 写入一条新的邮箱验证码记录。
func (d *EmailCodeDAO) CreateCode(ctx context.Context, email, scene, code string, expiredAt time.Time) error {
	query := `INSERT INTO email_codes (email, scene, code, expired_at, used, attempts) VALUES (?, ?, ?, ?, 0, 0)`
	_, err := d.mysql.db.ExecContext(ctx, query, email, scene, code, expiredAt)
	if err != nil {
		log.Error(ctx, "EmailCodeDAO.CreateCode: email=%s scene=%s err=%v", email, scene, err)
		return err
	}
	return nil
}

// FindLatestValidCode 读取同邮箱同场景最新一条验证码记录。
func (d *EmailCodeDAO) FindLatestValidCode(ctx context.Context, email, scene string) (*EmailCode, error) {
	record := &EmailCode{}
	query := `SELECT id, email, scene, code, expired_at, used, attempts, created_at
		FROM email_codes
		WHERE email = ? AND scene = ?
		ORDER BY id DESC
		LIMIT 1`
	err := d.mysql.db.QueryRowContext(ctx, query, email, scene).Scan(
		&record.ID, &record.Email, &record.Scene, &record.Code,
		&record.ExpiredAt, &record.Used, &record.Attempts, &record.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "EmailCodeDAO.FindLatestValidCode: email=%s scene=%s err=%v", email, scene, err)
		return nil, err
	}
	return record, nil
}

func (d *EmailCodeDAO) MarkUsed(ctx context.Context, id int64) error {
	_, err := d.mysql.db.ExecContext(ctx, `UPDATE email_codes SET used = 1 WHERE id = ?`, id)
	if err != nil {
		log.Error(ctx, "EmailCodeDAO.MarkUsed: id=%d err=%v", id, err)
		return err
	}
	return nil
}

func (d *EmailCodeDAO) IncreaseAttempts(ctx context.Context, id int64) error {
	_, err := d.mysql.db.ExecContext(ctx, `UPDATE email_codes SET attempts = attempts + 1 WHERE id = ?`, id)
	if err != nil {
		log.Error(ctx, "EmailCodeDAO.IncreaseAttempts: id=%d err=%v", id, err)
		return err
	}
	return nil
}
