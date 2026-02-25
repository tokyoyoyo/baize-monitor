# Baize Monitor Makefile

# 默认目标
.DEFAULT_GOAL := help

# 测试相关目标
.PHONY: test test-sequential test-parallel test-cover clean-test

# 清除测试缓存
clean-test:
	@echo "Cleaning test cache..."
	go clean -testcache

# 顺序执行所有测试（解决数据库竞争问题）
test-sequential: clean-test
	@echo "Running all tests sequentially..."
	go test ./... -p 1 -v

# 生成测试覆盖率报告
test-cover: clean-test
	@echo "Generating test coverage report..."
	go test ./... -p 1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"


# 显示帮助信息
help:
	@echo "Baize Monitor Test Commands:"
	@echo "  make test-sequential   - Run all tests sequentially (recommended)"
	@echo "  make test-cover        - Generate coverage report"
	@echo "  make clean-test        - Clean test cache"