# Tasks: GORM and kart-io/logger Integration

**Feature Branch**: `002-gorm-kart-io`
**Input**: Design documents from `/specs/002-gorm-kart-io/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Tests are NOT explicitly requested in the feature specification. Test tasks are included only for validation purposes (integration tests, benchmarks, contract tests) as specified in the success criteria.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

All paths are relative to `auth-service/` directory:
- Models: `internal/model/`
- Services: `internal/service/`
- Storage: `internal/storage/`
- Logger: `pkg/logger/`
- Config: `internal/config/`
- Tests: `tests/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency management

- [ ] T001 [P] Add GORM dependencies to `auth-service/go.mod`: `gorm.io/gorm` v1.25+, `gorm.io/driver/postgres` latest
- [ ] T002 [P] Add kart-io/logger dependency to `auth-service/go.mod`: `github.com/kart-io/logger` (Go 1.21+ compatible)
- [ ] T003 [P] Remove old dependencies from `auth-service/go.mod`: direct `database/sql` usage, `github.com/sirupsen/logrus`
- [ ] T004 Run `go mod tidy` to clean up `auth-service/go.mod` and `auth-service/go.sum`
- [ ] T005 [P] Create directory structure: `auth-service/internal/model/` for GORM models
- [ ] T006 [P] Create directory structure: `auth-service/tests/integration/` for integration tests
- [ ] T007 [P] Create directory structure: `auth-service/tests/benchmark/` for performance benchmarks
- [ ] T008 [P] Create directory structure: `auth-service/tests/contract/` for API contract tests
- [ ] T009 [P] Create directory structure: `auth-service/scripts/` for validation and deployment scripts

**Checkpoint**: Dependencies installed, directory structure ready

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T010 Extend `auth-service/internal/config/config.go` to add `LoggingConfig` struct with fields: Engine (string), Level (string), Format (string), Output (string), OTLPConfig (nested struct with Enabled, Endpoint, Insecure)
- [ ] T011 Update `auth-service/configs/config.yaml` to add logging section: engine="zap", level="info", format="json", output="stdout", otlp.enabled=false
- [ ] T012 Add `auth-service/internal/config/config.go` field: `DatabaseConfig.AutoMigrate` (bool) for environment-based migration control
- [ ] T013 [P] Create GORM logger adapter in `auth-service/pkg/logger/gorm_adapter.go` that implements `gorm.io/gorm/logger.Interface` using kart-io/logger
- [ ] T014 Replace `auth-service/pkg/logger/logger.go` implementation: remove logrus, add kart-io/logger initialization with config loading, create singleton GetLogger() function
- [ ] T015 Update `auth-service/cmd/server/main.go`: Initialize kart-io/logger at startup before any other services using config from config.yaml
- [ ] T016 Refactor `auth-service/internal/storage/postgres.go`: Replace `*sql.DB` with `*gorm.DB`, initialize GORM with PostgreSQL driver and custom logger, configure connection pool (MaxOpenConns: 25, MaxIdleConns: 5, ConnMaxLifetime: 1 hour)
- [ ] T017 Add environment-based AutoMigrate control to `auth-service/internal/storage/postgres.go`: Run `db.AutoMigrate()` only if `cfg.Database.AutoMigrate == true` (dev/test only, false in production)

**Checkpoint**: Foundation ready - GORM database layer functional, kart-io/logger initialized, user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Database Operations with GORM (Priority: P1) 🎯 MVP

**Goal**: Replace all raw SQL queries with GORM query methods while maintaining identical query results (FR-001 to FR-009)

**Independent Test**: All existing database operations (user CRUD, role management, permission queries, API key management) work identically using GORM models and methods instead of raw SQL

### GORM Model Definition for User Story 1

