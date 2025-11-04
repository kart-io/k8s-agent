package fixtures

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// AgentBuilder helps build test agent data.
type AgentBuilder struct {
	id          string
	clusterID   string
	name        string
	version     string
	status      string
	lastSeen    time.Time
	labels      map[string]string
	annotations map[string]string
}

// NewAgentBuilder creates a new agent builder with defaults.
func NewAgentBuilder() *AgentBuilder {
	return &AgentBuilder{
		id:          "test-agent-001",
		clusterID:   "test-cluster-001",
		name:        "test-agent",
		version:     "v1.0.0",
		status:      "online",
		lastSeen:    time.Now(),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}
}

// WithID sets the agent ID.
func (b *AgentBuilder) WithID(id string) *AgentBuilder {
	b.id = id
	return b
}

// WithClusterID sets the cluster ID.
func (b *AgentBuilder) WithClusterID(clusterID string) *AgentBuilder {
	b.clusterID = clusterID
	return b
}

// WithName sets the agent name.
func (b *AgentBuilder) WithName(name string) *AgentBuilder {
	b.name = name
	return b
}

// WithVersion sets the agent version.
func (b *AgentBuilder) WithVersion(version string) *AgentBuilder {
	b.version = version
	return b
}

// WithStatus sets the agent status.
func (b *AgentBuilder) WithStatus(status string) *AgentBuilder {
	b.status = status
	return b
}

// WithLastSeen sets the last seen timestamp.
func (b *AgentBuilder) WithLastSeen(lastSeen time.Time) *AgentBuilder {
	b.lastSeen = lastSeen
	return b
}

// WithLabel adds a label.
func (b *AgentBuilder) WithLabel(key, value string) *AgentBuilder {
	b.labels[key] = value
	return b
}

// WithAnnotation adds an annotation.
func (b *AgentBuilder) WithAnnotation(key, value string) *AgentBuilder {
	b.annotations[key] = value
	return b
}

// Build returns the built agent data (can be adapted to actual proto message).
func (b *AgentBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"id":          b.id,
		"cluster_id":  b.clusterID,
		"name":        b.name,
		"version":     b.version,
		"status":      b.status,
		"last_seen":   timestamppb.New(b.lastSeen),
		"labels":      b.labels,
		"annotations": b.annotations,
	}
}

// WorkflowBuilder helps build test workflow data.
type WorkflowBuilder struct {
	id          string
	name        string
	description string
	strategy    string
	steps       []map[string]interface{}
	enabled     bool
	createdAt   time.Time
}

// NewWorkflowBuilder creates a new workflow builder with defaults.
func NewWorkflowBuilder() *WorkflowBuilder {
	return &WorkflowBuilder{
		id:          "test-workflow-001",
		name:        "test-workflow",
		description: "Test workflow",
		strategy:    "sequential",
		steps:       make([]map[string]interface{}, 0),
		enabled:     true,
		createdAt:   time.Now(),
	}
}

// WithID sets the workflow ID.
func (b *WorkflowBuilder) WithID(id string) *WorkflowBuilder {
	b.id = id
	return b
}

// WithName sets the workflow name.
func (b *WorkflowBuilder) WithName(name string) *WorkflowBuilder {
	b.name = name
	return b
}

// WithDescription sets the workflow description.
func (b *WorkflowBuilder) WithDescription(description string) *WorkflowBuilder {
	b.description = description
	return b
}

// WithStrategy sets the execution strategy.
func (b *WorkflowBuilder) WithStrategy(strategy string) *WorkflowBuilder {
	b.strategy = strategy
	return b
}

// AddStep adds a step to the workflow.
func (b *WorkflowBuilder) AddStep(stepType, name string, config map[string]interface{}) *WorkflowBuilder {
	b.steps = append(b.steps, map[string]interface{}{
		"type":   stepType,
		"name":   name,
		"config": config,
	})
	return b
}

// WithEnabled sets the enabled status.
func (b *WorkflowBuilder) WithEnabled(enabled bool) *WorkflowBuilder {
	b.enabled = enabled
	return b
}

// Build returns the built workflow data.
func (b *WorkflowBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"id":          b.id,
		"name":        b.name,
		"description": b.description,
		"strategy":    b.strategy,
		"steps":       b.steps,
		"enabled":     b.enabled,
		"created_at":  timestamppb.New(b.createdAt),
	}
}

// AnalysisRequestBuilder helps build test analysis request data.
type AnalysisRequestBuilder struct {
	incidentID  string
	clusterID   string
	severity    string
	events      []map[string]interface{}
	logs        []string
	metrics     map[string]interface{}
	context     map[string]string
	requestedAt time.Time
}

// NewAnalysisRequestBuilder creates a new analysis request builder.
func NewAnalysisRequestBuilder() *AnalysisRequestBuilder {
	return &AnalysisRequestBuilder{
		incidentID:  "test-incident-001",
		clusterID:   "test-cluster-001",
		severity:    "high",
		events:      make([]map[string]interface{}, 0),
		logs:        make([]string, 0),
		metrics:     make(map[string]interface{}),
		context:     make(map[string]string),
		requestedAt: time.Now(),
	}
}

// WithIncidentID sets the incident ID.
func (b *AnalysisRequestBuilder) WithIncidentID(id string) *AnalysisRequestBuilder {
	b.incidentID = id
	return b
}

// WithClusterID sets the cluster ID.
func (b *AnalysisRequestBuilder) WithClusterID(id string) *AnalysisRequestBuilder {
	b.clusterID = id
	return b
}

// WithSeverity sets the severity.
func (b *AnalysisRequestBuilder) WithSeverity(severity string) *AnalysisRequestBuilder {
	b.severity = severity
	return b
}

// AddEvent adds an event.
func (b *AnalysisRequestBuilder) AddEvent(eventType, message string) *AnalysisRequestBuilder {
	b.events = append(b.events, map[string]interface{}{
		"type":    eventType,
		"message": message,
	})
	return b
}

// AddLog adds a log line.
func (b *AnalysisRequestBuilder) AddLog(log string) *AnalysisRequestBuilder {
	b.logs = append(b.logs, log)
	return b
}

// WithMetric adds a metric.
func (b *AnalysisRequestBuilder) WithMetric(name string, value interface{}) *AnalysisRequestBuilder {
	b.metrics[name] = value
	return b
}

// WithContext adds context.
func (b *AnalysisRequestBuilder) WithContext(key, value string) *AnalysisRequestBuilder {
	b.context[key] = value
	return b
}

// Build returns the built analysis request data.
func (b *AnalysisRequestBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"incident_id":  b.incidentID,
		"cluster_id":   b.clusterID,
		"severity":     b.severity,
		"events":       b.events,
		"logs":         b.logs,
		"metrics":      b.metrics,
		"context":      b.context,
		"requested_at": timestamppb.New(b.requestedAt),
	}
}
