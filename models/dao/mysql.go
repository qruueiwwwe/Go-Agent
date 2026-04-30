package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"agent/global"
	"agent/library/log"

	_ "github.com/go-sql-driver/mysql"
)

// MySQL MySQL连接管理
type MySQL struct {
	db *sql.DB
}

// NewMySQL 创建MySQL连接
func NewMySQL(cfg global.DatabaseConfig) (*MySQL, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpen)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL连接测试失败: %v", err)
	}

	return &MySQL{db: db}, nil
}

// Close 关闭连接
func (m *MySQL) Close() error {
	return m.db.Close()
}

// NbnhhshDAO 缩写词猜测数据访问
type NbnhhshDAO struct {
	mysql *MySQL
}

// NewNbnhhshDAO 创建NbnhhshDAO
func NewNbnhhshDAO(mysql *MySQL) *NbnhhshDAO {
	return &NbnhhshDAO{mysql: mysql}
}

// NbnhhshRecord 缩写词记录
type NbnhhshRecord struct {
	Name        string
	Trans       []string
	CreateTime  time.Time
	UpdatedTime time.Time
}

// GetByName 根据名称查询记录
func (d *NbnhhshDAO) GetByName(ctx context.Context, name string) (*NbnhhshRecord, error) {
	query := "SELECT name, trans, create_time, updated_time FROM nbnhhsh WHERE name = ?"

	var record NbnhhshRecord
	var transJSON string

	err := d.mysql.db.QueryRowContext(ctx, query, name).Scan(&record.Name, &transJSON, &record.CreateTime, &record.UpdatedTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.GetByName: 查询失败 name=%s, err=%v", name, err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(transJSON), &record.Trans); err != nil {
		log.Error(ctx, "NbnhhshDAO.GetByName: 解析JSON失败 name=%s, err=%v", name, err)
		return nil, err
	}

	return &record, nil
}

// Save 保存记录
func (d *NbnhhshDAO) Save(ctx context.Context, name string, trans []string) error {
	transJSON, err := json.Marshal(trans)
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.Save: 序列化JSON失败 name=%s, err=%v", name, err)
		return err
	}

	query := `
		INSERT INTO nbnhhsh (name, trans, create_time, updated_time) 
		VALUES (?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE trans = VALUES(trans), updated_time = NOW()
	`

	_, err = d.mysql.db.ExecContext(ctx, query, name, transJSON)
	if err != nil {
		log.Error(ctx, "NbnhhshDAO.Save: 保存失败 name=%s, err=%v", name, err)
		return err
	}

	log.Info(ctx, "NbnhhshDAO.Save: 保存成功 name=%s, trans=%v", name, trans)
	return nil
}

// IsCacheValid 检查缓存是否有效（3天内）
func (d *NbnhhshDAO) IsCacheValid(record *NbnhhshRecord) bool {
	if record == nil {
		return false
	}
	return time.Since(record.UpdatedTime) <= 3*24*time.Hour
}
