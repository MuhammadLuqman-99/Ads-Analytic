package analytics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Service handles analytics and metrics aggregation
type Service struct {
	metricsRepo  repository.MetricsRepository
	campaignRepo repository.CampaignRepository
}

// NewService creates a new analytics service
func NewService(
	metricsRepo repository.MetricsRepository,
	campaignRepo repository.CampaignRepository,
) *Service {
	return &Service{
		metricsRepo:  metricsRepo,
		campaignRepo: campaignRepo,
	}
}

// DashboardMetrics represents the main dashboard metrics
type DashboardMetrics struct {
	TotalSpend        decimal.Decimal                 `json:"total_spend"`
	TotalImpressions  int64                           `json:"total_impressions"`
	TotalClicks       int64                           `json:"total_clicks"`
	TotalConversions  int64                           `json:"total_conversions"`
	TotalRevenue      decimal.Decimal                 `json:"total_revenue"`
	AverageCTR        float64                         `json:"average_ctr"`
	AverageCPC        decimal.Decimal                 `json:"average_cpc"`
	AverageCPA        decimal.Decimal                 `json:"average_cpa"`
	OverallROAS       float64                         `json:"overall_roas"`
	ActiveCampaigns   int                             `json:"active_campaigns"`
	PlatformBreakdown []entity.PlatformMetricsSummary `json:"platform_breakdown"`
	DailyTrend        []entity.DailyMetricsTrend      `json:"daily_trend"`
	TopCampaigns      []entity.TopPerformer           `json:"top_campaigns"`
	Comparison        *entity.MetricsComparison       `json:"comparison,omitempty"`
}

// GetDashboardMetrics returns comprehensive dashboard metrics
func (s *Service) GetDashboardMetrics(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) (*DashboardMetrics, error) {
	filter := entity.MetricsFilter{
		OrganizationID: orgID,
		DateRange:      dateRange,
	}

	// Get aggregated metrics
	aggregated, err := s.metricsRepo.GetAggregatedMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get platform breakdown
	platformBreakdown, err := s.metricsRepo.GetMetricsByPlatform(ctx, orgID, dateRange)
	if err != nil {
		return nil, err
	}

	// Get daily trend
	filter.GroupBy = "day"
	dailyTrend, err := s.metricsRepo.GetDailyTrend(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get top performing campaigns
	topCampaigns, err := s.metricsRepo.GetTopPerformingCampaigns(ctx, orgID, dateRange, 10)
	if err != nil {
		return nil, err
	}

	// Get active campaign count
	campaignFilter := entity.CampaignFilter{
		OrganizationID: orgID,
		Statuses:       []entity.CampaignStatus{entity.CampaignStatusActive},
	}
	campaigns, _, err := s.campaignRepo.List(ctx, campaignFilter)
	if err != nil {
		return nil, err
	}

	// Calculate period comparison
	comparison, err := s.getPeriodComparison(ctx, orgID, dateRange)
	if err != nil {
		// Non-fatal, just log
		comparison = nil
	}

	dashboard := &DashboardMetrics{
		TotalSpend:        aggregated.TotalSpend,
		TotalImpressions:  aggregated.TotalImpressions,
		TotalClicks:       aggregated.TotalClicks,
		TotalConversions:  aggregated.TotalConversions,
		TotalRevenue:      aggregated.TotalRevenue,
		AverageCTR:        aggregated.AverageCTR,
		AverageCPC:        aggregated.AverageCPC,
		AverageCPA:        aggregated.AverageCPA,
		OverallROAS:       aggregated.OverallROAS,
		ActiveCampaigns:   len(campaigns),
		PlatformBreakdown: platformBreakdown,
		DailyTrend:        dailyTrend,
		TopCampaigns:      topCampaigns,
		Comparison:        comparison,
	}

	return dashboard, nil
}

// getPeriodComparison calculates metrics comparison with previous period
func (s *Service) getPeriodComparison(ctx context.Context, orgID uuid.UUID, currentRange entity.DateRange) (*entity.MetricsComparison, error) {
	// Calculate previous period
	duration := currentRange.EndDate.Sub(currentRange.StartDate)
	previousRange := entity.DateRange{
		StartDate: currentRange.StartDate.Add(-duration),
		EndDate:   currentRange.StartDate,
	}

	// Get current period metrics
	currentFilter := entity.MetricsFilter{
		OrganizationID: orgID,
		DateRange:      currentRange,
	}
	currentMetrics, err := s.metricsRepo.GetAggregatedMetrics(ctx, currentFilter)
	if err != nil {
		return nil, err
	}

	// Get previous period metrics
	previousFilter := entity.MetricsFilter{
		OrganizationID: orgID,
		DateRange:      previousRange,
	}
	previousMetrics, err := s.metricsRepo.GetAggregatedMetrics(ctx, previousFilter)
	if err != nil {
		return nil, err
	}

	comparison := &entity.MetricsComparison{
		CurrentPeriod:  *currentMetrics,
		PreviousPeriod: *previousMetrics,
	}
	comparison.CalculateComparison()

	return comparison, nil
}

// CampaignPerformance represents detailed campaign performance
type CampaignPerformance struct {
	Campaign     entity.Campaign               `json:"campaign"`
	Metrics      entity.AggregatedMetrics      `json:"metrics"`
	DailyMetrics []entity.CampaignMetricsDaily `json:"daily_metrics"`
	AdSets       []AdSetPerformance            `json:"ad_sets,omitempty"`
}

// AdSetPerformance represents ad set performance
type AdSetPerformance struct {
	AdSet   entity.AdSet             `json:"ad_set"`
	Metrics entity.AggregatedMetrics `json:"metrics"`
}

// GetCampaignPerformance returns detailed performance for a campaign
func (s *Service) GetCampaignPerformance(ctx context.Context, campaignID uuid.UUID, dateRange entity.DateRange) (*CampaignPerformance, error) {
	campaign, err := s.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	// Get campaign metrics
	dailyMetrics, err := s.metricsRepo.GetCampaignMetrics(ctx, campaignID, dateRange)
	if err != nil {
		return nil, err
	}

	// Aggregate metrics
	aggregated := s.aggregateCampaignMetrics(dailyMetrics)

	return &CampaignPerformance{
		Campaign:     *campaign,
		Metrics:      aggregated,
		DailyMetrics: dailyMetrics,
	}, nil
}

// aggregateCampaignMetrics aggregates daily metrics into totals
func (s *Service) aggregateCampaignMetrics(dailyMetrics []entity.CampaignMetricsDaily) entity.AggregatedMetrics {
	aggregated := entity.AggregatedMetrics{}

	for _, m := range dailyMetrics {
		aggregated.TotalSpend = aggregated.TotalSpend.Add(m.Spend)
		aggregated.TotalImpressions += m.Impressions
		aggregated.TotalClicks += m.Clicks
		aggregated.TotalConversions += m.Conversions
		aggregated.TotalRevenue = aggregated.TotalRevenue.Add(m.ConversionValue)
		aggregated.Currency = m.Currency
	}

	// Calculate averages
	if aggregated.TotalImpressions > 0 {
		aggregated.AverageCTR = float64(aggregated.TotalClicks) / float64(aggregated.TotalImpressions) * 100
	}

	if aggregated.TotalClicks > 0 {
		aggregated.AverageCPC = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalClicks))
	}

	if aggregated.TotalImpressions > 0 {
		aggregated.AverageCPM = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalImpressions)).Mul(decimal.NewFromInt(1000))
	}

	if aggregated.TotalConversions > 0 {
		aggregated.AverageCPA = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalConversions))
	}

	if !aggregated.TotalSpend.IsZero() {
		aggregated.OverallROAS, _ = aggregated.TotalRevenue.Div(aggregated.TotalSpend).Float64()
	}

	return aggregated
}

