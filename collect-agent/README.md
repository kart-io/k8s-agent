# Aetherius Collect Agent

Aetherius 采集代理是一个轻量级的 Kubernetes Agent,负责从集群中收集事件和指标,并通过 NATS 消息总线上报到中央控制平面。

---

## 目录

- [功能特性](#功能特性)
- [架构设计](#架构设计)
- [配置说明](#配置说明)
- [部署指南](#部署指南)
- [开发指南](#开发指南)
- [监控运维](#监控运维)
- [故障排查](#故障排查)

---

## 功能特性

### 核心功能

- **事件监控**: 智能过滤 Kubernetes 事件,重点关注关键问题 (85+ 种事件类型)
- **指标采集**: 收集集群、节点、Pod 级别的性能指标
- **命令执行**: 安全执行只读诊断命令 (kubectl, ps, df 等)
- **NATS 通信**: 可靠的消息传输,支持自动重连
- **云平台检测**: 自动检测集群 ID (支持 AWS EKS, GCP GKE, Azure AKS)
- **健康监控**: 内置健康检查和 Prometheus 指标
- **安全加固**: 非 root 运行,最小 RBAC 权限

### 技术特性

- **轻量级**: 内存占用 < 128MB,CPU < 100m
- **高可用**: 自动重连,故障自愈
- **可观测**: 结构化日志,Prometheus 指标
- **云原生**: 容器化部署,支持 DaemonSet/Deployment

---

## 架构设计

### 组件架构

```plaintext
┌─────────────────────────────────────────────────────────────┐
│               Aetherius Collect Agent                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Event        │  │ Metrics      │  │ Command      │     │
│  │ Watcher      │  │ Collector    │  │ Executor     │     │
│  │              │  │              │  │              │     │
│  │ - Monitor    │  │ - Cluster    │  │ - kubectl    │     │
│  │ - Filter     │  │ - Node       │  │ - Diagnostic │     │
│  │ - Classify   │  │ - Pod        │  │ - Safe Exec  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │          Communication Manager                      │    │
│  │  - NATS Connection                                  │    │
│  │  - Message Publishing                               │    │
│  │  - Command Subscription                             │    │
│  │  - Auto Reconnection                                │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │          Health Server                              │    │
│  │  - Liveness Probe                                   │    │
│  │  - Readiness Probe                                  │    │
│  │  - Prometheus Metrics                               │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
         ↓ NATS Publish                    ↑ NATS Subscribe
         ↓                                  ↑
    agent-manager                      Commands
```

### 数据流

```plaintext
K8s Events → Event Watcher → Filter → NATS → Agent Manager
K8s Metrics → Metrics Collector → Aggregate → NATS → Agent Manager
Commands ← NATS ← Agent Manager → Command Executor → K8s API
```

---

## 配置说明

### 配置文件

Agent 支持通过 YAML 文件或环境变量配置:

```yaml
# 集群标识 (为空时自动检测)
cluster_id: ""

# NATS 中央服务器地址
central_endpoint: "nats://agent-manager.aetherius.svc:4222"

# 连接配置
reconnect_delay: 5s          # 重连延迟
max_retries: 10              # 最大重试次数

# 心跳和指标间隔
heartbeat_interval: 30s      # 心跳发送间隔
metrics_interval: 60s        # 指标采集间隔

# 缓冲配置
buffer_size: 1000            # 消息缓冲区大小

# 功能开关
enable_metrics: true         # 启用指标采集
enable_events: true          # 启用事件监控

# 日志配置
log_level: "info"            # 日志级别: debug, info, warn, error, fatal
```

### 环境变量

所有配置项都可以通过环境变量覆盖:

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `CLUSTER_ID` | 集群 ID | 自动检测 |
| `CENTRAL_ENDPOINT` | NATS 服务器地址 | - |
| `LOG_LEVEL` | 日志级别 | `info` |
| `RECONNECT_DELAY` | 重连延迟 | `5s` |
| `HEARTBEAT_INTERVAL` | 心跳间隔 | `30s` |
| `METRICS_INTERVAL` | 指标采集间隔 | `60s` |
| `BUFFER_SIZE` | 缓冲区大小 | `1000` |
| `MAX_RETRIES` | 最大重试次数 | `10` |
| `ENABLE_METRICS` | 启用指标采集 | `true` |
| `ENABLE_EVENTS` | 启用事件监控 | `true` |

### 云平台自动检测

Agent 可以自动从以下云平台检测集群 ID:

- **AWS EKS**: 从 EC2 元数据获取集群名称
- **GCP GKE**: 从 GCE 元数据获取集群名称
- **Azure AKS**: 从 Azure IMDS 获取集群名称
- **本地集群**: 使用 `kube-system` namespace UID

---

## 部署指南

### 前置要求

- Kubernetes 集群 v1.20+
- NATS 服务器可访问
- kubectl 命令行工具

### 快速部署

#### 1. 自定义配置

编辑 ConfigMap 中的 NATS 端点:

```bash
# 修改 manifests/03-configmap.yaml
sed -i 's|central.aetherius.io:4222|YOUR-NATS-ENDPOINT:4222|g' manifests/03-configmap.yaml
```

#### 2. 部署 Agent

```bash
# 部署所有清单
kubectl apply -f manifests/

# 或使用 Makefile
cd collect-agent
make k8s-deploy
```

#### 3. 验证部署

```bash
# 查看 Pod 状态
kubectl -n aetherius-agent get pods

# 查看日志
kubectl -n aetherius-agent logs deployment/aetherius-agent

# 检查健康状态
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080
curl http://localhost:8080/health/status
```

### 部署方式

#### DaemonSet 部署 (推荐)

每个节点运行一个 Agent 实例,适合大规模集群:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: aetherius-agent
  namespace: aetherius-agent
spec:
  selector:
    matchLabels:
      app: aetherius-agent
  template:
    spec:
      containers:
      - name: agent
        image: aetherius/collect-agent:v1.0.0
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
```

#### Deployment 部署

单个 Agent 实例,适合小型集群:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-agent
spec:
  replicas: 1
  # ... 其他配置
```

### 构建自定义镜像

```bash
# 1. 构建镜像
docker build -t your-registry/collect-agent:v1.0.0 .

# 2. 推送到镜像仓库
docker push your-registry/collect-agent:v1.0.0

# 3. 更新部署清单
sed -i 's|aetherius/collect-agent:v1.0.0|your-registry/collect-agent:v1.0.0|g' \
  manifests/04-deployment.yaml

# 4. 重新部署
kubectl apply -f manifests/04-deployment.yaml
```

---

## 开发指南

### 本地构建

```bash
# 安装依赖
go mod download

# 构建二进制
go build -o collect-agent ./main.go

# 或使用 Makefile
make build
```

### 本地运行

```bash
# 1. 准备配置文件
cp config.example.yaml config.yaml
# 编辑 config.yaml 设置 NATS 地址

# 2. 运行 Agent (需要 kubeconfig)
./collect-agent --config=config.yaml

# 或使用 Makefile
make run
```

### 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
make test-integration

# 代码覆盖率
make test-coverage
```

### 代码格式化

```bash
# 格式化代码
make fmt

# 代码检查
make lint
make vet
```

### 项目结构

```
collect-agent/
├── cmd/
│   └── agent/              # 主程序入口
├── internal/
│   ├── agent/              # Agent 核心逻辑
│   ├── config/             # 配置管理
│   ├── communication/      # NATS 通信
│   ├── watcher/            # 事件监控
│   ├── collector/          # 指标采集
│   └── executor/           # 命令执行
├── pkg/
│   └── types/              # 公共类型定义
├── manifests/              # Kubernetes 清单
├── Dockerfile              # 容器镜像定义
├── Makefile                # 构建脚本
└── README.md               # 本文档
```

---

## 监控运维

### 健康检查端点

Agent 提供以下健康检查端点 (端口 8080):

| 端点 | 说明 | 用途 |
|------|------|------|
| `GET /health/live` | 存活探针 | Kubernetes Liveness Probe |
| `GET /health/ready` | 就绪探针 | Kubernetes Readiness Probe |
| `GET /health/status` | 详细状态 | 监控和调试 |
| `GET /metrics` | Prometheus 指标 | 监控指标 |

### 健康状态响应

```json
{
  "status": "healthy",
  "cluster_id": "prod-us-west-2",
  "version": "v1.0.0",
  "uptime": "2h30m15s",
  "nats": {
    "connected": true,
    "url": "nats://agent-manager:4222",
    "reconnects": 0
  },
  "components": {
    "event_watcher": "running",
    "metrics_collector": "running",
    "command_executor": "ready"
  },
  "queues": {
    "event_queue": 5,
    "metrics_queue": 2,
    "command_queue": 0
  },
  "statistics": {
    "events_sent": 1250,
    "metrics_sent": 150,
    "commands_executed": 25
  }
}
```

### Prometheus 指标

```prometheus
# Agent 状态
agent_running{cluster_id="prod"} 1
agent_connected{cluster_id="prod"} 1
agent_uptime_seconds{cluster_id="prod"} 9015

# 队列大小
agent_event_queue_size{cluster_id="prod"} 5
agent_metrics_queue_size{cluster_id="prod"} 2
agent_command_queue_size{cluster_id="prod"} 0

# 统计信息
agent_events_sent_total{cluster_id="prod"} 1250
agent_metrics_sent_total{cluster_id="prod"} 150
agent_commands_executed_total{cluster_id="prod"} 25
agent_errors_total{cluster_id="prod",type="nats"} 2
```

### 日志格式

Agent 使用结构化 JSON 日志:

```json
{
  "timestamp": "2025-10-01T10:00:00Z",
  "level": "info",
  "logger": "event-watcher",
  "message": "Critical event detected",
  "cluster_id": "prod-us-west-2",
  "event": {
    "reason": "OOMKilled",
    "namespace": "production",
    "pod": "api-server-xyz"
  }
}
```

---

## NATS 消息主题

Agent 使用以下 NATS 主题进行通信:

### 发布主题 (Agent → Agent Manager)

| 主题 | 说明 | 频率 |
|------|------|------|
| `aetherius.agent.{cluster_id}.register` | Agent 注册 | 启动时 |
| `aetherius.agent.{cluster_id}.heartbeat` | 心跳消息 | 每 30 秒 |
| `aetherius.agent.{cluster_id}.event` | 事件上报 | 实时 |
| `aetherius.agent.{cluster_id}.metrics` | 指标上报 | 每 60 秒 |
| `aetherius.agent.{cluster_id}.result` | 命令结果 | 按需 |

### 订阅主题 (Agent Manager → Agent)

| 主题 | 说明 |
|------|------|
| `aetherius.agent.{cluster_id}.command` | 接收命令 |

---

## 安全设计

### 容器安全

- ✅ 非 root 用户运行 (UID: 65534, GID: 65534)
- ✅ 只读根文件系统
- ✅ 禁用特权模式
- ✅ 丢弃所有 Linux Capabilities
- ✅ 禁用特权升级

### RBAC 权限

Agent 仅需要以下最小权限:

```yaml
rules:
- apiGroups: [""]
  resources: ["events", "pods", "nodes", "namespaces"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]
```

### 命令白名单

仅允许执行以下安全命令:

- `kubectl get`
- `kubectl describe`
- `kubectl logs`
- `kubectl top`
- `ps aux`
- `df -h`
- `netstat -an`

**禁止**: `kubectl delete`, `kubectl edit`, `kubectl apply` 等写操作。

---

## 故障排查

### 常见问题

#### 1. Agent 无法连接到 NATS

**症状**:
```
ERROR Failed to connect to NATS: dial tcp: connection refused
```

**排查步骤**:
```bash
# 1. 检查 NATS 服务器地址
kubectl -n aetherius-agent get cm aetherius-agent-config -o yaml

# 2. 测试网络连通性
kubectl -n aetherius-agent exec deployment/aetherius-agent -- \
  nc -zv agent-manager 4222

# 3. 检查 NATS 服务器状态
kubectl -n aetherius get pods -l app=nats
kubectl -n aetherius logs deployment/nats
```

**解决方案**:
- 确认 `central_endpoint` 配置正确
- 检查网络策略和防火墙规则
- 确认 NATS 服务器正常运行

#### 2. 没有事件上报

**症状**: Agent 正常运行,但 Agent Manager 未收到事件。

**排查步骤**:
```bash
# 1. 检查 Agent 日志
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep -i event

# 2. 检查 RBAC 权限
kubectl -n aetherius-agent get sa aetherius-agent -o yaml
kubectl get clusterrole aetherius-agent -o yaml

# 3. 检查事件过滤配置
kubectl -n aetherius-agent describe cm aetherius-agent-config

# 4. 手动触发测试事件
kubectl run test-pod --image=busybox --restart=Never -- sh -c "exit 1"
```

**解决方案**:
- 检查 RBAC 权限是否正确
- 确认事件过滤规则未过度严格
- 查看 Agent 日志中的错误信息

#### 3. 内存占用过高

**症状**: Agent Pod 内存使用超过 512MB。

**排查步骤**:
```bash
# 1. 检查内存使用
kubectl -n aetherius-agent top pod

# 2. 查看队列大小
curl http://localhost:8080/health/status | jq .queues

# 3. 检查是否有事件洪泛
kubectl -n aetherius-agent logs deployment/aetherius-agent | \
  grep "queue full" | wc -l
```

**解决方案**:
- 减小 `buffer_size` 配置
- 增加内存限制
- 检查是否有大量重复事件
- 优化事件过滤规则

#### 4. Agent 频繁重启

**症状**: Pod 状态显示 CrashLoopBackOff。

**排查步骤**:
```bash
# 1. 查看 Pod 事件
kubectl -n aetherius-agent describe pod <pod-name>

# 2. 查看容器日志
kubectl -n aetherius-agent logs <pod-name> --previous

# 3. 检查资源限制
kubectl -n aetherius-agent get pod <pod-name> -o yaml | \
  grep -A 10 resources
```

**常见原因**:
- OOMKilled: 增加内存限制
- kubeconfig 无效: 检查 ServiceAccount
- NATS 连接失败: 增加 `max_retries`

### 调试模式

启用 debug 日志级别:

```bash
# 临时启用
kubectl -n aetherius-agent set env deployment/aetherius-agent \
  LOG_LEVEL=debug

# 永久修改
kubectl -n aetherius-agent edit cm aetherius-agent-config
# 修改 log_level: "debug"
kubectl -n aetherius-agent rollout restart deployment/aetherius-agent
```

### 查看详细状态

```bash
# 获取 Pod 状态
kubectl -n aetherius-agent get pods -o wide

# 查看实时日志
kubectl -n aetherius-agent logs deployment/aetherius-agent --follow

# 检查健康状态
kubectl -n aetherius-agent port-forward service/aetherius-agent 8080:8080
curl http://localhost:8080/health/status | jq

# 查看 Prometheus 指标
curl http://localhost:8080/metrics
```

---

## 性能调优

### 资源配置建议

| 场景 | CPU 请求 | 内存请求 | CPU 限制 | 内存限制 |
|------|---------|---------|---------|---------|
| 小型集群 (< 50 节点) | 50m | 64Mi | 200m | 256Mi |
| 中型集群 (50-200 节点) | 100m | 128Mi | 500m | 512Mi |
| 大型集群 (> 200 节点) | 200m | 256Mi | 1000m | 1Gi |

### 配置优化

```yaml
# 高负载场景
buffer_size: 2000              # 增加缓冲区
heartbeat_interval: 60s        # 降低心跳频率
metrics_interval: 120s         # 降低指标采集频率

# 低延迟场景
buffer_size: 500               # 减小缓冲区
heartbeat_interval: 15s        # 提高心跳频率
metrics_interval: 30s          # 提高指标采集频率
```

---

## 许可证

本项目是 Aetherius 智能运维平台的一部分,采用 MIT License 开源。

---

## 相关文档

- [Agent Manager 文档](../agent-manager/README.md)
- [系统架构文档](../docs/architecture/SYSTEM_ARCHITECTURE.md)
- [API 参考文档](../docs/api/API_REFERENCE.md)
- [部署指南](../deployments/k8s/README.md)
