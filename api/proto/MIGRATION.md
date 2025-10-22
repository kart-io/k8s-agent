# Proto Migration Guide

## Overview

All Protocol Buffer (protobuf) files have been migrated to a centralized `protos/` directory in the project root. This provides better organization and centralized management of all API definitions.

## Changes

### Before Migration

```
k8s-agent/
├── agent-manager/
│   └── proto/
│       └── agentmanager/
│           ├── agent/v1/agent.proto
│           ├── command/v1/command.proto
│           └── event/v1/event.proto
└── common/
    └── proto/
        └── common/
            ├── example/v1/example.proto
            └── health/v1/health.proto
```

### After Migration

```
k8s-agent/
└── protos/                    # NEW: Centralized proto directory
    ├── agentmanager/
    │   ├── agent/v1/agent.proto
    │   ├── command/v1/command.proto
    │   └── event/v1/event.proto
    ├── common/
    │   ├── example/v1/example.proto
    │   └── health/v1/health.proto
    ├── gen/                   # Generated Go code
    ├── Makefile              # Build automation
    ├── README.md             # Documentation
    └── .gitignore
```

## Go Package Path Changes

All proto files' `go_package` option has been updated:

**Before:**
```protobuf
option go_package = "github.com/kart-io/k8s-agent/agent-manager/proto/gen/agentmanager/agent/v1;agentv1";
option go_package = "github.com/kart-io/k8s-agent/common/proto/gen/common/example/v1;examplev1";
```

**After:**
```protobuf
option go_package = "github.com/kart-io/k8s-agent/protos/gen/agentmanager/agent/v1;agentv1";
option go_package = "github.com/kart-io/k8s-agent/protos/gen/common/example/v1;examplev1";
```

## How to Use

### 1. Generate Proto Files

```bash
cd protos
make gen-go
```

This generates Go code in `protos/gen/` directory.

### 2. Update Import Statements

In your Go files, update import paths:

**Before:**
```go
import (
    agentv1 "github.com/kart-io/k8s-agent/agent-manager/proto/gen/agentmanager/agent/v1"
    commandv1 "github.com/kart-io/k8s-agent/agent-manager/proto/gen/agentmanager/command/v1"
    eventv1 "github.com/kart-io/k8s-agent/agent-manager/proto/gen/agentmanager/event/v1"
)
```

**After:**
```go
import (
    agentv1 "github.com/kart-io/k8s-agent/protos/gen/agentmanager/agent/v1"
    commandv1 "github.com/kart-io/k8s-agent/protos/gen/agentmanager/command/v1"
    eventv1 "github.com/kart-io/k8s-agent/protos/gen/agentmanager/event/v1"
)
```

### 3. Run Tests

After updating imports:

```bash
# In agent-manager
cd agent-manager
go mod tidy
go build ./...
go test ./...

# In common
cd common
go mod tidy
go build ./...
go test ./...
```

## Benefits

1. **Centralized Management**: All proto files in one location
2. **Clear Ownership**: Easy to see all API definitions
3. **Simplified Builds**: Single Makefile for all protos
4. **Better Organization**: Clear separation from service code
5. **Easier Versioning**: Centralized version management
6. **Cross-Service Sharing**: Easier to share protos between services

## Makefile Commands

```bash
# Show all available commands
make help

# Generate all proto files
make gen-go

# Generate specific services
make gen-agentmanager
make gen-common

# Validate proto syntax
make validate

# List all proto files
make list

# Clean generated files
make clean

# Install required tools
make install-tools
```

## Migration Checklist

- [x] Create centralized `protos/` directory
- [x] Copy all proto files to new location
- [x] Update `go_package` paths in all proto files
- [x] Create unified Makefile for proto generation
- [x] Create comprehensive README
- [x] Create .gitignore for generated files
- [ ] Update all service imports to use new paths
- [ ] Update all service Makefiles to reference protos/
- [ ] Test all services build successfully
- [ ] Update CI/CD pipelines
- [ ] Update documentation

## Next Steps

1. **Update Service Imports**: Update all Go files in services to use new import paths
2. **Update Service Makefiles**: Update service-specific Makefiles to use centralized protos
3. **Remove Old Proto Directories**: After verification, remove old proto directories:
   - `agent-manager/proto/`
   - `common/proto/`
4. **Update CI/CD**: Update build pipelines to generate protos from new location
5. **Update Documentation**: Update architecture and API documentation

## Rollback (if needed)

If you need to rollback:

1. Keep the old proto directories as they are
2. Continue using old import paths
3. The new `protos/` directory can coexist without breaking existing code

## Questions?

See `protos/README.md` for detailed usage instructions.
