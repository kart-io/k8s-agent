# Aetherius AI Agent 文档索引图

## 文档结构概览

```
Aetherius AI Agent 文档体系
│
├── 📚 核心文档 (docs/)
│   ├── README.md ························ 文档中心与导航
│   ├── REQUIREMENTS.md ··················· 需求规格索引
│   └── ai_agent.md ······················· 完整系统需求规格说明书 (5073 行)
│
├── 📖 规格文档 (docs/specs/)
│   ├── 00_overview.md ···················· 系统概览与快速开始
│   ├── 02_architecture.md ················ 架构设计文档
│   ├── 03_data_models.md ················· 数据模型定义
│   ├── 04_deployment.md ·················· 部署配置指南
│   └── 05_operations.md ·················· 运维安全文档
│
└── 🎯 项目文档 (根目录)
    ├── CLAUDE.md ························· Claude Code 开发指南
    └── go.mod ···························· Go 模块定义
```

## 文档导航矩阵

### 按角色导航

| 角色 | 主要文档 | 辅助文档 |
|------|----------|----------|
| **产品经理 / 项目经理** | [需求规格索引](../REQUIREMENTS.md) | [系统概览](./00_overview.md) |
| **架构师** | [架构设计](./02_architecture.md) | [数据模型](./03_data_models.md) |
| **开发工程师** | [数据模型](./03_data_models.md) <br> [架构设计](./02_architecture.md) | [完整需求文档](../ai_agent.md) |
| **运维工程师 (SRE)** | [运维安全](./05_operations.md) <br> [部署配置](./04_deployment.md) | [架构设计](./02_architecture.md) |
| **测试工程师** | [数据模型](./03_data_models.md) | [需求规格](../REQUIREMENTS.md) |
| **安全工程师** | [运维安全](./05_operations.md) | [架构设计](./02_architecture.md) |

### 按任务导航

#### 🎯 快速开始

```
新用户入门路径:
1. README.md ············· 了解文档体系
2. 00_overview.md ········ 理解系统概览
3. REQUIREMENTS.md ······· 浏览功能需求
4. 02_architecture.md ···· 理解技术架构
```

#### 📋 需求分析

```
需求分析路径:
1. REQUIREMENTS.md ······· 需求总索引
2. ai_agent.md ··········· 详细需求说明
   ├─ 第 3 章: 功能需求
   ├─ 第 4 章: 非功能需求
   └─ 第 11 章: 迭代路线图
```

#### 🏗️ 系统设计

```
系统设计路径:
1. 02_architecture.md ···· 架构设计
   ├─ 系统架构
   ├─ 核心组件
   ├─ 多集群架构
   ├─ 数据流
   └─ 安全架构
2. 03_data_models.md ····· 数据模型
   ├─ 9 个核心数据模型
   ├─ Go 结构体定义
   ├─ 数据库 Schema
   └─ API 接口定义
```

#### 💻 开发实现

```
开发实现路径:
1. CLAUDE.md ············· 开发环境配置
2. 03_data_models.md ····· 数据结构参考
3. 02_architecture.md ···· 组件接口定义
4. ai_agent.md ··········· 业务逻辑详情
   ├─ 第 5 章: 技术实现细节
   └─ 第 6 章: 数据模型定义
```

#### 🚀 部署运维

```
部署运维路径:
1. 04_deployment.md ······ 部署配置
   ├─ 前置条件检查
   ├─ 基础设施部署
   ├─ 核心服务部署
   ├─ 多环境配置
   └─ 部署验证
2. 05_operations.md ······ 运维管理
   ├─ 监控告警
   ├─ 日志管理
   ├─ 故障排查
   └─ 安全加固
```

#### 🔒 安全合规

```
安全合规路径:
1. 05_operations.md ······ 安全文档
   ├─ 第 2 章: 安全架构
   ├─ 第 3 章: 威胁建模
   ├─ 第 4 章: 监控可观测
   └─ 第 6 章: 灾难恢复
2. 02_architecture.md ···· 安全设计
   └─ 第 6 章: 安全架构
```

## 文档关系图谱

