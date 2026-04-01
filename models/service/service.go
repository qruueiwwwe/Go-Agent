package service

import (
	"context"

	"agent/global"
	"agent/log"
	"agent/models/dao"
)

// UserService 用户服务层（暂未实现）
type UserService struct {
	userDAO *dao.UserDAO
}

// NewUserService 创建UserService实例
func NewUserService() *UserService {
	return &UserService{
		userDAO: dao.NewUserDAO(),
	}
}

// Init 初始化服务
func (s *UserService) Init(ctx context.Context, cfg global.DatabaseConfig) error {
	log.Info(ctx, "UserService 初始化（占位）")
	return s.userDAO.Init(ctx, cfg)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(ctx context.Context, id int64) (interface{}, error) {
	log.Debug(ctx, "UserService GetUserByID: id=%d", id)
	return s.userDAO.QueryByID(ctx, id)
}

// GetUserByName 根据名称获取用户
func (s *UserService) GetUserByName(ctx context.Context, name string) (interface{}, error) {
	log.Debug(ctx, "UserService GetUserByName: name=%s", name)
	return s.userDAO.QueryByName(ctx, name)
}

// ChatHistoryService 聊天历史服务层（暂未实现）
type ChatHistoryService struct {
	chatDAO *dao.ChatHistoryDAO
}

func NewChatHistoryService() *ChatHistoryService {
	return &ChatHistoryService{
		chatDAO: dao.NewChatHistoryDAO(),
	}
}

func (s *ChatHistoryService) Init(ctx context.Context, cfg global.DatabaseConfig) error {
	log.Info(ctx, "ChatHistoryService 初始化（占位）")
	return s.chatDAO.Init(ctx, cfg)
}

func (s *ChatHistoryService) SaveChat(ctx context.Context, data interface{}) error {
	log.Debug(ctx, "ChatHistoryService SaveChat")
	return s.chatDAO.Save(ctx, data)
}

func (s *ChatHistoryService) GetUserChats(ctx context.Context, userID int64) ([]interface{}, error) {
	log.Debug(ctx, "ChatHistoryService GetUserChats: userID=%d", userID)
	return s.chatDAO.QueryByUserID(ctx, userID)
}