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

### 配置文件结构

配置文件位于 `configs/config.yaml`,支持通过环境变量覆盖。

### HTTP 服务器配置

```yaml
server:
  host: "0.0.0.0"              # 监听地址
  port: 8081                   # 服务端口
  read_timeout: 30s            # 读取超时
  write_timeout: 30s           # 写入超时
  graceful_stop: 10s           # 优雅关闭等待时间
```

**环境变量**:
- `SERVER_HOST` - 覆盖监听地址
- `SERVER_PORT` - 覆盖服务端口

### NATS 消息总线配置

```yaml
nats:
  url: "nats://localhost:4222"  # NATS 服务器地址
  max_reconnect: -1             # 最大重连次数 (-1 表示无限重连)
  reconnect_wait: 2s            # 重连等待时间
```

**订阅主题**:
- `internal.event.critical` - 关键事件 (severity >= high)
- `internal.event.anomaly` - 异常事件

**环境变量**:
- `NATS_URL` - 覆盖 NATS 服务器地址

**注意事项**:
- 必须连接到 agent-manager 的内部 NATS 服务器
- 确保网络策略允许 orchestrator-service 访问 agent-manager 的 4222 端口
- 重连机制确保临时网络故障不会导致事件丢失

### PostgreSQL 数据库配置

```yaml
database:
  host: "localhost"             # 数据库主机
  port: 5432                    # 数据库端口
  user: "aetherius"             # 数据库用户
  password: "aetherius123"      # 数据库密码
  database: "aetherius_orchestrator"  # 数据库名称
  ssl_mode: "disable"           # SSL 模式 (disable/require/verify-ca/verify-full)
  max_open_conns: 25            # 最大打开连接数
  max_idle_conns: 5             # 最大空闲连接数
  conn_max_lifetime: 300s       # 连接最大生命周期
```

**环境变量**:
- `DB_HOST` - 数据库主机
- `DB_PORT` - 数据库端口
- `DB_USER` - 数据库用户
- `DB_PASSWORD` - 数据库密码
- `DB_NAME` - 数据库名称
- `DB_SSL_MODE` - SSL 模式

**数据库表结构**:
- `workflows` - 工作流定义
- `strategies` - 诊断策略
- `workflow_executions` - 工作流执行历史
- `step_executions` - 步骤执行详情

**性能建议**:
- 生产环境建议 `max_open_conns: 100`
- 启用连接池可显著提升性能
- 定期清理历史执行记录 (保留 30 天)

### Redis 缓存配置

```yaml
redis:
  addr: "localhost:6379"        # Redis 地址
  password: ""                  # Redis 密码 (可选)
  db: 1                         # Redis 数据库编号
  pool_size: 10                 # 连接池大小
  min_idle_conns: 3             # 最小空闲连接数
  dial_timeout: 5s              # 连接超时
```

**环境变量**:
- `REDIS_ADDR` - Redis 地址
- `REDIS_PASSWORD` - Redis 密码
- `REDIS_DB` - 数据库编号

**用途**:
- 缓存活跃策略定义 (TTL: 5分钟)
- 缓存工作流定义 (TTL: 10分钟)
- 分布式锁 (防止重复执行)
- 执行状态临时存储

### AI 服务配置

```yaml
ai:
  reasoning_service_url: "http://localhost:8082"  # Reasoning Service 地址
  timeout: 30s                                     # 请求超时
  max_retries: 3                                   # 最大重试次数
```

**环境变量**:
- `AI_SERVICE_URL` - Reasoning Service 地址
- `AI_TIMEOUT` - 请求超时时间
- `AI_MAX_RETRIES` - 最大重试次数

**调用场景**:
- 工作流步骤类型为 `ai_analysis` 时调用
- 自动进行根因分析
- 获取推荐修复动作

**API 端点**:
- `POST /api/v1/analyze/root-cause` - 根因分析
- `POST /api/v1/analyze/predict` - 故障预测
- `POST /api/v1/analyze/recommend` - 修复建议

### Temporal 工作流引擎配置 (可选)

