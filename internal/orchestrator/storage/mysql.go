package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/types"
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

// MySQLStore implements MySQL storage
type MySQLStore struct {
	db          *gorm.DB
	logger      core.Logger
	mysqlClient *commondb.MySQLClient
}

// NewMySQLStore creates a new MySQL store using common/db
func NewMySQLStore(opts *options.MySQLOptions, log core.Logger) (*MySQLStore, error) {
	// 直接使用 db 包创建 MySQL 客户端
	mysqlClient, err := commondb.NewMySQL(log,
		commondb.WithHost(opts.Host),
		commondb.WithPort(opts.Port),
		commondb.WithUser(opts.User),
		commondb.WithPassword(opts.Password),
		commondb.WithDatabase(opts.Database),
		commondb.WithMaxOpenConns(opts.MaxOpenConns),
		commondb.WithMaxIdleConns(opts.MaxIdleConns),
		commondb.WithConnMaxLifetime(opts.ConnMaxLifetime),
		commondb.WithLogLevel(opts.LogLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL client: %w", err)
	}

	store := &MySQLStore{
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

func (s *MySQLStore) migrate() error {
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

// Workflow operations.
func (s *MySQLStore) SaveWorkflow(ctx context.Context, workflow *types.Workflow) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return s.db.WithContext(ctx).Save(workflow).Error
}

func (s *MySQLStore) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var workflow types.Workflow
	if err := s.db.WithContext(ctx).First(&workflow, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get workflow %s: %w", id, err)
	}
	return &workflow, nil
}

func (s *MySQLStore) ListWorkflows(ctx context.Context) ([]*types.Workflow, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var workflows []*types.Workflow
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&workflows).Error; err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	return workflows, nil
}

// WorkflowExecution operations.
func (s *MySQLStore) SaveWorkflowExecution(ctx context.Context, execution *types.WorkflowExecution) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return s.db.WithContext(ctx).Save(execution).Error
}

func (s *MySQLStore) GetWorkflowExecution(ctx context.Context, id string) (*types.WorkflowExecution, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var execution types.WorkflowExecution
	if err := s.db.WithContext(ctx).First(&execution, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get workflow execution %s: %w", id, err)
	}
	return &execution, nil
}

func (s *MySQLStore) UpdateWorkflowExecutionStatus(ctx context.Context, id string, status types.ExecutionStatus) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	return s.db.WithContext(ctx).Model(&types.WorkflowExecution{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Strategy operations.
func (s *MySQLStore) SaveStrategy(ctx context.Context, strategy *types.Strategy) error {
	ctx, cancel := withTimeout(ctx)
	defer cancel()
	return s.db.WithContext(ctx).Save(strategy).Error
}

func (s *MySQLStore) GetStrategy(ctx context.Context, id string) (*types.Strategy, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var strategy types.Strategy
	if err := s.db.WithContext(ctx).First(&strategy, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get strategy %s: %w", id, err)
	}
	return &strategy, nil
}

func (s *MySQLStore) ListStrategies(ctx context.Context, enabledOnly bool) ([]*types.Strategy, error) {
	ctx, cancel := withTimeout(ctx)
	defer cancel()

	var strategies []*types.Strategy
	query := s.db.WithContext(ctx)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	if err := query.Order("priority DESC").Find(&strategies).Error; err != nil {
		return nil, fmt.Errorf("failed to list strategies: %w", err)
	}
	return strategies, nil
}

func (s *MySQLStore) Close() error {
	if s.mysqlClient != nil {
		return s.mysqlClient.Close()
	}
	return nil
}

func (s *MySQLStore) Health(ctx context.Context) error {
	if s.mysqlClient != nil {
		return s.mysqlClient.Health(ctx)
	}
	return fmt.Errorf("mysql client not initialized")
}
