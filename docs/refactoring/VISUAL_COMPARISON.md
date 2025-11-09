# Startup Flow Visualization

## Before: Complex Multi-File Flow

```
┌────────────────────────────────────────────────────────────────────────────┐
│                          ENTRY POINT: cmd/agent-manager/app/app.go         │
│                                                                            │
│  func Execute() {                                                          │
│      opts := commonapp.NewStandardOptions(...)                             │
│      app := &AgentManagerApp{}                                             │
│      commonapp.RunWithBootstrap(app, opts, ..., app.registerComponents)   │
│  }                                                                          │
│                                                                            │
│  func (a *AgentManagerApp) registerComponents(bs *bootstrap.Bootstrap) {  │
│      container, err := InitializeAgentManagerContainer(a.opts) ────┐      │
│      for _, init := range container.GetInitializers() {             │      │
│          bs.Register(init)                                          │      │
│      }                                                               │      │
│  }                                                                   │      │
└──────────────────────────────────────────────────────────────────────┼─────┘
                                                                       │
                                                                       ↓
┌────────────────────────────────────────────────────────────────────────────┐
│                      WIRE CONFIG: cmd/agent-manager/app/wire.go            │
│                                                                            │
│  var InitializerSet = wire.NewSet(                                         │
│      ProvideLogger,                                                        │
│      initializers.NewDatabaseInitializer,           ← Wrapper 1            │
│      initializers.NewRedisInitializer,              ← Wrapper 2            │
│      initializers.NewServiceInitializer,            ← Wrapper 3            │
│      initializers.NewRegistryInitializer,           ← Wrapper 4            │
│      initializers.NewNATSInitializer,               ← Wrapper 5            │
│      initializers.NewDispatcherInitializer,         ← Wrapper 6            │
│      initializers.NewHTTPServerInitializer,         ← Wrapper 7            │
│      initializers.NewGRPCServerInitializer,         ← Wrapper 8            │
│  )                                                                          │
│                                                                            │
│  func InitializeAgentManagerContainer(opts) (*AgentManagerContainer, error)│
│      wire.Build(InitializerSet, HealthInitializerSet, ...)                │
│  }                                                                          │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓ wire generate
┌────────────────────────────────────────────────────────────────────────────┐
│                   GENERATED: cmd/agent-manager/app/wire_gen.go             │
│                                                                            │
│  func InitializeAgentManagerContainer(opts) (*AgentManagerContainer, error)│
│      logger := ProvideLogger(opts)                                         │
│      dbInit := initializers.NewDatabaseInitializer(opts, logger)           │
│      redisInit := initializers.NewRedisInitializer(opts, logger)           │
│      serviceInit := initializers.NewServiceInitializer(...)                │
│      registryInit := initializers.NewRegistryInitializer(...)              │
│      natsInit := initializers.NewNATSInitializer(...)                      │
│      dispatcherInit := initializers.NewDispatcherInitializer(...)          │
│      httpInit := initializers.NewHTTPServerInitializer(...)                │
│      grpcInit := initializers.NewGRPCServerInitializer(...)                │
│      healthInit := NewHealthCheckInitializer(...)                          │
│      return NewAgentManagerContainer(db, redis, ..., health)               │
│  }                                                                          │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│                   CONTAINER: cmd/agent-manager/app/container.go            │
│                                                                            │
│  type AgentManagerContainer struct {                                       │
│      db         *initializers.DatabaseInitializer                          │
│      redis      *initializers.RedisInitializer                             │
│      service    *initializers.ServiceInitializer                           │
│      registry   *initializers.RegistryInitializer                          │
│      nats       *initializers.NATSInitializer                              │
│      dispatcher *initializers.DispatcherInitializer                        │
│      http       *initializers.HTTPServerInitializer                        │
│      grpc       *initializers.GRPCServerInitializer                        │
│      health     *HealthCheckInitializer                                    │
│  }                                                                          │
│                                                                            │
│  func (c *AgentManagerContainer) GetInitializers() []Initializer {        │
│      return []Initializer{c.db, c.redis, ..., c.health}                   │
│  }                                                                          │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│           WRAPPERS: internal/agent-manager/initializers/*.go               │
│                                                                            │
│  database.go (50 LOC)                                                      │
│  ├── type DatabaseInitializer struct {                                     │
│  │       *pkg.DatabaseInitializer ← Wraps generic initializer             │
│  │       store *storage.MySQLStore                                        │
│  │   }                                                                     │
│  └── func (d *DatabaseInitializer) Store() *storage.MySQLStore {          │
│          return d.store                                                    │
│      }                                                                      │
│                                                                            │
│  redis.go (35 LOC)                                                         │
│  ├── type RedisInitializer struct {                                        │
│  │       *pkg.RedisInitializer ← Wraps generic initializer                │
│  │       store *storage.RedisStore                                        │
│  │   }                                                                     │
│  └── Similar wrapper pattern                                               │
│                                                                            │
│  servers.go (262 LOC)                                                      │
│  ├── type HTTPServerInitializer struct {                                   │
│  │       standardInit *pkg.HTTPServerInitializer                           │
│  │       serviceInit  *ServiceInitializer                                 │
│  │       dbInit       *DatabaseInitializer                                │
│  │       redisInit    *RedisInitializer                                   │
│  │   }                                                                     │
│  └── Delegates to pkg.HTTPServerInitializer                                │
│                                                                            │
│  business_services.go (169 LOC)                                            │
│  ├── type ServiceInitializer struct {                                      │
│  │       registry       *agent.Registry                                   │
│  │       eventProcessor *event.Processor                                  │
│  │       dispatcher     *command.Dispatcher                               │
│  │   }                                                                     │
│  └── Creates business services                                             │
│                                                                            │
│  service_facades.go (217 LOC)                                              │
│  ├── type RegistryInitializer    - Delegates to ServiceInitializer        │
│  ├── type NATSInitializer         - Delegates to ServiceInitializer       │
│  └── type DispatcherInitializer   - Delegates to ServiceInitializer       │
│                                                                            │
│  TOTAL: 733 LOC of wrapper boilerplate                                    │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│                  GENERIC: pkg/initializers/*.go (unchanged)                │
│                                                                            │
│  database.go (150 LOC)  - Actual MySQL connection logic                   │
│  redis.go (120 LOC)     - Actual Redis connection logic                   │
│  http_server.go (200)   - Actual HTTP server creation logic               │
│  grpc_server.go (180)   - Actual gRPC server creation logic               │
│  health.go (100 LOC)    - Actual health check logic                       │
└────────────────────────────────────────────────────────────────────────────┘

TOTAL COMPLEXITY:
  - 9 files in cmd/agent-manager/app/
  - 5 files in internal/agent-manager/initializers/
  - 18 files to understand complete flow
  - 7 abstraction layers
  - 5 concepts (Wire, Bootstrap, Container, Wrapper, Generic)
```

