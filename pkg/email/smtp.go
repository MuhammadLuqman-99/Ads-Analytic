package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	From     string
	FromName string
}

// SMTPProvider implements email sending via SMTP
type SMTPProvider struct {
	config SMTPConfig
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(config SMTPConfig) (*SMTPProvider, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	return &SMTPProvider{config: config}, nil
}

// Send sends an email via SMTP
func (p *SMTPProvider) Send(ctx context.Context, msg *Message) error {
	// Build email headers and body
	var builder strings.Builder

	// From header
	if p.config.FromName != "" {
		builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", p.config.FromName, p.config.From))
	} else {
		builder.WriteString(fmt.Sprintf("From: %s\r\n", p.config.From))
	}

	// To header
	if msg.ToName != "" {
		builder.WriteString(fmt.Sprintf("To: %s <%s>\r\n", msg.ToName, msg.To))
	} else {
		builder.WriteString(fmt.Sprintf("To: %s\r\n", msg.To))
	}

	// Subject header
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// MIME headers
	builder.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" {
		builder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.HTMLBody)
	} else {
		builder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		builder.WriteString("\r\n")
		builder.WriteString(msg.TextBody)
	}

	emailBody := builder.String()
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	// Create authentication
	var auth smtp.Auth
	if p.config.Username != "" && p.config.Password != "" {
		auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	}

	// Send with TLS if enabled
	if p.config.UseTLS {
		return p.sendWithTLS(addr, auth, p.config.From, []string{msg.To}, []byte(emailBody))
	}

	// Send without TLS
	return smtp.SendMail(addr, auth, p.config.From, []string{msg.To}, []byte(emailBody))
}

// sendWithTLS sends email using explicit TLS connection
func (p *SMTPProvider) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: p.config.Host,
	})
	if err != nil {
		// Fall back to STARTTLS
		return p.sendWithSTARTTLS(addr, auth, from, to, msg)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, p.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate if credentials provided
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// sendWithSTARTTLS sends email using STARTTLS
func (p *SMTPProvider) sendWithSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName: p.config.Host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate if credentials provided
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// Name returns the provider name
func (p *SMTPProvider) Name() string {
	return "smtp"
}
