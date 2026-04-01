package agent

import (
	"context"
	"strings"

	"agent/library/log"

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

【必须遵守的规则】
1. 当用户问天气时：必须使用 weather 工具，禁止用知识库回答
2. 当用户问计算时：必须使用 calculator 工具
3. 当用户问文件处理时：必须使用 file 工具
4. 你的回复必须以 JSON 格式开始：{"tool":"工具名","input":"参数"}
5. 不能说"我无法访问"或"建议查阅"，你有工具就必须使用

【调用工具时】
你必须输出 JSON，格式如下：
{"tool":"weather","input":"城市"}

【工具返回后】
- 直接把工具返回的原始结果返回给用户
- 禁止做任何修改、总结或解释
- 原样返回结果
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

	// 降级策略：如果大模型没有输出 JSON，则根据用户消息关键词判断是否应该调用工具
	toolName, toolInput, shouldCallTool := s.inferToolCall(userMessage)
	if shouldCallTool {
		log.Info(ctx, "使用降级策略调用工具: %s, 参数: %s", toolName, toolInput)
		result := s.toolManager.Execute(ctx, toolName, toolInput)
		log.Info(ctx, "工具执行结果: %s", result)
		return result
	}

	return resp
}

// inferToolCall 根据用户消息推断是否需要调用工具
func (s *AgentService) inferToolCall(userMessage string) (toolName, toolInput string, shouldCall bool) {
	// 天气关键词
	weatherKeywords := []string{"天气", "几度", "温度", "气温", "下雨", "下雪", "晴天", "阴天", "明天", "后天", "周", "天"}
	for _, keyword := range weatherKeywords {
		if strings.Contains(userMessage, keyword) {
			return "weather", userMessage, true
		}
	}

	// 计算关键词
	calcKeywords := []string{"加", "减", "乘", "除", "×", "÷", "等于", "多少"}
	for _, keyword := range calcKeywords {
		if strings.Contains(userMessage, keyword) {
			// 但要排除"几度"这样的天气问题
			if !strings.Contains(userMessage, "度") || strings.Contains(userMessage, "几度") {
				return "calculator", userMessage, true
			}
		}
	}

	return "", "", false
}
