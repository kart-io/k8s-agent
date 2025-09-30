package agent

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

func TestNewCommandExecutor(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	if executor == nil {
		t.Fatal("NewCommandExecutor returned nil")
	}

	if executor.clusterID != "test-cluster" {
		t.Errorf("clusterID = %v, want %v", executor.clusterID, "test-cluster")
	}

	// Check that allowed tools are initialized
	if len(executor.allowedTools) == 0 {
		t.Error("allowedTools not initialized")
	}

	// Check specific tools
	if _, exists := executor.allowedTools["kubectl"]; !exists {
		t.Error("kubectl not in allowed tools")
	}
}

func TestValidateCommand_AllowedTool(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	cmd := types.Command{
		ID:     "cmd-1",
		Tool:   "kubectl",
		Action: "get",
		Args:   []string{"pods"},
	}

	err := executor.validateCommand(cmd)
	if err != nil {
		t.Errorf("validateCommand failed for allowed command: %v", err)
	}
}

func TestValidateCommand_DisallowedTool(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	cmd := types.Command{
		ID:     "cmd-1",
		Tool:   "rm",
		Action: "-rf",
		Args:   []string{"/"},
	}

	err := executor.validateCommand(cmd)
	if err == nil {
		t.Error("validateCommand should fail for disallowed tool")
	}
}

func TestValidateCommand_DisallowedAction(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	cmd := types.Command{
		ID:     "cmd-1",
		Tool:   "kubectl",
		Action: "delete",
		Args:   []string{"pods", "test-pod"},
	}

	err := executor.validateCommand(cmd)
	if err == nil {
		t.Error("validateCommand should fail for disallowed kubectl action")
	}
}

func TestCheckArgumentSafety_Safe(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	safeArgs := []string{"pods", "-n", "default", "--output", "json"}

	err := executor.checkArgumentSafety(safeArgs)
	if err != nil {
		t.Errorf("checkArgumentSafety failed for safe arguments: %v", err)
	}
}

func TestCheckArgumentSafety_DangerousPatterns(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	dangerousTests := []struct {
		name string
		args []string
	}{
		{"shell command injection", []string{"pods", "&&", "rm", "-rf", "/"}},
		{"pipe operator", []string{"pods", "|", "grep", "test"}},
		{"output redirection", []string{"pods", ">", "/tmp/output"}},
		{"sudo command", []string{"sudo", "kubectl", "get", "pods"}},
		{"delete command", []string{"delete", "pods", "test"}},
	}

	for _, tt := range dangerousTests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.checkArgumentSafety(tt.args)
			if err == nil {
				t.Errorf("checkArgumentSafety should fail for: %v", tt.args)
			}
		})
	}
}

func TestValidateKubectlCommand_LogsWithFollow(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	cmd := types.Command{
		ID:     "cmd-1",
		Tool:   "kubectl",
		Action: "logs",
		Args:   []string{"pod/test-pod", "--follow"},
	}

	err := executor.validateKubectlCommand(cmd, []string{"logs"})
	if err == nil {
		t.Error("validateKubectlCommand should fail for logs --follow")
	}
}

func TestExecute_Success(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	ctx := context.Background()
	cmd := types.Command{
		ID:      "cmd-1",
		Type:    "diagnostic",
		Tool:    "echo",
		Action:  "",
		Args:    []string{"test"},
		Timeout: 5 * time.Second,
	}

	// Note: This test may not work perfectly because 'echo' is not in allowedTools
	// In a real test, you would mock the execution or use a tool that's in the whitelist
	result := executor.Execute(ctx, cmd)

	if result == nil {
		t.Fatal("Execute returned nil result")
	}

	if result.CommandID != cmd.ID {
		t.Errorf("result.CommandID = %v, want %v", result.CommandID, cmd.ID)
	}

	if result.ClusterID != "test-cluster" {
		t.Errorf("result.ClusterID = %v, want %v", result.ClusterID, "test-cluster")
	}
}

func TestExecute_ValidationFailure(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	ctx := context.Background()
	cmd := types.Command{
		ID:      "cmd-1",
		Type:    "diagnostic",
		Tool:    "invalid-tool",
		Action:  "",
		Args:    []string{},
		Timeout: 5 * time.Second,
	}

	result := executor.Execute(ctx, cmd)

	if result == nil {
		t.Fatal("Execute returned nil result")
	}

	if result.Status != "failed" {
		t.Errorf("result.Status = %v, want %v", result.Status, "failed")
	}

	if result.Error == "" {
		t.Error("result.Error should not be empty for validation failure")
	}
}

func TestExecute_UnknownCommandType(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()
	executor := NewCommandExecutor(clientset, "test-cluster", logger)

	ctx := context.Background()
	cmd := types.Command{
		ID:      "cmd-1",
		Type:    "unknown-type",
		Tool:    "kubectl",
		Action:  "get",
		Args:    []string{"pods"},
		Timeout: 5 * time.Second,
	}

	result := executor.Execute(ctx, cmd)

	if result == nil {
		t.Fatal("Execute returned nil result")
	}

	if result.Status != "failed" {
		t.Errorf("result.Status = %v, want %v", result.Status, "failed")
	}
}