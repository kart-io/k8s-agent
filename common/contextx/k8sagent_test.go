package contextx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestK8sAgentContext(t *testing.T) {
	ctx := context.Background()

	t.Run("AgentID", func(t *testing.T) {
		agentID := "agent-12345"
		ctx = WithAgentID(ctx, agentID)
		assert.Equal(t, agentID, GetAgentID(ctx))

		// Test empty context
		emptyCtx := context.Background()
		assert.Empty(t, GetAgentID(emptyCtx))
	})

	t.Run("ClusterID", func(t *testing.T) {
		clusterID := "cluster-prod-us-west-2"
		ctx = WithClusterID(ctx, clusterID)
		assert.Equal(t, clusterID, GetClusterID(ctx))
	})

	t.Run("WorkflowID", func(t *testing.T) {
		workflowID := "wf-diagnose-pod-crash"
		ctx = WithWorkflowID(ctx, workflowID)
		assert.Equal(t, workflowID, GetWorkflowID(ctx))
	})

	t.Run("TaskID", func(t *testing.T) {
		taskID := "task-collect-logs"
		ctx = WithTaskID(ctx, taskID)
		assert.Equal(t, taskID, GetTaskID(ctx))
	})

	t.Run("EventID", func(t *testing.T) {
		eventID := "event-crashloopbackoff-123"
		ctx = WithEventID(ctx, eventID)
		assert.Equal(t, eventID, GetEventID(ctx))
	})

	t.Run("CommandID", func(t *testing.T) {
		commandID := "cmd-kubectl-logs-456"
		ctx = WithCommandID(ctx, commandID)
		assert.Equal(t, commandID, GetCommandID(ctx))
	})

	t.Run("ExtractK8sAgentInfo - full context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req-001")
		ctx = WithUserID(ctx, "user-123")
		ctx = WithTraceID(ctx, "trace-abc")
		ctx = WithAgentID(ctx, "agent-789")
		ctx = WithClusterID(ctx, "cluster-prod")
		ctx = WithWorkflowID(ctx, "wf-001")
		ctx = WithTaskID(ctx, "task-001")
		ctx = WithEventID(ctx, "event-001")
		ctx = WithCommandID(ctx, "cmd-001")

		info := ExtractK8sAgentInfo(ctx)

		assert.Equal(t, "req-001", info["request_id"])
		assert.Equal(t, "user-123", info["user_id"])
		assert.Equal(t, "trace-abc", info["trace_id"])
		assert.Equal(t, "agent-789", info["agent_id"])
		assert.Equal(t, "cluster-prod", info["cluster_id"])
		assert.Equal(t, "wf-001", info["workflow_id"])
		assert.Equal(t, "task-001", info["task_id"])
		assert.Equal(t, "event-001", info["event_id"])
		assert.Equal(t, "cmd-001", info["command_id"])
	})

	t.Run("ExtractK8sAgentInfo - partial context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req-002")
		ctx = WithAgentID(ctx, "agent-456")
		ctx = WithClusterID(ctx, "cluster-staging")

		info := ExtractK8sAgentInfo(ctx)

		assert.Equal(t, "req-002", info["request_id"])
		assert.Equal(t, "agent-456", info["agent_id"])
		assert.Equal(t, "cluster-staging", info["cluster_id"])

		// Fields not set should not be in map
		_, hasUserID := info["user_id"]
		_, hasWorkflowID := info["workflow_id"]
		assert.False(t, hasUserID)
		assert.False(t, hasWorkflowID)
	})

	t.Run("ExtractK8sAgentInfo - empty context", func(t *testing.T) {
		ctx := context.Background()
		info := ExtractK8sAgentInfo(ctx)

		// Should return empty map
		assert.Empty(t, info)
	})

	t.Run("Workflow execution scenario", func(t *testing.T) {
		// Simulate a typical workflow execution context
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req-workflow-123")
		ctx = WithUserID(ctx, "user-admin")
		ctx = WithTraceID(ctx, "trace-workflow-exec")
		ctx = WithAgentID(ctx, "agent-prod-01")
		ctx = WithClusterID(ctx, "cluster-prod-us-east-1")
		ctx = WithWorkflowID(ctx, "wf-diagnose-crashloop")
		ctx = WithTaskID(ctx, "task-collect-pod-logs")
		ctx = WithEventID(ctx, "event-pod-crashloopbackoff")

		// Verify all fields are accessible
		assert.Equal(t, "req-workflow-123", GetRequestID(ctx))
		assert.Equal(t, "agent-prod-01", GetAgentID(ctx))
		assert.Equal(t, "cluster-prod-us-east-1", GetClusterID(ctx))
		assert.Equal(t, "wf-diagnose-crashloop", GetWorkflowID(ctx))
		assert.Equal(t, "task-collect-pod-logs", GetTaskID(ctx))
		assert.Equal(t, "event-pod-crashloopbackoff", GetEventID(ctx))

		// Verify info extraction for logging
		info := ExtractK8sAgentInfo(ctx)
		assert.Len(t, info, 8) // All fields except command_id
	})

	t.Run("Command execution scenario", func(t *testing.T) {
		// Simulate a command execution context
		ctx := context.Background()
		ctx = WithAgentID(ctx, "agent-prod-02")
		ctx = WithClusterID(ctx, "cluster-prod-us-west-2")
		ctx = WithCommandID(ctx, "cmd-kubectl-describe-pod")
		ctx = WithTraceID(ctx, "trace-cmd-exec-456")

		assert.Equal(t, "agent-prod-02", GetAgentID(ctx))
		assert.Equal(t, "cmd-kubectl-describe-pod", GetCommandID(ctx))
		assert.Equal(t, "trace-cmd-exec-456", GetTraceID(ctx))

		info := ExtractK8sAgentInfo(ctx)
		assert.Equal(t, "agent-prod-02", info["agent_id"])
		assert.Equal(t, "cmd-kubectl-describe-pod", info["command_id"])
	})
}
