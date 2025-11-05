// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer initializes the database connection and schema.
type DatabaseInitializer struct {
	opts    *options.ServerOptions
	logger  core.Logger
	storage *storage.MySQLStorage
}

// NewDatabaseInitializer creates a new database initializer.
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Initialize initializes the database connection.
func (i *DatabaseInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing database connection",
		"host", i.opts.Database.Host,
		"port", i.opts.Database.Port,
		"database", i.opts.Database.Database,
	)

	// 创建数据库连接
	store, err := storage.NewMySQLStorage(i.opts.Database, i.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	i.storage = store

	i.logger.Infow("Database connected successfully",
		"host", i.opts.Database.Host,
		"port", i.opts.Database.Port,
		"database", i.opts.Database.Database,
	)

	// 初始化数据库 schema
	if err := i.storage.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}
	i.logger.Info("Database schema initialized successfully")

	return nil
}

// Shutdown closes the database connection.
func (i *DatabaseInitializer) Shutdown(ctx context.Context) error {
	i.logger.Info("Closing database connection")
	if i.storage != nil {
		// MySQL storage doesn't have explicit close method
		// Connection will be closed when the process exits
	}
	return nil
}

// Priority returns the initialization priority (higher = earlier).
func (i *DatabaseInitializer) Priority() int {
	return 300 // Database should be initialized early
}

// Name returns the name of this initializer.
func (i *DatabaseInitializer) Name() string {
	return "Database"
}

// GetStorage returns the initialized storage instance.
// Deprecated: Use Store() instead for consistency across services.
func (i *DatabaseInitializer) GetStorage() *storage.MySQLStorage {
	return i.storage
}

// Store returns the initialized storage instance.
// This is the standard method name across all database initializers.
func (i *DatabaseInitializer) Store() interface{} {
	return i.storage
}
