package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/kart-io/k8s-agent/internal/auth/types"
	"gorm.io/gorm"
)

// PostgresRepository implements Repository using PostgreSQL via GORM
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL-based notification repository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CreateNotification inserts a new notification record
func (r *PostgresRepository) CreateNotification(ctx context.Context, notification *types.ForcedLogoutNotification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

// UpdateStatus updates the delivery status of a notification
func (r *PostgresRepository) UpdateStatus(ctx context.Context, notificationID string, status string, sentAt *time.Time, errorMsg *string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "sent" && sentAt != nil {
		updates["sent_at"] = sentAt
	}

	if status == "failed" {
		now := time.Now()
		updates["failed_at"] = &now
		if errorMsg != nil {
			updates["error_message"] = errorMsg
		}
	}

	if err := r.db.WithContext(ctx).
		Model(&types.ForcedLogoutNotification{}).
		Where("notification_id = ?", notificationID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update notification status: %w", err)
	}

	return nil
}

// GetPendingNotifications retrieves notifications that need to be sent/retried
func (r *PostgresRepository) GetPendingNotifications(ctx context.Context, maxAttempts int, limit int) ([]types.ForcedLogoutNotification, error) {
	var notifications []types.ForcedLogoutNotification

	// Find notifications that are:
	// 1. Status is 'pending' OR
	// 2. Status is 'failed' and attempts < maxAttempts
	if err := r.db.WithContext(ctx).
		Where("status = ? OR (status = ? AND attempts < ?)", "pending", "failed", maxAttempts).
		Order("created_at ASC").
		Limit(limit).
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("get pending notifications: %w", err)
	}

	return notifications, nil
}

// IncrementAttempts increments the attempt counter and updates last_attempt_at
func (r *PostgresRepository) IncrementAttempts(ctx context.Context, notificationID string) error {
	now := time.Now()

	if err := r.db.WithContext(ctx).
		Model(&types.ForcedLogoutNotification{}).
		Where("notification_id = ?", notificationID).
		Updates(map[string]interface{}{
			"attempts":        gorm.Expr("attempts + 1"),
			"last_attempt_at": &now,
			"updated_at":      now,
		}).Error; err != nil {
		return fmt.Errorf("increment attempts: %w", err)
	}

	return nil
}

// GetNotification retrieves a single notification by ID
func (r *PostgresRepository) GetNotification(ctx context.Context, notificationID string) (*types.ForcedLogoutNotification, error) {
	var notification types.ForcedLogoutNotification

	if err := r.db.WithContext(ctx).
		Where("notification_id = ?", notificationID).
		First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found: %s", notificationID)
		}
		return nil, fmt.Errorf("get notification: %w", err)
	}

	return &notification, nil
}

// GetNotificationsByEventID retrieves all notifications for a specific event
func (r *PostgresRepository) GetNotificationsByEventID(ctx context.Context, eventID string) ([]types.ForcedLogoutNotification, error) {
	var notifications []types.ForcedLogoutNotification

	if err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("get notifications by event ID: %w", err)
	}

	return notifications, nil
}
