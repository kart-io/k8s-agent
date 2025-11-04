// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer initializes the database connection and schema.
type DatabaseInitializer struct {
	opts    *commonoptions.DatabaseOptions
	logger  core.Logger
	storage *storage.MySQLStorage
}

// NewDatabaseInitializer creates a new database initializer.
func NewDatabaseInitializer(opts *commonoptions.DatabaseOptions, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Initialize initializes the database connection.
func (i *DatabaseInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing database connection",
		"host", i.opts.Host,
		"port", i.opts.Port,
		"database", i.opts.Database,
	)

	// 创建数据库连接
	store, err := storage.NewMySQLStorage(i.opts, i.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	i.storage = store

	i.logger.Infow("Database connected successfully",
		"host", i.opts.Host,
		"port", i.opts.Port,
		"database", i.opts.Database,
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
func (i *DatabaseInitializer) GetStorage() *storage.MySQLStorage {
	return i.storage
}
