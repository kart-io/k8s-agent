package storage

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

type MySQLStorage struct {
	db          *sql.DB
	gormDB      *gorm.DB // GORM DB for ORM operations
	log         core.Logger
	mysqlClient *commondb.MySQLClient
	ownedClient bool // Whether this storage owns the client and should close it
}

// NewMySQLStorage creates a new MySQL storage using common/db.
// This method creates its own database connection and should only be used
// when a connection is not already available.
func NewMySQLStorage(opts *options.DatabaseOptions, logger core.Logger) (*MySQLStorage, error) {
	// 直接使用 db 包创建 MySQL 客户端
	mysqlClient, err := commondb.NewMySQL(logger,
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
		gormDB:      mysqlClient.DB,
		log:         logger,
		mysqlClient: mysqlClient,
		ownedClient: true, // We own this client
	}

	// 初始化数据库表结构
	if err := storage.InitSchema(); err != nil {
		return nil, err
	}

	return storage, nil
}

// NewMySQLStorageWithDB creates a new MySQL storage using an existing GORM DB connection.
// This is the preferred method when reusing an existing database connection.
func NewMySQLStorageWithDB(gormDB *gorm.DB, logger core.Logger) (*MySQLStorage, error) {
	if gormDB == nil {
		return nil, fmt.Errorf("gormDB cannot be nil")
	}

	// 获取 *sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if logger != nil {
		logger.Info("Reusing existing MySQL connection for cluster storage")
	}

	storage := &MySQLStorage{
		db:          sqlDB,
		gormDB:      gormDB,
		log:         logger,
		mysqlClient: nil,      // No client since we're reusing connection
		ownedClient: false,    // We don't own this connection
	}

	// 初始化数据库表结构
	if err := storage.InitSchema(); err != nil {
		return nil, err
	}

	return storage, nil
}

// NewMySQLStorageForTesting creates a MySQLStorage instance with an existing *sql.DB connection.
// This is useful for testing with mocked databases (sqlmock).
// Note: This bypasses GORM and schema initialization.
func NewMySQLStorageForTesting(db *sql.DB, logger core.Logger) *MySQLStorage {
	return &MySQLStorage{
		db:          db,
		gormDB:      nil, // No GORM in test mode
		log:         logger,
		mysqlClient: nil,
		ownedClient: false, // Test code owns the connection
	}
}

func (s *MySQLStorage) DB() *sql.DB {
	return s.db
}

// GormDB returns the GORM database instance for ORM operations.
func (s *MySQLStorage) GormDB() *gorm.DB {
	return s.gormDB
}

func (s *MySQLStorage) Close() error {
	// Only close if we own the client
	if s.ownedClient && s.mysqlClient != nil {
		return s.mysqlClient.Close()
	}
	// If we're reusing a connection, don't close it
	return nil
}

// InitSchema initializes database schema.
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

	_, err := s.db.ExecContext(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	if s.log != nil {
		s.log.Info("Database schema initialized")
	}
	return nil
}
