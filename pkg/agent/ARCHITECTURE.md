# Agent Framework Architecture

## 目录结构

```
pkg/agent/
├── core/                          # 核心接口和基础实现
│   ├── agent.go                   # Agent 接口及 BaseAgent
│   ├── chain.go                   # Chain 接口及 BaseChain
│   ├── orchestrator.go            # Orchestrator 接口及 BaseOrchestrator
│   └── errors.go                  # 错误定义
├── memory/                        # 记忆管理
│   ├── manager.go                 # Manager 接口定义
│   └── inmemory.go                # 内存实现
├── llm/                           # LLM 抽象层
│   └── client.go                  # LLM Client 接口
├── utils/                         # 工具函数
│   ├── prompt.go                  # Prompt 构建工具
│   └── parser.go                  # 响应解析工具
├── example/                       # 示例代码
│   └── main.go                    # 完整示例
├── README.md                      # 使用文档
└── IMPLEMENTATION_SUMMARY.md      # 实现总结
```

## 组件关系图

```
┌─────────────────────────────────────────────────────────────────┐
│                         Orchestrator                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Coordinates execution of Agents, Chains, and Tools      │   │
│  │  - Sequential/Parallel execution                         │   │
│  │  - Retry and error handling                              │   │
│  │  - Execution tracking                                    │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                    │                    │              │
│         ▼                    ▼                    ▼              │
│    ┌────────┐          ┌────────┐          ┌────────┐          │
│    │ Agent  │          │ Chain  │          │  Tool  │          │
│    └────────┘          └────────┘          └────────┘          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                            Agent                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Intelligent entity with reasoning capability            │   │
│  │  - Receives input                                        │   │
│  │  - Uses LLM for reasoning                                │   │
│  │  - Calls Tools for additional info                       │   │
│  │  - Accesses Memory for context                           │   │
│  │  - Returns structured output                             │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                    │                    │              │
│         ▼                    ▼                    ▼              │
│    ┌────────┐          ┌────────┐          ┌────────┐          │
│    │  LLM   │          │  Tool  │          │ Memory │          │
│    └────────┘          └────────┘          └────────┘          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                            Chain                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Sequential processing pattern                           │   │
│  │  - Multiple steps executed in order                      │   │
│  │  - Each step processes previous output                   │   │
│  │  - Error handling and step skipping                      │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                    │                    │              │
│         ▼                    ▼                    ▼              │
│    ┌────────┐          ┌────────┐          ┌────────┐          │
│    │ Step 1 │  ───►    │ Step 2 │  ───►    │ Step 3 │          │
│    └────────┘          └────────┘          └────────┘          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                           Memory                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Memory management system                                │   │
│  │  - Conversation history (short-term)                     │   │
│  │  - Case memory (long-term)                               │   │
│  │  - Vector store (similarity search)                      │   │
│  │  - Key-value store (generic)                             │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                    │                    │              │
│         ▼                    ▼                    ▼              │
│  ┌───────────┐      ┌──────────────┐      ┌──────────┐         │
│  │Conversation│      │VectorStore   │      │ KVStore  │         │
│  │   Store   │      │(Embedding)   │      │          │         │
│  └───────────┘      └──────────────┘      └──────────┘         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                             LLM                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Large Language Model abstraction                        │   │
│  │  - Provider-agnostic interface                           │   │
│  │  - Supports multiple providers                           │   │
│  │  - Unified request/response format                       │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │          │          │          │          │            │
│         ▼          ▼          ▼          ▼          ▼            │
│    ┌───────┐  ┌───────┐  ┌────────┐ ┌──────┐  ┌──────┐        │
│    │OpenAI │  │Gemini │  │DeepSeek│ │Ollama│  │ Kimi │        │
│    └───────┘  └───────┘  └────────┘ └──────┘  └──────┘        │
└─────────────────────────────────────────────────────────────────┘
```

## 数据流图

```
┌──────────────────────────────────────────────────────────────────┐
│                        Typical Workflow                          │
└──────────────────────────────────────────────────────────────────┘

User Request
    │
    ▼
┌────────────────────┐
│  Orchestrator      │  ◄─── OrchestratorRequest
│                    │       - TaskType
│  1. Plan execution │       - Parameters
│  2. Execute steps  │       - Strategy
│  3. Track progress │
└─────┬──────────────┘
      │
      ├─────────────────────────┬───────────────────────┐
      │                         │                       │
      ▼                         ▼                       ▼
┌────────────┐           ┌────────────┐         ┌────────────┐
│   Chain    │           │   Agent    │         │    Tool    │
│            │           │            │         │            │
│ Step1 ──►  │           │ 1. Memory  │         │  Execute   │
│ Step2 ──►  │           │ 2. LLM     │         │  Action    │
│ Step3 ──►  │           │ 3. Tool    │         │            │
└─────┬──────┘           └─────┬──────┘         └─────┬──────┘
      │                        │                       │
      │                        ▼                       │
      │                  ┌────────────┐               │
      │                  │   Memory   │               │
      │                  │            │               │
      │                  │ - History  │               │
      │                  │ - Cases    │               │
      │                  └────────────┘               │
      │                        │                       │
      │                        ▼                       │
      │                  ┌────────────┐               │
      │                  │    LLM     │               │
      │                  │            │               │
      │                  │ - Prompt   │               │
      │                  │ - Complete │               │
      │                  └────────────┘               │
      │                        │                       │
      └────────────────────────┴───────────────────────┘
                               │
                               ▼
                    ┌──────────────────┐
                    │ Orchestrator     │
                    │ Response         │
                    │                  │
                    │ - Result         │
                    │ - Steps          │
                    │ - Latency        │
                    └──────────────────┘
                               │
                               ▼
                         User Response
```

