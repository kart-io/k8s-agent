package notification

import (
	"context"
	"time"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// Repository defines the interface for notification storage operations
type Repository interface {
	// CreateNotification inserts a new notification record
	CreateNotification(ctx context.Context, notification *types.ForcedLogoutNotification) error

	// UpdateStatus updates the delivery status of a notification
	UpdateStatus(ctx context.Context, notificationID string, status string, sentAt *time.Time, errorMsg *string) error

	// GetPendingNotifications retrieves notifications that need to be sent/retried
	// Returns notifications with status 'pending' or 'failed' with attempts < max_attempts
	GetPendingNotifications(ctx context.Context, maxAttempts int, limit int) ([]types.ForcedLogoutNotification, error)

	// IncrementAttempts increments the attempt counter and updates last_attempt_at
	IncrementAttempts(ctx context.Context, notificationID string) error

	// GetNotification retrieves a single notification by ID
	GetNotification(ctx context.Context, notificationID string) (*types.ForcedLogoutNotification, error)

	// GetNotificationsByEventID retrieves all notifications for a specific event
	GetNotificationsByEventID(ctx context.Context, eventID string) ([]types.ForcedLogoutNotification, error)
}
