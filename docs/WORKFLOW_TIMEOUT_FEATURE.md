# Workflow Timeout Handling Feature

## Overview

This document describes the workflow timeout handling mechanism implemented in the orchestrator service. The feature provides comprehensive timeout control at both the workflow and step levels, with automatic cleanup and retry capabilities.

## Features

### 1. Multi-Level Timeout Control

- **Global Workflow Timeout**: Maximum time a complete workflow can run
- **Step-Level Timeout**: Individual timeout for each workflow step
- **Configurable Timeouts**: Both global and step timeouts can be configured
- **Context-Based Cancellation**: Uses Go's context package for proper cancellation propagation

### 2. Timeout Detection and Handling

- **Real-time Monitoring**: Active monitoring of workflow and step execution time
- **Graceful Cancellation**: Proper context cancellation for in-flight operations
- **Status Tracking**: Separate status (`ExecutionStatusTimeout`) for timed-out workflows
- **Metrics Collection**: Track timeout occurrences for monitoring and alerting

### 3. Automatic Cleanup

- **Resource Release**: Cleanup of held resources (database connections, locks, etc.)
- **Step Cancellation**: Mark incomplete steps as cancelled
- **State Persistence**: Save cleanup state to database
- **Error Context**: Preserve timeout information for debugging

### 4. Retry Mechanism

- **Configurable Retry**: Enable/disable retry on timeout
- **Retry Limits**: Configure maximum number of retry attempts
- **Retry Context**: Preserve retry count and previous timeout information
- **Progressive Retry**: Track retry attempts in workflow context

## Configuration

### Configuration File

Add the following section to `configs/orchestrator/config.yaml`:

```yaml
workflow:
  global_timeout: 30m          # Maximum time a workflow can run (default: 30 minutes)
  step_default_timeout: 5m     # Default timeout for workflow steps (default: 5 minutes)
  retry_on_timeout: true       # Whether to retry workflows that timeout
  max_retries: 3               # Maximum number of retry attempts for timed-out workflows
```

### Command-Line Flags

```bash
# Set global workflow timeout
./orchestrator --workflow.global-timeout=45m

# Set default step timeout
./orchestrator --workflow.step-default-timeout=10m

# Enable/disable retry on timeout
./orchestrator --workflow.retry-on-timeout=true

# Set maximum retry attempts
./orchestrator --workflow.max-retries=5
```

### Environment Variables

```bash
export WORKFLOW_GLOBAL_TIMEOUT=30m
export WORKFLOW_STEP_DEFAULT_TIMEOUT=5m
export WORKFLOW_RETRY_ON_TIMEOUT=true
export WORKFLOW_MAX_RETRIES=3
```

## Implementation Details

### Code Structure

#### 1. Configuration Options (`cmd/orchestrator/app/options/options.go`)

```go
type WorkflowOptions struct {
    GlobalTimeout      time.Duration // Maximum workflow execution time
    StepDefaultTimeout time.Duration // Default timeout for steps
    RetryOnTimeout     bool          // Enable retry on timeout
    MaxRetries         int           // Maximum retry attempts
}
```

#### 2. Workflow Engine (`internal/orchestrator/workflow/engine.go`)

**Key Components**:

- `cancelFuncs map[string]context.CancelFunc`: Store cancel functions for running workflows
- `executionsTimedOut int64`: Metric tracking timeout occurrences
- `SetTimeoutConfig()`: Configure timeout settings
- `handleWorkflowTimeout()`: Handle workflow timeout events
- `cleanupExecution()`: Perform cleanup operations
- `CancelExecutionWithCleanup()`: Manual cancellation with cleanup

**Workflow Execution Flow**:

```go
// 1. Create context with timeout
workflowCtx, cancel := context.WithTimeout(context.Background(), workflowTimeout)

// 2. Store cancel function for potential cancellation
e.cancelFuncs[executionID] = cancel

// 3. Execute workflow asynchronously
go e.executeWorkflow(workflowCtx, workflow, execution)

// 4. Monitor for timeout in execution loop
select {
case <-ctx.Done():
    e.handleWorkflowTimeout(ctx, execution)
    return
default:
    // Continue execution
}
```

#### 3. Step Execution (`internal/orchestrator/workflow/engine.go`)

