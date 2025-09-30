# Aetherius Orchestrator Service

任务编排服务 (Layer 3),负责工作流编排、诊断策略执行和自动化修复。

---

## 功能特性

### 核心功能

- **工作流引擎**: 灵活的步骤编排和执行
- **诊断策略**: 自动匹配故障模式并执行对应工作流
- **步骤执行器**: 支持多种步骤类型 (命令、AI分析、决策、修复等)
- **事件订阅**: 监听 agent-manager 的内部事件总线
- **AI 集成**: 调用 reasoning-service 进行智能分析

### 支持的步骤类型

- **Command**: 通过 agent-manager 执行远程命令
- **AI Analysis**: 调用 AI 服务进行根因分析
- **Decision**: 条件判断和分支控制
- **Remediation**: 执行自动化修复动作
- **Notification**: 发送告警通知
- **Wait**: 等待指定时间
- **Parallel**: 并行执行多个步骤 (规划中)

---

## 架构设计

```plaintext
┌──────────────────────────────────────────────────────────────┐
│             Orchestrator Service (Layer 3)                    │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ Workflow       │  │ Strategy     │  │ Event          │  │
│  │ Engine         │  │ Manager      │  │ Subscriber     │  │
│  │                │  │              │  │                │  │
│  │ - Execute      │  │ - Match      │  │ - Listen       │  │
│  │ - Track        │  │ - Trigger    │  │ - Route        │  │
│  └────────────────┘  └──────────────┘  └────────────────┘  │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐ │
│  │             Step Executor                               │ │
│  │  - Command (via agent-manager)                         │ │
│  │  - AI Analysis (via reasoning-service)                 │ │
│  │  - Decision (condition evaluation)                     │ │
│  │  - Remediation (auto-fix actions)                      │ │
│  │  - Notification (alerts)                               │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                               │
│  ↓ NATS Subscribe                   ↑ HTTP Calls             │
└───┼──────────────────────────────────┼──────────────────────┘
    │                                  │
    │ internal.event.critical          │
    │ internal.event.anomaly           │
    │                                  │
  agent-manager                   reasoning-service
```

---

## 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- NATS Server 2.10+
- agent-manager (运行中)
- reasoning-service (可选,用于 AI 分析)

### 本地运行

```bash
# 1. 安装依赖
go mod download

# 2. 启动依赖服务 (PostgreSQL, Redis, NATS)
docker-compose up -d postgres redis nats

# 3. 配置文件
cp configs/config.yaml configs/config.local.yaml
# 编辑 config.local.yaml 设置正确的连接信息

# 4. 运行服务
go run ./cmd/server --config=configs/config.local.yaml
```

### 验证运行

```bash
# 查看日志
# 服务启动后会自动订阅内部事件总线
# 等待 agent-manager 发布关键事件
```

---

## 工作流定义

### 工作流结构

```yaml
id: "diagnose_pod_crashloop"
name: "诊断 Pod CrashLoopBackOff"
description: "自动诊断和修复 Pod 崩溃循环问题"
trigger_type: "event"
trigger_config:
  event_reason: "CrashLoopBackOff"
  severity: "high"
status: "active"
priority: 10
timeout: "5m"

steps:
  - id: "collect_logs"
    type: "command"
    name: "收集容器日志"
    config:
      tool: "kubectl"
      action: "logs"
      args:
        - "--tail=100"
        - "--previous"
    timeout: "30s"
    on_success: ["check_resources"]
    on_failure: ["notify_failure"]

  - id: "check_resources"
    type: "command"
    name: "检查资源状态"
    config:
      tool: "kubectl"
      action: "describe"
    timeout: "30s"
    on_success: ["ai_analysis"]

  - id: "ai_analysis"
    type: "ai_analysis"
    name: "AI 根因分析"
    config:
      analysis_type: "root_cause"
    timeout: "60s"
    on_success: ["decide_action"]

  - id: "decide_action"
    type: "decision"
    name: "决策修复动作"
    config:
      conditions:
        - if: "analysis.root_cause == 'OOM'"
          then: "increase_memory"
        - if: "analysis.root_cause == 'Config'"
          then: "notify_owner"
    on_success: ["execute_remediation"]

  - id: "execute_remediation"
    type: "remediation"
    name: "执行修复"
    config:
      action_type: "kubectl"
      action: "patch_deployment"
    on_success: ["notify_success"]
    on_failure: ["notify_failure"]

  - id: "notify_success"
    type: "notification"
    name: "成功通知"
    config:
      channel: "slack"
      message: "Pod 问题已自动修复"

  - id: "notify_failure"
    type: "notification"
    name: "失败通知"
    config:
      channel: "slack"
      message: "自动修复失败,需要人工介入"
```

### 创建工作流

工作流通过 PostgreSQL 存储,可以通过以下方式创建:

1. **数据库直接插入**:

```sql
INSERT INTO workflows (id, name, description, trigger_type, status, steps, created_at, updated_at)
VALUES (
  'diagnose_pod_crashloop',
  '诊断 Pod CrashLoopBackOff',
  '自动诊断和修复 Pod 崩溃循环问题',
  'event',
  'active',
  '[]'::jsonb,  -- 步骤定义
  NOW(),
  NOW()
);
```

2. **通过 API** (需要先实现 API 端点):

```bash
curl -X POST http://localhost:8081/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d @workflow-definition.json
```

---

## 诊断策略

### 策略结构

