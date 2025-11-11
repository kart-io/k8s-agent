# Generic AI Agent Framework

A **generic, reusable AI Agent Framework** that provides core capabilities for building intelligent agents. **Completely independent of Kubernetes**, suitable for use in any Go project.

## 概述

`pkg/agent` 是一个通用的 AI Agent 框架，从 Aetherius Reasoning Service 中提取并抽象化。它提供了构建智能代理系统所需的核心组件：

- **Agent**: 具有推理能力的智能体
- **Chain**: 链式处理模式，用于多步骤任务
- **Orchestrator**: 编排器，协调多个 Agent、Chain 和 Tool
- **Memory**: 记忆系统，支持对话历史和案例检索
- **LLM**: 大语言模型抽象层
- **Utils**: 实用工具（Prompt 构建、响应解析）

## 架构

```
pkg/agent/
├── core/            # 核心接口和基础类型
│   ├── agent.go          # Agent 接口
│   ├── chain.go          # Chain 接口
│   ├── orchestrator.go   # Orchestrator 接口
│   └── errors.go         # 错误定义
├── memory/          # 记忆管理
│   ├── manager.go        # Memory Manager 接口
│   └── inmemory.go       # 内存实现
├── llm/             # LLM 抽象层
│   └── client.go         # LLM Client 接口
└── utils/           # 工具函数
    ├── prompt.go         # Prompt 构建工具
    └── parser.go         # 响应解析工具
```

## 核心概念

### 1. Agent

Agent 是具有推理能力的智能体，能够：
- 接收输入并进行处理
- 调用工具获取额外信息
- 使用 LLM 进行推理
- 返回结构化输出

```go
type Agent interface {
    Execute(ctx context.Context, input *AgentInput) (*AgentOutput, error)
    Name() string
    Description() string
    Capabilities() []string
}
```

### 2. Chain

Chain 是串行执行的处理模式，适用于：
- 多步骤的数据处理流程
- 需要按顺序执行的分析任务
- 每个步骤依赖前一步骤的输出

```go
type Chain interface {
    Process(ctx context.Context, input *ChainInput) (*ChainOutput, error)
    Name() string
    Steps() int
}
```

### 3. Orchestrator

Orchestrator 负责协调多个组件的执行：
- 复杂的多步骤工作流
- 多个 Agent 协作场景
- 动态决策和条件分支

```go
type Orchestrator interface {
    Execute(ctx context.Context, request *OrchestratorRequest) (*OrchestratorResponse, error)
    RegisterAgent(name string, agent Agent) error
    RegisterChain(name string, chain Chain) error
    RegisterTool(name string, tool Tool) error
    Name() string
}
```

### 4. Memory

Memory 系统提供：
- 对话历史管理
- 案例记忆和检索
- 向量存储（可选）
- 通用键值存储

```go
type Manager interface {
    AddConversation(ctx context.Context, conv *Conversation) error
    GetConversationHistory(ctx context.Context, sessionID string, limit int) ([]*Conversation, error)
    AddCase(ctx context.Context, caseMemory *Case) error
    SearchSimilarCases(ctx context.Context, query string, limit int) ([]*Case, error)
    Store(ctx context.Context, key string, value interface{}) error
    Retrieve(ctx context.Context, key string) (interface{}, error)
}
```

## 使用示例

### 创建简单 Agent

```go
package main

import (
    "context"
    "log"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
)

// CustomAgent 自定义 Agent
type CustomAgent struct {
    *core.BaseAgent
}

func NewCustomAgent() *CustomAgent {
    return &CustomAgent{
        BaseAgent: core.NewBaseAgent(
            "custom-agent",
            "A custom agent for specific tasks",
            []string{"analysis", "reasoning"},
        ),
    }
}

func (a *CustomAgent) Execute(ctx context.Context, input *core.AgentInput) (*core.AgentOutput, error) {
    // 实现具体的执行逻辑
    output := &core.AgentOutput{
        Result:  "Processing completed",
        Status:  "success",
        Message: "Task executed successfully",
    }

    return output, nil
}

func main() {
    agent := NewCustomAgent()

    input := &core.AgentInput{
        Task: "Analyze system logs",
        Context: map[string]interface{}{
            "log_file": "/var/log/system.log",
        },
        Options: core.DefaultAgentOptions(),
    }

    output, err := agent.Execute(context.Background(), input)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %v", output.Result)
}
```

### 使用 Chain 处理多步骤任务

