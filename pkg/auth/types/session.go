package types

import (
	"time"
)

// SessionInfo represents an active user session.
type SessionInfo struct {
	JTI            string    `json:"jti"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"` // User email for notifications
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	DeviceType     string    `json:"device_type"` // desktop, mobile, tablet
	DeviceName     string    `json:"device_name"` // "Chrome 118 on Windows"
	Location       string    `json:"location"`    // "City, State, Country"
	LoginAt        time.Time `json:"login_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// RevokedSession represents a blacklisted session.
type RevokedSession struct {
	JTI       string    `json:"jti"`
	UserID    string    `json:"user_id"`
	RevokedAt time.Time `json:"revoked_at"`
	RevokedBy string    `json:"revoked_by"` // admin user ID or "system"
	Reason    string    `json:"reason"`
	EventID   string    `json:"event_id"` // Links to audit event
}
