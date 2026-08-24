# Agent — Go 智能体服务

> **当前版本**：`dev-001-agent`（尚未引入语义化版本 tag，暂以开发分支名标记；发布 tag 后替换为 `vX.Y.Z`）

基于 Go 原生 `net/http` 构建的智能体服务：以本地 Ollama 为主模型、智谱清言云端 API 为后备，内置天气查询、数学计算、文件处理、缩写词猜测等工具，并提供 JWT 认证与三级角色权限、会话持久化、流式深度思考、AI 角色卡对话等能力。前端为内置静态页面，可编译进单一二进制文件。

## 核心功能

### 大模型服务
- **Ollama 本地模型**：主要使用本地 Ollama 服务（默认 `gemma3:4b`）
- **智谱清言后备**：Ollama 不可用时自动切换到云端 API（免费模型 `glm-4-flash`）
- **付费推理模型**：配置 `ZHIPU_NEW_API_KEY` 后，深度思考模式使用 `glm-4.5` 原生推理能力
- **无缝切换**：用户无感知，保证服务持续可用

### 统一鉴权与角色权限
- **JWT 认证**：登录成功返回 Token，有效期 24 小时；受保护接口需携带 `Authorization: Bearer <token>`
- **三级角色**：`user` / `vip` / `admin`
- **账号状态校验**：`status != 1` 的账号一律拒绝访问
- **多种注册方式**：用户名密码、邮箱验证码、手机号验证码（注册时手机号非必填）
- **找回密码**：支持邮箱验证码与手机号验证码两种重置路径

### 对话模式与会话持久化
- **普通模式**（`normal`）：非流式，所有登录用户可用
- **深度思考模式**（`thinking`）：SSE 流式输出思维链与答案，仅 VIP / 管理员可用
- **Auto 模式**（`auto`）：仅管理员可用
- **会话持久化**：会话与消息落库 MySQL，服务端按 `session_id` 自动加载最近 12 条历史作为上下文
- **上下文预算**：通过 `CONTEXT_BUDGET` 控制，可选 `2k` / `4k` / `8k` / `16k` / `32k` / `64k`，留空或非法值均按 `64k`（64000 字符 / 40 条消息）处理
- **自动标题**：首轮对话后异步生成会话标题

### AI 角色卡（仅 VIP / 管理员）
- 角色卡的创建、编辑、软删除与详情查询
- 支持公开 / 私有可见性、对话示例（few-shot）配置
- 角色卡对话拥有独立的会话列表与历史记录
- 支持流式与非流式两种对话方式

### 内容安全与限流
- **屏蔽词审核**：角色卡创建时校验，支持中文与拼音匹配、不区分大小写，词表可通过 `BLOCKED_WORDS` 覆盖
- **频率限制**：普通用户 5 分钟 1 次，VIP 用户 1 分钟 2 次，管理员无限制；超限返回 `errno=429`

### 天气查询工具（Weather）
- 支持 **100+ 中国城市** 天气查询
- 支持多天预报（今天、明天、后天、七天）
- 双 API 策略：接口盒子 API（中国气象局数据）→ 高德天气 API
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
- 猜测网络缩写词含义，支持拼音首字母缩写
- 结果写入 MySQL 缓存，需要数据库可用
- 示例：
  ```
  "yyds是什么意思"
  "xswl的含义"
  "awsl"
  ```

### 智能文件处理工具（FileTool）

文件读写限制在 `./data` 目录内，支持扩展名：`.txt`、`.md`、`.json`、`.go`、`.py`、`.js`、`.pdf`

支持的操作：

1. **文件解析与总结**（`parse`）
   ```
   "帮我总结一下 data/example.txt 文件的内容"
   "阅读 data/example.json"
   ```

2. **代码分析**（`code_analyze`）— Explain（解释）/ Error（错误检查）/ Optimize（性能优化）
   ```
   "分析一下 data/example.py 文件"
   ```

3. **格式转换**（`convert`）— 仅支持以下四种组合：
   - `.md` → `html`
   - `.md` → `word`
   - `.json` → `csv`
   - `.txt` / `.md` → `mindmap`
   ```
   "把 data/example.md 转换成 HTML"
   ```

### 文件上传与管理
- 前端直接上传文件，单文件最大 10MB
- 支持文件列表查看与删除
- 文件类型白名单校验，防目录穿越
- Web UI 集成文件管理面板

