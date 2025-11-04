package types

import "time"

// MetricsSummary 监控概览.
type MetricsSummary struct {
	TotalAgents     int       `json:"total_agents"`
	OnlineAgents    int       `json:"online_agents"`
	OfflineAgents   int       `json:"offline_agents"`
	TotalEvents     int       `json:"total_events"`
	CriticalEvents  int       `json:"critical_events"`
	TotalCommands   int       `json:"total_commands"`
	RunningCommands int       `json:"running_commands"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	DiskUsage       float64   `json:"disk_usage"`
	NetworkIn       float64   `json:"network_in"`
	NetworkOut      float64   `json:"network_out"`
	LastUpdateTime  time.Time `json:"last_update_time"`
}

// AgentMetrics Agent 指标.
type AgentMetrics struct {
	AgentID       string    `json:"agent_id"`
	AgentName     string    `json:"agent_name"`
	ClusterID     string    `json:"cluster_id"`
	Status        string    `json:"status"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	EventCount    int       `json:"event_count"`
	CommandCount  int       `json:"command_count"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Uptime        int64     `json:"uptime"` // seconds
}

// EventMetrics 事件指标.
type EventMetrics struct {
	EventType      string    `json:"event_type"`
	Severity       string    `json:"severity"`
	Count          int       `json:"count"`
	LastOccurrence time.Time `json:"last_occurrence"`
}

// TrendData 趋势数据.
type TrendData struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// Alert 告警规则.
type Alert struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Enabled     bool             `json:"enabled"`
	RuleType    string           `json:"rule_type"` // threshold, anomaly, pattern
	Conditions  []AlertCondition `json:"conditions"`
	Channels    []string         `json:"channels"` // email, webhook, slack
	Severity    string           `json:"severity"` // critical, warning, info
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AlertCondition 告警条件.
type AlertCondition struct {
	Metric    string      `json:"metric"`
	Operator  string      `json:"operator"` // gt, lt, eq, gte, lte
	Threshold interface{} `json:"threshold"`
	Duration  string      `json:"duration"` // 持续时间
}

// AlertHistory 告警历史.
type AlertHistory struct {
	ID          string     `json:"id"`
	AlertID     string     `json:"alert_id"`
	AlertName   string     `json:"alert_name"`
	Status      string     `json:"status"` // triggered, resolved, silenced
	Message     string     `json:"message"`
	TriggeredAt time.Time  `json:"triggered_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// DashboardOverview 仪表盘概览.
type DashboardOverview struct {
	Summary      MetricsSummary `json:"summary"`
	RecentAlerts []AlertHistory `json:"recent_alerts"`
	TopAgents    []AgentMetrics `json:"top_agents"`
	EventStats   []EventMetrics `json:"event_stats"`
	Trends       []TrendData    `json:"trends"`
}

// ChartData 图表数据.
type ChartData struct {
	Title  string                   `json:"title"`
	Type   string                   `json:"type"` // line, bar, pie, gauge
	Labels []string                 `json:"labels"`
	Series []map[string]interface{} `json:"series"`
}
