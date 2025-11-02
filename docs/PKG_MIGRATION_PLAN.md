# pkg/ Directory Migration Plan - Executable Actions

## Executive Summary

The pkg/ directory currently contains ONLY:
- **pkg/types/types.go** - Business domain models (CORRECT LOCATION)
- **pkg/api/** - Protocol Buffer definitions (CORRECT LOCATION)

However, according to CLAUDE.md architecture plans, the following packages are currently mislocated in common/ and should be moved to pkg/:

1. **bootstrap** (HIGH PRIORITY - Ready to move)
2. **contextx** (HIGH PRIORITY - Ready to move)
3. **idempotent** (HIGH PRIORITY - Ready to move)
4. **metrics** (MEDIUM PRIORITY - Needs analysis first)
5. **app** (MEDIUM PRIORITY - Needs refactoring first)

## Phase 1: High-Priority Migrations (Ready to Execute)

These packages are ready to move with minimal refactoring. They contain pure project-specific logic with no generic utility value.

### 1.1 Move: common/bootstrap → pkg/bootstrap

**Current Location**: `/common/bootstrap/`
**Files**:
- bootstrap.go (6150 bytes)
- helpers.go (3596 bytes)
- bootstrap_test.go (5134 bytes)

**Content**: Component lifecycle management for Aetherius services
- Bootstrap struct: Lifecycle management for initializers
- Start/Stop/Run methods
- AddInitializer, WaitForComponent, NotifyCompletion helpers

**Why Move**: Pure project-specific application initialization logic. Not useful in other Go projects.

**Classification**: Business Logic

**Steps**:
```bash
# 1. Create target directory
mkdir -p /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/bootstrap

# 2. Copy files
cp /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/bootstrap/* \
   /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/bootstrap/

# 3. Update imports in moved files (if any internal references)
# Currently: none that reference other common packages

# 4. Find and update all imports across codebase
grep -r "github.com/kart-io/k8s-agent/common/bootstrap" /Users/costalong/code/go/src/github.com/kart/k8s-agent --include="*.go"

# 5. Replace in 15-20 files in:
#    - cmd/*/app/*.go
#    - internal/*/initializers/*.go
#    - internal/*/app.go
```

**Files That Import bootstrap**:
- cmd/agent-manager/app/options/server_options.go
- cmd/orchestrator/app/options/server_options.go
- cmd/auth/app/options/server_options.go
- cmd/cluster/app/options/server_options.go
- cmd/reasoning/app/options/server_options.go
- And service initializers...

**Effort**: 1-2 hours

---

### 1.2 Move: common/contextx → pkg/contextx

**Current Location**: `/common/contextx/`
**Files**:
- context.go (10671 bytes)
- timeout.go (3077 bytes)
- context_test.go (6254 bytes)
- k8sagent_test.go (4679 bytes)

**Content**: K8s-agent specific context enrichment
- ContextKey types for trace IDs, agent IDs, cluster IDs
- WithTimeout, WithDeadline helpers
- K8s-specific context values

**Why Move**: These context keys and values are Aetherius-specific. Generic context helpers should be in common/; these domain-specific extensions should be in pkg/.

**Classification**: Business Logic

**Steps**:
```bash
# 1. Create target directory
mkdir -p /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/contextx

