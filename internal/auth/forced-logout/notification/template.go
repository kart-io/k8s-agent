package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	texttemplate "text/template"
	"time"

	"github.com/kart-io/k8s-agent/internal/auth/types"
)

// TemplateEngine handles email template rendering
type TemplateEngine struct {
	templateDir string
	htmlTmpl    *template.Template
	textTmpl    *texttemplate.Template
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(templateDir string) (*TemplateEngine, error) {
	engine := &TemplateEngine{
		templateDir: templateDir,
	}

	// Load HTML template
	htmlPath := filepath.Join(templateDir, "forced-logout.html")
	htmlTmpl, err := template.ParseFiles(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("parse HTML template: %w", err)
	}
	engine.htmlTmpl = htmlTmpl

	// Load text template
	textPath := filepath.Join(templateDir, "forced-logout.txt")
	textTmpl, err := texttemplate.ParseFiles(textPath)
	if err != nil {
		return nil, fmt.Errorf("parse text template: %w", err)
	}
	engine.textTmpl = textTmpl

	return engine, nil
}

// TemplateData represents the data passed to email templates
type TemplateData struct {
	Username   string
	Timestamp  string
	Reason     string
	DeviceInfo string
	Location   string
	ActorName  string
	LoginURL   string
}

// RenderTemplate renders both HTML and text versions of the forced logout email
func (te *TemplateEngine) RenderTemplate(variables types.NotificationVariables) (*RenderedEmail, error) {
	// Validate required variables
	if err := te.validateVariables(variables); err != nil {
		return nil, fmt.Errorf("validate template variables: %w", err)
	}

	// Prepare template data
	data := TemplateData{
		Username:   variables.Username,
		Timestamp:  variables.Timestamp.Format("Monday, January 2, 2006 at 3:04 PM MST"),
		Reason:     variables.Reason,
		DeviceInfo: variables.DeviceInfo,
		Location:   variables.Location,
		ActorName:  variables.ActorName,
		LoginURL:   variables.LoginURL,
	}

	// Render HTML version
	htmlBody, err := te.renderHTML(data)
	if err != nil {
		return nil, fmt.Errorf("render HTML: %w", err)
	}

	// Render text version
	textBody, err := te.renderText(data)
	if err != nil {
		return nil, fmt.Errorf("render text: %w", err)
	}

	return &RenderedEmail{
		Subject:  "Security Alert: Your Session Has Been Terminated",
		HTMLBody: htmlBody,
		TextBody: textBody,
	}, nil
}

// renderHTML renders the HTML email body
func (te *TemplateEngine) renderHTML(data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := te.htmlTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute HTML template: %w", err)
	}
	return buf.String(), nil
}

// renderText renders the plain text email body
func (te *TemplateEngine) renderText(data TemplateData) (string, error) {
	var buf bytes.Buffer
	if err := te.textTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute text template: %w", err)
	}
	return buf.String(), nil
}

// validateVariables ensures all required template variables are present
func (te *TemplateEngine) validateVariables(vars types.NotificationVariables) error {
	if vars.Username == "" {
		return fmt.Errorf("username is required")
	}
	if vars.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	if vars.LoginURL == "" {
		return fmt.Errorf("login_url is required")
	}
	// DeviceInfo, Location, ActorName, Reason can be empty
	return nil
}

// RenderedEmail represents a fully rendered email with both HTML and text versions
type RenderedEmail struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// CreateNotificationVariables creates NotificationVariables from session and actor info
func CreateNotificationVariables(
	username string,
	timestamp time.Time,
	reason string,
	deviceInfo string,
	location string,
	actorName string,
	loginURL string,
) types.NotificationVariables {
	// Set defaults for empty values
	if deviceInfo == "" {
		deviceInfo = "Unknown Device"
	}
	if location == "" {
		location = "Unknown Location"
	}
	if actorName == "" {
		actorName = "System Administrator"
	}
	if reason == "" {
		reason = "Security policy enforcement"
	}

	return types.NotificationVariables{
		Username:   username,
		Timestamp:  timestamp,
		Reason:     reason,
		DeviceInfo: deviceInfo,
		Location:   location,
		ActorName:  actorName,
		LoginURL:   loginURL,
	}
}
