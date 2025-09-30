package storage

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)

// PostgresStore implements storage using PostgreSQL
type PostgresStore struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(config types.DatabaseConfig, log *zap.Logger) (*PostgresStore, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.Default
	if config.Host != "" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:     db,
		logger: log.With(zap.String("component", "postgres")),
	}

	// Auto-migrate schemas
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	store.logger.Info("PostgreSQL store initialized",
		zap.String("host", config.Host),
		zap.String("database", config.Database))

	return store, nil
}

// migrate runs database migrations
func (s *PostgresStore) migrate() error {
	return s.db.AutoMigrate(
		&types.Agent{},
		&types.Event{},
		&types.Metrics{},
		&types.Command{},
		&types.CommandResult{},
		&types.Cluster{},
		&types.AlertRule{},
		&types.Alert{},
	)
}

// Agent operations

// SaveAgent saves an agent to the database
func (s *PostgresStore) SaveAgent(ctx context.Context, agent *types.Agent) error {
	return s.db.WithContext(ctx).Save(agent).Error
}

// GetAgent retrieves an agent by ID
func (s *PostgresStore) GetAgent(ctx context.Context, id string) (*types.Agent, error) {
	var agent types.Agent
	if err := s.db.WithContext(ctx).First(&agent, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetAgentByClusterID retrieves an agent by cluster ID
func (s *PostgresStore) GetAgentByClusterID(ctx context.Context, clusterID string) (*types.Agent, error) {
	var agent types.Agent
	if err := s.db.WithContext(ctx).First(&agent, "cluster_id = ?", clusterID).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

// ListAgents lists all agents
func (s *PostgresStore) ListAgents(ctx context.Context, status *types.AgentStatus) ([]*types.Agent, error) {
	var agents []*types.Agent
	query := s.db.WithContext(ctx)

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Order("registered_at DESC").Find(&agents).Error; err != nil {
		return nil, err
	}
	return agents, nil
}

// UpdateAgentStatus updates agent status
func (s *PostgresStore) UpdateAgentStatus(ctx context.Context, id string, status types.AgentStatus) error {
	return s.db.WithContext(ctx).Model(&types.Agent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// UpdateAgentHeartbeat updates agent heartbeat timestamp
func (s *PostgresStore) UpdateAgentHeartbeat(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Model(&types.Agent{}).
		Where("id = ?", id).
		Update("last_heartbeat", time.Now()).Error
}

// DeleteAgent deletes an agent
func (s *PostgresStore) DeleteAgent(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&types.Agent{}, "id = ?", id).Error
}

// Event operations

// SaveEvent saves an event to the database
func (s *PostgresStore) SaveEvent(ctx context.Context, event *types.Event) error {
	return s.db.WithContext(ctx).Create(event).Error
}

// GetEvent retrieves an event by ID
func (s *PostgresStore) GetEvent(ctx context.Context, id string) (*types.Event, error) {
	var event types.Event
	if err := s.db.WithContext(ctx).First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// ListEvents lists events with filters
func (s *PostgresStore) ListEvents(ctx context.Context, filter EventFilter) ([]*types.Event, error) {
	var events []*types.Event
	query := s.db.WithContext(ctx)

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

// EventFilter defines filters for event queries
type EventFilter struct {
	ClusterID string
	Severity  string
	Namespace string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
}

// Command operations

// SaveCommand saves a command to the database
func (s *PostgresStore) SaveCommand(ctx context.Context, cmd *types.Command) error {
	return s.db.WithContext(ctx).Create(cmd).Error
}

// GetCommand retrieves a command by ID
func (s *PostgresStore) GetCommand(ctx context.Context, id string) (*types.Command, error) {
	var cmd types.Command
	if err := s.db.WithContext(ctx).First(&cmd, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cmd, nil
}

// UpdateCommandStatus updates command status
func (s *PostgresStore) UpdateCommandStatus(ctx context.Context, id string, status types.CommandStatus) error {
	return s.db.WithContext(ctx).Model(&types.Command{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// SaveCommandResult saves a command result
func (s *PostgresStore) SaveCommandResult(ctx context.Context, result *types.CommandResult) error {
	return s.db.WithContext(ctx).Create(result).Error
}

// GetCommandResult retrieves a command result
func (s *PostgresStore) GetCommandResult(ctx context.Context, commandID string) (*types.CommandResult, error) {
	var result types.CommandResult
	if err := s.db.WithContext(ctx).First(&result, "command_id = ?", commandID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

// Cluster operations

// SaveCluster saves a cluster to the database
func (s *PostgresStore) SaveCluster(ctx context.Context, cluster *types.Cluster) error {
	return s.db.WithContext(ctx).Save(cluster).Error
}

// GetCluster retrieves a cluster by ID
func (s *PostgresStore) GetCluster(ctx context.Context, id string) (*types.Cluster, error) {
	var cluster types.Cluster
	if err := s.db.WithContext(ctx).First(&cluster, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cluster, nil
}

// ListClusters lists all clusters
func (s *PostgresStore) ListClusters(ctx context.Context) ([]*types.Cluster, error) {
	var clusters []*types.Cluster
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&clusters).Error; err != nil {
		return nil, err
	}
	return clusters, nil
}

// UpdateClusterHealth updates cluster health
func (s *PostgresStore) UpdateClusterHealth(ctx context.Context, id string, health types.ClusterHealth) error {
	return s.db.WithContext(ctx).Model(&types.Cluster{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"health":     health,
			"updated_at": time.Now(),
		}).Error
}

// DeleteCluster deletes a cluster
func (s *PostgresStore) DeleteCluster(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&types.Cluster{}, "id = ?", id).Error
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health checks database health
func (s *PostgresStore) Health(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}