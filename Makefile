# Makefile for Agent Project

# 变量定义
APP_NAME=agent
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO=go
GOFLAGS=-ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -w -s"

# API Key（可通过 make release ZHIPU_KEY=xxx 注入）
ZHIPU_KEY?=
WEATHER_ID?=
WEATHER_KEY?=
AMAP_KEY?=

# 颜色输出
COLOR_RESET=\033[0m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: help
help: ## 显示帮助信息
	@echo "$(COLOR_BLUE)Agent 项目构建脚本$(COLOR_RESET)"
	@echo ""
	@echo "使用方法: make [目标]"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: run
run: ## 运行应用
	@echo "$(COLOR_BLUE)正在启动应用...$(COLOR_RESET)"
	$(GO) run main.go

.PHONY: build
build: ## 构建应用（当前平台）
	@echo "$(COLOR_BLUE)正在构建应用...$(COLOR_RESET)"
	$(GO) build $(GOFLAGS) -o $(APP_NAME) .
	@echo "$(COLOR_GREEN)构建完成: $(APP_NAME)$(COLOR_RESET)"

.PHONY: build-with-key
build-with-key: ## 构建应用并内置智谱API Key（用法: make build-with-key ZHIPU_KEY=your_key）
	@echo "$(COLOR_BLUE)正在构建应用（内置API Key）...$(COLOR_RESET)"
	@test -n "$(ZHIPU_KEY)" || { echo "$(COLOR_YELLOW)错误: 请提供 ZHIPU_KEY 参数$(COLOR_RESET)"; exit 1; }
	$(GO) build -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.ZhipuAPIKey=$(ZHIPU_KEY) -w -s" -o $(APP_NAME) .
	@echo "$(COLOR_GREEN)构建完成: $(APP_NAME)（已内置智谱API Key）$(COLOR_RESET)"

.PHONY: release
release: ## 构建完整发布版本（内置前端+所有API Key）
	@echo "$(COLOR_BLUE)正在构建发布版本...$(COLOR_RESET)"
	@test -n "$(ZHIPU_KEY)" || { echo "$(COLOR_YELLOW)错误: 请提供 ZHIPU_KEY 参数$(COLOR_RESET)"; exit 1; }
	$(GO) build -ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.ZhipuAPIKey=$(ZHIPU_KEY) -X main.WeatherAPIID=$(WEATHER_ID) -X main.WeatherAPIKey=$(WEATHER_KEY) -X main.AmapAPIKey=$(AMAP_KEY) -w -s" -o $(APP_NAME) .
	@echo "$(COLOR_GREEN)构建完成: $(APP_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)已内置: 前端静态文件 + API Keys$(COLOR_RESET)"
	@ls -lh $(APP_NAME)

.PHONY: build-all
build-all: ## 构建所有平台的二进制文件
	@echo "$(COLOR_BLUE)正在构建所有平台...$(COLOR_RESET)"
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(APP_NAME)-linux-amd64 .
	@GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/$(APP_NAME)-linux-arm64 .
	@GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(APP_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/$(APP_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe .
	@echo "$(COLOR_GREEN)构建完成，文件位于 dist/ 目录$(COLOR_RESET)"

.PHONY: test
test: ## 运行测试
	@echo "$(COLOR_BLUE)正在运行测试...$(COLOR_RESET)"
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test-coverage
test-coverage: test ## 生成测试覆盖率报告
	@$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "$(COLOR_GREEN)覆盖率报告已生成: coverage.html$(COLOR_RESET)"

.PHONY: lint
lint: ## 运行代码检查
	@echo "$(COLOR_BLUE)正在检查代码...$(COLOR_RESET)"
	@$(GO) vet ./...
	@test -z "$$(gofmt -l .)" || { echo "$(COLOR_YELLOW)以下文件需要格式化:$(COLOR_RESET)"; gofmt -d .; exit 1; }
	@echo "$(COLOR_GREEN)代码检查通过$(COLOR_RESET)"

.PHONY: fmt
fmt: ## 格式化代码
	@echo "$(COLOR_BLUE)正在格式化代码...$(COLOR_RESET)"
	$(GO) fmt ./...

.PHONY: clean
clean: ## 清理构建文件
	@echo "$(COLOR_BLUE)正在清理...$(COLOR_RESET)"
	@rm -f $(APP_NAME)
	@rm -rf dist/
	@rm -f coverage.txt coverage.html
	@echo "$(COLOR_GREEN)清理完成$(COLOR_RESET)"

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	@echo "$(COLOR_BLUE)正在构建 Docker 镜像...$(COLOR_RESET)"
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(COLOR_GREEN)Docker 镜像构建完成$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	@echo "$(COLOR_BLUE)正在启动 Docker 容器...$(COLOR_RESET)"
	docker-compose up -d

.PHONY: docker-stop
docker-stop: ## 停止 Docker 容器
	@echo "$(COLOR_BLUE)正在停止 Docker 容器...$(COLOR_RESET)"
	docker-compose down

.PHONY: docker-logs
docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

.PHONY: install-deps
install-deps: ## 安装依赖
	@echo "$(COLOR_BLUE)正在安装依赖...$(COLOR_RESET)"
	$(GO) mod download
	$(GO) mod tidy

.PHONY: update-deps
update-deps: ## 更新依赖
	@echo "$(COLOR_BLUE)正在更新依赖...$(COLOR_RESET)"
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: security
security: ## 安全扫描
	@echo "$(COLOR_BLUE)正在运行安全扫描...$(COLOR_RESET)"
	@command -v gosec >/dev/null 2>&1 && gosec ./... || echo "$(COLOR_YELLOW)gosec 未安装，跳过安全扫描$(COLOR_RESET)"

.PHONY: version
version: ## 显示版本信息
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

.PHONY: dev
dev: ## 开发模式（运行并监控）
	@echo "$(COLOR_BLUE)启动开发模式...$(COLOR_RESET)"
	@command -v air >/dev/null 2>&1 && air || { echo "$(COLOR_YELLOW)air 未安装，请运行: go install github.com/cosmtrek/air@latest$(COLOR_RESET)"; $(GO) run main.go; }