package dao

import (
	"context"

	"agent/global"
	"agent/log"
)

// UserDAO 用户数据访问层（暂未实现）
type UserDAO struct{}

// NewUserDAO 创建UserDAO实例
func NewUserDAO() *UserDAO {
	return &UserDAO{}
}

// Init 初始化DAO
func (d *UserDAO) Init(ctx context.Context, cfg global.DatabaseConfig) error {
	log.Info(ctx, "UserDAO 初始化（占位）")
	return nil
}

// QueryByID 根据ID查询用户
func (d *UserDAO) QueryByID(ctx context.Context, id int64) (interface{}, error) {
	log.Debug(ctx, "UserDAO QueryByID: id=%d", id)
	return nil, nil
}

// QueryByName 根据名称查询用户
func (d *UserDAO) QueryByName(ctx context.Context, name string) (interface{}, error) {
	log.Debug(ctx, "UserDAO QueryByName: name=%s", name)
	return nil, nil
}

// ChatHistoryDAO 聊天历史数据访问层（暂未实现）
type ChatHistoryDAO struct{}

func NewChatHistoryDAO() *ChatHistoryDAO {
	return &ChatHistoryDAO{}
}

func (d *ChatHistoryDAO) Init(ctx context.Context, cfg global.DatabaseConfig) error {
	log.Info(ctx, "ChatHistoryDAO 初始化（占位）")
	return nil
}

func (d *ChatHistoryDAO) Save(ctx context.Context, data interface{}) error {
	log.Debug(ctx, "ChatHistoryDAO Save: data=%v", data)
	return nil
}

func (d *ChatHistoryDAO) QueryByUserID(ctx context.Context, userID int64) ([]interface{}, error) {
	log.Debug(ctx, "ChatHistoryDAO QueryByUserID: userID=%d", userID)
	return nil, nil
}