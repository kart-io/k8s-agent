# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Aetherius** is an enterprise-grade intelligent Kubernetes operations platform (智能 Kubernetes 运维平台) with AI-driven fault diagnosis and automated remediation. The system is a **monorepo** with a modular build system, using a 4-layer architecture combining event-driven design with AI technology to create a complete operational loop from data collection to intelligent analysis.

### Core Capabilities

- **Auto-discovery**: Real-time monitoring of K8s cluster anomaly events
- **Root Cause Analysis**: AI-driven multi-modal analysis (events + logs + metrics)
- **Intelligent Recommendations**: Rule-based and historical case-driven repair suggestions
- **Automated Remediation**: Workflow-driven automated repair execution
- **Continuous Learning**: Learn from feedback to continuously improve accuracy
- **Multi-cluster Management**: Unified management of hundreds of K8s clusters
- **Knowledge Graph**: Store operational experience in Neo4j

## Architecture

The system follows a 4-layer architecture:

```
Layer 1: Collect Agent (边缘采集层)
  ↓ NATS messaging
Layer 2: Agent Manager (中央控制层)
  ↓ Internal event bus
Layer 3: Orchestrator Service (任务编排层)
  ↓ HTTP API
Layer 4: Reasoning Service (AI 智能层)
```

### Layer 1: Collect Agent

- **Purpose**: Edge data collection deployed in each K8s cluster
- **Tech**: Go 1.25+, client-go, NATS
- **Functions**: K8s event monitoring (85+ event types), resource metrics collection, secure command execution
- **Entry Point**: `cmd/collect-agent/`
- **Implementation**: `internal/collect-agent/`

### Layer 2: Agent Manager

- **Purpose**: Central control plane managing all agents
- **Tech**: Go 1.25+, MySQL, Redis, NATS, Gin
- **Functions**: Agent lifecycle management, event aggregation/routing, command scheduling, multi-cluster management
- **API Port**: 8080
- **Entry Point**: `cmd/agent-manager/`
- **Implementation**: `internal/agent-manager/`

### Layer 3: Orchestrator Service

- **Purpose**: Workflow orchestration for automated diagnosis and remediation
- **Tech**: Go 1.25+, MySQL, Redis, NATS
- **Functions**: Workflow engine, diagnostic strategies, 6 step types (Command/AI/Decision/Remediation/Notification/Wait), AI integration
- **API Port**: 8081
- **Entry Point**: `cmd/orchestrator/`
- **Implementation**: `internal/orchestrator/`

### Layer 4: Reasoning Service

- **Purpose**: AI-driven root cause analysis and intelligent recommendations
- **Tech**: Go 1.25+, Gin, Neo4j, OpenAI/Gemini/DeepSeek API
- **Functions**: Root cause analysis, recommendation engine (30+ rules), prediction engine, knowledge graph (Neo4j), continuous learning
- **API Port**: 8082
- **Entry Point**: `cmd/reasoning/`
- **Implementation**: `internal/reasoning/`

### Supporting Services

- **Auth Service**: JWT authentication, session management, forced logout functionality
  - Tech: Go 1.25, Gin, JWT, Redis, MySQL
  - Entry Point: `cmd/auth/`
  - Implementation: `internal/auth/`

- **Gateway Service**: API gateway with Traefik integration
  - Entry Point: `cmd/gateway/`
  - Implementation: `internal/gateway/`

- **Monitor Service**: Monitoring and metrics collection
  - Entry Point: `cmd/monitor/`
  - Implementation: `internal/monitor/`

- **Cluster Service**: Multi-cluster management
  - Entry Point: `cmd/cluster/`
  - Implementation: `internal/cluster/`

### Shared Code Organization

The project follows a strict separation between generic utilities and project-specific business logic:

#### common/ - Generic Foundation Package (Independent Module)

**Purpose**: Generic, reusable utilities that ANY Go project can use. This is an independent module that can be extracted and used in other projects.

**Characteristics**:
- Zero business logic, pure technical implementation
- Independent Go module with its own `go.mod`
- Similar to a third-party library
- Can be open-sourced as a standalone package