```yaml
temporal:
  host_port: "localhost:7233"          # Temporal 服务器地址
  namespace: "default"                  # 命名空间
  task_queue: "aetherius-orchestrator" # 任务队列名称
```

**注意**: 当前版本使用自研工作流引擎,Temporal 集成在规划中。

### 日志配置

```yaml
logging:
  level: "info"          # 日志级别 (debug/info/warn/error)
  format: "json"         # 日志格式 (json/console)
  output_path: "stdout"  # 输出路径 (stdout/文件路径)
```

**环境变量**:
- `LOG_LEVEL` - 日志级别
- `LOG_FORMAT` - 日志格式

**日志级别说明**:
- `debug` - 开发调试,包含详细执行信息
- `info` - 生产推荐,包含关键操作日志
- `warn` - 仅警告和错误
- `error` - 仅错误信息

### 完整配置示例

```yaml
# configs/config.yaml - 生产环境配置
server:
  host: "0.0.0.0"
  port: 8081
  read_timeout: 30s
  write_timeout: 30s
  graceful_stop: 10s

nats:
  url: "nats://agent-manager-svc.aetherius.svc.cluster.local:4222"
  max_reconnect: -1
  reconnect_wait: 2s

database:
  host: "postgres-svc.aetherius.svc.cluster.local"
  port: 5432
  user: "aetherius"
  password: "${DB_PASSWORD}"  # 从环境变量读取
  database: "aetherius_orchestrator"
  ssl_mode: "require"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 300s

redis:
  addr: "redis-svc.aetherius.svc.cluster.local:6379"
  password: "${REDIS_PASSWORD}"
  db: 1
  pool_size: 20
  min_idle_conns: 5
  dial_timeout: 5s

ai:
  reasoning_service_url: "http://reasoning-service-svc.aetherius.svc.cluster.local:8082"
  timeout: 60s
  max_retries: 3

logging:
  level: "info"
  format: "json"
  output_path: "/var/log/orchestrator/orchestrator.log"
```

---

## API 文档

Orchestrator Service 提供 RESTful API 用于工作流和策略管理。

### 基础信息

- **Base URL**: `http://localhost:8081`
- **Content-Type**: `application/json`
- **认证**: 暂未实现 (规划中)

### 健康检查

#### GET /health

检查服务健康状态。

**请求示例**:
```bash
curl http://localhost:8081/health
```

**响应示例**:
```json
{
  "status": "healthy",
  "service": "orchestrator-service",
  "version": "v1.0.0",
  "components": {
    "database": "connected",
    "redis": "connected",
    "nats": "connected"
  },
  "timestamp": "2025-10-01T10:00:00Z"
}
```

### 工作流管理

#### GET /api/v1/workflows

获取所有工作流列表。

**查询参数**:
- `status` (可选): 过滤状态 (`active`, `inactive`, `all`)
- `page` (可选): 页码,默认 1
- `page_size` (可选): 每页数量,默认 20

**请求示例**:
```bash
curl "http://localhost:8081/api/v1/workflows?status=active&page=1&page_size=10"
```

**响应示例**:
```json
{
  "workflows": [
    {
      "id": "diagnose_pod_crashloop",
      "name": "诊断 Pod CrashLoopBackOff",
      "description": "自动诊断和修复 Pod 崩溃循环问题",
      "trigger_type": "event",
      "status": "active",
      "priority": 10,
      "created_at": "2025-09-01T10:00:00Z",
      "updated_at": "2025-09-15T14:30:00Z"
    }
  ],
  "total": 15,
  "page": 1,
  "page_size": 10
}
```

#### GET /api/v1/workflows/:id

获取指定工作流详情。

**请求示例**:
```bash
curl http://localhost:8081/api/v1/workflows/diagnose_pod_crashloop
```

