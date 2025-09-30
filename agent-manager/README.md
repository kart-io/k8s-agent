# Aetherius Agent Manager

中央控制平面服务,负责管理所有边缘 collect-agent 实例,聚合数据并提供统一的管理接口。

---

## 目录

- [功能特性](#功能特性)
- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 文档](#api-文档)
- [部署指南](#部署指南)
- [开发指南](#开发指南)

---

## 功能特性

### 核心功能

- **Agent 生命周期管理**: 注册、心跳监控、状态追踪
- **事件处理引擎**: 事件接收、过滤、聚合、路由
- **指标存储**: 时序数据存储和查询
- **命令调度**: 安全的命令分发和结果收集
- **多集群管理**: 统一管理多个 Kubernetes 集群

### 技术特性

- **高可用**: 无状态设计,支持水平扩展
- **实时通信**: 基于 NATS 的消息总线
- **数据持久化**: PostgreSQL + Redis 双存储
- **RESTful API**: 完整的 HTTP API
- **可观测性**: Prometheus 指标 + 结构化日志

---

## 架构设计

```plaintext
┌─────────────────────────────────────────────────────────────┐
│                    Agent Manager                             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌────────────────┐  ┌──────────────┐  ┌────────────────┐ │
│  │ Agent Registry │  │ Event        │  │ Command        │ │
│  │                │  │ Processor    │  │ Dispatcher     │ │
│  │ - Register     │  │ - Filter     │  │ - Validate     │ │
│  │ - Heartbeat    │  │ - Enrich     │  │ - Dispatch     │ │
│  │ - Status Track │  │ - Aggregate  │  │ - Track Result │ │
│  └────────────────┘  └──────────────┘  └────────────────┘ │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │             NATS Message Bus                         │   │
│  │  - Agent Registration/Heartbeat                      │   │
│  │  - Event Ingestion                                   │   │
│  │  - Metrics Collection                                │   │
│  │  - Command Distribution                              │   │
│  │  - Internal Event Publishing                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌────────────────┐  ┌──────────────┐                      │
│  │ PostgreSQL     │  │ Redis        │                      │
│  │ - Agents       │  │ - Cache      │                      │
│  │ - Events       │  │ - Sessions   │                      │
│  │ - Commands     │  │ - Queues     │                      │
│  │ - Clusters     │  │ - Counters   │                      │
│  └────────────────┘  └──────────────┘                      │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │             RESTful API Server                       │   │
│  │  - Agent Management                                  │   │
│  │  - Cluster Management                                │   │
│  │  - Event Query                                       │   │
│  │  - Command API                                       │   │
│  │  - Health/Metrics                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         ↓                                    ↑
    NATS Publish                         HTTP Requests
         ↓                                    ↑
  collect-agents                         Users/Orchestrator
```

---

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- NATS Server 2.10+
- Docker & Docker Compose (可选)

### 使用 Docker Compose (推荐)

```bash
# 1. 克隆仓库
git clone https://github.com/kart-io/k8s-agent.git
cd k8s-agent/agent-manager

# 2. 启动所有服务
docker-compose up -d

# 3. 查看日志
docker-compose logs -f agent-manager

# 4. 访问 API
curl http://localhost:8080/health/status
```

### 本地开发

```bash
# 1. 安装依赖
make deps

# 2. 启动依赖服务
docker-compose up -d postgres redis nats

# 3. 运行应用
make run

# 或者开发模式 (热重载)
make dev
```

### 验证部署

```bash
# 健康检查
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
curl http://localhost:8080/health/status

# Prometheus 指标
curl http://localhost:8080/metrics

# 列出 Agents
curl http://localhost:8080/api/v1/agents
```

---

## 配置说明

配置文件: `configs/config.yaml`

### 核心配置项

#### HTTP 服务器

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  graceful_stop: 10s
```

#### NATS 连接

```yaml
nats:
  url: "nats://localhost:4222"
  cluster_id: "agent-manager"
  max_reconnect: -1  # 无限重试
  reconnect_wait: 2s
  ping_interval: 20s
  max_pings_out: 3
```

#### 数据库配置

```yaml
database:
  host: "localhost"
  port: 5432
  user: "aetherius"
  password: "aetherius123"
  database: "aetherius"
  ssl_mode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s
```

#### Redis 配置

```yaml
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 3
```

### 环境变量覆盖

```bash
# 数据库
export DB_HOST=postgres.example.com
export DB_PORT=5432
export DB_USER=aetherius
export DB_PASSWORD=secret
export DB_NAME=aetherius

# Redis
export REDIS_ADDR=redis.example.com:6379
export REDIS_PASSWORD=secret

# NATS
export NATS_URL=nats://nats.example.com:4222
```

---

## API 文档

### 健康检查

#### GET /health/live

存活检查 (用于 Kubernetes liveness probe)

```bash
curl http://localhost:8080/health/live
```

#### GET /health/ready

就绪检查 (用于 Kubernetes readiness probe)

```bash
curl http://localhost:8080/health/ready
```

#### GET /health/status

详细状态信息

```bash
curl http://localhost:8080/health/status | jq .
```

### Agent 管理

#### GET /api/v1/agents

列出所有 Agents

```bash
# 所有 Agents
curl http://localhost:8080/api/v1/agents

# 过滤在线 Agents
curl http://localhost:8080/api/v1/agents?status=online
```

#### GET /api/v1/agents/:id

获取 Agent 详情

```bash
curl http://localhost:8080/api/v1/agents/{agent-id}
```

#### DELETE /api/v1/agents/:id

删除 Agent

```bash
curl -X DELETE http://localhost:8080/api/v1/agents/{agent-id}
```

### 集群管理

#### GET /api/v1/clusters

列出所有集群

```bash
curl http://localhost:8080/api/v1/clusters
```

#### GET /api/v1/clusters/:id

获取集群详情

```bash
curl http://localhost:8080/api/v1/clusters/{cluster-id}
```

#### POST /api/v1/clusters

创建集群

```bash
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "id": "prod-us-west",
    "name": "Production US West",
    "environment": "prod",
    "provider": "eks"
  }'
```

#### GET /api/v1/clusters/:id/health

获取集群健康状态

```bash
curl http://localhost:8080/api/v1/clusters/{cluster-id}/health
```

### 事件查询

#### GET /api/v1/events

查询事件

```bash
# 查询特定集群的事件
curl "http://localhost:8080/api/v1/events?cluster_id=prod-us-west"

# 按严重性过滤
curl "http://localhost:8080/api/v1/events?severity=critical"

# 按命名空间过滤
curl "http://localhost:8080/api/v1/events?namespace=production"
```

#### GET /api/v1/events/:id

获取事件详情

```bash
curl http://localhost:8080/api/v1/events/{event-id}
```

#### POST /api/v1/events/search

高级搜索

```bash
curl -X POST http://localhost:8080/api/v1/events/search \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "prod-us-west",
    "severity": "critical",
    "start_time": "2025-09-30T00:00:00Z",
    "end_time": "2025-09-30T23:59:59Z",
    "limit": 100
  }'
```

### 命令管理

#### POST /api/v1/commands

发送命令到 Agent

```bash
curl -X POST http://localhost:8080/api/v1/commands \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "prod-us-west",
    "type": "diagnostic",
    "tool": "kubectl",
    "action": "get",
    "args": ["pods", "-n", "production"],
    "timeout": "30s"
  }'