**Contents** (after reorganization):
- `cache/`: Unified caching interface (memory and Redis backends)
- `client/`: Common clients (NATS, gRPC, HTTP)
- `config/`: Configuration management with Options pattern (53 config functions)
- `db/`: Database client wrappers (MySQL, Redis)
- `errors/`: Standardized error handling
- `k8sutils/`: Generic Kubernetes resource conversion utilities
- `logger/`: Legacy logger (should use github.com/kart-io/logger instead)
- `middleware/`: Common HTTP middleware (CORS, rate limiting, auth)
- `mq/`: Message queue abstractions (NATS)
- `pagination/`: Generic pagination utilities
- `response/`: Unified API response format
- `server/`: HTTP/gRPC server wrappers (Gin, gRPC)
- `telemetry/`: OpenTelemetry integration (generic parts)
- `utils/`: Generic utility functions
- `validator/`: Data validation utilities

**Note**: `app/` and `types/` have been moved to `pkg/` as they contain business logic.

#### pkg/ - Project-Specific Package (Business Logic)

**Purpose**: Business logic and domain models specific to the **Aetherius (k8s-agent)** project. Not suitable for use in other projects.

**Characteristics**:
- Contains business logic and domain-specific code
- Tightly coupled with project requirements
- Can import from `common/` but not vice versa
- Located at root level `pkg/` (not `internal/pkg/`)

**Current Contents** (migrated from internal/pkg/ and common/):
- `app/`: Application startup and command initialization (from common/)
- `bootstrap/`: Application bootstrapping logic (from internal/pkg/)
- `contextx/`: Project-specific context management (from internal/pkg/)
- `idempotent/`: Business idempotency handling (from internal/pkg/)
- `metrics/`: Project-specific Prometheus metrics (from internal/pkg/)
- `types/`: Business domain models - Agent, Event, Command, Metrics (from common/)
- `k8s/`: Kubernetes business logic (created for future use)
- `agent/`: Agent domain models and business rules (created for future use)
- `workflow/`: Workflow orchestration business logic (created for future use)
- `diagnosis/`: Diagnostic strategies and rules (created for future use)

#### api/proto/ - Protocol Buffer Definitions (Independent Module)

- Protocol Buffer definitions for inter-service communication
- Generates Go code to `api/proto/gen/go/`
- Generates OpenAPI v2 specs to `api/proto/gen/openapiv2/`
- Independent module with its own `go.mod`

### Code Organization Decision Criteria

**Place in common/ if ALL are true**:
1. Can be used by ANY Go project
2. Contains ZERO business logic
3. Could be published as an independent library
4. Would still be useful if moved to a different project

**Place in pkg/ if ANY is true**:
1. Contains business logic
2. Specific to Aetherius project
3. Depends on project domain models (Agent, Workflow, etc.)
4. Not useful for other projects

**See**: [docs/CODE_REORGANIZATION.md](docs/CODE_REORGANIZATION.md) for detailed reorganization plan

## Common Development Commands

The project uses a **modular Makefile system** (inspired by OneX) with rules split across `scripts/make-rules/*.mk` files. All commands should be run from the repository root.

### Build Commands

```bash
# Build all services (outputs to _output/bin/)
make build

# Build specific service
make go.build.agent-manager
make go.build.orchestrator
make go.build.reasoning
make go.build.auth

# Build for multiple architectures
make go.build.multiarch

# Quick rebuild (clean + build)
make rebuild
```

### Run Commands

```bash
# Run specific service directly (from root)
make run-agent-manager
make run-orchestrator
make run-reasoning

# Run with hot reload (requires air)
make dev

# Run all services with Docker Compose (recommended for development)
make docker-compose-up
```

### Testing Commands

```bash
# Test all services
make test

# Test specific service
make go.test.agent-manager
make go.test.orchestrator

# Test with coverage (generates HTML reports in _output/coverage/)
make test-coverage

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Quick test (unit tests only)
make quick-test

# Full test suite (unit + integration)
make full-test
```

### Code Quality Commands

```bash
# Format all code
make fmt

# Run all linters (golangci-lint with 58 enabled linters)
make lint

# Run go vet
make vet

# Run all checks (lint + vet)
make check

# Pre-commit checks (format + lint + test)
make pre-commit

# Pre-push checks (lint + test + build)
make pre-push
```

### Docker Commands

