# Aetherius AI Agent - 微服务架构设计

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025年9月28日
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: 微服务架构设计

## 目录

- [1. 微服务架构概览](#1-微服务架构概览)
- [2. 核心微服务设计](#2-核心微服务设计)
- [3. 服务间通信](#3-服务间通信)
- [4. 数据管理策略](#4-数据管理策略)
- [5. 服务治理](#5-服务治理)
- [6. 部署架构](#6-部署架构)

## 1. 微服务架构概览

### 1.1 架构原则

基于领域驱动设计(DDD)和单一职责原则,将 Aetherius 系统拆分为以下微服务:

**设计原则**:

- ✅ **单一职责**: 每个服务专注于一个业务领域
- ✅ **松耦合**: 服务间通过定义良好的API通信
- ✅ **高内聚**: 相关功能聚合在同一服务内
- ✅ **独立部署**: 每个服务可独立发布和扩展
- ✅ **故障隔离**: 单个服务故障不影响整体系统

### 1.2 微服务全景图

> **架构说明**: 下图展示完整的微服务组件关系和数据流向
>
> **图例**:
> - **实线箭头**: 同步调用
> - **虚线箭头**: 异步消息
> - **双向箭头**: 双向通信

```
                    ┌───────────────────────────────────────┐
                    │       Ingress Layer                   │
                    │   Kong/Nginx/Traefik + Load Balancer │
                    └──────────────┬────────────────────────┘
                                   │ HTTP/HTTPS Traffic
                    ┌──────────────┼────────────────────────┐
                    │              │                        │
                    ▼              ▼                        ▼
        ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
        │  Event Gateway  │ │   API Gateway   │ │ Dashboard Web   │
        │  (事件入口)      │ │  (API聚合)      │ │ (前端界面)      │
        │  - Webhook接收  │ │  - REST API     │ │  - React SPA    │
        │  - 事件过滤     │ │  - GraphQL      │ │  - WebSocket    │
        │  - 格式标准化   │ │  - 身份认证     │ │  - 实时推送     │
        └────────┬────────┘ └─────────────────┘ └─────────────────┘
                 │                    ▲                    ▲
                 │ 发布事件           │ API调用            │ UI更新
                 ▼                    │                    │
    ┌────────────────────────────────────────────────────────────────┐
    │              Message Bus (NATS Streaming)                      │
    │        事件总线 • 发布/订阅 • 流处理 • 消息持久化                │
    └──────┬─────────────────────────┬─────────────────────────┬──────┘
           │                         │                         │
           │ 订阅事件                │ 订阅事件                │ 订阅事件
           ▼                         ▼                         ▼
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  Orchestrator   │      │ Reasoning       │      │ Execution       │
│  (任务编排)      │      │ Service         │      │ Gateway         │
│  - 任务调度     │◄────►│ (AI推理)        │◄────►│ (安全执行)      │
│  - 优先级队列   │      │  - LLM调用      │      │  - MCP协议      │
│  - 工作流引擎   │      │  - 策略生成     │      │  - 命令验证     │
│  - 状态管理     │      │  - 置信度评估   │      │  - 结果收集     │
└─────────┬───────┘      └─────────┬───────┘      └─────────┬───────┘
          │                        │                        │
          │                        │ 知识检索               │
          │                        ▼                        │
          │              ┌─────────────────┐                │
          │              │ Knowledge       │                │
          │              │ Service         │                │
          │              │ (知识管理)       │                │
          │              │  - 向量存储     │                │
          │              │  - 语义检索     │                │
          │              │  - 知识更新     │                │
          │              │  - RAG引擎      │                │
          │              └─────────────────┘                │
          │                                                 │
          │ 报告请求                                        │ 执行结果
          ▼                                                 ▼
┌─────────────────┐                              ┌─────────────────┐
│  Report Service │                              │ Notification    │
│  (报告生成)      │◄────────────────────────────►│ Service         │
│  - 报告模板     │            通知分发          │ (多渠道通知)     │
│  - 多格式输出   │                              │  - Slack        │
│  - 历史记录     │                              │  - Email        │
│  - 文件生成     │                              │  - Webhook      │
└─────────────────┘                              └─────────────────┘
```
┌───────────────────┐                                   ┌────────────────────┐
│  Report Service   │                                   │  Credential Service│
│  (报告服务)        │                                   │  (凭证服务)         │
│  - 报告生成       │                                   │  - Vault集成       │
│  - 通知发送       │                                   │  - 动态凭证        │
│  - 历史管理       │                                   │  - 凭证轮转        │
└─────────┬─────────┘                                   └────────────────────┘
          │
          ▼
┌───────────────────┐       ┌────────────────────┐      ┌────────────────────┐
│  Notification     │       │  Audit Service     │      │  Monitoring Service│
│  Service          │       │  (审计服务)         │      │  (监控服务)         │
│  (通知服务)        │       │  - 审计日志        │      │  - 指标收集        │
│  - 多渠道通知     │       │  - 合规检查        │      │  - 健康检查        │
│  - 消息队列       │       │  - 事件追踪        │      │  - 性能分析        │
└───────────────────┘       └────────────────────┘      └────────────────────┘
          │                            │                            │
          └────────────────────────────┼────────────────────────────┘
                                       ▼
┌──────────────────────────────────────────────────────────────────────┐
│                      Data Layer (数据层)                              │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐    │
│  │ PostgreSQL │  │   Redis    │  │  Weaviate  │  │   S3/Minio │    │
│  │ (关系数据) │  │  (缓存队列) │  │ (向量数据库)│  │ (对象存储) │    │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘    │
└──────────────────────────────────────────────────────────────────────┘
```

### 1.3 服务职责总览表

| 服务名称 | 端口 | 主要职责 | 数据存储 | 外部依赖 |
|---------|------|----------|----------|----------|
| **Event Gateway** | 8080 | 事件接收、验证、标准化 | Redis(缓存) | Alertmanager |
| **API Service** | 8081 | REST/GraphQL API、请求路由 | 无状态 | 所有服务 |
| **Dashboard Web** | 3000 | 前端UI、实时更新 | 无状态 | API Service |
| **Orchestrator** | 8082 | 任务调度、工作流控制 | PostgreSQL, Redis | Message Bus |
| **Reasoning Service** | 8083 | AI推理、策略生成 | 无状态 | OpenAI, Knowledge Service |
| **Execution Service** | 8084 | 命令执行、结果收集 | 无状态 | Kubernetes API, Credential Service |
| **Knowledge Service** | 8085 | 知识存储、语义检索 | Weaviate, PostgreSQL | - |
| **Credential Service** | 8086 | 凭证管理、动态获取 | Vault | HashiCorp Vault |
| **Report Service** | 8087 | 报告生成、历史管理 | PostgreSQL, S3 | - |
| **Notification Service** | 8088 | 多渠道通知 | Redis | Slack, Email SMTP |
| **Audit Service** | 8089 | 审计日志、合规检查 | PostgreSQL | - |
| **Monitoring Service** | 9100 | 指标收集、健康检查 | Prometheus | - |

## 2. 核心微服务设计

### 2.1 Event Gateway Service (事件网关服务)

#### 职责范围

- **核心功能**: 统一接收和处理外部事件源（Alertmanager Webhook等）
- **业务边界**: 事件入口 → 验证 → 标准化 → 发布到消息总线

**部署模式差异**：
- **单集群In-Cluster模式**:
  - Event Gateway接收Alertmanager告警
  - K8s事件由独立的Event Watcher组件处理并直接发布到消息总线
- **Agent代理模式**:
  - Event Gateway仅处理非K8s事件源（Alertmanager等）
  - K8s事件由各Agent采集并通过Agent Manager转发

#### 技术规格

```go
// 服务配置
type EventGatewayConfig struct {
    ServerPort         int           `yaml:"server_port" default:"8080"`
    WebhookTimeout     time.Duration `yaml:"webhook_timeout" default:"30s"`
    MaxPayloadSize     int64         `yaml:"max_payload_size" default:"10485760"` // 10MB
    DeduplicationTTL   time.Duration `yaml:"deduplication_ttl" default:"5m"`
    RateLimitPerMinute int           `yaml:"rate_limit_per_minute" default:"1000"`
}

// 核心接口
type EventGatewayService interface {
    // Webhook处理
    HandleAlertmanagerWebhook(ctx context.Context, payload []byte) error
    HandleCustomWebhook(ctx context.Context, source string, payload []byte) error

    // Kubernetes事件监听
    WatchK8sEvents(ctx context.Context, clusterID string) error

    // 事件验证和标准化
    ValidateEvent(event RawEvent) error
    StandardizeEvent(event RawEvent) (*StandardEvent, error)

    // 去重和过滤
    IsDuplicate(event StandardEvent) bool
    ShouldFilter(event StandardEvent) bool

    // 事件发布
    PublishEvent(ctx context.Context, event StandardEvent) error
}

// 数据模型
type StandardEvent struct {
    ID          string                 `json:"id"`
    Type        string                 `json:"type"` // alert, event, manual
    Source      string                 `json:"source"`
    Timestamp   time.Time              `json:"timestamp"`
    ClusterID   string                 `json:"cluster_id"`
    Namespace   string                 `json:"namespace"`
    Severity    string                 `json:"severity"`
    Labels      map[string]string      `json:"labels"`
    Annotations map[string]string      `json:"annotations"`
    Fingerprint string                 `json:"fingerprint"` // 用于去重
    RawData     map[string]interface{} `json:"raw_data"`
}
```

#### API端点

```
POST   /api/v1/webhook/alertmanager     # Alertmanager webhook
POST   /api/v1/webhook/custom/:source   # 自定义webhook
GET    /health                   # 健康检查
GET    /metrics                  # Prometheus指标
```

#### 部署配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: event-gateway
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: event-gateway
        image: aetherius/event-gateway:v1.0
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        env:
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: url
```

### 2.2 Orchestrator Service (编排服务)

#### 职责范围

- **核心功能**: 任务生命周期管理、调度和工作流控制
- **业务边界**: 任务创建 → 调度 → 状态管理 → 完成

#### 技术规格

```go
type OrchestratorConfig struct {
    MaxConcurrentTasks int           `yaml:"max_concurrent_tasks" default:"50"`
    TaskTimeout        time.Duration `yaml:"task_timeout" default:"10m"`
    RetryAttempts      int           `yaml:"retry_attempts" default:"3"`
    RetryDelay         time.Duration `yaml:"retry_delay" default:"30s"`
    WorkerPoolSize     int           `yaml:"worker_pool_size" default:"10"`
}

type OrchestratorService interface {
    // 任务生命周期
    CreateTask(ctx context.Context, event StandardEvent) (*DiagnosticTask, error)
    StartTask(ctx context.Context, taskID string) error
    CancelTask(ctx context.Context, taskID string) error
    CompleteTask(ctx context.Context, taskID string, result TaskResult) error

    // 任务调度
    ScheduleTasks(ctx context.Context) error
    SelectNextTask(ctx context.Context) (*DiagnosticTask, error)

    // 优先级管理
    CalculatePriority(task DiagnosticTask) int
    ReorderQueue(ctx context.Context) error

    // 状态管理
    UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error
    GetTaskStatus(ctx context.Context, taskID string) (TaskStatus, error)

    // 工作流控制
    ExecuteWorkflow(ctx context.Context, task DiagnosticTask) error
}

// 任务状态机
type TaskStatus string

const (
    StatusPending    TaskStatus = "pending"
    StatusScheduled  TaskStatus = "scheduled"
    StatusRunning    TaskStatus = "running"
    StatusCompleted  TaskStatus = "completed"
    StatusFailed     TaskStatus = "failed"
    StatusTimeout    TaskStatus = "timeout"
    StatusCancelled  TaskStatus = "cancelled"
)

// 优先级计算
func (o *Orchestrator) CalculatePriority(task DiagnosticTask) int {
    priority := 0

    // 基础优先级 (严重程度)
    switch task.Severity {
    case "critical":
        priority += 100
    case "high":
        priority += 50
    case "medium":
        priority += 10
    case "low":
        priority += 1
    }

    // 集群权重
    if clusterWeight, ok := o.clusterWeights[task.ClusterID]; ok {
        priority += clusterWeight
    }

    // 命名空间权重
    if nsWeight, ok := o.namespaceWeights[task.Namespace]; ok {
        priority += nsWeight
    }

    // 历史频率因子 (频繁出现的问题优先级提高)
    if frequency := o.getHistoricalFrequency(task.AlertName); frequency > 5 {
        priority += frequency * 2
    }

    return priority
}
```

#### API端点

```
POST   /tasks                    # 创建任务
GET    /tasks/:id                # 获取任务详情
PUT    /tasks/:id/cancel         # 取消任务
GET    /tasks                    # 查询任务列表
GET    /queue/status             # 队列状态
POST   /tasks/:id/retry          # 重试任务
```

#### 数据库Schema

```sql
CREATE TABLE diagnostic_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) NOT NULL,
    cluster_id VARCHAR(255) NOT NULL,
    namespace VARCHAR(255),
    priority INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    timeout_at TIMESTAMP,
    retry_count INTEGER DEFAULT 0,
    context JSONB,
    result JSONB,
    error_message TEXT,
    INDEX idx_status (status),
    INDEX idx_priority (priority DESC),
    INDEX idx_cluster_namespace (cluster_id, namespace),
    INDEX idx_created_at (created_at DESC)
);
```

### 2.3 Reasoning Service (推理服务)

#### 职责范围

- **核心功能**: AI驱动的智能诊断和策略生成
- **业务边界**: 任务分析 → 知识检索 → AI推理 → 执行计划

#### 技术规格

```go
type ReasoningConfig struct {
    AIProvider      string        `yaml:"ai_provider" default:"openai"`
    DefaultModel    string        `yaml:"default_model" default:"gpt-4"`
    FallbackModel   string        `yaml:"fallback_model" default:"gpt-3.5-turbo"`
    MaxTokens       int           `yaml:"max_tokens" default:"4000"`
    Temperature     float64       `yaml:"temperature" default:"0.1"`
    RequestTimeout  time.Duration `yaml:"request_timeout" default:"60s"`
    MaxRetries      int           `yaml:"max_retries" default:"3"`
}

type ReasoningService interface {
    // 诊断分析
    AnalyzeTask(ctx context.Context, task DiagnosticTask) (*AnalysisResult, error)

    // 策略生成
    GenerateStrategy(ctx context.Context, task DiagnosticTask, analysis AnalysisResult) (*DiagnosticStrategy, error)

    // 执行计划创建
    CreateExecutionPlan(ctx context.Context, strategy DiagnosticStrategy) (*ExecutionPlan, error)

    // 结果分析
    AnalyzeStepResult(ctx context.Context, step DiagnosticStep, result StepResult) (*StepAnalysis, error)

    // 决策逻辑
    DecideNextStep(ctx context.Context, currentState TaskState, analysis StepAnalysis) (NextAction, error)
}

type AnalysisResult struct {
    Summary           string                 `json:"summary"`
    PossibleCauses    []string               `json:"possible_causes"`
    SuggestedSteps    []string               `json:"suggested_steps"`
    Confidence        float64                `json:"confidence"`
    KnowledgeMatches  []KnowledgeEntry       `json:"knowledge_matches"`
    TokensUsed        int                    `json:"tokens_used"`
    Model             string                 `json:"model"`
    ResponseTime      time.Duration          `json:"response_time"`
}

type DiagnosticStrategy struct {
    ID               string              `json:"id"`
    Type             string              `json:"type"` // knowledge_based, ai_generated, rule_based
    Steps            []DiagnosticStep    `json:"steps"`
    EstimatedTime    time.Duration       `json:"estimated_time"`
    EstimatedCost    float64             `json:"estimated_cost"`
    Confidence       float64             `json:"confidence"`
}

type ExecutionPlan struct {
    ID               string              `json:"id"`
    TaskID           string              `json:"task_id"`
    Strategy         DiagnosticStrategy  `json:"strategy"`
    CurrentStepIndex int                 `json:"current_step_index"`
    TotalSteps       int                 `json:"total_steps"`
    MaxSteps         int                 `json:"max_steps"`
    TokenBudget      int                 `json:"token_budget"`
    TokensUsed       int                 `json:"tokens_used"`
    CreatedAt        time.Time           `json:"created_at"`
}
```

#### 推理流程

```go
func (r *ReasoningService) AnalyzeTask(ctx context.Context, task DiagnosticTask) (*AnalysisResult, error) {
    // 1. 知识库检索
    knowledgeMatches, err := r.knowledgeService.Search(ctx, SearchQuery{
        Text:      task.Description,
        ClusterID: task.ClusterID,
        TopK:      5,
        Threshold: 0.7,
    })
    if err != nil {
        return nil, fmt.Errorf("knowledge search failed: %w", err)
    }

    // 2. 决定使用哪种策略
    if len(knowledgeMatches) > 0 && knowledgeMatches[0].Score >= 0.8 {
        // 高置信度知识库匹配,直接使用知识库策略
        return r.generateKnowledgeBasedAnalysis(knowledgeMatches[0])
    }

    // 3. 检查AI服务可用性和预算
    if !r.aiAvailable || !r.budgetOK() {
        // 降级到规则引擎
        return r.generateRuleBasedAnalysis(task)
    }

    // 4. 构建LLM上下文
    context := r.buildContext(task, knowledgeMatches)

    // 5. 调用LLM
    model := r.selectModel(task.Complexity)
    response, err := r.llmClient.Analyze(ctx, model, context)
    if err != nil {
        log.Error("LLM analysis failed", zap.Error(err))
        // 降级到规则引擎
        return r.generateRuleBasedAnalysis(task)
    }

    // 6. 解析和验证响应
    analysis := r.parseAnalysisResponse(response)
    analysis.KnowledgeMatches = knowledgeMatches
    analysis.TokensUsed = response.TokensUsed
    analysis.Model = model

    return analysis, nil
}

func (r *ReasoningService) selectModel(complexity int) string {
    if complexity >= 7 {
        return r.config.DefaultModel // gpt-4 for complex issues
    }
    return r.config.FallbackModel // gpt-3.5-turbo for simple issues
}
```

#### API端点

```
POST   /analyze                  # 分析任务
POST   /strategies               # 生成策略
POST   /plans                    # 创建执行计划
POST   /steps/analyze            # 分析步骤结果
POST   /decisions                # 决策下一步
GET    /health                   # 健康检查
```

### 2.4 Execution Service (执行服务)

#### 职责范围

- **核心功能**: 安全执行诊断命令并收集结果
- **业务边界**: 执行计划 → 凭证获取 → 命令执行 → 结果收集

#### 技术规格

```go
type ExecutionConfig struct {
    MaxConcurrentExecutions int           `yaml:"max_concurrent_executions" default:"20"`
    CommandTimeout          time.Duration `yaml:"command_timeout" default:"5m"`
    MaxOutputSize           int64         `yaml:"max_output_size" default:"1048576"` // 1MB
    RetryAttempts           int           `yaml:"retry_attempts" default:"3"`
}

type ExecutionService interface {
    // 执行计划
    ExecutePlan(ctx context.Context, plan ExecutionPlan) (*PlanResult, error)

    // 执行单步
    ExecuteStep(ctx context.Context, step DiagnosticStep) (*StepResult, error)

    // 命令执行
    ExecuteCommand(ctx context.Context, cmd Command) (*CommandOutput, error)

    // 安全验证
    ValidateCommand(cmd Command) error
    CheckSafety(cmd Command) (SafetyScore, error)

    // 凭证管理
    GetCredentials(ctx context.Context, clusterID string) (*Credentials, error)
}

type Command struct {
    ID          string            `json:"id"`
    Tool        string            `json:"tool"` // kubectl, helm, etc.
    Action      string            `json:"action"`
    Args        []string          `json:"args"`
    Flags       map[string]string `json:"flags"`
    ClusterID   string            `json:"cluster_id"`
    Namespace   string            `json:"namespace"`
    Timeout     time.Duration     `json:"timeout"`
}

type StepResult struct {
    StepID      string            `json:"step_id"`
    Status      string            `json:"status"` // success, failed, timeout
    Output      string            `json:"output"`
    Error       string            `json:"error,omitempty"`
    ExitCode    int               `json:"exit_code"`
    Duration    time.Duration     `json:"duration"`
    Timestamp   time.Time         `json:"timestamp"`
}

// 5层安全检查
func (e *ExecutionService) ValidateCommand(cmd Command) error {
    // 1. 工具注册表验证
    if !e.toolRegistry.IsRegistered(cmd.Tool) {
        return fmt.Errorf("tool not registered: %s", cmd.Tool)
    }

    // 2. 只读操作检查
    if e.toolRegistry.IsDestructive(cmd.Tool, cmd.Action) {
        return fmt.Errorf("destructive operation not allowed: %s %s", cmd.Tool, cmd.Action)
    }

    // 3. 参数安全验证
    if err := e.validateArguments(cmd); err != nil {
        return fmt.Errorf("argument validation failed: %w", err)
    }

    // 4. RBAC权限验证
    if err := e.checkRBAC(cmd); err != nil {
        return fmt.Errorf("RBAC check failed: %w", err)
    }

    // 5. 集群权限验证
    if err := e.checkClusterAccess(cmd.ClusterID); err != nil {
        return fmt.Errorf("cluster access denied: %w", err)
    }

    return nil
}
```

#### MCP协议实现

```go
// MCP (Model Context Protocol) 服务端实现
type MCPServer struct {
    credentialService CredentialService
    k8sClientFactory  K8sClientFactory
    toolRegistry      ToolRegistry
}

func (m *MCPServer) ExecuteCommand(ctx context.Context, cmd Command) (*Result, error) {
    // 1. 验证命令
    if err := m.ValidateCommand(cmd); err != nil {
        return nil, err
    }

    // 2. 获取短期凭证
    creds, err := m.credentialService.GetCredentials(ctx, cmd.ClusterID)
    if err != nil {
        return nil, fmt.Errorf("failed to get credentials: %w", err)
    }
    defer creds.Release()

    // 3. 创建K8s客户端
    client, err := m.k8sClientFactory.Create(creds)
    if err != nil {
        return nil, fmt.Errorf("failed to create k8s client: %w", err)
    }

    // 4. 执行命令
    output, err := m.executeWithTimeout(ctx, client, cmd)
    if err != nil {
        return nil, err
    }

    return &Result{
        Output:   output,
        Duration: time.Since(start),
    }, nil
}
```

#### API端点

```
POST   /plans/execute            # 执行计划
POST   /steps/execute            # 执行单步
POST   /commands/execute         # 执行命令
POST   /commands/validate        # 验证命令
GET    /tools                    # 工具列表
GET    /health                   # 健康检查
```

### 2.5 Knowledge Service (知识库服务)

#### 职责范围

- **核心功能**: 知识存储、语义检索、知识更新
- **业务边界**: 知识管理 ← → 向量搜索

#### 技术规格

```go
type KnowledgeConfig struct {
    VectorDBEndpoint string  `yaml:"vector_db_endpoint"`
    EmbeddingModel   string  `yaml:"embedding_model" default:"text-embedding-ada-002"`
    TopK             int     `yaml:"top_k" default:"5"`
    ScoreThreshold   float64 `yaml:"score_threshold" default:"0.7"`
}

type KnowledgeService interface {
    // 知识存储
    AddKnowledge(ctx context.Context, entry KnowledgeEntry) error
    UpdateKnowledge(ctx context.Context, id string, entry KnowledgeEntry) error
    DeleteKnowledge(ctx context.Context, id string) error

    // 语义检索
    Search(ctx context.Context, query SearchQuery) ([]KnowledgeMatch, error)
    SearchByVector(ctx context.Context, vector []float64, topK int) ([]KnowledgeMatch, error)

    // 知识管理
    GetKnowledge(ctx context.Context, id string) (*KnowledgeEntry, error)
    ListKnowledge(ctx context.Context, filter KnowledgeFilter) ([]KnowledgeEntry, error)

    // 反馈处理
    ProcessFeedback(ctx context.Context, feedback UserFeedback) error
    UpdateConfidence(ctx context.Context, id string, delta float64) error
}

type KnowledgeEntry struct {
    ID          string                 `json:"id"`
    Title       string                 `json:"title"`
    Content     string                 `json:"content"`
    Category    string                 `json:"category"`
    Tags        []string               `json:"tags"`
    Embedding   []float64              `json:"embedding"`
    Confidence  float64                `json:"confidence"`
    UsageCount  int                    `json:"usage_count"`
    SuccessRate float64                `json:"success_rate"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type SearchQuery struct {
    Text         string            `json:"text"`
    ClusterID    string            `json:"cluster_id,omitempty"`
    Namespace    string            `json:"namespace,omitempty"`
    Category     string            `json:"category,omitempty"`
    Tags         []string          `json:"tags,omitempty"`
    TopK         int               `json:"top_k"`
    Threshold    float64           `json:"threshold"`
}

type KnowledgeMatch struct {
    Entry KnowledgeEntry `json:"entry"`
    Score float64        `json:"score"`
}
```

#### 向量检索实现

```go
func (k *KnowledgeService) Search(ctx context.Context, query SearchQuery) ([]KnowledgeMatch, error) {
    // 1. 生成查询向量
    embedding, err := k.embeddingClient.Embed(ctx, query.Text)
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }

    // 2. 构建Weaviate查询
    whereFilter := k.buildWhereFilter(query)

    // 3. 执行向量搜索
    results, err := k.vectorDB.Search(ctx, VectorSearchRequest{
        Vector:      embedding,
        Limit:       query.TopK,
        WhereFilter: whereFilter,
        Distance:    query.Threshold,
    })
    if err != nil {
        return nil, fmt.Errorf("vector search failed: %w", err)
    }

    // 4. 转换结果
    matches := make([]KnowledgeMatch, 0, len(results))
    for _, result := range results {
        if result.Distance >= query.Threshold {
            matches = append(matches, KnowledgeMatch{
                Entry: result.Entry,
                Score: result.Distance,
            })
        }
    }

    return matches, nil
}
```

#### API端点

```
POST   /knowledge                # 添加知识
PUT    /knowledge/:id            # 更新知识
DELETE /knowledge/:id            # 删除知识
GET    /knowledge/:id            # 获取知识
GET    /knowledge                # 列表查询
POST   /search                   # 语义搜索
POST   /feedback                 # 处理反馈
```

### 2.6 其他支撑服务

#### Credential Service (凭证服务)

```go
type CredentialService interface {
    // Vault集成
    GetK8sCredentials(ctx context.Context, clusterID string) (*K8sCredentials, error)
    RefreshCredentials(ctx context.Context, clusterID string) error
    RevokeCredentials(ctx context.Context, clusterID string) error

    // 动态凭证
    GenerateShortTermToken(ctx context.Context, clusterID string, ttl time.Duration) (*Token, error)
}

type K8sCredentials struct {
    ClusterID   string    `json:"cluster_id"`
    Token       string    `json:"token"`
    CA          string    `json:"ca"`
    Endpoint    string    `json:"endpoint"`
    ExpiresAt   time.Time `json:"expires_at"`
}
```

#### Report Service (报告服务)

```go
type ReportService interface {
    // 报告生成
    GenerateReport(ctx context.Context, task DiagnosticTask, result TaskResult) (*Report, error)
    RenderReport(ctx context.Context, report Report, format string) ([]byte, error)

    // 报告管理
    GetReport(ctx context.Context, id string) (*Report, error)
    ListReports(ctx context.Context, filter ReportFilter) ([]Report, error)
    ArchiveReport(ctx context.Context, id string) error
}
```

#### Notification Service (通知服务)

```go
type NotificationService interface {
    // 多渠道通知
    SendSlack(ctx context.Context, msg SlackMessage) error
    SendEmail(ctx context.Context, msg EmailMessage) error
    SendWebhook(ctx context.Context, url string, payload interface{}) error

    // 通知管理
    CreateNotification(ctx context.Context, notif Notification) error
    GetNotificationStatus(ctx context.Context, id string) (*NotificationStatus, error)
}
```

#### Audit Service (审计服务)

```go
type AuditService interface {
    // 审计日志
    LogEvent(ctx context.Context, event AuditEvent) error
    QueryLogs(ctx context.Context, query AuditQuery) ([]AuditEvent, error)

    // 合规检查
    CheckCompliance(ctx context.Context, standard string) (*ComplianceReport, error)
}
```

## 3. 服务间通信

**重要说明**: 本节描述的是微服务内部的通信机制,与Agent代理模式中的Agent-Central通信是独立的。

### 3.1 通信模式

| 通信场景 | 模式 | 技术选型 | 说明 |
|---------|------|----------|------|
| **事件分发** | 发布/订阅 | NATS/Kafka | 事件驱动架构,异步处理 |
| **同步调用** | REST/gRPC | HTTP/gRPC | 实时查询和命令 |
| **服务发现** | DNS/Registry | K8s Service / Consul | 动态服务发现 |
| **负载均衡** | Client-Side | K8s Service | 内置负载均衡 |

### 3.2 消息总线架构

**重要说明**: 在多集群Agent代理模式下，系统使用两个独立的NATS集群：
1. **内部NATS**: 用于微服务间的事件通信（本节描述）
2. **Agent NATS**: 用于Central与Agent间的通信（参见 09_agent_proxy_mode.md）

```text
Event Gateway
      │
      ▼ publish(event.received)
┌──────────────────────────────┐
│   内部Message Bus (NATS)     │
│                              │
│  Topics:                     │
│  - event.received            │
│  - task.created              │
│  - task.completed            │
│  - step.executed             │
│  - notification.send         │
└──────────────────────────────┘
      │
      ├─ subscribe(event.received) ──→ Orchestrator
      ├─ subscribe(task.created) ────→ Reasoning Service
      ├─ subscribe(step.executed) ───→ Report Service
      └─ subscribe(notification.send) → Notification Service
```

### 3.3 API Gateway配置

```yaml
# Kong/Traefik 路由配置
routes:
  - name: event-gateway
    paths: ["/webhook/*"]
    service: event-gateway:8080

  - name: api-service
    paths: ["/api/*"]
    service: api-service:8081
    plugins:
      - name: rate-limiting
        config:
          minute: 1000
      - name: jwt
      - name: cors

  - name: dashboard
    paths: ["/", "/ui/*"]
    service: dashboard-web:3000
```

### 3.4 服务网格 (可选)

```yaml
# Istio VirtualService 示例
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reasoning-service
spec:
  hosts:
  - reasoning-service
  http:
  - match:
    - headers:
        priority:
          exact: "high"
    route:
    - destination:
        host: reasoning-service
        subset: high-priority
      weight: 100
    timeout: 120s
  - route:
    - destination:
        host: reasoning-service
        subset: normal
      weight: 100
    timeout: 60s
```

## 4. 数据管理策略

### 4.1 数据库分配

| 服务 | 数据库 | 表/集合 | 共享/独立 |
|------|--------|---------|-----------|
| **Orchestrator** | PostgreSQL | diagnostic_tasks, task_history | 独立Schema |
| **Knowledge Service** | Weaviate + PostgreSQL | knowledge_entries (双写) | 独立 |
| **Report Service** | PostgreSQL + S3 | reports, report_artifacts | 独立Schema |
| **Audit Service** | PostgreSQL | audit_logs | 独立Schema |
| **Notification Service** | Redis | notification_queue | 独立DB |

### 4.2 数据一致性策略

```go
// 1. 强一致性: 使用数据库事务
func (o *Orchestrator) CreateTask(ctx context.Context, event StandardEvent) error {
    tx, err := o.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 创建任务记录
    task := &DiagnosticTask{...}
    if err := tx.Insert(task); err != nil {
        return err
    }

    // 写入队列
    if err := o.queue.Push(tx, task.ID); err != nil {
        return err
    }

    return tx.Commit()
}

// 2. 最终一致性: 使用消息队列
func (k *KnowledgeService) AddKnowledge(ctx context.Context, entry KnowledgeEntry) error {
    // 先写PostgreSQL
    if err := k.db.Insert(entry); err != nil {
        return err
    }

    // 异步写向量数据库
    return k.messageBus.Publish("knowledge.sync", entry)
}
```

### 4.3 缓存策略

```go
// Redis 缓存分层
type CacheStrategy struct {
    // L1: 本地内存缓存 (go-cache)
    LocalCache *cache.Cache

    // L2: Redis缓存
    RedisCache *redis.Client

    // L3: 数据库
    Database *sql.DB
}

func (c *CacheStrategy) Get(key string) (interface{}, error) {
    // L1 cache hit
    if val, found := c.LocalCache.Get(key); found {
        return val, nil
    }

    // L2 cache hit
    if val, err := c.RedisCache.Get(key).Result(); err == nil {
        c.LocalCache.Set(key, val, 5*time.Minute)
        return val, nil
    }

    // L3 database
    val, err := c.Database.Query(key)
    if err != nil {
        return nil, err
    }

    // 回写缓存
    c.RedisCache.Set(key, val, 1*time.Hour)
    c.LocalCache.Set(key, val, 5*time.Minute)

    return val, nil
}
```

## 5. 服务治理

### 5.1 服务健康检查

```go
// 健康检查接口
type HealthChecker struct {
    db        Database
    redis     RedisClient
    messageBus MessageBus
}

func (h *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
    checks := []HealthCheck{
        {Name: "database", Func: h.checkDatabase},
        {Name: "redis", Func: h.checkRedis},
        {Name: "message_bus", Func: h.checkMessageBus},
    }

    results := make([]CheckResult, len(checks))
    for i, check := range checks {
        start := time.Now()
        err := check.Func(ctx)
        results[i] = CheckResult{
            Name:     check.Name,
            Healthy:  err == nil,
            Duration: time.Since(start),
            Error:    err,
        }
    }

    return HealthStatus{
        Status:  h.calculateOverallStatus(results),
        Checks:  results,
        Version: version.Version,
    }
}
```

```yaml
# Kubernetes 健康检查配置
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3
```

### 5.2 熔断与限流

#### 重试策略
系统采用指数退避重试策略，避免级联故障：
- **默认重试次数**: 3次
- **初始重试延迟**: 30秒
- **退避因子**: 2.0（每次重试延迟翻倍）
- **最大重试延迟**: 5分钟
- **可重试错误类型**: 网络超时、服务暂时不可用(503)、限流(429)

```go
// 重试配置
type RetryConfig struct {
    MaxAttempts     int           `yaml:"max_attempts" default:"3"`
    InitialDelay    time.Duration `yaml:"initial_delay" default:"30s"`
    BackoffFactor   float64       `yaml:"backoff_factor" default:"2.0"`
    MaxDelay        time.Duration `yaml:"max_delay" default:"5m"`
    RetryableErrors []int         `yaml:"retryable_errors"` // HTTP状态码
}

// 熔断器配置
type CircuitBreakerConfig struct {
    MaxRequests       uint32        `yaml:"max_requests"`
    Interval          time.Duration `yaml:"interval"`
    Timeout           time.Duration `yaml:"timeout"`
    ReadyToTrip       func(counts Counts) bool
    OnStateChange     func(name string, from State, to State)
}

// 限流器配置
type RateLimiterConfig struct {
    RequestsPerSecond int           `yaml:"requests_per_second"`
    Burst             int           `yaml:"burst"`
    TokenBucketSize   int           `yaml:"token_bucket_size"`
}

// 使用示例
func (r *ReasoningService) AnalyzeTask(ctx context.Context, task DiagnosticTask) (*AnalysisResult, error) {
    // 限流检查
    if !r.rateLimiter.Allow() {
        return nil, ErrRateLimitExceeded
    }

    // 熔断器包装
    result, err := r.circuitBreaker.Execute(func() (interface{}, error) {
        return r.llmClient.Analyze(ctx, task)
    })

    if err != nil {
        return nil, err
    }

    return result.(*AnalysisResult), nil
}
```

### 5.3 分布式追踪

```go
// OpenTelemetry 集成
func (o *Orchestrator) CreateTask(ctx context.Context, event StandardEvent) (*DiagnosticTask, error) {
    ctx, span := tracer.Start(ctx, "orchestrator.create_task",
        trace.WithAttributes(
            attribute.String("event.id", event.ID),
            attribute.String("cluster.id", event.ClusterID),
        ),
    )
    defer span.End()

    // 业务逻辑...

    // 传播trace context到下游服务
    return o.reasoningService.AnalyzeTask(ctx, task)
}
```

### 5.4 服务监控

```yaml
# Prometheus ServiceMonitor
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aetherius-services
spec:
  selector:
    matchLabels:
      app.kubernetes.io/part-of: aetherius
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### 5.5 边界条件处理

#### 5.5.1 队列溢出处理

```go
// 事件队列溢出保护
type BoundedQueue struct {
    maxSize   int
    queue     chan interface{}
    droppedCounter prometheus.Counter
}

func (q *BoundedQueue) Push(item interface{}) error {
    select {
    case q.queue <- item:
        return nil
    default:
        // 队列已满，执行降级策略
        q.droppedCounter.Inc()

        // 策略1：丢弃最旧的事件
        select {
        case <-q.queue:
            q.queue <- item
            return nil
        default:
            return ErrQueueFull
        }
    }
}

// 任务队列优先级管理
func (o *Orchestrator) HandleQueueOverflow() {
    if o.queue.Size() > o.maxQueueSize*0.8 {
        // 80%容量时开始清理低优先级任务
        o.purgeLowPriorityTasks()

        // 提高准入门槛
        o.minAcceptPriority = PriorityHigh
    }
}
```

#### 5.5.2 并发限制处理

```go
// 并发执行限制
type ConcurrencyLimiter struct {
    semaphore chan struct{}
    timeout   time.Duration
}

func (c *ConcurrencyLimiter) Acquire(ctx context.Context) error {
    select {
    case c.semaphore <- struct{}{}:
        return nil
    case <-time.After(c.timeout):
        return ErrConcurrencyLimitExceeded
    case <-ctx.Done():
        return ctx.Err()
    }
}

// 服务级并发控制
func (e *ExecutionService) ExecuteWithLimit(ctx context.Context, task Task) error {
    if err := e.limiter.Acquire(ctx); err != nil {
        // 记录被拒绝的任务
        e.metrics.RejectedTasks.Inc()
        return fmt.Errorf("concurrency limit exceeded: %w", err)
    }
    defer e.limiter.Release()

    return e.execute(ctx, task)
}
```

#### 5.5.3 内存压力处理

```go
// 内存压力检测和响应
type MemoryMonitor struct {
    threshold   uint64  // 内存使用阈值
    checkPeriod time.Duration
}

func (m *MemoryMonitor) Start(ctx context.Context) {
    ticker := time.NewTicker(m.checkPeriod)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            var memStats runtime.MemStats
            runtime.ReadMemStats(&memStats)

            if memStats.Alloc > m.threshold {
                // 触发内存清理
                m.handleMemoryPressure(&memStats)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (m *MemoryMonitor) handleMemoryPressure(stats *runtime.MemStats) {
    // 1. 强制GC
    runtime.GC()

    // 2. 清理缓存
    if cache != nil {
        cache.Purge()
    }

    // 3. 缩减工作池
    workerPool.Shrink(0.5)

    // 4. 拒绝新请求
    circuitBreaker.Open()
}
```

#### 5.5.4 超时和死锁处理

```go
// 超时控制包装器
func WithTimeout(timeout time.Duration, fn func() error) error {
    done := make(chan error, 1)

    go func() {
        done <- fn()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        // 记录超时位置用于调试
        stack := make([]byte, 4096)
        runtime.Stack(stack, false)
        log.Error("Operation timeout",
            zap.Duration("timeout", timeout),
            zap.String("stack", string(stack)))
        return ErrTimeout
    }
}

// 死锁检测
type DeadlockDetector struct {
    locks    map[string]*LockInfo
    mu       sync.RWMutex
    timeout  time.Duration
}

func (d *DeadlockDetector) AcquireLock(id string) {
    d.mu.Lock()
    d.locks[id] = &LockInfo{
        AcquiredAt: time.Now(),
        Goroutine:  getGoroutineID(),
    }
    d.mu.Unlock()

    // 设置超时检查
    time.AfterFunc(d.timeout, func() {
        d.checkDeadlock(id)
    })
}
```

#### 5.5.5 级联故障防护

```go
// 级联故障断路器
type CascadeProtection struct {
    services    map[string]*ServiceHealth
    threshold   float64  // 故障服务比例阈值
}

func (c *CascadeProtection) CheckCascade() bool {
    failedCount := 0
    for _, health := range c.services {
        if !health.IsHealthy() {
            failedCount++
        }
    }

    failureRate := float64(failedCount) / float64(len(c.services))
    if failureRate > c.threshold {
        // 触发降级模式
        c.enableDegradedMode()
        return true
    }
    return false
}

func (c *CascadeProtection) enableDegradedMode() {
    // 1. 停止非关键服务
    c.stopNonCriticalServices()

    // 2. 降低请求接收率
    rateLimiter.SetRate(rateLimiter.GetRate() * 0.5)

    // 3. 启用快速失败模式
    for _, service := range c.services {
        service.EnableFastFail()
    }
}
```

#### 5.5.6 数据一致性边界处理

```go
// 分布式锁超时处理
func (l *DistributedLock) AcquireWithRetry(ctx context.Context, key string) error {
    maxRetries := 3
    backoff := time.Second

    for i := 0; i < maxRetries; i++ {
        err := l.tryAcquire(ctx, key)
        if err == nil {
            return nil
        }

        if errors.Is(err, ErrLockTimeout) {
            // 检查锁的持有者是否还活着
            if !l.isHolderAlive(key) {
                // 强制释放死锁
                l.forceRelease(key)
                continue
            }
        }

        select {
        case <-time.After(backoff):
            backoff *= 2
        case <-ctx.Done():
            return ctx.Err()
        }
    }

    return ErrLockAcquisitionFailed
}

// 最终一致性补偿机制
type EventualConsistency struct {
    reconciler *Reconciler
}

func (e *EventualConsistency) ReconcileData(ctx context.Context) {
    // 定期扫描数据不一致
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            inconsistencies := e.detectInconsistencies()
            for _, item := range inconsistencies {
                if err := e.reconciler.Reconcile(item); err != nil {
                    log.Error("Failed to reconcile",
                        zap.Error(err),
                        zap.Any("item", item))
                }
            }
        case <-ctx.Done():
            return
        }
    }
}
```

## 6. 部署架构

### 6.1 K8s部署拓扑

```
Namespace: aetherius-system
│
├── API Gateway (Ingress)
│   └── nginx-ingress-controller
│
├── Frontend Tier
│   ├── dashboard-web (replicas: 2)
│   └── api-service (replicas: 2)
│
├── Event Processing Tier
│   ├── event-gateway (replicas: 3)
│   └── orchestrator (replicas: 3)
│
├── Intelligence Tier
│   ├── reasoning-service (replicas: 2)
│   └── knowledge-service (replicas: 2)
│
├── Execution Tier
│   ├── execution-service (replicas: 2)
│   └── credential-service (replicas: 2)
│
├── Support Services
│   ├── report-service (replicas: 2)
│   ├── notification-service (replicas: 2)
│   ├── audit-service (replicas: 2)
│   └── monitoring-service (replicas: 1)
│
└── Data Layer
    ├── postgresql (StatefulSet, replicas: 3)
    ├── redis (StatefulSet, replicas: 3)
    ├── weaviate (Deployment, replicas: 2)
    └── nats (StatefulSet, replicas: 3)
```

### 6.2 服务依赖图

```
       ┌──────────────┐
       │ API Gateway  │
       └──────┬───────┘
              │
      ┌───────┴────────┐
      │                │
      ▼                ▼
┌───────────┐    ┌──────────┐
│ Dashboard │    │   API    │
│    Web    │    │ Service  │
└───────────┘    └─────┬────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ▼              ▼              ▼
  ┌──────────┐  ┌────────────┐  ┌──────────┐
  │  Event   │  │Orchestrator│  │  Report  │
  │ Gateway  │  │            │  │ Service  │
  └────┬─────┘  └──────┬─────┘  └──────────┘
       │               │
       │        ┌──────┴──────┐
       │        │             │
       ▼        ▼             ▼
  ┌──────────┐ ┌────────┐ ┌──────────┐
  │ Message  │ │Reasoning│ │Execution │
  │   Bus    │ │ Service │ │ Service  │
  └──────────┘ └────┬────┘ └────┬─────┘
                    │           │
                    ▼           ▼
              ┌──────────┐ ┌──────────┐
              │Knowledge │ │Credential│
              │ Service  │ │ Service  │
              └──────────┘ └──────────┘
```

### 6.3 资源配置建议

| 服务 | 副本数 | CPU Request | Memory Request | CPU Limit | Memory Limit |
|------|--------|-------------|----------------|-----------|--------------|
| **event-gateway** | 3 | 250m | 256Mi | 500m | 512Mi |
| **api-service** | 2 | 250m | 256Mi | 500m | 512Mi |
| **orchestrator** | 3 | 250m | 512Mi | 500m | 1Gi |
| **reasoning-service** | 2 | 500m | 1Gi | 1000m | 2Gi |
| **execution-service** | 2 | 250m | 512Mi | 500m | 1Gi |
| **knowledge-service** | 2 | 500m | 1Gi | 1000m | 2Gi |
| **credential-service** | 2 | 100m | 256Mi | 250m | 512Mi |
| **report-service** | 2 | 250m | 512Mi | 500m | 1Gi |
| **notification-service** | 2 | 100m | 256Mi | 250m | 512Mi |
| **audit-service** | 2 | 250m | 512Mi | 500m | 1Gi |

### 6.4 自动扩缩容

```yaml
# HorizontalPodAutoscaler 配置
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: reasoning-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: reasoning-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: aetherius_queue_depth
      target:
        type: AverageValue
        averageValue: "50"
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 30
      - type: Pods
        value: 2
        periodSeconds: 30
```

## 7. 服务开发指南

### 7.1 服务模板结构

```
service-template/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── config/
│   │   └── config.go            # 配置管理
│   ├── handler/
│   │   └── handler.go           # HTTP/gRPC处理器
│   ├── service/
│   │   └── service.go           # 业务逻辑
│   ├── repository/
│   │   └── repository.go        # 数据访问
│   └── client/
│       └── client.go            # 外部服务客户端
├── pkg/
│   └── api/
│       └── api.pb.go            # API定义 (protobuf)
├── deployments/
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── configmap.yaml
│   └── docker/
│       └── Dockerfile
├── go.mod
└── README.md
```

### 7.2 统一配置参数规范

为确保所有服务配置的一致性，所有配置参数必须遵循以下规范：

#### 7.2.1 配置参数命名规范

```yaml
# 配置参数命名规则
# 1. 使用snake_case（下划线分隔）
# 2. 层级结构清晰
# 3. 参数名称见名知意

# 正确示例
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

database:
  max_connections: 50
  connection_timeout: 5s

# 错误示例（不要使用）
serverPort: 8080              # 应该使用 server.port
readTimeout: 30s               # 应该使用 server.read_timeout
MaxConnections: 50             # 应该使用 database.max_connections
```

#### 7.2.2 时间参数格式

```yaml
# 时间参数必须包含单位
# 支持的单位: ns, us(µs), ms, s, m, h

timeout: 30s                   # 正确：30秒
retry_delay: 100ms             # 正确：100毫秒
max_idle_time: 5m              # 正确：5分钟
session_timeout: 24h           # 正确：24小时

# 错误示例
timeout: 30                    # 错误：缺少单位
retry_delay: "100"             # 错误：缺少单位
```

#### 7.2.3 资源配置格式

```yaml
# Kubernetes资源配置格式
resources:
  requests:
    memory: "256Mi"            # 必须使用引号包裹
    cpu: "250m"                # 必须使用引号包裹
  limits:
    memory: "1Gi"              # 必须使用引号包裹
    cpu: "1000m"               # 必须使用引号包裹

# 错误示例
resources:
  requests:
    memory: 256Mi              # 错误：缺少引号
    cpu: 0.25                  # 错误：应该使用250m
```

#### 7.2.4 默认值规范

```go
// Go 结构体默认值定义
type ServerConfig struct {
    Port            int           `yaml:"port" default:"8080"`
    ReadTimeout     time.Duration `yaml:"read_timeout" default:"30s"`
    WriteTimeout    time.Duration `yaml:"write_timeout" default:"30s"`
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout" default:"10s"`
}

type DatabaseConfig struct {
    Host              string        `yaml:"host" default:"localhost"`
    Port              int           `yaml:"port" default:"5432"`
    MaxConnections    int           `yaml:"max_connections" default:"50"`
    IdleConnections   int           `yaml:"idle_connections" default:"10"`
    ConnectionTimeout time.Duration `yaml:"connection_timeout" default:"5s"`
}

type RetryConfig struct {
    MaxAttempts   int           `yaml:"max_attempts" default:"3"`
    InitialDelay  time.Duration `yaml:"initial_delay" default:"30s"`
    BackoffFactor float64       `yaml:"backoff_factor" default:"2.0"`
    MaxDelay      time.Duration `yaml:"max_delay" default:"5m"`
}
```

#### 7.2.5 环境变量映射

```yaml
# 环境变量命名规范
# 格式: {SERVICE_NAME}_{PARAMETER_PATH}
# 使用大写字母和下划线

# 配置文件参数 -> 环境变量映射
server.port           -> AETHERIUS_SERVER_PORT
database.host         -> AETHERIUS_DATABASE_HOST
redis.password        -> AETHERIUS_REDIS_PASSWORD
ai_service.api_key    -> AETHERIUS_AI_SERVICE_API_KEY

# Kubernetes Deployment 示例
env:
- name: AETHERIUS_SERVER_PORT
  value: "8080"
- name: AETHERIUS_DATABASE_HOST
  value: "postgresql"
- name: AETHERIUS_REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: redis-secret
      key: password
```

#### 7.2.6 配置文件模板

```yaml
# config.yaml - 标准配置文件模板
# 所有服务必须遵循此结构

# 服务器配置
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 10s

# 数据库配置
database:
  host: localhost
  port: 5432
  database: aetherius
  username: aetherius
  password: ""  # 从环境变量或密钥管理系统获取
  max_connections: 50
  idle_connections: 10
  connection_lifetime: 1h
  connection_timeout: 5s

# Redis配置
redis:
  host: localhost
  port: 6379
  password: ""  # 从环境变量或密钥管理系统获取
  db: 0
  pool_size: 20
  idle_timeout: 5m
  dial_timeout: 5s

# 消息总线配置
message_bus:
  type: nats  # nats, kafka
  endpoints:
    - "nats://localhost:4222"
  reconnect_delay: 2s
  max_reconnect_attempts: 10

# 日志配置
logging:
  level: info  # debug, info, warn, error
  format: json  # json, text
  output: stdout  # stdout, file
  file_path: "/var/log/aetherius/service.log"
  max_size: 100  # MB
  max_backups: 10
  max_age: 30  # days

# 监控配置
monitoring:
  enabled: true
  metrics_port: 9090
  metrics_path: "/metrics"
  health_check_port: 8080
  health_check_path: "/health"

# 安全配置
security:
  enable_tls: false
  cert_file: "/etc/aetherius/certs/tls.crt"
  key_file: "/etc/aetherius/certs/tls.key"
  enable_auth: true
  auth_type: jwt  # jwt, oauth2, basic

# 服务发现配置
discovery:
  type: kubernetes  # kubernetes, consul, etcd
  namespace: aetherius-system
  service_name: ""  # 自动从POD_NAME获取

# 限流配置
rate_limiting:
  enabled: true
  requests_per_second: 100
  burst: 200

# 重试配置
retry:
  max_attempts: 3
  initial_delay: 30s
  backoff_factor: 2.0
  max_delay: 5m
  retryable_errors:
    - 503  # Service Unavailable
    - 504  # Gateway Timeout
    - 429  # Too Many Requests
```

#### 7.2.7 配置验证规则

```go
// 配置验证函数
func ValidateConfig(cfg *Config) error {
    var errs []error

    // 端口范围验证
    if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
        errs = append(errs, fmt.Errorf("invalid port: %d", cfg.Server.Port))
    }

    // 超时时间验证
    if cfg.Server.ReadTimeout < time.Second {
        errs = append(errs, fmt.Errorf("read_timeout too short: %v", cfg.Server.ReadTimeout))
    }

    // 数据库连接池验证
    if cfg.Database.MaxConnections < cfg.Database.IdleConnections {
        errs = append(errs, fmt.Errorf("max_connections must be >= idle_connections"))
    }

    // 重试配置验证
    if cfg.Retry.BackoffFactor < 1.0 {
        errs = append(errs, fmt.Errorf("backoff_factor must be >= 1.0"))
    }

    if len(errs) > 0 {
        return fmt.Errorf("config validation failed: %v", errs)
    }

    return nil
}
```

### 7.3 服务开发检查清单

```markdown
## 功能开发
- [ ] 定义清晰的服务职责边界
- [ ] 实现核心业务逻辑
- [ ] 添加输入验证
- [ ] 实现错误处理和重试机制
- [ ] 添加单元测试 (覆盖率 > 80%)

## API设计
- [ ] 定义RESTful/gRPC API
- [ ] 编写API文档 (Swagger/OpenAPI)
- [ ] 版本化API (v1, v2...)
- [ ] 实现向后兼容

## 可观测性
- [ ] 添加结构化日志
- [ ] 暴露Prometheus指标
- [ ] 集成分布式追踪
- [ ] 实现健康检查接口

## 安全性
- [ ] 实现认证和授权
- [ ] 添加输入验证和清理
- [ ] 使用安全的默认配置
- [ ] 敏感信息加密存储

## 部署
- [ ] 编写Dockerfile
- [ ] 创建K8s manifests
- [ ] 配置资源限制
- [ ] 设置自动扩缩容规则
- [ ] 实施滚动更新策略

## 运维
- [ ] 编写运维文档
- [ ] 添加告警规则
- [ ] 实现优雅关闭
- [ ] 准备故障排查手册
```

## 附录

### A. 服务端口分配表

| 服务 | HTTP端口 | Metrics端口 | gRPC端口 (可选) |
|------|----------|-------------|----------------|
| event-gateway | 8080 | 9090 | - |
| api-service | 8081 | 9091 | - |
| orchestrator | 8082 | 9092 | - |
| reasoning-service | 8083 | 9093 | - |
| execution-service | 8084 | 9094 | - |
| knowledge-service | 8085 | 9095 | 50051 |
| credential-service | 8086 | 9096 | - |
| report-service | 8087 | 9097 | - |
| notification-service | 8088 | 9098 | - |
| audit-service | 8089 | 9099 | - |
| dashboard-web | 3000 | - | - |

### B. 相关文档

- [架构设计文档](./02_architecture.md) - 系统架构总览
- [数据模型文档](./03_data_models.md) - 数据结构定义
- [部署配置文档](./04_deployment.md) - 部署指南
- [运维安全文档](./05_operations.md) - 运维和安全

### C. 技术债务与未来规划

#### 当前限制

- 消息总线使用NATS,需要评估Kafka的适用性
- 部分服务间同步调用,可优化为异步
- 缺少完整的服务网格集成

#### 未来改进

- [ ] 引入服务网格 (Istio/Linkerd)
- [ ] 实施更细粒度的服务拆分
- [ ] 增强服务间的弹性设计
- [ ] 实现智能路由和流量管理
- [ ] 引入Saga模式处理分布式事务