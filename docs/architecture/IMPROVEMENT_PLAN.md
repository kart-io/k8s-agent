# Aetherius 项目完善方案

基于 [OneX Cloud-Native Platform](https://github.com/onexstack/onex) 的最佳实践和 Go 项目标准布局

## 文档版本

- **版本**: v1.0.0
- **创建日期**: 2025-10-23
- **作者**: Aetherius Team
- **参考**: OneX v2, golang-standards/project-layout

---

## 1. 执行摘要

本文档基于对 OneX 云原生平台的深入分析,结合 Go 语言社区最佳实践,为 Aetherius 项目提供系统性改进方案。OneX 作为企业级云原生学习平台,在代码质量、项目结构、开发规范等方面树立了标杆,其经验对 Aetherius 的完善具有重要参考价值。

### 核心改进领域

1. **项目结构优化** - 采用标准化目录布局,提升代码可维护性
2. **构建流程增强** - 参考 OneX Makefile,完善自动化工作流
3. **代码质量提升** - 引入严格的代码规范和质量检查
4. **API 设计优化** - 统一 API 设计模式和错误处理
5. **文档体系完善** - 建立全面的技术文档和开发指南
6. **配置管理改进** - 优化多环境配置管理机制
7. **测试策略强化** - 建立完整的测试金字塔
8. **可观测性增强** - 完善日志、指标、追踪体系

---

## 2. OneX 项目分析总结

### 2.1 OneX 核心特点

#### 技术栈全面性

OneX 覆盖四大技术栈:

- **基础软件开发**: Linux、Shell、Makefile、Git
- **Go 语言开发**: Web 编程、SDK、认证、CLI 工具、Web 服务、分布式任务
- **云原生技术**: Kubernetes 编程、Docker 实践、声明式编程、不可变基础设施
- **微服务架构**: 服务网格、API 网关、服务治理

#### 代码质量标准

- **编程模式**: 支持命令式和声明式两种编程范式
- **架构简洁**: 清晰的分层架构,职责分离明确
- **代码健壮**: 完善的错误处理和边界检查
- **性能优异**: 高性能接口设计,优化关键路径
- **可维护性**: 高内聚低耦合,易于扩展和修改

#### 规范化体系

- **目录规范**: 遵循 golang-standards/project-layout
- **代码规范**: 统一的代码风格和命名约定
- **日志规范**: 结构化日志,统一日志级别
- **错误码规范**: 统一的错误码体系
- **文档规范**: 完整的 API 文档和开发文档
- **提交规范**: 规范的 Git Commit Message
- **版本规范**: 语义化版本管理

### 2.2 OneX 项目结构

基于 Go 包文档和搜索结果,OneX 采用标准布局:

```
onex/
├── cmd/                    # 主应用程序
│   ├── onex-apiserver/    # API 服务器 (主控制面)
│   ├── onex-gateway/      # 网关服务 (后端门户)
│   └── onex-usercenter/   # 用户中心
├── internal/              # 私有应用和库代码
│   ├── apiserver/        # API 服务器内部实现
│   ├── gateway/          # 网关内部实现
│   └── usercenter/       # 用户中心内部实现
│       └── store/        # 存储层
├── pkg/                   # 可被外部项目导入的库
│   ├── api/              # API 定义和客户端
│   ├── streams/          # 流处理 (使用泛型)
│   └── ...
├── api/                   # API 协议定义
│   ├── v1beta1/          # API v1beta1 版本
│   └── v1beta2/          # API v1beta2 版本 (计划中)
├── build/                 # 打包和 CI
├── configs/               # 配置文件模板
├── deployments/           # 部署配置 (IaaS, PaaS, Kubernetes)
├── docs/                  # 设计和用户文档
├── examples/              # 示例应用
├── scripts/               # 构建、安装、分析等脚本
├── test/                  # 额外的外部测试应用和数据
├── tools/                 # 项目支持工具
├── vendor/                # 依赖管理 (可选)
├── Makefile               # 构建自动化
└── go.mod                 # Go 模块定义
```

### 2.3 OneX Makefile 核心能力

#### 完整的构建流程

- `make all` - 完整流程: format → tidy → gen → copyright → lint → cover → build
- `make build` / `make build.multiarch` - 单平台/多平台构建
- `make gen` / `make gen-k8s` - 代码生成 (CI 配置、K8s 资源)
- `make protoc` - Protocol Buffers 编译

#### Docker 和部署

- `make image` / `make image.multiarch` - 容器镜像构建
- `make push` / `make push.multiarch` - 镜像推送
- `make deploy` - K8s 部署
- `make docker-install/uninstall` - Docker 部署
- `make sbs-install/uninstall` - 手动分步部署

#### 质量保证

- `make lint` - 代码检查 (`A=1` 启用全面检查)
- `make format` - 代码格式化 (gofmt, gofumpt, goimports)
- `make test` / `make cover` - 单元测试和覆盖率
- `make add-copyright` - 添加许可证头

#### 灵活配置

- `BINS` - 指定构建的服务
- `IMAGES` - 指定构建的镜像
- `PLATFORMS` - 目标平台 (linux/amd64, linux/arm64 等)
- `REGISTRY_PREFIX` - 镜像仓库前缀
- `VERSION` - 版本号
- `DBG` - 调试模式

---

## 3. Aetherius 现状分析

### 3.1 优势

#### 架构设计

- ✅ **清晰的四层架构**: Collect Agent → Agent Manager → Orchestrator → Reasoning Service
- ✅ **事件驱动设计**: 基于 NATS 的消息总线
- ✅ **微服务架构**: 服务职责分离明确
- ✅ **多集群管理**: 支持统一管理数百个 K8s 集群

#### 技术栈

- ✅ **现代技术栈**: Go 1.21+, MySQL 8.0+, Redis 6+, NATS 2.10+
- ✅ **AI 集成**: Reasoning Service 集成 OpenAI/Gemini/DeepSeek
- ✅ **云原生**: K8s 原生设计,支持容器化部署
- ✅ **监控完善**: Prometheus + Grafana 监控体系

#### 功能实现

- ✅ **功能完整**: FR-1 到 FR-18 全部实现 (100%)
- ✅ **工作流引擎**: 支持 6 种步骤类型的复杂工作流
- ✅ **知识图谱**: Neo4j 存储运维经验
- ✅ **认证授权**: JWT + RBAC + 会话管理

### 3.2 改进空间

#### 项目结构

- ⚠️ **目录结构不够标准化**: 混合使用 `internal/` 和服务目录,未充分利用 `pkg/`
- ⚠️ **模块划分不够清晰**: 部分共享代码放在 `common/` 而非 `pkg/` 或 `internal/pkg/`
- ⚠️ **服务代码组织**: 服务内部结构不够一致

#### 构建流程

- ⚠️ **Makefile 功能简单**: 缺少代码生成、格式化、版权检查等目标
- ⚠️ **CI/CD 流程不完整**: 缺少完整的 CI pipeline 定义
- ⚠️ **多平台构建支持**: 虽然支持但流程不够自动化
- ⚠️ **版本注入**: 存在但可以更加规范化

#### 代码质量

- ⚠️ **代码规范**: 缺少统一的代码规范文档
- ⚠️ **错误处理**: 错误码体系存在但不够统一
- ⚠️ **日志规范**: 使用了 kart-io/logger 但规范化程度可提升
- ⚠️ **注释和文档**: 部分代码缺少充分的注释

#### API 设计

- ⚠️ **API 版本管理**: 缺少明确的 API 版本演进策略
- ⚠️ **API 文档**: 缺少 OpenAPI/Swagger 规范
- ⚠️ **错误响应**: 错误响应格式需要统一

#### 测试

- ⚠️ **测试覆盖率**: 整体测试覆盖率需要提升
- ⚠️ **测试类型**: 缺少完整的集成测试和 E2E 测试
- ⚠️ **性能测试**: 缺少基准测试和性能测试

#### 文档

- ⚠️ **开发文档**: 缺少详细的开发指南和贡献指南
- ⚠️ **API 文档**: API 文档不够完善
- ⚠️ **运维文档**: 运维手册需要补充

---

## 4. 改进方案详细设计

### 4.1 项目结构重组

#### 4.1.1 目标结构

```
k8s-agent/ (Aetherius)
├── api/                          # API 定义 (公共接口)
│   ├── proto/                   # Protocol Buffers 定义
│   │   ├── v1/                 # API v1 版本
│   │   │   ├── agent/          # Agent API
│   │   │   ├── orchestrator/   # Orchestrator API
│   │   │   └── reasoning/      # Reasoning API
│   │   └── common/             # 通用 Proto 定义
│   └── openapi/                 # OpenAPI/Swagger 规范
│       └── v1/                 # API v1 OpenAPI 定义
├── build/                        # 打包和持续集成
│   ├── docker/                  # Dockerfiles
│   │   ├── agent-manager.Dockerfile
│   │   ├── orchestrator.Dockerfile
│   │   └── ...
│   ├── ci/                      # CI 配置
│   │   ├── .github/            # GitHub Actions
│   │   ├── .gitlab-ci.yml      # GitLab CI
│   │   └── jenkins/            # Jenkins 配置
│   └── package/                 # 系统包 (deb, rpm, etc.)
├── cmd/                          # 主应用程序入口
│   ├── agent-manager/           # Agent Manager 服务
│   │   └── main.go
│   ├── orchestrator/            # Orchestrator 服务
│   │   └── main.go
│   ├── reasoning/               # Reasoning 服务
│   │   └── main.go
│   ├── auth/                    # Auth 服务
│   │   └── main.go
│   ├── collect-agent/           # Collect Agent
│   │   └── main.go
│   └── tools/                   # 辅助工具
│       ├── migration/          # 数据库迁移工具
│       └── cli/                # CLI 工具
├── configs/                      # 配置文件模板和默认配置
│   ├── default.yaml             # 默认配置
│   ├── agent-manager/
│   │   ├── config.yaml
│   │   ├── config-dev.yaml
│   │   ├── config-staging.yaml
│   │   └── config-prod.yaml
│   └── ...
├── deployments/                  # IaaS, PaaS, 系统和容器编排部署配置和模板
│   ├── docker-compose/          # Docker Compose 部署
│   ├── k8s/                     # Kubernetes 部署
│   │   ├── base/               # 基础配置
│   │   └── overlays/           # 环境覆盖 (dev/staging/prod)
│   ├── helm/                    # Helm Charts (新增)
│   └── terraform/               # Terraform 配置 (新增)
├── docs/                         # 设计和用户文档
│   ├── devel/                   # 开发文档
│   │   ├── conventions/        # 规范和约定
│   │   ├── guide/              # 开发指南
│   │   └── api/                # API 开发文档
│   ├── user/                    # 用户文档
│   │   ├── guide/              # 用户指南
│   │   └── tutorials/          # 教程
│   ├── architecture/            # 架构文档
│   │   ├── SYSTEM_ARCHITECTURE.md
│   │   ├── ADR/                # Architecture Decision Records
│   │   └── diagrams/           # 架构图
│   └── operations/              # 运维文档
│       ├── deployment/         # 部署指南
│       ├── monitoring/         # 监控指南
│       └── troubleshooting/    # 故障排查
├── examples/                     # 应用示例和配置示例
│   ├── configs/                 # 配置示例
│   ├── workflows/               # 工作流示例
│   └── scripts/                 # 脚本示例
├── init/                         # 系统初始化 (systemd, upstart, sysv)
│   ├── systemd/
│   └── docker-init/
├── internal/                     # 私有应用和库代码
│   ├── agent-manager/           # Agent Manager 服务实现
│   │   ├── agent/              # Agent 管理
│   │   ├── cluster/            # 集群管理
│   │   ├── command/            # 命令调度
│   │   ├── event/              # 事件处理
│   │   ├── server/             # HTTP/gRPC 服务器
│   │   └── store/              # 存储层
│   ├── orchestrator/            # Orchestrator 服务实现
│   │   ├── workflow/           # 工作流引擎
│   │   ├── strategy/           # 诊断策略
│   │   ├── executor/           # 执行器
│   │   ├── server/             # HTTP/gRPC 服务器
│   │   └── store/              # 存储层
│   ├── reasoning/               # Reasoning 服务实现
│   │   ├── analyzer/           # 根因分析
│   │   ├── recommender/        # 推荐引擎
│   │   ├── knowledge/          # 知识图谱
│   │   ├── llm/                # LLM 集成
│   │   ├── server/             # HTTP/gRPC 服务器
│   │   └── store/              # 存储层
│   ├── auth/                    # Auth 服务实现
│   ├── collect-agent/           # Collect Agent 实现
│   └── pkg/                     # 内部共享包 (不可被外部导入)
│       ├── config/             # 配置管理
│       ├── db/                 # 数据库工具
│       ├── logger/             # 日志工具 (包装 kart-io/logger)
│       ├── middleware/         # 中间件
│       ├── response/           # 响应处理
│       ├── telemetry/          # 遥测
│       ├── validator/          # 验证器
│       └── version/            # 版本管理
├── pkg/                          # 可被外部项目导入的公共库
│   ├── client/                  # 客户端库
│   │   ├── agent-manager/      # Agent Manager 客户端
│   │   ├── orchestrator/       # Orchestrator 客户端
│   │   └── reasoning/          # Reasoning 客户端
│   ├── types/                   # 公共类型定义
│   ├── utils/                   # 通用工具函数
│   └── errors/                  # 错误定义
├── scripts/                      # 构建、安装、分析等脚本
│   ├── make-rules/              # Makefile 子规则
│   ├── install/                 # 安装脚本
│   ├── migration/               # 迁移脚本
│   └── lib/                     # 脚本库
├── test/                         # 额外的外部测试应用和数据
│   ├── integration/             # 集成测试
│   ├── e2e/                     # 端到端测试
│   ├── testdata/                # 测试数据
│   └── fixtures/                # 测试夹具
├── third_party/                  # 外部辅助工具、分支代码和其他第三方工具
├── tools/                        # 项目支持工具
│   ├── codegen/                 # 代码生成工具
│   └── migration/               # 迁移工具
├── vendor/                       # 应用依赖 (可选,使用 Go modules)
├── web/                          # Web 应用特定组件 (如有前端)
│   ├── static/                  # 静态资源
│   └── templates/               # 模板
├── .gitignore
├── .golangci.yml                # golangci-lint 配置
├── CHANGELOG.md                 # 变更日志
├── CODE_OF_CONDUCT.md           # 行为准则
├── CONTRIBUTING.md              # 贡献指南
├── LICENSE                      # 许可证
├── Makefile                     # 构建自动化
├── README.md                    # 项目说明
├── SECURITY.md                  # 安全政策
├── go.mod                       # Go 模块定义
└── go.sum                       # Go 模块校验和
```

#### 4.1.2 迁移策略

##### 阶段 1: 准备工作 (1-2 天)

1. **备份当前代码**: 创建 feature branch `feature/restructure`
2. **依赖分析**: 分析现有代码的依赖关系
3. **制定迁移清单**: 列出所有需要移动的文件和包

##### 阶段 2: 目录创建和基础迁移 (2-3 天)

1. **创建新目录结构**: 按照目标结构创建所有目录
2. **迁移 cmd/**: 保持 `cmd/` 相对简单,主要包含 `main.go`
3. **迁移 internal/**: 将服务实现代码移至 `internal/<service>/`
4. **创建 internal/pkg/**: 将 `common/` 中的私有包移至 `internal/pkg/`

##### 阶段 3: 创建公共库 (2-3 天)

1. **创建 pkg/**: 提取可被外部使用的库至 `pkg/`
2. **pkg/client/**: 创建各服务的客户端库
3. **pkg/types/**: 提取公共类型定义
4. **pkg/errors/**: 统一错误定义

##### 阶段 4: API 重组 (2-3 天)

1. **api/proto/v1/**: 按服务重组 Proto 定义
2. **生成代码更新**: 更新 Proto 生成代码的路径
3. **创建 api/openapi/**: 添加 OpenAPI 规范

##### 阶段 5: 更新导入路径 (3-4 天)

1. **全局搜索替换**: 更新所有导入路径
2. **go.mod 更新**: 更新模块路径和 replace 指令
3. **编译验证**: 确保所有服务可以编译

##### 阶段 6: 配置和脚本迁移 (1-2 天)

1. **configs/**: 重组配置文件
2. **scripts/**: 迁移和优化脚本
3. **deployments/**: 更新部署配置

##### 阶段 7: 测试和文档 (2-3 天)

1. **test/**: 创建测试目录结构
2. **docs/**: 重组文档
3. **运行测试**: 确保所有测试通过

##### 阶段 8: 收尾和发布 (1-2 天)

1. **代码审查**: 全面代码审查
2. **文档更新**: 更新所有文档中的路径引用
3. **合并主分支**: 合并 feature branch

**总计**: 约 14-22 天 (3-4 周)

#### 4.1.3 迁移工具脚本

创建 `tools/migration/restructure.sh` 自动化迁移脚本:

```bash
#!/bin/bash
# 项目结构重组脚本

set -e

PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"

echo "Starting project restructure..."

# 阶段 1: 创建新目录结构
echo "Creating new directory structure..."
mkdir -p api/proto/v1/{agent,orchestrator,reasoning,common}
mkdir -p api/openapi/v1
mkdir -p build/{ci,package}
mkdir -p docs/devel/{conventions,guide,api}
mkdir -p docs/user/{guide,tutorials}
mkdir -p docs/architecture/ADR
mkdir -p docs/operations/{deployment,monitoring,troubleshooting}
mkdir -p internal/pkg
mkdir -p pkg/{client,types,utils,errors}
mkdir -p scripts/{make-rules,install,migration,lib}
mkdir -p test/{integration,e2e,testdata,fixtures}
mkdir -p tools/codegen

# 阶段 2: 迁移 common/ 到 internal/pkg/
echo "Migrating common/ to internal/pkg/..."
for dir in common/*/; do
  dirname=$(basename "$dir")
  if [ -d "$dir" ]; then
    git mv "$dir" "internal/pkg/$dirname"
  fi
done

# 阶段 3: 更新导入路径 (示例)
echo "Updating import paths..."
find . -name "*.go" -type f -not -path "./vendor/*" -exec sed -i \
  's|github.com/kart-io/k8s-agent/common/|github.com/kart-io/k8s-agent/internal/pkg/|g' {} \;

echo "Restructure phase 1 complete!"
echo "Next steps: Run tests, update documentation, review changes"
```

### 4.2 Makefile 增强

#### 4.2.1 参考 OneX 的 Makefile 结构

创建模块化的 Makefile 系统:

```makefile
# scripts/make-rules/common.mk - 通用变量和函数
# scripts/make-rules/golang.mk - Go 构建规则
# scripts/make-rules/docker.mk - Docker 规则
# scripts/make-rules/gen.mk - 代码生成规则
# scripts/make-rules/tools.mk - 工具安装
```

根 Makefile 引入子规则:

```makefile
include scripts/make-rules/common.mk
include scripts/make-rules/golang.mk
include scripts/make-rules/docker.mk
include scripts/make-rules/gen.mk
include scripts/make-rules/tools.mk
```

#### 4.2.2 新增目标

```makefile
# 代码生成
.PHONY: gen
gen: gen-proto gen-openapi gen-mock ## Generate all code

.PHONY: gen-proto
gen-proto: ## Generate protobuf code
	@echo "Generating protobuf..."
	@cd api/proto && buf generate

.PHONY: gen-openapi
gen-openapi: ## Generate OpenAPI specs
	@echo "Generating OpenAPI..."
	@swag init -g cmd/agent-manager/main.go -o api/openapi/v1/agent-manager

.PHONY: gen-mock
gen-mock: ## Generate mock code
	@echo "Generating mocks..."
	@go generate ./...

# 代码格式化
.PHONY: format
format: ## Format code (gofmt, gofumpt, goimports)
	@echo "Formatting code..."
	@gofmt -s -w .
	@gofumpt -l -w .
	@goimports -w -local github.com/kart-io/k8s-agent .

# 版权检查
.PHONY: copyright
copyright: ## Add/check copyright headers
	@echo "Checking copyright headers..."
	@./scripts/add-copyright.sh

# 完整 CI 流程
.PHONY: all
all: format deps gen copyright lint test build ## Complete CI pipeline
```

#### 4.2.3 多平台构建增强

```makefile
# Platform-specific builds
.PHONY: build.multiarch
build.multiarch: ## Build for multiple architectures
	@echo "Building for multiple platforms..."
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) \
			-ldflags "$(LDFLAGS)" \
			-o $(BIN_DIR)/$(BINS)_$${os}_$${arch} \
			$(CMD_DIR)/$(BINS); \
	done
