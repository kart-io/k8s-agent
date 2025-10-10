package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type PostgresStorage struct {
	db  *sql.DB
	log *logrus.Logger
}

func NewPostgresStorage(cfg *Config, logger *logrus.Logger) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
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

	logger.Info("Successfully connected to PostgreSQL")

	return &PostgresStorage{
		db:  db,
		log: logger,
	}, nil
}

func (s *PostgresStorage) DB() *sql.DB {
	return s.db
}

func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// InitSchema initializes database schema
func (s *PostgresStorage) InitSchema() error {
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
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters(status);
	CREATE INDEX IF NOT EXISTS idx_clusters_provider ON clusters(provider);
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	s.log.Info("Database schema initialized")
	return nil
}
