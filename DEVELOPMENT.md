# Developer Guide

[简体中文](DEVELOPMENT.zh-CN.md) | [English](DEVELOPMENT.md)

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Quality](#code-quality)
- [Testing](#testing)
- [Building](#building)
- [Docker](#docker)
- [Debugging](#debugging)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **Go**: 1.21 or later
- **Docker**: 20.10 or later
- **Make**: GNU Make 4.0+
- **Git**: 2.30+

### Optional Tools

- **Air**: For hot reload development
- **golangci-lint**: For code linting
- **Buf**: For proto management
- **kubectl**: For Kubernetes deployment

## Project Structure

```
k8s-agent/
├── api/                    # API definitions (proto, OpenAPI)
│   └── proto/              # Protocol Buffer definitions
├── cmd/                    # Service entry points
│   ├── agent-manager/      # Agent Manager main
│   ├── orchestrator-service/ # Orchestrator main
│   └── reasoning-service-go/ # Reasoning Service main
├── pkg/                    # Public libraries
│   ├── client/             # gRPC client libraries
│   ├── errors/             # Error handling
│   └── version/            # Version management
├── internal/               # Private application code
├── scripts/                # Build and automation scripts
│   └── make-rules/         # Modular Makefile rules
├── test/                   # Test infrastructure
│   ├── fixtures/           # Test fixtures and helpers
│   ├── integration/        # Integration tests
│   └── e2e/                # End-to-end tests
├── deployments/            # Deployment configurations
│   ├── docker-compose/     # Docker Compose files
│   └── k8s/                # Kubernetes manifests
├── docs/                   # Documentation
├── githooks/               # Git hooks for code quality
└── Makefile                # Root Makefile
```

## Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/kart-io/k8s-agent.git
cd k8s-agent
```

### 2. Install Development Tools

```bash
# Install all development tools
make dev-setup

# Or install tools individually
make tools.install.golangci-lint
make tools.install.buf
make tools.install.air
```

### 3. Install Git Hooks

```bash
# Install pre-commit and commit-msg hooks
make hooks.install
```

### 4. Start Dependencies

```bash
cd deployments/docker-compose
docker-compose up -d mysql redis nats neo4j
```

### 5. Verify Setup

```bash
# Check tool installation
make tools.verify

# Check database connectivity
docker-compose ps
```

## Development Workflow

### Running Services Locally

#### Option 1: Run with Make (recommended)

```bash
# Run specific service
make run-agent-manager
make run-orchestrator-service
make run-reasoning-service-go
```

#### Option 2: Run with Hot Reload (using Air)

```bash
# Install air if not already installed
make tools.install.air

# Run with hot reload
cd agent-manager
air

# Or use make target
make dev
```

#### Option 3: Run with Go

```bash
cd agent-manager
go run cmd/server/main.go
```

### Code Generation

#### Generate Proto Code

```bash
# Generate all proto code
make proto.generate

# Lint proto files
make proto.lint

# Check for breaking changes
make proto.breaking

# Format proto files
make proto.format
```

### Code Quality Checks

#### Formatting

```bash
# Format all Go code
make go.fmt

# Check if code is formatted
gofmt -l .
```

#### Linting

```bash
# Run all linters
make go.lint

# Run linters with auto-fix
make go.lint.fix

# Run specific linter
golangci-lint run --disable-all --enable=errcheck ./...
```

#### Static Analysis

```bash
# Run go vet
make go.vet

# Run staticcheck
staticcheck ./...
```

### Git Workflow

#### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test changes
- `chore`: Maintenance tasks
- `build`: Build system changes
- `ci`: CI/CD changes

**Examples:**
```bash
git commit -m "feat(agent-manager): add agent health check endpoint"
git commit -m "fix(orchestrator): resolve workflow execution timeout"
git commit -m "docs: update README with installation steps"
```

#### Pre-commit Checks

Git hooks automatically run these checks:
- Go formatting (`gofmt`)
- Trailing whitespace
- `go vet`
- Basic security checks

To bypass hooks (not recommended):
```bash
git commit --no-verify
```

## Testing

### Unit Tests

```bash
# Run all tests
make go.test

# Run tests for specific service
cd agent-manager && go test ./...

# Run tests with coverage
make go.test.coverage

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./pkg/mypackage -run TestMyFunction
```

### Integration Tests

```bash
# Run integration tests
make go.test.integration

# Or with go test
go test -v -tags=integration ./test/integration/...
```

### End-to-End Tests

```bash
# Run E2E tests
make test-e2e

# Or with go test
go test -v -tags=e2e ./test/e2e/...
```

### Writing Tests

#### Unit Test Example

```go
package mypackage

import (
    "testing"
    "github.com/kart-io/k8s-agent/test/fixtures"
)

func TestMyFunction(t *testing.T) {
    // Arrange
    ctx := fixtures.TestContext(t)
    input := "test"

    // Act
    result, err := MyFunction(ctx, input)

    // Assert
    fixtures.AssertNoError(t, err, "MyFunction failed")
    fixtures.AssertEqual(t, "expected", result, "unexpected result")
}
```

#### Integration Test Example

```go
// +build integration

package integration

import (
    "testing"
    "github.com/kart-io/k8s-agent/test/fixtures"
)

func TestServiceIntegration(t *testing.T) {
    ctx := fixtures.TestContext(t)

    // Setup
    client := setupTestClient(t)
    defer client.Close()

    // Test
    result, err := client.DoSomething(ctx)

    // Verify
    fixtures.AssertNoError(t, err, "integration test failed")
    fixtures.AssertNotNil(t, result, "result should not be nil")
}
```

## Building

### Build All Services

```bash
# Build all services
make go.build

# Build specific service
make go.build.agent-manager

# Build with version injection
VERSION=v1.0.0 make go.build
```

### Build for Multiple Platforms

```bash
# Build for all platforms
make go.build.multiarch

# Build for specific platform
GOOS=linux GOARCH=arm64 make go.build
```

### Output

Built binaries are located in `_output/bin/`:
```
_output/bin/
├── agent-manager
├── orchestrator-service
├── reasoning-service-go
└── collect-agent
```

## Docker

### Build Docker Images

```bash
# Build all images
make docker.build

# Build specific image
make docker.build.agent-manager

# Build multi-platform images
make docker.buildx

# Build and push
make docker.buildx.push
```

### Run with Docker Compose

```bash
cd deployments/docker-compose
docker-compose up -d

# View logs
docker-compose logs -f agent-manager

# Stop services
docker-compose down
```

## Debugging

### Enable Debug Logging

```bash
# Set log level in config file
log:
  level: debug

# Or via environment variable
export LOG_LEVEL=debug
```

### Debug with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start service in debug mode
dlv debug ./cmd/agent-manager/main.go

# Or attach to running process
dlv attach <pid>
```

### Common Debug Commands

```bash
# List goroutines
(dlv) goroutines

# Set breakpoint
(dlv) break main.main

# Continue execution
(dlv) continue

# Print variable
(dlv) print myVar

# Stack trace
(dlv) stack
```

## Troubleshooting

### Common Issues

#### 1. Build Failures

**Problem**: `cannot find package`
```bash
# Solution: Update dependencies
make go.mod.download
make go.mod.tidy
```

**Problem**: `protoc-gen-go: command not found`
```bash
# Solution: Install proto tools
make tools.install.protoc-plugins
export PATH="$PATH:$(go env GOPATH)/bin"
```

#### 2. Test Failures

**Problem**: Database connection errors
```bash
# Solution: Start dependencies
cd deployments/docker-compose
docker-compose up -d mysql redis

# Wait for services to be ready
docker-compose ps
```

#### 3. Docker Issues

**Problem**: Port already in use
```bash
# Solution: Stop conflicting services
docker-compose down
lsof -ti:8080 | xargs kill -9
```

**Problem**: Out of disk space
```bash
# Solution: Clean up Docker
docker system prune -a
```

### Getting Help

- Check logs: `make logs`
- View health status: `curl http://localhost:8080/health`
- Check GitHub Issues: https://github.com/kart-io/k8s-agent/issues
- Ask on Discussions: https://github.com/kart-io/k8s-agent/discussions

## Additional Resources

- [Architecture Documentation](docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Proto Buffer Guide](docs/devel/proto-buf-guide.md)
- [Implementation Guide](docs/devel/implementation-guide.md)
- [Improvement Plan](docs/architecture/IMPROVEMENT_PLAN.md)

---

Happy coding! 🚀
