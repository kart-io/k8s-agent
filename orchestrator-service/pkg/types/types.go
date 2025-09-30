package types

import (
	"time"
)

// Workflow represents a diagnostic workflow
type Workflow struct {
	ID            string                 `json:"id" gorm:"primaryKey"`
	Name          string                 `json:"name" gorm:"index;not null"`
	Description   string                 `json:"description"`
	TriggerType   string                 `json:"trigger_type" gorm:"index"` // event, schedule, manual
	TriggerConfig map[string]interface{} `json:"trigger_config" gorm:"type:jsonb"`
	Steps         []WorkflowStep         `json:"steps" gorm:"type:jsonb"`
	Status        WorkflowStatus         `json:"status" gorm:"index"`
	Priority      int                    `json:"priority" gorm:"index"`
	Timeout       time.Duration          `json:"timeout"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// WorkflowStatus represents workflow status
type WorkflowStatus string

const (
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusDraft    WorkflowStatus = "draft"
)

// WorkflowStep represents a step in workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Type        StepType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Conditions  []Condition            `json:"conditions,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	OnSuccess   []string               `json:"on_success,omitempty"` // Next step IDs
	OnFailure   []string               `json:"on_failure,omitempty"` // Next step IDs
}

// StepType represents workflow step type
type StepType string

const (
	StepTypeCommand      StepType = "command"
	StepTypeAIAnalysis   StepType = "ai_analysis"
	StepTypeDecision     StepType = "decision"
	StepTypeRemediation  StepType = "remediation"
	StepTypeNotification StepType = "notification"
	StepTypeWait         StepType = "wait"
	StepTypeParallel     StepType = "parallel"
)

// Condition represents execution condition
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, matches
	Value    interface{} `json:"value"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries     int           `json:"max_retries"`
	InitialDelay   time.Duration `json:"initial_delay"`
	MaxDelay       time.Duration `json:"max_delay"`
	BackoffFactor  float64       `json:"backoff_factor"`
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID              string                 `json:"id" gorm:"primaryKey"`
	WorkflowID      string                 `json:"workflow_id" gorm:"index;not null"`
	TriggerEvent    map[string]interface{} `json:"trigger_event" gorm:"type:jsonb"`
	Status          ExecutionStatus        `json:"status" gorm:"index"`
	CurrentStepID   string                 `json:"current_step_id"`
	StepExecutions  []StepExecution        `json:"step_executions" gorm:"type:jsonb"`
	Context         map[string]interface{} `json:"context" gorm:"type:jsonb"`
	Result          map[string]interface{} `json:"result" gorm:"type:jsonb"`
	Error           string                 `json:"error,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Duration        time.Duration          `json:"duration"`
}

// ExecutionStatus represents execution status
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