```

### 4.3 代码质量规范

#### 4.3.1 golangci-lint 配置

创建 `.golangci.yml`:

```yaml
run:
  timeout: 5m
  go: '1.25'

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - bodyclose
    - gocyclo
    - misspell
    - unparam
    - unconvert
    - gocognit
    - goconst
    - gofumpt
    - revive

linters-settings:
  gocyclo:
    min-complexity: 15
  gocognit:
    min-complexity: 20
  goconst:
    min-len: 3
    min-occurrences: 3
  goimports:
    local-prefixes: github.com/kart-io/k8s-agent

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

#### 4.3.2 错误码规范

创建统一的错误码体系 `pkg/errors/codes.go`:

```go
package errors

import "fmt"

// ErrorCode 错误码类型
type ErrorCode int

const (
	// 通用错误 (1000-1999)
	ErrCodeSuccess        ErrorCode = 0
	ErrCodeUnknown        ErrorCode = 1000
	ErrCodeInternal       ErrorCode = 1001
	ErrCodeInvalidParam   ErrorCode = 1002
	ErrCodeNotFound       ErrorCode = 1003
	ErrCodeAlreadyExists  ErrorCode = 1004
	ErrCodeUnauthorized   ErrorCode = 1005
	ErrCodeForbidden      ErrorCode = 1006

	// Agent Manager 错误 (2000-2999)
	ErrCodeAgentNotFound     ErrorCode = 2001
	ErrCodeAgentOffline      ErrorCode = 2002
	ErrCodeClusterNotFound   ErrorCode = 2003
	ErrCodeCommandFailed     ErrorCode = 2004

	// Orchestrator 错误 (3000-3999)
	ErrCodeWorkflowNotFound  ErrorCode = 3001
	ErrCodeWorkflowFailed    ErrorCode = 3002
	ErrCodeStrategyNotFound  ErrorCode = 3003

	// Reasoning 错误 (4000-4999)
	ErrCodeAnalysisFailed    ErrorCode = 4001
	ErrCodeLLMError          ErrorCode = 4002
	ErrCodeKnowledgeNotFound ErrorCode = 4003
)

// Error 实现 error 接口
type Error struct {
	Code    ErrorCode
	Message string
	Details interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建新错误
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WithDetails 添加详细信息
func (e *Error) WithDetails(details interface{}) *Error {
	e.Details = details
	return e
}
```

