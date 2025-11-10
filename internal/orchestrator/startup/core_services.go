// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/orchestrator/service"
	"github.com/kart-io/k8s-agent/internal/orchestrator/strategy"
	"github.com/kart-io/k8s-agent/internal/orchestrator/workflow"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// CoreServices contains all core business services.
type CoreServices struct {
	WorkflowEngine  *workflow.Engine
	StrategyManager *strategy.Manager
	WorkflowService *service.WorkflowServiceServer
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
	return "orchestrator-services"
}

// Priority returns initialization priority.
func (s *CoreServicesInitializer) Priority() int {
	return 600 // After infrastructure (300-500), before event subscriber (700)
}

// Initialize creates all core service instances.
func (s *CoreServicesInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing orchestrator service layer components",
		"agent_manager_url", s.opts.AI.AgentManagerURL,
		"reasoning_service_url", s.opts.AI.ReasoningServiceURL,
		"global_timeout", s.opts.Workflow.GlobalTimeout,
		"step_default_timeout", s.opts.Workflow.StepDefaultTimeout,
		"retry_on_timeout", s.opts.Workflow.RetryOnTimeout,
		"max_retries", s.opts.Workflow.MaxRetries,
	)

	// Step 1: Get direct DB and Redis clients from pkg/initializers
	db := s.dbInit.DB()
	if db == nil {
		return fmt.Errorf("mysql db not initialized")
	}

	redisClient := s.redisInit.Client()
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Step 2: Create workflow engine with direct clients
	executor := workflow.NewExecutor(
		s.opts.AI.AgentManagerURL,
		s.opts.AI.ReasoningServiceURL,
		s.logger,
	)

	workflowEngine := workflow.NewEngine(
		db,
		redisClient,
		executor,
		s.logger,
	)

	// Set timeout configuration
	workflowEngine.SetTimeoutConfig(
		s.opts.Workflow.GlobalTimeout,
		s.opts.Workflow.StepDefaultTimeout,
		s.opts.Workflow.RetryOnTimeout,
		s.opts.Workflow.MaxRetries,
	)

	// Step 3: Create strategy manager with direct DB
	strategyManager := strategy.NewManager(
		db,
		workflowEngine,
		s.logger,
	)

	// Step 4: Create workflow service with direct DB
	workflowService := service.NewWorkflowServiceServer(
		workflowEngine,
		db,
		s.logger,
	)

	s.services = &CoreServices{
		WorkflowEngine:  workflowEngine,
		StrategyManager: strategyManager,
		WorkflowService: workflowService,
	}

	s.logger.Infow("Orchestrator service layer initialized successfully",
		"services", []string{"WorkflowEngine", "StrategyManager", "WorkflowService"},
	)

	return nil
}

// Services returns the initialized core services.
func (s *CoreServicesInitializer) Services() *CoreServices {
	return s.services
}

// Close cleans up resources (no-op as services don't need cleanup).
func (s *CoreServicesInitializer) Close(ctx context.Context) error {
	s.logger.Infow("Closing orchestrator service layer components")
	// Services don't hold resources that need explicit cleanup
	return nil
}
