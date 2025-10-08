# Orchestrator Service 测试指南

## 快速测试 - 删除 Pod 验证

### 前置条件检查

```bash
# 1. 检查 K8s 基础设施
kubectl get pods -n aetherius-dev

# 应该看到:
# - postgres-xxx (Running)
# - nats-xxx (Running)
# - redis-xxx (Running)

# 2. 检查要删除的测试 Pods
kubectl get pods -n default | grep test
```

### 步骤 1: 启动端口转发（3个终端窗口）

**终端 1 - PostgreSQL:**
```bash
cd deployments/kustomize
make forward-dev-postgres
# 保持运行，监听 localhost:5432
```

**终端 2 - NATS:**
```bash
cd deployments/kustomize
make forward-dev-nats
# 保持运行，监听 localhost:4222
```

**终端 3 - Redis (可选):**
```bash
cd deployments/kustomize
make forward-dev-redis
# 保持运行，监听 localhost:6379
```

### 步骤 2: 初始化数据库和策略（新终端）

```bash
cd orchestrator-service

# 2.1 创建数据库
make db-create

# 2.2 创建测试策略
./scripts/test-strategy.sh

# 验证策略已创建
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, name, enabled FROM strategies;"
```

### 步骤 3: 启动 orchestrator-service（终端 4）

```bash
cd orchestrator-service
make run
```

**期望看到的日志:**
```
==========================================================
     Aetherius Orchestrator Service - Initialization
==========================================================
📦 [1/6] Initializing PostgreSQL host=localhost port=5432
✅ PostgreSQL initialized successfully
📦 [2/6] Initializing Redis addr=localhost:6379
✅ Redis initialized successfully
📡 [3/6] Connecting to NATS url=nats://localhost:4222
✅ NATS connected successfully
⚙️  [4/6] Initializing workflow engine
✅ Workflow engine initialized
🎯 [5/6] Initializing strategy manager
✅ Strategy manager initialized
📬 [6/6] Initializing event subscriber
========== Starting event subscriber ==========
✓ Subscribed to critical events subject=internal.event.critical
✓ Subscribed to anomaly events subject=internal.event.anomaly
✓ Subscribed to all internal events (debug)
========== Event subscriber started successfully ==========
==========================================================
✅ Orchestrator Service started successfully!
==========================================================
🎧 Listening for events on NATS channels:
   - internal.event.critical
   - internal.event.anomaly
   - internal.event.* (debug)
==========================================================
```

### 步骤 4: 方式 A - 手动发送测试事件（推荐）

在新终端中发送测试事件，模拟 Pod 故障：

```bash
cd orchestrator-service
./scripts/test-event.sh
```

**期望在 orchestrator-service 日志中看到:**
```
📨 Received message on critical channel subject=internal.event.critical size=xxx
========== Processing Event ==========
✓ Event parsed successfully type=pod_failure cluster_id=test-cluster-001 severity=critical
🔍 Starting strategy matching event_type=pod_failure severity=critical
📋 Retrieved strategies from database total_strategies=1
✅ Strategy matched successfully strategy_id=strat-pod-crashloop
🚀 Starting strategy execution strategy_id=strat-pod-crashloop
🎬 Starting workflow execution workflow_id=wf-pod-failure-diagnostic
✅ Workflow execution started
========== Strategy execution started successfully ==========
```

### 步骤 5: 方式 B - 删除真实 Pods 触发事件

**注意:** 这需要 collect-agent 正在运行并监听 K8s 事件。

```bash
# 删除 Pod
kubectl delete pod final-correct-test -n default

# 或批量删除
kubectl delete pod final-correct-test final-test test-command-correlation \
  test-correlation-final test-event-correlation test-final-verification -n default
```

如果 collect-agent 配置正确，它会：
1. 捕获 Pod 删除/重启事件
2. 通过 NATS 发送到 `internal.event.critical` 或 `internal.event.anomaly`
3. orchestrator-service 接收并处理

### 步骤 6: 验证结果

**6.1 检查工作流执行记录:**
```bash
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, workflow_id, status, started_at
   FROM workflow_executions
   ORDER BY started_at DESC
   LIMIT 5;"
```

