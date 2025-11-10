package mysql

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// Client wraps GORM DB with lifecycle management.
// It provides MySQL connection pooling, health checks, auto-migration,
// and graceful shutdown capabilities.
type Client struct {
	db     *gorm.DB
	logger core.Logger
	config options.MySQLOptions
}

// Option is a function that configures the MySQL client.
type Option func(*Client) error

// WithAutoMigrate adds models for automatic migration.
// This option should only be used in development/testing environments.
func WithAutoMigrate(models ...interface{}) Option {
	return func(c *Client) error {
		if len(models) == 0 {
			return nil
		}
		c.logger.Info("Running GORM AutoMigrate for models")
		if err := c.db.AutoMigrate(models...); err != nil {
			return fmt.Errorf("failed to auto migrate: %w", err)
		}
		c.logger.Info("GORM AutoMigrate completed successfully")
		return nil
	}
}

// NewClient creates a new MySQL client with the given configuration.
// It establishes a connection, configures the connection pool, and verifies connectivity.
//
// Example usage:
//
//	mysqlOpts := options.NewMySQLOptions()
//	mysqlOpts.Host = "localhost"
//	mysqlOpts.Database = "mydb"
//	client, err := mysql.NewClient(mysqlOpts, logger, mysql.WithAutoMigrate(&User{}, &Post{}))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(mysqlOpts *options.MySQLOptions, log core.Logger, opts ...Option) (*Client, error) {
	// Validate configuration
	if err := mysqlOpts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create GORM logger adapter
	gormLogger := NewGormLogger(log, mysqlOpts.LogLevel, mysqlOpts.SlowQueryThreshold)

	// Build DSN and connect
	dsn := mysqlOpts.DSN()
	log.Infow("Connecting to MySQL",
		"host", mysqlOpts.Host,
		"port", mysqlOpts.Port,
		"database", mysqlOpts.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// Get underlying *sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(mysqlOpts.ConnMaxLifetime)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), mysqlOpts.SlowQueryThreshold*2)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	client := &Client{
		db:     db,
		logger: log.With("component", "mysql"),
		config: *mysqlOpts,
	}

	// Apply options (e.g., AutoMigrate)
	for _, opt := range opts {
		if err := opt(client); err != nil {
			// Close connection on error
			_ = client.Close()
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	client.logger.Infow("MySQL connected successfully",
		"host", mysqlOpts.Host,
		"port", mysqlOpts.Port,
		"database", mysqlOpts.Database,
	)

	return client, nil
}

// DB returns the underlying GORM DB instance for direct access.
func (c *Client) DB() *gorm.DB {
	return c.db
}

// Health checks if the database connection is healthy.
// It pings the database with the provided context.
func (c *Client) Health(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Close closes the database connection.
// It should be called when shutting down the application.
func (c *Client) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	c.logger.Info("Closing MySQL connection")
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// AutoMigrate runs automatic migration for the given models.
// This is a convenience method equivalent to using WithAutoMigrate option.
func (c *Client) AutoMigrate(models ...interface{}) error {
	if len(models) == 0 {
		return nil
	}

	c.logger.Info("Running GORM AutoMigrate")
	if err := c.db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	c.logger.Info("GORM AutoMigrate completed successfully")
	return nil
}
