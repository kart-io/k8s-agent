package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/logger/core"
)

// Engine manages workflow execution.
type Engine struct {
	db       *gorm.DB        // Direct GORM DB access
	cache    *goredis.Client // Direct Redis client access
	executor *Executor
	logger   core.Logger

	// Timeout configuration
	globalTimeout      time.Duration
	stepDefaultTimeout time.Duration
	retryOnTimeout     bool
	maxRetries         int

	// Execution tracking
	mu          sync.RWMutex
	executions  map[string]*types.WorkflowExecution
	cancelFuncs map[string]context.CancelFunc // Cancel functions for running workflows

	// Metrics
	executionsStarted   int64
	executionsCompleted int64
	executionsFailed    int64
	executionsTimedOut  int64
}

// NewEngine creates a new workflow engine.
func NewEngine(
	db *gorm.DB,
	cache *goredis.Client,
	executor *Executor,
	logger core.Logger,
) *Engine {
	return &Engine{
		db:                 db,
		cache:              cache,
		executor:           executor,
		logger:             logger,
		globalTimeout:      30 * time.Minute, // Default 30 minutes
		stepDefaultTimeout: 5 * time.Minute,  // Default 5 minutes
		retryOnTimeout:     true,
		maxRetries:         3,
		executions:         make(map[string]*types.WorkflowExecution),
		cancelFuncs:        make(map[string]context.CancelFunc),
	}
}

// SetTimeoutConfig sets timeout configuration for the engine.
func (e *Engine) SetTimeoutConfig(globalTimeout, stepDefaultTimeout time.Duration, retryOnTimeout bool, maxRetries int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if globalTimeout > 0 {
		e.globalTimeout = globalTimeout
	}
	if stepDefaultTimeout > 0 {
		e.stepDefaultTimeout = stepDefaultTimeout
	}
	e.retryOnTimeout = retryOnTimeout
	e.maxRetries = maxRetries

	e.logger.Info("Workflow timeout configuration updated",
		"global_timeout", globalTimeout,
		"step_default_timeout", stepDefaultTimeout,
		"retry_on_timeout", retryOnTimeout,
		"max_retries", maxRetries)
}

// StartWorkflow starts a new workflow execution.
func (e *Engine) StartWorkflow(ctx context.Context, workflowID string, triggerEvent map[string]interface{}) (*types.WorkflowExecution, error) {
	e.logger.Info("🎬 Starting workflow execution",
		"workflow_id", workflowID)

	// Load workflow definition directly from DB
	var workflow types.Workflow
	if err := e.db.WithContext(ctx).First(&workflow, "id = ?", workflowID).Error; err != nil {
		e.logger.Error("❌ Failed to load workflow from database",
			"workflow_id", workflowID,
			"error", err)
		return nil, fmt.Errorf("failed to load workflow: %w", err)
	}

	e.logger.Info("✓ Workflow loaded",
		"workflow_name", workflow.Name,
		"status", string(workflow.Status),
		"steps", len(workflow.Steps))

	if workflow.Status != types.WorkflowStatusActive {
		e.logger.Warn("⚠️  Workflow is not active",
			"workflow_id", workflowID,
			"status", string(workflow.Status))
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
		"execution_id", executionID)

	// Save execution directly to DB
	if err := e.db.WithContext(ctx).Save(execution).Error; err != nil {
		e.logger.Error("❌ Failed to save execution to database",
			"execution_id", executionID,
			"error", err)
		return nil, fmt.Errorf("failed to save execution: %w", err)
	}

	e.logger.Info("✓ Execution saved to database",
		"execution_id", executionID)

	// Track in memory
	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.executionsStarted++
	currentStats := e.executionsStarted
	e.mu.Unlock()

	e.logger.Info("📊 Execution tracking updated",
		"execution_id", executionID,
		"total_started", currentStats)

	// Create context with timeout for workflow execution
	workflowTimeout := e.globalTimeout
	if workflow.Timeout > 0 {
		workflowTimeout = workflow.Timeout
	}

	workflowCtx, cancel := context.WithTimeout(context.Background(), workflowTimeout)

	// Store cancel function for potential cancellation
	e.mu.Lock()
	e.cancelFuncs[executionID] = cancel
	e.mu.Unlock()

	// Start execution asynchronously
	go e.executeWorkflow(workflowCtx, &workflow, execution)

	e.logger.Info("Workflow execution started",
		"execution_id", execution.ID,
		"workflow_id", workflowID,
		"timeout", workflowTimeout)

	return execution, nil
}