**6.2 检查策略匹配情况:**
```bash
# 在 orchestrator-service 日志中搜索
grep "Strategy matched"
grep "📨 Received"
grep "🚀 Executing"
```

**6.3 检查 NATS 消息（可选）:**
```bash
# 安装 nats CLI
go install github.com/nats-io/natscli/nats@latest

# 监听所有 internal 事件
nats sub "internal.event.*" --server="nats://localhost:4222"
```

## 常见问题排查

### 问题 1: orchestrator 启动失败

**症状:**
```
failed to connect to database
```

**解决:**
```bash
# 确保端口转发正在运行
lsof -i :5432  # PostgreSQL
lsof -i :4222  # NATS

# 重新启动端口转发
cd deployments/kustomize
make forward-dev-postgres
make forward-dev-nats
```

### 问题 2: 收不到事件

**症状:** orchestrator 启动成功，但没有 `📨 Received message` 日志

**原因和解决:**

1. **NATS 连接问题:**
   ```bash
   # 测试 NATS 连接
   nats-bench -s nats://localhost:4222 pub test --msgs=1
   ```

2. **没有策略配置:**
   ```bash
   # 检查策略
   kubectl exec -n aetherius-dev deployment/postgres -- \
     psql -U postgres -d aetherius_orchestrator -c \
     "SELECT COUNT(*) FROM strategies WHERE enabled = true;"

   # 如果为 0，运行:
   cd orchestrator-service && ./scripts/test-strategy.sh
   ```

3. **collect-agent 未运行或未发送事件:**
   ```bash
   # 检查 collect-agent
   kubectl get pods -n aetherius-dev | grep collect-agent

   # 直接发送测试事件
   cd orchestrator-service && ./scripts/test-event.sh
   ```

### 问题 3: 收到事件但没有匹配策略

**症状:**
```
⚠️  No strategy matched for event
📋 Retrieved strategies from database total_strategies=0
```

**解决:**
```bash
# 重新初始化策略
cd orchestrator-service
./scripts/test-strategy.sh

# 验证
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, name, category, enabled FROM strategies;"
```

### 问题 4: 策略匹配但工作流失败

**症状:**
```
❌ Failed to load workflow from database
```

**解决:**
```bash
# 检查工作流是否存在
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, name, status FROM workflows;"

# 重新运行策略脚本（包含工作流）
cd orchestrator-service && ./scripts/test-strategy.sh
```

## 完整测试场景示例

### 场景 1: 模拟 CrashLoopBackOff Pod

```bash
# 1. 创建会崩溃的 Pod
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: crash-test-pod
  namespace: default
spec:
  restartPolicy: Always
  containers:
  - name: crash-container
    image: busybox:latest
    command: ["/bin/sh", "-c", "exit 1"]
EOF

# 2. 等待 Pod 进入 CrashLoopBackOff
sleep 10
kubectl get pod crash-test-pod -n default

# 3. 如果 collect-agent 运行，应该会捕获事件
# 4. 查看 orchestrator-service 日志确认处理

# 5. 清理
kubectl delete pod crash-test-pod -n default
```

### 场景 2: 批量删除 Pod

```bash
# 1. 确认 orchestrator 运行
# 2. 删除所有测试 Pods
kubectl delete pod final-correct-test final-test \
  test-command-correlation test-correlation-final \
  test-event-correlation test-final-verification -n default

# 3. 实时查看 orchestrator 日志
# 应该看到多个事件被处理

# 4. 查询数据库验证
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT COUNT(*) FROM workflow_executions
   WHERE started_at > NOW() - INTERVAL '5 minutes';"
```

## 成功标志

如果一切正常，你应该看到：

1. ✅ orchestrator-service 成功启动
2. ✅ 订阅了 3 个 NATS 频道
3. ✅ 收到 `📨 Received message` 日志
4. ✅ 策略成功匹配 `✅ Strategy matched`
5. ✅ 工作流启动 `🎬 Starting workflow execution`
6. ✅ 数据库中有执行记录

## 下一步

完成基础测试后，可以：
- 创建自定义策略
- 定义复杂的工作流
- 集成 agent-manager 进行命令执行
- 添加 AI 分析步骤
