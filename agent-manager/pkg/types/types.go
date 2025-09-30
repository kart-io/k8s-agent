package types

import (
	"time"
)

// Agent represents a registered agent
type Agent struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	ClusterID       string                 `json:"cluster_id" gorm:"index;not null"`
	ClusterName     string                 `json:"cluster_name"`
	Version         string                 `json:"version"`
	Status          AgentStatus            `json:"status" gorm:"index"`
	LastHeartbeat   time.Time              `json:"last_heartbeat" gorm:"index"`
	RegisteredAt    time.Time              `json:"registered_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	Capabilities    []string               `json:"capabilities" gorm:"type:jsonb"`
	ConnectionInfo  *ConnectionInfo        `json:"connection_info" gorm:"type:jsonb"`
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusOnline      AgentStatus = "online"
	AgentStatusOffline     AgentStatus = "offline"
	AgentStatusRegistering AgentStatus = "registering"
	AgentStatusError       AgentStatus = "error"
)

// ConnectionInfo contains agent connection details
type ConnectionInfo struct {
	Endpoint      string    `json:"endpoint"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastSeen      time.Time `json:"last_seen"`
	ReconnectCount int      `json:"reconnect_count"`
}

// Event represents a Kubernetes event
type Event struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	ClusterID string                 `json:"cluster_id" gorm:"index;not null"`
	Timestamp time.Time              `json:"timestamp" gorm:"index"`
	Type      string                 `json:"type" gorm:"index"`
	Source    string                 `json:"source"`
	Severity  string                 `json:"severity" gorm:"index"`
	Reason    string                 `json:"reason" gorm:"index"`
	Message   string                 `json:"message"`
	Namespace string                 `json:"namespace" gorm:"index"`
	Labels    map[string]string      `json:"labels" gorm:"type:jsonb"`
	RawData   map[string]interface{} `json:"raw_data" gorm:"type:jsonb"`
	ProcessedAt time.Time            `json:"processed_at"`
}

// Metrics represents cluster metrics
type Metrics struct {
	ID               string                 `json:"id" gorm:"primaryKey"`
	ClusterID        string                 `json:"cluster_id" gorm:"index;not null"`
	Timestamp        time.Time              `json:"timestamp" gorm:"index"`
	ClusterMetrics   map[string]interface{} `json:"cluster_metrics" gorm:"type:jsonb"`
	NodeMetrics      []map[string]interface{} `json:"node_metrics" gorm:"type:jsonb"`
	PodMetrics       []map[string]interface{} `json:"pod_metrics" gorm:"type:jsonb"`
	NamespaceMetrics []map[string]interface{} `json:"namespace_metrics" gorm:"type:jsonb"`
}

