package constants

import "time"

// Registry constants.
const (
	// HeartbeatTimeout is the duration after which an agent is considered offline
	// if no heartbeat is received (2x the expected heartbeat interval).
	HeartbeatTimeout = 60 * time.Second

	// CleanupInterval is the interval for cleaning up stale agents.
	CleanupInterval = 30 * time.Second

	// StaleAgentThreshold is the duration after which offline agents are removed.
	StaleAgentThreshold = 24 * time.Hour

	// AgentCacheTTL is the TTL for agent records in Redis cache.
	AgentCacheTTL = 30 * time.Minute

	// AgentOnlineTTL is the TTL for agent online status in Redis.
	AgentOnlineTTL = 2 * time.Minute
)

// Event processor constants.
const (
	// DuplicateEventTTL is the TTL for duplicate event detection.
	DuplicateEventTTL = 5 * time.Minute

	// MaxEventsPerGroup is the maximum number of events to keep in an aggregation group.
	MaxEventsPerGroup = 10

	// EventChannelBuffer is the buffer size for event channels.
	EventChannelBuffer = 100
)

// Command dispatcher constants.
const (
	// DefaultCommandTimeout is the default timeout for command execution.
	DefaultCommandTimeout = 30 * time.Second

	// CommandTimeoutCleanupInterval is the interval for cleaning up expired timers.
	CommandTimeoutCleanupInterval = 5 * time.Minute

	// MaxCommandRetries is the maximum number of retry attempts for failed commands.
	MaxCommandRetries = 3
)

// Database operation timeouts.
const (
	// DatabaseOperationTimeout is the timeout for database operations in background tasks.
	DatabaseOperationTimeout = 30 * time.Second

	// CacheOperationTimeout is the timeout for cache operations.
	CacheOperationTimeout = 2 * time.Second
)

// Workflow engine constants.
const (
	// DefaultStepTimeout is the default timeout for workflow step execution.
	DefaultStepTimeout = 60 * time.Second

	// WorkflowExecutionTimeout is the maximum time a workflow can run.
	WorkflowExecutionTimeout = 30 * time.Minute

	// DefaultRetryDelay is the default delay between retry attempts.
	DefaultRetryDelay = 5 * time.Second

	// MaxRetryAttempts is the default maximum number of retry attempts.
	MaxRetryAttempts = 3
)

// Allowed tools whitelist.
var AllowedTools = map[string]bool{
	"kubectl": true,
	"ps":      true,
	"df":      true,
	"netstat": true,
	"curl":    true,
	"ping":    true,
	"top":     true,
}

// AllowedKubectlActions defines safe kubectl actions.
var AllowedKubectlActions = map[string]bool{
	"get":      true,
	"describe": true,
	"logs":     true,
	"top":      true,
	"explain":  true,
	// Dangerous actions disabled by default
	// "delete": false,
	// "apply": false,
	// "create": false,
}

// AllowedCurlFlags defines safe curl flags.
var AllowedCurlFlags = map[string]bool{
	"-I":                true, // Head request
	"-s":                true, // Silent
	"--connect-timeout": true, // Connection timeout
	"-m":                true, // Max time
	"-L":                true, // Follow redirects
	// Dangerous flags disabled
	// "-H": false,  // Custom headers
	// "-d": false,  // POST data
	// "-X": false,  // Custom method
}
