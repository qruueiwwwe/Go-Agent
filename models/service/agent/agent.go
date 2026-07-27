package agent

import (
	"context"
	"fmt"
	"strings"

	"agent/library/log"

	"github.com/ollama/ollama/api"
)

// AgentService Agent 服务
type AgentService struct {
	ollamaSvc    *OllamaService
	zhipuSvc     *ZhipuService
	toolManager  *ToolManager
	systemPrompt string
	maxChars     int
	maxMsgs      int
}

// deepThinkingPromptSuffix 深度思考模式（免费模型）注入的提示
// 要求模型输出 【思考】...【回答】... 结构，供后端流式解析
const deepThinkingPromptSuffix = `

【深度思考模式（严格遵守）】
你必须严格按下面的两段式输出，两个标签必须成对出现，且严格使用中文全角括号"【】"：

【思考】
在此处逐步分析用户问题：
- 拆解意图
- 判断是否需要调用工具（如天气、缩写词、计算、文件解析等实时/外部信息一律必须调用工具，不得凭空作答）
- 若需要工具，说明为什么，并想好工具名与 input 参数

【回答】
- 若需要调用工具：本段**只能**输出严格的 JSON，形如 {"tool":"工具名","input":"参数"}，不得包含任何其他文字、Markdown、代码块、解释或标签。
- 若无需工具：本段直接给出面向用户的自然语言最终答复。

【严禁事项】
- 严禁编造实时数据（例如天气、股价、时间）。任何实时或外部信息必须走工具。
- 【回答】段之后不得再输出任何内容。
- 不得省略或改写"【思考】"与"【回答】"这两个标签。
`

// NewAgentService 创建 Agent 服务
func NewAgentService(ollamaSvc *OllamaService, zhipuSvc *ZhipuService, toolManager *ToolManager, budgetMode string) *AgentService {
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

	【天气任务专用规则（强制）】
	1. 涉及天气比较（如“北京和西安相比”“哪个更热”）时，必须先拆分城市，再分别调用 weather 查询。
	2. weather 工具 input 只允许“单城市名 + 可选时间词”（如“北京今天”“西安明天”），禁止传整句自然语言。
	3. 禁止把比较词或问句词传给 weather：和、相比、怎么样、哪个、哪个更、对比。
	4. 多城市比较时，先分别查询每个城市天气，再输出结构化对比结论。
	5. 城市不明确时先追问澄清，不要盲目调用 weather。

	【调用工具时的格式】
	{"tool":"工具名","input":"参数"}

	示例：
	- 天气：{"tool":"weather","input":"北京"}
	- 计算：{"tool":"calculator","input":"1+2"}
	- 缩写词：{"tool":"nbnhhsh","input":"yyds"}
	- 天气比较（今天北京和西安相比怎么样）：
	  1) {"tool":"weather","input":"北京今天"}
	  2) {"tool":"weather","input":"西安今天"}
	  3) 根据两次结果输出对比结论
	- 天气比较（上海和杭州明天哪个更热）：
	  1) {"tool":"weather","input":"上海明天"}
	  2) {"tool":"weather","input":"杭州明天"}
	- 城市不明确（“今天和明天哪个城市更冷”）：先追问城市，不调用 weather

	【工具返回后】
	- 单工具场景：忠实返回工具事实，不编造数据。
	- 多次工具调用场景（如天气对比）：允许基于工具结果做结构化总结（天气/温度/风力/湿度）。
