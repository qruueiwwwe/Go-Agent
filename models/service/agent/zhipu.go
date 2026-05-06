package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"agent/global"
	"agent/library/log"
)

// ZhipuService 智谱清言服务
type ZhipuService struct {
	config     global.ZhipuConfig
	httpClient *http.Client
}

// zhipuRequest 智谱API请求结构
type zhipuRequest struct {
	Model       string         `json:"model"`
	Messages    []zhipuMessage `json:"messages"`
	Temperature float64        `json:"temperature"`
	Stream      bool           `json:"stream"`
}

// zhipuMessage 智谱消息结构
type zhipuMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// zhipuResponse 智谱API响应结构
type zhipuResponse struct {
	Choices []struct {
		Message zhipuMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// NewZhipuService 创建智谱服务
func NewZhipuService(cfg global.ZhipuConfig) *ZhipuService {
	return &ZhipuService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Chat 与智谱大模型对话
func (s *ZhipuService) Chat(ctx context.Context, messages []zhipuMessage) (string, error) {
	if s.config.APIKey == "" {
		return "", ErrNoAPIKey
	}

	req := zhipuRequest{
		Model:       s.config.Model,
		Messages:    messages,
		Temperature: s.config.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Error(ctx, "序列化请求失败: %v", err)
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		log.Error(ctx, "创建请求失败: %v", err)
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		log.Error(ctx, "请求智谱API失败: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	var zhipuResp zhipuResponse
	if err := json.NewDecoder(resp.Body).Decode(&zhipuResp); err != nil {
		log.Error(ctx, "解析响应失败: %v", err)
		return "", err
	}

	if zhipuResp.Error != nil {
		log.Error(ctx, "智谱API错误: %s - %s", zhipuResp.Error.Code, zhipuResp.Error.Message)
		return "", ErrAPIError
	}

	if len(zhipuResp.Choices) == 0 {
		log.Error(ctx, "智谱API返回空响应")
		return "", ErrEmptyResponse
	}

	content := zhipuResp.Choices[0].Message.Content
	log.Info(ctx, "智谱API响应成功, 长度: %d", len(content))
	return content, nil
}

// ParseToolCall 解析工具调用（与OllamaService保持一致）
func (s *ZhipuService) ParseToolCall(response string) (toolName, toolInput string, isToolCall bool) {
	return ParseToolCallFromResponse(response)
}

// ParseToolCallFromResponse 从响应中解析工具调用
func ParseToolCallFromResponse(response string) (toolName, toolInput string, isToolCall bool) {
	start := indexOf(response, "{")
	end := lastIndexOf(response, "}")

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

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
