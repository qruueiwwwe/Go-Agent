# 本地化文件智能处理功能 - 详细需求文档

## 1. 功能概述

为现有 Ollama AI Agent 项目实现「本地化文件智能处理」功能，通过新增 FileTool，支持：
- **文件内容解析与总结**：支持 TXT、Markdown、JSON、PDF 格式
- **文件格式转换**：Markdown→HTML/Word、JSON→Excel、文本→思维导图大纲
- **代码文件辅助**：支持 Go、Python、JS 代码的解释、错误查找和优化建议

---

## 2. 系统架构与技术方案

### 2.1 架构设计

```
用户请求 (ChatAPI)
    ↓
AgentService (现有)
    ↓
ToolManager (现有)
    ↓
FileTool (新增) ← 核心组件
    ├─ FileReader (文件读取+权限校验)
    ├─ ContentParser (内容解析)
    │   ├─ TXTParser
    │   ├─ MarkdownParser
    │   ├─ JSONParser
    │   └─ PDFParser
    ├─ Processor (处理器)
    │   ├─ SummaryProcessor (总结)
    │   ├─ CodeAnalyzer (代码分析)
    │   └─ FormatConverter (格式转换)
    └─ OllamaIntegration (Ollama API调用)
```

### 2.2 实现策略

#### 文件读取与权限控制
- **允许目录**：限制读取 `data/` 目录及其子目录
- **路径校验**：使用 `filepath.Abs()` 和 `filepath.Join()` 防止路径穿透 (`../` 绕过)
- **错误处理**：禁止读取限制目录外的文件，返回权限错误

#### 内容解析
- **TXT/MD**：使用 Go 原生 `os.ReadFile()`
- **JSON**：使用 `encoding/json` 标准库
- **PDF**：使用 `github.com/pdfcpu/pdfcpu` 库提取文本

#### AI 处理
- **Prompt 构建**：将解析后的内容拼接为 Prompt，通过 OllamaService 调用大模型
- **格式转换**：由 AI 生成转换结果（Markdown→HTML 通过 AI 输出；JSON→Excel 通过 AI 生成 CSV）
- **代码分析**：AI 执行代码解释、错误检查、优化建议

---

## 3. 详细功能设计

### 3.1 FileTool 核心结构

#### Tool 接口实现
```go
type FileTool struct {
    allowedBasePath string  // 允许的基础路径，如 "./data"
}

// Name 返回工具名称
func (f *FileTool) Name() string {
    return "file"
}

// Description 返回工具描述
func (f *FileTool) Description() string {
    return "智能文件处理工具，支持TXT/MD/JSON/PDF解析、格式转换和代码分析。使用格式：{\"action\":\"parse\",\"file\":\"data/example.txt\",\"mode\":\"summary\"}或{\"action\":\"code_analyze\",\"file\":\"data/example.go\",\"type\":\"explain\"}"
}

// Execute 执行文件操作
func (f *FileTool) Execute(ctx context.Context, input string) string {
    // 解析 input JSON，执行相应操作
}
```

#### 输入格式（JSON）
```json
{
    "action": "parse",           // parse/code_analyze/convert
    "file": "data/file.txt",     // 相对允许基础路径的文件路径
    "mode": "summary",           // parse模式：summary/extract/full
    "type": "explain",           // code_analyze类型：explain/error/optimize
    "target_format": "html"      // convert目标格式：html/word/csv等
}
```

### 3.2 文件读取与权限控制

#### FileReader 接口
```go
type FileReader struct {
    allowedBasePath string
}

// ReadFile 读取文件并返回内容
func (fr *FileReader) ReadFile(filePath string) (content string, err error) {
    // 1. 规范化路径：filepath.Abs()
    // 2. 校验权限：检查是否在 allowedBasePath 内
    // 3. 检查文件是否存在和可读
    // 4. 读取内容：os.ReadFile()
    // 5. 返回内容或错误
}

// ValidatePath 校验路径权限
func (fr *FileReader) ValidatePath(filePath string) (bool, error) {
    // 1. 规范化路径
    // 2. 确保路径在 allowedBasePath 内
    // 3. 返回校验结果
}
```

