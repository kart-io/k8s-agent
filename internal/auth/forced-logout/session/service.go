package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// Service provides business logic for session management.
type Service struct {
	repo Repository
}

// NewService creates a new session service.
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateSession validates and stores a new session.
func (s *Service) CreateSession(ctx context.Context, session *types.SessionInfo) error {
	// Validate required fields
	if session.JTI == "" || session.UserID == "" {
		return fmt.Errorf("jti and user_id are required")
	}

	// Enrich session with device type detection
	session.DeviceType = detectDeviceType(session.UserAgent)
	session.DeviceName = parseDeviceName(session.UserAgent)

	// Store in Redis
	return s.repo.StoreSession(ctx, session)
}

// GetSession retrieves a single session by JTI.
func (s *Service) GetSession(ctx context.Context, jti string) (*types.SessionInfo, error) {
	return s.repo.GetSession(ctx, jti)
}

// GetUserSessions retrieves all active sessions for a user.
func (s *Service) GetUserSessions(ctx context.Context, userID string, limit, offset int) (*types.SessionListResponse, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	sessions, total, err := s.repo.ListUserSessions(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}

	// Get username from first session if available
	username := ""
	if len(sessions) > 0 {
		username = sessions[0].Username
	}

	return &types.SessionListResponse{
		UserID:   userID,
		Username: username,
		Total:    total,
		Sessions: sessions,
		Pagination: &types.Pagination{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasMore: offset+len(sessions) < total,
		},
	}, nil
}

// ValidateSession checks if a session is valid and not revoked.
func (s *Service) ValidateSession(ctx context.Context, jti string) (bool, error) {
	// Check if revoked
	revoked, err := s.repo.IsRevoked(ctx, jti)
	if err != nil {
		return false, fmt.Errorf("check revoked: %w", err)
	}
	if revoked {
		return false, nil
	}

	// Check if session exists
	_, err = s.repo.GetSession(ctx, jti)
	if err != nil {
		// Session not found means it's not active (could be expired or never existed)
		_ = err // Explicitly ignore error
		return false, nil
	}

	return true, nil
}

// TerminateSession revokes a single session.
func (s *Service) TerminateSession(ctx context.Context, jti, userID, revokedBy, reason, eventID string) error {
	// Verify session exists
	session, err := s.repo.GetSession(ctx, jti)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Verify it belongs to the specified user
	if session.UserID != userID {
		return fmt.Errorf("session belongs to different user")
	}

	return s.repo.RevokeSession(ctx, jti, userID, revokedBy, reason, eventID)
}

// TerminateUserSessions revokes all sessions for a user.
func (s *Service) TerminateUserSessions(ctx context.Context, userID, revokedBy, reason, eventID string) (int, error) {
	// Get all user sessions
	sessions, _, err := s.repo.ListUserSessions(ctx, userID, 1000, 0) // Max 1000 sessions
	if err != nil {
		return 0, fmt.Errorf("list user sessions: %w", err)
	}

	if len(sessions) == 0 {
		return 0, nil
	}

	// Extract JTIs
	jtis := make([]string, len(sessions))
	for i, session := range sessions {
		jtis[i] = session.JTI
	}

	// Bulk revoke
	if err := s.repo.BulkRevokeSessions(ctx, jtis, userID, revokedBy, reason, eventID); err != nil {
		return 0, fmt.Errorf("bulk revoke: %w", err)
	}

	return len(sessions), nil
}

// detectDeviceType determines device type from User-Agent.
func detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

// parseDeviceName extracts readable device name from User-Agent.
func parseDeviceName(userAgent string) string {
	// Simple implementation - can be enhanced with proper UA parser
	ua := userAgent

	// Extract browser
	browser := "Unknown Browser"
	if strings.Contains(ua, "Chrome/") {
		parts := strings.Split(ua, "Chrome/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			browser = "Chrome " + version
		}
	} else if strings.Contains(ua, "Firefox/") {
		parts := strings.Split(ua, "Firefox/")
		if len(parts) > 1 {
			version := strings.Split(parts[1], " ")[0]
			browser = "Firefox " + version
		}
	} else if strings.Contains(ua, "Safari/") && !strings.Contains(ua, "Chrome") {
		browser = "Safari"
	}

	// Extract OS (check specific patterns first before generic ones)
	os := "Unknown OS"
	if strings.Contains(ua, "Windows NT") {
		os = "Windows"
	} else if strings.Contains(ua, "Android") {
		os = "Android"
	} else if strings.Contains(ua, "iOS") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
		os = "iOS"
	} else if strings.Contains(ua, "Mac OS X") {
		os = "macOS"
	} else if strings.Contains(ua, "Linux") {
		os = "Linux"
	}

	return fmt.Sprintf("%s on %s", browser, os)
}
