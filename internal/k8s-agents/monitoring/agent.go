// Package monitoring provides K8s-specific monitoring agent implementation
package monitoring

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
)

// K8sMonitoringAgent is a Kubernetes-specific agent for monitoring and alerting
type K8sMonitoringAgent struct {
	*core.BaseAgent

	k8sClient        kubernetes.Interface
	logger           *zap.Logger
	namespace        string
	metricsCollector *MetricsCollector
	alertManager     *AlertManager
	checkInterval    time.Duration
}

// MetricType defines types of metrics to monitor
type MetricType string

// MetricType constants define the types of metrics that can be monitored
const (
	MetricCPU        MetricType = "cpu"
	MetricMemory     MetricType = "memory"
	MetricDisk       MetricType = "disk"
	MetricNetwork    MetricType = "network"
	MetricPodStatus  MetricType = "pod_status"
	MetricNodeStatus MetricType = "node_status"
	MetricAPILatency MetricType = "api_latency"
	MetricErrorRate  MetricType = "error_rate"
)

// Alert represents a monitoring alert
type Alert struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Severity    AlertSeverity     `json:"severity"`
	MetricType  MetricType        `json:"metric_type"`
	Resource    string            `json:"resource"`
	Namespace   string            `json:"namespace"`
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	TriggeredAt time.Time         `json:"triggered_at"`
}

// AlertSeverity defines alert severity levels
type AlertSeverity string

// AlertSeverity constants define the severity levels for alerts
const (
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityInfo     AlertSeverity = "info"
)

// MonitoringResult represents the result of monitoring
type MonitoringResult struct {
	Timestamp    time.Time                        `json:"timestamp"`
	Metrics      map[MetricType][]MetricDataPoint `json:"metrics"`
	Alerts       []Alert                          `json:"alerts"`
	HealthStatus HealthStatus                     `json:"health_status"`
	Summary      string                           `json:"summary"`
}

