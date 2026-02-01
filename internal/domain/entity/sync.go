package entity

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Sync Status Types
// ============================================================================

// SyncStatus represents the status of a sync operation
type SyncStatus string

const (
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusRunning    SyncStatus = "running"
	SyncStatusCompleted  SyncStatus = "completed"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusCancelled  SyncStatus = "cancelled"
	SyncStatusRetrying   SyncStatus = "retrying"
	SyncStatusRateLimited SyncStatus = "rate_limited"
)

// SyncType represents the type of sync operation
type SyncType string

const (
	SyncTypeHourly     SyncType = "hourly"      // Active campaigns only
	SyncTypeDaily      SyncType = "daily"       // Full historical sync
	SyncTypeManual     SyncType = "manual"      // User-triggered sync
	SyncTypeWebhook    SyncType = "webhook"     // Webhook-triggered update
	SyncTypeInitial    SyncType = "initial"     // First-time full sync
	SyncTypeIncremental SyncType = "incremental" // Delta changes only
)

// SyncScope represents what to sync
type SyncScope string

const (
	SyncScopeAccount   SyncScope = "account"   // Sync entire account
	SyncScopeCampaign  SyncScope = "campaign"  // Sync specific campaign
	SyncScopeMetrics   SyncScope = "metrics"   // Sync metrics only
	SyncScopeStructure SyncScope = "structure" // Sync campaign structure only
)

// ============================================================================
// Sync State Entity
// ============================================================================

