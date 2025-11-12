# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Critical Build Rules

**ALL commands MUST be run from repository root.** The modular Makefile system will fail if commands are run from subdirectories.

```bash
# ✅ CORRECT - Always from root
make build
make test
make run-agent-manager-local

# ❌ WRONG - Never do this
cd cmd/agent-manager && make build
```

## Essential Commands

### Build and Run

```bash
# Build all services (outputs to _output/bin/)
make build

# Build specific service
make go.build.agent-manager
make go.build.orchestrator
make go.build.reasoning

# Run services locally (each needs separate terminal)
make run-agent-manager-local   # Port 8080
make run-orchestrator-local     # Port 8081
make run-reasoning-local        # Port 8082
make run-auth-local            # Port 8080
make run-cluster-local         # Port 8084

# Start dependencies (MySQL, Redis, NATS)
make run-deps
make run-check-deps  # Verify they're running
```

### Testing

```bash
# Run all tests
make test

# Test specific service
make go.test.agent-manager
make go.test.orchestrator

# Run single test function
go test -v ./internal/agent-manager/agent -run TestAgentRegistry_Register

# Coverage with HTML reports (_output/coverage/)
make test-coverage

# Integration tests
make test-integration

# E2E tests
make test-e2e
```

### Code Quality

```bash
# Format code (uses gofumpt + gci)
make fmt

# Run all linters (58 enabled)
make lint

# Run specific linters
make go.lint

# Pre-commit checks (fmt + lint + test)
make pre-commit
```

### Docker Operations

```bash
# Build images
make docker-build

# Build specific service
make docker.build.agent-manager

# Multi-platform build (amd64 + arm64)
make docker-buildx VERSION=v1.2.3

# Push images
make docker-push
```

### Development Workflow

```bash
# Initial setup (installs tools, git hooks)
make dev-setup

# Verify environment
make dev-ready

# Clean build artifacts
make clean

# Complete rebuild
make rebuild  # clean + build
```

## Architecture Overview

### 4-Layer System Architecture

```
Layer 1: Collect Agent → NATS → Layer 2: Agent Manager
                                        ↓
                                  Internal Bus
                                        ↓
                         Layer 3: Orchestrator Service
                                        ↓
                                    HTTP API
                                        ↓
                          Layer 4: Reasoning Service
```

**Data Flow Pattern**:

1. Collect Agent monitors K8s events in each cluster
2. Events sent via NATS to Agent Manager (central control)
3. Agent Manager evaluates and publishes to internal bus
4. Orchestrator subscribes, matches strategies, executes workflows
5. Workflows call Reasoning Service for AI analysis
6. Results trigger remediation or notifications

### Service Startup Patterns

The codebase uses three startup patterns (post-refactoring):

**Pattern 1: Ultra-Simple** (agent-manager, orchestrator, auth)

- Single `cmd/{service}/app/app.go` (~500 LOC)
- Bootstrap framework with priority-based init
- Direct instantiation, no Wire DI
- Used for complex services with many dependencies

**Pattern 2: Simple** (collect-agent, gateway, monitor)

- Basic `run()` function, no Bootstrap
- Linear initialization
- Used for lightweight services

**Pattern 3: Simplified Bootstrap** (cluster, reasoning)

- Bootstrap framework without complexity
- Mid-complexity services

**Key Design Decision**: Wire DI was eliminated - explicit instantiation is preferred for clarity.

### Code Organization

```
k8s-agent/
├── cmd/                    # Service entry points (main.go + app/)
├── internal/               # Private service implementations
│   └── {service}/
│       ├── api/           # HTTP handlers (Gin)
│       ├── grpc/          # gRPC services
│       ├── storage/       # Database layer
│       └── config/        # Service config
├── common/                 # Generic utilities (independent module)
├── pkg/                    # Business logic specific to project
│   ├── initializers/      # Shared infrastructure init
│   ├── bootstrap/         # App lifecycle management
│   └── types/            # Domain models
└── api/proto/             # Protocol buffers
```

