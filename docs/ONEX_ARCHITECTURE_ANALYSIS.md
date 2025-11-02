# OneX Architecture Analysis: Service Entry Points, Initialization Patterns, and Best Practices

This document provides a comprehensive analysis of the OneX codebase architecture, focusing on proven patterns that should be adopted in the Aetherius k8s-agent project.

## Executive Summary

OneX employs **two complementary service entry patterns** based on complexity:

1. **Complex Services** (APIServer, Controller Manager, Blockchain Controller, Job Controller): Use Cobra commands with explicit configuration and initialization
2. **Simpler Services** (Gateway, Cacheserver, etc.): Use lightweight App wrapper pattern with Viper configuration

Additionally, OneX uses a **sophisticated modular Makefile system** with clean separation of build concerns.

---

## 1. SERVICE ENTRY POINT PATTERNS

### Pattern 1: Kubernetes-Style Command Pattern (Complex Services)

**Used by**: onex-apiserver, onex-controller-manager, onex-blockchain-controller, onex-job-controller

**Characteristics**:
- Direct Cobra command definition in `cmd/<service>/app/`
- Explicit RunE function with full error handling
- Configuration options struct with Complete/Validate methods
- Flags registration with k8s component-base style
- Graceful shutdown with signal context

**File Structure**:
```
cmd/onex-apiserver/
├── apiserver.go          # main.go - entry point
├── app/
│   ├── server.go         # NewAPIServerCommand(), Run(), CreateServerChain()
│   ├── options/
│   │   ├── options.go    # ServerRunOptions struct, Complete(), Validate()
│   │   ├── globalflags.go
│   │   ├── validation.go
│   │   └── completion.go
│   ├── config.go         # NewConfig(), Complete()
│   ├── aggregator.go
│   └── helper.go
```

**Key Files**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-apiserver/apiserver.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-apiserver/app/server.go`

**Code Example**:

```go
// cmd/onex-apiserver/apiserver.go
func main() {
    var informerFactory informers.SharedInformerFactory

    command := app.NewAPIServerCommand(
        app.WithEtcdOptions("/registry/onex.io", appsv1beta1.SchemeGroupVersion, batchv1beta1.SchemeGroupVersion),
        app.WithRESTStorageProviders(appsrest.RESTStorageProvider{}, batchrest.RESTStorageProvider{}),
        app.WithAlternateDNS("onex.io"),
        app.WithAdmissionPlugin(minerset.PluginName, minerset.Register),
        app.WithPostStartHook("start-external-informers", func(ctx genericapiserver.PostStartHookContext) error {
            if informerFactory != nil {
                informerFactory.Start(ctx.Done())
            }
            return nil
        }),
    )

    code := cli.Run(command)
    os.Exit(code)
}

// cmd/onex-apiserver/app/server.go
func NewAPIServerCommand(serverRunOptions ...Option) *cobra.Command {
    s := options.NewServerRunOptions()
    for _, opt := range serverRunOptions {
        opt(s)
    }

    cmd := &cobra.Command{
        Use:   "onex-apiserver",
        Short: "Launch a onex API server",
        RunE: func(cmd *cobra.Command, args []string) error {
            version.PrintAndExitIfRequested()
            fs := cmd.Flags()

            if err := logsapi.ValidateAndApply(s.Logs, utilfeature.DefaultFeatureGate); err != nil {
                return err
            }
            cliflag.PrintFlags(fs)

            completedOptions, err := s.Complete()
            if err != nil {
                return err
            }

            if errs := completedOptions.Validate(); len(errs) != 0 {
                return utilerrors.NewAggregate(errs)
            }

            return Run(cmd.Context(), completedOptions)
        },
    }

    fs := cmd.Flags()
    namedFlagSets := s.Flags()
    version.AddFlags(namedFlagSets.FlagSet("global"))
    globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name(), logs.SkipLoggingConfigurationFlags())
    
    for _, f := range namedFlagSets.FlagSets {
        fs.AddFlagSet(f)
    }

    return cmd
}

func Run(ctx context.Context, opts options.CompletedOptions) error {
    klog.Infof("Version: %+v", version.Get().String())
    
    config, err := NewConfig(opts)
    if err != nil {
        return err
    }
    completed, err := config.Complete()
    if err != nil {
        return err
    }
    server, err := CreateServerChain(completed)
    if err != nil {
        return err
    }

    prepared, err := server.PrepareRun()
    if err != nil {
        return err
    }

    return prepared.Run(ctx)
}
```

**Option Pattern** (Functional Options):

```go
// Type definitions for options
type (
    Option       func(*options.ServerRunOptions)
    RegisterFunc func(plugins *admission.Plugins)
)

