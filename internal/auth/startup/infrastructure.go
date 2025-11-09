// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"

	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// InfrastructureInitializers contains all infrastructure-level initializers.
type InfrastructureInitializers struct {
	Database *pkginitializers.DatabaseInitializer
	Redis    *pkginitializers.RedisInitializer
	Email    *EmailClientInitializer
}

// NewInfrastructureInitializers creates all infrastructure initializers.
func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
	// Create database initializer
	dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)
	if opts.Database.AutoMigrate {
		// TODO: Add Auth models for auto-migration when available
		// dbInit.WithAutoMigrate(&models.User{}, &models.Role{}, ...)
	}

	// Create Redis initializer
	redisInit := pkginitializers.NewRedisInitializer(opts.Redis, logger)

	// Create email client initializer
	emailInit := &EmailClientInitializer{
		opts:   opts,
		logger: logger,
	}

	return &InfrastructureInitializers{
		Database: dbInit,
		Redis:    redisInit,
		Email:    emailInit,
	}
}

// EmailClientInitializer initializes email client configuration.
type EmailClientInitializer struct {
	opts   *commonapp.StandardOptions
	logger core.Logger
}

// Name returns the initializer name.
func (e *EmailClientInitializer) Name() string {
	return "email-client"
}

// Priority returns initialization priority.
func (e *EmailClientInitializer) Priority() int {
	return 500
}

// Initialize validates email configuration is loaded.
func (e *EmailClientInitializer) Initialize(ctx context.Context) error {
	e.logger.Infow("Initializing Email client",
		"smtp_host", e.opts.Email.SMTPHost,
		"smtp_port", e.opts.Email.SMTPPort,
		"enabled", e.opts.Email.Enabled,
	)

	// Email client will be created in notification service initializer
	// This initializer just validates the configuration is loaded

	e.logger.Infow("Email client configuration loaded")
	return nil
}

// Close cleans up resources (no-op for email config).
func (e *EmailClientInitializer) Close(ctx context.Context) error {
	return nil
}
