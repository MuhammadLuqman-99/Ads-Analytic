package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Mock implementations
// ============================================================================

type mockPlatformConnector struct {
	getCampaignsCalled  int32
	getAdSetsCalled     int32
	getAdsCalled        int32
	getInsightsCalled   int32
	getAdAccountsCalled int32
	shouldFail          bool
	failCount           int32
	mu                  sync.Mutex
}

func (m *mockPlatformConnector) Platform() entity.Platform {
	return entity.PlatformMeta
}

func (m *mockPlatformConnector) GetAuthURL(state string) string {
	return "https://example.com/oauth"
}

func (m *mockPlatformConnector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	return &entity.OAuthToken{
		AccessToken: "test_token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *mockPlatformConnector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	return &entity.OAuthToken{
		AccessToken: "refreshed_token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *mockPlatformConnector) RevokeToken(ctx context.Context, accessToken string) error {
	return nil
}

func (m *mockPlatformConnector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	return &entity.PlatformUser{
		ID:    "user123",
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}

func (m *mockPlatformConnector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	atomic.AddInt32(&m.getAdAccountsCalled, 1)
	return []entity.PlatformAccount{
		{
			ID:       "12345",
			Name:     "Test Ad Account",
			Currency: "USD",
			Status:   "active",
		},
	}, nil
}

func (m *mockPlatformConnector) GetCampaigns(ctx context.Context, accessToken string, adAccountID string) ([]entity.Campaign, error) {
	atomic.AddInt32(&m.getCampaignsCalled, 1)
	return []entity.Campaign{
		{
			Platform:             entity.PlatformMeta,
			PlatformCampaignID:   "campaign_1",
			PlatformCampaignName: "Test Campaign",
			Status:               entity.CampaignStatusActive,
			DailyBudget:          decimalPtr(decimal.NewFromInt(100)),
		},
	}, nil
}

func (m *mockPlatformConnector) GetCampaign(ctx context.Context, accessToken string, campaignID string) (*entity.Campaign, error) {
	return &entity.Campaign{
		Platform:           entity.PlatformMeta,
		PlatformCampaignID: campaignID,
		Status:             entity.CampaignStatusActive,
	}, nil
}

func (m *mockPlatformConnector) GetAdSets(ctx context.Context, accessToken string, campaignID string) ([]entity.AdSet, error) {
	atomic.AddInt32(&m.getAdSetsCalled, 1)
	return []entity.AdSet{
		{
			Platform:          entity.PlatformMeta,
			PlatformAdSetID:   "adset_1",
			PlatformAdSetName: "Test Ad Set",
			Status:            entity.CampaignStatusActive,
		},
	}, nil
}

func (m *mockPlatformConnector) GetAdSet(ctx context.Context, accessToken string, adSetID string) (*entity.AdSet, error) {
	return &entity.AdSet{
		Platform:        entity.PlatformMeta,
		PlatformAdSetID: adSetID,
		Status:          entity.CampaignStatusActive,
	}, nil
}

func (m *mockPlatformConnector) GetAds(ctx context.Context, accessToken string, adSetID string) ([]entity.Ad, error) {
	atomic.AddInt32(&m.getAdsCalled, 1)
	return []entity.Ad{
		{
			Platform:       entity.PlatformMeta,
			PlatformAdID:   "ad_1",
			PlatformAdName: "Test Ad",
			Status:         entity.CampaignStatusActive,
		},
	}, nil
}

func (m *mockPlatformConnector) GetAd(ctx context.Context, accessToken string, adID string) (*entity.Ad, error) {
	return &entity.Ad{
		Platform:     entity.PlatformMeta,
		PlatformAdID: adID,
		Status:       entity.CampaignStatusActive,
	}, nil
}

func (m *mockPlatformConnector) GetCampaignInsights(ctx context.Context, accessToken string, campaignID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	atomic.AddInt32(&m.getInsightsCalled, 1)
	return []entity.CampaignMetricsDaily{
		{
			Platform:    entity.PlatformMeta,
			MetricDate:  time.Now().AddDate(0, 0, -1),
			Impressions: 1000,
			Clicks:      100,
			Spend:       decimal.NewFromFloat(50.00),
		},
	}, nil
}

func (m *mockPlatformConnector) GetAdSetInsights(ctx context.Context, accessToken string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	return []entity.AdSetMetricsDaily{}, nil
}

func (m *mockPlatformConnector) GetAdInsights(ctx context.Context, accessToken string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	return []entity.AdMetricsDaily{}, nil
}

func (m *mockPlatformConnector) GetAccountInsights(ctx context.Context, accessToken string, adAccountID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	return &entity.AggregatedMetrics{
		TotalSpend:       decimal.NewFromFloat(1000),
		TotalImpressions: 50000,
		TotalClicks:      5000,
		TotalConversions: 100,
	}, nil
}

func (m *mockPlatformConnector) GetRateLimit() service.RateLimitStatus {
	return service.RateLimitStatus{
		Platform:  entity.PlatformMeta,
		Limit:     200,
		Remaining: 100,
	}
}

func (m *mockPlatformConnector) HealthCheck(ctx context.Context) error {
	return nil
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// Mock repository implementations
type mockConnectedAccountRepo struct {
	accounts []entity.ConnectedAccount
}

func (m *mockConnectedAccountRepo) Create(ctx context.Context, account *entity.ConnectedAccount) error {
	return nil
}

func (m *mockConnectedAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ConnectedAccount, error) {
	for _, acc := range m.accounts {
		if acc.ID == id {
			return &acc, nil
		}
	}
	return nil, nil
}

func (m *mockConnectedAccountRepo) GetByPlatformAccountID(ctx context.Context, orgID uuid.UUID, platform entity.Platform, platformAccountID string) (*entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepo) Update(ctx context.Context, account *entity.ConnectedAccount) error {
	return nil
}

func (m *mockConnectedAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockConnectedAccountRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error) {
	return m.accounts, nil
}

func (m *mockConnectedAccountRepo) ListByPlatform(ctx context.Context, orgID uuid.UUID, platform entity.Platform) ([]entity.ConnectedAccount, error) {
	return m.accounts, nil
}

func (m *mockConnectedAccountRepo) ListExpiring(ctx context.Context, withinMinutes int) ([]entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepo) ListActive(ctx context.Context) ([]entity.ConnectedAccount, error) {
	return m.accounts, nil
}

func (m *mockConnectedAccountRepo) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *interface{}) error {
	return nil
}

func (m *mockConnectedAccountRepo) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus, syncError string) error {
	return nil
}

func (m *mockConnectedAccountRepo) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockAdAccountRepo struct {
	accounts []entity.AdAccount
}

func (m *mockAdAccountRepo) Create(ctx context.Context, account *entity.AdAccount) error {
	return nil
}

func (m *mockAdAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AdAccount, error) {
	for _, acc := range m.accounts {
		if acc.ID == id {
			return &acc, nil
		}
	}
	return nil, nil
}

func (m *mockAdAccountRepo) GetByPlatformID(ctx context.Context, connectedAccountID uuid.UUID, platformAdAccountID string) (*entity.AdAccount, error) {
	return nil, nil
}

func (m *mockAdAccountRepo) Update(ctx context.Context, account *entity.AdAccount) error {
	return nil
}

func (m *mockAdAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAdAccountRepo) ListByConnectedAccount(ctx context.Context, connectedAccountID uuid.UUID) ([]entity.AdAccount, error) {
	return m.accounts, nil
}

func (m *mockAdAccountRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.AdAccount, error) {
	return m.accounts, nil
}

func (m *mockAdAccountRepo) Upsert(ctx context.Context, account *entity.AdAccount) error {
	account.ID = uuid.New()
	m.accounts = append(m.accounts, *account)
	return nil
}

type mockCampaignRepo struct {
	campaigns []entity.Campaign
}

func (m *mockCampaignRepo) Create(ctx context.Context, campaign *entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Campaign, error) {
	return nil, nil
}

func (m *mockCampaignRepo) GetByPlatformID(ctx context.Context, adAccountID uuid.UUID, platformCampaignID string) (*entity.Campaign, error) {
	return nil, nil
}

func (m *mockCampaignRepo) Update(ctx context.Context, campaign *entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockCampaignRepo) List(ctx context.Context, filter entity.CampaignFilter) ([]entity.Campaign, int64, error) {
	return m.campaigns, int64(len(m.campaigns)), nil
}

func (m *mockCampaignRepo) ListByAdAccount(ctx context.Context, adAccountID uuid.UUID) ([]entity.Campaign, error) {
	return m.campaigns, nil
}

func (m *mockCampaignRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.Campaign, error) {
	return m.campaigns, nil
}

func (m *mockCampaignRepo) Upsert(ctx context.Context, campaign *entity.Campaign) error {
	campaign.ID = uuid.New()
	m.campaigns = append(m.campaigns, *campaign)
	return nil
}

func (m *mockCampaignRepo) BulkUpsert(ctx context.Context, campaigns []entity.Campaign) error {
	for i := range campaigns {
		campaigns[i].ID = uuid.New()
	}
	m.campaigns = campaigns
	return nil
}

func (m *mockCampaignRepo) GetSummaries(ctx context.Context, filter entity.CampaignFilter) ([]entity.CampaignSummary, error) {
	return nil, nil
}

func (m *mockCampaignRepo) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockAdSetRepo struct {
	adSets []entity.AdSet
}

func (m *mockAdSetRepo) Create(ctx context.Context, adSet *entity.AdSet) error {
	return nil
}

func (m *mockAdSetRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AdSet, error) {
	return nil, nil
}

func (m *mockAdSetRepo) GetByPlatformID(ctx context.Context, campaignID uuid.UUID, platformAdSetID string) (*entity.AdSet, error) {
	return nil, nil
}

func (m *mockAdSetRepo) Update(ctx context.Context, adSet *entity.AdSet) error {
	return nil
}

func (m *mockAdSetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAdSetRepo) List(ctx context.Context, filter entity.AdSetFilter) ([]entity.AdSet, int64, error) {
	return m.adSets, int64(len(m.adSets)), nil
}

func (m *mockAdSetRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]entity.AdSet, error) {
	return m.adSets, nil
}