- [ ] T018 [P] [US1] Create `auth-service/internal/model/user.go`: Define User model with GORM tags (id, username, password, email, real_name, phone, avatar, status, created_at, updated_at), add many2many relationship with Roles, add TableName() method returning "users"
- [ ] T019 [P] [US1] Create `auth-service/internal/model/role.go`: Define Role model with GORM tags (id, name, code, description, status, sort, created_at, updated_at), add many2many relationships with Users and Permissions, add TableName() method returning "roles"
- [ ] T020 [P] [US1] Create `auth-service/internal/model/permission.go`: Define Permission model with GORM tags (id, parent_id, name, code, type, path, method, component, icon, sort, status, description, created_at, updated_at), add self-referencing Parent/Children relationships, add many2many relationship with Roles, add TableName() method returning "permissions"
- [ ] T021 [P] [US1] Create `auth-service/internal/model/apikey.go`: Define APIKey model with GORM tags (id, name, key, secret, user_id, description, expires_at, status, last_used_at, created_at, updated_at), use pointer types for nullable fields (expires_at, last_used_at), add belongs-to User relationship, add TableName() method returning "api_keys"
- [ ] T022 [P] [US1] Create `auth-service/internal/model/associations.go`: Define UserRole model (user_id, role_id with composite primary key) and RolePermission model (role_id, permission_id with composite primary key), add TableName() methods
- [ ] T023 [US1] Register all models in `auth-service/internal/storage/postgres.go` AutoMigrate call: User, Role, Permission, APIKey, UserRole, RolePermission (only runs in dev/test per T017)

### Service Layer Migration for User Story 1

- [ ] T024 [US1] Refactor `auth-service/internal/service/auth_service.go`: Replace `SELECT` query in `Login()` with `db.Where("username = ?", username).Preload("Roles").First(&user)`, replace permission query with `db.Joins()` for role_permissions and user_roles tables
- [ ] T025 [US1] Refactor `auth-service/internal/service/user_service.go` - `Create()`: Replace `INSERT` with `db.Create(&user)`, wrap role assignment in `db.Transaction()` for atomicity
- [ ] T026 [US1] Refactor `auth-service/internal/service/user_service.go` - `GetByID()`: Replace `SELECT` with `db.Preload("Roles").First(&user, "id = ?", id)`
- [ ] T027 [US1] Refactor `auth-service/internal/service/user_service.go` - `List()`: Replace `SELECT` with pagination using `db.Where(filters).Count(&total)` then `db.Limit(pageSize).Offset((page-1)*pageSize).Find(&users)`
- [ ] T028 [US1] Refactor `auth-service/internal/service/user_service.go` - `Update()`: Replace `UPDATE` with `db.Model(&user).Updates(updateData)`, handle role assignment updates with `db.Association("Roles").Replace(newRoles)` in transaction
- [ ] T029 [US1] Refactor `auth-service/internal/service/user_service.go` - `Delete()`: Replace `DELETE` with `db.Delete(&user, "id = ?", id)` (soft delete using gorm.DeletedAt per FR-008)
- [ ] T030 [US1] Refactor `auth-service/internal/service/user_service.go` - `AssignRoles()`: Replace manual `INSERT INTO user_roles` with `db.Model(&user).Association("Roles").Replace(roles)` inside transaction
- [ ] T031 [US1] Refactor `auth-service/internal/service/role_service.go` - `Create()`: Replace `INSERT` with `db.Create(&role)`
- [ ] T032 [US1] Refactor `auth-service/internal/service/role_service.go` - `GetByID()`: Replace `SELECT` with `db.Preload("Permissions").First(&role, "id = ?", id)`
- [ ] T033 [US1] Refactor `auth-service/internal/service/role_service.go` - `List()`: Replace `SELECT` with `db.Where(filters).Limit(pageSize).Offset((page-1)*pageSize).Find(&roles)` and `db.Model(&Role{}).Where(filters).Count(&total)`
- [ ] T034 [US1] Refactor `auth-service/internal/service/role_service.go` - `Update()`: Replace `UPDATE` with `db.Model(&role).Updates(updateData)`
- [ ] T035 [US1] Refactor `auth-service/internal/service/role_service.go` - `Delete()`: Replace `DELETE` with `db.Delete(&role, "id = ?", id)`, add validation to prevent deletion of system roles (super_admin, admin, user)
- [ ] T036 [US1] Refactor `auth-service/internal/service/role_service.go` - `AssignPermissions()`: Replace manual `INSERT INTO role_permissions` with `db.Model(&role).Association("Permissions").Replace(permissions)` inside transaction
- [ ] T037 [US1] Refactor `auth-service/internal/service/role_service.go` - `GetPermissions()`: Replace `SELECT with JOIN` with `db.Model(&role).Association("Permissions").Find(&permissions)`
- [ ] T038 [US1] Refactor `auth-service/internal/service/permission_service.go` - `Create()`: Replace `INSERT` with `db.Create(&permission)`
- [ ] T039 [US1] Refactor `auth-service/internal/service/permission_service.go` - `GetByID()`: Replace `SELECT` with `db.First(&permission, "id = ?", id)`
- [ ] T040 [US1] Refactor `auth-service/internal/service/permission_service.go` - `GetTree()`: Replace recursive SQL with `db.Preload("Children", "status = ?", 1).Where("parent_id IS NULL AND status = ?", 1).Order("sort ASC").Find(&permissions)`
- [ ] T041 [US1] Refactor `auth-service/internal/service/permission_service.go` - `List()`: Replace `SELECT` with `db.Where(filters).Limit(pageSize).Offset((page-1)*pageSize).Find(&permissions)` and count query
- [ ] T042 [US1] Refactor `auth-service/internal/service/permission_service.go` - `Update()`: Replace `UPDATE` with `db.Model(&permission).Updates(updateData)`
- [ ] T043 [US1] Refactor `auth-service/internal/service/permission_service.go` - `Delete()`: Replace `DELETE` with `db.Delete(&permission, "id = ?", id)` (cascade handled by GORM relationships)
- [ ] T044 [US1] Refactor `auth-service/internal/service/apikey_service.go` - `Create()`: Replace `INSERT` with `db.Create(&apiKey)`, handle nullable expires_at field using pointer type
- [ ] T045 [US1] Refactor `auth-service/internal/service/apikey_service.go` - `GetByKey()`: Replace `SELECT` with `db.Where("key = ?", key).First(&apiKey)`
- [ ] T046 [US1] Refactor `auth-service/internal/service/apikey_service.go` - `List()`: Replace `SELECT` with `db.Where("user_id = ?", userID).Find(&apiKeys)`
- [ ] T047 [US1] Refactor `auth-service/internal/service/apikey_service.go` - `UpdateLastUsed()`: Replace `UPDATE` with `db.Model(&apiKey).Update("last_used_at", time.Now())`, handle nullable pointer field
- [ ] T048 [US1] Refactor `auth-service/internal/service/apikey_service.go` - `Delete()`: Replace `DELETE` with `db.Delete(&apiKey, "id = ?", id)`

