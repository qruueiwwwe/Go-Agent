package agent

import (
	"agent/global"
	"context"
	"testing"
	"time"

	"github.com/ollama/ollama/api"
)

func TestZhipuService_NewZhipuService(t *testing.T) {
	cfg := global.ZhipuConfig{
		APIKey:      "test-key",
		Model:       "glm-4-flash",
		BaseURL:     "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		Timeout:     60 * time.Second,
		Temperature: 0.3,
		Enable:      true,
	}

	svc := NewZhipuService(cfg)
	if svc == nil {
		t.Error("NewZhipuService() should not return nil")
	}
	if svc.config.APIKey != "test-key" {
		t.Errorf("Expected APIKey 'test-key', got '%s'", svc.config.APIKey)
	}
}

func TestZhipuService_Chat_NoAPIKey(t *testing.T) {
	cfg := global.ZhipuConfig{
		APIKey:  "", // Empty API key
		Model:   "glm-4-flash",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		Timeout: 5 * time.Second,
	}

	svc := NewZhipuService(cfg)
	ctx := context.Background()

	_, err := svc.Chat(ctx, []zhipuMessage{
		{Role: "user", Content: "hello"},
	})

	if err != ErrNoAPIKey {
		t.Errorf("Expected ErrNoAPIKey, got %v", err)
	}
}

func TestParseToolCallFromResponse(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectedTool  string
		expectedInput string
		expectedBool  bool
	}{
		{
			name:          "Valid tool call",
			response:      `{"tool":"weather","input":"北京"}`,
			expectedTool:  "weather",
			expectedInput: "北京",
			expectedBool:  true,
		},
		{
			name:          "Tool call with prefix text",
			response:      `好的，我来查询天气 {"tool":"weather","input":"上海"}`,
			expectedTool:  "weather",
			expectedInput: "上海",
			expectedBool:  true,
		},
		{
			name:          "No tool call",
			response:      `今天天气不错`,
			expectedTool:  "",
			expectedInput: "",
			expectedBool:  false,
		},
		{
			name:          "Invalid JSON",
			response:      `{tool:weather}`,
			expectedTool:  "",
			expectedInput: "",
			expectedBool:  false,
		},
		{
			name:          "Empty response",
			response:      ``,
			expectedTool:  "",
			expectedInput: "",
			expectedBool:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, input, isToolCall := ParseToolCallFromResponse(tt.response)
			if tool != tt.expectedTool {
				t.Errorf("Expected tool '%s', got '%s'", tt.expectedTool, tool)
			}
			if input != tt.expectedInput {
				t.Errorf("Expected input '%s', got '%s'", tt.expectedInput, input)
			}
			if isToolCall != tt.expectedBool {
				t.Errorf("Expected isToolCall %v, got %v", tt.expectedBool, isToolCall)
			}
		})
	}
}

func TestConvertToZhipuMessages(t *testing.T) {
	msgs := []api.Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	result := convertToZhipuMessages(msgs)

	if len(result) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(result))
	}

	if result[0].Role != "system" || result[0].Content != "You are a helpful assistant" {
		t.Errorf("First message not converted correctly")
	}

	if result[1].Role != "user" || result[1].Content != "Hello" {
		t.Errorf("Second message not converted correctly")
	}
}
