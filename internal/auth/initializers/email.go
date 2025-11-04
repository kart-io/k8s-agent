package initializers

import (
	"context"
	"time"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	"github.com/kart-io/k8s-agent/internal/auth/email"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// EmailClientInitializer Email 客户端初始化器.
type EmailClientInitializer struct {
	cfg    *options.ServerOptions
	logger core.Logger
	client email.Client
}

// NewEmailClientInitializer 创建 Email 客户端初始化器.
func NewEmailClientInitializer(
	cfg *options.ServerOptions,
	logger core.Logger,
) *EmailClientInitializer {
	return &EmailClientInitializer{
		cfg:    cfg,
		logger: logger,
	}
}

// Name 返回初始化器名称.
func (e *EmailClientInitializer) Name() string {
	return "email-client"
}

// Priority 返回初始化优先级.
func (e *EmailClientInitializer) Priority() int {
	return bootstrap.PriorityMQ - 50 // 450
}

// Initialize 执行初始化.
func (e *EmailClientInitializer) Initialize(ctx context.Context) error {
	e.logger.Infow("Initializing Email client",
		"smtp_host", e.cfg.Email.SMTPHost,
		"smtp_port", e.cfg.Email.SMTPPort,
		"enabled", e.cfg.Email.Enabled,
	)

	var err error
	if e.cfg.Email.Enabled {
		// Create email client with SMTP config
		emailConfig := &email.Config{
			Host:     e.cfg.Email.SMTPHost,
			Port:     e.cfg.Email.SMTPPort,
			Username: e.cfg.Email.SMTPUser,
			Password: e.cfg.Email.SMTPPassword,
			From:     e.cfg.Email.FromAddress,
			UseTLS:   true,
			Timeout:  30 * time.Second,
		}

		e.client, err = email.NewClient(emailConfig)
		if err != nil {
			return err
		}
		e.logger.Infow("Email client initialized with SMTP",
			"host", e.cfg.Email.SMTPHost,
			"port", e.cfg.Email.SMTPPort,
		)
	} else {
		// Create a no-op client
		e.client, _ = email.NewClient(nil)
		e.logger.Infow("Email notifications disabled - using no-op mode")
	}

	return nil
}

// Close 关闭客户端.
func (e *EmailClientInitializer) Close(ctx context.Context) error {
	e.logger.Infow("Closing Email client")
	return nil
}

// Client 获取客户端实例.
func (e *EmailClientInitializer) Client() email.Client {
	return e.client
}