func (m *mockAdSetRepo) Upsert(ctx context.Context, adSet *entity.AdSet) error {
	adSet.ID = uuid.New()
	m.adSets = append(m.adSets, *adSet)
	return nil
}

func (m *mockAdSetRepo) BulkUpsert(ctx context.Context, adSets []entity.AdSet) error {
	for i := range adSets {
		adSets[i].ID = uuid.New()
	}
	m.adSets = adSets
	return nil
}

type mockAdRepo struct {
	ads []entity.Ad
}

func (m *mockAdRepo) Create(ctx context.Context, ad *entity.Ad) error {
	return nil
}

func (m *mockAdRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Ad, error) {
	return nil, nil
}

func (m *mockAdRepo) GetByPlatformID(ctx context.Context, adSetID uuid.UUID, platformAdID string) (*entity.Ad, error) {
	return nil, nil
}

func (m *mockAdRepo) Update(ctx context.Context, ad *entity.Ad) error {
	return nil
}

func (m *mockAdRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAdRepo) List(ctx context.Context, filter entity.AdFilter) ([]entity.Ad, int64, error) {
	return m.ads, int64(len(m.ads)), nil
}

func (m *mockAdRepo) ListByAdSet(ctx context.Context, adSetID uuid.UUID) ([]entity.Ad, error) {
	return m.ads, nil
}

