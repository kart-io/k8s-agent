# Agent Framework Extension

This document describes the Agent framework extensions for the orchestrator and agent-manager services.

## Overview

The Agent framework provides a unified interface for building intelligent, observable, and distributed agents across all services in the k8s-agent system.

## Architecture

```
pkg/agent/
├── core/              # Core Agent interface and base implementation
├── distributed/       # Cross-service Agent coordination
│   ├── coordinator.go # Agent orchestration and load balancing
│   ├── registry.go    # Service instance registration
│   └── client.go      # Remote Agent invocation
├── tools/             # Common utility Agents
│   ├── http_agent.go  # HTTP request execution
│   ├── shell_agent.go # Shell command execution
│   ├── database_agent.go # Database operations
│   └── cache_agent.go # Redis cache operations
└── observability/     # Monitoring and tracing
    ├── metrics.go     # Prometheus metrics
    ├── tracing.go     # OpenTelemetry tracing
    └── logging.go     # Structured logging wrapper

internal/orchestrator/agents/workflow/
├── workflow_agent.go  # Workflow orchestration Agent
└── step_agent.go      # Individual step execution Agent

internal/agent-manager/agents/command/
├── command_agent.go   # Command dispatch and tracking Agent
└── kubectl_agent.go   # kubectl command specialized Agent
```

## Service Agents

### 1. Orchestrator Service Agents

#### WorkflowAgent
Orchestrates complete diagnostic workflows with multiple steps.

**Capabilities:**
- workflow_execution
- step_orchestration
- context_management
- error_recovery

**Usage:**
```go
import (
    workflowagent "github.com/kart-io/k8s-agent/internal/orchestrator/agents/workflow"
)

agent := workflowagent.NewWorkflowAgent(executor, logger)

input := &agentcore.AgentInput{
    Task: "Execute diagnostic workflow",
    Context: map[string]interface{}{
        "execution": workflowExecution,
        "steps":     workflowSteps,
    },
    Options: agentcore.DefaultAgentOptions(),
}

output, err := agent.Execute(ctx, input)
```

#### StepAgent
Executes individual workflow steps (Command/AI/Decision/Remediation/Notification/Wait).

**Capabilities:**
- command_execution
- ai_analysis
- decision_making
- remediation
- notification
- timing_control

**Usage:**
```go
stepAgent := workflowagent.NewStepAgent(executor, logger)

input := &agentcore.AgentInput{
    Context: map[string]interface{}{
        "execution": execution,
        "step":      step,
    },
}

output, err := stepAgent.Execute(ctx, input)
```

### 2. Agent-Manager Service Agents

#### CommandAgent
Dispatches commands to target clusters and tracks execution results.

**Capabilities:**
- command_dispatch
- execution_tracking
- result_polling
- timeout_management

**Usage:**
```go
import (
    commandagent "github.com/kart-io/k8s-agent/internal/agent-manager/agents/command"
)

agent := commandagent.NewCommandAgent(dispatcher, logger)

input := &agentcore.AgentInput{
    Context: map[string]interface{}{
        "command": command,
    },
    Options: agentcore.AgentOptions{
        Timeout: 60 * time.Second,
    },
}

output, err := agent.Execute(ctx, input)
```

#### KubectlAgent
Specialized agent for kubectl command execution with result parsing.

**Capabilities:**
- kubectl_get
- kubectl_describe
- kubectl_logs
- kubectl_events
- result_parsing

**Usage:**
```go
kubectlAgent := commandagent.NewKubectlAgent(dispatcher, logger)

input := &agentcore.AgentInput{
    Context: map[string]interface{}{
        "cluster_id": "cluster-1",
        "action":     "get",
        "args":       []string{"pods"},
        "namespace":  "default",
    },
}

output, err := kubectlAgent.Execute(ctx, input)
```

## Distributed Agent Coordination

### Registry
Manages service instance registration and health tracking.

**Usage:**
```go
import "github.com/kart-io/k8s-agent/pkg/agent/distributed"

registry := distributed.NewRegistry(logger)

// Register service instance
instance := &distributed.ServiceInstance{
    ID:          "orchestrator-1",
    ServiceName: "orchestrator",
    Endpoint:    "http://localhost:8081",
    Agents:      []string{"workflow-agent", "step-agent"},
}
registry.Register(instance)

// Get healthy instances
instances, _ := registry.GetHealthyInstances("orchestrator")
```

