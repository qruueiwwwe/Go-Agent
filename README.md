# Ollama AI Agent

一个基于 Ollama 大模型的智能助手，支持多种工具扩展，包括天气查询、数学计算、本地文件处理、缩写词猜测等功能。当 Ollama 不可用时，自动切换到智谱清言云端API作为后备方案。

## 核心功能

### 用户认证与账号安全（新增）
- **邮箱验证码注册**：通过邮箱 + 验证码完成注册
- **手机号可选**：注册时手机号非必填，可为空
- **登录认证**：用户名密码登录，返回 JWT Token
- **忘记密码**：通过邮箱验证码重置密码
- **权限校验**：主页与 API 支持登录态校验

### 大模型服务
- **Ollama 本地模型**：主要使用本地 Ollama 服务（默认 gemma3:4b）
- **智谱清言后备**：当 Ollama 不可用时，自动切换到智谱清言云端API
- **无缝切换**：用户无感知，保证服务持续可用

### 天气查询工具（Weather）
- 支持 **100+ 中国城市** 天气查询
- 支持多天气预报（今天、明天、后天、七天）
- 双API策略：接口盒子API（中国气象局数据）→ 高德天气API
- 示例：
  ```
  "西安今天几度"
  "北京明天天气"
  "上海七天预报"
  ```

### 数学计算工具（Calculator）
- 支持基础四则运算（+、-、*、/）
- 支持负数和小数计算
- 示例：
  ```
  "100 + 50"
  "3.14 * 2"
  "1000 - 200 / 4"
  ```

### 缩写词猜测工具（Nbnhhsh）
- 猜测网络缩写词含义
- 支持拼音首字母缩写
- 示例：
  ```
  "yyds是什么意思"
  "xswl的含义"
  "awsl"
  ```

### 智能文件处理工具（FileTool）
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

### 文件上传与管理
- 支持前端直接上传文件（最大 10MB）
- 支持文件查看和删除
- 文件类型白名单（安全性）
- Web UI 集成文件管理面板

## 快速开始

### 前提条件
- Go 1.24.1 或更高版本
- Ollama 服务运行中（默认地址：`localhost:11434`）
- 已安装 `gemma3:4b` 模型（或其他兼容模型）

### 安装和运行

1. **克隆项目**
```bash
git clone <repo-url>
cd agent
```

2. **配置环境变量**
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑 .env 文件，填入你的API密钥
# 必填项：
# - ZHIPU_API_KEY：智谱清言API密钥（后备服务）
# 
# 可选项：
# - WEATHER_API_ID/KEY：接口盒子天气API
# - AMAP_API_KEY：高德天气API
# - DATABASE_PASSWORD：MySQL密码
# - JWT_SECRET：登录签名密钥
# - MAIL_SMTP_*：邮箱验证码SMTP配置
```

3. **安装依赖**
```bash
go mod download
```

4. **构建并运行**
```bash
go build -o agent .
./agent
```

服务将在 `http://localhost:25565` 启动

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
docker-compose up -d

# 或手动构建
docker build -t agent .
docker run -p 25565:25565 --env-file .env agent
```

## 环境变量配置

创建 `.env` 文件配置敏感信息：

```bash
# ========== 必填配置 ==========

# 智谱清言API配置（后备服务）
# 获取地址: https://open.bigmodel.cn/
ZHIPU_API_KEY=your_zhipu_api_key

# ========== 可选配置 ==========

# 天气API配置（接口盒子）
WEATHER_API_ID=your_weather_api_id
WEATHER_API_KEY=your_weather_api_key

# 高德天气API配置
# 获取地址: https://lbs.amap.com/
AMAP_API_KEY=your_amap_api_key

# 数据库配置
DATABASE_PASSWORD=your_db_password

# JWT 配置
JWT_SECRET=your_super_secret_jwt_key

# 邮箱验证码 SMTP 配置
MAIL_SMTP_HOST=smtp.qq.com
MAIL_SMTP_PORT=465
MAIL_USERNAME=your_email@qq.com
MAIL_PASSWORD=your_smtp_auth_code
MAIL_FROM_NAME=Agent验证码
MAIL_FROM_ADDR=your_email@qq.com

# 开发环境配置
APP_ENV=dev
EMAIL_DEBUG_FIXED_CODE=
```

**注意**：`.env` 文件已被 `.gitignore` 忽略，不会提交到代码仓库。

## 使用指南

### Web 界面
访问 `http://localhost:25565` 打开前端 Web 界面

**登录注册相关页面：**
- 登录页：`/login.html`
- 邮箱注册页：`/register.html`
- 忘记密码页：`/forgot-password.html`