func (m *mockAdRepo) Upsert(ctx context.Context, ad *entity.Ad) error {
	ad.ID = uuid.New()
	m.ads = append(m.ads, *ad)
	return nil
}

func (m *mockAdRepo) BulkUpsert(ctx context.Context, ads []entity.Ad) error {
	for i := range ads {
		ads[i].ID = uuid.New()
	}
	m.ads = ads
	return nil
}

type mockMetricsRepo struct {
	campaignMetrics []entity.CampaignMetricsDaily
}

func (m *mockMetricsRepo) CreateCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) GetCampaignMetrics(ctx context.Context, campaignID uuid.UUID, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	return m.campaignMetrics, nil
}

func (m *mockMetricsRepo) UpsertCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) BulkUpsertCampaignMetrics(ctx context.Context, metrics []entity.CampaignMetricsDaily) error {
	m.campaignMetrics = metrics
	return nil
}

func (m *mockMetricsRepo) CreateAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) GetAdSetMetrics(ctx context.Context, adSetID uuid.UUID, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	return nil, nil
}

func (m *mockMetricsRepo) UpsertAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) BulkUpsertAdSetMetrics(ctx context.Context, metrics []entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) CreateAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) GetAdMetrics(ctx context.Context, adID uuid.UUID, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	return nil, nil
}

