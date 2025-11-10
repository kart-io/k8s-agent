# Cluster Service GORM Migration

**Date**: 2025-11-10
**Status**: ✅ COMPLETED
**Service**: cluster

## Summary

Successfully migrated Cluster service database operations from raw SQL to GORM ORM. This simplifies code modifications and maintains consistency with other services like orchestrator.

## Problem Statement

The Cluster service had a GORM model definition in `internal/models/cluster/clusters.go` but was still using raw SQL queries via `sqlDB.ExecContext()` and `sqlDB.QueryRowContext()`. This caused:

- ❌ Code changes were unnecessarily complex
- ❌ Inconsistent database access patterns across services
- ❌ Manual SQL query writing prone to errors
- ❌ No ORM benefits (type safety, automatic mapping, etc.)

## Changes Made

### Files Modified

1. **`internal/cluster/service/cluster.go`** (3 methods migrated)
   - `AddCluster()`: Raw SQL INSERT → GORM Create
   - `getClient()`: Raw SQL SELECT → GORM Where/First

2. **`internal/cluster/service/k8s_cluster.go`** (6 methods migrated)
   - `ListClusters()`: Raw SQL SELECT → GORM Find with Order/Offset/Limit
   - `GetCluster()`: Raw SQL SELECT → GORM Where/First
   - `CreateCluster()`: Raw SQL INSERT → GORM Create
   - `UpdateCluster()`: Raw SQL UPDATE → GORM Model/Where/Updates
   - `DeleteCluster()`: Raw SQL DELETE → GORM Delete
   - `GetClusterOptions()`: Raw SQL SELECT → GORM Find with Select/Order
   - `getClient()`: Raw SQL SELECT → GORM Select/Where/First

**Total**: 9 methods migrated from raw SQL to GORM ORM

## Code Examples

### Before: Raw SQL (cluster.go)

```go
// AddCluster - lines 56-74
sqlDB, err := s.db.DB()
if err != nil {
    return fmt.Errorf("failed to get sql.DB: %w", err)
}

query := `
    INSERT INTO clusters (id, name, description, endpoint, version, status, region, provider, kubeconfig, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
_, err = sqlDB.ExecContext(ctx, query,
    cluster.ID, cluster.Name, cluster.Description, cluster.Endpoint,
    cluster.Version, cluster.Status, cluster.Region, cluster.Provider,
    cluster.KubeConfig, cluster.CreatedAt, cluster.UpdatedAt,
)
if err != nil {
    return err
}
```

### After: GORM ORM (cluster.go)

```go
// AddCluster - lines 57-74
clusterModel := &clustermodel.Cluster{
    ID:          cluster.ID,
    Name:        cluster.Name,
    Description: cluster.Description,
    Endpoint:    cluster.Endpoint,
    Version:     cluster.Version,
    Status:      cluster.Status,
    Region:      cluster.Region,
    Provider:    cluster.Provider,
    KubeConfig:  cluster.KubeConfig,
    CreatedAt:   cluster.CreatedAt,
    UpdatedAt:   cluster.UpdatedAt,
}

if err := s.db.WithContext(ctx).Create(clusterModel).Error; err != nil {
    return fmt.Errorf("failed to create cluster: %w", err)
}
```

### Before: Raw SQL (k8s_cluster.go)

```go
// ListClusters - lines 63-102
var total int64
countQuery := "SELECT COUNT(*) FROM clusters"
if err := s.storage.DB().QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
    return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to count clusters: %w", err))
}

query := `
    SELECT id, name, description, endpoint, version, status, region, provider, created_at, updated_at
    FROM clusters
    ORDER BY created_at DESC
    LIMIT ? OFFSET ?
`

rows, err := s.storage.DB().QueryContext(ctx, query, limit, offset)
// ... manual row scanning ...
```

### After: GORM ORM (k8s_cluster.go)

```go
// ListClusters - lines 63-108
var total int64
if err := s.storage.GormDB().WithContext(ctx).Model(&cluster.Cluster{}).Count(&total).Error; err != nil {
    return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to count clusters: %w", err))
}

var clusterList []*cluster.Cluster
if err := s.storage.GormDB().WithContext(ctx).
    Order("created_at DESC").
    Offset(offset).
    Limit(limit).
    Find(&clusterList).Error; err != nil {
    return nil, 0, errors.NewDatabaseError(fmt.Errorf("failed to query clusters: %w", err))
}

