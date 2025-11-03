package initializers

import (
	"context"

	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/k8s-agent/internal/auth/config"
	forcedlogout "github.com/kart-io/k8s-agent/internal/auth/forced-logout"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/audit"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/notification"
	"github.com/kart-io/k8s-agent/internal/auth/forced-logout/session"
	"github.com/kart-io/logger/core"
)

// SessionServiceInitializer Session 服务初始化器
type SessionServiceInitializer struct {
	cfg       *config.Config
	logger    core.Logger
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer
	service   *session.Service
}

// NewSessionServiceInitializer 创建 Session 服务初始化器
func NewSessionServiceInitializer(
	cfg *config.Config,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *SessionServiceInitializer {
	return &SessionServiceInitializer{
		cfg:       cfg,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}
}

// Name 返回初始化器名称
func (s *SessionServiceInitializer) Name() string {
	return "session-service"
}

// Priority 返回初始化优先级
func (s *SessionServiceInitializer) Priority() int {
	return bootstrap.PriorityMQ - 50 // 450
}

// Initialize 执行初始化
func (s *SessionServiceInitializer) Initialize(ctx context.Context) error {
	s.logger.Infow("Initializing Session service")

	// 创建 session repository (使用Redis存储)
	sessionRepo := session.NewRedisRepository(s.redisInit.Client(), s.cfg.JWT.ExpiresHours)
	s.service = session.NewService(sessionRepo)

	s.logger.Infow("Session service initialized successfully")
	return nil
}

// Close 关闭服务
func (s *SessionServiceInitializer) Close(ctx context.Context) error {
	s.logger.Infow("Closing Session service")
	return nil
}

// Service 获取服务实例
func (s *SessionServiceInitializer) Service() *session.Service {
	return s.service
}

// AuditServiceInitializer Audit 服务初始化器
type AuditServiceInitializer struct {
	cfg     *config.Config
	logger  core.Logger
	dbInit  *DatabaseInitializer
	service *audit.Service
}

// NewAuditServiceInitializer 创建 Audit 服务初始化器
func NewAuditServiceInitializer(
	cfg *config.Config,
	logger core.Logger,
	dbInit *DatabaseInitializer,
) *AuditServiceInitializer {
	return &AuditServiceInitializer{
		cfg:    cfg,
		logger: logger,
		dbInit: dbInit,
	}
}

// Name 返回初始化器名称
func (a *AuditServiceInitializer) Name() string {
	return "audit-service"
}

// Priority 返回初始化优先级
func (a *AuditServiceInitializer) Priority() int {
	return bootstrap.PriorityMQ - 40 // 460
}

// Initialize 执行初始化
func (a *AuditServiceInitializer) Initialize(ctx context.Context) error {
	a.logger.Infow("Initializing Audit service")

	// 创建 audit repository
	auditRepo := audit.NewPostgresRepository(a.dbInit.DB())
	a.service = audit.NewService(auditRepo)

	a.logger.Infow("Audit service initialized successfully")
	return nil
}

// Close 关闭服务
func (a *AuditServiceInitializer) Close(ctx context.Context) error {
	a.logger.Infow("Closing Audit service")
	return nil
}

// Service 获取服务实例
func (a *AuditServiceInitializer) Service() *audit.Service {
	return a.service
}

// NotificationServiceInitializer Notification 服务初始化器
type NotificationServiceInitializer struct {
	cfg       *config.Config
	logger    core.Logger
	dbInit    *DatabaseInitializer
	emailInit *EmailClientInitializer
	service   *notification.Service
}

// NewNotificationServiceInitializer 创建 Notification 服务初始化器
func NewNotificationServiceInitializer(
	cfg *config.Config,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	emailInit *EmailClientInitializer,
) *NotificationServiceInitializer {
	return &NotificationServiceInitializer{
		cfg:       cfg,
		logger:    logger,
		dbInit:    dbInit,
		emailInit: emailInit,
	}
}

// Name 返回初始化器名称
func (n *NotificationServiceInitializer) Name() string {
	return "notification-service"
}

// Priority 返回初始化优先级
func (n *NotificationServiceInitializer) Priority() int {
	return bootstrap.PriorityMQ - 30 // 470
}

// Initialize 执行初始化
func (n *NotificationServiceInitializer) Initialize(ctx context.Context) error {
	n.logger.Infow("Initializing Notification service")

	// 创建 notification repository
	notificationRepo := notification.NewPostgresRepository(n.dbInit.DB())

	// 创建 template engine
	templateEngine, err := notification.NewTemplateEngine(n.cfg.Email.TemplateDir)
	if err != nil {
		return err
	}

	// 创建 notification service
	n.service = notification.NewService(
		notificationRepo,
		n.emailInit.Client(),
		templateEngine,
	)

	n.logger.Infow("Notification service initialized successfully")
	return nil
}

// Close 关闭服务
func (n *NotificationServiceInitializer) Close(ctx context.Context) error {
	n.logger.Infow("Closing Notification service")
	return nil
}

// Service 获取服务实例
func (n *NotificationServiceInitializer) Service() *notification.Service {
	return n.service
}

// ForcedLogoutServiceInitializer ForcedLogout 服务初始化器
type ForcedLogoutServiceInitializer struct {
	cfg              *config.Config
	logger           core.Logger
	sessionInit      *SessionServiceInitializer
	auditInit        *AuditServiceInitializer
	notificationInit *NotificationServiceInitializer
	service          *forcedlogout.Service
}

// NewForcedLogoutServiceInitializer 创建 ForcedLogout 服务初始化器
func NewForcedLogoutServiceInitializer(
	cfg *config.Config,
	logger core.Logger,
	sessionInit *SessionServiceInitializer,
	auditInit *AuditServiceInitializer,
	notificationInit *NotificationServiceInitializer,
) *ForcedLogoutServiceInitializer {
	return &ForcedLogoutServiceInitializer{
		cfg:              cfg,
		logger:           logger,
		sessionInit:      sessionInit,
		auditInit:        auditInit,
		notificationInit: notificationInit,
	}
}

// Name 返回初始化器名称
func (f *ForcedLogoutServiceInitializer) Name() string {
	return "forced-logout-service"
}

// Priority 返回初始化优先级
func (f *ForcedLogoutServiceInitializer) Priority() int {
	return bootstrap.PriorityMQ - 10 // 490 (在notification之后)
}

// Initialize 执行初始化
func (f *ForcedLogoutServiceInitializer) Initialize(ctx context.Context) error {
	f.logger.Infow("Initializing ForcedLogout service")

	// Determine login URL (use environment variable or default)
	loginURL := "http://localhost:8090/login"

	// 创建 forced logout service
	f.service = forcedlogout.NewService(
		f.sessionInit.Service(),
		f.auditInit.Service(),
		f.notificationInit.Service(),
		loginURL,
	)

	f.logger.Infow("ForcedLogout service initialized successfully")
	return nil
}

// Close 关闭服务
func (f *ForcedLogoutServiceInitializer) Close(ctx context.Context) error {
	f.logger.Infow("Closing ForcedLogout service")
	return nil
}

// Service 获取服务实例
func (f *ForcedLogoutServiceInitializer) Service() *forcedlogout.Service {
	return f.service
}
