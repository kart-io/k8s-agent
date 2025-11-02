# PKG Directory Analysis Report

## Current State of pkg/ Directory

### Actual Current Contents
The `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg` directory currently contains **ONLY**:

1. **pkg/types/** - Business domain models
   - Single file: `types.go` (295 lines)
   - Contains: Agent, Event, Command, Metrics, Cluster, Alert, CommandResult, InternalEvent, HealthStatus, Config structs
   - Status: ACTIVELY USED across multiple services

2. **pkg/api/** - Protocol Buffer definitions and generated code
   - Subdirectories:
     - `api/agent/v1/` - Agent Manager API (proto + generated)
     - `api/orchestrator/v1/` - Orchestrator API (proto + generated)
     - `api/reasoning/v1/` - Reasoning Service API (proto + generated)
     - `api/common/{health,error,pagination}/v1/` - Common API messages
     - `api/docs/swagger/` - Swagger documentation
   - Contains: .proto files + generated .pb.go, .pb.gw.go, .pb.http.go files
   - Status: ACTIVELY USED, auto-generated from protobuf

### What's MISSING from pkg/ (According to CLAUDE.md)
The CLAUDE.md indicates pkg/ should have been populated with these packages (from internal/pkg/):
- `bootstrap/` - Application bootstrapping logic
- `contextx/` - Project-specific context management
- `idempotent/` - Business idempotency handling
- `metrics/` - Project-specific Prometheus metrics

BUT these packages are CURRENTLY IN common/, not moved yet:
- `/common/bootstrap/` 
- `/common/contextx/`
- `/common/idempotent/`
- `/common/metrics/`

## Analysis of Current common/ vs Desired pkg/

### Currently in common/ that should be in pkg/

| Package | Current Location | Should Move? | Reason |
|---------|------------------|-------------|--------|
| bootstrap | common/bootstrap | YES | App initialization logic - project-specific |
| contextx | common/contextx | YES | K8s agent context - project-specific |
| idempotent | common/idempotent | YES | Business idempotency - project-specific |
| metrics | common/metrics | YES | Project-specific metrics - NOT generic |
| app | common/app | YES | App startup logic - project-specific |
| k8sutils | common/k8sutils | PARTIAL | Keep generic K8s utils, move business logic |

### Currently in common/ that should STAY

| Package | Location | Reason |
|---------|----------|--------|
| cache | common/cache | Generic caching interface (memory/Redis) |
| client | common/client | Generic client implementations |
| config/options | common/options | Generic Options pattern config |
| db | common/db | Generic database client wrappers |
| errors | common/errors | Generic error handling |
| logger | common/logger | Generic logging (deprecated for kart-io/logger) |
| middleware | common/middleware | Generic HTTP middleware |
| mq | common/mq | Generic NATS/message queue |
| pagination | common/pagination | Generic pagination utilities |
| response | common/response | Generic API response format |
| server | common/server | Generic HTTP/gRPC server |
| validator | common/validator | Generic data validation |
| initializers | common/initializers | Generic component initialization |
| serializers | common/serializers | Generic data serialization |
| utils | common/utils | Generic utility functions |

## Currently in pkg/types - Classification

### File: pkg/types/types.go (295 lines)

**Content**: Business domain models for the Aetherius project
- Agent, AgentStatus, ConnectionInfo
- Event, InternalEvent, InternalEventType
- Metrics, Command, CommandStatus, CommandResult
- Cluster, ClusterStatus, ClusterHealth
- AlertRule, Alert, AlertStatus
- HealthStatus
- Config structures (Server, NATS, Database, Redis, Logging, Metrics)

**Classification**: ✅ SHOULD STAY IN pkg/
**Reason**: 
- Pure business domain models specific to Aetherius
- No business logic, just data structures
- Imported by internal services and handlers
- Project-specific, not reusable in other Go projects
- Currently used by: agent-manager, config, grpc, storage, nats, api, command, event handlers

## Currently in pkg/api/ - Classification

### Files: Protocol Buffer definitions

**Content**:
- Agent Manager APIs (agent/command)
- Orchestrator APIs (workflow)
- Reasoning Service APIs (analysis)
- Common shared messages (health, pagination, error)

**Classification**: ✅ SHOULD STAY IN pkg/api/
**Reason**:
- Generated code from protobuf definitions
- Service contract definitions
- Inter-service communication schemas
- Generated files should be in pkg/ (not moved to common/)
- Part of project's API layer

## Conflicts & Issues Found

### 1. Import Paths Not Updated
According to CLAUDE.md, bootstrap, contextx, idempotent, metrics should be in pkg/, but they're still in common/. Import paths in code would need updating.

### 2. common/metrics vs pkg/metrics
- common/metrics exists (generic Prometheus metrics)
- Docs say metrics should be in pkg/ (project-specific metrics)
- Need to clarify: generic vs project-specific metrics

### 3. common/app vs pkg/app
- common/app exists and contains application startup logic
- CLAUDE.md says app should be in pkg/
- Contains RunWithRunner, RunWithOptions, bootstrap patterns
- Partially generic (Runner pattern), partially project-specific

## Migration Assessment

### Packages Ready to Move (HIGH PRIORITY)
1. ✅ **bootstrap** (common/bootstrap → pkg/bootstrap)
   - Status: Pure project-specific application bootstrapping
   - Files: bootstrap.go, helpers.go, bootstrap_test.go
   - Usage: Internal project use only
   - Estimated effort: Low (1-2 hours)

2. ✅ **contextx** (common/contextx → pkg/contextx)
   - Status: K8s-agent specific context management
   - Files: context.go, timeout.go, k8sagent_test.go, context_test.go
   - Usage: Internal project use only
   - Estimated effort: Low (1-2 hours)

3. ✅ **idempotent** (common/idempotent → pkg/idempotent)
   - Status: Business idempotency logic
   - Files: idempotent.go, memory_store.go, redis_store.go, idempotent_test.go
   - Usage: Internal project use only
   - Estimated effort: Low (1-2 hours)

### Packages Requiring Analysis (MEDIUM PRIORITY)
4. ⚠️ **metrics** (common/metrics - HYBRID)
   - registry.go, metrics.go
   - Issue: Could be generic (Prometheus metrics) or project-specific
   - Recommended: Keep generic parts in common/, move project-specific to pkg/
   - Estimated effort: Medium (2-3 hours)

5. ⚠️ **app** (common/app - MIXED)
   - Contains: runner.go, bootstrap_app.go, interfaces.go, health.go, app.go, middleware.go
   - Hybrid: RunWithRunner/RunWithOptions patterns are reusable but app startup is project-specific
   - Recommended: Refactor to extract generic patterns to common/, move project logic to pkg/
   - Estimated effort: High (4-6 hours)

6. ⚠️ **k8sutils** (common/k8sutils - PARTIAL)
   - Review K8s resource conversion utilities
   - Separate generic K8s utilities from business logic
   - Estimated effort: Medium (2-3 hours)

### Keep in common/ (NO ACTION NEEDED)
- cache, client, config, db, errors, logger, middleware, mq, pagination, response, server, validator, initializers, serializers, utils

## pkg/ Directory Structure - What's Missing

According to CLAUDE.md plans, these directories should exist in pkg/ for future use:
- `pkg/k8s/` - Kubernetes business logic (created for future use)
- `pkg/agent/` - Agent domain models and business rules (created for future use)
- `pkg/workflow/` - Workflow orchestration business logic (created for future use)
- `pkg/diagnosis/` - Diagnostic strategies and rules (created for future use)
- `pkg/telemetry/` - Project-specific telemetry (created for future use)
- `pkg/utils/` - Business-specific utilities (created for future use)

## Dependencies That Will Be Affected

When moving packages from common/ to pkg/, these import paths change:

**Before**:
```go
import "github.com/kart-io/k8s-agent/common/bootstrap"
import "github.com/kart-io/k8s-agent/common/contextx"
import "github.com/kart-io/k8s-agent/common/idempotent"
import "github.com/kart-io/k8s-agent/common/metrics"
```

**After**:
```go
import "github.com/kart-io/k8s-agent/pkg/bootstrap"
import "github.com/kart-io/k8s-agent/pkg/contextx"
import "github.com/kart-io/k8s-agent/pkg/idempotent"
import "github.com/kart-io/k8s-agent/pkg/metrics"
```

### Files That Import These Packages
- Services: agent-manager, orchestrator, auth, cluster, reasoning
- Command apps: cmd/{service}/app/ directories
- Initializers and internal service logic

## Recommended Migration Priority

### Phase 1: Move Straightforward Packages (Week 1)
1. Move bootstrap (common/bootstrap → pkg/bootstrap)
2. Move contextx (common/contextx → pkg/contextx)
3. Move idempotent (common/idempotent → pkg/idempotent)
4. Update imports across codebase
5. Test all services
- **Effort**: 3-5 hours
- **Risk**: Low
- **Testing**: Run full test suite + integration tests

### Phase 2: Refactor Complex Packages (Week 2)
5. Analyze and split common/metrics
6. Refactor common/app to separate concerns
7. Create generic/project-specific split where needed
- **Effort**: 6-10 hours
- **Risk**: Medium
- **Testing**: Full regression testing required

### Phase 3: Future Structure (Week 3+)
8. Create planned pkg/ subdirectories (k8s/, agent/, workflow/, diagnosis/)
9. Populate with business domain logic as services evolve
- **Effort**: 2-3 hours (structure setup)
- **Risk**: Low (just directory creation)

## Summary Table

| Package | Current | Target | Move? | Priority | Status | Effort |
|---------|---------|--------|-------|----------|--------|--------|
| bootstrap | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| contextx | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| idempotent | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| metrics | common/ | pkg/ | PARTIAL | HIGH | Needs review | 2-3h |
| app | common/ | pkg/ | SPLIT | MEDIUM | Needs refactor | 4-6h |
| k8sutils | common/ | mixed | PARTIAL | MEDIUM | Needs review | 2-3h |
| types | pkg/ | pkg/ | NO | N/A | Current | - |
| api | pkg/ | pkg/ | NO | N/A | Current | - |

## Conflicts with Existing common/ Packages

### Direct Conflicts (Same Name)
- common/metrics (generic) vs pkg/metrics (project-specific) - Need to distinguish

### Indirect Conflicts (Related Functionality)
- common/contextx (generic context) vs project-specific extensions in pkg/contextx
- common/bootstrap (component lifecycle) vs pkg/bootstrap (app startup)
- common/initializers (generic) vs project initializers in services

## Recommendations

1. **Immediate Actions**:
   - Move bootstrap, contextx, idempotent to pkg/
   - Update all import paths in codebase
   - Run full test suite to validate

2. **Short-term (Week 2)**:
   - Refactor common/metrics to separate generic from project-specific
   - Review and potentially refactor common/app
   - Add comments in pkg/ packages explaining why they're project-specific

3. **Medium-term**:
   - Create planned pkg/ subdirectories structure
   - Move business domain logic as it's developed
   - Document the pkg/ package structure

4. **Document**:
   - Update CLAUDE.md with actual migration status
   - Create migration completion report
   - Document why each package stays in common/ or moves to pkg/
