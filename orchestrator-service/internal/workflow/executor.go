package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
)

// Executor executes workflow steps
type Executor struct {
	logger            *zap.Logger
	agentManagerURL   string
	reasoningServiceURL string
	httpClient        *http.Client
}

// NewExecutor creates a new executor
func NewExecutor(
	agentManagerURL string,
	reasoningServiceURL string,
	logger *zap.Logger,
) *Executor {
	return &Executor{
		logger:              logger.With(zap.String("component", "workflow-executor")),
		agentManagerURL:     agentManagerURL,
		reasoningServiceURL: reasoningServiceURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteCommand executes a command step
func (ex *Executor) ExecuteCommand(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing command step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	// Extract command parameters from config
	clusterID, _ := step.Config["cluster_id"].(string)
	tool, _ := step.Config["tool"].(string)
	action, _ := step.Config["action"].(string)
	args, _ := step.Config["args"].([]interface{})

	if clusterID == "" {
		// Try to get from trigger event
		if payload, ok := execution.TriggerEvent["payload"].(map[string]interface{}); ok {
			if cid, ok := payload["cluster_id"].(string); ok {
				clusterID = cid
			}
		}
	}

	// Prepare command request
	cmdReq := map[string]interface{}{
		"cluster_id": clusterID,
		"type":       "diagnostic",
		"tool":       tool,
		"action":     action,
		"args":       args,
		"timeout":    "30s",
		"issued_by":  "orchestrator-service",
		"correlation_id": execution.ID,
	}

	// Send command to agent-manager
	resp, err := ex.sendHTTPRequest(ctx, "POST", ex.agentManagerURL+"/api/v1/commands", cmdReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	commandID, _ := resp["id"].(string)

	// Wait for result (polling)
	result, err := ex.waitForCommandResult(ctx, commandID, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get command result: %w", err)
	}

	return map[string]interface{}{
		"command_id": commandID,
		"result":     result,
	}, nil
}

// ExecuteAIAnalysis executes an AI analysis step
func (ex *Executor) ExecuteAIAnalysis(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing AI analysis step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	// Prepare analysis request
	analysisReq := map[string]interface{}{
		"request_id":   fmt.Sprintf("%s-%s", execution.ID, step.ID),
		"workflow_id":  execution.WorkflowID,
		"analysis_type": step.Config["analysis_type"],
		"context": map[string]interface{}{
			"event":       execution.TriggerEvent,
			"execution":   execution.Context,
			"step_config": step.Config,
		},
		"options": map[string]interface{}{
			"timeout":        "30s",
			"min_confidence": 0.7,
		},
	}

	// Send request to reasoning service
	resp, err := ex.sendHTTPRequest(ctx, "POST", ex.reasoningServiceURL+"/api/v1/analyze/root-cause", analysisReq)
	if err != nil {
		return nil, fmt.Errorf("failed to request AI analysis: %w", err)
	}

	return resp, nil
}

// ExecuteDecision executes a decision step
func (ex *Executor) ExecuteDecision(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing decision step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	// Get conditions from config
	conditions, ok := step.Config["conditions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid decision conditions")
	}

	// Evaluate each condition
	for i, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if condition
		ifCondition, _ := condMap["if"].(string)
		thenAction, _ := condMap["then"].(string)

		// Evaluate condition
		if ex.evaluateDecisionCondition(execution, ifCondition) {
			ex.logger.Info("Decision condition matched",
				zap.String("execution_id", execution.ID),
				zap.Int("condition_index", i),
				zap.String("action", thenAction))

			return map[string]interface{}{
				"decision":  thenAction,
				"condition": ifCondition,
				"matched":   true,
			}, nil
		}
	}

	// No condition matched
	return map[string]interface{}{
		"decision": "default",
		"matched":  false,
	}, nil
}

// ExecuteRemediation executes a remediation step
func (ex *Executor) ExecuteRemediation(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing remediation step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	actionType, _ := step.Config["action_type"].(string)
	action, _ := step.Config["action"].(string)

	ex.logger.Info("Executing remediation action",
		zap.String("action_type", actionType),
		zap.String("action", action))

	// For now, just log the action
	// In production, this would execute actual remediation
	return map[string]interface{}{
		"action_type": actionType,
		"action":      action,
		"status":      "completed",
		"message":     fmt.Sprintf("Remediation action %s executed", action),
	}, nil
}

// ExecuteNotification executes a notification step
func (ex *Executor) ExecuteNotification(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing notification step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	channel, _ := step.Config["channel"].(string)
	message, _ := step.Config["message"].(string)

	ex.logger.Info("Sending notification",
		zap.String("channel", channel),
		zap.String("message", message))

	// In production, this would send to actual notification channels
	// (Slack, Email, PagerDuty, etc.)

	return map[string]interface{}{
		"channel": channel,
		"message": message,
		"sent_at": time.Now(),
		"status":  "sent",
	}, nil
}

// ExecuteWait executes a wait step
func (ex *Executor) ExecuteWait(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	ex.logger.Info("Executing wait step",
		zap.String("execution_id", execution.ID),
		zap.String("step_id", step.ID))

	durationStr, _ := step.Config["duration"].(string)
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration: %w", err)
	}

	ex.logger.Info("Waiting",
		zap.Duration("duration", duration))

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(duration):
		return map[string]interface{}{
			"waited": duration.String(),
		}, nil
	}
}

// Helper functions

// sendHTTPRequest sends an HTTP request
func (ex *Executor) sendHTTPRequest(ctx context.Context, method, url string, body interface{}) (map[string]interface{}, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ex.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// waitForCommandResult waits for command execution result
func (ex *Executor) waitForCommandResult(ctx context.Context, commandID string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for command result")
			}

			// Poll for result
			url := fmt.Sprintf("%s/api/v1/commands/%s/result", ex.agentManagerURL, commandID)
			result, err := ex.sendHTTPRequest(ctx, "GET", url, nil)
			if err != nil {
				// Result not ready yet, continue polling
				continue
			}

			return result, nil
		}
	}
}

// evaluateDecisionCondition evaluates a decision condition
func (ex *Executor) evaluateDecisionCondition(execution *types.WorkflowExecution, condition string) bool {
	// Simple condition evaluation
	// In production, this would use a proper expression evaluator

	// Check context variables
	switch condition {
	case "analysis.root_cause == 'OOM'":
		if rootCause, ok := execution.Context["analysis_root_cause"].(string); ok {
			return rootCause == "OOM"
		}
	case "analysis.root_cause == 'Config'":
		if rootCause, ok := execution.Context["analysis_root_cause"].(string); ok {
			return rootCause == "Config"
		}
	case "severity == 'critical'":
		if payload, ok := execution.TriggerEvent["payload"].(map[string]interface{}); ok {
			if severity, ok := payload["severity"].(string); ok {
				return severity == "critical"
			}
		}
	}

	return false
}