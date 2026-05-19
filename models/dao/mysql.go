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

// DB 获取底层数据库连接
func (m *MySQL) DB() *sql.DB {
	return m.db
}

// AutoMigrate 自动迁移数据库表
func (m *MySQL) AutoMigrate(ctx context.Context) error {
	// 创建用户表
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
		password VARCHAR(255) NOT NULL COMMENT '密码(bcrypt加密)',
		nickname VARCHAR(100) COMMENT '昵称',
		phone VARCHAR(20) UNIQUE NULL COMMENT '手机号',
		email VARCHAR(255) UNIQUE NULL COMMENT '邮箱',
		role VARCHAR(20) DEFAULT 'user' COMMENT '角色: admin/user',
		status TINYINT DEFAULT 1 COMMENT '状态: 1-正常 0-禁用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_phone (phone),
		INDEX idx_email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
	`

	_, err := m.db.ExecContext(ctx, createUserTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建用户表失败 err=%v", err)
		return fmt.Errorf("创建用户表失败: %v", err)
	}

	createEmailCodeTable := `
	CREATE TABLE IF NOT EXISTS email_codes (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		email VARCHAR(255) NOT NULL COMMENT '邮箱',
		scene VARCHAR(32) NOT NULL COMMENT '场景: register/reset_password',
		code VARCHAR(20) NOT NULL COMMENT '验证码',
		expired_at DATETIME NOT NULL COMMENT '过期时间',
		used TINYINT DEFAULT 0 COMMENT '是否已使用',
		attempts INT DEFAULT 0 COMMENT '尝试次数',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_email_scene (email, scene),
		INDEX idx_expired_at (expired_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱验证码表';
	`

	_, err = m.db.ExecContext(ctx, createEmailCodeTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建邮箱验证码表失败 err=%v", err)
		return fmt.Errorf("创建邮箱验证码表失败: %v", err)
	}

	// 创建频率限制记录表
	createRateLimitTable := `
	CREATE TABLE IF NOT EXISTS rate_limits (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT NOT NULL COMMENT '用户ID',
		endpoint VARCHAR(50) NOT NULL COMMENT '接口标识',
		request_at DATETIME NOT NULL COMMENT '请求时间',
		INDEX idx_user_endpoint_time (user_id, endpoint, request_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频率限制记录表';
	`

	_, err = m.db.ExecContext(ctx, createRateLimitTable)
	if err != nil {
		log.Error(ctx, "AutoMigrate: 创建频率限制表失败 err=%v", err)
		return fmt.Errorf("创建频率限制表失败: %v", err)
	}

	log.Info(ctx, "AutoMigrate: 数据库迁移完成")
	return nil
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
