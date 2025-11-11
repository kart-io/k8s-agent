# Agent Framework Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         Agent Framework Extension                                │
│                      (Unified Intelligent Agent System)                          │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Service Layer                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────────────────┐              ┌──────────────────────────┐        │
│  │   Orchestrator Service   │              │   Agent-Manager Service  │        │
│  ├──────────────────────────┤              ├──────────────────────────┤        │
│  │                          │              │                          │        │
│  │  ┌────────────────────┐  │              │  ┌────────────────────┐  │        │
│  │  │  WorkflowAgent     │  │              │  │  CommandAgent      │  │        │
│  │  │  - workflow_exec   │  │              │  │  - dispatch        │  │        │
│  │  │  - step_orchestr   │  │              │  │  - tracking        │  │        │
│  │  │  - context_mgmt    │  │              │  │  - polling         │  │        │
│  │  └────────────────────┘  │              │  └────────────────────┘  │        │
│  │                          │              │                          │        │
│  │  ┌────────────────────┐  │              │  ┌────────────────────┐  │        │
│  │  │  StepAgent         │  │              │  │  KubectlAgent      │  │        │
│  │  │  - command         │  │              │  │  - kubectl_get     │  │        │
│  │  │  - ai_analysis     │  │              │  │  - kubectl_logs    │  │        │
│  │  │  - decision        │  │              │  │  - result_parsing  │  │        │
│  │  │  - remediation     │  │              │  └────────────────────┘  │        │
│  │  │  - notification    │  │              │                          │        │
│  │  └────────────────────┘  │              │                          │        │
│  │                          │              │                          │        │
│  └─────────────┬────────────┘              └─────────────┬────────────┘        │
│                │                                         │                      │
└────────────────┼─────────────────────────────────────────┼──────────────────────┘
                 │                                         │
                 └────────────┬────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         Distributed Agent Layer                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          Coordinator                                       │  │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐          │  │
