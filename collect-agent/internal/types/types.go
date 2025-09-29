package types

import (
	"time"
)

// AgentInfo contains information about the agent registration
type AgentInfo struct {
	ClusterID    string    `json:"cluster_id"`
	Version      string    `json:"version"`
	StartTime    time.Time `json:"start_time"`
	Capabilities []string  `json:"capabilities"`
}

// Event represents a Kubernetes event that needs to be reported
type Event struct {
	ID         string                 `json:"id"`
	ClusterID  string                 `json:"cluster_id"`
	Type       string                 `json:"type"`
	Source     string                 `json:"source"`
	Namespace  string                 `json:"namespace"`
	Severity   string                 `json:"severity"`
	Reason     string                 `json:"reason"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	ReportedAt time.Time              `json:"reported_at"`
	Labels     map[string]string      `json:"labels"`
	RawData    map[string]interface{} `json:"raw_data"`
}

// Metrics represents collected metrics from the cluster
type Metrics struct {
	ClusterID string                 `json:"cluster_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Command represents a command to be executed by the agent
type Command struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Tool      string            `json:"tool"`
	Action    string            `json:"action"`
	Args      []string          `json:"args"`
	Env       map[string]string `json:"env,omitempty"`
	Timeout   time.Duration     `json:"timeout"`
	CreatedAt time.Time         `json:"created_at"`
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	CommandID string        `json:"command_id"`
	ClusterID string        `json:"cluster_id"`
	Status    string        `json:"status"` // success, failed, timeout
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// Heartbeat represents agent health status
type Heartbeat struct {
	ClusterID string            `json:"cluster_id"`
	Timestamp time.Time         `json:"timestamp"`
	Status    string            `json:"status"`
	Metrics   HeartbeatMetrics  `json:"metrics"`
}

// HeartbeatMetrics contains internal metrics for heartbeat
type HeartbeatMetrics struct {
	EventQueueSize   int `json:"event_queue_size"`
	MetricsQueueSize int `json:"metrics_queue_size"`
	CommandQueueSize int `json:"command_queue_size"`
	UptimeSeconds    int `json:"uptime_seconds"`
}

// AgentConfig represents the agent configuration
type AgentConfig struct {
	ClusterID         string        `yaml:"cluster_id"`
	CentralEndpoint   string        `yaml:"central_endpoint"`
	ReconnectDelay    time.Duration `yaml:"reconnect_delay"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	MetricsInterval   time.Duration `yaml:"metrics_interval"`
	BufferSize        int           `yaml:"buffer_size"`
	MaxRetries        int           `yaml:"max_retries"`
	LogLevel          string        `yaml:"log_level"`
	EnableMetrics     bool          `yaml:"enable_metrics"`
	EnableEvents      bool          `yaml:"enable_events"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		CentralEndpoint:   "nats://localhost:4222",
		ReconnectDelay:    5 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MetricsInterval:   60 * time.Second,
		BufferSize:        1000,
		MaxRetries:        10,
		LogLevel:          "info",
		EnableMetrics:     true,
		EnableEvents:      true,
	}
}