// Functional option implementations
func WithEtcdOptions(prefix string, versions ...schema.GroupVersion) Option {
    return func(s *options.ServerRunOptions) {
        codec := legacyscheme.Codecs.LegacyCodec(versions...)
        s.RecommendedOptions.Etcd = genericoptions.NewEtcdOptions(
            storagebackend.NewDefaultConfig(prefix, codec),
        )
    }
}

func WithAdmissionPlugin(name string, registerFunc RegisterFunc) Option {
    return func(s *options.ServerRunOptions) {
        s.RecommendedOptions.Admission.RecommendedPluginOrder = 
            append(s.RecommendedOptions.Admission.RecommendedPluginOrder, name)
        registerFunc(s.RecommendedOptions.Admission.Plugins)
    }
}

func WithPostStartHook(name string, hook genericapiserver.PostStartHookFunc) Option {
    return func(s *options.ServerRunOptions) {
        s.ExternalPostStartHooks[name] = hook
    }
}
```

---

### Pattern 2: Lightweight App Wrapper Pattern (Simpler Services)

**Used by**: onex-gateway, onex-cacheserver, onex-nightwatch

**Characteristics**:
- Minimal main.go (just calls app.NewApp().Run())
- Uses `github.com/onexstack/onexstack/pkg/app.App` wrapper
- Options implement NamedFlagSetOptions interface
- Configuration through Viper with environment variable support
- Implements Complete() and Validate() methods

**File Structure**:
```
cmd/onex-gateway/
├── main.go                    # Minimal entry point
└── app/
    ├── server.go              # NewApp(), run() function
    ├── options/
    │   └── options.go         # ServerOptions with Complete/Validate
    └── ...
```

**Key Files**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-gateway/main.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-gateway/app/server.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-gateway/app/options/options.go`

**Code Example**:

```go
// cmd/onex-gateway/main.go - Minimal entry point
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/onexstack/onex/cmd/onex-gateway/app"
)

func main() {
    app.NewApp().Run()
}

// cmd/onex-gateway/app/server.go
const commandDesc = `The gateway server is the back-end portal server of onex. 
All requests from the front-end will arrive at the gateway, requests will be 
uniformly processed and distributed by the gateway.`

func NewApp() *app.App {
    opts := options.NewServerOptions()
    application := app.NewApp(
        gateway.Name,
        "Launch a onex gateway server",
        app.WithDescription(commandDesc),
        app.WithOptions(opts),
        app.WithDefaultValidArgs(),
        app.WithRunFunc(run(opts)),
        app.WithLoggerContextExtractor(map[string]func(context.Context) string{
            known.XTraceID: contextx.TraceID,
            known.XUserID:  contextx.UserID,
        }),
    )
    return application
}

func run(opts *options.ServerOptions) app.RunFunc {
    return func() error {
        cfg, err := opts.Config()
        if err != nil {
            return fmt.Errorf("failed to load configuration: %w", err)
        }

        ctx := genericapiserver.SetupSignalContext()

        server, err := cfg.NewServer(ctx)
        if err != nil {
            return fmt.Errorf("failed to create server: %w", err)
        }

        return server.Run(ctx)
    }
}

// cmd/onex-gateway/app/options/options.go
type ServerOptions struct {
    GRPCOptions       *genericoptions.GRPCOptions
    HTTPOptions       *genericoptions.HTTPOptions
    TLSOptions        *genericoptions.TLSOptions
    MySQLOptions      *genericoptions.MySQLOptions
    RedisOptions      *genericoptions.RedisOptions
    EtcdOptions       *genericoptions.EtcdOptions
    JaegerOptions     *genericoptions.JaegerOptions
    ConsulOptions     *genericoptions.ConsulOptions
    UserCenterOptions *usercenter.UserCenterOptions
    MetricsOptions    *genericoptions.MetricsOptions
    EnableTLS         bool
    Kubeconfig        string
    FeatureGates      map[string]bool
    Log               *log.Options
}

// Ensure ServerOptions implements the app.NamedFlagSetOptions interface
var _ app.NamedFlagSetOptions = (*ServerOptions)(nil)

func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        GRPCOptions:       genericoptions.NewGRPCOptions(),
        HTTPOptions:       genericoptions.NewHTTPOptions(),
        TLSOptions:        genericoptions.NewTLSOptions(),
        MySQLOptions:      genericoptions.NewMySQLOptions(),
        RedisOptions:      genericoptions.NewRedisOptions(),
        EtcdOptions:       genericoptions.NewEtcdOptions(),
        JaegerOptions:     genericoptions.NewJaegerOptions(),
        ConsulOptions:     genericoptions.NewConsulOptions(),
        UserCenterOptions: usercenter.NewUserCenterOptions(),
        MetricsOptions:    genericoptions.NewMetricsOptions(),
        Log:               log.NewOptions(),
    }
}

func (o *ServerOptions) Flags() (fss cliflag.NamedFlagSets) {
    o.GRPCOptions.AddFlags(fss.FlagSet("grpc"))
    o.HTTPOptions.AddFlags(fss.FlagSet("http"))
    o.TLSOptions.AddFlags(fss.FlagSet("tls"))
    o.MySQLOptions.AddFlags(fss.FlagSet("mysql"))
    // ... more flags
    return fss
}

func (o *ServerOptions) Complete() error {
    if o.JaegerOptions.ServiceName == "" {
        o.JaegerOptions.ServiceName = UserAgent
    }
    _ = feature.DefaultMutableFeatureGate.SetFromMap(o.FeatureGates)
    return nil
}

func (o *ServerOptions) Validate() error {
    errs := []error{}
    errs = append(errs, o.GRPCOptions.Validate()...)
    errs = append(errs, o.HTTPOptions.Validate()...)
    // ... more validations
    return utilerrors.NewAggregate(errs)
}

func (o *ServerOptions) Config() (*gateway.Config, error) {
    kubeconfig, err := clientcmd.BuildConfigFromFlags("", o.Kubeconfig)
    if err != nil {
        return nil, err
    }
    
    cfg := &gateway.Config{
        GRPCOptions:       o.GRPCOptions,
        HTTPOptions:       o.HTTPOptions,
        // ... other options
    }
    return cfg, nil
}
```

