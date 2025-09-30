package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/agent"
	"github.com/kart-io/k8s-agent/agent-manager/internal/event"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// Server manages NATS server connection and subscriptions
type Server struct {
	conn          *nats.Conn
	logger        *zap.Logger
	config        types.NATSConfig

	// Components
	registry      *agent.Registry
	eventProcessor *event.Processor

	// Subscriptions
	subscriptions []*nats.Subscription
	mu            sync.RWMutex
	stopCh        chan struct{}
	wg            sync.WaitGroup

	// Metrics
	messagesReceived int64
	messagesSent     int64
	errorCount       int64
}

// NewServer creates a new NATS server instance
func NewServer(
	config types.NATSConfig,
	registry *agent.Registry,
	eventProcessor *event.Processor,
	logger *zap.Logger,
) *Server {
	return &Server{
		config:         config,
		registry:       registry,
		eventProcessor: eventProcessor,
		logger:         logger.With(zap.String("component", "nats-server")),
		stopCh:         make(chan struct{}),
	}
}

// Start starts the NATS server and subscriptions
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting NATS server", zap.String("url", s.config.URL))

	// Connect to NATS
	if err := s.connect(); err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Setup subscriptions
	if err := s.setupSubscriptions(); err != nil {
		return fmt.Errorf("failed to setup subscriptions: %w", err)
	}

	// Start connection monitor
	s.wg.Add(1)
	go s.connectionMonitor()

	s.logger.Info("NATS server started successfully")

	return nil
}

// Stop stops the NATS server
func (s *Server) Stop() error {
	s.logger.Info("Stopping NATS server")

	close(s.stopCh)
	s.wg.Wait()

	// Unsubscribe all
	s.mu.Lock()
	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warn("Failed to unsubscribe", zap.Error(err))
		}
	}
	s.subscriptions = nil
	s.mu.Unlock()

	// Close connection
	if s.conn != nil {
		s.conn.Close()
	}

	s.logger.Info("NATS server stopped")

	return nil
}

// connect establishes connection to NATS server
func (s *Server) connect() error {
	opts := []nats.Option{
		nats.Name("agent-manager"),
		nats.MaxReconnects(s.config.MaxReconnect),
		nats.ReconnectWait(s.config.ReconnectWait),
		nats.PingInterval(s.config.PingInterval),
		nats.MaxPingsOutstanding(s.config.MaxPingsOut),
		nats.DisconnectErrHandler(s.handleDisconnect),
		nats.ReconnectHandler(s.handleReconnect),
		nats.ErrorHandler(s.handleError),
	}

	conn, err := nats.Connect(s.config.URL, opts...)
	if err != nil {
		return err
	}

	s.conn = conn
	s.logger.Info("Connected to NATS", zap.String("url", s.config.URL))

	return nil
}

// setupSubscriptions sets up all NATS subscriptions
func (s *Server) setupSubscriptions() error {
	// Subscribe to agent registration
	if err := s.subscribeRegister(); err != nil {
		return fmt.Errorf("failed to subscribe to register: %w", err)
	}

	// Subscribe to agent heartbeat
	if err := s.subscribeHeartbeat(); err != nil {
		return fmt.Errorf("failed to subscribe to heartbeat: %w", err)
	}

	// Subscribe to agent events
	if err := s.subscribeEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Subscribe to agent metrics
	if err := s.subscribeMetrics(); err != nil {
		return fmt.Errorf("failed to subscribe to metrics: %w", err)
	}

	// Subscribe to command results
	if err := s.subscribeResults(); err != nil {
		return fmt.Errorf("failed to subscribe to results: %w", err)
	}

	return nil
}

