# Quick Start Guide - Using Reasoning Service with Agent Framework

## 5-Minute Setup

### Step 1: Import the Adapters

```go
import (
    "context"
    "github.com/kart-io/k8s-agent/internal/reasoning/adapters"
    "github.com/kart-io/k8s-agent/pkg/agent/core"
)
```

### Step 2: Create or Use Existing Components

```go
// You already have these from your service initialization
reasoningAgent := yourExistingReasoningAgent
rootCauseChain := yourExistingRootCauseChain
descriptionChain := yourExistingDescriptionChain
k8sTool := yourExistingK8sTool
memoryManager := yourExistingMemoryManager
```

### Step 3: Create the Orchestrator Adapter

```go
orchestrator, err := adapters.NewOrchestratorAdapter(
    reasoningAgent,
    rootCauseChain,
    descriptionChain,
    k8sTool,
    memoryManager,
    config, // your existing OrchestratorConfig
)
if err != nil {
    log.Fatalf("Failed to create orchestrator: %v", err)
}
```

### Step 4: Use Framework Interface

```go
// Build request using framework types
request := &core.OrchestratorRequest{
    TaskID:   "task-001",
    TaskType: "failure_analysis",
    Parameters: map[string]interface{}{
        "failure_type":  "CrashLoopBackOff",
        "resource_type": "pod",
        "resource_name": "my-pod",
        "namespace":     "default",
        "error_message": "OOMKilled",
    },
    SessionID: "session-123",
}

// Execute
response, err := orchestrator.Execute(context.Background(), request)

// Access results
rootCause := response.Metadata["root_cause"]
description := response.Metadata["description"]
```

## Common Use Cases

### Use Case 1: Quick Pod Failure Analysis

```go
request := &core.OrchestratorRequest{
    TaskID:   generateTaskID(),
    TaskType: "pod_failure",
    Parameters: map[string]interface{}{
        "failure_type":  "CrashLoopBackOff",
        "resource_type": "pod",
        "resource_name": podName,
        "namespace":     namespace,
    },
    Strategy: core.DefaultOrchestratorStrategy(),
}

response, _ := orchestrator.Execute(ctx, request)
if rc, ok := response.Metadata["root_cause"].(*root_cause.AnalysisOutput); ok {
    fmt.Printf("Root cause: %s (confidence: %.2f)\n",
        rc.RootCause, rc.Confidence)
}
```

### Use Case 2: Analysis with Similar Cases

```go
request := &core.OrchestratorRequest{
    TaskID:   generateTaskID(),
    TaskType: "failure_analysis",
    Parameters: map[string]interface{}{
        "failure_type":  "ServiceDown",
        "resource_type": "service",
        "resource_name": serviceName,
        "namespace":     namespace,
        "error_message": errorMsg,
    },
    SessionID: sessionID, // Enable similar case lookup
    Strategy: core.OrchestratorStrategy{
        Mode:          "sequential",
        GlobalTimeout: 2 * time.Minute,
    },
}

response, _ := orchestrator.Execute(ctx, request)
similarCases := response.Metadata["similar_cases"].([]*memory.CaseMemory)
fmt.Printf("Found %d similar cases\n", len(similarCases))
```

### Use Case 3: Direct Agent Usage

```go
// Use adapter for single agent
agentAdapter := adapters.NewReasoningAgentAdapter(reasoningAgent)

input := &core.AgentInput{
    Task: "Analyze failure",
    Context: map[string]interface{}{
        "failure_type":  "OOMKilled",
        "resource_name": "my-pod",
        // ...
    },
}

output, _ := agentAdapter.Execute(ctx, input)
analysisResult := output.Result.(*reasoning.AnalysisOutput)
```

## Migration Checklist

- [ ] Import adapters package
- [ ] Wrap existing components with adapters
- [ ] Test with framework types
- [ ] Verify results match original implementation
- [ ] Update API endpoints (optional)
- [ ] Deploy to staging
- [ ] Monitor performance

## Troubleshooting

### Issue: Type assertion fails
```go
// Wrong
rootCause := response.Metadata["root_cause"].(string)

// Correct
if rc, ok := response.Metadata["root_cause"].(*root_cause.AnalysisOutput); ok {
    // Use rc
}
```

### Issue: Missing parameters
```go
// Make sure all required parameters are provided
request.Parameters = map[string]interface{}{
    "failure_type":  "required",
    "resource_type": "required",
    "resource_name": "required",
    "namespace":     "required",
    // Optional ones can be omitted
}
```

### Issue: Timeout errors
```go
// Set appropriate timeouts
request.Strategy = core.OrchestratorStrategy{
    GlobalTimeout: 5 * time.Minute,  // Overall
    StepTimeout:   60 * time.Second,  // Per step
}
```

## Performance Tips

1. **Reuse adapters**: Create once, use multiple times
2. **Set reasonable timeouts**: Balance speed vs. accuracy
3. **Use context cancellation**: Enable early termination
4. **Monitor step execution**: Check which steps take longest

## Next Steps

- Read [README.md](./README.md) for comprehensive documentation
- Check [example/main.go](./example/main.go) for more examples
- Review [REFACTORING_SUMMARY.md](./REFACTORING_SUMMARY.md) for technical details

## Need Help?

- Architecture questions → See README.md
- Usage examples → See example/main.go
- Technical details → See REFACTORING_SUMMARY.md
- Migration guide → See COMPLETION_REPORT.md