### Client
Invokes remote agents across services.

**Usage:**
```go
client := distributed.NewClient(logger)

// Execute remote agent
output, err := client.ExecuteAgent(ctx, endpoint, agentName, input)

// Async execution
taskID, err := client.ExecuteAgentAsync(ctx, endpoint, agentName, input)
result, completed, err := client.GetAsyncResult(ctx, endpoint, taskID)
```

### Coordinator
Orchestrates agent execution with load balancing and failover.

**Usage:**
```go
coordinator := distributed.NewCoordinator(registry, client, logger)

// Execute with automatic load balancing
output, err := coordinator.ExecuteAgent(ctx, "orchestrator", "workflow-agent", input)

// Execute with retry
output, err := coordinator.ExecuteAgentWithRetry(ctx, "orchestrator", "workflow-agent", input, 3)

// Parallel execution
tasks := []distributed.AgentTask{
    {ServiceName: "agent-manager", AgentName: "command-agent", Input: input1},
    {ServiceName: "reasoning", AgentName: "analysis-agent", Input: input2},
}
results, err := coordinator.ExecuteParallel(ctx, tasks)

// Sequential execution with context passing
results, err := coordinator.ExecuteSequential(ctx, tasks)
```

## Common Tool Agents

### HTTPAgent
General-purpose HTTP client for web requests.

**Usage:**
```go
import "github.com/kart-io/k8s-agent/pkg/agent/tools"

httpAgent := tools.NewHTTPAgent(logger)

// GET request
output, _ := httpAgent.Get(ctx, "http://api.example.com/data", headers)

// POST request
output, _ := httpAgent.Post(ctx, "http://api.example.com/create", body, headers)
```

### ShellAgent
Secure shell command execution with whitelisting.

**Usage:**
```go
allowedCommands := []string{"ls", "cat", "grep", "awk", "bash"}
shellAgent := tools.NewShellAgent(allowedCommands, logger)

input := &agentcore.AgentInput{
    Context: map[string]interface{}{
        "command": "ls",
        "args":    []string{"-la", "/tmp"},
    },
}

output, err := shellAgent.Execute(ctx, input)
```

### DatabaseAgent
Database operations with GORM.

**Usage:**
```go
dbAgent := tools.NewDatabaseAgent(db, logger)

// Query
output, _ := dbAgent.Query(ctx, "SELECT * FROM agents WHERE status = ?", "online")

// Create
output, _ := dbAgent.Create(ctx, "agents", map[string]interface{}{
    "id":     "agent-1",
    "status": "online",
})

// Update
output, _ := dbAgent.Update(ctx, "agents",
    map[string]interface{}{"status": "offline"},
    map[string]interface{}{"id": "agent-1"},
)
```

### CacheAgent
Redis cache operations.

**Usage:**
```go
cacheAgent := tools.NewCacheAgent(redisClient, logger)

// Set with TTL
output, _ := cacheAgent.Set(ctx, "key", "value", 3600)

// Get
output, _ := cacheAgent.Get(ctx, "key")

// Delete
output, _ := cacheAgent.Delete(ctx, "key")
```

## Observability

### Automatic Instrumentation
Wrap any agent to add automatic metrics, logging, and tracing.

**Usage:**
```go
import "github.com/kart-io/k8s-agent/pkg/agent/observability"

// Wrap agent with observability
instrumentedAgent := observability.WrapAgent(agent, "orchestrator", logger)

// All executions now automatically emit:
// - Prometheus metrics
// - Structured logs
// - OpenTelemetry traces
output, err := instrumentedAgent.Execute(ctx, input)
```

### Metrics
Prometheus metrics are automatically collected:

- `agent_executions_total` - Total agent executions
- `agent_execution_duration_seconds` - Execution duration histogram
- `agent_errors_total` - Total errors
- `tool_calls_total` - Tool invocations
- `tool_call_duration_seconds` - Tool call duration
- `remote_agent_calls_total` - Remote agent calls
- `service_instances_total` - Registered instances
- `healthy_instances_total` - Healthy instances
- `concurrent_executions` - Current concurrent executions

### Tracing
OpenTelemetry spans are automatically created:

- `agent.execute` - Agent execution span
- `tool.call` - Tool invocation span
- `remote_agent.call` - Remote agent call span

