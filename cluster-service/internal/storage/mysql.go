package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string // Kept for compatibility but not used in MySQL
	MaxOpenConns int
	MaxIdleConns int
}

type MySQLStorage struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewMySQLStorage(cfg *Config, logger *logrus.Logger) (*MySQLStorage, error) {
	// MySQL DSN format: user:password@tcp(host:port)/dbname?parseTime=true&charset=utf8mb4
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	db.SetConnMaxLifetime(time.Hour)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to MySQL")

	return &MySQLStorage{
		db:  db,
		log: logger,
	}, nil
}

func (s *MySQLStorage) DB() *sql.DB {
	return s.db
}

func (s *MySQLStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
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

	s.log.Info("Database schema initialized")
	return nil
}

// NewMySQLStorageWithDB creates a MySQLStorage instance with an existing DB connection
// This is useful for testing with mocked databases
func NewMySQLStorageWithDB(db *sql.DB, logger *logrus.Logger) *MySQLStorage {
	if logger == nil {
		logger = logrus.New()
	}
	return &MySQLStorage{
		db:  db,
		log: logger,
	}
}
