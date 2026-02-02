package repository

import (
	"context"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates a user
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// VerifyEmail marks the email as verified
	VerifyEmail(ctx context.Context, id uuid.UUID) error

	// UpdatePassword updates the user's password hash
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// UpdateProfile updates the user's profile information
	UpdateProfile(ctx context.Context, id uuid.UUID, firstName, lastName, phone string) error
}

// VerificationTokenRepository defines the interface for verification token persistence
type VerificationTokenRepository interface {
	// Create creates a new verification token
	Create(ctx context.Context, token *entity.VerificationToken) error

	// GetByToken retrieves a token by its value
	GetByToken(ctx context.Context, token string) (*entity.VerificationToken, error)

	// GetByUserAndType retrieves the latest token for a user by type
	GetByUserAndType(ctx context.Context, userID uuid.UUID, tokenType entity.TokenType) (*entity.VerificationToken, error)

	// MarkAsUsed marks a token as used
	MarkAsUsed(ctx context.Context, id uuid.UUID) error

	// DeleteByUser deletes all tokens for a user
	DeleteByUser(ctx context.Context, userID uuid.UUID) error

	// DeleteByUserAndType deletes all tokens of a specific type for a user
	DeleteByUserAndType(ctx context.Context, userID uuid.UUID, tokenType entity.TokenType) error

	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context) error
}

// OrganizationRepository defines the interface for organization data persistence
type OrganizationRepository interface {
	// Create creates a new organization
	Create(ctx context.Context, org *entity.Organization) error

	// GetByID retrieves an organization by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error)

	// GetBySlug retrieves an organization by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Organization, error)

	// Update updates an organization
	Update(ctx context.Context, org *entity.Organization) error

	// Delete deletes an organization
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists organizations with pagination
	List(ctx context.Context, pagination *entity.Pagination) ([]entity.Organization, error)

	// GetByUserID retrieves all organizations a user belongs to
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Organization, error)
}

// OrganizationMemberRepository defines the interface for organization member data persistence
type OrganizationMemberRepository interface {
	// Create creates a new organization member
	Create(ctx context.Context, member *entity.OrganizationMember) error

	// GetByID retrieves a member by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.OrganizationMember, error)

	// GetByOrgAndUser retrieves a member by organization and user ID
	GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*entity.OrganizationMember, error)

	// Update updates a member
	Update(ctx context.Context, member *entity.OrganizationMember) error

	// Delete removes a member from an organization
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByOrganization lists all members of an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.OrganizationMember, error)

	// ListByUser lists all organizations a user is a member of
	ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.OrganizationMember, error)

	// UpdateRole updates a member's role
	UpdateRole(ctx context.Context, id uuid.UUID, role entity.UserRole) error
}

// ConnectedAccountRepository defines the interface for connected account data persistence
type ConnectedAccountRepository interface {
	// Create creates a new connected account
	Create(ctx context.Context, account *entity.ConnectedAccount) error

	// GetByID retrieves a connected account by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ConnectedAccount, error)

	// GetByPlatformAccountID retrieves by platform-specific account ID
	GetByPlatformAccountID(ctx context.Context, orgID uuid.UUID, platform entity.Platform, platformAccountID string) (*entity.ConnectedAccount, error)

	// Update updates a connected account
	Update(ctx context.Context, account *entity.ConnectedAccount) error

	// Delete deletes a connected account
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByOrganization lists all connected accounts for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error)

	// ListByPlatform lists all connected accounts for a platform
	ListByPlatform(ctx context.Context, orgID uuid.UUID, platform entity.Platform) ([]entity.ConnectedAccount, error)

	// ListExpiring lists accounts with tokens expiring soon
	ListExpiring(ctx context.Context, withinMinutes int) ([]entity.ConnectedAccount, error)

	// ListActive lists all active accounts that need syncing
	ListActive(ctx context.Context) ([]entity.ConnectedAccount, error)

	// UpdateTokens updates the OAuth tokens for an account
	UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *interface{}) error

	// UpdateSyncStatus updates the sync status for an account
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus, syncError string) error

	// UpdateLastSynced updates the last synced timestamp
	UpdateLastSynced(ctx context.Context, id uuid.UUID) error
}