#### 4.3.3 日志规范

基于 `kart-io/logger` 的统一日志规范:

```go
// internal/pkg/logger/logger.go
package logger

import (
	"context"

	kartlogger "github.com/kart-io/logger"
)

// 全局 logger 实例
var defaultLogger *kartlogger.Logger

// Init 初始化全局 logger
func Init(config *kartlogger.Config) error {
	logger, err := kartlogger.New(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// WithContext 从 context 获取 logger
func WithContext(ctx context.Context) *kartlogger.Logger {
	// 从 context 中提取 trace_id, user_id 等字段
	logger := defaultLogger
	if traceID := ctx.Value("trace_id"); traceID != nil {
		logger = logger.WithField("trace_id", traceID)
	}
	return logger
}

// 标准字段约定
const (
	FieldTraceID   = "trace_id"
	FieldUserID    = "user_id"
	FieldRequestID = "request_id"
	FieldService   = "service"
	FieldOperation = "operation"
	FieldDuration  = "duration_ms"
	FieldError     = "error"
)
```

使用示例:

```go
// 在 HTTP handler 中
logger := logger.WithContext(ctx)
logger.WithFields(map[string]interface{}{
	logger.FieldOperation: "CreateWorkflow",
	logger.FieldUserID:    userID,
}).Info("Creating workflow")

// 记录错误
if err != nil {
	logger.WithFields(map[string]interface{}{
		logger.FieldError:     err.Error(),
		logger.FieldOperation: "CreateWorkflow",
	}).Error("Failed to create workflow")
}
```

