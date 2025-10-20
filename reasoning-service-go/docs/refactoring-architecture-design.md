# Reasoning Service Go 重构架构设计

## 文档版本

- 版本: 1.0
- 日期: 2025-10-20
- 作者: Architecture Team

## 1. 执行摘要

### 1.1 重构目标

基于 go-llm-proxy 与 LangChainGo 重构当前 Reasoning Service,实现统一的 LLM 访问层、标准化的推理链、可扩展的 Agent 框架和智能内存管理。

### 1.2 关键收益

- **统一 LLM 接口**: 使用 go-llm-proxy 提供统一的多供应商 API 访问
- **标准化推理链**: 使用 LangChainGo 实现可组合的推理流程
- **可扩展性**: 模块化架构便于添加新的分析能力和 LLM 提供商
- **智能上下文管理**: 使用 LangChainGo Memory 实现对话历史和上下文持久化
- **降低维护成本**: 减少自定义代码,依赖成熟的开源框架

## 2. 当前架构分析

### 2.1 现有系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP API Server                          │
│                   (internal/api/server.go)                   │
└────────────────┬────────────────────┬────────────────────────┘
                 │                    │
                 ▼                    ▼
        ┌────────────────┐    ┌───────────────────┐
        │ Root Cause     │    │  Recommender      │
        │ Analyzer       │    │  Engine           │
        │ (analyzer/)    │    │  (recommender/)   │
        └────────┬───────┘    └────────┬──────────┘
                 │                     │
                 └──────────┬──────────┘
                            ▼
                 ┌─────────────────────┐
                 │   LLM Client Layer  │
                 │   (pkg/llm/)        │
                 │                     │
                 │  - Factory Pattern  │
                 │  - Custom Clients   │
                 │  - Direct API Call  │
                 └─────────────────────┘
```

### 2.2 核心模块职责

#### 2.2.1 API Server (internal/api/server.go:26-33)

- HTTP 路由和请求处理
- CORS 和日志中间件
- 健康检查端点
- 根因分析端点 (`/api/v1/analyze/root-cause`)
- K8s 事件分析端点 (`/api/v1/analyze/k8s-event`)

#### 2.2.2 Root Cause Analyzer (internal/analyzer/root_cause.go:15-31)

```go
type RootCauseAnalyzer struct {
    config     *config.Config
    llmClients []llm.Client
    patterns   map[types.RootCauseType][]*regexp.Regexp
}
```

分析策略:

1. **基于事件的分析** (analyzeEvent) - K8s 事件到根因类型的直接映射
2. **基于日志的分析** (analyzeLogs) - 正则表达式模式匹配
3. **基于指标的分析** (analyzeMetrics) - 资源使用率阈值检测
4. **基于 LLM 的分析** (analyzeLLM) - 调用 LLM 进行深度分析

#### 2.2.3 Recommender Engine (internal/recommender/engine.go:13-29)

```go
type Engine struct {
    config     *config.Config
    llmClients []llm.Client
    rules      map[types.RootCauseType][]types.Recommendation
}
```

功能:

- 基于规则的建议生成 (45+ 个根因类型的预定义规则)
- 可选的 LLM 增强建议
- 建议包含: 操作步骤、命令、YAML 配置、风险评估

#### 2.2.4 LLM Client Layer (pkg/llm/)

自定义实现:

```go
type Client interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    AnalyzeRootCause(ctx context.Context, event, logs, metrics string) (string, error)
    GenerateRecommendations(ctx context.Context, rootCause, context string) (string, error)
    Provider() Provider
    IsAvailable() bool
}
```

支持的提供商 (pkg/llm/factory.go:6-25):

- OpenAI
- Gemini
- DeepSeek
- Ollama
- SiliconFlow
- Kimi
- Custom

### 2.3 调用流程

```
用户请求
   │
   ▼
┌──────────────────────────────────────────────┐
│ 1. API Server 接收请求                        │
│    - 解析 JSON body                           │
│    - 设置超时 context                         │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│ 2. RootCauseAnalyzer.Analyze()               │
│    ├─ analyzeEvent() - 事件分析              │
│    ├─ analyzeLogs() - 日志模式匹配           │
│    ├─ analyzeMetrics() - 指标阈值检测        │
│    └─ analyzeLLM() - LLM 深度分析 (可选)     │
│       └─ 遍历 llmClients 直到成功            │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│ 3. RecommenderEngine.GenerateRecommendations()│
│    ├─ getRuleBasedRecommendations()          │
│    └─ getLLMRecommendations() (可选)         │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│ 4. 返回 AnalysisResult                        │
│    - root_cause                               │
│    - recommendations                          │
│    - confidence                               │
│    - evidence                                 │
└──────────────────────────────────────────────┘
```

### 2.4 现有架构的问题

#### 2.4.1 LLM 集成层面

- **手动实现多个 LLM 客户端**: 每个提供商需要单独实现 HTTP 客户端和错误处理
- **缺乏统一的错误处理**: 各个客户端的错误处理逻辑不一致
- **重试逻辑分散**: 手动在 analyzer 层面遍历 llmClients 实现回退
- **提示词管理**: 硬编码在 `pkg/llm/prompts.go`,缺乏版本控制和 A/B 测试能力

#### 2.4.2 推理流程层面

- **流程硬编码**: 分析步骤固定,难以调整执行顺序或添加新步骤
- **无状态上下文**: 每次请求独立处理,无法利用历史分析结果
- **缺乏复杂推理能力**: 无法实现 ReAct、Tree of Thought 等高级推理模式
- **工具集成困难**: 无标准化的外部工具调用接口

#### 2.4.3 可维护性层面

- **代码重复**: LLM 客户端实现有大量重复逻辑
- **测试困难**: LLM 调用难以模拟和测试
- **扩展性差**: 添加新的 LLM 提供商需要大量工作

## 3. 重构方案设计

### 3.1 整体架构

```
┌───────────────────────────────────────────────────────────────────┐
│                      HTTP API Gateway                              │
│                   (internal/api/server.go)                         │
└───────────────────────────┬───────────────────────────────────────┘
                            │
                            ▼
