# Start Here: New Developer Service Startup Guide

**Welcome to Aetherius!** 👋

This guide will help you understand how to create, modify, or maintain services in this project. Start here and follow the links for more details.

---

## 5-Minute Quick Start

### What Do I Need to Know?

There are **2 patterns** for service startup in this project:

1. **Ultra-Simple Pattern** (100-150 lines of code)
   - For services with **0-3 external dependencies** (rarely used)
   - Example: collect-agent (connects to NATS only)
   - No framework overhead

2. **Bootstrap Pattern** (500-600 lines of code)
   - For services with **5+ external dependencies** (most common)
   - Example: agent-manager (connects to Database, Redis, NATS)
   - Structured 5-layer initialization

### Which Pattern Do I Use?

```
Count your external dependencies (Database, Redis, NATS, APIs, etc.):

  < 5 dependencies?       → Ultra-Simple Pattern
  >= 5 dependencies?      → Bootstrap Pattern
```

That's it! The rest is details.

### Where Should I Look?

- **I'm implementing a new service** → Read [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md)
- **I need complete details** → Read [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md)
- **I want code templates** → See [templates/service_startup_template.go](templates/service_startup_template.go)
- **I'm confused** → This guide + the quick reference

---

## 30-Minute Understanding Path

### Step 1: Understand the Patterns (10 min)

**Bootstrap Pattern** (most common):
```go
// app.go - single file, ~500 lines

func Execute() {
    opts := commonapp.NewStandardOptions("Name", "agent").
        WithDatabase().WithRedis().WithNATS()
    app := &MyApp{}
    commonapp.RunWithBootstrap(app, opts, config, app.registerComponents)
}

type MyApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}

func (a *MyApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // Layer 1: Infrastructure (DB, Redis)
    // Layer 2: Business Services
    // Layer 3: Servers (HTTP, gRPC)
    // Layer 4: Monitoring (Health checks)
    // All with priority-based ordering
    return nil
}
```

**Ultra-Simple Pattern** (rare):
```go
// app.go - single file, ~100 lines

func Execute() {
    opts := commonapp.NewStandardOptions("Name", "agent").WithAgent()
    app := &MyApp{}
    commonapp.Run(app, opts, config)
}

func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    // Direct setup: create logger, client, server
    return nil
}

func (a *MyApp) Run(ctx context.Context) error {
    return a.client.Start(ctx)  // Blocks
}
```

### Step 2: Look at Real Code (10 min)

Open these files in your IDE:

1. **Complex Service** (Bootstrap): `cmd/agent-manager/app/app.go`
   - Multiple dependencies
   - 5 initialization layers
   - Real bootstrap pattern usage

2. **Simple Service** (Ultra-Simple): `cmd/collect-agent/app/app.go`
   - Minimal dependencies
   - Linear initialization
   - Straightforward shutdown

### Step 3: Read Quick Reference (10 min)

Open [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md) and read:
- Pattern Selection
- Pattern-specific code sections
- Your relevant pattern's "Key Code" section

---

## Ready to Implement? (1-2 Hours)

### For a New Service

1. **Decide pattern** (< 1 min)
   - Count dependencies
   - Use matrix in QUICK_SERVICE_STARTUP_REFERENCE.md

2. **Copy template** (< 5 min)
   - Bootstrap Pattern: Copy from [service_startup_template.go](templates/service_startup_template.go) section "Bootstrap Pattern Template"
   - Ultra-Simple Pattern: Copy from same file, "Ultra-Simple Pattern Template"

3. **Customize template** (15-30 min)
   - Replace `{service}` with your service name
   - Implement the Application interface methods
   - Add your business logic

4. **Follow checklist** (30 min)
   - See "Adding a New Service (Checklist)" in [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md)
   - ✅ Directory structure
   - ✅ File implementation
   - ✅ Makefile registration
   - ✅ Build test
   - ✅ Run test
   - ✅ Write tests

5. **Build and run** (5 min)
   ```bash
   make go.build.{myservice}
   make run-{myservice}
   ```

### For Modifying Existing Service

1. **Find your pattern** in cmd/{service}/app/app.go
2. **Look for example of what you want** in [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) "Common Patterns" section
3. **Apply to your service**
4. **Test**: `make go.test.{service}`

---

## Common Scenarios

### "I want to add a new database connection"

1. See: [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) "Pattern: Service with Database Migration"
2. Or: Look at auth service how it connects to multiple things

### "The service doesn't start. What's wrong?"

1. Check: [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md) "Troubleshooting" section
2. If not there: [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) "Troubleshooting" section

### "I need to add feature X"

1. Look for similar service that has it
2. Read: [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) "Common Patterns" section
3. Compare: Your service structure vs. the example

### "I'm getting initialization errors"

Usually means wrong priority or missing dependency registration.

1. Check: [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) "Initialization Priority Guidelines"
2. Fix: Move component to later priority
3. Rule: Infrastructure (300-400) → Services (600) → Servers (900+)

---

## Key Concepts Explained Simply

### StandardOptions

**What it is**: Configuration builder that sets up what your service needs.

```go
// I need database, Redis, and messaging
opts := commonapp.NewStandardOptions("MyService", "my-user-agent").
    WithDatabase().    // Add MySQL support
    WithRedis().       // Add Redis support
    .WithNATS()        // Add NATS support
    .WithMetrics()     // Add Prometheus metrics
```

