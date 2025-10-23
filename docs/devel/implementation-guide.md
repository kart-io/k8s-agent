# Aetherius 项目结构改进实施指南

基于 OneX 最佳实践的详细实施步骤

## 文档版本

- **版本**: v1.0.0
- **创建日期**: 2025-10-23
- **预计实施时间**: 3-4 周

---

## 第一阶段: Buf 集成 (Week 1, Days 1-2)

### 目标

将项目从传统 protoc 工作流迁移到现代化的 Buf 工作流。

### 步骤 1.1: 备份现有配置

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent

# 创建备份分支
git checkout -b backup/before-buf-migration
git push origin backup/before-buf-migration

# 创建工作分支
git checkout -b feature/buf-integration
```

### 步骤 1.2: 安装 Buf

```bash
cd api/proto

# 安装 buf
make install-buf

# 验证安装
buf --version
```

### 步骤 1.3: 应用新配置文件

```bash
# 备份旧 Makefile
mv Makefile Makefile.old

# 应用新 Makefile
mv Makefile.new Makefile

# buf.gen.yaml 已经创建,无需额外操作

# 验证配置
buf config ls-lint-rules
buf config ls-breaking-rules
```

### 步骤 1.4: 更新依赖

```bash
# 更新 buf 依赖
buf dep update

# 检查生成的 buf.lock
cat buf.lock
```

### 步骤 1.5: 运行 Buf Lint

```bash
# 首次运行 lint,可能会发现一些问题
buf lint

# 如果有问题,逐个修复
# 常见问题:
# 1. 包名不符合规范
# 2. 字段编号不连续
# 3. 缺少注释
```

### 步骤 1.6: 生成代码

```bash
# 清理旧的生成代码
make clean

# 使用 buf 生成新代码
make buf-generate

# 验证生成的代码
ls -R gen/
```

### 步骤 1.7: 测试编译

```bash
# 进入生成的 Go 代码目录
cd gen/go

# 尝试编译
go mod init github.com/kart-io/k8s-agent/api/proto/gen/go || true
go mod tidy
go build ./...
```

### 步骤 1.8: 更新 go.mod

更新主项目的 go.mod 以使用新的生成路径:

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent

# 更新 replace 指令
# 编辑 go.mod,确保 proto 路径正确
```

### 步骤 1.9: 运行完整测试

```bash
# 运行所有测试,确保没有破坏现有功能
make test

# 如果有失败,修复导入路径
find . -name "*.go" -type f -exec grep -l "github.com/kart-io/k8s-agent/protos" {} \;
```

### 步骤 1.10: 提交 Buf 集成

```bash
git add api/proto/buf.gen.yaml
git add api/proto/Makefile
git add api/proto/buf.lock
git commit -m "feat(proto): migrate to Buf for proto management

- Add buf.gen.yaml for code generation
- Update Makefile with buf commands
- Lock dependencies in buf.lock
- Maintain backward compatibility with legacy commands

Breaking: None (maintains compatibility)"
```

---

## 第二阶段: 添加新的 Proto 定义 (Week 1, Days 3-5)

### 目标

为 Orchestrator 和 Reasoning 服务添加完整的 Proto API 定义。

### 步骤 2.1: 创建通用 Proto 定义

```bash
cd api/proto

# 创建通用消息定义
mkdir -p common/pagination/v1
mkdir -p common/errors/v1
mkdir -p common/types/v1
```

#### common/pagination/v1/pagination.proto

```protobuf
syntax = "proto3";

package common.pagination.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/common/pagination/v1;paginationv1";

// PageRequest 分页请求
message PageRequest {
  int32 page = 1;           // 页码,从 1 开始
  int32 page_size = 2;      // 每页大小,默认 20,最大 100
  string order_by = 3;      // 排序字段
  bool desc = 4;            // 是否降序,默认 false (升序)
}

// PageResponse 分页响应
message PageResponse {
  int32 page = 1;           // 当前页
  int32 page_size = 2;      // 每页大小
  int64 total = 3;          // 总记录数
  int32 total_pages = 4;    // 总页数
}
```

#### common/errors/v1/errors.proto

