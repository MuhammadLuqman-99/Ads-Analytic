package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Mock Repository Implementations
// ============================================================================

type mockMetricsRepository struct {
	campaignMetrics map[uuid.UUID][]entity.CampaignMetricsDaily
	aggregated      *entity.AggregatedMetrics
	platformMetrics []entity.PlatformMetricsSummary
	dailyTrend      []entity.DailyMetricsTrend
	topPerformers   []entity.TopPerformer
}

func newMockMetricsRepo() *mockMetricsRepository {
	return &mockMetricsRepository{
		campaignMetrics: make(map[uuid.UUID][]entity.CampaignMetricsDaily),
	}
}

func (m *mockMetricsRepository) CreateCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) GetCampaignMetrics(ctx context.Context, campaignID uuid.UUID, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	if metrics, ok := m.campaignMetrics[campaignID]; ok {
		// Filter by date range
		var filtered []entity.CampaignMetricsDaily
		for _, metric := range metrics {
			if !metric.MetricDate.Before(dateRange.StartDate) && !metric.MetricDate.After(dateRange.EndDate) {
				filtered = append(filtered, metric)
			}
		}
		return filtered, nil
	}
	return []entity.CampaignMetricsDaily{}, nil
}

func (m *mockMetricsRepository) UpsertCampaignMetrics(ctx context.Context, metrics *entity.CampaignMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) BulkUpsertCampaignMetrics(ctx context.Context, metrics []entity.CampaignMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) CreateAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) GetAdSetMetrics(ctx context.Context, adSetID uuid.UUID, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	return nil, nil
}

func (m *mockMetricsRepository) UpsertAdSetMetrics(ctx context.Context, metrics *entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) BulkUpsertAdSetMetrics(ctx context.Context, metrics []entity.AdSetMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) CreateAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) GetAdMetrics(ctx context.Context, adID uuid.UUID, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	return nil, nil
}

func (m *mockMetricsRepository) UpsertAdMetrics(ctx context.Context, metrics *entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) BulkUpsertAdMetrics(ctx context.Context, metrics []entity.AdMetricsDaily) error {
	return nil
}

func (m *mockMetricsRepository) GetAggregatedMetrics(ctx context.Context, filter entity.MetricsFilter) (*entity.AggregatedMetrics, error) {
	if m.aggregated != nil {
		return m.aggregated, nil
	}
	return &entity.AggregatedMetrics{}, nil
}

func (m *mockMetricsRepository) GetMetricsByPlatform(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) ([]entity.PlatformMetricsSummary, error) {
	return m.platformMetrics, nil
}

func (m *mockMetricsRepository) GetDailyTrend(ctx context.Context, filter entity.MetricsFilter) ([]entity.DailyMetricsTrend, error) {
	return m.dailyTrend, nil
}

func (m *mockMetricsRepository) GetTopPerformingCampaigns(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange, limit int) ([]entity.TopPerformer, error) {
	return m.topPerformers, nil
}

type mockCampaignRepository struct {
	campaigns []entity.Campaign
}

func newMockCampaignRepo() *mockCampaignRepository {
	return &mockCampaignRepository{
		campaigns: []entity.Campaign{},
	}
}

