package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// EmailType represents different types of emails
type EmailType string

const (
	EmailTypeWelcome             EmailType = "welcome"
	EmailTypeVerification        EmailType = "verification"
	EmailTypePasswordReset       EmailType = "password_reset"
	EmailTypePlatformConnected   EmailType = "platform_connected"
	EmailTypeWeeklySummary       EmailType = "weekly_summary"
	EmailTypeTokenExpired        EmailType = "token_expired"
	EmailTypeSubscriptionConfirm EmailType = "subscription_confirm"
	EmailTypePaymentFailed       EmailType = "payment_failed"
	EmailTypeConnectReminder     EmailType = "connect_reminder"
	EmailTypeInactiveReminder    EmailType = "inactive_reminder"
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending   EmailStatus = "pending"
	EmailStatusSent      EmailStatus = "sent"
	EmailStatusFailed    EmailStatus = "failed"
	EmailStatusOpened    EmailStatus = "opened"
	EmailStatusClicked   EmailStatus = "clicked"
	EmailStatusBounced   EmailStatus = "bounced"
)

// Email represents an email to be sent
type Email struct {
	ID          string                 `json:"id"`
	To          string                 `json:"to"`
	ToName      string                 `json:"to_name"`
	Subject     string                 `json:"subject"`
	Type        EmailType              `json:"type"`
	TemplateID  string                 `json:"template_id"`
	Data        map[string]interface{} `json:"data"`
	Status      EmailStatus            `json:"status"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	OpenedAt    *time.Time             `json:"opened_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// EmailService interface for sending emails
type EmailService interface {
	Send(ctx context.Context, email *Email) error
	SendTemplate(ctx context.Context, email *Email, htmlContent string) error
	GetStatus(ctx context.Context, emailID string) (*EmailStatus, error)
}

// Config holds email service configuration
type Config struct {
	Provider    string // "resend" or "sendgrid"
	APIKey      string
	FromEmail   string
	FromName    string
	ReplyTo     string
	BaseURL     string // For tracking links
	Environment string // "production" or "development"
}

// NewEmailService creates a new email service based on provider
func NewEmailService(cfg *Config) (EmailService, error) {
	switch cfg.Provider {
	case "resend":
		return NewResendService(cfg), nil
	case "sendgrid":
		return NewSendGridService(cfg), nil
	default:
		return NewResendService(cfg), nil
	}
}

// ResendService implements EmailService using Resend API
type ResendService struct {
	config     *Config
	httpClient *http.Client
}

// NewResendService creates a new Resend email service
func NewResendService(cfg *Config) *ResendService {
	return &ResendService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ResendRequest represents a Resend API request
type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	ReplyTo string   `json:"reply_to,omitempty"`
	Tags    []Tag    `json:"tags,omitempty"`
}

// Tag represents an email tag for tracking
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ResendResponse represents a Resend API response
type ResendResponse struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

// Send sends an email using Resend
func (s *ResendService) Send(ctx context.Context, email *Email) error {
	// Render template
	htmlContent, err := RenderTemplate(email.Type, email.Data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.SendTemplate(ctx, email, htmlContent)
}

// SendTemplate sends an email with custom HTML content
func (s *ResendService) SendTemplate(ctx context.Context, email *Email, htmlContent string) error {
	from := fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail)

	reqBody := ResendRequest{
		From:    from,
		To:      []string{email.To},
		Subject: email.Subject,
		HTML:    htmlContent,
		ReplyTo: s.config.ReplyTo,
		Tags: []Tag{
			{Name: "email_type", Value: string(email.Type)},
			{Name: "email_id", Value: email.ID},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var resendResp ResendResponse
	if err := json.NewDecoder(resp.Body).Decode(&resendResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: %s", resendResp.Error)
	}

	return nil
}

// GetStatus gets the status of an email (Resend)
func (s *ResendService) GetStatus(ctx context.Context, emailID string) (*EmailStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.resend.com/emails/%s", emailID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		LastEvent string `json:"last_event"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var status EmailStatus
	switch result.LastEvent {
	case "delivered":
		status = EmailStatusSent
	case "opened":
		status = EmailStatusOpened
	case "clicked":
		status = EmailStatusClicked
	case "bounced":
		status = EmailStatusBounced
	default:
		status = EmailStatusPending
	}

	return &status, nil
}

// SendGridService implements EmailService using SendGrid API
type SendGridService struct {
	config     *Config
	httpClient *http.Client
}

// NewSendGridService creates a new SendGrid email service
func NewSendGridService(cfg *Config) *SendGridService {
	return &SendGridService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendGridRequest represents a SendGrid API request
type SendGridRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             EmailAddress      `json:"from"`
	ReplyTo          *EmailAddress     `json:"reply_to,omitempty"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
	TrackingSettings *TrackingSettings `json:"tracking_settings,omitempty"`
}

// Personalization represents SendGrid personalization
type Personalization struct {
	To []EmailAddress `json:"to"`
}

// EmailAddress represents an email address
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Content represents email content
type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// TrackingSettings for SendGrid
type TrackingSettings struct {
	OpenTracking  *OpenTracking  `json:"open_tracking,omitempty"`
	ClickTracking *ClickTracking `json:"click_tracking,omitempty"`
}

// OpenTracking settings
type OpenTracking struct {
	Enable bool `json:"enable"`
}

// ClickTracking settings
type ClickTracking struct {
	Enable bool `json:"enable"`
}

// Send sends an email using SendGrid
func (s *SendGridService) Send(ctx context.Context, email *Email) error {
	htmlContent, err := RenderTemplate(email.Type, email.Data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.SendTemplate(ctx, email, htmlContent)
}

// SendTemplate sends an email with custom HTML content
func (s *SendGridService) SendTemplate(ctx context.Context, email *Email, htmlContent string) error {
	reqBody := SendGridRequest{
		Personalizations: []Personalization{
			{
				To: []EmailAddress{
					{Email: email.To, Name: email.ToName},
				},
			},
		},
		From: EmailAddress{
			Email: s.config.FromEmail,
			Name:  s.config.FromName,
		},
		Subject: email.Subject,
		Content: []Content{
			{Type: "text/html", Value: htmlContent},
		},
		TrackingSettings: &TrackingSettings{
			OpenTracking:  &OpenTracking{Enable: true},
			ClickTracking: &ClickTracking{Enable: true},
		},
	}

	if s.config.ReplyTo != "" {
		reqBody.ReplyTo = &EmailAddress{Email: s.config.ReplyTo}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("sendgrid API error: %v", errResp)
	}

	return nil
}

// GetStatus gets the status of an email (SendGrid - simplified)
func (s *SendGridService) GetStatus(ctx context.Context, emailID string) (*EmailStatus, error) {
	// SendGrid doesn't have a simple status endpoint
	// Would need to use Activity API or webhooks
	status := EmailStatusSent
	return &status, nil
}

// NewEmail creates a new email with a unique ID
func NewEmail(to, toName, subject string, emailType EmailType, data map[string]interface{}) *Email {
	return &Email{
		ID:        uuid.New().String(),
		To:        to,
		ToName:    toName,
		Subject:   subject,
		Type:      emailType,
		Data:      data,
		Status:    EmailStatusPending,
		CreatedAt: time.Now(),
	}
}

// RenderTemplate renders an email template
func RenderTemplate(emailType EmailType, data map[string]interface{}) (string, error) {
	tmpl, ok := templates[emailType]
	if !ok {
		return "", fmt.Errorf("template not found for type: %s", emailType)
	}

	t, err := template.New("email").Parse(baseTemplate)
	if err != nil {
		return "", err
	}

	t, err = t.Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