func (m *mockMetricsRepo) UpsertAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) BulkUpsertAdMetrics(ctx context.Context, metrics []entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepo) GetAggregatedMetrics(ctx context.Context, filter entity.MetricsFilter) (*entity.AggregatedMetrics, error) {
	return nil, nil
}

func (m *mockMetricsRepo) GetMetricsByPlatform(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) ([]entity.PlatformMetricsSummary, error) {
	return nil, nil
}

func (m *mockMetricsRepo) GetDailyTrend(ctx context.Context, filter entity.MetricsFilter) ([]entity.DailyMetricsTrend, error) {
	return nil, nil
}

func (m *mockMetricsRepo) GetTopPerformingCampaigns(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange, limit int) ([]entity.TopPerformer, error) {
	return nil, nil
}

// Verify interfaces
var _ service.PlatformConnector = (*mockPlatformConnector)(nil)
var _ repository.ConnectedAccountRepository = (*mockConnectedAccountRepo)(nil)
var _ repository.AdAccountRepository = (*mockAdAccountRepo)(nil)
var _ repository.CampaignRepository = (*mockCampaignRepo)(nil)
var _ repository.AdSetRepository = (*mockAdSetRepo)(nil)
var _ repository.AdRepository = (*mockAdRepo)(nil)
var _ repository.MetricsRepository = (*mockMetricsRepo)(nil)

// ============================================================================
// Tests
// ============================================================================

func TestDefaultMetaSyncWorkerConfig(t *testing.T) {
	config := DefaultMetaSyncWorkerConfig()

	if config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5", config.MaxConcurrency)
	}
	if config.RateLimitCalls != 200 {
		t.Errorf("RateLimitCalls = %d, want 200", config.RateLimitCalls)
	}
	if config.RateLimitWindow != time.Hour {
		t.Errorf("RateLimitWindow = %v, want 1h", config.RateLimitWindow)
	}
	if config.SyncTimeout != 10*time.Minute {
		t.Errorf("SyncTimeout = %v, want 10m", config.SyncTimeout)
	}
	if config.BatchTimeout != 60*time.Minute {
		t.Errorf("BatchTimeout = %v, want 60m", config.BatchTimeout)
	}
}

func TestNewMetaSyncWorker(t *testing.T) {
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	worker := NewMetaSyncWorker(
		nil, // nil config should use defaults
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	if worker == nil {
		t.Fatal("Expected worker to be created")
	}
	if worker.config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5", worker.config.MaxConcurrency)
	}
	if worker.IsRunning() {
		t.Error("Worker should not be running initially")
	}
}

func TestMetaSyncWorker_StartStop(t *testing.T) {
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	worker := NewMetaSyncWorker(
		nil,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker
	err := worker.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start worker: %v", err)
	}

	// Give workers time to start
	time.Sleep(50 * time.Millisecond)

	if !worker.IsRunning() {
		t.Error("Worker should be running after Start")
	}

	// Stop worker
	worker.Stop()

	// Give workers time to stop
	time.Sleep(100 * time.Millisecond)
}

