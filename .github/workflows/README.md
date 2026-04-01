# CI/CD 流水线说明

本项目使用 GitHub Actions 实现 CI/CD 自动化流程。

## 工作流说明

### 1. CI（持续集成）- `.github/workflows/ci.yml`

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 针对这两个分支的 Pull Request

**包含任务：**
- **代码检查**：`go vet`、`gofmt` 格式检查
- **单元测试**：运行测试并生成覆盖率报告
- **多平台构建**：构建 Linux/macOS/Windows 多平台二进制文件
- **Docker 构建测试**：测试 Dockerfile 构建是否成功
- **安全扫描**：
  - Gosec：Go 代码安全扫描
  - Trivy：依赖漏洞扫描

**产物：**
- 上传构建的二进制文件（保留 7 天）
- 上传测试覆盖率到 Codecov
- 上传安全扫描结果到 GitHub Security 标签页

### 2. CD（持续部署）- `.github/workflows/cd.yml`

**触发条件：**
- 推送标签（如 `v1.0.0`）
- 手动触发

**包含任务：**
- **多平台 Docker 镜像构建和推送**：
  - 支持 linux/amd64 和 linux/arm64
  - 推送到 GitHub Container Registry (GHCR)
- **创建 GitHub Release**：
  - 自动构建所有平台的二进制文件
  - 创建压缩包并上传到 Release
  - 自动生成 Release Notes
- **部署到服务器**（可选）：
  - 通过 SSH 部署到配置的服务器
  - 使用 Docker 运行新版本

## 使用指南

### 1. 初次设置

#### 配置 Secrets（用于部署）

如果需要自动部署功能，需要在 GitHub 仓库中配置以下 Secrets：

1. 进入仓库的 `Settings` → `Secrets and variables` → `Actions`
2. 点击 `New repository secret` 添加以下 secrets：

```
DEPLOY_HOST        # 服务器地址
DEPLOY_USER        # SSH 用户名
DEPLOY_SSH_KEY     # SSH 私钥
DEPLOY_PORT        # SSH 端口（可选，默认 22）
```

#### 安装 Codecov（可选）

如果你想使用代码覆盖率功能：

1. 在 [Codecov](https://codecov.io/) 注册账号
2. 添加你的 GitHub 仓库
3. 获取上传令牌并添加到 GitHub Secrets：`CODECOV_TOKEN`

### 2. 日常使用

#### 开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/your-feature

# 2. 开发代码
vim your_file.go

# 3. 本地测试
make test
make lint

# 4. 提交并推送
git add .
git commit -m "feat: 添加新功能"
git push origin feature/your-feature

# 5. 在 GitHub 创建 Pull Request
# CI 会自动运行检查
```

#### 发布新版本

```bash
# 1. 确保在 main 分支且是最新的
git checkout main
git pull origin main

# 2. 创建标签（遵循语义化版本）
git tag v1.0.0
git push origin v1.0.0

# 3. CD 会自动执行：
#    - 构建多平台 Docker 镜像
#    - 推送到 GHCR
#    - 创建 GitHub Release
#    - 部署到服务器（如已配置）
```

#### 本地构建

使用 Makefile 进行本地构建：

```bash
# 查看所有可用命令
make help

# 构建当前平台
make build

# 构建所有平台
make build-all

# 运行测试
make test

# 代码检查
make lint

# 构建 Docker 镜像
make docker-build
```

## Docker 镜像使用

### 从 GHCR 拉取镜像

```bash
# 拉取最新版本
docker pull ghcr.io/your-username/agent:latest

# 拉取指定版本
docker pull ghcr.io/your-username/agent:v1.0.0

# 运行容器
docker run -d \
  --name agent \
  -p 8080:8080 \
  -v ./logs:/app/logs \
  ghcr.io/your-username/agent:latest
```

### 使用 Docker Compose

```bash
# 启动所有服务
make docker-run

# 查看日志
make docker-logs

# 停止服务
make docker-stop
```

## 环境变量配置

支持的环境变量：

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| SERVER_PORT | 服务端口 | 8080 |
| OLLAMA_HOST | Ollama 地址 | localhost:11434 |
| OLLAMA_MODEL | 模型名称 | qwen:7b |
| TZ | 时区 | Asia/Shanghai |

## 故障排查

### CI 失败

1. 查看 Actions 标签页中的失败日志
2. 常见问题：
   - 依赖下载失败：检查 `go.mod` 和 `go.sum` 是否正确
   - 测试失败：本地运行 `make test` 查看详细错误
   - 格式问题：运行 `make fmt` 修复格式

### CD 失败

1. 检查标签格式是否正确（如 `v1.0.0`）
2. 检查 Secrets 是否正确配置
3. 查看 SSH 连接和服务器日志

### Docker 构建失败

1. 确保所有依赖都在 `go.mod` 中
2. 检查 Dockerfile 语法
3. 本地测试：`docker build -t test .`

## 扩展和自定义

### 添加新的构建平台

修改 `.github/workflows/ci.yml` 中的 matrix：

```yaml
matrix:
  os: [linux, darwin, windows]
  arch: [amd64, arm64, riscv64]  # 添加 riscv64
```

### 添加新的部署目标

在 `.github/workflows/cd.yml` 中添加新的 job：

```yaml
deploy-k8s:
  name: Deploy to Kubernetes
  runs-on: ubuntu-latest
  needs: build-and-push
  steps:
    - uses: actions/checkout@v4
    - name: Deploy to K8s
      # 你的部署逻辑
```

## 最佳实践

1. **分支策略**：
   - `main`：生产环境
   - `develop`：开发环境
   - `feature/*`：功能分支

2. **提交信息**：使用语义化提交信息（如 `feat:`, `fix:`, `docs:`）

3. **版本号**：遵循语义化版本规范（SemVer）

4. **安全性**：
   - 定期更新依赖：`make update-deps`
   - 关注安全扫描结果
   - 不要在代码中硬编码密钥

## 联系方式

如有问题，请提交 Issue 或 Pull Request。
