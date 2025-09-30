# Aetherius Collect Agent - 快速开始指南

本指南将帮助你在 5 分钟内部署 Aetherius Collect Agent 到 Kubernetes 集群。

## 前置条件

- ✅ Kubernetes 集群 (v1.20+)
- ✅ kubectl 已配置并能访问集群
- ✅ NATS 服务器运行在中央控制平面
- ✅ (可选) Docker 用于构建镜像

## 方式 1: 使用 Makefile (推荐)

### 步骤 1: 克隆代码

```bash
cd /path/to/k8s-agent/collect-agent
```

### 步骤 2: 查看所有可用命令

```bash
make help
```

### 步骤 3: 生成配置示例

```bash
make config-example
# 编辑 config.example.yaml 修改 NATS 端点
```

### 步骤 4: 部署到 Kubernetes

```bash
# 方式 A: 使用默认配置
make k8s-deploy

# 方式 B: 自定义配置
export CENTRAL_ENDPOINT="nats://your-nats-server:4222"
export CLUSTER_ID="prod-us-west"
make k8s-deploy
```

### 步骤 5: 验证部署

```bash
# 查看状态
make k8s-status

# 查看日志
make k8s-logs

# 检查健康
make k8s-health
```

## 方式 2: 使用部署脚本

### 快速部署

```bash
# 使用脚本部署(交互式)
./scripts/deploy.sh

# 或指定参数
./scripts/deploy.sh \
  --cluster-id prod-us-west \
  --central-endpoint nats://nats.prod.com:4222 \
  --image-tag v1.0.0
```

### 脚本选项

```bash
./scripts/deploy.sh --help

Options:
  --cluster-id ID           设置集群 ID (不指定则自动检测)
  --central-endpoint URL    设置 NATS 端点
  --image-tag TAG           设置镜像标签 (默认: v1.0.0)
  --dry-run                 预演模式,不实际部署
  --skip-confirm            跳过确认提示
```

## 方式 3: 手动部署

### 步骤 1: 修改配置

编辑 `manifests/03-configmap.yaml`:

```yaml
data:
  config.yaml: |
    cluster_id: "prod-us-west"  # 或留空自动检测
    central_endpoint: "nats://your-nats-server:4222"
    reconnect_delay: 5s
    heartbeat_interval: 30s
    metrics_interval: 60s
    enable_metrics: true
    enable_events: true
    log_level: "info"
```

### 步骤 2: 应用 Manifests

```bash
kubectl apply -f manifests/01-namespace.yaml
kubectl apply -f manifests/02-rbac.yaml
kubectl apply -f manifests/03-configmap.yaml
kubectl apply -f manifests/04-deployment.yaml
```

### 步骤 3: 等待就绪

```bash
kubectl -n aetherius-agent rollout status deployment/aetherius-agent
```

## 验证部署

### 1. 检查 Pod 状态

```bash
kubectl -n aetherius-agent get pods
```

期望输出:
```
NAME                               READY   STATUS    RESTARTS   AGE
aetherius-agent-xxxxxxxxxx-xxxxx   1/1     Running   0          1m
```

### 2. 查看日志

```bash
kubectl -n aetherius-agent logs deployment/aetherius-agent -f
```

正常日志示例:
```json
{"timestamp":"2025-09-30T10:00:00Z","level":"info","message":"Starting Aetherius Collect Agent","cluster_id":"prod-us-west"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Connected to NATS","url":"nats://..."}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Agent registered","cluster_id":"prod-us-west"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Event watcher started"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Agent started successfully"}
```

### 3. 检查健康状态

```bash
# 端口转发
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &

# 检查健康
curl http://localhost:8080/health/status | jq .
```

期望输出:
```json
{
  "cluster_id": "prod-us-west",
  "running": true,
  "start_time": "2025-09-30T10:00:00Z",
  "uptime": 300000000000,
  "event_queue_size": 0,
  "metrics_queue_size": 0,
  "command_queue_size": 0,
  "result_queue_size": 0,
  "connected": true
}
```

### 4. 查看 Prometheus 指标

```bash
curl http://localhost:8080/metrics
```

## 本地开发

### 构建和运行

```bash
# 1. 下载依赖
make deps

# 2. 运行测试
make test

# 3. 本地构建
make build

# 4. 本地运行(需要 kubeconfig)
export CENTRAL_ENDPOINT="nats://localhost:4222"
./build/collect-agent --config=config.example.yaml
```

### Docker 本地测试

```bash
# 1. 构建镜像
make docker-build

# 2. 运行容器
make docker-run
```

## 常见问题

### 1. Agent 无法连接到 NATS

**症状**: 日志显示 "failed to connect to NATS"

**解决**:
```bash
# 检查 NATS 端点配置
kubectl -n aetherius-agent get configmap agent-config -o yaml

# 测试网络连通性
kubectl -n aetherius-agent exec deployment/aetherius-agent -- nc -zv nats-server 4222

# 检查 NATS 服务器状态
kubectl -n nats-system get pods  # 如果 NATS 在 K8s 中
```

### 2. Agent 无法检测集群 ID

**症状**: 日志显示 "failed to detect cluster ID"

