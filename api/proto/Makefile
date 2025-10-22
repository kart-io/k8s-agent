.PHONY: all clean gen-go gen-agentmanager gen-common help install-tools setup

# Proto 文件目录
PROTO_DIR := .
# 生成代码输出目录
GEN_DIR := gen

# protoc 编译器配置
PROTOC := protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY := $(shell go env GOPATH)/bin/protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 := $(shell go env GOPATH)/bin/protoc-gen-openapiv2

# Go module path
GO_MODULE := github.com/kart-io/k8s-agent

# Proto include paths
PROTO_INCLUDES := \
	-I$(PROTO_DIR) \
	-I$(shell go env GOPATH)/src \
	-I$(shell go env GOPATH)/src/github.com/googleapis/googleapis \
	-I$(shell go env GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway

# 默认目标
all: gen-go

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all                - Generate all proto files (default)"
	@echo "  setup              - Setup googleapis dependencies"
	@echo "  gen-go             - Generate Go code for all proto files"
	@echo "  gen-agentmanager   - Generate Go code for agent-manager protos"
	@echo "  gen-common         - Generate Go code for common protos"
	@echo "  install-tools      - Install required protoc plugins"
	@echo "  clean              - Remove generated files"
	@echo "  help               - Show this help message"

# 安装所需工具
install-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Done! Make sure $(shell go env GOPATH)/bin is in your PATH"

# 设置 googleapis 依赖
setup:
	@echo "Setting up googleapis dependencies..."
	@chmod +x setup-googleapis.sh
	@./setup-googleapis.sh

# 生成所有 Go 代码
gen-go: gen-agentmanager gen-common
	@echo "All proto files generated successfully!"

# 生成 agent-manager 的 proto 代码
gen-agentmanager:
	@echo "Generating agent-manager proto files..."
	@mkdir -p $(GEN_DIR)/agentmanager/agent/v1
	@mkdir -p $(GEN_DIR)/agentmanager/command/v1
	@mkdir -p $(GEN_DIR)/agentmanager/event/v1

	# Generate agent.proto
	$(PROTOC) $(PROTO_INCLUDES) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=logtostderr=true \
		agentmanager/agent/v1/agent.proto

	# Generate command.proto
	$(PROTOC) $(PROTO_INCLUDES) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=logtostderr=true \
		agentmanager/command/v1/command.proto

	# Generate event.proto
	$(PROTOC) $(PROTO_INCLUDES) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) \
		--grpc-gateway_opt=paths=source_relative \
		--grpc-gateway_opt=logtostderr=true \
		agentmanager/event/v1/event.proto

	@echo "agent-manager proto files generated!"

# 生成 common 的 proto 代码
gen-common:
	@echo "Generating common proto files..."
	@mkdir -p $(GEN_DIR)/common/example/v1
	@mkdir -p $(GEN_DIR)/common/health/v1

	# Generate example.proto
	$(PROTOC) $(PROTO_INCLUDES) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		common/example/v1/example.proto

	# Generate health.proto
	$(PROTOC) $(PROTO_INCLUDES) \
		--go_out=$(GEN_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) \
		--go-grpc_opt=paths=source_relative \
		common/health/v1/health.proto

	@echo "common proto files generated!"

# 清理生成的文件
clean:
	@echo "Cleaning generated files..."
	rm -rf $(GEN_DIR)
	@echo "Done!"

# 验证 proto 文件语法
validate:
	@echo "Validating proto files..."
	@find $(PROTO_DIR) -name "*.proto" -type f | while read proto; do \
		echo "Validating $$proto..."; \
		$(PROTOC) $(PROTO_INCLUDES) --descriptor_set_out=/dev/null $$proto || exit 1; \
	done
	@echo "All proto files are valid!"

# 列出所有 proto 文件
list:
	@echo "Proto files in project:"
	@find $(PROTO_DIR) -name "*.proto" -type f | sort
