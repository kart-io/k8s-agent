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

func NewPostgresStorage(config *Config, logger *logrus.Logger) (*PostgresStorage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{
		db:  db,
		log: logger,
	}

	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info("PostgreSQL storage initialized successfully")
	return storage, nil
}

func (s *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics_summary (
		id SERIAL PRIMARY KEY,
		total_agents INT,
		online_agents INT,
		offline_agents INT,
		total_events INT,
		critical_events INT,
		total_commands INT,
		running_commands INT,
		cpu_usage FLOAT,
		memory_usage FLOAT,
		disk_usage FLOAT,
		network_in FLOAT,
		network_out FLOAT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS agent_metrics (
		id SERIAL PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		agent_name VARCHAR(255),
		cluster_id VARCHAR(255),
		status VARCHAR(50),
		cpu_usage FLOAT,
		memory_usage FLOAT,
		event_count INT,
		command_count INT,
		last_heartbeat TIMESTAMP,
		uptime BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_agent_id (agent_id),
		INDEX idx_created_at (created_at)
	);

	CREATE TABLE IF NOT EXISTS event_metrics (
		id SERIAL PRIMARY KEY,
		event_type VARCHAR(100),
		severity VARCHAR(50),
		count INT,
		last_occurrence TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_event_type (event_type),
		INDEX idx_severity (severity)
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		enabled BOOLEAN DEFAULT true,
		rule_type VARCHAR(50),
		conditions JSONB,
		channels JSONB,
		severity VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alert_history (
		id VARCHAR(255) PRIMARY KEY,
		alert_id VARCHAR(255),
		alert_name VARCHAR(255),
		status VARCHAR(50),
		message TEXT,
		triggered_at TIMESTAMP,
		resolved_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_alert_id (alert_id),
		INDEX idx_status (status),
		INDEX idx_triggered_at (triggered_at)
	);

	CREATE TABLE IF NOT EXISTS trend_data (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMP NOT NULL,
		metrics JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_timestamp (timestamp)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) DB() *sql.DB {
	return s.db
}
