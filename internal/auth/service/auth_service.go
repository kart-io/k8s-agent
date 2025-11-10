package service

import (
	"context"
	stderrors "errors"
	"time"

	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/errors"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	authmodel "github.com/kart-io/k8s-agent/internal/models/auth"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/auth/crypto"
	"github.com/kart-io/k8s-agent/pkg/auth/jwt"
	"github.com/kart-io/logger/core"
)

// AuthService handles authentication business logic.
type AuthService struct {
	db             *storage.MySQLDB
	redis          *storage.RedisClient
	cfg            *commonapp.StandardOptions
	logger         core.Logger
	sessionService SessionServiceInterface
}

// SessionServiceInterface defines the interface for session operations.
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, session *types.SessionInfo) error
	ValidateSession(ctx context.Context, jti string) (bool, error)
}

// NewAuthService creates a new auth service.
func NewAuthService(db *storage.MySQLDB, redis *storage.RedisClient, cfg *commonapp.StandardOptions, logger core.Logger) *AuthService {
	return &AuthService{
		db:     db,
		redis:  redis,
		cfg:    cfg,
		logger: logger.With("component", "auth-service"),
	}
}

// SetSessionService sets the session service dependency (called after initialization).
func (s *AuthService) SetSessionService(sessionService SessionServiceInterface) {
	s.sessionService = sessionService
}