┌───────────────────────────────────────────────────────────────────┐
│                   Reasoning Orchestrator                           │
│                (internal/orchestrator/orchestrator.go)             │
│  ┌────────────────────────────────────────────────────┐           │
│  │         LangChain Chains & Agents                  │           │
│  │  - AnalysisChain                                   │           │
│  │  - RecommendationChain                             │           │
│  │  - K8sToolAgent                                    │           │
│  └────────────────┬───────────────────────────────────┘           │
└───────────────────┼───────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
┌──────────────┐ ┌─────────┐ ┌──────────────┐
│   LLM Proxy  │ │ Memory  │ │  Tool Layer  │
│   Adapter    │ │ Manager │ │              │
└──────┬───────┘ └─────────┘ └──────────────┘
       │
       ▼
┌────────────────────────────────────────────┐
│        Unified LLM Proxy Layer             │
│        (go-llm-router / gollm)             │
│  ┌─────────────────────────────────────┐  │
│  │  Unified OpenAI-Compatible API      │  │
│  │  - Automatic Fallback & Retry       │  │
│  │  - Cost Tracking                    │  │
│  │  - Rate Limiting                    │  │
│  └─────────────────────────────────────┘  │
└────────────┬───────────────────────────────┘
             │
    ┌────────┴────────┬────────┬────────┐
    ▼                 ▼        ▼        ▼
┌─────────┐  ┌──────────┐  ┌──────┐  ┌──────┐
│ OpenAI  │  │  Gemini  │  │Ollama│  │ ...  │
└─────────┘  └──────────┘  └──────┘  └──────┘
```

### 3.2 核心模块设计

#### 3.2.1 LLM Proxy Adapter 层

**职责**: 封装 go-llm-proxy (gollm) 提供统一的 LLM 访问接口

**文件**: `pkg/llm/proxy/adapter.go`

```go
// ProxyAdapter 封装 gollm 提供统一接口
type ProxyAdapter struct {
    client *gollm.Client
    config *ProxyConfig
}

type ProxyConfig struct {
    Providers []ProviderConfig
    Fallback  FallbackConfig
    Tracking  TrackingConfig
}