## After: Simple Single-File Flow

```
┌────────────────────────────────────────────────────────────────────────────┐
│                     EVERYTHING: cmd/agent-manager/app/app.go (505 LOC)    │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │ 1. ENTRY POINT                                                       │ │
│  │                                                                      │ │
│  │ func Execute() {                                                     │ │
│  │     opts := commonapp.NewStandardOptions(...)                        │ │
│  │     app := &AgentManagerApp{}                                        │ │
│  │     commonapp.RunWithBootstrap(app, opts, ..., registerComponents)  │ │
│  │ }                                                                     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │ 2. APP STRUCT (holds all services directly, no container)           │ │
│  │                                                                      │ │
│  │ type AgentManagerApp struct {                                        │ │
│  │     bootstrap *bootstrap.Bootstrap                                   │ │
│  │     opts      *commonapp.StandardOptions                             │ │
│  │     logger    core.Logger                                            │ │
│  │                                                                      │ │
│  │     // Business services (created once, shared)                     │ │
│  │     registry       *agent.Registry                                   │ │
│  │     eventProcessor *event.Processor                                  │ │
│  │     dispatcher     *command.Dispatcher                               │ │
│  │     natsServer     *nats.Server                                      │ │
│  │                                                                      │ │
│  │     // Storage instances                                            │ │
│  │     mysqlStore *storage.MySQLStore                                   │ │
│  │     redisStore *storage.RedisStore                                   │ │
│  │ }                                                                     │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │ 3. COMPONENT REGISTRATION (explicit, linear, clear)                 │ │
│  │                                                                      │ │
│  │ func (a *AgentManagerApp) registerComponents(bs *Bootstrap) error { │ │
│  │                                                                      │ │
│  │     // Step 1: Database (Priority 300)                              │ │
│  │     dbInit := pkginitializers.NewDatabaseInitializer(...)           │ │
│  │     if a.opts.Database.AutoMigrate {                                │ │
│  │         dbInit.WithAutoMigrate(&types.Agent{}, ...)                 │ │
│  │     }                                                                 │ │
│  │     bs.Register(dbInit)                                              │ │
│  │                                                                      │ │
│  │     // Step 2: Redis (Priority 400)                                 │ │
│  │     redisInit := pkginitializers.NewRedisInitializer(...)           │ │
│  │     bs.Register(redisInit)                                           │ │
│  │                                                                      │ │
│  │     // Step 3: Business Services (Priority 600)                     │ │
│  │     serviceInit := &serviceLayerInitializer{                        │ │
│  │         app:       a,                                                │ │
│  │         dbInit:    dbInit,    ← Explicit dependency                 │ │
│  │         redisInit: redisInit, ← Explicit dependency                 │ │
│  │     }                                                                 │ │
│  │     bs.Register(serviceInit)                                         │ │
│  │                                                                      │ │
│  │     // Step 4: NATS (Priority 500)                                  │ │
│  │     natsInit := &natsInitializer{app: a, opts: a.opts}              │ │
│  │     bs.Register(natsInit)                                            │ │
│  │                                                                      │ │
│  │     // Step 5: Dispatcher Setup (Priority 550)                      │ │
│  │     dispatcherInit := &dispatcherInitializer{app: a}                │ │
│  │     bs.Register(dispatcherInit)                                      │ │
│  │                                                                      │ │
│  │     // Step 6: HTTP Server (Priority 1000)                          │ │
│  │     httpInit := &httpServerInitializer{                             │ │
│  │         app:       a,                                                │ │
│  │         dbInit:    dbInit,                                           │ │
│  │         redisInit: redisInit,                                        │ │
│  │     }                                                                 │ │
│  │     bs.Register(httpInit)                                            │ │
│  │                                                                      │ │
│  │     // Step 7: gRPC Server (Priority 900)                           │ │
│  │     if a.opts.GRPC.Enable {                                          │ │
│  │         grpcInit := &grpcServerInitializer{app: a, dbInit: dbInit}  │ │
│  │         bs.Register(grpcInit)                                        │ │
│  │     }                                                                 │ │
│  │                                                                      │ │
│  │     // Step 8: Health Check (Priority 2000)                         │ │
│  │     healthInit := pkginitializers.NewHealthCheckInitializer(...)    │ │
│  │     bs.Register(healthInit)                                          │ │
│  │                                                                      │ │
│  │     return nil                                                       │ │
│  │ }                                                                     │ │
│  │                                                                      │ │
│  │ ↑ INITIALIZATION ORDER: Visible at a glance (8 clear steps)        │ │
│  │ ↑ DEPENDENCIES: Explicit in code (no Wire magic)                   │ │
│  │ ↑ PRIORITIES: Documented inline (300, 400, 500, ...)               │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌──────────────────────────────────────────────────────────────────────┐ │
│  │ 4. INLINE INITIALIZERS (service-specific logic, ~400 LOC)           │ │
│  │                                                                      │ │
│  │ type serviceLayerInitializer struct {                                │ │
│  │     app       *AgentManagerApp                                       │ │
│  │     dbInit    *pkginitializers.DatabaseInitializer                   │ │
│  │     redisInit *pkginitializers.RedisInitializer                      │ │
│  │ }                                                                     │ │
│  │                                                                      │ │
│  │ func (s *serviceLayerInitializer) Initialize(ctx) error {           │ │
│  │     // Create storage instances                                     │ │
│  │     s.app.mysqlStore = &storage.MySQLStore{...}                     │ │
│  │     s.app.redisStore = &storage.RedisStore{...}                     │ │
│  │                                                                      │ │
│  │     // Create business services                                     │ │
│  │     s.app.registry = agent.NewRegistry(...)                         │ │
│  │     s.app.registry.Start(ctx)                                        │ │
│  │                                                                      │ │
│  │     s.app.eventProcessor = event.NewProcessor(...)                  │ │
│  │     s.app.dispatcher = command.NewDispatcher(...)                   │ │
│  │     return nil                                                       │ │
│  │ }                                                                     │ │
│  │                                                                      │ │
│  │ type natsInitializer struct { ... }           ← 50 LOC               │ │
│  │ type dispatcherInitializer struct { ... }     ← 30 LOC               │ │
│  │ type httpServerInitializer struct { ... }     ← 120 LOC              │ │
│  │ type grpcServerInitializer struct { ... }     ← 100 LOC              │ │
│  └──────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  TOTAL: 505 LOC in ONE file                                               │
│  - Entry point: 30 LOC                                                     │
│  - App struct: 20 LOC                                                      │
│  - registerComponents(): 65 LOC (THE HEART - initialization order)       │
│  - Inline initializers: 390 LOC (service-specific logic)                  │
└────────────────────────────────────────────────────────────────────────────┘
                                    ↓
┌────────────────────────────────────────────────────────────────────────────┐
│                  GENERIC: pkg/initializers/*.go (unchanged)                │
│                                                                            │
│  database.go (150 LOC)  - Used directly by registerComponents()           │
│  redis.go (120 LOC)     - Used directly by registerComponents()           │
│  http_server.go (200)   - Used by httpServerInitializer                   │
│  grpc_server.go (180)   - Used by grpcServerInitializer                   │
│  health.go (100 LOC)    - Used directly by registerComponents()           │
└────────────────────────────────────────────────────────────────────────────┘

TOTAL COMPLEXITY:
  - 1 file in cmd/agent-manager/app/
  - 0 files in internal/agent-manager/initializers/
  - 6 files to understand complete flow
  - 3 abstraction layers
  - 2 concepts (Bootstrap, Inline Initializers)
```

