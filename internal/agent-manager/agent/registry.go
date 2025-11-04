package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kart-io/k8s-agent/internal/agent-manager/constants"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// Registry manages agent lifecycle and state.
type Registry struct {
	store  *storage.PostgresStore
	cache  *storage.RedisStore
	logger core.Logger
	mu     sync.RWMutex
	agents map[string]*types.Agent // In-memory cache
	stopCh chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context // Lifecycle context
	cancel context.CancelFunc

	// Configuration
	heartbeatTimeout time.Duration
	cleanupInterval  time.Duration

	// Metrics
	registrationCount int64
	heartbeatCount    int64
}

// NewRegistry creates a new agent registry.
func NewRegistry(
	store *storage.PostgresStore,
	cache *storage.RedisStore,
	logger core.Logger,
) *Registry {
	return &Registry{
		store:            store,
		cache:            cache,
		logger:           logger.With("component", "agent-registry"),
		agents:           make(map[string]*types.Agent),
		stopCh:           make(chan struct{}),
		heartbeatTimeout: constants.HeartbeatTimeout,
		cleanupInterval:  constants.CleanupInterval,
	}
}

// Start starts the registry background tasks.
func (r *Registry) Start(ctx context.Context) error {
	r.logger.Info("Starting agent registry")

	// Create lifecycle context
	r.ctx, r.cancel = context.WithCancel(ctx)

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

// Stop stops the registry.
func (r *Registry) Stop() error {
	r.logger.Info("Stopping agent registry")

	// Cancel context first
	if r.cancel != nil {
		r.cancel()
	}

	// Close stop channel
	close(r.stopCh)

	// Wait for goroutines to finish
	r.wg.Wait()

	return nil
}

// RegisterAgent registers a new agent or updates existing one.
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

		r.logger.Infow("Agent re-registered",
			"agent_id", agent.ID,
			"cluster_id", agent.ClusterID)
	} else {
		// New agent - keep the ID if provided, otherwise it will be empty and database will fail
		// The ID should come from the registration message
		agent.RegisteredAt = time.Now()
		agent.UpdatedAt = time.Now()

		if agent.ConnectionInfo == nil {
			agent.ConnectionInfo = &types.ConnectionInfo{}
		}
		agent.ConnectionInfo.ConnectedAt = time.Now()
		agent.ConnectionInfo.LastSeen = time.Now()

		r.logger.Infow("New agent registered",
			"agent_id", agent.ID,
			"cluster_id", agent.ClusterID,
			"version", agent.Version)

		r.registrationCount++
	}

	// Save to database
	if err := r.store.SaveAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to save agent: %w", err)
	}

	// Cache in Redis (30-minute TTL)
	if err := r.cache.CacheAgent(ctx, agent, constants.AgentCacheTTL); err != nil {
		r.logger.Warnw("Failed to cache agent", "error", err)
	}

	// Mark as online in Redis (2-minute TTL)
	if err := r.cache.SetAgentOnline(ctx, agent.ID, constants.AgentOnlineTTL); err != nil {
		r.logger.Warnw("Failed to set agent online", "error", err)
	}

	// Store in memory
	r.agents[agent.ID] = agent

	// Create or update cluster record
	if err := r.ensureCluster(ctx, agent); err != nil {
		r.logger.Warnw("Failed to ensure cluster",
			"cluster_id", agent.ClusterID,
			"error", err)
		// Don't fail agent registration if cluster creation fails
	}

	return nil
}

// UnregisterAgent unregisters an agent.
func (r *Registry) UnregisterAgent(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update status to offline
	if err := r.store.UpdateAgentStatus(ctx, agentID, types.AgentStatusOffline); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}

	// Remove from cache
	if err := r.cache.DeleteCachedAgent(ctx, agentID); err != nil {
		r.logger.Warnw("Failed to delete agent from cache", "agent_id", agentID, "error", err)
	}

	// Remove from memory
	delete(r.agents, agentID)

	r.logger.Infow("Agent unregistered", "agent_id", agentID)

	return nil
}

