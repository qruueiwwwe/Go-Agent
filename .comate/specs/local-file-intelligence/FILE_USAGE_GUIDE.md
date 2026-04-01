# 文件处理功能使用指南

## 🎯 功能概述

FileTool 是智能文件处理工具，现已支持完整的文件生命周期管理：

| 功能 | 说明 |
|-----|------|
| 📤 **上传文件** | 支持多种文件格式上传 |
| 📁 **管理文件** | 查看、删除已上传的文件 |
| 🔍 **解析文件** | 读取并分析文件内容 |
| 💬 **智能处理** | 文件总结、代码分析、格式转换 |

---

## 🚀 快速开始

### 1. 启动服务

```bash
cd /Users/xierunze/code/agent
./agent
```

访问: http://localhost:8080

### 2. 上传文件

**方式1：前端 UI 上传**
- 点击 "📤 上传文件" 按钮
- 选择需要的文件（支持 TXT、MD、JSON、GO、PY、JS、PDF）
- 文件自动上传到 `data/` 目录

**方式2：直接放置文件**
- 将文件直接放到项目的 `data/` 目录
- 无需上传，可直接分析

### 3. 分析文件

**方式1：点击文件名**
- 上传完成后，点击文件名自动生成分析请求

**方式2：手动输入 JSON 命令**

---

## 📋 API 接口文档

### 文件上传
```http
POST /api/upload
Content-Type: multipart/form-data

file: <二进制文件>
```

**响应示例**
```json
{
  "code": 1000,
  "message": "成功",
  "data": {
    "filename": "example.txt",
    "path": "data/example.txt",
    "size": 1024,
    "message": "文件上传成功，大小 1024 字节"
  }
}
```

### 获取文件列表
```http
GET /api/files
```

**响应示例**
```json
{
  "code": 1000,
  "message": "成功",
  "data": {
    "files": [
      {
        "name": "example.txt",
        "size": 1024
      }
    ],
    "count": 1
  }
}
```

### 删除文件
```http
DELETE /api/file/delete?filename=example.txt
```

**响应示例**
```json
{
  "code": 1000,
  "message": "成功",
  "data": {
    "message": "文件删除成功"
  }
}
```

### 文件处理（通过聊天接口）
```http
POST /api/chat
Content-Type: application/json

{
  "message": "{\"action\":\"parse\",\"file\":\"data/example.txt\",\"mode\":\"summary\"}"
}
```

---

## 🔧 文件处理命令格式

### 1. 文件解析和总结

**解析文件**
```json
{
  "action": "parse",
  "file": "data/example.txt",
  "mode": "summary"
}
```

| 参数 | 可选值 | 说明 |
|-----|-------|------|
| `action` | `parse` | 文件解析操作 |
| `file` | 相对路径 | 文件在 data/ 目录内的相对路径 |
| `mode` | `summary`(默认) | 生成文件摘要 |
| | `extract` | 提取关键信息 |
| | `full` | 返回完整内容 |

**示例请求**
```json
{
  "action": "parse",
  "file": "data/README.md",
  "mode": "summary"
}
```

### 2. 代码分析

**解释代码**
```json
{
  "action": "code_analyze",
  "file": "data/example.go",
  "type": "explain"
}
```

| 参数 | 可选值 | 说明 |
|-----|-------|------|
| `action` | `code_analyze` | 代码分析操作 |
| `file` | 相对路径 | 代码文件路径 |
| `type` | `explain` | 解释代码功能 |
| | `error` | 检查代码错误 |
| | `optimize` | 提供优化建议 |

**支持的语言**
- ✅ Go (`.go`)
- ✅ Python (`.py`)
- ✅ JavaScript (`.js`)

**示例请求**
```json
{
  "action": "code_analyze",
  "file": "data/calculator.go",
  "type": "explain"
}
```

### 3. 格式转换

**格式转换**
```json
{
  "action": "convert",
  "file": "data/example.md",
  "target": "html"
}
```

| 源格式 | 目标格式 | 说明 |
|-------|---------|------|
| Markdown | `html` | 转换为 HTML |
| Markdown | `word` | 转换为 Word 格式 |
| JSON | `csv` | 转换为 CSV |
| 文本/Markdown | `mindmap` | 转换为思维导图大纲 |

**示例请求**
```json
{
  "action": "convert",
  "file": "data/guide.md",
  "target": "html"
}
```

---

## 📝 支持的文件类型

| 格式 | 扩展名 | 用途 |
|-----|-------|------|
| 纯文本 | `.txt` | 文档、日志 |
| Markdown | `.md` | 文档、笔记 |
| JSON | `.json` | 数据配置 |
| Go | `.go` | 代码分析 |
| Python | `.py` | 代码分析 |
| JavaScript | `.js` | 代码分析 |
| PDF | `.pdf` | 文档处理 |

---

## 🛡️ 安全性说明

### 权限控制
- ✅ 所有文件操作限制在 `data/` 目录
- ✅ 防止路径穿透攻击
- ✅ 最大文件大小 10MB

### 支持的文件类型限制
- 只能上传/处理指定的文件类型
- 其他格式会被拒绝

---

## 💡 使用示例

### 示例1：上传并总结 Markdown 文档

```
1. 点击 "📤 上传文件"
2. 选择 guide.md
3. 上传后自动显示在文件列表
4. 点击 guide.md 文件名
5. AI 自动生成文档摘要
```

### 示例2：代码审查

```
1. 上传 calculator.go
2. 输入以下 JSON 命令检查错误：
{
  "action": "code_analyze",
  "file": "data/calculator.go",
  "type": "error"
}
3. AI 检查代码并提出改进建议
```

### 示例3：格式转换

```
1. 上传 documentation.md
2. 输入以下 JSON 命令转换为 HTML：
{
  "action": "convert",
  "file": "data/documentation.md",
  "target": "html"
}
3. 获取转换后的 HTML 内容
```

---

## 📂 文件存储位置

所有上传的文件存储在项目的 `data/` 目录：

```
agent/
└── data/
    ├── example.txt
    ├── guide.md
    ├── config.json
    └── calculator.go
```

---

## ⚠️ 常见问题

### Q: 文件上传失败提示"文件过大"
**A:** 单个文件最大限制为 10MB，请上传较小的文件。

### Q: 上传后找不到文件
**A:** 检查文件名是否正确，文件应该在 `data/` 目录中。

### Q: 代码分析不支持我的语言
**A:** 目前仅支持 Go、Python、JavaScript，其他语言可尝试作为文本分析。

### Q: 如何删除不需要的文件
**A:** 点击文件右侧的 "🗑️" 按钮即可删除。

---

## 🔄 工作流程图

```
上传文件
   ↓
[文件存储到 data/ 目录]
   ↓
选择操作
   ├→ 解析 → AI 总结/提取/返回完整内容
   ├→ 代码分析 → AI 解释/检查/优化
   ├→ 格式转换 → AI 生成转换结果
   └→ 删除 → 从 data/ 目录移除
   ↓
返回结果给用户
```

---

**最后更新**: 2026-04-01  
**版本**: 1.0.0 (包含文件上传功能)