## 快速开始

### 前提条件
- **Go 1.25.0** 或更高版本（见 `go.mod`）
- **MySQL 5.7+**：会话持久化、用户体系、角色卡、缩写词缓存均依赖数据库
- **Ollama 服务**（可选，推荐）：默认地址 `localhost:11434`，已拉取 `gemma3:4b` 或其他兼容模型
  - 若 Ollama 不可用，服务会自动降级到智谱清言后备，此时必须配置 `ZHIPU_API_KEY`

### 安装和运行

1. **克隆项目**
```bash
git clone <repo-url>
cd agent
```

2. **配置环境变量**
```bash
cp .env.example .env
```

编辑 `.env`：

必填：
- `ZHIPU_API_KEY` — 智谱清言 API 密钥（Ollama 不可用时的后备服务）
- `JWT_SECRET` — 登录签名密钥
- `DATABASE_PASSWORD` — MySQL 密码

可选：
- `ZHIPU_NEW_API_KEY` — 智谱付费 API 密钥，启用后深度思考模式使用 `glm-4.5`
- `WEATHER_API_ID` / `WEATHER_API_KEY` — 接口盒子天气 API
- `AMAP_API_KEY` — 高德天气 API
- `CONTEXT_BUDGET` — 对话上下文预算，`2k` / `4k` / `8k` / `16k` / `32k` / `64k`，默认 `64k`
- `BLOCKED_WORDS` — 屏蔽词，逗号分隔
- `MAIL_SMTP_*` — 邮箱验证码 SMTP 配置

3. **安装依赖**
```bash
go mod download
```

4. **构建并运行**
```bash
go build -o agent .
./agent
```

服务监听 `0.0.0.0:25565`，本地访问 `http://localhost:25565`

### 使用 Makefile

```bash
make help            # 查看所有可用目标
```

构建与运行：

| 目标 | 说明 |
|------|------|
| `make run` | `go run main.go` 直接运行 |
| `make build` | 构建当前平台二进制（注入 Version / BuildTime） |
| `make build-with-key ZHIPU_KEY=xxx` | 构建并将智谱 API Key 编译进二进制 |
| `make release ZHIPU_KEY=xxx [WEATHER_ID=] [WEATHER_KEY=] [AMAP_KEY=]` | 发布版本：内置前端静态文件 + 所有 API Key |
| `make build-all` | 交叉编译 linux/darwin（amd64+arm64）与 windows/amd64 到 `dist/` |
| `make dev` | 开发模式，存在 `air` 时热重载 |
| `make clean` | 清理二进制、`dist/` 与覆盖率文件 |

质量与依赖：

| 目标 | 说明 |
|------|------|
| `make test` | `go test -race` 并生成 `coverage.txt` |
| `make test-coverage` | 生成 HTML 覆盖率报告 `coverage.html` |
| `make lint` | `go vet` + `gofmt` 检查 |
| `make fmt` | 格式化代码 |
| `make security` | gosec 安全扫描（未安装则跳过） |
| `make install-deps` / `make update-deps` | 安装 / 更新依赖 |
| `make version` | 显示 Version 与 BuildTime |

Docker：

| 目标 | 说明 |
|------|------|
| `make docker-build` | 构建镜像并打 `latest` 标签 |
| `make docker-run` | `docker-compose up -d` |
| `make docker-stop` | `docker-compose down` |
| `make docker-logs` | 跟随查看容器日志 |

> `make release` 会把 `static/` 下的前端资源通过 `static_embed.go` 编译进二进制，产出物可脱离源码目录独立运行。

### Docker 运行

```bash
# 使用 docker-compose
docker-compose up -d

# 或手动构建
docker build -t agent .
docker run -p 25565:25565 --env-file .env agent
```

## 环境变量配置

在项目根目录创建 `.env` 文件（程序启动时自动加载）：

