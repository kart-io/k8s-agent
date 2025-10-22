package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
)

// MockPostgresStore is a mock for PostgresStore
type MockPostgresStore struct {
	mock.Mock
}

func (m *MockPostgresStore) GetWorkflow(ctx context.Context, workflowID string) (*types.Workflow, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Workflow), args.Error(1)
}

func (m *MockPostgresStore) SaveWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockPostgresStore) UpdateWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockPostgresStore) GetWorkflowExecution(ctx context.Context, executionID string) (*types.WorkflowExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.WorkflowExecution), args.Error(1)
}

// MockRedisStore is a mock for RedisStore
type MockRedisStore struct {
	mock.Mock
}

func (m *MockRedisStore) CacheWorkflow(ctx context.Context, workflow *types.Workflow, ttl time.Duration) error {
	args := m.Called(ctx, workflow, ttl)
	return args.Error(0)
}

func (m *MockRedisStore) GetCachedWorkflow(ctx context.Context, workflowID string) (*types.Workflow, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Workflow), args.Error(1)
}

// MockExecutor is a mock for Executor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) ExecuteCommand(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	args := m.Called(ctx, execution, step)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockExecutor) ExecuteAIAnalysis(ctx context.Context, execution *types.WorkflowExecution, step types.WorkflowStep) (map[string]interface{}, error) {
	args := m.Called(ctx, execution, step)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestNewEngine(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.executions)
	assert.Equal(t, int64(0), engine.executionsStarted)
}

func TestStartWorkflow(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	workflowID := "workflow-001"
	triggerEvent := map[string]interface{}{
		"type":   "pod_failure",
		"reason": "CrashLoopBackOff",
	}

	workflow := &types.Workflow{
		ID:     workflowID,
		Name:   "Diagnose Pod CrashLoop",
		Status: types.WorkflowStatusActive,
		Steps: []types.WorkflowStep{
			{
				ID:   "step1",
				Type: types.StepTypeCommand,
				Name: "Collect logs",
			},
		},
	}

	// Setup mocks
	mockStore.On("GetWorkflow", ctx, workflowID).Return(workflow, nil)
	mockStore.On("SaveWorkflowExecution", ctx, mock.AnythingOfType("*types.WorkflowExecution")).Return(nil)

	execution, err := engine.StartWorkflow(ctx, workflowID, triggerEvent)

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, workflowID, execution.WorkflowID)
	assert.Equal(t, types.ExecutionStatusPending, execution.Status)
	assert.NotEmpty(t, execution.ID)
	mockStore.AssertExpectations(t)

	// Allow async execution to start
	time.Sleep(50 * time.Millisecond)
}

func TestStartWorkflow_InactiveWorkflow(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	workflowID := "workflow-002"

	workflow := &types.Workflow{
		ID:     workflowID,
		Name:   "Inactive Workflow",
		Status: types.WorkflowStatusInactive,
	}

	// Setup mock
	mockStore.On("GetWorkflow", ctx, workflowID).Return(workflow, nil)

	execution, err := engine.StartWorkflow(ctx, workflowID, nil)

	assert.Error(t, err)
	assert.Nil(t, execution)
	assert.Contains(t, err.Error(), "not active")
	mockStore.AssertExpectations(t)
}

func TestStartWorkflow_WorkflowNotFound(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	workflowID := "non-existent-workflow"

	// Setup mock to return error
	mockStore.On("GetWorkflow", ctx, workflowID).Return(nil, assert.AnError)

	execution, err := engine.StartWorkflow(ctx, workflowID, nil)

	assert.Error(t, err)
	assert.Nil(t, execution)
	mockStore.AssertExpectations(t)
}

func TestGetExecution(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	executionID := "exec-001"
	expectedExecution := &types.WorkflowExecution{
		ID:         executionID,
		WorkflowID: "workflow-001",
		Status:     types.ExecutionStatusRunning,
		StartedAt:  time.Now(),
	}

	// Setup mock
	mockStore.On("GetWorkflowExecution", ctx, executionID).Return(expectedExecution, nil)

	execution, err := engine.GetExecution(ctx, executionID)

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, executionID, execution.ID)
	mockStore.AssertExpectations(t)
}

func TestCancelExecution(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	executionID := "exec-002"

	// Add execution to in-memory tracking
	execution := &types.WorkflowExecution{
		ID:         executionID,
		WorkflowID: "workflow-001",
		Status:     types.ExecutionStatusRunning,
	}
	engine.executions[executionID] = execution

	// Setup mock
	mockStore.On("UpdateWorkflowExecution", ctx, mock.AnythingOfType("*types.WorkflowExecution")).Return(nil)

	err := engine.CancelExecution(ctx, executionID)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestListExecutions(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)

	// Add multiple executions to in-memory tracking
	engine.executions["exec-1"] = &types.WorkflowExecution{ID: "exec-1"}
	engine.executions["exec-2"] = &types.WorkflowExecution{ID: "exec-2"}
	engine.executions["exec-3"] = &types.WorkflowExecution{ID: "exec-3"}

	executions := engine.ListExecutions()

	assert.Len(t, executions, 3)
}

func TestGetMetrics(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)

	// Simulate executions
	engine.executionsStarted = 100
	engine.executionsCompleted = 85
	engine.executionsFailed = 15

	metrics := engine.GetMetrics()

	assert.Equal(t, int64(100), metrics["executions_started"])
	assert.Equal(t, int64(85), metrics["executions_completed"])
	assert.Equal(t, int64(15), metrics["executions_failed"])
}

// Benchmark tests
func BenchmarkStartWorkflow(b *testing.B) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	mockExecutor := new(MockExecutor)
	logger := zap.NewNop()

	engine := NewEngine(mockStore, mockCache, mockExecutor, logger)
	ctx := context.Background()

	workflow := &types.Workflow{
		ID:     "benchmark-workflow",
		Name:   "Benchmark",
		Status: types.WorkflowStatusActive,
		Steps:  []types.WorkflowStep{},
	}

	// Setup mocks
	mockStore.On("GetWorkflow", ctx, "benchmark-workflow").Return(workflow, nil)
	mockStore.On("SaveWorkflowExecution", ctx, mock.AnythingOfType("*types.WorkflowExecution")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.StartWorkflow(ctx, "benchmark-workflow", nil)
	}
}
