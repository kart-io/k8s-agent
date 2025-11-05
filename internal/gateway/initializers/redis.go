package initializers

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/cmd/gateway/app/options"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the common Redis initializer for gateway service.
// Gateway has a special requirement: Redis is optional (for rate limiting).
// This wrapper handles the case where Redis is unavailable.
type RedisInitializer struct {
	baseInit *pkginitializers.RedisInitializer
	logger   core.Logger
}

// NewRedisInitializer creates a Redis initializer for gateway service.
func NewRedisInitializer(cfg *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// Use the common Redis initializer from pkg/initializers
	redisInit := pkginitializers.NewRedisInitializer(cfg.Redis, logger)

	return &RedisInitializer{
		baseInit: redisInit,
		logger:   logger,
	}
}

// Name returns the initializer name.
func (r *RedisInitializer) Name() string {
	return r.baseInit.Name()
}

// Priority returns initialization priority.
func (r *RedisInitializer) Priority() int {
	return r.baseInit.Priority()
}

// Initialize initializes the Redis connection.
// Unlike other services, gateway treats Redis as optional for rate limiting.
func (r *RedisInitializer) Initialize(ctx context.Context) error {
	err := r.baseInit.Initialize(ctx)
	if err != nil {
		// Gateway-specific: Redis is optional, fail gracefully
		r.logger.Warnw("Failed to connect to Redis, rate limiting will use local mode",
			"error", err,
		)
		return nil // Don't fail if Redis is unavailable
	}
	return nil
}

// Close closes the Redis connection.
func (r *RedisInitializer) Close(ctx context.Context) error {
	return r.baseInit.Close(ctx)
}

// HealthCheck checks Redis health.
func (r *RedisInitializer) HealthCheck(ctx context.Context) error {
	// Gateway doesn't require Redis for health check
	if r.baseInit.Client() == nil {
		return nil // Redis is optional
	}
	return r.baseInit.HealthCheck(ctx)
}

// Client returns the initialized Redis client.
func (r *RedisInitializer) Client() *redis.Client {
	return r.baseInit.Client()
}
