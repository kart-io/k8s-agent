// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer wraps the generic database initializer with monitor-specific configuration.
//
// DESIGN NOTE: This is a thin wrapper that:
// 1. Embeds pkg/initializers.DatabaseInitializer for standard lifecycle management
// 2. Provides Storage() method for backward compatibility with existing monitor code
// 3. Can be extended with monitor-specific models via WithAutoMigrate() when needed
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	opts   *commonapp.StandardOptions
	logger core.Logger
	store  *storage.MySQLStorage // Lazily initialized storage wrapper
}

// NewDatabaseInitializer creates a database initializer for monitor service.
func NewDatabaseInitializer(cfg *commonapp.StandardOptions, logger core.Logger) *DatabaseInitializer {
	// Create the base initializer with standard database configuration
	dbInit := pkginitializers.NewDatabaseInitializer(cfg.Database, logger)

	// TODO: Configure auto-migration when monitor models are defined
	// if cfg.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.Metric{}, &models.Alert{}, ...)
	// }

	return &DatabaseInitializer{
		DatabaseInitializer: dbInit,
		opts:                cfg,
		logger:              logger,
	}
}

// Storage returns the initialized storage (for backward compatibility).
// It lazily creates the storage wrapper on first call.
func (d *DatabaseInitializer) Storage() *storage.MySQLStorage {
	if d.store != nil {
		return d.store
	}

	// Create storage using the configuration (this will create its own connection)
	store, err := storage.NewMySQLStorage(d.opts.Database, d.logger)
	if err != nil {
		d.logger.Errorw("Failed to create MySQL storage", "error", err)
		return nil
	}

	d.store = store
	return d.store
}

// Store returns the initialized storage instance.
// This is the standard method name across all database initializers.
func (d *DatabaseInitializer) Store() interface{} {
	return d.Storage()
}
