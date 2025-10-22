package command

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/pkg/types"
)

// MockNATSServer is a mock for NATS Server
type MockNATSServer struct {
	mock.Mock
}

func (m *MockNATSServer) PublishCommand(clusterID string, cmd *types.Command) error {
	args := m.Called(clusterID, cmd)
	return args.Error(0)
}

func (m *MockNATSServer) SubscribeCommands(clusterID string, handler func(*types.Command)) error {
	args := m.Called(clusterID, handler)
	return args.Error(0)
}

// MockAgentRegistry is a mock for Agent Registry
type MockAgentRegistry struct {
	mock.Mock
}

func (m *MockAgentRegistry) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	args := m.Called(ctx, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Agent), args.Error(1)
}

// MockPostgresStore for commands
type MockCommandStore struct {
	mock.Mock
}

func (m *MockCommandStore) SaveCommand(ctx context.Context, cmd *types.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockCommandStore) UpdateCommand(ctx context.Context, cmd *types.Command) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockCommandStore) GetCommand(ctx context.Context, commandID string) (*types.Command, error) {
	args := m.Called(ctx, commandID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Command), args.Error(1)
}

func (m *MockCommandStore) ListCommands(ctx context.Context, filters map[string]interface{}) ([]*types.Command, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Command), args.Error(1)
}

// MockRedisStore for commands
type MockCommandCache struct {
	mock.Mock
}

func (m *MockCommandCache) CacheCommand(ctx context.Context, cmd *types.Command, ttl time.Duration) error {
	args := m.Called(ctx, cmd, ttl)
	return args.Error(0)
}

func TestNewDispatcher(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)

	assert.NotNil(t, dispatcher)
	assert.NotNil(t, dispatcher.pendingCommands)
	assert.NotNil(t, dispatcher.commandTimeouts)
}

