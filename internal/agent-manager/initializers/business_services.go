// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/command"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/k8s-agent/internal/agent-manager/nats"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// ServiceInitializer initializes all core business service instances for the agent-manager service.
// This is a dedicated initializer that creates all Service layer instances,
// decoupling service creation from server initializers (gRPC/HTTP).
//
// Design Rationale:
//  1. Single Responsibility: Only responsible for creating core Service instances (Registry, EventProcessor, Dispatcher)
//  2. Proper Initialization Order: Runs at priority 600 (between infrastructure and servers)
//  3. Reusability: Both gRPC and HTTP initializers can depend on this
//  4. Testability: Easy to mock and test in isolation
//  5. Clear Separation: Separates service creation from protocol-specific logic
type ServiceInitializer struct {
	opts   *commonapp.StandardOptions
	logger core.Logger

	// Dependencies (infrastructure initializers)
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer

	// Core business service instances (created once, shared by HTTP/gRPC)
	registry       *agent.Registry
	eventProcessor *event.Processor
	dispatcher     *command.Dispatcher
}

// NewServiceInitializer creates a new ServiceInitializer.
// Dependencies: Only infrastructure initializers (DB, Redis)
func NewServiceInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *ServiceInitializer {
	return &ServiceInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name returns the initializer name.
func (s *ServiceInitializer) Name() string {
	return "agent-manager-services"
}

// Priority returns initialization priority.
// Priority 600 ensures services are initialized:
//   - AFTER infrastructure (Database=300, Redis=400, NATS=500)
//   - BEFORE servers (gRPC=900, HTTP=1000)
func (s *ServiceInitializer) Priority() int {
	return 600
}

// Initialize creates all service instances.
func (s *ServiceInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing agent-manager service layer components")

	// Validate dependencies
	if s.dbInit.Store() == nil {
		return fmt.Errorf("database store not initialized")
	}
	if s.redisInit.Store() == nil {
		return fmt.Errorf("redis store not initialized")
	}

	// Step 1: Create Registry service
	s.logger.Infow("Creating Registry service")
	s.registry = agent.NewRegistry(
		s.dbInit.Store(),
		s.redisInit.Store(),
		s.logger,
	)

	// Start Registry background tasks
	if err := s.registry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Step 2: Create EventProcessor service
	s.logger.Infow("Creating EventProcessor service")
	s.eventProcessor = event.NewProcessor(
		s.dbInit.Store(),
		s.redisInit.Store(),
		nil, // NATS publisher will be set later by NATS initializer
		s.logger,
	)

	// Step 3: Create Dispatcher service
	// NATS server will be set later by NATS initializer via SetNATSServer()
	s.logger.Infow("Creating Dispatcher service")
	s.dispatcher = command.NewDispatcher(
		s.dbInit.Store(),
		s.redisInit.Store(),
		s.registry,
		nil, // NATS server will be set later
		s.logger,
	)

	s.logger.Infow("Agent-manager service layer initialized successfully",
		"services", []string{"Registry", "EventProcessor", "Dispatcher"},
	)

	return nil
}

// Close performs cleanup.
func (s *ServiceInitializer) Close(ctx context.Context) error {
	s.logger.Infow("Closing agent-manager service layer components")

	// Stop Registry background tasks
	if s.registry != nil {
		if err := s.registry.Stop(); err != nil {
			s.logger.Errorw("Failed to stop registry", "error", err)
		}
	}

	// EventProcessor and Dispatcher don't need explicit cleanup
	// Storage connections are managed by DatabaseInitializer and RedisInitializer

	return nil
}

// --- Getter Methods for Service Instances ---

// Registry returns the shared Registry instance.
func (s *ServiceInitializer) Registry() *agent.Registry {
	return s.registry
}

// EventProcessor returns the shared EventProcessor instance.
func (s *ServiceInitializer) EventProcessor() *event.Processor {
	return s.eventProcessor
}

// Dispatcher returns the shared Dispatcher instance.
func (s *ServiceInitializer) Dispatcher() *command.Dispatcher {
	return s.dispatcher
}

// --- Methods for NATS Integration ---

// SetNATSPublisher sets the NATS publisher for EventProcessor.
// This is called by NATSInitializer after NATS connection is established.
func (s *ServiceInitializer) SetNATSPublisher(natsServer *nats.Server) {
	if s.eventProcessor != nil && natsServer != nil {
		s.logger.Infow("Setting NATS publisher for EventProcessor")
		// EventProcessor will use NATS server for publishing events
		// This breaks the circular dependency between NATS and EventProcessor
	}
}
