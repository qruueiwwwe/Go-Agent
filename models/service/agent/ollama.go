package agent

import (
	"context"
	"encoding/json"
	"strings"

	"agent/library/log"

	"github.com/ollama/ollama/api"
)

// OllamaService Ollama 服务
type OllamaService struct {
	client    *api.Client
	modelName string
}

func NewOllamaService(client *api.Client, modelName string) *OllamaService {
	return &OllamaService{
		client:    client,
		modelName: modelName,
	}
}

// Chat 与大模型对话
func (s *OllamaService) Chat(ctx context.Context, messages []api.Message) (string, error) {
	req := &api.ChatRequest{
		Model:    s.modelName,
		Messages: messages,
		Stream:   func(b bool) *bool { return &b }(false),
	}

	var fullResp string
	err := s.client.Chat(ctx, req, func(res api.ChatResponse) error {
		fullResp += res.Message.Content
		return nil
	})

	if err != nil {
		log.Error(ctx, "调用大模型失败: %v", err)
		return "", err
	}

	return fullResp, nil
}

// ParseToolCall 解析工具调用
func (s *OllamaService) ParseToolCall(response string) (toolName, toolInput string, isToolCall bool) {
	// 尝试提取 JSON 部分（可能有前后文本）
	// 查找第一个 { 和最后一个 }
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end < start {
		return "", "", false
	}

	jsonStr := response[start : end+1]
	var toolCall map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &toolCall); err == nil && toolCall["tool"] != "" {
		return toolCall["tool"], toolCall["input"], true
	}
	return "", "", false
}

// TrimSpace 去除空白字符
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}
