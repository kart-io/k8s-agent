package strategy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
)

// MockPostgresStore is a mock for PostgresStore
type MockPostgresStore struct {
	mock.Mock
}

func (m *MockPostgresStore) ListStrategies(ctx context.Context, activeOnly bool) ([]*types.Strategy, error) {
	args := m.Called(ctx, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Strategy), args.Error(1)
}

func (m *MockPostgresStore) GetStrategy(ctx context.Context, strategyID string) (*types.Strategy, error) {
	args := m.Called(ctx, strategyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Strategy), args.Error(1)
}

func (m *MockPostgresStore) SaveStrategy(ctx context.Context, strategy *types.Strategy) error {
	args := m.Called(ctx, strategy)
	return args.Error(0)
}

// MockWorkflowEngine is a mock for workflow engine
type MockWorkflowEngine struct {
	mock.Mock
}

func (m *MockWorkflowEngine) StartWorkflow(ctx context.Context, workflowID string, params map[string]interface{}) (*types.WorkflowExecution, error) {
	args := m.Called(ctx, workflowID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WorkflowExecution), args.Error(1)
}

func (m *MockWorkflowEngine) CancelWorkflow(ctx context.Context, executionID string) error {
	args := m.Called(ctx, executionID)
	return args.Error(0)
}

func (m *MockWorkflowEngine) GetWorkflowStatus(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WorkflowExecution), args.Error(1)
}

func TestNewManager(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.store)
	assert.NotNil(t, manager.engine)
	assert.NotNil(t, manager.logger)
}

func TestMatchStrategy_Success(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "OOMKilled",
			},
		},
	}

	strategies := []*types.Strategy{
		{
			ID:         "strategy-001",
			Name:       "OOM Strategy",
			WorkflowID: "workflow-oom",
			Active:     true,
			Symptoms: []types.Symptom{
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
			},
		},
	}

	mockStore.On("ListStrategies", ctx, true).Return(strategies, nil)

	result, err := manager.MatchStrategy(ctx, event)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "strategy-001", result.ID)
	assert.Equal(t, "OOM Strategy", result.Name)
	mockStore.AssertExpectations(t)
}

func TestMatchStrategy_MultipleStrategies(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "OOMKilled",
			},
		},
	}

	strategies := []*types.Strategy{
		{
			ID:         "strategy-001",
			Name:       "Generic Memory Strategy",
			WorkflowID: "workflow-memory",
			Active:     true,
			Symptoms: []types.Symptom{
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
			},
		},
		{
			ID:         "strategy-002",
			Name:       "Specific OOM Strategy",
			WorkflowID: "workflow-oom-specific",
			Active:     true,
			Symptoms: []types.Symptom{
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
			},
		},
	}

	mockStore.On("ListStrategies", ctx, true).Return(strategies, nil)

	result, err := manager.MatchStrategy(ctx, event)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should pick the strategy with higher score (more matching symptoms)
	assert.Equal(t, "strategy-002", result.ID)
	mockStore.AssertExpectations(t)
}

func TestMatchStrategy_NoMatch(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "UnknownReason",
			},
		},
	}

	strategies := []*types.Strategy{
		{
			ID:         "strategy-001",
			Name:       "OOM Strategy",
			WorkflowID: "workflow-oom",
			Active:     true,
			Symptoms: []types.Symptom{
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
			},
		},
	}

	mockStore.On("ListStrategies", ctx, true).Return(strategies, nil)

	result, err := manager.MatchStrategy(ctx, event)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no matching strategy found")
	mockStore.AssertExpectations(t)
}

func TestMatchStrategy_NoActiveStrategies(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "OOMKilled",
			},
		},
	}

	// Return empty strategies list
	mockStore.On("ListStrategies", ctx, true).Return([]*types.Strategy{}, nil)

	result, err := manager.MatchStrategy(ctx, event)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no matching strategy found")
	mockStore.AssertExpectations(t)
}

func TestMatchStrategy_StoreError(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
	}

	mockStore.On("ListStrategies", ctx, true).Return(nil, assert.AnError)

	result, err := manager.MatchStrategy(ctx, event)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list strategies")
	mockStore.AssertExpectations(t)
}

func TestExecuteStrategy_Success(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	strategy := &types.Strategy{
		ID:         "strategy-001",
		Name:       "OOM Strategy",
		WorkflowID: "workflow-oom",
		Active:     true,
	}

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "OOMKilled",
			},
		},
	}

	expectedExecution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-oom",
		Status:     "running",
	}

	params := map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	}

	mockEngine.On("StartWorkflow", ctx, "workflow-oom", params).Return(expectedExecution, nil)

	result, err := manager.ExecuteStrategy(ctx, strategy, event)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "exec-001", result.ID)
	assert.Equal(t, "running", result.Status)
	mockEngine.AssertExpectations(t)
}