**App Framework** (`github.com/onexstack/onexstack/pkg/app.App`):

```go
// staging/src/github.com/onexstack/onexstack/pkg/app/app.go
type App struct {
    name        string
    shortDesc   string
    description string
    run         RunFunc
    cmd         *cobra.Command
    args        cobra.PositionalArgs
    
    healthCheckFunc HealthCheckFunc
    options         any
    silence         bool
    noConfig        bool
    watch           bool
    contextExtractors map[string]func(context.Context) string
}

type RunFunc func() error
type HealthCheckFunc func() error
type Option func(*App)

func NewApp(name string, shortDesc string, opts ...Option) *App {
    app := &App{
        name:      name,
        run:       func() error { return nil },
        shortDesc: shortDesc,
    }

    for _, o := range opts {
        o(app)
    }

    app.buildCommand()
    return app
}

// Functional options
func WithOptions(opts any) Option {
    return func(app *App) {
        app.options = opts
    }
}

func WithRunFunc(run RunFunc) Option {
    return func(app *App) {
        app.run = run
    }
}

func WithDescription(desc string) Option {
    return func(app *App) {
        app.description = desc
    }
}

func (app *App) Run() {
    os.Exit(cli.Run(app.cmd))
}

func (app *App) runCommand(cmd *cobra.Command, args []string) error {
    version.PrintAndExitIfRequested()

    if err := viper.BindPFlags(cmd.Flags()); err != nil {
        return err
    }

    if app.options != nil {
        if err := viper.Unmarshal(app.options); err != nil {
            return err
        }

        if complete, ok := app.options.(interface{ Complete() error }); ok {
            if err := complete.Complete(); err != nil {
                return err
            }
        }

        if validate, ok := app.options.(interface{ Validate() error }); ok {
            if err := validate.Validate(); err != nil {
                return err
            }
        }
    }

    app.initializeLogger()

    if !app.silence {
        log.Infow("Starting application", "name", app.name, "version", version.Get().ToJSON())
    }

    if app.healthCheckFunc != nil {
        if err := app.healthCheckFunc(); err != nil {
            return err
        }
    }

    return app.run()
}
```

---

## 2. DEPENDENCY INJECTION & INITIALIZATION PATTERNS

### Pattern: Google Wire for Dependency Injection

