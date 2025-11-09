// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package initializers provides factory functions for Wire dependency injection
// This eliminates the need for service-specific wrapper initializers

package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializerProvider creates a DatabaseInitializer from StandardOptions
// This eliminates the need for internal/*/initializers/database.go wrapper files
func DatabaseInitializerProvider(opts *commonapp.StandardOptions, logger core.Logger) *DatabaseInitializer {
	if opts.Database == nil {
		panic("Database options not configured")
	}
	return NewDatabaseInitializer(opts.Database, logger)
}

// RedisInitializerProvider creates a RedisInitializer from StandardOptions
// This eliminates the need for internal/*/initializers/redis.go wrapper files
func RedisInitializerProvider(opts *commonapp.StandardOptions, logger core.Logger) *RedisInitializer {
	if opts.Redis == nil {
		panic("Redis options not configured")
	}
	return NewRedisInitializer(opts.Redis, logger)
}

// NATSInitializerProvider creates a NATSInitializer from StandardOptions
// This eliminates the need for internal/*/initializers/nats.go wrapper files
func NATSInitializerProvider(opts *commonapp.StandardOptions, logger core.Logger) *NATSInitializer {
	if opts.NATS == nil {
		panic("NATS options not configured")
	}
	return NewNATSInitializer(opts.NATS, logger)
}

// Benefits of using these providers:
// 1. Eliminates 2-3 wrapper files per service (database.go, redis.go, nats.go)
// 2. Reduces code duplication across services
// 3. Centralizes initialization logic in one place
// 4. Reduces total codebase by ~375 lines (75 lines per service × 5 services)

// Usage in wire.go:
// var CoreProviderSet = wire.NewSet(
//     ProvideLogger,
//     pkginitializers.DatabaseInitializerProvider,  // Instead of initializers.NewDatabaseInitializer
//     pkginitializers.RedisInitializerProvider,     // Instead of initializers.NewRedisInitializer
//     pkginitializers.NATSInitializerProvider,      // Instead of initializers.NewNATSInitializer
//     // ... other initializers
// )
