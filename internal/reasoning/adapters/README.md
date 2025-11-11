# Reasoning Service Refactoring - Agent Framework Integration

## Overview

This directory contains adapters that integrate the `internal/reasoning` service with the generic `pkg/agent` framework. The adapters maintain backward compatibility while enabling the reasoning service to leverage the framework's standardized interfaces and capabilities.

## Architecture

### Framework Components (pkg/agent/core)

- **Agent Interface**: Generic AI agent with Execute() method
- **Chain Interface**: Sequential processing pipeline with Process() method
- **Orchestrator Interface**: Multi-component workflow coordination with Execute() method
- **Tool Interface**: Reusable utility with Execute() method

### Reasoning Service Components (internal/reasoning)

- **ReasoningAgent**: Kubernetes failure analysis agent
- **RootCauseChain**: LLM-based root cause analysis chain
- **DescriptionChain**: Failure description generation chain
- **K8sTool**: Kubernetes cluster interaction tool
- **Orchestrator**: Workflow coordinator for complete analysis flow

## Adapter Design Pattern

The adapters follow a **Wrapper Pattern** to bridge between domain-specific types and framework interfaces:

```
┌─────────────────────────────────────────────────────┐
│         pkg/agent/core Framework Interfaces          │
│  (Agent, Chain, Orchestrator, Tool)                 │
└──────────────────┬──��───────────────────────────────┘
                   │
                   │ implements
                   ▼
┌─────────────────────────────────────────────────────┐
│            Adapter Layer (this directory)            │
│  - ReasoningAgentAdapter                             │
│  - RootCauseChainAdapter                             │
│  - DescriptionChainAdapter                           │
│  - K8sToolAdapter                                    │
│  - OrchestratorAdapter                               │
└──────────────────┬──────────────────────────────────┘
                   │
                   │ wraps & delegates to
                   ▼
┌─────────────────────────────────────────────────────┐
│      Domain-Specific Implementations                 │
│  (internal/reasoning/agents, chains, orchestrator)   │
└─────────────────────────────────────────────────────┘
```

## Adapters

### 1. ReasoningAgentAdapter

**Purpose**: Adapts `reasoning.ReasoningAgent` to `core.Agent` interface.

**Type Conversions**:
- `core.AgentInput` → `reasoning.AnalysisInput`
- `reasoning.AnalysisOutput` → `core.AgentOutput`

**Example Usage**:
```go
// Create original reasoning agent
reasoningAgent, _ := reasoning.NewReasoningAgent(...)

// Wrap with adapter
adapter := adapters.NewReasoningAgentAdapter(reasoningAgent)

// Use through framework interface
input := &core.AgentInput{
    Task: "Analyze pod failure",
    Context: map[string]interface{}{
        "failure_type": "CrashLoopBackOff",
        "resource_type": "pod",
        "resource_name": "my-app-pod",
        "namespace": "production",
    },
}

output, err := adapter.Execute(ctx, input)
```

### 2. RootCauseChainAdapter

**Purpose**: Adapts `root_cause.RootCauseChain` to `core.Chain` interface.

**Type Conversions**:
- `core.ChainInput` → `root_cause.AnalysisInput`
- `root_cause.AnalysisOutput` → `core.ChainOutput`

**Example Usage**:
```go
// Create chain
chain, _ := root_cause.NewRootCauseChain(llmProxy, config)

// Wrap with adapter
adapter := adapters.NewRootCauseChainAdapter(chain)

// Use through framework interface
input := &core.ChainInput{
    Data: &root_cause.AnalysisInput{
        FailureType: "pod_failure",
        ResourceName: "my-pod",
        // ... other fields
    },
}

output, err := adapter.Process(ctx, input)
```

### 3. DescriptionChainAdapter

**Purpose**: Adapts `description.DescriptionChain` to `core.Chain` interface.

**Type Conversions**:
- `core.ChainInput` → `description.DescriptionInput`
- `description.DescriptionOutput` → `core.ChainOutput`

**Example Usage**:
```go
// Create chain
chain, _ := description.NewDescriptionChain(llmProxy, config)

// Wrap with adapter
adapter := adapters.NewDescriptionChainAdapter(chain)

// Use through framework interface
input := &core.ChainInput{
    Data: &description.DescriptionInput{
        FailureType: "pod_failure",
        Language: "en",
        // ... other fields
    },
}

output, err := adapter.Process(ctx, input)
```

### 4. K8sToolAdapter

**Purpose**: Adapts `k8s_tool.K8sTool` to `core.Tool` interface.

**Type Conversions**:
- `core.ToolInput` → `k8s_tool.ToolInput`
- `k8s_tool.ToolOutput` → `core.ToolOutput`

**Example Usage**:
```go
// Create tool
tool, _ := k8s_tool.NewK8sTool(config)

// Wrap with adapter
adapter := adapters.NewK8sToolAdapter(tool)

// Use through framework interface
input := &core.ToolInput{
    Action: "logs",
    Parameters: map[string]interface{}{
        "resource_type": "pod",
        "resource_name": "my-pod",
        "namespace": "default",
    },
}

output, err := adapter.Execute(ctx, input)
```

### 5. OrchestratorAdapter

**Purpose**: Adapts the orchestrator to use `core.BaseOrchestrator` and coordinate all adapted components.

