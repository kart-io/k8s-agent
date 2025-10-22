package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_service_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Authentication metrics
	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"type", "result"}, // type: jwt|apikey, result: success|failure
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation"}, // operation: select|insert|update|delete
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_service_db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Cache metrics
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"}, // cache_type: permission|role
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	// Active users gauge
	activeUsersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_service_active_users",
			Help: "Number of currently active users",
		},
	)

	// API key metrics
	apiKeyValidationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_service_apikey_validations_total",
			Help: "Total number of API key validations",
		},
		[]string{"result"}, // result: success|failure|expired
	)
)

// PrometheusMiddleware records HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// RecordAuthAttempt records authentication attempt
func RecordAuthAttempt(authType string, success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	authAttemptsTotal.WithLabelValues(authType, result).Inc()
}

// RecordDBQuery records database query
func RecordDBQuery(operation string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation).Inc()
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCacheHit records cache hit
func RecordCacheHit(cacheType string) {
	cacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records cache miss
func RecordCacheMiss(cacheType string) {
	cacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(count float64) {
	activeUsersGauge.Set(count)
}

// RecordAPIKeyValidation records API key validation result
func RecordAPIKeyValidation(result string) {
	apiKeyValidationsTotal.WithLabelValues(result).Inc()
}