**Location**: `cmd/onex-controller-manager/app/wire.go`

**Characteristics**:
- Uses Google Wire code generation tool
- ProviderSet pattern for grouping related providers
- Wire.Build() to declare dependencies
- Automatic initialization order

**Code Example**:

```go
//go:build wireinject
// +build wireinject

// cmd/onex-controller-manager/app/wire.go
package app

import (
    "github.com/google/wire"
    "github.com/onexstack/onex/internal/gateway/store"
    "github.com/onexstack/onexstack/pkg/db"
)

func wireStoreClient(*db.MySQLOptions) (store.IStore, error) {
    wire.Build(
        db.ProviderSet,
        store.ProviderSet,
    )

    return nil, nil
}

// Wire generation: go run github.com/google/wire/cmd/wire
// Outputs: cmd/onex-controller-manager/app/wire_gen.go
```

**How It Works**:
1. Define ProviderSets in each domain package
2. Declare wire.Build() with all ProviderSets
3. Run `go generate` to auto-generate `wire_gen.go`
4. Call generated initialization function

**Benefits**:
- Compile-time dependency injection (no reflection overhead)
- Clear dependency graph
- Type-safe initialization
- Easy to trace dependencies

---

## 3. ERROR HANDLING PATTERNS

### Pattern: ErrorX with HTTP Status Codes and Metadata

**Location**: `/staging/src/github.com/onexstack/onexstack/pkg/errorsx/errorsx.go`

**Key File**: `/Users/costalong/code/go/src/github.com/onexstack/onex/staging/src/github.com/onexstack/onexstack/pkg/errorsx/errorsx.go`

**Characteristics**:
- HTTP status code integration (Code field)
- Business error reason (Reason field)
- User-friendly message (Message field)
- Metadata for additional context
- gRPC status integration
- Error chain support with errors.Is()

**Code Example**:

```go
// ErrorX structure
type ErrorX struct {
    Code     int               `json:"code,omitempty"`      // HTTP status code
    Reason   string            `json:"reason,omitempty"`    // Business error code
    Message  string            `json:"message,omitempty"`   // User-friendly message
    Metadata map[string]string `json:"metadata,omitempty"`  // Additional context
}

// Creating errors
func New(code int, reason string, format string, args ...any) *ErrorX {
    return &ErrorX{
        Code:    code,
        Reason:  reason,
        Message: fmt.Sprintf(format, args...),
    }
}

// Chaining methods
func (err *ErrorX) WithMessage(format string, args ...any) *ErrorX {
    err.Message = fmt.Sprintf(format, args...)
    return err
}

func (err *ErrorX) WithMetadata(md map[string]string) *ErrorX {
    err.Metadata = md
    return err
}

func (err *ErrorX) KV(kvs ...string) *ErrorX {
    if err.Metadata == nil {
        err.Metadata = make(map[string]string)
    }
    for i := 0; i < len(kvs); i += 2 {
        if i+1 < len(kvs) {
            err.Metadata[kvs[i]] = kvs[i+1]
        }
    }
    return err
}

func (err *ErrorX) WithRequestID(requestID string) *ErrorX {
    return err.KV("X-Request-ID", requestID)
}

// gRPC integration
func (err *ErrorX) GRPCStatus() *status.Status {
    details := errdetails.ErrorInfo{Reason: err.Reason, Metadata: err.Metadata}
    s, _ := status.New(httpstatus.ToGRPCCode(err.Code), err.Message).WithDetails(&details)
    return s
}

// Error matching
func (err *ErrorX) Is(target error) bool {
    if errx := new(ErrorX); errors.As(target, &errx) {
        return errx.Code == err.Code && errx.Reason == err.Reason
    }
    return false
}

// Extracting error info
func Code(err error) int {
    if err == nil {
        return http.StatusOK
    }
    return FromError(err).Code
}

func Reason(err error) string {
    if err == nil {
        return ErrInternal.Reason
    }
    return FromError(err).Reason
}

// Converting any error to ErrorX
func FromError(err error) *ErrorX {
    if err == nil {
        return nil
    }

    if errx := new(ErrorX); errors.As(err, &errx) {
        return errx
    }

    gs, ok := status.FromError(err)
    if !ok {
        return New(ErrInternal.Code, ErrInternal.Reason, err.Error())
    }

    ret := New(httpstatus.FromGRPCCode(gs.Code()), ErrInternal.Reason, gs.Message())

    for _, detail := range gs.Details() {
        if typed, ok := detail.(*errdetails.ErrorInfo); ok {
            ret.Reason = typed.Reason
            return ret.WithMetadata(typed.Metadata)
        }
    }

    return ret
}

// Usage example
var (
    ErrInternal    = New(500, "InternalError", "internal server error")
    ErrNotFound    = New(404, "NotFound", "resource not found")
    ErrUnauthorized = New(401, "Unauthorized", "unauthorized")
)

// In handlers
if resource == nil {
    return ErrNotFound.
        WithMessage("user %d not found", userID).
        WithRequestID(requestID)
}
```