// UpdateHeartbeat updates agent heartbeat timestamp.
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
	if err := r.cache.SetAgentOnline(ctx, agentID, constants.AgentOnlineTTL); err != nil {
		r.logger.Warnw("Failed to set agent online in cache", "agent_id", agentID, "error", err)
	}

	r.heartbeatCount++

	return nil
}

// GetAgent retrieves agent by ID.
func (r *Registry) GetAgent(ctx context.Context, agentID string) (*types.Agent, error) {
	r.mu.RLock()
	// Check memory cache first
	if agent, ok := r.agents[agentID]; ok {
		r.mu.RUnlock()
		// Return a copy to prevent external modification
		return r.copyAgent(agent), nil
	}
	r.mu.RUnlock()

	// Check Redis cache
	agent, err := r.cache.GetCachedAgent(ctx, agentID)
	if err != nil {
		r.logger.Warnw("Failed to get agent from cache", "error", err)
	}
	if agent != nil {
		// Add to memory cache with write lock
		r.mu.Lock()
		r.agents[agentID] = agent
		r.mu.Unlock()
		return r.copyAgent(agent), nil
	}

	// Fallback to database
	agent, err = r.store.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Cache in memory with write lock
	r.mu.Lock()
	r.agents[agentID] = agent
	r.mu.Unlock()

	return r.copyAgent(agent), nil
}

// copyAgent creates a deep copy of an agent to prevent race conditions.
func (r *Registry) copyAgent(agent *types.Agent) *types.Agent {
	if agent == nil {
		return nil
	}

	copy := *agent

	// Deep copy metadata map
	if agent.Metadata != nil {
		copy.Metadata = make(map[string]interface{}, len(agent.Metadata))
		for k, v := range agent.Metadata {
			copy.Metadata[k] = v
		}
	}

	// Deep copy capabilities slice
	if agent.Capabilities != nil {
		copy.Capabilities = make([]string, len(agent.Capabilities))
		copySlice := copy.Capabilities
		for i, v := range agent.Capabilities {
			copySlice[i] = v
		}
	}

	// Deep copy connection info
	if agent.ConnectionInfo != nil {
		connInfo := *agent.ConnectionInfo
		copy.ConnectionInfo = &connInfo
	}

	return &copy
}

// GetAgentByClusterID retrieves agent by cluster ID.
func (r *Registry) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	r.mu.RLock()
	// Search in memory cache
	for _, agent := range r.agents {
		if agent.ClusterID == clusterID {
			r.mu.RUnlock()
			return r.copyAgent(agent), nil
		}
	}
	r.mu.RUnlock()

	// Fallback to database
	agent, err := r.store.GetAgentByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	return r.copyAgent(agent), nil
}

// ListAgents lists all agents with optional status filter.
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

// GetOnlineAgents returns list of online agents.
func (r *Registry) GetOnlineAgents(ctx context.Context) ([]*types.Agent, error) {
	status := types.AgentStatusOnline
	return r.ListAgents(ctx, &status)
}

// GetAgentCount returns count of agents by status.
func (r *Registry) GetAgentCount(ctx context.Context, status *types.AgentStatus) (int, error) {
	agents, err := r.ListAgents(ctx, status)
	if err != nil {
		return 0, err
	}
	return len(agents), nil
}

// loadAgents loads agents from database into memory.
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

	r.logger.Infow("Loaded agents from database", "count", len(agents))

	return nil
}

// heartbeatMonitor monitors agent heartbeats.
func (r *Registry) heartbeatMonitor() {
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Errorw("Panic in heartbeat monitor", "panic", rec)
		}
	}()

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

// checkHeartbeats checks for stale heartbeats.
func (r *Registry) checkHeartbeats() {
	// Use lifecycle context with timeout
	ctx, cancel := context.WithTimeout(r.ctx, constants.DatabaseOperationTimeout)
	defer cancel()

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	for agentID, agent := range r.agents {
		// Check if heartbeat is stale
		if now.Sub(agent.LastHeartbeat) > r.heartbeatTimeout {
			if agent.Status == types.AgentStatusOnline {
				r.logger.Warnw("Agent heartbeat timeout",
					"agent_id", agentID,
					"cluster_id", agent.ClusterID,
					"last_heartbeat", now.Sub(agent.LastHeartbeat))

				// Update status to offline
				agent.Status = types.AgentStatusOffline
				agent.UpdatedAt = now

				// Update in database
				if err := r.store.UpdateAgentStatus(ctx, agentID, types.AgentStatusOffline); err != nil {
					r.logger.Errorw("Failed to update agent status",
						"agent_id", agentID,
						"error", err)
				}
			}
		}
	}
}

