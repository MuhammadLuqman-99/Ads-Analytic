// Package email provides automated email functionality for AdsAnalytic
//
// Features:
// - Email service abstraction (Resend, SendGrid)
// - HTML email templates
// - Redis-based email queue
// - Automated triggers for various events
// - Background worker for processing emails
// - Cron jobs for scheduled emails
//
// Usage:
//
//	// Create email service
//	config := &email.Config{
//	    Provider:  "resend",
//	    APIKey:    os.Getenv("RESEND_API_KEY"),
//	    FromEmail: "noreply@adsanalytic.com",
//	    FromName:  "AdsAnalytic",
//	    BaseURL:   "https://adsanalytic.com",
//	}
//
//	service, err := email.NewEmailService(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create queue
//	queue := email.NewEmailQueue(redisClient, config)
//
//	// Create trigger
//	trigger := email.NewEmailTrigger(queue, config)
//
//	// Send welcome email
//	trigger.TriggerWelcomeEmail(ctx, &email.UserData{
//	    Email: "user@example.com",
//	    Name:  "John Doe",
//	})
//
//	// Start worker
//	worker := email.NewWorker(queue, service, config, nil)
//	worker.Start(ctx)
//
// Email Types:
//   - welcome: Sent after user registration
//   - verification: Email verification
//   - password_reset: Password reset request
//   - platform_connected: First platform connected
//   - weekly_summary: Weekly performance digest
//   - token_expired: OAuth token expired
//   - subscription_confirm: Subscription activated
//   - payment_failed: Payment failure warning
//   - connect_reminder: Reminder to connect platform
//   - inactive_reminder: Reminder for inactive users
package email

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Manager manages the entire email system
type Manager struct {
	Service EmailService
	Queue   *EmailQueue
	Trigger *EmailTrigger
	Worker  *Worker
	Cron    *EmailCron
	Config  *Config
}

// NewManager creates a new email manager with all components
func NewManager(redisClient *redis.Client, config *Config, workerConfig *WorkerConfig) (*Manager, error) {
	// Create service
	service, err := NewEmailService(config)
	if err != nil {
		return nil, err
	}

	// Create queue
	queue := NewEmailQueue(redisClient, config)

	// Create trigger
	trigger := NewEmailTrigger(queue, config)

	// Create worker
	worker := NewWorker(queue, service, config, workerConfig)

	// Create cron
	cron := NewEmailCron(trigger)

	return &Manager{
		Service: service,
		Queue:   queue,
		Trigger: trigger,
		Worker:  worker,
		Cron:    cron,
		Config:  config,
	}, nil
}

// Start starts the email worker and cron jobs
func (m *Manager) Start(ctx context.Context) error {
	// Start worker
	if err := m.Worker.Start(ctx); err != nil {
		return err
	}

	// Start cron
	m.Cron.Start(ctx)

	return nil
}

// Stop stops the email worker and cron jobs
func (m *Manager) Stop() {
	m.Cron.Stop()
	m.Worker.Stop()
}

// GetStats returns email system statistics
func (m *Manager) GetStats(ctx context.Context) (*WorkerStats, error) {
	return m.Worker.GetStats(ctx)
}
