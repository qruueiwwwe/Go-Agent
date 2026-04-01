package agent

import (
	"context"

	log "agent/library/log/logger"

	"github.com/ollama/ollama/api"
)

// AgentService Agent 服务
type AgentService struct {
	ollamaSvc    *OllamaService
	toolManager  *ToolManager
	systemPrompt string
}

// NewAgentService 创建 Agent 服务
func NewAgentService(ollamaSvc *OllamaService, toolManager *ToolManager) *AgentService {
	systemPrompt := `
你是一个智能助手，能使用工具。
你有以下工具：
` + toolManager.GetToolsDesc() + `
规则：
1. 如果需要使用工具，必须调用工具，不要自己算或编造
2. 当需要使用工具时，你必须严格按照以下 JSON 格式输出：
   {"tool":"工具名","input":"参数"}
3. 不需要工具时，直接回答用户问题
4. 记住对话历史

【重要】当工具返回结果后，你必须：
- 直接把工具返回的原始结果返回给用户
- 禁止做任何修改、总结、解释或添加任何文字
- 即使工具返回的结果不完整，也要原样返回
- 例如：如果工具返回"北京：晴天，最高20°C"，你必须返回"北京：晴天，最高20°C"，不能说"北京天气是晴天，最高温度20度"
`
	return &AgentService{
		ollamaSvc:    ollamaSvc,
		toolManager:  toolManager,
		systemPrompt: systemPrompt,
	}
}

// Process 处理用户消息
func (s *AgentService) Process(ctx context.Context, userMessage string, history []api.Message) string {
	// 构建消息列表
	msgs := []api.Message{
		{Role: "system", Content: s.systemPrompt},
	}

	// 添加历史消息
	for _, m := range history {
		msgs = append(msgs, m)
	}

	// 添加当前用户消息
	msgs = append(msgs, api.Message{Role: "user", Content: userMessage})

	log.Info(ctx, "开始处理用户消息: %s", userMessage)

	// 调用大模型
	resp, err := s.ollamaSvc.Chat(ctx, msgs)
	if err != nil {
		log.Error(ctx, "调用大模型失败: %v", err)
		return "服务暂时不可用，请稍后重试"
	}

	log.Info(ctx, "大模型响应: %s", resp)

	// 检查是否需要调用工具
	toolName, toolInput, isToolCall := s.ollamaSvc.ParseToolCall(resp)
	if isToolCall {
		log.Info(ctx, "调用工具: %s, 参数: %s", toolName, toolInput)

		// 执行工具
		result := s.toolManager.Execute(ctx, toolName, toolInput)
		log.Info(ctx, "工具执行结果: %s", result)

		// 直接返回工具结果
		return result
	}

	return resp
}
