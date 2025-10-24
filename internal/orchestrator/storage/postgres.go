package storage

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
	"github.com/kart-io/logger/core"
)

// PostgresStore implements MySQL storage
// Note: Kept the name for backward compatibility, but now using MySQL
type PostgresStore struct {
	db          *gorm.DB
	logger      core.Logger
	mysqlClient *commondb.MySQLClient
}

// NewPostgresStore creates a new MySQL store using common/db
// Note: Kept the name for backward compatibility, but now using MySQL
func NewPostgresStore(opts *options.DatabaseOptions, log core.Logger) (*PostgresStore, error) {
	// 使用 common/db helper 函数创建 MySQL 客户端
	mysqlClient, err := commondb.NewMySQLFromOptions(log, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL client: %w", err)
	}

	store := &PostgresStore{
		db:          mysqlClient.DB,
		logger:      log,
		mysqlClient: mysqlClient,
	}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	store.logger.Info("MySQL store initialized successfully")
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
	if s.mysqlClient != nil {
		return s.mysqlClient.Close()
	}
	return nil
}

func (s *PostgresStore) Health(ctx context.Context) error {
	if s.mysqlClient != nil {
		return s.mysqlClient.Health(ctx)
	}
	return fmt.Errorf("mysql client not initialized")
}