func TestDispatchCommand_Success(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	cmd := &types.Command{
		ClusterID: "cluster-001",
		Type:      "diagnostic",
		Tool:      "kubectl",
		Action:    "get",
		Args:      []string{"pods"},
		IssuedBy:  "orchestrator",
	}

	targetAgent := &types.Agent{
		ID:        "agent-001",
		ClusterID: "cluster-001",
		Status:    types.AgentStatusOnline,
	}

	// Setup mocks
	mockRegistry.On("GetAgentByClusterID", ctx, "cluster-001").Return(targetAgent, nil)
	mockStore.On("SaveCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)
	mockNATS.On("PublishCommand", "cluster-001", mock.AnythingOfType("*types.Command")).Return(nil)

	err := dispatcher.DispatchCommand(ctx, cmd)

	assert.NoError(t, err)
	assert.NotEmpty(t, cmd.ID)
	assert.Equal(t, types.CommandStatusPending, cmd.Status)
	assert.NotZero(t, cmd.CreatedAt)
	mockRegistry.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockNATS.AssertExpectations(t)
}

func TestDispatchCommand_AgentOffline(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	cmd := &types.Command{
		ClusterID: "cluster-001",
		Type:      "diagnostic",
		Tool:      "kubectl",
		Action:    "get",
	}

	offlineAgent := &types.Agent{
		ID:        "agent-001",
		ClusterID: "cluster-001",
		Status:    types.AgentStatusOffline,
	}

	// Setup mock
	mockRegistry.On("GetAgentByClusterID", ctx, "cluster-001").Return(offlineAgent, nil)

	err := dispatcher.DispatchCommand(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "offline")
	mockRegistry.AssertExpectations(t)
}

func TestDispatchCommand_ClusterNotFound(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	cmd := &types.Command{
		ClusterID: "non-existent-cluster",
		Type:      "diagnostic",
		Tool:      "kubectl",
		Action:    "get",
	}

	// Setup mock to return error
	mockRegistry.On("GetAgentByClusterID", ctx, "non-existent-cluster").Return(nil, assert.AnError)

	err := dispatcher.DispatchCommand(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockRegistry.AssertExpectations(t)
}

func TestDispatchCommand_ValidationFailure(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	// Invalid command - missing ClusterID
	cmd := &types.Command{
		Type:   "diagnostic",
		Tool:   "kubectl",
		Action: "get",
	}

	err := dispatcher.DispatchCommand(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestGetCommand(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	commandID := "cmd-001"
	expectedCmd := &types.Command{
		ID:        commandID,
		ClusterID: "cluster-001",
		Type:      "diagnostic",
		Status:    types.CommandStatusCompleted,
	}

	// Setup mock
	mockStore.On("GetCommand", ctx, commandID).Return(expectedCmd, nil)

	cmd, err := dispatcher.GetCommand(ctx, commandID)

	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, commandID, cmd.ID)
	mockStore.AssertExpectations(t)
}

func TestUpdateCommandStatus(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	commandID := "cmd-002"
	cmd := &types.Command{
		ID:        commandID,
		ClusterID: "cluster-001",
		Status:    types.CommandStatusPending,
	}

	// Add to pending commands
	dispatcher.pendingCommands[commandID] = cmd

	// Setup mock
	mockStore.On("UpdateCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)

	err := dispatcher.UpdateCommandStatus(ctx, commandID, types.CommandStatusCompleted, "Success", nil)

	assert.NoError(t, err)
	assert.Equal(t, types.CommandStatusCompleted, cmd.Status)
	mockStore.AssertExpectations(t)
}

func TestCancelCommand(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	commandID := "cmd-003"
	cmd := &types.Command{
		ID:        commandID,
		ClusterID: "cluster-001",
		Status:    types.CommandStatusRunning,
	}

	// Add to pending commands
	dispatcher.pendingCommands[commandID] = cmd

	// Setup mock
	mockStore.On("UpdateCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)

	err := dispatcher.CancelCommand(ctx, commandID)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestListCommands(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	filters := map[string]interface{}{
		"cluster_id": "cluster-001",
		"status":     "pending",
	}

	expectedCommands := []*types.Command{
		{ID: "cmd-1", ClusterID: "cluster-001", Status: types.CommandStatusPending},
		{ID: "cmd-2", ClusterID: "cluster-001", Status: types.CommandStatusPending},
	}

	// Setup mock
	mockStore.On("ListCommands", ctx, filters).Return(expectedCommands, nil)

	commands, err := dispatcher.ListCommands(ctx, filters)

	assert.NoError(t, err)
	assert.Len(t, commands, 2)
	mockStore.AssertExpectations(t)
}

func TestGetMetrics(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)

	// Simulate command metrics
	dispatcher.commandsIssued = 100
	dispatcher.commandsCompleted = 85
	dispatcher.commandsFailed = 10
	dispatcher.commandsTimeout = 5

	metrics := dispatcher.GetMetrics()

	assert.Equal(t, int64(100), metrics["commands_issued"])
	assert.Equal(t, int64(85), metrics["commands_completed"])
	assert.Equal(t, int64(10), metrics["commands_failed"])
	assert.Equal(t, int64(5), metrics["commands_timeout"])
}

func TestCommandTimeout(t *testing.T) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	cmd := &types.Command{
		ID:        "cmd-timeout",
		ClusterID: "cluster-001",
		Type:      "diagnostic",
		Tool:      "kubectl",
		Action:    "get",
		Timeout:   100 * time.Millisecond, // Very short timeout for testing
		IssuedBy:  "test",
	}

	targetAgent := &types.Agent{
		ID:        "agent-001",
		ClusterID: "cluster-001",
		Status:    types.AgentStatusOnline,
	}

	// Setup mocks
	mockRegistry.On("GetAgentByClusterID", ctx, "cluster-001").Return(targetAgent, nil)
	mockStore.On("SaveCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)
	mockStore.On("UpdateCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)
	mockNATS.On("PublishCommand", "cluster-001", mock.AnythingOfType("*types.Command")).Return(nil)

	err := dispatcher.DispatchCommand(ctx, cmd)
	assert.NoError(t, err)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Check if command was marked as timeout
	dispatcher.mu.RLock()
	_, exists := dispatcher.pendingCommands[cmd.ID]
	dispatcher.mu.RUnlock()

	// Command should be removed from pending after timeout
	assert.False(t, exists)
}

// Benchmark tests
func BenchmarkDispatchCommand(b *testing.B) {
	mockStore := new(MockCommandStore)
	mockCache := new(MockCommandCache)
	mockRegistry := new(MockAgentRegistry)
	mockNATS := new(MockNATSServer)
	logger := zap.NewNop()

	dispatcher := NewDispatcher(mockStore, mockCache, mockRegistry, mockNATS, logger)
	ctx := context.Background()

	targetAgent := &types.Agent{
		ID:        "agent-001",
		ClusterID: "cluster-001",
		Status:    types.AgentStatusOnline,
	}

	// Setup mocks
	mockRegistry.On("GetAgentByClusterID", ctx, "cluster-001").Return(targetAgent, nil)
	mockStore.On("SaveCommand", ctx, mock.AnythingOfType("*types.Command")).Return(nil)
	mockNATS.On("PublishCommand", "cluster-001", mock.AnythingOfType("*types.Command")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := &types.Command{
			ClusterID: "cluster-001",
			Type:      "diagnostic",
			Tool:      "kubectl",
			Action:    "get",
		}
		_ = dispatcher.DispatchCommand(ctx, cmd)
	}
}
