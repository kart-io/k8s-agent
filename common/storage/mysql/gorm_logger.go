package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/kart-io/logger/core"
)

// GormLogger adapts kart-io/logger to GORM's logger interface.
// It provides structured logging for SQL queries with configurable levels,
// slow query detection, and error filtering.
type GormLogger struct {
	log                       core.Logger
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormLogger creates a new GORM logger adapter.
// Parameters:
//   - log: The kart-io logger instance
//   - level: Log level string (silent, error, warn, info)
//   - slowThreshold: Threshold for slow query logging (e.g., 200ms)
func NewGormLogger(log core.Logger, level string, slowThreshold time.Duration) *GormLogger {
	// Parse log level
	var gormLevel gormlogger.LogLevel
	switch level {
	case "silent":
		gormLevel = gormlogger.Silent
	case "error":
		gormLevel = gormlogger.Error
	case "warn":
		gormLevel = gormlogger.Warn
	case "info":
		gormLevel = gormlogger.Info
	default:
		gormLevel = gormlogger.Silent
	}

	return &GormLogger{
		log:                       log,
		logLevel:                  gormLevel,
		slowThreshold:             slowThreshold,
		ignoreRecordNotFoundError: true, // Don't log "record not found" as errors
	}
}

// LogMode implements gorm.io/gorm/logger.Interface.
// It sets the log level and returns a new logger instance.
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info implements gorm.io/gorm/logger.Interface.
// Logs informational messages.
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.log.Infow(fmt.Sprintf(msg, data...))
	}
}

// Warn implements gorm.io/gorm/logger.Interface.
// Logs warning messages.
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.log.Warnw(fmt.Sprintf(msg, data...))
	}
}

// Error implements gorm.io/gorm/logger.Interface.
// Logs error messages.
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.log.Errorw(fmt.Sprintf(msg, data...))
	}
}

// Trace implements gorm.io/gorm/logger.Interface.
// This is called for every SQL query and logs query execution details with structured fields.
// It categorizes queries as:
//   - Errors (with optional filtering of "record not found")
//   - Slow queries (exceeding slowThreshold)
//   - Normal queries (at debug level)
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []interface{}{
		"elapsed_ms", elapsed.Milliseconds(),
		"rows", rows,
		"sql", sql,
	}

	if err != nil && l.logLevel >= gormlogger.Error {
		// Check if we should ignore this error
		if l.ignoreRecordNotFoundError && errors.Is(err, gormlogger.ErrRecordNotFound) {
			return
		}

		// Log errors with error field
		l.log.Errorw("GORM query error",
			append(fields, "error", err.Error())...,
		)
	} else if elapsed > l.slowThreshold && l.logLevel >= gormlogger.Warn {
		// Log slow queries as warnings
		l.log.Warnw("GORM slow query",
			append(fields, "slow_query", true)...,
		)
	} else if l.logLevel >= gormlogger.Info {
		// Log normal queries at debug level
		l.log.Debugw("GORM query executed", fields...)
	}
}
