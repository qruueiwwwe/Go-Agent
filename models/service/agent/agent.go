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
3. 当用户涉及文件操作（解析、总结、分析、查看文件等）：必须使用 file 工具
4. 你的回复必须以 JSON 格式开始：{"tool":"工具名","input":"参数"}
5. 不能说"我无法访问"或"建议查阅"，你有工具就必须使用它

【调用工具时的格式】
- 天气：{"tool":"weather","input":"城市"}
- 计算：{"tool":"calculator","input":"1+2"}
- 文件：{"tool":"file","input":{"action":"parse","file":"文件路径","mode":"summary"}}

【工具返回后】
- 直接把工具返回的原始结果返回给用户
- 禁止做任何修改、总结或解释
- 原样返回结果

【重点】
- 用户提到文件、文档、数据、代码时，必须使用 file 工具
- 不能假装无法访问文件，因为你有 file 工具
- 确保 action 字段填写正确：parse(解析)/code_analyze(代码分析)/convert(格式转换)
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
	// 文件工具关键词（优先检查）
	fileKeywords := []string{"文件", "解析", "总结", "分析", "阅读", "读取", "查看", ".txt", ".md", ".json", ".py", ".go", ".js", ".pdf", "data/"}
	for _, keyword := range fileKeywords {
		if strings.Contains(userMessage, keyword) {
			// 尝试提取文件名
			fileName := extractFileName(userMessage)

			// 判断操作类型
			action := "parse" // 默认
			mode := "summary"
			if strings.Contains(userMessage, "分析代码") || strings.Contains(userMessage, "代码分析") {
				action = "code_analyze"
			} else if strings.Contains(userMessage, "转换") || strings.Contains(userMessage, "格式") {
				action = "convert"
			}

			// 如果没有提取到文件名，使用原消息作为参数
			if fileName == "" {
				fileName = userMessage
			}

			// 构造 file 工具的 JSON 输入
			if action == "parse" {
				toolInput = `{"action":"parse","file":"` + fileName + `","mode":"` + mode + `"}`
			} else if action == "code_analyze" {
				toolInput = `{"action":"code_analyze","file":"` + fileName + `","type":"explain"}`
			} else {
				toolInput = `{"action":"convert","file":"` + fileName + `","target":"html"}`
			}

			return "file", toolInput, true
		}
	}

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
			if !strings.Contains(userMessage, "度") || strings.Contains(userMessage, "几度") {
				return "calculator", userMessage, true
			}
		}
	}

	return "", "", false
}

// extractFileName 从用户消息中提取文件名
func extractFileName(userMessage string) string {
	// 查找 data/ 开头的路径
	if idx := strings.Index(userMessage, "data/"); idx != -1 {
		// 提取从 data/ 开始到空格或特殊符号为止的文件路径
		end := len(userMessage)
		for i := idx; i < len(userMessage); i++ {
			c := userMessage[i : i+1]
			if c == " " || c == "," || strings.HasPrefix(userMessage[i:], "。") || strings.HasPrefix(userMessage[i:], "，") {
				end = i
				break
			}
		}
		return userMessage[idx:end]
	}

	// 查找常见的文件扩展名
	extensions := []string{".txt", ".md", ".json", ".py", ".go", ".js", ".pdf"}
	for _, ext := range extensions {
		if idx := strings.LastIndex(userMessage, ext); idx != -1 {
			// 从这个位置往前查找文件名的开始
			start := idx
			for start > 0 {
				c := userMessage[start-1 : start]
				if c == " " || c == "/" {
					break
				}
				start--
			}
			// 如果前面有 data/ 那么包含它
			if start > 0 && strings.Contains(userMessage[:idx], "data/") {
				dataIdx := strings.LastIndex(userMessage[:idx], "data/")
				return userMessage[dataIdx : idx+len(ext)]
			}
			// 否则检查是否以 / 开头（相对路径）
			if start > 0 && userMessage[start-1:start] == "/" {
				return userMessage[start-1 : idx+len(ext)]
			}
			return userMessage[start : idx+len(ext)]
		}
	}

	return ""
}
