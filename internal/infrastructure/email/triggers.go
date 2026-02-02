package email

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmailTrigger handles automated email triggers
type EmailTrigger struct {
	queue   *EmailQueue
	config  *Config
}

// NewEmailTrigger creates a new email trigger handler
func NewEmailTrigger(queue *EmailQueue, config *Config) *EmailTrigger {
	return &EmailTrigger{
		queue:  queue,
		config: config,
	}
}

// UserData contains user information for emails
type UserData struct {
	ID        uuid.UUID
	Email     string
	Name      string
	CreatedAt time.Time
	LastLogin *time.Time
}

// PlatformData contains platform connection information
type PlatformData struct {
	Name           string
	ConnectedAt    time.Time
	TokenExpiresAt *time.Time
}

// SubscriptionData contains subscription information
type SubscriptionData struct {
	PlanName        string
	Amount          float64
	NextBillingDate time.Time
	Features        []string
}

// PaymentFailureData contains payment failure information
type PaymentFailureData struct {
	PlanName       string
	Amount         float64
	FailureReason  string
	AttemptCount   int
	GracePeriodEnd *time.Time
}

// WeeklySummaryData contains weekly summary information
type WeeklySummaryData struct {
	WeekStart         string
	WeekEnd           string
	TotalSpend        float64
	TotalRevenue      float64
	ROAS              float64
	Conversions       int
	SpendChange       float64
	RevenueChange     float64
	ROASChange        float64
	ConversionsChange float64
	SpendUp           bool
	RevenueUp         bool
	ROASUp            bool
	ConversionsUp     bool
	Platforms         []PlatformSummary
	TopCampaign       *CampaignSummary
	Insights          []string
}

// PlatformSummary contains platform-specific summary
type PlatformSummary struct {
	Name        string
	Spend       float64
	ROAS        float64
	Conversions int
}

// CampaignSummary contains campaign summary
type CampaignSummary struct {
	Name    string
	ROAS    float64
	Revenue float64
}

