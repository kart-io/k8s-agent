package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"

	commondb "github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

type PostgresStorage struct {
	db          *sql.DB
	log         *logrus.Logger
	mysqlClient *commondb.MySQLClient
}

// NewPostgresStorage creates a new storage using common/db
func NewPostgresStorage(opts *options.DatabaseOptions, logger core.Logger) (*PostgresStorage, error) {
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

	// 将 core.Logger 转换为 logrus.Logger (临时兼容)
	logrusLogger, ok := logger.(core.Logger)
	if !ok {
		// 如果不是 logrus.Logger，创建一个默认的
		logrusLogger = logrus.New()
	}

	storage := &PostgresStorage{
		db:          sqlDB,
		log:         logrusLogger,
		mysqlClient: mysqlClient,
	}

	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logrusLogger.Info("MySQL storage initialized successfully")
	return storage, nil
}

func (s *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS metrics_summary (
		id INT AUTO_INCREMENT PRIMARY KEY,
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS agent_metrics (
		id INT AUTO_INCREMENT PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		agent_name VARCHAR(255),
		cluster_id VARCHAR(255),
		status VARCHAR(50),
		cpu_usage FLOAT,
		memory_usage FLOAT,
		event_count INT,
		command_count INT,
		last_heartbeat TIMESTAMP NULL,
		uptime BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_agent_id (agent_id),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS event_metrics (
		id INT AUTO_INCREMENT PRIMARY KEY,
		event_type VARCHAR(100),
		severity VARCHAR(50),
		count INT,
		last_occurrence TIMESTAMP NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_event_type (event_type),
		INDEX idx_severity (severity)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS alerts (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		enabled BOOLEAN DEFAULT true,
		rule_type VARCHAR(50),
		conditions JSON,
		channels JSON,
		severity VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS alert_history (
		id VARCHAR(255) PRIMARY KEY,
		alert_id VARCHAR(255),
		alert_name VARCHAR(255),
		status VARCHAR(50),
		message TEXT,
		triggered_at TIMESTAMP NULL,
		resolved_at TIMESTAMP NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_alert_id (alert_id),
		INDEX idx_status (status),
		INDEX idx_triggered_at (triggered_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS trend_data (
		id INT AUTO_INCREMENT PRIMARY KEY,
		timestamp TIMESTAMP NOT NULL,
		metrics JSON,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_timestamp (timestamp)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
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
