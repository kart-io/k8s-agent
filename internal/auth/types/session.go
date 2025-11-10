package types

import (
	"time"

	authtypes "github.com/kart-io/k8s-agent/pkg/auth/types"
)

// Re-export types from pkg/auth/types for backward compatibility
type SessionInfo = authtypes.SessionInfo
type RevokedSession = authtypes.RevokedSession

// SessionListResponse represents the response for listing user sessions.
type SessionListResponse struct {
	UserID     string        `json:"user_id"`
	Username   string        `json:"username"`
	Total      int           `json:"total"`
	Sessions   []SessionInfo `json:"sessions"`
	Pagination *Pagination   `json:"pagination,omitempty"`
}

// Pagination holds pagination metadata.
type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	Total   int  `json:"total"`
	HasMore bool `json:"has_more"`
}

// ForceLogoutRequest represents a forced logout request.
type ForceLogoutRequest struct {
	Reason        string `json:"reason,omitempty"`                                                       // Optional reason
	TriggeredBy   string `json:"triggered_by" binding:"omitempty,oneof=manual policy security_incident"` // manual, policy, security_incident
	CorrelationID string `json:"correlation_id,omitempty"`                                               // Optional correlation ID
}

// ForceLogoutResponse represents the response after forced logout.
type ForceLogoutResponse struct {
	EventID        string    `json:"event_id"`
	Success        bool      `json:"success"`
	SessionCount   int       `json:"session_count"`
	TargetUserID   string    `json:"target_user_id,omitempty"`
	TargetUsername string    `json:"target_username,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	Message        string    `json:"message"`
}

// BulkForceLogoutRequest represents a bulk logout request.
type BulkForceLogoutRequest struct {
	SessionJTIs []string `json:"session_jtis" binding:"required,min=1,max=100"`
	Reason      string   `json:"reason,omitempty"`
	TriggeredBy string   `json:"triggered_by" binding:"omitempty,oneof=manual policy security_incident"`
}

// BulkForceLogoutResponse represents bulk logout results.
type BulkForceLogoutResponse struct {
	EventID        string                `json:"event_id"`
	TotalRequested int                   `json:"total_requested"`
	Successful     int                   `json:"successful"`
	Failed         int                   `json:"failed"`
	Results        []SessionLogoutResult `json:"results"`
}

// SessionLogoutResult represents per-session logout result.
type SessionLogoutResult struct {
	JTI     string `json:"jti"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// AuditEventListResponse represents audit event list response.
type AuditEventListResponse struct {
	Total      int           `json:"total"`
	Events     []interface{} `json:"events"` // Will be ForcedLogoutEvent
	Pagination *Pagination   `json:"pagination,omitempty"`
}

// ErrorResponse represents API error response.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}
