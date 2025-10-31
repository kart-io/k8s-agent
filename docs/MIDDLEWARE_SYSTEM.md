# Middleware System

## Overview

The middleware system provides a flexible, priority-based extension mechanism for Bootstrap applications. Middleware functions execute **before component registration**, allowing you to inject additional initializers dynamically based on runtime conditions.

## Key Concepts

### Middleware Function

A middleware function has access to the bootstrap instance, logger, and application options:

```go
type MiddlewareFunc func(bs *bootstrap.Bootstrap, logger core.Logger, opts Options) error
```

### Middleware Configuration

Each middleware has a name, priority, and function:

```go
type MiddlewareConfig struct {
    Name     string         // Middleware name (for logging)
    Priority int            // Execution priority (lower = earlier)
    Func     MiddlewareFunc // Middleware function
}
```

### Execution Order

Middlewares execute in **priority order** (ascending):
- Priority 100: Very early (metrics, tracing)
- Priority 200-400: Infrastructure components
- Priority 500: Default priority (business logic)
- Priority 600+: Late-stage components

## Usage

### Basic Usage

```go
import (
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
)

type MyApp struct {
    *commonapp.StandardBootstrapApplication
}

func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    if a.StandardBootstrapApplication == nil {
        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("MyService", a).
            // Add single middleware
            WithMiddleware(commonapp.Middleware("CustomInit", 300, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
                // Your initialization logic
                logger.Infow("Custom middleware executing")
                return nil
            })).
            // Add simple middleware with default priority (500)
            WithMiddlewareFunc("SimpleInit", func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
                logger.Infow("Simple middleware")
                return nil
            })
    }
    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

### Batch Middleware Registration

```go
func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    middlewares := []commonapp.MiddlewareConfig{
        commonapp.Middleware("Metrics", 100, metricsMiddleware),
        commonapp.Middleware("Tracing", 150, tracingMiddleware),
        commonapp.Middleware("Cache", 200, cacheMiddleware),
    }

    if a.StandardBootstrapApplication == nil {
        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("MyService", a).
            WithMiddlewares(middlewares...)
    }
    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

### Conditional Middleware

Use `ConditionalMiddleware` to apply middleware only when conditions are met:

```go
import commonapp "github.com/kart-io/k8s-agent/pkg/app"

func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    if a.StandardBootstrapApplication == nil {
        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("MyService", a).
            WithMiddleware(
                commonapp.ConditionalMiddleware(
                    "DevTools",
                    100,
                    func(opts commonapp.Options) bool {
                        // Only enable in development environment
                        return opts.GetEnv() == "development"
                    },
                    func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
                        logger.Infow("Development tools enabled")
                        // Register development initializers
                        return nil
                    },
                ),
            )
    }
    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

## Predefined Middleware Examples

The framework provides predefined middleware helpers in `pkg/app/middleware.go`:

### Metrics Middleware

```go
func metricsInitializer(port int, logger core.Logger) bootstrap.Initializer {
    return &MyMetricsInitializer{port: port, logger: logger}
}

app.WithMiddleware(
    commonapp.MetricsMiddleware(9090, metricsInitializer),
)
```

### Tracing Middleware

```go
app.WithMiddleware(
    commonapp.TracingMiddleware("my-service", "http://jaeger:14268/api/traces"),
)
```

### Profiling Middleware

```go
app.WithMiddleware(
    commonapp.ProfilingMiddleware(6060),  // Enable pprof on port 6060
)
```

### Rate Limiting Middleware

```go
app.WithMiddleware(
    commonapp.RateLimitMiddleware(1000),  // 1000 requests per second
)
```

## Creating Custom Middleware

### Simple Custom Middleware

```go
func MyCustomMiddleware() commonapp.MiddlewareConfig {
    return commonapp.CustomMiddleware("MyMiddleware", 500, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
        logger.Infow("Executing custom middleware")

        // Add initializers to bootstrap
        init := &MyCustomInitializer{...}
        bs.Register(init)

        return nil
    })
}

// Usage
app.WithMiddleware(MyCustomMiddleware())
```

### Advanced Custom Middleware

```go
type CacheMiddlewareConfig struct {
    Type     string
    Host     string
    Port     int
    Priority int
}

