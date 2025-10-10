package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kart-io/notifyhub/pkg/message"
	"github.com/kart-io/notifyhub/pkg/notifyhub"
	"github.com/kart-io/notifyhub/pkg/target"
	"github.com/kart/k8s-agent/auth-service/pkg/types"
)

// Service orchestrates notification delivery and tracking using NotifyHub
type Service struct {
	repo            Repository
	notifyHubClient notifyhub.Client
	templateEngine  *TemplateEngine
}

// NewService creates a new notification service with NotifyHub integration
func NewService(repo Repository, notifyHubClient notifyhub.Client, templateEngine *TemplateEngine) *Service {
	return &Service{
		repo:            repo,
		notifyHubClient: notifyHubClient,
		templateEngine:  templateEngine,
	}
}

// NotifyUserParams contains parameters for sending a notification
type NotifyUserParams struct {
	EventID      string
	UserID       string
	EmailAddress string
	Username     string
	Timestamp    time.Time
	Reason       string
	DeviceInfo   string
	Location     string
	ActorName    string
	LoginURL     string
}

// NotifyUser creates a notification record and sends an email using NotifyHub
func (s *Service) NotifyUser(ctx context.Context, params NotifyUserParams) error {
	// Validate parameters
	if err := s.validateParams(params); err != nil {
		return fmt.Errorf("validate params: %w", err)
	}

	// Create notification variables
	variables := CreateNotificationVariables(
		params.Username,
		params.Timestamp,
		params.Reason,
		params.DeviceInfo,
		params.Location,
		params.ActorName,
		params.LoginURL,
	)

	// Render email template
	renderedEmail, err := s.templateEngine.RenderTemplate(variables)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// Create notification record
	notificationID := uuid.New().String()
	notification := &types.ForcedLogoutNotification{
		NotificationID: notificationID,
		EventID:        params.EventID,
		UserID:         params.UserID,
		EmailAddress:   params.EmailAddress,
		Channel:        "email",
		TemplateName:   "forced-logout",
		Subject:        &renderedEmail.Subject,
		Body:           &renderedEmail.TextBody,
		Variables:      &variables,
		Status:         "pending",
		Attempts:       0,
	}

	// Store notification record
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("create notification record: %w", err)
	}

	// Send email asynchronously using NotifyHub
	s.sendEmailAsync(ctx, notificationID, params.EmailAddress, renderedEmail)

	return nil
}

// sendEmailAsync sends email in background using NotifyHub
func (s *Service) sendEmailAsync(ctx context.Context, notificationID, emailAddress string, renderedEmail *RenderedEmail) {
	go func() {
		// Use background context with timeout
		bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// Increment attempts
		if err := s.repo.IncrementAttempts(bgCtx, notificationID); err != nil {
			fmt.Printf("Failed to increment attempts for notification %s: %v\n", notificationID, err)
		}

		// Create message for NotifyHub
		msg := &message.Message{
			ID:     notificationID,
			Title:  renderedEmail.Subject,
			Body:   renderedEmail.HTMLBody,
			Format: message.FormatHTML,
			Targets: []target.Target{
				{
					Type:     "email",
					Value:    emailAddress,
					Platform: "email",
				},
			},
			Metadata: map[string]interface{}{
				"text_body": renderedEmail.TextBody,
				"category":  "security-alert",
			},
		}

		// Send via NotifyHub
		receipt, err := s.notifyHubClient.Send(bgCtx, msg)

		// Update status based on result
		if err != nil {
			errMsg := err.Error()
			if updateErr := s.repo.UpdateStatus(bgCtx, notificationID, "failed", nil, &errMsg); updateErr != nil {
				fmt.Printf("Failed to update notification status to failed for %s: %v\n", notificationID, updateErr)
			}
		} else {
			// Check if email delivery succeeded
			success := false
			for _, result := range receipt.Results {
				if result.Platform == "email" && result.Success {
					success = true
					break
				}
			}

			if success {
				sentAt := time.Now()
				if updateErr := s.repo.UpdateStatus(bgCtx, notificationID, "sent", &sentAt, nil); updateErr != nil {
					fmt.Printf("Failed to update notification status to sent for %s: %v\n", notificationID, updateErr)
				}
			} else {
				errMsg := "Email delivery failed"
				if len(receipt.Results) > 0 && receipt.Results[0].Error != "" {
					errMsg = receipt.Results[0].Error
				}
				if updateErr := s.repo.UpdateStatus(bgCtx, notificationID, "failed", nil, &errMsg); updateErr != nil {
					fmt.Printf("Failed to update notification status to failed for %s: %v\n", notificationID, updateErr)
				}
			}
		}
	}()
}

