package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kart-io/logger/core"
	"github.com/nats-io/nats.go"

	"github.com/kart-io/k8s-agent/internal/collect-agent/types"
)

// CommunicationManager handles all NATS communication for the agent
type CommunicationManager struct {
	config     *types.AgentConfig
	clusterID  string
	k8sVersion string
	apiServer  string
	natsConn   *nats.Conn
	logger     core.Logger
	mu         sync.RWMutex
	connected  bool
	stopCh     chan struct{}
	wg         sync.WaitGroup

	// Channels for different message types
	eventChan      <-chan *types.Event
	metricsChan    <-chan *types.Metrics
	resultChan     <-chan *types.CommandResult
	commandHandler func(*types.Command)
}

// NewCommunicationManager creates a new communication manager
func NewCommunicationManager(
	config *types.AgentConfig,
	clusterID string,
	k8sVersion string,
	apiServer string,
	eventChan <-chan *types.Event,
	metricsChan <-chan *types.Metrics,
	resultChan <-chan *types.CommandResult,
	commandHandler func(*types.Command),
	logger core.Logger,
) *CommunicationManager {
	return &CommunicationManager{
		config:         config,
		clusterID:      clusterID,
		k8sVersion:     k8sVersion,
		apiServer:      apiServer,
		eventChan:      eventChan,
		metricsChan:    metricsChan,
		resultChan:     resultChan,
		commandHandler: commandHandler,
		logger:         logger, // Logger already has component context from caller
		stopCh:         make(chan struct{}),
	}
}

// Start initializes the NATS connection and starts message handling
func (cm *CommunicationManager) Start(ctx context.Context) error {
	if err := cm.connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	if err := cm.register(); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	// Start message handlers
	cm.wg.Add(4)
	go cm.handleEvents(ctx)
	go cm.handleMetrics(ctx)
	go cm.handleResults(ctx)
	go cm.handleHeartbeat(ctx)

	// Subscribe to commands
	if err := cm.subscribeToCommands(); err != nil {
		return fmt.Errorf("failed to subscribe to commands: %w", err)
	}

	cm.logger.Infow("Communication manager started",
		"cluster_id", cm.clusterID,
		"endpoint", cm.config.CentralEndpoint)

	return nil
}

// Stop stops the communication manager and closes connections
func (cm *CommunicationManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.connected {
		return nil
	}

	cm.logger.Infow("Stopping communication manager")

	close(cm.stopCh)
	cm.wg.Wait()

	if cm.natsConn != nil {
		cm.natsConn.Close()
	}

	cm.connected = false
	cm.logger.Infow("Communication manager stopped")
	return nil
}

// connect establishes connection to NATS server
func (cm *CommunicationManager) connect() error {
	cm.logger.Infow("Connecting to NATS", "endpoint", cm.config.CentralEndpoint)

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("agent-%s", cm.clusterID)),
		nats.ReconnectWait(cm.config.ReconnectDelay),
		nats.MaxReconnects(cm.config.MaxRetries),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			cm.logger.Warnw("Disconnected from NATS",
				"cluster_id", cm.clusterID,
				"error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			cm.logger.Infow("Reconnected to NATS",
				"cluster_id", cm.clusterID,
				"url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			cm.logger.Warnw("NATS connection closed",
				"cluster_id", cm.clusterID)
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			cm.logger.Errorw("NATS error",
				"cluster_id", cm.clusterID,
				"subject", sub.Subject,
				"error", err)
		}),
	}

	nc, err := nats.Connect(cm.config.CentralEndpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	cm.natsConn = nc
	cm.connected = true
	cm.logger.Infow("Connected to NATS", "url", nc.ConnectedUrl())

	return nil
}

// register sends agent registration information to central
func (cm *CommunicationManager) register() error {
	now := time.Now()

	// Get local IP address
	localIP := getLocalIP()
	serviceAddress := fmt.Sprintf("http://%s:%d", localIP, cm.config.HealthPort)

	// Create agent registration matching agent-manager's Agent type
	agentInfo := map[string]interface{}{
		"id":             cm.clusterID, // Using clusterID as agent ID
		"cluster_id":     cm.clusterID,
		"cluster_name":   cm.config.ClusterName,
		"version":        "v1.0.0", // Agent version
		"status":         "online",
		"registered_at":  now,
		"last_heartbeat": now,
		"capabilities":   []string{"event_watch", "metrics_collect", "command_execute"},
		"connection_info": map[string]interface{}{
			"endpoint":        cm.config.CentralEndpoint,
			"connected_at":    now,
			"last_seen":       now,
			"reconnect_count": 0,
			"service_address": serviceAddress,
			"local_ip":        localIP,
		},
		// Cluster information for agent-manager to update cluster record
		"k8s_version": cm.k8sVersion,
		"api_server":  cm.apiServer,
	}

	data, err := json.Marshal(agentInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal agent info: %w", err)
	}

	subject := fmt.Sprintf("aetherius.agent.%s.register", cm.clusterID)
	if err := cm.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish register message: %w", err)
	}

	cm.logger.Infow("Agent registered", "cluster_id", cm.clusterID)
	return nil
}

