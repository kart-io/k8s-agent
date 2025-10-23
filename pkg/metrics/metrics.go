package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Common metric names
const (
	// HTTP metrics
	HTTPRequestsTotal   = "http_requests_total"
	HTTPRequestDuration = "http_request_duration_seconds"
	HTTPResponseSize    = "http_response_size_bytes"
	HTTPRequestSize     = "http_request_size_bytes"

	// gRPC metrics
	GRPCRequestsTotal   = "grpc_requests_total"
	GRPCRequestDuration = "grpc_request_duration_seconds"

	// Database metrics
	DBQueriesTotal    = "db_queries_total"
	DBQueryDuration   = "db_query_duration_seconds"
	DBConnectionsOpen = "db_connections_open"
	DBConnectionsIdle = "db_connections_idle"

	// Cache metrics
	CacheHitsTotal   = "cache_hits_total"
	CacheMissesTotal = "cache_misses_total"
	CacheLatency     = "cache_latency_seconds"

	// Message queue metrics
	MQMessagesPublished = "mq_messages_published_total"
	MQMessagesConsumed  = "mq_messages_consumed_total"
	MQMessagesDLQ       = "mq_messages_dlq_total"
	MQPublishDuration   = "mq_publish_duration_seconds"
	MQConsumeDuration   = "mq_consume_duration_seconds"

	// Business metrics
	EventsProcessed   = "events_processed_total"
	WorkflowsExecuted = "workflows_executed_total"
	TasksCompleted    = "tasks_completed_total"
	ErrorsTotal       = "errors_total"

	// System metrics
	GoroutinesActive = "goroutines_active"
	MemoryAllocated  = "memory_allocated_bytes"
	CPUUsage         = "cpu_usage_percent"
)

// Common label names
const (
	LabelMethod    = "method"
	LabelPath      = "path"
	LabelStatus    = "status"
	LabelService   = "service"
	LabelOperation = "operation"
	LabelResult    = "result"
	LabelErrorType = "error_type"
	LabelCluster   = "cluster"
	LabelNamespace = "namespace"
	LabelResource  = "resource"
	LabelEventType = "event_type"
	LabelWorkflow  = "workflow"
	LabelTask      = "task"
)

// HTTPMetrics manages HTTP-related metrics.
type HTTPMetrics struct {
	registry        *Registry
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
}

// NewHTTPMetrics creates HTTP metrics.
func NewHTTPMetrics(registry *Registry) *HTTPMetrics {
	return &HTTPMetrics{
		registry: registry,
		requestsTotal: registry.Counter(
			HTTPRequestsTotal,
			"Total number of HTTP requests",
			[]string{LabelMethod, LabelPath, LabelStatus},
		),
		requestDuration: registry.Histogram(
			HTTPRequestDuration,
			"HTTP request duration in seconds",
			[]string{LabelMethod, LabelPath, LabelStatus},
			[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		),
		responseSize: registry.Histogram(
			HTTPResponseSize,
			"HTTP response size in bytes",
			[]string{LabelMethod, LabelPath, LabelStatus},
			[]float64{100, 1000, 10000, 100000, 1000000, 10000000},
		),
		requestSize: registry.Histogram(
			HTTPRequestSize,
			"HTTP request size in bytes",
			[]string{LabelMethod, LabelPath},
			[]float64{100, 1000, 10000, 100000, 1000000, 10000000},
		),
	}
}

// RecordRequest records an HTTP request.
func (h *HTTPMetrics) RecordRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	labels := prometheus.Labels{
		LabelMethod: method,
		LabelPath:   path,
		LabelStatus: statusCodeToString(statusCode),
	}

	h.requestsTotal.With(labels).Inc()
	h.requestDuration.With(labels).Observe(duration.Seconds())
	h.responseSize.With(labels).Observe(float64(responseSize))

	requestLabels := prometheus.Labels{
		LabelMethod: method,
		LabelPath:   path,
	}
	h.requestSize.With(requestLabels).Observe(float64(requestSize))
}

// DatabaseMetrics manages database-related metrics.
type DatabaseMetrics struct {
	registry        *Registry
	queriesTotal    *prometheus.CounterVec
	queryDuration   *prometheus.HistogramVec
	connectionsOpen *prometheus.GaugeVec
	connectionsIdle *prometheus.GaugeVec
}

// NewDatabaseMetrics creates database metrics.
func NewDatabaseMetrics(registry *Registry) *DatabaseMetrics {
	return &DatabaseMetrics{
		registry: registry,
		queriesTotal: registry.Counter(
			DBQueriesTotal,
			"Total number of database queries",
			[]string{LabelOperation, LabelResult},
		),
		queryDuration: registry.Histogram(
			DBQueryDuration,
			"Database query duration in seconds",
			[]string{LabelOperation, LabelResult},
			[]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		),
		connectionsOpen: registry.Gauge(
			DBConnectionsOpen,
			"Number of open database connections",
			[]string{},
		),
		connectionsIdle: registry.Gauge(
			DBConnectionsIdle,
			"Number of idle database connections",
			[]string{},
		),
	}
}

