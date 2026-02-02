package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ResendProvider implements email sending via Resend API
type ResendProvider struct {
	apiKey   string
	from     string
	fromName string
	client   *http.Client
}

// NewResendProvider creates a new Resend email provider
func NewResendProvider(apiKey, from, fromName string) (*ResendProvider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Resend API key is required")
	}
	return &ResendProvider{
		apiKey:   apiKey,
		from:     from,
		fromName: fromName,
		client:   &http.Client{},
	}, nil
}

// resendRequest represents a Resend API request
type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

// Send sends an email via Resend
func (p *ResendProvider) Send(ctx context.Context, msg *Message) error {
	// Build from address
	from := p.from
	if p.fromName != "" {
		from = fmt.Sprintf("%s <%s>", p.fromName, p.from)
	}

	// Build request
	req := resendRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		HTML:    msg.HTMLBody,
		Text:    msg.TextBody,
		ReplyTo: msg.ReplyTo,
	}

	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal Resend request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send Resend request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		return fmt.Errorf("Resend API error: status %d", resp.StatusCode)
	}

	return nil
}

// Name returns the provider name
func (p *ResendProvider) Name() string {
	return "resend"
}