```bash
# ========== 必填配置 ==========

# 智谱清言 API（Ollama 不可用时的后备服务）
# 获取地址: https://open.bigmodel.cn/
ZHIPU_API_KEY=your_zhipu_api_key

# 登录签名密钥
JWT_SECRET=your_super_secret_jwt_key

# MySQL 密码
DATABASE_PASSWORD=your_db_password

# ========== 可选配置 ==========

# 智谱付费 API 密钥。配置后深度思考模式使用 glm-4.5 原生 reasoning
# 留空、NULL、null 均视为未配置
ZHIPU_NEW_API_KEY=

# 对话上下文预算：2k / 4k / 8k / 16k / 32k / 64k（留空即默认 64k）
CONTEXT_BUDGET=64k

# 屏蔽词，逗号分隔，用于角色卡创建时的内容审核
# 支持中文与拼音匹配，不区分大小写
BLOCKED_WORDS=loser

# 天气 API（接口盒子）
WEATHER_API_ID=your_weather_api_id
WEATHER_API_KEY=your_weather_api_key

# 高德天气 API
# 获取地址: https://lbs.amap.com/
AMAP_API_KEY=your_amap_api_key

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

# 测试用：请求头携带 key: <此值> 可绕过 /api/chat 的登录校验
# 生产环境请勿设置
CHAT_BYPASS_KEY=
```

**注意事项：**
- `.env` 已被 `.gitignore` 忽略，不会提交到代码仓库
- `.env.example` 中部分默认值与代码不一致（`SERVER_PORT=8080`、`OLLAMA_MODEL=qwen:7b`、`DATABASE_DBNAME=agent`）。实际生效值以代码为准：端口 `25565`、模型 `gemma3:4b`、库名 `goagent`
- `CHAT_BYPASS_KEY` 会绕过鉴权，仅用于本地测试

## 使用指南

### Web 界面

访问 `http://localhost:25565`

| 页面 | 路径 |
|------|------|
| 主界面（聊天 / 文件管理 / 角色卡） | `/index.html` |
| 登录 | `/login.html` |
| 注册 | `/register.html` |
| 忘记密码 | `/forgot-password.html` |
| 管理后台 | `/admin.html` |

### 统一响应格式

除 SSE 流式接口外，所有接口**返回的 HTTP 状态码恒为 200**，业务结果通过 `errno` 判断：

```json
{
  "data": { },
  "errmsg": "",
  "errno": 0,
  "logid": "xxxxxxxx"
}
```

| 字段 | 说明 |
|------|------|
| `errno` | 业务错误码，`0` 表示成功；`401` 未登录 / 登录过期、`403` 无权限、`429` 频率超限、`404` 参数或方法错误、`500` 服务内部错误 |
| `errmsg` | 错误描述，成功时为空 |
| `data` | 业务数据，失败时可能不返回 |
| `logid` | 本次请求日志 ID，排查问题时提供该值 |

### 鉴权方式

先调用登录或注册接口获得 Token，随后在受保护接口的请求头中携带：

```
Authorization: Bearer <token>
```

Token 有效期 24 小时。账号 `status` 必须为 `1`，否则一律返回 `errno=401`。

---

### 一、公开接口（无需鉴权）

#### 健康检查
```bash
GET /api/health
```

#### 用户名密码注册 / 登录
```bash
POST /api/auth/register
Content-Type: application/json

{"username": "testuser", "password": "password123", "nickname": "测试用户"}
```

```bash
POST /api/auth/login
Content-Type: application/json

{"username": "testuser", "password": "password123"}
```

登录成功后 `data` 中返回 Token 与用户信息。

#### 短信验证码注册 / 重置密码

| 接口 | 请求体 |
|------|--------|
| `POST /api/auth/sms/send` | `{"phone":"13800000000","scene":"register"}` |
| `POST /api/auth/register-by-phone` | `{"phone":"...","code":"123456","username":"...","password":"...","nickname":"..."}` |
| `POST /api/auth/password/reset` | `{"phone":"...","code":"123456","new_password":"..."}` |

`scene` 用于区分验证码用途（注册 / 重置密码）。

#### 邮箱验证码注册 / 重置密码

| 接口 | 请求体 |
|------|--------|
| `POST /api/auth/email/send` | `{"email":"a@b.com","scene":"register"}` |
| `POST /api/auth/register-by-email` | `{"email":"...","code":"123456","username":"...","password":"...","nickname":"...","phone":"..."}` |
| `POST /api/auth/password/reset-by-email` | `{"email":"...","code":"123456","new_password":"..."}` |

