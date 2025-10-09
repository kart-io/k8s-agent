# Quickstart: GORM and kart-io/logger Integration

**Feature**: Code Optimization - GORM and kart-io/logger Integration
**Branch**: `002-gorm-kart-io`
**Target Audience**: Developers implementing this refactoring

## Overview

This guide provides a quick reference for implementing the GORM and kart-io/logger integration. For detailed technical decisions, see `research.md`. For data models, see `data-model.md`.

---

## Prerequisites

- Go 1.21+
- Existing auth-service codebase (on branch `001-auth-service-a` or `master`)
- PostgreSQL 13+ (existing instance)
- Redis 6+ (existing instance)
- Docker and Docker Compose (for testing)

---

## Quick Setup (Development)

### 1. Install Dependencies

```bash
cd auth-service

# Add GORM dependencies
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres

# Add kart-io/logger
go get -u github.com/kart-io/logger

# Remove old dependencies
go mod tidy
```

### 2. Update Configuration

Edit `configs/config.yaml`:

```yaml
# Add logging section
logging:
  engine: "zap"      # or "slog" for testing
  level: "debug"     # debug in dev, info in prod
  format: "json"     # or "text" for local dev
  output: "stdout"
  otlp:
    enabled: false   # Enable in prod with collector
    endpoint: "http://localhost:4317"
    insecure: true

# Add database migration control
database:
  auto_migrate: true  # true in dev/test, false in prod
```

### 3. Run with Docker Compose

```bash
# Start all services (PostgreSQL, Redis, auth-service)
docker-compose up -d

# Check logs
docker-compose logs -f auth-service

# Verify health
curl http://localhost:8080/health
```

---

## Implementation Checklist

### Phase 1: GORM Models (Priority: P1)

- [ ] Create `internal/model/` directory if not exists
- [ ] Define GORM models in `internal/model/*.go`
  - [ ] `user.go` - User model with Roles relationship
  - [ ] `role.go` - Role model with Users and Permissions
  - [ ] `permission.go` - Permission model with Parent/Children
  - [ ] `apikey.go` - APIKey model with User relationship
  - [ ] `associations.go` - UserRole and RolePermission models
- [ ] Add `TableName()` methods to all models
- [ ] Test models with AutoMigrate in dev environment

**Validation**:
```bash
# Run in dev mode, check logs for table creation
go run cmd/server/main.go

# Verify tables match existing schema
psql -h localhost -U postgres -d auth_db -c "\d users"
```

---

### Phase 2: Database Layer Refactoring (Priority: P1)

- [ ] Update `internal/storage/postgres.go`
  - [ ] Replace `sql.DB` with `*gorm.DB`
  - [ ] Initialize GORM with PostgreSQL driver
  - [ ] Configure connection pool (25 max open, 5 max idle)
  - [ ] Add environment-based AutoMigrate control
- [ ] Create GORM logger adapter in `pkg/logger/gorm_adapter.go`
- [ ] Remove old migration scripts (keep backup for rollback)

**Example**:
```go
// internal/storage/postgres.go
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: NewGormLogger(kartLogger),
    })
    if err != nil {
        return nil, err
    }

    // Configure pool
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

    // AutoMigrate only in dev/test
    if cfg.AutoMigrate {
        db.AutoMigrate(&model.User{}, &model.Role{}, /* ... */)
    }

    return &PostgresDB{DB: db}, nil
}
```

---

### Phase 3: Service Layer Migration (Priority: P1)

For each service file in `internal/service/*`:

- [ ] **auth_service.go**
  - [ ] Replace raw SQL with `db.Where().First()`
  - [ ] Use `db.Preload("Roles")` for user with roles
  - [ ] Update permission queries with Joins
- [ ] **user_service.go**
  - [ ] Use `db.Create()` instead of INSERT
  - [ ] Use `db.Updates()` instead of UPDATE
  - [ ] Use transactions for role assignments: `db.Transaction()`
  - [ ] Use `db.Limit().Offset()` for pagination
- [ ] **role_service.go**
  - [ ] Replace SQL queries with GORM methods
  - [ ] Use `db.Association()` for permission management
- [ ] **permission_service.go**
  - [ ] Use `db.Preload("Children")` for tree structure
  - [ ] Replace recursive SQL with GORM relationships
- [ ] **apikey_service.go**
  - [ ] Replace SQL queries with GORM methods
  - [ ] Handle pointer types for nullable fields

**Before (Raw SQL)**:
```go
func (s *UserService) GetByID(id string) (*types.User, error) {
    var user types.User
    err := s.db.DB.QueryRow(
        "SELECT id, username, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Username, &user.Email)
    return &user, err
}
```

**After (GORM)**:
```go
func (s *UserService) GetByID(id string) (*model.User, error) {
    var user model.User
    err := s.db.DB.Preload("Roles").First(&user, "id = ?", id).Error
    return &user, err
}
```

---

### Phase 4: Logger Migration (Priority: P2)

- [ ] Update `pkg/logger/logger.go`
  - [ ] Replace logrus imports with kart-io/logger
  - [ ] Add initialization function with config
  - [ ] Create singleton GetLogger() function
- [ ] Update all service files
  - [ ] Replace `log := logger.GetLogger()` calls
  - [ ] Update log method calls (logrus → kart-io/logger)
  - [ ] Update structured field syntax if needed
- [ ] Update middleware
  - [ ] `internal/middleware/logging.go` - Use kart-io/logger
  - [ ] `internal/middleware/jwt.go` - Update log calls
  - [ ] `internal/middleware/permission.go` - Update log calls
