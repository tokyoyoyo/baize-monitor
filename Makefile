# Baize Monitor Makefile

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=baize-agent
BUILD_DIR=bin
SOURCE_DIR=cmd/agent

# 构建Agent
build-agent:
	@echo "Building BaiZe Agent..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./$(SOURCE_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  build-agent  - Build the BaiZe Agent binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help message"
