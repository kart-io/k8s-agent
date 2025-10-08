# Orchestrator Service 日志说明

## 日志级别和符号

orchestrator-service 现在包含详细的日志输出，使用表情符号便于快速识别日志类型：

### 启动和初始化
- `📦` - 组件初始化
- `✅` - 操作成功
- `❌` - 操作失败
- `⚠️` - 警告信息

### 事件处理
- `📨` - 收到 NATS 消息
- `📬` - 调试消息（所有事件）
- `🔍` - 策略匹配
- `📋` - 数据库查询
- `🚀` - 执行操作
- `🎬` - 工作流启动
- `📝` - 创建记录
- `📊` - 统计信息
- `🎧` - 监听状态

## 日志输出示例

### 1. 启动日志

```
==========================================================
     Aetherius Orchestrator Service - Initialization
==========================================================
📦 [1/6] Initializing PostgreSQL host=localhost port=5432 database=aetherius_orchestrator
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
✓ Subscribed to all internal events (debug) subject=internal.event.>
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

### 2. 接收事件日志

```
📨 Received message on critical channel subject=internal.event.critical size=245
========== Processing Event ==========
✓ Event parsed successfully type=pod_failure cluster_id=test-cluster-001 severity=critical
```

### 3. 策略匹配日志

```
🔍 Starting strategy matching event_type=pod_failure severity=critical
📋 Retrieved strategies from database total_strategies=3
✓ New best match found strategy_name="Pod CrashLoopBackOff Strategy" score=10
✅ Strategy matched successfully strategy_id=strat-pod-crashloop category=pod_failure final_score=10
```

### 4. 执行工作流日志

```
🚀 Starting strategy execution strategy_id=strat-pod-crashloop workflow_id=wf-pod-failure-diagnostic
🎬 Starting workflow execution workflow_id=wf-pod-failure-diagnostic
✓ Workflow loaded workflow_name="Pod Failure Diagnostic Workflow" status=active steps=3
📝 Created workflow execution instance execution_id=a1b2c3d4-...
✓ Execution saved to database
📊 Execution tracking updated total_started=1
✅ Workflow execution started execution_id=a1b2c3d4-... status=pending
========== Strategy execution started successfully ==========
```

### 5. 没有策略匹配的日志

```
🔍 Starting strategy matching event_type=unknown_event severity=low
📋 Retrieved strategies from database total_strategies=3
⚠️  No matching strategy found event_type=unknown_event
⚠️  No strategy matched for event event_type=unknown_event severity=low
```

### 6. 数据库为空的日志

```
🔍 Starting strategy matching event_type=pod_failure
📋 Retrieved strategies from database total_strategies=0
⚠️  No active strategies found in database
```

## 调试技巧

### 1. 查看是否接收到事件

查找 `📨 Received message` 日志：
```bash
make run | grep "📨"
```

### 2. 查看策略匹配

查找策略相关日志：
```bash
make run | grep "🔍\|📋\|✅.*Strategy"
```

### 3. 查看工作流执行

查找工作流日志：
```bash
make run | grep "🎬\|🚀\|📝"
```

### 4. 只看错误和警告

```bash
make run | grep "❌\|⚠️"
```

### 5. 查看所有事件（调试模式）

如果日志级别设置为 debug，可以看到：
```bash
# configs/config.yaml
logging:
  level: "debug"  # 改为 debug
```

然后运行：
```bash
make run | grep "📬"
```

## 日志配置

在 `configs/config.yaml` 中配置日志：

```yaml
logging:
  level: "info"      # debug, info, warn, error
  format: "json"     # json, console
  output_path: "stdout"  # stdout, stderr, or file path
```

### 日志级别说明

- **debug**: 显示所有日志，包括 `📬` 调试消息
- **info**: 显示信息、警告和错误（推荐）
- **warn**: 只显示警告和错误
- **error**: 只显示错误

## 常见问题排查

### 问题 1: 服务启动但没有处理事件

**症状**：看到启动日志但没有 `📨 Received message`

**排查步骤**：
1. 检查 NATS 连接是否成功：`✅ NATS connected successfully`
2. 检查订阅是否成功：`✓ Subscribed to critical events`
3. 使用测试脚本发送事件：`./scripts/test-event.sh`
4. 检查 NATS 是否有消息：`nats sub 'internal.event.*'`

### 问题 2: 收到事件但没有执行策略

**症状**：看到 `📨 Received message` 但显示 `⚠️  No strategy matched`

**排查步骤**：
1. 检查数据库是否有策略：`📋 Retrieved strategies total_strategies=0`
2. 运行策略初始化脚本：`./scripts/test-strategy.sh`
3. 检查事件类型是否匹配策略条件

### 问题 3: 策略匹配但工作流失败

**症状**：看到 `✅ Strategy matched` 但显示 `❌ Failed to start workflow`

**排查步骤**：
1. 检查 workflow_id 是否存在于数据库
2. 检查工作流状态是否为 active
3. 查看详细错误信息

## 性能监控

查看统计信息：
```bash
make run | grep "📊"
```

这会显示：
- 总共启动的执行次数
- 执行状态等信息

## 日志文件输出

如果需要将日志保存到文件：

```yaml
# configs/config.yaml
logging:
  level: "info"
  format: "json"
  output_path: "/var/log/orchestrator-service.log"
```

然后可以使用 `jq` 查看：
```bash
tail -f /var/log/orchestrator-service.log | jq .
```