**响应示例**:
```json
{
  "id": "diagnose_pod_crashloop",
  "name": "诊断 Pod CrashLoopBackOff",
  "description": "自动诊断和修复 Pod 崩溃循环问题",
  "trigger_type": "event",
  "trigger_config": {
    "event_reason": "CrashLoopBackOff",
    "severity": "high"
  },
  "status": "active",
  "priority": 10,
  "timeout": "5m",
  "steps": [
    {
      "id": "collect_logs",
      "type": "command",
      "name": "收集容器日志",
      "config": {
        "tool": "kubectl",
        "action": "logs",
        "args": ["--tail=100", "--previous"]
      },
      "timeout": "30s",
      "on_success": ["check_resources"],
      "on_failure": ["notify_failure"]
    }
  ],
  "created_at": "2025-09-01T10:00:00Z",
  "updated_at": "2025-09-15T14:30:00Z"
}
```

#### POST /api/v1/workflows

创建新工作流。

**请求体**:
```json
{
  "id": "diagnose_custom_issue",
  "name": "诊断自定义问题",
  "description": "自定义诊断工作流",
  "trigger_type": "event",
  "trigger_config": {
    "event_reason": "CustomReason"
  },
  "status": "active",
  "priority": 5,
  "timeout": "10m",
  "steps": [
    {
      "id": "step1",
      "type": "command",
      "name": "执行诊断命令",
      "config": {
        "tool": "kubectl",
        "action": "get"
      }
    }
  ]
}
```

**响应示例**:
```json
{
  "id": "diagnose_custom_issue",
  "message": "Workflow created successfully",
  "created_at": "2025-10-01T10:00:00Z"
}
```

#### PUT /api/v1/workflows/:id

更新工作流。

**请求体**: 同创建工作流

**响应示例**:
```json
{
  "id": "diagnose_custom_issue",
  "message": "Workflow updated successfully",
  "updated_at": "2025-10-01T10:30:00Z"
}
```

#### DELETE /api/v1/workflows/:id

删除工作流。

**请求示例**:
```bash
curl -X DELETE http://localhost:8081/api/v1/workflows/diagnose_custom_issue
```

**响应示例**:
```json
{
  "message": "Workflow deleted successfully"
}
```

### 工作流执行

#### GET /api/v1/executions

获取工作流执行历史。

**查询参数**:
- `workflow_id` (可选): 按工作流ID过滤
- `status` (可选): 按状态过滤 (`running`, `completed`, `failed`)
- `start_time` (可选): 开始时间 (RFC3339格式)
- `end_time` (可选): 结束时间
- `page` (可选): 页码
- `page_size` (可选): 每页数量

**请求示例**:
```bash
curl "http://localhost:8081/api/v1/executions?workflow_id=diagnose_pod_crashloop&status=completed&page=1"
```

**响应示例**:
```json
{
  "executions": [
    {
      "id": "exec-20251001-001",
      "workflow_id": "diagnose_pod_crashloop",
      "workflow_name": "诊断 Pod CrashLoopBackOff",
      "status": "completed",
      "trigger_event": {
        "type": "k8s_event",
        "reason": "CrashLoopBackOff",
        "cluster_id": "prod-cluster-01"
      },
      "started_at": "2025-10-01T09:00:00Z",
      "completed_at": "2025-10-01T09:02:30Z",
      "duration": "2m30s",
      "step_count": 7,
      "steps_completed": 7,
      "steps_failed": 0
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 20
}
```

#### GET /api/v1/executions/:id

获取执行详情。

**请求示例**:
```bash
curl http://localhost:8081/api/v1/executions/exec-20251001-001
```

**响应示例**:
```json
{
  "id": "exec-20251001-001",
  "workflow_id": "diagnose_pod_crashloop",
  "workflow_name": "诊断 Pod CrashLoopBackOff",
  "status": "completed",
  "trigger_event": {
    "type": "k8s_event",
    "reason": "CrashLoopBackOff",
    "resource_type": "Pod",
    "resource_name": "my-app-7d8f9c",
    "namespace": "production",
    "cluster_id": "prod-cluster-01"
  },
  "context": {
    "cluster_id": "prod-cluster-01",
    "namespace": "production",
    "pod_name": "my-app-7d8f9c"
  },
  "steps": [
    {
      "id": "collect_logs",
      "name": "收集容器日志",
      "type": "command",
      "status": "completed",
      "started_at": "2025-10-01T09:00:01Z",
      "completed_at": "2025-10-01T09:00:25Z",
      "duration": "24s",
      "output": {
        "command_id": "cmd-001",
        "logs": "..."
      },
      "error": null
    },
    {
      "id": "ai_analysis",
      "name": "AI 根因分析",
      "type": "ai_analysis",
      "status": "completed",
      "started_at": "2025-10-01T09:00:45Z",
      "completed_at": "2025-10-01T09:01:30Z",
      "duration": "45s",
      "output": {
        "root_cause": "OOMKilled",
        "confidence": 0.95,
        "recommendations": [
          "增加内存限制至 512Mi"
        ]
      }
    }
  ],
  "started_at": "2025-10-01T09:00:00Z",
  "completed_at": "2025-10-01T09:02:30Z",
  "duration": "2m30s"
}
```

