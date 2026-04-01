package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"agent/library/log"
	"agent/models/service/file"
)

// FileTool 文件智能处理工具
// 支持文件内容解析、总结、代码分析、格式转换等功能
//
// 使用示例：
//
//  1. 解析和总结文件：
//     {"action":"parse","file":"data/example.txt","mode":"summary"}
//     mode 可选值: summary(默认)、extract、full
//
//  2. 代码分析（支持 Go/Python/JavaScript）：
//     {"action":"code_analyze","file":"data/example.go","type":"explain"}
//     type 可选值: explain(默认)、error、optimize
//
//  3. 格式转换：
//     {"action":"convert","file":"data/example.md","target":"html"}
//     支持的转换: md->html、md->word、json->csv、text->mindmap
//
// 安全性：
// - 所有文件访问都限制在 ./data 目录内
// - 防止路径穿透攻击（../、绝对路径等）
// - 最大文件大小限制为 10MB
type FileTool struct {
	fileReader      *file.FileReader
	summaryProc     *file.SummaryProcessor
	codeAnalyzer    *file.CodeAnalyzer
	formatConverter *file.FormatConverter
}

// NewFileTool 创建 FileTool 实例
func NewFileTool(ollamaSvc *OllamaService, allowedBasePath string) *FileTool {
	return &FileTool{
		fileReader:      file.NewFileReader(allowedBasePath),
		summaryProc:     file.NewSummaryProcessor(ollamaSvc),
		codeAnalyzer:    file.NewCodeAnalyzer(ollamaSvc),
		formatConverter: file.NewFormatConverter(ollamaSvc),
	}
}

// Name 返回工具名称
func (ft *FileTool) Name() string {
	return "file"
}

// Description 返回工具描述
func (ft *FileTool) Description() string {
	return `智能文件处理工具，支持TXT/MD/JSON/PDF解析、代码分析和格式转换。
使用格式：
- 解析和总结：{"action":"parse","file":"data/example.txt","mode":"summary"}
- 代码分析：{"action":"code_analyze","file":"data/example.go","type":"explain"}
- 格式转换：{"action":"convert","file":"data/example.md","target":"html"}`
}

// Execute 执行文件操作
func (ft *FileTool) Execute(ctx context.Context, input string) string {
	log.Info(ctx, "FileTool.Execute: 收到请求: %s", input)

	// 解析 JSON 输入
	var request map[string]string
	if err := json.Unmarshal([]byte(input), &request); err != nil {
		log.Error(ctx, "FileTool.Execute: JSON 解析失败: %v", err)
		return "输入错误：无效的 JSON 格式"
	}

	// 获取 action
	action := request["action"]
	if action == "" {
		return "输入错误：缺少 action 参数 (parse/code_analyze/convert)"
	}

	// 获取文件路径
	filePath := request["file"]
	if filePath == "" {
		return "输入错误：缺少 file 参数"
	}

	// 分发到不同的处理器
	switch action {
	case "parse":
		return ft.handleParse(ctx, filePath, request["mode"])
	case "code_analyze":
		return ft.handleCodeAnalyze(ctx, filePath, request["type"])
	case "convert":
		return ft.handleConvert(ctx, filePath, request["target"])
	default:
		return fmt.Sprintf("输入错误：不支持的 action (%s)", action)
	}
}

// handleParse 处理文件解析和总结
func (ft *FileTool) handleParse(ctx context.Context, filePath string, mode string) string {
	log.Info(ctx, "FileTool.handleParse: 文件=%s, 模式=%s", filePath, mode)

	// 读取文件
	content, err := ft.fileReader.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "FileTool.handleParse: 文件读取失败: %v", err)
		return err.Error()
	}

	// 解析内容
	parser := file.GetParser(filePath)
	parseResult, err := parser.Parse(content)
	if err != nil {
		log.Error(ctx, "FileTool.handleParse: 解析失败: %v", err)
		return fmt.Sprintf("文件错误：%v", err)
	}

	// 根据 mode 返回结果
	if mode == "" || mode == "summary" {
		// 生成摘要
		summary, err := ft.summaryProc.Process(ctx, parseResult.RawContent)
		if err != nil {
			log.Error(ctx, "FileTool.handleParse: 摘要生成失败: %v", err)
			return err.Error()
		}
		return summary
	} else if mode == "extract" {
		// 返回关键信息
		result := fmt.Sprintf("文件: %s\n标题: %s\n关键词: %v\n字数: %d",
			filePath, parseResult.Title, parseResult.Keywords, len(parseResult.RawContent))
		return result
	} else if mode == "full" {
		// 返回完整内容
		return parseResult.RawContent
	} else {
		return fmt.Sprintf("输入错误：不支持的 mode (%s)", mode)
	}
}

// handleCodeAnalyze 处理代码分析
func (ft *FileTool) handleCodeAnalyze(ctx context.Context, filePath string, analysisType string) string {
	log.Info(ctx, "FileTool.handleCodeAnalyze: 文件=%s, 类型=%s", filePath, analysisType)

	// 检查文件后缀名并获取语言
	ext := strings.ToLower(filepath.Ext(filePath))
	var language string

	switch ext {
	case ".go":
		language = "go"
	case ".py":
		language = "python"
	case ".js", ".ts", ".jsx", ".tsx":
		language = "javascript"
	default:
		return fmt.Sprintf("格式错误：不支持的代码文件类型 (%s)", ext)
	}

	// 读取代码文件
	code, err := ft.fileReader.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "FileTool.handleCodeAnalyze: 文件读取失败: %v", err)
		return err.Error()
	}

	// 执行代码分析
	if analysisType == "" {
		analysisType = "explain"
	}

	result, err := ft.codeAnalyzer.Process(ctx, analysisType, code, language)
	if err != nil {
		log.Error(ctx, "FileTool.handleCodeAnalyze: 分析失败: %v", err)
		return err.Error()
	}

	return result
}

// handleConvert 处理格式转换
func (ft *FileTool) handleConvert(ctx context.Context, filePath string, targetFormat string) string {
	log.Info(ctx, "FileTool.handleConvert: 文件=%s, 目标格式=%s", filePath, targetFormat)

	if targetFormat == "" {
		return "输入错误：缺少 target 参数"
	}

	// 读取文件
	content, err := ft.fileReader.ReadFile(filePath)
	if err != nil {
		log.Error(ctx, "FileTool.handleConvert: 文件读取失败: %v", err)
		return err.Error()
	}

	// 根据文件类型和目标格式进行转换
	ext := strings.ToLower(filepath.Ext(filePath))

	var result string
	switch {
	case ext == ".md" && targetFormat == "html":
		result, err = ft.formatConverter.ConvertMarkdownToHTML(ctx, content)
	case ext == ".md" && targetFormat == "word":
		result, err = ft.formatConverter.ConvertMarkdownToWord(ctx, content)
	case ext == ".json" && targetFormat == "csv":
		result, err = ft.formatConverter.ConvertJSONToCSV(ctx, content)
	case (ext == ".txt" || ext == ".md") && targetFormat == "mindmap":
		result, err = ft.formatConverter.ConvertTextToMindMap(ctx, content)
	default:
		return fmt.Sprintf("输入错误：不支持的格式转换 (%s → %s)", ext, targetFormat)
	}

	if err != nil {
		log.Error(ctx, "FileTool.handleConvert: 转换失败: %v", err)
		return err.Error()
	}

	return result
}