**Key Features**:
- Registers all adapted agents, chains, and tools with `BaseOrchestrator`
- Implements `core.Orchestrator.Execute()` interface
- Maintains original orchestrator workflow logic
- Provides type conversion between framework and domain types

**Example Usage**:
```go
// Create orchestrator adapter
adapter, err := adapters.NewOrchestratorAdapter(
    reasoningAgent,
    rootCauseChain,
    descriptionChain,
    k8sTool,
    memoryManager,
    config,
)

// Use through framework interface
request := &core.OrchestratorRequest{
    TaskID: "task-123",
    TaskType: "failure_analysis",
    Description: "Analyze pod crash",
    Parameters: map[string]interface{}{
        "failure_type": "CrashLoopBackOff",
        "resource_type": "pod",
        "resource_name": "my-pod",
        "namespace": "production",
    },
    Strategy: core.DefaultOrchestratorStrategy(),
    SessionID: "session-456",
}

response, err := adapter.Execute(ctx, request)
```

## Benefits of This Approach

### 1. **Backward Compatibility**
- Existing `internal/reasoning` code remains unchanged
- No breaking changes to current service implementation
- Can migrate gradually without disrupting production

### 2. **Framework Integration**
- Reasoning components now implement standard framework interfaces
- Can be used in other orchestration scenarios
- Enables composition with other framework-compatible agents

### 3. **Type Safety**
- Adapters handle all type conversions explicitly
- Compile-time type checking ensures correctness
- Clear error messages for type mismatches

### 4. **Maintainability**
- Single responsibility: adapters only handle type conversion
- Original business logic unchanged in domain components
- Easy to understand adapter code (each ~200-300 lines)

### 5. **Testability**
- Adapters can be unit tested independently
- Mock framework interfaces for testing
- Domain components maintain existing tests

### 6. **Future Extensibility**
- Easy to add new adapters for new components
- Can swap implementations without changing adapters
- Framework evolves independently of domain code

## Migration Path

### Phase 1: Adapter Creation (Current)
- ✅ Create adapters for all reasoning components
- ✅ Ensure backward compatibility
- ✅ Document adapter usage

### Phase 2: Integration Testing
- Test adapters with existing reasoning service
- Verify all workflows work correctly
- Benchmark performance impact (should be minimal)

### Phase 3: Service Layer Update
- Update `internal/reasoning/service` to use adapters
- Maintain existing API contracts
- Add framework-based endpoints (optional)

### Phase 4: Gradual Migration (Optional)
- Migrate domain types to framework types where beneficial
- Reduce adapter complexity over time
- Consider consolidating similar types

## Performance Considerations

### Adapter Overhead
- Type conversions are lightweight (mostly struct field mapping)
- No deep copying unless necessary
- Expected overhead: < 1ms per operation

### Memory Usage
- Adapters hold references, not copies
- No significant memory overhead
- Garbage collection impact: minimal

## Testing Strategy

### Unit Tests for Adapters
```go
func TestReasoningAgentAdapter_Execute(t *testing.T) {
    // Test input conversion
    // Test output conversion
    // Test error handling
}
```

### Integration Tests
```go
func TestOrchestratorAdapter_FullWorkflow(t *testing.T) {
    // Test complete analysis workflow
    // Verify all steps execute correctly
    // Check result accuracy
}
```

## Future Enhancements

### 1. **Direct Framework Implementation**
Once proven stable, consider implementing domain components directly with framework interfaces:
```go
type ReasoningAgent struct {
    *core.BaseAgent
    // ... fields
}

func (a *ReasoningAgent) Execute(ctx context.Context, input *core.AgentInput) (*core.AgentOutput, error) {
    // Direct implementation
}
```

### 2. **Unified Type System**
Gradually align domain types with framework types where it makes sense:
- Use `core.ReasoningStep` instead of domain-specific step types
- Use `core.ToolCall` for tool invocations
- Reduce type conversion complexity

### 3. **Enhanced Orchestration**
Leverage framework's orchestrator capabilities:
- Retry strategies
- Parallel execution
- Dynamic workflow planning
- Conditional branching

## FAQ

### Q: Why use adapters instead of changing domain code?
**A**: Adapters provide a safe, non-breaking way to integrate with the framework while preserving existing battle-tested code. Migration can happen gradually.

### Q: Is there performance overhead?
**A**: Minimal. Type conversion is lightweight struct field mapping. Expected < 1ms per operation.

### Q: Can I use both adapted and original interfaces?
**A**: Yes! The original interfaces (`reasoning.Agent`, `root_cause.Chain`) remain fully functional. Adapters provide additional framework compatibility.

### Q: When should I use adapters vs. direct framework implementation?
**A**: Use adapters initially for proven production code. Consider direct implementation for new components or when refactoring anyway.

### Q: How do I test adapted components?
**A**: Test adapters separately with unit tests. Integration tests can use either original or adapted interfaces.

## References

- Framework: `/home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/pkg/agent/core/`
- Original implementations: `/home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/internal/reasoning/`
- Adapter pattern: https://refactoring.guru/design-patterns/adapter

## Conclusion

This adapter-based approach provides a pragmatic migration path to the `pkg/agent` framework while maintaining stability and backward compatibility. It enables the reasoning service to benefit from framework capabilities without requiring risky rewrites of production code.
