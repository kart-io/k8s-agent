# Makefile 使用案例大全

本文档提供了项目中所有 Makefile 命令的实际使用案例和最佳实践。

## 目录

- [开发环境设置](#开发环境设置)
- [构建命令](#构建命令)
- [运行服务](#运行服务)
- [测试命令](#测试命令)
- [代码质量](#代码质量)
- [Docker 操作](#docker-操作)
- [依赖管理](#依赖管理)
- [数据库操作](#数据库操作)
- [协议生成](#协议生成)
- [部署操作](#部署操作)
- [CI/CD](#cicd)
- [实用工具](#实用工具)
- [组合命令](#组合命令)

---

## 开发环境设置

### 初次设置开发环境

```bash
# 1. 安装所有必需的开发工具（golangci-lint, protoc, etc.）
make dev-setup

# 2. 验证工具是否正确安装
make tools.verify

# 3. 安装 Git hooks（自动格式化、lint 等）
make hooks.install

# 4. 下载所有 Go 依赖
make deps

# 5. 启动依赖服务（MySQL, Redis, NATS）
make run-deps

# 6. 验证依赖服务是否运行
make run-check-deps
```

### 安装额外的开发工具

```bash
# 安装所有工具（包括可选工具如 air, mockgen）
make install-tools A=1

# 单独安装特定工具
make tools.install.air        # 热重载工具
make tools.install.mockgen    # Mock 生成工具
make tools.install.golangci-lint  # Linter
```

### 验证开发环境就绪

```bash
# 一键检查开发环境是否准备好
make dev-ready
```

---

## 构建命令

### 基础构建

```bash
# 构建所有服务（输出到 _output/bin/）
make build

# 或使用新格式（推荐）
make go.build

# 查看构建输出
ls -lh _output/bin/
```

### 构建特定服务

```bash
# 新格式（推荐 - 清晰明确）
make go.build.agent-manager
make go.build.orchestrator
make go.build.reasoning
make go.build.auth
make go.build.gateway
make go.build.monitor
make go.build.cluster
make go.build.collect-agent

# 旧格式（兼容性）
make build-agent-manager
make build-orchestrator
```

### 多平台构建

```bash
# 为所有支持平台构建（linux/amd64, linux/arm64, darwin/amd64, darwin/arm64）
make go.build.multiarch

# 查看输出
ls -lh _output/bin/
# 输出示例:
# agent-manager-linux-amd64
# agent-manager-linux-arm64
# agent-manager-darwin-amd64
# agent-manager-darwin-arm64
```

### 快速重建

```bash
# 清理并重新构建（适合解决缓存问题）
make rebuild

# 或分步执行
make clean
make build
```

### 带详细输出的构建

```bash
# 启用详细模式（显示完整命令）
make V=1 build

# 构建特定服务并显示详细信息
make V=1 go.build.agent-manager
```

---

## 运行服务

### 运行单个服务（默认配置）

```bash
# 使用默认配置运行服务
make run-auth
make run-agent-manager
make run-orchestrator
make run-reasoning
make run-gateway
make run-monitor
make run-cluster
```

### 使用特定环境配置运行服务

```bash
# Auth 服务
make run-auth-local     # 本地开发配置
make run-auth-dev       # 开发环境配置
make run-auth-test      # 测试环境配置
make run-auth-prod      # 生产环境配置

# Agent Manager 服务
make run-agent-manager-local
make run-agent-manager-dev
make run-agent-manager-test
make run-agent-manager-prod

# Orchestrator 服务
make run-orchestrator-local
make run-orchestrator-dev
make run-orchestrator-test
make run-orchestrator-prod

# Reasoning 服务
make run-reasoning-local
make run-reasoning-dev
make run-reasoning-test
make run-reasoning-prod
```

### 通用运行方式（灵活指定服务和环境）

```bash
# 语法: make run SERVICE=<服务名> ENV=<环境>
make run SERVICE=auth ENV=local
make run SERVICE=agent-manager ENV=dev
make run SERVICE=orchestrator ENV=test
make run SERVICE=reasoning ENV=prod

# 不指定 ENV 则使用默认配置
make run SERVICE=auth
```

### 热重载开发模式

```bash
# 使用 air 进行热重载（代码变更自动重启）
make dev

# 注意：需要在项目根目录有 .air.toml 配置文件
```

### 管理依赖服务

```bash
# 启动所有依赖服务（MySQL, Redis, NATS）
make run-deps

# 检查依赖服务状态
make run-check-deps

# 停止所有依赖服务
make run-stop-deps

# 单独管理服务（使用 docker-compose）
cd deployments/docker-compose
docker-compose up -d mysql     # 只启动 MySQL
docker-compose up -d redis     # 只启动 Redis
docker-compose up -d nats      # 只启动 NATS
docker-compose ps              # 查看服务状态
docker-compose logs -f mysql   # 查看 MySQL 日志
```

### 完整开发流程示例

```bash
# Terminal 1: 启动依赖服务
make run-deps

# Terminal 2: 运行 Agent Manager
make run-agent-manager-local

# Terminal 3: 运行 Orchestrator
make run-orchestrator-local

# Terminal 4: 运行 Reasoning Service
make run-reasoning-local

# Terminal 5: 运行 Auth Service
make run-auth-local
```

---

## 测试命令

### 运行所有测试

```bash
# 运行所有单元测试
make test

# 或使用新格式
make go.test
```

### 测试特定服务

```bash
# 测试特定服务
make go.test.agent-manager
make go.test.orchestrator
make go.test.reasoning
make go.test.auth
```

### 测试单个包或函数

```bash
# 测试特定包
go test -v ./internal/agent-manager/agent

# 测试特定函数
go test -v ./internal/agent-manager/agent -run TestAgentRegistry_Register

# 测试时显示详细输出
go test -v -count=1 ./internal/orchestrator/workflow/...
```

### 覆盖率测试

```bash
# 生成所有服务的覆盖率报告（HTML 格式在 _output/coverage/）
make test-coverage

# 查看特定服务的覆盖率报告
open _output/coverage/agent-manager.html
open _output/coverage/orchestrator.html

# 或在 Linux 上
xdg-open _output/coverage/agent-manager.html
```

### 集成测试

```bash
# 运行集成测试（需要数据库等外部依赖）
make test-integration

# 运行特定服务的集成测试
go test -tags=integration -v -timeout 10m ./internal/agent-manager/test/integration/...
```

### 端到端测试

```bash
# 运行 E2E 测试
make test-e2e
```

### 快速测试和完整测试

```bash
# 快速测试（仅单元测试）
make quick-test

# 完整测试套件（单元 + 集成）
make full-test
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. -benchmem ./internal/agent-manager/...

# 运行特定基准测试
go test -bench=BenchmarkAgentRegistry -benchmem ./internal/agent-manager/agent
```

### 竞态检测

```bash
# 使用竞态检测运行测试
go test -race ./...

# 测试特定包的竞态条件
go test -race -v ./internal/orchestrator/workflow/...
```

---

## 代码质量

### 代码格式化

```bash
# 标准 Go 格式化
make fmt
make go.fmt

# 使用 gofumpt 和 gci 进行增强格式化（推荐）
make fmt  # 会自动使用 gofumpt + gci

# 只格式化特定文件
gofmt -w internal/agent-manager/agent/registry.go
```

### 代码检查（Linting）

```bash
# 运行所有 linters（58 个启用的 linters）
make lint
make go.lint

# 自动修复可修复的问题
make go.lint.fix

# 运行特定 linter
golangci-lint run --disable-all --enable=errcheck ./...
golangci-lint run --disable-all --enable=staticcheck ./...
```

### 代码审查（Vet）

```bash
# 运行 go vet
make vet
make go.vet

# 检查特定包
go vet ./internal/agent-manager/...
```

### 完整代码质量检查

```bash
# 运行所有检查（lint + vet）
make check
make validate
```

### 预提交检查

```bash
# 提交前运行（格式化 + lint + 测试）
make pre-commit

# 推送前运行（lint + 测试 + 构建）
make pre-push
```

---

## Docker 操作

### 构建 Docker 镜像

```bash
# 构建所有服务的 Docker 镜像
make docker-build

# 或使用新格式
make docker.build

# 构建特定服务的镜像
make docker.build.agent-manager
make docker.build.orchestrator
make docker.build.reasoning

# 指定版本标签
make docker-build VERSION=v1.2.3
```

### 多平台 Docker 构建

```bash
# 构建多平台镜像（linux/amd64, linux/arm64）
make docker-buildx VERSION=v1.2.3

# 构建并推送多平台镜像
make docker-buildx-push VERSION=v1.2.3

# 为特定环境构建
make docker-buildx-env ENV=dev
make docker-buildx-push-env ENV=prod
```

### 推送 Docker 镜像

```bash
# 推送所有镜像
make docker-push

# 推送特定服务镜像
docker push <registry>/<namespace>/agent-manager:latest
```

### 使用 Docker Compose

```bash
# 启动所有服务
make docker-compose-up

# 或手动操作
cd deployments/docker-compose
docker-compose up -d

# 查看日志
docker-compose logs -f agent-manager
docker-compose logs -f orchestrator --tail=100

# 停止所有服务
docker-compose down

# 停止并删除数据卷（警告：会删除所有数据！）
docker-compose down -v
```

---

## 依赖管理

### 下载依赖

```bash
# 下载所有依赖
make deps
make go.mod.download
```

### 整理依赖

```bash
# 整理 go.mod（移除未使用的依赖）
make tidy
make go.mod.tidy

# 这会整理：
# - 根目录 go.mod
# - common/go.mod
# - api/proto/go.mod
```

### 验证依赖

```bash
# 验证依赖完整性
make deps-verify
make go.mod.verify
```

### 完整依赖管理流程

```bash
# 添加新依赖后
go get github.com/new/package@latest
make tidy
make deps-verify
```

---

## 数据库操作

### 启动数据库

```bash
# 启动所有数据库服务
make run-deps

# 或单独启动
cd deployments/docker-compose
docker-compose up -d mysql
docker-compose up -d redis
docker-compose up -d neo4j
```

### 连接到数据库

```bash
# 连接到 MySQL
docker-compose -f deployments/docker-compose/docker-compose.yaml exec mysql mysql -u aetherius -p
# 密码通常在 docker-compose.yaml 中

# 或从 docker-compose 目录
cd deployments/docker-compose
docker-compose exec mysql mysql -u aetherius -p

# 连接到 Redis
docker-compose -f deployments/docker-compose/docker-compose.yaml exec redis redis-cli

# 或使用 make 命令
make redis-cli
```

### 重置数据库

```bash
# 停止并删除所有数据（警告：会删除所有数据！）
cd deployments/docker-compose
docker-compose down -v

# 重新启动
docker-compose up -d mysql redis

# 等待初始化完成
sleep 10

# 验证服务状态
docker-compose ps
```

### 数据库迁移

```bash
# 运行数据库迁移（如果配置了迁移脚本）
# 通常在服务启动时自动执行
make run-agent-manager  # 会自动运行迁移
```

---

## 协议生成

### 生成 Protocol Buffer 代码

```bash
# 生成所有 protobuf 代码
make gen-proto
make proto.generate

# 输出位置：
# - Go 代码: api/proto/gen/go/
# - OpenAPI v2: api/proto/gen/openapiv2/
```

### 清理生成的代码

```bash
# 清理所有生成的 proto 代码
make proto.clean
```

### 构建 protoc 插件

```bash
# 构建 protoc 插件
make proto.build
```

### 完整 proto 开发流程

```bash
# 1. 编辑 proto 文件
vim api/proto/orchestrator/workflow.proto

# 2. 生成代码
make gen-proto

# 3. 验证生成的代码
ls -lh api/proto/gen/go/orchestrator/

# 4. 在服务中导入使用
# import workflowpb "github.com/kart-io/k8s-agent/api/proto/gen/go/orchestrator"
```

---

## 部署操作

### Kubernetes 部署

```bash
# 部署到开发环境
make deploy ENV=dev

# 部署到 staging 环境
make deploy ENV=staging

# 部署到生产环境
make deploy ENV=prod

# 验证 manifests
make manifests-validate
```

### 使用 Kustomize 部署

```bash
# 部署到特定环境
cd deployments/k8s/overlays/dev
kubectl apply -k .

# 或从根目录
kubectl apply -k deployments/k8s/overlays/dev
```

### 查看部署状态

```bash
# 查看 pods
kubectl -n aetherius get pods

# 查看服务
kubectl -n aetherius get svc

# 查看日志
kubectl -n aetherius logs -f deployment/agent-manager
kubectl -n aetherius logs -f deployment/orchestrator --tail=100
```

---

## CI/CD

### 本地 CI 流程

```bash
# 运行完整 CI 流程（deps + fmt + vet + lint + test + build）
make ci

# 这等同于：
make deps
make go.fmt
make go.vet
make go.lint
make go.test
make go.build
```

### 创建发布

```bash
# 创建新版本发布
make release VERSION=v1.2.3

# 这会执行：
# 1. 清理旧构建
# 2. 下载依赖
# 3. 运行测试
# 4. 构建所有服务
# 5. 构建 Docker 镜像
```

### Git 提交工作流

```bash
# 1. 提交前检查
make pre-commit

# 2. 提交代码
git add .
git commit -m "feat: add new feature"

# 3. 推送前检查
make pre-push

# 4. 推送代码
git push origin feature-branch
```

---

## 实用工具

### 查看项目信息

```bash
# 查看版本信息
make version

# 查看项目详细信息
make info

# 查看项目统计
make stats
```

### 查看可用命令

```bash
# 显示帮助信息
make help

# 显示所有 targets
make targets

# 列出所有 makefile
make list-mk
```

### 重命名项目

```bash
# 重命名项目模块路径（批量替换所有文件）
make rename-project OLD=github.com/old/path NEW=github.com/new/path

# 执行后需要整理依赖
make tidy
```

### 清理操作

```bash
# 清理构建产物
make clean

# 清理所有（包括依赖缓存）
make clean-all

# 清理特定类型
make go.clean      # Go 构建缓存
make proto.clean   # Proto 生成代码
```

---

## 组合命令

### 完整开发流程

```bash
# 从零开始的完整流程
make dev-setup          # 1. 设置开发环境
make run-deps           # 2. 启动依赖服务
make build              # 3. 构建所有服务
make run-agent-manager  # 4. 运行服务（在新终端）
```

### 代码变更后的流程

```bash
# 修改代码后
make fmt              # 1. 格式化代码
make lint             # 2. 检查代码质量
make test             # 3. 运行测试
make rebuild          # 4. 重新构建
```

### 发布新版本流程

```bash
# 完整发布流程
make pre-push                    # 1. 预推送检查
git tag v1.2.3                   # 2. 创建标签
make release VERSION=v1.2.3      # 3. 构建发布
make docker-buildx-push VERSION=v1.2.3  # 4. 推送镜像
git push origin v1.2.3           # 5. 推送标签
```

### 调试流程

```bash
# 遇到问题时
make clean-all        # 1. 完全清理
make deps             # 2. 重新下载依赖
make build V=1        # 3. 详细构建（查看错误）
make test -v          # 4. 详细测试
```

### 性能分析流程

```bash
# 性能分析
make build                    # 1. 构建服务
make run-agent-manager        # 2. 启动服务

# 在另一个终端
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

---

## 常见场景示例

### 场景 1: 新开发者入职

```bash
# 1. 克隆仓库
git clone <repository-url>
cd k8s-agent

# 2. 设置开发环境
make dev-setup

# 3. 启动依赖
make run-deps

# 4. 构建并运行
make build
make run-agent-manager-local
```

### 场景 2: 修复 Bug

```bash
# 1. 创建分支
git checkout -b fix/bug-123

# 2. 修改代码
vim internal/agent-manager/agent/registry.go

# 3. 测试修复
make go.test.agent-manager

# 4. 代码质量检查
make pre-commit

# 5. 提交
git add .
git commit -m "fix: resolve bug-123"

# 6. 推送前检查
make pre-push
git push origin fix/bug-123
```

### 场景 3: 添加新功能

```bash
# 1. 创建功能分支
git checkout -b feature/new-feature

# 2. 添加代码和测试
vim internal/agent-manager/agent/new_feature.go
vim internal/agent-manager/agent/new_feature_test.go

# 3. 运行测试
make go.test.agent-manager

# 4. 更新 proto（如需要）
vim api/proto/agent/agent.proto
make gen-proto

# 5. 完整检查
make ci

# 6. 提交
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature
```

### 场景 4: 生产部署

```bash
# 1. 创建发布标签
git tag v1.2.3

# 2. 构建发布
make release VERSION=v1.2.3

# 3. 构建并推送多平台镜像
make docker-buildx-push VERSION=v1.2.3

# 4. 部署到 Kubernetes
make deploy ENV=prod

# 5. 验证部署
kubectl -n aetherius get pods
kubectl -n aetherius get svc
```

### 场景 5: 紧急热修复

```bash
# 1. 从主分支创建热修复分支
git checkout main
git pull
git checkout -b hotfix/critical-bug

# 2. 快速修复
vim internal/agent-manager/agent/registry.go

# 3. 快速测试
make quick-test

# 4. 快速构建
make quick-build

# 5. 紧急发布
git add .
git commit -m "hotfix: fix critical bug"
git push origin hotfix/critical-bug

# 6. 快速部署
make docker.build.agent-manager VERSION=v1.2.4
docker push <registry>/agent-manager:v1.2.4
kubectl -n aetherius set image deployment/agent-manager agent-manager=<registry>/agent-manager:v1.2.4
```

---

## 最佳实践

### 日常开发

```bash
# 每天开始工作
make run-check-deps  # 检查依赖服务
make build           # 确保代码可构建

# 编码过程中
make dev             # 使用热重载

# 提交前
make pre-commit      # 自动检查
```

### 持续集成

```bash
# 本地模拟 CI
make ci

# 如果 CI 失败，逐步排查
make go.fmt          # 格式问题
make go.lint         # Lint 问题
make go.test         # 测试失败
make go.build        # 构建失败
```

### 性能优化

```bash
# 并行构建（加速）
make -j4 build

# 使用构建缓存
make build  # 第二次会快很多

# 清理缓存（如果有问题）
make clean
make build
```

---

## 故障排除

### 构建失败

```bash
# 问题：cannot find package
# 解决：
make clean-all
make deps
make build

# 问题：版本冲突
# 解决：
make tidy
go mod download
```

### 测试失败

```bash
# 问题：数据库连接失败
# 解决：
make run-check-deps  # 检查服务
make run-deps        # 启动服务

# 问题：端口占用
# 解决：
lsof -ti:8080 | xargs kill -9  # macOS/Linux
```

### Docker 问题

```bash
# 问题：镜像构建失败
# 解决：
docker system prune -a  # 清理所有镜像
make docker-build

# 问题：docker-compose 失败
# 解决：
cd deployments/docker-compose
docker-compose down -v
docker-compose up -d
```

---

## 参考资料

- 根 Makefile: `Makefile`
- 模块化 Make 规则: `scripts/make-rules/*.mk`
- Docker Compose: `deployments/docker-compose/docker-compose.yml`
- Kubernetes 配置: `deployments/k8s/`
- CLAUDE.md: 项目架构和开发指南

---

**提示**:
- 所有命令必须从仓库根目录运行
- 使用 `make help` 查看所有可用命令
- 使用 `make targets` 查看所有 makefile 中的 targets
- 推荐使用新格式命令（如 `go.build.X`）而不是旧格式（`build-X`）