│  │  │ Load Balancing  │  │    Failover     │  │     Retry       │          │  │
│  │  │  (Round-robin)  │  │   (Auto-switch) │  │  (Exponential)  │          │  │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘          │  │
│  │                                                                            │  │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐          │  │
│  │  │    Parallel     │  │   Sequential    │  │   Distributed   │          │  │
│  │  │   Execution     │  │   Execution     │  │     Tracing     │          │  │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘          │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────┐                      ┌──────────────────────┐        │
│  │      Registry        │                      │       Client         │        │
│  ├──────────────────────┤                      ├──────────────────────┤        │
│  │ - Instance Reg       │                      │ - Sync Call          │        │
│  │ - Health Tracking    │                      │ - Async Call         │        │
│  │ - Service Discovery  │                      │ - Result Polling     │        │
│  │ - Heartbeat          │◄─────────────────────┤ - Health Check       │        │
│  └──────────────────────┘                      └──────────────────────┘        │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Tool Agent Layer                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐              │
│  │ HTTPAgent  │  │ ShellAgent │  │ Database   │  │ CacheAgent │              │
│  ├────────────┤  ├────────────┤  │   Agent    │  ├────────────┤              │
│  │ - GET      │  │ - Execute  │  ├────────────┤  │ - Get      │              │
│  │ - POST     │  │ - Script   │  │ - Query    │  │ - Set      │              │
│  │ - PUT      │  │ - Pipeline │  │ - Create   │  │ - Delete   │              │
│  │ - DELETE   │  │ - Timeout  │  │ - Update   │  │ - Exists   │              │
│  │ - PATCH    │  │ - Whitelist│  │ - Delete   │  │ - Expire   │              │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘              │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        Observability Layer                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                    InstrumentedAgent (Wrapper)                             │  │
│  │                 (Transparent observability for all agents)                 │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌────────────────┐     ┌────────────────┐     ┌────────────────┐            │
│  │    Metrics     │     │    Tracing     │     │    Logging     │            │
│  ├────────────────┤     ├────────────────┤     ├────────────────┤            │
│  │ - Executions   │     │ - Agent Spans  │     │ - Structured   │            │
│  │ - Duration     │     │ - Tool Spans   │     │ - Context      │            │
│  │ - Errors       │     │ - Remote Spans │     │ - Correlation  │            │
│  │ - Tool Calls   │     │ - Events       │     │ - Debug Info   │            │
│  │ - Concurrent   │     │ - Attributes   │     │ - Performance  │            │
│  └────────────────┘     └────────────────┘     └────────────────┘            │
│         │                      │                       │                        │
│         ▼                      ▼                       ▼                        │
│   Prometheus            OpenTelemetry           Logger (Zap/Slog)             │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           Core Agent Interface                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  interface Agent {                                                              │
│      Execute(ctx, input) (output, error)                                       │
│      Name() string                                                              │
│      Description() string                                                       │
│      Capabilities() []string                                                    │
│  }                                                                              │
│                                                                                  │
│  BaseAgent (implements common methods)                                          │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────────────┐
│                         Execution Flow Example                                   │
└─────────────────────────────────────────────────────────────────────────────────┘

  User Request
       │
       ▼
  Orchestrator API
       │
       ▼
  WorkflowAgent.Execute()
       │
       ├──► StepAgent.Execute(Command Step)
       │         │
       │         ▼
       │    Coordinator.ExecuteAgent("agent-manager", "kubectl-agent")
       │         │
       │         ├──► Registry.GetHealthyInstances("agent-manager")
       │         │
       │         ├──► Client.ExecuteAgent(endpoint, "kubectl-agent", input)
       │         │         │
       │         │         ▼
       │         │    Agent-Manager API
       │         │         │
       │         │         ▼
       │         │    KubectlAgent.Execute()
       │         │         │
       │         │         ▼
       │         │    CommandAgent.Execute()
       │         │         │
       │         │         ▼
       │         │    Dispatcher.DispatchCommand()
       │         │         │
       │         │         ▼
       │         │    NATS → Collect Agent → kubectl
       │         │         │
       │         │         ▼
       │         │    Result parsing & enhancement
       │         │
       │         ▼
       │    Output (with parsed kubectl data)
       │
       ├──► StepAgent.Execute(AI Step)
       │         │
       │         ▼
       │    Coordinator.ExecuteAgent("reasoning", "analysis-agent")
       │         │
       │         ▼
       │    AI Analysis Result
       │
       ▼
  Workflow Completed
       │
       ▼
  Metrics, Traces, Logs automatically recorded


┌─────────────────────────────────────────────────────────────────────────────────┐
│                            Data Flow                                             │
└─────────────────────────────────────────────────────────────────────────────────┘

  AgentInput {                          AgentOutput {
    task: "Diagnose pod"                  status: "success"
    instruction: "workflow-123"           result: {...}
    context: {                            reasoning_steps: [...]
      execution: {...}                    tool_calls: [...]
      steps: [...]                        latency: 2.5s
    }                                     timestamp: ...
    options: {...}                        metadata: {...}
  }                                     }
       │                                      ▲
       │                                      │
       └──────────► Agent.Execute() ─────────┘
                         │
                         ├──► Observability Wrapper
                         │         ├──► Metrics
                         │         ├──► Tracing
                         │         └──► Logging
                         │
                         ├──► Business Logic
                         │         ├──► Tool Calls
                         │         └──► Sub-Agents
                         │
                         └──► Result Assembly


┌─────────────────────────────────────────────────────────────────────────────────┐
│                       Benefits Summary                                           │
└─────────────────────────────────────────────────────────────────────────────────┘

  ✅ Unified Interface
     └──► All agents implement same core.Agent interface

  ✅ Distributed Execution
     └──► Cross-service agent calls with failover

  ✅ Automatic Observability
     └──► Metrics, traces, logs without manual instrumentation

  ✅ Tool Reusability
     └──► HTTP, Shell, Database, Cache agents shared across services

  ✅ Composability
     └──► Chain, parallelize, nest agents easily

  ✅ Load Balancing
     └──► Round-robin across service instances

  ✅ Health Management
     └��─► Automatic instance health tracking

  ✅ Context Propagation
     └──► Seamless context passing across agent boundaries

  ✅ Error Handling
     └──► Comprehensive error tracking and recovery

  ✅ Performance
     └──► < 100ms overhead, 1000+ concurrent agents
```