```

#### GET /api/v1/commands/:id

获取命令状态

```bash
curl http://localhost:8080/api/v1/commands/{command-id}
```

#### GET /api/v1/commands/:id/result

获取命令执行结果

```bash
curl http://localhost:8080/api/v1/commands/{command-id}/result
```

---

## 部署指南

### Docker 部署

```bash
# 构建镜像
make docker-build

# 推送镜像
make docker-push

# 运行容器
make docker-run
```

### Kubernetes 部署

```bash
# 部署到 Kubernetes
kubectl apply -f deployments/k8s/

# 查看状态
kubectl -n aetherius-system get pods
kubectl -n aetherius-system logs deployment/agent-manager

# 访问服务
kubectl -n aetherius-system port-forward svc/agent-manager 8080:8080
```

### 生产环境建议

#### 高可用配置

```yaml
# deployment.yaml
replicas: 3  # 至少 3 个副本

affinity:
  podAntiAffinity:  # 分散到不同节点
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app: agent-manager
          topologyKey: kubernetes.io/hostname
```

#### 资源配置

```yaml
resources:
  requests:
    cpu: 500m
    memory: 1Gi
  limits:
    cpu: 2000m
    memory: 4Gi
```

#### 监控和告警

```yaml
# ServiceMonitor for Prometheus
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: agent-manager
spec:
  selector:
    matchLabels:
      app: agent-manager
  endpoints:
  - port: http
    path: /metrics
