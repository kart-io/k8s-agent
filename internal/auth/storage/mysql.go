package storage

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/logger/core"
)

// MySQLDB wraps GORM database connection
type MySQLDB struct {
	DB     *gorm.DB
	Logger core.Logger
}

// NewMySQLDB creates a new MySQL connection using GORM
func NewMySQLDB(cfg *commonoptions.MySQLOptions, log core.Logger) (*MySQLDB, error) {
	// Build connection string (DSN) for MySQL
	// Format: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Create GORM logger adapter
	gormLogger := NewGormLogger(log)

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
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

	return &MySQLDB{
		DB:     db,
		Logger: log,
	}, nil
}

// Close closes the database connection.
func (p *MySQLDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GormLogger adapts kart-io/logger to GORM's logger interface.
type GormLogger struct {
	log                       core.Logger
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormLogger creates a new GORM logger adapter.
func NewGormLogger(log core.Logger) *GormLogger {
	return &GormLogger{
		log:                       log,
		slowThreshold:             200 * time.Millisecond, // Queries slower than 200ms are logged as warnings
		ignoreRecordNotFoundError: true,                   // Don't log "record not found" as errors
	}
}

// LogMode implements gorm.io/gorm/logger.Interface.
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Return a new instance with the specified log level
	// Note: kart-io/logger level is controlled globally, so we just return self
	return l
}

// Info implements gorm.io/gorm/logger.Interface.
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Infow(fmt.Sprintf(msg, data...))
}

// Warn implements gorm.io/gorm/logger.Interface.
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.Warnw(fmt.Sprintf(msg, data...))
}

// Error implements gorm.io/gorm/logger.Interface.
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.Errorw(fmt.Sprintf(msg, data...))
}

// Trace implements gorm.io/gorm/logger.Interface
// This is called for every SQL query and logs query execution details.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && (!l.ignoreRecordNotFoundError || err.Error() != "record not found") {
		// Log errors
		l.log.Errorw("GORM query error",
			"elapsed_ms", elapsed.Milliseconds(),
			"rows", rows,
			"sql", sql,
			"error", err.Error(),
		)
	} else if elapsed > l.slowThreshold {
		// Log slow queries as warnings
		l.log.Warnw("GORM slow query",
			"elapsed_ms", elapsed.Milliseconds(),
			"rows", rows,
			"sql", sql,
			"slow_query", true,
		)
	} else {
		// Log normal queries at debug level
		l.log.Debugw("GORM query executed",
			"elapsed_ms", elapsed.Milliseconds(),
			"rows", rows,
			"sql", sql,
		)
	}
}