func TestExecuteStrategy_WorkflowFailure(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	strategy := &types.Strategy{
		ID:         "strategy-001",
		Name:       "OOM Strategy",
		WorkflowID: "workflow-oom",
		Active:     true,
	}

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
	}

	params := map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	}

	mockEngine.On("StartWorkflow", ctx, "workflow-oom", params).Return(nil, assert.AnError)

	result, err := manager.ExecuteStrategy(ctx, strategy, event)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockEngine.AssertExpectations(t)
}

func TestCalculateMatchScore(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)

	tests := []struct {
		name     string
		event    types.InternalEvent
		strategy *types.Strategy
		expected int
	}{
		{
			name: "Single symptom match",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "OOMKilled",
					},
				},
			},
			strategy: &types.Strategy{
				Symptoms: []types.Symptom{
					{
						Type:    "event",
						Pattern: "OOMKilled",
					},
				},
			},
			expected: 10,
		},
		{
			name: "Multiple symptom matches",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "OOMKilled",
					},
				},
			},
			strategy: &types.Strategy{
				Symptoms: []types.Symptom{
					{
						Type:    "event",
						Pattern: "OOMKilled",
					},
					{
						Type:    "event",
						Pattern: "OOMKilled",
					},
				},
			},
			expected: 20,
		},
		{
			name: "No symptom match",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "BackOff",
					},
				},
			},
			strategy: &types.Strategy{
				Symptoms: []types.Symptom{
					{
						Type:    "event",
						Pattern: "OOMKilled",
					},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := manager.calculateMatchScore(tt.event, tt.strategy)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestSymptomMatches(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)

	tests := []struct {
		name     string
		event    types.InternalEvent
		symptom  types.Symptom
		expected bool
	}{
		{
			name: "Event type matches",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "OOMKilled",
					},
				},
			},
			symptom: types.Symptom{
				Type:    "event",
				Pattern: "OOMKilled",
			},
			expected: true,
		},
		{
			name: "Event type does not match",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "BackOff",
					},
				},
			},
			symptom: types.Symptom{
				Type:    "event",
				Pattern: "OOMKilled",
			},
			expected: false,
		},
		{
			name: "Invalid payload structure",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"invalid": "data",
				},
			},
			symptom: types.Symptom{
				Type:    "event",
				Pattern: "OOMKilled",
			},
			expected: false,
		},
		{
			name: "Non-event symptom type",
			event: types.InternalEvent{
				Payload: map[string]interface{}{
					"event": map[string]interface{}{
						"reason": "OOMKilled",
					},
				},
			},
			symptom: types.Symptom{
				Type:    "metric",
				Pattern: "high_memory",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := manager.symptomMatches(tt.event, tt.symptom)
			assert.Equal(t, tt.expected, matches)
		})
	}
}

// Benchmark tests
func BenchmarkMatchStrategy(b *testing.B) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
		Payload: map[string]interface{}{
			"event": map[string]interface{}{
				"reason": "OOMKilled",
			},
		},
	}

	strategies := []*types.Strategy{
		{
			ID:         "strategy-001",
			Name:       "OOM Strategy",
			WorkflowID: "workflow-oom",
			Active:     true,
			Symptoms: []types.Symptom{
				{
					Type:    "event",
					Pattern: "OOMKilled",
				},
			},
		},
	}

	mockStore.On("ListStrategies", ctx, true).Return(strategies, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.MatchStrategy(ctx, event)
	}
}

func BenchmarkExecuteStrategy(b *testing.B) {
	mockStore := new(MockPostgresStore)
	mockEngine := new(MockWorkflowEngine)
	logger := zap.NewNop()

	manager := NewManager(mockStore, mockEngine, logger)
	ctx := context.Background()

	strategy := &types.Strategy{
		ID:         "strategy-001",
		Name:       "OOM Strategy",
		WorkflowID: "workflow-oom",
		Active:     true,
	}

	event := types.InternalEvent{
		ID:   "event-001",
		Type: "k8s_event",
	}

	execution := &types.WorkflowExecution{
		ID:         "exec-001",
		WorkflowID: "workflow-oom",
		Status:     "running",
	}

	params := map[string]interface{}{
		"strategy_id": strategy.ID,
		"event":       event,
	}

	mockEngine.On("StartWorkflow", ctx, "workflow-oom", params).Return(execution, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.ExecuteStrategy(ctx, strategy, event)
	}
}
