# Aetherius 智能 Kubernetes 运维平台 - 系统架构图

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Kubernetes Clusters (多集群)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                      │
│  │   Cluster 1  │  │   Cluster 2  │  │   Cluster N  │                      │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │                      │
│  │  │ Pods   │  │  │  │ Pods   │  │  │  │ Pods   │  │                      │
│  │  │ Events │  │  │  │ Events │  │  │  │ Events │  │                      │
│  │  │Metrics │  │  │  │Metrics │  │  │  │Metrics │  │                      │
│  │  └────────┘  │  │  └────────┘  │  │  └────────┘  │                      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                      │
│         │                 │                 │                                │
│    ┌────▼─────────────────▼─────────────────▼────────┐                     │
│    │  Layer 1: Collect Agent (DaemonSet/Deployment)  │                     │
│    │  - K8s Event Monitoring (85+ types)             │                     │
│    │  - Resource Metrics Collection                   │                     │
│    │  - Command Execution                             │                     │
│    │  - Heartbeat Management                          │                     │
│    └──────────────────────┬───────────────────────────┘                     │
└───────────────────────────┼──────────────────────────────────────────────────┘
                            │
                            │ NATS Messages
                            │ (events, metrics, heartbeats)
                            │
┌───────────────────────────▼──────────────────────────────────────────────────┐
│                  Layer 2: Agent Manager (Port 8080)                          │
│                         中央控制平面                                          │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Agent Registration & Lifecycle Management                  │            │
│  │  - Agent registration, heartbeat monitoring                 │            │
│  │  - Agent status tracking (online/offline)                   │            │
│  │  - Multi-cluster metadata management                        │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Event Processing & Aggregation                             │            │
│  │  - Event filtering, deduplication                           │            │
│  │  - Severity evaluation                                      │            │
│  │  - Event correlation                                        │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Command Dispatch & Execution                                │            │
│  │  - Command validation & whitelisting                        │            │
│  │  - Secure command routing to agents                         │            │
│  │  - Execution result tracking                                │            │
│  └─────────────────────────────────────────────────────────────┘            │
│                                                                               │
│  Storage: MySQL (agents, clusters, events, commands)                        │
│  Cache: Redis (sessions, temporary data)                                    │
│                                                                               │
│  ┌─────────────────────┐                                                    │
│  │   RESTful API       │                                                    │
│  │  /api/v1/agents     │                                                    │
│  │  /api/v1/clusters   │                                                    │
│  │  /api/v1/events     │                                                    │
│  │  /api/v1/commands   │                                                    │
│  └─────────────────────┘                                                    │
└───────────────────────────┬──────────────────────────────────────────────────┘
                            │
                            │ Internal Event Bus (NATS)
                            │ (critical events, workflow triggers)
                            │
┌───────────────────────────▼──────────────────────────────────────────────────┐
│              Layer 3: Orchestrator Service (Port 8081)                       │
│                         任务编排层                                            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Workflow Engine                                             │            │
│  │  - Step execution (sequential/parallel)                     │            │
│  │  - Retry & error handling                                   │            │
│  │  - State management                                         │            │
│  │  - Conditional branching                                    │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Diagnostic Strategy Management                              │            │
│  │  - Event pattern matching                                   │            │
│  │  - Strategy-to-workflow mapping                             │            │
│  │  - Priority-based triggering                                │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Step Executors (6 Types)                                   │            │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │            │
│  │  │  Command    │  │     AI      │  │  Decision   │        │            │
│  │  │  Execute    │  │  Analysis   │  │  Branching  │        │            │
│  │  │  kubectl    │  │  Call RCA   │  │  If/Else    │        │            │
│  │  └─────────────┘  └─────────────┘  └─────────────┘        │            │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │            │
│  │  │Remediation  │  │Notification │  │    Wait     │        │            │
│  │  │Scale/Restart│  │Send Alerts  │  │   Delay     │        │            │
│  │  └─────────────┘  └─────────────┘  └─────────────┘        │            │
│  └─────────────────────────────────────────────────────────────┘            │
│                                                                               │
│  Storage: MySQL (workflows, strategies, executions, history)                │
│  Cache: Redis (workflow state, temporary results)                           │
│                                                                               │
│  ┌─────────────────────┐                                                    │
│  │   RESTful API       │                                                    │
│  │  /api/v1/workflows  │                                                    │
│  │  /api/v1/strategies │                                                    │
│  │  /api/v1/executions │                                                    │
│  └─────────────────────┘                                                    │
└───────────────────────────┬──────────────────────────────────────────────────┘
                            │
                            │ HTTP API Calls
                            │ (root cause analysis, recommendations)
                            │
