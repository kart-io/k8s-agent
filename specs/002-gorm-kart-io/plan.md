# Implementation Plan: GORM and kart-io/logger Integration

**Branch**: `002-gorm-kart-io` | **Date**: 2025-10-09 | **Spec**: [spec.md](./spec.md)

## Summary

This plan describes the refactoring of the auth-service codebase to use GORM for database operations and kart-io/logger for structured logging. The primary goals are: (1) Replace raw SQL queries with GORM's type-safe query builder, (2) Migrate from logrus to kart-io/logger for ecosystem integration and OTLP support, (3) Automate Redis caching through GORM hooks.

**Technical Approach**: All-at-once deployment with pre-validation, 5-10 minute downtime, Redis cache flush, and 10-minute rollback window. GORM AutoMigrate disabled in production; models designed to match existing schema without modifications.

---

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- GORM (`gorm.io/gorm` v1.25+, `gorm.io/driver/postgres` latest)
- kart-io/logger (`github.com/kart-io/logger` compatible with Go 1.21+)
- PostgreSQL driver for GORM
- Existing: Gin v1.9.1, Redis client, Prometheus client

**Storage**: PostgreSQL 13+ (existing schema, no changes)
**Testing**: Go testing framework, integration tests, contract tests, benchmarks
**Target Platform**: Linux server (Docker containers)
**Project Type**: Single web service (auth-service microservice)

**Performance Goals**:
- Query latency: Match or exceed current (<10ms per SC-003)
- Startup time: <5 seconds with GORM initialization (SC-004)
- Log throughput: Equivalent to or better than logrus (SC-006)
- OTLP export: <1 second to collector (SC-007)

**Constraints**:
- Zero data loss during migration (SC-008)
- No schema changes allowed (FR-017)
- All API endpoints must remain identical (SC-009)
- Brief downtime acceptable (5-10 minutes per clarifications)
- 30% code reduction target in database layers (SC-002)

**Scale/Scope**:
- ~4,730 lines of existing Go code
- 6 database tables, 2 junction tables
- 28 API endpoints to preserve
- ~3 user stories (P1: GORM, P2: Logger, P3: Caching)

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Status**: N/A - Project constitution is template-only, no active principles defined

**Note**: Since `.specify/memory/constitution.md` contains only placeholder templates, no specific gates apply to this feature. Standard software engineering best practices will be followed:
- Test-driven approach (integration tests before/after migration)
- Code review required
- Documentation updated
- Backward compatibility maintained

**Post-Design Re-Check**: Not applicable (no constitution gates to enforce)

---

## Project Structure

### Documentation (this feature)

```
specs/002-gorm-kart-io/
├── spec.md              # Feature specification (already created)
├── plan.md              # This file
├── research.md          # Phase 0 output (/speckit.plan - CREATED)
├── data-model.md        # Phase 1 output (/speckit.plan - CREATED)
├── quickstart.md        # Phase 1 output (/speckit.plan - CREATED)
├── contracts/           # Phase 1 output (/speckit.plan - CREATED)
│   └── api-contracts.md # API contract preservation requirements
├── checklists/          # Validation checklists
│   └── requirements.md  # Spec quality checklist (already created)
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT YET CREATED)
```

### Source Code (auth-service/)

**Existing Structure** (will be refactored, not restructured):

```
auth-service/
├── cmd/
│   └── server/
│       └── main.go              # MODIFY: Init kart-io/logger, GORM connection
├── internal/
│   ├── config/
│   │   └── config.go            # MODIFY: Add logger config struct
│   ├── model/                   # NEW: GORM model definitions
│   │   ├── user.go              # NEW: User GORM model
│   │   ├── role.go              # NEW: Role GORM model
│   │   ├── permission.go        # NEW: Permission GORM model
│   │   ├── apikey.go            # NEW: APIKey GORM model
│   │   └── associations.go      # NEW: UserRole, RolePermission models
│   ├── storage/
│   │   ├── postgres.go          # MODIFY: Use *gorm.DB instead of *sql.DB
│   │   ├── redis.go             # KEEP: No changes needed
│   │   └── migrate.go           # DEPRECATE: Remove after migration
│   ├── service/
│   │   ├── auth_service.go      # MODIFY: Replace SQL with GORM queries
│   │   ├── user_service.go      # MODIFY: Replace SQL with GORM queries
│   │   ├── role_service.go      # MODIFY: Replace SQL with GORM queries
│   │   ├── permission_service.go # MODIFY: Replace SQL with GORM queries
│   │   └── apikey_service.go    # MODIFY: Replace SQL with GORM queries
│   ├── handler/                 # MINIMAL: Update response handling if needed
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── role_handler.go
│   │   ├── permission_handler.go
│   │   └── apikey_handler.go
│   └── middleware/
│       ├── logging.go           # MODIFY: Use kart-io/logger
│       ├── jwt.go               # MODIFY: Use kart-io/logger
│       ├── permission.go        # MODIFY: Use kart-io/logger
│       ├── cors.go              # MINIMAL: Update logging calls
│       └── api_key.go           # MINIMAL: Update logging calls
├── pkg/
│   ├── logger/
│   │   ├── logger.go            # MODIFY: Replace logrus with kart-io/logger
│   │   └── gorm_adapter.go      # NEW: GORM logger adapter
│   ├── cache/
│   │   ├── permission_cache.go  # MODIFY: Integrate GORM hooks (P3)
│   │   └── gorm_cache.go        # NEW: GORM callback registration (P3)
│   ├── jwt/                     # KEEP: No changes
│   ├── crypto/                  # KEEP: No changes
│   ├── metrics/                 # KEEP: No changes
│   ├── response/                # KEEP: No changes
│   ├── filter/                  # KEEP: No changes
│   ├── validator/               # KEEP: No changes
│   └── pagination/              # KEEP: No changes
├── configs/
│   └── config.yaml              # MODIFY: Add logging section
├── tests/
│   ├── integration/             # NEW: GORM integration tests
│   │   └── gorm_test.go
│   ├── benchmark/               # NEW: Performance benchmarks
│   │   └── query_bench_test.go
│   └── contract/                # NEW: API contract tests
│       ├── api_test.go
│       └── run-contract-tests.sh
├── scripts/
│   ├── validate-gorm-schema.sh  # NEW: Pre-deployment validation
│   ├── measure-code-reduction.sh # NEW: Verify SC-002
│   └── verify-data-integrity.sh  # NEW: Verify SC-008
├── docker-compose.yml           # MINIMAL: May need env vars update
├── go.mod                       # MODIFY: Add GORM, kart-io/logger deps
└── go.sum                       # AUTO-GENERATED

```

