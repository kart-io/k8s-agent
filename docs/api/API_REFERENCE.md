# Aetherius API 参考文档

完整的 REST API 接口文档。

---

## 目录

- [Agent Manager API](#agent-manager-api)
- [Orchestrator Service API](#orchestrator-service-api)
- [Reasoning Service API](#reasoning-service-api)
- [认证](#认证)
- [错误处理](#错误处理)

---

## Agent Manager API

**Base URL**: `http://agent-manager:8080`

### 健康检查

#### GET /health

检查服务健康状态。

**响应**:

```json
{
  "status": "healthy",
  "service": "agent-manager",
  "timestamp": "2025-09-30T10:00:00Z"
}
```

---

### Agent 管理

#### GET /api/v1/agents

列出所有 Agent。

**查询参数**:

- `cluster_id` (optional): 按集群 ID 过滤
- `status` (optional): 按状态过滤 (online/offline/error)
- `page` (optional): 页码，默认 1
- `page_size` (optional): 每页数量，默认 20

**响应**:

```json
{
  "agents": [
    {
      "id": "agent-123",
      "cluster_id": "prod-cluster",
      "status": "online",
      "version": "v1.0.0",
      "last_heartbeat": "2025-09-30T10:00:00Z",
      "connection_info": {
        "ip": "192.168.1.100",
        "hostname": "node-1"
      },
      "metadata": {
        "node_count": 10,
        "pod_count": 150
      }
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 20
}
```

#### GET /api/v1/agents/:id

获取特定 Agent 详情。

**路径参数**:

- `id`: Agent ID

**响应**:

```json
{
  "id": "agent-123",
  "cluster_id": "prod-cluster",
  "status": "online",
  "version": "v1.0.0",
  "last_heartbeat": "2025-09-30T10:00:00Z",
  "connection_info": {
    "ip": "192.168.1.100",
    "hostname": "node-1"
  },
  "metadata": {
    "node_count": 10,
    "pod_count": 150,
    "k8s_version": "v1.28.0"
  },
  "created_at": "2025-09-01T00:00:00Z",
  "updated_at": "2025-09-30T10:00:00Z"
}
```

---

### 集群管理

#### GET /api/v1/clusters

列出所有集群。

**响应**:

```json
{
  "clusters": [
    {
      "id": "prod-cluster",
      "name": "Production Cluster",
      "region": "us-west-2",
      "status": "healthy",
      "agent_count": 1,
      "node_count": 10,
      "pod_count": 150,
      "metadata": {
        "k8s_version": "v1.28.0",
        "provider": "aws"
      }
    }
  ],
  "total": 5
}
```

#### GET /api/v1/clusters/:id

获取集群详情。

**路径参数**:

- `id`: 集群 ID

**响应**:

```json
{
  "id": "prod-cluster",
  "name": "Production Cluster",
  "region": "us-west-2",
  "status": "healthy",
  "agent_count": 1,
  "node_count": 10,
  "pod_count": 150,
  "recent_events": [
    {
      "id": "evt-123",
      "type": "Warning",
      "reason": "OOMKilled",
      "message": "Container was OOM killed",
      "timestamp": "2025-09-30T09:50:00Z"
    }
  ],
  "metadata": {
    "k8s_version": "v1.28.0",
    "provider": "aws",
    "total_cpu": "40 cores",
    "total_memory": "160Gi"
  }
}
```

---

### 事件查询

#### GET /api/v1/events

查询事件。

**查询参数**:

- `cluster_id` (optional): 集群 ID
- `severity` (optional): 严重级别 (critical/high/medium/low)
- `reason` (optional): 事件原因
- `namespace` (optional): 命名空间
- `start_time` (optional): 开始时间 (RFC3339)
- `end_time` (optional): 结束时间 (RFC3339)
- `page` (optional): 页码
- `page_size` (optional): 每页数量

**响应**:

```json
{
  "events": [
    {
      "id": "evt-123",
      "cluster_id": "prod-cluster",
      "timestamp": "2025-09-30T09:50:00Z",
      "type": "Warning",
      "severity": "high",
      "reason": "OOMKilled",
      "message": "Container api-server in pod api-server-7f9b8c-xyz was OOMKilled",
      "namespace": "production",
      "labels": {
        "kind": "Pod",
        "name": "api-server-7f9b8c-xyz",
        "node": "node-1"
      }
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

#### POST /api/v1/events/query

高级事件查询。

**请求体**:

```json
{
  "filters": {
    "cluster_ids": ["prod-cluster", "staging-cluster"],
    "severities": ["critical", "high"],
    "reasons": ["OOMKilled", "CrashLoopBackOff"],
    "time_range": {
      "start": "2025-09-30T00:00:00Z",
      "end": "2025-09-30T23:59:59Z"
    }
  },
  "aggregation": {
    "group_by": ["reason", "namespace"],
    "metrics": ["count", "avg_duration"]
  },
  "sort": {
    "field": "timestamp",
    "order": "desc"
  },
  "pagination": {
    "page": 1,
    "page_size": 50
  }
}
```

**响应**:

```json
{
  "events": [...],
  "aggregations": {
    "OOMKilled": {
      "count": 45,
      "namespaces": ["production", "staging"]
    }
  },
  "total": 100
}
```

---

### 命令管理

#### POST /api/v1/commands

发送命令到 Agent。

**请求体**:

```json
{
  "cluster_id": "prod-cluster",
  "type": "diagnostic",
  "tool": "kubectl",
  "action": "logs",
  "args": ["--tail=100", "--previous", "api-server-7f9b8c-xyz"],
  "namespace": "production",
  "timeout": "30s"
}
```

**响应**:

```json
{
  "command_id": "cmd-123",
  "cluster_id": "prod-cluster",
  "status": "pending",
  "issued_at": "2025-09-30T10:00:00Z",
  "timeout": "30s"
}
```

#### GET /api/v1/commands/:id

查询命令状态。

**路径参数**:

- `id`: 命令 ID

**响应**:

```json
{
  "command_id": "cmd-123",
  "cluster_id": "prod-cluster",
  "status": "completed",
  "issued_at": "2025-09-30T10:00:00Z",
  "completed_at": "2025-09-30T10:00:05Z",
  "execution_time": "5s",
  "result": {
    "exit_code": 0,
    "output": "Log output here...",
    "error": ""
  }
}
```

---

## Orchestrator Service API

**Base URL**: `http://orchestrator-service:8081`

### 工作流管理

#### GET /api/v1/workflows

列出所有工作流。

**查询参数**:

- `status` (optional): 状态过滤 (active/inactive)
- `category` (optional): 分类过滤

**响应**:

```json
{
  "workflows": [
    {
      "id": "wf-diagnose-oom",
      "name": "Diagnose OOM Killed",
      "description": "Automatically diagnose and fix OOM killed pods",
      "trigger_type": "event",
      "status": "active",
      "priority": 10,
      "steps_count": 6,
      "created_at": "2025-09-01T00:00:00Z"
    }
  ],
  "total": 25
}
```

#### GET /api/v1/workflows/:id

获取工作流详情。

**响应**:

```json
{
  "id": "wf-diagnose-oom",
  "name": "Diagnose OOM Killed",
  "description": "Automatically diagnose and fix OOM killed pods",
  "trigger_type": "event",
  "trigger_config": {
    "event_reason": "OOMKilled",
    "severity": "high"
  },
  "status": "active",
  "priority": 10,
  "timeout": "5m",
  "steps": [
    {
      "id": "collect_logs",
      "type": "command",
      "name": "Collect container logs",
      "config": {
        "tool": "kubectl",
        "action": "logs"
      },
      "on_success": ["ai_analysis"],
      "on_failure": ["notify_failure"]
    }
  ]
}
```

#### POST /api/v1/workflows/:id/execute

手动触发工作流执行。

**请求体**:

```json
{
  "trigger_event": {
    "cluster_id": "prod-cluster",
    "namespace": "production",
    "pod_name": "api-server-xyz",
    "reason": "OOMKilled"
  }
}
```

**响应**:

```json
{
  "execution_id": "exec-123",
  "workflow_id": "wf-diagnose-oom",
  "status": "running",
  "started_at": "2025-09-30T10:00:00Z"
}
```

#### GET /api/v1/workflows/executions/:id

查询工作流执行状态。

**响应**:

```json
{
  "execution_id": "exec-123",
  "workflow_id": "wf-diagnose-oom",
  "status": "completed",
  "started_at": "2025-09-30T10:00:00Z",
  "completed_at": "2025-09-30T10:02:30Z",
  "duration": "150s",
  "steps_completed": 6,
  "steps_total": 6,
  "result": {
    "root_cause": "OOMKiller",
    "recommendations": [
      {
        "action": "increase_memory_limit",
        "confidence": 0.92
      }
    ],
    "actions_taken": [
      "Updated memory limit from 512Mi to 1Gi"
    ]
  }
}
```

---

### 策略管理

#### GET /api/v1/strategies

列出所有诊断策略。

**响应**:

```json
{
  "strategies": [
    {
      "id": "strategy-oom",
      "name": "OOM Killer Strategy",
      "category": "pod_failure",
      "workflow_id": "wf-diagnose-oom",
      "priority": 10,
      "enabled": true,
      "match_count": 145
    }
  ],
  "total": 50
}
```

---

## Reasoning Service API

**Base URL**: `http://reasoning-service:8082`

### 根因分析

#### POST /api/v1/analyze/root-cause

执行根因分析。

**请求体**:

```json
{
  "request_id": "req-123",
  "analysis_type": "root_cause",
  "context": {
    "event": {
      "reason": "OOMKilled",
      "message": "Container was OOM killed"
    },
    "logs": "fatal error: runtime: out of memory...",
    "metrics": {
      "memory": {
        "usage_percent": 98
      }
    }
  },
  "options": {
    "min_confidence": 0.7,
    "include_similar_cases": true,
    "max_recommendations": 5
  }
}
```

**响应**:

```json
{
  "request_id": "req-123",
  "status": "completed",
  "result": {
    "root_cause": {
      "type": "OOMKiller",
      "description": "Container was killed due to out of memory (OOM)",
      "confidence": 0.95,
      "evidence": [
        "Event reason: OOMKilled",
        "Found 'Out of memory' in logs",
        "Memory usage at 98%"
      ]
    },
    "recommendations": [
      {
        "action": "increase_memory_limit",
        "description": "Increase container memory limits to prevent OOM kills",
        "confidence": 0.90,
        "risk": "low",
        "impact": "Prevents future OOM kills, may increase cluster resource usage",
        "steps": [
          "Analyze current memory usage patterns",
          "Calculate recommended memory limit (current + 50%)",
          "Update Deployment/StatefulSet memory limits",
          "kubectl apply -f updated-manifest.yaml",
          "Monitor for OOM recurrence"
        ],
        "rollback_steps": [
          "Revert to previous memory limits",
          "kubectl rollout undo deployment/<name>"
        ],
        "estimated_duration": "5 minutes"
      }
    ],
    "confidence": 0.95,
    "similar_cases": [
      {
        "case_id": "case-456",
        "description": "Similar OOM issue in production",
        "similarity_score": 0.88,
        "root_cause": "OOMKiller",
        "solution": "Increased memory limit to 1Gi",
        "outcome": "No more OOM kills"
      }
    ]
  },
  "processing_time": 0.523
}
```

---

### 故障预测

#### POST /api/v1/analyze/predict

预测潜在故障。

**请求体**:

```json
{
  "cluster_id": "prod-cluster",
  "resource_type": "pod",
  "resource_name": "api-server-xyz",
  "metrics": {
    "memory": {
      "usage_percent": 85
    },
    "cpu": {
      "usage_percent": 75,
      "throttling_percent": 60
    },
    "restart_count": 3,
    "history": [
      {
        "timestamp": "2025-09-30T09:00:00Z",
        "memory": {"usage_percent": 70},
        "cpu": {"usage_percent": 65}
      }
    ]
  },
  "time_window": "24h"
}
```

**响应**:

```json
{
  "failure_probability": 0.75,
  "predicted_failure_time": "2025-09-30T16:00:00Z",
  "failure_types": ["OOMKiller", "CPUThrottling"],
  "confidence": 0.82,
  "contributing_factors": [
    "Memory usage approaching limit (85%)",
    "CPU throttling at 60%",
    "Pod restarted 3 times",
    "Memory increasing at 2.5% per hour"
  ]
}
```

---

### 反馈提交

#### POST /api/v1/feedback

提交反馈用于学习。

**请求体**:

```json
{
  "feedback_id": "fb-123",
  "request_id": "req-123",
  "feedback_type": "diagnosis_accuracy",
  "rating": 5,
  "was_helpful": true,
  "actual_root_cause": "OOMKiller",
  "actual_solution": "Increased memory limit",
  "comments": "Diagnosis was accurate and recommendations worked perfectly",
  "submitted_by": "admin@example.com"
}
```

**响应**:

```json
{
  "status": "accepted",
  "message": "Feedback processed successfully"
}
```

---

### 知识图谱

#### GET /api/v1/cases/similar

查找相似案例。

**查询参数**:

- `event_reason` (optional): 事件原因
- `root_cause` (optional): 根因类型
- `limit` (optional): 返回数量，默认 5

**响应**:

```json
{
  "cases": [
    {
      "case_id": "case-456",
      "description": "API server OOM in production",
      "similarity_score": 0.88,
      "root_cause": "OOMKiller",
      "solution": "Increased memory limit from 512Mi to 1Gi",
      "outcome": "No more OOM kills after change",
      "timestamp": "2025-09-15T10:00:00Z"
    }
  ]
}
```

---

### 准确率指标

#### GET /api/v1/metrics/accuracy

获取诊断准确率指标。

**查询参数**:

- `root_cause_type` (optional): 根因类型过滤

**响应**:

```json
{
  "overall": 0.87,
  "by_root_cause": {
    "OOMKiller": {
      "total_diagnoses": 50,
      "correct_diagnoses": 47,
      "accuracy": 0.94,
      "last_updated": "2025-09-30T10:00:00Z"
    },
    "CPUThrottling": {
      "total_diagnoses": 30,
      "correct_diagnoses": 25,
      "accuracy": 0.83,
      "last_updated": "2025-09-30T10:00:00Z"
    }
  }
}
```

---

## 认证

### API Key 认证

所有 API 请求需要包含 API Key:

```bash
curl -H "X-API-Key: your-api-key" \
  http://agent-manager:8080/api/v1/agents
```

### JWT Token 认证

```bash
# 获取 Token
curl -X POST http://agent-manager:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'

# 使用 Token
curl -H "Authorization: Bearer <token>" \
  http://agent-manager:8080/api/v1/agents
```

---

## 错误处理

### 错误响应格式

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Agent not found",
    "details": {
      "agent_id": "agent-123"
    },
    "timestamp": "2025-09-30T10:00:00Z"
  }
}
```

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

### 错误代码

| 代码 | 说明 |
|------|------|
| INVALID_REQUEST | 请求参数无效 |
| RESOURCE_NOT_FOUND | 资源不存在 |
| UNAUTHORIZED | 未授权 |
| FORBIDDEN | 禁止访问 |
| RATE_LIMIT_EXCEEDED | 超出速率限制 |
| INTERNAL_ERROR | 内部错误 |
| SERVICE_UNAVAILABLE | 服务不可用 |

---

## 速率限制

默认速率限制:

- **未认证**: 10 requests/minute
- **已认证**: 100 requests/minute
- **Premium**: 1000 requests/minute

响应头:

```text
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1696070400
```

---

## 分页

支持分页的 API 使用以下参数:

- `page`: 页码 (从 1 开始)
- `page_size`: 每页数量 (最大 100)

响应包含分页信息:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

---

## Webhook

订阅事件通知:

```bash
POST /api/v1/webhooks
{
  "url": "https://example.com/webhook",
  "events": ["event.critical", "workflow.completed"],
  "secret": "webhook-secret"
}
```

Webhook 负载:

```json
{
  "event": "event.critical",
  "timestamp": "2025-09-30T10:00:00Z",
  "data": {
    "event_id": "evt-123",
    "cluster_id": "prod-cluster",
    "reason": "OOMKilled"
  },
  "signature": "sha256=..."
}
```