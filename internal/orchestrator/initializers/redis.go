// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/orchestrator/app/options"
	"github.com/kart-io/k8s-agent/internal/orchestrator/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the generic Redis initializer with orchestrator-specific configuration.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	opts   *options.ServerOptions
	logger core.Logger
	store  *storage.RedisStore // Cached storage instance
}

// NewRedisInitializer creates a Redis initializer for orchestrator service.
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// Create the base initializer
	redisInit := pkginitializers.NewRedisInitializer(opts.Redis, logger)

	return &RedisInitializer{
		RedisInitializer: redisInit,
		opts:             opts,
		logger:           logger,
	}
}

// Store returns the initialized storage instance.
// It lazily creates the storage wrapper on first call.
func (r *RedisInitializer) Store() *storage.RedisStore {
	if r.store != nil {
		return r.store
	}

	// Create storage using the configuration (this will create its own connection)
	store, err := storage.NewRedisStore(r.opts.Redis, r.logger)
	if err != nil {
		r.logger.Errorw("Failed to create Redis store", "error", err)
		return nil
	}

	r.store = store
	return r.store
}

