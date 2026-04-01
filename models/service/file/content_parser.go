package file

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ParseResult 解析结果结构体
type ParseResult struct {
	RawContent string                 // 原始内容
	Title      string                 // 文档标题
	Keywords   []string               // 关键词
	Summary    string                 // 内容摘要（由 AI 生成）
	Metadata   map[string]interface{} // 元数据
}

// ContentParser 内容解析器接口
type ContentParser interface {
	Parse(content string) (ParseResult, error)
}

// ============ TXTParser ============

// TXTParser 文本解析器
type TXTParser struct{}

// NewTXTParser 创建 TXTParser
func NewTXTParser() *TXTParser {
	return &TXTParser{}
}

// Parse 解析文本文件
func (p *TXTParser) Parse(content string) (ParseResult, error) {
	result := ParseResult{
		RawContent: content,
		Metadata:   make(map[string]interface{}),
	}

	// 提取第一行作为标题
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && lines[0] != "" {
		result.Title = lines[0]
	}

	// 统计行数和字数
	result.Metadata["lines"] = len(lines)
	result.Metadata["characters"] = len(content)

	return result, nil
}

// ============ MarkdownParser ============

// MarkdownParser Markdown 解析器
type MarkdownParser struct{}

// NewMarkdownParser 创建 MarkdownParser
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{}
}

// Parse 解析 Markdown 文件
func (p *MarkdownParser) Parse(content string) (ParseResult, error) {
	result := ParseResult{
		RawContent: content,
		Keywords:   []string{},
		Metadata:   make(map[string]interface{}),
	}

	lines := strings.Split(content, "\n")
	var headers []string
	var codeBlocks int
	var links []string

	// 提取标题
	headerRegex := regexp.MustCompile(`^(#+)\s+(.+)$`)
	// 提取代码块
	codeBlockRegex := regexp.MustCompile("```")
	// 提取链接
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)

	inCodeBlock := false
	for _, line := range lines {
		// 检查 Markdown 标题
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			title := matches[2]
			headers = append(headers, title)

			// 第一个 H1 作为文档标题
			if level == 1 && result.Title == "" {
				result.Title = title
			}
		}

		// 检查代码块
		if codeBlockRegex.MatchString(line) {
			inCodeBlock = !inCodeBlock
			if inCodeBlock {
				codeBlocks++
			}
		}

		// 提取链接
		if linkMatches := linkRegex.FindAllStringSubmatch(line, -1); linkMatches != nil {
			for _, match := range linkMatches {
				links = append(links, match[2])
			}
		}
	}

	// 从标题中提取关键词
	result.Keywords = headers

	// 保存元数据
	result.Metadata["headers"] = headers
	result.Metadata["code_blocks"] = codeBlocks
	result.Metadata["links"] = links
	result.Metadata["characters"] = len(content)
	result.Metadata["lines"] = len(lines)

	return result, nil
}

// ============ JSONParser ============

// JSONParser JSON 解析器
type JSONParser struct{}

// NewJSONParser 创建 JSONParser
func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

// Parse 解析 JSON 文件
func (p *JSONParser) Parse(content string) (ParseResult, error) {
	result := ParseResult{
		RawContent: content,
		Keywords:   []string{},
		Metadata:   make(map[string]interface{}),
	}

	// 验证 JSON 格式
	var jsonData interface{}
	err := json.Unmarshal([]byte(content), &jsonData)
	if err != nil {
		return result, fmt.Errorf("JSON解析错误：%v", err)
	}

	result.Metadata["valid_json"] = true

	// 提取顶级键作为关键词
	if jsonObj, ok := jsonData.(map[string]interface{}); ok {
		for key := range jsonObj {
			result.Keywords = append(result.Keywords, key)
		}
		result.Metadata["json_type"] = "object"
		result.Metadata["keys_count"] = len(jsonObj)
	} else if jsonArray, ok := jsonData.([]interface{}); ok {
		result.Metadata["json_type"] = "array"
		result.Metadata["array_length"] = len(jsonArray)
	}

	result.Metadata["characters"] = len(content)
	result.Title = "JSON 数据文件"

	return result, nil
}

// ============ PDFParser ============

// PDFParser PDF 解析器
type PDFParser struct{}

// NewPDFParser 创建 PDFParser
func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

// Parse 解析 PDF 文件
func (p *PDFParser) Parse(content string) (ParseResult, error) {
	result := ParseResult{
		RawContent: content,
		Metadata:   make(map[string]interface{}),
	}

	// 从 content 中提取页数信息
	pageCountRegex := regexp.MustCompile(`PDF 文件包含 (\d+) 页`)
	matches := pageCountRegex.FindStringSubmatch(content)
	if matches != nil {
		result.Metadata["pdf_pages"] = matches[1]
	}

	result.Metadata["characters"] = len(content)
	result.Title = "PDF 文档"

	return result, nil
}

// ============ GetParser 工厂函数 ============

// GetParser 根据文件扩展名获取对应的解析器
func GetParser(filePath string) ContentParser {
	switch {
	case strings.HasSuffix(filePath, ".txt"):
		return NewTXTParser()
	case strings.HasSuffix(filePath, ".md"):
		return NewMarkdownParser()
	case strings.HasSuffix(filePath, ".json"):
		return NewJSONParser()
	case strings.HasSuffix(filePath, ".pdf"):
		return NewPDFParser()
	default:
		// 默认使用文本解析器
		return NewTXTParser()
	}
}