┌───────────────────────────▼──────────────────────────────────────────────────┐
│              Layer 4: Reasoning Service (Port 8082)                          │
│                         AI 智能层 (Go)                                        │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Root Cause Analysis Engine                                  │            │
│  │  - Multi-modal analysis (events + logs + metrics)           │            │
│  │  - Pattern recognition                                      │            │
│  │  - Causal chain analysis                                    │            │
│  │  - Confidence scoring                                       │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  AI Agent Chain                                              │            │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │            │
│  │  │  Analyzer    │→ │ Recommender  │→ │  Validator   │      │            │
│  │  │  Agent       │  │   Agent      │  │   Agent      │      │            │
│  │  └──────────────┘  └──────────────┘  └──────────────┘      │            │
│  │         ↓                  ↓                  ↓             │            │
│  │  ┌──────────────────────────────────────────────────────┐  │            │
│  │  │         LLM Integration Layer                        │  │            │
│  │  │  OpenAI API │ Gemini API │ DeepSeek API             │  │            │
│  │  └──────────────────────────────────────────────────────┘  │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Recommendation Engine                                       │            │
│  │  - 30+ built-in repair rules                                │            │
│  │  - Similar case retrieval (from knowledge graph)            │            │
│  │  - Action prioritization                                    │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Prediction Engine                                           │            │
│  │  - Trend analysis                                           │            │
│  │  - Anomaly detection (Isolation Forest)                     │            │
│  │  - Proactive alerting                                       │            │
│  └─────────────────────────────────────────────────────────────┘            │
│  ┌─────────────────────────────────────────────────────────────┐            │
│  │  Continuous Learning System                                  │            │
│  │  - Feedback processing                                      │            │
│  │  - Accuracy tracking                                        │            │
│  │  - Model improvement                                        │            │
│  └─────────────────────────────────────────────────────────────┘            │
│                                                                               │
│  Storage: Neo4j Graph DB (knowledge graph, cases, relationships)            │
│  Memory: Vector embeddings for semantic search                              │
│                                                                               │
│  ┌─────────────────────┐                                                    │
│  │   RESTful API       │                                                    │
│  │  /api/v1/analyze    │                                                    │
│  │  /api/v1/recommend  │                                                    │
│  │  /api/v1/predict    │                                                    │
│  │  /api/v1/feedback   │                                                    │
│  └─────────────────────┘                                                    │
└───────────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────────┐
│                         Supporting Services                                    │
├───────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐           │
│  │  Auth Service    │  │ Gateway Service  │  │ Monitor Service  │           │
│  │  (JWT, Session)  │  │   (Traefik)      │  │  (Prometheus)    │           │
│  │  Port: 8083      │  │   Port: 80/443   │  │  Port: 9090      │           │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘           │
│  ┌──────────────────┐                                                        │
│  │ Cluster Service  │                                                        │
│  │ (Multi-cluster)  │                                                        │
│  │  Port: 8084      │                                                        │
│  └──────────────────┘                                                        │
└───────────────────────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────────────────────────┐
│                         Infrastructure Layer                                   │
├───────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │    MySQL     │  │    Redis     │  │     NATS     │  │    Neo4j     │    │
│  │   Port:3306  │  │  Port: 6379  │  │  Port: 4222  │  │ Port: 7474   │    │
│  │              │  │              │  │              │  │      7687    │    │
│  │ - Agents DB  │  │ - Sessions   │  │ - Events     │  │ - Knowledge  │    │
│  │ - Events DB  │  │ - Cache      │  │ - Commands   │  │   Graph      │    │
│  │ - Workflows  │  │ - Locks      │  │ - Pub/Sub    │  │ - Cases      │    │
│  │ - Auth DB    │  │              │  │              │  │              │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
└───────────────────────────────────────────────────────────────────────────────┘
```

## 数据流示例：CrashLoopBackOff 自动诊断与修复

```
1. Event Detection (Layer 1)
   Collect Agent detects Pod CrashLoopBackOff event
   ↓
   Publishes to NATS: {"type": "Warning", "reason": "BackOff", "pod": "app-xyz"}

