# Ollama AI Agent 🤖

一个基于 Ollama 大模型的智能助手，支持多种工具扩展，包括天气查询、数学计算、本地文件处理等功能。

## ✨ 核心功能

### 🌤️ 天气查询工具（Weather）
- 支持 **100+ 中国城市** 天气查询
- 支持多天气预报（今天、明天、后天、七天）
- 自动回退策略：Open-Meteo API → wttr.in API
- 示例：
  ```
  "西安今天几度"
  "北京明天天气"
  "上海七天预报"
  ```

### 🧮 数学计算工具（Calculator）
- 支持基础四则运算（+、-、*、/）
- 支持负数和小数计算
- 示例：
  ```
  "100 + 50"
  "3.14 * 2"
  "1000 - 200 / 4"
  ```

### 📄 智能文件处理工具（FileTool）
支持多种文件格式和操作方式：

#### 支持格式
- **文本文档**：TXT、MD、JSON
- **代码文件**：Go、Python、JavaScript
- **特殊格式**：PDF

#### 支持操作
1. **文件解析与总结** (`parse`)
   ```
   "帮我总结一下 data/example.txt 文件的内容"
   "阅读 data/example.json"
   ```

2. **代码分析** (`code_analyze`)
   - Explain（代码解释）
   - Error（错误检查）
   - Optimize（性能优化）
   ```
   "分析一下 data/example.py 文件"
   ```

3. **格式转换** (`convert`)
   - Markdown → HTML
   - Markdown → Word
   - JSON → CSV
   - Text → MindMap
   ```
   "把 data/example.md 转换成 HTML"
   ```

### 📁 文件上传与管理
- 支持前端直接上传文件（最大 10MB）
- 支持文件查看和删除
- 文件类型白名单（安全性）
- Web UI 集成文件管理面板

## 🚀 快速开始

### 前提条件
- Go 1.24.1 或更高版本
- Ollama 服务运行中（默认地址：`localhost:11434`）
- 已安装 `qwen:7b` 模型（或其他兼容模型）

### 安装和运行

1. **克隆项目**
```bash
git clone <repo-url>
cd agent
```

2. **安装依赖**
```bash
go mod download
```

3. **构建**
```bash
go build -o agent .
```

4. **运行**
```bash
./agent
```

服务将在 `http://localhost:8080` 启动

### 使用 Makefile

```bash
# 编译
make build

# 运行
make run

# 测试
make test

# 清理
make clean
```

### Docker 运行

```bash
# 使用 docker-compose
docker-compose up

# 或手动构建
docker build -t agent .
docker run -p 8080:8080 agent
```

## 📖 使用指南

### Web 界面
访问 `http://localhost:8080` 打开前端 Web 界面

**功能包括：**
- 💬 聊天对话框
- 📋 文件管理面板
- 📤 文件上传
- 📋 文件列表查看
- 🗑️ 文件删除

### API 接口

#### 1. 聊天接口
```bash
POST /api/chat
Content-Type: application/json

{
  "message": "今天西安几度",
  "history": []
}
```

**响应示例：**
```json
{
  "code": 1000,
  "message": "成功",
  "data": {
    "result": "西安的未来1天天气：\n今天：阴天，最高温度：24°C，最低温度：9°C"
  }
}
```

#### 2. 文件上传
```bash
POST /api/upload
Content-Type: multipart/form-data

file=@example.txt
```

**响应示例：**
```json
{
  "code": 1000,
  "data": {
    "filename": "example.txt",
    "path": "data/example.txt",
    "size": 1024
  }
}
```

#### 3. 文件列表
```bash
GET /api/files
```

**响应示例：**
```json
{
  "code": 1000,
  "data": {
    "files": [
      {"name": "example.txt", "size": 1024},
      {"name": "example.md", "size": 2048}
    ]
  }
}
```

#### 4. 删除文件
```bash
DELETE /api/file/delete
Content-Type: application/json

{
  "filename": "example.txt"
}
```

### 文件工具 JSON 格式

当大模型调用文件工具时，JSON 格式如下：

```json
{
  "action": "parse",
  "file": "data/example.txt",
  "mode": "summary"
}
```

**action 参数：**
- `parse` - 解析文件（默认）
- `code_analyze` - 代码分析
- `convert` - 格式转换

**mode 参数（仅 parse）：**
- `summary` - 摘要（默认）
- `extract` - 关键词提取
- `full` - 完整内容

**type 参数（仅 code_analyze）：**
- `explain` - 解释代码（默认）
- `error` - 检查错误
- `optimize` - 优化建议