// MetricDataPoint represents a single metric data point
type MetricDataPoint struct {
	Resource  string            `json:"resource"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// HealthStatus represents overall health status
type HealthStatus struct {
	Status     string          `json:"status"` // healthy, degraded, unhealthy
	Score      float64         `json:"score"`  // 0-100
	Components map[string]bool `json:"components"`
	Issues     []string        `json:"issues"`
}

// MetricsCollector collects metrics from various sources
type MetricsCollector struct {
	prometheusEndpoint string
	logger             *zap.Logger
}

// AlertManager manages alerts
type AlertManager struct {
	alerts map[string]*Alert
	rules  []AlertRule
	logger *zap.Logger
}

// AlertRule defines a rule for generating alerts
type AlertRule struct {
	Name       string
	MetricType MetricType
	Condition  string
	Threshold  float64
	Duration   time.Duration
	Severity   AlertSeverity
}

// NewK8sMonitoringAgent creates a new K8s monitoring agent
func NewK8sMonitoringAgent(k8sClient kubernetes.Interface, namespace string, logger *zap.Logger) *K8sMonitoringAgent {
	baseAgent := core.NewBaseAgent(
		"K8s Monitoring Agent",
		"Continuous monitoring and alerting for Kubernetes",
		[]string{"monitor", "alert", "metric-collection", "health-check"},
	)

	return &K8sMonitoringAgent{
		BaseAgent:        baseAgent,
		k8sClient:        k8sClient,
		logger:           logger,
		namespace:        namespace,
		metricsCollector: NewMetricsCollector("", logger),
		alertManager:     NewAlertManager(logger),
		checkInterval:    30 * time.Second,
	}
}

// Initialize initializes the monitoring agent
func (a *K8sMonitoringAgent) Initialize(ctx context.Context) error {
	// Setup default alert rules
	a.setupDefaultAlertRules()

	// Register monitoring tools
	if err := a.registerMonitoringTools(); err != nil {
		return fmt.Errorf("failed to register monitoring tools: %w", err)
	}

	a.logger.Info("K8s monitoring agent initialized",
		zap.String("namespace", a.namespace),
		zap.Duration("check_interval", a.checkInterval))

	return nil
}

// StartMonitoring starts continuous monitoring
func (a *K8sMonitoringAgent) StartMonitoring(ctx context.Context) error {
	ticker := time.NewTicker(a.checkInterval)
	defer ticker.Stop()

	a.logger.Info("Starting continuous monitoring")

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Monitoring stopped")
			return ctx.Err()

		case <-ticker.C:
			result, err := a.performMonitoringCycle(ctx)
			if err != nil {
				a.logger.Error("Monitoring cycle failed", zap.Error(err))
				continue
			}

			// Process results
			a.processMonitoringResult(ctx, result)
		}
	}
}

// PerformCheck performs a single monitoring check
func (a *K8sMonitoringAgent) PerformCheck(ctx context.Context) (*MonitoringResult, error) {
	return a.performMonitoringCycle(ctx)
}

// GetCurrentAlerts returns current active alerts
func (a *K8sMonitoringAgent) GetCurrentAlerts() []Alert {
	return a.alertManager.GetActiveAlerts()
}

// Private methods

func (a *K8sMonitoringAgent) performMonitoringCycle(ctx context.Context) (*MonitoringResult, error) {
	startTime := time.Now()

	result := &MonitoringResult{
		Timestamp: startTime,
		Metrics:   make(map[MetricType][]MetricDataPoint),
		Alerts:    make([]Alert, 0),
		HealthStatus: HealthStatus{
			Components: make(map[string]bool),
			Issues:     make([]string, 0),
		},
	}

	// Collect pod metrics
	podMetrics, err := a.collectPodMetrics(ctx)
	if err != nil {
		a.logger.Warn("Failed to collect pod metrics", zap.Error(err))
	} else {
		result.Metrics[MetricPodStatus] = podMetrics
	}

	// Collect node metrics
	nodeMetrics, err := a.collectNodeMetrics(ctx)
	if err != nil {
		a.logger.Warn("Failed to collect node metrics", zap.Error(err))
	} else {
		result.Metrics[MetricNodeStatus] = nodeMetrics
	}

	// Collect resource metrics (CPU, Memory)
	resourceMetrics, err := a.metricsCollector.CollectResourceMetrics(ctx)
	if err != nil {
		a.logger.Warn("Failed to collect resource metrics", zap.Error(err))
	} else {
		result.Metrics[MetricCPU] = resourceMetrics.CPU
		result.Metrics[MetricMemory] = resourceMetrics.Memory
	}

	// Check for alerts
	alerts := a.alertManager.EvaluateRules(result.Metrics)
	result.Alerts = alerts

	// Calculate health status
	result.HealthStatus = a.calculateHealthStatus(result)

	// Generate summary
	result.Summary = a.generateSummary(result)

	// Store in memory for historical analysis
	a.storeMonitoringResult(ctx, result)

	return result, nil
}

func (a *K8sMonitoringAgent) collectPodMetrics(ctx context.Context) ([]MetricDataPoint, error) {
	metrics := make([]MetricDataPoint, 0)

	// List pods
	pods, err := a.k8sClient.CoreV1().Pods(a.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		status := 0.0
		if pod.Status.Phase == v1.PodRunning {
			status = 1.0
		}

		metric := MetricDataPoint{
			Resource: pod.Name,
			Value:    status,
			Unit:     "status",
			Labels: map[string]string{
				"namespace": pod.Namespace,
				"phase":     string(pod.Status.Phase),
			},
			Timestamp: time.Now(),
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (a *K8sMonitoringAgent) collectNodeMetrics(ctx context.Context) ([]MetricDataPoint, error) {
	metrics := make([]MetricDataPoint, 0)

	// List nodes
	nodes, err := a.k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		ready := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				ready = true
				break
			}
		}

		value := 0.0
		if ready {
			value = 1.0
		}

		metric := MetricDataPoint{
			Resource: node.Name,
			Value:    value,
			Unit:     "status",
			Labels: map[string]string{
				"ready": fmt.Sprintf("%v", ready),
			},
			Timestamp: time.Now(),
		}

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (a *K8sMonitoringAgent) processMonitoringResult(ctx context.Context, result *MonitoringResult) {
	// Log summary
	a.logger.Info("Monitoring cycle completed",
		zap.String("health_status", result.HealthStatus.Status),
		zap.Float64("health_score", result.HealthStatus.Score),
		zap.Int("alerts", len(result.Alerts)))

	// Process alerts
	for _, alert := range result.Alerts {
		a.handleAlert(ctx, alert)
	}

	// Create task for any critical issues
	if result.HealthStatus.Status == "unhealthy" {
		// TODO: Implement task creation once Task type is available
		// task := core.Task{
		// 	ID:          fmt.Sprintf("investigate-%d", time.Now().Unix()),
		// 	Name:        "Investigate Cluster Health",
		// 	Description: "Cluster health is degraded, investigation needed",
		// 	Type:        core.TaskTypeAnalysis,
		// 	Input: map[string]interface{}{
		// 		"health_status": result.HealthStatus,
		// 		"alerts":        result.Alerts,
		// 	},
		// 	Priority:  1,
		// 	CreatedAt: time.Now(),
		// }

		// if _, err := a.Execute(ctx, task); err != nil {
		// 	a.logger.Error("Failed to execute investigation task", zap.Error(err))
		// }
	}
}

func (a *K8sMonitoringAgent) handleAlert(ctx context.Context, alert Alert) {
	a.logger.Warn("Alert triggered",
		zap.String("alert_name", alert.Name),
		zap.String("severity", string(alert.Severity)),
		zap.String("message", alert.Message))

	// Store alert in memory
	// TODO: Implement memory storage once Memory type is available
	// memory := core.Memory{
	// 	Type:      core.MemoryEpisodic,
	// 	Content:   alert,
	// 	Tags:      []string{"alert", string(alert.Severity), string(alert.MetricType)},
	// 	Relevance: a.calculateAlertRelevance(alert.Severity),
	// 	Timestamp: alert.TriggeredAt,
	// }

	// if err := a.Remember(ctx, memory); err != nil {
	// 	a.logger.Warn("Failed to store alert in memory", zap.Error(err))
	// }

	// For critical alerts, trigger immediate action
	if alert.Severity == AlertSeverityCritical {
		// Could trigger remediation agent here
		a.logger.Error("Critical alert requires immediate action",
			zap.String("alert_id", alert.ID))
	}
}

func (a *K8sMonitoringAgent) calculateHealthStatus(result *MonitoringResult) HealthStatus {
	healthScore := 100.0
	issues := make([]string, 0)
	components := make(map[string]bool)

	// Check pod health
	if podMetrics, exists := result.Metrics[MetricPodStatus]; exists {
		healthyPods := 0
		for _, metric := range podMetrics {
			if metric.Value == 1.0 {
				healthyPods++
			}
		}

		podHealth := float64(healthyPods) / float64(len(podMetrics))
		components["pods"] = podHealth > 0.8

		if podHealth < 0.8 {
			healthScore -= 20
			issues = append(issues, fmt.Sprintf("%.0f%% pods unhealthy", (1-podHealth)*100))
		}
	}

	// Check node health
	if nodeMetrics, exists := result.Metrics[MetricNodeStatus]; exists {
		healthyNodes := 0
		for _, metric := range nodeMetrics {
			if metric.Value == 1.0 {
				healthyNodes++
			}
		}

		nodeHealth := float64(healthyNodes) / float64(len(nodeMetrics))
		components["nodes"] = nodeHealth == 1.0

		if nodeHealth < 1.0 {
			healthScore -= 30
			issues = append(issues, fmt.Sprintf("%.0f%% nodes unhealthy", (1-nodeHealth)*100))
		}
	}

	// Factor in alerts
	for _, alert := range result.Alerts {
		switch alert.Severity {
		case AlertSeverityCritical:
			healthScore -= 20
		case AlertSeverityWarning:
			healthScore -= 10
		}
	}

	// Determine overall status
	status := "healthy"
	if healthScore < 50 {
		status = "unhealthy"
	} else if healthScore < 80 {
		status = "degraded"
	}

	return HealthStatus{
		Status:     status,
		Score:      healthScore,
		Components: components,
		Issues:     issues,
	}
}

func (a *K8sMonitoringAgent) generateSummary(result *MonitoringResult) string {
	totalPods := len(result.Metrics[MetricPodStatus])
	totalNodes := len(result.Metrics[MetricNodeStatus])
	totalAlerts := len(result.Alerts)

	return fmt.Sprintf("Health: %s (%.0f%%), Pods: %d, Nodes: %d, Alerts: %d",
		result.HealthStatus.Status,
		result.HealthStatus.Score,
		totalPods,
		totalNodes,
		totalAlerts)
}

func (a *K8sMonitoringAgent) storeMonitoringResult(ctx context.Context, result *MonitoringResult) {
	// TODO: Implement memory storage once Memory type is available
	// memory := core.Memory{
	// 	Type:    core.MemoryShortTerm,
	// 	Content: result,
	// 	Tags:    []string{"monitoring", "metrics", result.HealthStatus.Status},
	// 	Relevance: func() float64 {
	// 		if result.HealthStatus.Status == "unhealthy" {
	// 			return 1.0
	// 		} else if result.HealthStatus.Status == "degraded" {
	// 			return 0.7
	// 		}
	// 		return 0.3
	// 	}(),
	// 	Timestamp: result.Timestamp,
	// 	TTL: func() *time.Duration {
	// 		d := 24 * time.Hour
	// 		return &d
	// 	}(),
	// }

	// if err := a.Remember(ctx, memory); err != nil {
	// 	a.logger.Warn("Failed to store monitoring result", zap.Error(err))
	// }
}

func (a *K8sMonitoringAgent) setupDefaultAlertRules() {
	rules := []AlertRule{
		{
			Name:       "High CPU Usage",
			MetricType: MetricCPU,
			Condition:  "greater_than",
			Threshold:  80.0,
			Duration:   5 * time.Minute,
			Severity:   AlertSeverityWarning,
		},
		{
			Name:       "High Memory Usage",
			MetricType: MetricMemory,
			Condition:  "greater_than",
			Threshold:  90.0,
			Duration:   5 * time.Minute,
			Severity:   AlertSeverityCritical,
		},
		{
			Name:       "Pod Not Ready",
			MetricType: MetricPodStatus,
			Condition:  "equals",
			Threshold:  0.0,
			Duration:   2 * time.Minute,
			Severity:   AlertSeverityWarning,
		},
		{
			Name:       "Node Not Ready",
			MetricType: MetricNodeStatus,
			Condition:  "equals",
			Threshold:  0.0,
			Duration:   1 * time.Minute,
			Severity:   AlertSeverityCritical,
		},
	}

	a.alertManager.SetRules(rules)
}

func (a *K8sMonitoringAgent) registerMonitoringTools() error {
	// Register monitoring-specific tools with the base agent
	// This would include Prometheus queries, metric collectors, etc.
	return nil
}

// calculateAlertRelevance calculates the relevance score for an alert
//
//nolint:unused // This will be used in future alert prioritization
func (a *K8sMonitoringAgent) calculateAlertRelevance(severity AlertSeverity) float64 {
	switch severity {
	case AlertSeverityCritical:
		return 1.0
	case AlertSeverityWarning:
		return 0.7
	case AlertSeverityInfo:
		return 0.3
	default:
		return 0.1
	}
}

// MetricsCollector implementation

func NewMetricsCollector(prometheusEndpoint string, logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		prometheusEndpoint: prometheusEndpoint,
		logger:             logger,
	}
}

type ResourceMetrics struct {
	CPU    []MetricDataPoint
	Memory []MetricDataPoint
}

func (c *MetricsCollector) CollectResourceMetrics(ctx context.Context) (*ResourceMetrics, error) {
	// This would actually query Prometheus or metrics-server
	// For now, return mock data
	return &ResourceMetrics{
		CPU:    make([]MetricDataPoint, 0),
		Memory: make([]MetricDataPoint, 0),
	}, nil
}

// AlertManager implementation

func NewAlertManager(logger *zap.Logger) *AlertManager {
	return &AlertManager{
		alerts: make(map[string]*Alert),
		rules:  make([]AlertRule, 0),
		logger: logger,
	}
}

func (m *AlertManager) SetRules(rules []AlertRule) {
	m.rules = rules
}

func (m *AlertManager) EvaluateRules(metrics map[MetricType][]MetricDataPoint) []Alert {
	alerts := make([]Alert, 0)

	// Evaluate each rule against metrics
	// This is a simplified implementation
	for _, rule := range m.rules {
		if metricData, exists := metrics[rule.MetricType]; exists {
			for _, dataPoint := range metricData {
				if m.evaluateCondition(dataPoint.Value, rule.Condition, rule.Threshold) {
					alert := Alert{
						ID:          fmt.Sprintf("alert-%d", time.Now().UnixNano()),
						Name:        rule.Name,
						Severity:    rule.Severity,
						MetricType:  rule.MetricType,
						Resource:    dataPoint.Resource,
						Value:       dataPoint.Value,
						Threshold:   rule.Threshold,
						Message:     fmt.Sprintf("%s: %.2f exceeds threshold %.2f", rule.Name, dataPoint.Value, rule.Threshold),
						Labels:      dataPoint.Labels,
						TriggeredAt: time.Now(),
					}
					alerts = append(alerts, alert)
				}
			}
		}
	}

	return alerts
}

func (m *AlertManager) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case "greater_than":
		return value > threshold
	case "less_than":
		return value < threshold
	case "equals":
		return value == threshold
	default:
		return false
	}
}

func (m *AlertManager) GetActiveAlerts() []Alert {
	alerts := make([]Alert, 0)
	for _, alert := range m.alerts {
		alerts = append(alerts, *alert)
	}
	return alerts
}

// Execute implements the core.Agent interface
func (a *K8sMonitoringAgent) Execute(ctx context.Context, input *core.AgentInput) (*core.AgentOutput, error) {
	// Extract task from input
	task := input.Task
	if task == "" {
		task = "monitor"
	}

	var result interface{}
	var err error

	switch task {
	case "monitor":
		result, err = a.performMonitoringCycle(ctx)
		if err != nil {
			return &core.AgentOutput{
				Status:  "failed",
				Message: fmt.Sprintf("Monitoring failed: %v", err),
			}, err
		}

	case "get_alerts":
		result = a.GetCurrentAlerts()

	case "start":
		go a.StartMonitoring(ctx)
		result = "Monitoring started"

	default:
		return &core.AgentOutput{
			Status:  "failed",
			Message: fmt.Sprintf("Unknown task: %s", task),
		}, fmt.Errorf("unknown task: %s", task)
	}

	return &core.AgentOutput{
		Status:  "success",
		Result:  result,
		Message: fmt.Sprintf("Task %s completed successfully", task),
	}, nil
}
