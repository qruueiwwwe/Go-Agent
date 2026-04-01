# 本地化文件智能处理功能 - 任务清单

## 核心开发任务

- [x] Task 1: 添加 PDF 处理依赖并设置基础工具结构
    - 1.1: 在 go.mod 中添加 `github.com/pdfcpu/pdfcpu` 依赖
    - 1.2: 运行 `go mod download` 和 `go mod tidy` 验证依赖
    - 1.3: 创建 `models/service/agent/file_reader.go` 文件框架

- [x] Task 2: 实现文件读取和权限控制 (FileReader)
    - 2.1: 定义 FileReader 结构体和初始化函数
    - 2.2: 实现 ValidatePath 方法进行路径权限校验（防止路径穿透）
    - 2.3: 实现 ReadFile 方法进行文件读取和大小限制检查
    - 2.4: 添加 readPDFContent 辅助函数使用 pdfcpu 提取 PDF 文本
    - 2.5: 编写路径权限校验的单元测试

- [x] Task 3: 实现内容解析器 (ContentParser)
    - 3.1: 创建 `models/service/agent/content_parser.go` 文件
    - 3.2: 定义 ContentParser 接口和 ParseResult 结构体
    - 3.3: 实现 TXTParser（直接返回原文本）
    - 3.4: 实现 MarkdownParser（提取标题、代码块、链接等）
    - 3.5: 实现 JSONParser（验证格式并提取结构）
    - 3.6: 实现 PDFParser（调用 readPDFContent）
    - 3.7: 创建 GetParser 工厂函数根据文件后缀选择解析器

- [x] Task 4: 实现处理器 (Processors)
    - 4.1: 创建 `models/service/agent/processors.go` 文件
    - 4.2: 定义 Processor 接口
    - 4.3: 实现 SummaryProcessor（总结文件内容）
    - 4.4: 实现 CodeAnalyzer（支持 Go/Python/JS 的解释、错误检查、优化）
    - 4.5: 实现 FormatConverter（MD→HTML、JSON→CSV、文本→思维导图）
    - 4.6: 为各处理器添加对应的 Prompt 构建方法

- [x] Task 5: 实现 FileTool 核心逻辑
    - 5.1: 创建 `models/service/agent/file_tool.go` 文件
    - 5.2: 定义 FileTool 结构体（包含 FileReader、ContentParser、各 Processor）
    - 5.3: 实现 Name() 方法返回 "file"
    - 5.4: 实现 Description() 方法返回工具描述
    - 5.5: 实现 Execute() 方法：解析 JSON input → 执行对应 action
    - 5.6: 在 Execute 中处理所有错误情形并返回对应错误消息

- [x] Task 6: 集成 FileTool 到 Agent 框架
    - 6.1: 修改 `main.go`：导入 FileTool 相关包
    - 6.2: 在 main 函数中创建 FileTool 实例并注册到 ToolManager
    - 6.3: 验证 ToolManager.GetToolsDesc() 包含 file 工具描述

- [x] Task 7: 创建测试数据和验证功能
    - 7.1: 创建 `data/` 目录结构和测试文件（example.txt、example.md、example.json、example.go）
    - 7.2: 编写集成测试验证文件读取功能
    - 7.3: 编写集成测试验证路径权限控制
    - 7.4: 编写集成测试验证各内容解析器
    - 7.5: 编写集成测试验证各处理器（调用 AI）
    - 7.6: 验证权限拒绝场景（尝试读取 data/ 外的文件）

- [x] Task 8: 文档和边界情况处理
    - 8.1: 添加 FileTool 使用说明注释
    - 8.2: 补充错误消息的国际化或明确性
    - 8.3: 处理大文件边界情况（超过 10MB 限制）
    - 8.4: 处理空文件和畸形数据输入的边界情况