`
	maxChars, maxMsgs := getContextBudget(budgetMode)
	return &AgentService{
		ollamaSvc:    ollamaSvc,
		zhipuSvc:     zhipuSvc,
		toolManager:  toolManager,
		systemPrompt: systemPrompt,
		maxChars:     maxChars,
		maxMsgs:      maxMsgs,
	}
}

// Process 处理用户消息
func (s *AgentService) Process(ctx context.Context, userMessage string, history []api.Message) string {
	// 构建消息列表
	msgs := []api.Message{
		{Role: "system", Content: s.systemPrompt},
	}

	trimmedHistory := s.trimHistory(history, userMessage)
	for _, m := range trimmedHistory {
		msgs = append(msgs, m)
	}

	msgs = append(msgs, api.Message{Role: "user", Content: userMessage})

	log.Info(ctx, "开始处理用户消息: %s", userMessage)

	// 调用大模型：优先智谱，失败降级 Ollama
	var resp string
	var err error

	if s.zhipuSvc != nil {
		resp, err = s.zhipuSvc.Chat(ctx, convertToZhipuMessages(msgs))
		if err != nil {
			log.Warn(ctx, "智谱调用失败: %v，尝试 Ollama 兜底", err)
		} else if resp != "" {
			log.Info(ctx, "智谱清言响应成功")
		}
	}

	if resp == "" && s.ollamaSvc != nil {
		resp, err = s.ollamaSvc.Chat(ctx, msgs)
		if err != nil {
			log.Error(ctx, "Ollama 调用失败: %v", err)
		}
	}

	if resp == "" {
		log.Error(ctx, "智谱和 Ollama 均不可用")
		return "服务暂时不可用，请稍后重试"
	}

	log.Info(ctx, "大模型响应: %s", resp)

	// 检查是否需要调用工具（优先使用智谱解析，因为它可能更准确）
	var toolName, toolInput string
	var isToolCall bool
	if s.zhipuSvc != nil {
		toolName, toolInput, isToolCall = s.zhipuSvc.ParseToolCall(resp)
	} else if s.ollamaSvc != nil {
		toolName, toolInput, isToolCall = s.ollamaSvc.ParseToolCall(resp)
	}
	if !isToolCall {
		toolName, toolInput, isToolCall = ParseToolCallFromResponse(resp)
	}

	if isToolCall {
		log.Info(ctx, "调用工具: %s, 参数: %s", toolName, toolInput)
		result := s.executeToolAndSummarize(ctx, userMessage, toolName, toolInput)
		log.Info(ctx, "工具执行结果: %s", result)
		return result
	}

	// 强制工具调用检测：当大模型未遵循指令时
	if forceToolName, forceToolInput, shouldForce := s.shouldForceToolCall(userMessage, resp); shouldForce {
		log.Info(ctx, "强制调用工具（大模型未遵循指令）: %s, 参数: %s", forceToolName, forceToolInput)
		result := s.executeToolAndSummarize(ctx, userMessage, forceToolName, forceToolInput)
		log.Info(ctx, "工具执行结果: %s", result)
		return result
	}

	// 降级策略：只有在大模型响应质量差的情况下，才根据关键词调用工具
	// 检查大模型的响应是否表明无法处理
	if s.shouldUseFallbackStrategy(resp) {
		toolName, toolInput, shouldCallTool := s.inferToolCall(userMessage)
		if shouldCallTool {
			log.Info(ctx, "使用降级策略调用工具: %s, 参数: %s", toolName, toolInput)
			result := s.executeToolAndSummarize(ctx, userMessage, toolName, toolInput)
			log.Info(ctx, "工具执行结果: %s", result)
			return result
		}
	}

	return resp
}

// shouldForceToolCall 检测是否应该强制调用工具
// 当用户问天气或缩写词含义但大模型没有调用工具却返回了答案时，强制调用工具
func (s *AgentService) shouldForceToolCall(userMessage, resp string) (toolName, toolInput string, shouldCall bool) {
	if _, _, ok := ParseToolCallFromResponse(resp); ok {
		return "", "", false
	}

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
				return "weather", userMessage, true
			}
		}
	}

	// ========== 缩写词猜测工具检测 ==========
	// 检测用户是否在问缩写词含义
	if strings.Contains(userMessage, "是什么意思") || strings.Contains(userMessage, "是什么") ||
		strings.Contains(userMessage, "含义") || strings.Contains(userMessage, "指的是") ||
		strings.Contains(userMessage, "缩写") {
		// 提取缩写词，强制调用 nbnhhsh 工具查询
		abbr := extractAbbreviation(userMessage)
		if abbr != "" {
			return "nbnhhsh", abbr, true
		}
	}

	// 检测纯字母/数字组合（可能是缩写词）
	abbr := extractAbbreviation(userMessage)
	if abbr != "" && len(abbr) <= 10 {
		// 大模型直接回答了但没有调用工具，强制使用 nbnhhsh 获取准确结果
		return "nbnhhsh", abbr, true
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

// convertToZhipuMessages 将Ollama消息格式转换为智谱消息格式
func convertToZhipuMessages(msgs []api.Message) []zhipuMessage {
	result := make([]zhipuMessage, len(msgs))
	for i, msg := range msgs {
		result[i] = zhipuMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return result
}

func (s *AgentService) executeToolAndSummarize(ctx context.Context, userMessage, toolName, toolInput string) string {
	if toolName == "weather" {
		cities := extractCandidateCities(toolInput)
		if len(cities) == 0 {
			normalized := normalizeWeatherInput(toolInput)
			if normalized == "" {
				return "请明确要查询的城市，例如：北京今天、西安明天。"
			}
			cities = []string{normalized}
		}

		results := make([]string, 0, len(cities))
		for _, city := range cities {
			query := city
			if strings.Contains(toolInput, "明天") {
				query = city + "明天"
			} else if strings.Contains(toolInput, "后天") {
				query = city + "后天"
			} else if strings.Contains(toolInput, "七天") || strings.Contains(toolInput, "7天") {
				query = city + "七天"
			} else if strings.Contains(toolInput, "今天") {
				query = city + "今天"
			}
			results = append(results, s.toolManager.Execute(ctx, "weather", query))
		}

		merged := strings.Join(results, "\n\n")
		if len(results) > 1 {
			if summary, ok := s.runModelSecondPass(ctx, userMessage, merged); ok {
				return summary
			}
		}
		return merged
	}

	result := s.toolManager.Execute(ctx, toolName, toolInput)
	if summary, ok := s.runModelSecondPass(ctx, userMessage, result); ok {
		return summary
	}
	return result
}

func (s *AgentService) runModelSecondPass(ctx context.Context, userMessage, toolResult string) (string, bool) {
	prompt := fmt.Sprintf("用户问题：%s\n\n工具结果：%s\n\n请仅基于工具结果回答用户问题，禁止编造。若是对比问题，请给出结构化对比结论。", userMessage, toolResult)
	msgs := []api.Message{
		{Role: "system", Content: s.systemPrompt},
		{Role: "user", Content: prompt},
	}

	if s.ollamaSvc != nil {
		if resp, err := s.ollamaSvc.Chat(ctx, msgs); err == nil && strings.TrimSpace(resp) != "" {
			return resp, true
		}
	}
	if s.zhipuSvc != nil {
		if resp, err := s.zhipuSvc.Chat(ctx, convertToZhipuMessages(msgs)); err == nil && strings.TrimSpace(resp) != "" {
			return resp, true
		}
	}
	return "", false
}

func normalizeWeatherInput(input string) string {
	s := strings.TrimSpace(input)
	replacements := []string{"天气", "今天", "明天", "后天", "七天", "7天", "一周", "未来", "怎么样", "如何", "相比", "对比", "哪个", "哪个更", "更", "吗", "呢", "？", "?"}
	for _, r := range replacements {
		s = strings.ReplaceAll(s, r, "")
	}
	s = strings.ReplaceAll(s, "和", " ")
	s = strings.ReplaceAll(s, "与", " ")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func extractCandidateCities(input string) []string {
	s := strings.TrimSpace(input)
	replacements := []string{"天气", "今天", "明天", "后天", "七天", "7天", "一周", "未来", "怎么样", "如何", "相比", "对比", "哪个", "哪个更", "更", "吗", "呢", "？", "?", "的"}
	for _, r := range replacements {
		s = strings.ReplaceAll(s, r, "")
	}
	separators := []string{"和", "与", "、", ",", "，"}
	for _, sep := range separators {
		s = strings.ReplaceAll(s, sep, "|")
	}
	items := strings.Split(s, "|")
	cities := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		c := strings.TrimSpace(item)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		cities = append(cities, c)
	}
	return cities
}

func getContextBudget(mode string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "4k":
		return 4000, 12
	case "2k", "":
		return 2000, 8
	default:
		return 2000, 8
	}
}

func (s *AgentService) trimHistory(history []api.Message, userMessage string) []api.Message {
	if len(history) == 0 {
		return nil
	}

	usedChars := len(s.systemPrompt) + len(userMessage)
	if usedChars >= s.maxChars {
		return nil
	}

	keptReversed := make([]api.Message, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		if len(keptReversed) >= s.maxMsgs {
			break
		}
		contentLen := len(msg.Content)
		if usedChars+contentLen > s.maxChars {
			break
		}
		usedChars += contentLen
		keptReversed = append(keptReversed, msg)
	}

	trimmed := make([]api.Message, len(keptReversed))
	for i := range keptReversed {
		trimmed[len(keptReversed)-1-i] = keptReversed[i]
	}
	return trimmed
}

// StreamMode 流式模式
type StreamMode string

const (
	StreamModeThinking StreamMode = "thinking" // 深度思考：注入思考标签，流式切分 thought/answer
	StreamModeAuto     StreamMode = "auto"     // Auto：纯流式答案，不注入思考标签
)

// ProcessStream 流式处理用户消息
// history: 历史消息（不含当前 user 消息）
// mode: thinking / auto
// chunkCh: 用于推送流式片段；调用方负责关闭（本函数结束时不 close）
// 返回值：完整的 thought、answer（供上层持久化）
func (s *AgentService) ProcessStream(ctx context.Context, userMessage string, history []api.Message, mode StreamMode, chunkCh chan<- StreamChunk) (thought string, answer string) {
	// 组装系统提示
	sysPrompt := s.systemPrompt
	if mode == StreamModeThinking {
		// 免费模型走标签方案；付费模型走原生 reasoning，标签方案对付费模型无害
		sysPrompt = s.systemPrompt + deepThinkingPromptSuffix
	}

	msgs := []api.Message{{Role: "system", Content: sysPrompt}}
	trimmed := s.trimHistory(history, userMessage)
	msgs = append(msgs, trimmed...)
	msgs = append(msgs, api.Message{Role: "user", Content: userMessage})

	// 第一次流式生成
	var streamErr error
	if s.zhipuSvc != nil {
		if err := s.streamViaZhipu(ctx, msgs, mode, chunkCh, &thought, &answer); err == nil {
			streamErr = nil
		} else {
			log.Warn(ctx, "ProcessStream: 智谱流式失败，降级 Ollama: %v", err)
			streamErr = err
		}
	}
	if answer == "" && s.ollamaSvc != nil {
		if err := s.streamViaOllama(ctx, msgs, mode, chunkCh, &thought, &answer); err == nil {
			streamErr = nil
		} else {
			log.Error(ctx, "ProcessStream: Ollama 流式失败: %v", err)
			streamErr = err
		}
	}

	// 记录首次响应
	log.Info(ctx, "ProcessStream: 第一次响应 thoughtLen=%d answerLen=%d answer=%q", len(thought), len(answer), truncateForLog(answer, 500))

	if answer == "" && thought == "" {
		if streamErr != nil {
			chunkCh <- StreamChunk{Type: ChunkError, Content: streamErr.Error()}
		} else {
			chunkCh <- StreamChunk{Type: ChunkError, Content: "服务暂时不可用，请稍后重试"}
		}
		return
	}

	// 工具调用检测：answer 是否是工具调用格式
	toolName, toolInput, isToolCall := ParseToolCallFromResponse(strings.TrimSpace(answer))
	if isToolCall {
		log.Info(ctx, "ProcessStream: 检测到工具调用 tool=%s input=%s", toolName, toolInput)

		// 通知前端清空 answer 累积（把工具 JSON 抹掉）
		chunkCh <- StreamChunk{Type: ChunkAnswerReset}
		// 推 tool_call
		chunkCh <- StreamChunk{Type: ChunkToolCall, Tool: toolName, ToolInput: toolInput, Content: toolInput}

		// 执行工具（Process 里的 executeToolAndSummarize 会对 weather 做多城市拆分并返回汇总文本；这里我们直接调 toolManager 保持简单，多城市能力后续补）
		toolResult := s.toolManager.Execute(ctx, toolName, toolInput)
		log.Info(ctx, "ProcessStream: 工具结果 tool=%s resultLen=%d", toolName, len(toolResult))

		// 推 tool_result（一次性）
		chunkCh <- StreamChunk{Type: ChunkToolResult, Tool: toolName, Content: toolResult}

		// 第二次流式：让模型基于工具结果给出面向用户的答复
		secondMsgs := []api.Message{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: fmt.Sprintf("用户问题：%s\n\n工具「%s」的执行结果如下：\n%s\n\n请仅基于工具结果，用自然语言回答用户问题。禁止再输出任何工具调用 JSON。", userMessage, toolName, toolResult)},
		}
		// 二次调用时使用普通答案（不再输出思考标签），采用 Auto 语义
		var secondThought, secondAnswer string
		if s.zhipuSvc != nil {
			if err := s.streamViaZhipu(ctx, secondMsgs, StreamModeAuto, chunkCh, &secondThought, &secondAnswer); err != nil {
				log.Warn(ctx, "ProcessStream: 二次流式失败，降级 Ollama: %v", err)
				secondAnswer = ""
			}
		}
		if secondAnswer == "" && s.ollamaSvc != nil {
			_ = s.streamViaOllama(ctx, secondMsgs, StreamModeAuto, chunkCh, &secondThought, &secondAnswer)
		}
		if secondAnswer == "" {
			// 兜底：直接把工具原文当答案
			secondAnswer = toolResult
			chunkCh <- StreamChunk{Type: ChunkAnswer, Content: toolResult}
		}
		log.Info(ctx, "ProcessStream: 二次响应 answerLen=%d answer=%q", len(secondAnswer), truncateForLog(secondAnswer, 500))
		// 覆盖 answer 用于上层持久化
		answer = secondAnswer
	}

	return
}

// truncateForLog 截断字符串用于日志输出
func truncateForLog(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "...(truncated)"
}

// streamViaZhipu 通过智谱流式生成
func (s *AgentService) streamViaZhipu(ctx context.Context, msgs []api.Message, mode StreamMode, chunkCh chan<- StreamChunk, thoughtOut, answerOut *string) error {
	// 判断是否走原生 reasoning
	preferReasoning := mode == StreamModeThinking
	model, apiKey, isReasoning := s.zhipuSvc.SelectModelAndKey(preferReasoning)
	log.Info(ctx, "streamViaZhipu: model=%s reasoning=%v mode=%s", model, isReasoning, mode)

	tokenCh := make(chan string, 32)
	reasoningCh := make(chan string, 32)
	errCh := make(chan error, 1)

	go func() {
		errCh <- s.zhipuSvc.ChatStream(ctx, convertToZhipuMessages(msgs), model, apiKey, tokenCh, reasoningCh)
		close(tokenCh)
		close(reasoningCh)
	}()

	// Auto 模式：不解析思考标签，全部作为 answer
	if mode == StreamModeAuto {
		var ab strings.Builder
		tokenCount := 0
		for tokenCh != nil || reasoningCh != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case tk, ok := <-tokenCh:
				if !ok {
					tokenCh = nil
					continue
				}
				ab.WriteString(tk)
				tokenCount++
				chunkCh <- StreamChunk{Type: ChunkAnswer, Content: tk}
			case _, ok := <-reasoningCh:
				if !ok {
					reasoningCh = nil
					continue
				}
			}
		}
		log.Info(ctx, "streamViaZhipu(auto): tokens=%d totalLen=%d", tokenCount, ab.Len())
		*answerOut = ab.String()
		return <-errCh
	}

	// Thinking 模式
	if isReasoning {
		// 付费模型原生 reasoning：直接分发
		var tb, ab strings.Builder
		for tokenCh != nil || reasoningCh != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case tk, ok := <-tokenCh:
				if !ok {
					tokenCh = nil
					continue
				}
				ab.WriteString(tk)
				chunkCh <- StreamChunk{Type: ChunkAnswer, Content: tk}
			case rc, ok := <-reasoningCh:
				if !ok {
					reasoningCh = nil
					continue
				}
				tb.WriteString(rc)
				chunkCh <- StreamChunk{Type: ChunkThought, Content: rc}
			}
		}
		*thoughtOut = tb.String()
		*answerOut = ab.String()
		return <-errCh
	}

	// 免费模型：通过标签解析器切分
	parser := NewTagStreamParser()
	relayCh := make(chan StreamChunk, 64)
	var tb, ab strings.Builder
	done := make(chan struct{})
	go func() {
		for c := range relayCh {
			if c.Type == ChunkThought {
				tb.WriteString(c.Content)
			} else if c.Type == ChunkAnswer {
				ab.WriteString(c.Content)
			}
			chunkCh <- c
		}
		close(done)
	}()

	for tokenCh != nil {
		select {
		case <-ctx.Done():
			close(relayCh)
			<-done
			return ctx.Err()
		case tk, ok := <-tokenCh:
			if !ok {
				tokenCh = nil
				continue
			}
			parser.Feed(tk, relayCh)
		}
	}
	parser.Flush(relayCh)
	close(relayCh)
	<-done
	*thoughtOut = tb.String()
	*answerOut = ab.String()
	return <-errCh
}

// streamViaOllama 通过 Ollama 流式生成（兜底）
func (s *AgentService) streamViaOllama(ctx context.Context, msgs []api.Message, mode StreamMode, chunkCh chan<- StreamChunk, thoughtOut, answerOut *string) error {
	tokenCh := make(chan string, 32)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.ollamaSvc.ChatStream(ctx, msgs, tokenCh)
		close(tokenCh)
	}()

	if mode == StreamModeAuto {
		var ab strings.Builder
		for tk := range tokenCh {
			ab.WriteString(tk)
			chunkCh <- StreamChunk{Type: ChunkAnswer, Content: tk}
		}
		*answerOut = ab.String()
		return <-errCh
	}

	parser := NewTagStreamParser()
	relayCh := make(chan StreamChunk, 64)
	var tb, ab strings.Builder
	done := make(chan struct{})
	go func() {
		for c := range relayCh {
			if c.Type == ChunkThought {
				tb.WriteString(c.Content)
			} else if c.Type == ChunkAnswer {
				ab.WriteString(c.Content)
			}
			chunkCh <- c
		}
		close(done)
	}()

	for tk := range tokenCh {
		parser.Feed(tk, relayCh)
	}
	parser.Flush(relayCh)
	close(relayCh)
	<-done
	*thoughtOut = tb.String()
	*answerOut = ab.String()
	return <-errCh
}

// GenerateTitle 根据首条 user 消息和 AI 答案生成 15 字内的会话标题
// 若生成失败或不可用，返回空串（调用方处理）
func (s *AgentService) GenerateTitle(ctx context.Context, userMessage, aiAnswer string) string {
	if strings.TrimSpace(userMessage) == "" {
		return ""
	}
	// 截取 AI 答案前 200 字，避免过长
	ai := aiAnswer
	if len([]rune(ai)) > 200 {
		r := []rune(ai)
		ai = string(r[:200])
	}
	prompt := fmt.Sprintf("请为以下对话生成一个15个字以内的标题，不要引号、不要标点、不要解释：\n\n用户：%s\nAI：%s\n\n标题：", userMessage, ai)
	msgs := []zhipuMessage{
		{Role: "system", Content: "你是标题生成器，只输出标题本身，不含引号和标点。最多15个字。"},
		{Role: "user", Content: prompt},
	}

	if s.zhipuSvc != nil {
		// 使用免费模型即可，快速便宜
		model := s.zhipuSvc.config.Model
		apiKey := s.zhipuSvc.config.APIKey
		if title, err := s.zhipuSvc.ChatWithModel(ctx, msgs, model, apiKey, 40); err == nil {
			return cleanTitle(title)
		} else {
			log.Warn(ctx, "GenerateTitle: 智谱调用失败 err=%v", err)
		}
	}

	if s.ollamaSvc != nil {
		apiMsgs := []api.Message{
			{Role: "system", Content: "你是标题生成器，只输出标题本身，不含引号和标点。最多15个字。"},
			{Role: "user", Content: prompt},
		}
		if title, err := s.ollamaSvc.Chat(ctx, apiMsgs); err == nil {
			return cleanTitle(title)
		}
	}
	return ""
}

// cleanTitle 清理标题结果
func cleanTitle(s string) string {
	s = strings.TrimSpace(s)
	// 去除首尾引号
	trimChars := "\"'`「」【】《》。，,.！!？?：: \n\r\t"
	s = strings.Trim(s, trimChars)
	// 截断到 15 字符（rune）
	r := []rune(s)
	if len(r) > 15 {
		s = string(r[:15])
	}
	return s
}
