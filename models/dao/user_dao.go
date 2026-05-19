package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agent/library/log"
)

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // 不序列化到 JSON
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDAO 用户数据访问对象
type UserDAO struct {
	mysql *MySQL
}

// NewUserDAO 创建 UserDAO 实例
func NewUserDAO(mysql *MySQL) *UserDAO {
	return &UserDAO{mysql: mysql}
}

// Create 创建用户
func (d *UserDAO) Create(ctx context.Context, user *User) error {
	query := `INSERT INTO users (username, password, nickname, phone, email, role, status) 
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := d.mysql.db.ExecContext(ctx, query,
		user.Username, user.Password, user.Nickname, nullable(user.Phone), nullable(user.Email), user.Role, user.Status)
	if err != nil {
		log.Error(ctx, "UserDAO.Create: 创建用户失败 username=%s, err=%v", user.Username, err)
		return err
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	log.Info(ctx, "UserDAO.Create: 用户创建成功 id=%d, username=%s", user.ID, user.Username)
	return nil
}

// FindByUsername 根据用户名查找用户
func (d *UserDAO) FindByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, nickname, phone, email, role, status, created_at, updated_at 
	          FROM users WHERE username = ?`

	var phone, email sql.NullString
	err := d.mysql.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Nickname, &phone, &email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	user.Phone = nullStringValue(phone)
	user.Email = nullStringValue(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "UserDAO.FindByUsername: 查询失败 username=%s, err=%v", username, err)
		return nil, err
	}

	return user, nil
}

// FindByID 根据 ID 查找用户
func (d *UserDAO) FindByID(ctx context.Context, id int64) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, nickname, phone, email, role, status, created_at, updated_at 
	          FROM users WHERE id = ?`

	var phone, email sql.NullString
	err := d.mysql.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Nickname, &phone, &email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	user.Phone = nullStringValue(phone)
	user.Email = nullStringValue(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "UserDAO.FindByID: 查询失败 id=%d, err=%v", id, err)
		return nil, err
	}

	return user, nil
}

// FindByPhone 根据手机号查找用户
func (d *UserDAO) FindByPhone(ctx context.Context, phone string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, nickname, phone, email, role, status, created_at, updated_at 
	          FROM users WHERE phone = ?`

	var dbPhone, email sql.NullString
	err := d.mysql.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID, &user.Username, &user.Password, &user.Nickname, &dbPhone, &email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	user.Phone = nullStringValue(dbPhone)
	user.Email = nullStringValue(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "UserDAO.FindByPhone: 查询失败 phone=%s, err=%v", phone, err)
		return nil, err
	}

	return user, nil
}

// BindPhone 绑定手机号
func (d *UserDAO) BindPhone(ctx context.Context, userID int64, phone string) error {
	query := `UPDATE users SET phone = ?, updated_at = NOW() WHERE id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, phone, userID)
	if err != nil {
		log.Error(ctx, "UserDAO.BindPhone: 绑定失败 userID=%d, phone=%s, err=%v", userID, phone, err)
		return err
	}
	return nil
}

// UpdatePasswordByPhone 根据手机号更新密码
func (d *UserDAO) UpdatePasswordByPhone(ctx context.Context, phone, hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = NOW() WHERE phone = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, hashedPassword, phone)
	if err != nil {
		log.Error(ctx, "UserDAO.UpdatePasswordByPhone: 更新失败 phone=%s, err=%v", phone, err)
		return err
	}
	return nil
}

// FindByEmail 根据邮箱查询用户，不存在返回 nil, nil。
func (d *UserDAO) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, nickname, phone, email, role, status, created_at, updated_at 
	          FROM users WHERE email = ?`

	var phone, dbEmail sql.NullString
	err := d.mysql.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Nickname, &phone, &dbEmail,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Error(ctx, "UserDAO.FindByEmail: 查询失败 email=%s, err=%v", email, err)
		return nil, err
	}
	user.Phone = nullStringValue(phone)
	user.Email = nullStringValue(dbEmail)
	return user, nil
}

// BindEmail 绑定邮箱
func (d *UserDAO) BindEmail(ctx context.Context, userID int64, email string) error {
	query := `UPDATE users SET email = ?, updated_at = NOW() WHERE id = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, nullable(email), userID)
	if err != nil {
		log.Error(ctx, "UserDAO.BindEmail: 绑定失败 userID=%d, email=%s, err=%v", userID, email, err)
		return err
	}
	return nil
}

// UpdatePasswordByEmail 根据邮箱更新密码
func (d *UserDAO) UpdatePasswordByEmail(ctx context.Context, email, hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = NOW() WHERE email = ?`
	_, err := d.mysql.db.ExecContext(ctx, query, hashedPassword, email)
	if err != nil {
		log.Error(ctx, "UserDAO.UpdatePasswordByEmail: 更新失败 email=%s, err=%v", email, err)
		return err
	}
	return nil
}

// nullable 将空字符串转为 nil，用于可空字段写库。
func nullable(v string) interface{} {
	if v == "" {
		return nil
	}
	return v
}

// nullStringValue 将 SQL 可空字符串转换为普通 string。
func nullStringValue(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

// Update 更新用户信息
func (d *UserDAO) Update(ctx context.Context, user *User) error {
	query := `UPDATE users SET nickname = ?, phone = ?, email = ?, role = ?, status = ?, updated_at = NOW() 
	          WHERE id = ?`

	_, err := d.mysql.db.ExecContext(ctx, query,
		user.Nickname, nullable(user.Phone), nullable(user.Email), user.Role, user.Status, user.ID)
	if err != nil {
		log.Error(ctx, "UserDAO.Update: 更新失败 id=%d, err=%v", user.ID, err)
		return err
	}

	log.Info(ctx, "UserDAO.Update: 用户更新成功 id=%d", user.ID)
	return nil
}
