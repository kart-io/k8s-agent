# Aetherius 监控配置

完整的监控解决方案,包含 Prometheus、Grafana 和 Alertmanager。

---

## 目录

- [架构概览](#架构概览)
- [组件说明](#组件说明)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [告警规则](#告警规则)
- [Grafana 仪表板](#grafana-仪表板)
- [最佳实践](#最佳实践)

---

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     Aetherius Services                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Agent     │  │ Orchestrator │  │  Reasoning   │      │
│  │   Manager    │  │   Service    │  │   Service    │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │               │
│         └──────────────────┼──────────────────┘               │
│                            │ /metrics                         │
└────────────────────────────┼──────────────────────────────────┘
                             │
                  ┌──────────▼──────────┐
                  │    Prometheus       │
                  │  (采集 & 存储)       │
                  └──────────┬──────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
    ┌─────────▼────────┐    │    ┌────────▼────────┐
    │   Alertmanager   │    │    │     Grafana     │
    │  (告警管理)       │    │    │  (可视化)        │
    └─────────┬────────┘    │    └─────────────────┘
              │             │
              │             │
    ┌─────────▼─────────────▼──────┐
    │   通知渠道                    │
    │  - Email                     │
    │  - Slack                     │
    │  - PagerDuty                 │
    │  - Webhook                   │
    └──────────────────────────────┘
```

---

## 组件说明

### Prometheus

**功能**:
- 指标采集和存储
- 告警规则评估
- 时间序列数据库

**采集目标**:
- Agent Manager (端口 8080)
- Orchestrator Service (端口 8081)
- Reasoning Service (端口 8082)
- Collect Agent (边缘部署)
- PostgreSQL (通过 postgres_exporter)
- Redis (通过 redis_exporter)
- NATS (原生支持)
- Neo4j (原生支持)
- 系统指标 (通过 node_exporter)
- 容器指标 (通过 cAdvisor)

**访问地址**: <http://localhost:9090>

### Grafana

**功能**:
- 数据可视化
- 仪表板管理
- 告警可视化

**内置仪表板**:
- Aetherius System Overview
- Service Performance
- Resource Usage
- Business Metrics

**默认凭据**:

```text
用户名: admin
密码: admin
```

**访问地址**: <http://localhost:3000>

### Alertmanager

**功能**:
- 告警路由和分组
- 告警抑制和静默
- 多渠道通知

**通知渠道**:
- Email (SMTP)
- Slack
- PagerDuty
- Webhook

**访问地址**: <http://localhost:9093>

### Exporters

**Node Exporter** (端口 9100):
- 系统 CPU、内存、磁盘、网络指标

**Postgres Exporter** (端口 9187):
- 数据库连接数、查询性能、锁等待

**Redis Exporter** (端口 9121):
- 内存使用、命令统计、键空间信息

**cAdvisor** (端口 8080):
- 容器 CPU、内存、网络、磁盘 I/O

---

## 快速开始

### Docker Compose 部署

#### 1. 启动监控栈

```bash
cd deployments/monitoring
docker-compose -f docker-compose.monitoring.yml up -d
```

#### 2. 验证服务状态

```bash
# 检查所有服务
docker-compose -f docker-compose.monitoring.yml ps

# 查看 Prometheus 日志
docker logs aetherius-prometheus

# 查看 Grafana 日志
docker logs aetherius-grafana
```

#### 3. 访问服务

- **Prometheus**: <http://localhost:9090>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Alertmanager**: <http://localhost:9093>

#### 4. 配置 Grafana

1. 登录 Grafana (admin/admin)
2. 添加 Prometheus 数据源 (已自动配置)
3. 导入仪表板 (已自动导入)

### Kubernetes 部署

#### 1. 创建监控命名空间

```bash
kubectl create namespace monitoring
```

#### 2. 部署 Prometheus

```bash
# 创建 ConfigMap
kubectl create configmap prometheus-config \
  --from-file=prometheus/prometheus.yml \
  -n monitoring

kubectl create configmap prometheus-rules \
  --from-file=prometheus/rules/ \
  -n monitoring

# 部署 Prometheus
kubectl apply -f k8s/prometheus-deployment.yaml
```

#### 3. 部署 Grafana

```bash
# 创建 ConfigMap
kubectl create configmap grafana-datasources \
  --from-file=grafana/datasources/ \
  -n monitoring

kubectl create configmap grafana-dashboards \
  --from-file=grafana/dashboards/ \
  -n monitoring

# 部署 Grafana
kubectl apply -f k8s/grafana-deployment.yaml
```

#### 4. 部署 Alertmanager

```bash
# 创建 ConfigMap
kubectl create configmap alertmanager-config \
  --from-file=alertmanager/alertmanager.yml \
  -n monitoring

# 部署 Alertmanager
kubectl apply -f k8s/alertmanager-deployment.yaml
```

#### 5. 访问服务

```bash
# Port Forward Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# Port Forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Port Forward Alertmanager
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
```

---

## 配置说明

### Prometheus 配置

**文件位置**: `prometheus/prometheus.yml`

**关键配置**:

```yaml
global:
  scrape_interval: 15s       # 采集间隔
  evaluation_interval: 15s   # 规则评估间隔

scrape_configs:
  - job_name: 'agent-manager'
    kubernetes_sd_configs:   # Kubernetes 服务发现
      - role: pod
```

**自定义采集间隔**:

```yaml
scrape_configs:
  - job_name: 'custom-service'
    scrape_interval: 30s     # 覆盖全局配置
    static_configs:
      - targets: ['service:8080']
```

### 告警规则配置

**文件位置**: `prometheus/rules/aetherius-alerts.yml`

**规则结构**:

```yaml
groups:
  - name: service_availability
    interval: 30s
    rules:
      - alert: ServiceDown
        expr: up{job="agent-manager"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "服务不可用"
          description: "详细描述"
```

**告警级别**:

- `critical`: 严重 - 需要立即处理
- `high`: 高 - 1 小时内处理
- `warning`: 警告 - 当天处理
- `info`: 信息 - 记录即可

### Alertmanager 配置

**文件位置**: `alertmanager/alertmanager.yml`

**路由配置**:

```yaml
route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s          # 等待同组告警
  group_interval: 10s      # 同组告警发送间隔
  repeat_interval: 12h     # 重复发送间隔
  receiver: 'default'
```

**Email 通知配置**:

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@example.com'
  smtp_auth_username: 'alerts@example.com'
  smtp_auth_password: 'your_password'
  smtp_require_tls: true

receivers:
  - name: 'email'
    email_configs:
      - to: 'team@example.com'
        headers:
          Subject: '[Aetherius] {{ .GroupLabels.alertname }}'
```

**Slack 通知配置**:

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK'
        channel: '#alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## 告警规则

### 服务可用性

| 告警名称 | 条件 | 严重级别 | 持续时间 |
|---------|------|---------|---------|
| AgentManagerDown | Agent Manager 不响应 | critical | 1m |
| OrchestratorServiceDown | Orchestrator 不响应 | critical | 1m |
| ReasoningServiceDown | Reasoning Service 不响应 | critical | 1m |
| CollectAgentDown | Collect Agent 不响应 | high | 5m |

### 性能指标

| 告警名称 | 条件 | 严重级别 | 持续时间 |
|---------|------|---------|---------|
| HighAPILatency | P95 延迟 > 1s | warning | 5m |
| HighErrorRate | 5xx 错误率 > 5% | warning | 5m |
| HighCPUUsage | CPU 使用率 > 80% | warning | 10m |
| HighMemoryUsage | 内存使用 > 3GB | warning | 10m |

### 业务指标

| 告警名称 | 条件 | 严重级别 | 持续时间 |
|---------|------|---------|---------|
| HighAgentRegistrationFailureRate | 注册失败率 > 10% | warning | 5m |
| HighEventProcessingDelay | 事件处理延迟 > 5s | warning | 5m |
| HighWorkflowFailureRate | 工作流失败率 > 20% | warning | 10m |
| LowRootCauseAccuracy | 根因分析准确率 < 70% | warning | 30m |

### 依赖服务

| 告警名称 | 条件 | 严重级别 | 持续时间 |
|---------|------|---------|---------|
| PostgreSQLDown | PostgreSQL 不响应 | critical | 1m |
| RedisDown | Redis 不响应 | critical | 1m |
| NATSDown | NATS 不响应 | critical | 1m |
| Neo4jDown | Neo4j 不响应 | high | 5m |

---

## Grafana 仪表板

### System Overview

**包含面板**:

1. **Service Status**: 所有服务的健康状态
2. **Request Rate**: 各服务的请求速率
3. **Error Rate**: 4xx/5xx 错误率
4. **API Response Time**: P95 响应时间
5. **Agent Registration**: 已注册 Agent 数量
6. **Active Workflows**: 活跃工作流数量
7. **Events Processed**: 已处理事件数量
8. **Root Cause Accuracy**: 根因分析准确率
9. **Memory Usage**: 各服务内存使用
10. **CPU Usage**: 各服务 CPU 使用
11. **Goroutines**: Go 服务的 Goroutine 数量
12. **NATS Message Rate**: 消息吞吐量

### 自定义仪表板

#### 创建新仪表板

1. 登录 Grafana
2. 点击 "+" → "Dashboard"
3. 添加面板
4. 选择 Prometheus 数据源
5. 输入 PromQL 查询

#### 常用 PromQL 查询

**请求速率**:

```promql
rate(http_requests_total[5m])
```

**错误率**:

```promql
rate(http_requests_total{status=~"5.."}[5m])
  / rate(http_requests_total[5m])
```

**P95 延迟**:

```promql
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket[5m]))
```

**内存使用**:

```promql
process_resident_memory_bytes / 1024 / 1024
```

**CPU 使用率**:

```promql
rate(process_cpu_seconds_total[5m]) * 100
```

---

## 最佳实践

### 1. 告警配置

**避免告警疲劳**:
- 设置合理的阈值
- 使用 `for` 参数避免瞬时波动
- 实施告警分级
- 配置告警抑制规则

**示例**:

```yaml
# 不好的做法 - 容易误报
- alert: HighCPU
  expr: cpu_usage > 50
  for: 10s

# 好的做法
- alert: HighCPU
  expr: cpu_usage > 80
  for: 10m  # 持续 10 分钟
  labels:
    severity: warning
```

### 2. 指标采集

**采集间隔建议**:
- 核心服务: 15s
- 依赖服务: 30s
- 系统指标: 60s

**保留时间建议**:
- 开发环境: 7 天
- 生产环境: 30-90 天

### 3. 仪表板设计

**遵循 RED 方法**:
- **Rate**: 请求速率
- **Errors**: 错误率
- **Duration**: 响应时间

**遵循 USE 方法**:
- **Utilization**: 利用率
- **Saturation**: 饱和度
- **Errors**: 错误

### 4. 容量规划

**Prometheus 存储**:

```text
所需磁盘空间 = 指标数量 × 采集间隔 × 保留时间 × 压缩率
```

**示例计算**:

```text
1000 个指标 × 15s 采集间隔 × 30 天保留 × 1.5 压缩率
= 约 40GB 存储空间
```

### 5. 高可用部署

**Prometheus HA**:

```yaml
# 部署多个 Prometheus 实例
# 使用相同配置,独立采集

# Instance 1
prometheus-1:
  external_labels:
    replica: 1

# Instance 2
prometheus-2:
  external_labels:
    replica: 2
```

**Alertmanager 集群**:

```yaml
# 配置 Alertmanager 集群
alertmanager:
  cluster:
    peers:
      - alertmanager-1:9094
      - alertmanager-2:9094
      - alertmanager-3:9094
```

---

## 故障排查

### Prometheus 无法采集指标

**检查步骤**:

1. 验证目标可达性:

```bash
curl http://agent-manager:8080/metrics
```

2. 检查 Prometheus 目标状态:
   - 访问 <http://localhost:9090/targets>
   - 查看目标状态和错误信息

3. 检查服务发现:

```bash
# Kubernetes 环境
kubectl get pods -n aetherius -o wide
```

### 告警未触发

**检查步骤**:

1. 验证告警规则加载:
   - 访问 <http://localhost:9090/rules>
   - 确认规则已加载

2. 检查表达式:
   - 在 Prometheus UI 执行 PromQL
   - 验证返回结果

3. 检查 Alertmanager 连接:

```bash
curl http://localhost:9093/api/v1/status
```

### Grafana 无数据

**检查步骤**:

1. 验证数据源:
   - Configuration → Data Sources
   - Test 按钮验证连接

2. 检查时间范围:
   - 确保选择的时间范围内有数据

3. 检查 PromQL 语法:
   - 在 Prometheus UI 测试查询

---

## 维护和优化

### 日常维护

**每周**:
- 检查告警规则有效性
- 清理过期的静默规则
- 审查告警通知记录

**每月**:
- 优化告警阈值
- 更新仪表板
- 清理无用指标

**每季度**:
- 容量规划评估
- 性能优化
- 告警规则审计

### 性能优化

**减少指标基数**:

```yaml
# 使用 metric_relabel_configs 过滤不需要的指标
metric_relabel_configs:
  - source_labels: [__name__]
    regex: 'go_.*'
    action: drop
```

**优化查询性能**:

```promql
# 不好的做法 - 范围太大
rate(http_requests_total[1d])

# 好的做法 - 合理的范围
rate(http_requests_total[5m])
```

---

## 参考资料

- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
- [Alertmanager 文档](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [PromQL 查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)

---

## 获取帮助

如有问题:

- 查看 [主文档](../../README.md)
- 搜索 [Issues](https://github.com/kart-io/k8s-agent/issues)
- 提交 [新 Issue](https://github.com/kart-io/k8s-agent/issues/new)