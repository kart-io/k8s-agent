# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Aetherius** is an enterprise-grade intelligent Kubernetes operations platform (智能 Kubernetes 运维平台) with AI-driven fault diagnosis and automated remediation. The system uses a 4-layer architecture combining event-driven design with AI technology to create a complete operational loop from data collection to intelligent analysis.

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
- **Tech**: Go 1.21+, client-go, NATS
- **Functions**: K8s event monitoring (85+ event types), resource metrics collection, secure command execution
- **Location**: `collect-agent/`

### Layer 2: Agent Manager

- **Purpose**: Central control plane managing all agents
- **Tech**: Go 1.21+, MySQL, Redis, NATS, Gin
- **Functions**: Agent lifecycle management, event aggregation/routing, command scheduling, multi-cluster management
- **API Port**: 8080
- **Location**: `agent-manager/`

### Layer 3: Orchestrator Service

- **Purpose**: Workflow orchestration for automated diagnosis and remediation
- **Tech**: Go 1.21+, MySQL, Redis, NATS
- **Functions**: Workflow engine, diagnostic strategies, 6 step types (Command/AI/Decision/Remediation/Notification/Wait), AI integration
- **API Port**: 8081
- **Location**: `orchestrator-service/`

### Layer 4: Reasoning Service

- **Purpose**: AI-driven root cause analysis and intelligent recommendations
- **Tech**: Go 1.24+, Gin, Neo4j, OpenAI/Gemini/DeepSeek API
- **Functions**: Root cause analysis, recommendation engine (30+ rules), prediction engine, knowledge graph (Neo4j), continuous learning
- **API Port**: 8082
- **Location**: `reasoning-service-go/`

### Supporting Services

- **Auth Service**: JWT authentication, session management, forced logout functionality
  - Tech: Go 1.21, Gin, JWT, Redis, MySQL
  - Location: `auth-service/`

- **Gateway Service**: API gateway with Traefik integration
  - Location: `gateway-service/`

- **Monitor Service**: Monitoring and metrics collection
  - Location: `monitor-service/`

- **Cluster Service**: Multi-cluster management
  - Location: `cluster-service/`

- **Common**: Shared libraries and utilities
  - Location: `common/`

## Common Development Commands

### Build and Run Commands

```bash
# Build all services
make build

# Build specific service
cd agent-manager && make build
cd orchestrator-service && make build
cd reasoning-service-go && make build
cd auth-service && make build

# Run all services with Docker Compose (recommended for development)
make docker-compose-up

# Run specific service in development mode
cd agent-manager && make run-dev
cd orchestrator-service && make run-dev
cd reasoning-service-go && make run-dev
cd auth-service && make run-dev

# Run with hot reload (requires air)
cd agent-manager && make dev
cd orchestrator-service && make dev
```

### Testing Commands

```bash
# Test all services
make test

# Test specific service
cd agent-manager && make test
cd orchestrator-service && make test
cd reasoning-service-go && make test

# Test with coverage
make test-coverage
cd <service> && make test-coverage

# Run integration tests
cd orchestrator-service && make test-integration
```

### Code Quality Commands

```bash
# Format all code
make fmt

# Format specific service
cd <service> && make fmt

# Run linters
make lint
cd <service> && make lint

# Run go vet
make vet
cd <service> && make vet
```

### Docker Commands

```bash
# Build all Docker images
make docker-build VERSION=v1.0.0

# Build multi-platform images (linux/amd64,linux/arm64)
make docker-buildx VERSION=v1.0.0

# Build and push multi-platform images
make docker-buildx-push VERSION=v1.0.0

# Build specific service image
cd <service> && make docker-build

# Build with environment tag
cd <service> && make docker-buildx-env ENV=dev
cd <service> && make docker-buildx-push-env ENV=dev
```

### Development Environment

```bash
# Setup development environment (install Go/Python tools)
make dev-setup

# Start all services in development mode
make dev-start

# Stop all development services
make dev-stop

# Setup databases only (MySQL, Redis, NATS, Neo4j)
make db-setup

# Reset all databases
make db-reset

# Check service health
make health-check

# View logs
make logs
```

### Kubernetes Deployment

```bash
# Deploy all services to Kubernetes
make k8s-deploy

# Delete all services from Kubernetes
make k8s-delete

# Show Kubernetes status
make k8s-status

# Show logs for specific service
make k8s-logs SERVICE=agent-manager

# Restart all deployments
make k8s-restart

# Deploy with Kustomize (Traefik, Prometheus, Grafana)
cd deployments/kustomize && make deploy-all
```

### Database Operations

