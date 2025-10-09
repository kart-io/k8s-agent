# Research: GORM and kart-io/logger Integration

**Feature**: Code Optimization - GORM and kart-io/logger Integration
**Branch**: `002-gorm-kart-io`
**Date**: 2025-10-09

## Research Questions

### 1. GORM Model Design for Existing Schema

**Question**: How to design GORM models that match the existing PostgreSQL schema without requiring schema changes?

**Decision**: Use GORM struct tags to map to existing column names and table structures

**Rationale**:
- GORM supports custom column names via `column:` tag
- Can disable auto-migration in production via environment variable
- Supports existing foreign key relationships through `foreignKey:` and `references:` tags
- Pointer types for nullable fields (expires_at, last_used_at) match SQL NULL behavior

**Implementation Approach**:
```go
type User struct {
    ID        string    `gorm:"column:id;primaryKey"`
    Username  string    `gorm:"column:username;uniqueIndex;not null"`
    Password  string    `gorm:"column:password;not null"`
    Email     string    `gorm:"column:email;uniqueIndex"`
    // ... other fields
    CreatedAt time.Time `gorm:"column:created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at"`

    // Many-to-many relationship
    Roles     []Role `gorm:"many2many:user_roles;foreignKey:ID;joinForeignKey:UserID;References:ID;joinReferences:RoleID"`
}
```

**Alternatives Considered**:
- **Schema migration**: Rejected - violates FR-017 (no schema changes)
- **Dual models (GORM + raw SQL)**: Rejected - creates inconsistency and defeats purpose of migration

---

### 2. kart-io/logger Configuration Integration

**Question**: How to configure kart-io/logger to match existing logrus behavior while adding new features?

**Decision**: Extend existing config.yaml with logger section and use kart-io/logger's built-in configuration loading

**Rationale**:
- kart-io/logger supports YAML configuration loading
- Can map logrus log levels (Debug, Info, Warn, Error) to kart-io/logger equivalents
- Zap engine provides better performance than logrus
- OTLP integration is optional via configuration flag

**Implementation Approach**:
```yaml
logging:
  engine: "zap"  # or "slog"
  level: "info"  # debug, info, warn, error
  format: "json" # json or text
  output: "stdout"
  otlp:
    enabled: true
    endpoint: "http://localhost:4317"
    insecure: true
```

**Code Integration**:
```go
import "github.com/kart-io/logger"

// Initialize logger
logger.Init(&logger.Config{
    Engine: cfg.Logging.Engine,
    Level:  cfg.Logging.Level,
    Format: cfg.Logging.Format,
    OTLP:   cfg.Logging.OTLP,
})

// Get logger instance
log := logger.GetLogger()
log.Info("Service started", "port", cfg.Server.Port)
```

**Alternatives Considered**:
- **Keep logrus**: Rejected - doesn't meet integration goal with kart-io ecosystem
- **Custom logger wrapper**: Rejected - adds unnecessary complexity, kart-io/logger already provides unified interface

---

### 3. GORM Caching Integration with Redis

**Question**: How to implement automatic caching with GORM hooks without breaking existing cache logic?

**Decision**: Use GORM callbacks (AfterFind, AfterCreate, AfterUpdate, AfterDelete) with feature flag for gradual rollout

**Rationale**:
- GORM callbacks execute automatically on database operations
- Can implement caching logic once and reuse across all models
- Feature flag allows testing in development before production deployment
- Existing cache will be flushed on deployment per clarification decision

**Implementation Approach**:
```go
// Register global callback for permission caching
db.Callback().Query().After("gorm:after_query").Register("cache:after_find", func(db *gorm.DB) {
    if db.Statement.Schema != nil && db.Statement.Schema.Table == "permissions" {
        // Cache the result in Redis
        cacheKey := generateCacheKey(db.Statement)
        cacheData(cacheKey, db.Statement.Dest)
    }
})

// Register callback for cache invalidation
db.Callback().Update().After("gorm:after_update").Register("cache:invalidate", func(db *gorm.DB) {
    if db.Statement.Schema != nil {
        invalidateRelatedCaches(db.Statement.Schema.Table)
    }
})
```

**Alternatives Considered**:
- **Manual caching**: Current approach - rejected due to code duplication and inconsistency risk
- **Third-party GORM cache plugin**: Considered but rejected - adds external dependency and less control

---

### 4. Production Deployment Strategy

**Question**: How to safely deploy GORM migration with minimal risk?

**Decision**: All-at-once deployment with pre-deployment validation and 10-minute rollback window (from clarifications)

**Rationale**:
- Pre-deployment script validates GORM models match existing schema
- Brief downtime (5-10 min) is acceptable per spec assumptions
- Rollback plan keeps previous version ready for quick revert
- Redis cache flush ensures consistency with new caching logic

**Implementation Approach**:

**Pre-Deployment Validation Script**:
```bash
#!/bin/bash
# validate-gorm-migration.sh

# 1. Run GORM AutoMigrate in dry-run mode (dev environment)
# 2. Compare generated schema with production schema
# 3. Verify no column drops or type changes
# 4. Generate migration report
# 5. Require manual approval if differences detected
```

