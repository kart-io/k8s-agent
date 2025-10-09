package service

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/auth-service/internal/config"
	"github.com/kart-io/k8s-agent/auth-service/internal/model"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	"github.com/kart-io/k8s-agent/auth-service/pkg/crypto"
	"github.com/kart-io/k8s-agent/auth-service/pkg/jwt"
	"github.com/kart-io/k8s-agent/auth-service/pkg/types"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	db    *storage.PostgresDB
	redis *storage.RedisClient
	cfg   *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(db *storage.PostgresDB, redis *storage.RedisClient, cfg *config.Config) *AuthService {
	return &AuthService{
		db:    db,
		redis: redis,
		cfg:   cfg,
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (*types.LoginResponse, error) {
	// Find user by username using GORM with Preload for roles
	var user model.User
	err := s.db.DB.Preload("Roles").
		Where("username = ? AND status = ?", username, 1).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("invalid username or password")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Verify password
	if err := crypto.CheckPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Convert model.Role to types.Role
	roles := make([]types.Role, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Status:      role.Status,
			Sort:        role.Sort,
		}
	}

	// Generate JWT token
	token, expiresAt, err := jwt.GenerateToken(user.ID, user.Username, s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Build response
	userInfo := &types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Avatar:   user.Avatar,
		Roles:    roles,
	}

	return &types.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      userInfo,
	}, nil
}

// Logout blacklists a JWT token
func (s *AuthService) Logout(token string) error {
	// Calculate TTL based on token expiration
	claims, err := jwt.ValidateToken(token, s.cfg.JWT.Secret)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Add token to Redis blacklist
	ctx := context.Background()
	if err := s.redis.BlacklistToken(ctx, token, ttl); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// GetCurrentUser retrieves user information by user ID
func (s *AuthService) GetCurrentUser(userID string) (*types.UserInfo, error) {
	var user model.User
	err := s.db.DB.Preload("Roles").
		Where("id = ? AND status = ?", userID, 1).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Convert model.Role to types.Role
	roles := make([]types.Role, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Status:      role.Status,
			Sort:        role.Sort,
		}
	}

	return &types.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Avatar:   user.Avatar,
		Roles:    roles,
	}, nil
}

// GetUserMenus builds hierarchical menu tree from user permissions
func (s *AuthService) GetUserMenus(userID string) ([]*types.MenuItem, error) {
	// Get all menu permissions for the user using GORM Joins
	var permissions []model.Permission
	err := s.db.DB.Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.type = ? AND permissions.status = ?", userID, "menu", 1).
		Order("permissions.sort").
		Find(&permissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query menu permissions: %w", err)
	}

	// Convert model.Permission to types.MenuItem
	var menus []*types.MenuItem
	for _, perm := range permissions {
		menu := &types.MenuItem{
			ID:        perm.ID,
			Name:      perm.Name,
			Path:      perm.Path,
			Component: perm.Component,
			Icon:      perm.Icon,
			Sort:      perm.Sort,
		}
		if perm.ParentID != nil {
			menu.ParentID = *perm.ParentID
		}
		menus = append(menus, menu)
	}

	// Build tree structure
	return buildMenuTree(menus), nil
}

// getUserRoles retrieves roles for a user (helper method, kept for compatibility)
func (s *AuthService) getUserRoles(userID string) ([]types.Role, error) {
	var user model.User
	err := s.db.DB.Preload("Roles", "status = ?", 1).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	// Convert model.Role to types.Role
	roles := make([]types.Role, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = types.Role{
			ID:          role.ID,
			Name:        role.Name,
			Code:        role.Code,
			Description: role.Description,
			Status:      role.Status,
			Sort:        role.Sort,
		}
	}

	return roles, nil
}

// buildMenuTree builds hierarchical menu structure
func buildMenuTree(menus []*types.MenuItem) []*types.MenuItem {
	menuMap := make(map[string]*types.MenuItem)
	var roots []*types.MenuItem

	// First pass: create map
	for _, menu := range menus {
		menuMap[menu.ID] = menu
	}

	// Second pass: build hierarchy
	for _, menu := range menus {
		if menu.ParentID == "" {
			roots = append(roots, menu)
		} else {
			if parent, exists := menuMap[menu.ParentID]; exists {
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return roots
}
