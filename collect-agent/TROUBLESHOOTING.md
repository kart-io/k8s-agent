# 故障排查指南

本文档提供 Aetherius Collect Agent 常见问题的诊断和解决方法。

---

## 目录

- [连接问题](#连接问题)
- [性能问题](#性能问题)
- [部署问题](#部署问题)
- [配置问题](#配置问题)
- [日志分析](#日志分析)
- [调试技巧](#调试技巧)

---

## 连接问题

### 问题 1: Agent 无法连接到 NATS

**症状**:
```json
{"level":"error","message":"failed to connect to NATS","error":"dial tcp: connection refused"}
```

**原因分析**:
1. NATS 服务器未运行或不可访问
2. 错误的 NATS 端点配置
3. 网络策略阻止连接
4. NATS 服务器资源不足

**排查步骤**:

```bash
# 1. 检查配置
kubectl -n aetherius-agent get configmap agent-config -o yaml | grep central_endpoint

# 2. 测试网络连通性
kubectl -n aetherius-agent exec deployment/aetherius-agent -- nc -zv nats-server 4222

# 3. 检查 NATS 服务器状态
kubectl -n nats-system get pods
kubectl -n nats-system logs deployment/nats

# 4. 检查 DNS 解析
kubectl -n aetherius-agent exec deployment/aetherius-agent -- nslookup nats-server

# 5. 检查网络策略
kubectl get networkpolicies -A
```

**解决方法**:

**方法 A**: 修正 NATS 端点
```bash
kubectl -n aetherius-agent edit configmap agent-config
# 修改 central_endpoint: "nats://correct-nats-server:4222"

kubectl -n aetherius-agent rollout restart deployment/aetherius-agent
```

**方法 B**: 确保 NATS 服务运行
```bash
# 检查 NATS 服务
kubectl -n nats-system get svc nats

# 如果不存在,部署 NATS
kubectl apply -f https://raw.githubusercontent.com/nats-io/k8s/master/nats-server/simple-nats.yml
```

**方法 C**: 允许网络访问
```yaml
# networkpolicy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-nats
  namespace: aetherius-agent
spec:
  podSelector:
    matchLabels:
      app: aetherius-agent
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: nats-system
    ports:
    - protocol: TCP
      port: 4222
```

### 问题 2: Agent 频繁重连

**症状**:
```json
{"level":"warn","message":"Disconnected from NATS"}
{"level":"info","message":"Reconnected to NATS"}
```

**原因分析**:
1. NATS 服务器不稳定
2. 网络抖动
3. NATS 连接数达到上限
4. Agent 资源不足导致超时

**排查步骤**:

```bash
# 1. 查看重连频率
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep -E "Disconnected|Reconnected"

# 2. 检查 Agent 资源使用
kubectl -n aetherius-agent top pods

# 3. 检查 NATS 连接数
kubectl -n nats-system exec deployment/nats -- nats-server -sl | grep connections

# 4. 查看网络延迟
kubectl -n aetherius-agent exec deployment/aetherius-agent -- ping -c 5 nats-server
```

**解决方法**:

**方法 A**: 调整重连参数
```yaml
# configmap
reconnect_delay: 10s  # 增加重连延迟
max_retries: 30       # 增加重试次数
```

**方法 B**: 增加 NATS 连接限制
```yaml
# nats-config
max_connections: 10000  # 增加最大连接数
```

**方法 C**: 增加 Agent 资源
```yaml
# deployment
resources:
  requests:
    memory: 256Mi
    cpu: 200m
  limits:
    memory: 512Mi
    cpu: 500m
```

---

## 性能问题

### 问题 3: Agent 内存占用过高

**症状**:
```bash
NAME                              READY   STATUS    RESTARTS   AGE
aetherius-agent-xxx   1/1     Running   0          1h
# Memory: 400Mi / 256Mi (156%)
```

**原因分析**:
1. 事件缓冲区过大
2. 事件处理速度慢于生成速度
3. 内存泄漏
4. 指标收集过于频繁

**排查步骤**:

```bash
# 1. 检查队列大小
curl http://localhost:8080/health/status | jq '.event_queue_size, .metrics_queue_size'

# 2. 查看内存使用趋势
kubectl -n aetherius-agent top pod --watch

# 3. 检查 goroutine 泄漏
kubectl -n aetherius-agent port-forward svc/aetherius-agent 6060:6060 &
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 4. 生成内存 profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

**解决方法**:

**方法 A**: 减小缓冲区
```yaml
buffer_size: 500  # 从 1000 减小到 500
```

**方法 B**: 降低收集频率
```yaml
heartbeat_interval: 60s   # 从 30s 增加到 60s
metrics_interval: 120s    # 从 60s 增加到 120s
```

**方法 C**: 增加内存限制
```yaml
resources:
  limits:
    memory: 512Mi  # 从 256Mi 增加
```

**方法 D**: 禁用非必要功能
```yaml
enable_metrics: false  # 暂时禁用指标收集
```

### 问题 4: Agent CPU 使用率高

**症状**:
Agent CPU 使用率持续超过 80%

**原因分析**:
1. 事件处理量过大
2. 频繁的 K8s API 调用
3. JSON 序列化开销
4. 日志级别过低(debug)

**排查步骤**:

```bash
# 1. CPU profiling
kubectl -n aetherius-agent port-forward svc/aetherius-agent 6060:6060 &
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 2. 检查事件处理速率
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Event sent" | wc -l

# 3. 查看 K8s API 调用频率
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Failed to list"
```

**解决方法**:

**方法 A**: 优化日志级别
```yaml
log_level: "info"  # 从 debug 改为 info
```

**方法 B**: 增加事件过滤
修改事件过滤规则,只上报最关键的事件

**方法 C**: 调整资源限制
```yaml
resources:
  requests:
    cpu: 200m
  limits:
    cpu: 500m
```

---

## 部署问题

### 问题 5: Pod 处于 CrashLoopBackOff

**症状**:
```bash
NAME                              READY   STATUS             RESTARTS   AGE
aetherius-agent-xxx   0/1     CrashLoopBackOff   5          3m
```

**排查步骤**:

```bash
# 1. 查看 Pod 详情
kubectl -n aetherius-agent describe pod <pod-name>

# 2. 查看容器日志
kubectl -n aetherius-agent logs <pod-name>

# 3. 查看之前容器的日志
kubectl -n aetherius-agent logs <pod-name> --previous

# 4. 检查事件
kubectl -n aetherius-agent get events --sort-by='.lastTimestamp'
```

**常见原因和解决**:

**原因 A**: 配置错误
```bash
# 检查配置
kubectl -n aetherius-agent get configmap agent-config -o yaml

# 验证配置格式
kubectl -n aetherius-agent create configmap test-config --from-file=config.yaml --dry-run=client
```

**原因 B**: RBAC 权限不足
```bash
# 检查 ServiceAccount
kubectl -n aetherius-agent get sa aetherius-agent

# 检查 ClusterRoleBinding
kubectl get clusterrolebinding | grep aetherius-agent

# 重新应用 RBAC
kubectl apply -f manifests/02-rbac.yaml
```

**原因 C**: 镜像拉取失败
```bash
# 检查镜像
kubectl -n aetherius-agent describe pod <pod-name> | grep -A 5 "Failed to pull image"

# 使用本地镜像
docker save aetherius/collect-agent:v1.0.0 | kubectl -n aetherius-agent load
```

### 问题 6: Readiness 探针失败

**症状**:
```bash
Readiness probe failed: HTTP probe failed with statuscode: 503
```

**排查步骤**:

```bash
# 1. 检查健康端点
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &
curl -v http://localhost:8080/health/ready

# 2. 查看详细状态
curl http://localhost:8080/health/status | jq .

# 3. 检查 Agent 是否连接到 NATS
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Connected to NATS"
```

**解决方法**:

**方法 A**: 增加初始延迟
```yaml
readinessProbe:
  initialDelaySeconds: 10  # 从 5 增加到 10
  periodSeconds: 10        # 从 5 增加到 10
```

**方法 B**: 修复 NATS 连接问题
参考 [问题 1](#问题-1-agent-无法连接到-nats)

---

## 配置问题

### 问题 7: 集群 ID 自动检测失败

**症状**:
```json
{"level":"error","message":"failed to detect cluster ID"}
```

**解决方法**:

**方法 A**: 手动指定集群 ID
```bash
kubectl -n aetherius-agent set env deployment/aetherius-agent CLUSTER_ID=my-cluster-id
```

**方法 B**: 通过 ConfigMap 指定
```yaml
# configmap
cluster_id: "prod-us-west"
```

**方法 C**: 使用云平台标签
确保节点有正确的云平台标签:
- EKS: `eks.amazonaws.com/cluster-name`
- GKE: `cloud.google.com/gke-cluster-name`
- AKS: `kubernetes.azure.com/cluster`

### 问题 8: 事件没有上报

**症状**:
Agent 运行正常,但中央没有收到事件

**排查步骤**:

```bash
# 1. 检查事件监听是否启用
kubectl -n aetherius-agent get configmap agent-config -o yaml | grep enable_events

# 2. 检查 K8s 是否有事件
kubectl get events --all-namespaces | head -20

# 3. 查看 Agent 日志
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Event sent"

# 4. 检查队列状态
curl http://localhost:8080/health/status | jq '.event_queue_size'
```

**解决方法**:

**方法 A**: 启用事件监听
```yaml
enable_events: true
```

**方法 B**: 检查事件过滤规则
某些事件可能被过滤掉了,降低过滤阈值或查看代码中的过滤逻辑

**方法 C**: 增加日志级别诊断
```yaml
log_level: "debug"
```

---

## 日志分析

### 正常日志模式

```json
{"timestamp":"2025-09-30T10:00:00Z","level":"info","message":"Starting Aetherius Collect Agent"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Connected to NATS","url":"nats://..."}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Agent registered","cluster_id":"prod-us-west"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Event watcher started"}
{"timestamp":"2025-09-30T10:00:01Z","level":"info","message":"Agent started successfully"}
```

### 异常日志模式

**错误 1**: 配置错误
```json
{"level":"error","message":"invalid configuration","error":"central_endpoint is required"}
```

**错误 2**: 权限不足
```json
{"level":"error","message":"Failed to list nodes","error":"forbidden: User cannot list nodes"}
```

**错误 3**: 资源不足
```json
{"level":"error","message":"Failed to publish event","error":"channel full"}
```

### 日志分析命令

```bash
# 统计错误数量
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep -c '"level":"error"'

# 查看最近的错误
kubectl -n aetherius-agent logs deployment/aetherius-agent --tail=100 | grep '"level":"error"'

# 分析事件上报率
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep "Event sent" | wc -l

# 查看重连历史
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep -E "Disconnected|Reconnected"
```

---

## 调试技巧

### 技巧 1: 启用调试模式

```bash
# 临时启用 debug 日志
kubectl -n aetherius-agent set env deployment/aetherius-agent LOG_LEVEL=debug

# 查看详细日志
kubectl -n aetherius-agent logs deployment/aetherius-agent -f
```

### 技巧 2: 使用 port-forward 调试

```bash
# 转发健康检查端口
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &

# 查看实时状态
watch -n 1 'curl -s http://localhost:8080/health/status | jq .'

# 查看 Prometheus 指标
curl http://localhost:8080/metrics
```

### 技巧 3: 在本地运行 Agent

```bash
# 1. 导出 kubeconfig
export KUBECONFIG=~/.kube/config

# 2. 设置环境变量
export CENTRAL_ENDPOINT="nats://localhost:4222"
export LOG_LEVEL="debug"

# 3. 本地运行
go run ./main.go --config=config.example.yaml
```

### 技巧 4: 使用 tcpdump 抓包

```bash
# 抓取 NATS 通信
kubectl -n aetherius-agent exec deployment/aetherius-agent -- tcpdump -i any -w /tmp/nats.pcap port 4222

# 下载抓包文件
kubectl cp aetherius-agent/<pod-name>:/tmp/nats.pcap ./nats.pcap

# 使用 Wireshark 分析
wireshark nats.pcap
```

### 技巧 5: 性能分析

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存 profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 生成火焰图
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## 获取帮助

如果以上方法都无法解决问题:

1. **收集诊断信息**:
```bash
# 生成诊断报告
kubectl -n aetherius-agent describe pod <pod-name> > pod-describe.txt
kubectl -n aetherius-agent logs <pod-name> > agent-logs.txt
kubectl -n aetherius-agent get events > events.txt
curl http://localhost:8080/health/status > health-status.json
```

2. **联系支持**:
- GitHub Issues: https://github.com/kart-io/k8s-agent/issues
- 邮件: support@aetherius.io
- 文档: ./README.md

3. **社区资源**:
- 常见问题: ./FAQ.md
- 快速开始: ./QUICKSTART.md
- 实现文档: ./IMPLEMENTATION_SUMMARY.md