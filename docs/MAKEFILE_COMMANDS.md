# Makefile 命令指南

本文档说明项目中新旧 Makefile 命令格式的使用。

## 命令格式说明

项目采用 **OneX 风格的模块化 Makefile 系统**，有两套命令格式：

### 1. 新格式（推荐）✅

**格式**: `make <module>.<action>[.<service>]`

**特点**:
- 模块化清晰，易于理解命令归属
- 与 `scripts/make-rules/*.mk` 文件结构对应
- 支持命令补全和文档生成

**示例**:
```bash
# 构建
make go.build                    # 构建所有服务
make go.build.agent-manager      # 构建特定服务

# 测试
make go.test                     # 运行所有测试
make go.test.orchestrator        # 测试特定服务

# Docker
make docker.build                # 构建所有镜像
make docker.build.reasoning      # 构建特定镜像

# Proto
make proto.generate              # 生成 protobuf 代码
make proto.lint                  # 检查 proto 文件
```

### 2. 旧格式（兼容）

**格式**: `make <action>[-<service>]`

**特点**:
- 简短，易于输入
- 向后兼容旧脚本和习惯
- 本质上是新格式的别名

**示例**:
```bash
# 构建
make build                       # 等同于 make go.build
make build-agent-manager         # 等同于 make go.build.agent-manager

# 测试
make test                        # 等同于 make go.test
make test-coverage               # 等同于 make go.test.coverage

# Docker
make docker                      # 等同于 make docker.build
make docker-orchestrator         # 等同于 make docker.build.orchestrator
```

## 命令对照表

| 旧格式 (Legacy) | 新格式 (Recommended) | 说明 |
|----------------|---------------------|------|
| `make build` | `make go.build` | 构建所有服务 |
| `make build-<service>` | `make go.build.<service>` | 构建特定服务 |
| `make test` | `make go.test` | 运行所有测试 |
| `make test-coverage` | `make go.test.coverage` | 测试覆盖率 |
| `make fmt` | `make go.fmt` | 格式化代码 |
| `make lint` | `make go.lint` | 代码检查 |
| `make vet` | `make go.vet` | 静态分析 |
| `make docker` | `make docker.build` | 构建 Docker 镜像 |
| `make docker-<service>` | `make docker.build.<service>` | 构建特定镜像 |
| `make gen-proto` | `make proto.generate` | 生成 protobuf |

## 完整命令列表

### Go 构建与测试

```bash
# 构建
make go.build                    # 构建所有服务
make go.build.<service>          # 构建特定服务
make go.build.multiarch          # 多架构构建

# 测试
make go.test                     # 单元测试
make go.test.<service>           # 测试特定服务
make go.test.coverage            # 测试覆盖率
make go.test.integration         # 集成测试

# 代码质量
make go.fmt                      # 格式化
make go.lint                     # Lint 检查
make go.vet                      # 静态分析

# 依赖管理
make go.mod.download             # 下载依赖
make go.mod.tidy                 # 整理依赖
make go.mod.verify               # 验证依赖
```

### Docker 镜像

```bash
# 构建
make docker.build                # 构建所有镜像
make docker.build.<service>      # 构建特定镜像

# 推送
make docker.push                 # 推送所有镜像
make docker.push.<service>       # 推送特定镜像

# 多架构
make image.build.multiarch.<service>     # 多架构构建
make image.push.multiarch.<service>      # 推送多架构镜像
```

### Protocol Buffers

```bash
make proto.generate              # 生成代码
make proto.lint                  # 检查 proto 文件
make proto.format                # 格式化 proto 文件
make proto.breaking              # 检查破坏性变更
make proto.clean                 # 清理生成的代码
```

### 开发工具

```bash
make tools.install               # 安装所有工具
make tools.install.<tool>        # 安装特定工具
make tools.verify                # 验证工具安装

# 示例
make tools.install.golangci-lint
make tools.install.buf
make tools.install.air
```

### Git Hooks

```bash
make hooks.install               # 安装 Git hooks
make hooks.uninstall             # 卸载 Git hooks
```

### 版本管理

```bash
make version                     # 显示版本信息
make release VERSION=v1.0.0      # 创建发布版本
```

## 推荐实践

### ✅ 推荐

```bash
# 1. 使用新格式命令（清晰明确）
make go.build.agent-manager
make go.test.coverage
make docker.build.orchestrator

# 2. 在 CI/CD 脚本中使用新格式
# .github/workflows/build.yml
- run: make go.build
- run: make go.test.coverage
- run: make docker.build
```

### ⚠️ 可接受（向后兼容）

```bash
# 在个人开发中使用旧格式（更短）
make build
make test
make docker
```

### ❌ 不推荐

```bash
# 不要混用格式（容易混淆）
make go.build.agent-manager
make build-orchestrator  # 应该统一使用新格式

# 不要直接调用 scripts/make-rules/*.mk 中的目标
# 应通过根 Makefile 调用
```

## 为什么保留两套格式？

### 设计理念

1. **渐进式迁移**: 允许用户逐步从旧格式迁移到新格式
2. **向后兼容**: 不破坏现有脚本和文档
3. **用户友好**: 新手可用简短命令，专家可用明确命令
4. **OneX 兼容**: 遵循 OneX 项目的模块化设计

### 何时会移除旧格式？

- 短期（6-12 个月）: 保留所有兼容性别名
- 中期（1-2 年）: 在文档中标记为 "Deprecated"
- 长期（2+ 年）: 根据社区反馈决定是否移除

## 查看所有可用命令

```bash
# 查看所有命令和说明
make help

# 查看所有模块的目标
make targets

# 列出所有包含的 makefile
make list-mk

# 查看项目统计信息
make stats
```

## 故障排除

### 命令未找到

```bash
# 问题
make: *** No rule to make target 'go.build.myservice'

# 解决
# 1. 检查服务名是否正确
ls cmd/  # 查看可用服务

# 2. 确保从项目根目录运行
pwd  # 应该是 /path/to/k8s-agent

# 3. 查看可用命令
make help
```

### 命令执行失败

```bash
# 查看详细输出
V=1 make go.build.agent-manager

# 查看 Makefile 变量
make info
```

## 参考资料

- [Makefile](../Makefile) - 根 Makefile
- [scripts/make-rules/](../scripts/make-rules/) - 模块化规则定义
- [OneX Project](https://github.com/superproj/onex) - 参考项目

---

**建议**: 在新代码和文档中统一使用新格式（`go.build.*`），以保持一致性和可维护性。
