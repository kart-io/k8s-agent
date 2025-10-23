package strategy

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
)

// Manager manages diagnostic strategies
type Manager struct {
	store  *storage.PostgresStore
	engine *workflow.Engine
	logger *zap.Logger
}

// NewManager creates a new strategy manager
func NewManager(
	store *storage.PostgresStore,
	engine *workflow.Engine,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		store:  store,
		engine: engine,
		logger: logger.With(zap.String("component", "strategy-manager")),
	}
}

// MatchStrategy finds matching strategy for an event
func (m *Manager) MatchStrategy(ctx context.Context, event types.InternalEvent) (*types.Strategy, error) {
	m.logger.Info("🔍 Starting strategy matching",
		zap.String("event_type", event.Type),
		zap.String("severity", event.Severity))

	// Get all active strategies
	strategies, err := m.store.ListStrategies(ctx, true)
	if err != nil {
		m.logger.Error("❌ Failed to list strategies from database", zap.Error(err))
		return nil, fmt.Errorf("failed to list strategies: %w", err)
	}

	m.logger.Info("📋 Retrieved strategies from database",
		zap.Int("total_strategies", len(strategies)))

	if len(strategies) == 0 {
		m.logger.Warn("⚠️  No active strategies found in database")
		return nil, fmt.Errorf("no active strategies available")
	}

	// Find best matching strategy
	var bestMatch *types.Strategy
	var bestScore int

	for i, strategy := range strategies {
		score := m.calculateMatchScore(event, strategy)
		m.logger.Debug("Evaluating strategy",
			zap.Int("index", i),
			zap.String("strategy_id", strategy.ID),
			zap.String("strategy_name", strategy.Name),
			zap.String("category", strategy.Category),
			zap.Int("score", score))

		if score > bestScore {
			bestScore = score
			bestMatch = strategy
			m.logger.Info("✓ New best match found",
				zap.String("strategy_name", strategy.Name),
				zap.Int("score", score))
		}
	}

	if bestMatch == nil {
		m.logger.Warn("⚠️  No matching strategy found",
			zap.String("event_type", event.Type))
		return nil, fmt.Errorf("no matching strategy found")
	}

	m.logger.Info("✅ Strategy matched successfully",
		zap.String("strategy_id", bestMatch.ID),
		zap.String("strategy_name", bestMatch.Name),
		zap.String("category", bestMatch.Category),
		zap.Int("final_score", bestScore))

	return bestMatch, nil
}

// ExecuteStrategy executes a matched strategy
func (m *Manager) ExecuteStrategy(ctx context.Context, strategy *types.Strategy, event types.InternalEvent) (*types.WorkflowExecution, error) {
	m.logger.Info("🚀 Starting strategy execution",
		zap.String("strategy_id", strategy.ID),
		zap.String("strategy_name", strategy.Name),
		zap.String("workflow_id", strategy.WorkflowID))

	// Start workflow execution
	execution, err := m.engine.StartWorkflow(ctx, strategy.WorkflowID, map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	})

	if err != nil {
		m.logger.Error("❌ Failed to start workflow",
			zap.String("strategy_id", strategy.ID),
			zap.String("workflow_id", strategy.WorkflowID),
			zap.Error(err))
		return nil, err
	}

	m.logger.Info("✅ Workflow execution started",
		zap.String("execution_id", execution.ID),
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("status", string(execution.Status)))

	return execution, nil
}

// calculateMatchScore calculates match score between event and strategy
func (m *Manager) calculateMatchScore(event types.InternalEvent, strategy *types.Strategy) int {
	score := 0

	// Check each symptom
	for _, symptom := range strategy.Symptoms {
		if m.symptomMatches(event, symptom) {
			score += 10
		}
	}

	return score
}

// symptomMatches checks if symptom matches event
func (m *Manager) symptomMatches(event types.InternalEvent, symptom types.Symptom) bool {
	payload, ok := event.Payload["event"].(map[string]interface{})
	if !ok {
		return false
	}

	// Match by event type
	if symptom.Type == "event" {
		if reason, ok := payload["reason"].(string); ok {
			return reason == symptom.Pattern
		}
	}

	return false
}