#### POST /api/v1/executions/:id/cancel

取消正在运行的工作流执行。

**请求示例**:
```bash
curl -X POST http://localhost:8081/api/v1/executions/exec-20251001-002/cancel
```

**响应示例**:
```json
{
  "id": "exec-20251001-002",
  "message": "Execution cancelled",
  "status": "cancelled"
}
```

### 策略管理

#### GET /api/v1/strategies

获取所有诊断策略。

**查询参数**:
- `enabled` (可选): 过滤启用状态 (`true`, `false`, `all`)
- `category` (可选): 按类别过滤

**请求示例**:
```bash
curl "http://localhost:8081/api/v1/strategies?enabled=true"
```

**响应示例**:
```json
{
  "strategies": [
    {
      "id": "strategy_pod_crashloop",
      "name": "Pod CrashLoopBackOff 策略",
      "category": "pod_failure",
      "workflow_id": "diagnose_pod_crashloop",
      "priority": 10,
      "enabled": true,
      "created_at": "2025-09-01T10:00:00Z"
    }
  ],
  "total": 12
}
```

#### GET /api/v1/strategies/:id

获取策略详情。

**响应示例**:
```json
{
  "id": "strategy_pod_crashloop",
  "name": "Pod CrashLoopBackOff 策略",
  "category": "pod_failure",
  "description": "检测并处理 Pod 崩溃循环",
  "symptoms": [
    {
      "type": "event",
      "pattern": "CrashLoopBackOff",
      "conditions": {
        "severity": "high",
        "resource_type": "Pod"
      }
    }
  ],
  "workflow_id": "diagnose_pod_crashloop",
  "priority": 10,
  "enabled": true,
  "match_count": 45,
  "success_rate": 0.89,
  "created_at": "2025-09-01T10:00:00Z",
  "updated_at": "2025-09-15T14:30:00Z"
}
```

#### POST /api/v1/strategies

创建新策略。

**请求体**:
```json
{
  "id": "strategy_custom",
  "name": "自定义策略",
  "category": "custom",
  "symptoms": [
    {
      "type": "event",
      "pattern": "CustomPattern",
      "conditions": {
        "severity": "medium"
      }
    }
  ],
  "workflow_id": "diagnose_custom_issue",
  "priority": 5,
  "enabled": true
}
```

#### PUT /api/v1/strategies/:id/enable

启用策略。

#### PUT /api/v1/strategies/:id/disable

禁用策略。

### 统计信息

#### GET /api/v1/stats/overview

获取系统概览统计。

**响应示例**:
```json
{
  "workflows": {
    "total": 15,
    "active": 12,
    "inactive": 3
  },
  "strategies": {
    "total": 12,
    "enabled": 10,
    "disabled": 2
  },
  "executions": {
    "total": 1250,
    "today": 45,
    "running": 3,
    "completed": 1100,
    "failed": 147,
    "success_rate": 0.88
  },
  "performance": {
    "avg_execution_time": "2m15s",
    "avg_steps_per_workflow": 5.2
  }
}
```

#### GET /api/v1/stats/workflows/:id

获取指定工作流统计。

