package service

import (
	"context"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/google/uuid"
)

// PlatformConnector defines the interface for all ad platform integrations
// This interface allows for easy testing and mocking of platform connectors
type PlatformConnector interface {
	// Platform returns the platform type
	Platform() entity.Platform

	// OAuth methods
	// GetAuthURL generates the OAuth authorization URL for the platform
	GetAuthURL(state string) string

	// ExchangeCode exchanges an authorization code for OAuth tokens
	ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error)

	// RefreshToken refreshes an expired access token using the refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error)

	// RevokeToken revokes an access token
	RevokeToken(ctx context.Context, accessToken string) error

	// Account methods
	// GetUserInfo retrieves the authenticated user's information
	GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error)

	// GetAdAccounts retrieves all ad accounts accessible by the token
	GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error)

	// Campaign methods
	// GetCampaigns retrieves all campaigns for an ad account
	GetCampaigns(ctx context.Context, accessToken string, adAccountID string) ([]entity.Campaign, error)

	// GetCampaign retrieves a single campaign by ID
	GetCampaign(ctx context.Context, accessToken string, campaignID string) (*entity.Campaign, error)

	// AdSet methods
	// GetAdSets retrieves all ad sets for a campaign
	GetAdSets(ctx context.Context, accessToken string, campaignID string) ([]entity.AdSet, error)

	// GetAdSet retrieves a single ad set by ID
	GetAdSet(ctx context.Context, accessToken string, adSetID string) (*entity.AdSet, error)

	// Ad methods
	// GetAds retrieves all ads for an ad set
	GetAds(ctx context.Context, accessToken string, adSetID string) ([]entity.Ad, error)

	// GetAd retrieves a single ad by ID
	GetAd(ctx context.Context, accessToken string, adID string) (*entity.Ad, error)

	// Metrics methods
	// GetCampaignInsights retrieves performance metrics for a campaign
	GetCampaignInsights(ctx context.Context, accessToken string, campaignID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error)

	// GetAdSetInsights retrieves performance metrics for an ad set
	GetAdSetInsights(ctx context.Context, accessToken string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error)

	// GetAdInsights retrieves performance metrics for an ad
	GetAdInsights(ctx context.Context, accessToken string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error)

	// GetAccountInsights retrieves aggregated insights for an ad account
	GetAccountInsights(ctx context.Context, accessToken string, adAccountID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error)

	// Rate limiting
	// GetRateLimit returns the current rate limit status
	GetRateLimit() RateLimitStatus

	// Health check
	// HealthCheck verifies the connector can connect to the platform
	HealthCheck(ctx context.Context) error
}

// RateLimitStatus represents the current rate limit status for a platform
type RateLimitStatus struct {
	Platform  entity.Platform
	Limit     int
	Remaining int
	ResetAt   int64 // Unix timestamp
	IsLimited bool
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Platform        entity.Platform
	AccountID       uuid.UUID
	CampaignsSynced int
	AdSetsSynced    int
	AdsSynced       int
	MetricsSynced   int
	Errors          []error
	StartedAt       int64
	CompletedAt     int64
}

// IsSuccess returns true if the sync completed without errors
func (r *SyncResult) IsSuccess() bool {
	return len(r.Errors) == 0
}

// HasPartialSuccess returns true if some items were synced despite errors
func (r *SyncResult) HasPartialSuccess() bool {
	return len(r.Errors) > 0 && (r.CampaignsSynced > 0 || r.AdSetsSynced > 0 || r.AdsSynced > 0)
}

// ConnectorConfig holds common configuration for platform connectors
type ConnectorConfig struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	APIVersion      string
	RateLimitCalls  int
	RateLimitWindow int // seconds
	Timeout         int // seconds
	MaxRetries      int
}

// ConnectorFactory creates platform connectors
type ConnectorFactory interface {
	// CreateConnector creates a connector for the specified platform
	CreateConnector(platform entity.Platform) (PlatformConnector, error)

	// GetConnector returns an existing connector for the platform
	GetConnector(platform entity.Platform) (PlatformConnector, bool)

	// RegisterConnector registers a connector for a platform
	RegisterConnector(platform entity.Platform, connector PlatformConnector)
}

// BatchSyncRequest represents a request to sync multiple accounts
type BatchSyncRequest struct {
	OrganizationID uuid.UUID
	AccountIDs     []uuid.UUID
	Platforms      []entity.Platform
	DateRange      entity.DateRange
	SyncCampaigns  bool
	SyncAdSets     bool
	SyncAds        bool
	SyncMetrics    bool
}

// BatchSyncResult represents the result of a batch sync operation
type BatchSyncResult struct {
	OrganizationID uuid.UUID
	Results        []SyncResult
	TotalAccounts  int
	SuccessCount   int
	PartialCount   int
	FailureCount   int
	StartedAt      int64
	CompletedAt    int64
}

// DataSyncer defines the interface for syncing data from platforms
type DataSyncer interface {
	// SyncAccount syncs all data for a single connected account
	SyncAccount(ctx context.Context, accountID uuid.UUID) (*SyncResult, error)

	// SyncCampaigns syncs campaigns for an account
	SyncCampaigns(ctx context.Context, accountID uuid.UUID) (int, error)

	// SyncMetrics syncs metrics for an account within a date range
	SyncMetrics(ctx context.Context, accountID uuid.UUID, dateRange entity.DateRange) (int, error)

	// BatchSync syncs multiple accounts
	BatchSync(ctx context.Context, request BatchSyncRequest) (*BatchSyncResult, error)
}

// TokenManager defines the interface for managing OAuth tokens
type TokenManager interface {
	// RefreshTokenIfNeeded checks if a token needs refresh and refreshes it
	RefreshTokenIfNeeded(ctx context.Context, accountID uuid.UUID) error

	// RefreshAllExpiring refreshes all tokens that are about to expire
	RefreshAllExpiring(ctx context.Context) (int, error)

	// RevokeToken revokes a token for an account
	RevokeToken(ctx context.Context, accountID uuid.UUID) error
}
