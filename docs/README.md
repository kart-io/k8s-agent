# Aetherius 项目文档

欢迎来到 Aetherius (k8s-agent) 项目文档中心

---

## 📖 文档导航

### 🚀 快速开始

**新手必读**:
1. [快速开始指南](QUICK_START.md) - 5分钟快速上手
2. [gRPC-HTTP统一Handler](grpc-http-unification/README.md) - **最新功能**

**开发参考**:
- [Makefile命令参考](MAKEFILE_COMMANDS.md) - 所有make命令说明
- [端口使用说明](PORT_USAGE.md) - 服务端口分配
- [故障排查指南](TROUBLESHOOTING.md) - 常见问题解决

---

## 📐 架构文档

### 系统架构
- [系统架构](architecture/SYSTEM_ARCHITECTURE.md) - 完整的4层架构设计
- [架构决策记录](decisions/) - ADR文档

### 代码组织
- [代码重组说明](CODE_REORGANIZATION.md) - common/ vs pkg/ vs internal/
- [代码标准化](CODE_STANDARDIZATION.md) - 统一代码规范

### 最佳实践
- [最佳实践总结](BEST_PRACTICE_SUMMARY.md) - 项目最佳实践
- [服务标准模式](SERVICE_STANDARD_PATTERN.md) - 服务入口模式
- [中间件系统](MIDDLEWARE_SYSTEM.md) - HTTP中间件使用

---

## 🎯 最新功能

### gRPC-HTTP统一Handler (2025-11-01)

**核心特性**:
- ✅ 双协议支持 - HTTP/JSON + gRPC/Protobuf
- ✅ 零代码重复 - 一个实现，两种协议
- ✅ 自动转换 - gRPC-Gateway处理

**服务端口**:
- Orchestrator: HTTP :8092 / gRPC :9092
- Reasoning: HTTP :8082 / gRPC :9093

**文档**:
- [完整报告](grpc-http-unification/README.md) - **从这里开始！**
- [实现总结](grpc-http-unification/grpc_http_unified_handler_summary.md)
- [API使用示例](grpc-http-unification/api_usage_examples.md)
- [架构设计](grpc-http-unification/architecture_design.md)
- [快速启动](grpc-http-unification/quickstart_guide.md)
- [交付清单](grpc-http-unification/delivery_checklist.md)

---

## 📚 服务文档

### 核心服务

| 服务 | 文档位置 | 说明 |
|------|---------|------|
| Agent Manager | - | 中央控制层，管理所有采集代理 |
| Orchestrator | [grpc-http-unification](grpc-http-unification/) | 工作流编排服务（支持gRPC+HTTP） |
| Reasoning | [grpc-http-unification](grpc-http-unification/) | AI智能分析服务（支持gRPC+HTTP） |
| Auth | - | 认证授权服务（Bootstrap模式） |
| Cluster | - | 集群管理服务（Bootstrap模式） |
| Collect Agent | - | 边缘采集代理（Simple模式） |
| Gateway | - | API网关（Simple模式） |
| Monitor | - | 监控服务（Simple模式） |

### gRPC服务
- [gRPC文档](grpc/) - gRPC服务说明

---

## 💻 开发文档

### 开发指南
- [开发指南](devel/) - 完整开发文档
- [API参考](api/) - API文档
- [技术规范](specs/) - 技术规范文档

### API文档
- [API快速参考](API_QUICK_REFERENCE.md) - API速查表
- [gRPC-HTTP API](grpc-http-unification/api_usage_examples.md) - HTTP/gRPC API示例

### 参考资料
- [OneX学习总结](ONEX_LEARNINGS.md) - OneX项目学习笔记
- [OneX实施指南](ONEX_IMPLEMENTATION_GUIDE.md) - **OneX功能补充详细指南**
- [OneX代码示例](ONEX_CODE_EXAMPLES.md) - **OneX实战代码示例**
- [OneX实施总结](ONEX_IMPLEMENTATION_SUMMARY.md) - **Phase 1实施完成报告** ✨
- [幂等性集成报告](IDEMPOTENCY_INTEGRATION_REPORT.md) - **Agent Manager幂等性中间件集成** 🆕
- [幂等性测试指南](IDEMPOTENCY_TESTING_GUIDE.md) - **测试和验证指南** 🆕
- [最终完成报告](FINAL_COMPLETION_REPORT.md) - **OneX实施与幂等性集成最终报告** 🎉
- [项目需求](REQUIREMENTS.md) - 功能需求文档

---

## 📋 文档分类

### 按用户角色

**新手开发者**:
1. QUICK_START.md
2. grpc-http-unification/README.md
3. grpc-http-unification/quickstart_guide.md
4. TROUBLESHOOTING.md

**高级开发者**:
1. CODE_REORGANIZATION.md
2. CODE_STANDARDIZATION.md
3. SERVICE_STANDARD_PATTERN.md
4. grpc-http-unification/architecture_design.md

**架构师**:
1. architecture/SYSTEM_ARCHITECTURE.md
2. grpc-http-unification/architecture_design.md
3. BEST_PRACTICE_SUMMARY.md
4. decisions/

**API使用者**:
1. API_QUICK_REFERENCE.md
2. grpc-http-unification/api_usage_examples.md
3. grpc/

---

## 📊 文档统计

- **核心文档**: 18个MD文件
- **文档目录**: 6个
- **总文档数**: 41个MD文件
- **总大小**: ~1.1MB

---

## 🔍 快速查找

### 我想...

**快速上手**:
→ [QUICK_START.md](QUICK_START.md)

**使用gRPC或HTTP API**:
→ [grpc-http-unification/api_usage_examples.md](grpc-http-unification/api_usage_examples.md)

**了解系统架构**:
→ [architecture/SYSTEM_ARCHITECTURE.md](architecture/SYSTEM_ARCHITECTURE.md)

**查看Makefile命令**:
→ [MAKEFILE_COMMANDS.md](MAKEFILE_COMMANDS.md)

**解决问题**:
→ [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

**了解最新功能**:
→ [grpc-http-unification/README.md](grpc-http-unification/README.md)

---

## 📝 文档贡献

### 文档规范
- 使用Markdown格式
- 每个文档顶部包含目录
- 代码示例要可运行
- 保持文档更新

### 文档位置
```
docs/
├── *.md                      # 核心文档（根目录14个）
├── architecture/             # 架构设计文档
├── api/                      # API文档
├── devel/                    # 开发指南
├── specs/                    # 规范文档
├── decisions/                # 架构决策记录 (ADR)
├── grpc-http-unification/   # gRPC-HTTP统一handler文档
└── grpc/                     # gRPC文档
```

---

## 🎉 最近更新

### 2025-11-01
- ✅ 完成gRPC-HTTP统一Handler实现
- ✅ 新增6份gRPC-HTTP文档（grpc-http-unification/）
- ✅ 清理25个过时文档（refactoring/、优化报告等）
- ✅ 重组文档结构（从62个减少到37个）
- ✅ 文档大小优化（从1.2MB减少到800KB）

### 历史更新
- 2025-10-31: 完成服务标准化
- 2025-10-30: 完成代码重组
- 2025-10-24: 完成初始化器重构

---

**最后更新**: 2025-11-01
**文档版本**: v2.1
**维护者**: Aetherius Team