// AdAccountRepository defines the interface for ad account data persistence
type AdAccountRepository interface {
	// Create creates a new ad account
	Create(ctx context.Context, account *entity.AdAccount) error

	// GetByID retrieves an ad account by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AdAccount, error)

	// GetByPlatformID retrieves by platform-specific ad account ID
	GetByPlatformID(ctx context.Context, connectedAccountID uuid.UUID, platformAdAccountID string) (*entity.AdAccount, error)

	// Update updates an ad account
	Update(ctx context.Context, account *entity.AdAccount) error

	// Delete deletes an ad account
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByConnectedAccount lists all ad accounts for a connected account
	ListByConnectedAccount(ctx context.Context, connectedAccountID uuid.UUID) ([]entity.AdAccount, error)

	// ListByOrganization lists all ad accounts for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.AdAccount, error)

	// Upsert creates or updates an ad account
	Upsert(ctx context.Context, account *entity.AdAccount) error
}

// CampaignRepository defines the interface for campaign data persistence
type CampaignRepository interface {
	// Create creates a new campaign
	Create(ctx context.Context, campaign *entity.Campaign) error

	// GetByID retrieves a campaign by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Campaign, error)

	// GetByPlatformID retrieves by platform-specific campaign ID
	GetByPlatformID(ctx context.Context, adAccountID uuid.UUID, platformCampaignID string) (*entity.Campaign, error)

	// Update updates a campaign
	Update(ctx context.Context, campaign *entity.Campaign) error

	// Delete deletes a campaign
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists campaigns with filters
	List(ctx context.Context, filter entity.CampaignFilter) ([]entity.Campaign, int64, error)

	// ListByAdAccount lists all campaigns for an ad account
	ListByAdAccount(ctx context.Context, adAccountID uuid.UUID) ([]entity.Campaign, error)

	// ListByOrganization lists all campaigns for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.Campaign, error)

	// Upsert creates or updates a campaign
	Upsert(ctx context.Context, campaign *entity.Campaign) error

	// BulkUpsert creates or updates multiple campaigns
	BulkUpsert(ctx context.Context, campaigns []entity.Campaign) error

	// GetSummaries retrieves campaign summaries with metrics
	GetSummaries(ctx context.Context, filter entity.CampaignFilter) ([]entity.CampaignSummary, error)

	// UpdateLastSynced updates the last synced timestamp
	UpdateLastSynced(ctx context.Context, id uuid.UUID) error
}

// AdSetRepository defines the interface for ad set data persistence
type AdSetRepository interface {
	// Create creates a new ad set
	Create(ctx context.Context, adSet *entity.AdSet) error

	// GetByID retrieves an ad set by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AdSet, error)

	// GetByPlatformID retrieves by platform-specific ad set ID
	GetByPlatformID(ctx context.Context, campaignID uuid.UUID, platformAdSetID string) (*entity.AdSet, error)

	// Update updates an ad set
	Update(ctx context.Context, adSet *entity.AdSet) error

	// Delete deletes an ad set
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists ad sets with filters
	List(ctx context.Context, filter entity.AdSetFilter) ([]entity.AdSet, int64, error)

	// ListByCampaign lists all ad sets for a campaign
	ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]entity.AdSet, error)

	// Upsert creates or updates an ad set
	Upsert(ctx context.Context, adSet *entity.AdSet) error

	// BulkUpsert creates or updates multiple ad sets
	BulkUpsert(ctx context.Context, adSets []entity.AdSet) error
}

// AdRepository defines the interface for ad data persistence
type AdRepository interface {
	// Create creates a new ad
	Create(ctx context.Context, ad *entity.Ad) error

	// GetByID retrieves an ad by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Ad, error)

	// GetByPlatformID retrieves by platform-specific ad ID
	GetByPlatformID(ctx context.Context, adSetID uuid.UUID, platformAdID string) (*entity.Ad, error)

	// Update updates an ad
	Update(ctx context.Context, ad *entity.Ad) error

	// Delete deletes an ad
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists ads with filters
	List(ctx context.Context, filter entity.AdFilter) ([]entity.Ad, int64, error)

	// ListByAdSet lists all ads for an ad set
	ListByAdSet(ctx context.Context, adSetID uuid.UUID) ([]entity.Ad, error)

	// Upsert creates or updates an ad
	Upsert(ctx context.Context, ad *entity.Ad) error

	// BulkUpsert creates or updates multiple ads
	BulkUpsert(ctx context.Context, ads []entity.Ad) error
}

