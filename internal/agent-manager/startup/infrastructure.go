// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// InfrastructureInitializers contains all infrastructure-level initializers.
type InfrastructureInitializers struct {
	Database *DatabaseInitializer
	Redis    *RedisInitializer
}

// NewInfrastructureInitializers creates all infrastructure initializers.
func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
	// Create database initializer with auto-migration
	dbInit := &DatabaseInitializer{
		DatabaseInitializer: pkginitializers.NewDatabaseInitializer(opts.Database, logger),
		logger:              logger,
	}
	if opts.Database.AutoMigrate {
		dbInit.WithAutoMigrate(
			&types.Agent{},
			&types.Event{},
			&types.Metrics{},
			&types.Command{},
			&types.CommandResult{},
			&types.Cluster{},
			&types.AlertRule{},
			&types.Alert{},
		)
	}

	// Create Redis initializer
	redisInit := &RedisInitializer{
		RedisInitializer: pkginitializers.NewRedisInitializer(opts.Redis, logger),
		logger:           logger,
	}

	return &InfrastructureInitializers{
		Database: dbInit,
		Redis:    redisInit,
	}
}

// DatabaseInitializer wraps the generic database initializer with service-specific storage.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	logger core.Logger
	store  *storage.MySQLStore
}

// Store returns the storage instance (creates on first call using common/storage/mysql.Client).
func (d *DatabaseInitializer) Store() *storage.MySQLStore {
	if d.store == nil && d.Client() != nil {
		// ✅ 使用 common/storage/mysql.Client 创建业务 storage
		d.store = storage.NewMySQLStore(d.Client(), d.logger)
	}
	return d.store
}

// RedisInitializer wraps the generic Redis initializer with service-specific storage.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	logger core.Logger
	store  *storage.RedisStore
}

// Store returns the storage instance (creates on first call using redis.Client).
func (r *RedisInitializer) Store() *storage.RedisStore {
	if r.store == nil && r.Client() != nil {
		// ✅ 使用 redis.Client 创建业务 storage
		r.store = storage.NewRedisStore(r.Client(), r.logger)
	}
	return r.store
}