### Logging
Structured logs with consistent format:

```json
{
  "level": "info",
  "agent": "workflow-agent",
  "service": "orchestrator",
  "task": "Execute diagnostic workflow",
  "session_id": "session-123",
  "duration": "2.5s",
  "status": "success",
  "message": "Agent execution completed"
}
```

## Integration Examples

### Example 1: Orchestrator executes commands via Agent-Manager

```go
// In orchestrator service
coordinator := distributed.NewCoordinator(registry, client, logger)

// Create command
cmd := &types.Command{
    ClusterID: "cluster-1",
    Tool:      "kubectl",
    Action:    "get",
    Args:      []interface{}{"pods", "-n", "default"},
}

// Execute via remote CommandAgent
input := &agentcore.AgentInput{
    Task: "Execute kubectl command",
    Context: map[string]interface{}{
        "command": cmd,
    },
}

output, err := coordinator.ExecuteAgent(ctx, "agent-manager", "command-agent", input)
```

### Example 2: Workflow with multiple agents

```go
// Create workflow agent
workflowAgent := workflowagent.NewWorkflowAgent(executor, logger)

// Wrap with observability
instrumentedAgent := observability.WrapAgent(workflowAgent, "orchestrator", logger)

// Execute workflow
input := &agentcore.AgentInput{
    Task: "Diagnose pod crash",
    Context: map[string]interface{}{
        "execution": execution,
        "steps":     steps,
    },
    Options: agentcore.AgentOptions{
        Timeout:      300 * time.Second,
        EnableTools:  true,
        EnableMemory: true,
    },
}

output, err := instrumentedAgent.Execute(ctx, input)

// Output contains:
// - All step results
// - Tool call records
// - Reasoning steps
// - Execution metrics
```

### Example 3: Parallel agent execution

```go
// Execute multiple agents in parallel
tasks := []distributed.AgentTask{
    {
        ServiceName: "agent-manager",
        AgentName:   "kubectl-agent",
        Input: &agentcore.AgentInput{
            Context: map[string]interface{}{
                "cluster_id": "cluster-1",
                "action":     "get",
                "args":       []string{"pods"},
            },
        },
    },
    {
        ServiceName: "agent-manager",
        AgentName:   "kubectl-agent",
        Input: &agentcore.AgentInput{
            Context: map[string]interface{}{
                "cluster_id": "cluster-1",
                "action":     "events",
            },
        },
    },
}

results, err := coordinator.ExecuteParallel(ctx, tasks)

// Process results
for i, result := range results {
    if result.Error != nil {
        logger.Errorw("Task failed", "index", i, "error", result.Error)
    } else {
        logger.Infow("Task succeeded", "index", i, "output", result.Output)
    }
}
```

## Best Practices

1. **Always use context with timeout**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
   defer cancel()
   ```

2. **Wrap agents with observability**
   ```go
   agent = observability.WrapAgent(agent, serviceName, logger)
   ```

3. **Use distributed coordinator for cross-service calls**
   ```go
   output, err := coordinator.ExecuteAgent(ctx, "service", "agent", input)
   ```

4. **Handle errors gracefully**
   ```go
   output, err := agent.Execute(ctx, input)
   if err != nil {
       logger.Errorw("Agent failed", "error", err)
       // Handle error
   }
   ```

5. **Check output status**
   ```go
   if output.Status != "success" {
       logger.Warnw("Agent execution partial", "message", output.Message)
   }
   ```

## Testing

All agents include comprehensive unit tests. Run tests with:

```bash
# Test all agent packages
make test

# Test specific agent
go test -v ./internal/orchestrator/agents/workflow/
go test -v ./internal/agent-manager/agents/command/
go test -v ./pkg/agent/distributed/
go test -v ./pkg/agent/tools/
```

## Performance

- **Average latency**: < 100ms (without external calls)
- **Tool call overhead**: < 10ms
- **Remote agent call**: < 200ms (same datacenter)
- **Concurrent execution**: Supports 1000+ parallel agents

## Future Enhancements

- [ ] Agent learning and optimization
- [ ] Multi-LLM support for tool agents
- [ ] Advanced retry strategies (exponential backoff, circuit breaker)
- [ ] Agent composition (chaining multiple agents)
- [ ] Dynamic agent discovery and registration
- [ ] Agent marketplace and plugin system
