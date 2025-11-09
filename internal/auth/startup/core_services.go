// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// CoreServices contains all core business services.
type CoreServices struct {
	Auth       *service.AuthService
	User       *service.UserService
	Role       *service.RoleService
	Permission *service.PermissionService
	APIKey     *service.APIKeyService

	// Storage wrappers (shared by all services)
	MySQLDB     *storage.MySQLDB
	RedisClient *storage.RedisClient
}

// CoreServicesInitializer creates all core business services.
type CoreServicesInitializer struct {
	opts      *commonapp.StandardOptions
	logger    core.Logger
	dbInit    *pkginitializers.DatabaseInitializer
	redisInit *pkginitializers.RedisInitializer

	services *CoreServices
}

// NewCoreServicesInitializer creates a new core services initializer.
func NewCoreServicesInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *pkginitializers.DatabaseInitializer,
	redisInit *pkginitializers.RedisInitializer,
) *CoreServicesInitializer {
	return &CoreServicesInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name returns the initializer name.
func (s *CoreServicesInitializer) Name() string {
	return "auth-services"
}

// Priority returns initialization priority.
func (s *CoreServicesInitializer) Priority() int {
	return 600 // After infrastructure, before forced-logout services
}

// Initialize creates all core service instances.
func (s *CoreServicesInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing auth service layer components")

	// Step 1: Create storage wrappers (shared by all services)
	mysqlDB := &storage.MySQLDB{
		DB:     s.dbInit.DB(),
		Logger: s.logger,
	}
	redisClient := &storage.RedisClient{
		Client: s.redisInit.Client(),
	}

	// Step 2: Create business service instances
	authService := service.NewAuthService(
		mysqlDB,
		redisClient,
		s.opts,
		s.logger,
	)

	userService := service.NewUserService(
		mysqlDB,
		s.logger,
	)

	roleService := service.NewRoleService(
		mysqlDB,
		s.logger,
	)

	permissionService := service.NewPermissionService(
		mysqlDB,
		s.logger,
	)

	apiKeyService := service.NewAPIKeyService(
		mysqlDB,
	)

	s.services = &CoreServices{
		Auth:        authService,
		User:        userService,
		Role:        roleService,
		Permission:  permissionService,
		APIKey:      apiKeyService,
		MySQLDB:     mysqlDB,
		RedisClient: redisClient,
	}

	s.logger.Infow("Auth service layer initialized successfully",
		"services", []string{"Auth", "User", "Role", "Permission", "APIKey"},
	)

	return nil
}

// Services returns the initialized core services.
func (s *CoreServicesInitializer) Services() *CoreServices {
	return s.services
}

// Close cleans up resources (no-op as services don't need cleanup).
func (s *CoreServicesInitializer) Close(ctx context.Context) error {
	return nil
}