```bash
# Build all Docker images
make docker-build VERSION=v1.0.0

# Build specific service image
make docker.build.agent-manager

# Build multi-platform images (linux/amd64, linux/arm64)
make docker-buildx VERSION=v1.0.0

# Build and push multi-platform images
make docker-buildx-push VERSION=v1.0.0

# Push Docker images
make docker-push
```

### Protocol Buffer Generation

```bash
# Generate all protobuf code
make gen-proto

# Clean generated code
make proto.clean

# Build protoc plugins
make proto.build
```

### Development Environment

```bash
# Setup development environment (install tools + git hooks)
make dev-setup

# Install all development tools
make install-tools

# Install additional tools (air, mockgen, etc.)
make install-tools A=1

# Verify tools installation
make tools.verify

# Install git hooks
make hooks.install

# Check dev environment readiness
make dev-ready
```

### Kubernetes Deployment

```bash
# Deploy to specific environment
make deploy ENV=dev
make deploy ENV=staging
make deploy ENV=prod

# Validate Kubernetes manifests
make manifests-validate

# Deploy with Kustomize
cd deployments/k8s/overlays/dev && kubectl apply -k .
```

### Database Operations

```bash
# Start database dependencies only
make db-setup

# Reset all databases
make db-reset

# Connect to MySQL
docker-compose -f deployments/docker-compose/docker-compose.yaml exec mysql mysql -u aetherius -p

# Connect to Redis
make redis-cli
```

### Dependency Management

```bash
# Download dependencies
make deps

# Tidy dependencies
make tidy

# Verify dependencies
make deps-verify
```

### Utilities

```bash
# Show version information
make version

# Show project information
make info

# Show project statistics (files, lines, targets, etc.)
make stats

# Show all available targets
make help

# Show targets from all sub-makefiles
make targets

# List all included makefiles
make list-mk

# Clean build artifacts
make clean

# Clean everything (including vendor)
make clean-all
```

### CI/CD

```bash
# Run full CI pipeline (deps + fmt + vet + lint + test + build)
make ci

# Create release
make release VERSION=v1.0.0
```

## Project Structure

The repository is a **monorepo** with services organized by domain, not by separate directories:

```
k8s-agent/
├── go.mod                         # Root Go module (Go 1.25.0)
├── Makefile                       # Root orchestration (includes modular make rules)
├── scripts/
│   └── make-rules/               # Modular Makefile system (inspired by OneX)
│       ├── common.mk             # Common variables and functions
│       ├── golang.mk             # Go build/test/quality targets
│       ├── docker.mk             # Docker build targets
│       ├── proto.mk              # Protobuf generation
│       ├── tools.mk              # Development tools management
│       ├── hooks.mk              # Git hooks installation
│       ├── k8s.mk                # Kubernetes deployment
│       ├── version.mk            # Version injection
│       ├── lint.mk               # Linting configuration
│       └── ...                   # Other modular rules
├── cmd/                          # Service entry points
│   ├── agent-manager/
│   │   ├── app/                  # Application setup
│   │   └── main.go               # Entry point
│   ├── orchestrator/
│   ├── reasoning/
│   ├── auth/
│   ├── gateway/
│   ├── monitor/
│   ├── cluster/
│   └── collect-agent/
├── internal/                     # Service implementations (private)
│   ├── agent-manager/
│   │   ├── api/                  # HTTP API handlers
│   │   ├── config/               # Configuration
│   │   ├── storage/              # Database/storage layer
│   │   ├── agent/                # Agent registry and management
│   │   ├── command/              # Command dispatcher
│   │   ├── event/                # Event processor
│   │   ├── grpc/                 # gRPC services
│   │   └── nats/                 # NATS integration
│   ├── orchestrator/
│   ├── reasoning/
│   ├── auth/
│   └── ...
├── common/                       # Generic foundation package (Go module)
│   ├── cache/                    # Unified caching
│   ├── client/                   # Common clients
│   ├── config/                   # Options-based configuration
│   ├── db/                       # Database clients
│   ├── errors/                   # Error handling
│   ├── k8sutils/                 # K8s utilities (generic only)
│   ├── middleware/               # HTTP middleware
│   ├── mq/                       # Message queue
│   ├── pagination/               # Pagination
│   ├── response/                 # Response format
│   ├── server/                   # Server setup
│   ├── telemetry/                # Telemetry (generic)
│   ├── types/                    # Common types
│   ├── utils/                    # Generic utils
│   ├── validator/                # Validation
│   └── go.mod                    # Independent module
├── pkg/                          # Project-specific business logic
│   ├── bootstrap/                # App initialization
│   ├── contextx/                 # Context management
│   ├── idempotent/               # Idempotency
│   ├── metrics/                  # Project metrics
│   ├── k8s/                      # K8s business logic
│   ├── agent/                    # Agent domain
│   ├── workflow/                 # Workflow logic
│   ├── diagnosis/                # Diagnosis logic
│   ├── types/                    # Business types
│   ├── telemetry/                # Project telemetry
│   └── utils/                    # Business utils
├── api/proto/                    # Protocol Buffer definitions (Go module)
│   ├── common/                   # Common proto messages
│   ├── orchestrator/             # Orchestrator service protos
│   ├── reasoning/                # Reasoning service protos
│   ├── gen/go/                   # Generated Go code
│   └── gen/openapiv2/            # Generated OpenAPI specs
├── deployments/
│   ├── docker-compose/           # Docker Compose setup
│   ├── k8s/                      # Kubernetes manifests
│   │   ├── base/                 # Base configurations
│   │   └── overlays/             # Environment-specific overlays
│   └── kustomize/                # Kustomize configurations
├── docs/
│   ├── architecture/             # Architecture documentation
│   └── CODE_REORGANIZATION.md    # Code organization plan
├── _output/                      # Build outputs (created by make)
│   ├── bin/                      # Compiled binaries
│   └── coverage/                 # Test coverage reports
└── configs/                      # Configuration files (if present)
```