**Structure Decision**: Existing single-service structure is preserved. The refactoring adds GORM models to `internal/model/` and updates service layer to use GORM methods. No major restructuring required since this is a code-level migration, not an architectural change.

---

## Complexity Tracking

*No constitution violations to justify - constitution file is template-only.*

---

## Implementation Phases

### Phase 0: Research ✅ COMPLETED

**Output**: `research.md`

**Key Decisions Made**:
1. **GORM Model Design**: Use struct tags to map existing schema, pointer types for nullable fields
2. **Logger Configuration**: Extend config.yaml with logging section, Zap engine by default
3. **Caching Strategy**: GORM callbacks for auto-caching, feature flag for gradual rollout
4. **Deployment**: All-at-once with validation, Redis flush, 10-min rollback window
5. **Connection Pool**: Configure GORM DB instance to match existing settings (25/5)
6. **Query Logging**: Custom GORM logger using kart-io/logger interface

**Alternatives Considered**: All documented in research.md with rationale

---

### Phase 1: Design & Contracts ✅ COMPLETED

**Output**: `data-model.md`, `contracts/api-contracts.md`, `quickstart.md`

**Data Model**:
- 6 GORM models defined (User, Role, Permission, APIKey, UserRole, RolePermission)
- All models map to existing tables via `gorm:` tags
- Relationships defined (many-to-many, belongs-to, has-many)
- Indexes preserved in struct tags
- Nullable fields use pointer types

**API Contracts**:
- All 28 existing endpoints documented
- Request/response formats must remain identical
- Error response format standardized
- Contract test requirements defined

**Quickstart Guide**:
- Development setup instructions
- Implementation checklist (6 phases)
- Deployment guide with pre/post validation
- Troubleshooting common issues
- Success validation commands

---

### Phase 2: Task Breakdown (NEXT STEP)

**Command**: `/speckit.tasks`

**Expected Output**: `tasks.md` with detailed implementation tasks

**Task Categories**:
1. **Setup Tasks** (T001-T005):
   - Add dependencies to go.mod
   - Update config.yaml structure
   - Create internal/model directory
   - Add test directories

2. **GORM Model Tasks** (T006-T011):
   - Implement User model
   - Implement Role model
   - Implement Permission model
   - Implement APIKey model
   - Implement junction models
   - Add validation methods

3. **Database Layer Tasks** (T012-T016):
   - Refactor postgres.go for GORM
   - Create GORM logger adapter
   - Add environment-based AutoMigrate control
   - Configure connection pool
   - Remove old migration scripts

4. **Service Layer Tasks** (T017-T026):
   - Migrate auth_service.go
   - Migrate user_service.go
   - Migrate role_service.go
   - Migrate permission_service.go
   - Migrate apikey_service.go
   - Update handler response mapping

5. **Logger Migration Tasks** (T027-T035):
   - Replace pkg/logger with kart-io/logger
   - Update all service logging calls
   - Update middleware logging
   - Update main.go initialization
   - Configure OTLP (optional)

6. **Caching Integration Tasks** (T036-T040):
   - Create GORM cache callbacks
   - Register callbacks in postgres.go
   - Update permission_cache.go
   - Add feature flag
   - Test cache invalidation

7. **Testing Tasks** (T041-T050):
   - Write integration tests
   - Write benchmark tests
   - Write contract tests
   - Create baseline capture script
   - Create comparison script

8. **Deployment Tasks** (T051-T060):
   - Create validation script
   - Create deployment script
   - Create rollback procedure
   - Update Docker Compose
   - Document monitoring

---

