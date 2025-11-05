package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/cmd/monitor/app/options"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer handles database initialization.
type DatabaseInitializer struct {
	cfg     *options.ServerOptions
	logger  core.Logger
	storage *storage.PostgresStorage
}

// NewDatabaseInitializer creates a new database initializer.
func NewDatabaseInitializer(cfg *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		cfg:    cfg,
		logger: logger,
	}
}

// Name returns the initializer name.
func (d *DatabaseInitializer) Name() string {
	return "monitor-database"
}

// Priority returns initialization priority (lower runs first).
func (d *DatabaseInitializer) Priority() int {
	return bootstrap.PriorityDatabase
}

// Initialize initializes the database connection.
func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
	d.logger.Infow("Initializing database connection")

	storage, err := storage.NewPostgresStorage(d.cfg.Database, d.logger)
	if err != nil {
		return err
	}

	d.storage = storage
	d.logger.Infow("Database initialized successfully")
	return nil
}

// Run does nothing - database is passive.
func (d *DatabaseInitializer) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Close closes the database connection.
func (d *DatabaseInitializer) Close(ctx context.Context) error {
	if d.storage != nil {
		d.logger.Infow("Closing database connection")
		return d.storage.Close()
	}
	return nil
}

// Storage returns the initialized storage.
// Deprecated: Use Store() instead for consistency across services.
func (d *DatabaseInitializer) Storage() *storage.PostgresStorage {
	return d.storage
}

// Store returns the initialized storage instance.
// This is the standard method name across all database initializers.
func (d *DatabaseInitializer) Store() interface{} {
	return d.storage
}

