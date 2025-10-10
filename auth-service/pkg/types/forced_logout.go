package types

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ForcedLogoutEvent represents an audit event
type ForcedLogoutEvent struct {
	ID              int64           `json:"id" gorm:"primaryKey"`
	EventID         string          `json:"event_id" gorm:"uniqueIndex;not null"`
	Timestamp       time.Time       `json:"timestamp" gorm:"not null"`
	ActorType       string          `json:"actor_type" gorm:"not null"` // admin, system
	ActorID         *string         `json:"actor_id,omitempty"`
	ActorUsername   *string         `json:"actor_username,omitempty"`
	ActorIPAddress  *string         `json:"actor_ip_address,omitempty"`
	TargetUserID    string          `json:"target_user_id" gorm:"not null"`
	TargetUsername  string          `json:"target_username" gorm:"not null"`
	SessionJTI      *string         `json:"session_jti,omitempty"`
	SessionCount    int             `json:"session_count" gorm:"not null"`
	SessionMetadata SessionMetadata `json:"session_metadata" gorm:"type:jsonb"`
	Reason          *string         `json:"reason,omitempty"`
	LogoutType      string          `json:"logout_type" gorm:"not null"`  // single, all, bulk
	TriggeredBy     string          `json:"triggered_by" gorm:"not null"` // manual, policy, security_incident
	PreviousHash    *string         `json:"previous_hash,omitempty"`
	CurrentHash     string          `json:"current_hash" gorm:"not null"`
	CorrelationID   *string         `json:"correlation_id,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// TableName specifies the table name
func (ForcedLogoutEvent) TableName() string {
	return "forced_logout_events"
}

// SessionMetadata is a JSON array of session details
type SessionMetadata []SessionDetail

// SessionDetail represents a single session in the metadata
type SessionDetail struct {
	JTI        string    `json:"jti"`
	IPAddress  string    `json:"ip_address"`
	DeviceName string    `json:"device_name"`
	LoginAt    time.Time `json:"login_at"`
}

// Scan implements sql.Scanner for JSONB
func (sm *SessionMetadata) Scan(value interface{}) error {
	if value == nil {
		*sm = SessionMetadata{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sm)
}

// Value implements driver.Valuer for JSONB
func (sm SessionMetadata) Value() (driver.Value, error) {
	if len(sm) == 0 {
		return nil, nil
	}
	return json.Marshal(sm)
}

// ForcedLogoutNotification represents a notification delivery record
type ForcedLogoutNotification struct {
	ID             int64                  `json:"id" gorm:"primaryKey"`
	NotificationID string                 `json:"notification_id" gorm:"uniqueIndex;not null"`
	EventID        string                 `json:"event_id" gorm:"index;not null"`
	UserID         string                 `json:"user_id" gorm:"index;not null"`
	EmailAddress   string                 `json:"email_address" gorm:"not null"`
	Channel        string                 `json:"channel" gorm:"not null"` // email, sms, push
	TemplateName   string                 `json:"template_name" gorm:"not null"`
	Subject        *string                `json:"subject,omitempty"`
	Body           *string                `json:"body,omitempty"`
	Variables      *NotificationVariables `json:"variables,omitempty" gorm:"type:jsonb"`
	Status         string                 `json:"status" gorm:"default:pending"` // pending, sent, failed
	Attempts       int                    `json:"attempts" gorm:"default:0"`
	LastAttemptAt  *time.Time             `json:"last_attempt_at,omitempty"`
	SentAt         *time.Time             `json:"sent_at,omitempty"`
	FailedAt       *time.Time             `json:"failed_at,omitempty"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TableName specifies the table name
func (ForcedLogoutNotification) TableName() string {
	return "forced_logout_notifications"
}

// NotificationVariables holds template variables
type NotificationVariables struct {
	Username   string    `json:"username"`
	Timestamp  time.Time `json:"timestamp"`
	Reason     string    `json:"reason"`
	DeviceInfo string    `json:"device_info"`
	Location   string    `json:"location"`
	ActorName  string    `json:"actor_name"`
	LoginURL   string    `json:"login_url"`
}

// Scan implements sql.Scanner for JSONB
func (nv *NotificationVariables) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, nv)
}

// Value implements driver.Valuer for JSONB
func (nv NotificationVariables) Value() (driver.Value, error) {
	return json.Marshal(nv)
}
