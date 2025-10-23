# Aetherius 项目改进快速开始

快速参考和常用命令

## 📚 核心文档

| 文档 | 用途 | 阅读时间 |
|------|------|----------|
| **IMPROVEMENT_SUMMARY.md** | 改进概述和总结 | 10 分钟 |
| **architecture/IMPROVEMENT_PLAN.md** | 完整改进方案 | 45 分钟 |
| **devel/proto-buf-guide.md** | Buf 使用指南 | 30 分钟 |
| **devel/implementation-guide.md** | 实施步骤 | 30 分钟 |

## 🚀 快速开始 (5 分钟)

### 步骤 1: 安装 Buf

```bash
cd api/proto
make install-buf
buf --version
```

### 步骤 2: 更新 Makefile

```bash
# 备份旧 Makefile
mv Makefile Makefile.old

# 使用新 Makefile
mv Makefile.new Makefile
```

### 步骤 3: 生成代码

```bash
# 更新依赖
make buf-dep-update

# Lint
make buf-lint

# 生成代码
make buf-generate
```

### 步骤 4: 验证

```bash
# 检查生成的代码
tree gen/go/ -L 3

# 编译测试
cd gen/go
go mod tidy
go build ./...
```

## 📋 常用命令

### Proto 管理

```bash
cd api/proto

# Lint proto 文件
make buf-lint

# 检查破坏性变更
make buf-breaking

# 格式化 proto 文件
make buf-format-fix

# 生成代码
make buf-generate

# 运行所有检查
make ci
```

### 项目构建

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent

# 构建所有服务
make build

# 构建特定服务
make build BINS=agent-manager

# 运行测试
make test

# 代码格式化
make fmt

# Lint
make lint

# 完整 CI 流程
make ci
```

### 开发工作流

```bash
# 1. 创建功能分支
git checkout -b feature/your-feature

# 2. 修改代码

# 3. 运行检查
make fmt
make vet
make test

# 4. 如果修改了 proto
cd api/proto
make buf-lint
make buf-generate
cd ../..

# 5. 提交
git add .
git commit -m "feat: your feature description"

# 6. 推送
git push origin feature/your-feature
```

## 🎯 实施优先级

### P0 - 立即开始 (本周)

```bash
# 1. Buf 集成
cd api/proto
make install-buf
mv Makefile Makefile.old
mv Makefile.new Makefile
make buf-generate

# 2. 验证
make ci
```

### P1 - 近期完成 (2-4 周)

- 添加通用 Proto 定义 (pagination, errors)
- 创建 Orchestrator 和 Reasoning Proto API
- 开始项目结构重组

### P2 - 中期目标 (1-2 月)

- 完成项目结构重组
- 提升测试覆盖率
- 完善 CI/CD

## 🔧 故障排查

### Buf 相关

```bash
# buf 命令找不到
make install-buf

# 依赖问题
cd api/proto
buf dep update

# Lint 错误
buf lint --config buf.yaml

# 查看详细日志
buf -v generate
```

### 构建相关

```bash
# 清理并重新构建
make clean
make build

# 更新依赖
go mod tidy

# 检查导入路径
go list -m all
```

## 📖 学习路径

### Day 1: 了解改进方案

- ✅ 阅读 IMPROVEMENT_SUMMARY.md
- ✅ 理解核心改进点
- ✅ 查看目标目录结构

### Day 2: Buf 实践

- ✅ 阅读 proto-buf-guide.md
- ✅ 安装和配置 Buf
- ✅ 生成代码并验证

### Day 3-5: 试点实施

- ✅ 阅读 implementation-guide.md
- ✅ 选择一个服务试点
- ✅ 按照第一阶段执行

### Week 2+: 全面推进

- ✅ 按照实施计划逐步推进
- ✅ 定期回顾和调整
- ✅ 记录经验和问题

## 💡 最佳实践

### Proto 文件

- ✅ 每个 RPC 独特的请求/响应消息
- ✅ 枚举第一个值是 UNSPECIFIED = 0
- ✅ 保留已删除字段的编号
- ✅ 使用 google.protobuf.Timestamp 表示时间

### Go 代码

- ✅ 遵循 Effective Go
- ✅ 使用 golangci-lint
- ✅ 保持函数简短 (< 100 行)
- ✅ 添加充分的注释

### 提交规范

```
type(scope): subject

body

footer
```

类型:
- feat: 新功能
- fix: 修复
- refactor: 重构
- docs: 文档
- test: 测试
- chore: 构建/工具

## 🆘 获取帮助

### 内部资源

- 改进方案文档: docs/architecture/IMPROVEMENT_PLAN.md
- Buf 指南: docs/devel/proto-buf-guide.md
- 实施指南: docs/devel/implementation-guide.md

### 外部资源

- [Buf 官方文档](https://buf.build/docs/)
- [Go 项目布局](https://github.com/golang-standards/project-layout)
- [gRPC-Go 文档](https://grpc.io/docs/languages/go/)

### 社区

- GitHub Issues: 报告问题
- 团队会议: 每周同步进度

---

**提示**: 遇到问题先查看 docs/devel/proto-buf-guide.md 的"故障排查"章节!

**文档版本**: v1.0.0
**最后更新**: 2025-10-23