// 统一的方法
func (a *ProxyAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
func (a *ProxyAdapter) CompleteWithFallback(ctx context.Context, req *CompletionRequest, providers []string) (*CompletionResponse, error)
func (a *ProxyAdapter) GetProviderStatus() map[string]ProviderStatus
```

**关键特性**:

- 自动故障转移到备用提供商
- 统一的错误处理和重试逻辑
- 成本跟踪和使用统计
- 提供商健康检查

#### 3.2.2 LangChain Integration 层

**职责**: 集成 LangChainGo 实现可组合的推理链

**文件**: `internal/chains/`

##### Analysis Chain

```go
// AnalysisChain 根因分析链
type AnalysisChain struct {
    llm    llms.LLM
    memory memory.Memory
    tools  []tools.Tool
}

// 分析步骤
// 1. 数据预处理 (PreprocessChain)
// 2. 规则匹配 (RuleBasedAnalysis)
// 3. LLM 深度分析 (LLMAnalysisChain)
// 4. 结果聚合 (AggregationChain)

func (c *AnalysisChain) Run(ctx context.Context, input AnalysisInput) (*AnalysisOutput, error)
```

##### Recommendation Chain

```go
// RecommendationChain 建议生成链
type RecommendationChain struct {
    llm    llms.LLM
    memory memory.Memory
    rules  *RuleEngine
}

// 建议步骤
// 1. 加载规则模板 (LoadRulesChain)
// 2. LLM 增强建议 (LLMEnhancementChain)
// 3. 风险评估 (RiskAssessmentChain)
// 4. 优先级排序 (PrioritizationChain)

func (c *RecommendationChain) Run(ctx context.Context, input RecommendationInput) (*RecommendationOutput, error)
```

##### Tool Agent

```go
// K8sToolAgent 具有工具调用能力的 Agent
type K8sToolAgent struct {
    llm     llms.LLM
    memory  memory.Memory
    tools   []tools.Tool
    executor agents.Executor
}

// 可用工具
// - KubectlDescribe: 获取资源详细信息
// - KubectlLogs: 获取日志
// - KubectlTop: 获取资源使用情况
// - MetricsQuery: 查询 Prometheus 指标

func (a *K8sToolAgent) Execute(ctx context.Context, task string) (*ExecutionResult, error)
```

**关键特性**:

- 模块化的链组合
- 可插拔的工具集成
- 上下文和记忆管理
- 支持复杂的推理模式 (ReAct, Tree of Thought)

#### 3.2.3 Memory Manager 层

**职责**: 管理对话历史和分析上下文

**文件**: `internal/memory/manager.go`

```go
// MemoryManager 管理多种类型的记忆
type MemoryManager struct {
    conversationMemory memory.ConversationBufferMemory
    vectorStore        memory.VectorStoreMemory
    structuredMemory   *StructuredMemory
}

// StructuredMemory 结构化的历史分析数据
type StructuredMemory struct {
    RecentAnalyses   []AnalysisResult
    CommonPatterns   map[types.RootCauseType][]Pattern
    SuccessfulActions map[string]ActionOutcome
}

func (m *MemoryManager) SaveAnalysis(ctx context.Context, result *AnalysisResult) error
func (m *MemoryManager) FindSimilarCases(ctx context.Context, query AnalysisContext) ([]CaseStudy, error)
func (m *MemoryManager) GetConversationHistory(ctx context.Context, sessionID string) ([]Message, error)
```

**关键特性**:

- 短期对话记忆 (ConversationBufferMemory)
- 长期向量存储 (VectorStoreMemory)
- 结构化历史数据存储
- 相似案例检索

#### 3.2.4 Orchestrator 层

**职责**: 协调各个模块完成完整的分析流程

**文件**: `internal/orchestrator/orchestrator.go`

```go
// Orchestrator 协调分析流程
type Orchestrator struct {
    analysisChain       *chains.AnalysisChain
    recommendationChain *chains.RecommendationChain
    toolAgent           *agents.K8sToolAgent
    memoryManager       *memory.MemoryManager
    config              *config.Config
}

func (o *Orchestrator) AnalyzeRootCause(ctx context.Context, req *types.AnalysisRequest) (*types.AnalysisResult, error) {
    // 1. 加载历史上下文
    history, _ := o.memoryManager.GetConversationHistory(ctx, req.SessionID)

    // 2. 执行分析链
    analysisResult, err := o.analysisChain.Run(ctx, chains.AnalysisInput{
        Request: req,
        History: history,
    })

    // 3. 如果需要更多信息,使用 Agent 调用工具
    if analysisResult.NeedsMoreInfo {
        toolResult, _ := o.toolAgent.Execute(ctx, analysisResult.ToolRequest)
        analysisResult = o.mergeToolResults(analysisResult, toolResult)
    }

    // 4. 生成建议
    recommendations, err := o.recommendationChain.Run(ctx, chains.RecommendationInput{
        RootCause: analysisResult.RootCause,
        Context:   req.Context,
    })

    // 5. 保存到记忆
    result := o.buildAnalysisResult(analysisResult, recommendations)
    o.memoryManager.SaveAnalysis(ctx, result)

    return result, nil
}
```

### 3.3 数据流设计

#### 3.3.1 分析请求流程

```
用户请求
   │
   ▼
┌────────────────────────────────────────────┐
│ 1. API Gateway 接收请求                    │
│    - 解析并验证请求                        │
│    - 创建 session context                  │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 2. Orchestrator 开始协调                   │
│    ├─ 加载 Memory (历史分析、相似案例)    │
│    └─ 初始化 Chains                        │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 3. AnalysisChain 执行                      │
│    ├─ PreprocessChain                      │
│    │   └─ 数据清洗、格式化                 │
│    ├─ RuleBasedAnalysis                    │
│    │   └─ 快速模式匹配 (无 LLM 调用)      │
│    ├─ LLMAnalysisChain (如需要)            │
│    │   └─ LLM Proxy Adapter                │
│    │       └─ Unified LLM API Call         │
│    │           └─ Automatic Fallback       │
│    └─ AggregationChain                     │
│        └─ 聚合多个分析结果                 │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 4. K8sToolAgent 补充信息 (可选)            │
│    ├─ Agent 决定需要调用哪些工具           │
│    ├─ 执行工具: kubectl describe/logs/top  │
│    └─ 将工具结果反馈给 Chain               │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 5. RecommendationChain 执行                │
│    ├─ LoadRulesChain (45+ 规则模板)       │
│    ├─ LLMEnhancementChain (可选)           │
│    │   └─ 使用 LLM 增强建议描述            │
│    ├─ RiskAssessmentChain                  │
│    │   └─ 评估每个建议的风险级别           │
│    └─ PrioritizationChain                  │
│        └─ 根据置信度和风险排序             │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 6. Memory Manager 保存结果                 │
│    ├─ 保存到 ConversationMemory            │
│    ├─ 保存到 VectorStore (相似性检索)     │
│    └─ 更新 StructuredMemory (统计分析)    │
└────────────┬───────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────┐
│ 7. 返回 AnalysisResult                     │
│    - root_cause                            │
│    - recommendations (已排序)              │
│    - confidence                            │
│    - evidence                              │
│    - similar_cases (来自 Memory)           │
└────────────────────────────────────────────┘
```

#### 3.3.2 LLM 调用优化流程

```
AnalysisChain 需要 LLM
         │
         ▼
┌─────────────────────────────────────────┐
│  LLM Proxy Adapter                      │
│                                         │
│  1. 构建 gollm Request                  │
│     - prompt engineering               │
│     - context injection                │
│     - parameter tuning                 │
│                                         │
│  2. 调用 gollm Client                   │
│     └─ Unified API (OpenAI-compatible) │
│                                         │
│  3. gollm 内部处理                      │
│     ├─ 选择 Primary Provider           │
│     ├─ 执行 API Call                   │
│     ├─ 错误检测                        │
│     │  └─ Timeout / Rate Limit / Error │
│     ├─ Automatic Retry (3 次)          │
│     └─ Fallback 到下一个 Provider      │
│                                         │
│  4. 返回 Response                       │
│     - content                          │
│     - provider (实际使用的)            │
│     - cost (调用成本)                  │
│     - latency                          │
└─────────────────────────────────────────┘
```

### 3.4 接口定义

#### 3.4.1 LLM Proxy Adapter 接口

```go
// pkg/llm/proxy/adapter.go

package proxy

import (
    "context"
    "github.com/teilomillet/gollm"
)

// ProxyAdapter 统一的 LLM 代理适配器
type ProxyAdapter struct {
    client *gollm.LLM
    config *AdapterConfig
}

type AdapterConfig struct {
    // 主提供商配置
    PrimaryProvider string
    // 备用提供商列表 (按优先级)
    FallbackProviders []string
    // 重试配置
    MaxRetries int
    RetryDelay time.Duration
    // 超时配置
    RequestTimeout time.Duration
    // 成本跟踪
    EnableCostTracking bool
}

// NewProxyAdapter 创建新的代理适配器
func NewProxyAdapter(config *AdapterConfig) (*ProxyAdapter, error) {
    llm, err := gollm.NewLLM(
        gollm.SetProvider(config.PrimaryProvider),
        gollm.SetModel("gpt-4"),
        gollm.SetMaxRetries(config.MaxRetries),
        gollm.SetMaxTokens(2000),
    )
    if err != nil {
        return nil, err
    }

    return &ProxyAdapter{
        client: llm,
        config: config,
    }, nil
}

// Complete 标准的补全请求
func (a *ProxyAdapter) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

// CompleteWithRetry 带自动重试的补全
func (a *ProxyAdapter) CompleteWithRetry(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

// StreamComplete 流式补全
func (a *ProxyAdapter) StreamComplete(ctx context.Context, req *CompletionRequest, handler StreamHandler) error

// GetMetrics 获取使用指标
func (a *ProxyAdapter) GetMetrics() *UsageMetrics

type CompletionRequest struct {
    Messages    []Message
    Temperature float64
    MaxTokens   int
    StopWords   []string
}

type CompletionResponse struct {
    Content      string
    Provider     string  // 实际使用的提供商
    Model        string
    TokensUsed   int
    Cost         float64
    Latency      time.Duration
    FinishReason string
}

type UsageMetrics struct {
    TotalRequests   int
    SuccessfulCalls int
    FailedCalls     int
    TotalCost       float64
    ProviderStats   map[string]ProviderMetrics
}
```

#### 3.4.2 LangChain Chains 接口

```go
// internal/chains/analysis_chain.go

package chains

import (
    "context"
    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/memory"
)

// AnalysisChain 根因分析链
type AnalysisChain struct {
    llm          llms.LLM
    memory       memory.Memory
    preprocessor *PreprocessChain
    ruleEngine   *RuleBasedAnalysis
    llmAnalyzer  *LLMAnalysisChain
    aggregator   *AggregationChain
}

// AnalysisInput 分析输入
type AnalysisInput struct {
    Request *types.AnalysisRequest
    History []memory.Message
}

// AnalysisOutput 分析输出
type AnalysisOutput struct {
    RootCause      *types.RootCause
    Confidence     float64
    Evidence       []string
    NeedsMoreInfo  bool
    ToolRequest    string
    IntermediateResults map[string]interface{}
}

// NewAnalysisChain 创建分析链
func NewAnalysisChain(llm llms.LLM, mem memory.Memory) *AnalysisChain {
    return &AnalysisChain{
        llm:          llm,
        memory:       mem,
        preprocessor: NewPreprocessChain(),
        ruleEngine:   NewRuleBasedAnalysis(),
        llmAnalyzer:  NewLLMAnalysisChain(llm),
        aggregator:   NewAggregationChain(),
    }
}

// Run 执行分析链
func (c *AnalysisChain) Run(ctx context.Context, input AnalysisInput) (*AnalysisOutput, error) {
    // 1. 预处理
    preprocessed, err := c.preprocessor.Process(ctx, input)
    if err != nil {
        return nil, err
    }

    // 2. 规则匹配 (快速路径)
    ruleResult, err := c.ruleEngine.Analyze(ctx, preprocessed)
    if err == nil && ruleResult.Confidence > 0.8 {
        // 高置信度,直接返回
        return ruleResult, nil
    }

    // 3. LLM 深度分析
    llmResult, err := c.llmAnalyzer.Analyze(ctx, preprocessed, ruleResult)
    if err != nil {
        return nil, err
    }

    // 4. 聚合结果
    finalResult, err := c.aggregator.Aggregate(ctx, ruleResult, llmResult)
    if err != nil {
        return nil, err
    }

    return finalResult, nil
}

// PreprocessChain 数据预处理链
type PreprocessChain struct{}

func (c *PreprocessChain) Process(ctx context.Context, input AnalysisInput) (*PreprocessedData, error)

// RuleBasedAnalysis 基于规则的分析
type RuleBasedAnalysis struct {
    patterns map[types.RootCauseType][]*regexp.Regexp
}

func (a *RuleBasedAnalysis) Analyze(ctx context.Context, data *PreprocessedData) (*AnalysisOutput, error)

// LLMAnalysisChain LLM 分析链
type LLMAnalysisChain struct {
    llm    llms.LLM
    prompt *prompts.PromptTemplate
}

func (c *LLMAnalysisChain) Analyze(ctx context.Context, data *PreprocessedData, hint *AnalysisOutput) (*AnalysisOutput, error)

// AggregationChain 结果聚合链
type AggregationChain struct{}

func (c *AggregationChain) Aggregate(ctx context.Context, results ...*AnalysisOutput) (*AnalysisOutput, error)
```

#### 3.4.3 Tool Agent 接口

```go
// internal/agents/k8s_tool_agent.go

package agents

import (
    "context"
    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/tools"
)

// K8sToolAgent K8s 工具调用 Agent
type K8sToolAgent struct {
    llm      llms.LLM
    executor agents.Executor
    tools    map[string]tools.Tool
}

// NewK8sToolAgent 创建 K8s Tool Agent
func NewK8sToolAgent(llm llms.LLM) *K8sToolAgent {
    toolList := []tools.Tool{
        NewKubectlDescribeTool(),
        NewKubectlLogsTool(),
        NewKubectlTopTool(),
        NewMetricsQueryTool(),
    }

    executor := agents.NewExecutor(
        agents.NewZeroShotAgent(llm, toolList),
    )

    return &K8sToolAgent{
        llm:      llm,
        executor: executor,
        tools:    buildToolMap(toolList),
    }
}

// Execute 执行任务
func (a *K8sToolAgent) Execute(ctx context.Context, task string) (*ExecutionResult, error) {
    result, err := a.executor.Call(ctx, map[string]any{
        "input": task,
    })
    if err != nil {
        return nil, err
    }

    return &ExecutionResult{
        Output:      result["output"].(string),
        ToolsUsed:   result["intermediate_steps"],
        Reasoning:   result["reasoning"].(string),
    }, nil
}

// Tool 定义

// KubectlDescribeTool 获取资源详细信息
type KubectlDescribeTool struct {
    k8sClient *kubernetes.Clientset
}

func (t *KubectlDescribeTool) Name() string { return "kubectl_describe" }
func (t *KubectlDescribeTool) Description() string {
    return "Get detailed information about a Kubernetes resource. Input should be resource type and name, e.g., 'pod/nginx-xxx'"
}
func (t *KubectlDescribeTool) Call(ctx context.Context, input string) (string, error)

// KubectlLogsTool 获取 Pod 日志
type KubectlLogsTool struct {
    k8sClient *kubernetes.Clientset
}

func (t *KubectlLogsTool) Name() string { return "kubectl_logs" }
func (t *KubectlLogsTool) Description() string {
    return "Get logs from a pod. Input should be pod name and optional container name."
}
func (t *KubectlLogsTool) Call(ctx context.Context, input string) (string, error)

// 其他工具类似...
```

#### 3.4.4 Memory Manager 接口

```go
// internal/memory/manager.go

package memory

import (
    "context"
    "github.com/tmc/langchaingo/memory"
    "github.com/tmc/langchaingo/vectorstores"
)

// MemoryManager 记忆管理器
type MemoryManager struct {
    // 短期对话记忆
    conversationMemory memory.ConversationBufferMemory

    // 长期向量存储
    vectorStore vectorstores.VectorStore

    // 结构化历史数据
    structuredStore *StructuredMemoryStore
}

// NewMemoryManager 创建记忆管理器
func NewMemoryManager(vectorStore vectorstores.VectorStore) *MemoryManager {
    return &MemoryManager{
        conversationMemory: memory.NewConversationBufferMemory(),
        vectorStore:        vectorStore,
        structuredStore:    NewStructuredMemoryStore(),
    }
}

// SaveAnalysis 保存分析结果
func (m *MemoryManager) SaveAnalysis(ctx context.Context, result *types.AnalysisResult) error {
    // 1. 保存到对话记忆
    m.conversationMemory.SaveContext(ctx, map[string]any{
        "input":  result.RequestID,
        "output": result.Result,
    })

    // 2. 保存到向量存储 (用于相似性检索)
    embedding := m.generateEmbedding(result)
    m.vectorStore.AddDocuments(ctx, []schema.Document{{
        PageContent: result.toText(),
        Metadata:    result.Metadata,
    }})

    // 3. 保存到结构化存储
    m.structuredStore.Save(result)

    return nil
}

// FindSimilarCases 查找相似案例
func (m *MemoryManager) FindSimilarCases(ctx context.Context, query types.AnalysisContext) ([]types.CaseStudy, error) {
    // 使用向量相似度搜索
    results, err := m.vectorStore.SimilaritySearch(ctx, query.toText(), 5)
    if err != nil {
        return nil, err
    }

    var cases []types.CaseStudy
    for _, doc := range results {
        case := m.documentToCase(doc)
        cases = append(cases, case)
    }

    return cases, nil
}

// GetConversationHistory 获取对话历史
func (m *MemoryManager) GetConversationHistory(ctx context.Context, sessionID string) ([]memory.ChatMessage, error) {
    messages, err := m.conversationMemory.LoadMemoryVariables(ctx, map[string]any{
        "session_id": sessionID,
    })
    if err != nil {
        return nil, err
    }

    return messages["history"].([]memory.ChatMessage), nil
}

// StructuredMemoryStore 结构化存储
type StructuredMemoryStore struct {
    recentAnalyses []types.AnalysisResult
    patterns       map[types.RootCauseType][]Pattern
    outcomes       map[string]ActionOutcome
}

func (s *StructuredMemoryStore) Save(result *types.AnalysisResult) error
func (s *StructuredMemoryStore) GetRecentAnalyses(limit int) []types.AnalysisResult
func (s *StructuredMemoryStore) GetCommonPatterns(rootCause types.RootCauseType) []Pattern
```

#### 3.4.5 Orchestrator 接口

```go
// internal/orchestrator/orchestrator.go

package orchestrator

import (
    "context"
    "reasoning-service-go/internal/chains"
    "reasoning-service-go/internal/agents"
    "reasoning-service-go/internal/memory"
    "reasoning-service-go/pkg/types"
)

// Orchestrator 协调器
type Orchestrator struct {
    analysisChain       *chains.AnalysisChain
    recommendationChain *chains.RecommendationChain
    toolAgent           *agents.K8sToolAgent
    memoryManager       *memory.MemoryManager
    config              *config.Config
}

// NewOrchestrator 创建协调器
func NewOrchestrator(
    analysisChain *chains.AnalysisChain,
    recommendationChain *chains.RecommendationChain,
    toolAgent *agents.K8sToolAgent,
    memoryManager *memory.MemoryManager,
    config *config.Config,
) *Orchestrator {
    return &Orchestrator{
        analysisChain:       analysisChain,
        recommendationChain: recommendationChain,
        toolAgent:           toolAgent,
        memoryManager:       memoryManager,
        config:              config,
    }
}

// AnalyzeRootCause 执行根因分析
func (o *Orchestrator) AnalyzeRootCause(ctx context.Context, req *types.AnalysisRequest) (*types.AnalysisResult, error) {
    // 1. 加载上下文
    history, err := o.memoryManager.GetConversationHistory(ctx, req.SessionID)
    if err != nil {
        history = []memory.ChatMessage{}
    }

    similarCases, err := o.memoryManager.FindSimilarCases(ctx, req.Context)
    if err != nil {
        similarCases = []types.CaseStudy{}
    }

    // 2. 执行分析链
    analysisInput := chains.AnalysisInput{
        Request:      req,
        History:      history,
        SimilarCases: similarCases,
    }

    analysisOutput, err := o.analysisChain.Run(ctx, analysisInput)
    if err != nil {
        return o.buildErrorResult(req.RequestID, err), nil
    }

    // 3. 如果需要更多信息,使用 Agent
    if analysisOutput.NeedsMoreInfo {
        toolResult, err := o.toolAgent.Execute(ctx, analysisOutput.ToolRequest)
        if err == nil {
            analysisOutput = o.mergeToolResults(analysisOutput, toolResult)
        }
    }

    // 4. 生成建议
    recommendationInput := chains.RecommendationInput{
        RootCause: analysisOutput.RootCause,
        Context:   req.Context,
        History:   history,
    }

    recommendations, err := o.recommendationChain.Run(ctx, recommendationInput)
    if err != nil {
        // 降级到规则引擎
        recommendations = o.getRuleBasedRecommendations(analysisOutput.RootCause)
    }

    // 5. 构建最终结果
    result := &types.AnalysisResult{
        RequestID:      req.RequestID,
        Status:         "completed",
        Result: &types.DetailedResult{
            RootCause:       analysisOutput.RootCause,
            Recommendations: recommendations,
            Confidence:      analysisOutput.Confidence,
            Evidence:        analysisOutput.Evidence,
            SimilarCases:    similarCases,
            LLMAnalysis:     analysisOutput.LLMAnalysis,
        },
        ProcessingTime: time.Since(start).Seconds(),
    }

    // 6. 保存到记忆
    o.memoryManager.SaveAnalysis(ctx, result)

    return result, nil
}

// 辅助方法
func (o *Orchestrator) mergeToolResults(analysis *chains.AnalysisOutput, toolResult *agents.ExecutionResult) *chains.AnalysisOutput
func (o *Orchestrator) getRuleBasedRecommendations(rootCause *types.RootCause) []types.Recommendation
func (o *Orchestrator) buildErrorResult(requestID string, err error) *types.AnalysisResult
```

### 3.5 依赖库选择

#### 3.5.1 LLM Proxy 库

**选择**: `github.com/teilomillet/gollm`

**理由**:

- ✅ 统一的 API 接口支持多个提供商 (OpenAI, Anthropic, Groq, Ollama)
- ✅ 自动故障转移和重试逻辑
- ✅ 内置的 Prompt 管理
- ✅ 结构化输出和验证
- ✅ 活跃维护,文档完善
- ✅ 纯 Go 实现,无 CGO 依赖

**替代方案**:

- `github.com/ScalePortal/llm-router`: 更适合作为独立服务部署,非库集成
- `github.com/kaiban-ai/kaiban-llm-proxy`: 功能较少,主要用于前端 API 隐藏

#### 3.5.2 LangChain 库

**选择**: `github.com/tmc/langchaingo`

**理由**:

- ✅ 官方 LangChain Go 实现
- ✅ 完整的 Chains、Agents、Memory 抽象
- ✅ 丰富的工具集成
- ✅ 向量存储支持 (Pinecone, Chroma, Weaviate)
- ✅ 活跃的社区和文档
- ✅ 支持流式输出和回调

**核心模块**:

- `github.com/tmc/langchaingo/chains` - 链式调用
- `github.com/tmc/langchaingo/agents` - Agent 框架
- `github.com/tmc/langchaingo/memory` - 记忆管理
- `github.com/tmc/langchaingo/llms` - LLM 抽象
- `github.com/tmc/langchaingo/vectorstores` - 向量存储
- `github.com/tmc/langchaingo/embeddings` - Embedding 生成

#### 3.5.3 向量数据库

**选择**: `github.com/chroma-core/chroma-go` (Chroma)

**理由**:

- ✅ 轻量级,可嵌入或独立部署
- ✅ LangChainGo 原生支持
- ✅ 易于开发和测试
- ✅ 支持持久化和内存模式

**替代方案**:

- `github.com/pinecone-io/go-pinecone`: 云服务,需要额外成本
- `github.com/weaviate/weaviate-go-client`: 更重量级,适合大规模

### 3.6 项目目录结构

```
reasoning-service-go/
├── cmd/
│   └── server/
│       └── main.go                    # 入口文件 (修改)
├── internal/
│   ├── api/
│   │   └── server.go                  # API 服务器 (修改)
│   ├── orchestrator/                  # 新增
│   │   └── orchestrator.go            # 协调器
│   ├── chains/                        # 新增
│   │   ├── analysis_chain.go          # 分析链
│   │   ├── recommendation_chain.go    # 建议链
│   │   ├── preprocess_chain.go        # 预处理
│   │   ├── aggregation_chain.go       # 聚合
│   │   └── prompts/                   # Prompt 模板
│   │       ├── analysis_prompts.go
│   │       └── recommendation_prompts.go
│   ├── agents/                        # 新增
│   │   ├── k8s_tool_agent.go          # K8s 工具 Agent
│   │   └── tools/
│   │       ├── kubectl_describe.go
│   │       ├── kubectl_logs.go
│   │       ├── kubectl_top.go
│   │       └── metrics_query.go
│   ├── memory/                        # 新增
│   │   ├── manager.go                 # 记忆管理器
│   │   ├── structured_store.go        # 结构化存储
│   │   └── embeddings.go              # Embedding 生成
│   ├── analyzer/                      # 保留 (重构)
│   │   ├── root_cause.go              # 简化为规则引擎
│   │   └── patterns.go                # 模式定义
│   ├── recommender/                   # 保留 (重构)
│   │   ├── engine.go                  # 规则引擎
│   │   └── rules.go                   # 规则定义
│   └── config/
│       └── config.go                  # 配置 (扩展)
├── pkg/
│   ├── llm/
│   │   ├── proxy/                     # 新增
│   │   │   ├── adapter.go             # gollm 适配器
│   │   │   ├── config.go              # 配置
│   │   │   └── metrics.go             # 指标收集
│   │   ├── interface.go               # 保留但简化
│   │   └── factory.go                 # 迁移到 proxy/adapter
│   └── types/
│       └── types.go                   # 类型定义 (扩展)
├── configs/
│   └── config.yaml                    # 配置文件 (更新)
├── docs/
│   ├── architecture.md                # 架构文档
│   ├── api.md                         # API 文档
│   └── refactoring-architecture-design.md  # 本文档
├── examples/
│   ├── basic_analysis.go              # 基础分析示例
│   ├── agent_execution.go             # Agent 执行示例
│   └── memory_usage.go                # Memory 使用示例
├── tests/
│   ├── integration/
│   │   ├── analysis_test.go
│   │   └── recommendation_test.go
│   └── unit/
│       ├── chains_test.go
│       ├── agents_test.go
│       └── memory_test.go
├── go.mod                             # 依赖 (更新)
├── go.sum
├── Makefile                           # 构建脚本 (更新)
└── README.md                          # 说明文档 (更新)
```

## 4. 迁移策略

### 4.1 分阶段实施

#### Phase 1: 基础设施 (Week 1-2)

**目标**: 建立新的基础架构层

**任务**:

1. 添加依赖库
   - `go get github.com/teilomillet/gollm`
   - `go get github.com/tmc/langchaingo`
   - `go get github.com/chroma-core/chroma-go`

2. 实现 LLM Proxy Adapter
   - `pkg/llm/proxy/adapter.go`
   - `pkg/llm/proxy/config.go`
   - `pkg/llm/proxy/metrics.go`

3. 设置测试环境
   - 单元测试框架
   - Mock LLM 响应
   - 集成测试环境

**验证标准**:

- ✅ gollm 客户端可以成功调用 OpenAI/Gemini
- ✅ 自动故障转移工作正常
- ✅ 成本跟踪和指标收集正常

#### Phase 2: Chains 实现 (Week 3-4)

**目标**: 实现核心分析和建议链

**任务**:

1. 实现 AnalysisChain
   - PreprocessChain
   - RuleBasedAnalysis (从现有代码迁移)
   - LLMAnalysisChain (使用 Proxy Adapter)
   - AggregationChain

2. 实现 RecommendationChain
   - LoadRulesChain (从现有 recommender 迁移)
   - LLMEnhancementChain
   - RiskAssessmentChain
   - PrioritizationChain

3. Prompt Engineering
   - 设计分析 Prompt 模板
   - 设计建议 Prompt 模板
   - A/B 测试框架

**验证标准**:

- ✅ AnalysisChain 可以处理各种输入
- ✅ RecommendationChain 生成高质量建议
- ✅ 链式调用的延迟在可接受范围内

#### Phase 3: Agent & Tools (Week 5-6)

**目标**: 实现工具调用能力

**任务**:

1. 实现 K8s Tools
   - KubectlDescribeTool
   - KubectlLogsTool
   - KubectlTopTool
   - MetricsQueryTool

2. 实现 K8sToolAgent
   - 集成 langchaingo agents
   - 实现 Executor
   - 工具选择逻辑

3. 集成到 Orchestrator

**验证标准**:

- ✅ Agent 可以根据需求调用正确的工具
- ✅ 工具返回有效的 K8s 信息
- ✅ Agent 的推理过程可追踪

#### Phase 4: Memory System (Week 7-8)

**目标**: 实现记忆和上下文管理

**任务**:

1. 设置 Chroma 向量数据库
   - 本地嵌入式模式
   - 持久化配置

2. 实现 MemoryManager
   - ConversationMemory 集成
   - VectorStore 集成
   - StructuredMemory 实现

3. 实现相似案例检索
   - Embedding 生成
   - 相似度计算
   - 结果排序

**验证标准**:

- ✅ 可以保存和检索对话历史
- ✅ 相似案例检索准确率 > 80%
- ✅ 向量搜索延迟 < 100ms

#### Phase 5: Orchestrator & Integration (Week 9-10)

**目标**: 集成所有模块

**任务**:

1. 实现 Orchestrator
   - 协调 Chains, Agents, Memory
   - 错误处理和降级
   - 性能优化

2. 更新 API Server
   - 使用 Orchestrator 替换现有逻辑
   - 保持 API 兼容性
   - 添加新的功能端点

3. 端到端测试
   - 各种分析场景
   - 压力测试
   - 性能基准测试

**验证标准**:

- ✅ 所有现有 API 测试通过
- ✅ 新功能正常工作
- ✅ 性能不低于现有实现

#### Phase 6: 清理与优化 (Week 11-12)

**目标**: 清理旧代码,优化性能

**任务**:

1. 删除旧的 LLM 客户端实现
   - `pkg/llm/openai.go`
   - `pkg/llm/gemini.go`
   - 等

2. 代码优化
   - 减少不必要的 LLM 调用
   - 缓存常用结果
   - 并发优化

3. 文档更新
   - API 文档
   - 架构文档
   - 示例代码

4. 部署准备
   - Docker 镜像更新
   - K8s 配置更新
   - 监控和告警

**验证标准**:

- ✅ 代码覆盖率 > 80%
- ✅ 文档完整且最新
- ✅ 生产环境就绪

### 4.2 向后兼容性

#### 4.2.1 API 兼容性

保持现有 API 端点不变:

- `POST /api/v1/analyze/root-cause`
- `POST /api/v1/analyze/k8s-event`
- `GET /health`

响应格式保持一致,但可以添加新字段:

```json
{
  "request_id": "xxx",
  "status": "completed",
  "result": {
    "root_cause": {...},
    "recommendations": [...],
    "confidence": 0.85,
    "evidence": [...],
    "similar_cases": [...],        // 新增
    "llm_provider": "openai",      // 新增
    "processing_details": {...}    // 新增
  }
}
```

#### 4.2.2 配置兼容性

新的配置向后兼容,旧配置文件仍然有效:

```yaml
# 旧配置继续工作
llm:
  enabled: true
  providers:
    - name: openai
      api_key: ${OPENAI_API_KEY}
      model: gpt-4

# 新配置选项 (可选)
llm_proxy:
  enable_fallback: true
  fallback_providers: [gemini, deepseek]
  enable_cost_tracking: true

memory:
  enable_vector_store: true
  vector_store_type: chroma
  vector_store_path: ./data/chroma
```

### 4.3 回滚策略

#### 4.3.1 功能开关

使用配置开关控制新旧实现:

```yaml
features:
  use_new_orchestrator: true    # 控制是否使用新的 Orchestrator
  use_llm_proxy: true            # 控制是否使用 LLM Proxy
  use_memory_system: true        # 控制是否使用 Memory 系统
  use_tool_agent: false          # 控制是否使用 Tool Agent
```

#### 4.3.2 A/B 测试

```go
// internal/api/server.go
func (s *Server) handleRootCauseAnalysis(w http.ResponseWriter, r *http.Request) {
    if s.config.Features.UseNewOrchestrator {
        // 新实现
        result, err := s.orchestrator.AnalyzeRootCause(ctx, req)
    } else {
        // 旧实现
        result, err := s.analyzer.Analyze(ctx, req)
    }
}
```

#### 4.3.3 监控和报警

- 错误率监控: 新实现错误率不应高于旧实现
- 延迟监控: P99 延迟不应增加超过 20%
- 成本监控: LLM 调用成本不应显著增加

## 5. 风险评估

### 5.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| LangChainGo 库不成熟 | 高 | 中 | 提前进行 POC,准备自定义实现 |
| gollm 性能不足 | 中 | 低 | 性能基准测试,必要时优化或替换 |
| 向量数据库成本 | 中 | 中 | 使用本地 Chroma,控制索引大小 |
| LLM 调用成本增加 | 高 | 中 | 智能缓存,减少不必要调用 |
| 复杂度增加导致 Bug | 高 | 高 | 充分的单元和集成测试 |

### 5.2 时间风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 学习曲线陡峭 | 中 | 高 | 提前学习和 POC |
| 集成问题 | 高 | 中 | 分阶段实施,早期发现问题 |
| 测试时间不足 | 高 | 中 | 并行开发和测试 |

### 5.3 运维风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 部署失败 | 高 | 低 | 蓝绿部署,快速回滚 |
| 性能下降 | 高 | 中 | 性能监控,灰度发布 |
| 依赖服务故障 | 中 | 低 | 降级策略,回退到规则引擎 |

## 6. 成功标准

### 6.1 功能指标

- ✅ 所有现有功能正常工作
- ✅ 新增功能 (Memory, Agent, Tools) 可用
- ✅ 代码覆盖率 ≥ 80%
- ✅ API 兼容性 100%

### 6.2 性能指标

- ✅ P50 延迟 ≤ 当前实现
- ✅ P99 延迟 ≤ 当前实现 * 1.2
- ✅ LLM 调用成功率 ≥ 95%
- ✅ 故障转移时间 < 1s

### 6.3 质量指标

- ✅ 根因分析准确率 ≥ 85%
- ✅ 建议有效性评分 ≥ 4.0/5.0
- ✅ 相似案例检索准确率 ≥ 80%
- ✅ 生产环境稳定性 ≥ 99.9%

### 6.4 成本指标

- ✅ LLM API 调用成本增长 < 30%
- ✅ 计算资源消耗增长 < 20%
- ✅ 存储成本增长 < 10%

## 7. 附录

### 7.1 参考资源

#### LangChainGo 文档

- 官方文档: <https://tmc.github.io/langchaingo/docs/>
- GitHub: <https://github.com/tmc/langchaingo>
- 示例代码: <https://pkg.go.dev/github.com/tmc/langchaingo/examples>

#### gollm 文档

- GitHub: <https://github.com/teilomillet/gollm>
- API 文档: <https://pkg.go.dev/github.com/teilomillet/gollm>

#### 其他资源

- LangChain 概念: <https://docs.langchain.com/docs/>
- LLM Router: <https://github.com/ScalePortal/llm-router>
- Chroma: <https://docs.trychroma.com/>

### 7.2 术语表

- **Chain**: 链式调用,将多个处理步骤串联
- **Agent**: 具有决策能力的智能体,可以调用工具
- **Memory**: 记忆系统,保存历史对话和分析结果
- **Tool**: 外部工具接口 (如 kubectl, API 调用)
- **Vector Store**: 向量数据库,用于相似性搜索
- **Embedding**: 文本的向量表示
- **Fallback**: 故障转移,主服务失败时切换到备用服务
- **Orchestrator**: 协调器,管理多个模块的交互

### 7.3 代码示例

#### 基础 LLM 调用

```go
package main

import (
    "context"
    "fmt"
    "github.com/teilomillet/gollm"
)

func main() {
    // 创建 LLM 客户端
    llm, err := gollm.NewLLM(
        gollm.SetProvider("openai"),
        gollm.SetModel("gpt-4"),
        gollm.SetMaxTokens(1000),
    )
    if err != nil {
        panic(err)
    }

    // 调用 LLM
    ctx := context.Background()
    response, err := llm.Generate(ctx, gollm.NewPrompt("Analyze this K8s event..."))
    if err != nil {
        panic(err)
    }

    fmt.Println(response)
}
```

#### 使用 LangChain 创建简单的 Chain

```go
package main

import (
    "context"
    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    // 创建 LLM
    llm, err := openai.New()
    if err != nil {
        panic(err)
    }

    // 创建 LLM Chain
    llmChain := chains.NewLLMChain(llm, prompts.NewPromptTemplate(
        "Analyze the following K8s event: {{.event}}",
        []string{"event"},
    ))

    // 运行 Chain
    ctx := context.Background()
    result, err := llmChain.Call(ctx, map[string]any{
        "event": eventData,
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
}
```

#### 使用 Agent 和 Tools

```go
package main

import (
    "context"
    "github.com/tmc/langchaingo/agents"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/tools"
)

func main() {
    // 创建 LLM
    llm, err := openai.New()
    if err != nil {
        panic(err)
    }

    // 定义工具
    kubectlTool := tools.Tool{
        Name: "kubectl",
        Description: "Execute kubectl commands",
        Func: func(ctx context.Context, input string) (string, error) {
            // 执行 kubectl 命令
            return executeKubectl(input)
        },
    }

    // 创建 Agent
    agent := agents.NewZeroShotAgent(
        llm,
        []tools.Tool{kubectlTool},
    )

    executor := agents.NewExecutor(agent)

    // 执行任务
    ctx := context.Background()
    result, err := executor.Call(ctx, map[string]any{
        "input": "Get details about the failing pod",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
}
```

## 8. 总结

本重构方案通过引入 go-llm-proxy (gollm) 和 LangChainGo,实现了:

1. **统一的 LLM 接口**: 简化多提供商集成,自动故障转移
2. **可组合的推理链**: 标准化分析流程,便于扩展和优化
3. **智能 Agent 系统**: 自动调用外部工具,增强分析能力
4. **持久化记忆**: 利用历史数据和相似案例,提高准确率

通过分阶段实施和充分的测试,可以在保持系统稳定性的同时,实现架构的现代化升级。
