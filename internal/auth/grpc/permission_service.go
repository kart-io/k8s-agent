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

// PermissionServiceServer implements the PermissionService gRPC service.
type PermissionServiceServer struct {
	authv1.UnimplementedPermissionServiceServer
	permissionService *service.PermissionService
	logger            core.Logger
}

// NewPermissionServiceServer creates a new PermissionService gRPC server.
func NewPermissionServiceServer(
	permissionService *service.PermissionService,
	logger core.Logger,
) *PermissionServiceServer {
	return &PermissionServiceServer{
		permissionService: permissionService,
		logger:            logger,
	}
}

// GetPermission retrieves a permission by ID.
func (s *PermissionServiceServer) GetPermission(ctx context.Context, req *authv1.GetPermissionRequest) (*authv1.Permission, error) {
	s.logger.Debugw("gRPC GetPermission request", "permission_id", req.PermissionId)

	if req.PermissionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "permission_id is required")
	}

	perm, err := s.permissionService.GetByID(idToString(req.PermissionId))
	if err != nil {
		s.logger.Errorw("GetPermission failed", "permission_id", req.PermissionId, "error", err)
		return nil, status.Error(codes.NotFound, "permission not found")
	}

	return convertPermissionToProto(perm), nil
}

// ListPermissions retrieves a list of permissions.
func (s *PermissionServiceServer) ListPermissions(ctx context.Context, req *authv1.ListPermissionsRequest) (*authv1.ListPermissionsResponse, error) {
	s.logger.Debugw("gRPC ListPermissions request")

	perms, err := s.permissionService.List("", "")
	if err != nil {
		s.logger.Errorw("ListPermissions failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to list permissions")
	}

	// Convert to protobuf format
	pbPerms := make([]*authv1.Permission, len(perms))
	for i, p := range perms {
		pbPerms[i] = convertPermissionToProto(&p)
	}

	return &authv1.ListPermissionsResponse{
		Permissions: pbPerms,
		Pagination: &commonpb.PaginationMetadata{
			Total:      int64(len(perms)),
			PageSize:   int32(len(perms)),
			Page:       1,
			TotalPages: 1,
		},
	}, nil
}

// GetPermissionTree retrieves the permission tree.
func (s *PermissionServiceServer) GetPermissionTree(ctx context.Context, req *authv1.GetPermissionTreeRequest) (*authv1.GetPermissionTreeResponse, error) {
	s.logger.Debugw("gRPC GetPermissionTree request")

	tree, err := s.permissionService.GetTree()
	if err != nil {
		s.logger.Errorw("GetPermissionTree failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to get permission tree")
	}

	// Convert tree nodes to flat permission list with parent-child relationships
	perms := flattenPermissionTree(tree)

	return &authv1.GetPermissionTreeResponse{
		Permissions: perms,
	}, nil
}

// CreatePermission creates a new permission.
func (s *PermissionServiceServer) CreatePermission(ctx context.Context, req *authv1.CreatePermissionRequest) (*authv1.Permission, error) {
	s.logger.Debugw("gRPC CreatePermission request", "name", req.Name)

	if req.Name == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "name and code are required")
	}

	perm, err := s.permissionService.Create(&types.PermissionRequest{
		Name:        req.Name,
		Code:        req.Code,
		Type:        "api",
		Path:        req.Resource,
		Method:      req.Action,
		Description: req.Description,
		ParentID:    idToString(req.ParentId),
	})
	if err != nil {
		s.logger.Errorw("CreatePermission failed", "name", req.Name, "error", err)
		return nil, status.Error(codes.Internal, "failed to create permission")
	}

	return convertPermissionToProto(perm), nil
}

// UpdatePermission updates a permission.
func (s *PermissionServiceServer) UpdatePermission(ctx context.Context, req *authv1.UpdatePermissionRequest) (*authv1.Permission, error) {
	s.logger.Debugw("gRPC UpdatePermission request", "permission_id", req.PermissionId)

	if req.PermissionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "permission_id is required")
	}

	updateReq := &types.PermissionRequest{}
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Code != nil {
		updateReq.Code = *req.Code
	}
	if req.Resource != nil {
		updateReq.Path = *req.Resource
	}
	if req.Action != nil {
		updateReq.Method = *req.Action
	}
	if req.Description != nil {
		updateReq.Description = *req.Description
	}
	if req.ParentId != nil {
		updateReq.ParentID = idToString(*req.ParentId)
	}

	if err := s.permissionService.Update(idToString(req.PermissionId), updateReq); err != nil {
		s.logger.Errorw("UpdatePermission failed", "permission_id", req.PermissionId, "error", err)
		return nil, status.Error(codes.Internal, "failed to update permission")
	}

	// Fetch updated permission
	return s.GetPermission(ctx, &authv1.GetPermissionRequest{PermissionId: req.PermissionId})
}

// DeletePermission deletes a permission.
func (s *PermissionServiceServer) DeletePermission(ctx context.Context, req *authv1.DeletePermissionRequest) (*authv1.DeletePermissionResponse, error) {
	s.logger.Debugw("gRPC DeletePermission request", "permission_id", req.PermissionId)

	if req.PermissionId == 0 {
		return nil, status.Error(codes.InvalidArgument, "permission_id is required")
	}

	if err := s.permissionService.Delete(idToString(req.PermissionId)); err != nil {
		s.logger.Errorw("DeletePermission failed", "permission_id", req.PermissionId, "error", err)
		return nil, status.Error(codes.Internal, "failed to delete permission")
	}

	return &authv1.DeletePermissionResponse{
		Message: "Permission deleted successfully",
	}, nil
}

// Helper functions

// convertPermissionToProto converts internal permission to protobuf format.
func convertPermissionToProto(p *types.Permission) *authv1.Permission {
	perm := &authv1.Permission{
		Id:          stringToID(p.ID),
		Name:        p.Name,
		Code:        p.Code,
		Resource:    p.Path,
		Action:      p.Method,
		Description: p.Description,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	if p.ParentID != "" {
		perm.ParentId = stringToID(p.ParentID)
	}

	return perm
}

// flattenPermissionTree flattens permission tree nodes to protobuf format.
func flattenPermissionTree(nodes []*types.PermissionNode) []*authv1.Permission {
	var result []*authv1.Permission

	var flatten func([]*types.PermissionNode)
	flatten = func(nodes []*types.PermissionNode) {
		for _, node := range nodes {
			perm := &authv1.Permission{
				Id:       stringToID(node.ID),
				Name:     node.Name,
				Code:     node.Code,
				Resource: node.Path,
				Action:   node.Method,
			}

			if node.ParentID != "" {
				perm.ParentId = stringToID(node.ParentID)
			}

			result = append(result, perm)

			// Recursively flatten children
			if len(node.Children) > 0 {
				flatten(node.Children)
			}
		}
	}

	flatten(nodes)
	return result
}
