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

【工具使用规则】
1. 天气查询：用户问实时天气时，使用 weather 工具获取准确数据
2. 数学计算：用户需要计算时，使用 calculator 工具
3. 文件操作：用户需要解析/分析文件时，使用 file 工具
4. 缩写词猜测：用户问字母缩写/网络用语的含义时，使用 nbnhhsh 工具

【重要】
- 先判断用户意图，再决定是否需要工具
- 如果用知识库能回答的问题，直接回答，不必调用工具
- 只有需要实时数据或外部资源时才调用工具

【调用工具时的格式】
{"tool":"工具名","input":"参数"}

示例：
- 天气：{"tool":"weather","input":"北京"}
- 计算：{"tool":"calculator","input":"1+2"}
- 缩写词：{"tool":"nbnhhsh","input":"yyds"}

【工具返回后】
- 直接把工具返回的原始结果返回给用户
- 原样返回结果，不做修改
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

	// 强制工具调用检测：当大模型未遵循指令时
	if forceToolName, forceToolInput, shouldForce := s.shouldForceToolCall(userMessage, resp); shouldForce {
		log.Info(ctx, "强制调用工具（大模型未遵循指令）: %s, 参数: %s", forceToolName, forceToolInput)
		result := s.toolManager.Execute(ctx, forceToolName, forceToolInput)
		log.Info(ctx, "工具执行结果: %s", result)
		return result
	}

	// 降级策略：只有在大模型响应质量差的情况下，才根据关键词调用工具
	// 检查大模型的响应是否表明无法处理
	if s.shouldUseFallbackStrategy(resp) {
		toolName, toolInput, shouldCallTool := s.inferToolCall(userMessage)
		if shouldCallTool {
			log.Info(ctx, "使用降级策略调用工具: %s, 参数: %s", toolName, toolInput)
			result := s.toolManager.Execute(ctx, toolName, toolInput)
			log.Info(ctx, "工具执行结果: %s", result)
			return result
		}
	}

	return resp
}

// shouldForceToolCall 检测是否应该强制调用工具
// 当用户问天气或缩写词含义但大模型没有调用工具却返回了答案时，强制调用工具
func (s *AgentService) shouldForceToolCall(userMessage, resp string) (toolName, toolInput string, shouldCall bool) {
	// ========== 天气工具检测 ==========
	weatherKeywords := []string{"天气", "几度", "温度", "气温", "下雨", "下雪", "晴", "阴", "多云"}
	hasWeatherKeyword := false
	for _, kw := range weatherKeywords {
		if strings.Contains(userMessage, kw) {
			hasWeatherKeyword = true
			break
		}
	}
	if hasWeatherKeyword {
		weatherResponsePatterns := []string{
			"天气", "气温", "温度", "晴", "阴", "多云", "雨", "雪",
			"最高温度", "最低温度", "°C", "摄氏",
		}
		for _, pattern := range weatherResponsePatterns {
			if strings.Contains(resp, pattern) {
				city := extractCityFromMessage(userMessage)
				return "weather", city, true
			}
		}
	}

	// ========== 缩写词猜测工具检测 ==========
	// 检测用户是否在问缩写词含义
	if strings.Contains(userMessage, "是什么意思") || strings.Contains(userMessage, "意思") ||
		strings.Contains(userMessage, "含义") || strings.Contains(userMessage, "指的是") {
		// 检测大模型是否返回了类似缩写词解释的格式但没有调用工具
		// 例如：【xxx】可能的含义：1. xxx
		if strings.Contains(resp, "可能的含义") || strings.Contains(resp, "含义：") ||
			strings.Contains(resp, "意思是") || strings.Contains(resp, "指的是") {
			// 提取缩写词
			abbr := extractAbbreviation(userMessage)
			if abbr != "" {
				return "nbnhhsh", abbr, true
			}
		}
	}

	// 检测纯字母/数字组合（可能是缩写词）
	abbr := extractAbbreviation(userMessage)
	if abbr != "" && len(abbr) <= 10 {
		// 如果大模型返回了类似解释但没有调用工具
		if strings.Contains(resp, "可能的含义") || strings.Contains(resp, "含义：") ||
			strings.Contains(resp, "意思是") || (strings.Contains(resp, "【") && strings.Contains(resp, "】")) {
			return "nbnhhsh", abbr, true
		}
	}

	return "", "", false
}

// extractAbbreviation 从用户消息中提取缩写词
func extractAbbreviation(msg string) string {
	// 移除常见的问题后缀
	replacements := []string{
		"是什么意思", "是什么含义", "是什么",
		"的意思", "的含义", "意思", "含义",
		"指的是什么", "指的是",
		"？", "?", "。", "，",
		"请问", "告诉我", "查一下",
	}
	result := msg
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r, "")
	}
	result = strings.TrimSpace(result)

	// 检查是否是纯字母/数字组合（可含符号）
	valid := true
	letterCount := 0
	for _, c := range result {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			letterCount++
		} else if c != '\'' && c != '-' && c != '_' && c != ' ' && c != ',' {
			valid = false
			break
		}
	}

	// 至少要有2个字母/数字才算有效缩写词
	if !valid || letterCount < 2 {
		return ""
	}

	// 移除空格和逗号
	result = strings.ReplaceAll(result, " ", "")
	result = strings.ReplaceAll(result, ",", "")

	if len(result) > 20 {
		return ""
	}

	return result
}

// extractCityFromMessage 从用户消息中提取城市名
func extractCityFromMessage(msg string) string {
	// 移除常见的关键词，提取城市名
	replacements := []string{
		"今天", "明天", "后天", "大后天",
		"的天气怎么样", "天气怎么样", "的天气", "天气",
		"温度", "气温", "几度",
		"怎么样", "如何", "呢", "吗", "？", "?",
		"请问", "查一下", "告诉我",
	}
	result := msg
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r, "")
	}
	result = strings.TrimSpace(result)

	// 如果结果为空或太长，返回原始消息让工具自己处理
	if result == "" || len(result) > 20 {
		return msg
	}

	return result
}

// shouldUseFallbackStrategy 判断是否应该使用降级策略
// 只有当大模型响应表明无法处理时，才使用降级策略
func (s *AgentService) shouldUseFallbackStrategy(resp string) bool {
	// 如果响应为空，使用降级策略
	if strings.TrimSpace(resp) == "" {
		return true
	}

	// 如果包含这些词汇，说明大模型无法处理，使用降级策略
	negativeKeywords := []string{
		"无法",
		"不能",
		"无法访问",
		"无法获取",
		"无法查询",
		"不支持",
		"稍后重试",
		"请稍后",
		"暂时不可用",
		"无法直接",
		"我无法",
		"我不能",
	}

	resp = strings.ToLower(resp)
	for _, keyword := range negativeKeywords {
		if strings.Contains(resp, strings.ToLower(keyword)) {
			return true
		}
	}

	// 否则大模型已经给出了有意义的答案，不使用降级策略
	return false
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

	// 注意：不添加计算工具的降级策略
	// 原因：大模型可以直接进行数学计算，而计算工具只支持 "3762+57778*6/333" 这样的数学表达式格式
	// 如果用户用自然语言（如"三千七百六十二"）表达数字，直接让大模型处理更好
	// 只有在大模型无法处理时，用户才需要用标准数学表达式格式

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
