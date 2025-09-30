package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/kart-io/k8s-agent/agent-manager/internal/storage"
	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// Registry manages agent lifecycle and state
type Registry struct {
	store       *storage.PostgresStore
	cache       *storage.RedisStore
	logger      *zap.Logger
	mu          sync.RWMutex
	agents      map[string]*types.Agent // In-memory cache
	stopCh      chan struct{}
	wg          sync.WaitGroup

	// Configuration
	heartbeatTimeout time.Duration
	cleanupInterval  time.Duration

	// Metrics
	registrationCount int64
	heartbeatCount    int64
}

// NewRegistry creates a new agent registry
func NewRegistry(
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	logger *zap.Logger,
) *Registry {
	return &Registry{
		store:            store,
		cache:            cache,
		logger:           logger.With(zap.String("component", "agent-registry")),
		agents:           make(map[string]*types.Agent),
		stopCh:           make(chan struct{}),
		heartbeatTimeout: 60 * time.Second, // 2x heartbeat interval
		cleanupInterval:  30 * time.Second,
	}
}

// Start starts the registry background tasks
func (r *Registry) Start(ctx context.Context) error {
	r.logger.Info("Starting agent registry")

	// Load existing agents from database
	if err := r.loadAgents(ctx); err != nil {
		return fmt.Errorf("failed to load agents: %w", err)
	}

	// Start background tasks
	r.wg.Add(2)
	go r.heartbeatMonitor()
	go r.cleanupStaleAgents()

	return nil
}

// Stop stops the registry
func (r *Registry) Stop() error {
	r.logger.Info("Stopping agent registry")
	close(r.stopCh)
	r.wg.Wait()
	return nil
}

// RegisterAgent registers a new agent or updates existing one
func (r *Registry) RegisterAgent(ctx context.Context, agent *types.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent.Status = types.AgentStatusOnline
	agent.LastHeartbeat = time.Now()

	// Check if agent already exists
	existing, err := r.store.GetAgentByClusterID(ctx, agent.ClusterID)
	if err == nil && existing != nil {
		// Update existing agent
		agent.ID = existing.ID
		agent.RegisteredAt = existing.RegisteredAt
		agent.UpdatedAt = time.Now()

		r.logger.Info("Agent re-registered",
			zap.String("agent_id", agent.ID),
			zap.String("cluster_id", agent.ClusterID))
	} else {
		// New agent
		agent.RegisteredAt = time.Now()
		agent.UpdatedAt = time.Now()

		if agent.ConnectionInfo == nil {
			agent.ConnectionInfo = &types.ConnectionInfo{}
		}
		agent.ConnectionInfo.ConnectedAt = time.Now()
		agent.ConnectionInfo.LastSeen = time.Now()

		r.logger.Info("New agent registered",
			zap.String("agent_id", agent.ID),
			zap.String("cluster_id", agent.ClusterID))

		r.registrationCount++
	}

	// Save to database
	if err := r.store.SaveAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to save agent: %w", err)
	}

	// Cache in Redis (30-minute TTL)
	if err := r.cache.CacheAgent(ctx, agent, 30*time.Minute); err != nil {
		r.logger.Warn("Failed to cache agent", zap.Error(err))
	}

	// Mark as online in Redis (2-minute TTL)
	if err := r.cache.SetAgentOnline(ctx, agent.ID, 2*time.Minute); err != nil {
		r.logger.Warn("Failed to set agent online", zap.Error(err))
	}

	// Store in memory
	r.agents[agent.ID] = agent

	return nil
}

// UnregisterAgent unregisters an agent
func (r *Registry) UnregisterAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update status to offline
	if err := r.store.UpdateAgentStatus(ctx, agentID, types.AgentStatusOffline); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	// Remove from cache
	r.cache.DeleteCachedAgent(ctx, agentID)

	// Remove from memory
	delete(r.agents, agentID)

	r.logger.Info("Agent unregistered", zap.String("agent_id", agentID))

	return nil
}

// UpdateHeartbeat updates agent heartbeat timestamp
func (r *Registry) UpdateHeartbeat(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update in database
	if err := r.store.UpdateAgentHeartbeat(ctx, agentID); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// Update in memory
	if agent, ok := r.agents[agentID]; ok {
		agent.LastHeartbeat = time.Now()
		agent.Status = types.AgentStatusOnline
		if agent.ConnectionInfo != nil {
			agent.ConnectionInfo.LastSeen = time.Now()
		}
	}

	// Extend TTL in Redis
	r.cache.SetAgentOnline(ctx, agentID, 2*time.Minute)

	r.heartbeatCount++

	return nil
}

