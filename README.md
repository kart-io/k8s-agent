# Aetherius - 智能 Kubernetes 运维平台

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Python Version](https://img.shields.io/badge/Python-3.11+-3776AB?logo=python)
![Kubernetes](https://img.shields.io/badge/Kubernetes-1.23+-326CE5?logo=kubernetes)

> 基于 AI 的智能 Kubernetes 故障诊断与自动修复平台

---

## 🎯 项目简介

Aetherius 是一个企业级智能 Kubernetes 运维平台，采用 4 层架构设计，结合事件驱动和 AI 技术，实现从数据采集到智能分析的完整闭环。

### 核心能力

- ✅ **自动发现**: 实时监控 K8s 集群异常事件
- ✅ **根因分析**: AI 驱动的多模态根因分析 (事件+日志+指标)
- ✅ **智能推荐**: 基于规则和历史案例的修复建议
- ✅ **自动修复**: 工作流驱动的自动化修复执行
- ✅ **持续学习**: 从反馈中学习，持续提高准确率
- ✅ **多集群管理**: 统一管理数百个 K8s 集群
- ✅ **知识沉淀**: 知识图谱存储运维经验

---

## 🏗️ 架构设计

```plaintext
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes Clusters                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                  │
│  │  Cluster │  │  Cluster │  │  Cluster │  ...              │
│  │    1     │  │    2     │  │    N     │                   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘                  │
│       │             │             │                          │
│  ┌────▼─────────────▼─────────────▼─────┐                  │
│  │     Layer 1: Collect Agent           │                  │
│  │   (事件监控 + 指标采集 + 命令执行)   │                  │
│  └──────────────────┬───────────────────┘                  │
└─────────────────────┼──────────────────────────────────────┘
                      │ NATS
┌─────────────────────▼──────────────────────────────────────┐
│       Layer 2: Agent Manager (中央控制平面)                 │
│  - Agent 注册管理    - 事件处理    - 命令分发              │
│  - 多集群管理        - 数据存储    - REST API              │
└─────────────────────┬──────────────────────────────────────┘
                      │ Internal Events
┌─────────────────────▼──────────────────────────────────────┐
│     Layer 3: Orchestrator Service (任务编排)                │
│  - 工作流引擎        - 诊断策略    - 自动修复              │
│  - 任务调度          - AI 集成     - 事件订阅              │
└─────────────────────┬──────────────────────────────────────┘
                      │ HTTP API
┌─────────────────────▼──────────────────────────────────────┐
│      Layer 4: Reasoning Service (AI 智能)                   │
│  - 根因分析          - 故障预测    - 智能推荐              │
│  - 知识图谱          - 持续学习    - 案例检索              │
└─────────────────────────────────────────────────────────────┘
```

详细架构设计请查看: [系统架构文档](docs/architecture/SYSTEM_ARCHITECTURE.md)

---

## 🚀 快速开始

### 使用 Docker Compose (推荐用于开发/测试)

```bash
# 1. 启动所有服务
cd deployments/docker-compose
docker-compose up -d

# 2. 检查服务状态
docker-compose ps

# 3. 验证健康状态
curl http://localhost:8080/health  # Agent Manager
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8082/health  # Reasoning Service
```

详细说明: [Docker Compose 部署指南](deployments/docker-compose/README.md)

### 使用 Kubernetes (推荐用于生产)

```bash
# 1. 创建命名空间
kubectl apply -f deployments/k8s/namespace.yaml

# 2. 部署依赖服务 (PostgreSQL, Redis, NATS, Neo4j)
kubectl apply -f deployments/k8s/dependencies.yaml

# 3. 等待依赖服务就绪
kubectl -n aetherius wait --for=condition=ready pod -l app=postgres --timeout=300s

# 4. 部署应用服务
kubectl apply -f deployments/k8s/agent-manager.yaml
kubectl apply -f deployments/k8s/orchestrator-service.yaml
kubectl apply -f deployments/k8s/reasoning-service.yaml

# 5. 验证部署
kubectl -n aetherius get pods
```

详细说明: [Kubernetes 部署指南](deployments/k8s/README.md)

---

## 📦 组件说明

### Layer 1: Collect Agent (边缘采集层)

部署在每个 Kubernetes 集群中，负责数据采集和命令执行。

- **技术栈**: Go 1.21+, client-go, NATS
- **核心功能**:
  - K8s 事件监控 (85+ 种关键事件)
  - 资源指标采集 (集群/节点/Pod/命名空间)
  - 安全命令执行 (kubectl/诊断工具)
- **部署**: DaemonSet 或 Deployment
- **文档**: [Collect Agent README](collect-agent/README.md)

```bash
cd collect-agent
make build
make run
```

---

### Layer 2: Agent Manager (中央控制层)

中央控制平面，管理所有 Agent，处理事件，分发命令。

- **技术栈**: Go 1.21+, PostgreSQL, Redis, NATS, Gin
- **核心功能**:
  - Agent 生命周期管理 (注册/心跳/状态)
  - 事件聚合与路由 (过滤/去重/关联)
  - 命令调度与分发 (验证/安全/跟踪)
  - 多集群管理
  - RESTful API
- **API 端口**: 8080
- **文档**: [Agent Manager README](agent-manager/README.md)

```bash
cd agent-manager
make build
make run
```

---

### Layer 3: Orchestrator Service (任务编排层)

工作流编排，自动诊断和修复。

- **技术栈**: Go 1.21+, PostgreSQL, Redis, NATS
- **核心功能**:
  - 工作流引擎 (步骤执行/重试/分支)
  - 诊断策略 (模式匹配/工作流触发)
  - 步骤执行器 (6 种类型: Command/AI/Decision/Remediation/Notification/Wait)
  - AI 集成 (调用 reasoning-service)
  - 事件订阅
- **API 端口**: 8081
- **文档**: [Orchestrator Service README](orchestrator-service/README.md)

```bash
cd orchestrator-service
make build
make run
```

---

### Layer 4: Reasoning Service (AI 智能层)

AI 驱动的根因分析、故障预测和智能推荐。

- **技术栈**: Python 3.11+, FastAPI, PyTorch, scikit-learn, Neo4j
- **核心功能**:
  - 根因分析引擎 (多模态: 事件+日志+指标)
  - 推荐引擎 (30+ 修复建议规则)
  - 预测引擎 (趋势分析+异常检测)
  - 知识图谱 (历史案例存储)
  - 持续学习系统
- **API 端口**: 8082
- **文档**: [Reasoning Service README](reasoning-service/README.md)

```bash
cd reasoning-service
make install
make run
```

---

## 🔧 开发指南

### 环境要求

- **Go**: 1.21+
- **Python**: 3.11+
- **Docker**: 20.10+
- **Kubernetes**: 1.23+
- **PostgreSQL**: 14+
- **Redis**: 6+
- **NATS**: 2.10+
- **Neo4j**: 5+ (可选)

### 本地开发

1. **启动依赖服务**:

```bash
cd deployments/docker-compose
docker-compose up -d postgres redis nats neo4j
```

2. **运行各个服务**:

```bash
# Terminal 1: Agent Manager
cd agent-manager && make run

# Terminal 2: Orchestrator Service
cd orchestrator-service && make run

# Terminal 3: Reasoning Service
cd reasoning-service && make dev

# Terminal 4: Collect Agent (可选)
cd collect-agent && make run
```

### 构建镜像

```bash
# Agent Manager
cd agent-manager
make docker-build

# Orchestrator Service
cd orchestrator-service
make docker-build

# Reasoning Service
cd reasoning-service
make docker-build

# Collect Agent
cd collect-agent
make docker-build
```

---

## 📊 监控指标

### 系统指标

- Agent 在线数量
- 事件处理速率 (events/sec)
- 工作流执行成功率
- API 响应时间
- 资源使用率 (CPU/Memory)

### 业务指标

- 根因分析准确率: **~85-95%**
- 自动修复成功率: **~80-90%**
- 平均故障发现时间 (MTTD): **< 1 分钟**
- 平均修复时间 (MTTR): **< 5 分钟** (自动修复)

---

## 🎓 使用示例

### 示例 1: 自动诊断 Pod CrashLoopBackOff

1. Collect Agent 发现 CrashLoopBackOff 事件
2. Agent Manager 评估为关键事件，发布到内部总线
3. Orchestrator Service 匹配策略，启动诊断工作流:
   - 收集 Pod 日志
   - 获取资源描述
   - 调用 AI 分析
   - 识别根因: OOM Killer
   - 推荐修复: 增加内存限制
4. 执行自动修复或通知运维人员

### 示例 2: 预测性维护

1. Collect Agent 定期采集资源指标
2. Agent Manager 检测内存使用率持续上升
3. 发布异常事件到 Orchestrator
4. Orchestrator 调用 Reasoning Service 预测
5. 预测结果: 2 小时后可能 OOM
6. 提前告警并建议扩容

---

## 📈 性能指标

### 处理能力

- **单个 Agent Manager**: 支持 1000+ Agents, 处理 10000+ events/min
- **单个 Orchestrator**: 并发 500+ 工作流, 吞吐 5000+ tasks/min
- **单个 Reasoning Service**: 100+ 分析请求/min, P99 延迟 < 5s

### 扩展性

- 支持 **数百个** Kubernetes 集群
- 支持 **数万个** Pod 监控
- 事件处理延迟 **< 1 秒**
- 工作流触发延迟 **< 5 秒**

---

## 🔒 安全设计

- **认证**: JWT Token, mTLS
- **授权**: RBAC, 命令白名单
- **传输加密**: TLS 1.3
- **存储加密**: 数据库 TDE
- **审计日志**: 所有关键操作记录

---

## 🗺️ 路线图

### Phase 1: 核心功能 (已完成 ✅)

- [x] 4 层架构实现
- [x] 事件采集和处理
- [x] 根因分析 (多模态)
- [x] 智能推荐 (30+ 规则)
- [x] 工作流编排
- [x] 自动修复

### Phase 2: 增强功能 (进行中 🚧)

- [ ] Web UI 界面
- [ ] 完整的 RBAC
- [ ] 多租户支持
- [ ] 高级工作流 (并行、循环)
- [ ] 更多内置策略 (100+)

### Phase 3: 智能升级 (规划中 📝)

- [ ] LLM 集成 (GPT/Claude)
- [ ] 深度学习模型
- [ ] 自然语言查询
- [ ] 智能对话修复
- [ ] 预测性维护

### Phase 4: 生态完善 (未来)

- [ ] 插件系统
- [ ] 自定义 Operator
- [ ] 多云支持
- [ ] 可视化编排器
- [ ] 社区知识库

---

## 📚 文档

- [系统架构](docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Docker Compose 部署](deployments/docker-compose/README.md)
- [Kubernetes 部署](deployments/k8s/README.md)
- [Collect Agent](collect-agent/README.md)
- [Agent Manager](agent-manager/README.md)
- [Orchestrator Service](orchestrator-service/README.md)
- [Reasoning Service](reasoning-service/README.md)

---

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

### 贡献者

感谢所有贡献者的付出！

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

---

## 💬 社区

- **Issues**: [GitHub Issues](https://github.com/kart-io/k8s-agent/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kart-io/k8s-agent/discussions)
- **Slack**: [加入 Slack](https://aetherius-slack.example.com)

---

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=kart-io/k8s-agent&type=Date)](https://star-history.com/#kart-io/k8s-agent&Date)

---

## 🙏 致谢

感谢以下开源项目:

- [Kubernetes](https://kubernetes.io/)
- [NATS](https://nats.io/)
- [FastAPI](https://fastapi.tiangolo.com/)
- [Neo4j](https://neo4j.com/)
- [PyTorch](https://pytorch.org/)

---

**Built with ❤️ by Aetherius Team**