func CacheMiddleware(cfg CacheMiddlewareConfig) commonapp.MiddlewareConfig {
    return commonapp.Middleware("Cache", cfg.Priority, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
        logger.Infow("Initializing cache middleware",
            "type", cfg.Type,
            "host", cfg.Host,
            "port", cfg.Port,
        )

        var init bootstrap.Initializer
        switch cfg.Type {
        case "redis":
            init = initializers.NewRedisInitializer(cfg.Host, cfg.Port, logger)
        case "memcached":
            init = initializers.NewMemcachedInitializer(cfg.Host, cfg.Port, logger)
        default:
            return fmt.Errorf("unsupported cache type: %s", cfg.Type)
        }

        bs.Register(init)
        return nil
    })
}

// Usage
app.WithMiddleware(CacheMiddleware(CacheMiddlewareConfig{
    Type:     "redis",
    Host:     "localhost",
    Port:     6379,
    Priority: 400,
}))
```

## Middleware Execution Flow

```
1. StandardBootstrapApplication.Initialize() called
   ↓
2. Create middlewareRegistrar (wraps ComponentRegistrar)
   ↓
3. BaseBootstrapApp.BaseInitialize() called
   ↓
4. Initialize logger
   ↓
5. Inject logger into middlewareRegistrar
   ↓
6. Call middlewareRegistrar.RegisterComponents()
   ↓
7. Sort middlewares by priority (ascending)
   ↓
8. Execute each middleware function in order
   │  - Each middleware can register initializers
   │  - Errors halt the process
   ↓
9. Call original ComponentRegistrar.RegisterComponents()
   ↓
10. Bootstrap initializes all registered components
```

## Priority Guidelines

Use these priority ranges for common middleware types:

| Priority | Purpose | Examples |
|----------|---------|----------|
| 50-150 | Infrastructure monitoring | Metrics, tracing, profiling |
| 200-400 | Core infrastructure | Cache, message queue, config center |
| 500-700 | Business components | Default priority, business initializers |
| 800-900 | Post-initialization | Health checks, admin interfaces |

## Best Practices

### 1. Use Descriptive Names

```go
// Good
commonapp.Middleware("MetricsCollection", 100, ...)

// Bad
commonapp.Middleware("m1", 100, ...)
```

### 2. Handle Errors Properly

```go
func MyMiddleware() commonapp.MiddlewareConfig {
    return commonapp.Middleware("MyMiddleware", 500, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
        init, err := createInitializer(opts)
        if err != nil {
            return fmt.Errorf("failed to create initializer: %w", err)
        }
        bs.Register(init)
        return nil
    })
}
```

### 3. Use Conditional Middleware for Optional Features

```go
// Enable feature only in specific environments
app.WithMiddleware(
    commonapp.ConditionalMiddleware(
        "DebugAPI",
        900,
        func(opts commonapp.Options) bool {
            return opts.GetEnv() != "production"
        },
        func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
            // Register debug API initializer
            return nil
        },
    ),
)
```

### 4. Choose Appropriate Priorities

```go
// Metrics should start early to capture all initialization metrics
app.WithMiddleware(commonapp.Middleware("Metrics", 100, ...))

// Database should initialize before services that depend on it
app.WithMiddleware(commonapp.Middleware("Database", 300, ...))

// Health checks should start last (after all services are ready)
app.WithMiddleware(commonapp.Middleware("Health", 900, ...))
```

### 5. Log Middleware Execution

```go
func MyMiddleware() commonapp.MiddlewareConfig {
    return commonapp.Middleware("MyFeature", 500, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
        logger.Infow("Initializing MyFeature middleware",
            "config_param", someValue,
            "enabled", true,
        )

        // Initialization logic...

        logger.Infow("MyFeature middleware initialized successfully")
        return nil
    })
}
```

## Examples from Real Services

### Agent Manager with Metrics Middleware

```go
func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    serverOpts := opts.(*options.ServerOptions)

    if a.StandardBootstrapApplication == nil {
        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Agent Manager", a).
            WithMiddleware(
                commonapp.MetricsMiddleware(9090, func(port int, logger core.Logger) bootstrap.Initializer {
                    return pkginitializers.NewMetricsInitializer(
                        fmt.Sprintf(":%d", port),
                        logger,
                    )
                }),
            )
    }

    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

### Orchestrator with Multiple Middlewares

