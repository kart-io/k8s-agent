package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HealthServer provides HTTP health check endpoints
type HealthServer struct {
	agent  *Agent
	server *http.Server
	logger *zap.Logger
}

// NewHealthServer creates a new health server
func NewHealthServer(agent *Agent, port int, logger *zap.Logger) *HealthServer {
	mux := http.NewServeMux()
	hs := &HealthServer{
		agent:  agent,
		logger: logger.With(zap.String("component", "health-server")),
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// Register handlers
	mux.HandleFunc("/health/live", hs.handleLiveness)
	mux.HandleFunc("/health/ready", hs.handleReadiness)
	mux.HandleFunc("/health/status", hs.handleStatus)
	mux.HandleFunc("/metrics", hs.handleMetrics)

	return hs
}

// Start starts the health server
func (hs *HealthServer) Start() error {
	hs.logger.Info("Starting health server", zap.String("addr", hs.server.Addr))

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.logger.Error("Health server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the health server
func (hs *HealthServer) Stop() error {
	hs.logger.Info("Stopping health server")
	return hs.server.Close()
}

// handleLiveness handles liveness probe requests
func (hs *HealthServer) handleLiveness(w http.ResponseWriter, r *http.Request) {
	if hs.agent.IsHealthy() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not healthy"))
	}
}

// handleReadiness handles readiness probe requests
func (hs *HealthServer) handleReadiness(w http.ResponseWriter, r *http.Request) {
	if hs.agent.IsReady() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not ready"))
	}
}

// handleStatus handles detailed status requests
func (hs *HealthServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := hs.agent.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		hs.logger.Error("Failed to encode status", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// handleMetrics handles metrics endpoint for Prometheus scraping
func (hs *HealthServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	status := hs.agent.GetStatus()

	// Generate Prometheus metrics format
	metrics := fmt.Sprintf(`# HELP agent_running Agent running status (1 = running, 0 = not running)
# TYPE agent_running gauge
agent_running{cluster_id="%s"} %d

# HELP agent_connected Agent connection status (1 = connected, 0 = not connected)
# TYPE agent_connected gauge
agent_connected{cluster_id="%s"} %d

# HELP agent_uptime_seconds Agent uptime in seconds
# TYPE agent_uptime_seconds gauge
agent_uptime_seconds{cluster_id="%s"} %f

# HELP agent_event_queue_size Number of events in queue
# TYPE agent_event_queue_size gauge
agent_event_queue_size{cluster_id="%s"} %d

# HELP agent_metrics_queue_size Number of metrics in queue
# TYPE agent_metrics_queue_size gauge
agent_metrics_queue_size{cluster_id="%s"} %d

# HELP agent_command_queue_size Number of commands in queue
# TYPE agent_command_queue_size gauge
agent_command_queue_size{cluster_id="%s"} %d

# HELP agent_result_queue_size Number of results in queue
# TYPE agent_result_queue_size gauge
agent_result_queue_size{cluster_id="%s"} %d
`,
		status.ClusterID, boolToInt(status.Running),
		status.ClusterID, boolToInt(status.Connected),
		status.ClusterID, status.Uptime.Seconds(),
		status.ClusterID, status.EventQueueSize,
		status.ClusterID, status.MetricsQueueSize,
		status.ClusterID, status.CommandQueueSize,
		status.ClusterID, status.ResultQueueSize,
	)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

// boolToInt converts boolean to int for metrics
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}