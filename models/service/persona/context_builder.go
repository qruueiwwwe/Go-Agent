package persona

import (
	"context"
	"fmt"
	"strings"

	"agent/library/log"
	"agent/models/entity"

	"github.com/ollama/ollama/api"
)

const (
	// ContextBudget16K 16k 字符预算
	ContextBudget16K = 16000
	// MaxRecentMessages 最近保留的消息条数（8轮对话）
	MaxRecentMessages = 16
	// SummaryThreshold 超过此消息数时触发摘要
	SummaryThreshold = 24
	// MaxExamples 最多注入的示例数量
	MaxExamples = 3
)

// LLMService LLM服务接口（用于摘要生成）
type LLMService interface {
	Chat(ctx context.Context, messages []api.Message) (string, error)
}

// ContextBuilder 上下文构建器
type ContextBuilder struct {
	llmService LLMService
}

// NewContextBuilder 创建上下文构建器
func NewContextBuilder(llmService LLMService) *ContextBuilder {
	return &ContextBuilder{llmService: llmService}
}

// Build 构建完整上下文
func (b *ContextBuilder) Build(persona *entity.Persona, history []entity.PersonaChat, userInput string) []api.Message {
	budget := ContextBudget16K
	used := 0
	msgs := []api.Message{}

	// Layer 1: 角色定义（始终保留）
	systemPrompt := b.buildSystemPrompt(persona)
	msgs = append(msgs, api.Message{Role: "system", Content: systemPrompt})
	used += len(systemPrompt)

	// Layer 2: Few-shot 示例（动态选择）
	examples := b.selectExamples(persona.Examples, budget-used-2000)
	for _, ex := range examples {
		msgs = append(msgs, api.Message{Role: "user", Content: ex.UserInput})
		msgs = append(msgs, api.Message{Role: "assistant", Content: ex.AIResponse})
		used += len(ex.UserInput) + len(ex.AIResponse)
	}

	// Layer 3 & 4: 历史消息处理
	historyMsgs := b.processHistory(history, budget-used-len(userInput)-500)
	msgs = append(msgs, historyMsgs...)

	// Layer 5: 当前输入
	msgs = append(msgs, api.Message{Role: "user", Content: userInput})

	return msgs
}

// buildSystemPrompt 构建角色系统提示词
func (b *ContextBuilder) buildSystemPrompt(persona *entity.Persona) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("你是 %s。\n\n", persona.Name))

	if persona.Tagline != "" {
		sb.WriteString(fmt.Sprintf("【角色介绍】\n%s\n\n", persona.Tagline))
	}

	if persona.Personality != "" {
		sb.WriteString(fmt.Sprintf("【性格特质】\n%s\n\n", persona.Personality))
	}

	sb.WriteString("【对话准则】\n")
	sb.WriteString("- 始终保持角色设定，不要跳出角色\n")
	sb.WriteString("- 使用符合角色的语气和表达方式\n")
	sb.WriteString("- 根据用户输入灵活响应，但不要偏离角色\n")
	sb.WriteString("- 如果用户的问题超出角色能力范围，可以礼貌地引导回角色领域\n")

	sb.WriteString(fmt.Sprintf("\n%s", persona.SystemPrompt))

	return sb.String()
}

// selectExamples 根据预算选择示例
func (b *ContextBuilder) selectExamples(examples []entity.Example, budget int) []entity.Example {
	if len(examples) == 0 || budget <= 0 {
		return nil
	}

	selected := []entity.Example{}
	used := 0

	for _, ex := range examples {
		cost := len(ex.UserInput) + len(ex.AIResponse)
		if used+cost <= budget {
			selected = append(selected, ex)
			used += cost
		} else {
			break
		}
	}

	// 最多选择 MaxExamples 个示例
	if len(selected) > MaxExamples {
		selected = selected[:MaxExamples]
	}

	return selected
}

// processHistory 处理历史消息
func (b *ContextBuilder) processHistory(history []entity.PersonaChat, budget int) []api.Message {
	if len(history) == 0 {
		return nil
	}

	// 如果历史消息过多，生成摘要
	if len(history) > SummaryThreshold {
		// 取最近的 MaxRecentMessages 条保留，其余生成摘要
		recentStart := len(history) - MaxRecentMessages
		olderHistory := history[:recentStart]
		recentHistory := history[recentStart:]

		// 计算摘要预算
		summaryBudget := budget / 2

		// 生成历史摘要
		summary := b.summarizeHistory(olderHistory, summaryBudget)
		if summary != "" {
			summaryMsg := api.Message{
				Role:    "system",
				Content: fmt.Sprintf("【对话背景摘要】\n%s", summary),
			}

			// 将摘要 + 最近历史组合
			result := []api.Message{summaryMsg}
			for _, h := range recentHistory {
				result = append(result, api.Message{Role: h.Role, Content: h.Content})
			}
			return result
		}
	}

	// 直接使用最近的历史消息
	start := 0
	if len(history) > MaxRecentMessages {
		start = len(history) - MaxRecentMessages
	}

	result := []api.Message{}
	for _, h := range history[start:] {
		result = append(result, api.Message{Role: h.Role, Content: h.Content})
	}
	return result
}

// summarizeHistory 生成历史摘要
func (b *ContextBuilder) summarizeHistory(history []entity.PersonaChat, budget int) string {
	if b.llmService == nil {
		return ""
	}

	// 拼接历史消息
	var sb strings.Builder
	for _, h := range history {
		roleName := "用户"
		if h.Role == "assistant" {
			roleName = "AI"
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", roleName, h.Content))
	}

	historyText := sb.String()
	// 如果历史太长，截取
	if len(historyText) > 4000 {
		historyText = historyText[:4000] + "..."
	}

	// 调用 LLM 生成摘要
	maxLen := budget / 2
	if maxLen > 500 {
		maxLen = 500
	}

	prompt := fmt.Sprintf(`请用简洁的语言概括以下对话的关键信息，控制在%d字以内：

%s

摘要要点：
1. 主要讨论的话题
2. 用户的关注点
3. 已得出的结论或共识`, maxLen, historyText)

	msgs := []api.Message{
		{Role: "user", Content: prompt},
	}

	summary, err := b.llmService.Chat(context.Background(), msgs)
	if err != nil {
		log.Error(context.Background(), "ContextBuilder.summarizeHistory: 生成摘要失败 err=%v", err)
		return ""
	}

	// 截取到预算范围内
	if len(summary) > maxLen {
		summary = summary[:maxLen] + "..."
	}

	return summary
}