```bash
# Connect to MySQL
docker-compose -f deployments/docker-compose/docker-compose.yaml exec mysql mysql -u aetherius -p

# Connect to Redis
make redis-cli

# Create database in K8s
cd agent-manager && make db-create
cd orchestrator-service && make db-create

# Create database locally
cd agent-manager && make db-create-local
```

## Project Structure

All Go services follow a consistent internal structure:

```
<service>/
├── cmd/
│   └── server/          # Main application entry point
├── internal/            # Private application code
│   ├── api/             # HTTP API handlers
│   ├── config/          # Configuration management
│   ├── storage/         # Database/storage layer
│   └── <domain>/        # Domain-specific packages
├── pkg/                 # Public packages (can be imported by other services)
├── configs/             # Configuration files
│   ├── config.yaml      # Default configuration
│   ├── config-dev.yaml  # Development configuration
│   ├── config-test.yaml # Test configuration
│   └── config-prod.yaml # Production configuration
├── deployments/         # Deployment manifests
├── Makefile             # Build automation
├── Dockerfile           # Container build
└── README.md            # Service documentation
```

## Technology Stack

### Backend (Go Services)

- **Go Version**: 1.21+ (most services), 1.24+ (reasoning-service-go)
- **Web Framework**: Gin v1.9.1
- **Messaging**: NATS 2.10+
- **Database**: MySQL 8.0+ (migrated from PostgreSQL)
- **Cache**: Redis 6+
- **ORM**: GORM (for services using it)
- **Logging**: Logrus (structured logging)
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **AI Integration**: OpenAI/Gemini/DeepSeek API (reasoning-service-go)

### Infrastructure

- **Container**: Docker 20.10+
- **Orchestration**: Kubernetes 1.23+
- **Service Mesh**: Traefik (in gateway-service)
- **Monitoring**: Prometheus + Grafana
- **Tracing**: OpenTelemetry (planned)

## Configuration Management

All services use YAML configuration files with environment variable overrides:

```bash
# Run with default config (configs/config.yaml)
go run cmd/server/main.go

# Run with specific config
go run cmd/server/main.go -config configs/config-dev.yaml
go run cmd/server/main.go -c configs/config-dev.yaml

# Or use Makefile
make run        # Uses default config
make run-dev    # Uses config-dev.yaml
```

Configuration priority (highest to lowest):
1. Environment variables
2. Command-line flags
3. Configuration file
4. Default values

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
- Located alongside implementation files (e.g., `service_test.go`)
- Run with: `go test ./...`

### Integration Tests

- Test service interactions
- Located in `test/integration/` (orchestrator-service)
- Run with: `make test-integration`

### End-to-End Tests

- Test complete workflows
- Scripts in root directory: `test-end-to-end.sh`, `verify-orchestrator.sh`
- Requires running services

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

The repository is a Go workspace with multiple modules:

```
k8s-agent/
├── go.mod                    # Workspace module (Go 1.24.0)
├── Makefile                  # Root orchestration
├── deployments/              # Docker Compose & K8s manifests
│   ├── docker-compose/
│   ├── k8s/
│   └── kustomize/
├── agent-manager/            # Go module
├── orchestrator-service/     # Go module
├── reasoning-service-go/     # Go module (AI智能层)
├── auth-service/             # Go module
├── gateway-service/          # Go module
├── monitor-service/          # Go module
├── cluster-service/          # Go module
├── collect-agent/            # Go module
├── common/                   # Go module (shared libraries)
└── docs/                     # Architecture documentation
```

## Quick Start for New Developers

```bash
# 1. Install prerequisites (Go 1.21+, Docker)
make dev-setup

# 2. Start dependencies (MySQL, Redis, NATS, Neo4j)
cd deployments/docker-compose
docker-compose up -d mysql redis nats neo4j

# 3. Run services in separate terminals
cd agent-manager && make run-dev
cd orchestrator-service && make run-dev
cd reasoning-service-go && make run-dev

# 4. Verify services are healthy
make health-check

# 5. (Optional) Deploy complete stack with Docker Compose
make docker-compose-up
```

## Important Notes

- **Database Migration**: Project has migrated from PostgreSQL to MySQL 8.0+
- **Reasoning Service**: Fully implemented in Go (`reasoning-service-go/`) with AI API integration (OpenAI/Gemini/DeepSeek)
- **Configuration Files**: Always use `-c` or `-config` flag for non-default configs
- **Multi-platform Support**: All Docker builds support linux/amd64 and linux/arm64
- **Authentication**: Auth service provides JWT authentication and forced logout capabilities
- **Event Flow**: Collect Agent → Agent Manager → Orchestrator → Reasoning Service

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
