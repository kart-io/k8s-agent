package strategy

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	"github.com/kart-io/logger/core"
)

// Manager manages diagnostic strategies.
type Manager struct {
	store  *storage.MySQLStore
	engine *workflow.Engine
	logger core.Logger
}

// NewManager creates a new strategy manager.
func NewManager(
	store *storage.MySQLStore,
	engine *workflow.Engine,
	logger core.Logger,
) *Manager {
	return &Manager{
		store:  store,
		engine: engine,
		logger: logger,
	}
}

// MatchStrategy finds matching strategy for an event.
func (m *Manager) MatchStrategy(ctx context.Context, event types.InternalEvent) (*types.Strategy, error) {
	m.logger.Info("🔍 Starting strategy matching",
		"event_type", event.Type,
		"severity", event.Severity)

	// Get all active strategies
	strategies, err := m.store.ListStrategies(ctx, true)
	if err != nil {
		m.logger.Error("❌ Failed to list strategies from database", "error", err)
		return nil, fmt.Errorf("failed to list strategies: %w", err)
	}

	m.logger.Info("📋 Retrieved strategies from database",
		"total_strategies", len(strategies))

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
			"index", i,
			"strategy_id", strategy.ID,
			"strategy_name", strategy.Name,
			"category", strategy.Category,
			"score", score)

		if score > bestScore {
			bestScore = score
			bestMatch = strategy
			m.logger.Info("✓ New best match found",
				"strategy_name", strategy.Name,
				"score", score)
		}
	}

	if bestMatch == nil {
		m.logger.Warn("⚠️  No matching strategy found",
			"event_type", event.Type)
		return nil, fmt.Errorf("no matching strategy found")
	}

	m.logger.Info("✅ Strategy matched successfully",
		"strategy_id", bestMatch.ID,
		"strategy_name", bestMatch.Name,
		"category", bestMatch.Category,
		"final_score", bestScore)

	return bestMatch, nil
}

// ExecuteStrategy executes a matched strategy.
func (m *Manager) ExecuteStrategy(ctx context.Context, strategy *types.Strategy, event types.InternalEvent) (*types.WorkflowExecution, error) {
	m.logger.Info("🚀 Starting strategy execution",
		"strategy_id", strategy.ID,
		"strategy_name", strategy.Name,
		"workflow_id", strategy.WorkflowID)

	// Start workflow execution
	execution, err := m.engine.StartWorkflow(ctx, strategy.WorkflowID, map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	})
	if err != nil {
		m.logger.Error("❌ Failed to start workflow",
			"strategy_id", strategy.ID,
			"workflow_id", strategy.WorkflowID,
			"error", err)
		return nil, err
	}

	m.logger.Info("✅ Workflow execution started",
		"execution_id", execution.ID,
		"workflow_id", execution.WorkflowID,
		"status", string(execution.Status))

	return execution, nil
}

// calculateMatchScore calculates match score between event and strategy.
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

// symptomMatches checks if symptom matches event.
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