// executeWorkflow executes a workflow.
func (e *Engine) executeWorkflow(ctx context.Context, workflow *types.Workflow, execution *types.WorkflowExecution) {
	defer func() {
		// Cleanup tracking
		e.mu.Lock()
		delete(e.executions, execution.ID)
		if cancel, ok := e.cancelFuncs[execution.ID]; ok {
			cancel() // Cancel the context
			delete(e.cancelFuncs, execution.ID)
		}
		e.mu.Unlock()
	}()

	// Update status to running
	execution.Status = types.ExecutionStatusRunning
	if err := e.db.WithContext(ctx).Model(&types.WorkflowExecution{}).
		Where("id = ?", execution.ID).
		Update("status", types.ExecutionStatusRunning).Error; err != nil {
		e.logger.Errorw("Failed to update workflow execution status to running",
			"execution_id", execution.ID,
			"error", err)
		// Continue execution despite update failure
	}

	// Execute steps in sequence with timeout monitoring
	for i, step := range workflow.Steps {
		// Check context for timeout or cancellation
		select {
		case <-ctx.Done():
			// Context cancelled or timed out
			e.handleWorkflowTimeout(ctx, execution)
			return
		default:
			// Continue execution
		}

		e.logger.Info("Executing workflow step",
			"execution_id", execution.ID,
			"step_id", step.ID,
			"step_name", step.Name,
			"step_index", i)

		// Check if we should execute this step
		if !e.shouldExecuteStep(execution, step) {
			e.logger.Debug("Skipping step due to conditions",
				"step_id", step.ID)
			continue
		}

		// Execute step with timeout
		stepExec, err := e.executeStepWithTimeout(ctx, execution, step)
		if err != nil {
			// Check if error is due to context cancellation/timeout
			if ctx.Err() != nil {
				e.handleWorkflowTimeout(ctx, execution)
				return
			}

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
		if err := e.db.WithContext(ctx).Save(execution).Error; err != nil {
			e.logger.Error("Failed to save workflow execution progress",
				"execution_id", execution.ID,
				"error", err)
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

// executeStepWithTimeout executes a single workflow step with timeout.
func (e *Engine) executeStepWithTimeout(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (*types.StepExecution, error) {
	// Determine step timeout
	stepTimeout := e.stepDefaultTimeout
	if step.Timeout > 0 {
		stepTimeout = step.Timeout
	}

	// Create context with timeout for this step
	stepCtx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	e.logger.Debug("Executing step with timeout",
		"step_id", step.ID,
		"timeout", stepTimeout)

	// Execute step
	return e.executeStep(stepCtx, execution, step)
}

// executeStep executes a single workflow step.
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
			"step_id", step.ID,
			"retry_count", retryCount,
			"error", err)

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

// shouldExecuteStep checks if step conditions are met.
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

// evaluateCondition evaluates a single condition.
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

// prepareStepInput prepares input for step execution.
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

// calculateRetryDelay calculates retry delay with exponential backoff.
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

// handleStepFailure handles step execution failure.
func (e *Engine) handleStepFailure(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep, err error) {
	e.logger.Error("Step execution failed",
		"execution_id", execution.ID,
		"step_id", step.ID,
		"error", err)

	e.completeExecution(ctx, execution, types.ExecutionStatusFailed, err.Error())
}

// executeFailureBranch executes failure branch.
func (e *Engine) executeFailureBranch(ctx context.Context, workflow *types.Workflow, execution *types.WorkflowExecution, failedStep types.WorkflowStep) {
	e.logger.Info("Executing failure branch",
		"execution_id", execution.ID,
		"failed_step_id", failedStep.ID,
		"failure_steps", failedStep.OnFailure)

	// Execute each step in the failure branch
	for _, stepID := range failedStep.OnFailure {
		// Find the step in the workflow
		var nextStep *types.WorkflowStep
		for i := range workflow.Steps {
			if workflow.Steps[i].ID == stepID {
				nextStep = &workflow.Steps[i] // Use index to avoid implicit memory aliasing
				break
			}
		}

		if nextStep == nil {
			e.logger.Warn("Failure branch step not found in workflow",
				"step_id", stepID,
				"execution_id", execution.ID)
			continue
		}

		e.logger.Info("Executing failure branch step",
			"execution_id", execution.ID,
			"step_id", nextStep.ID,
			"step_name", nextStep.Name)

		// Execute the failure handling step
		stepExec, err := e.executeStepWithTimeout(ctx, execution, *nextStep)
		if err != nil {
			e.logger.Error("Failure branch step execution error",
				"execution_id", execution.ID,
				"step_id", nextStep.ID,
				"error", err)
			// Continue to next failure step even if this one fails
			continue
		}

		// Add step execution to history
		execution.StepExecutions = append(execution.StepExecutions, *stepExec)

		// Update context with step output
		if stepExec.Output != nil {
			for k, v := range stepExec.Output {
				execution.Context[fmt.Sprintf("step_%s_%s", nextStep.ID, k)] = v
			}
		}

		// Save progress
		if err := e.db.WithContext(ctx).Save(execution).Error; err != nil {
			e.logger.Error("Failed to save workflow execution progress",
				"execution_id", execution.ID,
				"error", err)
		}

		e.logger.Info("Failure branch step completed",
			"execution_id", execution.ID,
			"step_id", nextStep.ID,
			"status", stepExec.Status)
	}

	// After executing all failure branch steps, mark workflow as failed
	e.completeExecution(ctx, execution, types.ExecutionStatusFailed,
		fmt.Sprintf("Step %s failed, failure branch executed", failedStep.ID))
}

// completeExecution completes workflow execution.
func (e *Engine) completeExecution(ctx context.Context, execution *types.WorkflowExecution, status types.ExecutionStatus, errorMsg string) {
	completedAt := time.Now()
	execution.Status = status
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(execution.StartedAt)

	if errorMsg != "" {
		execution.Error = errorMsg
	}

	// Save final state - critical operation
	if err := e.db.WithContext(ctx).Save(execution).Error; err != nil {
		e.logger.Error("CRITICAL: Failed to save final workflow execution state",
			"execution_id", execution.ID,
			"status", string(status),
			"error", err)
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
		"execution_id", execution.ID,
		"status", string(status),
		"duration", execution.Duration)
}

// CancelExecution cancels a running execution.
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

	return e.db.WithContext(ctx).Save(execution).Error
}

// GetExecution retrieves execution details.
func (e *Engine) GetExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	// Check in-memory first
	e.mu.RLock()
	if exec, ok := e.executions[executionID]; ok {
		e.mu.RUnlock()
		return exec, nil
	}
	e.mu.RUnlock()

	// Fallback to database
	var execution types.WorkflowExecution
	if err := e.db.WithContext(ctx).First(&execution, "id = ?", executionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get workflow execution %s: %w", executionID, err)
	}
	return &execution, nil
}

// GetStatistics returns engine statistics.
func (e *Engine) GetStatistics() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"active_executions":    len(e.executions),
		"executions_started":   e.executionsStarted,
		"executions_completed": e.executionsCompleted,
		"executions_failed":    e.executionsFailed,
		"executions_timed_out": e.executionsTimedOut,
		"global_timeout":       e.globalTimeout.String(),
		"step_default_timeout": e.stepDefaultTimeout.String(),
		"retry_on_timeout":     e.retryOnTimeout,
		"max_retries":          e.maxRetries,
	}
}

