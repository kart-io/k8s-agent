package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/kart/k8s-agent/collect-agent/internal/types"
)

// CommunicationManager handles all NATS communication for the agent
type CommunicationManager struct {
	config      *types.AgentConfig
	clusterID   string
	natsConn    *nats.Conn
	logger      *zap.Logger
	mu          sync.RWMutex
	connected   bool
	stopCh      chan struct{}
	wg          sync.WaitGroup

	// Channels for different message types
	eventChan   <-chan *types.Event
	metricsChan <-chan *types.Metrics
	resultChan  <-chan *types.CommandResult
	commandHandler func(*types.Command)
}

// NewCommunicationManager creates a new communication manager
func NewCommunicationManager(
	config *types.AgentConfig,
	clusterID string,
	eventChan <-chan *types.Event,
	metricsChan <-chan *types.Metrics,
	resultChan <-chan *types.CommandResult,
	commandHandler func(*types.Command),
	logger *zap.Logger,
) *CommunicationManager {
	return &CommunicationManager{
		config:         config,
		clusterID:      clusterID,
		eventChan:      eventChan,
		metricsChan:    metricsChan,
		resultChan:     resultChan,
		commandHandler: commandHandler,
		logger:         logger.With(zap.String("component", "communication")),
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

	cm.logger.Info("Communication manager started",
		zap.String("cluster_id", cm.clusterID),
		zap.String("endpoint", cm.config.CentralEndpoint))

	return nil
}

// Stop stops the communication manager and closes connections
func (cm *CommunicationManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.connected {
		return nil
	}

	cm.logger.Info("Stopping communication manager")

	close(cm.stopCh)
	cm.wg.Wait()

	if cm.natsConn != nil {
		cm.natsConn.Close()
	}

	cm.connected = false
	cm.logger.Info("Communication manager stopped")
	return nil
}

// connect establishes connection to NATS server
func (cm *CommunicationManager) connect() error {
	cm.logger.Info("Connecting to NATS", zap.String("endpoint", cm.config.CentralEndpoint))

	opts := []nats.Option{
		nats.Name(fmt.Sprintf("agent-%s", cm.clusterID)),
		nats.ReconnectWait(cm.config.ReconnectDelay),
		nats.MaxReconnects(cm.config.MaxRetries),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			cm.logger.Warn("Disconnected from NATS",
				zap.String("cluster_id", cm.clusterID),
				zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			cm.logger.Info("Reconnected to NATS",
				zap.String("cluster_id", cm.clusterID),
				zap.String("url", nc.ConnectedUrl()))
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			cm.logger.Warn("NATS connection closed",
				zap.String("cluster_id", cm.clusterID))
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			cm.logger.Error("NATS error",
				zap.String("cluster_id", cm.clusterID),
				zap.String("subject", sub.Subject),
				zap.Error(err))
		}),
	}

	nc, err := nats.Connect(cm.config.CentralEndpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	cm.natsConn = nc
	cm.connected = true
	cm.logger.Info("Connected to NATS", zap.String("url", nc.ConnectedUrl()))

	return nil
}

// register sends agent registration information to central
func (cm *CommunicationManager) register() error {
	agentInfo := types.AgentInfo{
		ClusterID:   cm.clusterID,
		Version:     "v1.0.0", // This should come from version package
		StartTime:   time.Now(),
		Capabilities: []string{"event_watch", "metrics_collect", "command_execute"},
	}

	data, err := json.Marshal(agentInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal agent info: %w", err)
	}

	subject := fmt.Sprintf("agent.register.%s", cm.clusterID)
	if err := cm.natsConn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish register message: %w", err)
	}

	cm.logger.Info("Agent registered", zap.String("cluster_id", cm.clusterID))
	return nil
}

// subscribeToCommands subscribes to command messages from central
func (cm *CommunicationManager) subscribeToCommands() error {
	subject := fmt.Sprintf("agent.command.%s", cm.clusterID)

	_, err := cm.natsConn.Subscribe(subject, func(msg *nats.Msg) {
		var cmd types.Command
		if err := json.Unmarshal(msg.Data, &cmd); err != nil {
			cm.logger.Error("Failed to unmarshal command", zap.Error(err))
			return
		}

		cm.logger.Info("Received command",
			zap.String("cluster_id", cm.clusterID),
			zap.String("command_id", cmd.ID),
			zap.String("tool", cmd.Tool),
			zap.String("action", cmd.Action))

		// Handle command asynchronously
		if cm.commandHandler != nil {
			go cm.commandHandler(&cmd)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to commands: %w", err)
	}

	cm.logger.Info("Subscribed to commands", zap.String("subject", subject))
	return nil
}

// handleEvents handles event publishing
func (cm *CommunicationManager) handleEvents(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("agent.event.%s", cm.clusterID)

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
				cm.logger.Error("Failed to publish event",
					zap.Error(err),
					zap.String("event_id", event.ID))
			}
		}
	}
}

// handleMetrics handles metrics publishing
func (cm *CommunicationManager) handleMetrics(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("agent.metrics.%s", cm.clusterID)

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
				cm.logger.Error("Failed to publish metrics", zap.Error(err))
			}
		}
	}
}

// handleResults handles command result publishing
func (cm *CommunicationManager) handleResults(ctx context.Context) {
	defer cm.wg.Done()

	subject := fmt.Sprintf("agent.result.%s", cm.clusterID)

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
				cm.logger.Error("Failed to publish result",
					zap.Error(err),
					zap.String("command_id", result.CommandID))
			}
		}
	}
}

// handleHeartbeat sends periodic heartbeat messages
func (cm *CommunicationManager) handleHeartbeat(ctx context.Context) {
	defer cm.wg.Done()

	ticker := time.NewTicker(cm.config.HeartbeatInterval)
	defer ticker.Stop()

	subject := fmt.Sprintf("agent.heartbeat.%s", cm.clusterID)

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
		zap.String("event_id", event.ID),
		zap.String("subject", subject))

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

	cm.logger.Debug("Metrics published", zap.String("subject", subject))
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

	cm.logger.Info("Result published",
		zap.String("command_id", result.CommandID),
		zap.String("status", result.Status),
		zap.String("subject", subject))

	return nil
}

// sendHeartbeat sends a heartbeat message
func (cm *CommunicationManager) sendHeartbeat(subject string) {
	// Get queue sizes (these would be passed from the main agent)
	// For now, we'll use placeholder values
	heartbeat := types.Heartbeat{
		ClusterID: cm.clusterID,
		Timestamp: time.Now(),
		Status:    "healthy",
		Metrics: types.HeartbeatMetrics{
			EventQueueSize:   0, // This should be actual queue size
			MetricsQueueSize: 0, // This should be actual queue size
			CommandQueueSize: 0, // This should be actual queue size
			UptimeSeconds:    int(time.Since(time.Now()).Seconds()), // This should be actual uptime
		},
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		cm.logger.Error("Failed to marshal heartbeat", zap.Error(err))
		return
	}

	if err := cm.natsConn.Publish(subject, data); err != nil {
		cm.logger.Error("Failed to publish heartbeat", zap.Error(err))
		return
	}

	cm.logger.Debug("Heartbeat sent", zap.String("cluster_id", cm.clusterID))
}

// IsConnected returns true if connected to NATS
func (cm *CommunicationManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connected && cm.natsConn != nil && cm.natsConn.IsConnected()
}