// MetricsRepository defines the interface for metrics data persistence
type MetricsRepository interface {
	// Campaign metrics
	CreateCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error
	GetCampaignMetrics(ctx context.Context, campaignID uuid.UUID, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error)
	UpsertCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error
	BulkUpsertCampaignMetrics(ctx context.Context, metrics []entity.CampaignMetricsDaily) error

	// Ad set metrics
	CreateAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error
	GetAdSetMetrics(ctx context.Context, adSetID uuid.UUID, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error)
	UpsertAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error
	BulkUpsertAdSetMetrics(ctx context.Context, metrics []entity.AdSetMetricsDaily) error

	// Ad metrics
	CreateAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error
	GetAdMetrics(ctx context.Context, adID uuid.UUID, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error)
	UpsertAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error
	BulkUpsertAdMetrics(ctx context.Context, metrics []entity.AdMetricsDaily) error

	// Aggregated metrics
	GetAggregatedMetrics(ctx context.Context, filter entity.MetricsFilter) (*entity.AggregatedMetrics, error)
	GetMetricsByPlatform(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) ([]entity.PlatformMetricsSummary, error)
	GetDailyTrend(ctx context.Context, filter entity.MetricsFilter) ([]entity.DailyMetricsTrend, error)
	GetTopPerformingCampaigns(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange, limit int) ([]entity.TopPerformer, error)
}

// TokenRefreshLogRepository defines the interface for token refresh log persistence
type TokenRefreshLogRepository interface {
	// Create creates a new token refresh log entry
	Create(ctx context.Context, log *entity.TokenRefreshLog) error

	// GetByAccountID retrieves logs for a connected account
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]entity.TokenRefreshLog, error)

	// DeleteOld deletes logs older than the specified days
	DeleteOld(ctx context.Context, olderThanDays int) error
}

// SyncStateRepository defines the interface for sync state data persistence
type SyncStateRepository interface {
	// Create creates a new sync state
	Create(ctx context.Context, state *entity.SyncState) error

	// GetByID retrieves a sync state by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SyncState, error)

	// GetByConnectedAccount retrieves sync state by connected account ID
	GetByConnectedAccount(ctx context.Context, connectedAccountID uuid.UUID) (*entity.SyncState, error)

	// Update updates a sync state
	Update(ctx context.Context, state *entity.SyncState) error

	// ListByOrganization lists all sync states for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.SyncState, error)

	// ListDueForHourlySync lists accounts due for hourly sync
	ListDueForHourlySync(ctx context.Context, platform entity.Platform) ([]entity.SyncState, error)

	// ListDueForDailySync lists accounts due for daily sync
	ListDueForDailySync(ctx context.Context, platform entity.Platform, hour int) ([]entity.SyncState, error)

	// UpdateSyncStarted marks a sync as started
	UpdateSyncStarted(ctx context.Context, id uuid.UUID, syncJobID uuid.UUID) error

	// UpdateSyncCompleted marks a sync as completed
	UpdateSyncCompleted(ctx context.Context, id uuid.UUID, syncType entity.SyncType) error

	// UpdateSyncFailed marks a sync as failed
	UpdateSyncFailed(ctx context.Context, id uuid.UUID, err string) error

	// UpdateRateLimit updates rate limit information
	UpdateRateLimit(ctx context.Context, id uuid.UUID, resetAt *time.Time) error

	// GetDataFreshness retrieves data freshness info for all accounts in org
	GetDataFreshness(ctx context.Context, orgID uuid.UUID) ([]entity.DataFreshnessInfo, error)
}