### 4.4 API 设计规范

#### 4.4.1 RESTful API 规范

##### URL 设计

```
# 资源命名 (使用复数)
GET    /api/v1/agents                 # 列表
GET    /api/v1/agents/{id}            # 详情
POST   /api/v1/agents                 # 创建
PUT    /api/v1/agents/{id}            # 更新
DELETE /api/v1/agents/{id}            # 删除
PATCH  /api/v1/agents/{id}            # 部分更新

# 子资源
GET    /api/v1/agents/{id}/commands   # Agent 的命令列表
POST   /api/v1/agents/{id}/commands   # 发送命令

# 动作 (非 CRUD)
POST   /api/v1/agents/{id}/restart    # 重启 Agent
POST   /api/v1/workflows/{id}/execute # 执行工作流
```

##### 请求/响应格式

```go
// 统一响应格式
type Response struct {
	Code    int         `json:"code"`              // 业务错误码
	Message string      `json:"message"`           // 消息
	Data    interface{} `json:"data,omitempty"`    // 数据
	TraceID string      `json:"trace_id,omitempty"` // 追踪 ID
}

// 分页响应
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Pagination Pagination `json:"pagination"`
	TraceID string      `json:"trace_id,omitempty"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// 错误响应
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	TraceID string      `json:"trace_id"`
}
```

#### 4.4.2 API 版本管理

```
# 目录结构
api/
├── proto/
│   ├── v1/          # Stable API
│   └── v2alpha1/    # Alpha API (新功能预览)
└── openapi/
    ├── v1/
    └── v2alpha1/

