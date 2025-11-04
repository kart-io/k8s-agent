package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the generic Redis initializer with auth-specific configuration.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
}

// NewRedisInitializer creates a Redis initializer for auth service.
func NewRedisInitializer(cfg *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// Create the base initializer
	redisInit := pkginitializers.NewRedisInitializer(cfg.Redis, logger)

	return &RedisInitializer{
		RedisInitializer: redisInit,
	}
}