// SyncJobRepository defines the interface for sync job data persistence
type SyncJobRepository interface {
	// Create creates a new sync job
	Create(ctx context.Context, job *entity.SyncJob) error

	// GetByID retrieves a sync job by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SyncJob, error)

	// Update updates a sync job
	Update(ctx context.Context, job *entity.SyncJob) error

	// ListPending lists pending jobs ready to run
	ListPending(ctx context.Context, limit int) ([]entity.SyncJob, error)

	// ListByConnectedAccount lists jobs for a connected account
	ListByConnectedAccount(ctx context.Context, connectedAccountID uuid.UUID, limit int) ([]entity.SyncJob, error)

	// ListRecent lists recent jobs (for dashboard)
	ListRecent(ctx context.Context, orgID uuid.UUID, limit int) ([]entity.SyncJob, error)

	// GetRunningJobs gets currently running jobs
	GetRunningJobs(ctx context.Context) ([]entity.SyncJob, error)

	// ClaimJob atomically claims a pending job for processing
	ClaimJob(ctx context.Context, jobID uuid.UUID, workerID string) error

	// UpdateProgress updates job progress
	UpdateProgress(ctx context.Context, id uuid.UUID, percent int, message string) error

	// MarkCompleted marks job as completed
	MarkCompleted(ctx context.Context, id uuid.UUID, recordsProcessed, recordsFailed int) error

	// MarkFailed marks job as failed
	MarkFailed(ctx context.Context, id uuid.UUID, err, errorCode string) error

	// ScheduleRetry schedules a retry for a failed job
	ScheduleRetry(ctx context.Context, id uuid.UUID, retryAfter time.Time, lastError string) error

	// DeleteOldJobs deletes jobs older than the specified days
	DeleteOldJobs(ctx context.Context, olderThanDays int) error

	// GetJobStats gets sync job statistics
	GetJobStats(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) (*SyncJobStats, error)
}

// SyncJobStats represents sync job statistics
type SyncJobStats struct {
	TotalJobs       int64 `json:"total_jobs"`
	PendingJobs     int64 `json:"pending_jobs"`
	RunningJobs     int64 `json:"running_jobs"`
	CompletedJobs   int64 `json:"completed_jobs"`
	FailedJobs      int64 `json:"failed_jobs"`
	AverageDuration int   `json:"average_duration_seconds"`
}

// RetryQueueRepository defines the interface for retry queue persistence
type RetryQueueRepository interface {
	// Enqueue adds an entry to the retry queue
	Enqueue(ctx context.Context, entry *entity.RetryQueueEntry) error

	// Dequeue gets entries ready for retry
	Dequeue(ctx context.Context, limit int) ([]entity.RetryQueueEntry, error)

	// Remove removes an entry from the queue
	Remove(ctx context.Context, id uuid.UUID) error

	// GetByJobID gets retry entry by job ID
	GetByJobID(ctx context.Context, jobID uuid.UUID) (*entity.RetryQueueEntry, error)

	// UpdateRetryCount updates retry count
	UpdateRetryCount(ctx context.Context, id uuid.UUID, count int, nextRetryAt time.Time, lastError string) error

	// CleanupExpired removes entries that exceeded max retries
	CleanupExpired(ctx context.Context) (int, error)
}

// WebhookEventRepository defines the interface for webhook event persistence
type WebhookEventRepository interface {
	// Create creates a new webhook event
	Create(ctx context.Context, event *entity.WebhookEvent) error

	// GetByID retrieves a webhook event by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WebhookEvent, error)

	// Update updates a webhook event
	Update(ctx context.Context, event *entity.WebhookEvent) error

	// ListPendingEvents lists unprocessed webhook events
	ListPendingEvents(ctx context.Context, limit int) ([]entity.WebhookEvent, error)

	// MarkProcessed marks an event as processed
	MarkProcessed(ctx context.Context, id uuid.UUID) error

	// MarkFailed marks an event as failed
	MarkFailed(ctx context.Context, id uuid.UUID, err string) error

	// DeleteOldEvents deletes events older than the specified days
	DeleteOldEvents(ctx context.Context, olderThanDays int) error
}

// ManualSyncRateLimitRepository defines the interface for manual sync rate limit tracking
type ManualSyncRateLimitRepository interface {
	// GetOrCreate gets or creates a rate limit entry for a user
	GetOrCreate(ctx context.Context, userID, orgID uuid.UUID) (*entity.ManualSyncRateLimit, error)

	// IncrementCount increments the sync count for current hour
	IncrementCount(ctx context.Context, id uuid.UUID) error

	// ResetForNewHour resets count for a new hour window
	ResetForNewHour(ctx context.Context, id uuid.UUID) error

	// CleanupOldRecords removes old rate limit records
	CleanupOldRecords(ctx context.Context) (int, error)
}

