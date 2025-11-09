// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package templates

/*
SERVICE STARTUP TEMPLATES

This file documents the standard service startup patterns used in the Aetherius project.
Choose the appropriate pattern based on your service's complexity and dependencies.

For complete implementation examples, see:
- cmd/agent-manager/app/app.go        - Bootstrap Pattern (Complex Service)
- cmd/auth/app/app.go                 - Bootstrap Pattern (with multiple features)
- cmd/collect-agent/app/app.go        - Ultra-Simple Pattern (minimal dependencies)
- cmd/gateway/app/app.go              - Ultra-Simple Pattern (stateless service)

=============================================================================
PATTERN 1: BOOTSTRAP PATTERN (Complex Services)
=============================================================================

Use this pattern when:
- Service has 5+ external dependencies (DB, Redis, NATS, gRPC, etc.)
- Complex initialization order dependencies
- Multiple feature modules with interdependencies
- Service needs fine-grained lifecycle management
- Service complexity score >= 10

Examples: agent-manager, orchestrator, auth, cluster, reasoning

STRUCTURE:
cmd/{service}/app/
├── app.go              - Main application entry point, implements Application interface
├── components.go       - Component registration and wiring
└── wire.go            - Optional: Wire dependency injection (if using Wire)

KEY COMPONENTS:
1. Execute() function - Entry point called from main.go
2. {Service}App struct - Implements commonapp.Application interface
3. StandardOptions - Configured with required capabilities
4. registerComponents() - Define initialization order using Bootstrap
5. Service-specific startup/ package - Grouped initializers in internal/{service}/startup/

BENEFITS:
- Clear dependency management
- Priority-based initialization (300-2000)
- Automatic cleanup in reverse order
- Health check integration
- Server auto-start support

TYPICAL INITIALIZATION LAYERS:
  Layer 1: Infrastructure (Priority 300-500)
    - Database, Redis, Email, external services
  Layer 2: Business Services (Priority 600-700)
    - Core service logic, repositories
  Layer 3: Advanced Services (Priority 650-800)
    - Session management, feature services
  Layer 4: Server Layer (Priority 900-1000)
    - gRPC servers, HTTP servers
  Layer 5: Monitoring (Priority 2000)
    - Health checks, metrics

---

TEMPLATE CODE:

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/{service}/startup"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

const (
	UserAgent = "aetherius-{service}"
)

// Execute runs the {service} command.
func Execute() {
	// Step 1: Create StandardOptions with required capabilities
	opts := commonapp.NewStandardOptions("{Service}", UserAgent).
		WithDatabase().
		WithRedis().
		// WithNATS(), WithGRPC(), WithEmail(), WithJWT(), WithMetrics() as needed
		WithMetrics()

	// Step 2: Create application instance
	app := &{Service}App{}

	// Step 3: Run with Bootstrap framework
	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "{service}",
			Short:     "{Service} Service",
			Long:      "Detailed description of {Service} Service",
			EnvPrefix: "{SERVICE_NAME}",
		},
		app.registerComponents,
	)
}

// {Service}App implements commonapp.Application interface.
type {Service}App struct {
	bootstrap *bootstrap.Bootstrap
	opts      *commonapp.StandardOptions
	logger    core.Logger
}

// Name returns the application name.
func (a *{Service}App) Name() string {
	return "{Service} Service"
}

// Initialize initializes the application.
func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	return nil
}

// Run runs the application.
func (a *{Service}App) Run(ctx context.Context) error {
	// Bootstrap.Run() is already called by RunWithBootstrap
	// We just wait for context cancellation
	<-ctx.Done()
	return ctx.Err()
}

// Shutdown gracefully shuts down the application.
func (a *{Service}App) Shutdown(ctx context.Context) error {
	// Bootstrap shutdown is handled automatically
	return nil
}

// registerComponents registers all component initializers with bootstrap.
// This method defines the complete initialization order for the service.
func (a *{Service}App) registerComponents(bs *bootstrap.Bootstrap) error {
	a.bootstrap = bs

	// Layer 1: Infrastructure (Priority 300-500)
	infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
	bs.Register(infra.Database)
	bs.Register(infra.Redis)

	// Layer 2: Core Business Services (Priority 600-700)
	coreServices := startup.NewCoreServicesInitializer(
		a.opts,
		a.logger,
		infra.Database,
		infra.Redis,
	)
	bs.Register(coreServices)

	// Layer 3: Advanced Services (Priority 650-800) - if needed
	if a.opts.NATS.Enable {
		natsInit := startup.NewNATSInitializer(a.opts, a.logger, coreServices)
		bs.Register(natsInit)
	}

	// Layer 4: Server Layer (Priority 900-1000)
	if a.opts.GRPC.Enable {
		grpcInit := startup.NewGRPCServerInitializer(
			a.opts,
			a.logger,
			infra,
			coreServices,
		)
		bs.Register(grpcInit)
	}

	httpInit := startup.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		infra,
		coreServices,
	)
	bs.Register(httpInit)

	// Layer 5: Monitoring (Priority 2000)
	healthInit := pkginitializers.NewHealthCheckInitializer(a.opts.Health, a.logger)
	bs.Register(healthInit)

	return nil
}

---

SERVICE-SPECIFIC STARTUP PACKAGE: internal/{service}/startup/

This package contains grouped initializers organized by functionality.

TYPICAL STRUCTURE:

internal/{service}/startup/
├── infrastructure.go      - Database, Redis, external client initializers
├── core_services.go       - Main business logic services
├── nats.go               - NATS message queue setup (if needed)
├── servers.go            - HTTP and gRPC server initializers
└── [feature]_services.go - Feature-specific services (e.g., session, audit)

EXAMPLE: infrastructure.go

package startup

type InfrastructureInitializers struct {
	Database *DatabaseInitializer
	Redis    *RedisInitializer
	Email    *EmailInitializer  // if needed
}

func NewInfrastructureInitializers(opts *StandardOptions, logger Logger) *InfrastructureInitializers {
	dbInit := &DatabaseInitializer{
		DatabaseInitializer: pkginitializers.NewDatabaseInitializer(opts.Database, logger),
	}
	if opts.Database.AutoMigrate {
		dbInit.WithAutoMigrate(&types.Agent{}, &types.Event{}, &types.Metrics{})
	}

	redisInit := &RedisInitializer{
		RedisInitializer: pkginitializers.NewRedisInitializer(opts.Redis, logger),
	}

	return &InfrastructureInitializers{
		Database: dbInit,
		Redis:    redisInit,
	}
}

=============================================================================
PATTERN 2: ULTRA-SIMPLE PATTERN (Simple Services)
=============================================================================

Use this pattern when:
- Service has 0-3 external dependencies
- Linear, straightforward initialization logic
- Lightweight services (agents, gateways, monitoring)
- Service complexity score < 10
- No complex interdependencies

Examples: collect-agent, gateway, monitor

STRUCTURE:
cmd/{service}/
├── main.go          - Minimal entry point
└── app/
    └── app.go       - Application with direct initialization

KEY CHARACTERISTICS:
1. No Bootstrap framework
2. Direct component initialization in Initialize()
3. No separate startup/ package needed
4. Simple context-based shutdown
5. All initialization in one logical flow

BENEFITS:
- Minimal boilerplate
- Easy to understand at a glance
- Fast startup with no dependency resolution overhead
- Perfect for edge agents and simple services

---

TEMPLATE CODE:

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/{service}/agent"
	"github.com/kart-io/k8s-agent/internal/{service}/types"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	UserAgent = "aetherius-{service}"
)

// Execute runs the {service} command.
func Execute() {
	// Step 1: Create StandardOptions with only required capabilities
	opts := commonapp.NewStandardOptions("{Service}", UserAgent).
		WithAgent() // or WithServer(), WithLogging(), etc.

	// Step 2: Create application instance
	app := &{Service}App{}

	// Step 3: Run with standard Run() function (no Bootstrap)
	commonapp.Run(
		app,
		opts,
		commonapp.Config{
			Use:       "{service}",
			Short:     "{Service}",
			Long:      "Description of {Service}",
			EnvPrefix: "{SERVICE_NAME}",
		},
	)
}

// {Service}App implements commonapp.Application interface.
// This is the ULTRA-SIMPLE pattern: no Bootstrap, direct initialization.
type {Service}App struct {
	opts   *commonapp.StandardOptions
	logger core.Logger
	// Service-specific components
	server *http.Server
	client *MyClient
}

// Name returns the application name.
func (a *{Service}App) Name() string {
	return "{Service}"
}

// Initialize initializes the application components directly.
func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing {Service}",
		"port", a.opts.Server.Port,
		"debug", a.opts.Logging.Debug,
	)

	// Create service components in order
	client, err := NewMyClient(a.opts, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	a.client = client

	// Create HTTP server if needed
	a.server = NewHTTPServer(a.opts, a.logger, a.client)

	return nil
}

// Run runs the application.
func (a *{Service}App) Run(ctx context.Context) error {
	a.logger.Info("Starting {Service}")

	// Start server (blocks until context cancelled)
	return a.server.Start(ctx)
}

// Shutdown gracefully shuts down the application.
func (a *{Service}App) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down {Service}")

	// Shutdown in reverse order of initialization
	if a.server != nil {
		if err := a.server.Stop(ctx); err != nil {
			a.logger.Warnw("Server shutdown error", "error", err)
		}
	}

	if a.client != nil {
		if err := a.client.Close(ctx); err != nil {
			a.logger.Warnw("Client close error", "error", err)
		}
	}

	a.logger.Info("{Service} shutdown complete")
	return nil
}

---

MAIN.GO (same for both patterns):

package main

import (
	_ "go.uber.org/automaxprocs/maxprocs"
	"github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
	app.Execute()
}

=============================================================================
DECISION FLOWCHART
=============================================================================

1. Does service have 5+ external dependencies (DB, Redis, NATS, etc.)?
   YES -> Use BOOTSTRAP PATTERN
   NO  -> Continue

2. Does service have complex initialization order requirements?
   YES -> Use BOOTSTRAP PATTERN
   NO  -> Continue

3. Does service have multiple feature modules with interdependencies?
   YES -> Use BOOTSTRAP PATTERN
   NO  -> Continue

4. Is service complexity >= 10?
   YES -> Use BOOTSTRAP PATTERN
   NO  -> Use ULTRA-SIMPLE PATTERN

=============================================================================
COMMON PITFALLS & SOLUTIONS
=============================================================================

PITFALL 1: Initialization Order Issues (Bootstrap Pattern)

PROBLEM: Components fail to initialize because dependencies are not ready.
SOLUTION:
- Use Priority() to control initialization order
- Layer 1 (Priority 300-500): Infrastructure only
- Layer 2 (Priority 600-700): Services that depend on infrastructure
- Layer 3 (Priority 800): Services that depend on Layer 2
- Layer 4 (Priority 900-1000): Servers that depend on everything
- Layer 5 (Priority 2000): Monitoring

Example:
	db := 300
	redis := 350
	services := 600
	nats := 650
	servers := 900
	health := 2000

PITFALL 2: Circular Dependencies

PROBLEM: Service A needs Service B, Service B needs Service A
SOLUTION:
- Avoid circular dependencies in initialization
- Use interface-based design to break cycles
- Register components in correct order
- Pass initializers, not fully initialized components

BAD:
	serviceA := startup.NewServiceA(serviceB)
	serviceB := startup.NewServiceB(serviceA)

GOOD:
	init := startup.NewServiceBInitializer(...)
	bs.Register(init)
	// ServiceA is registered after and gets reference to initialized ServiceB

PITFALL 3: Resource Leaks (Ultra-Simple Pattern)

PROBLEM: Shutdown() not called properly or partial cleanup
SOLUTION:
- Always implement Shutdown() even if empty
- Track all resources that need cleanup
- Implement in reverse order of initialization
- Use defer for guaranteed cleanup
- Test shutdown path explicitly

PITFALL 4: Configuration Not Applied

PROBLEM: StandardOptions created but not used during initialization
SOLUTION:
- Always check opts.{Feature}.Enable before initializing
- Example:
	if a.opts.NATS.Enable {
		natsInit := startup.NewNATSInitializer(...)
		bs.Register(natsInit)
	}

PITFALL 5: Logger Not Initialized Early

PROBLEM: Logger not available during initialization
SOLUTION:
- Initialize logger in Initialize() before anything else
- Use logger throughout initialization for debugging

func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// FIRST: Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	// THEN: Use logger for subsequent initialization
	a.logger.Info("Initializing components...")
}

=============================================================================
TESTING PATTERNS
=============================================================================

BOOTSTRAP PATTERN TESTING:

func TestBootstrap_InitializationOrder(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	opts := testutil.NewTestOptions()

	app := &{Service}App{}
	bs := bootstrap.New(logger)

	// Register components in correct order
	err := app.registerComponents(bs)
	require.NoError(t, err)

	// Initialize all components
	ctx := context.Background()
	err = bs.Initialize(ctx)
	require.NoError(t, err)

	// Verify all components are initialized
	// ...

	// Shutdown
	err = bs.Shutdown(ctx, 5*time.Second)
	require.NoError(t, err)
}

ULTRA-SIMPLE PATTERN TESTING:

func TestApp_Initialize(t *testing.T) {
	logger := testutil.NewTestLogger(t)
	opts := testutil.NewTestOptions()

	app := &{Service}App{}
	err := app.Initialize(context.Background(), opts)
	require.NoError(t, err)

	// Verify components are initialized
	require.NotNil(t, app.client)
	require.NotNil(t, app.server)

	// Test shutdown
	err = app.Shutdown(context.Background())
	require.NoError(t, err)
}

=============================================================================
MIGRATION GUIDE: ADDING TO EXISTING PATTERN
=============================================================================

TO ADD A NEW COMPONENT TO BOOTSTRAP PATTERN:

1. Create initializer in internal/{service}/startup/{component}.go:

	type MyComponentInitializer struct {
		*pkginitializers.BaseInitializer
		logger core.Logger
		// Other fields
	}

	func NewMyComponentInitializer(opts Options, logger core.Logger) *MyComponentInitializer {
		return &MyComponentInitializer{
			BaseInitializer: pkginitializers.NewBaseInitializer(
				"MyComponent",
				700,  // Priority
			),
			logger: logger,
		}
	}

	func (i *MyComponentInitializer) Initialize(ctx context.Context) error {
		// Initialize logic here
		return nil
	}

	func (i *MyComponentInitializer) Close(ctx context.Context) error {
		// Cleanup logic here
		return nil
	}

2. Register in registerComponents():

	myCompInit := startup.NewMyComponentInitializer(a.opts, a.logger)
	bs.Register(myCompInit)

3. Use the component in dependent services:

	serviceInit := startup.NewServiceInitializer(
		a.opts,
		a.logger,
		myCompInit,
	)
	bs.Register(serviceInit)

=============================================================================
STANDARDOPTIONS CAPABILITIES
=============================================================================

Available options builders:

opts := commonapp.NewStandardOptions("Name", "UserAgent")
	.WithServer()        // HTTP server (port, health check, metrics)
	.WithDatabase()      // MySQL client
	.WithRedis()         // Redis client
	.WithNATS()          // NATS messaging
	.WithGRPC()          // gRPC server
	.WithJWT()           // JWT authentication
	.WithEmail()         // Email client
	.WithMetrics()       // Prometheus metrics
	.WithAgent()         // Agent configuration (for collect-agent)
	.WithCORS()          // CORS middleware
	.WithLogging()       // Logging configuration

Environment variables override config file settings:
	SERVER_PORT=9000
	DATABASE_HOST=localhost
	REDIS_ADDR=redis:6379
	NATS_URL=nats://localhost:4222
	PREFIX_OPTION=value (where PREFIX is EnvPrefix from Config)

=============================================================================
BOOTSTRAP INITIALIZATION CONTEXT
=============================================================================

In Bootstrap pattern, registerComponents() can access:

	func (a *{Service}App) registerComponents(bs *bootstrap.Bootstrap) error {
		// Access via a.opts which is already populated
		opts := a.opts

		// All StandardOptions are available:
		opts.Server.Port
		opts.Database.Host
		opts.Redis.Addr
		opts.NATS.URL
		opts.Health.Port
		opts.Metrics.Enable

		// Feature flags:
		opts.GRPC.Enable
		opts.NATS.Enable

		// Logger:
		a.logger

		// Register components based on configuration
		return nil
	}

*/
