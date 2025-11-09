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

// RoleServiceServer implements the RoleService gRPC service.
type RoleServiceServer struct {
	authv1.UnimplementedRoleServiceServer
	roleService *service.RoleService
	logger      core.Logger
}

// NewRoleServiceServer creates a new RoleService gRPC server.
func NewRoleServiceServer(
	roleService *service.RoleService,
	logger core.Logger,
) *RoleServiceServer {
	return &RoleServiceServer{
		roleService: roleService,
		logger:      logger,
	}
}

// GetRole retrieves a role by ID.
func (s *RoleServiceServer) GetRole(ctx context.Context, req *authv1.GetRoleRequest) (*authv1.Role, error) {
	s.logger.Debugw("gRPC GetRole request", "role_id", req.RoleId)

	if req.RoleId == 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}

	role, err := s.roleService.GetByID(idToString(req.RoleId))
	if err != nil {
		s.logger.Errorw("GetRole failed", "role_id", req.RoleId, "error", err)
		return nil, status.Error(codes.NotFound, "role not found")
	}

	return &authv1.Role{
		Id:          req.RoleId,
		Name:        role.Name,
		Description: role.Description,
		Status:      int32(role.Status),
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}, nil
}

// ListRoles retrieves a list of roles.
func (s *RoleServiceServer) ListRoles(ctx context.Context, req *authv1.ListRolesRequest) (*authv1.ListRolesResponse, error) {
	s.logger.Debugw("gRPC ListRoles request")

	roles, err := s.roleService.List()
	if err != nil {
		s.logger.Errorw("ListRoles failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to list roles")
	}

	// Convert to protobuf format
	pbRoles := make([]*authv1.Role, len(roles))
	for i, r := range roles {
		pbRoles[i] = &authv1.Role{
			Id:          uint64(stringToID(r.ID)),
			Name:        r.Name,
			Description: r.Description,
			Status:      int32(r.Status),
			CreatedAt:   timestamppb.New(r.CreatedAt),
			UpdatedAt:   timestamppb.New(r.UpdatedAt),
		}
	}

	return &authv1.ListRolesResponse{
		Roles: pbRoles,
		Pagination: &commonpb.PaginationMetadata{
			Total:      int64(len(roles)),
			PageSize:   int32(len(roles)),
			Page:       1,
			TotalPages: 1,
		},
	}, nil
}

// CreateRole creates a new role.
func (s *RoleServiceServer) CreateRole(ctx context.Context, req *authv1.CreateRoleRequest) (*authv1.Role, error) {
	s.logger.Debugw("gRPC CreateRole request", "name", req.Name)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	role, err := s.roleService.Create(&types.RoleRequest{
		Name:        req.Name,
		Code:        req.Name, // Use name as code for simplicity
		Description: req.Description,
	})
	if err != nil {
		s.logger.Errorw("CreateRole failed", "name", req.Name, "error", err)
		return nil, status.Error(codes.Internal, "failed to create role")
	}

	return &authv1.Role{
		Id:          uint64(stringToID(role.ID)),
		Name:        role.Name,
		Description: role.Description,
		Status:      int32(role.Status),
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}, nil
}

// UpdateRole updates a role.
func (s *RoleServiceServer) UpdateRole(ctx context.Context, req *authv1.UpdateRoleRequest) (*authv1.Role, error) {
	s.logger.Debugw("gRPC UpdateRole request", "role_id", req.RoleId)

	if req.RoleId == 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}

	updateReq := &types.RoleRequest{}
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Description != nil {
		updateReq.Description = *req.Description
	}
	if req.Status != nil {
		st := int(*req.Status)
		updateReq.Status = &st
	}

	if err := s.roleService.Update(idToString(req.RoleId), updateReq); err != nil {
		s.logger.Errorw("UpdateRole failed", "role_id", req.RoleId, "error", err)
		return nil, status.Error(codes.Internal, "failed to update role")
	}

	// Fetch updated role
	return s.GetRole(ctx, &authv1.GetRoleRequest{RoleId: req.RoleId})
}

// DeleteRole deletes a role.
func (s *RoleServiceServer) DeleteRole(ctx context.Context, req *authv1.DeleteRoleRequest) (*authv1.DeleteRoleResponse, error) {
	s.logger.Debugw("gRPC DeleteRole request", "role_id", req.RoleId)

	if req.RoleId == 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}

	if err := s.roleService.Delete(idToString(req.RoleId)); err != nil {
		s.logger.Errorw("DeleteRole failed", "role_id", req.RoleId, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete role")
	}

	return &authv1.DeleteRoleResponse{
		Message: "Role deleted successfully",
	}, nil
}

// AssignPermissions assigns permissions to a role.
func (s *RoleServiceServer) AssignPermissions(ctx context.Context, req *authv1.AssignPermissionsRequest) (*authv1.AssignPermissionsResponse, error) {
	s.logger.Debugw("gRPC AssignPermissions request", "role_id", req.RoleId, "permission_count", len(req.PermissionIds))

	if req.RoleId == 0 {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}

	// Convert permission IDs
	permissionIDs := make([]string, len(req.PermissionIds))
	for i, id := range req.PermissionIds {
		permissionIDs[i] = idToString(id)
	}

	if err := s.roleService.AssignPermissions(idToString(req.RoleId), permissionIDs); err != nil {
		s.logger.Errorw("AssignPermissions failed", "role_id", req.RoleId, "error", err)
		return nil, status.Error(codes.Internal, "failed to assign permissions")
	}

	return &authv1.AssignPermissionsResponse{
		Message: "Permissions assigned successfully",
	}, nil
}
