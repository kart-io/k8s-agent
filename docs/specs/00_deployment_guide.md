# Aetherius 部署架构选择指南

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025年9月27日
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: 部署架构选择指南

## 目录

- [1. 快速选择](#1-快速选择)
- [2. 架构模式对比](#2-架构模式对比)
- [3. 部署决策树](#3-部署决策树)
- [4. 文档导航](#4-文档导航)

## 1. 快速选择

根据您的集群规模,选择合适的部署架构:

| 集群数量 | 推荐架构 | 参考文档 | 主要优势 | 适用场景说明 |
|---------|---------|---------|----------|--------------|
| **1个集群** | 单集群In-Cluster | [08_in_cluster_deployment.md](./08_in_cluster_deployment.md) | 简单可靠,全功能 | 独立的生产/开发/测试集群 |
| **2-10个集群** | Agent代理模式 | [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) | 轻量Agent,统一管理 | 多集群统一监控和诊断 |
| **10+个集群** | 分层Agent架构 | 待规划 | 分层管理,可扩展 | 超大规模多集群环境 |

## 2. 架构模式对比

### 2.1 单集群 In-Cluster 模式

**架构示意**:

```text
┌─────────────────────────────────────┐
│         Kubernetes Cluster          │
│                                     │
│  ┌──────────────────────────────┐  │
│  │   Aetherius Full Stack       │  │
│  │                              │  │
│  │  • Event Gateway (Webhook)   │  │
│  │  • K8s Event Watcher         │  │
│  │  • Orchestrator              │  │
│  │  • Reasoning Service         │  │
│  │  • Execution Service         │  │
│  │  • Knowledge Service         │  │
│  │  • ... (所有服务)            │  │
│  └──────────────────────────────┘  │
│              ▲                      │
│              │ Watch                │
│         K8s API Server              │
└─────────────────────────────────────┘
```

**适用场景**:
- ✅ 单一生产集群
- ✅ 需要完整功能
- ✅ 资源充足

**参考文档**:
- [08_in_cluster_deployment.md](./08_in_cluster_deployment.md) - In-Cluster部署实现
- [06_microservices.md](./06_microservices.md) - 微服务架构
- [07_k8s_event_watcher.md](./07_k8s_event_watcher.md) - K8s事件监听

### 2.2 Agent 代理模式

**架构示意**:

```
┌──────────────────────────────────────┐
│   Central Control Plane              │
│   (部署在任意位置)                    │
│                                      │
│  • Agent Manager                     │
│  • Orchestrator                      │
│  • Reasoning Service                 │
│  • Knowledge Service                 │
│  • ... (核心服务)                    │
└──────────┬───────────────────────────┘
           │ NATS
           ▼
  ┌────────┴────────┐
  │                 │
  ▼                 ▼
┌─────────┐    ┌─────────┐
│Cluster A│    │Cluster B│
│         │    │         │
│ Agent   │    │ Agent   │
│ (轻量)  │    │ (轻量)  │
└─────────┘    └─────────┘
```

**适用场景**:
- ✅ 2-10个集群
- ✅ 多云/混合云
- ✅ Agent主动连接(无需Central访问集群)

**参考文档**:
- [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) - Agent代理模式完整实现

### 2.3 架构模式详细对比

| 对比项 | 单集群In-Cluster | Agent代理模式 |
|-------|----------------|--------------|
| **部署位置** | 集群内部 | 中心+多个边缘Agent |
| **网络要求** | 无特殊要求 | Agent能访问Central的NATS |
| **资源消耗** | 较高(全功能) | 低(Agent轻量) |
| **集群隔离** | 完全隔离 | Agent隔离,数据集中 |
| **扩展性** | 每集群独立部署 | 新增Agent即可 |
| **管理复杂度** | 低(单集群) | 中等(多集群统一管理) |
| **适用规模** | 1个集群 | 2-10个集群 |

## 3. 部署决策树

```
开始
  │
  ├─ 需要监控几个集群?
  │
  ├─ 1个集群
  │   └─► 使用"单集群In-Cluster模式"
  │       └─► 参考: 08_in_cluster_deployment.md
  │
  ├─ 2-10个集群
  │   │
  │   ├─ 集群是否分布在多云/混合云?
  │   │   ├─ 是
  │   │   │   └─► 使用"Agent代理模式"
  │   │   │       └─► 参考: 09_agent_proxy_mode.md
  │   │   │
  │   │   └─ 否(同一网络环境)
  │   │       ├─ Central能否访问所有集群API?
  │   │       │   ├─ 能
  │   │       │   │   └─► 考虑"Out-of-Cluster中心模式"
  │   │       │   │       (需要维护所有kubeconfig)
  │   │       │   │
  │   │       │   └─ 不能
  │   │       │       └─► 使用"Agent代理模式"
  │   │       │           └─► 参考: 09_agent_proxy_mode.md
  │   │       │
  │   │       └─► 推荐使用"Agent代理模式"(更灵活)
  │
  └─ 10+个集群
      └─► 使用"分层Agent架构"
          └─► 待规划(可先用Agent代理模式验证)
```

## 4. 文档导航

### 4.1 核心架构文档

| 文档 | 描述 | 适用场景 |
|-----|------|---------|
| [02_architecture.md](./02_architecture.md) | 系统整体架构 | 所有场景 |
| [03_data_models.md](./03_data_models.md) | 数据模型定义 | 所有场景 |
| [06_microservices.md](./06_microservices.md) | 微服务架构设计 | 单集群/Central |

### 4.2 技术实现文档

| 文档 | 描述 | 适用场景 |
|-----|------|---------|
| [07_k8s_event_watcher.md](./07_k8s_event_watcher.md) | K8s事件监听实现 | 所有场景(核心技术) |
| [08_in_cluster_deployment.md](./08_in_cluster_deployment.md) | In-Cluster部署 | 单集群 |
| [09_agent_proxy_mode.md](./09_agent_proxy_mode.md) | Agent代理模式 | 多集群 |

### 4.3 运维文档

| 文档 | 描述 | 适用场景 |
|-----|------|---------|
| [04_deployment.md](./04_deployment.md) | 部署配置 | 所有场景 |
| [05_operations.md](./05_operations.md) | 运维与安全 | 所有场景 |

## 5. 常见问题

### Q1: 我有3个集群,应该选哪个架构?

**A**: 推荐使用Agent代理模式([09_agent_proxy_mode.md](./09_agent_proxy_mode.md))。
- 在中心集群部署完整控制平面
- 在其他2个集群各部署一个轻量级Agent
- Agent自动上报事件到中心,统一管理

### Q2: Agent代理模式和单集群模式能否共存?

**A**: 可以,但需要注意:
- 如果中心控制平面部署在单集群In-Cluster模式的集群内,该集群监控两种方式:
  - 本地直接监控(Event Gateway + K8s Event Watcher)
  - 也可以部署Agent上报给本地Central(可选,通常不需要)
- 其他集群统一使用Agent模式

### Q3: Event Gateway 在Agent代理模式下还需要吗?

**A**: 需要,但职责保持一致(仅处理外部告警源):

| 部署模式 | Event Gateway职责 | K8s事件处理 |
|---------|------------------|-------------|
| **单集群模式** | 接收外部告警源<br>(Alertmanager Webhook) | 由Event Watcher组件<br>独立处理 |
| **Agent代理模式** | 接收外部告警源<br>(Alertmanager Webhook) | 由各Agent组件<br>独立处理和上报 |

**关键点**: Event Gateway在两种模式下都**不直接处理K8s事件**,始终专注于外部告警源。

参考: [06_microservices.md 第2.1节](./06_microservices.md#21-event-gateway-service-事件网关服务) 了解详细职责说明

### Q4: NATS消息总线是同一个吗?

**A**: 不是,有两个独立的NATS集群:
- **微服务内部NATS**: 用于微服务间通信(Orchestrator、Reasoning等)
- **Agent通信NATS**: 专用于Agent与Central通信,需要对外暴露

### Q5: 如何从单集群扩展到多集群?

**A**: 平滑迁移步骤:
1. 保持现有单集群In-Cluster部署不变
2. 在现有集群外部署Agent通信NATS(LoadBalancer)
3. 部署Agent Manager服务
4. 在新集群部署Agent,连接到中心
5. (可选)将原集群也改造为Agent模式

## 6. 架构演进路径

```
阶段1: 单集群验证
   使用: 单集群In-Cluster
   目标: 验证功能,熟悉系统
   └─► 08_in_cluster_deployment.md

阶段2: 多集群扩展
   使用: Agent代理模式
   目标: 统一管理2-10个集群
   └─► 09_agent_proxy_mode.md

阶段3: 超大规模(未来)
   使用: 分层Agent架构
   目标: 10+个集群,分层管理
   └─► 待规划
```

## 附录

### A. 术语对照

| 术语 | 说明 |
|-----|------|
| **In-Cluster** | 部署在Kubernetes集群内部 |
| **Out-of-Cluster** | 部署在Kubernetes集群外部 |
| **Agent代理模式** | 轻量Agent+中心控制平面的分布式架构 |
| **Event Gateway** | 事件网关服务,接收外部事件源 |
| **K8s Event Watcher** | K8s事件监听组件,使用Informer机制 |
| **Agent Manager** | 管理所有Agent的中心服务 |

### B. 快速链接

- **我是新手**: 从[08_in_cluster_deployment.md](./08_in_cluster_deployment.md)开始
- **我需要多集群**: 查看[09_agent_proxy_mode.md](./09_agent_proxy_mode.md)
- **我要了解架构**: 阅读[02_architecture.md](./02_architecture.md)和[06_microservices.md](./06_microservices.md)