---

## 4. CONTEXT MANAGEMENT

### Pattern: Type-Safe Context Values

**Location**: `internal/pkg/contextx/contextx.go`

**Key File**: `/Users/costalong/code/go/src/github.com/onexstack/onex/internal/pkg/contextx/contextx.go`

**Characteristics**:
- Unexported struct types as context keys (prevents collisions)
- Type-safe getters and setters
- JWT claims extraction
- User ID and access token management
- Trace ID propagation

**Code Example**:

```go
// Unexported types as context keys (thread-safe, collision-proof)
type (
    transCtx     struct{}
    noTransCtx   struct{}
    transLockCtx struct{}
    userIDCtx    struct{}
    traceIDCtx   struct{}
)

type (
    claimsKey      struct{}
    userKey        struct{}
    userMKey       struct{}
    accessTokenKey struct{}
    traceIDKey     struct{}
)

// Type-safe setters and getters
func WithClaims(ctx context.Context, claims *jwt.RegisteredClaims) context.Context {
    return context.WithValue(ctx, claimsKey{}, claims)
}

func Claims(ctx context.Context) *jwt.RegisteredClaims {
    claims, _ := ctx.Value(claimsKey{}).(*jwt.RegisteredClaims)
    return claims
}

func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userKey{}, userID)
}

func UserID(ctx context.Context) string {
    userID, _ := ctx.Value(userKey{}).(string)
    return userID
}

func WithAccessToken(ctx context.Context, accessToken string) context.Context {
    return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

func AccessToken(ctx context.Context) string {
    accessToken, _ := ctx.Value(accessTokenKey{}).(string)
    return accessToken
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceID(ctx context.Context) string {
    traceID, _ := ctx.Value(traceIDKey{}).(string)
    return traceID
}
```

---

## 5. MIDDLEWARE PATTERNS

### Pattern: Gin Middleware Functions

**Location**: `internal/pkg/middleware/gin/`

**Key File**: `/Users/costalong/code/go/src/github.com/onexstack/onex/internal/pkg/middleware/gin/traceid.go`

**Example: TraceID Middleware**:

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/onexstack/onex/internal/pkg/contextx"
    known "github.com/onexstack/onex/internal/pkg/known/toyblc"
)

// TraceID is a Gin middleware that injects Trace-ID into request/response
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if Trace-ID exists in request headers
        traceID := c.Request.Header.Get(known.TraceIDKey)

        if traceID == "" {
            traceID = uuid.New().String()
            c.Request.Header.Set(known.TraceIDKey, traceID)
        }

        // Set Trace-ID in response headers
        c.Writer.Header().Set(known.TraceIDKey, traceID)

        // Store in Gin context
        c.Set("trace.id", traceID)

        // Store in request context for downstream use
        ctx := contextx.WithTraceID(c.Request.Context(), traceID)
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

**Middleware Registration Pattern**:

```go
// In server setup
router := gin.New()
router.Use(middleware.TraceID())
router.Use(middleware.Authentication())
router.Use(middleware.Authorization())
router.Use(middleware.Logging())
```

---

## 6. MODULAR MAKEFILE SYSTEM

### Architecture: Script-based Makefile Modules

**Location**: `scripts/make-rules/`

**Key Files**:
- `scripts/make-rules/common.mk` - Shared variables and functions
- `scripts/make-rules/golang.mk` - Go build/test targets
- `scripts/make-rules/image.mk` - Docker build targets
- `scripts/make-rules/tools.mk` - Development tool management
- `scripts/make-rules/generate.mk` - Code generation

**Pattern**:
1. Root Makefile includes modular make-rules files
2. Each .mk file focuses on one concern
3. Common variables defined in common.mk
4. Make targets use modular naming: `<module>.<action>[.<service>]`

**Root Makefile Structure**:

