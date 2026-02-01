package analytics

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/pkg/currency"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Service handles analytics and metrics aggregation
type Service struct {
	metricsRepo       repository.MetricsRepository
	campaignRepo      repository.CampaignRepository
	currencyConverter *currency.Converter
}

// NewService creates a new analytics service
func NewService(
	metricsRepo repository.MetricsRepository,
	campaignRepo repository.CampaignRepository,
) *Service {
	return &Service{
		metricsRepo:       metricsRepo,
		campaignRepo:      campaignRepo,
		currencyConverter: currency.NewDefaultConverter(),
	}
}

// NewServiceWithCurrency creates a new analytics service with custom currency converter
func NewServiceWithCurrency(
	metricsRepo repository.MetricsRepository,
	campaignRepo repository.CampaignRepository,
	converter *currency.Converter,
) *Service {
	if converter == nil {
		converter = currency.NewDefaultConverter()
	}
	return &Service{
		metricsRepo:       metricsRepo,
		campaignRepo:      campaignRepo,
		currencyConverter: converter,
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

// ============================================================================
// CalculateAnalytics - Main Analytics Calculation Method
// ============================================================================

// CalculateAnalytics performs comprehensive metrics calculation based on the request
// Returns ROAS, CPA, CTR per platform and combined with zero-division protection
func (s *Service) CalculateAnalytics(ctx context.Context, req entity.AnalyticsRequest) (*entity.AnalyticsResponse, error) {
	// 1. Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. Build metrics filter from request
	filter := entity.MetricsFilter{
		OrganizationID: req.OrganizationID,
		DateRange:      req.DateRange,
		Platforms:      req.Platforms,
		CampaignIDs:    req.CampaignIDs,
	}

	// 3. Get daily metrics based on filter
	dailyMetrics, err := s.getDailyMetricsForFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	// 4. Get campaign count
	campaignFilter := entity.CampaignFilter{
		OrganizationID: req.OrganizationID,
		Platforms:      req.Platforms,
	}
	campaigns, _, _ := s.campaignRepo.List(ctx, campaignFilter)
	campaignCount := len(campaigns)

	// 5. Build response
	response := &entity.AnalyticsResponse{
		DateRange:       req.DateRange,
		TargetCurrency:  req.TargetCurrency,
		GeneratedAt:     time.Now(),
		PlatformMetrics: make(map[entity.Platform]*entity.CalculatedMetrics),
	}

	// 6. Calculate overall metrics (all platforms combined)
	response.OverallMetrics = s.calculateMetricsFromDaily(dailyMetrics, req.TargetCurrency, nil, campaignCount)

	// 7. Calculate per-platform metrics
	platformGroups := s.groupMetricsByPlatform(dailyMetrics)
	platformCampaignCounts := s.getCampaignCountsByPlatform(campaigns)

	for platform, metrics := range platformGroups {
		count := platformCampaignCounts[platform]
		platformCopy := platform // Create copy for pointer
		calculated := s.calculateMetricsFromDaily(metrics, req.TargetCurrency, &platformCopy, count)
		response.PlatformMetrics[platform] = calculated
	}

	// 8. Generate platform comparison
	if len(response.PlatformMetrics) > 1 {
		response.Comparison = s.generatePlatformComparison(response.PlatformMetrics)
	}

	// 9. Include daily trend if requested
	if req.IncludeDetails {
		filter.GroupBy = "day"
		dailyTrend, err := s.metricsRepo.GetDailyTrend(ctx, filter)
		if err == nil {
			response.DailyTrend = dailyTrend
		}
	}

	// 10. Build data quality report
	response.DataQuality = s.buildDataQualityReport(dailyMetrics, req)

	return response, nil
}

// getDailyMetricsForFilter retrieves daily metrics based on the filter
// It handles campaign-specific queries vs organization-wide queries
func (s *Service) getDailyMetricsForFilter(ctx context.Context, filter entity.MetricsFilter) ([]entity.CampaignMetricsDaily, error) {
	var allMetrics []entity.CampaignMetricsDaily

	// If specific campaigns requested, get metrics for each
	if len(filter.CampaignIDs) > 0 {
		for _, campaignID := range filter.CampaignIDs {
			metrics, err := s.metricsRepo.GetCampaignMetrics(ctx, campaignID, filter.DateRange)
			if err != nil {
				continue // Skip campaigns that fail, include in data quality report
			}
			// Filter by platform if specified
			for _, m := range metrics {
				if len(filter.Platforms) == 0 || containsPlatform(filter.Platforms, m.Platform) {
					allMetrics = append(allMetrics, m)
				}
			}
		}
	} else {
		// Get all campaigns for organization
		campaignFilter := entity.CampaignFilter{
			OrganizationID: filter.OrganizationID,
			Platforms:      filter.Platforms,
		}
		campaigns, _, err := s.campaignRepo.List(ctx, campaignFilter)
		if err != nil {
			return nil, err
		}

		for _, campaign := range campaigns {
			metrics, err := s.metricsRepo.GetCampaignMetrics(ctx, campaign.ID, filter.DateRange)
			if err != nil {
				continue // Skip campaigns that fail
			}
			allMetrics = append(allMetrics, metrics...)
		}
	}

	return allMetrics, nil
}

// calculateMetricsFromDaily calculates aggregated metrics from daily data
// with zero-division protection for all ratio metrics
func (s *Service) calculateMetricsFromDaily(
	dailyMetrics []entity.CampaignMetricsDaily,
	targetCurrency string,
	platform *entity.Platform,
	campaignCount int,
) *entity.CalculatedMetrics {
	metrics := &entity.CalculatedMetrics{
		Platform:      platform,
		Currency:      targetCurrency,
		CampaignCount: campaignCount,
	}

	if len(dailyMetrics) == 0 {
		metrics.CalculateDerivedFields()
		return metrics
	}

	// Track unique campaigns and date range
	campaignsSeen := make(map[uuid.UUID]bool)
	var firstDate, lastDate time.Time

	// Aggregate all daily metrics
	for _, m := range dailyMetrics {
		// Currency conversion if needed
		spend := s.convertCurrency(m.Spend, m.Currency, targetCurrency)
		revenue := s.convertCurrency(m.ConversionValue, m.Currency, targetCurrency)

		metrics.TotalSpend = metrics.TotalSpend.Add(spend)
		metrics.TotalRevenue = metrics.TotalRevenue.Add(revenue)
		metrics.TotalImpressions += m.Impressions
		metrics.TotalClicks += m.Clicks
		metrics.TotalConversions += m.Conversions
		metrics.TotalReach += m.Reach
		metrics.TotalLikes += m.Likes
		metrics.TotalComments += m.Comments
		metrics.TotalShares += m.Shares

		// Track campaigns
		campaignsSeen[m.CampaignID] = true

		// Track date range
		if firstDate.IsZero() || m.MetricDate.Before(firstDate) {
			firstDate = m.MetricDate
		}
		if lastDate.IsZero() || m.MetricDate.After(lastDate) {
			lastDate = m.MetricDate
		}
	}

	// Set date range
	if !firstDate.IsZero() {
		metrics.FirstDate = &firstDate
	}
	if !lastDate.IsZero() {
		metrics.LastDate = &lastDate
	}

	// Update campaign count from actual data if not provided
	if campaignCount == 0 {
		metrics.CampaignCount = len(campaignsSeen)
	}

	// Calculate derived fields with zero-division protection
	metrics.CalculateDerivedFields()

	return metrics
}

// groupMetricsByPlatform groups metrics by platform
func (s *Service) groupMetricsByPlatform(metrics []entity.CampaignMetricsDaily) map[entity.Platform][]entity.CampaignMetricsDaily {
	grouped := make(map[entity.Platform][]entity.CampaignMetricsDaily)
	for _, m := range metrics {
		grouped[m.Platform] = append(grouped[m.Platform], m)
	}
	return grouped
}

// getCampaignCountsByPlatform counts campaigns per platform
func (s *Service) getCampaignCountsByPlatform(campaigns []entity.Campaign) map[entity.Platform]int {
	counts := make(map[entity.Platform]int)
	for _, c := range campaigns {
		counts[c.Platform]++
	}
	return counts
}

// convertCurrency converts an amount to target currency using the currency converter
func (s *Service) convertCurrency(amount decimal.Decimal, from, to string) decimal.Decimal {
	if from == to || from == "" || to == "" {
		return amount
	}

	if amount.IsZero() {
		return amount
	}

	// Use the currency converter
	converted, err := s.currencyConverter.Convert(amount, from, to)
	if err != nil {
		// On error, return original amount (with warning in production)
		return amount
	}
	return converted
}

// GetSupportedCurrencies returns list of supported currencies for conversion
func (s *Service) GetSupportedCurrencies() []string {
	return s.currencyConverter.GetSupportedCurrencies()
}

// IsSupportedCurrency checks if a currency is supported
func (s *Service) IsSupportedCurrency(code string) bool {
	return s.currencyConverter.IsSupportedCurrency(code)
}

// generatePlatformComparison generates comparison rankings between platforms
func (s *Service) generatePlatformComparison(platformMetrics map[entity.Platform]*entity.CalculatedMetrics) *entity.PlatformComparison {
	comparison := &entity.PlatformComparison{
		PlatformCount: len(platformMetrics),
	}

	// Convert to slice for sorting
	type platformEntry struct {
		platform entity.Platform
		metrics  *entity.CalculatedMetrics
	}
	entries := make([]platformEntry, 0, len(platformMetrics))
	for p, m := range platformMetrics {
		entries = append(entries, platformEntry{platform: p, metrics: m})
	}

	// Best ROAS (highest)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].metrics.ROAS == nil {
			return false
		}
		if entries[j].metrics.ROAS == nil {
			return true
		}
		return *entries[i].metrics.ROAS > *entries[j].metrics.ROAS
	})
	if len(entries) > 0 && entries[0].metrics.ROAS != nil {
		comparison.BestROAS = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        *entries[0].metrics.ROAS,
			DisplayValue: fmt.Sprintf("%.2fx", *entries[0].metrics.ROAS),
		}
	}

	// Highest CTR
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].metrics.CTR == nil {
			return false
		}
		if entries[j].metrics.CTR == nil {
			return true
		}
		return *entries[i].metrics.CTR > *entries[j].metrics.CTR
	})
	if len(entries) > 0 && entries[0].metrics.CTR != nil {
		comparison.HighestCTR = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        *entries[0].metrics.CTR,
			DisplayValue: fmt.Sprintf("%.2f%%", *entries[0].metrics.CTR),
		}
	}

	// Lowest CPA (best)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].metrics.CPA == nil {
			return false
		}
		if entries[j].metrics.CPA == nil {
			return true
		}
		return entries[i].metrics.CPA.LessThan(*entries[j].metrics.CPA)
	})
	if len(entries) > 0 && entries[0].metrics.CPA != nil {
		cpaFloat, _ := entries[0].metrics.CPA.Float64()
		comparison.LowestCPA = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        cpaFloat,
			DisplayValue: fmt.Sprintf("%s %s", entries[0].metrics.Currency, entries[0].metrics.CPA.StringFixed(2)),
		}
	}

	// Lowest CPC (best)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].metrics.CPC == nil {
			return false
		}
		if entries[j].metrics.CPC == nil {
			return true
		}
		return entries[i].metrics.CPC.LessThan(*entries[j].metrics.CPC)
	})
	if len(entries) > 0 && entries[0].metrics.CPC != nil {
		cpcFloat, _ := entries[0].metrics.CPC.Float64()
		comparison.LowestCPC = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        cpcFloat,
			DisplayValue: fmt.Sprintf("%s %s", entries[0].metrics.Currency, entries[0].metrics.CPC.StringFixed(2)),
		}
	}

	// Most Spend
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].metrics.TotalSpend.GreaterThan(entries[j].metrics.TotalSpend)
	})
	if len(entries) > 0 {
		spendFloat, _ := entries[0].metrics.TotalSpend.Float64()
		comparison.MostSpend = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        spendFloat,
			DisplayValue: fmt.Sprintf("%s %s", entries[0].metrics.Currency, entries[0].metrics.TotalSpend.StringFixed(2)),
		}
	}

	// Most Revenue
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].metrics.TotalRevenue.GreaterThan(entries[j].metrics.TotalRevenue)
	})
	if len(entries) > 0 {
		revFloat, _ := entries[0].metrics.TotalRevenue.Float64()
		comparison.MostRevenue = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        revFloat,
			DisplayValue: fmt.Sprintf("%s %s", entries[0].metrics.Currency, entries[0].metrics.TotalRevenue.StringFixed(2)),
		}
	}

	// Most Clicks
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].metrics.TotalClicks > entries[j].metrics.TotalClicks
	})
	if len(entries) > 0 {
		comparison.MostClicks = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        float64(entries[0].metrics.TotalClicks),
			DisplayValue: formatNumber(entries[0].metrics.TotalClicks),
		}
	}

	// Best Conversion Rate
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].metrics.ConversionRate == nil {
			return false
		}
		if entries[j].metrics.ConversionRate == nil {
			return true
		}
		return *entries[i].metrics.ConversionRate > *entries[j].metrics.ConversionRate
	})
	if len(entries) > 0 && entries[0].metrics.ConversionRate != nil {
		comparison.BestConvRate = &entity.PlatformRank{
			Platform:     entries[0].platform,
			Value:        *entries[0].metrics.ConversionRate,
			DisplayValue: fmt.Sprintf("%.2f%%", *entries[0].metrics.ConversionRate),
		}
	}

	return comparison
}

