package file

import (
	"context"
	"fmt"

	"agent/library/log"
)

// OllamaService 接口，用于避免循环导入
type OllamaService interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// Processor 处理器接口
type Processor interface {
	Process(ctx context.Context, content string) (string, error)
}

// ============ SummaryProcessor 摘要处理器 ============

// SummaryProcessor 文件摘要处理器
type SummaryProcessor struct {
	ollamaSvc OllamaService
}

// NewSummaryProcessor 创建 SummaryProcessor
func NewSummaryProcessor(ollamaSvc OllamaService) *SummaryProcessor {
	return &SummaryProcessor{
		ollamaSvc: ollamaSvc,
	}
}

// Process 生成文件摘要
func (p *SummaryProcessor) Process(ctx context.Context, content string) (string, error) {
	// 如果内容过长，截断
	if len(content) > 5000 {
		content = content[:5000] + "\n[内容已截断...]"
	}

	prompt := "请总结以下文档的核心内容：\n\n" + content + "\n\n请提供不超过200字的摘要。"

	// 构建消息
	msg := map[string]string{
		"tool":  "summary",
		"input": content,
	}

	log.Info(ctx, "SummaryProcessor: 调用 AI 生成摘要")

	// 调用 Ollama API
	resp, err := p.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}

	_ = msg // 避免未使用变量警告

	return resp, nil
}

// ============ CodeAnalyzer 代码分析器 ============

// CodeAnalyzer 代码分析器，支持 Go/Python/JavaScript
type CodeAnalyzer struct {
	ollamaSvc OllamaService
}

// NewCodeAnalyzer 创建 CodeAnalyzer
func NewCodeAnalyzer(ollamaSvc OllamaService) *CodeAnalyzer {
	return &CodeAnalyzer{
		ollamaSvc: ollamaSvc,
	}
}

// AnalyzeExplain 解释代码
func (ca *CodeAnalyzer) AnalyzeExplain(ctx context.Context, code, language string) (string, error) {
	prompt := "请解释以下 " + language + " 代码的功能和逻辑：\n\n" +
		"```" + language + "\n" + code + "\n```\n\n" +
		"请用简洁的语言说明主要逻辑和用途。"

	log.Info(ctx, "CodeAnalyzer: 解释 %s 代码", language)
	resp, err := ca.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// AnalyzeError 查找代码错误
func (ca *CodeAnalyzer) AnalyzeError(ctx context.Context, code, language string) (string, error) {
	prompt := "请检查以下 " + language + " 代码的语法错误和逻辑问题：\n\n" +
		"```" + language + "\n" + code + "\n```\n\n" +
		"列出所有发现的问题，并提供修复建议。"

	log.Info(ctx, "CodeAnalyzer: 检查 %s 代码错误", language)
	resp, err := ca.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// AnalyzeOptimize 优化代码
func (ca *CodeAnalyzer) AnalyzeOptimize(ctx context.Context, code, language string) (string, error) {
	langHint := language
	if language == "go" {
		langHint = "Go（优先考虑 Go 最佳实践）"
	}

	prompt := "请为以下 " + langHint + " 代码提供优化建议：\n\n" +
		"```" + language + "\n" + code + "\n```\n\n" +
		"列出可以改进的地方和具体建议。"

	log.Info(ctx, "CodeAnalyzer: 优化 %s 代码", language)
	resp, err := ca.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// Process 处理代码分析请求（通用入口）
func (ca *CodeAnalyzer) Process(ctx context.Context, analysisType, code, language string) (string, error) {
	switch analysisType {
	case "explain":
		return ca.AnalyzeExplain(ctx, code, language)
	case "error":
		return ca.AnalyzeError(ctx, code, language)
	case "optimize":
		return ca.AnalyzeOptimize(ctx, code, language)
	default:
		return "", fmt.Errorf("输入错误：不支持的分析类型 (%s)", analysisType)
	}
}

// ============ FormatConverter 格式转换器 ============

// FormatConverter 格式转换器
type FormatConverter struct {
	ollamaSvc OllamaService
}

// NewFormatConverter 创建 FormatConverter
func NewFormatConverter(ollamaSvc OllamaService) *FormatConverter {
	return &FormatConverter{
		ollamaSvc: ollamaSvc,
	}
}

// ConvertMarkdownToHTML 将 Markdown 转换为 HTML
func (fc *FormatConverter) ConvertMarkdownToHTML(ctx context.Context, markdown string) (string, error) {
	if len(markdown) > 5000 {
		markdown = markdown[:5000] + "\n[内容已截断...]"
	}

	prompt := "请将以下 Markdown 转换为 HTML：\n\n" +
		"```markdown\n" + markdown + "\n```\n\n" +
		"返回完整的 HTML 代码。"

	log.Info(ctx, "FormatConverter: 将 Markdown 转换为 HTML")
	resp, err := fc.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// ConvertJSONToCSV 将 JSON 转换为 CSV
func (fc *FormatConverter) ConvertJSONToCSV(ctx context.Context, jsonContent string) (string, error) {
	if len(jsonContent) > 5000 {
		jsonContent = jsonContent[:5000] + "\n[内容已截断...]"
	}

	prompt := "请将以下 JSON 数据转换为 CSV 格式：\n\n" +
		"```json\n" + jsonContent + "\n```\n\n" +
		"返回 CSV 格式的数据。"

	log.Info(ctx, "FormatConverter: 将 JSON 转换为 CSV")
	resp, err := fc.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// ConvertTextToMindMap 将文本转换为思维导图大纲
func (fc *FormatConverter) ConvertTextToMindMap(ctx context.Context, text string) (string, error) {
	if len(text) > 5000 {
		text = text[:5000] + "\n[内容已截断...]"
	}

	prompt := "请根据以下文本生成思维导图大纲（使用 Markdown 列表格式）：\n\n" +
		text + "\n\n" +
		"返回层级清晰的 Markdown 列表。"

	log.Info(ctx, "FormatConverter: 将文本转换为思维导图大纲")
	resp, err := fc.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// ConvertMarkdownToWord 将 Markdown 转换为 Word（格式化文本）
func (fc *FormatConverter) ConvertMarkdownToWord(ctx context.Context, markdown string) (string, error) {
	if len(markdown) > 5000 {
		markdown = markdown[:5000] + "\n[内容已截断...]"
	}

	prompt := "请将以下 Markdown 转换为适合 Word 文档的格式（返回带格式的文本）：\n\n" +
		"```markdown\n" + markdown + "\n```\n\n" +
		"返回带有标题、段落等格式的文本内容。"

	log.Info(ctx, "FormatConverter: 将 Markdown 转换为 Word 格式")
	resp, err := fc.ollamaSvc.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("服务错误：AI 处理失败，请稍后重试")
	}
	return resp, nil
}

// Process 处理格式转换请求（通用入口）
func (fc *FormatConverter) Process(ctx context.Context, content, sourceFormat, targetFormat string) (string, error) {
	switch {
	case sourceFormat == "markdown" && targetFormat == "html":
		return fc.ConvertMarkdownToHTML(ctx, content)
	case sourceFormat == "markdown" && targetFormat == "word":
		return fc.ConvertMarkdownToWord(ctx, content)
	case sourceFormat == "json" && targetFormat == "csv":
		return fc.ConvertJSONToCSV(ctx, content)
	case sourceFormat == "text" && targetFormat == "mindmap":
		return fc.ConvertTextToMindMap(ctx, content)
	default:
		return "", fmt.Errorf("输入错误：不支持的格式转换 (%s → %s)", sourceFormat, targetFormat)
	}
}