```makefile
.DEFAULT_GOAL := help

.PHONY: all
all: format tidy gen add-copyright lint cover build

# Include modular rules
include scripts/make-rules/common.mk
include scripts/make-rules/all.mk

# Build targets
.PHONY: build
build: tidy
    $(MAKE) go.build

.PHONY: build.multiarch
build.multiarch:
    $(MAKE) go.build.multiarch

# Docker targets
.PHONY: image
image:
    $(MAKE) image.build

.PHONY: image.multiarch
image.multiarch:
    $(MAKE) image.build.multiarch
```

**Benefits**:
- Modular organization (easier to maintain)
- Consistent naming across projects
- Reusable rules across services
- Clear separation of concerns

---

## 7. CONTROLLER RUNTIME INTEGRATION

### Pattern: Kubebuilder-style Manager Setup

**Location**: `cmd/onex-controller-manager/app/controllermanager.go`

**Characteristics**:
- Uses `sigs.k8s.io/controller-runtime`
- Manager initialization with cache, client, webhook configuration
- Health check setup (liveness and readiness probes)
- Metrics collection via Prometheus
- Leader election configuration
- Namespace filtering for multi-tenancy

**Code Example**:

```go
import (
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/cache"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller"
    "sigs.k8s.io/controller-runtime/pkg/healthz"
    ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
)

func Run(ctx context.Context, c *config.CompletedConfig) error {
    // Store controller configs
    cfgz, err := configz.New(ConfigzName)
    if err != nil {
        return err
    }
    cfgz.Set(c.ComponentConfig)

    // Configure namespace watching
    var watchNamespaces map[string]cache.Config
    if c.ComponentConfig.Generic.Namespace != "" {
        watchNamespaces = map[string]cache.Config{
            c.ComponentConfig.Generic.Namespace: {},
        }
    }

    // Create manager
    mgr, err := ctrl.NewManager(c.Kubeconfig, ctrl.Options{
        Scheme:                 scheme,
        LeaderElection:         c.ComponentConfig.Generic.LeaderElection.LeaderElect,
        LeaderElectionID:       c.ComponentConfig.Generic.LeaderElection.ResourceName,
        LeaseDuration:          &c.ComponentConfig.Generic.LeaderElection.LeaseDuration.Duration,
        RenewDeadline:          &c.ComponentConfig.Generic.LeaderElection.RenewDeadline.Duration,
        RetryPeriod:            &c.ComponentConfig.Generic.LeaderElection.RetryPeriod.Duration,
        LeaderElectionResourceLock: c.ComponentConfig.Generic.LeaderElection.ResourceLock,
        LeaderElectionNamespace:    c.ComponentConfig.Generic.LeaderElection.ResourceNamespace,
        HealthProbeBindAddress:     c.ComponentConfig.Generic.HealthzBindAddress,
        PprofBindAddress:           c.ComponentConfig.Generic.PprofBindAddress,
        Cache: cache.Options{
            DefaultNamespaces: watchNamespaces,
            SyncPeriod:        &c.ComponentConfig.Generic.SyncPeriod.Duration,
            ByObject: map[client.Object]cache.ByObject{
                &corev1.ConfigMap{}: {Label: chainSecretCacheSelector},
                &corev1.Secret{}:    {Label: chainSecretCacheSelector},
            },
        },
        Client: client.Options{
            Cache: &client.CacheOptions{
                DisableFor: []client.Object{
                    &corev1.ConfigMap{},
                    &corev1.Secret{},
                },
            },
        },
        WebhookServer: webhook.NewServer(webhook.Options{}),
    })
    if err != nil {
        return err
    }

    // Setup health checks
    if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
        klog.ErrorS(err, "Unable to set up health check")
        return err
    }
    if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
        klog.ErrorS(err, "Unable to set up ready check")
        return err
    }

    // Register metrics
    machineMetricsCollector := metrics.NewMinerCollector(mgr.GetClient(), c.ComponentConfig.Generic.Namespace)
    ctrlmetrics.Registry.MustRegister(machineMetricsCollector)

    // Initialize event recorder
    record.InitFromRecorder(mgr.GetEventRecorderFor("onex-controller-manager"))

    // Add controllers
    addControllers(ctx, cctx, mgr, NewControllerDescriptors())

    // Start informers
    cctx.InformerFactory.Start(ctx.Done())
    cctx.ObjectOrMetadataInformerFactory.Start(ctx.Done())
    close(cctx.InformersStarted)

    return mgr.Start(ctx)
}
```

