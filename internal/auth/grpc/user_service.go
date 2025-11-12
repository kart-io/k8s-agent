package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/types"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	commonpb "github.com/kart-io/k8s-agent/pkg/api/common/pagination/v1"
	"github.com/kart-io/logger/core"
)

// UserServiceServer implements the UserService gRPC service.
type UserServiceServer struct {
	authv1.UnimplementedUserServiceServer
	userService *service.UserService
	logger      core.Logger
}

// NewUserServiceServer creates a new UserService gRPC server.
func NewUserServiceServer(
	userService *service.UserService,
	logger core.Logger,
) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
		logger:      logger,
	}
}

// GetUser retrieves a user by ID.
func (s *UserServiceServer) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.User, error) {
	s.logger.Debugw("gRPC GetUser request", "user_id", req.UserId)

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := s.userService.GetByID(idToString(req.UserId))
	if err != nil {
		s.logger.Errorw("GetUser failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &authv1.User{
		Id:        req.UserId,
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.RealName,
		Phone:     user.Phone,
		Status:    int32(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// ListUsers retrieves a paginated list of users.
func (s *UserServiceServer) ListUsers(ctx context.Context, req *authv1.ListUsersRequest) (*authv1.ListUsersResponse, error) {
	s.logger.Debugw("gRPC ListUsers request")

	// Set defaults
	page := int(1)
	pageSize := int(20)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	// Extract status filter
	var statusFilter *int
	if req.Status != nil {
		status := int(*req.Status)
		statusFilter = &status
	}

	// Call service
	paginatedResp, err := s.userService.List(types.PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Sort:     "created_at",
		Order:    "desc",
	}, statusFilter)
	if err != nil {
		s.logger.Errorw("ListUsers failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	// Convert users to protobuf format
	usersSlice, ok := paginatedResp.Items.([]types.User)
	if !ok {
		s.logger.Errorw("ListUsers failed to convert items to []types.User")
		return nil, status.Error(codes.Internal, "failed to convert user list")
	}

	users := make([]*authv1.User, len(usersSlice))
	for i, u := range usersSlice {
		users[i] = &authv1.User{
			Id:        stringToID(u.ID),
			Username:  u.Username,
			Email:     u.Email,
			Nickname:  u.RealName,
			Phone:     u.Phone,
			Status:    int32(u.Status),
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt),
		}
	}

	return &authv1.ListUsersResponse{
		Users: users,
		Pagination: &commonpb.PaginationMetadata{
			Total:      paginatedResp.Total,
			PageSize:   int32(paginatedResp.PageSize),
			Page:       int32(paginatedResp.Page),
			TotalPages: int32(paginatedResp.TotalPages),
		},
	}, nil
}

// CreateUser creates a new user.
func (s *UserServiceServer) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.User, error) {
	s.logger.Debugw("gRPC CreateUser request", "username", req.Username)

	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// Convert role IDs
	roleIDs := make([]string, len(req.RoleIds))
	for i, id := range req.RoleIds {
		roleIDs[i] = idToString(id)
	}

	user, err := s.userService.Create(&types.UserCreateRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		RealName: req.Nickname,
		Phone:    req.Phone,
		RoleIDs:  roleIDs,
	})
	if err != nil {
		s.logger.Errorw("CreateUser failed", "username", req.Username, "error", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &authv1.User{
		Id:        stringToID(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		Nickname:  user.RealName,
		Phone:     user.Phone,
		Status:    int32(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// UpdateUser updates a user's information.
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.User, error) {
	s.logger.Debugw("gRPC UpdateUser request", "user_id", req.UserId)

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Convert role IDs
	var roleIDs []string
	if len(req.RoleIds) > 0 {
		roleIDs = make([]string, len(req.RoleIds))
		for i, id := range req.RoleIds {
			roleIDs[i] = idToString(id)
		}
	}

	// Build update request
	updateReq := &types.UserUpdateRequest{}
	if req.Email != nil {
		updateReq.Email = *req.Email
	}
	if req.Nickname != nil {
		updateReq.RealName = *req.Nickname
	}
	if req.Phone != nil {
		updateReq.Phone = *req.Phone
	}
	if req.Status != nil {
		status := int(*req.Status)
		updateReq.Status = &status
	}
	if len(roleIDs) > 0 {
		updateReq.RoleIDs = roleIDs
	}

	if err := s.userService.Update(idToString(req.UserId), updateReq); err != nil {
		s.logger.Errorw("UpdateUser failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	// Fetch updated user
	return s.GetUser(ctx, &authv1.GetUserRequest{UserId: req.UserId})
}

// DeleteUser deletes a user.
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	s.logger.Debugw("gRPC DeleteUser request", "user_id", req.UserId)

	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := s.userService.Delete(idToString(req.UserId)); err != nil {
		s.logger.Errorw("DeleteUser failed", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	return &authv1.DeleteUserResponse{
		Message: "User deleted successfully",
	}, nil
}

// Helper function to convert uint64 to string ID
func idToString(id uint64) string {
	// In production, implement proper ID conversion based on your ID scheme
	// For UUID-based IDs, you may need to maintain an ID mapping
	// For now, using a simple placeholder
	return "placeholder_id" // TODO: Implement proper ID conversion
}