// handleWorkflowTimeout handles workflow timeout.
func (e *Engine) handleWorkflowTimeout(ctx context.Context, execution *types.WorkflowExecution) {
	e.logger.Warn("Workflow execution timed out",
		"execution_id", execution.ID,
		"duration", time.Since(execution.StartedAt))

	// Perform cleanup operations
	e.cleanupExecution(ctx, execution)

	// Update metrics
	e.mu.Lock()
	e.executionsTimedOut++
	e.mu.Unlock()

	// Check if we should retry
	if e.retryOnTimeout && execution.Context != nil {
		retryCount, _ := execution.Context["retry_count"].(int)
		if retryCount < e.maxRetries {
			e.logger.Info("Scheduling workflow retry after timeout",
				"execution_id", execution.ID,
				"retry_count", retryCount+1,
				"max_retries", e.maxRetries)

			// Update retry count in context
			execution.Context["retry_count"] = retryCount + 1
			execution.Context["previous_timeout"] = true

			// Complete current execution as timeout
			e.completeExecution(ctx, execution, types.ExecutionStatusTimeout,
				fmt.Sprintf("Workflow timed out after %s (retry %d/%d)",
					time.Since(execution.StartedAt), retryCount+1, e.maxRetries))

			// Note: Retry scheduling would be handled by an external scheduler
			// For now, we just mark the execution as timed out with retry info
			return
		}
	}

	// Complete execution as timeout (no retry)
	e.completeExecution(ctx, execution, types.ExecutionStatusTimeout,
		fmt.Sprintf("Workflow timed out after %s", time.Since(execution.StartedAt)))
}