**Key Principle**:

- `common/` = Zero business logic, reusable anywhere
- `pkg/` = Project-specific business logic
- `internal/` = Private service code

### Service Communication

- **Agent → Manager**: NATS pub/sub (events, metrics, heartbeats)
- **Manager → Orchestrator**: Internal NATS bus
- **Orchestrator → Reasoning**: HTTP REST API
- **All services → DB**: MySQL for persistence, Redis for cache/sessions

### Workflow Engine (Orchestrator)

Supports 6 step types:

- **Command**: Execute kubectl via Agent Manager
- **AI**: Call Reasoning Service
- **Decision**: Conditional branching
- **Remediation**: Apply fixes
- **Notification**: Send alerts
- **Wait**: Delay/observation

## Custom SpecKit Commands

The project includes workflow commands for structured development:

- `/speckit.specify` - Create feature specification
- `/speckit.plan` - Generate implementation plan
- `/speckit.tasks` - Create task list
- `/speckit.implement` - Execute tasks
- `/speckit.clarify` - Gather requirements
- `/speckit.analyze` - Cross-artifact analysis
- `/speckit.checklist` - Generate checklist
- `/speckit.constitution` - Update project principles

See `.claude/commands/speckit.*.md` for details.

## Key Technical Details

### Technology Stack

- **Go**: 1.25.0 (strict requirement)
- **Framework**: Gin v1.11.0
- **Database**: MySQL 8.0+
- **Cache**: Redis 6+
- **Messaging**: NATS 2.10+
- **Kubernetes**: client-go v0.31.3
- **Logging**: github.com/kart-io/logger (Zap/Slog dual-engine)
- **AI**: OpenAI/Gemini/DeepSeek APIs via gollm

### Service Ports

- Agent Manager: 8080
- Orchestrator: 8081
- Reasoning: 8082
- Auth: 8080
- Cluster: 8084
- Monitor: 8085
- Gateway: 8086

### Performance Targets

- Agent Manager: 1000+ agents, 10K events/min
- Orchestrator: 500+ concurrent workflows
- Reasoning: 100+ analysis requests/min
- MTTD < 1 minute, MTTR < 5 minutes

### Testing Strategy

- Unit tests: Alongside implementation
- Integration: `internal/{service}/test/integration/`
- Coverage: HTML reports in `_output/coverage/`
- Mocking: testify + go-sqlmock

## Important Patterns

### Error Handling

```go
import "github.com/kart-io/k8s-agent/common/errors"

// Wrap with context and code
return errors.Wrap(err, errors.CodeInternal, "operation failed")
```

### Service Initialization

```go
// All services follow this pattern in cmd/{service}/app/app.go
func (a *App) registerComponents(bs *bootstrap.Bootstrap) error {
    // Infrastructure - Priority 300-400
    dbInit := pkginitializers.NewDatabaseInitializer(...)
    bs.Register(dbInit)

    // Business logic - Priority 600
    serviceInit := &serviceInitializer{app: a}
    bs.Register(serviceInit)

    // Servers - Priority 1000
    httpInit := &httpServerInitializer{app: a}
    bs.Register(httpInit)
}
```

### Common Initializers (pkg/initializers/)

- DatabaseInitializer - MySQL/GORM with migrations
- RedisInitializer - Redis client with health checks
- NATSInitializer - NATS with auto-reconnect
- HTTPServerInitializer - Gin server with middleware
- GRPCServerInitializer - gRPC with graceful shutdown

## Quick Troubleshooting

**Build fails**: `make clean && make deps`

**Port conflicts**: Check with `lsof -ti:8080`

**Database issues**: `make run-check-deps`

**Service crashes**: Enable debug logging in configs/

**NATS connection**: Verify with `docker-compose ps nats`

For detailed documentation see `docs/` directory.
