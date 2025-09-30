# Aetherius 性能测试工具

本目录包含 Aetherius 平台的性能测试和基准测试工具。

---

## 目录

- [工具说明](#工具说明)
- [快速开始](#快速开始)
- [负载测试](#负载测试)
- [基准测试](#基准测试)
- [性能指标](#性能指标)
- [优化建议](#优化建议)

---

## 工具说明

### load-test.sh

**用途**: HTTP 负载测试,模拟高并发场景

**功能**:
- 使用 wrk 进行压力测试
- 测试各服务的吞吐量和延迟
- 支持并发测试多个服务
- 生成性能报告

**依赖**:
- wrk (HTTP 压测工具)
- jq (JSON 处理)
- curl (HTTP 客户端)

### benchmark.sh

**用途**: 性能基准测试,测量 API 响应时间

**功能**:
- 测量平均/最小/最大响应时间
- 计算成功率
- 生成 HTML 和 Markdown 报告
- 导出详细的 CSV 数据

**依赖**:
- curl (HTTP 客户端)
- bc (数学计算)

---

## 快速开始

### 1. 安装依赖

#### macOS

```bash
brew install wrk jq bc
```

#### Ubuntu/Debian

```bash
sudo apt-get install wrk jq bc curl
```

#### 从源码安装 wrk

```bash
git clone https://github.com/wg/wrk
cd wrk
make
sudo cp wrk /usr/local/bin/
```

### 2. 启动服务

确保 Aetherius 服务正在运行:

```bash
# Docker Compose
cd deployments/docker-compose
docker-compose up -d

# Kubernetes
kubectl get pods -n aetherius
```

### 3. 运行测试

```bash
cd examples/scripts/performance

# 负载测试 (默认 30 秒)
./load-test.sh

# 基准测试 (默认 100 个样本)
./benchmark.sh
```

---

## 负载测试

### 基本用法

```bash
# 使用默认配置
./load-test.sh

# 自定义配置
./load-test.sh --duration 60s --threads 20 --connections 200
```

### 测试特定服务

```bash
# 仅测试 Agent Manager
./load-test.sh --agent-manager

# 仅测试 Orchestrator Service
./load-test.sh --orchestrator

# 仅测试 Reasoning Service
./load-test.sh --reasoning

# 并发压力测试
./load-test.sh --concurrent
```

### 自定义服务地址

```bash
export AGENT_MANAGER_URL=http://agent-manager:8080
export ORCHESTRATOR_URL=http://orchestrator:8081
export REASONING_URL=http://reasoning:8082

./load-test.sh
```

### 输出示例

```text
========================================
测试: Agent Manager List Agents
========================================
URL: http://localhost:8080/api/v1/agents
Method: GET
Duration: 30s
Threads: 10
Connections: 100

Running 30s test @ http://localhost:8080/api/v1/agents
  10 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    12.45ms    3.21ms  45.32ms   87.43%
    Req/Sec   825.43    124.56     1.02k    69.23%
  247532 requests in 30.01s, 45.23MB read
Requests/sec:   8248.76
Transfer/sec:      1.51MB
```

### 性能目标

| 指标 | 目标值 |
|------|--------|
| 请求成功率 | ≥ 99% |
| 平均延迟 (健康检查) | < 10ms |
| 平均延迟 (简单查询) | < 50ms |
| 平均延迟 (复杂分析) | < 500ms |
| 吞吐量 (每服务) | > 1000 req/s |
| P95 延迟 | < 2x 平均延迟 |
| P99 延迟 | < 3x 平均延迟 |

---

## 基准测试

### 基本用法

```bash
# 使用默认配置 (100 个样本)
./benchmark.sh

# 增加样本数
./benchmark.sh --samples 1000

# 自定义输出目录
./benchmark.sh --output my-benchmark-results
```

### 测试特定服务

```bash
# 仅测试 Agent Manager
./benchmark.sh --agent-manager --samples 500

# 仅测试 Reasoning Service
./benchmark.sh --reasoning --samples 200
```

### 输出文件

基准测试会生成以下文件:

```text
benchmark-results/
├── benchmark-report.html          # HTML 报告
├── benchmark-report.md            # Markdown 报告
├── Agent_Manager_Health_Check.csv
├── Agent_Manager_List_Agents.csv
├── Orchestrator_List_Workflows.csv
├── Reasoning_Root_Cause_Analysis.csv
└── ...
```

### 报告示例

#### Markdown 报告

```markdown
# Aetherius 性能基准测试报告

**生成时间**: 2025-09-30 15:30:00

## 性能摘要

| 测试项 | 平均响应时间 | 最小响应时间 | 最大响应时间 | 成功率 |
|--------|-------------|-------------|-------------|--------|
| Agent Manager Health Check | 8ms | 5ms | 15ms | 100.0% |
| Agent Manager List Agents | 42ms | 28ms | 89ms | 100.0% |
| Reasoning Root Cause Analysis | 387ms | 245ms | 612ms | 99.0% |
```

#### HTML 报告

HTML 报告包含:
- 测试配置摘要
- 性能指标表格
- 优化建议
- 详细数据链接

在浏览器中打开:

```bash
open benchmark-results/benchmark-report.html
# 或
xdg-open benchmark-results/benchmark-report.html
```

### CSV 数据分析

CSV 文件格式:

```csv
sample,time_ms,status
1,42,200
2,38,200
3,45,200
...
```

可以使用 Excel、Python、R 等工具进行进一步分析:

```python
import pandas as pd
import matplotlib.pyplot as plt

# 读取数据
df = pd.read_csv('benchmark-results/Agent_Manager_List_Agents.csv')

# 绘制响应时间分布
df['time_ms'].hist(bins=50)
plt.xlabel('Response Time (ms)')
plt.ylabel('Frequency')
plt.title('Response Time Distribution')
plt.show()

# 计算百分位数
print(f"P50: {df['time_ms'].quantile(0.50)}ms")
print(f"P95: {df['time_ms'].quantile(0.95)}ms")
print(f"P99: {df['time_ms'].quantile(0.99)}ms")
```

---

## 性能指标

### 关键指标

#### 1. 延迟 (Latency)

**定义**: 从发送请求到收到响应的时间

**指标**:
- **平均延迟**: 所有请求的平均响应时间
- **P50 延迟**: 50% 的请求响应时间小于此值
- **P95 延迟**: 95% 的请求响应时间小于此值
- **P99 延迟**: 99% 的请求响应时间小于此值

**目标值**:

| API 类型 | 平均延迟 | P95 延迟 | P99 延迟 |
|---------|---------|---------|---------|
| 健康检查 | < 10ms | < 20ms | < 30ms |
| 简单查询 | < 50ms | < 100ms | < 150ms |
| 复杂查询 | < 200ms | < 400ms | < 600ms |
| AI 分析 | < 500ms | < 1000ms | < 1500ms |

#### 2. 吞吐量 (Throughput)

**定义**: 每秒处理的请求数 (RPS/QPS)

**目标值**:
- Agent Manager: > 1000 req/s
- Orchestrator Service: > 500 req/s
- Reasoning Service: > 100 req/s (AI 分析较重)

#### 3. 错误率 (Error Rate)

**定义**: 失败请求占总请求的比例

**目标值**:
- 正常情况: < 0.1%
- 高负载: < 1%
- 限流场景: 根据配置

#### 4. 并发能力

**定义**: 同时处理的请求数

**目标值**:
- 每个服务支持 ≥ 100 并发连接
- 集群支持 ≥ 1000 并发连接

---

## 优化建议

### 1. 延迟优化

#### 数据库查询优化

```sql
-- 添加索引
CREATE INDEX idx_events_cluster_severity ON events(cluster_id, severity);
CREATE INDEX idx_events_timestamp ON events(timestamp DESC);

-- 使用连接池
database:
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime: 60
```

#### 缓存优化

```yaml
# 启用 Redis 缓存
cache:
  enabled: true
  backend: redis
  ttl:
    agent: 600        # Agent 信息缓存 10 分钟
    event: 60         # 事件缓存 1 分钟
    analysis: 300     # 分析结果缓存 5 分钟
```

#### 批处理优化

```yaml
# 使用批处理减少网络往返
event:
  batch_size: 100       # 批量处理 100 个事件
  batch_timeout: 1000   # 1 秒超时
```

### 2. 吞吐量优化

#### 增加工作协程

```yaml
# Agent Manager
event:
  workers: 20          # 增加事件处理协程

# Orchestrator Service
executor:
  workers: 50          # 增加工作流执行协程
  queue_size: 2000     # 增加队列大小
```

#### 水平扩展

```yaml
# Kubernetes HPA
spec:
  replicas: 3
  autoscaling:
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
```

### 3. 资源优化

#### 内存优化

```yaml
# 限制缓存大小
cache:
  max_size: 10000     # 最多缓存 10000 个条目

# 定期清理
data:
  retention_days: 30  # 保留 30 天数据
```

#### CPU 优化

```go
// 使用 pprof 分析热点
import _ "net/http/pprof"

// 优化热路径代码
// - 减少内存分配
// - 使用 sync.Pool 复用对象
// - 避免不必要的锁竞争
```

### 4. 网络优化

#### 启用 HTTP/2

```yaml
server:
  enable_http2: true
```

#### 启用压缩

```yaml
server:
  enable_compression: true
  compression_level: 6
```

#### 优化 NATS 配置

```yaml
nats:
  max_reconnects: -1
  reconnect_wait: 2
  ping_interval: 20
  max_pings_out: 2
```

---

## 性能测试最佳实践

### 1. 测试环境

- **隔离环境**: 在独立的测试环境进行,避免影响生产
- **真实数据**: 使用接近生产的数据量和结构
- **网络条件**: 模拟真实的网络延迟和带宽

### 2. 测试策略

- **基线测试**: 建立性能基线,用于对比
- **渐进加压**: 逐步增加负载,找到性能拐点
- **持续时间**: 测试足够长时间,观察稳定性
- **峰值测试**: 模拟峰值流量场景

### 3. 监控指标

在测试期间监控:

```bash
# CPU 和内存
kubectl top pods -n aetherius

# 请求指标
curl http://localhost:9090/api/v1/query?query=rate(http_requests_total[1m])

# 数据库连接
psql -c "SELECT count(*) FROM pg_stat_activity"

# Redis 内存
redis-cli INFO memory
```

### 4. 结果分析

- **对比基线**: 与历史数据对比
- **识别瓶颈**: CPU、内存、网络、数据库?
- **分析日志**: 查找错误和异常
- **优化迭代**: 优化后重新测试

---

## 示例场景

### 场景 1: 日常性能验证

```bash
# 快速基准测试 (100 样本)
./benchmark.sh --samples 100

# 查看报告
cat benchmark-results/benchmark-report.md
```

### 场景 2: 版本发布前测试

```bash
# 全面负载测试 (5 分钟)
./load-test.sh --duration 300s --threads 20 --connections 200

# 详细基准测试 (1000 样本)
./benchmark.sh --samples 1000

# 对比历史数据
diff old-results/benchmark-report.md benchmark-results/benchmark-report.md
```

### 场景 3: 性能调优验证

```bash
# 优化前测试
./benchmark.sh --output before-optimization

# 应用优化
# ... 修改配置,重启服务 ...

# 优化后测试
./benchmark.sh --output after-optimization

# 对比结果
diff before-optimization/benchmark-report.md \
     after-optimization/benchmark-report.md
```

### 场景 4: 容量规划

```bash
# 逐步增加负载
for connections in 50 100 200 400 800; do
  echo "Testing with $connections connections..."
  ./load-test.sh \
    --duration 60s \
    --threads 10 \
    --connections $connections | \
    tee "load-test-c${connections}.log"
  sleep 10
done

# 分析结果,找到性能拐点
```

---

## 故障排查

### 测试失败

```bash
# 检查服务是否运行
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health

# 检查日志
docker logs aetherius-agent-manager
kubectl logs -n aetherius deployment/agent-manager
```

### 工具未找到

```bash
# 检查 wrk
which wrk
wrk --version

# 检查 curl
which curl
curl --version

# 检查 bc
which bc
bc --version
```

### 权限问题

```bash
# 确保脚本可执行
chmod +x load-test.sh
chmod +x benchmark.sh
```

---

## 参考资料

- [wrk 文档](https://github.com/wg/wrk)
- [性能测试最佳实践](https://docs.microsoft.com/en-us/azure/architecture/best-practices/performance-testing)
- [HTTP 负载测试工具对比](https://k6.io/blog/comparing-best-open-source-load-testing-tools/)
- [Aetherius 监控指南](../../deployments/monitoring/README.md)
- [Aetherius 故障排查](../../docs/TROUBLESHOOTING.md)