// cleanupExecution performs cleanup operations for a timed-out or cancelled execution.
func (e *Engine) cleanupExecution(ctx context.Context, execution *types.WorkflowExecution) {
	e.logger.Info("Performing cleanup for execution",
		"execution_id", execution.ID,
		"status", execution.Status)

	// 1. Release any locks or resources held by the execution
	// (This would be service-specific - add as needed)

	// 2. Cancel any in-flight HTTP requests
	// (Already handled by context cancellation)

	// 3. Update execution status to indicate cleanup in progress
	execution.Context["cleanup_started"] = time.Now()

	// 4. Mark any incomplete steps as cancelled
	for i := range execution.StepExecutions {
		if execution.StepExecutions[i].Status == types.ExecutionStatusRunning {
			now := time.Now()
			execution.StepExecutions[i].Status = types.ExecutionStatusCancelled
			execution.StepExecutions[i].CompletedAt = &now
			execution.StepExecutions[i].Error = "Cancelled due to workflow timeout"
		}
	}

	// 5. Save cleanup state
	if err := e.db.WithContext(context.Background()).Save(execution).Error; err != nil {
		e.logger.Error("Failed to save execution cleanup state",
			"execution_id", execution.ID,
			"error", err)
	}

	e.logger.Info("Cleanup completed for execution",
		"execution_id", execution.ID)
}

// CancelExecutionWithCleanup cancels a running execution and performs cleanup.
func (e *Engine) CancelExecutionWithCleanup(ctx context.Context, executionID string, reason string) error {
	e.mu.Lock()
	execution, ok := e.executions[executionID]
	cancel, hasCancel := e.cancelFuncs[executionID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("execution not found or not running")
	}

	e.logger.Info("Cancelling workflow execution",
		"execution_id", executionID,
		"reason", reason)

	// Cancel the context
	if hasCancel {
		cancel()
	}

	// Perform cleanup
	e.cleanupExecution(ctx, execution)

	// Update execution status
	execution.Status = types.ExecutionStatusCancelled
	completedAt := time.Now()
	execution.CompletedAt = &completedAt
	execution.Duration = completedAt.Sub(execution.StartedAt)
	execution.Error = reason

	// Save final state
	if err := e.db.WithContext(ctx).Save(execution).Error; err != nil {
		e.logger.Error("Failed to save cancelled execution state",
			"execution_id", executionID,
			"error", err)
		return fmt.Errorf("failed to save cancelled state: %w", err)
	}

	return nil
}

// Helper functions

// contains checks if string s contains substr.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
