package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	"github.com/kart-io/logger/core"
)

// AuthServiceServer implements the AuthService gRPC service.
type AuthServiceServer struct {
	authv1.UnimplementedAuthServiceServer
	authService *service.AuthService
	logger      core.Logger
}

// NewAuthServiceServer creates a new AuthService gRPC server.
func NewAuthServiceServer(
	authService *service.AuthService,
	logger core.Logger,
) *AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
		logger:      logger,
	}
}

// Login authenticates a user and returns JWT tokens.
func (s *AuthServiceServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	s.logger.Debugw("gRPC Login request", "username", req.Username)

	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	loginResp, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		s.logger.Errorw("Login failed", "username", req.Username, "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid username or password")
	}

	// Convert roles
	roles := make([]*authv1.Role, len(loginResp.User.Roles))
	for i, r := range loginResp.User.Roles {
		// Validate status value before conversion to prevent overflow
		roleStatus := r.Status
		if roleStatus < 0 {
			roleStatus = 0
		}
		if roleStatus > 0x7FFFFFFF { // Max int32
			roleStatus = 0x7FFFFFFF
		}

		roles[i] = &authv1.Role{
			Id:          stringToID(r.ID),
			Name:        r.Name,
			Description: r.Description,
			Status:      int32(roleStatus),
			CreatedAt:   timestamppb.New(r.CreatedAt),
			UpdatedAt:   timestamppb.New(r.UpdatedAt),
		}
	}

	user := &authv1.User{
		Id:       stringToID(loginResp.User.ID),
		Username: loginResp.User.Username,
		Email:    loginResp.User.Email,
		Nickname: loginResp.User.RealName,
		Roles:    roles,
	}

	return &authv1.LoginResponse{
		Token:     loginResp.Token,
		ExpiresAt: loginResp.ExpiresAt.Unix(),
		User:      user,
	}, nil
}

// Logout invalidates a user's session.
func (s *AuthServiceServer) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	s.logger.Debugw("gRPC Logout request", "token_prefix", truncateToken(req.Token))

	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := s.authService.Logout(req.Token); err != nil {
		s.logger.Errorw("Logout failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to logout")
	}

	return &authv1.LogoutResponse{
		Message: "Logout successful",
	}, nil
}

// RefreshToken refreshes an access token using a refresh token.
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	s.logger.Debugw("gRPC RefreshToken request")

	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	refreshResp, err := s.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Errorw("RefreshToken failed", "error", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	return &authv1.RefreshTokenResponse{
		Token:     refreshResp.AccessToken,
		ExpiresAt: refreshResp.ExpiresAt.Unix(),
	}, nil
}

// GetMe retrieves the current user's information.
func (s *AuthServiceServer) GetMe(ctx context.Context, req *authv1.GetMeRequest) (*authv1.User, error) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	s.logger.Debugw("gRPC GetMe request", "user_id", userID)

	userInfo, err := s.authService.GetCurrentUser(userID)
	if err != nil {
		s.logger.Errorw("GetMe failed", "user_id", userID, "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Convert roles
	roles := make([]*authv1.Role, len(userInfo.Roles))
	for i, r := range userInfo.Roles {
		// Validate status value before conversion to prevent overflow
		roleStatus := r.Status
		if roleStatus < 0 {
			roleStatus = 0
		}
		if roleStatus > 0x7FFFFFFF { // Max int32
			roleStatus = 0x7FFFFFFF
		}

		roles[i] = &authv1.Role{
			Id:          stringToID(r.ID),
			Name:        r.Name,
			Description: r.Description,
			Status:      int32(roleStatus),
			CreatedAt:   timestamppb.New(r.CreatedAt),
			UpdatedAt:   timestamppb.New(r.UpdatedAt),
		}
	}

	return &authv1.User{
		Id:       stringToID(userInfo.ID),
		Username: userInfo.Username,
		Email:    userInfo.Email,
		Nickname: userInfo.RealName,
		Roles:    roles,
	}, nil
}

// GetMenus retrieves the current user's menu permissions.
func (s *AuthServiceServer) GetMenus(ctx context.Context, req *authv1.GetMenusRequest) (*authv1.GetMenusResponse, error) {
	// Extract user ID from context
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	s.logger.Debugw("gRPC GetMenus request", "user_id", userID)

	menuItems, err := s.authService.GetUserMenus(userID)
	if err != nil {
		s.logger.Errorw("GetMenus failed", "user_id", userID, "error", err)
		return nil, status.Error(codes.Internal, "failed to retrieve menus")
	}

	// Convert to protobuf Menu
	menus := convertMenuItemsToProto(menuItems)

	return &authv1.GetMenusResponse{
		Menus: menus,
	}, nil
}

// CheckPermission checks if a user has permission for a specific resource and action.
func (s *AuthServiceServer) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
	s.logger.Debugw("gRPC CheckPermission request",
		"user_id", req.UserId,
		"resource", req.Resource,
		"action", req.Action,
	)

	// TODO: Implement permission check logic
	// This would typically query user's roles and their associated permissions
	return &authv1.CheckPermissionResponse{
		Allowed: true,
		Reason:  "Permission check not implemented yet",
	}, nil
}

// Helper functions

// convertMenuItemsToProto converts internal menu items to protobuf format.
func convertMenuItemsToProto(items []*types.MenuItem) []*authv1.Menu {
	if len(items) == 0 {
		return nil
	}

	menus := make([]*authv1.Menu, len(items))
	for i, item := range items {
		menu := &authv1.Menu{
			Id:   item.ID,
			Name: item.Name,
			Path: item.Path,
			Icon: item.Icon,
		}

		// Recursively convert children
		if len(item.Children) > 0 {
			menu.Children = convertMenuItemsToProto(item.Children)
		}

		menus[i] = menu
	}

	return menus
}

// truncateToken truncates token for logging (security).
func truncateToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:10] + "..."
}

// stringToID converts string ID to uint64 (simple hash for demo).
func stringToID(id string) uint64 {
	// In production, use proper ID conversion or keep IDs as strings
	hash := uint64(0)
	for _, c := range id {
		hash = hash*31 + uint64(c)
	}
	return hash
}
