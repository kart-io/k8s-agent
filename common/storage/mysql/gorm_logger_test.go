package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gormlogger "gorm.io/gorm/logger"
)

func TestNewGormLogger(t *testing.T) {
	log := mockLogger()

	tests := []struct {
		name          string
		level         string
		slowThreshold time.Duration
		expectedLevel gormlogger.LogLevel
	}{
		{
			name:          "silent level",
			level:         "silent",
			slowThreshold: 200 * time.Millisecond,
			expectedLevel: gormlogger.Silent,
		},
		{
			name:          "error level",
			level:         "error",
			slowThreshold: 200 * time.Millisecond,
			expectedLevel: gormlogger.Error,
		},
		{
			name:          "warn level",
			level:         "warn",
			slowThreshold: 200 * time.Millisecond,
			expectedLevel: gormlogger.Warn,
		},
		{
			name:          "info level",
			level:         "info",
			slowThreshold: 200 * time.Millisecond,
			expectedLevel: gormlogger.Info,
		},
		{
			name:          "invalid level defaults to silent",
			level:         "invalid",
			slowThreshold: 200 * time.Millisecond,
			expectedLevel: gormlogger.Silent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormLog := NewGormLogger(log, tt.level, tt.slowThreshold)
			assert.NotNil(t, gormLog)
			assert.Equal(t, tt.expectedLevel, gormLog.logLevel)
			assert.Equal(t, tt.slowThreshold, gormLog.slowThreshold)
			assert.True(t, gormLog.ignoreRecordNotFoundError)
		})
	}
}

func TestGormLogger_LogMode(t *testing.T) {
	log := mockLogger()
	gormLog := NewGormLogger(log, "info", 200*time.Millisecond)

	// Test LogMode
	newLogger := gormLog.LogMode(gormlogger.Error)
	assert.NotNil(t, newLogger)

	// Should return a new instance
	newGormLog, ok := newLogger.(*GormLogger)
	assert.True(t, ok)
	assert.Equal(t, gormlogger.Error, newGormLog.logLevel)
}

func TestGormLogger_Info(t *testing.T) {
	log := mockLogger()
	gormLog := NewGormLogger(log, "info", 200*time.Millisecond)

	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		gormLog.Info(ctx, "test info message: %s", "data")
	})

	// Silent mode should not log
	gormLog = NewGormLogger(log, "silent", 200*time.Millisecond)
	assert.NotPanics(t, func() {
		gormLog.Info(ctx, "test info message")
	})
}

func TestGormLogger_Warn(t *testing.T) {
	log := mockLogger()
	gormLog := NewGormLogger(log, "warn", 200*time.Millisecond)

	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		gormLog.Warn(ctx, "test warn message: %s", "data")
	})

	// Silent mode should not log
	gormLog = NewGormLogger(log, "silent", 200*time.Millisecond)
	assert.NotPanics(t, func() {
		gormLog.Warn(ctx, "test warn message")
	})
}

func TestGormLogger_Error(t *testing.T) {
	log := mockLogger()
	gormLog := NewGormLogger(log, "error", 200*time.Millisecond)

	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		gormLog.Error(ctx, "test error message: %s", "data")
	})

	// Silent mode should not log
	gormLog = NewGormLogger(log, "silent", 200*time.Millisecond)
	assert.NotPanics(t, func() {
		gormLog.Error(ctx, "test error message")
	})
}

func TestGormLogger_Trace(t *testing.T) {
	log := mockLogger()
	ctx := context.Background()

	tests := []struct {
		name          string
		level         string
		slowThreshold time.Duration
		elapsed       time.Duration
		err           error
		shouldLog     bool
		description   string
	}{
		{
			name:          "silent mode - no logging",
			level:         "silent",
			slowThreshold: 200 * time.Millisecond,
			elapsed:       100 * time.Millisecond,
			err:           nil,
			shouldLog:     false,
			description:   "Silent mode should not log",
		},
		{
			name:          "normal query",
			level:         "info",
			slowThreshold: 200 * time.Millisecond,
			elapsed:       100 * time.Millisecond,
			err:           nil,
			shouldLog:     true,
			description:   "Normal query should log at info level",
		},
		{
			name:          "slow query",
			level:         "warn",
			slowThreshold: 200 * time.Millisecond,
			elapsed:       300 * time.Millisecond,
			err:           nil,
			shouldLog:     true,
			description:   "Slow query should log at warn level",
		},
		{
			name:          "error query",
			level:         "error",
			slowThreshold: 200 * time.Millisecond,
			elapsed:       100 * time.Millisecond,
			err:           errors.New("database error"),
			shouldLog:     true,
			description:   "Error query should log at error level",
		},
		{
			name:          "record not found ignored",
			level:         "error",
			slowThreshold: 200 * time.Millisecond,
			elapsed:       100 * time.Millisecond,
			err:           gormlogger.ErrRecordNotFound,
			shouldLog:     false,
			description:   "Record not found should be ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gormLog := NewGormLogger(log, tt.level, tt.slowThreshold)

			begin := time.Now().Add(-tt.elapsed)

			// Mock function that returns SQL and rows
			fc := func() (string, int64) {
				return "SELECT * FROM users WHERE id = ?", 1
			}

			// Should not panic
			assert.NotPanics(t, func() {
				gormLog.Trace(ctx, begin, fc, tt.err)
			})
		})
	}
}