// TriggerWelcomeEmail sends a welcome email after registration
func (t *EmailTrigger) TriggerWelcomeEmail(ctx context.Context, user *UserData) error {
	data := map[string]interface{}{
		"Name":    user.Name,
		"Email":   user.Email,
		"BaseURL": t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeWelcome, data),
		EmailTypeWelcome,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerVerificationEmail sends an email verification email
func (t *EmailTrigger) TriggerVerificationEmail(ctx context.Context, user *UserData, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", t.config.BaseURL, token)

	data := map[string]interface{}{
		"Name":            user.Name,
		"Email":           user.Email,
		"VerificationURL": verificationURL,
		"BaseURL":         t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeVerification, data),
		EmailTypeVerification,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerPasswordResetEmail sends a password reset email
func (t *EmailTrigger) TriggerPasswordResetEmail(ctx context.Context, user *UserData, token, ipAddress string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", t.config.BaseURL, token)

	data := map[string]interface{}{
		"Name":        user.Name,
		"Email":       user.Email,
		"ResetURL":    resetURL,
		"BaseURL":     t.config.BaseURL,
		"IPAddress":   ipAddress,
		"RequestTime": time.Now().Format("2 Jan 2006, 3:04 PM"),
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypePasswordReset, data),
		EmailTypePasswordReset,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerPlatformConnectedEmail sends a congratulations email when platform is connected
func (t *EmailTrigger) TriggerPlatformConnectedEmail(ctx context.Context, user *UserData, platform *PlatformData, hasMultiplePlatforms bool) error {
	data := map[string]interface{}{
		"Name":                 user.Name,
		"PlatformName":         platform.Name,
		"HasMultiplePlatforms": hasMultiplePlatforms,
		"BaseURL":              t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypePlatformConnected, data),
		EmailTypePlatformConnected,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerTokenExpiredEmail sends a reminder when token expires
func (t *EmailTrigger) TriggerTokenExpiredEmail(ctx context.Context, user *UserData, platform *PlatformData) error {
	data := map[string]interface{}{
		"Name":         user.Name,
		"PlatformName": platform.Name,
		"BaseURL":      t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeTokenExpired, data),
		EmailTypeTokenExpired,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerSubscriptionConfirmEmail sends subscription confirmation
func (t *EmailTrigger) TriggerSubscriptionConfirmEmail(ctx context.Context, user *UserData, subscription *SubscriptionData) error {
	data := map[string]interface{}{
		"Name":            user.Name,
		"PlanName":        subscription.PlanName,
		"Amount":          fmt.Sprintf("%.2f", subscription.Amount),
		"NextBillingDate": subscription.NextBillingDate.Format("2 Jan 2006"),
		"Features":        subscription.Features,
		"BaseURL":         t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeSubscriptionConfirm, data),
		EmailTypeSubscriptionConfirm,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerPaymentFailedEmail sends payment failed warning
func (t *EmailTrigger) TriggerPaymentFailedEmail(ctx context.Context, user *UserData, payment *PaymentFailureData) error {
	data := map[string]interface{}{
		"Name":          user.Name,
		"PlanName":      payment.PlanName,
		"Amount":        fmt.Sprintf("%.2f", payment.Amount),
		"FailureReason": payment.FailureReason,
		"AttemptCount":  payment.AttemptCount,
		"BaseURL":       t.config.BaseURL,
	}

	if payment.GracePeriodEnd != nil {
		data["GracePeriodEnd"] = payment.GracePeriodEnd.Format("2 Jan 2006")
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypePaymentFailed, data),
		EmailTypePaymentFailed,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerWeeklySummaryEmail sends weekly performance summary
func (t *EmailTrigger) TriggerWeeklySummaryEmail(ctx context.Context, user *UserData, summary *WeeklySummaryData) error {
	data := map[string]interface{}{
		"Name":              user.Name,
		"WeekStart":         summary.WeekStart,
		"WeekEnd":           summary.WeekEnd,
		"TotalSpend":        fmt.Sprintf("%.2f", summary.TotalSpend),
		"TotalRevenue":      fmt.Sprintf("%.2f", summary.TotalRevenue),
		"ROAS":              fmt.Sprintf("%.2f", summary.ROAS),
		"Conversions":       summary.Conversions,
		"SpendChange":       fmt.Sprintf("%.1f", summary.SpendChange),
		"RevenueChange":     fmt.Sprintf("%.1f", summary.RevenueChange),
		"ROASChange":        fmt.Sprintf("%.2f", summary.ROASChange),
		"ConversionsChange": fmt.Sprintf("%.1f", summary.ConversionsChange),
		"SpendUp":           summary.SpendUp,
		"RevenueUp":         summary.RevenueUp,
		"ROASUp":            summary.ROASUp,
		"ConversionsUp":     summary.ConversionsUp,
		"Platforms":         summary.Platforms,
		"TopCampaign":       summary.TopCampaign,
		"Insights":          summary.Insights,
		"BaseURL":           t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeWeeklySummary, data),
		EmailTypeWeeklySummary,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// TriggerConnectReminderEmail sends a reminder to connect platform (24h after register)
func (t *EmailTrigger) TriggerConnectReminderEmail(ctx context.Context, user *UserData) error {
	data := map[string]interface{}{
		"Name":    user.Name,
		"BaseURL": t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeConnectReminder, data),
		EmailTypeConnectReminder,
		data,
	)

	// Schedule for 24 hours after registration
	delay := 24 * time.Hour
	if time.Since(user.CreatedAt) < delay {
		delay = delay - time.Since(user.CreatedAt)
	} else {
		delay = 0
	}

	return t.queue.EnqueueWithDelay(ctx, email, delay)
}

// TriggerInactiveReminderEmail sends a reminder for inactive users (3 days)
func (t *EmailTrigger) TriggerInactiveReminderEmail(ctx context.Context, user *UserData, recentSpend float64, recentConversions int) error {
	lastLoginDays := 3
	if user.LastLogin != nil {
		lastLoginDays = int(time.Since(*user.LastLogin).Hours() / 24)
	}

	data := map[string]interface{}{
		"Name":              user.Name,
		"LastLoginDays":     lastLoginDays,
		"HasData":           recentSpend > 0 || recentConversions > 0,
		"RecentSpend":       fmt.Sprintf("%.2f", recentSpend),
		"RecentConversions": recentConversions,
		"BaseURL":           t.config.BaseURL,
	}

	email := NewEmail(
		user.Email,
		user.Name,
		GetSubject(EmailTypeInactiveReminder, data),
		EmailTypeInactiveReminder,
		data,
	)

	return t.queue.Enqueue(ctx, email)
}

// ScheduleConnectReminder schedules a connect reminder for a new user
func (t *EmailTrigger) ScheduleConnectReminder(ctx context.Context, user *UserData) error {
	// Schedule for 24 hours from now
	return t.queue.EnqueueWithDelay(ctx, &Email{
		ID:      uuid.New().String(),
		To:      user.Email,
		ToName:  user.Name,
		Type:    EmailTypeConnectReminder,
		Status:  EmailStatusPending,
		Data: map[string]interface{}{
			"Name":    user.Name,
			"BaseURL": t.config.BaseURL,
		},
		CreatedAt: time.Now(),
	}, 24*time.Hour)
}
