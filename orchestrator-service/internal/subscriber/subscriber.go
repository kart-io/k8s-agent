package subscriber

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/orchestrator-service/internal/strategy"
	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
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
	s.logger.Info("Starting event subscriber")

	// Subscribe to critical events
	if err := s.subscribeCriticalEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to critical events: %w", err)
	}

	// Subscribe to anomaly events
	if err := s.subscribeAnomalyEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to anomaly events: %w", err)
	}

	s.logger.Info("Event subscriber started")
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
	sub, err := s.conn.Subscribe("internal.event.critical", func(msg *nats.Msg) {
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("Subscribed to critical events")
	return nil
}

func (s *Subscriber) subscribeAnomalyEvents() error {
	sub, err := s.conn.Subscribe("internal.event.anomaly", func(msg *nats.Msg) {
		s.handleEvent(msg)
	})
	if err != nil {
		return err
	}

	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Info("Subscribed to anomaly events")
	return nil
}

func (s *Subscriber) handleEvent(msg *nats.Msg) {
	var event types.InternalEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		s.logger.Error("Failed to unmarshal event", zap.Error(err))
		return
	}

	s.logger.Info("Received internal event",
		zap.String("type", event.Type),
		zap.String("cluster_id", event.ClusterID),
		zap.String("severity", event.Severity))

	// Match strategy
	ctx := context.Background()
	matchedStrategy, err := s.strategyManager.MatchStrategy(ctx, event)
	if err != nil {
		s.logger.Warn("No strategy matched for event",
			zap.String("event_type", event.Type),
			zap.Error(err))
		return
	}

	// Execute strategy
	execution, err := s.strategyManager.ExecuteStrategy(ctx, matchedStrategy, event)
	if err != nil {
		s.logger.Error("Failed to execute strategy",
			zap.String("strategy_id", matchedStrategy.ID),
			zap.Error(err))
		return
	}

	s.logger.Info("Strategy execution started",
		zap.String("strategy_id", matchedStrategy.ID),
		zap.String("execution_id", execution.ID))
}