// Command represents a command to be executed
type Command struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	ClusterID     string                 `json:"cluster_id" gorm:"index;not null"`
	Type          string                 `json:"type"`
	Tool          string                 `json:"tool"`
	Action        string                 `json:"action"`
	Args          []string               `json:"args" gorm:"type:jsonb"`
	Namespace     string                 `json:"namespace"`
	Timeout       time.Duration          `json:"timeout"`
	IssuedBy      string                 `json:"issued_by"`
	CorrelationID string                 `json:"correlation_id" gorm:"index"`
	Status        CommandStatus          `json:"status" gorm:"index"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// CommandStatus represents the status of a command
type CommandStatus string

const (
	CommandStatusPending   CommandStatus = "pending"
	CommandStatusSent      CommandStatus = "sent"
	CommandStatusExecuting CommandStatus = "executing"
	CommandStatusCompleted CommandStatus = "completed"
	CommandStatusFailed    CommandStatus = "failed"
	CommandStatusTimeout   CommandStatus = "timeout"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	ID            string        `json:"id" gorm:"primaryKey"`
	CommandID     string        `json:"command_id" gorm:"index;not null"`
	ClusterID     string        `json:"cluster_id" gorm:"index"`
	Status        string        `json:"status"`
	ExitCode      int           `json:"exit_code"`
	Output        string        `json:"output" gorm:"type:text"`
	Error         string        `json:"error" gorm:"type:text"`
	ExecutionTime time.Duration `json:"execution_time"`
	Timestamp     time.Time     `json:"timestamp" gorm:"index"`
}

// Cluster represents a managed Kubernetes cluster
type Cluster struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index;not null"`
	Description string                 `json:"description"`
	Environment string                 `json:"environment" gorm:"index"` // dev, staging, prod
	Region      string                 `json:"region"`
	Provider    string                 `json:"provider"` // eks, gke, aks, onprem
	Status      ClusterStatus          `json:"status" gorm:"index"`
	Health      ClusterHealth          `json:"health" gorm:"index"`
	Version     string                 `json:"version"`
	AgentCount  int                    `json:"agent_count"`
	NodeCount   int                    `json:"node_count"`
	PodCount    int                    `json:"pod_count"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ClusterStatus represents the status of a cluster
type ClusterStatus string

const (
	ClusterStatusActive      ClusterStatus = "active"
	ClusterStatusInactive    ClusterStatus = "inactive"
	ClusterStatusMaintenance ClusterStatus = "maintenance"
	ClusterStatusError       ClusterStatus = "error"
)

// ClusterHealth represents the health of a cluster
type ClusterHealth string

const (
	ClusterHealthHealthy   ClusterHealth = "healthy"
	ClusterHealthDegraded  ClusterHealth = "degraded"
	ClusterHealthUnhealthy ClusterHealth = "unhealthy"
	ClusterHealthUnknown   ClusterHealth = "unknown"
)

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index;not null"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled" gorm:"index"`
	Severity    string                 `json:"severity" gorm:"index"`
	Conditions  map[string]interface{} `json:"conditions" gorm:"type:jsonb"`
	Actions     []string               `json:"actions" gorm:"type:jsonb"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	RuleID      string                 `json:"rule_id" gorm:"index"`
	ClusterID   string                 `json:"cluster_id" gorm:"index"`
	Severity    string                 `json:"severity" gorm:"index"`
	Status      AlertStatus            `json:"status" gorm:"index"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context" gorm:"type:jsonb"`
	FiredAt     time.Time              `json:"fired_at" gorm:"index"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// InternalEvent represents an event published to the internal event bus
type InternalEvent struct {
	Type      string                 `json:"type"`
	ClusterID string                 `json:"cluster_id"`
	Severity  string                 `json:"severity"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// InternalEventType represents different types of internal events
type InternalEventType string

const (
	InternalEventTypeCritical      InternalEventType = "critical"
	InternalEventTypeAnomaly       InternalEventType = "anomaly"
	InternalEventTypeSLOBreach     InternalEventType = "slo_breach"
	InternalEventTypeCommandResult InternalEventType = "command_result"
	InternalEventTypeMetricsAlert  InternalEventType = "metrics_alert"
)

// HealthStatus represents the health status of the agent-manager
type HealthStatus struct {
	Status           string                 `json:"status"`
	Version          string                 `json:"version"`
	Uptime           time.Duration          `json:"uptime"`
	ActiveAgents     int                    `json:"active_agents"`
	TotalClusters    int                    `json:"total_clusters"`
	EventsProcessed  int64                  `json:"events_processed"`
	MetricsProcessed int64                  `json:"metrics_processed"`
	CommandsIssued   int64                  `json:"commands_issued"`
	Components       map[string]interface{} `json:"components"`
	Timestamp        time.Time              `json:"timestamp"`
}

// Config represents the agent-manager configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	NATS     NATSConfig     `yaml:"nats"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	GracefulStop time.Duration `yaml:"graceful_stop"`
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL             string        `yaml:"url"`
	ClusterID       string        `yaml:"cluster_id"`
	MaxReconnect    int           `yaml:"max_reconnect"`
	ReconnectWait   time.Duration `yaml:"reconnect_wait"`
	PingInterval    time.Duration `yaml:"ping_interval"`
	MaxPingsOut     int           `yaml:"max_pings_out"`
	EnableJetStream bool          `yaml:"enable_jetstream"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"` // json, console
	OutputPath string `yaml:"output_path"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
	Port    int    `yaml:"port"`
}