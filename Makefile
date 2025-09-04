.PHONY: build test test-verbose test-coverage clean help

# 默认目标
all: build

# 构建项目
build:
	go build -o gchangelog.exe main.go

# 运行测试
test:
	go test ./...

# 运行详细测试
test-verbose:
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	go tool cover -func=coverage/coverage.out

# 运行性能测试
test-bench:
	go test -bench=. ./...

# 运行所有测试（包括性能测试和覆盖率）
test-all: test test-bench test-coverage

# 清理构建文件
clean:
	rm -f gchangelog.exe
	rm -rf coverage/

# 安装依赖
deps:
	go mod download
	go mod tidy

# 显示帮助信息
help:
	@echo "可用的命令:"
	@echo "  build        - 构建项目"
	@echo "  test         - 运行测试"
	@echo "  test-verbose - 运行详细测试"
	@echo "  test-coverage- 运行测试并生成覆盖率报告"
	@echo "  test-bench   - 运行性能测试"
	@echo "  test-all     - 运行所有测试（包括性能测试和覆盖率）"
	@echo "  clean        - 清理构建文件"
	@echo "  deps         - 安装依赖"
	@echo "  help         - 显示此帮助信息"