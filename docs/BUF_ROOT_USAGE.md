# 根目录 Buf 管理 - 快速参考

## 概述

项目使用**根目录的 buf 配置**来管理 proto 文件，无需切换到 `api/proto` 目录即可执行所有操作。

## 配置文件

### 根目录配置（推荐）
```
项目根目录/
├── buf.yaml          # Buf 配置（指向 api/proto）
├── buf.gen.yaml      # 代码生成配置
└── buf.lock          # 依赖锁定文件
```

### Proto 源文件
```
api/proto/
├── agent/v1/
├── orchestrator/v1/
├── reasoning/v1/
└── common/{health,error,pagination}/v1/
```

### 生成的代码
```
pkg/api/
├── agent/v1/
├── orchestrator/v1/
├── reasoning/v1/
└── docs/swagger/
```

## 常用命令

### 从项目根目录执行

```bash
# 生成代码
make proto.generate

# Lint 检查
make proto.lint

# 破坏性变更检查
make proto.breaking

# 格式化 proto 文件
make proto.format

# 清理生成的代码
make proto.clean

# 更新依赖
make proto.dep.update

# 显示配置信息
make proto.info
```

### 直接使用 buf 命令

```bash
# 从根目录执行（推荐）
buf generate
buf lint
buf format -w
buf breaking --against '.git#branch=master'

# 或者指定目录
buf generate --template buf.gen.yaml
buf lint api/proto
```

## 工作流程

### 1. 修改 Proto 文件

```bash
# 编辑 proto 文件
vim api/proto/agent/v1/agent.proto
```

### 2. 生成代码

```bash
# 从项目根目录
make proto.generate

# 或直接使用 buf
buf generate
```

### 3. 验证代码

```bash
# 检查生成的文件
ls -lh pkg/api/agent/v1/

# 编译验证
go build ./pkg/api/...
```

### 4. Lint 和格式化

```bash
# Lint 检查
make proto.lint

# 自动格式化
make proto.format
```

## 配置说明

### buf.yaml (根目录)

```yaml
version: v2

# 指向 proto 源文件目录
modules:
  - path: api/proto
    name: buf.build/kart-io/k8s-agent

# 依赖管理
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc-ecosystem/grpc-gateway

# Lint 规则
lint:
  use:
    - STANDARD
```

### buf.gen.yaml (根目录)

```yaml
version: v2

# 管理模式配置
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/kart-io/k8s-agent/pkg/api

# 代码生成插件
plugins:
  - remote: buf.build/protocolbuffers/go
    out: pkg/api
  - remote: buf.build/grpc/go
    out: pkg/api
  - remote: buf.build/grpc-ecosystem/gateway
    out: pkg/api
  - remote: buf.build/grpc-ecosystem/openapiv2
    out: pkg/api/docs/swagger

# 输入源
inputs:
  - directory: api/proto
```

## 与旧方式的对比

### 旧方式（api/proto 目录）
```bash
cd api/proto
buf generate
cd ../..
```

### 新方式（根目录）✅
```bash
# 直接在根目录执行
make proto.generate

# 或
buf generate
```

## 优势

1. **统一入口**: 所有操作从根目录执行
2. **简化流程**: 不需要切换目录
3. **CI/CD 友好**: 脚本更简单
4. **与 Makefile 集成**: 使用统一的构建系统

## 故障排查

### 问题 1: "module not found"

**原因**: buf.yaml 中的 path 配置错误

**解决**:
```yaml
# 确保 path 指向正确的目录
modules:
  - path: api/proto  # 正确
    # NOT: path: pkg/api/proto
```

### 问题 2: "output directory does not exist"

**原因**: 输出目录未创建

**解决**:
```bash
mkdir -p pkg/api/docs/swagger
make proto.generate
```

### 问题 3: 生成的代码路径错误

**原因**: buf.gen.yaml 中的 out 配置错误

**解决**:
```yaml
plugins:
  - remote: buf.build/protocolbuffers/go
    out: pkg/api  # 从根目录的相对路径
```

## 最佳实践

1. **始终从根目录执行**: 使用 `make proto.generate` 而不是切换到子目录
2. **提交生成的代码**: 确保 `pkg/api/` 下的生成代码被 git 追踪
3. **定期运行 lint**: `make proto.lint` 确保代码质量
4. **检查 breaking change**: 合并前运行 `make proto.breaking`
5. **格式化代码**: 提交前运行 `make proto.format`

## CI/CD 集成

```yaml
# .github/workflows/proto.yml
name: Proto

on: [push, pull_request]

jobs:
  proto:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Buf
        uses: bufbuild/buf-setup-action@v1

      - name: Lint Proto
        run: make proto.lint

      - name: Check Breaking Changes
        if: github.event_name == 'pull_request'
        run: make proto.breaking

      - name: Generate Code
        run: make proto.generate

      - name: Verify Generated Code
        run: go build ./pkg/api/...
```

## 参考文档

- [完整文档](api/proto/README.md)
- [实施总结](docs/PROTO_IMPLEMENTATION.md)
- [示例代码](examples/README.md)
- [Buf 官方文档](https://docs.buf.build/)
