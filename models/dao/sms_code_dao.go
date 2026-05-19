package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agent/library/log"
)

// SMSCode 短信验证码记录
type SMSCode struct {
	ID        int64
	Phone     string
	Scene     string
	Code      string
	ExpiredAt time.Time
	Used      int
	Attempts  int
	CreatedAt time.Time
}

// SMSCodeDAO 验证码 DAO
type SMSCodeDAO struct {
	mysql *MySQL
}

func NewSMSCodeDAO(mysql *MySQL) *SMSCodeDAO {
	return &SMSCodeDAO{mysql: mysql}
}

func (d *SMSCodeDAO) CreateCode(ctx context.Context, phone, scene, code string, expiredAt time.Time) error {
	query := `INSERT INTO sms_codes (phone, scene, code, expired_at, used, attempts) VALUES (?, ?, ?, ?, 0, 0)`
	_, err := d.mysql.db.ExecContext(ctx, query, phone, scene, code, expiredAt)
	if err != nil {
		log.Error(ctx, "SMSCodeDAO.CreateCode: phone=%s scene=%s err=%v", phone, scene, err)
		return err
	}
	return nil
}

func (d *SMSCodeDAO) FindLatestValidCode(ctx context.Context, phone, scene string) (*SMSCode, error) {
	record := &SMSCode{}
	query := `SELECT id, phone, scene, code, expired_at, used, attempts, created_at
		FROM sms_codes
		WHERE phone = ? AND scene = ?
		ORDER BY id DESC
		LIMIT 1`
	err := d.mysql.db.QueryRowContext(ctx, query, phone, scene).Scan(
		&record.ID, &record.Phone, &record.Scene, &record.Code,
		&record.ExpiredAt, &record.Used, &record.Attempts, &record.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "SMSCodeDAO.FindLatestValidCode: phone=%s scene=%s err=%v", phone, scene, err)
		return nil, err
	}
	return record, nil
}

func (d *SMSCodeDAO) MarkUsed(ctx context.Context, id int64) error {
	_, err := d.mysql.db.ExecContext(ctx, `UPDATE sms_codes SET used = 1 WHERE id = ?`, id)
	if err != nil {
		log.Error(ctx, "SMSCodeDAO.MarkUsed: id=%d err=%v", id, err)
		return err
	}
	return nil
}

func (d *SMSCodeDAO) IncreaseAttempts(ctx context.Context, id int64) error {
	_, err := d.mysql.db.ExecContext(ctx, `UPDATE sms_codes SET attempts = attempts + 1 WHERE id = ?`, id)
	if err != nil {
		log.Error(ctx, "SMSCodeDAO.IncreaseAttempts: id=%d err=%v", id, err)
		return err
	}
	return nil
}