**路径校验逻辑**：
```
输入：data/subdir/file.txt
1. filepath.Join(allowedBasePath, input) → ./data/data/subdir/file.txt (错误处理)
   正确做法：直接检查 input 是否在 allowedBasePath 内
2. abs_input = filepath.Abs(input)
3. abs_base = filepath.Abs(allowedBasePath)
4. 检查 abs_input 是否以 abs_base 开头
5. 检查不存在 ../.. 等路径穿透
```

### 3.3 内容解析器

#### Parser 接口
```go
type ContentParser interface {
    Parse(content string) (ParseResult, error)
}

type ParseResult struct {
    RawContent  string  // 原始内容
    Title       string  // 文档标题（若有）
    Keywords    []string // 关键词
    Summary     string  // AI 生成的摘要
    Metadata    map[string]interface{} // 元数据
}
```

#### 各种格式解析器

**TXTParser**：
```go
// 简单返回原文本内容
type TXTParser struct {}

func (p *TXTParser) Parse(content string) (ParseResult, error) {
    return ParseResult{
        RawContent: content,
    }, nil
}
```

**MarkdownParser**：
```go
// 提取 Markdown 标题、代码块、链接等结构化信息
type MarkdownParser struct {}

func (p *MarkdownParser) Parse(content string) (ParseResult, error) {
    // 1. 使用正则表达式提取标题
    // 2. 识别代码块、链接、图片
    // 3. 构建目录结构
}
```

**JSONParser**：
```go
// 验证 JSON 格式，提取结构
type JSONParser struct {}

func (p *JSONParser) Parse(content string) (ParseResult, error) {
    // 1. 使用 json.Unmarshal 验证格式
    // 2. 提取 key 作为关键词
    // 3. 构建数据结构摘要
}
```

**PDFParser**：
```go
// 使用 pdfcpu 库提取文本
type PDFParser struct {}

func (p *PDFParser) Parse(content string) (ParseResult, error) {
    // 注：PDF 需要先从文件读取二进制数据
    // 然后使用 pdfcpu 提取文本
}
```

### 3.4 处理器

#### SummaryProcessor
**场景**：用户请求文件总结
```go
type SummaryProcessor struct {
    ollamaSvc *OllamaService
}

func (p *SummaryProcessor) Process(ctx context.Context, content string) (string, error) {
    // 1. 构建 Prompt：
    //    "请总结以下文档的核心内容：\n{content}\n\n请提供不超过200字的摘要。"
    // 2. 调用 ollamaSvc.Chat()
    // 3. 返回 AI 总结结果
}
```

#### CodeAnalyzer
**支持语言**：Go、Python、JavaScript

**场景1 - 解释代码**：
```go
prompt := `请解释以下 Go/Python/JavaScript 代码的功能和逻辑：
\`\`\`
{code}
\`\`\`
请用简洁的语言说明主要逻辑和用途。`
```

**场景2 - 查找错误**：
```go
prompt := `请检查以下 Go/Python/JavaScript 代码的语法错误和逻辑问题：
\`\`\`
{code}
\`\`\`
列出所有发现的问题，并提供修复建议。`
```

**场景3 - 优化代码**：
```go
prompt := `请为以下 Go/Python/JavaScript 代码提供优化建议（优先考虑 Go 最佳实践）：
\`\`\`
{code}
\`\`\`
列出可以改进的地方和具体建议。`
```

#### FormatConverter
**Markdown → HTML**：
```go
prompt := `请将以下 Markdown 转换为 HTML：
\`\`\`markdown
{markdown_content}
\`\`\`
返回完整的 HTML 代码。`
```

**JSON → CSV**：
```go
prompt := `请将以下 JSON 数据转换为 CSV 格式：
\`\`\`json
{json_content}
\`\`\`
返回 CSV 格式的数据。`
```

**文本 → 思维导图大纲**：
```go
prompt := `请根据以下文本生成思维导图大纲（使用 Markdown 列表格式）：
\`\`\`
{text_content}
\`\`\`
返回层级清晰的 Markdown 列表。`
```

---

## 4. 受影响的文件

| 文件 | 修改类型 | 说明 |
|-----|--------|------|
| `models/service/agent/tool.go` | 修改 | 无需改动（Tool 接口已存在） |
| `models/service/agent/file_tool.go` | **新增** | FileTool 主文件，包含 Execute 逻辑 |
| `models/service/agent/file_reader.go` | **新增** | 文件读取和权限校验 |
| `models/service/agent/content_parser.go` | **新增** | 各种格式的内容解析器 |
| `models/service/agent/processors.go` | **新增** | 摘要、代码分析、格式转换处理器 |
| `go.mod` | 修改 | 添加 `github.com/pdfcpu/pdfcpu` 依赖 |
| `main.go` | 修改 | 注册 FileTool |

---

## 5. 实现细节

### 5.1 FileTool.Execute 方法流程

```
输入 JSON 字符串
    ↓
