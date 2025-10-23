# Protocol Buffers 和 Buf 管理指南

本文档详细说明如何使用 Buf 工具管理 Aetherius 项目的 Protocol Buffers 定义。

## 文档版本

- **版本**: v1.0.0
- **创建日期**: 2025-10-23
- **参考**: [Buf 官方文档](https://buf.build/docs/)

---

## 1. Buf 简介

### 1.1 为什么使用 Buf

Buf 是现代化的 Protocol Buffers 工具链,相比传统 protoc 有以下优势:

#### 优势对比

| 特性 | protoc (传统) | Buf (现代) |
|------|--------------|-----------|
| **配置复杂度** | 高 (需要手动管理 include 路径) | 低 (buf.yaml 统一配置) |
| **依赖管理** | 手动克隆或下载 | 自动下载并缓存 |
| **代码检查** | 无内置 linter | 内置强大的 linter |
| **破坏性检测** | 需要第三方工具 | 内置 breaking change 检测 |
| **代码格式化** | 无 | 内置 formatter |
| **性能** | 一般 | 更快 (并行编译) |
| **BSR 集成** | 无 | 可发布到 Buf Schema Registry |

### 1.2 核心功能

1. **Linting**: 强制执行 API 设计最佳实践
2. **Breaking Change Detection**: 确保向后兼容
3. **Code Generation**: 简化的代码生成流程
4. **Dependency Management**: 自动管理 protobuf 依赖
5. **Format**: 统一的代码格式化

---

## 2. 项目 Proto 目录结构

### 2.1 当前结构

```
api/proto/
├── buf.yaml                    # Buf 配置文件
├── buf.gen.yaml               # 代码生成配置 (待创建)
├── buf.lock                   # 依赖锁定文件 (自动生成)
├── Makefile                   # 构建脚本 (兼容 Buf)
├── README.md                  # Proto 使用文档
├── agentmanager/              # Agent Manager 服务 API
│   ├── agent/v1/
│   │   └── agent.proto        # Agent 管理 API
│   ├── command/v1/
│   │   └── command.proto      # 命令执行 API
│   └── event/v1/
│       └── event.proto        # 事件处理 API
├── orchestrator/              # Orchestrator 服务 API (待添加)
│   ├── workflow/v1/
│   ├── strategy/v1/
│   └── execution/v1/
├── reasoning/                 # Reasoning 服务 API (待添加)
│   ├── analysis/v1/
│   ├── recommendation/v1/
│   └── knowledge/v1/
├── common/                    # 通用定义
│   ├── health/v1/
│   │   └── health.proto       # 健康检查
│   ├── pagination/v1/         # 分页 (待添加)
│   │   └── pagination.proto
│   └── errors/v1/             # 错误定义 (待添加)
│       └── errors.proto
└── gen/                       # 生成的代码 (gitignored)
    ├── go/                    # Go 代码
    └── openapiv2/             # OpenAPI 文档
```

### 2.2 建议的改进结构

```
api/proto/
├── buf.yaml                   # Buf workspace 配置
├── buf.gen.yaml              # 代码生成配置
├── buf.lock                  # 依赖锁定
├── .gitignore                # 忽略生成的代码
├── README.md
├── Makefile                  # Buf 集成的 Makefile
├── v1/                       # API v1 版本 (稳定)
│   ├── agentmanager/
│   │   ├── buf.yaml          # 模块特定配置 (可选)
│   │   ├── agent/
│   │   │   └── agent.proto
│   │   ├── command/
│   │   │   └── command.proto
│   │   └── event/
│   │       └── event.proto
│   ├── orchestrator/
│   │   ├── workflow/
│   │   ├── strategy/
│   │   └── execution/
│   ├── reasoning/
│   │   ├── analysis/
│   │   ├── recommendation/
│   │   └── knowledge/
│   └── common/
│       ├── health/
│       ├── pagination/
│       └── errors/
├── v2alpha1/                 # API v2alpha1 (实验性,可选)
│   └── ...
└── gen/                      # 生成的代码
    ├── go/
    ├── openapiv2/
    └── doc/
```

---

## 3. Buf 配置详解

### 3.1 buf.yaml (主配置)

当前配置:

```yaml
# api/proto/buf.yaml
version: v2
modules:
  - path: .
    name: buf.build/kart-io/k8s-agent
deps:
  - buf.build/googleapis/googleapis
lint:
  use:
    - STANDARD
  except:
    - PACKAGE_VERSION_SUFFIX
breaking:
  use:
    - FILE
```

#### 推荐的增强配置

```yaml
# api/proto/buf.yaml
version: v2

# 模块定义
modules:
  - path: .
    name: buf.build/kart-io/k8s-agent
    # 排除生成的代码和第三方代码
    excludes:
      - gen
      - vendor
      - third_party

# 依赖管理
deps:
  - buf.build/googleapis/googleapis           # Google APIs
  - buf.build/grpc-ecosystem/grpc-gateway     # gRPC-Gateway

# Linting 配置
lint:
  use:
    - STANDARD                    # 标准规则集
  except:
    - PACKAGE_VERSION_SUFFIX     # 允许不带版本后缀的包名
    - IMPORT_NO_PUBLIC           # 允许不使用 public import (根据需求)
  # 自定义规则
  enum_zero_value_suffix: _UNSPECIFIED
  rpc_allow_same_request_response: false
  rpc_allow_google_protobuf_empty_requests: true
  rpc_allow_google_protobuf_empty_responses: true
  service_suffix: Service

# 破坏性变更检测
breaking:
  use:
    - FILE                        # 基于文件的检测
  except:
    - EXTENSION_MESSAGE_NO_DELETE # 根据项目需求调整
  # 检测范围配置
  ignore:
    - v2alpha1                    # 忽略 alpha 版本的破坏性检测
  # 对比目标 (通常是主分支或上一个版本)
  # against: '.git#branch=main,subdir=api/proto'
```

### 3.2 buf.gen.yaml (代码生成配置)

创建 `api/proto/buf.gen.yaml`:

```yaml
# api/proto/buf.gen.yaml
version: v2

# 管理的插件
managed:
  enabled: true
  override:
    # 覆盖所有文件的 Go import 路径
    - file_option: go_package_prefix
      value: github.com/kart-io/k8s-agent/api/proto/gen/go

# 插件配置
plugins:
  # Go protobuf 插件
  - remote: buf.build/protocolbuffers/go
    out: gen/go
    opt:
      - paths=source_relative

  # Go gRPC 插件
  - remote: buf.build/grpc/go
    out: gen/go
    opt:
      - paths=source_relative

  # gRPC-Gateway 插件
  - remote: buf.build/grpc-ecosystem/gateway
    out: gen/go
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

  # OpenAPI v2 (Swagger) 插件
  - remote: buf.build/grpc-ecosystem/openapiv2
    out: gen/openapiv2
    opt:
      - allow_merge=true
      - merge_file_name=api

  # 文档生成插件 (可选)
  - remote: buf.build/bufbuild/doc
    out: gen/doc
    opt:
      - template=html

# 输入配置
inputs:
  - directory: .
```

#### 本地插件配置 (如果不使用 BSR)

```yaml
# api/proto/buf.gen.yaml (本地版本)
version: v2

managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/kart-io/k8s-agent/api/proto/gen/go

plugins:
  # 使用本地安装的插件
  - local: protoc-gen-go
    out: gen/go
    opt:
      - paths=source_relative

  - local: protoc-gen-go-grpc
    out: gen/go
    opt:
      - paths=source_relative

  - local: protoc-gen-grpc-gateway
    out: gen/go
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

  - local: protoc-gen-openapiv2
    out: gen/openapiv2
    opt:
      - allow_merge=true
      - merge_file_name=api
```

### 3.3 .gitignore

```gitignore
# api/proto/.gitignore

# 生成的代码
gen/

# Buf 缓存
.buf/

# 依赖锁定文件 (如果使用 BSR)
# buf.lock

# IDE
.idea/
.vscode/
*.swp
*.swo
```

---

## 4. Buf 工作流

### 4.1 安装 Buf

```bash
# macOS
brew install bufbuild/buf/buf

# Linux (使用 binary)
BUF_VERSION=1.28.1
curl -sSL \
  "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf

# 验证安装
buf --version
```

### 4.2 初始化项目

```bash
cd api/proto

# 初始化 buf.yaml (如果还没有)
buf config init

# 更新依赖
buf dep update

# 这会生成 buf.lock 文件
```

### 4.3 日常开发工作流

#### Lint Proto 文件

```bash
# 检查所有 proto 文件
buf lint

# 只检查特定目录
buf lint --path agentmanager/
```

#### 破坏性变更检测

```bash
# 对比当前代码与 main 分支
buf breaking --against '.git#branch=main,subdir=api/proto'

# 对比当前代码与特定 tag
buf breaking --against '.git#tag=v1.0.0,subdir=api/proto'

# 对比当前代码与本地目录
buf breaking --against ../proto-v1/
```

#### 格式化 Proto 文件

```bash
# 格式化所有文件
buf format -w

# 检查格式 (不修改)
buf format --diff
```

#### 生成代码

```bash
# 使用 buf.gen.yaml 生成代码
buf generate

# 生成特定路径
buf generate --path agentmanager/

# 指定配置文件
buf generate --template buf.gen.yaml
```

### 4.4 集成到 Makefile

更新 `api/proto/Makefile`:

```makefile
.PHONY: all buf-lint buf-breaking buf-format buf-generate clean help

# 默认目标
all: buf-lint buf-generate

help:
	@echo "Available targets:"
	@echo "  all              - Lint and generate all proto files"
	@echo "  buf-lint         - Run buf lint"
	@echo "  buf-breaking     - Check for breaking changes"
	@echo "  buf-format       - Format proto files"
	@echo "  buf-generate     - Generate code from proto files"
	@echo "  buf-dep-update   - Update dependencies"
	@echo "  clean            - Remove generated files"
	@echo "  install-buf      - Install buf tool"

# Buf lint
buf-lint:
	@echo "Running buf lint..."
	buf lint

# Buf breaking change detection
buf-breaking:
	@echo "Checking for breaking changes..."
	buf breaking --against '.git#branch=main,subdir=api/proto'

# Buf format
buf-format:
	@echo "Formatting proto files..."
	buf format -w

# Buf generate
buf-generate:
	@echo "Generating code..."
	buf generate

# Update buf dependencies
buf-dep-update:
	@echo "Updating buf dependencies..."
	buf dep update

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf gen/

# Install buf
install-buf:
	@echo "Installing buf..."
	@if command -v brew >/dev/null 2>&1; then \
		brew install bufbuild/buf/buf; \
	else \
		curl -sSL \
		  "https://github.com/bufbuild/buf/releases/latest/download/buf-$$(uname -s)-$$(uname -m)" \
		  -o /tmp/buf && \
		chmod +x /tmp/buf && \
		sudo mv /tmp/buf /usr/local/bin/buf; \
	fi
	@buf --version

# 向后兼容: 保留 gen-go 目标
gen-go: buf-generate

# CI 目标
ci: buf-lint buf-breaking buf-generate
	@echo "CI checks passed!"
```

---

## 5. Proto 文件编写规范

### 5.1 文件组织

#### 每个文件一个服务或消息组

```protobuf
// ✅ Good: agent.proto - 只包含 Agent 相关定义
syntax = "proto3";

package agentmanager.agent.v1;

service AgentService {
  rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
  rpc GetAgent(GetAgentRequest) returns (Agent);
}

message Agent { ... }
message ListAgentsRequest { ... }
message ListAgentsResponse { ... }
```

```protobuf
// ❌ Bad: services.proto - 混合多个不相关的服务
service AgentService { ... }
service CommandService { ... }
service WorkflowService { ... }
```

### 5.2 命名规范

#### 包命名

```protobuf
// ✅ Good: 使用组织.服务.API.版本
package agentmanager.agent.v1;

// ❌ Bad
package agent;  // 太简短
package com.kart.io.k8s.agent.v1;  // 太长,Java 风格
```

#### 消息命名

```protobuf
// ✅ Good: 每个 RPC 都有独特的请求/响应消息
service AgentService {
  rpc CreateAgent(CreateAgentRequest) returns (Agent);
  rpc UpdateAgent(UpdateAgentRequest) returns (Agent);
  rpc DeleteAgent(DeleteAgentRequest) returns (DeleteAgentResponse);
}

message CreateAgentRequest {
  string cluster_id = 1;
  string name = 2;
}

// ❌ Bad: 重用消息或使用通用名称
service AgentService {
  rpc CreateAgent(AgentRequest) returns (Agent);  // 不够具体
  rpc UpdateAgent(AgentRequest) returns (Agent);  // 重用消息
}
```

#### 字段命名

```protobuf
// ✅ Good: 使用 snake_case
message Agent {
  string agent_id = 1;
  string cluster_id = 2;
  int64 created_at = 3;
  AgentStatus status = 4;
}

// ❌ Bad
message Agent {
  string AgentID = 1;       // PascalCase
  string clusterId = 2;     // camelCase
  int64 createdAt = 3;      // camelCase
}
```

### 5.3 字段编号管理

#### 保留已删除字段的编号

```protobuf
message Agent {
  reserved 2, 15, 9 to 11;    // 保留已删除的字段编号
  reserved "foo", "bar";       // 保留已删除的字段名

  string agent_id = 1;
  // string old_field = 2;     // 已删除,编号已保留
  string cluster_id = 3;
}
```

#### 字段编号分配策略

```protobuf
message Agent {
  // 1-15: 最常用的字段 (单字节编码)
  string agent_id = 1;
  string cluster_id = 2;
  AgentStatus status = 3;

  // 16-2047: 常用字段 (双字节编码)
  string name = 16;
  map<string, string> labels = 17;

  // 2048+: 不常用或扩展字段
  string internal_notes = 2048;
}
```

### 5.4 枚举定义

```protobuf
// ✅ Good: 第一个值是 UNSPECIFIED,值为 0
enum AgentStatus {
  AGENT_STATUS_UNSPECIFIED = 0;  // 必须有,且为 0
  AGENT_STATUS_ONLINE = 1;
  AGENT_STATUS_OFFLINE = 2;
  AGENT_STATUS_ERROR = 3;
}

// ❌ Bad
enum AgentStatus {
  ONLINE = 0;   // 没有 UNSPECIFIED
  OFFLINE = 1;
  ERROR = 2;
}
```

### 5.5 默认值和可选字段

```protobuf
// Proto3 中使用 optional 关键字
message UpdateAgentRequest {
  string agent_id = 1;                    // 必填
  optional string name = 2;               // 可选,可以区分未设置和空字符串
  optional AgentStatus status = 3;        // 可选
  map<string, string> labels = 4;         // map 本身可选,可以为空
}
```

### 5.6 时间和日期

```protobuf
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";

message Agent {
  string agent_id = 1;

  // ✅ Good: 使用 google.protobuf.Timestamp
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;

  // ✅ Good: 使用 google.protobuf.Duration
  google.protobuf.Duration heartbeat_interval = 4;

  // ❌ Bad: 使用 int64 或 string
  // int64 created_timestamp = 2;
  // string created_date = 3;
}
```

### 5.7 分页

```protobuf
// common/pagination/v1/pagination.proto
syntax = "proto3";

package common.pagination.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/common/pagination/v1;paginationv1";

// PageRequest 分页请求
message PageRequest {
  int32 page = 1;           // 页码,从 1 开始
  int32 page_size = 2;      // 每页大小,默认 20
  string order_by = 3;      // 排序字段
  bool desc = 4;            // 是否降序
}

// PageResponse 分页响应
message PageResponse {
  int32 page = 1;           // 当前页
  int32 page_size = 2;      // 每页大小
  int64 total = 3;          // 总记录数
  int32 total_pages = 4;    // 总页数
}
```

使用分页:

```protobuf
import "common/pagination/v1/pagination.proto";

message ListAgentsRequest {
  string cluster_id = 1;
  common.pagination.v1.PageRequest page = 2;
}

message ListAgentsResponse {
  repeated Agent agents = 1;
  common.pagination.v1.PageResponse page = 2;
}
```

### 5.8 错误处理

```protobuf
// common/errors/v1/errors.proto
syntax = "proto3";

package common.errors.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/common/errors/v1;errorsv1";

// Error 错误详情
message Error {
  int32 code = 1;           // 错误码
  string message = 2;       // 错误消息
  repeated ErrorDetail details = 3;  // 错误详情
}

// ErrorDetail 错误详细信息
message ErrorDetail {
  string field = 1;         // 字段名
  string issue = 2;         // 问题描述
  string value = 3;         // 当前值
}

// 标准错误码
enum ErrorCode {
  ERROR_CODE_UNSPECIFIED = 0;
  ERROR_CODE_INVALID_ARGUMENT = 1;
  ERROR_CODE_NOT_FOUND = 2;
  ERROR_CODE_ALREADY_EXISTS = 3;
  ERROR_CODE_PERMISSION_DENIED = 4;
  ERROR_CODE_INTERNAL = 5;
}
```

---

## 6. 完整示例

### 6.1 创建新的 Proto 文件

假设我们要为 Orchestrator 服务添加 Workflow API:

#### 步骤 1: 创建目录结构

```bash
mkdir -p api/proto/orchestrator/workflow/v1
```

#### 步骤 2: 创建 workflow.proto

```protobuf
// api/proto/orchestrator/workflow/v1/workflow.proto
syntax = "proto3";

package orchestrator.workflow.v1;

option go_package = "github.com/kart-io/k8s-agent/api/proto/gen/go/orchestrator/workflow/v1;workflowv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";
import "common/pagination/v1/pagination.proto";

// WorkflowService Workflow 管理服务
service WorkflowService {
  // CreateWorkflow 创建工作流
  rpc CreateWorkflow(CreateWorkflowRequest) returns (Workflow) {
    option (google.api.http) = {
      post: "/api/v1/workflows"
      body: "*"
    };
  }

  // GetWorkflow 获取工作流详情
  rpc GetWorkflow(GetWorkflowRequest) returns (Workflow) {
    option (google.api.http) = {
      get: "/api/v1/workflows/{workflow_id}"
    };
  }

  // ListWorkflows 列出工作流
  rpc ListWorkflows(ListWorkflowsRequest) returns (ListWorkflowsResponse) {
    option (google.api.http) = {
      get: "/api/v1/workflows"
    };
  }

  // ExecuteWorkflow 执行工作流
  rpc ExecuteWorkflow(ExecuteWorkflowRequest) returns (WorkflowExecution) {
    option (google.api.http) = {
      post: "/api/v1/workflows/{workflow_id}/execute"
      body: "*"
    };
  }

  // CancelWorkflow 取消工作流执行
  rpc CancelWorkflow(CancelWorkflowRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      post: "/api/v1/workflows/{workflow_id}/cancel"
      body: "*"
    };
  }
}

// WorkflowStatus 工作流状态
enum WorkflowStatus {
  WORKFLOW_STATUS_UNSPECIFIED = 0;
  WORKFLOW_STATUS_DRAFT = 1;
  WORKFLOW_STATUS_ACTIVE = 2;
  WORKFLOW_STATUS_INACTIVE = 3;
  WORKFLOW_STATUS_DELETED = 4;
}

// ExecutionStatus 执行状态
enum ExecutionStatus {
  EXECUTION_STATUS_UNSPECIFIED = 0;
  EXECUTION_STATUS_PENDING = 1;
  EXECUTION_STATUS_RUNNING = 2;
  EXECUTION_STATUS_SUCCEEDED = 3;
  EXECUTION_STATUS_FAILED = 4;
  EXECUTION_STATUS_CANCELLED = 5;
}

// Workflow 工作流定义
message Workflow {
  string workflow_id = 1;
  string name = 2;
  string description = 3;
  WorkflowStatus status = 4;
  repeated WorkflowStep steps = 5;
  map<string, string> variables = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
  string created_by = 9;
}

// WorkflowStep 工作流步骤
message WorkflowStep {
  string step_id = 1;
  string name = 2;
  StepType type = 3;
  map<string, string> parameters = 4;
  repeated string depends_on = 5;  // 依赖的步骤 ID
}

// StepType 步骤类型
enum StepType {
  STEP_TYPE_UNSPECIFIED = 0;
  STEP_TYPE_COMMAND = 1;
  STEP_TYPE_AI_ANALYSIS = 2;
  STEP_TYPE_DECISION = 3;
  STEP_TYPE_REMEDIATION = 4;
  STEP_TYPE_NOTIFICATION = 5;
  STEP_TYPE_WAIT = 6;
}

// WorkflowExecution 工作流执行
message WorkflowExecution {
  string execution_id = 1;
  string workflow_id = 2;
  ExecutionStatus status = 3;
  repeated StepExecution steps = 4;
  google.protobuf.Timestamp started_at = 5;
  google.protobuf.Timestamp completed_at = 6;
  string error_message = 7;
}

// StepExecution 步骤执行
message StepExecution {
  string step_id = 1;
  string name = 2;
  ExecutionStatus status = 3;
  google.protobuf.Timestamp started_at = 4;
  google.protobuf.Timestamp completed_at = 5;
  string output = 6;
  string error = 7;
}

// CreateWorkflowRequest 创建工作流请求
message CreateWorkflowRequest {
  string name = 1;
  string description = 2;
  repeated WorkflowStep steps = 3;
  map<string, string> variables = 4;
}

// GetWorkflowRequest 获取工作流请求
message GetWorkflowRequest {
  string workflow_id = 1;
}

// ListWorkflowsRequest 列出工作流请求
message ListWorkflowsRequest {
  optional WorkflowStatus status = 1;
  common.pagination.v1.PageRequest page = 2;
}

// ListWorkflowsResponse 列出工作流响应
message ListWorkflowsResponse {
  repeated Workflow workflows = 1;
  common.pagination.v1.PageResponse page = 2;
}

// ExecuteWorkflowRequest 执行工作流请求
message ExecuteWorkflowRequest {
  string workflow_id = 1;
  map<string, string> variables = 2;  // 运行时变量
  string triggered_by = 3;
}

// CancelWorkflowRequest 取消工作流请求
message CancelWorkflowRequest {
  string workflow_id = 1;
  string execution_id = 2;
  string reason = 3;
}
```

#### 步骤 3: 运行 Buf 检查和生成

```bash
cd api/proto

# Lint
buf lint --path orchestrator/

# 生成代码
buf generate --path orchestrator/

# 或者使用 Makefile
make buf-lint
make buf-generate
```

#### 步骤 4: 使用生成的代码

```go
// internal/orchestrator/server/workflow_server.go
package server

import (
	"context"

	workflowv1 "github.com/kart-io/k8s-agent/api/proto/gen/go/orchestrator/workflow/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type WorkflowServer struct {
	workflowv1.UnimplementedWorkflowServiceServer
	// dependencies
}

func (s *WorkflowServer) CreateWorkflow(
	ctx context.Context,
	req *workflowv1.CreateWorkflowRequest,
) (*workflowv1.Workflow, error) {
	// Implementation
	return &workflowv1.Workflow{
		WorkflowId: "wf-123",
		Name:       req.Name,
		// ...
	}, nil
}

func (s *WorkflowServer) GetWorkflow(
	ctx context.Context,
	req *workflowv1.GetWorkflowRequest,
) (*workflowv1.Workflow, error) {
	// Implementation
	return nil, nil
}

// ... 其他方法实现
```

---

## 7. CI/CD 集成

### 7.1 GitHub Actions

创建 `.github/workflows/proto.yml`:

```yaml
name: Proto CI

on:
  pull_request:
    paths:
      - 'api/proto/**'
  push:
    branches:
      - main
    paths:
      - 'api/proto/**'

jobs:
  proto-lint:
    name: Lint Proto Files
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Buf
        uses: bufbuild/buf-setup-action@v1
        with:
          version: '1.28.1'

      - name: Lint
        working-directory: api/proto
        run: buf lint

      - name: Format Check
        working-directory: api/proto
        run: buf format --diff --exit-code

  proto-breaking:
    name: Breaking Change Detection
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4

      - name: Setup Buf
        uses: bufbuild/buf-setup-action@v1
        with:
          version: '1.28.1'

      - name: Breaking Change Detection
        working-directory: api/proto
        run: |
          buf breaking --against \
            'https://github.com/${{ github.repository }}.git#branch=main,subdir=api/proto'

  proto-generate:
    name: Generate Proto Code
    runs-on: ubuntu-latest
    needs: [proto-lint]
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Setup Buf
        uses: bufbuild/buf-setup-action@v1
        with:
          version: '1.28.1'

      - name: Generate
        working-directory: api/proto
        run: buf generate

      - name: Verify Generated Code
        run: |
          # 确保生成的代码可以编译
          cd api/proto/gen/go
          go mod tidy
          go build ./...
```

### 7.2 Pre-commit Hook

创建 `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Proto pre-commit hook

echo "Running proto checks..."

cd api/proto

# Lint
echo "Linting proto files..."
if ! buf lint; then
  echo "❌ Buf lint failed"
  exit 1
fi

# Format
echo "Checking proto format..."
if ! buf format --diff --exit-code; then
  echo "❌ Proto files are not formatted"
  echo "Run: cd api/proto && buf format -w"
  exit 1
fi

echo "✅ Proto checks passed"
exit 0
```

使其可执行:

```bash
chmod +x .git/hooks/pre-commit
```

---

## 8. 最佳实践总结

### 8.1 DO (应该做的)

- ✅ 每个 RPC 使用独特的请求/响应消息
- ✅ 枚举的第一个值是 `UNSPECIFIED = 0`
- ✅ 保留已删除字段的编号
- ✅ 使用 `google.protobuf.Timestamp` 表示时间
- ✅ 使用 `optional` 标记可选字段
- ✅ 为所有服务添加 gRPC-Gateway 注解
- ✅ 定期运行 `buf lint` 和 `buf breaking`
- ✅ 使用语义化版本 (v1, v2)

### 8.2 DON'T (不应该做的)

- ❌ 不要重用字段编号
- ❌ 不要在稳定版本中引入破坏性变更
- ❌ 不要使用通用的请求/响应消息名称
- ❌ 不要忽略 lint 警告
- ❌ 不要手动编辑生成的代码
- ❌ 不要将生成的代码提交到 Git (除非必要)

### 8.3 版本演进策略

```
v1 (Stable)
  ↓ 添加新字段 (向后兼容)
v1.1 (Stable)
  ↓ 添加新 RPC (向后兼容)
v1.2 (Stable)
  ↓ 需要破坏性变更
v2alpha1 (Experimental)
  ↓ 收集反馈,迭代
v2beta1 (Pre-release)
  ↓ 稳定后
v2 (Stable)
```

---

## 9. 故障排查

### 9.1 常见问题

#### 问题 1: `buf: command not found`

```bash
# 解决方案: 安装 buf
make install-buf

# 或手动安装
brew install bufbuild/buf/buf
```

#### 问题 2: `module not found: buf.build/googleapis/googleapis`

```bash
# 解决方案: 更新依赖
cd api/proto
buf dep update
```

#### 问题 3: Lint 错误 `PACKAGE_VERSION_SUFFIX`

```yaml
# buf.yaml 中添加例外
lint:
  except:
    - PACKAGE_VERSION_SUFFIX
```

#### 问题 4: 生成的代码导入路径不正确

```yaml
# buf.gen.yaml 中配置 managed mode
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/kart-io/k8s-agent/api/proto/gen/go
```

### 9.2 调试技巧

```bash
# 查看 buf 配置
buf config ls-lint-rules
buf config ls-breaking-rules

# 查看依赖树
buf dep tree

# 导出模块信息
buf export api/proto -o /tmp/proto-export

# 查看详细日志
buf -v lint
buf -v generate
```

---

## 10. 参考资源

### 官方文档

- [Buf 官方文档](https://buf.build/docs/)
- [Buf CLI 参考](https://buf.build/docs/reference/cli/)
- [Buf Schema Registry](https://buf.build/docs/bsr/)
- [Protocol Buffers 官方文档](https://protobuf.dev/)
- [gRPC-Gateway 文档](https://grpc-ecosystem.github.io/grpc-gateway/)

### 最佳实践

- [Protocol Buffers Style Guide](https://protobuf.dev/programming-guides/style/)
- [API Design Guide (Google)](https://cloud.google.com/apis/design)
- [Buf Best Practices](https://buf.build/docs/best-practices/style-guide/)

### 工具

- [Buf CLI](https://github.com/bufbuild/buf)
- [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go)
- [protoc-gen-go-grpc](https://github.com/grpc/grpc-go)
- [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)

---

**文档版本**: v1.0.0
**最后更新**: 2025-10-23
**维护者**: Aetherius Team
