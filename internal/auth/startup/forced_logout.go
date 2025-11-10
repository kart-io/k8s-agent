// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package startup

import (
	"context"
	"fmt"
	"time"

	forcedlogout "github.com/kart-io/k8s-agent/internal/auth/forced-logout"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/audit"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/notification"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/email"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// ForcedLogoutServices contains all forced-logout related services.
type ForcedLogoutServices struct {
	Session      *session.Service
	Audit        *audit.Service
	Notification *notification.Service
	ForcedLogout *forcedlogout.Service
}

// SessionServiceInitializer creates session service.
type SessionServiceInitializer struct {
	opts      *commonapp.StandardOptions
	logger    core.Logger
	redisInit *pkginitializers.RedisInitializer

	service *session.Service
}

// NewSessionServiceInitializer creates a new session service initializer.
func NewSessionServiceInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	redisInit *pkginitializers.RedisInitializer,
) *SessionServiceInitializer {
	return &SessionServiceInitializer{
		opts:      opts,
		logger:    logger,
		redisInit: redisInit,
	}
}

// Name returns the initializer name.
func (s *SessionServiceInitializer) Name() string {
	return "session-service"
}

// Priority returns initialization priority.
func (s *SessionServiceInitializer) Priority() int {
	return 650 // After core services
}

// Initialize creates the session service.
func (s *SessionServiceInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing Session service")

	sessionRepo := session.NewRedisRepository(s.redisInit.Client(), s.opts.JWT.ExpiresHours)
	s.service = session.NewService(sessionRepo)

	s.logger.Infow("Session service initialized successfully")
	return nil
}

// Service returns the initialized session service.
func (s *SessionServiceInitializer) Service() *session.Service {
	return s.service
}

// Close cleans up resources.
func (s *SessionServiceInitializer) Close(ctx context.Context) error {
	return nil
}

// AuditServiceInitializer creates audit service.
type AuditServiceInitializer struct {
	logger core.Logger
	dbInit *pkginitializers.DatabaseInitializer

	service *audit.Service
}

// NewAuditServiceInitializer creates a new audit service initializer.
func NewAuditServiceInitializer(
	logger core.Logger,
	dbInit *pkginitializers.DatabaseInitializer,
) *AuditServiceInitializer {
	return &AuditServiceInitializer{
		logger: logger,
		dbInit: dbInit,
	}
}

// Name returns the initializer name.
func (a *AuditServiceInitializer) Name() string {
	return "audit-service"
}

// Priority returns initialization priority.
func (a *AuditServiceInitializer) Priority() int {
	return 660 // After session service
}

// Initialize creates the audit service.
func (a *AuditServiceInitializer) Initialize(ctx context.Context) error {
	a.logger.Infow("Initializing Audit service")

	auditRepo := audit.NewMySQLRepository(a.dbInit.DB())
	a.service = audit.NewService(auditRepo)

	a.logger.Infow("Audit service initialized successfully")
	return nil
}

// Service returns the initialized audit service.
func (a *AuditServiceInitializer) Service() *audit.Service {
	return a.service
}

// Close cleans up resources.
func (a *AuditServiceInitializer) Close(ctx context.Context) error {
	return nil
}

// NotificationServiceInitializer creates notification service.
type NotificationServiceInitializer struct {
	opts      *commonapp.StandardOptions
	logger    core.Logger
	dbInit    *pkginitializers.DatabaseInitializer
	emailInit *EmailClientInitializer

	service *notification.Service
}

// NewNotificationServiceInitializer creates a new notification service initializer.
func NewNotificationServiceInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *pkginitializers.DatabaseInitializer,
	emailInit *EmailClientInitializer,
) *NotificationServiceInitializer {
	return &NotificationServiceInitializer{
		opts:      opts,
		logger:    logger,
		dbInit:    dbInit,
		emailInit: emailInit,
	}
}