// cleanupStaleAgents removes old offline agents.
func (r *Registry) cleanupStaleAgents() {
	defer r.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			r.logger.Errorw("Panic in cleanup goroutine", "panic", rec)
		}
	}()

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

// performCleanup removes agents that have been offline for too long.
func (r *Registry) performCleanup() {
	// Use lifecycle context with timeout
	ctx, cancel := context.WithTimeout(r.ctx, constants.DatabaseOperationTimeout)
	defer cancel()

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	for agentID, agent := range r.agents {
		if agent.Status == types.AgentStatusOffline {
			offlineDuration := now.Sub(agent.LastHeartbeat)
			if offlineDuration > constants.StaleAgentThreshold {
				r.logger.Infow("Cleaning up stale agent",
					"agent_id", agentID,
					"offline_duration", offlineDuration)

				// Delete from database
				if err := r.store.DeleteAgent(ctx, agentID); err != nil {
					r.logger.Errorw("Failed to delete agent",
						"agent_id", agentID,
						"error", err)
					continue
				}

				// Delete from cache
				if err := r.cache.DeleteCachedAgent(ctx, agentID); err != nil {
					r.logger.Warnw("Failed to delete agent from cache", "agent_id", agentID, "error", err)
				}

				// Remove from memory
				delete(r.agents, agentID)
			}
		}
	}
}

// ensureCluster creates or updates a cluster record for the agent.
func (r *Registry) ensureCluster(ctx context.Context, agent *types.Agent) error {
	// Check if cluster already exists
	cluster, err := r.store.GetCluster(ctx, agent.ClusterID)
	if err == nil && cluster != nil {
		// Update existing cluster
		cluster.Status = types.ClusterStatusActive
		cluster.Health = types.ClusterHealthHealthy
		cluster.Version = agent.Version
		cluster.AgentCount = 1 // This should be counted from agents table
		cluster.UpdatedAt = time.Now()

		return r.store.SaveCluster(ctx, cluster)
	}

	// Create new cluster
	newCluster := &types.Cluster{
		ID:          agent.ClusterID,
		Name:        agent.ClusterID, // Use cluster ID as name by default
		Description: "Auto-created from agent registration",
		Environment: "unknown", // Can be updated later
		Status:      types.ClusterStatusActive,
		Health:      types.ClusterHealthHealthy,
		Version:     agent.Version,
		AgentCount:  1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return r.store.SaveCluster(ctx, newCluster)
}

// UpdateClusterInfo updates cluster K8s version and API server.
func (r *Registry) UpdateClusterInfo(ctx context.Context, clusterID, k8sVersion, apiServer string) error {
	cluster, err := r.store.GetCluster(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	// Update fields if provided
	if k8sVersion != "" {
		cluster.Version = k8sVersion
	}
	if apiServer != "" {
		cluster.APIServer = apiServer
	}
	cluster.UpdatedAt = time.Now()

	return r.store.SaveCluster(ctx, cluster)
}

// UpdateClusterMetrics updates cluster node and pod counts.
func (r *Registry) UpdateClusterMetrics(ctx context.Context, clusterID string, nodeCount, podCount int) error {
	cluster, err := r.store.GetCluster(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	if cluster == nil {
		return fmt.Errorf("cluster not found: %s", clusterID)
	}

	// Update metrics
	cluster.NodeCount = nodeCount
	cluster.PodCount = podCount
	cluster.UpdatedAt = time.Now()

	return r.store.SaveCluster(ctx, cluster)
}

// GetStatistics returns registry statistics.
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
		"total_agents":       len(r.agents),
		"online_agents":      onlineCount,
		"offline_agents":     offlineCount,
		"registration_count": r.registrationCount,
		"heartbeat_count":    r.heartbeatCount,
		"heartbeat_timeout":  r.heartbeatTimeout.String(),
	}
}