```
                    ┌──────────────────────┐
                    │   README.md          │
                    │  (文档中心)           │
                    └──────────┬───────────┘
                              │
                 ┌────────────┼────────────┐
                 │            │            │
          ┌──────▼─────┐ ┌───▼──────┐ ┌──▼─────────┐
          │00_overview │ │REQUIRE-  │ │ ai_agent   │
          │    .md     │ │MENTS.md  │ │   .md      │
          │ (系统概览) │ │(需求索引)│ │(完整SRS)   │
          └──────┬─────┘ └───┬──────┘ └──┬─────────┘
                 │           │            │
        ┌────────┴───────────┴────────────┴────────┐
        │                                           │
  ┌─────▼──────────┐                    ┌──────────▼─────┐
  │02_architecture │                    │03_data_models  │
  │      .md       │◄──────引用──────────┤      .md       │
  │  (架构设计)    │                    │  (数据模型)    │
  └────────┬───────┘                    └────────┬───────┘
           │                                     │
           │         ┌──────────────┐            │
           └────────►│04_deployment │◄───────────┘
                     │      .md     │
                     │  (部署配置)  │
                     └──────┬───────┘
                            │
                   ┌────────▼────────┐
                   │05_operations.md │
                   │   (运维安全)    │
                   └─────────────────┘
```

## 主要章节对照表

### ai_agent.md 章节分布

| 章节 | 标题 | 行数范围 | 主题文档 |
|------|------|----------|----------|
| 1 | 引言 | 1-50 | [00_overview.md](./00_overview.md) |
| 2 | 系统概览 | 51-150 | [00_overview.md](./00_overview.md) |
| 3 | 功能需求 | 151-1200 | [REQUIREMENTS.md](../REQUIREMENTS.md) |
| 4 | 非功能需求 | 1201-1600 | [REQUIREMENTS.md](../REQUIREMENTS.md) |
| 5 | 技术实现细节 | 1601-2400 | [02_architecture.md](./02_architecture.md) |
| 6 | 数据模型 | 2401-2470 | [03_data_models.md](./03_data_models.md) |
| 7 | 部署与配置 | 2471-3200 | [04_deployment.md](./04_deployment.md) |
| 8 | 可观测性与监控 | 3201-3900 | [05_operations.md](./05_operations.md) (第4章) |
| 9 | 测试策略 | 3901-4200 | [02_architecture.md](./02_architecture.md) (第8章) |
| 10 | 安全考量 | 4201-4670 | [05_operations.md](./05_operations.md) (第2-3章) |
| 11 | 迭代路线图 | 4671-4680 | [00_overview.md](./00_overview.md) |
| 12 | 常见问题解答 | 4681-5073 | 各文档附录 |

### 专题文档内容对照

#### 02_architecture.md 架构设计文档

```
第 1 章: 概述
第 2 章: 系统架构
   2.1 架构概览
   2.2 分层架构
   2.3 服务职责划分
第 3 章: 核心组件设计
   3.1 事件网关 (Event Gateway)
   3.2 编排器 (Orchestrator)
   3.3 推理服务 (Reasoning Service)
   3.4 执行服务 (Execution Service)
   3.5 报告服务 (Report Service)
第 4 章: 多集群架构
   4.1 管理架构
   4.2 动态凭证管理
第 5 章: 数据流与交互
   5.1 端到端数据流
   5.2 服务间交互
第 6 章: 安全架构
   6.1 Only-Read 权限控制
   6.2 凭证管理
第 7 章: 关键设计决策
   7.1 技术选型
   7.2 架构权衡
```

#### 03_data_models.md 数据模型文档

```
第 1 章: 概述
第 2 章: 核心数据模型
   2.1 DiagnosticTask (诊断任务)
   2.2 DiagnosticStep (诊断步骤)
   2.3 Tool (工具定义)
   2.4 KnowledgeBase (知识库)
   2.5 DiagnosticReport (诊断报告)
   2.6 UserFeedback (用户反馈)
   2.7 ResourceUsage (资源使用)
   2.8 SystemConfig (系统配置)
   2.9 HistoryRecord (历史记录)
第 3 章: 数据模型关系
第 4 章: Go 结构体定义
第 5 章: 数据库 Schema
第 6 章: API 接口定义
```

#### 04_deployment.md 部署配置文档

```
第 1 章: 概述
第 2 章: 前置条件
   2.1 环境要求
   2.2 前置条件检查脚本
第 3 章: 基础设施部署
   3.1 创建基础资源
   3.2 部署 PostgreSQL
   3.3 部署 Redis
   3.4 部署 Vault
   3.5 部署向量数据库
第 4 章: 核心服务部署
   4.1 配置 ConfigMap
   4.2 创建应用密钥
   4.3 部署各微服务
第 5 章: Ingress 配置
第 6 章: 部署验证
第 7 章: 多环境配置
第 8 章: Helm Chart 部署
第 9 章: 配置管理
第 10 章: 故障排查
第 11 章: 升级和回滚
第 12 章: 卸载和清理
```

