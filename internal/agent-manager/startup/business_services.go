// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// BusinessServices contains all core business services.
type BusinessServices struct {
	Registry       *agent.Registry
	EventProcessor *event.Processor
	Dispatcher     *command.Dispatcher
}

// BusinessServicesInitializer creates all core business services.
type BusinessServicesInitializer struct {
	opts   *commonapp.StandardOptions
	logger core.Logger
	infra  *InfrastructureInitializers

	services *BusinessServices
}

// NewBusinessServicesInitializer creates a new business services initializer.
func NewBusinessServicesInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	infra *InfrastructureInitializers,
) *BusinessServicesInitializer {
	return &BusinessServicesInitializer{
		opts:   opts,
		logger: logger,
		infra:  infra,
	}
}

// Name returns the initializer name.
func (s *BusinessServicesInitializer) Name() string {
	return "agent-manager-services"
}

// Priority returns initialization priority.
func (s *BusinessServicesInitializer) Priority() int {
	return 600 // After infrastructure, before NATS and servers
}

// Initialize creates all service instances.
func (s *BusinessServicesInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing agent-manager service layer components")

	// Validate dependencies
	if s.infra.Database.Store() == nil {
		return fmt.Errorf("database store not initialized")
	}
	if s.infra.Redis.Store() == nil {
		return fmt.Errorf("redis store not initialized")
	}

	// Step 1: Create Registry service
	s.logger.Infow("Creating Registry service")
	registry := agent.NewRegistry(
		s.infra.Database.Store(),
		s.infra.Redis.Store(),
		s.logger,
	)

	// Start Registry background tasks
	if err := registry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Step 2: Create EventProcessor service
	s.logger.Infow("Creating EventProcessor service")
	eventProcessor := event.NewProcessor(
		s.infra.Database.Store(),
		s.infra.Redis.Store(),
		nil, // NATS publisher will be set later by NATS initializer
		s.logger,
	)

	// Step 3: Create Dispatcher service
	s.logger.Infow("Creating Dispatcher service")
	dispatcher := command.NewDispatcher(
		s.infra.Database.Store(),
		s.infra.Redis.Store(),
		registry,
		nil, // NATS server will be set later by NATS initializer
		s.logger,
	)

	s.services = &BusinessServices{
		Registry:       registry,
		EventProcessor: eventProcessor,
		Dispatcher:     dispatcher,
	}

	s.logger.Infow("Agent-manager service layer initialized successfully",
		"services", []string{"Registry", "EventProcessor", "Dispatcher"},
	)

	return nil
}

// Services returns the initialized business services.
func (s *BusinessServicesInitializer) Services() *BusinessServices {
	return s.services
}

// Close performs cleanup.
func (s *BusinessServicesInitializer) Close(ctx context.Context) error {
	s.logger.Infow("Closing agent-manager service layer components")

	// Stop Registry background tasks
	if s.services != nil && s.services.Registry != nil {
		if err := s.services.Registry.Stop(); err != nil {
			s.logger.Errorw("Failed to stop registry", "error", err)
		}
	}

	// EventProcessor and Dispatcher don't need explicit cleanup
	return nil
}
