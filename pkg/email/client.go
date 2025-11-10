package email

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"
)

// Client provides email sending functionality.
type Client interface {
	Send(ctx context.Context, msg *Message) (*Receipt, error)
}

// Config holds email configuration.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
	Timeout  time.Duration
}

// Message represents an email message.
type Message struct {
	ID       string
	Title    string
	Body     string
	Format   string // "html" or "text"
	Targets  []Target
	Metadata map[string]interface{}
}

// Target represents an email recipient.
type Target struct {
	Type     string
	Value    string // email address
	Platform string
}

// Receipt represents the result of sending an email.
type Receipt struct {
	MessageID string
	Results   []Result
}

// Result represents the result for a single recipient.
type Result struct {
	Platform string
	Success  bool
	Error    string
}

// smtpClient implements the Client interface using SMTP.
type smtpClient struct {
	config *Config
}

// NewClient creates a new email client.
func NewClient(config *Config) (Client, error) {
	if config == nil {
		return &noopClient{}, nil
	}

	if config.Host == "" {
		return &noopClient{}, nil
	}

	return &smtpClient{
		config: config,
	}, nil
}

// Send sends an email message.
func (c *smtpClient) Send(ctx context.Context, msg *Message) (*Receipt, error) {
	if len(msg.Targets) == 0 {
		return nil, fmt.Errorf("no targets specified")
	}

	receipt := &Receipt{
		MessageID: msg.ID,
		Results:   make([]Result, 0, len(msg.Targets)),
	}

	// Send to each target
	for _, target := range msg.Targets {
		result := Result{
			Platform: target.Platform,
			Success:  false,
		}

		if target.Type != "email" {
			result.Error = "unsupported target type"
			receipt.Results = append(receipt.Results, result)
			continue
		}

		// Compose email
		var body string
		if msg.Format == "html" {
			body = fmt.Sprintf("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\nFrom: %s\nTo: %s\nSubject: %s\n\n%s",
				c.config.From, target.Value, msg.Title, msg.Body)
		} else {
			body = fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
				c.config.From, target.Value, msg.Title, msg.Body)
		}

		// Send via SMTP
		addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
		var auth smtp.Auth
		if c.config.Username != "" && c.config.Password != "" {
			auth = smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
		}

		err := smtp.SendMail(addr, auth, c.config.From, []string{target.Value}, []byte(body))
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}

		receipt.Results = append(receipt.Results, result)
	}

	return receipt, nil
}

// noopClient is a no-op implementation for testing/development.
type noopClient struct{}

// Send logs the message instead of sending it.
func (c *noopClient) Send(ctx context.Context, msg *Message) (*Receipt, error) {
	log.Printf("[EMAIL NO-OP] Would send email: %s to %d recipients", msg.Title, len(msg.Targets))

	receipt := &Receipt{
		MessageID: msg.ID,
		Results:   make([]Result, 0, len(msg.Targets)),
	}

	for _, target := range msg.Targets {
		receipt.Results = append(receipt.Results, Result{
			Platform: target.Platform,
			Success:  true, // Simulate success in no-op mode
		})
	}

	return receipt, nil
}