## Migration Strategy

### Pre-Migration Phase

**Baseline Capture**:
```bash
# Capture current API responses
./tests/contract/capture-baseline.sh > baseline.json

# Measure current code size
cloc internal/service internal/storage > code-size-before.txt

# Benchmark current query performance
go test -bench=. ./tests/benchmark > bench-before.txt
```

### Migration Execution

**Development/Test First**:
1. Implement on feature branch `002-gorm-kart-io`
2. Run all tests in Docker Compose environment
3. Validate with integration tests
4. Benchmark performance vs baseline
5. Capture new API responses and compare
6. Code review

**Production Deployment**:
1. Database backup
2. Run validation script (dry-run migration check)
3. Manual approval if validation passes
4. Brief downtime (5-10 min)
5. Redis flush
6. Deploy new version
7. Health checks
8. 10-minute monitoring window
9. Rollback if critical issues
10. Complete if successful

### Post-Migration Phase

**Validation**:
```bash
# Verify all success criteria
./tests/validate-success-criteria.sh

# Monitor for 24 hours
# - Error rates
# - Query latencies
# - Cache hit rates
# - Log volume

# Confirm code reduction
./scripts/measure-code-reduction.sh

# Archive baseline for future reference
```

---

## Risk Mitigation

| Risk | Mitigation | Verification |
|------|------------|--------------|
| Query behavior differs | Comprehensive integration tests | SC-001 validation |
| Performance regression | Pre/post benchmarks, SC-003 enforcement | Benchmark comparison |
| Null handling issues | Pointer types in models, existing data tests | Integration tests with production data copy |
| Cache inconsistency | Flush on deployment, monitoring | Cache hit/miss metrics |
| AutoMigrate in prod | Environment check, SC-004 validation | Config validation in deployment script |
| Logger perf degradation | Benchmark tests, SC-006 enforcement | Logging benchmark suite |
| API contract breakage | Contract tests, SC-009 enforcement | Before/after response comparison |
| Data corruption | Database backup, record count validation | SC-008 verification script |

---

## Success Criteria Validation Plan

Each SC will be validated as follows:

| Criterion | Validation Method | Tooling |
|-----------|-------------------|---------|
| SC-001: Identical results | Integration tests comparing GORM vs raw SQL | `tests/integration/gorm_test.go` |
| SC-002: 30% code reduction | Line count diff in service/storage packages | `scripts/measure-code-reduction.sh` |
| SC-003: Query perf <10ms | Benchmark suite comparing old vs new | `tests/benchmark/query_bench_test.go` |
| SC-004: Startup <5s | Service startup timer | Health check response time |
| SC-005: Structured JSON logs | Log format validation | `tests/logging/validate-log-format.sh` |
| SC-006: Logger perf equivalent | Logging benchmark suite | `tests/benchmark/logger_bench_test.go` |
| SC-007: OTLP <1s | OTLP collector integration test | Manual verification with Jaeger |
| SC-008: Zero data loss | Record count comparison before/after | `scripts/verify-data-integrity.sh` |
| SC-009: Identical API responses | Contract test suite | `tests/contract/compare-responses.sh` |
| SC-010: Docker Compose health | Health check all services <30s | `docker-compose ps`, health endpoints |

---

## Rollback Procedure

**Trigger Conditions**:
- Critical API failures (>5% error rate)
- Database connection failures
- Query performance degradation (>2x baseline)
- Data integrity issues detected

**Rollback Steps**:
1. Stop new service version
2. Deploy previous binary (kept ready)
3. Restart service
4. Verify health checks pass
5. Monitor for 10 minutes
6. Investigation and root cause analysis
7. Fix issues and schedule re-deployment

**Rollback Time**: Target <5 minutes (service restart only, no DB changes to revert)

---

## Dependencies Graph

```
Phase 1: GORM Models
    ↓
Phase 2: Database Layer (postgres.go, GORM logger)
    ↓
Phase 3: Service Layer (auth, user, role, permission, apikey services)
    ↓
Phase 4: Logger Migration (kart-io/logger everywhere)
    ↓
Phase 5: Caching Integration (GORM hooks) [OPTIONAL - P3]
    ↓
Phase 6: Testing (integration, benchmark, contract)
    ↓
Phase 7: Deployment (validation, scripts, monitoring)
```

**Critical Path**: Phases 1-4, 6-7 (Caching is P3, can be deferred)

---

## Next Steps

1. **Run `/speckit.tasks`** to generate detailed task breakdown
2. **Review tasks.md** for implementation sequence
3. **Begin implementation** following quickstart.md checklist
4. **Execute tests** as each phase completes
5. **Deploy to staging** for final validation
6. **Production deployment** following deployment guide

---

## Reference Documents

- **Feature Specification**: [spec.md](./spec.md)
- **Research Decisions**: [research.md](./research.md)
- **Data Models**: [data-model.md](./data-model.md)
- **API Contracts**: [contracts/api-contracts.md](./contracts/api-contracts.md)
- **Quick Reference**: [quickstart.md](./quickstart.md)
- **Requirements Checklist**: [checklists/requirements.md](./checklists/requirements.md)
