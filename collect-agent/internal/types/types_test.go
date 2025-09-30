package types

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// Test default values
	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"CentralEndpoint", config.CentralEndpoint, "nats://localhost:4222"},
		{"ReconnectDelay", config.ReconnectDelay, 5 * time.Second},
		{"HeartbeatInterval", config.HeartbeatInterval, 30 * time.Second},
		{"MetricsInterval", config.MetricsInterval, 60 * time.Second},
		{"BufferSize", config.BufferSize, 1000},
		{"MaxRetries", config.MaxRetries, 10},
		{"LogLevel", config.LogLevel, "info"},
		{"EnableMetrics", config.EnableMetrics, true},
		{"EnableEvents", config.EnableEvents, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestAgentConfigFields(t *testing.T) {
	config := &AgentConfig{
		ClusterID:         "test-cluster",
		CentralEndpoint:   "nats://test:4222",
		ReconnectDelay:    10 * time.Second,
		HeartbeatInterval: 60 * time.Second,
		MetricsInterval:   120 * time.Second,
		BufferSize:        500,
		MaxRetries:        5,
		LogLevel:          "debug",
		EnableMetrics:     false,
		EnableEvents:      false,
	}

	if config.ClusterID != "test-cluster" {
		t.Errorf("ClusterID = %v, want %v", config.ClusterID, "test-cluster")
	}

	if config.BufferSize != 500 {
		t.Errorf("BufferSize = %v, want %v", config.BufferSize, 500)
	}
}

func TestEventStructure(t *testing.T) {
	event := &Event{
		ID:        "event-123",
		ClusterID: "cluster-1",
		Type:      "k8s_event",
		Source:    "kubernetes",
		Namespace: "default",
		Severity:  "high",
		Reason:    "CrashLoopBackOff",
		Message:   "Pod is crashing",
		Timestamp: time.Now(),
		Labels: map[string]string{
			"kind": "Pod",
			"name": "test-pod",
		},
		RawData: map[string]interface{}{
			"count": 5,
		},
	}

	if event.ID != "event-123" {
		t.Errorf("Event.ID = %v, want %v", event.ID, "event-123")
	}

	if event.Severity != "high" {
		t.Errorf("Event.Severity = %v, want %v", event.Severity, "high")
	}

	if event.Labels["kind"] != "Pod" {
		t.Errorf("Event.Labels[kind] = %v, want %v", event.Labels["kind"], "Pod")
	}
}

func TestMetricsStructure(t *testing.T) {
	metrics := &Metrics{
		ClusterID: "cluster-1",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"nodes": map[string]interface{}{
				"total": 3,
				"ready": 3,
			},
		},
	}

	if metrics.ClusterID != "cluster-1" {
		t.Errorf("Metrics.ClusterID = %v, want %v", metrics.ClusterID, "cluster-1")
	}

	if metrics.Data == nil {
		t.Fatal("Metrics.Data is nil")
	}
}

func TestCommandStructure(t *testing.T) {
	cmd := &Command{
		ID:      "cmd-123",
		Type:    "diagnostic",
		Tool:    "kubectl",
		Action:  "get",
		Args:    []string{"pods", "-n", "default"},
		Timeout: 30 * time.Second,
		Env: map[string]string{
			"KUBECONFIG": "/path/to/config",
		},
	}

	if cmd.Tool != "kubectl" {
		t.Errorf("Command.Tool = %v, want %v", cmd.Tool, "kubectl")
	}

	if len(cmd.Args) != 3 {
		t.Errorf("len(Command.Args) = %v, want %v", len(cmd.Args), 3)
	}
}

func TestCommandResultStructure(t *testing.T) {
	result := &CommandResult{
		CommandID: "cmd-123",
		ClusterID: "cluster-1",
		Status:    "success",
		Output:    "pod/test-pod Running",
		Error:     "",
		Duration:  500 * time.Millisecond,
		Timestamp: time.Now(),
	}

	if result.Status != "success" {
		t.Errorf("CommandResult.Status = %v, want %v", result.Status, "success")
	}

	if result.Duration != 500*time.Millisecond {
		t.Errorf("CommandResult.Duration = %v, want %v", result.Duration, 500*time.Millisecond)
	}
}

func TestHeartbeatStructure(t *testing.T) {
	hb := &Heartbeat{
		ClusterID: "cluster-1",
		Timestamp: time.Now(),
		Status:    "healthy",
		Metrics: HeartbeatMetrics{
			EventQueueSize:   10,
			MetricsQueueSize: 5,
			CommandQueueSize: 2,
			UptimeSeconds:    3600,
		},
	}

	if hb.Status != "healthy" {
		t.Errorf("Heartbeat.Status = %v, want %v", hb.Status, "healthy")
	}

	if hb.Metrics.EventQueueSize != 10 {
		t.Errorf("Heartbeat.Metrics.EventQueueSize = %v, want %v", hb.Metrics.EventQueueSize, 10)
	}
}