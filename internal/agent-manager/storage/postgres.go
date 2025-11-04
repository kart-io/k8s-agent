package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

const (
	// defaultDBTimeout 数据库操作默认超时时间.
	defaultDBTimeout = 5 * time.Second
)

// withTimeout 为context添加默认超时（如果尚未设置deadline）.
func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		// context已经有deadline，直接返回
		return ctx, func() {}
	}
	// 添加默认超时
	return context.WithTimeout(ctx, defaultDBTimeout)
}

// PostgresStore implements storage using MySQL
// Note: Kept the name for backward compatibility, but now using MySQL.
type PostgresStore struct {
	*db.MySQLClient // Embed common MySQL client
	logger          core.Logger
}

// NewPostgresStore creates a new MySQL store
// Note: Kept the name for backward compatibility, but now using MySQL.
func NewPostgresStore(config types.DatabaseConfig, log core.Logger) (*PostgresStore, error) {
	// Create MySQL client using common package with Options pattern
	mysqlClient, err := db.NewMySQL(log,
		db.WithHost(config.Host),
		db.WithPort(config.Port),
		db.WithUser(config.User),
		db.WithPassword(config.Password),
		db.WithDatabase(config.Database),
		db.WithMaxOpenConns(config.MaxOpenConns),
		db.WithMaxIdleConns(config.MaxIdleConns),
		db.WithConnMaxLifetime(config.ConnMaxLifetime),
		db.WithLogLevel("info"),
	)
	if err != nil {
		return nil, err
	}

	store := &PostgresStore{
		MySQLClient: mysqlClient,
		logger:      log.With("component", "postgres"),
	}

	// Auto-migrate schemas
	if err := store.AutoMigrate(
		&types.Agent{},
		&types.Event{},
		&types.Metrics{},
		&types.Command{},
		&types.CommandResult{},
		&types.Cluster{},
		&types.AlertRule{},
		&types.Alert{},
	); err != nil {
		return nil, err
	}

	store.logger.Infow("PostgreSQL store initialized",
		"host", config.Host,
		"database", config.Database)

	return store, nil
}

// Agent operations

// SaveAgent saves an agent to the database.
func (s *PostgresStore) SaveAgent(ctx context.Context, agent *types.Agent) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return s.DB.WithContext(ctx).Save(agent).Error
}

// GetAgent retrieves an agent by ID.
func (s *PostgresStore) GetAgent(ctx context.Context, id string) (*types.Agent, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var agent types.Agent
	if err := s.DB.WithContext(ctx).First(&agent, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent %s: %w", id, err)
	}
	return &agent, nil
}

// GetAgentByClusterID retrieves an agent by cluster ID.
func (s *PostgresStore) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var agent types.Agent
	if err := s.DB.WithContext(ctx).First(&agent, "cluster_id = ?", clusterID).Error; err != nil {
		return nil, fmt.Errorf("failed to get agent by cluster_id %s: %w", clusterID, err)
	}
	return &agent, nil
}

// ListAgents lists all agents.
func (s *PostgresStore) ListAgents(ctx context.Context, status *types.AgentStatus) ([]*types.Agent, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var agents []*types.Agent
	query := s.DB.WithContext(ctx)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("registered_at DESC").Find(&agents).Error; err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return agents, nil
}

// UpdateAgentStatus updates agent status.
func (s *PostgresStore) UpdateAgentStatus(ctx context.Context, id string, status types.AgentStatus) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return s.DB.WithContext(ctx).Model(&types.Agent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateAgentHeartbeat updates agent heartbeat timestamp.
func (s *PostgresStore) UpdateAgentHeartbeat(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return s.DB.WithContext(ctx).Model(&types.Agent{}).
		Where("id = ?", id).
		Update("last_heartbeat", time.Now()).Error
}

// DeleteAgent deletes an agent.
func (s *PostgresStore) DeleteAgent(ctx context.Context, id string) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return s.DB.WithContext(ctx).Delete(&types.Agent{}, "id = ?", id).Error
}

// Event operations

// SaveEvent saves an event to the database.
func (s *PostgresStore) SaveEvent(ctx context.Context, event *types.Event) error {
	return s.DB.WithContext(ctx).Create(event).Error
}

// GetEvent retrieves an event by ID.
func (s *PostgresStore) GetEvent(ctx context.Context, id string) (*types.Event, error) {
	var event types.Event
	if err := s.DB.WithContext(ctx).First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents lists events with filters.
func (s *PostgresStore) ListEvents(ctx context.Context, filter EventFilter) ([]*types.Event, error) {
	var events []*types.Event
	query := s.DB.WithContext(ctx)

	if filter.ClusterID != "" {
		query = query.Where("cluster_id = ?", filter.ClusterID)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.Namespace != "" {
		query = query.Where("namespace = ?", filter.Namespace)
	}
	if !filter.StartTime.IsZero() {
		query = query.Where("timestamp >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		query = query.Where("timestamp <= ?", filter.EndTime)
	}

	query = query.Order("timestamp DESC").Limit(filter.Limit)
	if err := query.Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// EventFilter defines filters for event queries.
type EventFilter struct {
	ClusterID string
	Severity  string
	Namespace string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// Command operations

// SaveCommand saves a command to the database.
func (s *PostgresStore) SaveCommand(ctx context.Context, cmd *types.Command) error {
	return s.DB.WithContext(ctx).Create(cmd).Error
}

// GetCommand retrieves a command by ID.
func (s *PostgresStore) GetCommand(ctx context.Context, id string) (*types.Command, error) {
	var cmd types.Command
	if err := s.DB.WithContext(ctx).First(&cmd, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

// UpdateCommandStatus updates command status.
func (s *PostgresStore) UpdateCommandStatus(ctx context.Context, id string, status types.CommandStatus) error {
	return s.DB.WithContext(ctx).Model(&types.Command{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// SaveCommandResult saves a command result.
func (s *PostgresStore) SaveCommandResult(ctx context.Context, result *types.CommandResult) error {
	return s.DB.WithContext(ctx).Create(result).Error
}

// GetCommandResult retrieves a command result.
func (s *PostgresStore) GetCommandResult(ctx context.Context, commandID string) (*types.CommandResult, error) {
	var result types.CommandResult
	if err := s.DB.WithContext(ctx).First(&result, "command_id = ?", commandID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// Cluster operations

// SaveCluster saves a cluster to the database.
func (s *PostgresStore) SaveCluster(ctx context.Context, cluster *types.Cluster) error {
	return s.DB.WithContext(ctx).Save(cluster).Error
}

// GetCluster retrieves a cluster by ID.
func (s *PostgresStore) GetCluster(ctx context.Context, id string) (*types.Cluster, error) {
	var cluster types.Cluster
	if err := s.DB.WithContext(ctx).First(&cluster, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

// ListClusters lists all clusters.
func (s *PostgresStore) ListClusters(ctx context.Context) ([]*types.Cluster, error) {
	var clusters []*types.Cluster
	if err := s.DB.WithContext(ctx).Order("created_at DESC").Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

// UpdateClusterHealth updates cluster health.
func (s *PostgresStore) UpdateClusterHealth(ctx context.Context, id string, health types.ClusterHealth) error {
	return s.DB.WithContext(ctx).Model(&types.Cluster{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"health":     health,
			"updated_at": time.Now(),
		}).Error
}

// DeleteCluster deletes a cluster.
func (s *PostgresStore) DeleteCluster(ctx context.Context, id string) error {
	return s.DB.WithContext(ctx).Delete(&types.Cluster{}, "id = ?", id).Error
}
