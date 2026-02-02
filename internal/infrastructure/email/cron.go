package email

import (
	"context"
	"log"
	"time"
)

// CronJob represents a scheduled email job
type CronJob struct {
	Name     string
	Schedule string // cron expression
	Handler  func(ctx context.Context) error
}

// EmailCron manages scheduled email jobs
type EmailCron struct {
	trigger *EmailTrigger
	jobs    []CronJob
	stopCh  chan struct{}
	running bool
}

// NewEmailCron creates a new email cron manager
func NewEmailCron(trigger *EmailTrigger) *EmailCron {
	return &EmailCron{
		trigger: trigger,
		stopCh:  make(chan struct{}),
	}
}

// RegisterJobs registers all scheduled email jobs
func (c *EmailCron) RegisterJobs() {
	c.jobs = []CronJob{
		{
			Name:     "weekly_summary",
			Schedule: "0 9 * * 1", // Every Monday at 9 AM
			Handler:  c.sendWeeklySummaries,
		},
		{
			Name:     "connect_reminders",
			Schedule: "0 10 * * *", // Every day at 10 AM
			Handler:  c.sendConnectReminders,
		},
		{
			Name:     "inactive_reminders",
			Schedule: "0 14 * * *", // Every day at 2 PM
			Handler:  c.sendInactiveReminders,
		},
		{
			Name:     "token_expiry_check",
			Schedule: "0 8 * * *", // Every day at 8 AM
			Handler:  c.checkTokenExpiry,
		},
	}
}

// Start starts the cron scheduler
// Note: In production, use a proper cron library like robfig/cron
func (c *EmailCron) Start(ctx context.Context) {
	c.running = true
	c.RegisterJobs()

	log.Println("[EmailCron] Starting cron jobs...")

	// Simple ticker-based implementation
	// In production, use robfig/cron for proper cron scheduling
	go c.runWeeklySummary(ctx)
	go c.runDailyJobs(ctx)
}

// Stop stops the cron scheduler
func (c *EmailCron) Stop() {
	if !c.running {
		return
	}
	c.running = false
	close(c.stopCh)
	log.Println("[EmailCron] Stopped")
}

// runWeeklySummary runs weekly summary job
func (c *EmailCron) runWeeklySummary(ctx context.Context) {
	// Calculate time until next Monday 9 AM
	now := time.Now()
	next := nextMonday9AM(now)
	timer := time.NewTimer(time.Until(next))

	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-c.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			log.Println("[EmailCron] Running weekly summary job")
			if err := c.sendWeeklySummaries(ctx); err != nil {
				log.Printf("[EmailCron] Weekly summary error: %v", err)
			}
			// Schedule next run
			timer.Reset(7 * 24 * time.Hour)
		}
	}
}

// runDailyJobs runs daily jobs
func (c *EmailCron) runDailyJobs(ctx context.Context) {
	// Run at 10 AM daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			hour := time.Now().Hour()

			// Connect reminders at 10 AM
			if hour == 10 {
				log.Println("[EmailCron] Running connect reminders job")
				if err := c.sendConnectReminders(ctx); err != nil {
					log.Printf("[EmailCron] Connect reminders error: %v", err)
				}
			}

			// Inactive reminders at 2 PM
			if hour == 14 {
				log.Println("[EmailCron] Running inactive reminders job")
				if err := c.sendInactiveReminders(ctx); err != nil {
					log.Printf("[EmailCron] Inactive reminders error: %v", err)
				}
			}

			// Token expiry check at 8 AM
			if hour == 8 {
				log.Println("[EmailCron] Running token expiry check")
				if err := c.checkTokenExpiry(ctx); err != nil {
					log.Printf("[EmailCron] Token expiry check error: %v", err)
				}
			}
		}
	}
}

// sendWeeklySummaries sends weekly performance summaries
func (c *EmailCron) sendWeeklySummaries(ctx context.Context) error {
	// TODO: Get all users with weekly digest enabled
	// For each user:
	// 1. Aggregate their metrics for the past week
	// 2. Generate insights
	// 3. Send email

	log.Println("[EmailCron] Weekly summaries: Implementation needed")
	// Example implementation:
	// users, err := userRepo.GetUsersWithWeeklyDigest(ctx)
	// for _, user := range users {
	//     summary := generateWeeklySummary(ctx, user.ID)
	//     c.trigger.TriggerWeeklySummaryEmail(ctx, user, summary)
	// }

	return nil
}

// sendConnectReminders sends reminders to users who haven't connected platforms
func (c *EmailCron) sendConnectReminders(ctx context.Context) error {
	// TODO: Get users who:
	// 1. Registered more than 24 hours ago
	// 2. Haven't connected any platform
	// 3. Haven't received this reminder yet

	log.Println("[EmailCron] Connect reminders: Implementation needed")
	// Example implementation:
	// users, err := userRepo.GetUsersWithoutPlatforms(ctx, 24*time.Hour)
	// for _, user := range users {
	//     c.trigger.TriggerConnectReminderEmail(ctx, user)
	// }

	return nil
}

// sendInactiveReminders sends reminders to inactive users
func (c *EmailCron) sendInactiveReminders(ctx context.Context) error {
	// TODO: Get users who:
	// 1. Haven't logged in for 3+ days
	// 2. Have connected platforms
	// 3. Haven't received this reminder recently (e.g., 7 days)

	log.Println("[EmailCron] Inactive reminders: Implementation needed")
	// Example implementation:
	// users, err := userRepo.GetInactiveUsers(ctx, 3*24*time.Hour)
	// for _, user := range users {
	//     recentSpend, recentConversions := getRecentMetrics(ctx, user.ID)
	//     c.trigger.TriggerInactiveReminderEmail(ctx, user, recentSpend, recentConversions)
	// }

	return nil
}

// checkTokenExpiry checks for expiring tokens and sends reminders
func (c *EmailCron) checkTokenExpiry(ctx context.Context) error {
	// TODO: Get platform connections with tokens expiring in:
	// 1. 7 days
	// 2. 3 days
	// 3. 1 day
	// 4. Already expired

	log.Println("[EmailCron] Token expiry check: Implementation needed")
	// Example implementation:
	// connections, err := connectionRepo.GetExpiringTokens(ctx, 7*24*time.Hour)
	// for _, conn := range connections {
	//     user, _ := userRepo.GetByID(ctx, conn.UserID)
	//     platform := &PlatformData{Name: conn.Platform, TokenExpiresAt: conn.ExpiresAt}
	//     c.trigger.TriggerTokenExpiredEmail(ctx, user, platform)
	// }

	return nil
}

// Helper function to calculate next Monday 9 AM
func nextMonday9AM(now time.Time) time.Time {
	// Find days until next Monday
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 && now.Hour() >= 9 {
		daysUntilMonday = 7
	}

	next := now.AddDate(0, 0, daysUntilMonday)
	next = time.Date(next.Year(), next.Month(), next.Day(), 9, 0, 0, 0, next.Location())

	return next
}