// SyncErrorLogRepository defines the interface for sync error log persistence
type SyncErrorLogRepository interface {
	// Create creates a new error log entry
	Create(ctx context.Context, log *entity.SyncErrorLog) error

	// ListByJob lists errors for a job
	ListByJob(ctx context.Context, jobID uuid.UUID) ([]entity.SyncErrorLog, error)

	// ListByAccount lists errors for a connected account
	ListByAccount(ctx context.Context, connectedAccountID uuid.UUID, limit int) ([]entity.SyncErrorLog, error)

	// ListRecent lists recent errors across all accounts
	ListRecent(ctx context.Context, orgID uuid.UUID, limit int) ([]entity.SyncErrorLog, error)

	// DeleteOld deletes logs older than the specified days
	DeleteOld(ctx context.Context, olderThanDays int) error

	// GetErrorStats gets error statistics
	GetErrorStats(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) (map[string]int, error)
}

// SubscriptionRepository defines the interface for subscription data persistence
type SubscriptionRepository interface {
	// Create creates a new subscription
	Create(ctx context.Context, sub *entity.Subscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Subscription, error)

	// GetByOrganization retrieves subscription by organization ID
	GetByOrganization(ctx context.Context, orgID uuid.UUID) (*entity.Subscription, error)

	// GetByStripeCustomerID retrieves by Stripe customer ID
	GetByStripeCustomerID(ctx context.Context, customerID string) (*entity.Subscription, error)

	// GetByStripeSubscriptionID retrieves by Stripe subscription ID
	GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*entity.Subscription, error)

	// Update updates a subscription
	Update(ctx context.Context, sub *entity.Subscription) error

	// UpdateStatus updates subscription status
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.SubscriptionStatus) error

	// UpdatePlan updates the plan tier
	UpdatePlan(ctx context.Context, id uuid.UUID, tier entity.PlanTier, cycle entity.BillingCycle) error

	// RecordPayment records a successful payment
	RecordPayment(ctx context.Context, id uuid.UUID, amount float64, paidAt time.Time) error

	// IncrementPaymentFails increments payment failure count
	IncrementPaymentFails(ctx context.Context, id uuid.UUID) error

	// ResetPaymentFails resets payment failure count
	ResetPaymentFails(ctx context.Context, id uuid.UUID) error

	// ListExpiring lists subscriptions expiring within given days
	ListExpiring(ctx context.Context, withinDays int) ([]entity.Subscription, error)

	// ListPastDue lists subscriptions with past due payments
	ListPastDue(ctx context.Context) ([]entity.Subscription, error)
}

// PaymentHistoryRepository defines the interface for payment history persistence
type PaymentHistoryRepository interface {
	// Create creates a new payment record
	Create(ctx context.Context, payment *entity.PaymentHistory) error

	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PaymentHistory, error)

	// GetByStripePaymentIntent retrieves by Stripe payment intent ID
	GetByStripePaymentIntent(ctx context.Context, intentID string) (*entity.PaymentHistory, error)

	// ListByOrganization lists payments for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit int) ([]entity.PaymentHistory, error)

	// ListBySubscription lists payments for a subscription
	ListBySubscription(ctx context.Context, subID uuid.UUID) ([]entity.PaymentHistory, error)

	// UpdateStatus updates payment status
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus) error

	// GetTotalRevenue gets total revenue for a period
	GetTotalRevenue(ctx context.Context, startDate, endDate time.Time) (float64, error)
}