2. Event Aggregation (Layer 2)
   Agent Manager receives event
   ↓
   Evaluates severity: HIGH
   ↓
   Stores in MySQL + Publishes to internal event bus

3. Workflow Triggering (Layer 3)
   Orchestrator subscribes to event
   ↓
   Matches diagnostic strategy: "pod-crash-diagnosis"
   ↓
   Executes workflow:

   Step 1 [Command]: Get pod logs
     → Agent Manager → Collect Agent → kubectl logs

   Step 2 [Command]: Describe pod
     → Agent Manager → Collect Agent → kubectl describe pod

   Step 3 [AI]: Analyze root cause
     → HTTP POST /api/v1/analyze
     → Reasoning Service

   Step 4 [Decision]: Check confidence > 0.8
     → If true: proceed to remediation
     → If false: notify human operator

   Step 5 [AI]: Get recommendations
     → HTTP POST /api/v1/recommend
     → Reasoning Service

   Step 6 [Remediation]: Apply fix
     → Update resource limits (if OOM)
     → Or restart pod (if config issue)
     → Agent Manager → Collect Agent → kubectl apply

   Step 7 [Wait]: Observe 30s

   Step 8 [Command]: Verify pod status
     → Check if pod is Running

   Step 9 [Notification]: Send result
     → Email/Slack notification

4. AI Analysis (Layer 4)
   Reasoning Service receives analysis request
   ↓
   Analyzer Agent:
   - Extracts features from logs & events
   - Identifies error patterns
   - Queries knowledge graph for similar cases
   ↓
   LLM Integration:
   - Sends context to OpenAI/Gemini/DeepSeek
   - Gets structured analysis result
   ↓
   Recommender Agent:
   - Retrieves 30+ rule-based suggestions
   - Ranks by relevance and confidence
   - Filters by feasibility
   ↓
   Returns: {
     "root_cause": "OOM Killer terminated container",
     "confidence": 0.92,
     "recommendations": [
       {"action": "increase_memory_limit", "priority": 1},
       {"action": "add_memory_request", "priority": 2}
     ]
   }

5. Feedback Loop
   Operator confirms fix worked
   ↓
   Feedback sent to Reasoning Service
   ↓
   Continuous Learning System:
   - Updates case in Neo4j knowledge graph
   - Increases confidence for similar patterns
   - Improves future accuracy
