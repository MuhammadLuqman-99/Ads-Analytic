package entity

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Usage Tracking
// ============================================================================

// UsageType represents the type of usage being tracked
type UsageType string

const (
	UsageTypeAPICall       UsageType = "api_call"
	UsageTypeDataSync      UsageType = "data_sync"
	UsageTypeReportGenerate UsageType = "report_generate"
	UsageTypeWebhookReceive UsageType = "webhook_receive"
	UsageTypeExport        UsageType = "export"
)

// OrganizationUsage tracks usage per organization
type OrganizationUsage struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_org_usage_date"`
	Date           time.Time `json:"date" gorm:"type:date;not null;uniqueIndex:idx_org_usage_date"`

	// API Usage
	APICallsCount      int64 `json:"api_calls_count" gorm:"default:0"`
	APICallsLimit      int64 `json:"api_calls_limit" gorm:"default:100"`

	// Data Sync
	DataSyncCount      int64 `json:"data_sync_count" gorm:"default:0"`
	RecordsSynced      int64 `json:"records_synced" gorm:"default:0"`

	// Reports
	ReportsGenerated   int64 `json:"reports_generated" gorm:"default:0"`
	ExportsCount       int64 `json:"exports_count" gorm:"default:0"`

	// Webhooks
	WebhooksReceived   int64 `json:"webhooks_received" gorm:"default:0"`

	// Storage (in bytes)
	StorageUsedBytes   int64 `json:"storage_used_bytes" gorm:"default:0"`
	StorageLimitBytes  int64 `json:"storage_limit_bytes" gorm:"default:104857600"` // 100MB default

	// Connected accounts
	ConnectedAccounts  int   `json:"connected_accounts" gorm:"default:0"`
	AccountsLimit      int   `json:"accounts_limit" gorm:"default:1"`

	// Active users
	ActiveUsersCount   int   `json:"active_users_count" gorm:"default:0"`
	UsersLimit         int   `json:"users_limit" gorm:"default:1"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// IsAPILimitExceeded checks if API call limit is exceeded
func (u *OrganizationUsage) IsAPILimitExceeded() bool {
	if u.APICallsLimit <= 0 { // -1 or 0 means unlimited
		return false
	}
	return u.APICallsCount >= u.APICallsLimit
}

// IsStorageLimitExceeded checks if storage limit is exceeded
func (u *OrganizationUsage) IsStorageLimitExceeded() bool {
	if u.StorageLimitBytes <= 0 {
		return false
	}
	return u.StorageUsedBytes >= u.StorageLimitBytes
}

// IsAccountsLimitExceeded checks if connected accounts limit is exceeded
func (u *OrganizationUsage) IsAccountsLimitExceeded() bool {
	if u.AccountsLimit <= 0 {
		return false
	}
	return u.ConnectedAccounts >= u.AccountsLimit
}

// GetAPIUsagePercent returns API usage as percentage
func (u *OrganizationUsage) GetAPIUsagePercent() float64 {
	if u.APICallsLimit <= 0 {
		return 0
	}
	return float64(u.APICallsCount) / float64(u.APICallsLimit) * 100
}

// GetStorageUsagePercent returns storage usage as percentage
func (u *OrganizationUsage) GetStorageUsagePercent() float64 {
	if u.StorageLimitBytes <= 0 {
		return 0
	}
	return float64(u.StorageUsedBytes) / float64(u.StorageLimitBytes) * 100
}

// ============================================================================
// Usage Summary (for billing/reporting)
// ============================================================================

// UsageSummary provides a summary of usage for a period
type UsageSummary struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`

	// Totals
	TotalAPICalls     int64 `json:"total_api_calls"`
	TotalDataSyncs    int64 `json:"total_data_syncs"`
	TotalRecordsSynced int64 `json:"total_records_synced"`
	TotalReports      int64 `json:"total_reports"`
	TotalExports      int64 `json:"total_exports"`
	TotalWebhooks     int64 `json:"total_webhooks"`

	// Current state
	CurrentStorage    int64 `json:"current_storage_bytes"`
	CurrentAccounts   int   `json:"current_accounts"`
	CurrentUsers      int   `json:"current_users"`

	// Limits
	APICallsLimit     int64 `json:"api_calls_limit"`
	StorageLimit      int64 `json:"storage_limit_bytes"`
	AccountsLimit     int   `json:"accounts_limit"`
	UsersLimit        int   `json:"users_limit"`

	// Usage percentage
	APIUsagePercent     float64 `json:"api_usage_percent"`
	StorageUsagePercent float64 `json:"storage_usage_percent"`
	AccountsUsagePercent float64 `json:"accounts_usage_percent"`
}

// ============================================================================
// Usage Event Log (for detailed tracking)
// ============================================================================