// UsageRepository defines the interface for usage tracking persistence
type UsageRepository interface {
	// GetOrCreateDaily gets or creates daily usage record
	GetOrCreateDaily(ctx context.Context, orgID uuid.UUID, date time.Time) (*entity.OrganizationUsage, error)

	// GetByDate retrieves usage for a specific date
	GetByDate(ctx context.Context, orgID uuid.UUID, date time.Time) (*entity.OrganizationUsage, error)

	// IncrementAPICallCount increments API call counter
	IncrementAPICallCount(ctx context.Context, orgID uuid.UUID) error

	// IncrementSyncCount increments sync counter
	IncrementSyncCount(ctx context.Context, orgID uuid.UUID, recordsSynced int64) error

	// IncrementReportCount increments report counter
	IncrementReportCount(ctx context.Context, orgID uuid.UUID) error

	// UpdateStorageUsage updates storage usage in bytes
	UpdateStorageUsage(ctx context.Context, orgID uuid.UUID, bytes int64) error

	// UpdateAccountsCount updates connected accounts count
	UpdateAccountsCount(ctx context.Context, orgID uuid.UUID, count int) error

	// UpdateUsersCount updates active users count
	UpdateUsersCount(ctx context.Context, orgID uuid.UUID, count int) error

	// UpdateLimits updates usage limits based on plan
	UpdateLimits(ctx context.Context, orgID uuid.UUID, limits entity.PlanLimits) error

	// GetSummary gets usage summary for a period
	GetSummary(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*entity.UsageSummary, error)

	// GetDailyUsage gets daily usage for a date range
	GetDailyUsage(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]entity.OrganizationUsage, error)

	// CheckAPILimit checks if API call limit is exceeded
	CheckAPILimit(ctx context.Context, orgID uuid.UUID) (bool, int64, int64, error) // exceeded, current, limit
}

// UsageEventRepository defines the interface for usage event logging
type UsageEventRepository interface {
	// Create creates a new usage event
	Create(ctx context.Context, event *entity.UsageEvent) error

	// ListByOrganization lists events for an organization
	ListByOrganization(ctx context.Context, orgID uuid.UUID, eventType entity.UsageType, limit int) ([]entity.UsageEvent, error)

	// GetEventCounts gets event counts by type for a period
	GetEventCounts(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (map[entity.UsageType]int64, error)

	// DeleteOld deletes events older than specified days
	DeleteOld(ctx context.Context, olderThanDays int) error
}

// QuotaAlertRepository defines the interface for quota alerts
type QuotaAlertRepository interface {
	// Create creates a new quota alert
	Create(ctx context.Context, alert *entity.QuotaAlert) error

	// GetUnacknowledged gets unacknowledged alerts for an organization
	GetUnacknowledged(ctx context.Context, orgID uuid.UUID) ([]entity.QuotaAlert, error)

	// Acknowledge acknowledges an alert
	Acknowledge(ctx context.Context, alertID uuid.UUID, userID uuid.UUID) error

	// GetRecent gets recent alerts
	GetRecent(ctx context.Context, orgID uuid.UUID, limit int) ([]entity.QuotaAlert, error)
}

// CouponRepository defines the interface for coupon management
type CouponRepository interface {
	// Create creates a new coupon
	Create(ctx context.Context, coupon *entity.Coupon) error

	// GetByCode retrieves a coupon by code
	GetByCode(ctx context.Context, code string) (*entity.Coupon, error)

	// GetByStripeCouponID retrieves by Stripe coupon ID
	GetByStripeCouponID(ctx context.Context, stripeCouponID string) (*entity.Coupon, error)

	// IncrementRedemptions increments redemption count
	IncrementRedemptions(ctx context.Context, id uuid.UUID) error

	// Update updates a coupon
	Update(ctx context.Context, coupon *entity.Coupon) error

	// Deactivate deactivates a coupon
	Deactivate(ctx context.Context, id uuid.UUID) error

	// ListActive lists active coupons
	ListActive(ctx context.Context) ([]entity.Coupon, error)
}

// UnitOfWork defines a transactional unit of work pattern
type UnitOfWork interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) (UnitOfWork, error)

	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Users returns the user repository
	Users() UserRepository

	// Organizations returns the organization repository
	Organizations() OrganizationRepository

	// OrganizationMembers returns the organization member repository
	OrganizationMembers() OrganizationMemberRepository

	// ConnectedAccounts returns the connected account repository
	ConnectedAccounts() ConnectedAccountRepository

	// AdAccounts returns the ad account repository
	AdAccounts() AdAccountRepository

	// Campaigns returns the campaign repository
	Campaigns() CampaignRepository

	// AdSets returns the ad set repository
	AdSets() AdSetRepository

	// Ads returns the ad repository
	Ads() AdRepository

	// Metrics returns the metrics repository
	Metrics() MetricsRepository
}
