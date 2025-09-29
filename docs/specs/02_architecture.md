# Aetherius AI Agent - 系统架构设计

> **文档版本**: v1.6
> **最后更新**: 2025年9月28日
> **读者对象**: 系统架构师、开发工程师

---

## 目录

- [1. 架构总览](#1-架构总览)
- [2. 核心组件](#2-核心组件)
- [3. 多集群架构](#3-多集群架构)
- [4. 数据流架构](#4-数据流架构)
- [5. 安全架构](#5-安全架构)
- [6. 关键设计决策](#6-关键设计决策)

---

## 1. 架构总览

### 1.1 系统架构图

> **权威来源**: 完整架构详见 [ai_agent.md#5.1.1](../ai_agent.md#511-整体系统架构)
>
> **本节目的**: 提供系统架构的快速概览，详细的组件交互和技术实现请参考ai_agent.md

**逻辑分层架构**:

```
┌─────────────────────────────────────────────────┐
│ 输入层 (Input Layer)                            │
│ • Alertmanager Webhooks                         │
│ • K8s Event Streams                             │
│ • Manual Triggers (Dashboard/API)               │
├─────────────────────────────────────────────────┤
│ 事件网关层 (Event Gateway Layer)                │
│ • Webhook Handler                               │
│ • Event Filter & Validator                     │
│ • Event Standardization                        │
├─────────────────────────────────────────────────┤
│ 编排层 (Orchestration Layer)                    │
│ • Task Scheduler                                │
│ • Priority Queue Manager                       │
│ • Workflow Engine                              │
├─────────────────────────────────────────────────┤
│ 智能分析层 (AI Analysis Layer)                  │
│ • Reasoning Service (LLM)                      │
│ • Knowledge Base (RAG)                         │
│ • Strategy Generator                           │
├─────────────────────────────────────────────────┤
│ 执行层 (Execution Layer)                        │
│ • Execution Gateway (MCP Protocol)             │
│ • Tool Registry                                │
│ • Security Validator                           │
├─────────────────────────────────────────────────┤
│ 输出层 (Output Layer)                           │
│ • Report Generator                              │
│ • Notification Service (Multi-channel)         │
│ • Dashboard API                                 │
└─────────────────────────────────────────────────┘
```

### 1.2 架构特性

| 特性 | 说明 | 优势 | 体现 |
|------|------|------|------|
| **事件驱动** | 基于事件的异步架构 | 高吞吐、松耦合 | NATS消息总线、异步处理 |
| **微服务** | 组件独立部署和扩展 | 易维护、可扩展 | 6大核心服务独立运行 |
| **无状态设计** | 服务实例无状态 | 水平扩展、高可用 | 状态存储在数据层 |
| **分层架构** | 清晰的职责分离 | 易理解、可测试 | 6层逻辑架构设计 |
| **插件化** | 工具注册表机制 | 易扩展、可定制 | Tool Registry动态配置 |

> **架构一致性说明**: 本架构设计与 [ai_agent.md#5章](../ai_agent.md#5-系统设计-system-design) 和 [06_microservices.md](./06_microservices.md) 保持完全一致

### 1.3 技术栈

详见: [ai_agent.md#2.4](../ai_agent.md#24-运行环境-operating-environment)

| 层次 | 技术选型 | 版本 |
|------|----------|------|
| **编程语言** | Go | 1.21+ |
| **容器编排** | Kubernetes | v1.20-v1.28 |
| **数据库** | PostgreSQL | 14+ |
| **缓存队列** | Redis | 6+ |
| **向量数据库** | Weaviate/Qdrant | 最新稳定版 |
| **密钥管理** | HashiCorp Vault | 最新稳定版 |
| **AI服务** | OpenAI/Anthropic | GPT-4/Claude |

---

## 2. 核心组件

### 2.1 组件职责矩阵

详见: [ai_agent.md#5.1.2核心服务详细架构](../ai_agent.md#512-核心服务详细架构)

| 组件 | 职责 | 输入 | 输出 | 依赖 |
|------|------|------|------|------|
| **Event Gateway** | 接收和验证事件 | Webhook/Event | 标准化事件 | None |
| **Orchestrator** | 任务调度和编排 | 标准化事件 | DiagnosticTask | Redis, PostgreSQL |
| **Reasoning Service** | AI推理和策略生成 | DiagnosticTask | ExecutionPlan | LLM, Vector DB |
| **Execution Gateway** | 安全执行命令 | ExecutionPlan | ExecutionResult | Vault, K8s API |
| **Report Service** | 生成和分发报告 | ExecutionResult | DiagnosticReport | SMTP, Slack API |
| **Knowledge Base** | 存储和检索知识 | Query | RelevantDocs | Vector DB |

### 2.2 Event Gateway (事件网关)

**功能**: 统一接收和处理各类输入事件

**子组件**:
- Webhook Handler: 处理HTTP webhook请求
- Event Filter: 过滤和去重事件
- Schema Validator: 验证事件格式
- Event Publisher: 发布到内部事件总线

**关键接口**:

```go
type EventGateway interface {
    ReceiveWebhook(ctx context.Context, payload []byte) error
    ReceiveK8sEvent(ctx context.Context, event K8sEvent) error
    ValidateEvent(event Event) error
    PublishEvent(event StandardizedEvent) error
}
```

**配置参数**:
- `webhook_timeout`: webhook处理超时时间 (默认: 30s)
- `max_payload_size`: 最大载荷大小 (默认: 10MB)
- `deduplication_window`: 去重时间窗口 (默认: 5分钟)

### 2.3 Orchestrator (编排器)

**功能**: 任务调度、优先级管理、流程控制

详见: [ai_agent.md#5.3.1](../ai_agent.md#531-任务状态管理)

**子组件**:
- Task Scheduler: 从队列取出任务
- Priority Manager: 管理任务优先级
- Workflow Engine: 控制诊断流程
- State Manager: 管理任务状态

**状态机**:

```
pending → running → [completed | failed | timeout | cancelled]
```

**优先级算法**:

```go
priority = base_priority(severity)
         + cluster_weight
         + namespace_weight
         + historical_frequency_factor
```

### 2.4 Reasoning Service (推理服务)

**功能**: AI驱动的智能诊断和策略生成

详见: [ai_agent.md#5.9.2](../ai_agent.md#592-诊断策略选择逻辑-diagnostic-strategy-selection)

**核心能力**:
- 知识库语义检索 (RAG)
- LLM推理和分析
- 执行计划生成
- 置信度评估

**工作流程**:

```
1. 接收诊断任务
2. 查询知识库 (相似度检索)
3. 构建LLM上下文 (任务+知识+历史)
4. 生成诊断策略
5. 创建执行计划
6. 返回执行计划
```

**模型选择策略**:

```go
if knowledge_match_score >= 0.8 {
    strategy = "knowledge_based"  // 直接使用知识库
} else if ai_available && budget_ok {
    if complexity >= 7 {
        model = "gpt-4"  // 复杂问题
    } else {
        model = "gpt-3.5-turbo"  // 简单问题
    }
} else {
    strategy = "rule_based"  // 降级到规则引擎
}
```

### 2.5 Execution Gateway (执行网关)

**功能**: 安全执行诊断命令

详见: [ai_agent.md#5.4](../ai_agent.md#54-工具与能力-tools--capabilities)

**5层安全检查**:

```
1. 工具注册表验证 (命令必须已注册)
2. 只读操作检查 (禁止破坏性操作)
3. 参数安全验证 (无危险参数)
4. 权限验证 (RBAC检查)
5. 集群权限验证 (cluster_id匹配)
```

**MCP协议实现**:

```go
type MCPServer interface {
    ExecuteCommand(ctx context.Context, cmd Command) (*Result, error)
    ValidateCommand(cmd Command) error
    GetCredentials(clusterID string) (*Credentials, error)
}
```

**动态凭证获取**:

详见: [ai_agent.md#5.5.2](../ai_agent.md#552-动态凭证获取流程)

```
1. 从任务中提取 cluster_id
2. 向Vault请求短期Token (1小时有效)
3. 使用Token建立K8s连接
4. 执行命令
5. 清理临时凭证
```

### 2.6 Knowledge Base (知识库)

**功能**: 运维知识的存储和检索

详见: [ai_agent.md#6.4](../ai_agent.md#64-知识库模型-knowledge-base-model)

**存储架构**:

```
PostgreSQL (元数据)
    ├─ knowledge_entries (知识条目)
    ├─ knowledge_categories (分类)
    ├─ knowledge_tags (标签)
    └─ knowledge_versions (版本)

Vector DB (向量索引)
    ├─ embeddings (768维向量)
    ├─ similarity_index (HNSW索引)
    └─ metadata (元数据关联)
```

**检索流程**:

```
1. 查询文本向量化 (Embedding)
2. 向量相似度搜索 (Top-K)
3. 元数据过滤 (cluster_type, k8s_version)
4. 置信度评分
5. 返回相关知识
```

**更新策略**:

详见: [ai_agent.md#5.9.5](../ai_agent.md#595-知识库更新决策-knowledge-base-update-decision)

---

## 3. 多集群架构

### 3.1 多集群管理模式

详见: [ai_agent.md#5.5](../ai_agent.md#55-多集群管理策略-multi-cluster-management-strategy)

**设计原则**:
- 🏢 **中央化管理**: 单一控制平面
- 🌐 **分布式执行**: 每个集群独立执行
- 🔐 **凭证隔离**: 不同集群凭证严格隔离
- 📚 **知识共享**: 全局知识库跨集群复用

### 3.2 多集群架构图

完整架构图参见: [ai_agent.md#5.5.1](../ai_agent.md#551-多集群架构图)

```
中央管理平台 (Central Platform)
    ├─ Orchestrator
    ├─ Reasoning Service
    ├─ Execution Gateway
    └─ Shared Data Layer
        ├─ PostgreSQL (任务元数据)
        ├─ Redis (任务队列)
        ├─ Vector DB (全局知识库)
        └─ Vault (凭证管理)
         │
         ├──────────────┬──────────────┬──────────────
         │              │              │
    Cluster A      Cluster B      Cluster C
    (prod-us)      (prod-eu)      (staging)
```

### 3.3 集群注册与配置

详见: [ai_agent.md#5.5.3](../ai_agent.md#553-集群注册与配置管理)

**配置示例**:

```yaml
clusters:
  - id: "prod-us-west"
    name: "Production US West"
    region: "us-west-2"
    environment: "production"
    vault_path: "kubernetes/prod-us-west"
    monitoring:
      alertmanager_url: "https://alertmanager.prod-us-west.com"
    access_control:
      max_concurrent_tasks: 10
      allowed_operations: ["get", "list", "describe", "logs"]
      rate_limit: "100/hour"
```

### 3.4 跨集群诊断流程

详见: [ai_agent.md#3.2.3](../ai_agent.md#323-多集群协调流程-multi-cluster-coordination)

```
1. 中央调度器接收任务
2. 根据cluster_id路由到目标集群
3. 执行网关获取动态凭证
4. 通过MCP代理执行命令
5. 收集结果并汇聚
6. 生成统一诊断报告
```

---

## 4. 数据流架构

### 4.1 端到端数据流

详见: [ai_agent.md#5.6](../ai_agent.md#56-数据流向与处理架构-data-flow--processing-architecture)

```
数据输入层
    ↓ (JSON payload)
数据标准化层
    ↓ (StandardizedEvent)
任务编排层
    ↓ (DiagnosticTask)
智能处理层
    ↓ (ExecutionPlan)
安全执行层
    ↓ (ExecutionResult)
分析决策层
    ↓ (DiagnosticAnalysis)
输出生成层
    ↓ (DiagnosticReport)
反馈闭环
    ↓ (UserFeedback → Knowledge Base)
```

### 4.2 数据持久化架构

详见: [ai_agent.md#5.6.2](../ai_agent.md#562-数据持久化架构)

**存储分层**:

```
应用层 (Services)
    ↓
缓存层 (Redis Cluster)
    ├─ 任务队列 (LIST)
    ├─ 会话状态 (HASH)
    ├─ 查询缓存 (STRING, TTL 1h)
    └─ 配置缓存 (JSON)
    ↓
持久化层
    ├─ PostgreSQL (关系型数据)
    │   ├─ diagnostic_tasks
    │   ├─ diagnostic_steps
    │   ├─ diagnostic_results
    │   ├─ user_feedback
    │   ├─ audit_logs
    │   └─ system_config
    │
    ├─ Vector DB (向量索引)
    │   ├─ knowledge_entries
    │   ├─ embeddings (768d)
    │   └─ similarity_index (HNSW)
    │
    └─ HashiCorp Vault (密钥管理)
        ├─ cluster_credentials
        ├─ api_tokens
        └─ encryption_keys
```

### 4.3 数据一致性模型

详见: [ai_agent.md#5.7.3](../ai_agent.md#573-数据一致性保证机制-data-consistency-guarantees)

**强一致性** (PostgreSQL事务):
- 任务状态变更
- 审计日志写入
- 关键配置更新

**最终一致性** (缓存+数据库):
- 知识库更新 (1-5分钟延迟)
- 统计指标更新 (5-10分钟延迟)
- 监控数据同步 (15秒延迟)

---

## 5. 安全架构

### 5.1 安全分层设计

详见: [ai_agent.md#10章](../ai_agent.md#10-安全考量)

```
┌─────────────────────────────────────────┐
│ 边界层: 认证、授权、网络隔离             │
├─────────────────────────────────────────┤
│ 应用层: 只读约束、命令白名单、参数验证   │
├─────────────────────────────────────────┤
│ 数据层: 加密传输、加密存储、数据脱敏     │
├─────────────────────────────────────────┤
│ 审计层: 完整日志、行为追踪、合规检查     │
└─────────────────────────────────────────┘
```

### 5.2 认证与授权

**身份认证**: OIDC/OAuth2.0
**权限控制**: RBAC (基于角色)

**权限模型**:

```
User
  ├─ has Roles
  │   ├─ Admin (管理员)
  │   ├─ Operator (运维工程师)
  │   ├─ Developer (开发工程师)
  │   └─ Viewer (只读用户)
  │
  └─ has Permissions
      ├─ clusters:read
      ├─ tasks:create
      ├─ tasks:read
      ├─ tasks:control
      ├─ knowledge:read
      └─ knowledge:write
```

### 5.3 数据安全

**传输加密**: TLS 1.3
**存储加密**: AES-256
**敏感数据脱敏**: 自动识别和脱敏

**脱敏规则**:

```go
var SensitivePatterns = []string{
    `password\s*[:=]\s*\S+`,     // 密码
    `secret\s*[:=]\s*\S+`,       // 密钥
    `token\s*[:=]\s*\S+`,        // Token
    `api[_-]?key\s*[:=]\s*\S+`,  // API Key
    `\d{3}-\d{2}-\d{4}`,         // SSN
    `\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}`, // 信用卡
}
```

### 5.4 网络安全

**网络策略**: K8s NetworkPolicy
**服务网格**: Istio (可选)
**入口控制**: Ingress with mTLS

**网络隔离**:

```
Internet
    ↓ (TLS termination)
Ingress Controller
    ↓ (mTLS)
API Gateway (认证/授权)
    ↓ (内部网络)
Backend Services
    ↓ (mTLS)
Data Stores
```

---

## 6. 关键设计决策

### 6.1 为什么选择事件驱动架构?

**优势**:
- ✅ 高吞吐量: 异步处理提高并发能力
- ✅ 松耦合: 组件独立演进
- ✅ 可扩展: 易于添加新的事件处理器
- ✅ 弹性: 故障隔离和恢复

**权衡**:
- ⚠️ 复杂性: 调试和追踪更困难
- ⚠️ 一致性: 需要处理最终一致性

### 6.2 为什么选择只读执行模式?

详见: [ai_agent.md#3.6.2](../ai_agent.md#362-技术约束-technical-constraints)

**原因**:
- 🛡️ 安全: 防止AI误操作破坏生产环境
- 📋 合规: 符合审计和变更管理要求
- 🔍 专注: 聚焦诊断而非自动修复

**实现**:
- 5层安全检查机制
- 命令白名单
- 参数安全验证

### 6.3 为什么使用动态凭证?

详见: [ai_agent.md#5.5.2](../ai_agent.md#552-动态凭证获取流程)

**优势**:
- 🔐 安全: 短期Token (1小时)
- 🔄 动态: 按需获取，用后即弃
- 📊 审计: 每次访问可追踪
- 🚫 隔离: 不持久化长期凭证

**实现**: HashiCorp Vault + K8s Service Account

### 6.4 为什么采用中央化+分布式模式?

详见: [ai_agent.md#5.5](../ai_agent.md#55-多集群管理策略-multi-cluster-management-strategy)

**中央化**:
- 统一调度和管理
- 全局知识库共享
- 简化运维

**分布式**:
- 降低网络延迟
- 提高容错能力
- 支持大规模扩展

### 6.5 为什么选择RAG而非纯LLM?

**RAG优势**:
- 📚 知识增强: 结合专业运维知识
- 🎯 准确性: 提高诊断准确率
- 💰 成本: 减少LLM Token消耗
- 🔄 可更新: 知识库持续演进

**实现**: Vector DB + Semantic Search + LLM

### 6.6 架构演进路线

详见: [ai_agent.md#11章](../ai_agent.md#11-迭代路线图-roadmap)

**v1.6 (当前)**: 多集群管理 + 成本控制
**v1.7 (计划)**: 增强AI能力 + 自动修复建议
**v1.8 (计划)**: 多云支持 + 边缘部署
**v2.0 (愿景)**: 自主修复 + 预测性维护

---

## 相关文档

- [需求规格](../REQUIREMENTS.md) - 功能和非功能需求
- [数据模型](./03_data_models.md) - 核心数据结构
- [部署指南](./04_deployment.md) - 部署和配置
- [安全设计](../ai_agent.md#10-安全考量) - 详细安全设计

---

**文档维护**: Aetherius架构团队
**最后更新**: 2025年9月28日