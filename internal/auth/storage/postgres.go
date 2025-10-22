package storage

import (
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/auth/config"
	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// PostgresDB wraps GORM database connection
// Note: Kept the name for backward compatibility, but now using MySQL
type PostgresDB struct {
	DB *gorm.DB
}

// NewPostgresDB creates a new MySQL connection using GORM
// Note: Kept the name for backward compatibility, but now using MySQL
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	// Build connection string (DSN) for MySQL
	// Format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	// Get kart-io/logger instance for GORM
	log := logger.GetLogger()
	gormLogger := logger.NewGormLogger(log)

	// Open GORM connection with custom logger
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying *sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour) // Connections last max 1 hour

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run AutoMigrate only if enabled (dev/test only, not in production)
	if cfg.AutoMigrate {
		log.Infow("GORM AutoMigrate enabled, running migrations...")

		// Register all models for auto-migration
		if err := db.AutoMigrate(
			&model.User{},
			&model.Role{},
			&model.Permission{},
			&model.APIKey{},
			&model.UserRole{},
			&model.RolePermission{},
		); err != nil {
			return nil, fmt.Errorf("failed to run auto migrations: %w", err)
		}

		log.Infow("GORM AutoMigrate completed successfully")
	} else {
		log.Infow("GORM AutoMigrate disabled (production mode)")
	}

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Health checks database health
func (p *PostgresDB) Health() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