```go
package main

import (
    "context"
    "log"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
)

// Step 实现
type ValidationStep struct{}

func (s *ValidationStep) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // 验证逻辑
    return input, nil
}

func (s *ValidationStep) Name() string {
    return "validation"
}

func (s *ValidationStep) Description() string {
    return "Validate input data"
}

type ProcessingStep struct{}

func (s *ProcessingStep) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // 处理逻辑
    return "processed: " + input.(string), nil
}

func (s *ProcessingStep) Name() string {
    return "processing"
}

func (s *ProcessingStep) Description() string {
    return "Process data"
}

func main() {
    // 创建步骤
    steps := []core.Step{
        &ValidationStep{},
        &ProcessingStep{},
    }

    // 创建 Chain
    chain := core.NewBaseChain("processing-chain", steps)

    // 执行
    input := &core.ChainInput{
        Data:    "raw data",
        Options: core.DefaultChainOptions(),
    }

    output, err := chain.Process(context.Background(), input)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %v", output.Data)
}
```

### 使用 Memory 系统

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/kart-io/k8s-agent/pkg/agent/memory"
)

func main() {
    // 创建内存管理器
    memMgr := memory.NewInMemoryManager(memory.DefaultConfig())

    // 添加对话
    conv := &memory.Conversation{
        SessionID: "session-1",
        Role:      "user",
        Content:   "How do I fix this error?",
        Timestamp: time.Now(),
    }

    if err := memMgr.AddConversation(context.Background(), conv); err != nil {
        log.Fatal(err)
    }

    // 添加案例
    case1 := &memory.Case{
        Title:       "Database Connection Error",
        Description: "Unable to connect to database",
        Problem:     "Connection timeout",
        Solution:    "Check network settings and credentials",
        Category:    "database",
        Tags:        []string{"database", "connection"},
    }

    if err := memMgr.AddCase(context.Background(), case1); err != nil {
        log.Fatal(err)
    }

    // 搜索相似案例
    cases, err := memMgr.SearchSimilarCases(context.Background(), "database connection", 5)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d similar cases", len(cases))
}
```

### 使用 Prompt Builder

```go
package main

import (
    "log"

    "github.com/kart-io/k8s-agent/pkg/agent/utils"
)

func main() {
    // 构建 Prompt
    prompt := utils.NewPromptBuilder().
        WithSystemPrompt("You are a helpful assistant.").
        WithContext("User is experiencing a system error").
        WithExample("What's the error?", "Connection timeout to database").
        WithTask("Analyze the error and provide solution").
        WithConstraint("Keep response under 200 words").
        WithConstraint("Provide actionable steps").
        WithOutputFormat("JSON format with 'problem', 'solution', 'steps' fields").
        Build()

    log.Println(prompt)
}
```

### 使用 Response Parser

```go
package main

import (
    "log"

    "github.com/kart-io/k8s-agent/pkg/agent/utils"
)

func main() {
    response := `
Here's the analysis:

## Root Cause
Database connection pool exhausted

## Solution
1. Increase connection pool size
2. Implement connection timeout
3. Add connection retry logic

## JSON Output
` + "```json\n" + `{
  "root_cause": "Connection pool exhausted",
  "confidence": 0.95,
  "recommendations": ["Increase pool size", "Add timeouts"]
}
` + "```\n"

    parser := utils.NewResponseParser(response)

    // 提取 JSON
    jsonData, err := parser.ExtractJSON()
    if err != nil {
        log.Fatal(err)
    }
    log.Println("JSON:", jsonData)

    // 提取列表
    items := parser.ExtractList()
    log.Println("Items:", items)

    // 提取章节
    section, _ := parser.ExtractSection("Root Cause")
    log.Println("Root Cause:", section)
}
```

## 设计原则

1. **接口优先**: 定义清晰的接口，支持多种实现
2. **可组合性**: 组件可以灵活组合和扩展
3. **类型安全**: 使用强类型，避免运行时错误
4. **上下文感知**: 所有操作支持 context.Context
5. **可观测性**: 内置步骤跟踪和性能指标
6. **易用性**: 提供默认实现和配置

## 扩展性

框架设计为易于扩展：

1. **自定义 Agent**: 实现 `Agent` 接口
2. **自定义 Chain**: 实现 `Chain` 和 `Step` 接口
3. **自定义 Memory**: 实现 `Manager` 接口
4. **自定义 LLM**: 实现 `Client` 接口
5. **自定义 Tool**: 实现 `Tool` 接口

## 与 Reasoning Service 的关系

- `internal/reasoning/` 包含 K8s 特定的实现
- `pkg/agent/` 提供通用的框架
- Reasoning Service 使用 `pkg/agent/` 作为基础设施

## 未来计划

- [ ] 添加更多 Memory 实现（Redis, Vector DB）
- [ ] 添加更多 LLM Provider
- [ ] 实现并行 Chain 执行
- [ ] 添加 Agent 间通信机制
- [ ] 完善监控和追踪
- [ ] 添加更多工具和中间件

## 贡献

欢迎贡献代码、报告问题或提出建议。请遵循项目的代码规范和最佳实践。
