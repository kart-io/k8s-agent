# 集成测试指南 - Agent-Manager & Orchestrator-Service

## 🎯 测试目标

验证完整的事件处理流程：
```
事件源 → NATS → orchestrator-service → 策略匹配 → 工作流执行 → 数据库记录
```

## ✅ 前置条件检查

### 1. 服务运行状态

打开 **4 个终端窗口**：

**终端 1: PostgreSQL 端口转发**
```bash
cd deployments/kustomize
make forward-dev-postgres
# 保持运行，看到: Forwarding from 127.0.0.1:5432
```

**终端 2: NATS 端口转发**
```bash
cd deployments/kustomize
make forward-dev-nats
# 保持运行，看到: Forwarding from 127.0.0.1:4222
```

**终端 3: agent-manager**
```bash
cd agent-manager

# 首次运行需要创建数据库
make db-create

# 启动服务
make run

# 应该看到:
# {"level":"info","msg":"Starting Aetherius Agent Manager"}
# {"level":"info","msg":"API server starting","addr":"0.0.0.0:8080"}
```

**终端 4: orchestrator-service**
```bash
cd orchestrator-service

# 首次运行需要创建数据库和策略
make db-create
./scripts/test-strategy.sh

# 启动服务
make run

# 应该看到:
# ✅ Orchestrator Service started successfully!
# 🎧 Listening for events on NATS channels:
#   - internal.event.critical
#   - internal.event.anomaly
```

### 2. 验证连接

```bash
# 检查端口
lsof -i :5432  # PostgreSQL
lsof -i :4222  # NATS
lsof -i :6379  # Redis (可选)
lsof -i :8080  # agent-manager
```

## 🧪 测试方法

### 方法 1: 使用 orchestrator 测试脚本（最简单）

```bash
cd orchestrator-service
./scripts/test-event.sh
```

**期望在 orchestrator-service 终端看到**:
```
📨 Received message on critical channel subject=internal.event.critical size=xxx
========== Processing Event ==========
✓ Event parsed successfully type=pod_failure cluster_id=test-cluster-001
🔍 Starting strategy matching event_type=pod_failure severity=critical
📋 Retrieved strategies from database total_strategies=1
✅ Strategy matched successfully strategy_id=strat-pod-crashloop
🚀 Starting strategy execution strategy_id=strat-pod-crashloop
🎬 Starting workflow execution workflow_id=wf-pod-failure-diagnostic
✓ Workflow loaded workflow_name="Pod Failure Diagnostic Workflow"
📝 Created workflow execution instance execution_id=xxx
✓ Execution saved to database
========== Strategy execution started successfully ==========
```

### 方法 2: 手动构造事件（Python/curl）

**使用 Python**:
```python
#!/usr/bin/env python3
import nats
import asyncio
import json
from datetime import datetime

async def send_event():
    nc = await nats.connect("nats://localhost:4222")

    event = {
        "type": "pod_failure",
        "cluster_id": "test-cluster-001",
        "severity": "critical",
        "payload": {
            "namespace": "default",
            "pod_name": "test-pod",
            "reason": "CrashLoopBackOff"
        },
        "timestamp": datetime.utcnow().isoformat() + "Z"
    }

    await nc.publish("internal.event.critical", json.dumps(event).encode())
    await nc.flush()
    await nc.close()
    print("✅ Event sent!")

asyncio.run(send_event())
```

### 方法 3: 直接查询数据库验证

即使没有收到事件，也可以验证组件是否正常：

**检查 orchestrator 数据库**:
```bash
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c "
-- 查看策略
SELECT id, name, category, enabled FROM strategies;

-- 查看工作流
SELECT id, name, status FROM workflows;

-- 查看最近的执行记录
SELECT id, workflow_id, status, started_at
FROM workflow_executions
ORDER BY started_at DESC
LIMIT 5;
"
```

**检查 agent-manager 数据库**:
```bash
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius -c "
-- 查看代理
SELECT id, cluster_id, status, last_heartbeat FROM agents;

-- 查看事件（如果有表）
SELECT id, type, severity, created_at
FROM events
WHERE created_at > NOW() - INTERVAL '10 minutes'
ORDER BY created_at DESC;
"
```

## 📊 验证清单

