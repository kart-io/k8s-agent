package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// Client wraps go-redis client with lifecycle management.
// It provides connection pooling, health checks, and graceful shutdown.
type Client struct {
	client *redis.Client
	logger core.Logger
	config *options.RedisOptions
}

// NewClient creates a new Redis client with the given configuration.
// It establishes a connection and verifies connectivity.
//
// Example usage:
//
//	redisOpts := options.NewRedisOptions()
//	redisOpts.Addr = "localhost:6379"
//	client, err := redis.NewClient(redisOpts, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(opts *options.RedisOptions, log core.Logger) (*Client, error) {
	// Validate configuration
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	log.Infow("Connecting to Redis",
		"addr", opts.Addr,
		"db", opts.DB,
	)

	// Create Redis client
	rdb, err := opts.ConnectRedis(log)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.DialTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	client := &Client{
		client: rdb,
		logger: log.With("component", "redis"),
		config: opts,
	}

	client.logger.Infow("Redis connected successfully",
		"addr", opts.Addr,
		"db", opts.DB,
	)

	return client, nil
}

// Client returns the underlying go-redis client for direct access.
// This allows using all redis commands directly.
func (c *Client) Client() *redis.Client {
	return c.client
}

// Health checks if the Redis connection is healthy by pinging it.
func (c *Client) Health(ctx context.Context) error {
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}

// Close closes the Redis connection.
// It should be called when shutting down the application.
func (c *Client) Close() error {
	c.logger.Info("Closing Redis connection")
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}
	return nil
}