- [ ] Update main.go
  - [ ] Initialize kart-io/logger at startup
  - [ ] Configure OTLP if enabled

**Before (logrus)**:
```go
import "github.com/sirupsen/logrus"

log := logrus.WithFields(logrus.Fields{
    "user_id": userID,
    "action": "login",
})
log.Info("User logged in")
```

**After (kart-io/logger)**:
```go
import "github.com/kart-io/logger"

log := logger.GetLogger()
log.Info("User logged in",
    "user_id", userID,
    "action", "login",
)
```

---

### Phase 5: Caching Integration (Priority: P3)

- [ ] Create `pkg/cache/gorm_cache.go`
- [ ] Register GORM callbacks
  - [ ] `AfterFind` - Cache query results
  - [ ] `AfterCreate/Update/Delete` - Invalidate cache
- [ ] Update `internal/storage/postgres.go` to register callbacks
- [ ] Add feature flag for gradual rollout

**Example**:
```go
// Register caching callback
db.Callback().Query().After("gorm:after_query").Register("cache:after_find",
    func(db *gorm.DB) {
        // Cache logic here
    },
)
```

---

### Phase 6: Testing (Priority: P1)

- [ ] Create `tests/integration/gorm_test.go`
  - [ ] Test all service methods
  - [ ] Compare results with baseline (if available)
- [ ] Create `tests/benchmark/query_bench_test.go`
  - [ ] Benchmark GORM queries vs raw SQL
  - [ ] Verify SC-003 (no queries slower than 10ms)
- [ ] Create `tests/contract/api_test.go`
  - [ ] Test all API endpoints
  - [ ] Verify SC-009 (identical responses)
- [ ] Update Docker Compose for test environment

**Run Tests**:
```bash
# Unit tests
go test ./...

# Integration tests
go test ./tests/integration/...

# Benchmarks
go test -bench=. ./tests/benchmark/...

# Contract tests
./tests/contract/run-contract-tests.sh
```

---

## Deployment Guide

### Pre-Deployment Checklist

- [ ] All tests passing
- [ ] Code review complete
- [ ] Database backup taken
- [ ] Rollback procedure documented
- [ ] Monitoring alerts configured

### Deployment Steps

**1. Pre-Deployment Validation**:
```bash
# Run validation script (generates report)
./scripts/validate-gorm-schema.sh

# Review report for schema differences
cat migration-validation-report.txt

# Approve deployment if validation passes
```

**2. Deployment**:
```bash
# Stop service (brief downtime starts)
systemctl stop auth-service

# Flush Redis cache
redis-cli FLUSHALL

# Deploy new binary
cp bin/auth-service /usr/local/bin/auth-service

# Start service
systemctl start auth-service

# Wait for health check
curl --retry 5 --retry-delay 2 http://localhost:8080/health
```

**3. Post-Deployment Verification** (10-minute window):
```bash
# Check service logs
journalctl -u auth-service -f

# Run smoke tests
./tests/smoke/run-smoke-tests.sh

# Check error rate in metrics
curl http://localhost:8080/metrics | grep error_total

# Test critical endpoints
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

**4. Rollback (if needed)**:
```bash
# Stop new version
systemctl stop auth-service

# Deploy previous version
cp bin/auth-service.backup /usr/local/bin/auth-service

# Start service
systemctl start auth-service

# Verify rollback successful
curl http://localhost:8080/health
```

---

## Troubleshooting

### Common Issues

**1. GORM AutoMigrate runs in production**
```bash
# Check environment
env | grep AUTO_MIGRATE

# Should be: DATABASE_AUTO_MIGRATE=false

# If true, update config and restart
```

**2. Query performance regression**
```bash
# Enable GORM query logging
export LOG_LEVEL=debug

# Check slow queries in logs
grep "elapsed_ms" /var/log/auth-service.log | awk '$NF > 10'

# Identify missing indexes or N+1 queries
```

**3. Cache miss spike after deployment**
```bash
# Expected behavior (cache was flushed)
# Monitor cache hit rate

curl http://localhost:8080/metrics | grep cache_hits

# Should recover to normal within 15 minutes
```

**4. Logger OTLP connection failed**
```bash
# Check OTLP endpoint
curl http://otlp-collector:4317

# If unavailable, disable OTLP temporarily
export LOGGING_OTLP_ENABLED=false

# Restart service
```

---

## Success Validation

After deployment, verify success criteria:

```bash
# SC-001: Database operations work
./tests/integration/run-all-tests.sh

# SC-002: Code reduction (30%+)
./scripts/measure-code-reduction.sh

# SC-003: Query performance (<10ms)
./tests/benchmark/run-query-benchmarks.sh

# SC-004: Startup time (<5s)
time systemctl start auth-service

# SC-005-007: Logging validation
./tests/logging/validate-log-format.sh

# SC-008: Data integrity
./scripts/verify-data-integrity.sh

# SC-009: API contracts
./tests/contract/compare-responses.sh

# SC-010: Docker Compose health
docker-compose ps
```

---

## Reference

- **Full Plan**: `plan.md`
- **Research**: `research.md`
- **Data Models**: `data-model.md`
- **API Contracts**: `contracts/api-contracts.md`
- **Specification**: `spec.md`

---

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review research.md for technical decisions
3. Consult GORM documentation: https://gorm.io/docs/
4. Consult kart-io/logger: https://github.com/kart-io/logger