**Environment Override**:
```bash
# These override the config file
DATABASE_HOST=my-host
REDIS_ADDR=my-redis:6379
NATS_URL=nats://my-nats:4222
```

### Bootstrap Pattern Layers

**The 5 layers execute in order**:

```
Layer 1: Infrastructure (Priority 300-500)
  ↓ Database, Redis, Email clients

Layer 2: Business Services (Priority 600-700)
  ↓ Your main service logic

Layer 3: Advanced Services (Priority 650-800)
  ↓ Features like sessions, audit trails

Layer 4: Servers (Priority 900-1000)
  ↓ HTTP and gRPC servers

Layer 5: Monitoring (Priority 2000)
  ↓ Health checks
```

**Why priorities matter**:
- Database must start before services that use it
- Services must start before servers that expose them
- Health checks run last to verify everything works

### Application Interface

All services implement this:

```go
type Application interface {
    Name() string                                          // Your service name
    Initialize(ctx context.Context, opts Options) error   // Setup phase
    Run(ctx context.Context) error                        // Running phase (blocks)
    Shutdown(ctx context.Context) error                   // Cleanup phase
}
```

---

## File Organization

```
Your new service:
cmd/myservice/
├── main.go                 ← 5 lines (copy-paste)
└── app/
    └── app.go              ← 100-150 (ultra-simple) or 500-600 (bootstrap)

internal/myservice/
├── service/                ← Your business logic
├── storage/                ← Database access
├── api/                    ← HTTP handlers
└── grpc/                   ← gRPC services

(Optional for bootstrap pattern)
internal/myservice/startup/
├── infrastructure.go       ← DB, Redis setup
├── core_services.go        ← Main service creation
└── servers.go              ← HTTP/gRPC setup
```

---

## Documentation Map

```
START HERE
    ↓
This file (you are here!)
    ↓
[Choose your path]
    ↓
┌─────────────────────────────────────┐
│ Just need quick info?               │
│ → QUICK_SERVICE_STARTUP_REFERENCE   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Need complete understanding?        │
│ → SERVICE_STARTUP_GUIDE.md          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Ready to implement?                 │
│ → templates/service_startup_template │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Need navigation?                    │
│ → SERVICE_STARTUP_INDEX.md          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Need complete details?              │
│ → IMPLEMENTATION_SUMMARY.md         │
└─────────────────────────────────────┘
```

---

## Cheat Sheet

### Creating Ultra-Simple Service
```bash
mkdir -p cmd/{service}/app internal/{service}
# Copy template from templates/service_startup_template.go
# Implement in cmd/{service}/app/app.go
# Add {service} to SERVICES in Makefile
make go.build.{service}
```

### Creating Bootstrap Service
```bash
mkdir -p cmd/{service}/app internal/{service}/startup
# Copy template from templates/service_startup_template.go
# Implement registerComponents() with 5 layers
# Add {service} to SERVICES in Makefile
make go.build.{service}
```

### Testing
```bash
make go.test.{service}          # Run tests
make go.build.{service}         # Build it
make run-{service}              # Run it
```

### Troubleshooting
```bash
# Check logs
docker-compose logs -f {service}

# Port already in use?
lsof -ti:{port} | xargs kill -9

# Need databases?
cd deployments/docker-compose && docker-compose up -d mysql redis nats
```

---

## Before You Ask for Help

1. **Have you read** [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md)? (5 min)
2. **Have you looked at** an existing similar service? (10 min)
3. **Did you check** the troubleshooting section? (5 min)
4. **Did you verify** your implementation against the checklist? (5 min)

If you've done all 4 and still stuck, great! Then ask. You'll have more context.

---

## Success Metrics

You've succeeded when:
- ✅ `make go.build.{service}` runs without errors
- ✅ `make run-{service}` starts the service
- ✅ `curl http://localhost:PORT/health` returns 200 OK
- ✅ `make go.test.{service}` passes all tests
- ✅ Service shuts down gracefully with SIGTERM
- ✅ Code follows the patterns in QUICK_SERVICE_STARTUP_REFERENCE.md

---

## Next Steps

**Pick one**:

### I want the quick overview
→ Read [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md) (5-10 min)

### I want complete understanding
→ Read [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md) (30-45 min)

### I want to start implementing now
→ Copy template from [templates/service_startup_template.go](templates/service_startup_template.go) and use checklist

### I want to understand all the details
→ Read [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) (complete overview)

### I'm lost on navigation
→ Read [SERVICE_STARTUP_INDEX.md](SERVICE_STARTUP_INDEX.md) (map of all docs)

---

**Good luck! Welcome to the team! 👋**

You're now ready to implement services in Aetherius. The documentation has examples from 8 working services, comprehensive guides, and ready-to-use templates. You've got this!

---

**Quick Links**:
- Quick Reference: [QUICK_SERVICE_STARTUP_REFERENCE.md](QUICK_SERVICE_STARTUP_REFERENCE.md)
- Full Guide: [SERVICE_STARTUP_GUIDE.md](SERVICE_STARTUP_GUIDE.md)
- Code Templates: [templates/service_startup_template.go](templates/service_startup_template.go)
- Navigation: [SERVICE_STARTUP_INDEX.md](SERVICE_STARTUP_INDEX.md)

**Real Examples**:
- Complex (Bootstrap): `cmd/agent-manager/app/app.go`
- Simple (Ultra-Simple): `cmd/collect-agent/app/app.go`