## 接口层次结构

```
┌─────────────────────────────────────────────────────────────────┐
│                       Core Interfaces                            │
└─────────────────────────────────────────────────────────────────┘

Agent Interface
    │
    ├─ Execute(ctx, input) → (output, error)
    ├─ Name() → string
    ├─ Description() → string
    └─ Capabilities() → []string
         │
         └─ Implemented by
              ├─ BaseAgent (default impl)
              └─ CustomAgent (user impl)

Chain Interface
    │
    ├─ Process(ctx, input) → (output, error)
    ├─ Name() → string
    └─ Steps() → int
         │
         └─ Implemented by
              ├─ BaseChain (default impl)
              └─ CustomChain (user impl)

Orchestrator Interface
    │
    ├─ Execute(ctx, request) → (response, error)
    ├─ RegisterAgent(name, agent) → error
    ├─ RegisterChain(name, chain) → error
    ├─ RegisterTool(name, tool) → error
    └─ Name() → string
         │
         └─ Implemented by
              ├─ BaseOrchestrator (default impl)
              └─ CustomOrchestrator (user impl)

Memory Manager Interface
    │
    ├─ AddConversation(ctx, conv) → error
    ├─ GetConversationHistory(ctx, sessionID, limit) → ([]conv, error)
    ├─ AddCase(ctx, case) → error
    ├─ SearchSimilarCases(ctx, query, limit) → ([]case, error)
    ├─ Store(ctx, key, value) → error
    ├─ Retrieve(ctx, key) → (value, error)
    └─ Clear(ctx) → error
         │
         └─ Implemented by
              ├─ InMemoryManager (memory impl)
              ├─ RedisManager (future)
              └─ VectorDBManager (future)

LLM Client Interface
    │
    ├─ Complete(ctx, request) → (response, error)
    ├─ Chat(ctx, messages) → (response, error)
    ├─ Provider() → Provider
    └─ IsAvailable() → bool
         │
         └─ Implemented by
              ├─ OpenAIClient (future)
              ├─ GeminiClient (future)
              └─ CustomClient (user impl)
```

## 使用模式

```
┌─────────────────────────────────────────────────────────────────┐
│                        Usage Patterns                            │
└─────────────────────────────────────────────────────────────────┘

Pattern 1: Simple Agent
    User → Agent.Execute() → Result
         (直接使用，无 Memory/LLM)

Pattern 2: Agent with Memory
    User → Agent.Execute()
              ├─ Memory.GetHistory()
              ├─ Process with context
              └─ Memory.Save()
         → Result

Pattern 3: Agent with LLM
    User → Agent.Execute()
              ├─ Prompt.Build()
              ├─ LLM.Complete()
              └─ Parser.Parse()
         → Result

Pattern 4: Chain Processing
    User → Chain.Process()
              ├─ Step1.Execute()
              ├─ Step2.Execute(step1_output)
              └─ Step3.Execute(step2_output)
         → Result

Pattern 5: Orchestrated Workflow
    User → Orchestrator.Execute()
              ├─ Chain.Process()
              ├─ Agent1.Execute()
              ├─ Tool.Execute()
              └─ Agent2.Execute()
         → Response

Pattern 6: Full Stack
    User → Orchestrator.Execute()
              ├─ Memory.GetHistory()
              ├─ Chain.Process()
              │    └─ Multiple Steps
              ├─ Agent.Execute()
              │    ├─ Memory.GetCases()
              │    ├─ Prompt.Build()
              │    ├─ LLM.Complete()
              │    ├─ Parser.Parse()
              │    └─ Tool.Execute()
              └─ Memory.Save()
         → Response
```

## 扩展点

```
┌─────────────────────────────────────────────────────────────────┐
│                       Extension Points                           │
└─────────────────────────────────────────────────────────────────┘

1. Custom Agent Implementation
   ┌──────────────────────┐
   │ type MyAgent struct  │
   │   *core.BaseAgent    │
   │   myDependencies...  │
   │ }                    │
   │                      │
   │ func (a *MyAgent)    │
   │   Execute(...)       │
   │     // Custom logic  │
   │ }                    │
   └──────────────────────┘

2. Custom Chain Steps
   ┌──────────────────────┐
   │ type MyStep struct { │
   │   config ...         │
   │ }                    │
   │                      │
   │ func (s *MyStep)     │
   │   Execute(...)       │
   │     // Custom logic  │
   │ }                    │
   └──────────────────────┘

3. Custom Memory Backend
   ┌──────────────────────┐
   │ type RedisMemory     │
   │   client *redis.Cli  │
   │ }                    │
   │                      │
   │ func (m *RedisMemory)│
   │   AddConversation()  │
   │     // Redis ops     │
   │ }                    │
   └──────────────────────┘

4. Custom LLM Provider
   ┌──────────────────────┐
   │ type MyLLMClient     │
   │   apiKey string      │
   │   baseURL string     │
   │ }                    │
   │                      │
   │ func (c *MyLLMClient)│
   │   Complete(...)      │
   │     // API calls     │
   │ }                    │
   └──────────────────────┘

5. Custom Tool
   ┌──────────────────────┐
   │ type MyTool struct { │
   │   config ...         │
   │ }                    │
   │                      │
   │ func (t *MyTool)     │
   │   Execute(...)       │
   │     // Tool logic    │
   │ }                    │
   └──────────────────────┘
```
