package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"agent/global"
	"agent/library/log"

	"github.com/ollama/ollama/api"
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
	MaxTokens   int            `json:"max_tokens,omitempty"`
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
// 支持多种格式：
// 1. JSON格式：{"tool":"calculator","input":"2^16 - 1"}
// 2. 换行分隔格式：calculator\n2^16 - 1（智谱API常用格式）
// 3. 包装JSON格式：{"input":"2**16-1"}（智谱API有时返回）
func ParseToolCallFromResponse(response string) (toolName, toolInput string, isToolCall bool) {
	// 尝试解析 JSON 格式
	start := indexOf(response, "{")
	end := lastIndexOf(response, "}")

	if start != -1 && end != -1 && end > start {
		jsonStr := response[start : end+1]
		var toolCall map[string]string
		if err := json.Unmarshal([]byte(jsonStr), &toolCall); err == nil {
			// 检查是否是标准格式 {"tool":"...","input":"..."}
			if toolCall["tool"] != "" {
				return toolCall["tool"], toolCall["input"], true
			}
			// 检查是否是包装格式 {"input":"..."}，需要从参数推断工具
			if toolCall["input"] != "" {
				// 尝试从响应的 JSON 之前部分找到工具名称
				prefix := response[:start]
				prefix = strings.TrimSpace(prefix)
				if prefix != "" {
					// 检查是否是有效的工具名称
					validTools := map[string]bool{
						"calculator": true,
						"weather":    true,
						"file":       true,
						"nbnhhsh":    true,
					}
					if validTools[prefix] {
						return prefix, toolCall["input"], true
					}
				}
				// 如果没有找到工具名，尝试从参数内容推断
				input := toolCall["input"]
				return inferToolFromInput(input), input, true
			}
		}
	}

	// 尝试解析换行分隔格式（智谱API常用）
	// 格式：toolName\ninput
	lines := strings.Split(response, "\n")
	if len(lines) > 0 {
		toolName = strings.TrimSpace(lines[0])
		// 检查是否是有效的工具名称
		validTools := map[string]bool{
			"calculator": true,
			"weather":    true,
			"file":       true,
			"nbnhhsh":    true,
		}

		if validTools[toolName] {
			// 获取工具参数（如果有）
			if len(lines) > 1 {
				// 将剩余行合并作为参数
				toolInput = strings.TrimSpace(strings.Join(lines[1:], "\n"))
			}
			return toolName, toolInput, true
		}
	}

	return "", "", false
}

// inferToolFromInput 从输入内容推断应该使用的工具
func inferToolFromInput(input string) string {
	inputLower := strings.ToLower(input)

	// 检查是否是数学表达式
	// 包含数字和运算符
	hasNumbers := false
	operators := []string{"+", "-", "*", "/", "^", "**"}
	for _, op := range operators {
		if strings.Contains(input, op) {
			hasNumbers = true
			break
		}
	}
	if hasNumbers {
		// 检查是否包含数字
		for _, c := range input {
			if c >= '0' && c <= '9' {
				hasNumbers = true
				break
			}
		}
	}
	if hasNumbers {
		return "calculator"
	}

	// 检查是否是天气查询
	weatherKeywords := []string{"天气", "温度", "度", "晴", "阴", "雨", "雪", "城市"}
	for _, kw := range weatherKeywords {
		if strings.Contains(inputLower, kw) {
			return "weather"
		}
	}

	// 默认返回 calculator（大部分情况是数学计算）
	return "calculator"
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

// ChatWithAPIMessages 使用 api.Message 格式与智谱大模型对话（适配器方法）
func (s *ZhipuService) ChatWithAPIMessages(ctx context.Context, messages []api.Message) (string, error) {
	zhipuMsgs := make([]zhipuMessage, len(messages))
	for i, m := range messages {
		zhipuMsgs[i] = zhipuMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}
	return s.Chat(ctx, zhipuMsgs)
}

// ChatWithModel 允许指定模型与 API Key（用于付费推理模型）
func (s *ZhipuService) ChatWithModel(ctx context.Context, messages []zhipuMessage, model, apiKey string, maxTokens int) (string, error) {
	if apiKey == "" {
		return "", ErrNoAPIKey
	}
	req := zhipuRequest{
		Model:       model,
		Messages:    messages,
		Temperature: s.config.Temperature,
		Stream:      false,
		MaxTokens:   maxTokens,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var zhipuResp zhipuResponse
	if err := json.NewDecoder(resp.Body).Decode(&zhipuResp); err != nil {
		return "", err
	}
	if zhipuResp.Error != nil {
		log.Error(ctx, "智谱API错误(model=%s): %s - %s", model, zhipuResp.Error.Code, zhipuResp.Error.Message)
		return "", ErrAPIError
	}
	if len(zhipuResp.Choices) == 0 {
		return "", ErrEmptyResponse
	}
	return zhipuResp.Choices[0].Message.Content, nil
}

// zhipuStreamFrame 流式响应帧
type zhipuStreamFrame struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Role             string `json:"role"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// SelectModelAndKey 根据是否需要 reasoning 选择模型与 API Key
// 若付费 key 可用且 preferReasoning=true → 返回付费模型与 NewAPIKey
// 否则返回免费模型与 APIKey
func (s *ZhipuService) SelectModelAndKey(preferReasoning bool) (model, apiKey string, isReasoning bool) {
	if preferReasoning && s.config.IsReasoningEnabled() {
		return s.config.ReasoningModel, s.config.NewAPIKey, true
	}
	return s.config.Model, s.config.APIKey, false
}

// ChatStream 流式对话
// tokenCh 接收 delta.content（正文答案）
// reasoningCh 接收 delta.reasoning_content（付费模型原生思考链，可能为 nil）
// 使用 SelectModelAndKey 时传入的 model/apiKey
func (s *ZhipuService) ChatStream(ctx context.Context, messages []zhipuMessage, model, apiKey string, tokenCh chan<- string, reasoningCh chan<- string) error {
	if apiKey == "" {
		return ErrNoAPIKey
	}

	req := zhipuRequest{
		Model:       model,
		Messages:    messages,
		Temperature: s.config.Temperature,
		Stream:      true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		log.Error(ctx, "ChatStream: 请求失败 err=%v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := make([]byte, 512)
		n, _ := resp.Body.Read(buf)
		log.Error(ctx, "ChatStream: 状态码=%d body=%s", resp.StatusCode, string(buf[:n]))
		return errors.New("zhipu stream http status " + resp.Status)
	}

	reader := bufio.NewReaderSize(resp.Body, 8192)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		// SSE 帧格式：data: {...}
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		payload := bytes.TrimSpace(line[len("data:"):])
		if len(payload) == 0 {
			continue
		}
		if bytes.Equal(payload, []byte("[DONE]")) {
			return nil
		}

		var frame zhipuStreamFrame
		if err := json.Unmarshal(payload, &frame); err != nil {
			log.Warn(ctx, "ChatStream: 解析帧失败 err=%v payload=%s", err, string(payload))
			continue
		}
		if frame.Error != nil {
			return errors.New(frame.Error.Message)
		}
		for _, ch := range frame.Choices {
			if ch.Delta.ReasoningContent != "" && reasoningCh != nil {
				reasoningCh <- ch.Delta.ReasoningContent
			}
			if ch.Delta.Content != "" && tokenCh != nil {
				tokenCh <- ch.Delta.Content
			}
			// finish_reason 出现时不主动 break，等待 [DONE] 或 EOF
			_ = ch.FinishReason
		}
	}
}

// _ 保留 strings 依赖占位
var _ = strings.TrimSpace
