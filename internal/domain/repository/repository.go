package repository

import (
	"context"

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