**功能包括：**
- 聊天对话框
- 文件管理面板
- 文件上传
- 文件列表查看
- 文件删除

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
    "result": "西安未来1天天气：\n今天：阴天，24~9°C"
  }
}
```

#### 2. 文件上传
```bash
POST /api/upload
Content-Type: multipart/form-data

file=@example.txt
```

#### 3. 文件列表
```bash
GET /api/files
```

#### 4. 删除文件
```bash
DELETE /api/file/delete
Content-Type: application/json

{
  "filename": "example.txt"
}
```

#### 5. 发送邮箱验证码
```bash
POST /api/auth/email/send
Content-Type: application/json

{
  "email": "user@example.com",
  "scene": "register"
}
```

#### 6. 邮箱验证码注册（手机号可选）
```bash
POST /api/auth/register-by-email
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "1234",
  "username": "testuser",
  "password": "password123",
  "nickname": "测试用户",
  "phone": ""
}
```

#### 7. 邮箱验证码重置密码
```bash
POST /api/auth/password/reset-by-email
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "1234",
  "new_password": "newpass123"
}
```

## 项目结构

```
agent/
├── models/
│   └── service/
│       ├── agent/           # Agent 核心逻辑
│       │   ├── agent.go     # Agent 主逻辑
│       │   ├── ollama.go    # Ollama 客户端
│       │   ├── zhipu.go     # 智谱清言服务
│       │   ├── tool.go      # 工具接口
│       │   └── file_tool.go # 文件工具实现
│       ├── weather/         # 天气工具
│       ├── calculator/      # 计算工具
│       ├── nbnhhsh/         # 缩写词猜测工具
│       └── file/            # 文件处理模块
├── webapi/
│   └── controllers/         # HTTP 控制器
├── router/
│   └── router.go           # 路由配置
├── static/                  # 前端静态文件
│   ├── index.html
│   ├── css/
│   └── js/
├── library/
│   └── log/                # 日志系统
├── global/
│   ├── config.go           # 配置定义
│   └── error.go            # 错误定义
├── data/                   # 上传文件存储目录
├── logs/                   # 日志存储目录
├── .env                    # 环境变量配置（不提交）
├── .env.example            # 环境变量模板
├── main.go                 # 程序入口
├── go.mod
└── README.md
```

## 配置说明

配置通过 `global/config.go` 和环境变量管理：

### 服务配置
| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| Server.Port | 25565 | 服务端口 |
| Server.Timeout | 30s | 请求超时 |

### Ollama 配置
| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| Ollama.Host | localhost:11434 | Ollama 地址 |
| Ollama.Model | gemma3:4b | 使用的模型 |
| Ollama.Temperature | 0.3 | 温度参数 |

### 智谱清言配置
| 配置项 | 默认值 | 说明 |
|-------|-------|------|
| Zhipu.Model | glm-4-flash | 智谱模型 |
| Zhipu.Timeout | 60s | 请求超时 |
| Zhipu.Enable | true | 是否启用后备 |

## 安全特性

### 密钥安全
- 所有敏感密钥通过环境变量配置
- `.env` 文件不会被提交到代码仓库
- 配置文件中不硬编码任何密钥

### 文件访问安全
- 路径验证：所有文件访问限制在 `./data` 目录
- 防目录穿透：检测 `../` 路径尝试
- 文件类型白名单：仅允许特定扩展名
- 大小限制：单文件最大 10MB

### API 安全
- 默认本地访问（可配置 CORS）
- 请求超时保护
- 错误信息隐藏敏感信息
- 登录态使用 JWT 校验
- 邮箱验证码包含过期时间、尝试次数与使用状态控制

## 故障排除

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
3. 系统会自动切换到智谱清言后备服务

### 智谱清言API调用失败
- 检查 `ZHIPU_API_KEY` 是否正确配置
- 检查API余额是否充足
- 查看日志获取详细错误信息

### 模型不存在
```
错误：模型 gemma3:4b 不可用
```

**解决方案：**
```bash
ollama pull gemma3:4b
```

### 天气工具失败
天气工具使用双API策略，如果主API失败会自动尝试备用API

### 缩写词工具不可用
- 确保 MySQL 服务可用
- 检查 `DATABASE_PASSWORD` 配置

## 日志

日志文件位于 `./logs/` 目录

查看实时日志：
```bash
tail -f ./logs/agent_*.log
```

## 测试

运行测试：
```bash
go test -v ./...
```

## 版本信息

查看版本信息：
```bash
./agent version
```

## 许可证

MIT License

## 致谢

- [Ollama](https://ollama.ai) - 本地 LLM 框架
- [智谱清言](https://open.bigmodel.cn/) - 云端大模型API
- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF 处理库
- [接口盒子](https://cn.apihz.cn/) - 天气数据来源
- [高德地图](https://lbs.amap.com/) - 天气数据来源