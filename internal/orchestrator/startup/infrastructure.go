// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// InfrastructureInitializers contains all infrastructure-level initializers.
type InfrastructureInitializers struct {
	Database *DatabaseInitializer
	Redis    *RedisInitializer
	NATS     *pkginitializers.NATSInitializer // ✅ Use pkg/initializers version
}

// NewInfrastructureInitializers creates all infrastructure initializers.
func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
	// Create database initializer
	dbInit := &DatabaseInitializer{
		DatabaseInitializer: pkginitializers.NewDatabaseInitializer(opts.Database, logger),
		opts:                opts,
		logger:              logger,
	}
	// Note: Orchestrator models should be added when available
	// if opts.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.Workflow{}, &models.Strategy{}, ...)
	// }

	// Create Redis initializer
	redisInit := &RedisInitializer{
		RedisInitializer: pkginitializers.NewRedisInitializer(opts.Redis, logger),
		opts:             opts,
		logger:           logger,
	}

	// Create NATS initializer (using pkg/initializers)
	natsInit := pkginitializers.NewNATSInitializer(opts.NATS, logger)

	return &InfrastructureInitializers{
		Database: dbInit,
		Redis:    redisInit,
		NATS:     natsInit,
	}
}

// DatabaseInitializer wraps the generic database initializer with service-specific configuration.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	opts   *commonapp.StandardOptions
	logger core.Logger
	store  *storage.MySQLStore
}

// Store returns the MySQL storage instance (creates on first call).
func (d *DatabaseInitializer) Store() *storage.MySQLStore {
	if d.store == nil {
		var err error
		d.store, err = storage.NewMySQLStore(d.opts.Database, d.logger)
		if err != nil {
			d.logger.Errorw("Failed to create MySQL store", "error", err)
			return nil
		}
	}
	return d.store
}

// RedisInitializer wraps the generic Redis initializer with service-specific configuration.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	opts   *commonapp.StandardOptions
	logger core.Logger
	store  *storage.RedisStore
}

// Store returns the Redis storage instance (creates on first call).
func (r *RedisInitializer) Store() *storage.RedisStore {
	if r.store == nil {
		var err error
		r.store, err = storage.NewRedisStore(r.opts.Redis, r.logger)
		if err != nil {
			r.logger.Errorw("Failed to create Redis store", "error", err)
			return nil
		}
	}
	return r.store
}

// NATSInitializer has been replaced with pkg/initializers.NATSInitializer
// The custom implementation has been removed to eliminate code duplication.
// All NATS functionality is now provided by pkg/initializers/nats.go.
//
// Migration notes:
// - pkg/initializers.NATSInitializer provides all the same methods
// - Connection is available via Conn() or Connection() methods
// - Health checks and lifecycle management are handled automatically
// - Reconnect handlers are configured in pkg/initializers
