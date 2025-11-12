package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	commonpb "github.com/kart-io/k8s-agent/pkg/api/common/pagination/v1"
	"github.com/kart-io/logger/core"
)

// SessionServiceServer implements the SessionService gRPC service.
type SessionServiceServer struct {
	authv1.UnimplementedSessionServiceServer
	sessionService *session.Service
	logger         core.Logger
}

// NewSessionServiceServer creates a new SessionService gRPC server.
func NewSessionServiceServer(
	sessionService *session.Service,
	logger core.Logger,
) *SessionServiceServer {
	return &SessionServiceServer{
		sessionService: sessionService,
		logger:         logger,
	}
}

// GetSession retrieves a session by ID.
func (s *SessionServiceServer) GetSession(ctx context.Context, req *authv1.GetSessionRequest) (*authv1.Session, error) {
	s.logger.Debugw("gRPC GetSession request", "session_id", req.SessionId)

	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	sessionInfo, err := s.sessionService.GetSession(ctx, req.SessionId)
	if err != nil {
		s.logger.Errorw("GetSession failed", "session_id", req.SessionId, "error", err)
		return nil, status.Error(codes.NotFound, "session not found")
	}

	return &authv1.Session{
		Id:        sessionInfo.JTI,
		UserId:    stringToID(sessionInfo.UserID),
		Username:  sessionInfo.Username,
		IpAddress: sessionInfo.IPAddress,
		UserAgent: sessionInfo.UserAgent,
		CreatedAt: timestamppb.New(sessionInfo.LoginAt),
		ExpiresAt: timestamppb.New(sessionInfo.ExpiresAt),
	}, nil
}

// ListSessions retrieves a list of sessions for a user.
func (s *SessionServiceServer) ListSessions(ctx context.Context, req *authv1.ListSessionsRequest) (*authv1.ListSessionsResponse, error) {
	s.logger.Debugw("gRPC ListSessions request")

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

	// Extract user ID filter
	var userID string
	if req.UserId != nil {
		userID = idToString(*req.UserId)
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Call service
	sessionList, err := s.sessionService.GetUserSessions(ctx, userID, pageSize, offset)
	if err != nil {
		s.logger.Errorw("ListSessions failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to list sessions")
	}

	// Convert to protobuf format
	sessions := make([]*authv1.Session, len(sessionList.Sessions))
	for i, sess := range sessionList.Sessions {
		sessions[i] = &authv1.Session{
			Id:        sess.JTI,
			UserId:    stringToID(sess.UserID),
			Username:  sess.Username,
			IpAddress: sess.IPAddress,
			UserAgent: sess.UserAgent,
			CreatedAt: timestamppb.New(sess.LoginAt),
			ExpiresAt: timestamppb.New(sess.ExpiresAt),
		}
	}

	totalPages := uint64(sessionList.Total / pageSize)
	if sessionList.Total%pageSize > 0 {
		totalPages++
	}

	return &authv1.ListSessionsResponse{
		Sessions: sessions,
		Pagination: &commonpb.PaginationMetadata{
			Total:      int64(sessionList.Total),
			PageSize:   int32(pageSize),
			Page:       int32(page),
			TotalPages: int32(totalPages),
		},
	}, nil
}

// InvalidateSession terminates a session.
func (s *SessionServiceServer) InvalidateSession(ctx context.Context, req *authv1.InvalidateSessionRequest) (*authv1.InvalidateSessionResponse, error) {
	s.logger.Debugw("gRPC InvalidateSession request", "session_id", req.SessionId)

	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id is required")
	}

	// Get session to find user ID
	sessionInfo, err := s.sessionService.GetSession(ctx, req.SessionId)
	if err != nil {
		s.logger.Errorw("InvalidateSession - session not found", "session_id", req.SessionId, "error", err)
		return nil, status.Error(codes.NotFound, "session not found")
	}

	// Terminate session
	if err := s.sessionService.TerminateSession(ctx, req.SessionId, sessionInfo.UserID, "admin", "manual_invalidation", ""); err != nil {
		s.logger.Errorw("InvalidateSession failed", "session_id", req.SessionId, "error", err)
		return nil, status.Error(codes.Internal, "failed to invalidate session")
	}

	return &authv1.InvalidateSessionResponse{
		Message: "Session invalidated successfully",
	}, nil
}