// subscribeRegister subscribes to agent registration messages
func (s *Server) subscribeRegister() error {
	subject := "aetherius.agent.*.register"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleRegister(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Info("Subscribed to agent registration", zap.String("subject", subject))

	return nil
}

// subscribeHeartbeat subscribes to agent heartbeat messages
func (s *Server) subscribeHeartbeat() error {
	subject := "aetherius.agent.*.heartbeat"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleHeartbeat(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Info("Subscribed to agent heartbeat", zap.String("subject", subject))

	return nil
}

// subscribeEvents subscribes to agent event messages
func (s *Server) subscribeEvents() error {
	subject := "aetherius.agent.*.event"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Info("Subscribed to agent events", zap.String("subject", subject))

	return nil
}

// subscribeMetrics subscribes to agent metrics messages
func (s *Server) subscribeMetrics() error {
	subject := "aetherius.agent.*.metrics"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleMetrics(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Info("Subscribed to agent metrics", zap.String("subject", subject))

	return nil
}

// subscribeResults subscribes to command result messages
func (s *Server) subscribeResults() error {
	subject := "aetherius.agent.*.result"

	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.handleResult(msg)
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subscriptions = append(s.subscriptions, sub)
	s.mu.Unlock()

	s.logger.Info("Subscribed to command results", zap.String("subject", subject))

	return nil
}

// Message handlers

// handleRegister handles agent registration messages
func (s *Server) handleRegister(msg *nats.Msg) {
	s.messagesReceived++

	var agentInfo types.Agent
	if err := json.Unmarshal(msg.Data, &agentInfo); err != nil {
		s.logger.Error("Failed to unmarshal register message", zap.Error(err))
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.registry.RegisterAgent(ctx, &agentInfo); err != nil {
		s.logger.Error("Failed to register agent",
			zap.String("cluster_id", agentInfo.ClusterID),
			zap.Error(err))
		s.errorCount++
		return
	}

	s.logger.Info("Agent registered successfully",
		zap.String("agent_id", agentInfo.ID),
		zap.String("cluster_id", agentInfo.ClusterID))

	// Send acknowledgment
	ack := map[string]interface{}{
		"status":  "registered",
		"agent_id": agentInfo.ID,
	}
	s.sendResponse(msg, ack)
}

// handleHeartbeat handles agent heartbeat messages
func (s *Server) handleHeartbeat(msg *nats.Msg) {
	s.messagesReceived++

	var heartbeat struct {
		AgentID   string    `json:"agent_id"`
		ClusterID string    `json:"cluster_id"`
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(msg.Data, &heartbeat); err != nil {
		s.logger.Error("Failed to unmarshal heartbeat message", zap.Error(err))
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.registry.UpdateHeartbeat(ctx, heartbeat.AgentID); err != nil {
		s.logger.Warn("Failed to update heartbeat",
			zap.String("agent_id", heartbeat.AgentID),
			zap.Error(err))
		s.errorCount++
		return
	}

	s.logger.Debug("Heartbeat received",
		zap.String("agent_id", heartbeat.AgentID),
		zap.String("cluster_id", heartbeat.ClusterID))
}

// handleEvent handles agent event messages
func (s *Server) handleEvent(msg *nats.Msg) {
	s.messagesReceived++

	var event types.Event
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("Failed to unmarshal event message", zap.Error(err))
		s.errorCount++
		return
	}

	ctx := context.Background()
	if err := s.eventProcessor.ProcessEvent(ctx, &event); err != nil {
		s.logger.Error("Failed to process event",
			zap.String("event_id", event.ID),
			zap.String("cluster_id", event.ClusterID),
			zap.Error(err))
		s.errorCount++
		return
	}

	s.logger.Debug("Event processed",
		zap.String("event_id", event.ID),
		zap.String("cluster_id", event.ClusterID),
		zap.String("severity", event.Severity))
}

// handleMetrics handles agent metrics messages
func (s *Server) handleMetrics(msg *nats.Msg) {
	s.messagesReceived++

	var metrics types.Metrics
	if err := json.Unmarshal(msg.Data, &metrics); err != nil {
		s.logger.Error("Failed to unmarshal metrics message", zap.Error(err))
		s.errorCount++
		return
	}

	// TODO: Process metrics (store in Prometheus/VictoriaMetrics)
	s.logger.Debug("Metrics received",
		zap.String("cluster_id", metrics.ClusterID))
}

// handleResult handles command result messages
func (s *Server) handleResult(msg *nats.Msg) {
	s.messagesReceived++

	var result types.CommandResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		s.logger.Error("Failed to unmarshal result message", zap.Error(err))
		s.errorCount++
		return
	}

	// TODO: Process command result
	s.logger.Info("Command result received",
		zap.String("command_id", result.CommandID),
		zap.String("cluster_id", result.ClusterID),
		zap.String("status", result.Status))
}

// PublishCommand publishes a command to an agent
func (s *Server) PublishCommand(clusterID string, cmd *types.Command) error {
	subject := fmt.Sprintf("aetherius.agent.%s.command", clusterID)

	data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	if err := s.conn.Publish(subject, data); err != nil {
		s.errorCount++
		return fmt.Errorf("failed to publish command: %w", err)
	}

	s.messagesSent++
	s.logger.Info("Command published",
		zap.String("command_id", cmd.ID),
		zap.String("cluster_id", clusterID),
		zap.String("subject", subject))

	return nil
}

// sendResponse sends a response message
func (s *Server) sendResponse(msg *nats.Msg, response interface{}) {
	data, err := json.Marshal(response)
	if err != nil {
		s.logger.Error("Failed to marshal response", zap.Error(err))
		return
	}

	if err := msg.Respond(data); err != nil {
		s.logger.Error("Failed to send response", zap.Error(err))
		return
	}

	s.messagesSent++
}

// Connection event handlers

func (s *Server) handleDisconnect(conn *nats.Conn, err error) {
	s.logger.Warn("Disconnected from NATS",
		zap.Error(err),
		zap.String("url", s.config.URL))
}

func (s *Server) handleReconnect(conn *nats.Conn) {
	s.logger.Info("Reconnected to NATS",
		zap.String("url", conn.ConnectedUrl()))
}

func (s *Server) handleError(conn *nats.Conn, sub *nats.Subscription, err error) {
	s.logger.Error("NATS error",
		zap.Error(err),
		zap.String("subject", sub.Subject))
	s.errorCount++
}

// connectionMonitor monitors connection health
func (s *Server) connectionMonitor() {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.conn == nil || !s.conn.IsConnected() {
				s.logger.Warn("NATS connection lost, attempting reconnect")
			}
		}
	}
}

// GetStatistics returns server statistics
func (s *Server) GetStatistics() map[string]interface{} {
	var connected bool
	var connectedURL string

	if s.conn != nil {
		connected = s.conn.IsConnected()
		connectedURL = s.conn.ConnectedUrl()
	}

	return map[string]interface{}{
		"connected":          connected,
		"connected_url":      connectedURL,
		"messages_received":  s.messagesReceived,
		"messages_sent":      s.messagesSent,
		"error_count":        s.errorCount,
		"subscription_count": len(s.subscriptions),
	}
}

// Health checks NATS server health
func (s *Server) Health() error {
	if s.conn == nil {
		return fmt.Errorf("not connected")
	}
	if !s.conn.IsConnected() {
		return fmt.Errorf("connection lost")
	}
	return nil
}