### Key Architecture Patterns

1. **Monorepo Structure**: All services share the same root `go.mod` with replace directives for internal modules
2. **Modular Makefile**: Build system split into logical `.mk` files for maintainability
3. **Three-Layer Code Organization**:
   - `common/` = Generic utilities (can use in any project)
   - `pkg/` = Business logic (specific to Aetherius)
   - `internal/` = Service implementations (private)
4. **Domain-Driven Structure**: `internal/<service>/` follows domain-driven design patterns
5. **Protocol Buffer First**: API definitions in `api/proto/` generate code for multiple languages

## Technology Stack

### Backend (Go Services)

- **Go Version**: 1.25.0+ (workspace), services may specify minimum 1.21+
- **Web Framework**: Gin v1.11.0
- **Messaging**: NATS 2.10+
- **Database**: MySQL 8.0+ (migrated from PostgreSQL)
- **Cache**: Redis 6+
- **ORM**: GORM v1.31.0
- **Logging**:
  - `github.com/kart-io/logger` (dual-engine: Zap/Slog with OTLP integration)
  - Logrus (legacy, being phased out)
- **Version Management**: `github.com/kart-io/version` (build-time injection)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **AI Integration**:
  - OpenAI/Gemini/DeepSeek API (reasoning service)
  - `github.com/teilomillet/gollm` v0.1.9 (LLM abstraction)
- **Kubernetes**: client-go v0.31.3, api v0.31.3, metrics v0.31.3
- **Metrics**: Prometheus client v1.23.2
- **Testing**: testify v1.11.1, go-sqlmock v1.5.2

### Infrastructure

- **Container**: Docker 20.10+
- **Orchestration**: Kubernetes 1.23+
- **Service Mesh**: Traefik (in gateway-service)
- **Monitoring**: Prometheus + Grafana
- **Tracing**: OpenTelemetry (planned)

## Configuration Management

Services use YAML configuration with Viper, supporting environment variable overrides:

```bash
# Services read config from various sources (priority order):
# 1. Environment variables (highest priority)
# 2. Command-line flags (-c, --config)
# 3. Configuration file (configs/config.yaml or specified path)
# 4. Default values (lowest priority)

# Configuration files are typically in each service's configs/ directory
# Common config options are defined in common/config/ package
```

### Configuration Structure

The `common/config/` package provides standardized configuration options:

- **Server Options** (`server_options.go`): HTTP/gRPC server settings
- **Database Options** (`database_options.go`): MySQL connection pooling
- **Redis Options** (`redis_options.go`): Redis client configuration
- **NATS Options** (`nats_options.go`): NATS messaging setup
- **JWT Options** (`jwt_options.go`): JWT authentication settings
- **CORS Options** (`cors_options.go`): CORS middleware configuration
- **Metrics Options** (`metrics_options.go`): Prometheus metrics setup