# 2. Copy files
cp /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/contextx/* \
   /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/contextx/

# 3. Find all imports to replace
grep -r "github.com/kart-io/k8s-agent/common/contextx" /Users/costalong/code/go/src/github.com/kart/k8s-agent --include="*.go"
```

**Files That Import contextx**: ~20+ files across services

**Effort**: 1-2 hours

---

### 1.3 Move: common/idempotent → pkg/idempotent

**Current Location**: `/common/idempotent/`
**Files**:
- idempotent.go (5339 bytes)
- memory_store.go (2992 bytes)
- redis_store.go (2851 bytes)
- idempotent_test.go (6187 bytes)

**Content**: Business idempotency implementation
- Idempotent struct with logic for deduplicating requests
- MemoryStore, RedisStore backends for idempotency tracking
- Request deduplication logic specific to Agent Manager/Orchestrator patterns

**Why Move**: This is business idempotency logic specific to Aetherius workflows. Generic idempotency patterns should be in common/; this implementation with Agent Manager specific semantics belongs in pkg/.

**Classification**: Business Logic

**Steps**:
```bash
# 1. Create target directory
mkdir -p /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/idempotent

# 2. Copy files
cp /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/idempotent/* \
   /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/idempotent/

# 3. Find all imports to replace
grep -r "github.com/kart-io/k8s-agent/common/idempotent" /Users/costalong/code/go/src/github.com/kart/k8s-agent --include="*.go"
```

**Files That Import idempotent**: ~10+ files in agent-manager, orchestrator, auth

**Effort**: 1-2 hours

---

## Phase 2: Medium-Priority Migrations (Needs Analysis)

### 2.1 Analyze: common/metrics

**Current Location**: `/common/metrics/`
**Files**:
- metrics.go (xxx bytes)
- registry.go (xxx bytes)

**Content**:
- Generic Prometheus metrics
- OR project-specific metrics definitions?
- Need to review actual code

**Decision Required**:
- Is this purely generic Prometheus wrappers? → STAY in common/
- Is this Aetherius project-specific metrics? → MOVE to pkg/

**Analysis Steps**:
```bash
# 1. Read the files
cat /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/metrics/*.go

# 2. Check what metrics are defined
grep -E "^type|^func|NewMetrics" /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/metrics/*.go

# 3. Check if it's generic or project-specific
grep -i "agent\|aetherius\|orchestrator\|workflow" /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/metrics/*.go
```

**Likely Outcome**: SPLIT - Extract generic Prometheus metrics to common/, move project-specific metrics to pkg/

**Effort**: 2-3 hours

---

### 2.2 Refactor: common/app → pkg/app (SPLIT)

**Current Location**: `/common/app/`
**Files**:
- app.go
- bootstrap_app.go
- runner.go
- interfaces.go
- health.go
- middleware.go

**Content**:
- RunWithRunner() - Generic pattern for running apps with components
- RunWithOptions() - Generic pattern for simple apps
- Application interface
- HealthCheck interface
- HTTP middleware setup

**Issue**: HYBRID package
- Generic patterns (RunWithRunner, RunWithOptions) should stay in common/
- Project-specific app startup logic should move to pkg/

**Decision Required**: SPLIT the package

**Recommended Refactoring**:
1. Keep in common/app:
   - app.go - Base Application interface (generic)
   - runner.go - RunWithRunner, RunWithOptions patterns (generic)
   - interfaces.go - HealthCheck, etc. (generic)

2. Move to pkg/app:
   - bootstrap_app.go - BootstrapApp pattern (Aetherius-specific)
   - Project-specific startup logic
   - Health check implementations

**Effort**: 4-6 hours

---

### 2.3 Analyze: common/k8sutils

**Current Location**: `/common/k8sutils/`

**Content**:
- Generic K8s resource conversion utilities
- OR project-specific K8s logic?

**Decision Required**:
- Generic K8s conversion (JSON/YAML/Proto) → STAY in common/
- Project-specific K8s logic → MOVE to pkg/k8s/

**Effort**: 2-3 hours

---

## Packages That Should STAY in common/

Do NOT move these packages. They are generic utility packages suitable for any Go project:

| Package | Reason |
|---------|--------|
| cache | Generic caching (memory/Redis backends) - usable anywhere |
| client | Generic client implementations - usable anywhere |
| config | Generic Options pattern - usable anywhere |
| db | Generic database wrappers (MySQL, Redis) - usable anywhere |
| errors | Generic error handling - usable anywhere |
| logger | Generic logging - usable anywhere |
| middleware | Generic HTTP middleware (CORS, rate limit, etc) - usable anywhere |
| mq | Generic message queue (NATS) - usable anywhere |
| pagination | Generic pagination - usable anywhere |
| response | Generic API response format - usable anywhere |
| server | Generic HTTP/gRPC server - usable anywhere |
| validator | Generic data validation - usable anywhere |
| initializers | Generic component initialization framework - usable anywhere |
| serializers | Generic serialization - usable anywhere |
| utils | Generic utility functions - usable anywhere |

---

## Current State: pkg/types and pkg/api

### pkg/types/types.go - CORRECT LOCATION ✅

**Content**: Business domain models
- Agent, AgentStatus, ConnectionInfo
- Event, InternalEvent, InternalEventType
- Metrics, Command, CommandStatus, CommandResult
- Cluster, ClusterStatus, ClusterHealth
- AlertRule, Alert, AlertStatus
- HealthStatus, Config structures

**Action**: NO CHANGES NEEDED
- Correctly located in pkg/
- Actively used across services
- Pure domain models (no business logic, just data structures)

**Import Count**: ~15 services import this

---

### pkg/api/ - CORRECT LOCATION ✅

**Content**: Protocol Buffer definitions and generated code
- Agent Manager API (agent.proto, command.proto)
- Orchestrator API (workflow.proto)
- Reasoning Service API (analysis.proto)
- Common API (health.proto, pagination.proto, error.proto)
- Generated: .pb.go, .pb.gw.go, .pb.http.go files

**Action**: NO CHANGES NEEDED
- Correctly located in pkg/
- Service contract definitions (should be project-specific)
- Auto-generated from protobuf

---

## Import Path Updates Required

When moving packages, use a systematic approach:

### For each package being moved:

```bash
# Step 1: Count how many files need updating
grep -r "github.com/kart-io/k8s-agent/common/PACKAGE_NAME" --include="*.go" | wc -l

# Step 2: List all files
grep -rl "github.com/kart-io/k8s-agent/common/PACKAGE_NAME" --include="*.go"

# Step 3: Update all files using sed
find /Users/costalong/code/go/src/github.com/kart/k8s-agent -name "*.go" -type f \
  -exec sed -i '' 's|github\.com/kart-io/k8s-agent/common/PACKAGE_NAME|github.com/kart-io/k8s-agent/pkg/PACKAGE_NAME|g' {} \;

# Step 4: Remove original directory
rm -rf /Users/costalong/code/go/src/github.com/kart/k8s-agent/common/PACKAGE_NAME

# Step 5: Run tests
cd /Users/costalong/code/go/src/github.com/kart/k8s-agent
make test
```

---

## Execution Order

### Week 1: Phase 1 (High Priority)
**Timeline**: 3-5 hours
**Risk**: Low

1. Move bootstrap
   - Estimated time: 45 minutes
   - Files to update: ~20
   - Test: `make test`

2. Move contextx
   - Estimated time: 45 minutes
   - Files to update: ~20
   - Test: `make test`

3. Move idempotent
   - Estimated time: 45 minutes
   - Files to update: ~15
   - Test: `make test`

4. Validate and test all services
   - Estimated time: 1 hour
   - Run: Full test suite, integration tests, Docker Compose

### Week 2: Phase 2 (Medium Priority)
**Timeline**: 8-12 hours
**Risk**: Medium

5. Analyze common/metrics
   - Estimated time: 2-3 hours

6. Split common/app
   - Estimated time: 4-6 hours

7. Analyze common/k8sutils
   - Estimated time: 2-3 hours

8. Full validation
   - Estimated time: 1 hour

### Week 3: Phase 3 (Future)
**Timeline**: 2-3 hours
**Risk**: Low

9. Create pkg/ directory structure:
   - pkg/k8s/
   - pkg/agent/
   - pkg/workflow/
   - pkg/diagnosis/
   - pkg/telemetry/
   - pkg/utils/

10. Document final structure in CLAUDE.md

---

## Validation Checklist

After moving each package:

- [ ] Directory created in pkg/
- [ ] Files copied to new location
- [ ] Original files removed from common/
- [ ] All imports updated (search and verify)
- [ ] `go mod tidy` runs without errors
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] `make test-integration` passes (if applicable)
- [ ] Docker Compose deployment works
- [ ] All services start without errors
- [ ] Service health checks pass

---

## Rollback Plan

If any migration step fails:

```bash
# 1. Revert changes from git
git checkout -- common/ cmd/ internal/

# 2. Remove pkg/ directories that were created
rm -rf /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/bootstrap
rm -rf /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/contextx
rm -rf /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/idempotent

# 3. Re-run tests to verify rollback
make test
```

---

## Notes for Developers

1. **Import Change Impact**:
   - Services that use bootstrap, contextx, idempotent will have their imports changed
   - All tests will verify these changes work correctly

2. **No API Changes**:
   - Moving packages to pkg/ is a pure refactoring
   - No changes to public interfaces or behavior
   - No changes to API endpoints or gRPC services

3. **Documentation**:
   - Update CLAUDE.md with completion status
   - Add migration notes to git commit messages
   - Document pkg/ package structure

4. **Future Planning**:
   - After Phase 1 completes, plan Phase 2 with stakeholders
   - Communicate refactoring schedule to team
   - Minimize disruption during active development

---

## Success Criteria

After completing Phase 1:
- All 3 packages (bootstrap, contextx, idempotent) moved to pkg/
- All imports updated across codebase
- All tests passing (unit, integration, e2e)
- All services deploy and run successfully
- Documentation updated
- No functionality changes (pure refactoring)

After completing Phase 2:
- common/metrics and common/app analyzed and split appropriately
- pkg/ structure matches CLAUDE.md design
- Clear separation between generic (common/) and project-specific (pkg/)

