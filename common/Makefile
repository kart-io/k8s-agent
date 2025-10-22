# Common Library Makefile
.PHONY: help proto-gen proto-lint proto-format proto-breaking proto-clean test build clean

# 默认目标
.DEFAULT_GOAL := help

# 颜色定义
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

## help: 显示帮助信息
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${GREEN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Protobuf 相关命令:

## proto-gen: 生成 protobuf 代码
proto-gen:
	@echo "Generating protobuf code..."
	cd proto && buf generate
	@echo "Protobuf code generated successfully"

## proto-lint: 检查 protobuf 定义
proto-lint:
	@echo "Linting protobuf files..."
	cd proto && buf lint
	@echo "Protobuf lint completed"

## proto-format: 格式化 protobuf 文件
proto-format:
	@echo "Formatting protobuf files..."
	cd proto && buf format -w
	@echo "Protobuf files formatted"

## proto-breaking: 检查破坏性变更
proto-breaking:
	@echo "Checking for breaking changes..."
	cd proto && buf breaking --against '.git#branch=master'
	@echo "Breaking change check completed"

## proto-clean: 清理生成的代码
proto-clean:
	@echo "Cleaning generated protobuf code..."
	rm -rf proto/gen
	@echo "Generated code cleaned"

## proto-update: 更新 buf 依赖
proto-update:
	@echo "Updating buf dependencies..."
	cd proto && buf mod update
	@echo "Buf dependencies updated"

## 开发相关命令:

## test: 运行测试
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests completed"

## test-coverage: 运行测试并显示覆盖率
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## fmt: 格式化代码
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Code formatted"

## lint: 运行 golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "Lint completed"

## vet: 运行 go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet completed"

## tidy: 整理 go.mod
tidy:
	@echo "Tidying go.mod..."
	go mod tidy
	@echo "go.mod tidied"

## deps: 下载依赖
deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

## 清理相关命令:

## clean: 清理构建产物
clean: proto-clean
	@echo "Cleaning build artifacts..."
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "Build artifacts cleaned"

## 工具安装:

## install-buf: 安装 buf
install-buf:
	@echo "Installing buf..."
	go install github.com/bufbuild/buf/cmd/buf@latest
	@echo "buf installed successfully"

## install-tools: 安装所有开发工具
install-tools: install-buf
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
	@echo "All tools installed"

## CI/CD 相关命令:

## ci: 运行 CI 流程
ci: proto-lint proto-gen fmt vet lint test
	@echo "CI pipeline completed successfully"

## 快速开始:

## init: 初始化项目（安装工具 + 生成代码 + 下载依赖）
init: install-tools proto-gen deps
	@echo "Project initialized successfully"