func (m *mockCampaignRepository) Create(ctx context.Context, campaign *entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Campaign, error) {
	for _, c := range m.campaigns {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockCampaignRepository) GetByPlatformID(ctx context.Context, adAccountID uuid.UUID, platformCampaignID string) (*entity.Campaign, error) {
	return nil, nil
}

func (m *mockCampaignRepository) Update(ctx context.Context, campaign *entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockCampaignRepository) List(ctx context.Context, filter entity.CampaignFilter) ([]entity.Campaign, int64, error) {
	var filtered []entity.Campaign
	for _, c := range m.campaigns {
		if c.OrganizationID == filter.OrganizationID {
			// Platform filter
			if len(filter.Platforms) > 0 {
				matched := false
				for _, p := range filter.Platforms {
					if c.Platform == p {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			filtered = append(filtered, c)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func (m *mockCampaignRepository) ListByAdAccount(ctx context.Context, adAccountID uuid.UUID) ([]entity.Campaign, error) {
	return nil, nil
}

func (m *mockCampaignRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.Campaign, error) {
	var filtered []entity.Campaign
	for _, c := range m.campaigns {
		if c.OrganizationID == orgID {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

func (m *mockCampaignRepository) Upsert(ctx context.Context, campaign *entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepository) BulkUpsert(ctx context.Context, campaigns []entity.Campaign) error {
	return nil
}

func (m *mockCampaignRepository) GetSummaries(ctx context.Context, filter entity.CampaignFilter) ([]entity.CampaignSummary, error) {
	return nil, nil
}

func (m *mockCampaignRepository) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	return nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func createTestService() (*Service, *mockMetricsRepository, *mockCampaignRepository) {
	metricsRepo := newMockMetricsRepo()
	campaignRepo := newMockCampaignRepo()
	service := NewService(metricsRepo, campaignRepo)
	return service, metricsRepo, campaignRepo
}

func createTestCampaign(orgID uuid.UUID, platform entity.Platform) entity.Campaign {
	return entity.Campaign{
		BaseEntity:     entity.BaseEntity{ID: uuid.New()},
		OrganizationID: orgID,
		Platform:       platform,
		Status:         entity.CampaignStatusActive,
	}
}

func createTestMetric(campaignID, orgID uuid.UUID, platform entity.Platform, date time.Time, spend, revenue float64, impressions, clicks, conversions int64) entity.CampaignMetricsDaily {
	return entity.CampaignMetricsDaily{
		BaseEntity:      entity.BaseEntity{ID: uuid.New()},
		CampaignID:      campaignID,
		OrganizationID:  orgID,
		Platform:        platform,
		MetricDate:      date,
		Spend:           decimal.NewFromFloat(spend),
		ConversionValue: decimal.NewFromFloat(revenue),
		Impressions:     impressions,
		Clicks:          clicks,
		Conversions:     conversions,
		Currency:        "MYR",
	}
}

// ============================================================================
// CalculateAnalytics Tests
// ============================================================================

func TestCalculateAnalytics_Success(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()
	startDate := now.AddDate(0, 0, -7)
	endDate := now

	// Create test campaigns
	metaCampaign := createTestCampaign(orgID, entity.PlatformMeta)
	tiktokCampaign := createTestCampaign(orgID, entity.PlatformTikTok)
	campaignRepo.campaigns = []entity.Campaign{metaCampaign, tiktokCampaign}

	// Create test metrics
	metricsRepo.campaignMetrics[metaCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(metaCampaign.ID, orgID, entity.PlatformMeta, now.AddDate(0, 0, -1), 100, 500, 10000, 500, 50),
		createTestMetric(metaCampaign.ID, orgID, entity.PlatformMeta, now.AddDate(0, 0, -2), 150, 600, 15000, 750, 75),
	}
	metricsRepo.campaignMetrics[tiktokCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(tiktokCampaign.ID, orgID, entity.PlatformTikTok, now.AddDate(0, 0, -1), 80, 400, 8000, 400, 40),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Verify overall metrics
	if result.OverallMetrics == nil {
		t.Fatal("OverallMetrics should not be nil")
	}

	expectedSpend := decimal.NewFromFloat(330) // 100 + 150 + 80
	if !result.OverallMetrics.TotalSpend.Equal(expectedSpend) {
		t.Errorf("TotalSpend: got %s, want %s", result.OverallMetrics.TotalSpend.String(), expectedSpend.String())
	}

	expectedClicks := int64(1650) // 500 + 750 + 400
	if result.OverallMetrics.TotalClicks != expectedClicks {
		t.Errorf("TotalClicks: got %d, want %d", result.OverallMetrics.TotalClicks, expectedClicks)
	}

	// Verify ROAS is calculated
	if result.OverallMetrics.ROAS == nil {
		t.Error("ROAS should be calculated")
	}

	// Verify platform breakdown
	if len(result.PlatformMetrics) != 2 {
		t.Errorf("Expected 2 platforms, got %d", len(result.PlatformMetrics))
	}

	// Verify comparison exists when multiple platforms
	if result.Comparison == nil {
		t.Error("Comparison should be generated for multiple platforms")
	}

	// Verify data quality report
	if result.DataQuality == nil {
		t.Error("DataQuality should not be nil")
	}
}

func TestCalculateAnalytics_ZeroDivisionProtection(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	// Create metric with zero impressions, clicks, conversions
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 0, 0, 0, 0, 0),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// All ratio metrics should be nil (not calculated due to zero denominator)
	if result.OverallMetrics.ROAS != nil {
		t.Error("ROAS should be nil when spend is 0")
	}
	if result.OverallMetrics.CTR != nil {
		t.Error("CTR should be nil when impressions is 0")
	}
	if result.OverallMetrics.CPC != nil {
		t.Error("CPC should be nil when clicks is 0")
	}
	if result.OverallMetrics.CPA != nil {
		t.Error("CPA should be nil when conversions is 0")
	}
	if result.OverallMetrics.CPM != nil {
		t.Error("CPM should be nil when impressions is 0")
	}
}

func TestCalculateAnalytics_MissingData(t *testing.T) {
	service, _, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	// Create campaign but no metrics
	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -7),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Data quality report should indicate missing data
	if result.DataQuality == nil {
		t.Fatal("DataQuality should not be nil")
	}

	if result.DataQuality.HasCompleteData {
		t.Error("HasCompleteData should be false when no metrics exist")
	}

	if len(result.DataQuality.Warnings) == 0 {
		t.Error("Warnings should contain message about missing data")
	}
}

func TestCalculateAnalytics_PlatformFilter(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	// Create campaigns for different platforms
	metaCampaign := createTestCampaign(orgID, entity.PlatformMeta)
	tiktokCampaign := createTestCampaign(orgID, entity.PlatformTikTok)
	campaignRepo.campaigns = []entity.Campaign{metaCampaign, tiktokCampaign}

	// Create metrics for both
	metricsRepo.campaignMetrics[metaCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(metaCampaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}
	metricsRepo.campaignMetrics[tiktokCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(tiktokCampaign.ID, orgID, entity.PlatformTikTok, now, 80, 400, 8000, 400, 40),
	}

	// Request only Meta
	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		Platforms:      []entity.Platform{entity.PlatformMeta},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Should only have Meta data
	if len(result.PlatformMetrics) != 1 {
		t.Errorf("Expected 1 platform, got %d", len(result.PlatformMetrics))
	}

	if _, exists := result.PlatformMetrics[entity.PlatformMeta]; !exists {
		t.Error("Should have Meta platform metrics")
	}

	if _, exists := result.PlatformMetrics[entity.PlatformTikTok]; exists {
		t.Error("Should NOT have TikTok platform metrics")
	}

	// Total spend should only include Meta
	expectedSpend := decimal.NewFromFloat(100)
	if !result.OverallMetrics.TotalSpend.Equal(expectedSpend) {
		t.Errorf("TotalSpend: got %s, want %s (Meta only)", result.OverallMetrics.TotalSpend.String(), expectedSpend.String())
	}
}

func TestCalculateAnalytics_CampaignFilter(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	// Create multiple campaigns
	campaign1 := createTestCampaign(orgID, entity.PlatformMeta)
	campaign2 := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign1, campaign2}

	// Create metrics for both
	metricsRepo.campaignMetrics[campaign1.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign1.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}
	metricsRepo.campaignMetrics[campaign2.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign2.ID, orgID, entity.PlatformMeta, now, 200, 1000, 20000, 1000, 100),
	}

	// Request only campaign1
	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		CampaignIDs:    []uuid.UUID{campaign1.ID},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Total spend should only include campaign1
	expectedSpend := decimal.NewFromFloat(100)
	if !result.OverallMetrics.TotalSpend.Equal(expectedSpend) {
		t.Errorf("TotalSpend: got %s, want %s (campaign1 only)", result.OverallMetrics.TotalSpend.String(), expectedSpend.String())
	}
}

func TestCalculateAnalytics_InvalidRequest(t *testing.T) {
	service, _, _ := createTestService()
	ctx := context.Background()

	tests := []struct {
		name string
		req  entity.AnalyticsRequest
	}{
		{
			name: "missing organization ID",
			req: entity.AnalyticsRequest{
				DateRange: entity.DateRange{
					StartDate: time.Now().AddDate(0, 0, -7),
					EndDate:   time.Now(),
				},
			},
		},
		{
			name: "missing date range",
			req: entity.AnalyticsRequest{
				OrganizationID: uuid.New(),
			},
		},
		{
			name: "start date after end date",
			req: entity.AnalyticsRequest{
				OrganizationID: uuid.New(),
				DateRange: entity.DateRange{
					StartDate: time.Now(),
					EndDate:   time.Now().AddDate(0, 0, -7),
				},
			},
		},
		{
			name: "invalid platform",
			req: entity.AnalyticsRequest{
				OrganizationID: uuid.New(),
				DateRange: entity.DateRange{
					StartDate: time.Now().AddDate(0, 0, -7),
					EndDate:   time.Now(),
				},
				Platforms: []entity.Platform{"invalid_platform"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CalculateAnalytics(ctx, tt.req)
			if err == nil {
				t.Error("Expected error for invalid request")
			}
		})
	}
}

func TestCalculateAnalytics_PlatformComparison(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	// Create campaigns with different performance
	metaCampaign := createTestCampaign(orgID, entity.PlatformMeta)
	tiktokCampaign := createTestCampaign(orgID, entity.PlatformTikTok)
	campaignRepo.campaigns = []entity.Campaign{metaCampaign, tiktokCampaign}

	// Meta: Better ROAS (500/100 = 5x)
	metricsRepo.campaignMetrics[metaCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(metaCampaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}

	// TikTok: Lower ROAS (300/80 = 3.75x)
	metricsRepo.campaignMetrics[tiktokCampaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(tiktokCampaign.ID, orgID, entity.PlatformTikTok, now, 80, 300, 12000, 600, 30),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	if result.Comparison == nil {
		t.Fatal("Comparison should be generated")
	}

	// Best ROAS should be Meta
	if result.Comparison.BestROAS == nil {
		t.Fatal("BestROAS should be calculated")
	}
	if result.Comparison.BestROAS.Platform != entity.PlatformMeta {
		t.Errorf("BestROAS platform: got %s, want %s", result.Comparison.BestROAS.Platform, entity.PlatformMeta)
	}

	// Platform count should be 2
	if result.Comparison.PlatformCount != 2 {
		t.Errorf("PlatformCount: got %d, want 2", result.Comparison.PlatformCount)
	}
}

func TestCalculatedMetrics_CalculateDerivedFields(t *testing.T) {
	tests := []struct {
		name           string
		metrics        *entity.CalculatedMetrics
		expectROAS     bool
		expectCTR      bool
		expectCPC      bool
		expectCPA      bool
		expectCPM      bool
		expectConvRate bool
	}{
		{
			name: "all values present",
			metrics: &entity.CalculatedMetrics{
				TotalSpend:       decimal.NewFromInt(100),
				TotalRevenue:     decimal.NewFromInt(500),
				TotalImpressions: 10000,
				TotalClicks:      500,
				TotalConversions: 50,
			},
			expectROAS:     true,
			expectCTR:      true,
			expectCPC:      true,
			expectCPA:      true,
			expectCPM:      true,
			expectConvRate: true,
		},
		{
			name: "zero spend",
			metrics: &entity.CalculatedMetrics{
				TotalSpend:       decimal.Zero,
				TotalRevenue:     decimal.NewFromInt(500),
				TotalImpressions: 10000,
				TotalClicks:      500,
				TotalConversions: 50,
			},
			expectROAS:     false, // ROAS = Revenue/Spend, spend is zero
			expectCTR:      true,  // CTR = Clicks/Impressions, works with zero spend
			expectCPC:      true,  // CPC = Spend/Clicks, works with zero spend (result is 0)
			expectCPA:      true,  // CPA = Spend/Conversions, works (result is 0)
			expectCPM:      true,  // CPM = Spend/Impressions*1000, works (result is 0)
			expectConvRate: true,
		},
		{
			name: "zero impressions",
			metrics: &entity.CalculatedMetrics{
				TotalSpend:       decimal.NewFromInt(100),
				TotalRevenue:     decimal.NewFromInt(500),
				TotalImpressions: 0,
				TotalClicks:      500,
				TotalConversions: 50,
			},
			expectROAS:     true,
			expectCTR:      false,
			expectCPC:      true,
			expectCPA:      true,
			expectCPM:      false,
			expectConvRate: true,
		},
		{
			name: "zero clicks",
			metrics: &entity.CalculatedMetrics{
				TotalSpend:       decimal.NewFromInt(100),
				TotalRevenue:     decimal.NewFromInt(500),
				TotalImpressions: 10000,
				TotalClicks:      0,
				TotalConversions: 50,
			},
			expectROAS:     true,
			expectCTR:      true, // CTR will be 0
			expectCPC:      false,
			expectCPA:      true,
			expectCPM:      true,
			expectConvRate: false,
		},
		{
			name: "zero conversions",
			metrics: &entity.CalculatedMetrics{
				TotalSpend:       decimal.NewFromInt(100),
				TotalRevenue:     decimal.NewFromInt(500),
				TotalImpressions: 10000,
				TotalClicks:      500,
				TotalConversions: 0,
			},
			expectROAS:     true,
			expectCTR:      true,
			expectCPC:      true,
			expectCPA:      false,
			expectCPM:      true,
			expectConvRate: true, // Will be 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.metrics.CalculateDerivedFields()

			if (tt.metrics.ROAS != nil) != tt.expectROAS {
				t.Errorf("ROAS presence: got %v, want %v", tt.metrics.ROAS != nil, tt.expectROAS)
			}
			if (tt.metrics.CTR != nil) != tt.expectCTR {
				t.Errorf("CTR presence: got %v, want %v", tt.metrics.CTR != nil, tt.expectCTR)
			}
			if (tt.metrics.CPC != nil) != tt.expectCPC {
				t.Errorf("CPC presence: got %v, want %v", tt.metrics.CPC != nil, tt.expectCPC)
			}
			if (tt.metrics.CPA != nil) != tt.expectCPA {
				t.Errorf("CPA presence: got %v, want %v", tt.metrics.CPA != nil, tt.expectCPA)
			}
			if (tt.metrics.CPM != nil) != tt.expectCPM {
				t.Errorf("CPM presence: got %v, want %v", tt.metrics.CPM != nil, tt.expectCPM)
			}
			if (tt.metrics.ConversionRate != nil) != tt.expectConvRate {
				t.Errorf("ConversionRate presence: got %v, want %v", tt.metrics.ConversionRate != nil, tt.expectConvRate)
			}
		})
	}
}

func TestDataQualityReport_CalculateCompleteness(t *testing.T) {
	tests := []struct {
		name                  string
		totalDaysRequested    int
		totalDaysWithData     int
		expectHasCompleteData bool
		expectCompleteness    float64
	}{
		{
			name:                  "full data",
			totalDaysRequested:    7,
			totalDaysWithData:     7,
			expectHasCompleteData: true,
			expectCompleteness:    100,
		},
		{
			name:                  "no data",
			totalDaysRequested:    7,
			totalDaysWithData:     0,
			expectHasCompleteData: false,
			expectCompleteness:    0,
		},
		{
			name:                  "partial data",
			totalDaysRequested:    10,
			totalDaysWithData:     5,
			expectHasCompleteData: false,
			expectCompleteness:    50,
		},
		{
			name:                  "zero days requested",
			totalDaysRequested:    0,
			totalDaysWithData:     0,
			expectHasCompleteData: false,
			expectCompleteness:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &entity.DataQualityReport{
				TotalDaysRequested: tt.totalDaysRequested,
				TotalDaysWithData:  tt.totalDaysWithData,
			}
			report.CalculateCompleteness()

			if report.HasCompleteData != tt.expectHasCompleteData {
				t.Errorf("HasCompleteData: got %v, want %v", report.HasCompleteData, tt.expectHasCompleteData)
			}
			if report.DataCompleteness != tt.expectCompleteness {
				t.Errorf("DataCompleteness: got %f, want %f", report.DataCompleteness, tt.expectCompleteness)
			}
		})
	}
}

func TestAnalyticsRequest_Validate(t *testing.T) {
	validOrgID := uuid.New()
	now := time.Now()

	tests := []struct {
		name    string
		req     entity.AnalyticsRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: entity.AnalyticsRequest{
				OrganizationID: validOrgID,
				DateRange: entity.DateRange{
					StartDate: now.AddDate(0, 0, -7),
					EndDate:   now,
				},
			},
			wantErr: false,
		},
		{
			name: "missing org ID",
			req: entity.AnalyticsRequest{
				DateRange: entity.DateRange{
					StartDate: now.AddDate(0, 0, -7),
					EndDate:   now,
				},
			},
			wantErr: true,
		},
		{
			name: "missing start date",
			req: entity.AnalyticsRequest{
				OrganizationID: validOrgID,
				DateRange: entity.DateRange{
					EndDate: now,
				},
			},
			wantErr: true,
		},
		{
			name: "start after end",
			req: entity.AnalyticsRequest{
				OrganizationID: validOrgID,
				DateRange: entity.DateRange{
					StartDate: now,
					EndDate:   now.AddDate(0, 0, -7),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid platform",
			req: entity.AnalyticsRequest{
				OrganizationID: validOrgID,
				DateRange: entity.DateRange{
					StartDate: now.AddDate(0, 0, -7),
					EndDate:   now,
				},
				Platforms: []entity.Platform{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "valid with all optional fields",
			req: entity.AnalyticsRequest{
				OrganizationID: validOrgID,
				DateRange: entity.DateRange{
					StartDate: now.AddDate(0, 0, -7),
					EndDate:   now,
				},
				Platforms:      []entity.Platform{entity.PlatformMeta, entity.PlatformTikTok},
				CampaignIDs:    []uuid.UUID{uuid.New()},
				TargetCurrency: "USD",
				IncludeDetails: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Currency Conversion Tests
// ============================================================================

func TestCalculateAnalytics_CurrencyConversion(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	// Create campaign
	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	// Create metric with USD currency
	metric := createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50)
	metric.Currency = "USD"
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{metric}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR", // Convert from USD to MYR
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Spend should be converted from USD to MYR (100 USD * 4.47 = 447 MYR approx)
	// The exact value depends on exchange rate
	if result.OverallMetrics.TotalSpend.IsZero() {
		t.Error("TotalSpend should not be zero after conversion")
	}

	// Currency should be MYR
	if result.OverallMetrics.Currency != "MYR" {
		t.Errorf("Currency: got %s, want MYR", result.OverallMetrics.Currency)
	}
}

func TestCalculateAnalytics_SameCurrencyNoConversion(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	metric := createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50)
	metric.Currency = "MYR"
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{metric}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR", // Same currency, no conversion needed
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	// Spend should remain 100 MYR
	expectedSpend := decimal.NewFromFloat(100)
	if !result.OverallMetrics.TotalSpend.Equal(expectedSpend) {
		t.Errorf("TotalSpend: got %s, want %s", result.OverallMetrics.TotalSpend.String(), expectedSpend.String())
	}
}

func TestService_GetSupportedCurrencies(t *testing.T) {
	service, _, _ := createTestService()

	currencies := service.GetSupportedCurrencies()

	if len(currencies) == 0 {
		t.Error("Should return supported currencies")
	}

	// Check that MYR is supported (base currency)
	found := false
	for _, c := range currencies {
		if c == "MYR" {
			found = true
			break
		}
	}
	if !found {
		t.Error("MYR should be in supported currencies")
	}
}

func TestService_IsSupportedCurrency(t *testing.T) {
	service, _, _ := createTestService()

	tests := []struct {
		currency string
		want     bool
	}{
		{"MYR", true},
		{"USD", true},
		{"SGD", true},
		{"EUR", true},
		{"XXX", false}, // Invalid currency
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.currency, func(t *testing.T) {
			got := service.IsSupportedCurrency(tt.currency)
			if got != tt.want {
				t.Errorf("IsSupportedCurrency(%s): got %v, want %v", tt.currency, got, tt.want)
			}
		})
	}
}

// ============================================================================
// ROAS, CPA, CTR Calculation Tests
// ============================================================================

func TestROASCalculation(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	// Spend: 100, Revenue: 500 -> ROAS = 5.0
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	if result.OverallMetrics.ROAS == nil {
		t.Fatal("ROAS should be calculated")
	}

	expectedROAS := 5.0 // 500 / 100
	if *result.OverallMetrics.ROAS != expectedROAS {
		t.Errorf("ROAS: got %f, want %f", *result.OverallMetrics.ROAS, expectedROAS)
	}
}

func TestCPACalculation(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	// Spend: 100, Conversions: 50 -> CPA = 2.0
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	if result.OverallMetrics.CPA == nil {
		t.Fatal("CPA should be calculated")
	}

	expectedCPA := decimal.NewFromFloat(2.0) // 100 / 50
	if !result.OverallMetrics.CPA.Equal(expectedCPA) {
		t.Errorf("CPA: got %s, want %s", result.OverallMetrics.CPA.String(), expectedCPA.String())
	}
}

func TestCTRCalculation(t *testing.T) {
	service, metricsRepo, campaignRepo := createTestService()
	ctx := context.Background()

	orgID := uuid.New()
	now := time.Now()

	campaign := createTestCampaign(orgID, entity.PlatformMeta)
	campaignRepo.campaigns = []entity.Campaign{campaign}

	// Clicks: 500, Impressions: 10000 -> CTR = 5.0%
	metricsRepo.campaignMetrics[campaign.ID] = []entity.CampaignMetricsDaily{
		createTestMetric(campaign.ID, orgID, entity.PlatformMeta, now, 100, 500, 10000, 500, 50),
	}

	req := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: now.AddDate(0, 0, -1),
			EndDate:   now,
		},
		TargetCurrency: "MYR",
	}

	result, err := service.CalculateAnalytics(ctx, req)
	if err != nil {
		t.Fatalf("CalculateAnalytics failed: %v", err)
	}

	if result.OverallMetrics.CTR == nil {
		t.Fatal("CTR should be calculated")
	}

	expectedCTR := 5.0 // (500 / 10000) * 100
	if *result.OverallMetrics.CTR != expectedCTR {
		t.Errorf("CTR: got %f, want %f", *result.OverallMetrics.CTR, expectedCTR)
	}
}
