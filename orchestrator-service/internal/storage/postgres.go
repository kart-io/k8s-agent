package storage

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/orchestrator-service/pkg/types"
)

// PostgresStore implements PostgreSQL storage
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	store := &PostgresStore{
		db:     db,
		logger: log.With(zap.String("component", "postgres")),
	}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	store.logger.Info("PostgreSQL store initialized")
	return store, nil
}

func (s *PostgresStore) migrate() error {
	return s.db.AutoMigrate(
		&types.Workflow{},
		&types.WorkflowExecution{},
		&types.Strategy{},
		&types.Task{},
		&types.RemediationAction{},
		&types.RemediationExecution{},
		&types.AIAnalysisRequest{},
	)
}

// Workflow operations
func (s *PostgresStore) SaveWorkflow(ctx context.Context, workflow *types.Workflow) error {
	return s.db.WithContext(ctx).Save(workflow).Error
}

func (s *PostgresStore) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	var workflow types.Workflow
	if err := s.db.WithContext(ctx).First(&workflow, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (s *PostgresStore) ListWorkflows(ctx context.Context) ([]*types.Workflow, error) {
	var workflows []*types.Workflow
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, err
	}
	return workflows, nil
}

// WorkflowExecution operations
func (s *PostgresStore) SaveWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	return s.db.WithContext(ctx).Save(execution).Error
}

func (s *PostgresStore) GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error) {
	var execution types.WorkflowExecution
	if err := s.db.WithContext(ctx).First(&execution, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &execution, nil
}

func (s *PostgresStore) UpdateWorkflowExecutionStatus(ctx context.Context, id string, status types.ExecutionStatus) error {
	return s.db.WithContext(ctx).Model(&types.WorkflowExecution{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Strategy operations
func (s *PostgresStore) SaveStrategy(ctx context.Context, strategy *types.Strategy) error {
	return s.db.WithContext(ctx).Save(strategy).Error
}

func (s *PostgresStore) GetStrategy(ctx context.Context, id string) (*types.Strategy, error) {
	var strategy types.Strategy
	if err := s.db.WithContext(ctx).First(&strategy, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (s *PostgresStore) ListStrategies(ctx context.Context, enabledOnly bool) ([]*types.Strategy, error) {
	var strategies []*types.Strategy
	query := s.db.WithContext(ctx)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	if err := query.Order("priority DESC").Find(&strategies).Error; err != nil {
		return nil, err
	}
	return strategies, nil
}

func (s *PostgresStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *PostgresStore) Health(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}