### Example Usage

Services typically define a config structure embedding common options:

```go
import "github.com/kart-io/k8s-agent/common/config"

type Config struct {
    Server   config.ServerOptions
    Database config.DatabaseOptions
    Redis    config.RedisOptions
    NATS     config.NATSOptions
    // Service-specific options...
}
```

## Modular Make System

The project uses a sophisticated modular Makefile system inspired by the OneX project. Understanding this system is crucial for development:

### Make Rule Modules

Each `.mk` file in `scripts/make-rules/` provides specific functionality:

- **common.mk**: Core variables (VERSION, GIT_COMMIT, PLATFORM, directories)
- **golang.mk**: Go targets (build, test, vet, fmt, lint)
- **docker.mk**: Docker build targets (single/multi-platform, push)
- **proto.mk**: Protobuf code generation (Go, OpenAPI v2)
- **tools.mk**: Development tool installation (golangci-lint, air, mockgen, etc.)
- **hooks.mk**: Git hooks setup (pre-commit, commit-msg, etc.)
- **version.mk**: Version injection via ldflags
- **k8s.mk**: Kubernetes deployment helpers
- **lint.mk**: Linting configuration (58 enabled linters)
- **release.mk**: Release pipeline orchestration
- **deploy.mk**: Deployment automation
- **gen.mk**: Code generation orchestration
- **copyright.mk**: Copyright header management
- **swagger.mk**: Swagger/OpenAPI generation

### Target Naming Conventions

The modular system uses hierarchical target naming:

```bash
# Pattern: <module>.<action>[.<service>]
make go.build                    # Build all services
make go.build.agent-manager      # Build specific service
make go.test.orchestrator        # Test specific service
make docker.build.reasoning      # Build Docker image for service
make proto.generate              # Generate all protobuf code
make tools.install               # Install all required tools
make tools.install.air           # Install specific tool
```

### Legacy Compatibility

The root Makefile provides legacy-compatible aliases:

```bash
# These are equivalent:
make build          ≡  make go.build
make test           ≡  make go.test
make lint           ≡  make go.lint
make docker         ≡  make docker.build
make gen-proto      ≡  make proto.generate
```

### Adding New Services

To add a new service to the build system:

1. Create service entry point: `cmd/<service>/main.go`
2. Create implementation: `internal/<service>/`
3. Add service name to `SERVICES` variable in root Makefile (line 12)
4. Build system automatically picks up the new service

### Makefile Best Practices

1. **Always run from root**: Never `cd` into service directories
2. **Use modular targets**: Prefer `go.build.X` over `build-X` for clarity
3. **Check available targets**: Run `make help` or `make targets`
4. **Parallel builds**: The system supports parallel builds with `-j` flag
5. **Verbose mode**: Set `V=1` for verbose output (e.g., `make V=1 build`)

## Service Communication

### Internal Communication

- **Agent → Agent Manager**: NATS messages (events, metrics, heartbeats)
- **Agent Manager → Orchestrator**: Internal event bus (NATS)
- **Orchestrator → Reasoning**: HTTP API (port 8082)
- **Services → Databases**: MySQL (data persistence), Redis (cache/sessions)

### External Communication

- **API Access**: Through Gateway Service (Traefik) or direct service ports
- **Agent Manager API**: Port 8080
- **Orchestrator API**: Port 8081
- **Reasoning Service API**: Port 8082

## Key Workflows

### 1. Event Processing Flow

```
Collect Agent (monitors K8s event)
  → NATS message to Agent Manager
  → Agent Manager evaluates severity
  → Publishes to internal event bus
  → Orchestrator subscribes and matches strategy
  → Executes diagnostic workflow
  → Calls Reasoning Service for AI analysis
  → Executes remediation steps
  → Sends notifications
```

### 2. Diagnostic Workflow Execution

The Orchestrator Service uses 6 step types in workflows:

- **Command**: Execute kubectl/diagnostic commands via Agent Manager
- **AI**: Call Reasoning Service for root cause analysis
- **Decision**: Conditional branching based on analysis results
- **Remediation**: Execute repair actions (scale, restart, update)
- **Notification**: Send alerts to operators
- **Wait**: Delay for observation or rate limiting

