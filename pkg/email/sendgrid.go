package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SendGridProvider implements email sending via SendGrid API
type SendGridProvider struct {
	apiKey   string
	from     string
	fromName string
	client   *http.Client
}

// NewSendGridProvider creates a new SendGrid email provider
func NewSendGridProvider(apiKey, from, fromName string) (*SendGridProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("SendGrid API key is required")
	}
	return &SendGridProvider{
		apiKey:   apiKey,
		from:     from,
		fromName: fromName,
		client:   &http.Client{},
	}, nil
}

// sendGridRequest represents a SendGrid API request
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send sends an email via SendGrid
func (p *SendGridProvider) Send(ctx context.Context, msg *Message) error {
	// Build request
	req := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{
				To: []sendGridEmail{
					{Email: msg.To, Name: msg.ToName},
				},
			},
		},
		From: sendGridEmail{
			Email: p.from,
			Name:  p.fromName,
		},
		Subject: msg.Subject,
		Content: []sendGridContent{},
	}

	// Add content
	if msg.TextBody != "" {
		req.Content = append(req.Content, sendGridContent{
			Type:  "text/plain",
			Value: msg.TextBody,
		})
	}
	if msg.HTMLBody != "" {
		req.Content = append(req.Content, sendGridContent{
			Type:  "text/html",
			Value: msg.HTMLBody,
		})
	}

	// Add reply-to
	if msg.ReplyTo != "" {
		req.ReplyTo = &sendGridEmail{Email: msg.ReplyTo}
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal SendGrid request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send SendGrid request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		return fmt.Errorf("SendGrid API error: status %d", resp.StatusCode)
	}

	return nil
}

// Name returns the provider name
func (p *SendGridProvider) Name() string {
	return "sendgrid"
}
