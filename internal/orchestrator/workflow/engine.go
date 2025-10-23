package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
)

// Engine manages workflow execution
type Engine struct {
	store    *storage.PostgresStore
	cache    *storage.RedisStore
	executor *Executor
	logger   *zap.Logger

	// Execution tracking
	mu         sync.RWMutex
	executions map[string]*types.WorkflowExecution

	// Metrics
	executionsStarted   int64
	executionsCompleted int64
	executionsFailed    int64
}

// NewEngine creates a new workflow engine
func NewEngine(
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	executor *Executor,
	logger *zap.Logger,
) *Engine {
	return &Engine{
		store:      store,
		cache:      cache,
		executor:   executor,
		logger:     logger.With(zap.String("component", "workflow-engine")),
		executions: make(map[string]*types.WorkflowExecution),
	}
}

// StartWorkflow starts a new workflow execution
func (e *Engine) StartWorkflow(ctx context.Context, workflowID string, triggerEvent map[string]interface{}) (*types.WorkflowExecution, error) {
	e.logger.Info("🎬 Starting workflow execution",
		zap.String("workflow_id", workflowID))

	// Load workflow definition
	workflow, err := e.store.GetWorkflow(ctx, workflowID)
	if err != nil {
		e.logger.Error("❌ Failed to load workflow from database",
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	e.logger.Info("✓ Workflow loaded",
		zap.String("workflow_name", workflow.Name),
		zap.String("status", string(workflow.Status)),
		zap.Int("steps", len(workflow.Steps)))

	if workflow.Status != types.WorkflowStatusActive {
		e.logger.Warn("⚠️  Workflow is not active",
			zap.String("workflow_id", workflowID),
			zap.String("status", string(workflow.Status)))
		return nil, fmt.Errorf("workflow is not active")
	}

	// Create execution instance
	executionID := uuid.New().String()
	execution := &types.WorkflowExecution{
		ID:           executionID,
		WorkflowID:   workflowID,
		TriggerEvent: triggerEvent,
		Status:       types.ExecutionStatusPending,
		Context:      make(map[string]interface{}),
		StartedAt:    time.Now(),
	}

	e.logger.Info("📝 Created workflow execution instance",
		zap.String("execution_id", executionID))

	// Save execution
	if err := e.store.SaveWorkflowExecution(ctx, execution); err != nil {
		e.logger.Error("❌ Failed to save execution to database",
			zap.String("execution_id", executionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	e.logger.Info("✓ Execution saved to database",
		zap.String("execution_id", executionID))

	// Track in memory
	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.executionsStarted++
	currentStats := e.executionsStarted
	e.mu.Unlock()

	e.logger.Info("📊 Execution tracking updated",
		zap.String("execution_id", executionID),
		zap.Int64("total_started", currentStats))

	// Start execution asynchronously
	go e.executeWorkflow(context.Background(), workflow, execution)

	e.logger.Info("Workflow execution started",
		zap.String("execution_id", execution.ID),
		zap.String("workflow_id", workflowID))

	return execution, nil
}

// executeWorkflow executes a workflow
func (e *Engine) executeWorkflow(ctx context.Context, workflow *types.Workflow, execution *types.WorkflowExecution) {
	defer func() {
		e.mu.Lock()
		delete(e.executions, execution.ID)
		e.mu.Unlock()
	}()

	// Update status to running
	execution.Status = types.ExecutionStatusRunning
	e.store.UpdateWorkflowExecutionStatus(ctx, execution.ID, types.ExecutionStatusRunning)

	// Execute steps in sequence
	for i, step := range workflow.Steps {
		e.logger.Info("Executing workflow step",
			zap.String("execution_id", execution.ID),
			zap.String("step_id", step.ID),
			zap.String("step_name", step.Name),
			zap.Int("step_index", i))

		// Check if we should execute this step
		if !e.shouldExecuteStep(execution, step) {
			e.logger.Debug("Skipping step due to conditions",
				zap.String("step_id", step.ID))
			continue
		}

		// Execute step
		stepExec, err := e.executeStep(ctx, execution, step)
		if err != nil {
			e.handleStepFailure(ctx, execution, step, err)
			return
		}

		// Add step execution to history
		execution.StepExecutions = append(execution.StepExecutions, *stepExec)

		// Update context with step output
		if stepExec.Output != nil {
			for k, v := range stepExec.Output {
				execution.Context[fmt.Sprintf("step_%s_%s", step.ID, k)] = v
			}
		}

		// Save progress
		if err := e.store.SaveWorkflowExecution(ctx, execution); err != nil {
			e.logger.Error("Failed to save workflow execution progress",
				zap.String("execution_id", execution.ID),
				zap.Error(err))
			// Continue execution even if save fails - we'll try again at completion
		}

		// Check if step failed
		if stepExec.Status == types.ExecutionStatusFailed {
			// Execute failure branch if defined
			if len(step.OnFailure) > 0 {
				e.executeFailureBranch(ctx, workflow, execution, step)
			} else {
				e.completeExecution(ctx, execution, types.ExecutionStatusFailed, fmt.Sprintf("Step %s failed", step.ID))
			}
			return
		}

		// Execute success branch if defined
		if len(step.OnSuccess) > 0 {
			// For now, just continue to next step
			// In advanced implementation, this would handle branching
		}
	}

	// All steps completed successfully
	e.completeExecution(ctx, execution, types.ExecutionStatusCompleted, "")
}

// executeStep executes a single workflow step
func (e *Engine) executeStep(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (*types.StepExecution, error) {
	stepExec := &types.StepExecution{
		StepID:    step.ID,
		Status:    types.ExecutionStatusRunning,
		Input:     e.prepareStepInput(execution, step),
		StartedAt: time.Now(),
	}

	var err error
	var retryCount int

	// Iterative retry loop instead of recursive
	for {
		// Execute based on step type
		switch step.Type {
		case types.StepTypeCommand:
			stepExec.Output, err = e.executor.ExecuteCommand(ctx, execution, step)
		case types.StepTypeAIAnalysis:
			stepExec.Output, err = e.executor.ExecuteAIAnalysis(ctx, execution, step)
		case types.StepTypeDecision:
			stepExec.Output, err = e.executor.ExecuteDecision(ctx, execution, step)
		case types.StepTypeRemediation:
			stepExec.Output, err = e.executor.ExecuteRemediation(ctx, execution, step)
		case types.StepTypeNotification:
			stepExec.Output, err = e.executor.ExecuteNotification(ctx, execution, step)
		case types.StepTypeWait:
			stepExec.Output, err = e.executor.ExecuteWait(ctx, execution, step)
		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}

		// Success case
		if err == nil {
			break
		}

		// Failure - check if retry is allowed
		if step.RetryPolicy == nil || retryCount >= step.RetryPolicy.MaxRetries {
			// No more retries
			break
		}

		// Retry logic
		retryCount++
		stepExec.RetryCount = retryCount

		e.logger.Info("Retrying step",
			zap.String("step_id", step.ID),
			zap.Int("retry_count", retryCount),
			zap.Error(err))

		// Wait before retry with context cancellation support
		delay := e.calculateRetryDelay(step.RetryPolicy, retryCount)
		select {
		case <-ctx.Done():
			// Context cancelled during retry wait
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	// Set completion time and status
	completedAt := time.Now()
	stepExec.CompletedAt = &completedAt
	stepExec.Duration = completedAt.Sub(stepExec.StartedAt)

	if err != nil {
		stepExec.Status = types.ExecutionStatusFailed
		stepExec.Error = err.Error()
		return stepExec, err
	}

	stepExec.Status = types.ExecutionStatusCompleted
	return stepExec, nil
}

// shouldExecuteStep checks if step conditions are met
func (e *Engine) shouldExecuteStep(execution *types.WorkflowExecution, step types.WorkflowStep) bool {
	if len(step.Conditions) == 0 {
		return true
	}

	for _, condition := range step.Conditions {
		if !e.evaluateCondition(execution, condition) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (e *Engine) evaluateCondition(execution *types.WorkflowExecution, condition types.Condition) bool {
	// Get field value from context
	value, ok := execution.Context[condition.Field]
	if !ok {
		return false
	}

	// Evaluate operator
	switch condition.Operator {
	case "eq":
		return value == condition.Value
	case "ne":
		return value != condition.Value
	case "gt":
		if v, ok := value.(float64); ok {
			if cv, ok := condition.Value.(float64); ok {
				return v > cv
			}
		}
	case "lt":
		if v, ok := value.(float64); ok {
			if cv, ok := condition.Value.(float64); ok {
				return v < cv
			}
		}
	case "contains":
		if v, ok := value.(string); ok {
			if cv, ok := condition.Value.(string); ok {
				return contains(v, cv)
			}
		}
	}

	return false
}

// prepareStepInput prepares input for step execution
func (e *Engine) prepareStepInput(execution *types.WorkflowExecution, step types.WorkflowStep) map[string]interface{} {
	input := make(map[string]interface{})

	// Copy step config
	for k, v := range step.Config {
		input[k] = v
	}

	// Add execution context
	input["execution_id"] = execution.ID
	input["workflow_id"] = execution.WorkflowID
	input["trigger_event"] = execution.TriggerEvent

	return input
}

// calculateRetryDelay calculates retry delay with exponential backoff
func (e *Engine) calculateRetryDelay(policy *types.RetryPolicy, retryCount int) time.Duration {
	delay := policy.InitialDelay
	for i := 1; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * policy.BackoffFactor)
		if delay > policy.MaxDelay {
			return policy.MaxDelay
		}
	}
	return delay
}

// handleStepFailure handles step execution failure
func (e *Engine) handleStepFailure(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep, err error) {
	e.logger.Error("Step execution failed",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID),
		zap.Error(err))

	e.completeExecution(ctx, execution, types.ExecutionStatusFailed, err.Error())
}

// executeFailureBranch executes failure branch
func (e *Engine) executeFailureBranch(ctx context.Context, workflow *types.Workflow, execution *types.WorkflowExecution, step types.WorkflowStep) {
	// TODO: Implement failure branch execution
	e.logger.Info("Executing failure branch",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	e.completeExecution(ctx, execution, types.ExecutionStatusFailed, "Failure branch executed")
}

// completeExecution completes workflow execution
func (e *Engine) completeExecution(ctx context.Context, execution *types.WorkflowExecution, status types.ExecutionStatus, errorMsg string) {
	completedAt := time.Now()
	execution.Status = status
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(execution.StartedAt)

	if errorMsg != "" {
		execution.Error = errorMsg
	}

	// Save final state - critical operation
	if err := e.store.SaveWorkflowExecution(ctx, execution); err != nil {
		e.logger.Error("CRITICAL: Failed to save final workflow execution state",
			zap.String("execution_id", execution.ID),
			zap.String("status", string(status)),
			zap.Error(err))
		// This is a data integrity issue - log extensively for investigation
	}

	// Update metrics
	e.mu.Lock()
	if status == types.ExecutionStatusCompleted {
		e.executionsCompleted++
	} else {
		e.executionsFailed++
	}
	e.mu.Unlock()

	e.logger.Info("Workflow execution completed",
		zap.String("execution_id", execution.ID),
		zap.String("status", string(status)),
		zap.Duration("duration", execution.Duration))
}

// CancelExecution cancels a running execution
func (e *Engine) CancelExecution(ctx context.Context, executionID string) error {
	e.mu.Lock()
	execution, ok := e.executions[executionID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("execution not found")
	}

	execution.Status = types.ExecutionStatusCancelled
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(execution.StartedAt)

	return e.store.SaveWorkflowExecution(ctx, execution)
}

// GetExecution retrieves execution details
func (e *Engine) GetExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	// Check in-memory first
	e.mu.RLock()
	if exec, ok := e.executions[executionID]; ok {
		e.mu.RUnlock()
		return exec, nil
	}
	e.mu.RUnlock()

	// Fallback to database
	return e.store.GetWorkflowExecution(ctx, executionID)
}

// GetStatistics returns engine statistics
func (e *Engine) GetStatistics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"active_executions":    len(e.executions),
		"executions_started":   e.executionsStarted,
		"executions_completed": e.executionsCompleted,
		"executions_failed":    e.executionsFailed,
	}
}

// Helper functions

// contains checks if string s contains substr
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