```

## 服务间通信协议

| Source | Destination | Protocol | Purpose |
|--------|-------------|----------|---------|
| Collect Agent | Agent Mgr | NATS | Events, Metrics |
| Collect Agent | Agent Mgr | NATS | Heartbeat |
| Agent Manager | Orchestrator | NATS | Internal Events |
| Agent Manager | Collect Agent | NATS | Commands |
| Orchestrator | Reasoning | HTTP/REST | AI Analysis |
| Orchestrator | Agent Mgr | HTTP/REST | Command Dispatch |
| Gateway | All Services | HTTP/REST | API Routing |
| Auth Service | All Services | JWT Validation | Authentication |
| All Services | MySQL | TCP/3306 | Data Storage |
| All Services | Redis | TCP/6379 | Cache/Sessions |
| Reasoning Service | Neo4j | Bolt/7687 | Knowledge Graph |
| Reasoning Service | LLM APIs | HTTPS | AI Processing |

## 服务端口映射

| Service | Port | Protocol | Purpose |
|---------|------|----------|---------|
| Agent Manager | 8080 | HTTP | RESTful API |
| Orchestrator Service | 8081 | HTTP | RESTful API |
| Reasoning Service | 8082 | HTTP | RESTful API |
| Auth Service | 8083 | HTTP | Authentication API |
| Cluster Service | 8084 | HTTP | Cluster Management API |
| Gateway Service | 80/443 | HTTP/HTTPS | API Gateway |
| Monitor Service | 9090 | HTTP | Prometheus Metrics |
| MySQL | 3306 | TCP | Database |
| Redis | 6379 | TCP | Cache/Sessions |
| NATS | 4222 | TCP | Messaging |
| NATS Monitoring | 8222 | HTTP | NATS Admin |
| Neo4j HTTP | 7474 | HTTP | Graph DB Admin |
| Neo4j Bolt | 7687 | Bolt | Graph DB Query |

## 关键设计模式

### 1. Event-Driven Architecture (EDA)

- **NATS 作为事件总线**: 解耦服务间通信
- **Pub/Sub 模式**: 一对多事件分发
- **异步处理**: 提高系统吞吐量和响应能力
- **事件溯源**: 所有关键事件可追溯

### 2. Workflow as Code

- **声明式工作流**: YAML 定义诊断策略
- **步骤类型可扩展**: 6 种基础步骤类型 + 自定义扩展
- **条件分支**: 基于 AI 分析结果动态决策
- **状态管理**: Redis 存储工作流执行状态

### 3. AI Agent Chain

- **多 Agent 协作**: Analyzer → Recommender → Validator
- **责任链模式**: 每个 Agent 专注单一职责
- **LLM 集成抽象层**: 支持多个 LLM 提供商 (OpenAI/Gemini/DeepSeek)
- **上下文传递**: Agent 间共享分析上下文

### 4. Knowledge Graph

- **Neo4j 存储**: 图数据库存储历史案例和关系
- **相似案例检索**: 基于图查询找到类似问题
- **持续学习**: 从反馈中更新知识图谱
- **模式识别**: 识别故障模式和因果关系

### 5. Multi-tenancy & Multi-cluster

- **租户隔离**: 数据和权限按租户隔离
- **集群元数据管理**: 统一管理数百个 K8s 集群
- **统一控制平面**: Agent Manager 作为中心化控制点
- **水平扩展**: 支持大规模多集群场景

## 性能与扩展性

### 吞吐量指标

- **Agent Manager**: 支持 1000+ Agents, 处理 10,000+ events/min
- **Orchestrator**: 并发 500+ workflows, 吞吐 5,000+ tasks/min
- **Reasoning Service**: 100+ 分析请求/min

### 延迟指标

- **事件处理延迟**: < 1s (从检测到存储)
- **工作流触发延迟**: < 5s (从事件到工作流启动)
- **AI 分析延迟**: P99 < 5s

### 扩展性设计

- **无状态服务**: 所有服务设计为无状态，易于水平扩展
- **数据分片**: MySQL 数据库支持分片
- **缓存层**: Redis 减轻数据库压力
- **消息队列**: NATS 异步解耦，提高并发能力

## 安全设计

### 认证与授权

- **JWT Token**: 无状态认证机制
- **RBAC**: 基于角色的访问控制
- **mTLS**: 服务间通信加密
- **API Key**: 外部系统接入认证

### 数据安全

- **传输加密**: TLS 1.3
- **存储加密**: 数据库 TDE (Transparent Data Encryption)
- **敏感信息脱敏**: 日志和 API 响应中敏感信息脱敏
- **审计日志**: 所有关键操作记录审计日志

### 命令执行安全

- **命令白名单**: 只允许预定义的安全命令
- **权限验证**: 执行前验证操作者权限
- **沙箱执行**: 隔离环境中执行命令
- **结果审计**: 记录所有命令执行结果

## 高可用设计

### 服务高可用

- **多副本部署**: 所有服务至少 2 个副本
- **健康检查**: Liveness 和 Readiness 探针
- **滚动更新**: 零停机部署
- **熔断机制**: 防止级联故障

### 数据高可用

- **MySQL 主从复制**: 数据冗余
- **Redis 哨兵模式**: 自动故障转移
- **NATS 集群**: 消息队列高可用
- **Neo4j 集群**: 知识图谱数据冗余

### 灾难恢复

- **定期备份**: 数据库和知识图谱定期备份
- **异地容灾**: 支持多数据中心部署
- **快速恢复**: 自动化恢复流程
- **数据一致性**: 使用分布式事务保证数据一致性
