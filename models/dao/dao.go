package dao

import (
	"context"

	"agent/global"
	"agent/library/log"
)

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