**Deployment Steps**:
1. Take database backup
2. Run validation script
3. Stop current service
4. Flush Redis cache (`redis-cli FLUSHALL`)
5. Deploy new version with GORM
6. Run health checks
7. Monitor for 10 minutes
8. If critical issues: rollback to previous version
9. If successful: complete deployment

**Alternatives Considered**:
- **Blue-green deployment**: Rejected - adds complexity for refactoring task, all-at-once simpler
- **Feature flag per-endpoint**: Rejected - increases code complexity and testing surface

---

### 5. GORM Connection Pool Configuration

**Question**: How to ensure GORM connection pool matches existing database/sql settings?

**Decision**: Configure GORM DB instance with identical pool settings during initialization

**Rationale**:
- GORM exposes `DB()` method to access underlying `*sql.DB`
- Can apply same configuration as current implementation
- No performance regression expected

**Implementation Approach**:
```go
// Initialize GORM
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: gormLogger, // Custom logger using kart-io/logger
})

// Get underlying *sql.DB and configure pool
sqlDB, err := db.DB()
sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)  // 25
sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)  // 5
sqlDB.SetConnMaxLifetime(time.Hour)
```

**Alternatives Considered**:
- **GORM default settings**: Rejected - may differ from current configuration
- **Separate pool config**: Rejected - DRY violation

---

### 6. GORM Query Logging Integration

**Question**: How to log GORM queries using kart-io/logger instead of GORM's default logger?

**Decision**: Implement custom GORM logger interface using kart-io/logger

**Rationale**:
- GORM accepts custom logger via `gorm.Config{Logger: customLogger}`
- Can integrate query timing and context fields
- Consistent log format across entire application

**Implementation Approach**:
```go
type GormLogger struct {
    log *logger.Logger
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
    return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    l.log.Info(msg, "data", data)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    l.log.Warn(msg, "data", data)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    l.log.Error(msg, "data", data)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    l.log.Debug("query executed",
        "sql", sql,
        "rows", rows,
        "elapsed_ms", elapsed.Milliseconds(),
        "error", err,
    )
}
```

**Alternatives Considered**:
- **GORM default logger**: Rejected - doesn't integrate with kart-io/logger ecosystem
- **Separate logging for GORM**: Rejected - creates inconsistent log streams

---

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| ORM | GORM | Latest stable | Database abstraction, migrations (dev/test only) |
| GORM Driver | gorm.io/driver/postgres | Latest | PostgreSQL adapter for GORM |
| Logger | github.com/kart-io/logger | Latest compatible with Go 1.21+ | Unified structured logging |
| Database | PostgreSQL | 13+ | Primary data store (existing) |
| Cache | Redis | 6+ | Permission/role caching (existing) |
| HTTP Framework | Gin | Existing version | Web framework (no changes) |

## Best Practices Applied

### GORM Best Practices
1. **Use pointer types for nullable fields** - Prevents zero-value ambiguity
2. **Define indexes in struct tags** - Ensures proper index creation
3. **Use Preload for relationships** - Avoids N+1 query problem
4. **Use transactions for multi-table ops** - Ensures data consistency
5. **Disable AutoMigrate in production** - Prevents accidental schema changes

### kart-io/logger Best Practices
1. **Initialize once at startup** - Avoid re-initialization overhead
2. **Use structured fields** - Better query ability in log aggregators
3. **Use appropriate log levels** - Debug for dev, Info for prod, Error for failures
4. **Include context fields** - user_id, request_id for traceability
5. **Configure OTLP conditionally** - Optional in dev, required in prod

### Deployment Best Practices
1. **Pre-deployment validation** - Catch schema mismatches early
2. **Database backup before deployment** - Enable quick data recovery
3. **Cache flush coordination** - Prevent stale data issues
4. **Rollback procedure ready** - 10-minute window for detection and revert
5. **Health check monitoring** - Automated detection of deployment issues

## Migration Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| GORM query behavior differs from raw SQL | High | Medium | Comprehensive integration tests comparing results |
| Performance regression | High | Low | Benchmark tests before/after, SC-003 enforces <10ms |
| Null handling issues | Medium | Low | Use pointer types, test with existing data |
| Cache inconsistency after flush | Medium | Medium | Monitoring cache hit/miss rates post-deployment |
| AutoMigrate runs in production | High | Low | Environment variable check, SC-004 validates |
| Logger performance degradation | Medium | Low | Benchmark tests, SC-006 enforces equivalence |

## Success Validation

Each success criterion (SC-001 through SC-010) will be validated through:
- SC-001: Integration test suite comparing GORM vs raw SQL results
- SC-002: Line count diff in `internal/service/*` and `internal/storage/*`
- SC-003: Database query benchmark suite
- SC-004: Service startup timer in health check
- SC-005: Log output format validation
- SC-006: Logging benchmark suite
- SC-007: OTLP collector integration test
- SC-008: Database record count comparison script
- SC-009: API contract tests
- SC-010: Docker Compose health check script
