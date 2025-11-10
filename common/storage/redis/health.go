package redis

import (
	"context"
)

// HealthCheck checks if Redis is healthy using a default timeout.
// This is a convenience method that uses the dial timeout.
// For custom timeout, use Health(ctx) method on the Client.
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.DialTimeout)
	defer cancel()
	return c.Health(ctx)
}

// PoolStats returns connection pool statistics.
type PoolStats struct {
	Hits       uint32 // Number of times a connection was found in the pool
	Misses     uint32 // Number of times a connection was not found in the pool
	Timeouts   uint32 // Number of times a timeout occurred waiting for a connection
	TotalConns uint32 // Total number of connections in the pool
	IdleConns  uint32 // Number of idle connections in the pool
	StaleConns uint32 // Number of stale connections removed from the pool
}

// PoolStats returns the current connection pool statistics.
func (c *Client) PoolStats() *PoolStats {
	stats := c.client.PoolStats()
	return &PoolStats{
		Hits:       stats.Hits,
		Misses:     stats.Misses,
		Timeouts:   stats.Timeouts,
		TotalConns: stats.TotalConns,
		IdleConns:  stats.IdleConns,
		StaleConns: stats.StaleConns,
	}
}