// buildDataQualityReport creates a data quality report
func (s *Service) buildDataQualityReport(metrics []entity.CampaignMetricsDaily, req entity.AnalyticsRequest) *entity.DataQualityReport {
	report := &entity.DataQualityReport{
		TotalDaysRequested: int(req.DateRange.EndDate.Sub(req.DateRange.StartDate).Hours()/24) + 1,
	}

	if len(metrics) == 0 {
		report.AddWarning("No metrics data found for the specified date range")
		report.CalculateCompleteness()
		return report
	}

	// Track dates with data
	datesWithData := make(map[string]bool)
	platformsWithData := make(map[entity.Platform]bool)
	campaignsWithData := make(map[uuid.UUID]bool)
	var latestSync time.Time

	for _, m := range metrics {
		dateKey := m.MetricDate.Format("2006-01-02")
		datesWithData[dateKey] = true
		platformsWithData[m.Platform] = true
		campaignsWithData[m.CampaignID] = true

		if m.LastSyncedAt != nil && m.LastSyncedAt.After(latestSync) {
			latestSync = *m.LastSyncedAt
		}
	}

	report.TotalDaysWithData = len(datesWithData)

	// Check for missing dates
	for d := req.DateRange.StartDate; !d.After(req.DateRange.EndDate); d = d.AddDate(0, 0, 1) {
		dateKey := d.Format("2006-01-02")
		if !datesWithData[dateKey] {
			report.MissingDates = append(report.MissingDates, dateKey)
		}
	}

	// Check for platforms with no data (if specific platforms requested)
	if len(req.Platforms) > 0 {
		for _, p := range req.Platforms {
			if !platformsWithData[p] {
				report.PlatformsWithNoData = append(report.PlatformsWithNoData, p)
			}
		}
	}

	// Set last sync time
	if !latestSync.IsZero() {
		report.LastSyncTime = &latestSync
	}

	// Generate warnings
	if len(report.MissingDates) > 0 {
		report.AddWarning(fmt.Sprintf("%d days have missing data", len(report.MissingDates)))
	}
	if len(report.PlatformsWithNoData) > 0 {
		report.AddWarning(fmt.Sprintf("No data for platforms: %v", report.PlatformsWithNoData))
	}

	// Calculate completeness
	report.CalculateCompleteness()

	return report
}

// containsPlatform checks if a platform is in the list
func containsPlatform(platforms []entity.Platform, p entity.Platform) bool {
	for _, platform := range platforms {
		if platform == p {
			return true
		}
	}
	return false
}