## Side-by-Side Comparison

### HTTP Server Initialization

**BEFORE** (6-file trace):
```
app.go (line 88)
  → InitializeAgentManagerContainer(opts)
    → wire.go (line 38)
      → InitializerSet definition
        → wire_gen.go (line 30)
          → httpServerInit = NewHTTPServerInitializer(...)
            → container.go (line 37)
              → c.http = httpServerInit
                → servers.go (line 38)
                  → type HTTPServerInitializer
                    → Initialize() method (line 63)
                      → pkg/initializers/http_server.go
                        → Actual HTTP server creation
```

**AFTER** (1-file trace):
```
app.go (line 156)
  → httpInit := &httpServerInitializer{...}
    → app.go (line 348)
      → func (h *httpServerInitializer) Initialize(ctx)
        → Creates api.Server with routes
        → Calls pkg/initializers/http_server.go
```

### Dependency Graph Discovery

**BEFORE**: Read 5 files to understand dependencies
```
1. wire.go        - See InitializerSet
2. wire_gen.go    - See actual construction order
3. container.go   - See field dependencies
4. servers.go     - See HTTPServerInitializer dependencies
5. business_services.go - See ServiceInitializer dependencies
```

**AFTER**: Read 1 function in 1 file
```
1. app.go, registerComponents() - See complete initialization order
   - Step 1: dbInit
   - Step 2: redisInit
   - Step 3: serviceInit (depends on: dbInit, redisInit)
   - Step 4: natsInit
   - Step 5: dispatcherInit
   - Step 6: httpInit (depends on: dbInit, redisInit)
   - Step 7: grpcInit (depends on: dbInit)
   - Step 8: healthInit
```

## Key Insight

The visualization clearly shows:

**BEFORE**: Complexity distributed across 9 files with 7 abstraction layers
- Hard to understand
- Hard to debug
- Hard to modify

**AFTER**: Simplicity concentrated in 1 file with 3 layers
- Easy to understand (read one function)
- Easy to debug (set breakpoint in app.go)
- Easy to modify (change one file)

**The complete startup flow is visible in ~65 lines of `registerComponents()`**

---

**Conclusion**: Sometimes the best design is the simplest design.