// CrossPlatformReport represents a cross-platform comparison report
type CrossPlatformReport struct {
	DateRange    entity.DateRange                               `json:"date_range"`
	TotalMetrics entity.AggregatedMetrics                       `json:"total_metrics"`
	ByPlatform   []entity.PlatformMetricsSummary                `json:"by_platform"`
	Trends       map[entity.Platform][]entity.DailyMetricsTrend `json:"trends"`
	Insights     []string                                       `json:"insights"`
}

// GetCrossPlatformReport generates a cross-platform comparison report
func (s *Service) GetCrossPlatformReport(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) (*CrossPlatformReport, error) {
	// Get platform breakdown
	platformMetrics, err := s.metricsRepo.GetMetricsByPlatform(ctx, orgID, dateRange)
	if err != nil {
		return nil, err
	}

	// Get total aggregated metrics
	filter := entity.MetricsFilter{
		OrganizationID: orgID,
		DateRange:      dateRange,
	}
	totalMetrics, err := s.metricsRepo.GetAggregatedMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get trends by platform
	trends := make(map[entity.Platform][]entity.DailyMetricsTrend)
	platforms := []entity.Platform{entity.PlatformMeta, entity.PlatformTikTok, entity.PlatformShopee}

	for _, platform := range platforms {
		filter.Platforms = []entity.Platform{platform}
		filter.GroupBy = "day"
		trend, err := s.metricsRepo.GetDailyTrend(ctx, filter)
		if err != nil {
			continue
		}
		trends[platform] = trend
	}

	// Generate insights
	insights := s.generateInsights(platformMetrics, *totalMetrics)

	return &CrossPlatformReport{
		DateRange:    dateRange,
		TotalMetrics: *totalMetrics,
		ByPlatform:   platformMetrics,
		Trends:       trends,
		Insights:     insights,
	}, nil
}