```go
// executeStepWithTimeout wraps step execution with timeout
func (e *Engine) executeStepWithTimeout(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (*types.StepExecution, error) {
    // Use step-specific timeout or default
    stepTimeout := e.stepDefaultTimeout
    if step.Timeout > 0 {
        stepTimeout = step.Timeout
    }

    // Create context with timeout
    stepCtx, cancel := context.WithTimeout(ctx, stepTimeout)
    defer cancel()

    // Execute step
    return e.executeStep(stepCtx, execution, step)
}
```

### Timeout Handling Process

#### 1. Workflow Timeout Detection

```go
// In executeWorkflow loop
select {
case <-ctx.Done():
    // Context cancelled or timed out
    e.handleWorkflowTimeout(ctx, execution)
    return
default:
    // Continue execution
}
```

#### 2. Timeout Handling

```go
func (e *Engine) handleWorkflowTimeout(ctx context.Context, execution *types.WorkflowExecution) {
    // 1. Log timeout event
    e.logger.Warn("Workflow execution timed out", ...)

    // 2. Perform cleanup
    e.cleanupExecution(ctx, execution)

    // 3. Update metrics
    e.executionsTimedOut++

    // 4. Check for retry
    if e.retryOnTimeout && retryCount < e.maxRetries {
        // Schedule retry (external scheduler)
        execution.Context["retry_count"] = retryCount + 1
    }

    // 5. Complete execution with timeout status
    e.completeExecution(ctx, execution, types.ExecutionStatusTimeout, ...)
}
```

#### 3. Cleanup Operations

```go
func (e *Engine) cleanupExecution(ctx context.Context, execution *types.WorkflowExecution) {
    // 1. Release resources (locks, connections, etc.)

    // 2. Cancel in-flight HTTP requests (via context)

    // 3. Mark incomplete steps as cancelled
    for i := range execution.StepExecutions {
        if execution.StepExecutions[i].Status == types.ExecutionStatusRunning {
            execution.StepExecutions[i].Status = types.ExecutionStatusCancelled
            execution.StepExecutions[i].Error = "Cancelled due to workflow timeout"
        }
    }

    // 4. Save cleanup state to database
    e.store.SaveWorkflowExecution(context.Background(), execution)
}
```

## Usage Examples

### Example 1: Basic Workflow with Timeout

```yaml
workflow:
  name: "diagnose_pod_crashloop"
  timeout: 15m  # Override global timeout

  steps:
    - id: collect_logs
      type: command
      timeout: 2m  # Step-specific timeout
      command:
        tool: kubectl
        action: logs
        args: ["--tail=100", "${pod_name}"]

    - id: ai_analysis
      type: ai_analysis
      timeout: 5m  # AI analysis may take longer
      analysis_type: "root_cause"
```

### Example 2: Monitoring Timeout Metrics

```go
// Get engine statistics including timeout metrics
stats := engine.GetStatistics()

fmt.Printf("Executions timed out: %d\n", stats["executions_timed_out"])
fmt.Printf("Global timeout: %s\n", stats["global_timeout"])
fmt.Printf("Step default timeout: %s\n", stats["step_default_timeout"])
```

### Example 3: Manual Cancellation with Cleanup

```go
// Cancel a running workflow with cleanup
err := engine.CancelExecutionWithCleanup(ctx, executionID, "Manual cancellation by operator")
if err != nil {
    log.Errorf("Failed to cancel execution: %v", err)
}
```

## Monitoring and Observability

### Metrics

The workflow engine provides the following timeout-related metrics:

- `executions_timed_out`: Total number of workflows that timed out
- `active_executions`: Current number of running workflows
- `global_timeout`: Configured global timeout duration
- `step_default_timeout`: Configured default step timeout

### Logging

Timeout events are logged with the following information:

```json
{
  "level": "warn",
  "msg": "Workflow execution timed out",
  "execution_id": "uuid-1234",
  "duration": "35m12s",
  "workflow_id": "diagnose_pod_crashloop"
}
```

Cleanup operations are logged:

```json
{
  "level": "info",
  "msg": "Performing cleanup for execution",
  "execution_id": "uuid-1234",
  "status": "timeout"
}
```

### Database Records

Timed-out workflows are stored with:

- `status`: `"timeout"`
- `error`: Detailed timeout message with duration
- `context.cleanup_started`: Timestamp of cleanup start
- `context.retry_count`: Number of retry attempts (if applicable)
- `step_executions[].status`: Steps marked as `"cancelled"` if incomplete

## Best Practices

### 1. Timeout Configuration

- **Set realistic timeouts**: Base on observed execution times plus buffer
- **Consider step complexity**: Complex steps (AI analysis) need longer timeouts
- **Account for external dependencies**: Network latency, API rate limits, etc.
- **Progressive timeouts**: Step timeout < Global timeout

### 2. Retry Strategy

- **Enable retry for transient failures**: Temporary network issues, service unavailability
- **Disable retry for permanent failures**: Invalid configuration, missing resources
- **Set reasonable retry limits**: Avoid infinite retry loops
- **Monitor retry patterns**: High retry rates indicate systemic issues

### 3. Cleanup Operations

- **Test cleanup logic**: Ensure resources are properly released
- **Make cleanup idempotent**: Safe to run multiple times
- **Log cleanup actions**: Aid in debugging and auditing
- **Handle cleanup failures**: Log but don't fail the timeout handling

### 4. Monitoring

- **Alert on high timeout rates**: May indicate under-provisioned resources
- **Track timeout duration distribution**: Identify outliers
- **Monitor retry effectiveness**: Are retries succeeding?
- **Correlate with external systems**: Database, cache, external APIs

## Troubleshooting

### Issue: Workflows timing out frequently

**Possible Causes**:
- Timeout configured too short
- External service slow or unavailable
- Database performance issues
- High load on orchestrator service

**Solutions**:
1. Check execution duration distribution
2. Increase timeout if executions are close to limit
3. Investigate slow external services
4. Optimize database queries
5. Scale orchestrator service

### Issue: Timeouts not triggering

**Possible Causes**:
- Context not properly propagated
- Blocking operations not respecting context
- Timeout configured as 0 (disabled)

**Solutions**:
1. Verify timeout configuration
2. Check context propagation in custom step executors
3. Ensure all blocking operations use context-aware methods
4. Review logs for timeout configuration at startup

### Issue: Cleanup not executing

**Possible Causes**:
- Database connection issues
- Panic in cleanup code
- Context already cancelled

**Solutions**:
1. Use background context for cleanup operations
2. Add panic recovery in cleanup functions
3. Log all cleanup operations for debugging
4. Test cleanup logic in isolation

## Performance Considerations

### Memory Usage

- **Cancel functions**: One per active workflow (~8 bytes pointer + context overhead)
- **Execution tracking**: One entry per active workflow in memory map
- **Impact**: Minimal for typical workloads (<1000 concurrent workflows)

### CPU Usage

- **Timeout monitoring**: Passive (select statement, no polling)
- **Context cancellation**: O(1) operation
- **Cleanup operations**: Depends on cleanup logic (typically <100ms)

### Database Load

- **Status updates**: One write per timeout event
- **Cleanup saves**: One write per timeout event
- **Query impact**: Minimal (indexed lookups)

## Future Enhancements

### Planned Features

1. **Dynamic timeout adjustment**: Auto-adjust based on historical execution times
2. **Step timeout prediction**: ML-based timeout estimation
3. **Cascading timeouts**: Parent timeout affects child workflows
4. **Timeout policies**: Different policies for different workflow categories
5. **Graceful degradation**: Partial results on timeout
6. **Timeout budgets**: Distributed timeout across workflow DAG

### Potential Improvements

1. **Prometheus metrics integration**: Export timeout metrics
2. **Alerting integration**: Send alerts on timeout thresholds
3. **Timeout analysis dashboard**: Visualize timeout patterns
4. **Automated retry scheduling**: Built-in retry scheduler
5. **Timeout reason classification**: Categorize timeout causes

## References

- [Go Context Package Documentation](https://golang.org/pkg/context/)
- [Workflow Engine Architecture](../architecture/SYSTEM_ARCHITECTURE.md)
- [Orchestrator Service Documentation](../README.md)
- [Configuration Management Guide](../CONFIG.md)

## Version History

- **v1.0.0** (2025-11-06): Initial implementation
  - Global workflow timeout
  - Step-level timeout
  - Automatic cleanup
  - Retry mechanism
  - Configuration options
  - Metrics and logging
