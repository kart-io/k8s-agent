# Quick Start Guide for Restructured Project

## Overview

This project has been restructured to follow the [onex v2](https://github.com/onexstack/onex/tree/feature/onex-v2) monorepo pattern for better organization and maintainability.

## New Structure Benefits

✅ **Unified Build System** - Single Makefile for all services
✅ **Clear Package Boundaries** - Separation of cmd/, internal/, pkg/, api/
✅ **Easier Navigation** - Consistent structure across services
✅ **Better CI/CD** - Simplified build and deployment pipelines
✅ **Monorepo Best Practices** - Industry-standard layout

## Directory Structure

```
k8s-agent/
├── cmd/                    # Service entry points
│   ├── agent-manager/
│   ├── orchestrator/
│   ├── reasoning/
│   ├── auth/
│   ├── gateway/
│   ├── monitor/
│   ├── cluster/
│   └── collect-agent/
├── internal/               # Private packages (service-specific)
├── pkg/                    # Public packages (reusable)
├── api/                    # API definitions (proto files)
├── build/                  # Build scripts and Dockerfiles
├── configs/                # Configuration templates
├── manifests/              # Kubernetes deployment manifests
├── test/                   # Tests
├── tools/                  # Development tools
├── examples/               # Usage examples
└── docs/                   # Documentation
```

## Quick Commands

### Building Services

```bash
# Build all services
make build

# Build specific service
make build BINS=agent-manager

# Build multiple services
make build BINS="agent-manager orchestrator"

# Build specific service (shorthand)
make build-agent-manager
```

### Running Services

```bash
# Run specific service
make run-agent-manager
make run-orchestrator

# Or use go run directly
go run ./cmd/agent-manager/main.go
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run go vet
make vet

# Run all checks (CI)
make ci
```

### Docker

```bash
# Build all Docker images
make docker

# Build specific image
make docker BINS=agent-manager

# Build and push
make docker
make docker-push

# Build specific image (shorthand)
make docker-agent-manager
```

### Deployment

```bash
# Deploy to development
make deploy ENV=dev

# Deploy to staging
make deploy ENV=staging

# Deploy to production
make deploy ENV=prod

# Validate manifests
make manifests-validate
```

### Dependencies

```bash
# Download dependencies
make deps

# Verify dependencies
make deps-verify
```

### Code Generation

```bash
# Generate all code
make gen

# Generate protobuf code
make gen-proto
```

### Development Setup

```bash
# Install development tools
make dev-setup
```

### Cleanup

```bash
# Clean build artifacts
make clean

# Clean everything
make clean-all
```

## Migration from Old Structure

If you're working with code from the old structure:

### Import Path Changes

**Before:**
```go
import (
    "github.com/kart-io/k8s-agent/agent-manager/internal/agent"
    "github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)
```

**After:**
```go
import (
    "github.com/kart-io/k8s-agent/internal/agent-manager/agent"
    "github.com/kart-io/k8s-agent/pkg/types"
)
```

### Using the Migration Script

To migrate a service from old structure to new:

```bash
# Migrate a service
make migrate SERVICE=agent-manager

# Or run script directly
./tools/migration/migrate-service.sh agent-manager
```

## Service-Specific Notes

### Agent Manager
```bash
# Build
make build-agent-manager

# Run
make run-agent-manager

# Docker
make docker-agent-manager
```

### Orchestrator
```bash
# Build
make build-orchestrator

# Run
make run-orchestrator

# Docker
make docker-orchestrator
```

### Reasoning Service
```bash
# Build
make build-reasoning

# Run
make run-reasoning

# Docker
make docker-reasoning
```

## Environment Variables

Common environment variables:

```bash
# Version
VERSION=v1.0.0 make build

# Docker registry
DOCKER_REGISTRY=my-registry.io make docker

# Docker namespace
DOCKER_NAMESPACE=my-namespace make docker

# Image tag
IMAGE_TAG=latest make docker

# Environment for deployment
ENV=dev make deploy
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Test

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run CI
        run: make ci

      - name: Build all services
        run: make build-all

      - name: Build Docker images
        run: make docker
```

### GitLab CI Example

```yaml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  script:
    - make ci

build:
  stage: build
  script:
    - make build-all
    - make docker

deploy:
  stage: deploy
  script:
    - make deploy ENV=$CI_ENVIRONMENT_NAME
  only:
    - main
```

## Troubleshooting

### Build Issues

**Problem**: `package not found`
```bash
# Solution: Update dependencies
make deps
```

**Problem**: `undefined: types.Something`
```bash
# Solution: Check import paths, update to new structure
# Old: github.com/kart-io/k8s-agent/agent-manager/pkg/types
# New: github.com/kart-io/k8s-agent/pkg/types
```

### Docker Issues

**Problem**: `Dockerfile not found`
```bash
# Solution: Dockerfiles are now in build/docker/
# Use make docker instead of manual docker build
make docker BINS=agent-manager
```

### Deployment Issues

**Problem**: `manifests not found`
```bash
# Solution: Manifests are now in manifests/
# Use make deploy instead of kubectl apply
make deploy ENV=dev
```

## Getting Help

```bash
# Show all available commands
make help

# Show project information
make info

# Show version information
make version
```

## Best Practices

1. **Always use Make commands** instead of manual go/docker commands
2. **Update import paths** when moving code
3. **Run tests** before committing: `make ci`
4. **Format code** before committing: `make fmt`
5. **Use environment overlays** for deployments
6. **Keep cmd/ minimal** - put logic in internal/
7. **Use pkg/** only for truly reusable code

## Additional Resources

- [Full Restructuring Plan](./RESTRUCTURING_PLAN.md)
- [Migration Script](../tools/migration/migrate-service.sh)
- [onex v2 Reference](https://github.com/onexstack/onex/tree/feature/onex-v2)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## Next Steps

1. ✅ Review this guide
2. ✅ Try building a service: `make build-agent-manager`
3. ✅ Run tests: `make test`
4. ✅ Build Docker image: `make docker-agent-manager`
5. ✅ Deploy to dev: `make deploy ENV=dev`

For questions or issues, check the main documentation or create an issue.