# 版本演进路径
v1 -> v2alpha1 -> v2beta1 -> v2
```

#### 4.4.3 OpenAPI 规范

使用 Swag 生成 OpenAPI 文档:

```go
// cmd/agent-manager/main.go

// @title Aetherius Agent Manager API
// @version v1.0.0
// @description Enterprise-grade intelligent Kubernetes operations platform
// @contact.name API Support
// @contact.email support@aetherius.io
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// ...
}

// 在 handler 中添加注释
// @Summary List all agents
// @Description Get a list of all registered agents
// @Tags agents
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} PageResponse{data=[]Agent}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /agents [get]
// @Security BearerAuth
func (h *AgentHandler) List(c *gin.Context) {
	// ...
}
```

生成命令:

```bash
make gen-openapi
# 或
swag init -g cmd/agent-manager/main.go -o api/openapi/v1/agent-manager
```

### 4.5 测试策略

#### 4.5.1 测试金字塔

```
        /\
       /E2E\        (少量) - 端到端测试
      /------\
     /Integ. \      (适量) - 集成测试
    /----------\
   /Unit Tests \   (大量) - 单元测试
  /--------------\
```

#### 4.5.2 单元测试规范

```go
// internal/agent-manager/agent/service_test.go
package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

// 测试文件命名: <file>_test.go
// 测试函数命名: Test<FunctionName>_<Scenario>

func TestAgentService_Register_Success(t *testing.T) {
	// Arrange
	mockStore := new(MockAgentStore)
	service := NewAgentService(mockStore)

	agent := &Agent{
		ID:        "agent-1",
		ClusterID: "cluster-1",
		Status:    StatusOnline,
	}

	mockStore.On("Create", mock.Anything, agent).Return(nil)

	// Act
	err := service.Register(context.Background(), agent)

	// Assert
	require.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestAgentService_Register_AlreadyExists(t *testing.T) {
	// Arrange
	mockStore := new(MockAgentStore)
	service := NewAgentService(mockStore)

	agent := &Agent{ID: "agent-1"}
	mockStore.On("Create", mock.Anything, agent).
		Return(errors.New(ErrCodeAlreadyExists, "agent already exists"))

	// Act
	err := service.Register(context.Background(), agent)

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrCodeAlreadyExists, err.(*errors.Error).Code)
}

