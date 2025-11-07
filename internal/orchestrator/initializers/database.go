// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer wraps the generic database initializer with orchestrator-specific configuration.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	opts   *commonapp.StandardOptions
	logger core.Logger
	store  *storage.MySQLStore // Cached storage instance
}

// NewDatabaseInitializer creates a database initializer for orchestrator service.
func NewDatabaseInitializer(opts *commonapp.StandardOptions, logger core.Logger) *DatabaseInitializer {
	// Create the base initializer
	dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)

	// Configure auto-migration if enabled
	// Note: Orchestrator models should be defined in internal/orchestrator/models package
	// When models are created, uncomment the following:
	// if opts.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.Workflow{}, &models.Strategy{}, ...)
	// }

	return &DatabaseInitializer{
		DatabaseInitializer: dbInit,
		opts:                opts,
		logger:              logger,
	}
}

// Store returns the initialized storage instance.
// It lazily creates the storage wrapper on first call.
func (d *DatabaseInitializer) Store() *storage.MySQLStore {
	if d.store != nil {
		return d.store
	}

	// Create storage using the configuration (this will create its own connection)
	store, err := storage.NewMySQLStore(d.opts.Database, d.logger)
	if err != nil {
		d.logger.Errorw("Failed to create MySQL store", "error", err)
		return nil
	}

	d.store = store
	return d.store
}

