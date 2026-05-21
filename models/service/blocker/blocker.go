package blocker

import (
	"strings"
)

// Service 屏蔽词服务
type Service struct {
	blockedWords []string
}

// NewService 创建屏蔽词服务
func NewService(blockedWords []string) *Service {
	// 转为小写，方便匹配
	words := make([]string, len(blockedWords))
	for i, w := range blockedWords {
		words[i] = strings.ToLower(strings.TrimSpace(w))
	}
	return &Service{blockedWords: words}
}

// CheckText 检查文本是否包含屏蔽词
// 返回: 是否包含屏蔽词, 匹配到的屏蔽词
func (s *Service) CheckText(text string) (bool, string) {
	textLower := strings.ToLower(text)
	for _, word := range s.blockedWords {
		if word != "" && strings.Contains(textLower, word) {
			return true, word
		}
	}
	return false, ""
}

// CheckMultiple 检查多个文本字段
// 返回: 是否包含屏蔽词, 匹配到的屏蔽词, 字段名
func (s *Service) CheckMultiple(fields map[string]string) (bool, string, string) {
	for fieldName, text := range fields {
		if found, word := s.CheckText(text); found {
			return true, word, fieldName
		}
	}
	return false, "", ""
}

// GetBlockedWords 获取屏蔽词列表
func (s *Service) GetBlockedWords() []string {
	return s.blockedWords
}