// GetAgent retrieves agent by ID
func (r *Registry) GetAgent(ctx context.Context, agentID string) (*types.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check memory cache first
	if agent, ok := r.agents[agentID]; ok {
		return agent, nil
	}

	// Check Redis cache
	agent, err := r.cache.GetCachedAgent(ctx, agentID)
	if err != nil {
		r.logger.Warn("Failed to get agent from cache", zap.Error(err))
	}
	if agent != nil {
		// Add to memory cache
		r.agents[agentID] = agent
		return agent, nil
	}

	// Fallback to database
	agent, err = r.store.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Cache in memory
	r.agents[agentID] = agent

	return agent, nil
}

// GetAgentByClusterID retrieves agent by cluster ID
func (r *Registry) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Search in memory cache
	for _, agent := range r.agents {
		if agent.ClusterID == clusterID {
			return agent, nil
		}
	}

	// Fallback to database
	return r.store.GetAgentByClusterID(ctx, clusterID)
}

// ListAgents lists all agents with optional status filter
func (r *Registry) ListAgents(ctx context.Context, status *types.AgentStatus) ([]*types.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If no filter, return from memory
	if status == nil {
		agents := make([]*types.Agent, 0, len(r.agents))
		for _, agent := range r.agents {
			agents = append(agents, agent)
		}
		return agents, nil
	}

	// With filter, query database
	return r.store.ListAgents(ctx, status)
}

// GetOnlineAgents returns list of online agents
func (r *Registry) GetOnlineAgents(ctx context.Context) ([]*types.Agent, error) {
	status := types.AgentStatusOnline
	return r.ListAgents(ctx, &status)
}

// GetAgentCount returns count of agents by status
func (r *Registry) GetAgentCount(ctx context.Context, status *types.AgentStatus) (int, error) {
	agents, err := r.ListAgents(ctx, status)
	if err != nil {
		return 0, err
	}
	return len(agents), nil
}

// loadAgents loads agents from database into memory
func (r *Registry) loadAgents(ctx context.Context) error {
	agents, err := r.store.ListAgents(ctx, nil)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, agent := range agents {
		r.agents[agent.ID] = agent
	}

	r.logger.Info("Loaded agents from database", zap.Int("count", len(agents)))

	return nil
}

// heartbeatMonitor monitors agent heartbeats
func (r *Registry) heartbeatMonitor() {
	defer r.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.checkHeartbeats()
		}
	}
}

// checkHeartbeats checks for stale heartbeats
func (r *Registry) checkHeartbeats() {
	ctx := context.Background()
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	for agentID, agent := range r.agents {
		// Check if heartbeat is stale
		if now.Sub(agent.LastHeartbeat) > r.heartbeatTimeout {
			if agent.Status == types.AgentStatusOnline {
				r.logger.Warn("Agent heartbeat timeout",
					zap.String("agent_id", agentID),
					zap.String("cluster_id", agent.ClusterID),
					zap.Duration("last_heartbeat", now.Sub(agent.LastHeartbeat)))

				// Update status to offline
				agent.Status = types.AgentStatusOffline
				agent.UpdatedAt = now

				// Update in database
				if err := r.store.UpdateAgentStatus(ctx, agentID, types.AgentStatusOffline); err != nil {
					r.logger.Error("Failed to update agent status",
						zap.String("agent_id", agentID),
						zap.Error(err))
				}
			}
		}
	}
}

// cleanupStaleAgents removes old offline agents
func (r *Registry) cleanupStaleAgents() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.performCleanup()
		}
	}
}

// performCleanup removes agents that have been offline for too long
func (r *Registry) performCleanup() {
	ctx := context.Background()
	now := time.Now()
	threshold := 24 * time.Hour // Remove agents offline for 24+ hours

	r.mu.Lock()
	defer r.mu.Unlock()

	for agentID, agent := range r.agents {
		if agent.Status == types.AgentStatusOffline {
			offlineDuration := now.Sub(agent.LastHeartbeat)
			if offlineDuration > threshold {
				r.logger.Info("Cleaning up stale agent",
					zap.String("agent_id", agentID),
					zap.Duration("offline_duration", offlineDuration))

				// Delete from database
				if err := r.store.DeleteAgent(ctx, agentID); err != nil {
					r.logger.Error("Failed to delete agent",
						zap.String("agent_id", agentID),
						zap.Error(err))
					continue
				}

				// Delete from cache
				r.cache.DeleteCachedAgent(ctx, agentID)

				// Remove from memory
				delete(r.agents, agentID)
			}
		}
	}
}

// GetStatistics returns registry statistics
func (r *Registry) GetStatistics() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	onlineCount := 0
	offlineCount := 0

	for _, agent := range r.agents {
		if agent.Status == types.AgentStatusOnline {
			onlineCount++
		} else {
			offlineCount++
		}
	}

	return map[string]interface{}{
		"total_agents":        len(r.agents),
		"online_agents":       onlineCount,
		"offline_agents":      offlineCount,
		"registration_count":  r.registrationCount,
		"heartbeat_count":     r.heartbeatCount,
		"heartbeat_timeout":   r.heartbeatTimeout.String(),
	}
}