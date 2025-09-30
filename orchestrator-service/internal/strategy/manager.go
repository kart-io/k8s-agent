package strategy

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/orchestrator-service/internal/storage"
	"github.com/kart-io/k8s-agent/orchestrator-service/internal/workflow"
	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
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
	// Get all active strategies
	strategies, err := m.store.ListStrategies(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("failed to list strategies: %w", err)
	}

	// Find best matching strategy
	var bestMatch *types.Strategy
	var bestScore int

	for _, strategy := range strategies {
		score := m.calculateMatchScore(event, strategy)
		if score > bestScore {
			bestScore = score
			bestMatch = strategy
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no matching strategy found")
	}

	m.logger.Info("Strategy matched",
		zap.String("strategy_id", bestMatch.ID),
		zap.String("strategy_name", bestMatch.Name),
		zap.Int("score", bestScore))

	return bestMatch, nil
}

// ExecuteStrategy executes a matched strategy
func (m *Manager) ExecuteStrategy(ctx context.Context, strategy *types.Strategy, event types.InternalEvent) (*types.WorkflowExecution, error) {
	m.logger.Info("Executing strategy",
		zap.String("strategy_id", strategy.ID),
		zap.String("workflow_id", strategy.WorkflowID))

	// Start workflow execution
	return m.engine.StartWorkflow(ctx, strategy.WorkflowID, map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	})
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