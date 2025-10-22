package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/pkg/types"
)

// MockPostgresStore is a mock for PostgresStore
type MockPostgresStore struct {
	mock.Mock
}

func (m *MockPostgresStore) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	args := m.Called(ctx, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Agent), args.Error(1)
}

func (m *MockPostgresStore) SaveAgent(ctx context.Context, agent *types.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockPostgresStore) UpdateAgentStatus(ctx context.Context, agentID string, status types.AgentStatus) error {
	args := m.Called(ctx, agentID, status)
	return args.Error(0)
}

func (m *MockPostgresStore) ListAgents(ctx context.Context, filters map[string]interface{}) ([]*types.Agent, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Agent), args.Error(1)
}

// MockRedisStore is a mock for RedisStore
type MockRedisStore struct {
	mock.Mock
}

func (m *MockRedisStore) SetAgentHeartbeat(ctx context.Context, agentID string, timestamp time.Time) error {
	args := m.Called(ctx, agentID, timestamp)
	return args.Error(0)
}

func (m *MockRedisStore) GetAgentHeartbeat(ctx context.Context, agentID string) (time.Time, error) {
	args := m.Called(ctx, agentID)
	return args.Get(0).(time.Time), args.Error(1)
}

func TestNewRegistry(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.agents)
	assert.Equal(t, 60*time.Second, registry.heartbeatTimeout)
	assert.Equal(t, 30*time.Second, registry.cleanupInterval)
}

func TestRegisterAgent_NewAgent(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agent := &types.Agent{
		ClusterID:   "test-cluster-001",
		ClusterName: "Test Cluster",
		Version:     "v1.0.0",
	}

	// Mock: agent doesn't exist
	mockStore.On("GetAgentByClusterID", ctx, "test-cluster-001").Return(nil, nil)
	mockStore.On("SaveAgent", ctx, mock.AnythingOfType("*types.Agent")).Return(nil)
	mockCache.On("SetAgentHeartbeat", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	err := registry.RegisterAgent(ctx, agent)

	assert.NoError(t, err)
	assert.Equal(t, types.AgentStatusOnline, agent.Status)
	assert.NotZero(t, agent.LastHeartbeat)
	assert.NotZero(t, agent.RegisteredAt)
	mockStore.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestRegisterAgent_ExistingAgent(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	existingAgent := &types.Agent{
		ID:           "agent-123",
		ClusterID:    "test-cluster-001",
		RegisteredAt: time.Now().Add(-1 * time.Hour),
	}

	newAgent := &types.Agent{
		ClusterID:   "test-cluster-001",
		ClusterName: "Test Cluster Updated",
		Version:     "v1.1.0",
	}

	// Mock: agent exists
	mockStore.On("GetAgentByClusterID", ctx, "test-cluster-001").Return(existingAgent, nil)
	mockStore.On("SaveAgent", ctx, mock.AnythingOfType("*types.Agent")).Return(nil)
	mockCache.On("SetAgentHeartbeat", ctx, "agent-123", mock.AnythingOfType("time.Time")).Return(nil)

	err := registry.RegisterAgent(ctx, newAgent)

	assert.NoError(t, err)
	assert.Equal(t, existingAgent.ID, newAgent.ID)
	assert.Equal(t, existingAgent.RegisteredAt, newAgent.RegisteredAt)
	assert.Equal(t, types.AgentStatusOnline, newAgent.Status)
	mockStore.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestHeartbeat(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agentID := "agent-123"

	// Mock
	mockCache.On("SetAgentHeartbeat", ctx, agentID, mock.AnythingOfType("time.Time")).Return(nil)
	mockStore.On("UpdateAgentStatus", ctx, agentID, types.AgentStatusOnline).Return(nil)

	// Add agent to in-memory cache
	registry.agents[agentID] = &types.Agent{
		ID:        agentID,
		ClusterID: "test-cluster",
		Status:    types.AgentStatusOffline,
	}

	err := registry.Heartbeat(ctx, agentID)

	assert.NoError(t, err)
	assert.Equal(t, types.AgentStatusOnline, registry.agents[agentID].Status)
	mockCache.AssertExpectations(t)
}

func TestGetAgent(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agentID := "agent-123"
	expectedAgent := &types.Agent{
		ID:          agentID,
		ClusterID:   "test-cluster",
		ClusterName: "Test Cluster",
		Status:      types.AgentStatusOnline,
	}

	// Add to in-memory cache
	registry.agents[agentID] = expectedAgent

	agent, err := registry.GetAgent(ctx, agentID)

	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, agentID, agent.ID)
	assert.Equal(t, "test-cluster", agent.ClusterID)
}

func TestGetAgent_NotFound(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agent, err := registry.GetAgent(ctx, "non-existent-agent")

	assert.Error(t, err)
	assert.Nil(t, agent)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestListAgents(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)

	// Add multiple agents to in-memory cache
	registry.agents["agent-1"] = &types.Agent{ID: "agent-1", ClusterID: "cluster-1"}
	registry.agents["agent-2"] = &types.Agent{ID: "agent-2", ClusterID: "cluster-2"}
	registry.agents["agent-3"] = &types.Agent{ID: "agent-3", ClusterID: "cluster-3"}

	agents := registry.ListAgents()

	assert.Len(t, agents, 3)
}

func TestUnregisterAgent(t *testing.T) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agentID := "agent-123"

	// Add agent to in-memory cache
	registry.agents[agentID] = &types.Agent{
		ID:        agentID,
		ClusterID: "test-cluster",
	}

	// Mock
	mockStore.On("UpdateAgentStatus", ctx, agentID, types.AgentStatusOffline).Return(nil)

	err := registry.UnregisterAgent(ctx, agentID)

	assert.NoError(t, err)
	_, exists := registry.agents[agentID]
	assert.False(t, exists)
	mockStore.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkRegisterAgent(b *testing.B) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	// Setup mocks
	mockStore.On("GetAgentByClusterID", ctx, mock.AnythingOfType("string")).Return(nil, nil)
	mockStore.On("SaveAgent", ctx, mock.AnythingOfType("*types.Agent")).Return(nil)
	mockCache.On("SetAgentHeartbeat", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent := &types.Agent{
			ClusterID:   "test-cluster-001",
			ClusterName: "Test Cluster",
			Version:     "v1.0.0",
		}
		_ = registry.RegisterAgent(ctx, agent)
	}
}

func BenchmarkHeartbeat(b *testing.B) {
	mockStore := new(MockPostgresStore)
	mockCache := new(MockRedisStore)
	logger := zap.NewNop()

	registry := NewRegistry(mockStore, mockCache, logger)
	ctx := context.Background()

	agentID := "agent-123"
	registry.agents[agentID] = &types.Agent{
		ID:        agentID,
		ClusterID: "test-cluster",
	}

	// Setup mocks
	mockCache.On("SetAgentHeartbeat", ctx, agentID, mock.AnythingOfType("time.Time")).Return(nil)
	mockStore.On("UpdateAgentStatus", ctx, agentID, types.AgentStatusOnline).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.Heartbeat(ctx, agentID)
	}
}
