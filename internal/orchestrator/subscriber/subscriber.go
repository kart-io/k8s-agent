package subscriber

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
)

// Subscriber subscribes to internal events from agent-manager
type Subscriber struct {
	conn            *nats.Conn
	strategyManager *strategy.Manager
	logger          *zap.Logger
	subscriptions   []*nats.Subscription
}

// NewSubscriber creates a new subscriber
func NewSubscriber(
	conn *nats.Conn,
	strategyManager *strategy.Manager,
	logger *zap.Logger,
) *Subscriber {
	return &Subscriber{
		conn:            conn,
		strategyManager: strategyManager,
		logger:          logger.With(zap.String("component", "subscriber")),
	}
}

// Start starts subscribing to events
func (s *Subscriber) Start(ctx context.Context) error {
	s.logger.Info("========== Starting event subscriber ==========")

	// Subscribe to critical events
	if err := s.subscribeCriticalEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to critical events: %w", err)
	}

	// Subscribe to anomaly events
	if err := s.subscribeAnomalyEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to anomaly events: %w", err)
	}

	// Subscribe to all internal events for debugging
	if err := s.subscribeAllEvents(); err != nil {
		s.logger.Warn("Failed to subscribe to all events (debug)", zap.Error(err))
	}

	s.logger.Info("========== Event subscriber started successfully ==========",
		zap.Int("active_subscriptions", len(s.subscriptions)))
	return nil
}

// Stop stops the subscriber
func (s *Subscriber) Stop() error {
	s.logger.Info("Stopping event subscriber")

	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warn("Failed to unsubscribe", zap.Error(err))
		}
	}

	return nil
}

func (s *Subscriber) subscribeCriticalEvents() error {
	subject := "internal.event.critical"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Info("📨 Received message on critical channel",
			zap.String("subject", msg.Subject),
			zap.Int("size", len(msg.Data)))
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to critical events", zap.String("subject", subject))
	return nil
}

func (s *Subscriber) subscribeAnomalyEvents() error {
	subject := "internal.event.anomaly"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Info("📨 Received message on anomaly channel",
			zap.String("subject", msg.Subject),
			zap.Int("size", len(msg.Data)))
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to anomaly events", zap.String("subject", subject))
	return nil
}

func (s *Subscriber) subscribeAllEvents() error {
	subject := "internal.event.>"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Debug("📬 Debug: Received message on any internal.event channel",
			zap.String("subject", msg.Subject),
			zap.Int("size", len(msg.Data)),
			zap.ByteString("preview", msg.Data[:min(100, len(msg.Data))]))
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to all internal events (debug)", zap.String("subject", subject))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Subscriber) handleEvent(msg *nats.Msg) {
	s.logger.Info("========== Processing Event ==========",
		zap.String("subject", msg.Subject))

	var event types.InternalEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("❌ Failed to unmarshal event",
			zap.Error(err),
			zap.ByteString("raw_data", msg.Data))
		return
	}

	s.logger.Info("✓ Event parsed successfully",
		zap.String("type", event.Type),
		zap.String("cluster_id", event.ClusterID),
		zap.String("severity", event.Severity),
		zap.Time("timestamp", event.Timestamp))

	// Match strategy
	s.logger.Info("🔍 Matching strategy for event...",
		zap.String("event_type", event.Type))

	ctx := context.Background()
	matchedStrategy, err := s.strategyManager.MatchStrategy(ctx, event)
	if err != nil {
		s.logger.Warn("⚠️  No strategy matched for event",
			zap.String("event_type", event.Type),
			zap.String("severity", event.Severity),
			zap.Error(err))
		return
	}

	s.logger.Info("✓ Strategy matched",
		zap.String("strategy_id", matchedStrategy.ID),
		zap.String("strategy_name", matchedStrategy.Name),
		zap.String("workflow_id", matchedStrategy.WorkflowID))

	// Execute strategy
	s.logger.Info("🚀 Executing strategy...",
		zap.String("strategy_id", matchedStrategy.ID))

	execution, err := s.strategyManager.ExecuteStrategy(ctx, matchedStrategy, event)
	if err != nil {
		s.logger.Error("❌ Failed to execute strategy",
			zap.String("strategy_id", matchedStrategy.ID),
			zap.Error(err))
		return
	}

	s.logger.Info("========== Strategy execution started successfully ==========",
		zap.String("strategy_id", matchedStrategy.ID),
		zap.String("execution_id", execution.ID),
		zap.String("status", string(execution.Status)))
}
