package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
)

func TestNewExecutor(t *testing.T) {
	logger := zap.NewNop()
	agentManagerURL := "http://localhost:8080"
	reasoningServiceURL := "http://localhost:8081"

	executor := NewExecutor(agentManagerURL, reasoningServiceURL, logger)

	assert.NotNil(t, executor)
	assert.Equal(t, agentManagerURL, executor.agentManagerURL)
	assert.Equal(t, reasoningServiceURL, executor.reasoningServiceURL)
	assert.NotNil(t, executor.httpClient)
	assert.NotNil(t, executor.logger)
}

func TestExecuteCommand_Success(t *testing.T) {
	logger := zap.NewNop()

	// Mock agent-manager server
	commandID := "cmd-001"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/commands" {
			// Return command ID
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     commandID,
				"status": "pending",
			})
		} else if r.URL.Path == "/api/v1/commands/"+commandID+"/result" {
			// Return command result
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "completed",
				"output": "pod list",
			})
		}
	}))
	defer mockServer.Close()

	executor := NewExecutor(mockServer.URL, "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		TriggerEvent: map[string]interface{}{
			"payload": map[string]interface{}{
				"cluster_id": "cluster-001",
			},
		},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "command",
		Config: map[string]interface{}{
			"tool":   "kubectl",
			"action": "get",
			"args":   []interface{}{"pods"},
		},
	}

	result, err := executor.ExecuteCommand(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, commandID, result["command_id"])
}

func TestExecuteCommand_MissingClusterID(t *testing.T) {
	logger := zap.NewNop()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer mockServer.Close()

	executor := NewExecutor(mockServer.URL, "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:           "exec-001",
		WorkflowID:   "workflow-001",
		TriggerEvent: map[string]interface{}{},
	}

	step := types.WorkflowStep{
		ID:     "step-1",
		Type:   "command",
		Config: map[string]interface{}{},
	}

	result, err := executor.ExecuteCommand(ctx, execution, step)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecuteAIAnalysis_Success(t *testing.T) {
	logger := zap.NewNop()

	// Mock reasoning service server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/analyze/root-cause" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"root_cause": map[string]interface{}{
					"type":        "OOMKiller",
					"description": "Container was killed due to OOM",
					"confidence":  0.95,
				},
			})
		}
	}))
	defer mockServer.Close()

	executor := NewExecutor("", mockServer.URL, logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		TriggerEvent: map[string]interface{}{
			"reason": "OOMKilled",
		},
		Context: map[string]interface{}{},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "ai_analysis",
		Config: map[string]interface{}{
			"analysis_type": "root_cause",
		},
	}

	result, err := executor.ExecuteAIAnalysis(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result["root_cause"])
}

func TestExecuteAIAnalysis_ServiceError(t *testing.T) {
	logger := zap.NewNop()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	executor := NewExecutor("", mockServer.URL, logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:     "step-1",
		Type:   "ai_analysis",
		Config: map[string]interface{}{},
	}

	result, err := executor.ExecuteAIAnalysis(ctx, execution, step)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecuteDecision_ConditionMatch(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		Context: map[string]interface{}{
			"analysis_root_cause": "OOM",
		},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "decision",
		Config: map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"if":   "analysis.root_cause == 'OOM'",
					"then": "remediate_oom",
				},
			},
		},
	}

	result, err := executor.ExecuteDecision(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "remediate_oom", result["decision"])
	assert.True(t, result["matched"].(bool))
}

func TestExecuteDecision_NoMatch(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		Context:    map[string]interface{}{},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "decision",
		Config: map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"if":   "analysis.root_cause == 'OOM'",
					"then": "remediate_oom",
				},
			},
		},
	}

	result, err := executor.ExecuteDecision(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "default", result["decision"])
	assert.False(t, result["matched"].(bool))
}

func TestExecuteDecision_InvalidConditions(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "decision",
		Config: map[string]interface{}{
			"conditions": "invalid",
		},
	}

	result, err := executor.ExecuteDecision(ctx, execution, step)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecuteRemediation(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "remediation",
		Config: map[string]interface{}{
			"action_type": "scale",
			"action":      "increase_memory_limit",
		},
	}

	result, err := executor.ExecuteRemediation(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "scale", result["action_type"])
	assert.Equal(t, "increase_memory_limit", result["action"])
	assert.Equal(t, "completed", result["status"])
}

func TestExecuteNotification(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "notification",
		Config: map[string]interface{}{
			"channel": "slack",
			"message": "OOM detected and remediated",
		},
	}

	result, err := executor.ExecuteNotification(ctx, execution, step)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "slack", result["channel"])
	assert.Equal(t, "OOM detected and remediated", result["message"])
	assert.Equal(t, "sent", result["status"])
	assert.NotNil(t, result["sent_at"])
}

