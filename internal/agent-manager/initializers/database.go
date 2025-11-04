package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/agent-manager/app/options"
	"github.com/kart-io/k8s-agent/internal/agent-manager/storage"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/k8s-agent/pkg/types"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer wraps the generic database initializer with service-specific configuration.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
	store *storage.PostgresStore
}

// NewDatabaseInitializer creates a database initializer for agent-manager service.
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	// Create the base initializer
	dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)

	// Configure auto-migration if enabled
	if opts.Database.AutoMigrate {
		dbInit.WithAutoMigrate(
			&types.Agent{},
			&types.Event{},
			&types.Metrics{},
			&types.Command{},
			&types.CommandResult{},
			&types.Cluster{},
			&types.AlertRule{},
			&types.Alert{},
		)
	}

	return &DatabaseInitializer{
		DatabaseInitializer: dbInit,
	}
}

// Store returns the storage instance (creates on first call).
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
	if d.store == nil && d.Client() != nil {
		d.store = &storage.PostgresStore{
			MySQLClient: d.Client(),
		}
	}
	return d.store
}