// UsageEvent logs individual usage events
type UsageEvent struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	UserID         *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`

	// Event details
	EventType      UsageType `json:"event_type" gorm:"type:usage_type;not null;index"`
	EventAction    string    `json:"event_action" gorm:"size:100;not null"` // e.g., "GET /api/campaigns"
	ResourceType   string    `json:"resource_type,omitempty" gorm:"size:50"` // campaign, report, etc.
	ResourceID     string    `json:"resource_id,omitempty" gorm:"size:255"`

	// Request info
	RequestMethod  string `json:"request_method,omitempty" gorm:"size:10"`
	RequestPath    string `json:"request_path,omitempty" gorm:"size:500"`
	ResponseStatus int    `json:"response_status,omitempty"`
	ResponseTimeMs int    `json:"response_time_ms,omitempty"`

	// Size metrics
	RequestSizeBytes  int64 `json:"request_size_bytes,omitempty"`
	ResponseSizeBytes int64 `json:"response_size_bytes,omitempty"`

	// IP and user agent
	IPAddress  string `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent  string `json:"user_agent,omitempty" gorm:"size:500"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

// ============================================================================
// Feature Usage (for feature flags and limits)
// ============================================================================

// FeatureUsage tracks usage of specific features
type FeatureUsage struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_feature_usage"`
	FeatureKey     string    `json:"feature_key" gorm:"size:50;not null;uniqueIndex:idx_feature_usage"`

	// Usage
	UsageCount     int64     `json:"usage_count" gorm:"default:0"`
	UsageLimit     int64     `json:"usage_limit" gorm:"default:-1"` // -1 = unlimited
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`

	// Period
	PeriodType     string    `json:"period_type" gorm:"size:20"` // daily, monthly, lifetime
	PeriodStart    *time.Time `json:"period_start,omitempty"`

	// Status
	IsEnabled      bool      `json:"is_enabled" gorm:"default:true"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CanUseFeature checks if the feature can be used
func (f *FeatureUsage) CanUseFeature() bool {
	if !f.IsEnabled {
		return false
	}
	if f.UsageLimit <= 0 { // -1 or 0 means unlimited
		return true
	}
	return f.UsageCount < f.UsageLimit
}

// ============================================================================
// Quota Alert
// ============================================================================

// QuotaAlertType represents the type of quota alert
type QuotaAlertType string

const (
	QuotaAlertWarning  QuotaAlertType = "warning"  // 80% usage
	QuotaAlertCritical QuotaAlertType = "critical" // 95% usage
	QuotaAlertExceeded QuotaAlertType = "exceeded" // 100% usage
)

// QuotaAlert stores quota alert notifications
type QuotaAlert struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`

	// Alert info
	AlertType      QuotaAlertType `json:"alert_type" gorm:"type:quota_alert_type;not null"`
	QuotaType      string         `json:"quota_type" gorm:"size:50;not null"` // api_calls, storage, accounts
	CurrentUsage   int64          `json:"current_usage"`
	UsageLimit     int64          `json:"usage_limit"`
	UsagePercent   float64        `json:"usage_percent"`

	// Notification
	NotifiedAt     *time.Time `json:"notified_at,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *uuid.UUID `json:"acknowledged_by,omitempty" gorm:"type:uuid"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ============================================================================
// Tenant Context (for multi-tenancy)
// ============================================================================

// TenantContext holds tenant information for request context
type TenantContext struct {
	OrganizationID   uuid.UUID        `json:"organization_id"`
	UserID           uuid.UUID        `json:"user_id"`
	UserRole         UserRole         `json:"user_role"`
	SubscriptionTier PlanTier         `json:"subscription_tier"`
	Limits           PlanLimits       `json:"limits"`
	IsActive         bool             `json:"is_active"`
}

// HasFeature checks if the tenant has access to a feature
func (t *TenantContext) HasFeature(feature string) bool {
	switch feature {
	case "advanced_analytics":
		return t.Limits.AdvancedAnalytics
	case "custom_reports":
		return t.Limits.CustomReports
	case "webhooks":
		return t.Limits.WebhooksEnabled
	case "api_access":
		return t.Limits.APIAccessEnabled
	case "white_label":
		return t.Limits.WhiteLabelEnabled
	case "priority_support":
		return t.Limits.PrioritySupport
	default:
		return false
	}
}

// CanAddAccount checks if tenant can add more ad accounts
func (t *TenantContext) CanAddAccount(currentCount int) bool {
	if t.Limits.MaxAdAccounts <= 0 { // Unlimited
		return true
	}
	return currentCount < t.Limits.MaxAdAccounts
}

// CanAddUser checks if tenant can add more users
func (t *TenantContext) CanAddUser(currentCount int) bool {
	if t.Limits.MaxUsersPerOrg <= 0 { // Unlimited
		return true
	}
	return currentCount < t.Limits.MaxUsersPerOrg
}
