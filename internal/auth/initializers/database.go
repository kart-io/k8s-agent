package initializers

import (
	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// DatabaseInitializer wraps the generic database initializer with auth-specific configuration.
type DatabaseInitializer struct {
	*pkginitializers.DatabaseInitializer
}

// NewDatabaseInitializer creates a database initializer for auth service.
func NewDatabaseInitializer(cfg *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
	// Create the base initializer
	dbInit := pkginitializers.NewDatabaseInitializer(cfg.Database, logger)

	// Note: Models will be defined in internal/auth/models package
	// When models are created, uncomment the following:
	// if cfg.Database.AutoMigrate {
	//     dbInit.WithAutoMigrate(&models.User{}, &models.Session{}, ...)
	// }

	return &DatabaseInitializer{
		DatabaseInitializer: dbInit,
	}
}