// 表驱动测试
func TestValidateAgentID(t *testing.T) {
	tests := []struct {
		name    string
		agentID string
		wantErr bool
	}{
		{"valid ID", "agent-1", false},
		{"empty ID", "", true},
		{"too long", strings.Repeat("a", 256), true},
		{"invalid chars", "agent@123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentID(tt.agentID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
```

#### 4.5.3 集成测试

```go
// test/integration/agent_manager_test.go
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/internal/agent-manager"
	"github.com/stretchr/testify/suite"
)

type AgentManagerTestSuite struct {
	suite.Suite
	server *agentmanager.Server
	client *agentmanager.Client
}

func (s *AgentManagerTestSuite) SetupSuite() {
	// 启动测试服务器
	config := &agentmanager.Config{
		Port: 18080,
		DB: &db.Config{
			Host: "localhost",
			Port: 3306,
			Database: "aetherius_test",
		},
	}

	server, err := agentmanager.NewServer(config)
	s.Require().NoError(err)

	go server.Run()
	time.Sleep(time.Second) // 等待服务器启动

	s.server = server
	s.client = agentmanager.NewClient("http://localhost:18080")
}

func (s *AgentManagerTestSuite) TearDownSuite() {
	s.server.Shutdown(context.Background())
}

func (s *AgentManagerTestSuite) TestAgentLifecycle() {
	ctx := context.Background()

	// 注册 Agent
	agent := &agentmanager.Agent{
		ID:        "test-agent-1",
		ClusterID: "test-cluster",
	}

	err := s.client.RegisterAgent(ctx, agent)
	s.Require().NoError(err)

	// 获取 Agent
	got, err := s.client.GetAgent(ctx, "test-agent-1")
	s.Require().NoError(err)
	s.Equal(agent.ID, got.ID)

	// 更新 Agent
	agent.Status = agentmanager.StatusOffline
	err = s.client.UpdateAgent(ctx, agent)
	s.Require().NoError(err)

	// 删除 Agent
	err = s.client.DeleteAgent(ctx, "test-agent-1")
	s.Require().NoError(err)
}

func TestAgentManagerSuite(t *testing.T) {
	suite.Run(t, new(AgentManagerTestSuite))
}
```

运行集成测试:

```bash
make test-integration
# 或
go test -tags=integration ./test/integration/...
```

#### 4.5.4 E2E 测试

```go
// test/e2e/workflow_test.go
// +build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// E2E 测试: 完整的故障诊断流程
func TestWorkflowE2E_PodCrashDiagnosis(t *testing.T) {
	ctx := context.Background()

	// 1. 模拟 Pod Crash 事件
	event := &Event{
		Type:      "PodCrash",
		Namespace: "default",
		PodName:   "test-pod",
		Reason:    "CrashLoopBackOff",
	}

	err := publishEvent(ctx, event)
	require.NoError(t, err)

	// 2. 等待工作流触发
	time.Sleep(5 * time.Second)

	// 3. 检查工作流是否创建
	workflows, err := orchestratorClient.ListWorkflows(ctx, &ListOptions{
		Filters: map[string]string{
			"event_type": "PodCrash",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, workflows)

	workflow := workflows[0]

	// 4. 等待工作流完成
	err = waitForWorkflowCompletion(ctx, workflow.ID, 2*time.Minute)
	require.NoError(t, err)

	// 5. 检查诊断结果
	result, err := reasoningClient.GetAnalysisResult(ctx, workflow.ID)
	require.NoError(t, err)
	require.NotNil(t, result.RootCause)
	require.NotEmpty(t, result.Recommendations)

	// 6. 检查修复建议是否正确
	hasRestartRecommendation := false
	for _, rec := range result.Recommendations {
		if rec.Action == "RestartPod" {
			hasRestartRecommendation = true
			break
		}
	}
	require.True(t, hasRestartRecommendation)
}
```

### 4.6 配置管理优化

#### 4.6.1 配置结构

```yaml
# configs/agent-manager/config.yaml
server:
  port: 8080
  mode: release  # debug, release
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: mysql
  host: localhost
  port: 3306
  database: aetherius
  username: aetherius
  password: ${DB_PASSWORD}  # 支持环境变量
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  host: localhost
  port: 6379
  password: ${REDIS_PASSWORD}
  db: 0
  pool_size: 100

nats:
  urls:
    - nats://localhost:4222
  max_reconnects: 10
  reconnect_wait: 2s

logger:
  engine: zap  # zap, slog
  level: info  # debug, info, warn, error
  format: json  # json, text
  output:
    - stdout
    - /var/log/aetherius/agent-manager.log
  otlp:
    enabled: true
    endpoint: http://localhost:4317

telemetry:
  enabled: true
  prometheus:
    enabled: true
    port: 9090
  jaeger:
    enabled: true
    endpoint: http://localhost:14268/api/traces

feature_flags:
  enable_ai_diagnosis: true
  enable_auto_remediation: false
```

#### 4.6.2 配置加载

```go
// internal/pkg/config/loader.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Loader struct {
	v *viper.Viper
}

func NewLoader() *Loader {
	return &Loader{
		v: viper.New(),
	}
}

func (l *Loader) Load(configFile string, cfg interface{}) error {
	// 1. 设置配置文件
	l.v.SetConfigFile(configFile)

	// 2. 设置环境变量前缀
	l.v.SetEnvPrefix("AETHERIUS")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()

	// 3. 读取配置文件
	if err := l.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 4. 解析到结构体
	if err := l.v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 5. 验证配置
	if validator, ok := cfg.(interface{ Validate() error }); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	return nil
}

// LoadWithEnv 支持多环境配置
func (l *Loader) LoadWithEnv(baseConfig string, cfg interface{}) error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	// 加载基础配置
	if err := l.Load(baseConfig, cfg); err != nil {
		return err
	}

	// 加载环境特定配置 (如果存在)
	envConfig := strings.Replace(baseConfig, ".yaml", fmt.Sprintf("-%s.yaml", env), 1)
	if _, err := os.Stat(envConfig); err == nil {
		l.v.SetConfigFile(envConfig)
		if err := l.v.MergeInConfig(); err != nil {
			return fmt.Errorf("failed to merge env config: %w", err)
		}
		if err := l.v.Unmarshal(cfg); err != nil {
			return fmt.Errorf("failed to unmarshal env config: %w", err)
		}
	}

	return nil
}
```

使用示例:

```go
// cmd/agent-manager/main.go
func main() {
	var cfg config.AgentManagerConfig

	loader := config.NewLoader()
	if err := loader.LoadWithEnv("configs/agent-manager/config.yaml", &cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 使用配置
	server := agentmanager.NewServer(&cfg)
	server.Run()
}
```

### 4.7 文档体系完善

#### 4.7.1 文档目录结构

```
docs/
├── README.md                    # 文档索引
├── devel/                       # 开发文档
│   ├── conventions/            # 规范和约定
│   │   ├── coding-style.md     # 代码风格
│   │   ├── commit-message.md   # 提交信息规范
│   │   ├── error-handling.md   # 错误处理规范
│   │   └── logging.md          # 日志规范
│   ├── guide/                  # 开发指南
│   │   ├── getting-started.md  # 快速开始
│   │   ├── project-structure.md # 项目结构
│   │   ├── building.md         # 构建指南
│   │   ├── testing.md          # 测试指南
│   │   └── debugging.md        # 调试指南
│   └── api/                    # API 开发文档
│       ├── rest-api-design.md  # REST API 设计
│       ├── grpc-api-design.md  # gRPC API 设计
│       └── versioning.md       # API 版本管理
├── user/                        # 用户文档
│   ├── guide/                  # 用户指南
│   │   ├── installation.md     # 安装指南
│   │   ├── configuration.md    # 配置指南
│   │   └── usage.md            # 使用指南
│   └── tutorials/              # 教程
│       ├── quick-start.md      # 快速入门
│       └── advanced.md         # 高级教程
├── architecture/                # 架构文档
│   ├── SYSTEM_ARCHITECTURE.md  # 系统架构
│   ├── ADR/                    # Architecture Decision Records
│   │   ├── 0001-use-nats.md
│   │   └── 0002-ai-integration.md
│   └── diagrams/               # 架构图
└── operations/                  # 运维文档
    ├── deployment/             # 部署指南
    │   ├── docker-compose.md
    │   ├── kubernetes.md
    │   └── helm.md
    ├── monitoring/             # 监控指南
    │   ├── metrics.md
    │   ├── logging.md
    │   └── tracing.md
    └── troubleshooting/        # 故障排查
        └── common-issues.md
```

#### 4.7.2 ADR (Architecture Decision Record) 模板

```markdown
# ADR-XXXX: [简短标题]

## 状态

[提议 | 已接受 | 已弃用 | 已替代]

## 上下文

[描述需要决策的背景和问题]

## 决策

[描述做出的决策]

## 后果

### 优点

- [列出决策的优点]

### 缺点

- [列出决策的缺点]

### 权衡

- [描述权衡考虑]

## 替代方案

### 方案 A

[描述替代方案 A]

### 方案 B

[描述替代方案 B]

## 相关决策

- ADR-YYYY: [相关决策]
```

#### 4.7.3 代码规范文档

创建 `docs/devel/conventions/coding-style.md`:

```markdown
# Go 代码风格指南

## 1. 命名规范

### 1.1 包名

- 使用小写单个单词
- 避免下划线和驼峰
- 简洁、有意义

✅ Good:
```go
package agent
package storage
```

❌ Bad:
```go
package agentManager  // 驼峰
package agent_manager // 下划线
```

### 1.2 变量名

- 局部变量: 简短、有意义
- 全局变量: 完整描述性名称
- 缩写: 常见缩写全大写 (ID, HTTP, URL)

✅ Good:
```go
var userID string
var httpClient *http.Client
```

❌ Bad:
```go
var userId string       // 应该是 ID 而非 Id
var user_id string      // 下划线
var u string            // 全局变量太简短
```

### 1.3 函数名

- 导出函数: 驼峰命名,首字母大写
- 私有函数: 驼峰命名,首字母小写
- 方法接收者: 1-2 个字符,不使用 `this` 或 `self`

✅ Good:
```go
func (s *AgentService) Register(ctx context.Context, agent *Agent) error
func validateAgentID(id string) error
```

❌ Bad:
```go
func (this *AgentService) Register()  // 不使用 this
func Validate_Agent_ID()             // 下划线
```

## 2. 代码组织

### 2.1 导入顺序

1. 标准库
2. 第三方库
3. 本项目内部包

```go
import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/kart-io/k8s-agent/internal/pkg/logger"
	"github.com/kart-io/k8s-agent/pkg/types"
)
```

### 2.2 函数顺序

1. 构造函数 (New*)
2. 公共方法
3. 私有方法
4. 辅助函数

## 3. 错误处理

### 3.1 错误检查

总是检查错误,不要忽略

✅ Good:
```go
data, err := readFile(path)
if err != nil {
	return fmt.Errorf("failed to read file: %w", err)
}
```

❌ Bad:
```go
data, _ := readFile(path)  // 忽略错误
```

### 3.2 错误包装

使用 `%w` 包装错误,保留错误链

✅ Good:
```go
if err != nil {
	return fmt.Errorf("failed to process request: %w", err)
}
```

## 4. 注释

### 4.1 包注释

每个包都应该有包注释

```go
// Package agent provides agent management functionality.
// It handles agent registration, heartbeat, and command dispatch.
package agent
```

### 4.2 函数注释

导出函数必须有注释,以函数名开头

```go
// Register registers a new agent with the given configuration.
// It returns an error if the agent already exists or if validation fails.
func Register(ctx context.Context, agent *Agent) error {
	// ...
}
```

## 5. 并发

### 5.1 Goroutine

- 谨慎使用 goroutine
- 确保有退出机制
- 避免 goroutine 泄漏

✅ Good:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
	select {
	case <-ctx.Done():
		return
	case <-time.After(interval):
		// do work
	}
}()
```

### 5.2 Channel

- 明确发送和接收方
- 关闭 channel 时确保没有并发写入

## 6. 性能

### 6.1 避免不必要的分配

使用 `strings.Builder` 而非字符串拼接

✅ Good:
```go
var b strings.Builder
for _, s := range parts {
	b.WriteString(s)
}
return b.String()
```

❌ Bad:
```go
result := ""
for _, s := range parts {
	result += s  // 每次循环都分配新字符串
}
```

### 6.2 使用指针

对于大型结构体,使用指针避免拷贝

✅ Good:
```go
func ProcessAgent(agent *Agent) error {
	// ...
}
```
```

### 4.8 CI/CD 增强

#### 4.8.1 GitHub Actions 工作流

创建 `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ master, develop ]
  pull_request:
    branches: [ master, develop ]

env:
  GO_VERSION: '1.25'
  GOLANGCI_LINT_VERSION: 'v1.55'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: ${{ env.GOLANGCI_LINT_VERSION }}
          args: --timeout=5m

      - name: Check formatting
        run: |
          make format
          git diff --exit-code

  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: aetherius_test
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3

      redis:
        image: redis:7
        ports:
          - 6379:6379

      nats:
        image: nats:2.10
        ports:
          - 4222:4222

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Download dependencies
        run: make deps

      - name: Run unit tests
        run: make test
        env:
          DB_HOST: localhost
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: root
          DB_NAME: aetherius_test
          REDIS_HOST: localhost
          REDIS_PORT: 6379
          NATS_URL: nats://localhost:4222

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          flags: unittests

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test]
    strategy:
      matrix:
        service: [agent-manager, orchestrator, reasoning, auth, collect-agent]

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build ${{ matrix.service }}
        run: make build BINS=${{ matrix.service }}

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.service }}
          path: bin/${{ matrix.service }}

  docker:
    name: Docker Build
    runs-on: ubuntu-latest
    needs: [build]
    if: github.event_name == 'push'
    strategy:
      matrix:
        service: [agent-manager, orchestrator, reasoning, auth, collect-agent]

    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: build/docker/${{ matrix.service }}.Dockerfile
          platforms: linux/amd64,linux/arm64
          push: true
          tags: |
            aetherius/${{ matrix.service }}:${{ github.sha }}
            aetherius/${{ matrix.service }}:latest
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

#### 4.8.2 发布流程

创建 `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Generate changelog
        id: changelog
        run: |
          # 使用 git-chglog 或类似工具生成 changelog
          echo "changelog<<EOF" >> $GITHUB_OUTPUT
          git log --oneline $(git describe --tags --abbrev=0 HEAD^)..HEAD >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Create Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: Release ${{ github.ref }}
          body: ${{ steps.changelog.outputs.changelog }}
          draft: false
          prerelease: false
```

---

## 5. 实施计划

### 5.1 优先级划分

#### P0 (必须,立即实施) - 1-2 周

1. **项目结构重组**: 建立标准化的目录结构
2. **Makefile 增强**: 完善构建流程和自动化
3. **错误码统一**: 建立统一的错误码体系

#### P1 (重要,近期实施) - 3-4 周

4. **代码质量规范**: golangci-lint 配置和代码规范文档
5. **日志规范**: 统一日志格式和字段
6. **API 设计规范**: RESTful API 规范和 OpenAPI 文档

#### P2 (期望,中期实施) - 5-8 周

7. **测试策略**: 完善单元测试、集成测试、E2E 测试
8. **文档体系**: 完善开发文档和用户文档
9. **CI/CD 增强**: GitHub Actions 工作流

#### P3 (可选,长期实施) - 9-12 周

10. **配置管理优化**: 多环境配置管理
11. **性能优化**: 基准测试和性能调优
12. **安全加固**: 安全扫描和漏洞修复

### 5.2 时间线

```
Week 1-2:  项目结构重组 + Makefile 增强
Week 3-4:  错误码统一 + 代码质量规范
Week 5-6:  日志规范 + API 设计规范
Week 7-8:  单元测试完善
Week 9-10: 集成测试 + E2E 测试
Week 11-12: 文档体系完善 + CI/CD 增强
```

### 5.3 团队分工

- **架构师**: 项目结构设计、技术方案审核
- **后端开发 1**: Agent Manager 重构
- **后端开发 2**: Orchestrator 重构
- **后端开发 3**: Reasoning Service 重构
- **测试工程师**: 测试策略制定和测试用例编写
- **文档工程师**: 文档体系建设

### 5.4 风险评估

#### 高风险

- **项目结构重组**: 可能导致大量代码修改,需要充分测试
  - **缓解措施**: 分步骤迁移,每个步骤都进行测试验证

#### 中风险

- **API 变更**: 可能影响现有客户端
  - **缓解措施**: 保持向后兼容,使用 API 版本管理
- **测试覆盖率**: 现有测试可能不足
  - **缓解措施**: 逐步提升测试覆盖率,设定阶段性目标

#### 低风险

- **文档完善**: 不影响功能
- **CI/CD 增强**: 可以逐步完善

---

## 6. 成功指标

### 6.1 代码质量指标

- ✅ 代码覆盖率 > 70%
- ✅ golangci-lint 通过率 100%
- ✅ 循环复杂度 < 15
- ✅ 函数长度 < 100 行

### 6.2 项目规范指标

- ✅ 所有服务遵循统一目录结构
- ✅ 所有 API 有 OpenAPI 文档
- ✅ 所有错误使用统一错误码
- ✅ 所有日志使用结构化格式

### 6.3 开发效率指标

- ✅ 构建时间 < 5 分钟
- ✅ CI 流程时间 < 10 分钟
- ✅ 新功能开发周期减少 20%

### 6.4 文档完整性指标

- ✅ 所有公共 API 有文档
- ✅ 所有服务有开发指南
- ✅ 关键决策有 ADR 记录

---

## 7. 总结

本改进方案基于 OneX 云原生平台的最佳实践,结合 Go 语言社区的标准布局,为 Aetherius 项目提供了全面的优化路径。通过系统化的改进,项目将在以下方面获得显著提升:

1. **代码质量**: 更高的代码质量和一致性
2. **可维护性**: 更清晰的项目结构和模块划分
3. **开发效率**: 更完善的工具链和自动化流程
4. **团队协作**: 统一的规范和完善的文档

实施这些改进需要团队的共同努力和持续投入,但长期来看,这些投入将带来巨大的回报,使 Aetherius 成为真正企业级的云原生运维平台。

---

## 附录

### A. 参考资源

- [OneX Cloud-Native Platform](https://github.com/onexstack/onex)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)

### B. 工具清单

- **构建工具**: Make, Go 1.25+
- **代码质量**: golangci-lint, gofumpt, goimports
- **测试工具**: testify, gomock, go-sqlmock
- **文档工具**: swag (OpenAPI), godoc
- **CI/CD**: GitHub Actions, Docker, Kubernetes
- **监控工具**: Prometheus, Grafana, Jaeger

### C. 模板文件

所有模板文件可在 `tools/templates/` 目录找到:

- `service.go.tmpl` - 服务模板
- `handler.go.tmpl` - Handler 模板
- `test.go.tmpl` - 测试模板
- `Dockerfile.tmpl` - Dockerfile 模板
- `config.yaml.tmpl` - 配置模板

---

**文档版本**: v1.0.0
**最后更新**: 2025-10-23
**维护者**: Aetherius Team