### Validation for User Story 1

- [ ] T049 [P] [US1] Create integration test `auth-service/tests/integration/gorm_user_test.go`: Test User CRUD operations (create, get by ID, list with pagination, update, delete, assign roles) verify identical results to raw SQL baseline
- [ ] T050 [P] [US1] Create integration test `auth-service/tests/integration/gorm_role_test.go`: Test Role CRUD operations and permission assignment, verify role-permission associations work correctly
- [ ] T051 [P] [US1] Create integration test `auth-service/tests/integration/gorm_permission_test.go`: Test Permission CRUD, tree query with Preload("Children"), verify hierarchical structure maintained
- [ ] T052 [P] [US1] Create integration test `auth-service/tests/integration/gorm_apikey_test.go`: Test APIKey CRUD with nullable fields (expires_at, last_used_at), verify pointer type handling
- [ ] T053 [P] [US1] Create integration test `auth-service/tests/integration/gorm_auth_test.go`: Test login flow with GORM (username lookup, Preload roles, permission check with joins), verify authentication works identically
- [ ] T054 [P] [US1] Create benchmark `auth-service/tests/benchmark/query_bench_test.go`: Benchmark GORM queries vs raw SQL baseline for common operations (user login, permission check, list with pagination), verify SC-003 (no queries slower than 10ms)
- [ ] T055 [P] [US1] Create contract test `auth-service/tests/contract/api_contract_test.go`: Test all 28 API endpoints documented in contracts/api-contracts.md, compare responses before/after GORM migration, verify SC-009 (identical responses)
- [ ] T056 [P] [US1] Create script `auth-service/scripts/validate-gorm-schema.sh`: Run GORM AutoMigrate in dry-run mode, compare generated schema with production database schema, verify no column drops or type changes (SC-008)
- [ ] T057 [P] [US1] Create script `auth-service/scripts/measure-code-reduction.sh`: Count lines in `internal/service/` and `internal/storage/` before/after GORM migration, verify SC-002 (30%+ reduction)
- [ ] T058 [P] [US1] Create script `auth-service/scripts/verify-data-integrity.sh`: Compare database record counts before/after migration for all tables, verify SC-008 (zero data loss)