// Convert to ClusterInfo (automatic field mapping)
```

## Benefits

### 1. Code Simplification

- ✅ No more manual SQL query writing
- ✅ No more manual row scanning (`rows.Scan(...)`)
- ✅ Automatic struct field mapping
- ✅ Type-safe database operations

### 2. Consistency

- ✅ Same GORM pattern as orchestrator service
- ✅ Consistent error handling (`gorm.ErrRecordNotFound`)
- ✅ Unified database access approach across project

### 3. Maintainability

- ✅ Easier to modify database queries (fluent API)
- ✅ Less error-prone (no SQL typos, no scan mismatches)
- ✅ Better IDE support (autocomplete, refactoring)

### 4. Features

- ✅ Automatic context support (`WithContext(ctx)`)
- ✅ Query builder methods (`Where`, `Order`, `Limit`, `Offset`)
- ✅ Model-based operations (cleaner code)

## Line Count Impact

### cluster.go

| Method | Before (Raw SQL) | After (GORM) | Change |
|--------|------------------|--------------|--------|
| AddCluster | 19 lines | 17 lines | -2 |
| getClient | 12 lines | 9 lines | -3 |
| **Total** | **31 lines** | **26 lines** | **-5 (-16%)** |

### k8s_cluster.go

| Method | Before (Raw SQL) | After (GORM) | Change |
|--------|------------------|--------------|--------|
| ListClusters | 40 lines | 46 lines | +6* |
| GetCluster | 23 lines | 31 lines | +8* |
| CreateCluster | 14 lines | 18 lines | +4* |
| UpdateCluster | 11 lines | 14 lines | +3* |
| DeleteCluster | 11 lines | 10 lines | -1 |
| GetClusterOptions | 28 lines | 19 lines | -9 |
| getClient | 11 lines | 17 lines | +6* |
| **Total** | **138 lines** | **155 lines** | **+17 (+12%)** |

\* *Increases due to model-to-DTO conversion logic, not raw SQL vs GORM overhead*

### Net Impact

- **cluster.go**: -5 lines (-16%)
- **k8s_cluster.go**: +17 lines (+12%)
- **Combined**: +12 lines net increase

**Why net increase?**
- Most increases are from model-to-DTO conversion logic (ClusterInfo struct)
- GORM queries themselves are shorter and cleaner
- Additional error handling for `gorm.ErrRecordNotFound`
- Benefit: Code is more maintainable and easier to modify

## Verification

### Build Test

```bash
make go.build.cluster
```

**Result**: ✅ Build successful

### Code Quality

- ✅ All methods use GORM ORM consistently
- ✅ Proper error handling with `gorm.ErrRecordNotFound`
- ✅ Context support via `WithContext(ctx)`
- ✅ No raw SQL queries remaining

## Pattern Reference

### Standard GORM Query Pattern

```go
// 1. Query single record
var model cluster.Cluster
err := s.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
if err == gorm.ErrRecordNotFound {
    return ErrNotFound
}

// 2. Query multiple records
var models []cluster.Cluster
err := s.db.WithContext(ctx).
    Where("status = ?", "active").
    Order("created_at DESC").
    Offset(offset).
    Limit(limit).
    Find(&models).Error

// 3. Create record
model := &cluster.Cluster{...}
err := s.db.WithContext(ctx).Create(model).Error

// 4. Update record
err := s.db.WithContext(ctx).
    Model(&cluster.Cluster{}).
    Where("id = ?", id).
    Updates(map[string]interface{}{
        "name": newName,
        "updated_at": time.Now(),
    }).Error

// 5. Delete record
result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&cluster.Cluster{})
if result.RowsAffected == 0 {
    return ErrNotFound
}

// 6. Count records
var count int64
err := s.db.WithContext(ctx).Model(&cluster.Cluster{}).Count(&count).Error
```

## Migration Checklist for Other Services

If you need to migrate another service from raw SQL to GORM:

- [ ] Import GORM model: `import clustermodel "github.com/kart-io/k8s-agent/internal/models/cluster"`
- [ ] Add `gorm` import if not present
- [ ] Replace `s.db.DB()` calls with `s.db.WithContext(ctx)`
- [ ] Replace raw SQL queries with GORM methods:
  - [ ] `INSERT` → `Create()`
  - [ ] `SELECT ... WHERE` → `Where().First()` or `Find()`
  - [ ] `UPDATE` → `Model().Where().Updates()`
  - [ ] `DELETE` → `Where().Delete()`
  - [ ] `SELECT COUNT(*)` → `Model().Count()`
- [ ] Handle `gorm.ErrRecordNotFound` properly
- [ ] Test build: `make go.build.<service>`
- [ ] Run tests if available

## Related Documentation

- [Cluster gRPC Separation](CLUSTER_GRPC_SEPARATION.md) - Previous optimization
- [Service Startup Simplification](SERVICE_STARTUP_SIMPLIFICATION.md) - Service patterns
- [Initializer Unification](INITIALIZER_UNIFICATION_SUMMARY.md) - Database initializers

## Conclusion

Successfully migrated all Cluster service database operations from raw SQL to GORM ORM:

- ✅ 9 methods migrated (2 in cluster.go, 7 in k8s_cluster.go)
- ✅ Build verification successful
- ✅ Consistent with orchestrator service patterns
- ✅ Code is now easier to maintain and modify
- ✅ Resolves user's concern: "现在改动代码太复杂"

This completes the GORM migration for the Cluster service, addressing the issue raised by the user that code changes were too complex due to raw SQL usage.