// RecordQuery records a database query.
func (d *DatabaseMetrics) RecordQuery(operation string, duration time.Duration, err error) {
	result := "success"
	if err != nil {
		result = "error"
	}

	labels := prometheus.Labels{
		LabelOperation: operation,
		LabelResult:    result,
	}

	d.queriesTotal.With(labels).Inc()
	d.queryDuration.With(labels).Observe(duration.Seconds())
}

// SetConnections updates connection pool metrics.
func (d *DatabaseMetrics) SetConnections(open, idle int) {
	d.connectionsOpen.With(prometheus.Labels{}).Set(float64(open))
	d.connectionsIdle.With(prometheus.Labels{}).Set(float64(idle))
}

// CacheMetrics manages cache-related metrics.
type CacheMetrics struct {
	registry    *Registry
	hitsTotal   *prometheus.CounterVec
	missesTotal *prometheus.CounterVec
	latency     *prometheus.HistogramVec
}

// NewCacheMetrics creates cache metrics.
func NewCacheMetrics(registry *Registry) *CacheMetrics {
	return &CacheMetrics{
		registry: registry,
		hitsTotal: registry.Counter(
			CacheHitsTotal,
			"Total number of cache hits",
			[]string{LabelOperation},
		),
		missesTotal: registry.Counter(
			CacheMissesTotal,
			"Total number of cache misses",
			[]string{LabelOperation},
		),
		latency: registry.Histogram(
			CacheLatency,
			"Cache operation latency in seconds",
			[]string{LabelOperation},
			[]float64{.0001, .0005, .001, .0025, .005, .01, .025, .05, .1},
		),
	}
}

// RecordHit records a cache hit.
func (c *CacheMetrics) RecordHit(operation string, duration time.Duration) {
	c.hitsTotal.With(prometheus.Labels{LabelOperation: operation}).Inc()
	c.latency.With(prometheus.Labels{LabelOperation: operation}).Observe(duration.Seconds())
}

// RecordMiss records a cache miss.
func (c *CacheMetrics) RecordMiss(operation string, duration time.Duration) {
	c.missesTotal.With(prometheus.Labels{LabelOperation: operation}).Inc()
	c.latency.With(prometheus.Labels{LabelOperation: operation}).Observe(duration.Seconds())
}

// BusinessMetrics manages business-related metrics.
type BusinessMetrics struct {
	registry          *Registry
	eventsProcessed   *prometheus.CounterVec
	workflowsExecuted *prometheus.CounterVec
	tasksCompleted    *prometheus.CounterVec
	errorsTotal       *prometheus.CounterVec
}

// NewBusinessMetrics creates business metrics.
func NewBusinessMetrics(registry *Registry) *BusinessMetrics {
	return &BusinessMetrics{
		registry: registry,
		eventsProcessed: registry.Counter(
			EventsProcessed,
			"Total number of events processed",
			[]string{LabelEventType, LabelCluster, LabelResult},
		),
		workflowsExecuted: registry.Counter(
			WorkflowsExecuted,
			"Total number of workflows executed",
			[]string{LabelWorkflow, LabelResult},
		),
		tasksCompleted: registry.Counter(
			TasksCompleted,
			"Total number of tasks completed",
			[]string{LabelTask, LabelResult},
		),
		errorsTotal: registry.Counter(
			ErrorsTotal,
			"Total number of errors",
			[]string{LabelService, LabelOperation, LabelErrorType},
		),
	}
}

// RecordEvent records an event processing.
func (b *BusinessMetrics) RecordEvent(eventType, cluster string, err error) {
	result := "success"
	if err != nil {
		result = "error"
	}

	b.eventsProcessed.With(prometheus.Labels{
		LabelEventType: eventType,
		LabelCluster:   cluster,
		LabelResult:    result,
	}).Inc()
}

// RecordWorkflow records a workflow execution.
func (b *BusinessMetrics) RecordWorkflow(workflow string, err error) {
	result := "success"
	if err != nil {
		result = "error"
	}

	b.workflowsExecuted.With(prometheus.Labels{
		LabelWorkflow: workflow,
		LabelResult:   result,
	}).Inc()
}

// RecordError records an error occurrence.
func (b *BusinessMetrics) RecordError(service, operation, errorType string) {
	b.errorsTotal.With(prometheus.Labels{
		LabelService:   service,
		LabelOperation: operation,
		LabelErrorType: errorType,
	}).Inc()
}

// Helper functions

func statusCodeToString(code int) string {
	if code >= 200 && code < 300 {
		return "2xx"
	} else if code >= 300 && code < 400 {
		return "3xx"
	} else if code >= 400 && code < 500 {
		return "4xx"
	} else if code >= 500 && code < 600 {
		return "5xx"
	}
	return "unknown"
}
