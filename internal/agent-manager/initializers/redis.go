package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// RedisInitializer wraps the generic Redis initializer with service-specific configuration.
type RedisInitializer struct {
	*pkginitializers.RedisInitializer
	store *storage.RedisStore
}

// NewRedisInitializer creates a Redis initializer for agent-manager service.
func NewRedisInitializer(opts *options.ServerOptions, logger core.Logger) *RedisInitializer {
	// Create the base initializer
	redisInit := pkginitializers.NewRedisInitializer(opts.Redis, logger)

	return &RedisInitializer{
		RedisInitializer: redisInit,
	}
}

// Store returns the storage instance (creates on first call).
func (r *RedisInitializer) Store() *storage.RedisStore {
	if r.store == nil && r.RedisClient() != nil {
		r.store = &storage.RedisStore{
			RedisClient: r.RedisClient(),
		}
	}
	return r.store
}