**Checkpoint**: At this point, User Story 1 should be fully functional - all database operations use GORM, integration tests pass, performance benchmarks meet SC-003, API contracts validated per SC-009

---

## Phase 4: User Story 2 - Structured Logging with kart-io/logger (Priority: P2)

**Goal**: Replace all logrus calls with kart-io/logger throughout the codebase, enable OTLP integration, ensure structured JSON logging (FR-010 to FR-016)

**Independent Test**: Run service with kart-io/logger configured, perform operations (login, user creation, permission checks), verify logs emitted with correct structure, fields, and levels through console output and OTLP collectors

### Logger Migration for User Story 2

- [ ] T059 [P] [US2] Update `auth-service/internal/middleware/logging.go`: Replace `logrus.WithFields()` calls with `logger.GetLogger()` and structured field syntax (key-value pairs), update request logging to include method, path, status, latency_ms, client_ip, user_id fields per FR-016
- [ ] T060 [P] [US2] Update `auth-service/internal/middleware/jwt.go`: Replace logrus import and calls with kart-io/logger, update error logging with structured fields (error message, token, user_id)
- [ ] T061 [P] [US2] Update `auth-service/internal/middleware/permission.go`: Replace logrus calls with kart-io/logger, add structured fields for permission checks (user_id, permission_code, result)
- [ ] T062 [P] [US2] Update `auth-service/internal/middleware/cors.go`: Replace logrus calls with kart-io/logger
- [ ] T063 [P] [US2] Update `auth-service/internal/middleware/api_key.go`: Replace logrus calls with kart-io/logger, add structured fields (api_key, user_id)
- [ ] T064 [US2] Update `auth-service/internal/service/auth_service.go`: Replace all `log := logrus.WithFields()` with `log := logger.GetLogger()`, update structured field syntax to key-value pairs (logrus.Fields → variadic args)
- [ ] T065 [US2] Update `auth-service/internal/service/user_service.go`: Replace all logrus calls with kart-io/logger, maintain existing log levels (Debug, Info, Warn, Error) per FR-014
- [ ] T066 [US2] Update `auth-service/internal/service/role_service.go`: Replace all logrus calls with kart-io/logger
- [ ] T067 [US2] Update `auth-service/internal/service/permission_service.go`: Replace all logrus calls with kart-io/logger
- [ ] T068 [US2] Update `auth-service/internal/service/apikey_service.go`: Replace all logrus calls with kart-io/logger
- [ ] T069 [US2] Update `auth-service/internal/handler/auth_handler.go`: Replace logrus calls with kart-io/logger (minimal changes, mostly middleware handles logging)
- [ ] T070 [US2] Update `auth-service/internal/handler/user_handler.go`: Replace logrus calls with kart-io/logger
- [ ] T071 [US2] Update `auth-service/internal/handler/role_handler.go`: Replace logrus calls with kart-io/logger
- [ ] T072 [US2] Update `auth-service/internal/handler/permission_handler.go`: Replace logrus calls with kart-io/logger
- [ ] T073 [US2] Update `auth-service/internal/handler/apikey_handler.go`: Replace logrus calls with kart-io/logger
- [ ] T074 [US2] Configure GORM query logging in `auth-service/pkg/logger/gorm_adapter.go`: Implement Trace() method to log SQL queries with timing information when log level is Debug per FR-015

### OTLP Integration for User Story 2 (Optional)

- [ ] T075 [US2] Add OTLP configuration support to `auth-service/cmd/server/main.go`: If `cfg.Logging.OTLP.Enabled == true`, configure kart-io/logger to export logs to OTLP endpoint per FR-013
- [ ] T076 [P] [US2] Update `auth-service/docker-compose.yml`: Add environment variables for OTLP configuration (LOGGING_OTLP_ENABLED, LOGGING_OTLP_ENDPOINT) if not already present

### Validation for User Story 2