**解决**:
```bash
# 方式 A: 手动指定集群 ID
kubectl -n aetherius-agent set env deployment/aetherius-agent CLUSTER_ID=my-cluster

# 方式 B: 修改 ConfigMap
kubectl -n aetherius-agent edit configmap agent-config
# 设置 cluster_id: "my-cluster"

# 重启 Agent
kubectl -n aetherius-agent rollout restart deployment/aetherius-agent
```

### 3. Pod 处于 CrashLoopBackOff

**症状**: Pod 持续重启

**解决**:
```bash
# 查看 Pod 详情
kubectl -n aetherius-agent describe pod <pod-name>

# 查看容器日志
kubectl -n aetherius-agent logs <pod-name> --previous

# 常见原因:
# - 配置错误: 检查 ConfigMap
# - RBAC 权限不足: 检查 ServiceAccount
# - NATS 无法访问: 检查网络
```

### 4. 事件没有上报

**症状**: Agent 运行正常但没有事件

**解决**:
```bash
# 检查事件监听是否启用
kubectl -n aetherius-agent get configmap agent-config -o yaml | grep enable_events

# 检查 K8s 事件
kubectl get events --all-namespaces | head -20

# 查看 Agent 日志中的事件处理
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Event sent"
```

### 5. 高内存使用

**症状**: Agent 内存占用超过 256Mi

**解决**:
```bash
# 检查队列大小
curl http://localhost:8080/health/status | jq '.event_queue_size, .metrics_queue_size'

# 调整缓冲区大小
kubectl -n aetherius-agent edit configmap agent-config
# 减小 buffer_size: 500

# 增加资源限制
kubectl -n aetherius-agent edit deployment aetherius-agent
# 调整 resources.limits.memory: 512Mi
```

## 监控和告警

### Prometheus 指标

Agent 暴露以下 Prometheus 指标:

```
agent_running{cluster_id="xxx"}               # Agent 运行状态
agent_connected{cluster_id="xxx"}             # NATS 连接状态
agent_uptime_seconds{cluster_id="xxx"}        # 运行时长
agent_event_queue_size{cluster_id="xxx"}      # 事件队列大小
agent_metrics_queue_size{cluster_id="xxx"}    # 指标队列大小
agent_command_queue_size{cluster_id="xxx"}    # 命令队列大小
agent_result_queue_size{cluster_id="xxx"}     # 结果队列大小
```

### 推荐的告警规则

```yaml
groups:
  - name: aetherius-agent
    rules:
      - alert: AgentDown
        expr: agent_running == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Agent is down on cluster {{ $labels.cluster_id }}"

      - alert: AgentDisconnected
        expr: agent_connected == 0
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Agent disconnected from NATS on cluster {{ $labels.cluster_id }}"

      - alert: HighEventQueue
        expr: agent_event_queue_size > 800
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High event queue size on cluster {{ $labels.cluster_id }}"
```

## 配置调优

### 高负载集群

对于大规模集群(1000+ nodes):

```yaml
# config.yaml
buffer_size: 5000              # 增大缓冲区
heartbeat_interval: 60s        # 降低心跳频率
metrics_interval: 120s         # 降低指标收集频率
```

### 低延迟需求

对于需要快速响应的场景:

```yaml
# config.yaml
heartbeat_interval: 15s        # 提高心跳频率
metrics_interval: 30s          # 提高指标收集频率
reconnect_delay: 2s            # 加快重连
```

### 资源受限环境

对于资源受限的集群:

```yaml
# config.yaml
buffer_size: 100               # 减小缓冲区
enable_metrics: false          # 禁用指标收集
metrics_interval: 300s         # 降低收集频率

# Deployment
resources:
  requests:
    memory: 64Mi
    cpu: 50m
  limits:
    memory: 128Mi
    cpu: 100m
```

## 升级指南

### 滚动升级

```bash
# 1. 更新镜像标签
export IMAGE_TAG=v1.1.0
make manifests-update

# 2. 应用更新
kubectl apply -f manifests/04-deployment.yaml

# 3. 监控滚动升级
kubectl -n aetherius-agent rollout status deployment/aetherius-agent

# 4. 验证新版本
kubectl -n aetherius-agent get pods -o jsonpath='{.items[0].spec.containers[0].image}'
```

### 回滚

```bash
# 查看历史版本
kubectl -n aetherius-agent rollout history deployment/aetherius-agent

# 回滚到上一版本
kubectl -n aetherius-agent rollout undo deployment/aetherius-agent

# 回滚到指定版本
kubectl -n aetherius-agent rollout undo deployment/aetherius-agent --to-revision=2
```

## 卸载

### 完全删除

```bash
# 使用 Makefile
make k8s-delete

# 或手动删除
kubectl delete -f manifests/

# 确认删除
kubectl get ns aetherius-agent
# 应该显示 "NotFound"
```

## 下一步

- 📖 阅读 [完整文档](./README.md)
- 🏗️ 查看 [架构设计](./IMPLEMENTATION_SUMMARY.md)
- 🔧 了解 [配置选项](./config.example.yaml)
- 🐛 报告 [问题](https://github.com/kart-io/k8s-agent/issues)

## 获取帮助

- 📚 文档: [./README.md](./README.md)
- 💬 讨论: GitHub Discussions
- 🐛 问题: GitHub Issues
- 📧 邮件: support@aetherius.io