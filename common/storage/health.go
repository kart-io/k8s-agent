package storage

import "context"

// HealthChecker defines the interface for checking storage health.
type HealthChecker interface {
	// Health checks if the storage backend is healthy.
	// Returns nil if healthy, error otherwise.
	Health(ctx context.Context) error
}
