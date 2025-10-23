package storage

import (
	"database/sql"
	"fmt"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

type MySQLStorage struct {
	db          *sql.DB
	log         core.Logger
	mysqlClient *commondb.MySQLClient
}

// NewMySQLStorage creates a new MySQL storage using common/db
func NewMySQLStorage(opts *options.DatabaseOptions, logger core.Logger) (*MySQLStorage, error) {
	// 使用 common/db helper 函数创建 MySQL 客户端
	mysqlClient, err := commondb.NewMySQLFromOptions(logger, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create MySQL client: %w", err)
	}

	// 获取 *sql.DB
	sqlDB, err := mysqlClient.DB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if logger != nil {
		logger.Info("Successfully connected to MySQL")
	}

	storage := &MySQLStorage{
		db:          sqlDB,
		log:         logger,
		mysqlClient: mysqlClient,
	}

	// 初始化数据库表结构
	if err := storage.InitSchema(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MySQLStorage) DB() *sql.DB {
	return s.db
}

func (s *MySQLStorage) Close() error {
	if s.mysqlClient != nil {
		return s.mysqlClient.Close()
	}
	return nil
}

// InitSchema initializes database schema
func (s *MySQLStorage) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS clusters (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		endpoint VARCHAR(255) NOT NULL,
		version VARCHAR(50),
		status VARCHAR(50) NOT NULL DEFAULT 'unknown',
		region VARCHAR(100),
		provider VARCHAR(100),
		kubeconfig TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_clusters_status (status),
		INDEX idx_clusters_provider (provider)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	if s.log != nil {
		s.log.Info("Database schema initialized")
	}
	return nil
}

// NewMySQLStorageWithDB creates a MySQLStorage instance with an existing DB connection
// This is useful for testing with mocked databases
func NewMySQLStorageWithDB(db *sql.DB, logger core.Logger) *MySQLStorage {
	return &MySQLStorage{
		db:  db,
		log: logger,
	}
}