```go
type Strategy struct {
    ID          string
    Name        string
    Category    string  // pod_failure, node_issue, network, etc.
    Symptoms    []Symptom
    WorkflowID  string  // 关联的工作流ID
    Priority    int
    Enabled     bool
}

type Symptom struct {
    Type       string  // event, metric, log
    Pattern    string  // 匹配模式
    Conditions map[string]interface{}
}
```

### 内置策略示例

1. **Pod CrashLoopBackOff**
   - 症状: K8s 事件 reason = "CrashLoopBackOff"
   - 工作流: `diagnose_pod_crashloop`
   - 优先级: 10

2. **Node NotReady**
   - 症状: K8s 事件 reason = "NodeNotReady"
   - 工作流: `diagnose_node_not_ready`
   - 优先级: 9

3. **OOM Killed**
   - 症状: K8s 事件 reason = "OOMKilling"
   - 工作流: `diagnose_oom_killed`
   - 优先级: 10

4. **PVC Pending**
   - 症状: K8s 事件 reason = "FailedBinding"
   - 工作流: `diagnose_pvc_pending`
   - 优先级: 7

---

## 工作流执行流程

### 1. 事件触发

```plaintext
agent-manager 发布内部事件
    ↓ NATS: internal.event.critical
Orchestrator 接收事件
    ↓
策略匹配 (Strategy Manager)
    ↓
启动工作流执行
```

### 2. 步骤执行

```plaintext
Workflow Engine
    ↓
按顺序执行步骤
    ├─> 准备输入 (从执行上下文)
    ├─> 调用 Executor
    │   ├─> Command: 调用 agent-manager API
    │   ├─> AI Analysis: 调用 reasoning-service API
    │   ├─> Decision: 评估条件
    │   ├─> Remediation: 执行修复动作
    │   └─> Notification: 发送通知
    ├─> 处理输出
    ├─> 更新上下文
    └─> 继续下一步 or 完成
```

### 3. 错误处理

- **重试机制**: 支持指数退避重试
- **失败分支**: on_failure 定义失败后的步骤
- **超时控制**: 每个步骤和整个工作流都有超时设置

---

## 配置说明

### 关键配置项

```yaml
# NATS 连接 (订阅内部事件)
nats:
  url: "nats://agent-manager:4222"

# AI 服务
ai:
  reasoning_service_url: "http://reasoning-service:8082"
  timeout: 30s
  max_retries: 3

# 数据库 (存储工作流和执行历史)
database:
  host: "postgres"
  database: "aetherius_orchestrator"
```

---

## 监控和调试

### 日志

```bash
# 查看工作流执行日志
tail -f orchestrator.log | grep "Workflow execution"

# 查看步骤执行日志
tail -f orchestrator.log | grep "Executing.*step"

# 查看策略匹配日志
tail -f orchestrator.log | grep "Strategy matched"
```

### 数据库查询

```sql
-- 查看所有工作流
SELECT id, name, status, created_at FROM workflows;

-- 查看最近的执行
SELECT id, workflow_id, status, started_at, duration
FROM workflow_executions
ORDER BY started_at DESC
LIMIT 10;

-- 查看执行详情
SELECT * FROM workflow_executions WHERE id = 'execution-id';

-- 查看活跃策略
SELECT id, name, category, priority FROM strategies WHERE enabled = true;
```

---

## 开发指南

### 添加新的步骤类型

1. 在 `types.go` 添加步骤类型常量:

```go
const (
    StepTypeCustom StepType = "custom"
)
```

2. 在 `executor.go` 实现执行逻辑:

```go
func (ex *Executor) ExecuteCustom(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
    // 实现自定义逻辑
    return output, nil
}
```

3. 在 `engine.go` 的 `executeStep` 中添加分支:

```go
case types.StepTypeCustom:
    stepExec.Output, err = e.executor.ExecuteCustom(ctx, execution, step)
```

### 添加新的诊断策略

1. 在数据库中插入策略:

```sql
INSERT INTO strategies (id, name, category, symptoms, workflow_id, priority, enabled)
VALUES (...);
```

2. 创建对应的工作流
3. 重启服务或热加载

---

## 性能优化

### 并发执行

- 使用 goroutine 异步执行工作流
- 避免阻塞主线程

### 数据库优化

- 为高频查询字段添加索引
- 定期清理旧的执行记录

### 缓存策略

- Redis 缓存活跃策略
- 内存缓存工作流定义

---

## 故障排查

### 问题 1: 工作流未触发

**检查**:
- agent-manager 是否发布了内部事件
- NATS 连接是否正常
- 策略是否启用 (enabled = true)
- 症状匹配是否正确

### 问题 2: 步骤执行失败

**检查**:
- agent-manager API 是否可访问
- reasoning-service 是否运行
- 步骤配置是否正确
- 超时设置是否合理

### 问题 3: AI 分析无响应

**检查**:
- reasoning-service URL 配置
- 网络连通性
- AI 服务日志

---

## 路线图

- [ ] RESTful API 实现
- [ ] Temporal 工作流引擎集成 (替代自研引擎)
- [ ] 并行步骤执行
- [ ] 工作流可视化编辑器
- [ ] 更多内置策略
- [ ] 修复动作审批流程
- [ ] 工作流版本控制
- [ ] A/B 测试支持

---

## 许可证

MIT License

---

## 相关文档

- [系统架构](../../docs/architecture/SYSTEM_ARCHITECTURE.md)
- [agent-manager](../agent-manager/README.md)
- [collect-agent](../collect-agent/README.md)