// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the generic Redis initializer with monitor-specific configuration.
// Note: Monitor uses a custom storage layer (storage.RedisStorage) for metrics caching.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	store *storage.RedisStorage // Cached storage instance
}

// NewRedisInitializer creates a Redis initializer for monitor service.
func NewRedisInitializer(cfg *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// Create the base initializer
	redisInit := pkginitializers.NewRedisInitializer(cfg.Redis, logger)

	return &RedisInitializer{
		RedisInitializer: redisInit,
	}
}

// Storage returns the initialized storage (for backward compatibility).
// It wraps the Redis client in monitor's storage wrapper.
func (r *RedisInitializer) Storage() *storage.RedisStorage {
	if r.store != nil {
		return r.store
	}

	client := r.Client()
	if client == nil {
		return nil
	}

	// Create storage wrapper around the existing client
	// Note: The storage.RedisStorage has unexported fields, so we can't create it directly
	// We need to use NewRedisStorage, but that will create a new connection
	// TODO: Refactor storage.RedisStorage to accept an existing client
	// For now, return nil and let the service layer handle storage creation
	return nil
}
