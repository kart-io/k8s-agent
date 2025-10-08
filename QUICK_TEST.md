# 快速测试 - 5 分钟验证

## 🚀 最简单的测试方法

### 前提条件
确保 4 个终端窗口都在运行：
1. `make forward-dev-postgres`
2. `make forward-dev-nats`
3. `cd agent-manager && make run`
4. `cd orchestrator-service && make run`

### 一键测试

**在新的终端窗口执行**:
```bash
cd orchestrator-service
./scripts/test-event.sh
```

**立即查看 orchestrator-service 的终端（终端4）**，应该看到：

```json
{"level":"info","msg":"📨 Received message on critical channel","subject":"internal.event.critical"}
{"level":"info","msg":"========== Processing Event =========="}
{"level":"info","msg":"✓ Event parsed successfully","type":"pod_failure"}
{"level":"info","msg":"🔍 Starting strategy matching","event_type":"pod_failure"}
{"level":"info","msg":"📋 Retrieved strategies from database","total_strategies":1}
{"level":"info","msg":"✅ Strategy matched successfully","strategy_id":"strat-pod-crashloop"}
{"level":"info","msg":"🚀 Starting strategy execution"}
{"level":"info","msg":"🎬 Starting workflow execution"}
{"level":"info","msg":"✓ Workflow loaded","workflow_name":"Pod Failure Diagnostic Workflow"}
{"level":"info","msg":"========== Strategy execution started successfully =========="}
```

### ✅ 成功标志

如果看到上面的日志，说明：
1. ✅ NATS 消息传递正常
2. ✅ orchestrator-service 接收事件正常
3. ✅ 策略匹配成功
4. ✅ 工作流启动成功
5. ✅ 整个流程打通！

### ❌ 如果没有日志

**可能的原因**:

1. **NATS 端口转发没运行**
   ```bash
   lsof -i :4222  # 应该有输出
   ```

2. **策略没有创建**
   ```bash
   cd orchestrator-service
   ./scripts/test-strategy.sh
   ```

3. **orchestrator 没启动**
   - 检查终端 4 是否有 "Orchestrator Service started successfully"

### 🔍 验证数据库记录

```bash
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, workflow_id, status, started_at
   FROM workflow_executions
   ORDER BY started_at DESC LIMIT 1;"
```

应该看到一条新记录！

## 🎯 测试完成

恭喜！你已经成功验证了 agent-manager 和 orchestrator-service 的集成。

**整个流程**:
```
test-event.sh
  → 发送 JSON 到 NATS (internal.event.critical)
    → orchestrator-service 接收
      → 匹配策略
        → 启动工作流
          → 保存到数据库 ✅
```