### Level 1: 基础连接（必须）
- [ ] PostgreSQL 端口转发正常 (5432)
- [ ] NATS 端口转发正常 (4222)
- [ ] agent-manager 启动成功 (8080)
- [ ] orchestrator-service 启动成功

### Level 2: 数据库初始化（必须）
- [ ] orchestrator 数据库已创建 (`aetherius_orchestrator`)
- [ ] agent-manager 数据库已创建 (`aetherius`)
- [ ] 策略表有数据 (`strategies`)
- [ ] 工作流表有数据 (`workflows`)

### Level 3: 事件处理（核心测试）
- [ ] orchestrator 收到 NATS 消息 (📨)
- [ ] 事件解析成功 (✓ Event parsed)
- [ ] 策略匹配成功 (✅ Strategy matched)
- [ ] 工作流启动 (🎬 Starting workflow)
- [ ] 数据库有执行记录 (`workflow_executions`)

### Level 4: 完整流程（进阶）
- [ ] agent-manager 接收事件
- [ ] agent-manager 发布到 NATS
- [ ] orchestrator 接收并处理
- [ ] 工作流步骤执行
- [ ] 结果保存到数据库

## 🔍 故障排查

### 问题 1: orchestrator 收不到事件

**检查项**:
```bash
# 1. NATS 连接
nc -zv localhost 4222

# 2. orchestrator 订阅状态
# 在 orchestrator 日志中查找:
# ✓ Subscribed to critical events subject=internal.event.critical

# 3. 手动监听 NATS (需要 nats CLI)
nats sub "internal.event.*" --server=nats://localhost:4222
# 然后在另一个终端发送事件，看能否收到
```

**解决方案**:
- 确保 NATS 端口转发运行
- 重启 orchestrator-service
- 检查日志级别（应该能看到 debug 日志）

### 问题 2: 策略不匹配

**症状**: 收到事件但显示 `⚠️ No strategy matched`

**检查**:
```bash
# 查看策略
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, name, category, enabled, symptoms
   FROM strategies
   WHERE enabled = true;"
```

**解决方案**:
```bash
cd orchestrator-service
./scripts/test-strategy.sh  # 重新创建策略
```

### 问题 3: 工作流启动失败

**症状**: `❌ Failed to load workflow from database`

**检查**:
```bash
# 查看工作流
kubectl exec -n aetherius-dev deployment/postgres -- \
  psql -U postgres -d aetherius_orchestrator -c \
  "SELECT id, name, status FROM workflows;"
```

**解决方案**:
- 确保 `test-strategy.sh` 运行成功
- 检查工作流 ID 是否匹配策略中的 `workflow_id`

### 问题 4: 日志太多看不清

**agent-manager**:
```bash
# 修改 agent-manager/configs/config.yaml
logging:
  level: "info"  # 从 debug 改为 info
```

**orchestrator-service**:
```bash
# 过滤关键日志
make run | grep -E "📨|🔍|✅|🚀|❌"
```

## 📈 成功指标

完整的成功流程日志应该包含：

**orchestrator-service**:
```
[启动]
✅ PostgreSQL initialized successfully
✅ NATS connected successfully
✅ Event subscriber started successfully
🎧 Listening for events on NATS channels

[收到事件]
📨 Received message on critical channel
✓ Event parsed successfully

[处理事件]
🔍 Starting strategy matching
📋 Retrieved strategies from database total_strategies=1
✅ Strategy matched successfully

[执行工作流]
🚀 Starting strategy execution
🎬 Starting workflow execution
✓ Workflow loaded
📝 Created workflow execution instance
✓ Execution saved to database
========== Strategy execution started successfully ==========
```

**数据库记录**:
```sql
-- 应该有新记录
SELECT COUNT(*) FROM workflow_executions
WHERE started_at > NOW() - INTERVAL '1 minute';
-- 结果应该 > 0
```

## 🎉 测试成功后

恭喜！你已经验证了：
1. ✅ 两个服务可以正常通信
2. ✅ NATS 消息传递正常
3. ✅ 事件处理链路完整
4. ✅ 数据库读写正常
5. ✅ 策略匹配工作正常
6. ✅ 工作流引擎可以启动

## 🚀 下一步

- 部署 collect-agent 捕获真实 K8s 事件
- 创建自定义策略和工作流
- 集成 AI 分析服务
- 添加命令执行功能
- 配置告警通知
