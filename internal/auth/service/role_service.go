package service

import (
	stderrors "errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	authmodel "github.com/kart-io/k8s-agent/internal/models/auth"
	"github.com/kart-io/logger/core"
)

// RoleService handles role management business logic.
type RoleService struct {
	db     *storage.MySQLDB
	logger core.Logger
}

// NewRoleService creates a new role service.
func NewRoleService(db *storage.MySQLDB, logger core.Logger) *RoleService {
	return &RoleService{
		db:     db,
		logger: logger.With("component", "role-service"),
	}
}

// List retrieves all roles.
func (s *RoleService) List() ([]types.Role, error) {
	s.logger.Infow("List roles request")

	var modelRoles []authmodel.Role
	err := s.db.DB.Where("status = ?", 1).Order("sort").Find(&modelRoles).Error
	if err != nil {
		s.logger.Errorw("List roles failed: database error", "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert authmodel.Role to types.Role
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

	s.logger.Infow("List roles successful", "count", len(roles))
	return roles, nil
}

// GetByID retrieves a role by ID.
func (s *RoleService) GetByID(id string) (*types.Role, error) {
	s.logger.Infow("Get role by ID request", "role_id", id)

	var role authmodel.Role
	err := s.db.DB.Where("id = ?", id).First(&role).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Get role failed: role not found", "role_id", id)
		return nil, errors.ErrNotFound.WithMessage("role not found").KV("role_id", id)
	}
	if err != nil {
		s.logger.Errorw("Get role failed: database error", "role_id", id, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.Infow("Get role successful", "role_id", id, "role_code", role.Code)

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
	s.logger.Infow("Create role request", "role_code", req.Code, "role_name", req.Name)

	// Check if code already exists
	var count int64
	if err := s.db.DB.Model(&authmodel.Role{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		s.logger.Errorw("Create role failed: database error", "role_code", req.Code, "error", err)
		return nil, errors.NewDatabaseError(err)
	}
	if count > 0 {
		s.logger.Warnw("Create role failed: code already exists", "role_code", req.Code)
		return nil, errors.ErrAlreadyExists.WithMessage("role code already exists").KV("role_code", req.Code)
	}

	roleID := uuid.New().String()
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	role := &authmodel.Role{
		ID:          roleID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      status,
		Sort:        req.Sort,
	}

	if err := s.db.DB.Create(role).Error; err != nil {
		s.logger.Errorw("Create role failed: insert error", "role_code", req.Code, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.Infow("Create role successful", "role_id", roleID, "role_code", req.Code)
	return s.GetByID(roleID)
}

// Update updates a role.
func (s *RoleService) Update(id string, req *types.RoleRequest) error {
	s.logger.Infow("Update role request", "role_id", id)

	// Check if role is system role (cannot be modified)
	var role authmodel.Role
	if err := s.db.DB.Where("id = ?", id).First(&role).Error; stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Update role failed: role not found", "role_id", id)
		return errors.ErrNotFound.WithMessage("role not found").KV("role_id", id)
	} else if err != nil {
		s.logger.Errorw("Update role failed: database error", "role_id", id, "error", err)
		return errors.NewDatabaseError(err)
	}

	if role.Code == "super_admin" || role.Code == "admin" || role.Code == "user" {
		s.logger.Warnw("Update role failed: cannot modify system role", "role_id", id, "role_code", role.Code)
		return errors.ErrForbidden.WithMessage("cannot modify system role").KV("role_code", role.Code)
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

	result := s.db.DB.Model(&authmodel.Role{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		s.logger.Errorw("Update role failed: update error", "role_id", id, "error", result.Error)
		return errors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Warnw("Update role failed: no rows affected", "role_id", id)
		return errors.ErrNotFound.WithMessage("role not found").KV("role_id", id)
	}

	s.logger.Infow("Update role successful", "role_id", id)
	return nil
}

// Delete deletes a role (checks if in use).
func (s *RoleService) Delete(id string) error {
	s.logger.Infow("Delete role request", "role_id", id)

	// Check if role is system role
	var role authmodel.Role
	if err := s.db.DB.Where("id = ?", id).First(&role).Error; stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Delete role failed: role not found", "role_id", id)
		return errors.ErrNotFound.WithMessage("role not found").KV("role_id", id)
	} else if err != nil {
		s.logger.Errorw("Delete role failed: database error", "role_id", id, "error", err)
		return errors.NewDatabaseError(err)
	}

	if role.Code == "super_admin" || role.Code == "admin" || role.Code == "user" {
		s.logger.Warnw("Delete role failed: cannot delete system role", "role_id", id, "role_code", role.Code)
		return errors.ErrForbidden.WithMessage("cannot delete system role").KV("role_code", role.Code)
	}

	// Check if role is assigned to users
	var userCount int64
	if err := s.db.DB.Model(&authmodel.UserRole{}).Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		s.logger.Errorw("Delete role failed: database error checking usage", "role_id", id, "error", err)
		return errors.NewDatabaseError(err)
	}
	if userCount > 0 {
		s.logger.Warnw("Delete role failed: role in use", "role_id", id, "user_count", userCount)
		return errors.ErrConflict.WithMessage("role is assigned to users, cannot delete").KV("role_id", id)
	}

	// Delete role
	result := s.db.DB.Delete(&authmodel.Role{}, "id = ?", id)
	if result.Error != nil {
		s.logger.Errorw("Delete role failed: delete error", "role_id", id, "error", result.Error)
		return errors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Warnw("Delete role failed: no rows affected", "role_id", id)
		return errors.ErrNotFound.WithMessage("role not found").KV("role_id", id)
	}

	s.logger.Infow("Delete role successful", "role_id", id, "role_code", role.Code)
	return nil
}

// AssignPermissions assigns permissions to a role.
func (s *RoleService) AssignPermissions(roleID string, permissionIDs []string) error {
	s.logger.Infow("Assign permissions request", "role_id", roleID, "permissions_count", len(permissionIDs))

	var role authmodel.Role
	role.ID = roleID

	// Load permissions by IDs
	var permissions []authmodel.Permission
	if err := s.db.DB.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
		s.logger.Errorw("Assign permissions failed: database error", "role_id", roleID, "error", err)
		return errors.NewDatabaseError(err)
	}

	// Use GORM Association to replace permissions
	if err := s.db.DB.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		s.logger.Errorw("Assign permissions failed: association error", "role_id", roleID, "error", err)
		return errors.NewDatabaseError(err)
	}

	s.logger.Infow("Assign permissions successful", "role_id", roleID, "permissions_count", len(permissions))
	return nil
}

// GetPermissions retrieves permissions for a role.
func (s *RoleService) GetPermissions(roleID string) ([]types.Permission, error) {
	s.logger.Infow("Get permissions request", "role_id", roleID)

	var role authmodel.Role
	err := s.db.DB.Preload("Permissions", "status = ?", 1).
		Where("id = ?", roleID).
		First(&role).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Get permissions failed: role not found", "role_id", roleID)
		return nil, errors.ErrNotFound.WithMessage("role not found").KV("role_id", roleID)
	}
	if err != nil {
		s.logger.Errorw("Get permissions failed: database error", "role_id", roleID, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert authmodel.Permission to types.Permission
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

	s.logger.Infow("Get permissions successful", "role_id", roleID, "permissions_count", len(permissions))
	return permissions, nil
}
