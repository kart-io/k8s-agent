# Project Restructuring Plan

## Overview

This document outlines the plan to reorganize the k8s-agent project to follow the [onex v2](https://github.com/onexstack/onex/tree/feature/onex-v2) monorepo structure pattern.

## Current Structure

```
k8s-agent/
├── agent-manager/          # Standalone service directory
├── orchestrator-service/   # Standalone service directory
├── reasoning-service-go/   # Standalone service directory
├── auth-service/           # Standalone service directory
├── gateway-service/        # Standalone service directory
├── monitor-service/        # Standalone service directory
├── cluster-service/        # Standalone service directory
├── collect-agent/          # Standalone service directory
├── common/                 # Shared libraries
├── protos/                 # Proto definitions
├── deployments/            # K8s manifests
├── docs/                   # Documentation
└── scripts/                # Utility scripts
```

### Issues with Current Structure:
1. **Fragmented Service Layout**: Each service in its own top-level directory makes monorepo management difficult
2. **No Unified Build System**: Each service has its own Makefile and build process
3. **Inconsistent Configuration Management**: Configs scattered across services
4. **No Centralized Entry Points**: No unified `cmd/` directory
5. **Mixed Internal/Public Code**: Unclear package boundaries

## Target Structure (onex v2 Pattern)

```
k8s-agent/
├── cmd/                          # Service entry points (binaries)
│   ├── agent-manager/
│   │   ├── main.go
│   │   └── app/
│   ├── orchestrator/
│   ├── reasoning/
│   ├── auth/
│   ├── gateway/
│   ├── monitor/
│   ├── cluster/
│   └── collect-agent/
│
├── internal/                     # Private application code
│   ├── agent-manager/           # agent-manager specific logic
│   │   ├── grpc/
│   │   ├── api/
│   │   ├── agent/
│   │   ├── command/
│   │   ├── event/
│   │   ├── storage/
│   │   └── nats/
│   ├── orchestrator/            # orchestrator specific logic
│   ├── reasoning/               # reasoning service specific logic
│   ├── auth/                    # auth service specific logic
│   ├── gateway/                 # gateway specific logic
│   ├── monitor/                 # monitor specific logic
│   ├── cluster/                 # cluster service specific logic
│   ├── collect-agent/           # collect-agent specific logic
│   └── pkg/                     # Shared internal utilities
│       ├── db/
│       ├── cache/
│       └── messaging/
│
├── pkg/                         # Public reusable packages
│   ├── types/                   # Common types
│   ├── client/                  # Client libraries
│   │   ├── agentmanager/
│   │   └── orchestrator/
│   ├── logger/                  # Logging utilities
│   └── util/                    # Common utilities
│
├── api/                         # API definitions
│   └── proto/                   # Proto files (from protos/)
│       ├── agentmanager/
│       │   ├── agent/v1/
│       │   ├── command/v1/
│       │   └── event/v1/
│       ├── orchestrator/
│       └── common/
│
├── build/                       # Build-related files
│   ├── docker/                  # Dockerfiles for each service
│   │   ├── agent-manager.Dockerfile
│   │   ├── orchestrator.Dockerfile
│   │   └── ...
│   └── scripts/                 # Build scripts
│       ├── build.sh
│       └── version.sh
│
├── configs/                     # Configuration templates
│   ├── agent-manager/
│   │   ├── config.yaml
│   │   ├── config-dev.yaml
│   │   └── config-prod.yaml
│   ├── orchestrator/
│   └── ...
│
├── manifests/                   # Kubernetes deployment manifests
│   ├── base/                    # Base manifests
│   │   ├── agent-manager/
│   │   ├── orchestrator/
│   │   └── ...
│   └── overlays/                # Environment-specific overlays
│       ├── dev/
│       ├── staging/
│       └── prod/
│
├── test/                        # Test utilities and fixtures
│   ├── integration/
│   ├── e2e/
│   └── fixtures/
│
├── tools/                       # Development and utility tools
│   ├── codegen/
│   └── migration/
│
├── examples/                    # Usage examples
│
├── docs/                        # Documentation
│   ├── architecture/
│   ├── api/
│   └── guides/
│
├── scripts/                     # Utility scripts
│   ├── setup/
│   ├── test/
│   └── deploy/
│
├── Makefile                     # Root Makefile (orchestrates all)
├── go.mod                       # Root go module
├── go.work                      # Go workspace (optional)
└── README.md                    # Project README
```

## Benefits of New Structure

### 1. **Monorepo Best Practices**
- Single source of truth for all services
- Atomic commits across services
- Easier refactoring and dependency management

### 2. **Clear Package Boundaries**
- `cmd/`: Entry points only
- `internal/`: Private implementation (not importable externally)
- `pkg/`: Public, reusable code
- `api/`: API contracts

### 3. **Unified Build System**
- Single root Makefile
- Consistent build commands: `make build BINS=agent-manager`
- Environment-specific builds: `make docker ENV=prod`

### 4. **Simplified Deployment**
- Centralized Kubernetes manifests
- Kustomize-based overlays for environments
- Single source for deployment configs

### 5. **Better Developer Experience**
- Consistent directory layout across services
- Clear navigation: `cmd/` for entry, `internal/` for logic
- Easier onboarding with standardized structure

## Migration Steps

### Phase 1: Setup New Structure (Non-Breaking)
1. ✅ Create new directory structure
2. Create migration scripts in `tools/migration/`
3. Create new root Makefile
4. Update go.mod import paths planning document

### Phase 2: Move API Definitions
1. Move `protos/` to `api/proto/`
2. Update protobuf generation scripts
3. Update import paths in all services
4. Test proto generation

### Phase 3: Move Service Entry Points
1. Move `agent-manager/cmd/` to `cmd/agent-manager/`
2. Update import paths in main.go and app/
3. Move orchestrator, reasoning, etc.
4. Test each service builds

### Phase 4: Reorganize Internal Code
1. Move `agent-manager/internal/` to `internal/agent-manager/`
2. Move `common/` shared code to `internal/pkg/` or `pkg/`
3. Update all internal imports
4. Repeat for all services

### Phase 5: Consolidate Build System
1. Move Dockerfiles to `build/docker/`
2. Create unified build scripts in `build/scripts/`
3. Update CI/CD pipelines
4. Test Docker builds

### Phase 6: Centralize Configurations
1. Move all service configs to `configs/{service}/`
2. Standardize config file naming
3. Update config loading paths in code

### Phase 7: Reorganize Deployments
1. Move `deployments/k8s/` to `manifests/base/`
2. Create Kustomize overlays for dev/staging/prod
3. Update deployment documentation
4. Test K8s deployments

### Phase 8: Move Public Packages
1. Move `agent-manager/pkg/types/` to `pkg/types/`
2. Move reusable client code to `pkg/client/`
3. Update import paths
4. Test package imports

### Phase 9: Verification
1. Build all services: `make build`
2. Run all tests: `make test`
3. Test Docker builds: `make docker`
4. Deploy to dev environment
5. Run integration tests

### Phase 10: Cleanup
1. Remove old service directories
2. Update all documentation
3. Update README.md
4. Update CLAUDE.md
5. Create migration guide for contributors

## Implementation Commands

### Create Directory Structure
```bash
# Create cmd structure
mkdir -p cmd/{agent-manager,orchestrator,reasoning,auth,gateway,monitor,cluster,collect-agent}

# Create internal structure
mkdir -p internal/{agent-manager,orchestrator,reasoning,auth,gateway,monitor,cluster,collect-agent,pkg}

# Create api structure
mkdir -p api/proto

# Create build structure
mkdir -p build/{docker,scripts}

# Create manifests structure
mkdir -p manifests/{base,overlays/{dev,staging,prod}}

# Create other directories
mkdir -p tools/migration test/{integration,e2e,fixtures}
```

### Example Migration Script
```bash
#!/bin/bash
# tools/migration/move-service.sh

SERVICE=$1

echo "Migrating $SERVICE..."

# Move cmd
cp -r ${SERVICE}/cmd/* cmd/${SERVICE}/

# Move internal
cp -r ${SERVICE}/internal/* internal/${SERVICE}/

# Move configs
cp -r ${SERVICE}/configs/* configs/${SERVICE}/

# Move Dockerfile
cp ${SERVICE}/Dockerfile build/docker/${SERVICE}.Dockerfile

echo "Migration complete. Verify and test before removing old directories."
```

## Import Path Changes

### Before
```go
import (
    "github.com/kart-io/k8s-agent/agent-manager/internal/agent"
    "github.com/kart-io/k8s-agent/agent-manager/pkg/types"
)
```

### After
```go
import (
    "github.com/kart-io/k8s-agent/internal/agent-manager/agent"
    "github.com/kart-io/k8s-agent/pkg/types"
)
```

## Makefile Structure

### Root Makefile
```makefile
# Root Makefile - orchestrates all services

# Service list
SERVICES := agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent

# Default target
.PHONY: all
all: build

# Build all services
.PHONY: build
build:
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		go build -o bin/$$service ./cmd/$$service; \
	done

# Build specific service
.PHONY: build-%
build-%:
	go build -o bin/$* ./cmd/$*

# Docker build all
.PHONY: docker
docker:
	@for service in $(SERVICES); do \
		docker build -f build/docker/$$service.Dockerfile -t $$service:latest .; \
	done

# Docker build specific service
.PHONY: docker-%
docker-%:
	docker build -f build/docker/$*.Dockerfile -t $*:latest .

# Run tests
.PHONY: test
test:
	go test ./...

# Clean
.PHONY: clean
clean:
	rm -rf bin/
```

## Rollback Plan

If migration causes issues:

1. **Immediate Rollback**: Keep old structure until verification
2. **Gradual Migration**: Migrate one service at a time
3. **Feature Flags**: Use build tags to support both structures temporarily
4. **Documentation**: Maintain compatibility guide during transition

## Timeline

- **Phase 1-2**: Day 1 (Setup + API)
- **Phase 3-4**: Day 2-3 (Entry points + Internal)
- **Phase 5-7**: Day 4-5 (Build + Config + Deploy)
- **Phase 8-9**: Day 6 (Packages + Verification)
- **Phase 10**: Day 7 (Cleanup)

**Total Estimate**: 1 week with proper testing

## References

- [onex v2 Structure](https://github.com/onexstack/onex/tree/feature/onex-v2)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Monorepo Best Practices](https://monorepo.tools/)

## Next Steps

1. ✅ Create directory structure
2. Create migration scripts
3. Get team approval
4. Start Phase 2 (API migration)
