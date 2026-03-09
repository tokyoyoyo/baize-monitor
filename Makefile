# Baize Monitor Makefile

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=baize-agent

SERVER_BINARY_NAME=baize-server
BUILD_DIR=bin
AGENT_SOURCE_DIR=cmd/agent
SERVER_SOURCE_DIR=internal/server

# Go 工具
GO := go
GOFMT := gofmt
GOLINT := golangci-lint

# 颜色定义
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help build-agent build-server build clean clean-test test test-agent test-server test-coverage test-sequential fmt fmt-check lint tidy deps run-agent run-server

## 帮助信息
help:
	@echo "$(BLUE)================================$(NC)"
	@echo "$(BLUE)  BaiZe Monitor - Makefile Help $(NC)"
	@echo "$(BLUE)================================$(NC)"
	@echo ""
	@echo "$(GREEN)构建目标:$(NC)"
	@echo "  build-agent      - 构建 Agent 二进制文件"
	@echo "  build-server     - 构建 Server 二进制文件"
	@echo "  build            - 构建所有组件 (Agent + Server)"
	@echo ""
	@echo "$(GREEN)清理目标:$(NC)"
	@echo "  clean            - 清理构建产物"
	@echo "  clean-all        - 清理构建产物和缓存"
	@echo "  clean-test       - 清理测试缓存"
	@echo ""
	@echo "$(GREEN)测试目标:$(NC)"
	@echo "  test             - 运行所有测试（默认并发）"
	@echo "  test-agent       - 运行 Agent 测试"
	@echo "  test-server      - 运行 Server 测试"
	@echo "  test-sequential  - 顺序运行所有测试（推荐，避免数据库争用）"
	@echo "  test-coverage    - 运行测试并生成覆盖率报告"
	@echo ""
	@echo "$(GREEN)代码质量:$(NC)"
	@echo "  fmt              - 格式化代码"
	@echo "  fmt-check        - 检查代码格式（不修改）"
	@echo "  lint             - 运行代码检查"
	@echo "  tidy             - 清理依赖"
	@echo "  deps             - 下载依赖"
	@echo ""
	@echo "$(GREEN)运行目标:$(NC)"
	@echo "  run-agent        - 运行 Agent"
	@echo "  run-server       - 运行 Server"
	@echo ""
	@echo "$(GREEN)其他目标:$(NC)"
	@echo "  version          - 显示版本信息"
	@echo "  env              - 显示环境信息"
	@echo ""

## 构建 Agent
build-agent:
	@echo "$(BLUE)Building BaiZe Agent...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(AGENT_SOURCE_DIR)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## 构建 Server
build-server:
	@echo "$(BLUE)Building BaiZe Server...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) ./$(SERVER_SOURCE_DIR)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(SERVER_BINARY_NAME)$(NC)"

## 构建所有组件
build: build-agent build-server
	@echo "$(GREEN)All components built successfully$(NC)"

## 清理构建产物
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)Clean complete$(NC)"

## 清理所有（包括缓存）
clean-all: clean
	@echo "$(YELLOW)Cleaning Go cache...$(NC)"
	@$(GO) clean -cache -modcache
	@echo "$(GREEN)Full clean complete$(NC)"

## 清理测试缓存
clean-test:
	@echo "$(BLUE)Cleaning test cache...$(NC)"
	@$(GO) clean -testcache
	@echo "$(GREEN)Test cache cleaned$(NC)"

## 运行所有测试（默认并发）
test:
	@echo "$(BLUE)Running all tests...$(NC)"
	$(GO) test -v ./...

## 运行 Agent 测试
test-agent:
	@echo "$(BLUE)Running Agent tests...$(NC)"
	$(GO) test -v ./internal/agent/...

## 运行 Server 测试
test-server:
	@echo "$(BLUE)Running Server tests...$(NC)"
	$(GO) test -v ./internal/server/...

## 顺序运行所有测试（避免数据库争用）
test-sequential: clean-test
	@echo "$(BLUE)Running all tests sequentially (recommended for database tests)...$(NC)"
	$(GO) test -v ./... -p 1

## 运行测试并生成覆盖率报告
test-coverage: clean-test
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GO) test -v ./... -p 1 -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## 格式化代码
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' | xargs $(GOFMT) -w
	@echo "$(GREEN)Code formatted$(NC)"

## 检查代码格式
fmt-check:
	@echo "$(BLUE)Checking code format...$(NC)"
	@diff=$$(find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' | xargs $(GOFMT) -d); \
	if [ -n "$$diff" ]; then \
		echo "$(YELLOW)The following files have format issues:$(NC)"; \
		echo "$$diff"; \
		exit 1; \
	fi
	@echo "$(GREEN)Code format check passed$(NC)"

## 运行代码检查
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v $(GOLINT) > /dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## 清理依赖
tidy:
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	$(GO) mod tidy
	@echo "$(GREEN)Dependencies tidied$(NC)"

## 下载依赖
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO) mod download
	@echo "$(GREEN)Dependencies downloaded$(NC)"

## 运行 Agent
run-agent:
	@echo "$(BLUE)Starting BaiZe Agent...$(NC)"
	$(GO) run ./$(AGENT_SOURCE_DIR)

## 运行 Server
run-server:
	@echo "$(BLUE)Starting BaiZe Server...$(NC)"
	$(GO) run ./$(SERVER_SOURCE_DIR)

## 显示版本信息
version:
	@echo "$(BLUE)BaiZe Monitor Version Information$(NC)"
	@echo "Go Version: $$($(GO) version)"
	@echo "Module: $$(cat go.mod | head -1)"

## 显示环境信息
env:
	@echo "$(BLUE)Environment Information$(NC)"
	@echo "GOOS: $$($(GO) env GOOS)"
	@echo "GOARCH: $$($(GO) env GOARCH)"
	@echo "GOPATH: $$($(GO) env GOPATH)"
	@echo "GOMOD: $$($(GO) env GOMOD)"
	@echo "Go Version: $$($(GO) version)"
