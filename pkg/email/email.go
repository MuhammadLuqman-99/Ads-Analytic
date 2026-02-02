package email

import (
	"context"
	"fmt"
)

// Provider defines the interface for email providers
type Provider interface {
	// Send sends an email
	Send(ctx context.Context, msg *Message) error
	// Name returns the provider name
	Name() string
}

// Message represents an email message
type Message struct {
	To          string
	ToName      string
	Subject     string
	HTMLBody    string
	TextBody    string
	ReplyTo     string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// Config holds email configuration
type Config struct {
	Provider       string
	From           string
	FromName       string
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPassword   string
	SMTPUseTLS     bool
	SendGridAPIKey string
	ResendAPIKey   string
}

// Sender is the main email sending service
type Sender struct {
	provider Provider
	from     string
	fromName string
}

// NewSender creates a new email sender based on config
func NewSender(cfg Config) (*Sender, error) {
	var provider Provider
	var err error

	switch cfg.Provider {
	case "smtp":
		provider, err = NewSMTPProvider(SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUser,
			Password: cfg.SMTPPassword,
			UseTLS:   cfg.SMTPUseTLS,
			From:     cfg.From,
			FromName: cfg.FromName,
		})
	case "sendgrid":
		provider, err = NewSendGridProvider(cfg.SendGridAPIKey, cfg.From, cfg.FromName)
	case "resend":
		provider, err = NewResendProvider(cfg.ResendAPIKey, cfg.From, cfg.FromName)
	case "noop", "":
		// No-op provider for development/testing
		provider = &NoOpProvider{}
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}

	return &Sender{
		provider: provider,
		from:     cfg.From,
		fromName: cfg.FromName,
	}, nil
}

// Send sends an email message
func (s *Sender) Send(ctx context.Context, msg *Message) error {
	return s.provider.Send(ctx, msg)
}

// SendHTML sends an HTML email
func (s *Sender) SendHTML(ctx context.Context, to, toName, subject, htmlBody string) error {
	return s.Send(ctx, &Message{
		To:       to,
		ToName:   toName,
		Subject:  subject,
		HTMLBody: htmlBody,
	})
}

// SendText sends a plain text email
func (s *Sender) SendText(ctx context.Context, to, toName, subject, textBody string) error {
	return s.Send(ctx, &Message{
		To:       to,
		ToName:   toName,
		Subject:  subject,
		TextBody: textBody,
	})
}

// ProviderName returns the name of the underlying provider
func (s *Sender) ProviderName() string {
	return s.provider.Name()
}

// NoOpProvider is a no-operation provider for testing
type NoOpProvider struct{}

func (p *NoOpProvider) Send(ctx context.Context, msg *Message) error {
	// Log the email instead of sending
	fmt.Printf("[NoOp Email] To: %s, Subject: %s\n", msg.To, msg.Subject)
	return nil
}

func (p *NoOpProvider) Name() string {
	return "noop"
}