```

---

## 开发指南

### 项目结构

```plaintext
agent-manager/
├── cmd/
│   └── server/           # 主程序入口
├── internal/
│   ├── agent/            # Agent 注册中心
│   ├── api/              # HTTP API 服务
│   ├── command/          # 命令调度器
│   ├── event/            # 事件处理引擎
│   ├── nats/             # NATS 服务端
│   └── storage/          # 存储层 (PostgreSQL + Redis)
├── pkg/
│   ├── types/            # 数据类型定义
│   └── utils/            # 工具函数
├── configs/              # 配置文件
├── deployments/          # 部署配置
│   ├── docker/           # Docker 相关
│   └── k8s/              # Kubernetes 清单
├── docs/                 # 文档
├── go.mod
├── Makefile
└── README.md
```

### 添加新功能

1. 在 `internal/` 下创建新模块
2. 在 `pkg/types/` 添加类型定义
3. 在 `internal/api/` 添加 API 端点
4. 更新 `cmd/server/main.go` 初始化新组件
5. 添加测试用例
6. 更新文档

### 运行测试

```bash
# 运行所有测试
make test

# 生成覆盖率报告
make test-coverage

# 查看覆盖率
open coverage.html
```

### 代码格式化

```bash
# 格式化代码
make fmt

# 运行 linter
make lint
```

---

## 故障排查

### 常见问题

#### 1. 无法连接到 NATS

```bash
# 检查 NATS 状态
docker-compose ps nats

# 查看 NATS 日志
docker-compose logs nats

# 测试连接
nc -zv localhost 4222
```

#### 2. 数据库连接失败

```bash
# 检查 PostgreSQL
docker-compose ps postgres

# 连接测试
psql -h localhost -U aetherius -d aetherius -c "SELECT 1"
```

#### 3. Agent 未注册

- 检查 Agent 配置中的 `central_endpoint`
- 确认网络连通性
- 查看 Agent 日志

### 日志级别

```yaml
# config.yaml
logging:
  level: "debug"  # 临时设置为 debug
```

---

## 性能优化

### 数据库优化

```sql
-- 创建索引
CREATE INDEX idx_events_cluster_severity ON events(cluster_id, severity);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);
CREATE INDEX idx_agents_status ON agents(status);
```

### Redis 优化

```yaml
redis:
  pool_size: 20      # 增加连接池
  min_idle_conns: 5  # 保持最小空闲连接
```

### NATS 优化

```yaml
nats:
  ping_interval: 10s   # 更频繁的 ping
  max_pings_out: 5     # 增加 ping 容忍度
```

---

## 贡献指南

欢迎贡献!请遵循以下步骤:

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

---

## 许可证

MIT License

---

## 联系方式

- GitHub: https://github.com/kart-io/k8s-agent
- 问题反馈: https://github.com/kart-io/k8s-agent/issues
- 文档: [docs/architecture/SYSTEM_ARCHITECTURE.md](../../docs/architecture/SYSTEM_ARCHITECTURE.md)