```protobuf
syntax = "proto3";

package common.errors.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/common/errors/v1;errorsv1";

// Error 错误详情
message Error {
  int32 code = 1;               // 错误码
  string message = 2;           // 错误消息
  repeated ErrorDetail details = 3;  // 错误详情列表
  string trace_id = 4;          // 追踪 ID
}

// ErrorDetail 错误详细信息
message ErrorDetail {
  string field = 1;             // 字段名
  string issue = 2;             // 问题描述
  string value = 3;             // 当前值
}

// ErrorCode 标准错误码
enum ErrorCode {
  ERROR_CODE_UNSPECIFIED = 0;
  ERROR_CODE_INVALID_ARGUMENT = 1;
  ERROR_CODE_NOT_FOUND = 2;
  ERROR_CODE_ALREADY_EXISTS = 3;
  ERROR_CODE_PERMISSION_DENIED = 4;
  ERROR_CODE_UNAUTHENTICATED = 5;
  ERROR_CODE_RESOURCE_EXHAUSTED = 6;
  ERROR_CODE_FAILED_PRECONDITION = 7;
  ERROR_CODE_ABORTED = 8;
  ERROR_CODE_OUT_OF_RANGE = 9;
  ERROR_CODE_UNIMPLEMENTED = 10;
  ERROR_CODE_INTERNAL = 11;
  ERROR_CODE_UNAVAILABLE = 12;
  ERROR_CODE_DATA_LOSS = 13;
}
```

### 步骤 2.2: 创建 Orchestrator Proto 定义

```bash
mkdir -p orchestrator/workflow/v1
mkdir -p orchestrator/strategy/v1
mkdir -p orchestrator/execution/v1
```

参考前面生成的 `workflow.proto` 示例创建完整的 Orchestrator API。

### 步骤 2.3: 创建 Reasoning Proto 定义

```bash
mkdir -p reasoning/analysis/v1
mkdir -p reasoning/recommendation/v1
mkdir -p reasoning/knowledge/v1
```

#### reasoning/analysis/v1/analysis.proto

```protobuf
syntax = "proto3";

package reasoning.analysis.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/reasoning/analysis/v1;analysisv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

// AnalysisService 根因分析服务
service AnalysisService {
  // AnalyzeIncident 分析故障
  rpc AnalyzeIncident(AnalyzeIncidentRequest) returns (AnalysisResult) {
    option (google.api.http) = {
      post: "/api/v1/analysis/incidents"
      body: "*"
    };
  }

  // GetAnalysisResult 获取分析结果
  rpc GetAnalysisResult(GetAnalysisResultRequest) returns (AnalysisResult) {
    option (google.api.http) = {
      get: "/api/v1/analysis/results/{result_id}"
    };
  }
}

// AnalysisStatus 分析状态
enum AnalysisStatus {
  ANALYSIS_STATUS_UNSPECIFIED = 0;
  ANALYSIS_STATUS_PENDING = 1;
  ANALYSIS_STATUS_ANALYZING = 2;
  ANALYSIS_STATUS_COMPLETED = 3;
  ANALYSIS_STATUS_FAILED = 4;
}

// AnalyzeIncidentRequest 分析故障请求
message AnalyzeIncidentRequest {
  string incident_id = 1;
  string event_type = 2;
  map<string, string> context = 3;
  repeated string log_snippets = 4;
  map<string, double> metrics = 5;
}

// AnalysisResult 分析结果
message AnalysisResult {
  string result_id = 1;
  string incident_id = 2;
  AnalysisStatus status = 3;
  RootCause root_cause = 4;
  repeated Evidence evidences = 5;
  double confidence = 6;  // 置信度 0-1
  google.protobuf.Timestamp analyzed_at = 7;
}

// RootCause 根因
message RootCause {
  string category = 1;  // 类别: resource, network, application, etc.
  string description = 2;
  string component = 3;  // 出问题的组件
  map<string, string> attributes = 4;
}

// Evidence 证据
message Evidence {
  string type = 1;  // log, metric, event
  string source = 2;
  string content = 3;
  double weight = 4;  // 权重
}

// GetAnalysisResultRequest 获取分析结果请求
message GetAnalysisResultRequest {
  string result_id = 1;
}
```

### 步骤 2.4: 生成新的 Proto 代码

```bash
cd api/proto

# Lint 新的 proto 文件
buf lint

# 生成代码
buf generate

# 检查生成的代码
tree gen/go/
```

### 步骤 2.5: 提交新 Proto 定义

```bash
git add api/proto/common/
git add api/proto/orchestrator/
git add api/proto/reasoning/
git commit -m "feat(proto): add Orchestrator and Reasoning API definitions

- Add common proto definitions (pagination, errors)
- Add Orchestrator API (workflow, strategy, execution)
- Add Reasoning API (analysis, recommendation, knowledge)
- Generate Go code and OpenAPI specs"
```

---

## 第三阶段: 项目结构重组 (Week 2-3)

### 目标

将项目重组为标准 Go 项目布局。

### 步骤 3.1: 创建新目录结构

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent

# 创建 pkg/ 目录
mkdir -p pkg/client/agentmanager
mkdir -p pkg/client/orchestrator
mkdir -p pkg/client/reasoning
mkdir -p pkg/types
mkdir -p pkg/utils
mkdir -p pkg/errors

# 创建 internal/pkg/ 目录
mkdir -p internal/pkg/config
mkdir -p internal/pkg/db
mkdir -p internal/pkg/logger
mkdir -p internal/pkg/middleware
mkdir -p internal/pkg/response
mkdir -p internal/pkg/telemetry
mkdir -p internal/pkg/validator
mkdir -p internal/pkg/version

# 创建增强的文档结构
mkdir -p docs/devel/conventions
mkdir -p docs/devel/guide
mkdir -p docs/user/guide
mkdir -p docs/user/tutorials
mkdir -p docs/architecture/ADR

# 创建 scripts/make-rules/
mkdir -p scripts/make-rules
mkdir -p scripts/install
mkdir -p scripts/lib

# 创建测试目录
mkdir -p test/integration
mkdir -p test/e2e
mkdir -p test/testdata
mkdir -p test/fixtures
```

### 步骤 3.2: 迁移 common/ 到 internal/pkg/

```bash
# 创建迁移脚本
cat > tools/migration/migrate-common-to-internal-pkg.sh <<'EOF'
#!/bin/bash
# 迁移 common/ 到 internal/pkg/

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

echo "Migrating common/ to internal/pkg/..."

# 需要迁移的目录列表
DIRS=(
  "config"
  "db"
  "logger"
  "middleware"
  "pagination"
  "response"
  "telemetry"
  "utils"
  "validator"
)

for dir in "${DIRS[@]}"; do
  if [ -d "common/$dir" ]; then
    echo "Migrating common/$dir to internal/pkg/$dir..."
    git mv "common/$dir" "internal/pkg/$dir"
  fi
done

echo "Migration phase 1 complete!"
echo "Next: Update import paths in Go files"
EOF

chmod +x tools/migration/migrate-common-to-internal-pkg.sh
```

### 步骤 3.3: 更新导入路径

```bash
# 创建导入路径更新脚本
cat > tools/migration/update-import-paths.sh <<'EOF'
#!/bin/bash
# 更新导入路径

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

echo "Updating import paths..."

