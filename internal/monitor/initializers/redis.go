// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
        commonapp "github.com/kart-io/k8s-agent/pkg/app"
	
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the generic Redis initializer with monitor-specific configuration.
// Note: Monitor uses a custom storage layer (storage.RedisStorage) for metrics caching.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	logger core.Logger                // Logger for this wrapper
	store  *storage.RedisStorage // Cached storage instance
}

// NewRedisInitializer creates a Redis initializer for monitor service.
func NewRedisInitializer(cfg *commonapp.StandardOptions, logger core.Logger) *RedisInitializer {
	// Create the base initializer
	redisInit := pkginitializers.NewRedisInitializer(cfg.Redis, logger)

	return &RedisInitializer{
		RedisInitializer: redisInit,
		logger:           logger,
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

	// Use the existing Redis client from the base initializer
	// This avoids creating duplicate Redis connections
	store, err := storage.NewRedisStorageWithClient(client, r.logger)
	if err != nil {
		r.logger.Errorw("Failed to create Redis storage wrapper", "error", err)
		return nil
	}

	r.store = store
	r.logger.Infow("Monitor storage initialized with reused Redis connection")
	return r.store
}
