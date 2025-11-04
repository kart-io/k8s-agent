package session

import (
	"context"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// Repository defines the interface for session storage operations.
type Repository interface {
	// StoreSession stores a new session in Redis
	StoreSession(ctx context.Context, session *types.SessionInfo) error

	// GetSession retrieves session metadata by JTI
	GetSession(ctx context.Context, jti string) (*types.SessionInfo, error)

	// ListUserSessions retrieves all active sessions for a user with pagination
	ListUserSessions(ctx context.Context, userID string, limit, offset int) ([]types.SessionInfo, int, error)

	// RevokeSession adds session to blacklist and removes from active sets
	RevokeSession(ctx context.Context, jti, userID, revokedBy, reason, eventID string) error

	// IsRevoked checks if a session JTI is blacklisted
	IsRevoked(ctx context.Context, jti string) (bool, error)

	// BulkRevokeSession revokes multiple sessions in one operation
	BulkRevokeSessions(ctx context.Context, jtis []string, userID, revokedBy, reason, eventID string) error
}
