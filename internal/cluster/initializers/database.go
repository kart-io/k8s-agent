// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/cluster/app/options"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer wraps the generic database initializer with cluster-specific configuration.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	opts   *options.ServerOptions
	logger core.Logger
	store  *storage.MySQLStorage // Cached storage instance
}

// NewDatabaseInitializer creates a database initializer for cluster service.
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	// Create the base initializer
	dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)

	// Configure auto-migration if enabled
	// Note: Cluster models should be defined in internal/cluster/models package
	// When models are created, uncomment the following:
	// if opts.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.Cluster{}, ...)
	// }

	return &DatabaseInitializer{
		DatabaseInitializer: dbInit,
		opts:                opts,
		logger:              logger,
	}
}

// GetStorage returns the initialized storage instance (for backward compatibility).
// It lazily creates the storage wrapper on first call.
func (d *DatabaseInitializer) GetStorage() *storage.MySQLStorage {
	if d.store != nil {
		return d.store
	}

	// Create storage using the configuration (this will create its own connection)
	// TODO: This is not ideal as it creates a new connection instead of reusing
	// the one from DatabaseInitializer. A better approach would be to refactor
	// MySQLStorage to accept an existing connection.
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
	return d.GetStorage()
}