// StepExecution represents a step execution
type StepExecution struct {
	StepID      string                 `json:"step_id"`
	Status      ExecutionStatus        `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
}

// Strategy represents a diagnostic strategy
type Strategy struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index;not null"`
	Category    string                 `json:"category" gorm:"index"` // pod_failure, node_issue, network, etc.
	Description string                 `json:"description"`
	Symptoms    []Symptom              `json:"symptoms" gorm:"type:jsonb"`
	WorkflowID  string                 `json:"workflow_id" gorm:"index"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled" gorm:"index"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Symptom represents a failure symptom pattern
type Symptom struct {
	Type       string                 `json:"type"` // event, metric, log
	Pattern    string                 `json:"pattern"`
	Conditions map[string]interface{} `json:"conditions"`
}

// Task represents a scheduled or queued task
type Task struct {
	ID             string                 `json:"id" gorm:"primaryKey"`
	Type           TaskType               `json:"type" gorm:"index"`
	WorkflowID     string                 `json:"workflow_id" gorm:"index"`
	ExecutionID    string                 `json:"execution_id" gorm:"index"`
	Payload        map[string]interface{} `json:"payload" gorm:"type:jsonb"`
	Status         TaskStatus             `json:"status" gorm:"index"`
	Priority       int                    `json:"priority" gorm:"index"`
	ScheduledAt    time.Time              `json:"scheduled_at" gorm:"index"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	Error          string                 `json:"error,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// TaskType represents task type
type TaskType string

const (
	TaskTypeWorkflowStart TaskType = "workflow_start"
	TaskTypeStepExecution TaskType = "step_execution"
	TaskTypeRemediation   TaskType = "remediation"
	TaskTypeNotification  TaskType = "notification"
)

// TaskStatus represents task status
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// RemediationAction represents an automated remediation action
type RemediationAction struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"index;not null"`
	Category    string                 `json:"category" gorm:"index"`
	Description string                 `json:"description"`
	ActionType  string                 `json:"action_type"` // kubectl, api_call, script
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb"`
	RiskLevel   RiskLevel              `json:"risk_level" gorm:"index"`
	RequireApproval bool               `json:"require_approval"`
	Rollback    *RollbackConfig        `json:"rollback" gorm:"type:jsonb"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RiskLevel represents risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RollbackConfig defines rollback configuration
type RollbackConfig struct {
	Enabled       bool                   `json:"enabled"`
	ActionType    string                 `json:"action_type"`
	Config        map[string]interface{} `json:"config"`
	TriggerOn     []string               `json:"trigger_on"` // failure, timeout, manual
}

// RemediationExecution represents a remediation execution
type RemediationExecution struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	ActionID    string                 `json:"action_id" gorm:"index"`
	ExecutionID string                 `json:"execution_id" gorm:"index"`
	Status      ExecutionStatus        `json:"status" gorm:"index"`
	Input       map[string]interface{} `json:"input" gorm:"type:jsonb"`
	Output      map[string]interface{} `json:"output" gorm:"type:jsonb"`
	Error       string                 `json:"error,omitempty"`
	ApprovedBy  string                 `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time             `json:"approved_at,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	RolledBack  bool                   `json:"rolled_back"`
	RollbackAt  *time.Time             `json:"rollback_at,omitempty"`
}

// AIAnalysisRequest represents an AI analysis request
type AIAnalysisRequest struct {
	ID          string                 `json:"id" gorm:"primaryKey"`
	ExecutionID string                 `json:"execution_id" gorm:"index"`
	Type        AIAnalysisType         `json:"type" gorm:"index"`
	Context     map[string]interface{} `json:"context" gorm:"type:jsonb"`
	Status      ExecutionStatus        `json:"status" gorm:"index"`
	Result      *AIAnalysisResult      `json:"result" gorm:"type:jsonb"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// AIAnalysisType represents AI analysis type
type AIAnalysisType string

const (
	AIAnalysisTypeRootCause     AIAnalysisType = "root_cause"
	AIAnalysisTypePrediction    AIAnalysisType = "prediction"
	AIAnalysisTypeRecommendation AIAnalysisType = "recommendation"
)

// AIAnalysisResult represents AI analysis result
type AIAnalysisResult struct {
	RootCause       *RootCauseAnalysis     `json:"root_cause,omitempty"`
	Recommendations []Recommendation       `json:"recommendations,omitempty"`
	Confidence      float64                `json:"confidence"`
	Evidence        []string               `json:"evidence"`
	SimilarCases    []string               `json:"similar_cases,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// RootCauseAnalysis represents root cause analysis
type RootCauseAnalysis struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Evidence    []string `json:"evidence"`
}

// Recommendation represents a recommendation
type Recommendation struct {
	Action      string                 `json:"action"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Risk        RiskLevel              `json:"risk"`
	Impact      string                 `json:"impact"`
	Steps       []string               `json:"steps"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Config represents orchestrator configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	NATS     NATSConfig     `yaml:"nats"`
	Temporal TemporalConfig `yaml:"temporal"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	AI       AIConfig       `yaml:"ai"`
	Logging  LoggingConfig  `yaml:"logging"`
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
	URL           string        `yaml:"url"`
	MaxReconnect  int           `yaml:"max_reconnect"`
	ReconnectWait time.Duration `yaml:"reconnect_wait"`
}

// TemporalConfig represents Temporal workflow engine configuration
type TemporalConfig struct {
	HostPort  string `yaml:"host_port"`
	Namespace string `yaml:"namespace"`
	TaskQueue string `yaml:"task_queue"`
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
}

// AIConfig represents AI service configuration
type AIConfig struct {
	ReasoningServiceURL string        `yaml:"reasoning_service_url"`
	Timeout             time.Duration `yaml:"timeout"`
	MaxRetries          int           `yaml:"max_retries"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
}

// InternalEvent represents an event from agent-manager
type InternalEvent struct {
	Type      string                 `json:"type"`
	ClusterID string                 `json:"cluster_id"`
	Severity  string                 `json:"severity"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}