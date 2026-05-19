package dao

import (
	"context"
	"time"

	"agent/library/log"
)

// RateLimitDAO 频率限制数据访问对象
type RateLimitDAO struct {
	mysql *MySQL
}

// NewRateLimitDAO 创建 RateLimitDAO 实例
func NewRateLimitDAO(mysql *MySQL) *RateLimitDAO {
	return &RateLimitDAO{mysql: mysql}
}

// CreateRecord 记录一次请求
func (d *RateLimitDAO) CreateRecord(ctx context.Context, userID int64, endpoint string, requestAt time.Time) error {
	query := `INSERT INTO rate_limits (user_id, endpoint, request_at) VALUES (?, ?, ?)`
	_, err := d.mysql.db.ExecContext(ctx, query, userID, endpoint, requestAt)
	if err != nil {
		log.Error(ctx, "RateLimitDAO.CreateRecord: 记录失败 userID=%d, err=%v", userID, err)
		return err
	}
	return nil
}

// CountRequests 统计用户在指定时间窗口内的请求数
func (d *RateLimitDAO) CountRequests(ctx context.Context, userID int64, endpoint string, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM rate_limits WHERE user_id = ? AND endpoint = ? AND request_at >= ?`
	var count int
	err := d.mysql.db.QueryRowContext(ctx, query, userID, endpoint, since).Scan(&count)
	if err != nil {
		log.Error(ctx, "RateLimitDAO.CountRequests: 查询失败 userID=%d, err=%v", userID, err)
		return 0, err
	}
	return count, nil
}

// GetOldestInWindow 获取时间窗口内最早的请求时间
func (d *RateLimitDAO) GetOldestInWindow(ctx context.Context, userID int64, endpoint string, since time.Time) (time.Time, error) {
	query := `SELECT request_at FROM rate_limits WHERE user_id = ? AND endpoint = ? AND request_at >= ? ORDER BY request_at ASC LIMIT 1`
	var t time.Time
	err := d.mysql.db.QueryRowContext(ctx, query, userID, endpoint, since).Scan(&t)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// CleanExpired 清理过期的频率限制记录（超过24小时的）
func (d *RateLimitDAO) CleanExpired(ctx context.Context) error {
	cutoff := time.Now().Add(-24 * time.Hour)
	query := `DELETE FROM rate_limits WHERE request_at < ?`
	_, err := d.mysql.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		log.Error(ctx, "RateLimitDAO.CleanExpired: 清理失败 err=%v", err)
		return err
	}
	return nil
}