# 更新所有 Go 文件中的导入路径
find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./api/proto/gen/*" | while read -r file; do
  # 备份文件
  cp "$file" "$file.bak"

  # 替换导入路径
  sed -i 's|github.com/kart-io/k8s-agent/common/|github.com/kart-io/k8s-agent/internal/pkg/|g' "$file"

  # 检查是否有变更
  if ! diff "$file" "$file.bak" > /dev/null 2>&1; then
    echo "Updated: $file"
  fi

  # 删除备份
  rm "$file.bak"
done

echo "Import paths updated!"
echo "Please run 'go mod tidy' to update dependencies"
EOF

chmod +x tools/migration/update-import-paths.sh
```

### 步骤 3.4: 创建公共客户端库

在 `pkg/client/` 下创建各服务的客户端:

```bash
# pkg/client/agentmanager/client.go
cat > pkg/client/agentmanager/client.go <<'EOF'
package agentmanager

import (
	"context"

	agentv1 "github.com/kart-io/k8s-agent/api/proto/gen/go/agentmanager/agent/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client Agent Manager 客户端
type Client struct {
	conn         *grpc.ClientConn
	agentClient  agentv1.AgentServiceClient
}

// New 创建新的 Agent Manager 客户端
func New(addr string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:        conn,
		agentClient: agentv1.NewAgentServiceClient(conn),
	}, nil
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.conn.Close()
}

// Agent 返回 Agent 服务客户端
func (c *Client) Agent() agentv1.AgentServiceClient {
	return c.agentClient
}

// ListAgents 列出所有 Agents (便利方法)
func (c *Client) ListAgents(ctx context.Context, req *agentv1.ListAgentsRequest) (*agentv1.ListAgentsResponse, error) {
	return c.agentClient.ListAgents(ctx, req)
}

// GetAgent 获取 Agent 详情 (便利方法)
func (c *Client) GetAgent(ctx context.Context, agentID string) (*agentv1.Agent, error) {
	return c.agentClient.GetAgent(ctx, &agentv1.GetAgentRequest{
		AgentId: agentID,
	})
}
EOF
```

### 步骤 3.5: 执行迁移

```bash
# 1. 执行目录迁移
./tools/migration/migrate-common-to-internal-pkg.sh

# 2. 更新导入路径
./tools/migration/update-import-paths.sh

# 3. 更新 go.mod
go mod tidy

# 4. 验证编译
make build

# 5. 运行测试
make test
```

### 步骤 3.6: 提交结构重组

```bash
git add .
git commit -m "refactor: restructure project following Go standard layout

- Move common/ to internal/pkg/
- Create pkg/ for public libraries
- Add client libraries in pkg/client/
- Update all import paths
- Maintain backward compatibility

This follows golang-standards/project-layout and makes
the codebase more maintainable and idiomatic."
```

---

## 第四阶段: Makefile 模块化 (Week 3)

### 目标

将根 Makefile 拆分为模块化的子规则文件。

### 步骤 4.1: 创建 scripts/make-rules/

```bash
mkdir -p scripts/make-rules
```

### 步骤 4.2: 创建 common.mk

```makefile
# scripts/make-rules/common.mk
# 通用变量和函数

# 项目信息
PROJECT_NAME := k8s-agent
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# 目录
ROOT_DIR := $(shell pwd)
BIN_DIR := $(ROOT_DIR)/bin
BUILD_DIR := $(ROOT_DIR)/build
CMD_DIR := $(ROOT_DIR)/cmd
API_DIR := $(ROOT_DIR)/api
TOOLS_DIR := $(ROOT_DIR)/tools
SCRIPTS_DIR := $(ROOT_DIR)/scripts

# Go 配置
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.Version=$(VERSION)' \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/kart-io/k8s-agent/internal/pkg/version.BuildTime=$(BUILD_TIME)'

# 服务列表
ALL_SERVICES := agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent
BINS ?= $(ALL_SERVICES)

# 平台
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Docker
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= aetherius
IMAGE_TAG ?= $(VERSION)

# 颜色输出
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[36m
COLOR_RED := \033[31m

# 辅助函数
define log_info
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)$(1)$(COLOR_RESET)"
endef

define log_success
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)$(1)$(COLOR_RESET)"
endef

define log_warning
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)$(1)$(COLOR_RESET)"
endef

define log_error
	@echo "$(COLOR_BOLD)$(COLOR_RED)$(1)$(COLOR_RESET)"
endef
```

### 步骤 4.3: 创建 golang.mk

```makefile
# scripts/make-rules/golang.mk
# Go 相关规则

.PHONY: build build-all build-multiarch test test-coverage lint fmt vet

# 构建单个服务
build:
	$(call log_info,"Building services: $(BINS)")
	@mkdir -p $(BIN_DIR)
	@for service in $(BINS); do \
		echo "Building $$service..."; \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
			-o $(BIN_DIR)/$$service $(CMD_DIR)/$$service || exit 1; \
		$(call log_success,"✓ Built $$service"); \
	done

# 构建所有服务
build-all:
	@$(MAKE) build BINS="$(ALL_SERVICES)"

# 多平台构建
build-multiarch:
	$(call log_info,"Building for multiple platforms")
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		for service in $(BINS); do \
			echo "Building $$service for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) \
				-ldflags "$(LDFLAGS)" \
				-o $(BIN_DIR)/$$service-$$os-$$arch \
				$(CMD_DIR)/$$service || exit 1; \
		done; \
	done
	$(call log_success,"✓ Multi-platform build complete")

# 测试
test:
	$(call log_info,"Running tests")
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	$(call log_success,"✓ Tests passed")

# 测试覆盖率
test-coverage: test
	$(call log_info,"Generating coverage report")
	@$(GO) tool cover -html=coverage.out -o coverage.html
	$(call log_success,"✓ Coverage report: coverage.html")

# Lint
lint:
	$(call log_info,"Running linters")
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		$(call log_success,"✓ Lint passed"); \
	else \
		$(call log_warning,"golangci-lint not installed"); \
	fi

# 格式化
fmt:
	$(call log_info,"Formatting code")
	@$(GO) fmt ./...
	$(call log_success,"✓ Code formatted")

# Vet
vet:
	$(call log_info,"Running go vet")
	@$(GO) vet ./...
	$(call log_success,"✓ Vet passed")
```

### 步骤 4.4: 更新根 Makefile

```makefile
# Makefile - 根 Makefile

# 引入子规则
include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/docker.mk
include scripts/make-rules/gen.mk
include scripts/make-rules/tools.mk

.DEFAULT_GOAL := help

.PHONY: all help

all: fmt vet lint test build

help:
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╔══════════════════════════════════════════════════════════════╗$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)║     K8s-Agent (Aetherius) Build System                       ║$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)╚══════════════════════════════════════════════════════════════╝$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Project: $(PROJECT_NAME)$(COLOR_RESET)"
	@echo "$(COLOR_BOLD)Version: $(VERSION)$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_YELLOW)%-30s$(COLOR_RESET) %s\n", $$1, $$2}'
```

---

## 第五阶段: CI/CD 增强 (Week 4)

### 目标

添加完整的 GitHub Actions 工作流和质量检查。

### 步骤 5.1: 创建 Proto CI 工作流

参考前面创建的 `.github/workflows/proto.yml`。

### 步骤 5.2: 创建主 CI 工作流

创建 `.github/workflows/ci.yml` (已在改进方案文档中提供)。

### 步骤 5.3: 创建 Pre-commit Hook

```bash
# 创建 pre-commit 钩子
cat > .git/hooks/pre-commit <<'EOF'
#!/bin/bash
# Pre-commit hook

set -e

echo "Running pre-commit checks..."

# 1. Format check
echo "Checking code format..."
make fmt
if ! git diff --exit-code > /dev/null; then
  echo "❌ Code is not formatted. Changes have been made."
  echo "Please review and commit the formatted code."
  exit 1
fi

# 2. Proto checks (if proto files changed)
if git diff --cached --name-only | grep -q "^api/proto/.*\.proto$"; then
  echo "Proto files changed, running proto checks..."
  cd api/proto
  make buf-lint
  make buf-format
  cd ../..
fi

# 3. Lint (快速模式)
echo "Running quick lint..."
make vet

echo "✅ Pre-commit checks passed!"
exit 0
EOF

chmod +x .git/hooks/pre-commit
```

---

## 第六阶段: 文档完善 (Week 4)

### 目标

建立完整的文档体系。

### 步骤 6.1: 创建开发规范文档

已创建:
- `docs/devel/proto-buf-guide.md`
- `docs/architecture/IMPROVEMENT_PLAN.md`

待创建:
- `docs/devel/conventions/coding-style.md` (参考改进方案中的示例)
- `docs/devel/conventions/commit-message.md`
- `docs/devel/guide/getting-started.md`

### 步骤 6.2: 创建 ADR 模板

```bash
mkdir -p docs/architecture/ADR

cat > docs/architecture/ADR/template.md <<'EOF'
# ADR-XXXX: [Title]

## Status

[Proposed | Accepted | Deprecated | Superseded]

## Context

[Describe the issue that we're seeing that is motivating this decision or change]

## Decision

[Describe our response to these forces. It is stated in full sentences, with active voice]

## Consequences

### Positive

- [e.g., improvement of quality attribute]

### Negative

- [e.g., compromising quality attribute]

### Risks

- [e.g., potential risks and their mitigation]

## Alternatives Considered

### Alternative 1

[Description]

### Alternative 2

[Description]

## Related Decisions

- ADR-YYYY: [Related decision]

## References

- [External references]
EOF
```

---

## 验证清单

在完成所有阶段后,使用以下清单验证:

### Buf 集成

- [ ] `buf --version` 正常工作
- [ ] `make buf-lint` 通过
- [ ] `make buf-generate` 生成代码成功
- [ ] `make buf-breaking` 可以检测破坏性变更

### Proto 定义

- [ ] 所有服务都有完整的 Proto 定义
- [ ] 通用消息(pagination, errors)已创建
- [ ] 生成的 Go 代码可以编译

### 项目结构

- [ ] `internal/pkg/` 包含内部共享代码
- [ ] `pkg/` 包含公共客户端库
- [ ] 导入路径更新完成
- [ ] 所有测试通过

### 构建系统

- [ ] 模块化 Makefile 工作正常
- [ ] `make all` 执行完整 CI pipeline
- [ ] `make build` 可以构建所有服务
- [ ] `make test` 运行所有测试

### CI/CD

- [ ] GitHub Actions 工作流配置正确
- [ ] Pre-commit hook 工作正常
- [ ] CI 通过所有检查

### 文档

- [ ] Proto 指南完整
- [ ] 改进方案文档完整
- [ ] ADR 模板创建
- [ ] README 更新

---

## 回滚计划

如果遇到严重问题需要回滚:

```bash
# 1. 切换回备份分支
git checkout backup/before-buf-migration

# 2. 创建新的工作分支
git checkout -b fix/rollback-and-retry

# 3. 逐步重新应用改进
# ... 找出问题并修复
```

---

## 下一步

完成本实施指南后:

1. 开始实施改进方案中的 P1 优先级项目
2. 逐步提升测试覆盖率
3. 完善监控和可观测性
4. 进行性能优化

---

**文档版本**: v1.0.0
**最后更新**: 2025-10-23
**负责人**: Aetherius Team
