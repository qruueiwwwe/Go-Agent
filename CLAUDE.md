# Agent 项目 CI/CD 配置说明

## 已创建的文件

我为你的项目补充了完整的 CI/CD 流水线配置，包括以下文件：

### 1. `.gitignore` - Git 忽略文件配置
- 忽略编译产物（二进制文件、临时文件）
- 忽略日志文件
- 忽略 IDE 配置文件
- 忽略环境变量文件

### 2. `Dockerfile` - Docker 镜像构建文件
- 采用多阶段构建，优化镜像大小
- 基于 Alpine Linux，最终镜像非常轻量
- 包含健康检查
- 使用非 root 用户运行，提高安全性
- 支持 GHCR（GitHub Container Registry）

### 3. `docker-compose.yml` - Docker Compose 配置
- Agent 服务
- Ollama 服务（可选）
- MySQL 服务（可选）
- Redis 服务（可选）
- Nginx 反向代理（可选）
- 网络和数据卷配置

### 4. `.github/workflows/ci.yml` - 持续集成工作流
包含以下任务：
- 代码检查（go vet、gofmt）
- 单元测试（带覆盖率）
- 多平台构建（Linux、macOS、Windows）
- Docker 构建测试
- 安全扫描（Gosec、Trivy）

### 5. `.github/workflows/cd.yml` - 持续部署工作流
包含以下任务：
- 多平台 Docker 镜像构建和推送
- GitHub Release 自动创建
- 服务器自动部署（可选）

### 6. `Makefile` - 构建工具
提供便捷的命令：
- `make run` - 运行应用
- `make build` - 构建应用
- `make build-all` - 构建所有平台
- `make test` - 运行测试
- `make lint` - 代码检查
- `make docker-build` - 构建 Docker 镜像
- 等等...

### 7. `nginx.conf` - Nginx 配置文件
- 反向代理配置
- 静态文件服务
- API 路由配置
- 健康检查端点

### 8. `.air.toml` - 热重载配置
- 用于开发环境的热重载
- 监控文件变化自动重启

### 9. `.env.example` - 环境变量示例
- 所有可配置的环境变量
- 默认值说明

### 10. `.github/workflows/README.md` - 详细文档
- CI/CD 流程说明
- 使用指南
- 故障排查

## 使用步骤

### 1. 提交代码到 GitHub

```bash
git add .
git commit -m "feat: 添加 CI/CD 配置"
git push origin main
```

### 2. 配置 Secrets（可选）

如果需要自动部署功能，在 GitHub 仓库中配置以下 Secrets：

- `DEPLOY_HOST` - 服务器地址
- `DEPLOY_USER` - SSH 用户名
- `DEPLOY_SSH_KEY` - SSH 私钥

### 3. 触发 CI 工作流

推送代码后，CI 会自动运行：
- 进入 GitHub 仓库的 Actions 标签页
- 查看工作流执行状态

### 4. 发布新版本

```bash
git tag v1.0.0
git push origin v1.0.0
```

CD 会自动执行：
- 构建多平台 Docker 镜像
- 推送到 GHCR
- 创建 GitHub Release

### 5. 本地开发

```bash
# 查看所有可用命令
make help

# 运行应用
make run

# 运行测试
make test

# 构建所有平台
make build-all

# 使用 Docker Compose
make docker-run
```

## Docker 镜像使用

### 拉取镜像

```bash
# 拉取最新版本
docker pull ghcr.io/your-username/agent:latest

# 拉取指定版本
docker pull ghcr.io/your-username/agent:v1.0.0
```

### 运行容器

```bash
docker run -d \
  --name agent \
  -p 8080:8080 \
  -v ./logs:/app/logs \
  ghcr.io/your-username/agent:latest
```

## 功能特性

✅ 自动化测试和构建
✅ 多平台支持（Linux/macOS/Windows）
✅ Docker 镜像自动构建和推送
✅ GitHub Release 自动创建
✅ 代码质量检查
✅ 安全扫描
✅ 测试覆盖率报告
✅ 服务器自动部署（可选）
✅ 多架构支持（amd64/arm64）
✅ 热重载开发模式

## 注意事项

1. 首次使用需要在 GitHub 上启用 Actions
2. 如果使用 Codecov，需要注册账号并配置令牌
3. 确保本地已安装 Docker 和 Docker Compose
4. 如需自动部署，需要先配置服务器 SSH 访问

## 下一步建议

1. 根据实际需求修改 `docker-compose.yml` 中的服务配置
2. 调整 CI/CD 工作流中的检查规则
3. 配置代码审查保护规则
4. 设置环境特定的配置文件
5. 考虑添加集成测试
6. 配置通知机制（邮件、Slack 等）

祝你的项目 CI/CD 流程运行顺利！