// subscribeToCommands subscribes to command messages from central
func (cm *CommunicationManager) subscribeToCommands() error {
	subject := fmt.Sprintf("aetherius.agent.%s.command", cm.clusterID)

	_, err := cm.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var cmd types.Command
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			cm.logger.Errorw("Failed to unmarshal command", "error", err)
			return
		}

		cm.logger.Infow("Received command",
			"cluster_id", cm.clusterID,
			"command_id", cmd.ID,
			"tool", cmd.Tool,
			"action", cmd.Action)

		// Handle command asynchronously
		if cm.commandHandler != nil {
			go cm.commandHandler(&cmd)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to commands: %w", err)
	}

	cm.logger.Infow("Subscribed to commands", "subject", subject)
	return nil
}

// handleEvents handles event publishing
func (cm *CommunicationManager) handleEvents(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("aetherius.agent.%s.event", cm.clusterID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case event := <-cm.eventChan:
			if event == nil {
				continue
			}

			if err := cm.publishEvent(subject, event); err != nil {
				cm.logger.Errorw("Failed to publish event",
					"error", err,
					"event_id", event.ID)
			}
		}
	}
}

// handleMetrics handles metrics publishing
func (cm *CommunicationManager) handleMetrics(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("aetherius.agent.%s.metrics", cm.clusterID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case metrics := <-cm.metricsChan:
			if metrics == nil {
				continue
			}

			if err := cm.publishMetrics(subject, metrics); err != nil {
				cm.logger.Errorw("Failed to publish metrics", "error", err)
			}
		}
	}
}

// handleResults handles command result publishing
func (cm *CommunicationManager) handleResults(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("aetherius.agent.%s.result", cm.clusterID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case result := <-cm.resultChan:
			if result == nil {
				continue
			}

			if err := cm.publishResult(subject, result); err != nil {
				cm.logger.Errorw("Failed to publish result",
					"error", err,
					"command_id", result.CommandID)
			}
		}
	}
}

// handleHeartbeat sends periodic heartbeat messages
func (cm *CommunicationManager) handleHeartbeat(ctx context.Context) {
	defer cm.wg.Done()

	ticker := time.NewTicker(cm.config.HeartbeatInterval)
	defer ticker.Stop()

	subject := fmt.Sprintf("aetherius.agent.%s.heartbeat", cm.clusterID)

	// Send initial heartbeat
	cm.sendHeartbeat(subject)

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case <-ticker.C:
			cm.sendHeartbeat(subject)
		}
	}
}

// publishEvent publishes an event to NATS
func (cm *CommunicationManager) publishEvent(subject string, event *types.Event) error {
	event.ClusterID = cm.clusterID
	event.ReportedAt = time.Now()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := cm.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	cm.logger.Debug("Event published",
		"event_id", event.ID,
		"subject", subject)

	return nil
}

// publishMetrics publishes metrics to NATS
func (cm *CommunicationManager) publishMetrics(subject string, metrics *types.Metrics) error {
	metrics.ClusterID = cm.clusterID
	metrics.Timestamp = time.Now()

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := cm.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish metrics: %w", err)
	}

	cm.logger.Debug("Metrics published", "subject", subject)
	return nil
}

// publishResult publishes a command result to NATS
func (cm *CommunicationManager) publishResult(subject string, result *types.CommandResult) error {
	result.ClusterID = cm.clusterID

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := cm.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}

	cm.logger.Infow("Result published",
		"command_id", result.CommandID,
		"status", result.Status,
		"subject", subject)

	return nil
}

// sendHeartbeat sends a heartbeat message
func (cm *CommunicationManager) sendHeartbeat(subject string) {
	// Create heartbeat matching agent-manager's expected format
	heartbeat := map[string]interface{}{
		"agent_id":   cm.clusterID, // Using clusterID as agent ID
		"cluster_id": cm.clusterID,
		"timestamp":  time.Now(),
		"status":     "healthy",
		"metrics": map[string]interface{}{
			"event_queue_size":   0, // This should be actual queue size
			"metrics_queue_size": 0, // This should be actual queue size
			"command_queue_size": 0, // This should be actual queue size
			"uptime_seconds":     0, // This should be actual uptime
		},
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		cm.logger.Errorw("Failed to marshal heartbeat", "error", err)
		return
	}

	if err := cm.natsConn.Publish(subject, data); err != nil {
		cm.logger.Errorw("Failed to publish heartbeat", "error", err)
		return
	}

	cm.logger.Debug("Heartbeat sent",
		"agent_id", cm.clusterID,
		"cluster_id", cm.clusterID)
}

// IsConnected returns true if connected to NATS
func (cm *CommunicationManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connected && cm.natsConn != nil && cm.natsConn.IsConnected()
}

// getLocalIP gets the local non-loopback IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