#### 05_operations.md 运维安全文档

```
第 1 章: 概述
第 2 章: 安全架构
   2.1 身份认证
   2.2 授权与访问控制
   2.3 数据保护与加密
   2.4 输入验证与防护
   2.5 审计与合规
第 3 章: 威胁建模与风险评估
   3.1 威胁识别矩阵
   3.2 安全扫描集成
第 4 章: 监控与可观测性
   4.1 监控指标体系
   4.2 分布式追踪
   4.3 日志管理
第 5 章: 成本管理
   5.1 AI 服务成本控制
第 6 章: 灾难恢复
   6.1 备份策略
   6.2 故障转移机制
第 7 章: 故障排查指南
   7.1 常见问题诊断
   7.2 性能问题排查
第 8 章: 运维最佳实践
   8.1 日常维护清单
   8.2 安全加固建议
```

## 文档更新记录

| 日期 | 文档 | 版本 | 更新内容 |
|------|------|------|----------|
| 2025-09-27 | ai_agent.md | v1.6 | 完善多集群管理和成本控制机制 |
| 2025-09-28 | 各专题文档 | v1.0 | 从ai_agent.md拆分专题文档 |
| 2025-09-30 | 06_microservices.md | v1.1 | 修正Event Gateway职责说明 |
| 2025-09-30 | 07_k8s_event_watcher.md | v1.1 | 增强过滤器逻辑说明 |
| 2025-09-30 | 08_in_cluster_deployment.md | v1.1 | 扩展权限安全说明 |
| 2025-09-30 | 09_agent_proxy_mode.md | v1.1 | 修正NATS架构说明 |
| 2025-09-30 | 02_architecture.md | v1.1 | 统一组件职责描述 |
| 2025-09-30 | 04_deployment.md | v1.1 | 优化资源配置和启动依赖说明 |
| 2025-09-30 | 00_deployment_guide.md | v1.1 | 统一FAQ描述 |

## 快速查找表

### 按关键词查找

| 关键词 | 相关文档 | 章节 |
|--------|----------|------|
| **架构设计** | [02_architecture.md](./02_architecture.md) | 全文 |
| **事件网关** | [02_architecture.md](./02_architecture.md) | 3.1 |
| **编排器** | [02_architecture.md](./02_architecture.md) | 3.2 |
| **AI 推理** | [02_architecture.md](./02_architecture.md) | 3.3 |
| **命令执行** | [02_architecture.md](./02_architecture.md) | 3.4 |
| **多集群** | [02_architecture.md](./02_architecture.md) | 4 |
| **数据模型** | [03_data_models.md](./03_data_models.md) | 全文 |
| **DiagnosticTask** | [03_data_models.md](./03_data_models.md) | 2.1 |
| **知识库** | [03_data_models.md](./03_data_models.md) | 2.4 |
| **部署** | [04_deployment.md](./04_deployment.md) | 全文 |
| **Kubernetes** | [04_deployment.md](./04_deployment.md) | 3 |
| **PostgreSQL** | [04_deployment.md](./04_deployment.md) | 3.2 |
| **Redis** | [04_deployment.md](./04_deployment.md) | 3.3 |
| **Helm** | [04_deployment.md](./04_deployment.md) | 8 |
| **安全** | [05_operations.md](./05_operations.md) | 2 |
| **认证授权** | [05_operations.md](./05_operations.md) | 2.1-2.2 |
| **RBAC** | [05_operations.md](./05_operations.md) | 2.2 |
| **加密** | [05_operations.md](./05_operations.md) | 2.3 |
| **审计** | [05_operations.md](./05_operations.md) | 2.5 |
| **监控** | [05_operations.md](./05_operations.md) | 4 |
| **成本管理** | [05_operations.md](./05_operations.md) | 5 |
| **备份恢复** | [05_operations.md](./05_operations.md) | 6 |
| **故障排查** | [05_operations.md](./05_operations.md) | 7 |
| **功能需求** | [REQUIREMENTS.md](../REQUIREMENTS.md) | 全文 |
| **非功能需求** | [REQUIREMENTS.md](../REQUIREMENTS.md) | 全文 |

### 按技术栈查找

