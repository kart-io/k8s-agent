package service

import (
	stderrors "errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/errors"
	authmodel "github.com/kart-io/k8s-agent/internal/models/auth"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	"github.com/kart-io/logger/core"
)

// PermissionService handles permission management business logic.
type PermissionService struct {
	db     *storage.MySQLDB
	logger core.Logger
}

// NewPermissionService creates a new permission service.
func NewPermissionService(db *storage.MySQLDB, logger core.Logger) *PermissionService {
	return &PermissionService{
		db:     db,
		logger: logger.With("component", "permission-service"),
	}
}

// List retrieves all permissions with optional filtering.
func (s *PermissionService) List(typeFilter, statusFilter string) ([]types.Permission, error) {
	s.logger.Infow("List permissions request",
		"type_filter", typeFilter,
		"status_filter", statusFilter,
	)

	query := s.db.DB.Model(&authmodel.Permission{})

	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var modelPerms []authmodel.Permission
	if err := query.Order("sort").Find(&modelPerms).Error; err != nil {
		s.logger.Errorw("List permissions failed: database error", "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert authmodel.Permission to types.Permission
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

	s.logger.Infow("List permissions successful", "count", len(permissions))
	return permissions, nil
}

// GetTree builds hierarchical permission tree.
func (s *PermissionService) GetTree() ([]*types.PermissionNode, error) {
	s.logger.Infow("Get permission tree request")

	// Get root permissions (parent_id IS NULL) using GORM Preload
	var roots []authmodel.Permission
	err := s.db.DB.Preload("Children").
		Where("parent_id IS NULL AND status = ?", 1).
		Order("sort").
		Find(&roots).Error
	if err != nil {
		s.logger.Errorw("Get permission tree failed: database error", "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert to PermissionNode tree
	tree := make([]*types.PermissionNode, len(roots))
	for i := range roots {
		tree[i] = convertToPermissionNode(&roots[i])
	}

	s.logger.Infow("Get permission tree successful", "root_nodes", len(tree))
	return tree, nil
}

// convertToPermissionNode recursively converts authmodel.Permission to types.PermissionNode.
func convertToPermissionNode(perm *authmodel.Permission) *types.PermissionNode {
	node := &types.PermissionNode{
		ID:       perm.ID,
		Name:     perm.Name,
		Code:     perm.Code,
		Type:     perm.Type,
		Path:     perm.Path,
		Method:   perm.Method,
		Icon:     perm.Icon,
		Sort:     perm.Sort,
		Children: []*types.PermissionNode{},
	}

	if perm.ParentID != nil {
		node.ParentID = *perm.ParentID
	}

	// Recursively convert children
	for i := range perm.Children {
		node.Children = append(node.Children, convertToPermissionNode(&perm.Children[i]))
	}

	return node
}

// GetByID retrieves a permission by ID.
func (s *PermissionService) GetByID(id string) (*types.Permission, error) {
	s.logger.Infow("Get permission by ID request", "permission_id", id)

	var perm authmodel.Permission
	err := s.db.DB.Where("id = ?", id).First(&perm).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Get permission failed: permission not found", "permission_id", id)
		return nil, errors.ErrNotFound.WithMessage("permission not found").KV("permission_id", id)
	}
	if err != nil {
		s.logger.Errorw("Get permission failed: database error", "permission_id", id, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.Infow("Get permission successful", "permission_id", id, "permission_code", perm.Code)

	result := &types.Permission{
		ID:          perm.ID,
		Name:        perm.Name,
		Code:        perm.Code,
		Type:        perm.Type,
		Path:        perm.Path,
		Method:      perm.Method,
		Component:   perm.Component,
		Icon:        perm.Icon,
		Sort:        perm.Sort,
		Status:      perm.Status,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}
	if perm.ParentID != nil {
		result.ParentID = *perm.ParentID
	}

	return result, nil
}

// Create creates a new permission.
func (s *PermissionService) Create(req *types.PermissionRequest) (*types.Permission, error) {
	s.logger.Infow("Create permission request",
		"permission_code", req.Code,
		"permission_name", req.Name,
		"type", req.Type,
	)

	// Check if code already exists
	var count int64
	if err := s.db.DB.Model(&authmodel.Permission{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		s.logger.Errorw("Create permission failed: database error", "permission_code", req.Code, "error", err)
		return nil, errors.NewDatabaseError(err)
	}
	if count > 0 {
		s.logger.Warnw("Create permission failed: code already exists", "permission_code", req.Code)
		return nil, errors.ErrAlreadyExists.WithMessage("permission code already exists").KV("permission_code", req.Code)
	}

	// Validate parent exists if provided
	if req.ParentID != "" {
		var parentCount int64
		if err := s.db.DB.Model(&authmodel.Permission{}).Where("id = ?", req.ParentID).Count(&parentCount).Error; err != nil {
			s.logger.Errorw("Create permission failed: parent validation error", "parent_id", req.ParentID, "error", err)
			return nil, errors.NewDatabaseError(err)
		}
		if parentCount == 0 {
			s.logger.Warnw("Create permission failed: parent not found", "parent_id", req.ParentID)
			return nil, errors.ErrNotFound.WithMessage("parent permission not found").KV("parent_id", req.ParentID)
		}
	}

	// Validate type-specific fields
	if req.Type == "menu" && (req.Path == "" || req.Component == "") {
		s.logger.Warnw("Create permission failed: menu validation failed", "permission_code", req.Code)
		return nil, errors.ErrValidationFailed.WithMessage("menu type requires path and component")
	}
	if req.Type == "api" && (req.Path == "" || req.Method == "") {
		s.logger.Warnw("Create permission failed: api validation failed", "permission_code", req.Code)
		return nil, errors.ErrValidationFailed.WithMessage("api type requires path and method")
	}

	permID := uuid.New().String()
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	perm := &authmodel.Permission{
		ID:          permID,
		Name:        req.Name,
		Code:        req.Code,
		Type:        req.Type,
		Path:        req.Path,
		Method:      req.Method,
		Component:   req.Component,
		Icon:        req.Icon,
		Sort:        req.Sort,
		Status:      status,
		Description: req.Description,
	}

	if req.ParentID != "" {
		perm.ParentID = &req.ParentID
	}

	if err := s.db.DB.Create(perm).Error; err != nil {
		s.logger.Errorw("Create permission failed: insert error", "permission_code", req.Code, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	s.logger.Infow("Create permission successful",
		"permission_id", permID,
		"permission_code", req.Code,
		"type", req.Type,
	)

	return s.GetByID(permID)
}

// Update updates a permission.
func (s *PermissionService) Update(id string, req *types.PermissionRequest) error {
	s.logger.Infow("Update permission request", "permission_id", id)

	// Build update data map for non-empty fields
	updateData := make(map[string]interface{})

	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Type != "" {
		updateData["type"] = req.Type
	}
	if req.Path != "" {
		updateData["path"] = req.Path
	}
	if req.Method != "" {
		updateData["method"] = req.Method
	}
	if req.Component != "" {
		updateData["component"] = req.Component
	}
	if req.Icon != "" {
		updateData["icon"] = req.Icon
	}
	if req.Sort != 0 {
		updateData["sort"] = req.Sort
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}

	result := s.db.DB.Model(&authmodel.Permission{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		s.logger.Errorw("Update permission failed: update error", "permission_id", id, "error", result.Error)
		return errors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Warnw("Update permission failed: no rows affected", "permission_id", id)
		return errors.ErrNotFound.WithMessage("permission not found").KV("permission_id", id)
	}

	s.logger.Infow("Update permission successful", "permission_id", id)
	return nil
}

// Delete deletes a permission (checks if assigned to roles).
func (s *PermissionService) Delete(id string) error {
	s.logger.Infow("Delete permission request", "permission_id", id)

	// Check if permission is assigned to roles
	var roleCount int64
	if err := s.db.DB.Model(&authmodel.RolePermission{}).Where("permission_id = ?", id).Count(&roleCount).Error; err != nil {
		s.logger.Errorw("Delete permission failed: database error checking role usage", "permission_id", id, "error", err)
		return errors.NewDatabaseError(err)
	}
	if roleCount > 0 {
		s.logger.Warnw("Delete permission failed: permission in use by roles", "permission_id", id, "role_count", roleCount)
		return errors.ErrConflict.WithMessage("permission is assigned to roles, cannot delete").KV("permission_id", id)
	}

	// Check if permission has children
	var childCount int64
	if err := s.db.DB.Model(&authmodel.Permission{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		s.logger.Errorw("Delete permission failed: database error checking children", "permission_id", id, "error", err)
		return errors.NewDatabaseError(err)
	}
	if childCount > 0 {
		s.logger.Warnw("Delete permission failed: permission has children", "permission_id", id, "child_count", childCount)
		return errors.ErrConflict.WithMessage("permission has child permissions, cannot delete").KV("permission_id", id)
	}

	// Delete permission
	result := s.db.DB.Delete(&authmodel.Permission{}, "id = ?", id)
	if result.Error != nil {
		s.logger.Errorw("Delete permission failed: delete error", "permission_id", id, "error", result.Error)
		return errors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		s.logger.Warnw("Delete permission failed: no rows affected", "permission_id", id)
		return errors.ErrNotFound.WithMessage("permission not found").KV("permission_id", id)
	}

	s.logger.Infow("Delete permission successful", "permission_id", id)
	return nil
}