**target 参数（仅 convert）：**
- `html` - 转换为 HTML（Markdown）
- `word` - 转换为 Word（Markdown）
- `csv` - 转换为 CSV（JSON）
- `mindmap` - 转换为思维导图（文本）

## 🏗️ 项目结构

```
agent/
├── models/
│   └── service/
│       ├── agent/           # Agent 核心逻辑
│       │   ├── agent.go     # Agent 主逻辑
│       │   ├── ollama.go    # Ollama 客户端
│       │   ├── tool.go      # 工具接口
│       │   └── file_tool.go # 文件工具实现
│       ├── file/            # 文件处理模块
│       │   ├── file_reader.go
│       │   ├── content_parser.go
│       │   ├── processors.go
│       │   └── file_reader_test.go
│       ├── weather/         # 天气工具
│       │   └── weather.go
│       ├── calculator/      # 计算工具
│       │   └── calculator.go
│       └── ...
├── webapi/
│   └── controllers/         # HTTP 控制器
│       ├── controller.go    # 聊天控制器
│       └── file_controller.go  # 文件管理控制器
├── router/
│   └── router.go           # 路由配置
├── static/
│   └── index.html          # 前端 Web 界面
├── library/
│   └── log/                # 日志系统
├── global/
│   ├── config.go           # 配置定义
│   └── error.go            # 错误定义
├── data/                   # 上传文件存储目录
├── logs/                   # 日志存储目录
├── main.go                 # 程序入口
├── go.mod
└── README.md
```

## ⚙️ 配置说明

配置文件位于 `global/config.go`：

```go
// Ollama 配置
Ollama: OllamaConfig{
    Host:    "localhost:11434",   // Ollama 地址
    Model:   "qwen:7b",          // 使用的模型
    Timeout: 120 * time.Second,  // 请求超时
}

// 服务配置
Server: ServerConfig{
    Port:         "8080",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
}

// 日志配置
Log: LogConfig{
    Level:      "info",
    Path:       "./logs",
    MaxSize:    100,           // MB
    MaxBackups: 30,            // 保留文件数
    MaxAge:     7,             // 保留天数
    Compress:   true,
}
```

**修改配置方法：**
编辑 `global/config.go` 中的 `DefaultConfig` 变量，或通过环境变量覆盖：

```bash
export OLLAMA_HOST="192.168.1.100:11434"
export OLLAMA_MODEL="qwen:14b"
```

## 🔒 安全特性

### 文件访问安全
- **路径验证**：所有文件访问限制在 `./data` 目录
- **防目录穿透**：检测 `../` 路径尝试
- **文件类型白名单**：仅允许特定扩展名
- **大小限制**：单文件最大 10MB

### API 安全
- 默认本地访问（可配置 CORS）
- 请求超时保护
- 错误信息隐藏敏感信息

## 🔧 故障排除

### Ollama 连接失败
```
错误：初始化 Ollama 客户端失败
```

**解决方案：**
1. 确保 Ollama 服务运行中
   ```bash
   ollama serve
   ```
2. 检查 Ollama 地址配置
3. 确保端口 11434 未被占用

### 模型不存在
```
错误：模型 qwen:7b 不可用
```

**解决方案：**
```bash
ollama pull qwen:7b
```

### 天气工具失败
天气工具自动回退，如果主 API 失败会尝试备用 API

### 文件工具权限错误
```
错误：权限错误：文件路径不在允许目录内
```

**确保：**
1. 文件位于 `data/` 目录
2. 使用相对路径 `data/filename`
3. 文件名不包含 `../` 等特殊字符

### 大模型不调用工具
如果大模型直接用知识库回答而不调用工具，系统会自动使用降级策略：
- 根据关键词识别应该调用哪个工具
- 自动构造工具调用 JSON
- 确保功能正常运行

## 📊 日志

日志文件位于 `./logs/` 目录

查看实时日志：
```bash
tail -f ./logs/agent_*.log
```

日志格式：
```
[INFO] 2026/04/01 17:20:29 main.go:40: [logid] 日志信息
```

## 🧪 测试

运行测试：
```bash
go test -v ./...
```

特定包测试：
```bash
go test -v ./models/service/file/
go test -v ./models/service/calculator/
```

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request

## 📝 版本信息

查看版本信息：
```bash
./agent version
```

或访问 API：
```bash
curl http://localhost:8080/api/chat -X POST
```

## 📄 许可证

MIT License

## 🙏 致谢

- [Ollama](https://ollama.ai) - 本地 LLM 框架
- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF 处理库
- [Open-Meteo](https://open-meteo.com) - 天气数据来源

## 📧 联系方式

如有问题或建议，请通过 Issue 联系我们

---

**最后更新：** 2026 年 4 月 1 日  
**当前版本：** v1.0.0