// NotifyUserSync sends notification synchronously (for testing or critical notifications)
func (s *Service) NotifyUserSync(ctx context.Context, params NotifyUserParams) error {
	// Validate parameters
	if err := s.validateParams(params); err != nil {
		return fmt.Errorf("validate params: %w", err)
	}

	// Create notification variables
	variables := CreateNotificationVariables(
		params.Username,
		params.Timestamp,
		params.Reason,
		params.DeviceInfo,
		params.Location,
		params.ActorName,
		params.LoginURL,
	)

	// Render email template
	renderedEmail, err := s.templateEngine.RenderTemplate(variables)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	// Create notification record
	notificationID := uuid.New().String()
	notification := &types.ForcedLogoutNotification{
		NotificationID: notificationID,
		EventID:        params.EventID,
		UserID:         params.UserID,
		EmailAddress:   params.EmailAddress,
		Channel:        "email",
		TemplateName:   "forced-logout",
		Subject:        &renderedEmail.Subject,
		Body:           &renderedEmail.TextBody,
		Variables:      &variables,
		Status:         "pending",
		Attempts:       0,
	}

	// Store notification record
	if err := s.repo.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("create notification record: %w", err)
	}

	// Increment attempts
	if err := s.repo.IncrementAttempts(ctx, notificationID); err != nil {
		return fmt.Errorf("increment attempts: %w", err)
	}

	// Create message for NotifyHub
	msg := &message.Message{
		ID:     notificationID,
		Title:  renderedEmail.Subject,
		Body:   renderedEmail.HTMLBody,
		Format: message.FormatHTML,
		Targets: []target.Target{
			{
				Type:     "email",
				Value:    params.EmailAddress,
				Platform: "email",
			},
		},
		Metadata: map[string]interface{}{
			"text_body": renderedEmail.TextBody,
			"category":  "security-alert",
		},
	}

	// Send via NotifyHub synchronously
	receipt, err := s.notifyHubClient.Send(ctx, msg)

	// Update status
	if err != nil {
		errMsg := err.Error()
		if updateErr := s.repo.UpdateStatus(ctx, notificationID, "failed", nil, &errMsg); updateErr != nil {
			fmt.Printf("Failed to update notification status to failed for %s: %v\n", notificationID, updateErr)
		}
		return fmt.Errorf("send email: %w", err)
	}

	// Check if email delivery succeeded
	success := false
	for _, result := range receipt.Results {
		if result.Platform == "email" && result.Success {
			success = true
			break
		}
	}

	if success {
		sentAt := time.Now()
		if updateErr := s.repo.UpdateStatus(ctx, notificationID, "sent", &sentAt, nil); updateErr != nil {
			fmt.Printf("Failed to update notification status to sent for %s: %v\n", notificationID, updateErr)
		}
	} else {
		errMsg := "Email delivery failed"
		if len(receipt.Results) > 0 && receipt.Results[0].Error != "" {
			errMsg = receipt.Results[0].Error
		}
		if updateErr := s.repo.UpdateStatus(ctx, notificationID, "failed", nil, &errMsg); updateErr != nil {
			fmt.Printf("Failed to update notification status to failed for %s: %v\n", notificationID, updateErr)
		}
		return fmt.Errorf("email delivery failed: %s", errMsg)
	}

	return nil
}