- [ ] T077 [P] [US2] Create validation script `auth-service/tests/logging/validate-log-format.sh`: Parse log output, verify JSON format with required fields (timestamp, level, message, context fields), verify SC-005
- [ ] T078 [P] [US2] Create benchmark `auth-service/tests/benchmark/logger_bench_test.go`: Benchmark kart-io/logger (Zap engine) vs logrus baseline for logging operations, verify SC-006 (equivalent or better performance)
- [ ] T079 [P] [US2] Create integration test for OTLP `auth-service/tests/integration/otlp_test.go`: Configure OTLP endpoint, perform authentication events, verify logs exported to collector within 1 second per SC-007 (manual verification with Jaeger or use logger/otlp-docker setup)
- [ ] T080 [US2] Manual validation: Start service with kart-io/logger, perform login/user creation/permission checks, verify all logs contain structured fields (user_id, method, path, status, latency_ms) in JSON format per acceptance criteria

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - GORM handles all database operations, kart-io/logger handles all logging with structured JSON output and optional OTLP export

---

## Phase 5: User Story 3 - GORM Caching Integration (Priority: P3)

**Goal**: Integrate Redis caching with GORM hooks (AfterFind, AfterCreate, AfterUpdate, AfterDelete) for automatic cache population and invalidation of frequently accessed data (permissions, roles)

**Independent Test**: Perform operations that trigger cache population (permission queries) and cache invalidation (role updates), verify cache hit/miss metrics, confirm cached data matches database state

### GORM Caching Hooks for User Story 3

- [ ] T081 [P] [US3] Create `auth-service/pkg/cache/gorm_cache.go`: Implement GORM callback functions: `AfterFindCallback()` to cache query results in Redis with 15-minute TTL, `AfterCreateCallback()` to invalidate related caches, `AfterUpdateCallback()` to invalidate related caches, `AfterDeleteCallback()` to invalidate related caches
- [ ] T082 [US3] Register GORM callbacks in `auth-service/internal/storage/postgres.go`: Add `db.Callback().Query().After("gorm:after_query").Register("cache:after_find", AfterFindCallback)`, register Create/Update/Delete callbacks similarly, apply to Permission and Role tables only
- [ ] T083 [US3] Update `auth-service/pkg/cache/permission_cache.go`: Integrate with GORM hooks, remove manual caching logic if redundant, ensure cache key format matches GORM hook expectations
- [ ] T084 [P] [US3] Add feature flag to `auth-service/internal/config/config.go`: Add `CacheConfig.GormHooksEnabled` (bool) for gradual rollout
- [ ] T085 [US3] Add environment-based GORM cache control to `auth-service/internal/storage/postgres.go`: Register GORM cache callbacks only if `cfg.Cache.GormHooksEnabled == true`

### Validation for User Story 3

- [ ] T086 [P] [US3] Create integration test `auth-service/tests/integration/gorm_cache_test.go`: Test cache population (query permission → verify cached in Redis), test cache invalidation (update role → verify related permission cache cleared), verify cache hit/miss metrics
- [ ] T087 [US3] Manual validation: Enable GORM cache hooks in dev environment, perform permission queries (check Redis for cached data), update roles (verify cache invalidation), monitor cache hit rate recovery within 15 minutes

**Checkpoint**: All user stories should now be independently functional - GORM handles database operations, kart-io/logger provides structured logging with OTLP, GORM hooks automate Redis caching

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, deployment preparation, final validation

