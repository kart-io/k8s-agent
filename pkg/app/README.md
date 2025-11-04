# Application Framework

Unified application framework for standardized service startup, configuration management, and lifecycle management.

## Features

- ✅ Unified configuration loading (config file + command line + environment variables)
- ✅ Automatic configuration validation and defaults
- ✅ Standardized application lifecycle management
- ✅ Graceful shutdown support
- ✅ Signal handling (SIGINT, SIGTERM)
- ✅ Integrated logging initialization
- ✅ Reduced boilerplate code

## Architecture

The framework provides two patterns based on service complexity:

### 1. Bootstrap Pattern (Complex Services)

For services with multiple dependencies and complex initialization:

```go
package app

import (
    "context"
    "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
)

type MyApp struct {
    config *Config
    logger core.Logger
    // initializers...
}

func (a *MyApp) Name() string {
    return "My Service"
}

func (a *MyApp) Initialize(ctx context.Context, opts app.Options) error {
    // Initialize configuration and logger
    return nil
}

func (a *MyApp) Run(ctx context.Context) error {
    // Application main logic
    <-ctx.Done()
    return nil
}

func (a *MyApp) Shutdown(ctx context.Context) error {
    // Cleanup resources
    return nil
}

func Execute() {
    opts := NewOptions()
    myApp := &MyApp{}

    app.RunWithBootstrap(
        myApp,
        opts,
        app.Config{
            Use:       "my-service",
            Short:     "My Service",
            Long:      "My service does something awesome",
            EnvPrefix: "MY_SERVICE",
        },
        myApp.registerComponents,
    )
}

func (a *MyApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // Register all component initializers
    // bs.Register(dbInit)
    // bs.Register(redisInit)
    return nil
}
```

**Used by:** agent-manager, auth, cluster, orchestrator, reasoning

### 2. Simple Pattern (Lightweight Services)

For services with minimal dependencies:

```go
package app

import (
    "context"
    "github.com/kart-io/k8s-agent/pkg/app"
)

type MyApp struct {
    config *Config
    logger core.Logger
}

func (a *MyApp) Name() string {
    return "My Service"
}

func (a *MyApp) Initialize(ctx context.Context, opts app.Options) error {
    // Initialize configuration and logger
    return nil
}

func (a *MyApp) Run(ctx context.Context) error {
    // Application main logic
    return a.server.Run(ctx)
}

func (a *MyApp) Shutdown(ctx context.Context) error {
    // Cleanup resources
    return nil
}

func Execute() {
    opts := NewOptions()
    myApp := &MyApp{}

    app.Run(
        myApp,
        opts,
        app.Config{
            Use:       "my-service",
            Short:     "My Service",
            Long:      "Lightweight service",
            EnvPrefix: "MY_SERVICE",
        },
    )
}
```

**Used by:** collect-agent, gateway, monitor

## Application Interface

All services implement the same clean interface:

```go
type Application interface {
    // Name returns the application name
    Name() string

    // Initialize initializes the application
    Initialize(ctx context.Context, opts Options) error

    // Run runs the application (blocks until shutdown)
    Run(ctx context.Context) error

    // Shutdown gracefully shuts down the application
    Shutdown(ctx context.Context) error
}
```

## Options Interface

Configuration options must implement:

```go
type Options interface {
    // Complete completes the configuration with defaults
    Complete() error

    // Validate validates the configuration
    Validate() []error

    // AddFlags adds command-line flags
    AddFlags(fs *pflag.FlagSet)
}
```

## Key Benefits

1. **Consistent Pattern**: All services follow the same structure
2. **Explicit Dependencies**: No hidden getters or magic methods
3. **Type Safety**: Direct types instead of interface{} wrappers
4. **Clear Lifecycle**: Initialize → Run → Shutdown
5. **Flexibility**: Choose between Bootstrap and Simple patterns based on needs

## Migration from Old Framework

If migrating from the old framework:

1. Remove embedded `StandardBootstrapApplication`
2. Replace `GetLogger()` with direct `logger` field
3. Replace `GetOptions()` with local options variable
4. Implement all 4 methods of the Application interface
5. Use `RunWithBootstrap` for complex services or `Run` for simple ones