// NotifyMultipleUsers sends notifications to multiple users concurrently
func (s *Service) NotifyMultipleUsers(ctx context.Context, paramsSlice []NotifyUserParams) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(paramsSlice))

	for i, params := range paramsSlice {
		wg.Add(1)
		go func(index int, p NotifyUserParams) {
			defer wg.Done()
			if err := s.NotifyUser(ctx, p); err != nil {
				errors[index] = err
			}
		}(i, params)
	}

	wg.Wait()

	// Filter out nil errors
	var actualErrors []error
	for _, err := range errors {
		if err != nil {
			actualErrors = append(actualErrors, err)
		}
	}

	return actualErrors
}

// RetryFailedNotifications processes pending/failed notifications
// This should be called by a background worker periodically
func (s *Service) RetryFailedNotifications(ctx context.Context, maxAttempts int, batchSize int) (int, error) {
	// Get pending notifications
	notifications, err := s.repo.GetPendingNotifications(ctx, maxAttempts, batchSize)
	if err != nil {
		return 0, fmt.Errorf("get pending notifications: %w", err)
	}

	if len(notifications) == 0 {
		return 0, nil
	}

	successCount := 0

	// Process each notification
	for _, notification := range notifications {
		// Skip if max attempts reached
		if notification.Attempts >= maxAttempts {
			continue
		}

		// Skip if no variables
		if notification.Variables == nil {
			continue
		}

		// Render template with stored variables
		renderedEmail, err := s.templateEngine.RenderTemplate(*notification.Variables)
		if err != nil {
			continue
		}

		// Increment attempts
		if err := s.repo.IncrementAttempts(ctx, notification.NotificationID); err != nil {
			continue
		}

		// Create message for NotifyHub
		msg := &message.Message{
			ID:     notification.NotificationID,
			Title:  renderedEmail.Subject,
			Body:   renderedEmail.HTMLBody,
			Format: message.FormatHTML,
			Targets: []target.Target{
				{
					Type:     "email",
					Value:    notification.EmailAddress,
					Platform: "email",
				},
			},
			Metadata: map[string]interface{}{
				"text_body": renderedEmail.TextBody,
				"category":  "security-alert",
				"retry":     true,
			},
		}

		// Send via NotifyHub
		receipt, err := s.notifyHubClient.Send(ctx, msg)

		// Update status
		if err != nil {
			errMsg := err.Error()
			if updateErr := s.repo.UpdateStatus(ctx, notification.NotificationID, "failed", nil, &errMsg); updateErr != nil {
				fmt.Printf("Failed to update notification status to failed for %s: %v\n", notification.NotificationID, updateErr)
			}
		} else {
			// Check if email delivery succeeded
			success := false
			for _, result := range receipt.Results {
				if result.Platform == "email" && result.Success {
					success = true
					break
				}
			}

			if success {
				sentAt := time.Now()
				if updateErr := s.repo.UpdateStatus(ctx, notification.NotificationID, "sent", &sentAt, nil); updateErr != nil {
					fmt.Printf("Failed to update notification status to sent for %s: %v\n", notification.NotificationID, updateErr)
				}
				successCount++
			} else {
				errMsg := "Email delivery failed"
				if len(receipt.Results) > 0 && receipt.Results[0].Error != "" {
					errMsg = receipt.Results[0].Error
				}
				if updateErr := s.repo.UpdateStatus(ctx, notification.NotificationID, "failed", nil, &errMsg); updateErr != nil {
					fmt.Printf("Failed to update notification status to failed for %s: %v\n", notification.NotificationID, updateErr)
				}
			}
		}
	}

	return successCount, nil
}

// GetNotificationStatus retrieves the delivery status of a notification
func (s *Service) GetNotificationStatus(ctx context.Context, notificationID string) (*types.ForcedLogoutNotification, error) {
	return s.repo.GetNotification(ctx, notificationID)
}

// validateParams validates notification parameters
func (s *Service) validateParams(params NotifyUserParams) error {
	if params.EventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if params.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if params.EmailAddress == "" {
		return fmt.Errorf("email_address is required")
	}
	if params.Username == "" {
		return fmt.Errorf("username is required")
	}
	if params.LoginURL == "" {
		return fmt.Errorf("login_url is required")
	}
	return nil
}
