package agent

import "errors"

var (
	// ErrNoAPIKey API密钥未配置
	ErrNoAPIKey = errors.New("智谱API密钥未配置")
	// ErrAPIError API调用错误
	ErrAPIError = errors.New("智谱API调用失败")
	// ErrEmptyResponse API返回空响应
	ErrEmptyResponse = errors.New("智谱API返回空响应")
)
