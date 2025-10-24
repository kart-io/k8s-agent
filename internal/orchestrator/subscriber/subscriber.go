package subscriber

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"

	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/logger/core"
)

// Subscriber subscribes to internal events from agent-manager
type Subscriber struct {
	conn            *nats.Conn
	strategyManager *strategy.Manager
	logger          core.Logger
	subscriptions   []*nats.Subscription
}

// NewSubscriber creates a new subscriber
func NewSubscriber(
	conn *nats.Conn,
	strategyManager *strategy.Manager,
	logger core.Logger,
) *Subscriber {
	return &Subscriber{
		conn:            conn,
		strategyManager: strategyManager,
		logger:          logger,
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
		s.logger.Warn("Failed to subscribe to all events (debug)", "error", err)
	}

	s.logger.Info("========== Event subscriber started successfully ==========",
		"active_subscriptions", len(s.subscriptions))
	return nil
}

// Stop stops the subscriber
func (s *Subscriber) Stop() error {
	s.logger.Info("Stopping event subscriber")

	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warn("Failed to unsubscribe", "error", err)
		}
	}

	return nil
}

func (s *Subscriber) subscribeCriticalEvents() error {
	subject := "internal.event.critical"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Info("📨 Received message on critical channel",
			"subject", msg.Subject,
			"size", len(msg.Data))
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to critical events", "subject", subject)
	return nil
}

func (s *Subscriber) subscribeAnomalyEvents() error {
	subject := "internal.event.anomaly"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Info("📨 Received message on anomaly channel",
			"subject", msg.Subject,
			"size", len(msg.Data))
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to anomaly events", "subject", subject)
	return nil
}

func (s *Subscriber) subscribeAllEvents() error {
	subject := "internal.event.>"
	sub, err := s.conn.Subscribe(subject, func(msg *nats.Msg) {
		s.logger.Debug("📬 Debug: Received message on any internal.event channel",
			"subject", msg.Subject,
			"size", len(msg.Data),
			"preview", string(msg.Data[:min(100, len(msg.Data))]))
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("✓ Subscribed to all internal events (debug)", "subject", subject)
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
		"subject", msg.Subject)

	var event types.InternalEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("❌ Failed to unmarshal event",
			"error", err,
			"raw_data", string(msg.Data))
		return
	}

	s.logger.Info("✓ Event parsed successfully",
		"type", event.Type,
		"cluster_id", event.ClusterID,
		"severity", event.Severity,
		"timestamp", event.Timestamp)

	// Match strategy
	s.logger.Info("🔍 Matching strategy for event...",
		"event_type", event.Type)

	ctx := context.Background()
	matchedStrategy, err := s.strategyManager.MatchStrategy(ctx, event)
	if err != nil {
		s.logger.Warn("⚠️  No strategy matched for event",
			"event_type", event.Type,
			"severity", event.Severity,
			"error", err)
		return
	}

	s.logger.Info("✓ Strategy matched",
		"strategy_id", matchedStrategy.ID,
		"strategy_name", matchedStrategy.Name,
		"workflow_id", matchedStrategy.WorkflowID)

	// Execute strategy
	s.logger.Info("🚀 Executing strategy...",
		"strategy_id", matchedStrategy.ID)

	execution, err := s.strategyManager.ExecuteStrategy(ctx, matchedStrategy, event)
	if err != nil {
		s.logger.Error("❌ Failed to execute strategy",
			"strategy_id", matchedStrategy.ID,
			"error", err)
		return
	}

	s.logger.Info("========== Strategy execution started successfully ==========",
		"strategy_id", matchedStrategy.ID,
		"execution_id", execution.ID,
		"status", string(execution.Status))
}