// SyncState tracks the sync state for each connected account
type SyncState struct {
	BaseEntity
	ConnectedAccountID uuid.UUID  `json:"connected_account_id" gorm:"type:uuid;not null;uniqueIndex:idx_sync_state_account"`
	OrganizationID     uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index"`
	Platform           Platform   `json:"platform" gorm:"type:platform_type;not null;index"`

	// Last successful sync timestamps
	LastHourlySync   *time.Time `json:"last_hourly_sync,omitempty"`
	LastDailySync    *time.Time `json:"last_daily_sync,omitempty"`
	LastFullSync     *time.Time `json:"last_full_sync,omitempty"`
	LastMetricsSync  *time.Time `json:"last_metrics_sync,omitempty"`

	// Current sync status
	CurrentSyncID     *uuid.UUID `json:"current_sync_id,omitempty" gorm:"type:uuid"`
	CurrentSyncStatus SyncStatus `json:"current_sync_status" gorm:"type:sync_status;default:'pending'"`
	IsSyncing         bool       `json:"is_syncing" gorm:"default:false"`

	// Data freshness indicator (minutes since last update)
	DataFreshnessMinutes int `json:"data_freshness_minutes" gorm:"-"` // Calculated field

	// Sync statistics
	TotalSyncs         int64 `json:"total_syncs" gorm:"default:0"`
	SuccessfulSyncs    int64 `json:"successful_syncs" gorm:"default:0"`
	FailedSyncs        int64 `json:"failed_syncs" gorm:"default:0"`
	ConsecutiveFailures int  `json:"consecutive_failures" gorm:"default:0"`

	// Rate limit tracking
	RateLimitResetAt *time.Time `json:"rate_limit_reset_at,omitempty"`
	RateLimitHits    int        `json:"rate_limit_hits" gorm:"default:0"`

	// Next scheduled sync
	NextScheduledSync *time.Time `json:"next_scheduled_sync,omitempty"`

	// Error tracking
	LastError      string     `json:"last_error,omitempty"`
	LastErrorAt    *time.Time `json:"last_error_at,omitempty"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// CalculateDataFreshness calculates minutes since last successful sync
func (s *SyncState) CalculateDataFreshness() int {
	var lastSync time.Time

	// Use the most recent sync
	if s.LastHourlySync != nil {
		lastSync = *s.LastHourlySync
	}
	if s.LastMetricsSync != nil && s.LastMetricsSync.After(lastSync) {
		lastSync = *s.LastMetricsSync
	}

	if lastSync.IsZero() {
		return -1 // Never synced
	}

	return int(time.Since(lastSync).Minutes())
}

// GetFreshnessStatus returns a human-readable freshness status
func (s *SyncState) GetFreshnessStatus() string {
	minutes := s.CalculateDataFreshness()
	switch {
	case minutes < 0:
		return "never_synced"
	case minutes <= 60:
		return "fresh"
	case minutes <= 240: // 4 hours
		return "recent"
	case minutes <= 1440: // 24 hours
		return "stale"
	default:
		return "outdated"
	}
}

// IsRateLimited checks if currently rate limited
func (s *SyncState) IsRateLimited() bool {
	if s.RateLimitResetAt == nil {
		return false
	}
	return time.Now().Before(*s.RateLimitResetAt)
}

// CanSync checks if sync is allowed now
func (s *SyncState) CanSync() bool {
	if s.IsSyncing {
		return false
	}
	if s.IsRateLimited() {
		return false
	}
	return true
}

// ============================================================================
// Sync Job Entity
// ============================================================================

// SyncJob represents a single sync operation
type SyncJob struct {
	BaseEntity
	OrganizationID     uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null;index"`
	ConnectedAccountID uuid.UUID  `json:"connected_account_id" gorm:"type:uuid;not null;index"`
	Platform           Platform   `json:"platform" gorm:"type:platform_type;not null"`

	// Job configuration
	SyncType   SyncType  `json:"sync_type" gorm:"type:sync_type;not null"`
	SyncScope  SyncScope `json:"sync_scope" gorm:"type:sync_scope;not null"`
	Status     SyncStatus `json:"status" gorm:"type:sync_status;default:'pending';index"`
	Priority   int       `json:"priority" gorm:"default:0"` // Higher = more priority

	// Target resources (optional - for specific campaign sync)
	CampaignID *uuid.UUID `json:"campaign_id,omitempty" gorm:"type:uuid"`

	// Date range for metrics sync
	DateRangeStart *time.Time `json:"date_range_start,omitempty"`
	DateRangeEnd   *time.Time `json:"date_range_end,omitempty"`

	// Execution tracking
	ScheduledAt     time.Time  `json:"scheduled_at" gorm:"not null;index"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`

	// Retry management
	MaxRetries     int        `json:"max_retries" gorm:"default:3"`
	RetryCount     int        `json:"retry_count" gorm:"default:0"`
	RetryAfter     *time.Time `json:"retry_after,omitempty"`
	LastRetryError string     `json:"last_retry_error,omitempty"`

	// Progress tracking
	ProgressPercent int    `json:"progress_percent" gorm:"default:0"`
	ProgressMessage string `json:"progress_message,omitempty"`

	// Results
	RecordsProcessed int    `json:"records_processed" gorm:"default:0"`
	RecordsFailed    int    `json:"records_failed" gorm:"default:0"`
	ErrorMessage     string `json:"error_message,omitempty"`
	ErrorCode        string `json:"error_code,omitempty"`

	// Triggered by
	TriggeredBy      string     `json:"triggered_by,omitempty"` // "scheduler", "webhook", "user:{id}"
	TriggeredByUser  *uuid.UUID `json:"triggered_by_user,omitempty" gorm:"type:uuid"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// CanRetry checks if job can be retried
func (j *SyncJob) CanRetry() bool {
	return j.RetryCount < j.MaxRetries
}

// IncrementRetry increments retry count and sets retry after
func (j *SyncJob) IncrementRetry(backoffSeconds int) {
	j.RetryCount++
	retryAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
	j.RetryAfter = &retryAt
	j.Status = SyncStatusRetrying
}

// MarkCompleted marks job as completed
func (j *SyncJob) MarkCompleted() {
	now := time.Now()
	j.Status = SyncStatusCompleted
	j.CompletedAt = &now
	j.ProgressPercent = 100
	if j.StartedAt != nil {
		duration := int(now.Sub(*j.StartedAt).Seconds())
		j.DurationSeconds = &duration
	}
}

// MarkFailed marks job as failed
func (j *SyncJob) MarkFailed(err error, code string) {
	now := time.Now()
	j.Status = SyncStatusFailed
	j.CompletedAt = &now
	j.ErrorMessage = err.Error()
	j.ErrorCode = code
	if j.StartedAt != nil {
		duration := int(now.Sub(*j.StartedAt).Seconds())
		j.DurationSeconds = &duration
	}
}

// ============================================================================
// Sync Error Log
// ============================================================================

// SyncErrorLog represents a detailed error log for sync operations
type SyncErrorLog struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SyncJobID          uuid.UUID `json:"sync_job_id" gorm:"type:uuid;not null;index"`
	ConnectedAccountID uuid.UUID `json:"connected_account_id" gorm:"type:uuid;not null;index"`
	Platform           Platform  `json:"platform" gorm:"type:platform_type;not null"`

	ErrorType    string    `json:"error_type" gorm:"size:50;not null"` // "api_error", "rate_limit", "auth", "validation", "network"
	ErrorCode    string    `json:"error_code,omitempty" gorm:"size:50"`
	ErrorMessage string    `json:"error_message" gorm:"type:text;not null"`
	StackTrace   string    `json:"stack_trace,omitempty" gorm:"type:text"`

	// Context
	RequestPath   string  `json:"request_path,omitempty" gorm:"size:500"`
	RequestMethod string  `json:"request_method,omitempty" gorm:"size:10"`
	ResponseCode  int     `json:"response_code,omitempty"`

	// Recovery info
	IsRetryable   bool   `json:"is_retryable" gorm:"default:false"`
	RetryAfter    *time.Time `json:"retry_after,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ============================================================================
// Retry Queue Entry
// ============================================================================

// RetryQueueEntry represents an entry in the retry queue
type RetryQueueEntry struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SyncJobID          uuid.UUID  `json:"sync_job_id" gorm:"type:uuid;not null;index"`
	ConnectedAccountID uuid.UUID  `json:"connected_account_id" gorm:"type:uuid;not null;index"`
	Platform           Platform   `json:"platform" gorm:"type:platform_type;not null"`

	RetryAt     time.Time `json:"retry_at" gorm:"not null;index"`
	RetryCount  int       `json:"retry_count" gorm:"default:0"`
	MaxRetries  int       `json:"max_retries" gorm:"default:3"`
	LastError   string    `json:"last_error,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ============================================================================
// Manual Sync Rate Limit
// ============================================================================

// ManualSyncRateLimit tracks manual sync rate limits per user
type ManualSyncRateLimit struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_manual_sync_user_hour"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`

	HourWindow     time.Time `json:"hour_window" gorm:"not null;uniqueIndex:idx_manual_sync_user_hour"` // Truncated to hour
	SyncCount      int       `json:"sync_count" gorm:"default:0"`
	LastSyncAt     time.Time `json:"last_sync_at"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const MaxManualSyncsPerHour = 5

// CanTriggerManualSync checks if user can trigger another manual sync
func (r *ManualSyncRateLimit) CanTriggerManualSync() bool {
	// Check if the hour window has passed
	currentHour := time.Now().Truncate(time.Hour)
	if r.HourWindow.Before(currentHour) {
		return true // New hour window
	}
	return r.SyncCount < MaxManualSyncsPerHour
}

// RemainingManualSyncs returns remaining manual syncs for the hour
func (r *ManualSyncRateLimit) RemainingManualSyncs() int {
	currentHour := time.Now().Truncate(time.Hour)
	if r.HourWindow.Before(currentHour) {
		return MaxManualSyncsPerHour
	}
	remaining := MaxManualSyncsPerHour - r.SyncCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ============================================================================
// Webhook Event
// ============================================================================

// WebhookEvent represents an incoming webhook event
type WebhookEvent struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Platform         Platform  `json:"platform" gorm:"type:platform_type;not null;index"`
	EventType        string    `json:"event_type" gorm:"size:100;not null;index"`

	// Signature verification
	ReceivedAt       time.Time `json:"received_at" gorm:"not null"`
	Signature        string    `json:"signature,omitempty" gorm:"size:500"`
	SignatureValid   bool      `json:"signature_valid" gorm:"default:false"`

	// Payload
	RawPayload       string    `json:"raw_payload" gorm:"type:text"`
	ParsedPayload    JSONMap   `json:"parsed_payload,omitempty" gorm:"type:jsonb"`

	// Processing status
	ProcessingStatus string     `json:"processing_status" gorm:"size:20;default:'pending'"` // "pending", "processed", "failed", "ignored"
	ProcessedAt      *time.Time `json:"processed_at,omitempty"`
	ProcessingError  string     `json:"processing_error,omitempty"`

	// Related entities
	ConnectedAccountID *uuid.UUID `json:"connected_account_id,omitempty" gorm:"type:uuid"`
	CampaignID         *uuid.UUID `json:"campaign_id,omitempty" gorm:"type:uuid"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ============================================================================
// Data Freshness Response
// ============================================================================

// DataFreshnessInfo represents data freshness information for dashboard
type DataFreshnessInfo struct {
	Platform              Platform   `json:"platform"`
	ConnectedAccountID    uuid.UUID  `json:"connected_account_id"`
	ConnectedAccountName  string     `json:"connected_account_name"`

	// Freshness status
	FreshnessStatus       string     `json:"freshness_status"` // "fresh", "recent", "stale", "outdated", "never_synced"
	MinutesSinceLastSync  int        `json:"minutes_since_last_sync"`
	LastSyncedAt          *time.Time `json:"last_synced_at,omitempty"`

	// Current sync status
	IsSyncing             bool       `json:"is_syncing"`
	CurrentSyncProgress   int        `json:"current_sync_progress,omitempty"`

	// Next sync
	NextScheduledSync     *time.Time `json:"next_scheduled_sync,omitempty"`

	// Health status
	IsHealthy             bool       `json:"is_healthy"`
	ConsecutiveFailures   int        `json:"consecutive_failures"`
	LastError             string     `json:"last_error,omitempty"`

	// Rate limit
	IsRateLimited         bool       `json:"is_rate_limited"`
	RateLimitResetAt      *time.Time `json:"rate_limit_reset_at,omitempty"`
}

// ============================================================================
// Scheduler Configuration
// ============================================================================

// SyncScheduleConfig represents sync schedule configuration
type SyncScheduleConfig struct {
	Platform Platform `json:"platform"`

	// Hourly sync settings
	HourlySyncEnabled     bool `json:"hourly_sync_enabled"`
	HourlySyncMinuteOffset int `json:"hourly_sync_minute_offset"` // 0-59, when in the hour to run

	// Daily sync settings
	DailySyncEnabled     bool `json:"daily_sync_enabled"`
	DailySyncHour        int  `json:"daily_sync_hour"`        // 0-23, hour to run daily sync
	DailySyncLookbackDays int `json:"daily_sync_lookback_days"` // How many days back to sync

	// Rate limit settings (per minute)
	MaxRequestsPerMinute int `json:"max_requests_per_minute"`
	RequestBurstSize     int `json:"request_burst_size"`

	// Retry settings
	MaxRetries          int `json:"max_retries"`
	RetryBackoffSeconds int `json:"retry_backoff_seconds"`
	MaxRetryBackoff     int `json:"max_retry_backoff_seconds"`
}

// DefaultSyncScheduleConfigs returns default sync configurations per platform
func DefaultSyncScheduleConfigs() map[Platform]SyncScheduleConfig {
	return map[Platform]SyncScheduleConfig{
		PlatformMeta: {
			Platform:              PlatformMeta,
			HourlySyncEnabled:     true,
			HourlySyncMinuteOffset: 5,
			DailySyncEnabled:      true,
			DailySyncHour:         3, // 3 AM
			DailySyncLookbackDays: 7,
			MaxRequestsPerMinute:  200, // Meta allows ~200/min for insights
			RequestBurstSize:      50,
			MaxRetries:            3,
			RetryBackoffSeconds:   60,
			MaxRetryBackoff:       900, // 15 min max
		},
		PlatformTikTok: {
			Platform:              PlatformTikTok,
			HourlySyncEnabled:     true,
			HourlySyncMinuteOffset: 10,
			DailySyncEnabled:      true,
			DailySyncHour:         3,
			DailySyncLookbackDays: 7,
			MaxRequestsPerMinute:  10, // TikTok has strict limits
			RequestBurstSize:      5,
			MaxRetries:            3,
			RetryBackoffSeconds:   120, // TikTok needs longer backoff
			MaxRetryBackoff:       1800, // 30 min max
		},
		PlatformShopee: {
			Platform:              PlatformShopee,
			HourlySyncEnabled:     true,
			HourlySyncMinuteOffset: 15,
			DailySyncEnabled:      true,
			DailySyncHour:         3,
			DailySyncLookbackDays: 7,
			MaxRequestsPerMinute:  60, // Shopee moderate limits
			RequestBurstSize:      20,
			MaxRetries:            3,
			RetryBackoffSeconds:   60,
			MaxRetryBackoff:       900,
		},
	}
}