解析 JSON → 获取 action、file、mode/type
    ↓
FileReader.ReadFile(file) → 获取文件内容
    ↓
根据文件后缀名选择 ContentParser（TXT/MD/JSON/PDF）
    ↓
根据 action 选择 Processor（Parse/CodeAnalyze/Convert）
    ↓
Processor.Process(content) → 调用 AI 处理
    ↓
返回处理结果
```

### 5.2 错误处理

| 错误情形 | 处理方式 |
|--------|--------|
| 路径超出允许目录 | 返回 "权限错误：文件路径不在允许目录内" |
| 文件不存在 | 返回 "文件错误：指定文件不存在" |
| 不支持的文件格式 | 返回 "格式错误：不支持的文件类型" |
| JSON 解析失败 | 返回 "输入错误：无效的 JSON 格式" |
| AI 调用失败 | 返回 "服务错误：AI 处理失败，请稍后重试" |

### 5.3 PDF 处理特殊说明

PDF 解析使用 `pdfcpu` 库的特点：
- 需要从文件直接读取二进制数据（不能只传文本）
- 可提取文本和元数据
- 大文件可能需要性能优化

实现方式：
```go
// 在 FileReader.ReadFile 中添加 PDF 特殊处理
if strings.HasSuffix(filePath, ".pdf") {
    // 使用 pdfcpu 读取 PDF
    return readPDFContent(filePath)
}
```

---

## 6. 约束与安全性

### 6.1 路径权限校验

**防止路径穿透**：
```go
// 错误示例（容易被绕过）
joinedPath := filepath.Join(baseDir, userInput)

// 正确做法
absInput := filepath.Abs(userInput)
absBase := filepath.Abs(baseDir)
if !strings.HasPrefix(absInput, absBase) {
    return error("权限错误")
}
// 还需检查是否包含 ".."
if strings.Contains(userInput, "..") {
    return error("权限错误")
}
```

### 6.2 文件大小限制

建议设置最大读取大小，避免内存溢出：
```go
const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB
if fileInfo.Size() > MAX_FILE_SIZE {
    return error("文件过大，超过10MB限制")
}
```

---

## 7. 验收标准

| 标准 | 验证方法 |
|-----|--------|
| ✓ 能够读取 TXT/MD/JSON 文件 | 在 `data/` 目录下放置测试文件，通过 API 请求验证 |
| ✓ 能够解析 PDF 文件 | 在 `data/` 目录下放置 PDF，验证文本提取 |
| ✓ 能够总结文件内容 | AI 返回有意义的总结 |
| ✓ 能够分析代码 | 对 Go/Python/JS 代码返回正确的解释/错误/优化建议 |
| ✓ 能够转换格式 | MD→HTML、JSON→CSV、文本→大纲转换成功 |
| ✓ 权限控制有效 | 尝试读取 `data/` 外的文件被拒绝 |
| ✓ 依赖正确安装 | 所有必需的依赖已在 go.mod 中声明 |

---

## 8. 外部依赖

| 库 | 用途 | 版本 |
|----|------|------|
| `github.com/pdfcpu/pdfcpu` | PDF 文本提取 | 最新稳定版 |
| `encoding/json` (标准库) | JSON 解析 | - |
| `os/io` (标准库) | 文件读写 | - |
| `filepath` (标准库) | 路径处理 | - |
| `regexp` (标准库) | 正则匹配 | - |