#### 文件上传与管理

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/upload` | `multipart/form-data`，字段名 `file`，单文件上限 10MB，白名单 `.txt .md .json .go .py .js .pdf`，存放于 `./data` |
| GET | `/api/files` | 列出 `./data` 目录下已上传文件 |
| POST | `/api/file/delete?filename=xxx.md` | 删除文件，文件名通过 **query 参数**传入 |

> 注意：删除接口是 **POST**，不是 DELETE；文件名走 query，不是请求体。

#### 工具直连接口（无需鉴权）

用于本地调试单个工具，不经过大模型编排：

| 方法 | 路径 |
|------|------|
| GET | `/internal/tools/list` |
| POST | `/internal/tools/weather` |
| POST | `/internal/tools/calculator` |
| POST | `/internal/tools/file` |
| POST | `/internal/tools/nbnhhsh` |
| POST | `/internal/tools/execute/{toolName}` |

---

### 二、对话接口（需鉴权）

#### 非流式对话

```bash
POST /api/chat
Authorization: Bearer <token>
Content-Type: application/json

{
  "message": "今天西安几度",
  "session_id": "",
  "mode": "thinking"
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `message` | 是 | 用户输入 |
| `session_id` | 否 | 留空则新建会话，返回新的 `session_id` |
| `mode` | 否 | `normal` / `thinking` / `auto`，默认 `thinking` |

响应 `data`：

```json
{
  "result": "西安当前 28℃……",
  "session_id": "sess_xxx",
  "title": "西安天气"
}
```

> 请求体中**没有 `history` 字段**。历史上下文由服务端根据 `session_id` 自动从数据库加载，最近 12 条消息参与上下文。

#### 流式对话（SSE）

```bash
POST /api/chat/stream?mode=thinking
Authorization: Bearer <token>
```

`mode` 通过 **query 参数**传入，并由 `StreamModeMiddleware` 做权限校验：

| mode | 说明 | 允许角色 |
|------|------|----------|
| `normal` | 普通问答 | user / vip / admin |
| `thinking` | 深度思考（默认值） | vip / admin |
| `auto` | 自动选择模式 | admin |

不在允许范围内返回 `errno=403`；传入其他值返回 `errno=400`（`不支持的 mode 参数`）。

响应为 `text/event-stream`，帧格式：

```
event: <type>
data: <json>

```

事件类型共 9 种：

| type | 含义 |
|------|------|
| `session` | 下发本轮会话 ID |
| `thought` | 思考过程增量文本 |
| `answer` | 回答增量文本 |
| `answer_reset` | 清空已输出回答并重新开始 |
| `tool_call` | 触发工具调用（含 `tool` / `tool_input`） |
| `tool_result` | 工具返回结果 |
| `title` | 自动生成的会话标题 |
| `done` | 本轮结束 |
| `error` | 出错信息 |

响应头包含 `Cache-Control: no-cache`、`Connection: keep-alive`、`X-Accel-Buffering: no`（禁用 Nginx 缓冲）。

#### 会话列表与历史

```bash
POST /api/getchatlist
{"page": 1, "size": 20}
```

```bash
POST /api/getchathistory
{"session_id": "sess_xxx", "limit": 50}
```

#### 当前用户信息

```bash
GET /api/auth/me
Authorization: Bearer <token>
```

返回 `data`：`{"id": 1, "username": "testuser", "role": "vip"}`

---

### 三、AI 角色卡（需鉴权 + VIP 及以上）

该分组挂载了 `VIPMiddleware`，`user` 角色访问一律返回 `errno=403`。

| 方法 | 路径 | 参数 |
|------|------|------|
| GET | `/api/persona/list` | query：`type`（默认 `public`）、`page`（默认 1）、`size`（默认 10，上限 50） |
| GET | `/api/persona/detail` | query：`id`（必填） |
| POST | `/api/persona/create` | body：`{name, avatar, tagline, personality, system_prompt, is_public, examples}` |
| POST | `/api/persona/update` | body：同上 + `id`（必填） |
| POST | `/api/persona/delete` | body：`{"id": 1}`，软删除（`status = -1`） |
| POST | `/api/persona/chat` | body：`{persona_id, session_id?, message}` |
| POST | `/api/persona/chat/stream` | body：同上，返回 SSE 流 |
| GET | `/api/persona/sessions` | query：`persona_id` |
| GET | `/api/persona/history` | query：`session_id`（必填） |

`examples` 为对话示例数组，元素形如 `{"user_input": "...", "ai_response": "..."}`，用于给角色提供 few-shot 样例。

`creator_id` 为 `NULL` 的角色卡是系统预置角色。

---

### 四、管理员接口（需鉴权 + admin）

该分组挂载了 `AdminMiddleware`，非 `admin` 角色返回 `errno=403`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/users` | 用户列表 |
| POST | `/api/admin/users/{id}/role` | 修改用户角色（`user` / `vip` / `admin`） |
| POST | `/api/admin/users/{id}/status` | 启用 / 禁用用户（`1` 正常，`0` 禁用） |

路径参数 `{id}` 使用 Go 1.22+ `http.ServeMux` 的原生路径匹配能力。

---

## 项目结构

```
agent/
├── main.go                      # 程序入口
├── static_embed.go              # go:embed 将 static/ 打包进单一二进制
├── app/
│   └── app.go                   # 应用装配：初始化配置、DAO、Service、Controller、路由
├── global/
│   ├── config.go                # 配置结构体与默认值（环境变量覆盖）
│   └── error.go                 # 统一错误码
├── library/
│   ├── init.go                  # 基础库初始化
│   └── log/log.go               # 日志组件（按天切分 agent_YYYY-MM-DD.log）
├── router/
│   ├── router.go                # 路由总入口，创建 GroupRouter 并注册各分组
│   └── routes/
│       ├── grouprouter.go       # GroupRouter：前缀 + 中间件继承
│       ├── middleware.go        # AuthMiddleware / AdminMiddleware / VIPMiddleware / StreamModeMiddleware
│       ├── auth.go              # /api/auth 认证路由
│       ├── chat.go              # /api/chat 对话与文件路由
│       ├── persona.go           # /api/persona 角色卡路由
│       ├── admin.go             # /api/admin 管理员路由
│       └── internal.go          # /internal/tools 工具直调路由
├── webapi/controllers/
│   ├── controller.go            # ChatController、HealthController、统一 Response 与 SSE 写入
│   ├── auth_controller.go       # 注册/登录/短信/邮箱/重置密码/当前用户
│   ├── persona_controller.go    # 角色卡增删改查与对话
│   ├── admin_controller.go      # 用户管理
│   ├── file_controller.go       # 上传/列表/删除
│   └── tool_controller.go       # 工具直调
├── models/
│   ├── entity/                  # chat.go、persona.go 请求与响应结构体
│   ├── dao/                     # mysql.go、user_dao、chat_dao、persona_*、sms_code、email_code、rate_limit
│   └── service/
│       ├── agent/               # 核心编排：ollama.go、zhipu.go、tool.go、stream_parser.go、stream_types.go、file_tool.go
│       ├── auth/                # auth.go、rate_limiter.go、sms_service.go、email_service.go
│       ├── blocker/             # 屏蔽词审核
│       ├── persona/             # persona_service.go、chat_service.go、context_builder.go
│       ├── weather/             # 天气工具
│       ├── calculator/          # 计算器工具
│       ├── file/                # 文件读取与内容解析
│       └── nbnhhsh/             # 缩写词猜义工具
├── utils/utils.go
├── static/                      # 前端：index / login / register / forgot-password / admin + css/ js/
├── Dockerfile / docker-compose.yml / nginx.conf
├── Makefile / .air.toml / .env.example
└── .github/workflows/           # CI / CD
```

---

## 配置说明

以下为 `global/config.go` 中 `DefaultConfig` 的实际默认值，环境变量可覆盖。

### Server

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `Port` | `25565` | 监听端口 |
| `ReadTimeout` | 60s | 请求读取超时 |
| `WriteTimeout` | 120s | 响应写入超时（为 LLM 长响应预留） |
| `IdleTimeout` | 120s | 连接空闲超时 |

> 代码中并不存在 `Server.Timeout` 这一配置项，超时由上述三项分别控制。

### Ollama（本地模型）

| 配置项 | 默认值 |
|--------|--------|
| `Host` | `localhost:11434` |
| `Model` | `gemma3:4b` |
| `Timeout` | 120s |
| `Temperature` | 0.3 |

### Zhipu（云端后备模型）

| 配置项 | 默认值 |
|--------|--------|
| `Model` | `glm-4-flash`（免费） |
| `ReasoningModel` | `glm-4.5`（付费，需 `ZHIPU_NEW_API_KEY`） |
| `BaseURL` | `https://open.bigmodel.cn/api/paas/v4/chat/completions` |
| `Timeout` | 120s |
| `Temperature` | 0.3 |
| `Enable` | `true` |

`ZhipuConfig.IsReasoningEnabled()` 判定逻辑：`NewAPIKey` 非空且不等于 `NULL` / `null` 时才启用推理模型。

### Database

| 配置项 | 默认值 |
|--------|--------|
| `Host` / `Port` | `localhost` / `3306` |
| `User` / `Password` | `root` / 空 |
| `DBName` | `goagent` |
| `MaxOpen` / `MaxIdle` | 10 / 5 |

### JWT 与 Blocker

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `JWT.Secret` | 空 | 必须通过 `JWT_SECRET` 环境变量提供 |
| `JWT.ExpireTime` | 24h | Token 有效期 |
| `Blocker.Words` | 内置默认词表 | 可通过 `BLOCKED_WORDS` 覆盖，逗号分隔 |

---

## 安全特性

- **JWT 鉴权**：`Authorization: Bearer <token>`，有效期 24 小时；`JWT_SECRET` 由环境变量注入，不落代码
- **三级 RBAC**：`user` / `vip` / `admin`，分别由 `AuthMiddleware`、`VIPMiddleware`、`AdminMiddleware` 与 `StreamModeMiddleware` 把关
- **账号状态校验**：`status != 1` 的账号即使持有有效 Token 也会被拒（`errno=401`）
- **频率限制**：按角色滑动窗口限流，超限返回 `errno=429`

  | 角色 | 窗口 | 允许次数 |
  |------|------|----------|
  | `user` | 5 分钟 | 1 |
  | `vip` | 1 分钟 | 2 |
  | `admin` | — | 不限制 |

- **屏蔽词审核**：角色卡内容创建时校验，支持中文与拼音匹配、不区分大小写
- **上传限制**：白名单扩展名 `.txt .md .json .go .py .js .pdf`，单文件 10MB，统一落在 `./data`
- **容器安全**：Docker 镜像以非 root 用户运行，多阶段构建减小攻击面

> `CHAT_BYPASS_KEY` 会让携带匹配 `key` 请求头的请求跳过鉴权，仅用于本地测试，生产环境务必不要配置。

## 日志

日志按天切分，文件名形如：

```bash
tail -f ./logs/agent_$(date +%F).log
# 或查看全部
tail -f ./logs/agent_*.log
```

每条响应都会返回 `logid`，排查问题时用它在日志中检索对应请求链路。

---

## 故障排除

**服务启动失败：数据库连接错误**
检查 MySQL 是否运行、`DATABASE_*` 环境变量是否正确。注意默认库名是 `goagent`（不是 `agent`）。

**返回 `errno=401`**
Token 缺失、格式不对（必须是 `Bearer <token>`）、已过期（24 小时），或账号 `status` 不为 `1`。

**返回 `errno=403`**
角色权限不足：`thinking` 模式需要 VIP 及以上，`auto` 模式仅管理员，`/api/persona/*` 需要 VIP 及以上，`/api/admin/*` 仅管理员。

**返回 `errno=429`**
触发角色频率限制，等待窗口结束后重试，或提升账号角色。

**深度思考模式无思维链输出**
确认请求走的是 `/api/chat/stream` 且 `mode=thinking`；若需要 `glm-4.5` 原生推理，需配置 `ZHIPU_NEW_API_KEY`，否则回退到 `glm-4-flash`。

**流式响应被缓冲、前端收不到增量**
若经过 Nginx 等反向代理，确认未开启响应缓冲（服务端已下发 `X-Accel-Buffering: no`）。

**Ollama 相关报错**
确认 `ollama serve` 已启动、`gemma3:4b` 已 `ollama pull`。若本地模型不可用且 `ZHIPU_API_KEY` 已配置，会自动回退到云端模型。

**日志排查**
日志目录默认 `./logs`，按天切分为 `agent_YYYY-MM-DD.log`。用响应中的 `logid` 检索对应请求。

## 许可证

仓库中当前**没有 LICENSE 文件**，尚未声明开源许可证。如需对外分发或二次使用，请先补充许可证文件并在此处更新说明。