### 3. Multi-cluster Agent Management

Agent Manager tracks agent lifecycle:

1. **Registration**: Agent registers with cluster metadata
2. **Heartbeat**: Periodic health checks (default: 30s)
3. **Command Dispatch**: Secure command execution with whitelisting
4. **Status Tracking**: Monitor agent health and cluster status
5. **Deregistration**: Clean removal on shutdown

## Database Schemas

### MySQL Databases

- `aetherius`: Agent Manager data (agents, clusters, events, commands)
- `aetherius_orchestrator`: Orchestrator data (workflows, strategies, executions)
- `aetherius_auth`: Auth Service data (users, sessions, audit logs)

### Redis Data Structures

- Session storage: `session:{jti}`, `user:sessions:{user_id}`
- JWT blacklist: `revoked:{jti}`
- Cache: Various service-specific keys

### Neo4j Graph Database

- Knowledge graph for historical cases
- Incident relationships and patterns
- Continuous learning data

## Performance Characteristics

### Throughput Targets

- **Single Agent Manager**: 1000+ agents, 10,000+ events/min
- **Single Orchestrator**: 500+ concurrent workflows, 5,000+ tasks/min
- **Single Reasoning Service**: 100+ analysis requests/min

### Latency Targets

- Event processing delay: < 1s
- Workflow trigger delay: < 5s
- Root cause analysis: P99 < 5s

### Business Metrics

- Root cause analysis accuracy: ~85-95%
- Auto-remediation success rate: ~80-90%
- Mean Time to Detect (MTTD): < 1 minute
- Mean Time to Repair (MTTR): < 5 minutes (with auto-remediation)

## Security Considerations

- **Authentication**: JWT tokens, mTLS between services
- **Authorization**: RBAC, command whitelisting
- **Transport Encryption**: TLS 1.3
- **Data Encryption**: Database TDE (Transparent Data Encryption)
- **Audit Logging**: All critical operations logged with hash chain
- **Session Management**: Redis-based with forced logout capability
- **Rate Limiting**: API rate limits enforced (e.g., 100 req/min for admin endpoints)

## Multi-Platform Docker Builds

All services support multi-platform builds (linux/amd64, linux/arm64):

```bash
# Setup buildx builder (one-time)
cd <service> && make docker-buildx-setup

# Build for multiple platforms
cd <service> && make docker-buildx

# Build and push to registry
cd <service> && make docker-buildx-push

# Build with environment tag
cd <service> && make docker-buildx-env ENV=dev
cd <service> && make docker-buildx-push-env ENV=dev
```

## Testing Strategy

### Unit Tests

- Test individual functions and packages
- Located alongside implementation files (e.g., `agent_registry_test.go`)
- Run with: `make test` or `make go.test.agent-manager`
- Use testify for assertions and go-sqlmock for database mocking

### Integration Tests

- Test service interactions and external dependencies
- Located in `internal/<service>/test/integration/` or service-level test directories
- Run with: `make test-integration`
- Tagged with `// +build integration` or `-tags=integration`

### End-to-End Tests

- Test complete workflows across multiple services
- May be in root-level test directories or scripts
- Run with: `make test-e2e`
- Requires running services or Docker Compose environment

### Coverage Reports

```bash
# Generate coverage for all services (HTML reports in _output/coverage/)
make test-coverage

# View coverage for specific service
open _output/coverage/agent-manager.html
```

## Common Development Patterns

### Using Common Libraries

Import shared functionality from the `common/` module:

```go
import (
    "github.com/kart-io/k8s-agent/common/cache"
    "github.com/kart-io/k8s-agent/common/client"
    "github.com/kart-io/k8s-agent/common/config"
    "github.com/kart-io/k8s-agent/common/errors"
    "github.com/kart-io/k8s-agent/common/middleware"
)

// Use unified cache interface
memCache := cache.NewMemoryCache(cache.Options{Prefix: "myapp:"})
redisCache := cache.NewRedisCache(redisOptions, cache.Options{Prefix: "myapp:"})

// Use standardized error handling
if err := doSomething(); err != nil {
    return errors.Wrap(err, errors.CodeInternal, "operation failed")
}
```