**响应示例**:
```json
{
  "workflow_id": "diagnose_pod_crashloop",
  "workflow_name": "诊断 Pod CrashLoopBackOff",
  "executions": {
    "total": 320,
    "completed": 285,
    "failed": 35,
    "success_rate": 0.89
  },
  "performance": {
    "avg_duration": "2m30s",
    "min_duration": "1m15s",
    "max_duration": "4m50s"
  },
  "trend": {
    "last_7_days": [15, 18, 22, 19, 25, 20, 23],
    "last_30_days_success_rate": 0.91
  }
}
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

## 工作流调试指南

### 1. 启用调试日志

编辑配置文件,将日志级别设置为 `debug`:

```yaml
logging:
  level: "debug"
  format: "json"
```

重启服务后,将输出详细的执行信息。

### 2. 使用日志追踪工作流执行

**追踪工作流触发**:
```bash
# 查看工作流触发日志
tail -f orchestrator.log | grep "Workflow triggered"

# 示例输出:
# {"level":"info","msg":"Workflow triggered","workflow_id":"diagnose_pod_crashloop","execution_id":"exec-001","trigger":"event"}
```

**追踪步骤执行**:
```bash
# 查看步骤执行日志
tail -f orchestrator.log | grep "Executing.*step"

# 示例输出:
# {"level":"info","msg":"Executing step","execution_id":"exec-001","step_id":"collect_logs","step_type":"command"}
# {"level":"info","msg":"Step completed","execution_id":"exec-001","step_id":"collect_logs","duration":"24s","status":"success"}
```

**追踪策略匹配**:
```bash
# 查看策略匹配日志
tail -f orchestrator.log | grep "Strategy"

# 示例输出:
# {"level":"info","msg":"Strategy matched","strategy_id":"strategy_pod_crashloop","event_reason":"CrashLoopBackOff"}
```

### 3. 数据库调试查询

**查看工作流执行详情**:
```sql
-- 查看最近的执行
SELECT
  id,
  workflow_id,
  status,
  started_at,
  completed_at,
  EXTRACT(EPOCH FROM (completed_at - started_at)) AS duration_seconds
FROM workflow_executions
ORDER BY started_at DESC
LIMIT 10;

-- 查看失败的执行
SELECT
  id,
  workflow_id,
  status,
  error_message,
  started_at
FROM workflow_executions
WHERE status = 'failed'
ORDER BY started_at DESC
LIMIT 20;

-- 查看特定执行的步骤详情
SELECT
  step_id,
  step_name,
  status,
  started_at,
  completed_at,
  error_message,
  output
FROM step_executions
WHERE execution_id = 'exec-20251001-001'
ORDER BY step_order;
```

**分析工作流性能**:
```sql
-- 统计各工作流的平均执行时间
SELECT
  workflow_id,
  COUNT(*) as execution_count,
  AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration_seconds,
  MIN(EXTRACT(EPOCH FROM (completed_at - started_at))) as min_duration_seconds,
  MAX(EXTRACT(EPOCH FROM (completed_at - started_at))) as max_duration_seconds
FROM workflow_executions
WHERE status = 'completed'
  AND completed_at > NOW() - INTERVAL '7 days'
GROUP BY workflow_id
ORDER BY avg_duration_seconds DESC;

-- 统计步骤失败率
SELECT
  step_id,
  step_name,
  COUNT(*) as total_executions,
  SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count,
  ROUND(100.0 * SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) / COUNT(*), 2) as failure_rate
FROM step_executions
WHERE started_at > NOW() - INTERVAL '7 days'
GROUP BY step_id, step_name
ORDER BY failure_rate DESC;
```

### 4. 调试 NATS 连接

**检查 NATS 连接状态**:
```bash
# 查看 NATS 连接日志
tail -f orchestrator.log | grep "NATS"

# 正常输出:
# {"level":"info","msg":"NATS connected","url":"nats://agent-manager:4222"}
# {"level":"info","msg":"NATS subscribed","subject":"internal.event.critical"}
```

**手动测试 NATS 订阅**:
```bash
# 使用 nats CLI 工具测试
nats sub "internal.event.critical" --server=nats://agent-manager:4222

# 发布测试消息
nats pub "internal.event.critical" '{"type":"test"}' --server=nats://agent-manager:4222
```

### 5. 调试步骤执行

**Command 步骤调试**:

检查点:
1. agent-manager API 是否可访问
2. cluster_id 是否正确传递
3. 命令参数是否正确

```bash
# 手动测试 agent-manager API
curl -X POST http://agent-manager:8080/api/v1/commands \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "prod-cluster-01",
    "type": "diagnostic",
    "tool": "kubectl",
    "action": "get",
    "args": ["pods"]
  }'
