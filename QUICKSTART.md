# 🚀 Aetherius (k8s-agent) 快速启动指南

基于 OneX v2 最佳实践构建的企业级 Kubernetes 智能运维平台

## ⚡ 快速开始 (5 分钟)

### 1. 环境要求

- Go 1.21+
- Docker 20.10+
- Make 4.0+
- kubectl 1.23+ (可选)

### 2. 一键安装

```bash
# 克隆项目
git clone https://github.com/kart-io/k8s-agent.git
cd k8s-agent

# 自动化安装（推荐）
./scripts/install.sh --type development

# 或者手动安装
./scripts/env-setup.sh
make tools.install
```

### 3. 快速启动

```bash
# 启动所有依赖服务
cd deployments/docker-compose
docker-compose up -d mysql redis nats neo4j

# 返回项目根目录
cd ../..

# 启动开发服务器（热重载）
make dev

# 或启动特定服务
cd agent-manager && make run-dev
```

### 4. 验证安装

```bash
# 检查工具链
make tools.verify

# 查看项目信息
make info

# 查看项目统计
make stats

# 运行测试
make go.test
```

## 📊 常用命令

### 开发工作流

```bash
# 代码格式化
make go.fmt

# 代码检查
make go.lint
make go.vet

# 运行测试
make go.test                 # 单元测试
make go.test.coverage        # 带覆盖率
make go.test.integration     # 集成测试

# 热重载开发
make dev                     # 所有服务
cd <service> && make dev     # 单个服务
```

### 构建和部署

```bash
# 构建二进制
make go.build                # 所有服务
make go.build.agent-manager  # 单个服务

# Docker 构建
make docker.build            # 本地构建
make docker.buildx           # 多平台构建
make docker.buildx.push      # 构建并推送

# Kubernetes 部署
make k8s.lint                # Lint manifests
make k8s.validate            # 验证 manifests
make k8s.apply               # 部署到集群
make k8s.status              # 查看状态
```

### 代码生成

```bash
# Protocol Buffer
make proto.generate          # 生成 Proto 代码
make proto.lint              # Lint Proto 文件

# Mocks 和文档
make gen.mocks               # 生成测试 mocks
make gen.docs                # 生成文档
make gen.swagger             # 生成 Swagger 文档
```

### 数据库管理

```bash
make db.create               # 创建数据库
make db.migrate              # 运行迁移
make db.seed                 # 填充测试数据
make db.backup               # 备份数据库
make db.reset                # 重置数据库
```

### 质量和安全

```bash
# 代码质量
make quality.check           # 运行所有质量检查
make quality.report          # 生成质量报告

# 安全扫描
make security.scan           # 运行所有安全扫描
make security.gosec          # Go 安全扫描
make security.trivy          # Docker 镜像扫描

# 性能分析
make perf.benchmark          # 运行基准测试
make perf.profile            # 性能分析
```

### 版本管理

```bash
make version.info            # 查看版本信息
make version.bump.patch      # 升级 patch 版本
make version.bump.minor      # 升级 minor 版本
make version.bump.major      # 升级 major 版本
make version.changelog       # 生成 changelog
make version.release TYPE=minor  # 完整发布流程
```

### CI/CD

```bash
# 运行完整 CI 流水线
./scripts/ci-helper.sh full

# 运行单个步骤
./scripts/ci-helper.sh setup
./scripts/ci-helper.sh lint
./scripts/ci-helper.sh test
./scripts/ci-helper.sh build

# 部署到环境
./scripts/ci-helper.sh deploy staging
./scripts/ci-helper.sh deploy production
```

## 📂 项目结构

```
k8s-agent/
├── api/proto/               # Protocol Buffer 定义
├── build/                   # 构建产物
├── cmd/                     # 服务入口点
│   ├── agent-manager/
│   ├── orchestrator/
│   ├── reasoning/
│   └── ...
├── configs/                 # 配置文件
├── deployments/             # 部署清单
│   ├── docker-compose/
│   ├── k8s/
│   └── kustomize/
├── docs/                    # 文档
├── internal/                # 内部包
├── pkg/                     # 公共包
├── scripts/                 # 脚本和工具
│   ├── make-rules/          # Make 规则
│   └── lib/                 # 脚本库
└── test/                    # 测试工具
```

## 🎯 核心服务

| 服务 | 端口 | 说明 |
|------|------|------|
| agent-manager | 8080 | 中央控制层 - Agent 管理 |
| orchestrator-service | 8081 | 任务编排层 - 工作流引擎 |
| reasoning-service-go | 8082 | AI 智能层 - 根因分析 |
| auth-service | 8083 | 认证服务 - JWT 认证 |
| gateway-service | 8084 | API 网关 - Traefik 集成 |

## 🔧 配置管理

### 环境变量

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置
vim .env

# 关键配置
ENVIRONMENT=development
AGENT_MANAGER_PORT=8080
ORCHESTRATOR_PORT=8081
REASONING_PORT=8082
```

### 配置文件

```bash
# 服务配置（按环境）
configs/<service>/
├── config.yaml          # 默认配置
├── config-dev.yaml      # 开发环境
├── config-test.yaml     # 测试环境
├── config-staging.yaml  # 预发布环境
└── config-prod.yaml     # 生产环境
```

## 🐛 常见问题

### 工具未安装

```bash
# 安装所有开发工具
make tools.install

# 验证工具链
make tools.verify
```

### 依赖问题

```bash
# 更新依赖
make go.mod.tidy

# 检查过时依赖
make deps.check

# 更新所有依赖
make deps.update
```

### 构建错误

```bash
# 清理并重新构建
make clean
make go.build

# 清理所有缓存
make clean.all
make clean.cache
```

### 测试失败

```bash
# 运行特定服务测试
make go.test.agent-manager

# 查看详细输出
cd <service> && go test -v ./...

# 带覆盖率
make go.test.coverage
```

## 📚 更多资源

- [完整文档](docs/)
- [API 文档](docs/api/)
- [架构设计](docs/architecture/)
- [开发指南](DEVELOPMENT.md)
- [贡献指南](CONTRIBUTING.md)
- [安全策略](SECURITY.md)

## 🆘 获取帮助

```bash
# 查看所有可用命令
make help

# 查看项目信息
make info

# 查看项目统计
make stats

# 查看版本信息
make version.info
```

### 在线支持

- **Issues**: https://github.com/kart-io/k8s-agent/issues
- **Discussions**: https://github.com/kart-io/k8s-agent/discussions
- **Email**: dev@kart.io
- **Security**: security@kart.io

## ⭐ 特性亮点

✅ **100% OneX v2 对齐** - 完全遵循最佳实践  
✅ **122 个 Make Targets** - 全面的自动化工具链  
✅ **58 个 Linters** - 企业级代码质量  
✅ **热重载开发** - 极致开发体验  
✅ **自动化 CI/CD** - 完整流水线  
✅ **Kubernetes 就绪** - 生产级部署  
✅ **安全扫描** - 内置安全检查  
✅ **性能优化** - 基准测试和分析

---

**Made with ❤️ following OneX v2 patterns**