### Logging Best Practices

The project is transitioning to `github.com/kart-io/logger`:

```go
import "github.com/kart-io/logger"

// Initialize logger (choose engine: Zap or Slog)
log := logger.New(logger.Config{
    Engine: logger.EngineZap,  // or logger.EngineSlog
    Level:  logger.LevelInfo,
    Format: logger.FormatJSON,
})

// Structured logging with fields
log.Info("Agent registered",
    logger.String("agentID", id),
    logger.String("cluster", clusterName),
    logger.Int("version", version),
)

// OTLP integration for traces
log.WithOTLP(otlpConfig).Info("Processing event", fields...)
```

### Protocol Buffer Development

When adding or modifying APIs:

```bash
# 1. Edit .proto files in api/proto/
vim api/proto/orchestrator/workflow.proto

# 2. Regenerate code
make gen-proto

# 3. Generated files appear in:
#    - api/proto/gen/go/<package>/
#    - api/proto/gen/openapiv2/

# 4. Import in Go code
import workflowpb "github.com/kart-io/k8s-agent/api/proto/gen/go/orchestrator"
```

### Adding Middleware

Use common middleware from `common/middleware/`:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/middleware"
    "github.com/kart-io/k8s-agent/common/config"
)

router := gin.New()

// CORS middleware
router.Use(middleware.CORS(config.CORSOptions{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))

// Rate limiting
router.Use(middleware.RateLimit(100, time.Minute))  // 100 req/min
```

### Configuration Management

Follow the standard configuration pattern:

```go
import (
    "github.com/spf13/viper"
    "github.com/kart-io/k8s-agent/common/config"
)

type Config struct {
    Server   config.ServerOptions   `mapstructure:"server"`
    Database config.DatabaseOptions `mapstructure:"database"`
    Redis    config.RedisOptions    `mapstructure:"redis"`
    NATS     config.NATSOptions     `mapstructure:"nats"`
    // Service-specific fields
    MyFeature struct {
        Enabled bool   `mapstructure:"enabled"`
        Timeout int    `mapstructure:"timeout"`
    } `mapstructure:"my_feature"`
}

func LoadConfig(path string) (*Config, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.AutomaticEnv()  // Read from environment variables

    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### Service Entry Point Pattern

Services follow this main.go pattern:

```go
// cmd/<service>/main.go
package main

import (
    "flag"
    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/cmd/<service>/app"
)

func main() {
    configPath := flag.String("c", "configs/config.yaml", "config file path")
    flag.Parse()

    // Initialize logger
    log := logger.New(logger.Config{...})

    // Load config
    cfg, err := app.LoadConfig(*configPath)
    if err != nil {
        log.Fatal("Failed to load config", logger.Error(err))
    }

    // Run application
    if err := app.Run(cfg, log); err != nil {
        log.Fatal("Application failed", logger.Error(err))
    }
}
```

### NATS Messaging Patterns

Use `common/mq/` for NATS messaging:

```go
import (
    "github.com/kart-io/k8s-agent/common/mq"
    "github.com/kart-io/k8s-agent/common/config"
)

// Connect to NATS
nc, err := mq.NewNATSConnection(config.NATSOptions{
    URL:     "nats://localhost:4222",
    Timeout: 10 * time.Second,
})

// Publish event
event := &AgentEvent{...}
if err := nc.Publish("agent.events", event); err != nil {
    log.Error("Failed to publish", logger.Error(err))
}

// Subscribe to events
sub, err := nc.Subscribe("agent.events.*", func(msg *nats.Msg) {
    // Handle message
})
```

## Monitoring and Observability

### Health Checks

All services expose health endpoints:

```bash
curl http://localhost:8080/health  # Agent Manager
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8082/health  # Reasoning Service
```

### Metrics

Services expose Prometheus metrics (where implemented):

```bash
curl http://localhost:8081/metrics  # Orchestrator metrics
```

### Logging

- Structured logging with Logrus
- JSON format for production
- Debug mode available in development configs

## CI/CD

```bash
# Run full CI pipeline (clean, deps, fmt, vet, lint, test, build)
make ci

# Create release
make release VERSION=v1.0.0

# This will: clean, deps, test, build-all, docker-build
```

## Workspace Structure

The repository uses Go workspaces (Go 1.25.0) with internal modules:

```
go.mod (root)                     # Main workspace module
├── replace github.com/kart-io/k8s-agent/common => ./common
├── replace github.com/kart-io/k8s-agent/api/proto => ./api/proto
└── replace github.com/kart-io/logger => ../logger

common/
└── go.mod                        # Common utilities module

api/proto/
└── go.mod                        # Protocol Buffer definitions module
```

### Module Dependencies

- **Root module**: Contains all service code (cmd/, internal/)
- **common/ module**: Shared utilities imported by root module
- **api/proto/ module**: Generated protobuf code imported by root module
- **External logger**: `github.com/kart-io/logger` from parent workspace

### Important Notes

1. **Monorepo Build**: All services built from root directory using modular Makefile
2. **No Service-Level Makefiles**: Build orchestration is centralized in root Makefile
3. **Shared Dependencies**: All services use same dependency versions from root go.mod
4. **Local Development**: Use `make run-<service>` from root, not `cd <service> && make run`

## Quick Start for New Developers

```bash
# 1. Install prerequisites (Go 1.25+, Docker, golangci-lint)
make dev-setup

# 2. Start dependencies (MySQL, Redis, NATS, Neo4j)
cd deployments/docker-compose
docker-compose up -d mysql redis nats neo4j

# 3. Build all services (outputs to _output/bin/)
cd ../..  # Back to root
make build

# 4. Run services (each in separate terminal, all from root directory)
make run-agent-manager
make run-orchestrator
make run-reasoning

# 5. Verify services are healthy
curl http://localhost:8080/health  # Agent Manager
curl http://localhost:8081/health  # Orchestrator
curl http://localhost:8082/health  # Reasoning Service

# 6. (Alternative) Deploy complete stack with Docker Compose
make docker-compose-up
```

## Important Notes

- **Monorepo Structure**: All services in one repository with centralized build system
- **Three-Layer Code Organization**:
  - `common/` - Generic utilities (zero business logic, can be used in ANY project)
  - `pkg/` - Business logic (Aetherius-specific, contains domain models and workflows)
  - `internal/` - Service implementations (private, not exported)
- **Modular Makefile**: Build system uses `scripts/make-rules/*.mk` pattern from OneX project
- **Go 1.25.0**: Workspace requires Go 1.25.0+, uses modern Go features
- **Database**: MySQL 8.0+ (migrated from PostgreSQL)
- **Logging**: Transitioning to `github.com/kart-io/logger` (dual-engine Zap/Slog)
- **Version Injection**: Uses `github.com/kart-io/version` for build-time version information
- **Reasoning Service**: Fully implemented in Go with AI API integration (OpenAI/Gemini/DeepSeek)
- **Configuration**: YAML + Viper with standardized options in `common/config/` (53 config functions)
- **Multi-platform**: Docker builds support linux/amd64 and linux/arm64
- **Authentication**: JWT-based with Redis session management and forced logout
- **Event Flow**: Collect Agent → NATS → Agent Manager → NATS → Orchestrator → HTTP → Reasoning Service
- **Build Outputs**: All binaries output to `_output/bin/`, coverage to `_output/coverage/`
- **No Service-Level Builds**: Always run make commands from repository root
- **Code Reorganization**: See [docs/CODE_REORGANIZATION.md](docs/CODE_REORGANIZATION.md) for migration plan from `internal/pkg/` to `pkg/`

## Documentation

- [README.md](README.md) - Project overview and quick start
- [docs/architecture/SYSTEM_ARCHITECTURE.md](docs/architecture/SYSTEM_ARCHITECTURE.md) - Detailed architecture
- [deployments/docker-compose/README.md](deployments/docker-compose/README.md) - Docker Compose deployment
- [deployments/k8s/README.md](deployments/k8s/README.md) - Kubernetes deployment
- Service-specific READMEs in each service directory

## Feature Implementation Status

**Core Functionality (FR-1 to FR-18)**: 18/18 completed (100%)

- Event reception from Alertmanager/K8s
- Diagnostic task management
- Knowledge base integration (RAG)
- MCP secure execution
- Command analysis and planning
- Diagnostic report generation
- Feedback collection and learning
- Cost control and resource management
- Priority management
- State query and intervention interfaces
- Custom policy configuration
- Historical query and statistics