---

## 8. RECOMMENDED PATTERNS FOR AETHERIUS (k8s-agent)

### Pattern Selection by Service Complexity

Based on the Aetherius architecture in CLAUDE.md:

**Use Complex Pattern (Cobra + Options)** for:
- agent-manager (✓ already uses)
- orchestrator (✓ already uses)
- auth (✓ already uses)
- cluster (✓ already uses)
- reasoning (✓ already uses)

**Use Simple Pattern (App wrapper)** for:
- collect-agent (lightweight edge agent)
- gateway
- monitor

### Implementation Checklist

#### For Bootstrap-Pattern Services:

1. **Create cmd/<service>/app/options/ package**
   - ServerOptions struct with all configuration fields
   - Implement NamedFlagSetOptions interface (or similar)
   - Implement Complete() method
   - Implement Validate() method

2. **Create cmd/<service>/app/server.go**
   - NewCommand() or NewApp() function
   - Options setup with functional options
   - RunE function with proper error handling
   - Flag registration

3. **Initialize in cmd/<service>/(service).go or main.go**
   - Use `cli.Run()` for Cobra commands
   - Or use app.NewApp().Run() for simple pattern

#### For Simple-Pattern Services:

1. **Minimal main.go**
   ```go
   func main() {
       app.NewApp().Run()
   }
   ```

2. **Create app/server.go**
   ```go
   func NewApp() *app.App {
       opts := options.NewServerOptions()
       return app.NewApp(
           serviceName,
           "Short description",
           app.WithOptions(opts),
           app.WithRunFunc(run(opts)),
           // ... other options
       )
   }
   
   func run(opts *options.ServerOptions) app.RunFunc {
       return func() error {
           cfg, err := opts.Config()
           if err != nil {
               return err
           }
           ctx := genericapiserver.SetupSignalContext()
           server, err := cfg.NewServer(ctx)
           if err != nil {
               return err
           }
           return server.Run(ctx)
       }
   }
   ```

3. **Create app/options/options.go**
   - Implement NamedFlagSetOptions interface
   - Implement Complete(), Validate()
   - Implement Config() method that returns service-specific config

---

## 9. KEY TAKEAWAYS & RECOMMENDATIONS

### What Aetherius Should Adopt

1. **Service Entry Patterns**
   - Already following Bootstrap pattern correctly ✓
   - Continue using pkg/app.RunWithRunner() ✓
   - Ensure consistent Complete() and Validate() implementations

2. **Error Handling**
   - Adopt ErrorX pattern from OneX
   - Replace any ad-hoc error handling
   - Support error metadata and request ID tracking

3. **Context Management**
   - Use unexported struct types for context keys
   - Provide typed getters/setters
   - Propagate trace IDs through request context

4. **Middleware**
   - Standardize on Gin middleware functions
   - Implement TraceID middleware for all services
   - Support context extraction for structured logging

5. **Dependency Injection**
   - Consider Google Wire for complex services
   - Ensure clear dependency declaration
   - Document provider sets and initialization order

6. **Build System**
   - Enhance existing Makefile with modular rules
   - Follow OneX naming conventions: `<module>.<action>[.<service>]`
   - Organize make-rules into separate .mk files

7. **Configuration Management**
   - Use Viper for configuration loading
   - Support environment variable overrides
   - Implement Complete() and Validate() for options

8. **Controller Patterns**
   - Use sigs.k8s.io/controller-runtime for controllers
   - Implement health checks
   - Register metrics collectors
   - Support leader election

---

## 10. CRITICAL FILES REFERENCE

**OneX Service Entry Patterns**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-apiserver/apiserver.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-apiserver/app/server.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-gateway/main.go`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-gateway/app/server.go`

**Error Handling**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/staging/src/github.com/onexstack/onexstack/pkg/errorsx/errorsx.go`

**Dependency Injection**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/cmd/onex-controller-manager/app/wire.go`

**Context Management**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/internal/pkg/contextx/contextx.go`

**Middleware**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/internal/pkg/middleware/gin/traceid.go`

**App Framework**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/staging/src/github.com/onexstack/onexstack/pkg/app/app.go`

**Makefile System**:
- `/Users/costalong/code/go/src/github.com/onexstack/onex/Makefile`
- `/Users/costalong/code/go/src/github.com/onexstack/onex/scripts/make-rules/`

