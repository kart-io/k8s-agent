package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	"github.com/kart-io/k8s-agent/internal/auth/crypto"
	"github.com/kart-io/k8s-agent/internal/auth/jwt"
	"github.com/kart-io/k8s-agent/internal/auth/model"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// AuthService handles authentication business logic.
type AuthService struct {
	db    *storage.MySQLDB
	redis *storage.RedisClient
	cfg   *options.ServerOptions
}

// NewAuthService creates a new auth service.
func NewAuthService(db *storage.MySQLDB, redis *storage.RedisClient, cfg *options.ServerOptions) *AuthService {
	return &AuthService{
		db:    db,
		redis: redis,
		cfg:   cfg,
	}
}

// Login authenticates a user and returns JWT tokens (access + refresh).
func (s *AuthService) Login(username, password string) (*types.LoginResponse, error) {
	// Find user by username using GORM with Preload for roles
	var user model.User
	err := s.db.DB.Preload("Roles").
		Where("username = ? AND status = ?", username, 1).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// Generate JWT token pair (access + refresh)
	tokenPair, err := jwt.GenerateTokenPair(user.ID, user.Username, s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Store refresh token in Redis
	ctx := context.Background()
	refreshClaims, _ := jwt.ValidateRefreshToken(tokenPair.RefreshToken, s.cfg.JWT.Secret)
	refreshTTL := time.Until(tokenPair.RefreshTokenExpiresAt)
	if err := s.redis.StoreRefreshToken(ctx, refreshClaims.ID, user.ID, refreshTTL); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
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

	// Calculate expires_in (seconds until expiration)
	expiresIn := int(time.Until(tokenPair.AccessTokenExpiresAt).Seconds())

	return &types.LoginResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		JTI:          refreshClaims.ID,
		ExpiresAt:    tokenPair.AccessTokenExpiresAt,
		ExpiresIn:    expiresIn,
		User:         userInfo,
	}, nil
}

// Logout blacklists a JWT token.
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

// GetCurrentUser retrieves user information by user ID.
func (s *AuthService) GetCurrentUser(userID string) (*types.UserInfo, error) {
	var user model.User
	err := s.db.DB.Preload("Roles").
		Where("id = ? AND status = ?", userID, 1).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
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

// GetUserMenus builds hierarchical menu tree from user permissions.
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

// buildMenuTree builds hierarchical menu structure.
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

// RefreshToken refreshes an access token using a valid refresh token.
// Implements token rotation: the old refresh token is revoked and a new one is issued.
func (s *AuthService) RefreshToken(refreshToken string) (*types.RefreshTokenResponse, error) {
	ctx := context.Background()

	// 1. Validate refresh token
	claims, err := jwt.ValidateRefreshToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 2. Check if refresh token is blacklisted
	isBlacklisted, err := s.redis.IsRefreshTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	// 3. Verify refresh token exists in Redis and matches user
	storedUserID, err := s.redis.GetRefreshTokenOwner(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired: %w", err)
	}
	if storedUserID != claims.UserID {
		return nil, fmt.Errorf("refresh token does not belong to the user")
	}

	// 4. Verify user still exists and is active
	var user model.User
	err = s.db.DB.Where("id = ? AND status = ?", claims.UserID, 1).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user not found or disabled")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 5. Generate new token pair (token rotation)
	newTokenPair, err := jwt.GenerateTokenPair(user.ID, user.Username, s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	// 6. Store new refresh token in Redis
	newRefreshClaims, _ := jwt.ValidateRefreshToken(newTokenPair.RefreshToken, s.cfg.JWT.Secret)
	refreshTTL := time.Until(newTokenPair.RefreshTokenExpiresAt)
	if err := s.redis.StoreRefreshToken(ctx, newRefreshClaims.ID, user.ID, refreshTTL); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// 7. Revoke old refresh token (token rotation)
	// We blacklist it instead of just deleting to prevent replay attacks
	oldTokenTTL := time.Until(claims.ExpiresAt.Time)
	if oldTokenTTL > 0 {
		if err := s.redis.BlacklistRefreshToken(ctx, claims.ID, oldTokenTTL); err != nil {
			// Log error but don't fail the refresh operation
			// The old token will naturally expire
		}
	}
	// Also delete from active tokens
	_ = s.redis.RevokeRefreshToken(ctx, claims.ID)

	// 8. Calculate expires_in
	expiresIn := int(time.Until(newTokenPair.AccessTokenExpiresAt).Seconds())

	return &types.RefreshTokenResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresAt:    newTokenPair.AccessTokenExpiresAt,
		ExpiresIn:    expiresIn,
	}, nil
}
