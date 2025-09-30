# 监控配置指南

本目录包含 Aetherius Collect Agent 的完整监控配置,包括 Prometheus 规则和 Grafana 仪表板。

---

## 目录

- [文件说明](#文件说明)
- [Prometheus 配置](#prometheus-配置)
- [Grafana 配置](#grafana-配置)
- [监控指标说明](#监控指标说明)
- [告警规则说明](#告警规则说明)
- [快速开始](#快速开始)
- [自定义配置](#自定义配置)

---

## 文件说明

```plaintext
monitoring/
├── prometheus-rules.yaml      # Prometheus 记录规则和告警规则
├── grafana-dashboard.json     # Grafana 仪表板配置
└── README.md                  # 本文档
```

---

## Prometheus 配置

### 部署 Prometheus 规则

#### 方法 1: 使用 kubectl

```bash
# 创建规则 ConfigMap
kubectl apply -f monitoring/prometheus-rules.yaml

# 配置 Prometheus 加载规则
kubectl -n monitoring edit configmap prometheus-config
```

在 Prometheus 配置中添加:

```yaml
rule_files:
  - /etc/prometheus/rules/agent.rules
```

#### 方法 2: 使用 Prometheus Operator

```bash
# 创建 PrometheusRule 资源
cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: aetherius-agent-rules
  namespace: monitoring
spec:
  groups:
  - name: aetherius_agent_recording_rules
    interval: 30s
    rules:
    # ... (从 prometheus-rules.yaml 复制规则)
EOF
```

#### 方法 3: Prometheus 配置文件

如果使用独立 Prometheus 部署:

```bash
# 复制规则文件到 Prometheus 配置目录
cp monitoring/prometheus-rules.yaml /etc/prometheus/rules/

# 重新加载 Prometheus 配置
curl -X POST http://prometheus:9090/-/reload
```

### 验证规则加载

```bash
# 检查规则状态
curl http://prometheus:9090/api/v1/rules | jq '.data.groups[] | select(.name | contains("aetherius"))'

# 检查告警状态
curl http://prometheus:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.component == "aetherius-agent")'
```

---

## Grafana 配置

### 导入仪表板

#### 方法 1: 通过 Grafana UI

1. 登录 Grafana
2. 点击左侧 `+` → `Import`
3. 上传 `grafana-dashboard.json` 文件
4. 选择 Prometheus 数据源
5. 点击 `Import`

#### 方法 2: 通过 API

```bash
# 设置 Grafana 凭据
export GRAFANA_URL="http://grafana:3000"
export GRAFANA_API_KEY="your-api-key"

# 导入仪表板
curl -X POST "${GRAFANA_URL}/api/dashboards/db" \
  -H "Authorization: Bearer ${GRAFANA_API_KEY}" \
  -H "Content-Type: application/json" \
  -d @monitoring/grafana-dashboard.json
```

#### 方法 3: 使用 ConfigMap (Grafana Operator)

```bash
# 创建 ConfigMap
kubectl create configmap aetherius-agent-dashboard \
  --from-file=dashboard.json=monitoring/grafana-dashboard.json \
  -n monitoring

# 为 ConfigMap 添加标签 (Grafana Operator 会自动发现)
kubectl label configmap aetherius-agent-dashboard \
  grafana_dashboard=1 \
  -n monitoring
```

### 配置数据源

确保 Grafana 已配置 Prometheus 数据源:

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

---

## 监控指标说明

### 核心指标

#### 1. aetherius_agent_nats_connected

- **类型**: Gauge
- **说明**: NATS 连接状态 (0=断开, 1=已连接)
- **标签**: `cluster_id`
- **用途**: 监控 Agent 与中央服务器的连接状态

#### 2. aetherius_agent_events_sent_total

- **类型**: Counter
- **说明**: 已发送事件总数
- **标签**: `cluster_id`, `status` (success/failure)
- **用途**: 统计事件发送量和成功率

#### 3. aetherius_agent_metrics_sent_total

- **类型**: Counter
- **说明**: 已发送指标总数
- **标签**: `cluster_id`, `status` (success/failure)
- **用途**: 统计指标上报量和成功率

#### 4. aetherius_agent_commands_executed_total

- **类型**: Counter
- **说明**: 已执行命令总数
- **标签**: `cluster_id`, `command_type`, `status` (success/failure)
- **用途**: 统计命令执行情况

#### 5. aetherius_agent_event_queue_size

- **类型**: Gauge
- **说明**: 当前事件队列大小
- **标签**: `cluster_id`
- **用途**: 监控队列堆积情况

#### 6. aetherius_agent_last_heartbeat_timestamp

- **类型**: Gauge
- **说明**: 最后一次心跳时间戳 (Unix 时间)
- **标签**: `cluster_id`
- **用途**: 检测 Agent 活跃状态

#### 7. aetherius_agent_last_metrics_timestamp

- **类型**: Gauge
- **说明**: 最后一次指标收集时间戳 (Unix 时间)
- **标签**: `cluster_id`
- **用途**: 检测指标收集是否正常

#### 8. aetherius_agent_start_timestamp

- **类型**: Gauge
- **说明**: Agent 启动时间戳 (Unix 时间)
- **标签**: `cluster_id`
- **用途**: 计算运行时长,检测频繁重启

### 记录规则

#### aetherius:agent:events_rate_1m

- **表达式**: `rate(aetherius_agent_events_sent_total[1m]) * 60`
- **说明**: 事件发送速率 (每分钟)
- **用途**: 快速查询事件吞吐量

#### aetherius:agent:metrics_rate_1m

- **表达式**: `rate(aetherius_agent_metrics_sent_total[1m]) * 60`
- **说明**: 指标发送速率 (每分钟)
- **用途**: 快速查询指标吞吐量

#### aetherius:agent:command_success_rate_5m

- **表达式**: `sum(rate(...success...)) / sum(rate(...))`
- **说明**: 命令执行成功率 (5分钟平均)
- **用途**: 监控命令执行质量

#### aetherius:agent:avg_queue_size_5m

- **表达式**: `avg_over_time(aetherius_agent_event_queue_size[5m])`
- **说明**: 平均队列大小 (5分钟)
- **用途**: 监控队列压力趋势

---

## 告警规则说明

### 关键告警

#### 1. AetheriusAgentNATSDisconnected

- **级别**: Critical
- **条件**: NATS 连接断开超过 1 分钟
- **影响**: Agent 无法上报数据和接收命令
- **处理**:
  1. 检查 NATS 服务器状态
  2. 检查网络连通性
  3. 查看 Agent 日志

```bash
kubectl -n aetherius-agent logs deployment/aetherius-agent | grep NATS
```

#### 2. AetheriusAgentEventQueueHigh

- **级别**: Warning
- **条件**: 事件队列大小超过 800 (默认 1000)
- **影响**: 可能导致事件丢失
- **处理**:
  1. 检查 NATS 连接是否正常
  2. 增加 buffer_size 配置
  3. 降低事件生成频率

#### 3. AetheriusAgentCommandFailureRateHigh

- **级别**: Warning
- **条件**: 命令失败率超过 10% (5分钟平均)
- **影响**: 诊断功能受影响
- **处理**:
  1. 检查 RBAC 权限
  2. 查看命令执行日志
  3. 验证命令参数合法性

#### 4. AetheriusAgentHeartbeatMissing

- **级别**: Critical
- **条件**: 超过 2 分钟未发送心跳
- **影响**: Agent 可能已停止工作
- **处理**:
  1. 检查 Pod 状态
  2. 查看容器日志
  3. 检查资源限制是否触发 OOMKilled

```bash
kubectl -n aetherius-agent get pods
kubectl -n aetherius-agent describe pod <pod-name>
```

#### 5. AetheriusAgentEventSendFailureHigh

- **级别**: Warning
- **条件**: 每分钟事件发送失败超过 1 个
- **影响**: 部分事件未上报
- **处理**:
  1. 检查 NATS 服务器负载
  2. 检查网络延迟
  3. 增加 reconnect_delay

#### 6. AetheriusAgentFrequentRestarts

- **级别**: Warning
- **条件**: 30 分钟内重启超过 3 次
- **影响**: 数据收集不连续
- **处理**:
  1. 查看崩溃日志
  2. 检查内存限制
  3. 检查 NATS 连接稳定性

#### 7. AetheriusAgentMetricsCollectionSlow

- **级别**: Warning
- **条件**: 超过 3 分钟未收集指标
- **影响**: 指标数据不完整
- **处理**:
  1. 检查 K8s API 响应时间
  2. 增加 metrics_interval
  3. 检查 Agent CPU 使用率

---

## 快速开始

### 完整部署流程

```bash
# 1. 部署 Agent
kubectl apply -f manifests/

# 2. 等待 Agent 启动
kubectl -n aetherius-agent wait --for=condition=ready pod -l app=aetherius-agent --timeout=60s

# 3. 部署 Prometheus 规则
kubectl apply -f monitoring/prometheus-rules.yaml

# 4. 导入 Grafana 仪表板
# (通过 UI 或 API,见上文)

# 5. 验证指标可用
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &
curl http://localhost:8080/metrics | grep aetherius_agent
```

### 访问监控

```bash
# 访问 Prometheus (假设使用 port-forward)
kubectl -n monitoring port-forward svc/prometheus 9090:9090

# 访问 Grafana
kubectl -n monitoring port-forward svc/grafana 3000:3000

# 打开浏览器
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000
```

---

## 自定义配置

### 调整告警阈值

编辑 `prometheus-rules.yaml`:

```yaml
# 调整队列告警阈值
- alert: AetheriusAgentEventQueueHigh
  expr: aetherius_agent_event_queue_size > 500  # 从 800 改为 500
  for: 3m  # 从 5m 改为 3m
```

### 添加自定义面板

1. 在 Grafana 中手动创建面板
2. 导出仪表板 JSON
3. 合并到 `grafana-dashboard.json`

### 添加新的记录规则

```yaml
# 自定义规则示例
- record: my_custom_rule
  expr: your_prometheus_expression
```

### 集成 AlertManager

```yaml
# alertmanager-config.yaml
route:
  group_by: ['alertname', 'cluster_id']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'aetherius-alerts'

receivers:
- name: 'aetherius-alerts'
  webhook_configs:
  - url: 'http://your-webhook-url'
    send_resolved: true
```

---

## 最佳实践

### 1. 数据保留

```yaml
# Prometheus 配置
storage:
  tsdb:
    retention.time: 15d
    retention.size: 50GB
```

### 2. 查询优化

使用记录规则预计算常用查询:

```promql
# 不推荐 (每次查询都计算)
rate(aetherius_agent_events_sent_total[5m])

# 推荐 (使用预计算规则)
aetherius:agent:events_rate_1m
```

### 3. 告警分组

按集群和严重程度分组告警:

```yaml
route:
  group_by: ['cluster_id', 'severity']
  routes:
  - match:
      severity: critical
    receiver: pagerduty
  - match:
      severity: warning
    receiver: slack
```

### 4. 仪表板权限

为不同团队创建不同的仪表板:

- **运维团队**: 完整仪表板,所有集群
- **开发团队**: 只读仪表板,特定集群
- **管理层**: 高层次概览,SLA 指标

---

## 故障排查

### 问题 1: 指标未显示

```bash
# 检查 Agent 是否暴露指标
kubectl -n aetherius-agent port-forward svc/aetherius-agent 8080:8080 &
curl http://localhost:8080/metrics

# 检查 Prometheus 是否抓取了 Agent
curl http://prometheus:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job == "aetherius-agent")'
```

### 问题 2: 告警未触发

```bash
# 检查规则是否加载
curl http://prometheus:9090/api/v1/rules

# 手动测试告警表达式
curl -G http://prometheus:9090/api/v1/query \
  --data-urlencode 'query=aetherius_agent_nats_connected == 0'
```

### 问题 3: 仪表板显示 "No Data"

1. 检查数据源配置
2. 验证查询表达式
3. 检查时间范围
4. 确认 cluster_id 变量值

---

## 参考资源

- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [PromQL 教程](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboard 最佳实践](https://grafana.com/docs/grafana/latest/best-practices/)

---

## 维护

### 定期任务

- **每周**: 检查告警有效性
- **每月**: 审查仪表板,移除无用面板
- **每季度**: 优化 Prometheus 查询性能

### 版本更新

当 Agent 指标变化时:

1. 更新 `prometheus-rules.yaml`
2. 更新 `grafana-dashboard.json`
3. 通知团队检查自定义仪表板
4. 更新本文档

---

如有问题,请参考:

- 主文档: ../README.md
- 故障排查: ../TROUBLESHOOTING.md
- 实现文档: ../IMPLEMENTATION_SUMMARY.md