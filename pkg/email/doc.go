// Package email provides generic email sending functionality.
//
// This package is independent of business logic and can be used by any service
// that needs to send email notifications. It supports SMTP-based email delivery
// with optional TLS encryption.
//
// Key features:
//   - Simple interface for sending emails
//   - Support for both HTML and plain text formats
//   - Multiple recipients per message
//   - No-op client for testing/development
//   - Delivery receipt tracking
//
// Example usage:
//
//	config := &email.Config{
//	    Host:     "smtp.gmail.com",
//	    Port:     587,
//	    Username: "user@example.com",
//	    Password: "password",
//	    From:     "noreply@example.com",
//	    UseTLS:   true,
//	}
//
//	client, err := email.NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	msg := &email.Message{
//	    ID:     "msg-123",
//	    Title:  "Test Email",
//	    Body:   "<h1>Hello World</h1>",
//	    Format: "html",
//	    Targets: []email.Target{
//	        {Type: "email", Value: "recipient@example.com", Platform: "email"},
//	    },
//	}
//
//	receipt, err := client.Send(context.Background(), msg)
//	if err != nil {
//	    log.Fatal(err)
//	}
package email