func TestExecuteWait_Success(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "wait",
		Config: map[string]interface{}{
			"duration": "100ms",
		},
	}

	start := time.Now()
	result, err := executor.ExecuteWait(ctx, execution, step)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
}

func TestExecuteWait_InvalidDuration(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "wait",
		Config: map[string]interface{}{
			"duration": "invalid",
		},
	}

	result, err := executor.ExecuteWait(ctx, execution, step)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecuteWait_ContextCanceled(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx, cancel := context.WithCancel(context.Background())

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "wait",
		Config: map[string]interface{}{
			"duration": "5s",
		},
	}

	// Cancel context immediately
	cancel()

	result, err := executor.ExecuteWait(ctx, execution, step)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSendHTTPRequest_Success(t *testing.T) {
	logger := zap.NewNop()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
		})
	}))
	defer mockServer.Close()

	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	result, err := executor.sendHTTPRequest(ctx, "POST", mockServer.URL, map[string]interface{}{
		"test": "data",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ok", result["status"])
}

func TestSendHTTPRequest_ServerError(t *testing.T) {
	logger := zap.NewNop()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	result, err := executor.sendHTTPRequest(ctx, "POST", mockServer.URL, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestEvaluateDecisionCondition(t *testing.T) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)

	tests := []struct {
		name      string
		execution *types.WorkflowExecution
		condition string
		expected  bool
	}{
		{
			name: "OOM root cause match",
			execution: &types.WorkflowExecution{
				Context: map[string]interface{}{
					"analysis_root_cause": "OOM",
				},
			},
			condition: "analysis.root_cause == 'OOM'",
			expected:  true,
		},
		{
			name: "Config root cause match",
			execution: &types.WorkflowExecution{
				Context: map[string]interface{}{
					"analysis_root_cause": "Config",
				},
			},
			condition: "analysis.root_cause == 'Config'",
			expected:  true,
		},
		{
			name: "Critical severity match",
			execution: &types.WorkflowExecution{
				TriggerEvent: map[string]interface{}{
					"payload": map[string]interface{}{
						"severity": "critical",
					},
				},
			},
			condition: "severity == 'critical'",
			expected:  true,
		},
		{
			name: "No match",
			execution: &types.WorkflowExecution{
				Context: map[string]interface{}{
					"analysis_root_cause": "Network",
				},
			},
			condition: "analysis.root_cause == 'OOM'",
			expected:  false,
		},
		{
			name:      "Unknown condition",
			execution: &types.WorkflowExecution{},
			condition: "unknown.condition == 'value'",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.evaluateDecisionCondition(tt.execution, tt.condition)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWaitForCommandResult_Success(t *testing.T) {
	logger := zap.NewNop()

	commandID := "cmd-001"
	attempts := 0

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// First attempt, return not ready
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// Second attempt, return result
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "completed",
			"output": "result data",
		})
	}))
	defer mockServer.Close()

	executor := NewExecutor(mockServer.URL, "", logger)
	ctx := context.Background()

	result, err := executor.waitForCommandResult(ctx, commandID, 10*time.Second)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "completed", result["status"])
}

func TestWaitForCommandResult_Timeout(t *testing.T) {
	logger := zap.NewNop()

	commandID := "cmd-001"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return not ready
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	executor := NewExecutor(mockServer.URL, "", logger)
	ctx := context.Background()

	result, err := executor.waitForCommandResult(ctx, commandID, 100*time.Millisecond)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "timeout")
}

// Benchmark tests
func BenchmarkExecuteCommand(b *testing.B) {
	logger := zap.NewNop()

	commandID := "cmd-001"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/commands" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     commandID,
				"status": "pending",
			})
		} else if r.URL.Path == "/api/v1/commands/"+commandID+"/result" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "completed",
			})
		}
	}))
	defer mockServer.Close()

	executor := NewExecutor(mockServer.URL, "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		TriggerEvent: map[string]interface{}{
			"payload": map[string]interface{}{
				"cluster_id": "cluster-001",
			},
		},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "command",
		Config: map[string]interface{}{
			"tool":   "kubectl",
			"action": "get",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteCommand(ctx, execution, step)
	}
}

func BenchmarkExecuteDecision(b *testing.B) {
	logger := zap.NewNop()
	executor := NewExecutor("", "", logger)
	ctx := context.Background()

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-001",
		Context: map[string]interface{}{
			"analysis_root_cause": "OOM",
		},
	}

	step := types.WorkflowStep{
		ID:   "step-1",
		Type: "decision",
		Config: map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"if":   "analysis.root_cause == 'OOM'",
					"then": "remediate_oom",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteDecision(ctx, execution, step)
	}
}