// Name returns the initializer name.
func (n *NotificationServiceInitializer) Name() string {
	return "notification-service"
}

// Priority returns initialization priority.
func (n *NotificationServiceInitializer) Priority() int {
	return 670 // After audit service
}

// Initialize creates the notification service.
func (n *NotificationServiceInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Initializing Notification service")

	notificationRepo := notification.NewMySQLRepository(n.dbInit.DB())

	templateEngine, err := notification.NewTemplateEngine(n.opts.Email.TemplateDir)
	if err != nil {
		return fmt.Errorf("failed to create template engine: %w", err)
	}

	// Create email client based on configuration
	var emailClient email.Client
	if n.opts.Email.Enabled {
		emailConfig := &email.Config{
			Host:     n.opts.Email.SMTPHost,
			Port:     n.opts.Email.SMTPPort,
			Username: n.opts.Email.SMTPUser,
			Password: n.opts.Email.SMTPPassword,
			From:     n.opts.Email.FromAddress,
			UseTLS:   true,
			Timeout:  30 * time.Second,
		}

		emailClient, err = email.NewClient(emailConfig)
		if err != nil {
			return fmt.Errorf("failed to create email client: %w", err)
		}
		n.logger.Infow("Email client initialized with SMTP",
			"host", n.opts.Email.SMTPHost,
			"port", n.opts.Email.SMTPPort,
		)
	} else {
		// Create a no-op client
		emailClient, _ = email.NewClient(nil)
		n.logger.Infow("Email notifications disabled - using no-op mode")
	}

	n.service = notification.NewService(
		notificationRepo,
		emailClient,
		templateEngine,
	)

	n.logger.Infow("Notification service initialized successfully")
	return nil
}

// Service returns the initialized notification service.
func (n *NotificationServiceInitializer) Service() *notification.Service {
	return n.service
}

// Close cleans up resources.
func (n *NotificationServiceInitializer) Close(ctx context.Context) error {
	return nil
}

// ForcedLogoutServiceInitializer creates forced logout service.
type ForcedLogoutServiceInitializer struct {
	logger           core.Logger
	sessionInit      *SessionServiceInitializer
	auditInit        *AuditServiceInitializer
	notificationInit *NotificationServiceInitializer

	service *forcedlogout.Service
}

// NewForcedLogoutServiceInitializer creates a new forced logout service initializer.
func NewForcedLogoutServiceInitializer(
	logger core.Logger,
	sessionInit *SessionServiceInitializer,
	auditInit *AuditServiceInitializer,
	notificationInit *NotificationServiceInitializer,
) *ForcedLogoutServiceInitializer {
	return &ForcedLogoutServiceInitializer{
		logger:           logger,
		sessionInit:      sessionInit,
		auditInit:        auditInit,
		notificationInit: notificationInit,
	}
}

// Name returns the initializer name.
func (f *ForcedLogoutServiceInitializer) Name() string {
	return "forced-logout-service"
}

// Priority returns initialization priority.
func (f *ForcedLogoutServiceInitializer) Priority() int {
	return 680 // After notification service
}

// Initialize creates the forced logout service.
func (f *ForcedLogoutServiceInitializer) Initialize(ctx context.Context) error {
	f.logger.Infow("Initializing ForcedLogout service")

	// Determine login URL (use environment variable or default)
	loginURL := "http://localhost:8090/login"

	f.service = forcedlogout.NewService(
		f.sessionInit.Service(),
		f.auditInit.Service(),
		f.notificationInit.Service(),
		loginURL,
	)

	f.logger.Infow("ForcedLogout service initialized successfully")
	return nil
}

// Service returns the initialized forced logout service.
func (f *ForcedLogoutServiceInitializer) Service() *forcedlogout.Service {
	return f.service
}

// Close cleans up resources.
func (f *ForcedLogoutServiceInitializer) Close(ctx context.Context) error {
	return nil
}