// Login authenticates a user and returns JWT tokens (access + refresh).
func (s *AuthService) Login(username, password string) (*types.LoginResponse, error) {
	s.logger.Infow("Login request", "username", username)

	// Find user by username using GORM with Preload for roles
	var user authmodel.User
	err := s.db.DB.Preload("Roles").
		Where("username = ? AND status = ?", username, 1).
		First(&user).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Login failed: user not found", "username", username)
		return nil, errors.ErrUnauthorized.WithMessage("invalid username or password")
	}
	if err != nil {
		s.logger.Errorw("Login failed: database error", "username", username, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Verify password
	if err := crypto.CheckPassword(user.Password, password); err != nil {
		s.logger.Warnw("Login failed: invalid password", "username", username, "user_id", user.ID)
		return nil, errors.ErrUnauthorized.WithMessage("invalid username or password")
	}

	// Convert authmodel.Role to types.Role
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
		s.logger.Errorw("Login failed: token generation error", "user_id", user.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to generate token pair")
	}

	// Store refresh token in Redis
	ctx := context.Background()
	refreshClaims, _ := jwt.ValidateRefreshToken(tokenPair.RefreshToken, s.cfg.JWT.Secret)
	refreshTTL := time.Until(tokenPair.RefreshTokenExpiresAt)
	if err := s.redis.StoreRefreshToken(ctx, refreshClaims.ID, user.ID, refreshTTL); err != nil {
		s.logger.Errorw("Login failed: redis error", "user_id", user.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to store refresh token")
	}

	// Parse access token to get JTI and create session
	accessClaims, err := jwt.ValidateToken(tokenPair.AccessToken, s.cfg.JWT.Secret)
	if err != nil {
		s.logger.Errorw("Login failed: failed to parse access token", "user_id", user.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to parse access token")
	}

	// Create session in Redis for the access token
	sessionInfo := &types.SessionInfo{
		JTI:            accessClaims.ID, // Use access token's JTI
		UserID:         user.ID,
		Username:       user.Username,
		Email:          user.Email,
		IPAddress:      "", // TODO: Get from request context
		UserAgent:      "", // TODO: Get from request headers
		DeviceType:     "", // Will be set by session service
		DeviceName:     "", // Will be set by session service
		Location:       "", // TODO: Get from IP geolocation
		LoginAt:        time.Now(),
		LastActivityAt: time.Now(),
		ExpiresAt:      tokenPair.AccessTokenExpiresAt,
	}

	// Only create session if session service is available
	if s.sessionService != nil {
		if err := s.sessionService.CreateSession(ctx, sessionInfo); err != nil {
			s.logger.Errorw("Login failed: failed to create session", "user_id", user.ID, "error", err)
			return nil, errors.ErrInternalError.WithMessage("failed to create session")
		}
		s.logger.Infow("Session created successfully", "jti", accessClaims.ID, "user_id", user.ID)
	} else {
		s.logger.Warnw("Session service not available, skipping session creation", "user_id", user.ID)
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

	s.logger.Infow("Login successful",
		"user_id", user.ID,
		"username", username,
		"roles_count", len(roles),
		"expires_in", expiresIn,
	)

	return &types.LoginResponse{
		Token:        tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		JTI:          accessClaims.ID, // Use access token's JTI, not refresh token's
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
		s.logger.Warnw("Logout failed: invalid token", "error", err)
		return errors.ErrUnauthorized.WithMessage("invalid token")
	}

	s.logger.Infow("Logout request", "user_id", claims.UserID, "jti", claims.ID)

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		s.logger.Infow("Logout skipped: token already expired", "user_id", claims.UserID)
		return nil
	}

	// Add token to Redis blacklist
	ctx := context.Background()
	if err := s.redis.BlacklistToken(ctx, token, ttl); err != nil {
		s.logger.Errorw("Logout failed: redis error", "user_id", claims.UserID, "error", err)
		return errors.ErrInternalError.WithMessage("failed to blacklist token")
	}

	s.logger.Infow("Logout successful", "user_id", claims.UserID, "jti", claims.ID)
	return nil
}

// GetCurrentUser retrieves user information by user ID.
func (s *AuthService) GetCurrentUser(userID string) (*types.UserInfo, error) {
	s.logger.Infow("Get current user request", "user_id", userID)

	var user authmodel.User
	err := s.db.DB.Preload("Roles").
		Where("id = ? AND status = ?", userID, 1).
		First(&user).Error

	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Get current user failed: user not found", "user_id", userID)
		return nil, errors.ErrNotFound.WithMessage("user not found").KV("user_id", userID)
	}
	if err != nil {
		s.logger.Errorw("Get current user failed: database error", "user_id", userID, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert authmodel.Role to types.Role
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

	s.logger.Infow("Get current user successful", "user_id", userID, "username", user.Username)

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
	s.logger.Infow("Get user menus request", "user_id", userID)

	// Get all menu permissions for the user using GORM Joins
	var permissions []authmodel.Permission
	err := s.db.DB.Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.type = ? AND permissions.status = ?", userID, "menu", 1).
		Order("permissions.sort").
		Find(&permissions).Error
	if err != nil {
		s.logger.Errorw("Get user menus failed: database error", "user_id", userID, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// Convert authmodel.Permission to types.MenuItem
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
	tree := buildMenuTree(menus)

	s.logger.Infow("Get user menus successful",
		"user_id", userID,
		"menus_count", len(menus),
		"root_menus_count", len(tree),
	)

	return tree, nil
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

	s.logger.Infow("Refresh token request")

	// 1. Validate refresh token
	claims, err := jwt.ValidateRefreshToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		s.logger.Warnw("Refresh token failed: invalid token", "error", err)
		return nil, errors.ErrUnauthorized.WithMessage("invalid refresh token")
	}

	// 2. Check if refresh token is blacklisted
	isBlacklisted, err := s.redis.IsRefreshTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		s.logger.Errorw("Refresh token failed: redis error", "jti", claims.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to check token blacklist")
	}
	if isBlacklisted {
		s.logger.Warnw("Refresh token failed: token revoked", "jti", claims.ID, "user_id", claims.UserID)
		return nil, errors.ErrUnauthorized.WithMessage("refresh token has been revoked")
	}

	// 3. Verify refresh token exists in Redis and matches user
	storedUserID, err := s.redis.GetRefreshTokenOwner(ctx, claims.ID)
	if err != nil {
		s.logger.Warnw("Refresh token failed: token not found", "jti", claims.ID, "error", err)
		return nil, errors.ErrUnauthorized.WithMessage("refresh token not found or expired")
	}
	if storedUserID != claims.UserID {
		s.logger.Warnw("Refresh token failed: user mismatch",
			"jti", claims.ID,
			"token_user", claims.UserID,
			"stored_user", storedUserID,
		)
		return nil, errors.ErrUnauthorized.WithMessage("refresh token does not belong to the user")
	}

	// 4. Verify user still exists and is active
	var user authmodel.User
	err = s.db.DB.Where("id = ? AND status = ?", claims.UserID, 1).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Warnw("Refresh token failed: user not found or disabled", "user_id", claims.UserID)
		return nil, errors.ErrUnauthorized.WithMessage("user not found or disabled")
	}
	if err != nil {
		s.logger.Errorw("Refresh token failed: database error", "user_id", claims.UserID, "error", err)
		return nil, errors.NewDatabaseError(err)
	}

	// 5. Generate new token pair (token rotation)
	newTokenPair, err := jwt.GenerateTokenPair(user.ID, user.Username, s.cfg.JWT.Secret, s.cfg.JWT.ExpiresHours)
	if err != nil {
		s.logger.Errorw("Refresh token failed: token generation error", "user_id", user.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to generate new token pair")
	}

	// 6. Store new refresh token in Redis
	newRefreshClaims, _ := jwt.ValidateRefreshToken(newTokenPair.RefreshToken, s.cfg.JWT.Secret)
	refreshTTL := time.Until(newTokenPair.RefreshTokenExpiresAt)
	if err := s.redis.StoreRefreshToken(ctx, newRefreshClaims.ID, user.ID, refreshTTL); err != nil {
		s.logger.Errorw("Refresh token failed: redis error", "user_id", user.ID, "error", err)
		return nil, errors.ErrInternalError.WithMessage("failed to store new refresh token")
	}

	// 7. Revoke old refresh token (token rotation)
	// We blacklist it instead of just deleting to prevent replay attacks
	oldTokenTTL := time.Until(claims.ExpiresAt.Time)
	if oldTokenTTL > 0 {
		if err := s.redis.BlacklistRefreshToken(ctx, claims.ID, oldTokenTTL); err != nil {
			// Log error but don't fail the refresh operation
			// The old token will naturally expire
			s.logger.Warnw("Failed to blacklist old refresh token", "jti", claims.ID, "error", err)
		}
	}
	// Also delete from active tokens
	_ = s.redis.RevokeRefreshToken(ctx, claims.ID)

	// 8. Calculate expires_in
	expiresIn := int(time.Until(newTokenPair.AccessTokenExpiresAt).Seconds())

	s.logger.Infow("Refresh token successful",
		"user_id", user.ID,
		"old_jti", claims.ID,
		"new_jti", newRefreshClaims.ID,
		"expires_in", expiresIn,
	)

	return &types.RefreshTokenResponse{
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		ExpiresAt:    newTokenPair.AccessTokenExpiresAt,
		ExpiresIn:    expiresIn,
	}, nil
}
