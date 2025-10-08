# Orchestrator Service 快速启动指南

## 当前状态

✅ **测试 Pod 已就绪**: `crashloop-test-pod` 正在 CrashLoopBackOff（重启 6 次）
✅ **测试脚本已创建**: 可以直接使用
✅ **日志增强完成**: 详细的表情符号日志
✅ **配置已优化**: 支持环境变量

## 一键启动（推荐）

打开 **4 个终端窗口**，按顺序执行：

### 终端 1: PostgreSQL 端口转发
```bash
cd deployments/kustomize
make forward-dev-postgres
```
保持运行 ✓

### 终端 2: NATS 端口转发
```bash
cd deployments/kustomize
make forward-dev-nats
```
保持运行 ✓

### 终端 3: 初始化并启动 Orchestrator
```bash
cd orchestrator-service

# 创建数据库（仅首次需要）
make db-create

# 创建测试策略（仅首次需要）
./scripts/test-strategy.sh

# 启动服务
make run
```

### 终端 4: 发送测试事件
```bash
cd orchestrator-service

# 等待终端 3 显示 "Orchestrator Service started successfully"
# 然后发送测试事件
./scripts/test-event.sh
```

## 期望结果

### 终端 3 应该显示:

```
==========================================================
     Aetherius Orchestrator Service - Initialization
==========================================================
📦 [1/6] Initializing PostgreSQL host=localhost port=5432
✅ PostgreSQL initialized successfully
📦 [2/6] Initializing Redis addr=localhost:6379
✅ Redis initialized successfully
📡 [3/6] Connecting to NATS url=nats://localhost:4222
✅ NATS connected successfully server_info=nats://localhost:4222
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

📨 Received message on critical channel subject=internal.event.critical size=256
========== Processing Event ==========
✓ Event parsed successfully type=pod_failure cluster_id=test-cluster-001
🔍 Starting strategy matching event_type=pod_failure severity=critical
📋 Retrieved strategies from database total_strategies=1
✅ Strategy matched successfully strategy_id=strat-pod-crashloop
🚀 Starting strategy execution strategy_id=strat-pod-crashloop
🎬 Starting workflow execution workflow_id=wf-pod-failure-diagnostic
✓ Workflow loaded workflow_name="Pod Failure Diagnostic Workflow"
✅ Workflow execution started execution_id=xxx
========== Strategy execution started successfully ==========
```

## 验证工作流执行

```bash
# 查询数据库中的执行记录
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, workflow_id, status, started_at
   FROM workflow_executions
   ORDER BY started_at DESC
   LIMIT 5;"
```

## 测试真实 Pod 事件（可选）

如果 collect-agent 正在运行，可以删除 crashloop-test-pod 触发真实事件：

```bash
# 删除 Pod（会触发事件）
kubectl delete pod crashloop-test-pod -n default

# 或强制删除
kubectl delete pod crashloop-test-pod -n default --force --grace-period=0

# 观察 orchestrator-service 日志是否收到事件
```

## 故障排查速查表

| 问题 | 症状 | 解决方案 |
|------|------|----------|
| **数据库连接失败** | `failed to connect to database` | 检查终端 1 的端口转发是否运行 |
| **NATS 连接失败** | `failed to connect to NATS` | 检查终端 2 的端口转发是否运行 |
| **没有策略** | `⚠️ No active strategies` | 运行 `./scripts/test-strategy.sh` |
| **收不到事件** | 没有 `📨` 日志 | 运行 `./scripts/test-event.sh` 手动发送 |
| **策略不匹配** | `⚠️ No strategy matched` | 检查事件类型是否为 `pod_failure` |

## 日志过滤技巧

```bash
# 只看关键步骤
make run | grep -E "📨|🔍|🚀|✅.*success"

# 只看错误和警告
make run | grep -E "❌|⚠️"

# 保存日志到文件
make run | tee orchestrator-$(date +%Y%m%d-%H%M%S).log
```

## 完整清理

```bash
# 停止所有服务（Ctrl+C 所有终端）

# 删除测试 Pod
kubectl delete pod crashloop-test-pod -n default

# 清理数据库（可选）
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -c "DROP DATABASE aetherius_orchestrator;"
```

## 下一步

✅ 完成基础测试后，可以：
1. 创建自定义策略和工作流
2. 集成 agent-manager 执行实际命令
3. 添加 AI 分析步骤
4. 配置告警和通知

详细文档：
- `ORCHESTRATOR_TEST_GUIDE.md` - 完整测试指南
- `LOGGING.md` - 日志说明和调试技巧
- `README.md` - 完整功能文档
