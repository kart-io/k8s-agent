package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/logger/core"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger adapts kart-io/logger to GORM's logger interface
type GormLogger struct {
	log                       core.Logger
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// NewGormLogger creates a new GORM logger adapter
func NewGormLogger(log core.Logger) *GormLogger {
	return &GormLogger{
		log:                       log,
		slowThreshold:             200 * time.Millisecond, // Queries slower than 200ms are logged as warnings
		ignoreRecordNotFoundError: true,                   // Don't log "record not found" as errors
	}
}

// LogMode implements gorm.io/gorm/logger.Interface
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Return a new instance with the specified log level
	// Note: kart-io/logger level is controlled globally, so we just return self
	return l
}

// Info implements gorm.io/gorm/logger.Interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log.Infow(fmt.Sprintf(msg, data...))
}

// Warn implements gorm.io/gorm/logger.Interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log.Warnw(fmt.Sprintf(msg, data...))
}

// Error implements gorm.io/gorm/logger.Interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log.Errorw(fmt.Sprintf(msg, data...))
}

// Trace implements gorm.io/gorm/logger.Interface
// This is called for every SQL query and logs query execution details
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
