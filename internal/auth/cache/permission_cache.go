package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

const (
	// Cache key prefixes
	userPermissionsPrefix = "user:permissions:"
	rolePermissionsPrefix = "role:permissions:"
	userRolesPrefix       = "user:roles:"

	// Cache TTLs
	permissionCacheTTL = 15 * time.Minute
	roleCacheTTL       = 15 * time.Minute
)

// PermissionCache handles permission caching in Redis
type PermissionCache struct {
	redis *storage.RedisClient
	db    *storage.PostgresDB
}

// NewPermissionCache creates a new permission cache
func NewPermissionCache(redis *storage.RedisClient, db *storage.PostgresDB) *PermissionCache {
	return &PermissionCache{
		redis: redis,
		db:    db,
	}
}

// GetUserPermissions gets user permissions from cache or DB
func (pc *PermissionCache) GetUserPermissions(ctx context.Context, userID string) ([]types.Permission, error) {
	cacheKey := userPermissionsPrefix + userID

	// Try to get from cache
	cached, err := pc.redis.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var permissions []types.Permission
		if err := json.Unmarshal([]byte(cached), &permissions); err == nil {
			return permissions, nil
		}
	}

	// Get from database
	permissions, err := pc.fetchUserPermissionsFromDB(userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	data, _ := json.Marshal(permissions)
	pc.redis.Client.Set(ctx, cacheKey, data, permissionCacheTTL)

	return permissions, nil
}

// GetUserRoles gets user roles from cache or DB
func (pc *PermissionCache) GetUserRoles(ctx context.Context, userID string) ([]types.Role, error) {
	cacheKey := userRolesPrefix + userID

	// Try to get from cache
	cached, err := pc.redis.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var roles []types.Role
		if err := json.Unmarshal([]byte(cached), &roles); err == nil {
			return roles, nil
		}
	}

	// Get from database
	roles, err := pc.fetchUserRolesFromDB(userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	data, _ := json.Marshal(roles)
	pc.redis.Client.Set(ctx, cacheKey, data, roleCacheTTL)

	return roles, nil
}

// GetRolePermissions gets role permissions from cache or DB
func (pc *PermissionCache) GetRolePermissions(ctx context.Context, roleID string) ([]types.Permission, error) {
	cacheKey := rolePermissionsPrefix + roleID

	// Try to get from cache
	cached, err := pc.redis.Client.Get(ctx, cacheKey).Result()
	if err == nil {
		var permissions []types.Permission
		if err := json.Unmarshal([]byte(cached), &permissions); err == nil {
			return permissions, nil
		}
	}

	// Get from database
	permissions, err := pc.fetchRolePermissionsFromDB(roleID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	data, _ := json.Marshal(permissions)
	pc.redis.Client.Set(ctx, cacheKey, data, permissionCacheTTL)

	return permissions, nil
}

// InvalidateUserPermissions invalidates user permission cache
func (pc *PermissionCache) InvalidateUserPermissions(ctx context.Context, userID string) error {
	return pc.redis.Client.Del(ctx, userPermissionsPrefix+userID).Err()
}

// InvalidateUserRoles invalidates user roles cache
func (pc *PermissionCache) InvalidateUserRoles(ctx context.Context, userID string) error {
	return pc.redis.Client.Del(ctx, userRolesPrefix+userID).Err()
}

// InvalidateRolePermissions invalidates role permissions cache
func (pc *PermissionCache) InvalidateRolePermissions(ctx context.Context, roleID string) error {
	return pc.redis.Client.Del(ctx, rolePermissionsPrefix+roleID).Err()
}

// InvalidateAllUserCaches invalidates all caches for a user
func (pc *PermissionCache) InvalidateAllUserCaches(ctx context.Context, userID string) error {
	if err := pc.InvalidateUserPermissions(ctx, userID); err != nil {
		return err
	}
	return pc.InvalidateUserRoles(ctx, userID)
}

// InvalidateAllRoleCaches invalidates all caches for a role
func (pc *PermissionCache) InvalidateAllRoleCaches(ctx context.Context, roleID string) error {
	return pc.InvalidateRolePermissions(ctx, roleID)
}

// fetchUserPermissionsFromDB fetches user permissions from database
func (pc *PermissionCache) fetchUserPermissionsFromDB(userID string) ([]types.Permission, error) {
	var modelPerms []model.Permission
	err := pc.db.DB.Distinct().
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("INNER JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND permissions.status = ?", userID, 1).
		Order("permissions.sort ASC").
		Find(&modelPerms).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user permissions: %w", err)
	}

	// Convert model.Permission to types.Permission
	permissions := make([]types.Permission, len(modelPerms))
	for i, p := range modelPerms {
		perm := types.Permission{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Type:        p.Type,
			Path:        p.Path,
			Method:      p.Method,
			Component:   p.Component,
			Icon:        p.Icon,
			Sort:        p.Sort,
			Status:      p.Status,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
		if p.ParentID != nil {
			perm.ParentID = *p.ParentID
		}
		permissions[i] = perm
	}

	return permissions, nil
}

// fetchUserRolesFromDB fetches user roles from database
func (pc *PermissionCache) fetchUserRolesFromDB(userID string) ([]types.Role, error) {
	var modelRoles []model.Role
	err := pc.db.DB.
		Joins("INNER JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, 1).
		Order("roles.sort ASC").
		Find(&modelRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}

	// Convert model.Role to types.Role
	roles := make([]types.Role, len(modelRoles))
	for i, r := range modelRoles {
		roles[i] = types.Role{
			ID:          r.ID,
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
			Status:      r.Status,
			Sort:        r.Sort,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}

	return roles, nil
}

// fetchRolePermissionsFromDB fetches role permissions from database
func (pc *PermissionCache) fetchRolePermissionsFromDB(roleID string) ([]types.Permission, error) {
	var modelPerms []model.Permission
	err := pc.db.DB.
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.status = ?", roleID, 1).
		Order("permissions.sort ASC").
		Find(&modelPerms).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}

	// Convert model.Permission to types.Permission
	permissions := make([]types.Permission, len(modelPerms))
	for i, p := range modelPerms {
		perm := types.Permission{
			ID:          p.ID,
			Name:        p.Name,
			Code:        p.Code,
			Type:        p.Type,
			Path:        p.Path,
			Method:      p.Method,
			Component:   p.Component,
			Icon:        p.Icon,
			Sort:        p.Sort,
			Status:      p.Status,
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
		if p.ParentID != nil {
			perm.ParentID = *p.ParentID
		}
		permissions[i] = perm
	}

	return permissions, nil
}
