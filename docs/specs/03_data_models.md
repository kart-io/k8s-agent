# Aetherius AI Agent - 核心数据模型

> **文档版本**: v1.6
> **最后更新**: 2025年9月28日
> **读者对象**: 开发工程师、数据库管理员

> **⭐ 数据模型技术规范**: 本文档提供 Aetherius 项目**数据模型的详细技术实现**。Go结构体定义、数据库Schema、API接口定义以本文档为准。
>
> **文档关系**:
> - [ai_agent.md 第6章](../ai_agent.md#6-核心数据模型-core-data-models): 提供数据模型的概要说明和业务上下文
> - **本文档**: 提供完整的技术实现细节(Go结构体、数据库表、API接口)
> - [02_architecture.md](./02_architecture.md): 说明数据模型在系统架构中的使用方式

---

## 目录

- [1. 数据模型总览](#1-数据模型总览)
- [2. 任务相关模型](#2-任务相关模型)
- [3. 工具与执行模型](#3-工具与执行模型)
- [4. 知识库模型](#4-知识库模型)
- [5. 报告与反馈模型](#5-报告与反馈模型)
- [6. 资源与配置模型](#6-资源与配置模型)
- [7. API接口定义](#7-api接口定义)

---

## 1. 数据模型总览

### 1.1 数据模型关系图

> **阅读指南**: 下图展示核心数据模型之间的关系，使用标准ER图符号：
> - **1:1** = 一对一关系
> - **1:N** = 一对多关系
> - **N:N** = 多对多关系

```
┌─────────────────────────────────────────────────────────────────────┐
│                          核心业务模型                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  DiagnosticTask (诊断任务) ──────── 1:1 ──────── DiagnosticReport   │
│       │                                              (诊断报告)      │
│       ├─── 1:N ──── DiagnosticStep (诊断步骤)                       │
│       │                  │                                          │
│       │                  ├─── N:1 ──── Tool (工具)                  │
│       │                  └─── 1:1 ──── ExecutionResult (执行结果)    │
│       │                                                              │
│       └─── 1:N ──── UserFeedback (用户反馈)                         │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                          知识管理模型                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  KnowledgeBase (知识库)                                              │
│       ├─── 1:N ──── KnowledgeEntry (知识条目)                       │
│       │                  └─── 1:1 ──── Embedding (向量)             │
│       └─── 1:N ──── KnowledgeCategory (知识分类)                    │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│                          系统管理模型                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Tool (工具注册表)                                                   │
│       ├─── N:1 ──── ToolCategory (工具分类)                         │
│       └─── N:N ──── Cluster (集群) [通过ToolClusterMapping]          │
│                                                                     │
│  ResourceUsage (资源使用监控)                                        │
│       └─── N:1 ──── DiagnosticTask (关联诊断任务)                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 核心模型清单

> **文档层次说明**:
> - **完整业务定义**: [ai_agent.md#第6章](../ai_agent.md#6-核心数据模型-core-data-models) - 提供业务上下文和需求背景
> - **技术实现细节**: 本文档各章节 - 提供Go结构体、数据库Schema、API接口

#### 1.2.1 按业务领域分类

**核心业务模型** (诊断流程核心):

| 模型 | 业务职责 | ai_agent.md | 本文档 |
|------|----------|-------------|--------|
| DiagnosticTask | 诊断任务管理和状态跟踪 | [6.1节](../ai_agent.md#61-诊断任务模型-diagnostic-task-model) | [2.1节](#21-diagnostictask-诊断任务模型) |
| DiagnosticStep | 单步骤执行和结果记录 | [6.2节](../ai_agent.md#62-诊断步骤模型-diagnostic-step-model) | [2.2节](#22-diagnosticstep-诊断步骤模型) |
| DiagnosticReport | 诊断报告生成和分发 | [6.5节](../ai_agent.md#65-诊断报告模型-diagnostic-report-model) | [5.1节](#51-diagnosticreport-诊断报告模型) |

**支撑服务模型** (系统支撑功能):

| 模型 | 业务职责 | ai_agent.md | 本文档 |
|------|----------|-------------|--------|
| Tool | 工具注册和权限管理 | [6.3节](../ai_agent.md#63-工具注册表模型-tool-registry-model) | [3.1节](#31-tool-工具注册表模型) |
| KnowledgeBase | 知识存储和语义检索 | [6.4节](../ai_agent.md#64-知识库模型-knowledge-base-model) | [4.1节](#41-knowledgebase-知识库模型) |
| ResourceUsage | 成本监控和预算控制 | [6.7节](../ai_agent.md#67-资源使用模型-resource-usage-model) | [6.1节](#61-resourceusage-资源使用模型) |

**管理和审计模型** (运维和合规):

| 模型 | 业务职责 | ai_agent.md | 本文档 |
|------|----------|-------------|--------|
| UserFeedback | 反馈收集和知识更新 | [6.6节](../ai_agent.md#66-用户反馈模型-user-feedback-model) | [5.2节](#52-userfeedback-用户反馈模型) |
| SystemConfig | 配置管理和环境适配 | [6.8节](../ai_agent.md#68-配置管理模型-configuration-model) | [6.2节](#62-systemconfig-系统配置模型) |
| HistoryRecord | 审计追踪和合规记录 | [6.9节](../ai_agent.md#69-历史记录模型-history-record-model) | [6.3节](#63-historyrecord-历史记录模型) |

---

## 2. 任务相关模型

### 2.1 DiagnosticTask (诊断任务模型)

详见: [ai_agent.md#6.1](../ai_agent.md#61-诊断任务模型-diagnostic-task-model)

**Go结构定义**:

```go
type DiagnosticTask struct {
    ID          string                 `json:"id" db:"id"`
    AlertID     string                 `json:"alert_id" db:"alert_id"`
    ClusterID   string                 `json:"cluster_id" db:"cluster_id"`
    Namespace   string                 `json:"namespace" db:"namespace"`
    Resource    ResourceInfo           `json:"resource" db:"resource"`
    Priority    Priority               `json:"priority" db:"priority"`
    Status      TaskStatus             `json:"status" db:"status"`
    CreatedAt   time.Time              `json:"created_at" db:"created_at"`
    StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
    CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
    Context     TaskContext            `json:"context" db:"context"`
    Steps       []DiagnosticStep       `json:"steps" db:"steps"`
    Result      *DiagnosticResult      `json:"result,omitempty" db:"result"`
    Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

type ResourceInfo struct {
    Kind      string `json:"kind"`
    Name      string `json:"name"`
    Namespace string `json:"namespace,omitempty"`
}

type Priority int
const (
    PriorityLow Priority = iota       // P3
    PriorityMedium                    // P2
    PriorityHigh                      // P1
    PriorityEmergency                 // P0
)

type TaskStatus string
const (
    StatusPending   TaskStatus = "pending"
    StatusRunning   TaskStatus = "running"
    StatusCompleted TaskStatus = "completed"
    StatusFailed    TaskStatus = "failed"
    StatusCancelled TaskStatus = "cancelled"
    StatusTimeout   TaskStatus = "timeout"
)

// 状态转换规则
type StateTransition struct {
    From       TaskStatus
    To         TaskStatus
    Condition  string
}

// 合法的状态转换
var ValidTransitions = []StateTransition{
    // 从 Pending 状态
    {From: StatusPending, To: StatusRunning, Condition: "任务被调度器取出"},
    {From: StatusPending, To: StatusCancelled, Condition: "用户取消或系统关闭"},
    {From: StatusPending, To: StatusTimeout, Condition: "等待超时(>30分钟)"},

    // 从 Running 状态
    {From: StatusRunning, To: StatusCompleted, Condition: "诊断成功完成"},
    {From: StatusRunning, To: StatusFailed, Condition: "诊断失败或错误"},
    {From: StatusRunning, To: StatusCancelled, Condition: "用户手动终止"},
    {From: StatusRunning, To: StatusTimeout, Condition: "执行超时(>10分钟)"},

    // 终态不允许转换
    // StatusCompleted, StatusFailed, StatusCancelled, StatusTimeout 为终态
}

// 状态转换验证函数
func IsValidTransition(from, to TaskStatus) bool {
    for _, t := range ValidTransitions {
        if t.From == from && t.To == to {
            return true
        }
    }
    return false
}

type TaskContext struct {
    AlertInfo    AlertInfo              `json:"alert_info"`
    ClusterInfo  ClusterInfo            `json:"cluster_info"`
    Environment  string                 `json:"environment"`
    Labels       map[string]string      `json:"labels"`
    Annotations  map[string]string      `json:"annotations"`
}
```

**数据库表结构**:

```sql
CREATE TABLE diagnostic_tasks (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id         VARCHAR(255) NOT NULL,
    cluster_id       VARCHAR(255) NOT NULL,
    namespace        VARCHAR(255),
    resource_kind    VARCHAR(100),
    resource_name    VARCHAR(255),
    priority         INTEGER NOT NULL,
    status           VARCHAR(50) NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at       TIMESTAMP,
    completed_at     TIMESTAMP,
    context          JSONB,
    metadata         JSONB,
    INDEX idx_cluster_id (cluster_id),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_created_at (created_at)
);
```

**状态转换规则**:

```
pending → running → completed
                 → failed
                 → timeout
                 → cancelled
```

### 2.2 DiagnosticStep (诊断步骤模型)

详见: [ai_agent.md#6.2](../ai_agent.md#62-诊断步骤模型-diagnostic-step-model)

**Go结构定义**:

```go
type DiagnosticStep struct {
    ID          string        `json:"id" db:"id"`
    TaskID      string        `json:"task_id" db:"task_id"`
    StepNumber  int           `json:"step_number" db:"step_number"`
    Type        StepType      `json:"type" db:"type"`
    ToolID      string        `json:"tool_id" db:"tool_id"`
    Input       StepInput     `json:"input" db:"input"`
    Output      *StepOutput   `json:"output,omitempty" db:"output"`
    Status      StepStatus    `json:"status" db:"status"`
    StartedAt   time.Time     `json:"started_at" db:"started_at"`
    CompletedAt *time.Time    `json:"completed_at,omitempty" db:"completed_at"`
    Duration    time.Duration `json:"duration" db:"duration"`
    Reasoning   string        `json:"reasoning" db:"reasoning"`
    TokenUsage  TokenUsage    `json:"token_usage" db:"token_usage"`
}

type StepType string
const (
    StepTypeAnalysis   StepType = "analysis"
    StepTypeExecution  StepType = "execution"
    StepTypeValidation StepType = "validation"
)

type StepInput struct {
    Command    string            `json:"command"`
    Parameters map[string]string `json:"parameters"`
    Context    string            `json:"context"`
}

type StepOutput struct {
    Stdout   string `json:"stdout"`
    Stderr   string `json:"stderr"`
    ExitCode int    `json:"exit_code"`
    Duration int64  `json:"duration_ms"`
}

type TokenUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

---

## 3. 工具与执行模型

### 3.1 Tool (工具注册表模型)

详见: [ai_agent.md#5.4.1](../ai_agent.md#541-诊断工具注册表)

**Go结构定义**:

```go
type Tool struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Category    ToolCategory      `json:"category"`
    Command     string            `json:"command"`
    Parameters  []ToolParameter   `json:"parameters"`
    Timeout     time.Duration     `json:"timeout"`
    ReadOnly    bool              `json:"readonly"`
    Clusters    []string          `json:"clusters,omitempty"`
    Enabled     bool              `json:"enabled"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

type ToolCategory string
const (
    CategoryKubectl   ToolCategory = "kubectl"
    CategorySystem    ToolCategory = "system"
    CategoryNetwork   ToolCategory = "network"
    CategoryStorage   ToolCategory = "storage"
    CategoryCustom    ToolCategory = "custom"
)

type ToolParameter struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Description string   `json:"description"`
    Required    bool     `json:"required"`
    Default     string   `json:"default,omitempty"`
    Enum        []string `json:"enum,omitempty"`
    Pattern     string   `json:"pattern,omitempty"`
}
```

**预定义工具示例**:

```json
{
  "id": "kubectl-get-pods",
  "name": "Get Pods",
  "description": "List all pods in a namespace",
  "category": "kubectl",
  "command": "kubectl get pods -n {namespace} -o json",
  "parameters": [
    {
      "name": "namespace",
      "type": "string",
      "description": "Kubernetes namespace",
      "required": true
    }
  ],
  "timeout": "30s",
  "readonly": true,
  "enabled": true
}
```

### 3.2 CommandExecution (命令执行记录)

```go
type CommandExecution struct {
    ID          string                 `json:"id"`
    StepID      string                 `json:"step_id"`
    ToolID      string                 `json:"tool_id"`
    ClusterID   string                 `json:"cluster_id"`
    Command     string                 `json:"command"`
    Parameters  map[string]interface{} `json:"parameters"`
    StartedAt   time.Time              `json:"started_at"`
    CompletedAt time.Time              `json:"completed_at"`
    Duration    time.Duration          `json:"duration"`
    ExitCode    int                    `json:"exit_code"`
    Stdout      string                 `json:"stdout"`
    Stderr      string                 `json:"stderr"`
    Success     bool                   `json:"success"`
}
```

---

## 4. 知识库模型

### 4.1 KnowledgeEntry (知识条目模型)

详见: [ai_agent.md#6.4](../ai_agent.md#64-知识库模型-knowledge-base-model)

**Go结构定义**:

```go
type KnowledgeEntry struct {
    ID          string            `json:"id"`
    Title       string            `json:"title"`
    Content     string            `json:"content"`
    Category    string            `json:"category"`
    Tags        []string          `json:"tags"`
    Source      string            `json:"source"`
    Metadata    KnowledgeMetadata `json:"metadata"`
    Embedding   []float32         `json:"embedding,omitempty"`
    Confidence  float64           `json:"confidence"`
    UsageCount  int               `json:"usage_count"`
    SuccessRate float64           `json:"success_rate"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    Version     int               `json:"version"`
}

type KnowledgeMetadata struct {
    ClusterType      string   `json:"cluster_type,omitempty"`
    K8sVersion       []string `json:"k8s_version,omitempty"`
    Environment      string   `json:"environment,omitempty"`
    ProblemType      string   `json:"problem_type"`
    Severity         string   `json:"severity"`
    ApplicableScopes []string `json:"applicable_scopes"`
}
```

**向量索引结构** (Vector DB):

```
Collection: knowledge_base
  ├─ Vectors: 768维 (OpenAI ada-002)
  ├─ Index: HNSW (M=16, efConstruction=200)
  └─ Metadata:
      ├─ entry_id
      ├─ category
      ├─ tags
      └─ metadata (JSON)
```

### 4.2 KnowledgeQuery (知识查询)

```go
type KnowledgeQuery struct {
    Query          string            `json:"query"`
    TopK           int               `json:"top_k"`
    MinScore       float64           `json:"min_score"`
    Filters        map[string]string `json:"filters,omitempty"`
    IncludeContent bool              `json:"include_content"`
}

type KnowledgeSearchResult struct {
    Entry      KnowledgeEntry `json:"entry"`
    Score      float64        `json:"score"`
    Highlights []string       `json:"highlights,omitempty"`
}
```

---

## 5. 报告与反馈模型

### 5.1 DiagnosticReport (诊断报告模型)

详见: [ai_agent.md#6.5](../ai_agent.md#65-诊断报告模型-diagnostic-report-model)

**Go结构定义**:

```go
type DiagnosticReport struct {
    ID               string          `json:"id"`
    TaskID           string          `json:"task_id"`
    Summary          string          `json:"summary"`
    RootCause        RootCauseInfo   `json:"root_cause"`
    Recommendations  []Recommendation `json:"recommendations"`
    Timeline         []TimelineEvent `json:"timeline"`
    ExecutedSteps    []StepSummary   `json:"executed_steps"`
    Confidence       float64         `json:"confidence"`
    ImpactAssessment ImpactInfo      `json:"impact_assessment"`
    CreatedAt        time.Time       `json:"created_at"`
    Format           ReportFormat    `json:"format"`
}

type RootCauseInfo struct {
    Description string            `json:"description"`
    Category    string            `json:"category"`
    Evidence    []string          `json:"evidence"`
    Confidence  float64           `json:"confidence"`
    RelatedDocs []string          `json:"related_docs,omitempty"`
}

type Recommendation struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Priority    string   `json:"priority"`
    Actions     []Action `json:"actions"`
    Impact      string   `json:"impact"`
}

type Action struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Command     string `json:"command,omitempty"`
    Warning     string `json:"warning,omitempty"`
}

type ReportFormat string
const (
    FormatJSON     ReportFormat = "json"
    FormatMarkdown ReportFormat = "markdown"
    FormatHTML     ReportFormat = "html"
    FormatSlack    ReportFormat = "slack"
)
```

### 5.2 UserFeedback (用户反馈模型)

详见: [ai_agent.md#6.6](../ai_agent.md#66-用户反馈模型-user-feedback-model)

**Go结构定义**:

```go
type UserFeedback struct {
    ID         string         `json:"id"`
    TaskID     string         `json:"task_id"`
    ReportID   string         `json:"report_id"`
    UserID     string         `json:"user_id"`
    Rating     FeedbackRating `json:"rating"`
    Comment    string         `json:"comment,omitempty"`
    Categories []string       `json:"categories,omitempty"`
    CreatedAt  time.Time      `json:"created_at"`
    Processed  bool           `json:"processed"`
    ProcessedAt *time.Time    `json:"processed_at,omitempty"`
}

type FeedbackRating string
const (
    RatingPositive FeedbackRating = "positive"  // 👍
    RatingNegative FeedbackRating = "negative"  // 👎
    RatingNeutral  FeedbackRating = "neutral"   // 😐
)
```

---

## 6. 资源与配置模型

### 6.1 ResourceUsage (资源使用模型)

详见: [ai_agent.md#6.7](../ai_agent.md#67-资源使用模型-resource-usage-model)

**Go结构定义**:

```go
type ResourceUsage struct {
    ID            string        `json:"id"`
    TaskID        string        `json:"task_id"`
    TokenUsage    TokenCost     `json:"token_usage"`
    ExecutionTime time.Duration `json:"execution_time"`
    APICallCount  int           `json:"api_call_count"`
    CacheHitRate  float64       `json:"cache_hit_rate"`
    EstimatedCost float64       `json:"estimated_cost"`
    Timestamp     time.Time     `json:"timestamp"`
}

type TokenCost struct {
    Model            string  `json:"model"`
    PromptTokens     int     `json:"prompt_tokens"`
    CompletionTokens int     `json:"completion_tokens"`
    TotalTokens      int     `json:"total_tokens"`
    PromptCost       float64 `json:"prompt_cost"`
    CompletionCost   float64 `json:"completion_cost"`
    TotalCost        float64 `json:"total_cost"`
}
```

### 6.2 SystemConfig (系统配置模型)

详见: [ai_agent.md#6.8](../ai_agent.md#68-配置管理模型-configuration-model)

**Go结构定义**:

```go
type SystemConfig struct {
    ID          string                 `json:"id"`
    Key         string                 `json:"key"`
    Value       interface{}            `json:"value"`
    Type        ConfigType             `json:"type"`
    Description string                 `json:"description"`
    Category    string                 `json:"category"`
    Validation  *ConfigValidation      `json:"validation,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Version     int                    `json:"version"`
}

type ConfigType string
const (
    ConfigTypeString  ConfigType = "string"
    ConfigTypeNumber  ConfigType = "number"
    ConfigTypeBoolean ConfigType = "boolean"
    ConfigTypeJSON    ConfigType = "json"
)

type ConfigValidation struct {
    Min     *float64 `json:"min,omitempty"`
    Max     *float64 `json:"max,omitempty"`
    Pattern string   `json:"pattern,omitempty"`
    Enum    []string `json:"enum,omitempty"`
}
```

**配置示例**:

```json
{
  "key": "diagnostic.max_concurrent_tasks",
  "value": 50,
  "type": "number",
  "description": "Maximum concurrent diagnostic tasks",
  "category": "performance",
  "validation": {
    "min": 1,
    "max": 100
  }
}
```

---

## 7. API接口定义

### 7.1 REST API 端点

**任务管理API**:

```
POST   /api/v1/tasks                  # 创建诊断任务
GET    /api/v1/tasks                  # 列出任务
GET    /api/v1/tasks/:id              # 获取任务详情
PUT    /api/v1/tasks/:id/cancel       # 取消任务
DELETE /api/v1/tasks/:id              # 删除任务
GET    /api/v1/tasks/:id/steps        # 获取任务步骤
GET    /api/v1/tasks/:id/report       # 获取诊断报告
```

**工具管理API**:

```
GET    /api/v1/tools                  # 列出工具
GET    /api/v1/tools/:id              # 获取工具详情
POST   /api/v1/tools                  # 注册新工具
PUT    /api/v1/tools/:id              # 更新工具
DELETE /api/v1/tools/:id              # 删除工具
```

**知识库API**:

```
GET    /api/v1/knowledge              # 搜索知识
GET    /api/v1/knowledge/:id          # 获取知识详情
POST   /api/v1/knowledge              # 添加知识
PUT    /api/v1/knowledge/:id          # 更新知识
DELETE /api/v1/knowledge/:id          # 删除知识
POST   /api/v1/knowledge/search       # 语义搜索
```

**反馈API**:

```
POST   /api/v1/feedback               # 提交反馈
GET    /api/v1/feedback/:id           # 获取反馈
GET    /api/v1/tasks/:id/feedback     # 获取任务反馈
```

### 7.2 API请求示例

**创建诊断任务**:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "alert_id": "alert-12345",
    "cluster_id": "prod-us-west",
    "namespace": "ecommerce",
    "resource": {
      "kind": "Pod",
      "name": "payment-service-7d8f9b456-xyz12"
    },
    "priority": 1
  }'
```

**查询任务状态**:

```bash
curl -X GET http://localhost:8080/api/v1/tasks/task-uuid \
  -H "Authorization: Bearer $TOKEN"
```

**提交反馈**:

```bash
curl -X POST http://localhost:8080/api/v1/feedback \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "task_id": "task-uuid",
    "rating": "positive",
    "comment": "Accurately identified the root cause"
  }'
```

### 7.3 GraphQL Schema (可选)

```graphql
type Query {
  task(id: ID!): DiagnosticTask
  tasks(
    filter: TaskFilter
    limit: Int
    offset: Int
  ): TaskConnection

  tool(id: ID!): Tool
  tools(category: ToolCategory): [Tool!]!

  knowledge(id: ID!): KnowledgeEntry
  searchKnowledge(
    query: String!
    topK: Int
    filters: KnowledgeFilter
  ): [KnowledgeSearchResult!]!
}

type Mutation {
  createTask(input: CreateTaskInput!): DiagnosticTask!
  cancelTask(id: ID!): DiagnosticTask!

  submitFeedback(input: FeedbackInput!): UserFeedback!

  registerTool(input: ToolInput!): Tool!
  updateTool(id: ID!, input: ToolInput!): Tool!
}

type Subscription {
  taskUpdated(id: ID!): DiagnosticTask!
  taskCompleted(clusterID: String): DiagnosticTask!
}
```

---

## 相关文档

- [系统架构](./02_architecture.md) - 架构设计
- [需求规格](../REQUIREMENTS.md) - 功能需求
- [API文档](../ai_agent.md#6章) - 完整数据模型定义

---

**文档维护**: Aetherius开发团队
**最后更新**: 2025年9月28日