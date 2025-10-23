# Aetherius 项目文档

欢迎来到 Aetherius (k8s-agent) 项目文档中心

## 📖 文档导航

### 🚀 快速开始

**新手必读**:
1. [快速开始指南](QUICK_START.md) - 5 分钟快速上手
2. [改进方案总结](IMPROVEMENT_SUMMARY.md) - 10 分钟了解改进全貌

### 📐 架构文档

| 文档 | 描述 | 适合人群 |
|------|------|----------|
| [系统架构](architecture/SYSTEM_ARCHITECTURE.md) | 完整的系统架构设计 | 所有人 |
| [改进方案](architecture/IMPROVEMENT_PLAN.md) | 基于 OneX 的改进方案 | 架构师、Tech Lead |
| [ADR 目录](architecture/ADR/) | 架构决策记录 | 架构师、开发者 |

### 💻 开发文档

#### 规范和约定

| 文档 | 描述 |
|------|------|
| [Protocol Buffers 指南](devel/proto-buf-guide.md) | Buf 工具完整使用指南 |
| [代码风格](devel/conventions/coding-style.md) | Go 代码规范 |
| [提交规范](devel/conventions/commit-message.md) | Git 提交信息规范 |

#### 开发指南

| 文档 | 描述 |
|------|------|
| [实施指南](devel/implementation-guide.md) | 分阶段实施步骤 |
| [项目结构](devel/guide/project-structure.md) | 项目组织说明 |
| [构建指南](devel/guide/building.md) | 编译和构建 |
| [测试指南](devel/guide/testing.md) | 测试策略和方法 |

### 👥 用户文档

| 文档 | 描述 |
|------|------|
| [安装指南](user/guide/installation.md) | 如何安装 Aetherius |
| [配置指南](user/guide/configuration.md) | 配置说明 |
| [使用指南](user/guide/usage.md) | 基本使用方法 |

### 🔧 运维文档

| 文档 | 描述 |
|------|------|
| [Kubernetes 部署](operations/deployment/kubernetes.md) | K8s 部署指南 |
| [Docker Compose 部署](operations/deployment/docker-compose.md) | Docker Compose 部署 |
| [监控配置](operations/monitoring/metrics.md) | Prometheus + Grafana |
| [日志管理](operations/monitoring/logging.md) | 日志收集和分析 |
| [故障排查](operations/troubleshooting/common-issues.md) | 常见问题解决 |

### 📝 API 文档

| 文档 | 描述 |
|------|------|
| [Agent Manager API](api/agent-manager.md) | Agent 管理 API |
| [Orchestrator API](api/orchestrator.md) | 工作流编排 API |
| [Reasoning API](api/reasoning.md) | AI 分析 API |
| [OpenAPI Specs](../api/proto/gen/openapiv2/) | Swagger 文档 |

## 🎯 按角色查找文档

### 架构师

1. [系统架构](architecture/SYSTEM_ARCHITECTURE.md)
2. [改进方案](architecture/IMPROVEMENT_PLAN.md)
3. [ADR 记录](architecture/ADR/)

### Tech Lead

1. [改进方案总结](IMPROVEMENT_SUMMARY.md)
2. [实施指南](devel/implementation-guide.md)
3. [Proto Buf 指南](devel/proto-buf-guide.md)

### 后端开发者

1. [快速开始](QUICK_START.md)
2. [代码规范](devel/conventions/coding-style.md)
3. [Proto Buf 指南](devel/proto-buf-guide.md)
4. [开发指南](devel/guide/)

### DevOps 工程师

1. [部署指南](operations/deployment/)
2. [监控配置](operations/monitoring/)
3. [故障排查](operations/troubleshooting/)

### QA 测试工程师

1. [测试指南](devel/guide/testing.md)
2. [API 文档](api/)

## 📅 文档更新记录

### 2025-10-23 - 重大更新

**新增文档**:
- ✨ Protocol Buffers 和 Buf 管理指南
- ✨ 项目改进实施指南 (6 个阶段)
- ✨ 改进方案总结
- ✨ 快速开始指南

**新增配置**:
- 🔧 buf.gen.yaml (Buf 代码生成配置)
- 🔧 增强的 Proto Makefile

**改进内容**:
- 📝 基于 OneX 最佳实践的完整改进方案
- 📝 Buf 工具链集成方案
- 📝 标准 Go 项目布局迁移计划

## 🤝 贡献文档

### 文档规范

所有文档必须:
- ✅ 使用 Markdown 格式
- ✅ 遵循 MarkdownLint 规则
- ✅ 包含文档版本和更新日期
- ✅ 添加目录 (对于长文档)
- ✅ 提供代码示例 (如适用)

### 文档模板

参考现有文档的结构:

```markdown
# 文档标题

简要描述文档内容

## 文档版本

- **版本**: v1.0.0
- **创建日期**: YYYY-MM-DD
- **最后更新**: YYYY-MM-DD

---

## 1. 章节标题

内容...

### 1.1 子章节

内容...

---

**文档版本**: v1.0.0
**维护者**: 维护者名称
```

### 提交文档变更

```bash
# 1. 创建分支
git checkout -b docs/your-doc-update

# 2. 编辑文档

# 3. 检查 Markdown 格式
markdownlint docs/

# 4. 提交
git add docs/
git commit -m "docs: update documentation for ..."

# 5. 推送并创建 PR
git push origin docs/your-doc-update
```

## 🔍 查找文档

### 全文搜索

```bash
# 在所有文档中搜索关键字
grep -r "关键字" docs/

# 只搜索特定目录
grep -r "关键字" docs/devel/

# 忽略大小写
grep -ri "关键字" docs/
```

### 按主题浏览

- **Protocol Buffers**: docs/devel/proto-buf-guide.md
- **项目改进**: docs/architecture/IMPROVEMENT_PLAN.md, docs/IMPROVEMENT_SUMMARY.md
- **开发规范**: docs/devel/conventions/
- **部署运维**: docs/operations/

## 📧 反馈和建议

发现文档问题或有改进建议?

- 📝 提交 GitHub Issue
- 💬 在团队会议中讨论
- ✉️ 联系文档维护者

## 🌟 文档质量目标

- ✅ 所有公共 API 有文档
- ✅ 所有架构决策有 ADR
- ✅ 每个服务有部署指南
- ✅ 常见问题有故障排查文档
- ✅ 文档与代码同步更新

---

**文档中心版本**: v1.0.0
**最后更新**: 2025-10-23
**维护团队**: Aetherius Team

**下一步**: 从 [快速开始指南](QUICK_START.md) 开始你的 Aetherius 之旅!
