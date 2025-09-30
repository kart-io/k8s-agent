# Aetherius 故障排查指南

本文档提供 Aetherius 平台常见问题的诊断和解决方法。

---

## 目录

- [快速诊断](#快速诊断)
- [服务问题](#服务问题)
- [依赖服务问题](#依赖服务问题)
- [网络和连接问题](#网络和连接问题)
- [性能问题](#性能问题)
- [数据问题](#数据问题)
- [部署问题](#部署问题)
- [配置问题](#配置问题)
- [日志分析](#日志分析)
- [常用工具和命令](#常用工具和命令)

---

## 快速诊断

### 检查系统状态

#### Docker Compose 环境

```bash
# 检查所有服务状态
cd deployments/docker-compose
docker-compose ps

# 查看服务健康状态
curl http://localhost:8080/health  # Agent Manager
curl http://localhost:8081/health  # Orchestrator Service
curl http://localhost:8082/health  # Reasoning Service

# 检查依赖服务
docker ps | grep -E "postgres|redis|nats|neo4j"
```

#### Kubernetes 环境

```bash
# 检查所有 Pod 状态
kubectl get pods -n aetherius

# 检查服务状态
kubectl get svc -n aetherius

# 快速健康检查
kubectl exec -n aetherius deployment/agent-manager -- \
  wget -q -O- http://localhost:8080/health
```

### 诊断流程图

```text
服务无法访问
    │
    ├─> Pod/容器是否运行?
    │   ├─> NO -> 检查部署状态 -> 查看事件和日志
    │   └─> YES -> 继续
    │
    ├─> 健康检查是否通过?
    │   ├─> NO -> 检查依赖服务 -> 查看应用日志
    │   └─> YES -> 继续
    │
    ├─> 网络连接是否正常?
    │   ├─> NO -> 检查服务/端口 -> 检查网络策略
    │   └─> YES -> 继续
    │
    └─> 检查应用层问题 -> 查看日志和指标
```

---

## 服务问题

### Agent Manager

#### 问题: Agent Manager 无法启动

**症状**:
- 容器/Pod 启动后立即退出
- 日志显示启动错误

**诊断步骤**:

```bash
# 查看日志
docker logs aetherius-agent-manager
# 或
kubectl logs -n aetherius deployment/agent-manager

# 检查配置
docker exec aetherius-agent-manager cat /etc/aetherius/config.yaml
# 或
kubectl exec -n aetherius deployment/agent-manager -- \
  cat /etc/aetherius/config.yaml
```

**常见原因和解决方法**:

1. **数据库连接失败**

```bash
# 检查数据库是否可访问
docker exec aetherius-agent-manager \
  psql -h postgres -U aetherius -d agent_manager -c "SELECT 1"

# 检查数据库密码
echo $DATABASE_PASSWORD

# 解决方法: 修正数据库配置或重启数据库
```

2. **NATS 连接失败**

```bash
# 检查 NATS 服务
docker ps | grep nats
kubectl get pods -n aetherius | grep nats

# 测试 NATS 连接
docker exec aetherius-agent-manager \
  nats-top -s nats://nats:4222

# 解决方法: 检查 NATS 地址和端口配置
```

3. **端口被占用**

```bash
# 检查端口占用
lsof -i :8080
# 或
netstat -tuln | grep 8080

# 解决方法: 停止占用端口的进程或更改配置端口
```

#### 问题: Agent 无法注册

**症状**:
- Collect Agent 日志显示注册失败
- Agent Manager 日志无注册请求记录

**诊断步骤**:

```bash
# 检查 Agent Manager 日志
kubectl logs -n aetherius deployment/agent-manager | grep "register"

# 检查 Collect Agent 日志
kubectl logs -n kube-system daemonset/collect-agent | grep "register"

# 检查 NATS 连接
kubectl exec -n aetherius deployment/agent-manager -- \
  nats sub "aetherius.agent.*.register"
```

**解决方法**:

1. **NATS 主题配置错误**

```yaml
# 检查 Agent Manager 配置
nats:
  subjects:
    registration: "aetherius.agent.{cluster_id}.register"

# 检查 Collect Agent 配置
nats:
  subjects:
    registration: "aetherius.agent.{cluster_id}.register"
```

2. **认证问题**

```bash
# 检查认证配置
kubectl get secret -n aetherius agent-auth-token
kubectl get secret -n kube-system agent-auth-token

# 确保两者一致
```

#### 问题: 事件处理延迟过高

**症状**:
- 事件从产生到处理时间过长
- Prometheus 显示高延迟告警

**诊断步骤**:

```bash
# 检查事件处理指标
curl http://localhost:8080/metrics | grep event_processing

# 检查队列积压
curl http://localhost:8080/metrics | grep event_queue_size

# 检查 goroutine 数量
curl http://localhost:8080/metrics | grep go_goroutines
```

**解决方法**:

1. **增加工作协程数**

```yaml
# config.yaml
event:
  workers: 20  # 从 10 增加到 20
  buffer_size: 2000  # 从 1000 增加到 2000
```

2. **优化批处理**

```yaml
event:
  batch_size: 200  # 增加批处理大小
  batch_timeout: 500  # 减少批处理超时
```

3. **检查数据库性能**

```sql
-- 查看慢查询
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC
LIMIT 10;
```

---

### Orchestrator Service

#### 问题: 工作流无法执行

**症状**:
- 工作流提交后状态一直是 "pending"
- 日志显示执行器错误

**诊断步骤**:

```bash
# 查看工作流执行状态
curl http://localhost:8081/api/v1/workflows/executions | jq

# 查看执行器日志
kubectl logs -n aetherius deployment/orchestrator-service | grep "executor"

# 检查执行器队列
curl http://localhost:8081/metrics | grep executor_queue
```

**解决方法**:

1. **执行器资源不足**

```yaml
# config.yaml
executor:
  workers: 50  # 增加工作协程数
  queue_size: 2000  # 增加队列大小
```

2. **工作流定义错误**

```bash
# 验证工作流 YAML
yamllint workflows/diagnose-oom-killed.yaml

# 检查工作流加载
kubectl exec -n aetherius deployment/orchestrator-service -- \
  ls -l /etc/aetherius/workflows/
```

#### 问题: 步骤执行超时

**症状**:
- 工作流执行失败,错误信息显示超时
- 特定步骤一直无法完成

**诊断步骤**:

```bash
# 查看步骤执行历史
curl "http://localhost:8081/api/v1/workflows/executions/{execution_id}" | jq

# 检查步骤超时配置
grep -r "timeout" workflows/

# 查看相关服务日志
kubectl logs -n aetherius deployment/agent-manager | grep "command"
```

**解决方法**:

1. **增加步骤超时时间**

```yaml
# workflow.yaml
steps:
  - id: collect_logs
    type: diagnostic
    config:
      timeout: 300  # 从 120 增加到 300 秒
```

2. **优化命令执行**

```yaml
# Agent Manager config.yaml
agent:
  command:
    default_timeout: 600  # 增加默认超时
    max_concurrent: 200  # 增加并发数
```

---

### Reasoning Service

#### 问题: 根因分析失败

**症状**:
- API 返回 500 错误
- 日志显示分析异常

**诊断步骤**:

```bash
# 测试根因分析 API
curl -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d @test-request.json | jq

# 查看分析器日志
kubectl logs -n aetherius deployment/reasoning-service | grep "analyzer"

# 检查 Python 进程
kubectl exec -n aetherius deployment/reasoning-service -- ps aux
```

**解决方法**:

1. **内存不足**

```yaml
# kubernetes deployment
resources:
  limits:
    memory: 2Gi  # 从 1Gi 增加到 2Gi
  requests:
    memory: 1Gi
```

2. **模型加载失败**

```bash
# 检查模型文件
kubectl exec -n aetherius deployment/reasoning-service -- \
  ls -l /app/models/

# 重新下载模型
kubectl exec -n aetherius deployment/reasoning-service -- \
  python -c "from sklearn.ensemble import IsolationForest; print('OK')"
```

#### 问题: Neo4j 连接失败

**症状**:
- 知识图谱功能不可用
- 日志显示 Neo4j 连接错误

**诊断步骤**:

```bash
# 检查 Neo4j 服务
kubectl get pods -n aetherius | grep neo4j
docker ps | grep neo4j

# 测试 Neo4j 连接
kubectl exec -n aetherius deployment/reasoning-service -- \
  python -c "from neo4j import GraphDatabase; \
             driver = GraphDatabase.driver('bolt://neo4j:7687', \
                                          auth=('neo4j', 'neo4j_pass')); \
             driver.verify_connectivity(); \
             print('Connected')"
```

**解决方法**:

1. **Neo4j 服务未就绪**

```bash
# 等待 Neo4j 启动
kubectl wait --for=condition=ready pod \
  -l app=neo4j -n aetherius --timeout=300s

# 或重启 Neo4j
kubectl rollout restart statefulset/neo4j -n aetherius
```

2. **密码错误**

```bash
# 重置 Neo4j 密码
kubectl exec -n aetherius neo4j-0 -- \
  cypher-shell -u neo4j -p neo4j \
  "ALTER CURRENT USER SET PASSWORD FROM 'neo4j' TO 'new_password'"
```

---

## 依赖服务问题

### PostgreSQL

#### 问题: 数据库连接数过多

**症状**:
- 新连接失败,错误: "too many connections"
- 服务响应缓慢

**诊断步骤**:

```sql
-- 查看当前连接数
SELECT count(*) FROM pg_stat_activity;

-- 查看最大连接数
SHOW max_connections;

-- 查看各数据库连接分布
SELECT datname, count(*)
FROM pg_stat_activity
GROUP BY datname;
```

**解决方法**:

1. **增加最大连接数**

```bash
# 修改 PostgreSQL 配置
kubectl edit configmap postgres-config -n aetherius

# 添加或修改
max_connections = 200

# 重启 PostgreSQL
kubectl rollout restart statefulset/postgres -n aetherius
```

2. **优化连接池配置**

```yaml
# 各服务 config.yaml
database:
  max_open_conns: 50  # 减少每个服务的最大连接数
  max_idle_conns: 10
  conn_max_lifetime: 30  # 减少连接生命周期
```

#### 问题: 数据库查询缓慢

**诊断步骤**:

```sql
-- 启用查询统计
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 查看慢查询
SELECT query, calls, mean_exec_time, max_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC
LIMIT 20;

-- 查看锁等待
SELECT * FROM pg_locks WHERE NOT granted;
```

**解决方法**:

1. **添加索引**

```sql
-- 为事件表添加索引
CREATE INDEX idx_events_cluster_id ON events(cluster_id);
CREATE INDEX idx_events_severity ON events(severity);
CREATE INDEX idx_events_timestamp ON events(timestamp);

-- 为执行表添加索引
CREATE INDEX idx_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX idx_executions_status ON workflow_executions(status);
```

2. **清理旧数据**

```sql
-- 删除 30 天前的事件
DELETE FROM events
WHERE timestamp < NOW() - INTERVAL '30 days'
AND severity NOT IN ('critical', 'high');

-- 归档旧的工作流执行
INSERT INTO workflow_executions_archive
SELECT * FROM workflow_executions
WHERE created_at < NOW() - INTERVAL '90 days';

DELETE FROM workflow_executions
WHERE created_at < NOW() - INTERVAL '90 days';
```

---

### Redis

#### 问题: Redis 内存不足

**症状**:
- 写入失败
- 缓存命中率下降

**诊断步骤**:

```bash
# 连接 Redis
kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass

# 查看内存使用
INFO memory

# 查看键空间信息
INFO keyspace

# 查看缓存命中率
INFO stats | grep keyspace
```

**解决方法**:

1. **增加内存限制**

```yaml
# kubernetes deployment
spec:
  containers:
    - name: redis
      resources:
        limits:
          memory: 2Gi  # 增加内存
```

2. **启用键过期策略**

```bash
# 设置内存策略
kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass \
  CONFIG SET maxmemory-policy allkeys-lru

# 设置最大内存
kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass \
  CONFIG SET maxmemory 1gb
```

3. **清理无用键**

```bash
# 查找大键
kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass \
  --bigkeys

# 删除过期键
kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass \
  --scan --pattern "cache:*" | \
  xargs kubectl exec -n aetherius redis-0 -- redis-cli -a redis_pass DEL
```

---

### NATS

#### 问题: 消息堆积

**症状**:
- Slow consumer 告警
- 消息延迟增加

**诊断步骤**:

```bash
# 查看 NATS 状态
kubectl exec -n aetherius nats-0 -- nats-server -V

# 查看消息统计
curl http://nats:8222/varz | jq

# 查看慢消费者
curl http://nats:8222/connz | jq '.connections[] | select(.slow_consumer > 0)'
```

**解决方法**:

1. **增加消费者处理能力**

```yaml
# config.yaml
event:
  workers: 20  # 增加工作协程
  buffer_size: 2000  # 增加缓冲区
```

2. **启用流式传输**

```bash
# 创建 JetStream
kubectl exec -n aetherius nats-0 -- nats stream add \
  --subjects "aetherius.events.>" \
  --retention limits \
  --max-age 24h \
  --max-msgs -1 \
  --max-bytes -1 \
  EVENTS
```

---

## 网络和连接问题

### 服务间无法通信

**诊断步骤**:

```bash
# Kubernetes 环境
# 检查 Service
kubectl get svc -n aetherius

# 测试服务连接
kubectl exec -n aetherius deployment/agent-manager -- \
  wget -q -O- http://orchestrator-service:8081/health

# 检查 NetworkPolicy
kubectl get networkpolicy -n aetherius

# Docker Compose 环境
# 检查网络
docker network ls
docker network inspect aetherius

# 测试容器间连接
docker exec aetherius-agent-manager \
  wget -q -O- http://orchestrator-service:8081/health
```

**解决方法**:

1. **检查 DNS 解析**

```bash
# Kubernetes
kubectl exec -n aetherius deployment/agent-manager -- \
  nslookup orchestrator-service

# Docker Compose
docker exec aetherius-agent-manager \
  nslookup orchestrator-service
```

2. **检查防火墙规则**

```bash
# Kubernetes NetworkPolicy
kubectl describe networkpolicy -n aetherius

# 如果过于严格,临时删除测试
kubectl delete networkpolicy --all -n aetherius
```

---

## 性能问题

### CPU 使用率过高

**诊断步骤**:

```bash
# 检查 CPU 使用
kubectl top pods -n aetherius
docker stats

# 查看进程
kubectl exec -n aetherius deployment/agent-manager -- top -b -n 1

# 查看 Goroutine
curl http://localhost:8080/metrics | grep go_goroutines

# Go profiling
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof -http=:8081 cpu.prof
```

**解决方法**:

1. **优化代码热点**

```bash
# 使用 pprof 分析
go tool pprof -top cpu.prof
go tool pprof -list functionName cpu.prof
```

2. **增加资源限制**

```yaml
# kubernetes
resources:
  limits:
    cpu: "2"
  requests:
    cpu: "1"
```

### 内存泄漏

**诊断步骤**:

```bash
# 监控内存增长
watch kubectl top pods -n aetherius

# 内存 profiling
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof -http=:8081 heap.prof

# Python 内存分析
kubectl exec -n aetherius deployment/reasoning-service -- \
  python -m memory_profiler app.py
```

**解决方法**:

1. **分析内存泄漏点**

```bash
# Go 服务
go tool pprof -alloc_space heap.prof
go tool pprof -inuse_space heap.prof

# Python 服务
pip install memory_profiler
python -m memory_profiler main.py
```

2. **设置内存限制和自动重启**

```yaml
# kubernetes
resources:
  limits:
    memory: 2Gi
livenessProbe:
  exec:
    command:
      - sh
      - -c
      - "[ $(cat /proc/meminfo | grep MemAvailable | awk '{print $2}') -gt 102400 ]"
```

---

## 数据问题

### 数据不一致

**诊断步骤**:

```bash
# 检查数据库数据
psql -h localhost -U aetherius -d agent_manager \
  -c "SELECT COUNT(*) FROM events WHERE processed = false;"

# 检查 Redis 缓存
redis-cli -a redis_pass KEYS "agent:*"

# 检查知识图谱
kubectl exec -n aetherius neo4j-0 -- cypher-shell -u neo4j -p neo4j_pass \
  "MATCH (n) RETURN count(n)"
```

**解决方法**:

1. **清理缓存**

```bash
# 清理 Redis
redis-cli -a redis_pass FLUSHDB

# 或清理特定前缀
redis-cli -a redis_pass --scan --pattern "cache:*" | xargs redis-cli -a redis_pass DEL
```

2. **重建索引**

```sql
-- PostgreSQL
REINDEX DATABASE agent_manager;

-- 或重建特定表
REINDEX TABLE events;
```

---

## 部署问题

### Pod 无法启动

**诊断步骤**:

```bash
# 查看 Pod 状态
kubectl get pods -n aetherius

# 查看 Pod 详情
kubectl describe pod <pod-name> -n aetherius

# 查看事件
kubectl get events -n aetherius --sort-by='.lastTimestamp'
```

**常见原因**:

1. **ImagePullBackOff**

```bash
# 检查镜像
kubectl describe pod <pod-name> -n aetherius | grep Image

# 解决方法:
# - 检查镜像名称和标签
# - 检查镜像仓库认证
# - 使用 imagePullSecrets
```

2. **CrashLoopBackOff**

```bash
# 查看容器日志
kubectl logs <pod-name> -n aetherius

# 查看上一次崩溃日志
kubectl logs <pod-name> -n aetherius --previous
```

---

## 配置问题

### 配置未生效

**诊断步骤**:

```bash
# 检查 ConfigMap
kubectl get configmap -n aetherius
kubectl describe configmap agent-manager-config -n aetherius

# 检查 Pod 挂载
kubectl exec -n aetherius deployment/agent-manager -- \
  cat /etc/aetherius/config.yaml

# 检查环境变量
kubectl exec -n aetherius deployment/agent-manager -- env | sort
```

**解决方法**:

1. **重新加载配置**

```bash
# 更新 ConfigMap
kubectl edit configmap agent-manager-config -n aetherius

# 重启 Pod 使配置生效
kubectl rollout restart deployment/agent-manager -n aetherius
```

2. **验证配置优先级**

```text
优先级(从高到低):
1. 命令行参数
2. 环境变量
3. ConfigMap
4. 配置文件默认值
```

---

## 日志分析

### 收集日志

```bash
# Docker Compose
docker-compose logs > all-logs.txt
docker logs aetherius-agent-manager > agent-manager.log

# Kubernetes
kubectl logs -n aetherius deployment/agent-manager > agent-manager.log
kubectl logs -n aetherius deployment/orchestrator-service > orchestrator.log
kubectl logs -n aetherius deployment/reasoning-service > reasoning.log

# 收集所有 Pod 日志
for pod in $(kubectl get pods -n aetherius -o name); do
  kubectl logs -n aetherius $pod > ${pod}.log
done
```

### 日志分析工具

```bash
# 搜索错误
grep -i "error\|fatal\|panic" *.log

# 统计错误类型
grep -i "error" *.log | awk '{print $NF}' | sort | uniq -c | sort -rn

# 分析时间线
grep "2025-09-30" agent-manager.log | sort

# 使用 jq 分析 JSON 日志
cat agent-manager.log | jq 'select(.level == "error")'
```

---

## 常用工具和命令

### 健康检查脚本

```bash
#!/bin/bash
# health-check.sh

echo "=== Aetherius Health Check ==="

# 检查 Agent Manager
if curl -s http://localhost:8080/health | grep -q "ok"; then
  echo "✓ Agent Manager: OK"
else
  echo "✗ Agent Manager: FAILED"
fi

# 检查 Orchestrator Service
if curl -s http://localhost:8081/health | grep -q "ok"; then
  echo "✓ Orchestrator Service: OK"
else
  echo "✗ Orchestrator Service: FAILED"
fi

# 检查 Reasoning Service
if curl -s http://localhost:8082/health | grep -q "ok"; then
  echo "✓ Reasoning Service: OK"
else
  echo "✗ Reasoning Service: FAILED"
fi

# 检查 PostgreSQL
if psql -h localhost -U aetherius -d agent_manager -c "SELECT 1" > /dev/null 2>&1; then
  echo "✓ PostgreSQL: OK"
else
  echo "✗ PostgreSQL: FAILED"
fi

# 检查 Redis
if redis-cli -a redis_pass ping | grep -q "PONG"; then
  echo "✓ Redis: OK"
else
  echo "✗ Redis: FAILED"
fi

# 检查 NATS
if curl -s http://localhost:8222/varz | jq -e '.server_id' > /dev/null 2>&1; then
  echo "✓ NATS: OK"
else
  echo "✗ NATS: FAILED"
fi
```

### 日志收集脚本

```bash
#!/bin/bash
# collect-logs.sh

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_DIR="logs_${TIMESTAMP}"

mkdir -p "$LOG_DIR"

echo "Collecting logs to $LOG_DIR ..."

# Docker Compose
if command -v docker-compose &> /dev/null; then
  docker-compose logs > "$LOG_DIR/docker-compose.log"
fi

# Kubernetes
if command -v kubectl &> /dev/null; then
  kubectl get all -n aetherius > "$LOG_DIR/k8s-resources.txt"
  kubectl get events -n aetherius > "$LOG_DIR/k8s-events.txt"

  for pod in $(kubectl get pods -n aetherius -o name); do
    kubectl logs -n aetherius $pod > "$LOG_DIR/${pod}.log" 2>&1
  done
fi

# 压缩
tar -czf "logs_${TIMESTAMP}.tar.gz" "$LOG_DIR"
echo "Logs collected: logs_${TIMESTAMP}.tar.gz"
```

---

## 获取帮助

如果上述方法无法解决问题:

1. **查看文档**: [完整文档](../README.md)
2. **搜索 Issues**: [GitHub Issues](https://github.com/kart-io/k8s-agent/issues)
3. **提交 Issue**: [新建 Issue](https://github.com/kart-io/k8s-agent/issues/new)
4. **联系支持**: support@aetherius.example.com

提交问题时请包含:
- 问题详细描述
- 复现步骤
- 错误日志
- 环境信息 (版本、部署方式、配置等)
- 诊断结果