```

**AI Analysis 步骤调试**:

检查点:
1. reasoning-service URL 配置
2. 网络连通性
3. 请求超时设置

```bash
# 手动测试 reasoning-service API
curl -X POST http://reasoning-service:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test-001",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff"
      }
    }
  }'
```

### 6. 使用 Makefile 命令调试

```bash
# 查看健康状态
make health-check

# 查看服务日志
make logs

# 查看 Kubernetes 部署状态 (如果部署在 K8s 中)
make k8s-status

# 连接到数据库
make psql

# 连接到 Redis
make redis-cli
```

### 7. 常见调试场景

**场景 1: 工作流未按预期触发**

调试步骤:
```bash
# 1. 检查 NATS 连接
tail -f orchestrator.log | grep "NATS"

# 2. 检查策略匹配日志
tail -f orchestrator.log | grep "Strategy"

# 3. 验证策略启用状态
psql -c "SELECT id, name, enabled FROM strategies WHERE id = 'strategy_pod_crashloop';"

# 4. 检查事件是否被 agent-manager 发布
# (在 agent-manager 日志中查看)
```

**场景 2: 步骤执行超时**

调试步骤:
```bash
# 1. 检查步骤执行日志
tail -f orchestrator.log | grep "step_id=collect_logs"

# 2. 检查超时设置
cat configs/config.yaml | grep timeout

# 3. 增加超时时间 (如果需要)
# 编辑工作流定义,修改步骤的 timeout 字段

# 4. 查看步骤执行历史
psql -c "
  SELECT step_id, status,
         EXTRACT(EPOCH FROM (completed_at - started_at)) as duration
  FROM step_executions
  WHERE step_id = 'collect_logs'
    AND started_at > NOW() - INTERVAL '1 day'
  ORDER BY started_at DESC;
"
```

**场景 3: AI 分析返回错误结果**

调试步骤:
```bash
# 1. 查看 AI 分析步骤的输入
psql -c "
  SELECT input, output, error_message
  FROM step_executions
  WHERE step_type = 'ai_analysis'
    AND execution_id = 'exec-20251001-001';
"

# 2. 手动调用 reasoning-service 验证
curl -X POST http://reasoning-service:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d @test-analysis-request.json

# 3. 查看 reasoning-service 日志
kubectl -n aetherius logs -l app=reasoning-service --tail=100

# 4. 检查 AI 服务配置
cat configs/config.yaml | grep -A 3 "^ai:"
```

### 8. 性能分析

**识别慢步骤**:
```sql
-- 查找平均执行时间最长的步骤
SELECT
  step_id,
  step_type,
  COUNT(*) as execution_count,
  AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_seconds,
  MAX(EXTRACT(EPOCH FROM (completed_at - started_at))) as max_seconds
FROM step_executions
WHERE completed_at > NOW() - INTERVAL '7 days'
  AND status = 'completed'
GROUP BY step_id, step_type
ORDER BY avg_seconds DESC
LIMIT 10;
```

**分析工作流并发度**:
```sql
-- 查看同时运行的工作流数量
SELECT
  DATE_TRUNC('hour', started_at) as hour,
  COUNT(*) as concurrent_executions
FROM workflow_executions
WHERE status = 'running'
  AND started_at > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour;
```

### 9. 调试工具清单

| 工具 | 用途 | 命令示例 |
|------|------|----------|
| **日志分析** | 实时查看执行日志 | `tail -f orchestrator.log \| grep "workflow"` |
| **psql** | 数据库查询调试 | `make psql` |
| **redis-cli** | 缓存状态检查 | `make redis-cli` |
| **curl** | API 端点测试 | `curl http://localhost:8081/health` |
| **nats CLI** | NATS 消息测试 | `nats sub internal.event.critical` |
| **jq** | JSON 格式化 | `cat execution.json \| jq .` |

### 10. 启用分布式追踪 (规划中)