```go
func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    if a.StandardBootstrapApplication == nil {
        middlewares := []commonapp.MiddlewareConfig{
            // Metrics collection
            commonapp.MetricsMiddleware(9090, newMetricsInit),

            // Distributed tracing
            commonapp.TracingMiddleware("orchestrator", "http://jaeger:14268/api/traces"),

            // Performance profiling (development only)
            commonapp.ConditionalMiddleware(
                "Profiling",
                200,
                func(opts commonapp.Options) bool {
                    return opts.GetEnv() == "development"
                },
                func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
                    // Enable pprof on port 6060
                    return nil
                },
            ),
        }

        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Orchestrator", a).
            WithMiddlewares(middlewares...)
    }

    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

## Debugging Middleware

### Enable Middleware Execution Logging

The framework automatically logs middleware execution:

```
2025-10-31T15:27:45.123Z INFO Applying middleware {"name": "Metrics", "priority": 100}
2025-10-31T15:27:45.134Z INFO Applying middleware {"name": "Tracing", "priority": 150}
2025-10-31T15:27:45.145Z INFO Applying middleware {"name": "Database", "priority": 300}
```

### Check Middleware Registration

Add debug logging in your middleware:

```go
func MyMiddleware() commonapp.MiddlewareConfig {
    return commonapp.Middleware("MyMiddleware", 500, func(bs *bootstrap.Bootstrap, logger core.Logger, opts commonapp.Options) error {
        logger.Debugw("MyMiddleware starting",
            "priority", 500,
            "options", opts,
        )

        // ... middleware logic ...

        logger.Debugw("MyMiddleware completed successfully")
        return nil
    })
}
```

## Migration Guide

### Migrating Existing Services

If you have a service using `BaseBootstrapApp` without middleware:

**Before:**
```go
type MyApp struct {
    *commonapp.BaseBootstrapApp
    // ...
}

func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    app := commonapp.NewBaseBootstrapApp("MyService", a)
    // ... manual initialization ...
}
```

**After:**
```go
type MyApp struct {
    *commonapp.StandardBootstrapApplication
    // ...
}

func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    if a.StandardBootstrapApplication == nil {
        a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("MyService", a).
            WithMiddleware(/* your middlewares */)
    }
    return a.StandardBootstrapApplication.Initialize(ctx, opts)
}
```

## Limitations

1. **No Inter-Middleware Communication**: Middlewares cannot pass data to each other directly. Use the Options object or shared state in the App structure.

2. **Execution Before Components**: Middlewares execute before the original `RegisterComponents()`. If you need post-registration logic, implement it in `OnStartup()` hook instead.

3. **Logger Dependency**: Middlewares require the logger to be initialized first. Do not attempt to initialize the logger within middleware.

4. **Error Handling**: Any middleware error stops the entire initialization process. Handle errors gracefully or use conditional middleware to skip optional features.

## Troubleshooting

### Middleware Not Executing

**Problem**: Middleware function not being called.

**Solution**: Ensure middleware is registered before calling `Initialize()`:

```go
if a.StandardBootstrapApplication == nil {
    a.StandardBootstrapApplication = commonapp.NewStandardBootstrapApplication("Service", a).
        WithMiddleware(myMiddleware)  // Must be before Initialize()
}
return a.StandardBootstrapApplication.Initialize(ctx, opts)
```

### Wrong Execution Order

**Problem**: Middleware executing in unexpected order.

**Solution**: Check middleware priorities. Lower numbers execute first:

```go
// This middleware executes FIRST (priority 100)
commonapp.Middleware("Early", 100, ...)

// This middleware executes SECOND (priority 500)
commonapp.Middleware("Late", 500, ...)
```

### Import Cycle Errors

**Problem**: Importing `pkg/initializers` in middleware causes import cycle.

**Solution**: Pass initializer factory functions as parameters instead:

```go
// Don't do this (causes import cycle):
func MyMiddleware() commonapp.MiddlewareConfig {
    return commonapp.Middleware("My", 500, func(...) error {
        init := pkginitializers.NewMyInitializer()  // ❌ Import cycle
    })
}

// Do this instead:
func MyMiddleware(newInit func() bootstrap.Initializer) commonapp.MiddlewareConfig {
    return commonapp.Middleware("My", 500, func(...) error {
        init := newInit()  // ✅ No import cycle
    })
}
```

## See Also

- [pkg/app/bootstrap_app.go](../pkg/app/bootstrap_app.go) - Core middleware implementation
- [pkg/app/middleware.go](../pkg/app/middleware.go) - Predefined middleware helpers
- [pkg/bootstrap/](../pkg/bootstrap/) - Bootstrap framework
- [cmd/cluster/app/app.go](../cmd/cluster/app/app.go) - Example service using StandardBootstrapApplication