| 技术 | 相关文档 | 位置 |
|------|----------|------|
| **Go 1.21+** | [02_architecture.md](./02_architecture.md) <br> [03_data_models.md](./03_data_models.md) | 多处 |
| **Kubernetes** | [04_deployment.md](./04_deployment.md) <br> [05_operations.md](./05_operations.md) | 全文 |
| **PostgreSQL** | [04_deployment.md](./04_deployment.md) | 3.2, 5 |
| **Redis** | [04_deployment.md](./04_deployment.md) | 3.3, 5 |
| **HashiCorp Vault** | [04_deployment.md](./04_deployment.md) <br> [05_operations.md](./05_operations.md) | 3.4, 2.3.3 |
| **Weaviate** | [04_deployment.md](./04_deployment.md) | 3.5 |
| **OpenAI GPT-4** | [02_architecture.md](./02_architecture.md) | 3.3 |
| **Prometheus** | [05_operations.md](./05_operations.md) | 4.1 |
| **Grafana** | [05_operations.md](./05_operations.md) | 4 |
| **OpenTelemetry** | [05_operations.md](./05_operations.md) | 4.2 |

## 学习路径建议

### 🎓 初级 (1-2 天)

```
Day 1: 理解系统
├─ 上午: README.md + 00_overview.md
│  └─ 目标: 了解 Aetherius 是什么,能做什么
├─ 下午: REQUIREMENTS.md
│  └─ 目标: 理解系统功能需求
└─ 晚上: 02_architecture.md (第 1-2 章)
   └─ 目标: 理解系统架构概览

Day 2: 数据与部署
├─ 上午: 03_data_models.md
│  └─ 目标: 理解核心数据结构
├─ 下午: 04_deployment.md (第 1-3 章)
│  └─ 目标: 了解部署流程
└─ 晚上: 实践练习
   └─ 目标: 在本地环境尝试部署
```

### 🚀 中级 (3-5 天)

```
Day 3: 深入架构
├─ 02_architecture.md (第 3-5 章)
│  └─ 目标: 理解各核心组件设计
└─ ai_agent.md (第 5 章)
   └─ 目标: 深入技术实现细节

Day 4: 安全与运维
├─ 05_operations.md (第 2-3 章)
│  └─ 目标: 理解安全架构和威胁模型
└─ 04_deployment.md (第 9-10 章)
   └─ 目标: 掌握配置管理和故障排查

Day 5: 监控与优化
├─ 05_operations.md (第 4-5 章)
│  └─ 目标: 掌握监控和成本控制
└─ ai_agent.md (第 8 章)
   └─ 目标: 理解可观测性设计
```

### 🎯 高级 (1 周+)

```
Week 1: 系统精通
├─ 完整阅读所有专题文档
├─ 深入研究 ai_agent.md 完整内容
├─ 实践部署生产级环境
└─ 编写自定义组件

Week 2: 贡献与优化
├─ 识别系统改进点
├─ 设计优化方案
├─ 实施性能调优
└─ 编写技术文档
```

## 文档维护指南

### 更新文档时需要同步的内容

```
更新 ai_agent.md 时:
├─ 检查相关专题文档是否需要更新
├─ 更新 REQUIREMENTS.md 中的需求索引
├─ 更新 00_index_diagram.md 的版本记录
└─ 在 README.md 中添加变更说明

创建新文档时:
├─ 在 README.md 中添加导航链接
├─ 在 00_index_diagram.md 中添加文档关系
├─ 在 REQUIREMENTS.md 中添加交叉引用
└─ 更新相关文档的"相关文档"章节
```

### 文档质量检查清单

```
□ 标题层级正确 (从 # 开始,不跳级)
□ 代码块指定语言类型
□ 表格格式规范
□ 链接有效且指向正确
□ 中英文混排时有空格分隔
□ 使用列表符号一致 (-/*)
□ 文件路径使用相对路径
□ 图表 ASCII 绘制整齐
□ 版本信息及时更新
□ 符合 MarkdownLint 规则
```

## 反馈与贡献

### 文档问题报告

如发现文档问题,请通过以下方式反馈:

1. **内容错误**: 在 GitHub Issues 中标记为 `documentation`
2. **链接失效**: 在相关文档的评论区指出
3. **需求变更**: 更新 REQUIREMENTS.md 并提交 PR

### 文档改进建议

欢迎提交改进建议:

- 增加示例代码
- 补充最佳实践
- 完善故障排查指南
- 优化学习路径

---

**文档版本**: v1.6
**最后更新**: 2025年9月30日
**维护团队**: Aetherius开发团队