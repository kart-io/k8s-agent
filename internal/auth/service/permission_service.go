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

// PermissionService handles permission management business logic.
type PermissionService struct {
	db *storage.MySQLDB
}

// NewPermissionService creates a new permission service.
func NewPermissionService(db *storage.MySQLDB) *PermissionService {
	return &PermissionService{db: db}
}

// List retrieves all permissions with optional filtering.
func (s *PermissionService) List(typeFilter, statusFilter string) ([]types.Permission, error) {
	query := s.db.DB.Model(&model.Permission{})

	if typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var modelPerms []model.Permission
	if err := query.Order("sort").Find(&modelPerms).Error; err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
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

// GetTree builds hierarchical permission tree.
func (s *PermissionService) GetTree() ([]*types.PermissionNode, error) {
	// Get root permissions (parent_id IS NULL) using GORM Preload
	var roots []model.Permission
	err := s.db.DB.Preload("Children").
		Where("parent_id IS NULL AND status = ?", 1).
		Order("sort").
		Find(&roots).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}

	// Convert to PermissionNode tree
	tree := make([]*types.PermissionNode, len(roots))
	for i := range roots {
		tree[i] = convertToPermissionNode(&roots[i])
	}

	return tree, nil
}

// convertToPermissionNode recursively converts model.Permission to types.PermissionNode.
func convertToPermissionNode(perm *model.Permission) *types.PermissionNode {
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
	var perm model.Permission
	err := s.db.DB.Where("id = ?", id).First(&perm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query permission: %w", err)
	}

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
	// Check if code already exists
	var count int64
	if err := s.db.DB.Model(&model.Permission{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check permission code: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("permission code already exists")
	}

	// Validate parent exists if provided
	if req.ParentID != "" {
		var parentCount int64
		if err := s.db.DB.Model(&model.Permission{}).Where("id = ?", req.ParentID).Count(&parentCount).Error; err != nil {
			return nil, fmt.Errorf("failed to check parent permission: %w", err)
		}
		if parentCount == 0 {
			return nil, fmt.Errorf("parent permission not found")
		}
	}

	// Validate type-specific fields
	if req.Type == "menu" && (req.Path == "" || req.Component == "") {
		return nil, fmt.Errorf("menu type requires path and component")
	}
	if req.Type == "api" && (req.Path == "" || req.Method == "") {
		return nil, fmt.Errorf("api type requires path and method")
	}

	permID := uuid.New().String()
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	perm := &model.Permission{
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
		return nil, fmt.Errorf("failed to insert permission: %w", err)
	}

	return s.GetByID(permID)
}

// Update updates a permission.
func (s *PermissionService) Update(id string, req *types.PermissionRequest) error {
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

	result := s.db.DB.Model(&model.Permission{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		return fmt.Errorf("failed to update permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

// Delete deletes a permission (checks if assigned to roles).
func (s *PermissionService) Delete(id string) error {
	// Check if permission is assigned to roles
	var roleCount int64
	if err := s.db.DB.Model(&model.RolePermission{}).Where("permission_id = ?", id).Count(&roleCount).Error; err != nil {
		return fmt.Errorf("failed to check permission usage: %w", err)
	}
	if roleCount > 0 {
		return fmt.Errorf("permission is assigned to %d role(s), cannot delete", roleCount)
	}

	// Check if permission has children
	var childCount int64
	if err := s.db.DB.Model(&model.Permission{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to check child permissions: %w", err)
	}
	if childCount > 0 {
		return fmt.Errorf("permission has %d child permission(s), cannot delete", childCount)
	}

	// Delete permission
	result := s.db.DB.Delete(&model.Permission{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}