- [ ] T088 [P] Remove deprecated code: Delete old migration scripts in `auth-service/internal/storage/migrate.go` (replaced by GORM AutoMigrate in dev/test per FR-003)
- [ ] T089 [P] Update `auth-service/README.md`: Document GORM model usage, kart-io/logger configuration, environment variables (DATABASE_AUTO_MIGRATE, LOGGING_ENGINE, LOGGING_OTLP_ENABLED)
- [ ] T090 [P] Create deployment script `auth-service/scripts/deploy-gorm-migration.sh`: Pre-deployment validation (run validate-gorm-schema.sh), stop service, flush Redis (`redis-cli FLUSHALL`), deploy binary, start service, run health checks, 10-minute monitoring window per FR-023, FR-024
- [ ] T091 [P] Create rollback procedure document `auth-service/docs/rollback-gorm-migration.md`: Document rollback steps (stop new version, restore previous binary, restart, verify health), include trigger conditions (>5% error rate, query perf >2x baseline) per plan.md
- [ ] T092 [P] Update `auth-service/docker-compose.yml`: Verify all services (PostgreSQL, Redis, auth-service) start successfully with GORM and kart-io/logger, verify SC-010 (all healthy within 30 seconds)
- [ ] T093 Verify startup time: Measure service startup with GORM AutoMigrate enabled (dev mode), verify SC-004 (<5 seconds)
- [ ] T094 Run comprehensive validation: Execute all success criteria validation commands from quickstart.md (SC-001 through SC-010), generate validation report
- [ ] T095 Code cleanup: Run `gofmt -s -w` on all modified files, run `go vet ./...`, fix any linting issues
- [ ] T096 Final integration test: Start Docker Compose environment, run all contract tests, verify all 28 API endpoints return identical responses per SC-009
- [ ] T097 Performance validation: Run all benchmarks, verify SC-003 (no queries slower than 10ms), SC-006 (logger performance equivalent)
- [ ] T098 Documentation: Update `auth-service/docs/` with GORM migration guide, troubleshooting section (from quickstart.md), deployment checklist
- [ ] T099 [P] Create monitoring dashboard queries: Document Prometheus queries for monitoring GORM connection pool, query latencies, cache hit rates, error rates post-deployment
- [ ] T100 Final checkpoint: Run quickstart.md validation end-to-end, verify all acceptance scenarios from spec.md pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion (T001-T009) - BLOCKS all user stories
- **User Story 1 - GORM (Phase 3)**: Depends on Foundational phase completion (T010-T017) - Priority P1
- **User Story 2 - Logger (Phase 4)**: Depends on Foundational phase completion (T010-T017) - Priority P2, can proceed in parallel with US1
- **User Story 3 - Caching (Phase 5)**: Depends on US1 completion (needs GORM models and storage layer) - Priority P3
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - GORM)**: Can start after Foundational (T017) - No dependencies on other stories
- **User Story 2 (P2 - Logger)**: Can start after Foundational (T015) - Independent of US1 (different code paths), can run in parallel
- **User Story 3 (P3 - Caching)**: Requires US1 completion (needs GORM models T018-T023, GORM storage T016-T017) - Integrates with US1 GORM layer

### Within Each User Story

**User Story 1 (GORM)**:
- Model definition (T018-T022) → All parallelizable [P]
- Model registration (T023) → Requires all models complete
- Service refactoring (T024-T048) → Sequential within each service file, parallel across different services
- Validation tests (T049-T058) → All parallelizable [P] after implementation complete

**User Story 2 (Logger)**:
- Middleware updates (T059-T063) → All parallelizable [P]
- Service updates (T064-T068) → All parallelizable [P]
- Handler updates (T069-T073) → All parallelizable [P]
- GORM logger (T074) → After T064-T068 complete
- OTLP config (T075-T076) → Parallelizable [P]
- Validation (T077-T080) → All parallelizable [P] after implementation

**User Story 3 (Caching)**:
- Cache callback implementation (T081) → Independent [P]
- Callback registration (T082) → After T081 complete, requires T016-T017 from US1
- Permission cache update (T083) → After T082 complete
- Feature flag (T084-T085) → Parallelizable [P] with T081
- Validation (T086-T087) → After implementation complete

### Parallel Opportunities

- All Setup tasks (T001-T009) can run in parallel
- Within Foundational: T010-T012 [config files] parallel, T013-T014 [logger setup] parallel, T015-T017 sequential
- US1 models (T018-T022) all parallel
- US1 service refactoring: Different service files can be parallelized (T024 parallel with T031, T038, T044)
- US1 validation tests (T049-T058) all parallel
- US2 middleware updates (T059-T063) all parallel
- US2 service updates (T064-T068) all parallel
- US2 handler updates (T069-T073) all parallel
- US2 validation (T077-T079) all parallel
- US3 cache setup (T081, T084) parallel
- Polish tasks (T088-T092, T099) mostly parallel

---

## Parallel Example: User Story 1 - GORM Models