func TestMetaSyncWorker_GetStats(t *testing.T) {
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	worker := NewMetaSyncWorker(
		nil,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	stats := worker.GetStats()

	if stats["is_running"].(bool) != false {
		t.Error("Expected is_running to be false")
	}
	if stats["running_tasks"].(int32) != 0 {
		t.Error("Expected running_tasks to be 0")
	}
	if stats["concurrency"].(int) != 5 {
		t.Errorf("Expected concurrency to be 5, got %d", stats["concurrency"].(int))
	}
}

func TestSyncTaskResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   SyncTaskResult
		expected bool
	}{
		{
			name:     "success - no error",
			result:   SyncTaskResult{},
			expected: true,
		},
		{
			name: "failure - has error",
			result: SyncTaskResult{
				Error: context.DeadlineExceeded,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsSuccess(); got != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSyncTaskResult_HasPartialSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   SyncTaskResult
		expected bool
	}{
		{
			name: "partial success - has error but synced campaigns",
			result: SyncTaskResult{
				Error:           context.DeadlineExceeded,
				CampaignsSynced: 5,
			},
			expected: true,
		},
		{
			name: "no partial - has error but nothing synced",
			result: SyncTaskResult{
				Error: context.DeadlineExceeded,
			},
			expected: false,
		},
		{
			name: "full success - no error",
			result: SyncTaskResult{
				CampaignsSynced: 5,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasPartialSuccess(); got != tt.expected {
				t.Errorf("HasPartialSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMetaSyncWorker_SyncAllActiveAccounts_NoAccounts(t *testing.T) {
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{accounts: []entity.ConnectedAccount{}}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	worker := NewMetaSyncWorker(
		nil,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	ctx := context.Background()
	result, err := worker.SyncAllActiveAccounts(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.TotalTasks != 0 {
		t.Errorf("Expected 0 tasks, got %d", result.TotalTasks)
	}
}

func TestMetaSyncWorker_SyncAllActiveAccounts_WithAccounts(t *testing.T) {
	tokenExpiry := time.Now().Add(24 * time.Hour)
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{
		accounts: []entity.ConnectedAccount{
			{
				BaseEntity:        entity.BaseEntity{ID: uuid.New()},
				OrganizationID:    uuid.New(),
				Platform:          entity.PlatformMeta,
				PlatformAccountID: "12345",
				AccessToken:       "test_token",
				TokenExpiresAt:    &tokenExpiry,
				Status:            entity.AccountStatusActive,
			},
		},
	}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	config := &MetaSyncWorkerConfig{
		MaxConcurrency:  2,
		RateLimitCalls:  1000, // High limit for testing
		RateLimitWindow: time.Hour,
		SyncTimeout:     30 * time.Second,
		BatchTimeout:    1 * time.Minute,
		RetryConfig: &RetryConfig{
			MaxRetries:   1,
			BaseDelay:    10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			JitterFactor: 0,
		},
		DefaultDateRange: entity.DateRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
		},
	}

	worker := NewMetaSyncWorker(
		config,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	ctx := context.Background()
	result, err := worker.SyncAllActiveAccounts(ctx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.TotalTasks != 1 {
		t.Errorf("Expected 1 task, got %d", result.TotalTasks)
	}

	// Verify API was called
	if atomic.LoadInt32(&connector.getAdAccountsCalled) == 0 {
		t.Error("Expected GetAdAccounts to be called")
	}
}

func TestMetaSyncWorker_SyncAccount_ExpiredToken(t *testing.T) {
	expiredTime := time.Now().Add(-1 * time.Hour)
	connector := &mockPlatformConnector{}
	connectedAccRepo := &mockConnectedAccountRepo{
		accounts: []entity.ConnectedAccount{
			{
				BaseEntity:        entity.BaseEntity{ID: uuid.New()},
				OrganizationID:    uuid.New(),
				Platform:          entity.PlatformMeta,
				PlatformAccountID: "12345",
				AccessToken:       "expired_token",
				TokenExpiresAt:    &expiredTime,
				Status:            entity.AccountStatusActive,
			},
		},
	}
	adAccountRepo := &mockAdAccountRepo{}
	campaignRepo := &mockCampaignRepo{}
	adSetRepo := &mockAdSetRepo{}
	adRepo := &mockAdRepo{}
	metricsRepo := &mockMetricsRepo{}

	worker := NewMetaSyncWorker(
		nil,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		zerolog.Nop(),
	)

	ctx := context.Background()
	_, err := worker.SyncAccount(ctx, connectedAccRepo.accounts[0].ID)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}