// generateInsights generates automatic insights from metrics
func (s *Service) generateInsights(platformMetrics []entity.PlatformMetricsSummary, total entity.AggregatedMetrics) []string {
	insights := make([]string, 0)

	if len(platformMetrics) == 0 {
		return insights
	}

	// Find best performing platform by ROAS
	sort.Slice(platformMetrics, func(i, j int) bool {
		return platformMetrics[i].ROAS > platformMetrics[j].ROAS
	})
	bestROAS := platformMetrics[0]
	if bestROAS.ROAS > 0 {
		insights = append(insights,
			fmt.Sprintf("%s has the highest ROAS at %.2fx", bestROAS.Platform, bestROAS.ROAS))
	}

	// Find most cost-effective platform by CPC
	sort.Slice(platformMetrics, func(i, j int) bool {
		return platformMetrics[i].CPC.LessThan(platformMetrics[j].CPC)
	})
	lowestCPC := platformMetrics[0]
	if !lowestCPC.CPC.IsZero() {
		insights = append(insights,
			fmt.Sprintf("%s has the lowest CPC at %s", lowestCPC.Platform, lowestCPC.CPC.StringFixed(2)))
	}

	// Find highest CTR platform
	sort.Slice(platformMetrics, func(i, j int) bool {
		return platformMetrics[i].CTR > platformMetrics[j].CTR
	})
	highestCTR := platformMetrics[0]
	if highestCTR.CTR > 0 {
		insights = append(insights,
			fmt.Sprintf("%s has the highest CTR at %.2f%%", highestCTR.Platform, highestCTR.CTR))
	}

	// Overall performance insight
	if total.OverallROAS >= 3 {
		insights = append(insights, "Overall ad performance is excellent with ROAS above 3x")
	} else if total.OverallROAS >= 2 {
		insights = append(insights, "Overall ad performance is good with ROAS above 2x")
	} else if total.OverallROAS >= 1 {
		insights = append(insights, "Overall ad performance is break-even, consider optimization")
	} else if total.OverallROAS > 0 {
		insights = append(insights, "Overall ROAS is below 1, ads are not profitable")
	}

	return insights
}

// ReportExport represents an exportable report
type ReportExport struct {
	GeneratedAt     time.Time                       `json:"generated_at"`
	OrganizationID  uuid.UUID                       `json:"organization_id"`
	DateRange       entity.DateRange                `json:"date_range"`
	Summary         entity.AggregatedMetrics        `json:"summary"`
	PlatformMetrics []entity.PlatformMetricsSummary `json:"platform_metrics"`
	CampaignMetrics []CampaignReportRow             `json:"campaign_metrics"`
}

// CampaignReportRow represents a row in the campaign report
type CampaignReportRow struct {
	CampaignID   uuid.UUID       `json:"campaign_id"`
	CampaignName string          `json:"campaign_name"`
	Platform     entity.Platform `json:"platform"`
	Status       string          `json:"status"`
	Spend        decimal.Decimal `json:"spend"`
	Impressions  int64           `json:"impressions"`
	Clicks       int64           `json:"clicks"`
	CTR          float64         `json:"ctr"`
	CPC          decimal.Decimal `json:"cpc"`
	Conversions  int64           `json:"conversions"`
	Revenue      decimal.Decimal `json:"revenue"`
	ROAS         float64         `json:"roas"`
}

// GenerateReport generates an exportable report
func (s *Service) GenerateReport(ctx context.Context, orgID uuid.UUID, dateRange entity.DateRange) (*ReportExport, error) {
	// Get summary metrics
	filter := entity.MetricsFilter{
		OrganizationID: orgID,
		DateRange:      dateRange,
	}
	summary, err := s.metricsRepo.GetAggregatedMetrics(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Get platform breakdown
	platformMetrics, err := s.metricsRepo.GetMetricsByPlatform(ctx, orgID, dateRange)
	if err != nil {
		return nil, err
	}

	// Get campaign summaries
	campaignFilter := entity.CampaignFilter{
		OrganizationID: orgID,
		DateRange:      &dateRange,
	}
	summaries, err := s.campaignRepo.GetSummaries(ctx, campaignFilter)
	if err != nil {
		return nil, err
	}

	// Convert to report rows
	campaignMetrics := make([]CampaignReportRow, len(summaries))
	for i, s := range summaries {
		campaignMetrics[i] = CampaignReportRow{
			CampaignID:   s.ID,
			CampaignName: s.Name,
			Platform:     s.Platform,
			Status:       string(s.Status),
			Spend:        s.TotalSpend,
			Impressions:  s.Impressions,
			Clicks:       s.Clicks,
			CTR:          s.CTR,
			CPC:          s.CPC,
			Conversions:  s.Conversions,
			Revenue:      decimal.Decimal{}, // Would need to be calculated
			ROAS:         s.ROAS,
		}
	}

	return &ReportExport{
		GeneratedAt:     time.Now(),
		OrganizationID:  orgID,
		DateRange:       dateRange,
		Summary:         *summary,
		PlatformMetrics: platformMetrics,
		CampaignMetrics: campaignMetrics,
	}, nil
}

// Helper to format numbers
func formatNumber(n int64) string {
	return fmt.Sprintf("%d", n)
}
