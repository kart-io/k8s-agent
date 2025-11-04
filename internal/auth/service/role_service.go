package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// RoleService handles role management business logic.
type RoleService struct {
	db *storage.PostgresDB
}

// NewRoleService creates a new role service.
func NewRoleService(db *storage.PostgresDB) *RoleService {
	return &RoleService{db: db}
}

// List retrieves all roles.
func (s *RoleService) List() ([]types.Role, error) {
	var modelRoles []model.Role
	err := s.db.DB.Where("status = ?", 1).Order("sort").Find(&modelRoles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
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

// GetByID retrieves a role by ID.
func (s *RoleService) GetByID(id string) (*types.Role, error) {
	var role model.Role
	err := s.db.DB.Where("id = ?", id).First(&role).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return &types.Role{
		ID:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      role.Status,
		Sort:        role.Sort,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

// Create creates a new role.
func (s *RoleService) Create(req *types.RoleRequest) (*types.Role, error) {
	// Check if code already exists
	var count int64
	if err := s.db.DB.Model(&model.Role{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check role code: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("role code already exists")
	}

	roleID := uuid.New().String()
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	role := &model.Role{
		ID:          roleID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      status,
		Sort:        req.Sort,
	}

	if err := s.db.DB.Create(role).Error; err != nil {
		return nil, fmt.Errorf("failed to insert role: %w", err)
	}

	return s.GetByID(roleID)
}

// Update updates a role.
func (s *RoleService) Update(id string, req *types.RoleRequest) error {
	// Check if role is system role (cannot be modified)
	var role model.Role
	if err := s.db.DB.Where("id = ?", id).First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("role not found")
	} else if err != nil {
		return fmt.Errorf("failed to query role: %w", err)
	}

	if role.Code == "super_admin" || role.Code == "admin" || role.Code == "user" {
		return fmt.Errorf("cannot modify system role")
	}

	// Build update data map for non-empty fields
	updateData := make(map[string]interface{})

	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}
	if req.Sort != 0 {
		updateData["sort"] = req.Sort
	}

	result := s.db.DB.Model(&model.Role{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return fmt.Errorf("failed to update role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// Delete deletes a role (checks if in use).
func (s *RoleService) Delete(id string) error {
	// Check if role is system role
	var role model.Role
	if err := s.db.DB.Where("id = ?", id).First(&role).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("role not found")
	} else if err != nil {
		return fmt.Errorf("failed to query role: %w", err)
	}

	if role.Code == "super_admin" || role.Code == "admin" || role.Code == "user" {
		return fmt.Errorf("cannot delete system role")
	}

	// Check if role is assigned to users
	var userCount int64
	if err := s.db.DB.Model(&model.UserRole{}).Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("role is assigned to %d user(s), cannot delete", userCount)
	}

	// Delete role
	result := s.db.DB.Delete(&model.Role{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found")
	}

	return nil
}

// AssignPermissions assigns permissions to a role.
func (s *RoleService) AssignPermissions(roleID string, permissionIDs []string) error {
	var role model.Role
	role.ID = roleID

	// Load permissions by IDs
	var permissions []model.Permission
	if err := s.db.DB.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		return fmt.Errorf("failed to find permissions: %w", err)
	}

	// Use GORM Association to replace permissions
	if err := s.db.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		return fmt.Errorf("failed to assign permissions: %w", err)
	}

	return nil
}

// GetPermissions retrieves permissions for a role.
func (s *RoleService) GetPermissions(roleID string) ([]types.Permission, error) {
	var role model.Role
	err := s.db.DB.Preload("Permissions", "status = ?", 1).
		Where("id = ?", roleID).
		First(&role).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("role not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}

	// Convert model.Permission to types.Permission
	permissions := make([]types.Permission, len(role.Permissions))
	for i, p := range role.Permissions {
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