```bash
# Launch all GORM model definitions together:
Task: "Create auth-service/internal/model/user.go (User model)"
Task: "Create auth-service/internal/model/role.go (Role model)"
Task: "Create auth-service/internal/model/permission.go (Permission model)"
Task: "Create auth-service/internal/model/apikey.go (APIKey model)"
Task: "Create auth-service/internal/model/associations.go (UserRole, RolePermission models)"

# After models complete, launch service refactoring in parallel (different files):
Task: "Refactor auth-service/internal/service/user_service.go"
Task: "Refactor auth-service/internal/service/role_service.go"
Task: "Refactor auth-service/internal/service/permission_service.go"
Task: "Refactor auth-service/internal/service/apikey_service.go"
```

## Parallel Example: User Story 2 - Logger Migration

```bash
# Launch all middleware updates together (different files):
Task: "Update auth-service/internal/middleware/logging.go"
Task: "Update auth-service/internal/middleware/jwt.go"
Task: "Update auth-service/internal/middleware/permission.go"
Task: "Update auth-service/internal/middleware/cors.go"
Task: "Update auth-service/internal/middleware/api_key.go"

# Launch all service updates together (different files):
Task: "Update auth-service/internal/service/auth_service.go"
Task: "Update auth-service/internal/service/user_service.go"
Task: "Update auth-service/internal/service/role_service.go"
Task: "Update auth-service/internal/service/permission_service.go"
Task: "Update auth-service/internal/service/apikey_service.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only) - Recommended

1. Complete Phase 1: Setup (T001-T009)
2. Complete Phase 2: Foundational (T010-T017) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 - GORM (T018-T058)
4. **STOP and VALIDATE**: Test User Story 1 independently (run T049-T058)
5. Deploy/demo if ready - service now uses GORM for all database operations

### Incremental Delivery (Recommended Sequence)

1. **Foundation**: Setup + Foundational (T001-T017) → Database and logger infrastructure ready
2. **MVP Release**: Add User Story 1 (T018-T058) → Test independently → Deploy GORM migration
3. **Enhancement 1**: Add User Story 2 (T059-T080) → Test independently → Deploy logger migration
4. **Enhancement 2**: Add User Story 3 (T081-T087) → Test independently → Deploy GORM caching
5. **Polish**: Phase 6 (T088-T100) → Final validation and deployment hardening

Each story adds value without breaking previous stories.

### Parallel Team Strategy

With multiple developers:

1. **Week 1**: Team completes Setup + Foundational together (T001-T017)
2. **Week 2-3**: Once Foundational is done:
   - Developer A: User Story 1 - GORM models + service layer (T018-T048)
   - Developer B: User Story 2 - Logger migration (T059-T080) - can start in parallel
   - Developer C: Validation tests for US1 (T049-T058)
3. **Week 4**:
   - Developer A: User Story 3 - GORM caching (T081-T087) - requires US1 complete
   - Developer B: Polish tasks (T088-T100)
   - Developer C: Integration validation and deployment preparation
4. Stories integrate independently, final integration in Week 4

---

## Notes

- **[P] tasks**: Different files, no dependencies - safe to parallelize
- **[Story] label**: Maps task to specific user story (US1, US2, US3) for traceability
- **Each user story is independently testable**: US2 (Logger) can be tested without US1 (GORM) if US2 completes first, though US1 is higher priority
- **Tests are validation-focused**: Integration tests, benchmarks, contract tests verify success criteria (SC-001 through SC-010)
- **Commit strategy**: Commit after each logical group (all models, each service file, each middleware file)
- **Rollback safety**: FR-023 requires rollback plan; deployment includes 10-minute monitoring window
- **Data safety**: FR-017, SC-008 ensure zero data loss; validation scripts (T056, T058) verify schema compatibility and data integrity
- **Stop at checkpoints**: Each user story completion is a valid stopping point to validate independently before proceeding

### Critical Success Factors

- **SC-001**: All database operations work identically (validate with T049-T055)
- **SC-002**: 30%+ code reduction (validate with T057)
- **SC-003**: No performance regression <10ms (validate with T054)
- **SC-008**: Zero data loss (validate with T056, T058)
- **SC-009**: Identical API responses (validate with T055)
- **Deployment safety**: Pre-deployment validation (T056), Redis flush (T090), rollback procedure (T091)