未来版本将集成 OpenTelemetry 进行分布式追踪:

```yaml
# 规划中的配置
tracing:
  enabled: true
  exporter: "jaeger"
  endpoint: "http://jaeger:14268/api/traces"
  sample_rate: 0.1
```

---

## 故障排查

### 问题 1: 工作流未触发

**症状**: agent-manager 发布了事件,但 orchestrator 没有触发工作流

**检查清单**:
- [ ] agent-manager 是否发布了内部事件 (检查 agent-manager 日志)
- [ ] NATS 连接是否正常 (`grep "NATS" orchestrator.log`)
- [ ] 策略是否启用 (`SELECT * FROM strategies WHERE enabled = true`)
- [ ] 症状匹配是否正确 (检查 event_reason 是否匹配)
- [ ] 工作流状态是否为 active

**解决方法**:
```bash
# 1. 验证 NATS 连接
curl http://localhost:8081/health | jq .components.nats

# 2. 启用策略
psql -c "UPDATE strategies SET enabled = true WHERE id = 'strategy_pod_crashloop';"

# 3. 查看最近的事件订阅日志
tail -f orchestrator.log | grep "Event received"
```

### 问题 2: 步骤执行失败

**症状**: 工作流已触发,但某个步骤执行失败

**检查清单**:
- [ ] agent-manager API 是否可访问
- [ ] reasoning-service 是否运行
- [ ] 步骤配置是否正确 (tool, action, args)
- [ ] 超时设置是否合理
- [ ] 网络策略是否允许访问

**解决方法**:
```bash
# 1. 测试 agent-manager 连通性
curl http://agent-manager:8080/health

# 2. 测试 reasoning-service 连通性
curl http://reasoning-service:8082/health

# 3. 查看步骤失败详情
psql -c "
  SELECT step_id, error_message, output
  FROM step_executions
  WHERE status = 'failed'
  ORDER BY started_at DESC
  LIMIT 5;
"

# 4. 增加超时时间
# 编辑工作流定义,修改 timeout 字段
```

### 问题 3: AI 分析无响应

**症状**: AI 分析步骤超时或无结果

**检查清单**:
- [ ] reasoning-service URL 配置正确
- [ ] 网络连通性正常
- [ ] AI 服务有足够资源 (CPU/内存)
- [ ] 请求超时设置合理
- [ ] AI 服务依赖 (Neo4j) 正常

**解决方法**:
```bash
# 1. 检查 AI 服务配置
cat configs/config.yaml | grep -A 3 "^ai:"

# 2. 测试 AI 服务
curl -X POST http://reasoning-service:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "test",
    "context": {"event": {"reason": "CrashLoopBackOff"}}
  }'

# 3. 查看 AI 服务日志
kubectl -n aetherius logs -l app=reasoning-service --tail=50

# 4. 增加 AI 超时时间
# 编辑 configs/config.yaml:
# ai:
#   timeout: 60s  # 从 30s 增加到 60s
```

### 问题 4: 数据库连接错误

**症状**: 服务启动失败或频繁报数据库错误

**解决方法**:
```bash
# 1. 验证数据库连接信息
psql -h localhost -U aetherius -d aetherius_orchestrator -c "SELECT 1;"

# 2. 检查连接池配置
cat configs/config.yaml | grep -A 10 "^database:"

# 3. 查看数据库连接数
psql -c "SELECT count(*) FROM pg_stat_activity WHERE datname = 'aetherius_orchestrator';"

# 4. 调整连接池参数 (如果连接数过多)
# 编辑 configs/config.yaml:
# database:
#   max_open_conns: 50  # 减少最大连接数
```

### 问题 5: Redis 缓存失效

**症状**: 策略或工作流定义频繁从数据库加载

**解决方法**:
```bash
# 1. 检查 Redis 连接
redis-cli -h localhost -p 6379 PING

# 2. 查看缓存命中率
redis-cli INFO stats | grep keyspace

# 3. 手动清除缓存 (如果需要强制刷新)
redis-cli FLUSHDB

# 4. 检查 Redis 配置
cat configs/config.yaml | grep -A 8 "^redis:"
```

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