package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/pkg/currency"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Unified Metrics Types
// ============================================================================

// UnifiedMetrics represents normalized metrics from all platforms
type UnifiedMetrics struct {
	ID           uuid.UUID       `json:"id"`
	Date         time.Time       `json:"date"`
	Platform     entity.Platform `json:"platform"`
	CampaignID   uuid.UUID       `json:"campaign_id"`
	CampaignName string          `json:"campaign_name"`
	AdAccountID  uuid.UUID       `json:"ad_account_id,omitempty"`

	// Core metrics
	Spend       decimal.Decimal `json:"spend"`
	Impressions int64           `json:"impressions"`
	Clicks      int64           `json:"clicks"`
	Conversions int64           `json:"conversions"`
	Revenue     decimal.Decimal `json:"revenue"`

	// Calculated metrics (with zero-division protection)
	ROAS *float64         `json:"roas,omitempty"`
	CTR  *float64         `json:"ctr,omitempty"`
	CPC  *decimal.Decimal `json:"cpc,omitempty"`
	CPA  *decimal.Decimal `json:"cpa,omitempty"`
	CPM  *decimal.Decimal `json:"cpm,omitempty"`

	// Currency
	Currency string `json:"currency"`

	// Metadata
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

// CalculateDerived calculates ROAS, CTR, CPC, CPA, CPM with zero-division protection
func (u *UnifiedMetrics) CalculateDerived() {
	// ROAS = Revenue / Spend
	if u.Spend.IsPositive() {
		roas, _ := u.Revenue.Div(u.Spend).Float64()
		u.ROAS = &roas
	}

	// CTR = Clicks / Impressions * 100
	if u.Impressions > 0 {
		ctr := float64(u.Clicks) / float64(u.Impressions) * 100
		u.CTR = &ctr
	}

	// CPC = Spend / Clicks
	if u.Clicks > 0 {
		cpc := u.Spend.Div(decimal.NewFromInt(u.Clicks))
		u.CPC = &cpc
	}

	// CPA = Spend / Conversions
	if u.Conversions > 0 {
		cpa := u.Spend.Div(decimal.NewFromInt(u.Conversions))
		u.CPA = &cpa
	}

	// CPM = Spend / Impressions * 1000
	if u.Impressions > 0 {
		cpm := u.Spend.Div(decimal.NewFromInt(u.Impressions)).Mul(decimal.NewFromInt(1000))
		u.CPM = &cpm
	}
}

// ============================================================================
// Aggregation Types
// ============================================================================

// AggregationLevel defines the level of aggregation
type AggregationLevel string

const (
	AggregationByPlatform AggregationLevel = "platform"
	AggregationByDate     AggregationLevel = "date"
	AggregationByCampaign AggregationLevel = "campaign"
	AggregationTotal      AggregationLevel = "total"
)

// AggregatedResult represents aggregated metrics at any level
type AggregatedResult struct {
	Key       string          `json:"key"`        // Platform name, date string, campaign ID, or "total"
	Level     AggregationLevel `json:"level"`
	Platform  *entity.Platform `json:"platform,omitempty"`
	Date      *time.Time       `json:"date,omitempty"`
	CampaignID *uuid.UUID      `json:"campaign_id,omitempty"`

	// Aggregated metrics
	TotalSpend       decimal.Decimal `json:"total_spend"`
	TotalImpressions int64           `json:"total_impressions"`
	TotalClicks      int64           `json:"total_clicks"`
	TotalConversions int64           `json:"total_conversions"`
	TotalRevenue     decimal.Decimal `json:"total_revenue"`

	// Calculated metrics
	ROAS *float64         `json:"roas,omitempty"`
	CTR  *float64         `json:"ctr,omitempty"`
	CPC  *decimal.Decimal `json:"cpc,omitempty"`
	CPA  *decimal.Decimal `json:"cpa,omitempty"`
	CPM  *decimal.Decimal `json:"cpm,omitempty"`

	// Counts
	CampaignCount int `json:"campaign_count"`
	DayCount      int `json:"day_count"`

	Currency string `json:"currency"`
}

// CalculateDerived calculates derived metrics
func (a *AggregatedResult) CalculateDerived() {
	if a.TotalSpend.IsPositive() {
		roas, _ := a.TotalRevenue.Div(a.TotalSpend).Float64()
		a.ROAS = &roas
	}

	if a.TotalImpressions > 0 {
		ctr := float64(a.TotalClicks) / float64(a.TotalImpressions) * 100
		a.CTR = &ctr
		cpm := a.TotalSpend.Div(decimal.NewFromInt(a.TotalImpressions)).Mul(decimal.NewFromInt(1000))
		a.CPM = &cpm
	}

	if a.TotalClicks > 0 {
		cpc := a.TotalSpend.Div(decimal.NewFromInt(a.TotalClicks))
		a.CPC = &cpc
	}

	if a.TotalConversions > 0 {
		cpa := a.TotalSpend.Div(decimal.NewFromInt(a.TotalConversions))
		a.CPA = &cpa
	}
}

// ============================================================================
// Comparison Types
// ============================================================================

// PeriodComparisonResult represents comparison between two periods
type PeriodComparisonResult struct {
	CurrentPeriod  AggregatedResult `json:"current_period"`
	PreviousPeriod AggregatedResult `json:"previous_period"`

	// Changes (percentage)
	SpendChange       *float64 `json:"spend_change_percent,omitempty"`
	ImpressionsChange *float64 `json:"impressions_change_percent,omitempty"`
	ClicksChange      *float64 `json:"clicks_change_percent,omitempty"`
	ConversionsChange *float64 `json:"conversions_change_percent,omitempty"`
	RevenueChange     *float64 `json:"revenue_change_percent,omitempty"`
	ROASChange        *float64 `json:"roas_change_percent,omitempty"`
	CTRChange         *float64 `json:"ctr_change_percent,omitempty"`
}

// CalculateChanges calculates percentage changes between periods
func (p *PeriodComparisonResult) CalculateChanges() {
	// Spend change
	if !p.PreviousPeriod.TotalSpend.IsZero() {
		change, _ := p.CurrentPeriod.TotalSpend.Sub(p.PreviousPeriod.TotalSpend).
			Div(p.PreviousPeriod.TotalSpend).Mul(decimal.NewFromInt(100)).Float64()
		p.SpendChange = &change
	}

	// Impressions change
	if p.PreviousPeriod.TotalImpressions > 0 {
		change := float64(p.CurrentPeriod.TotalImpressions-p.PreviousPeriod.TotalImpressions) /
			float64(p.PreviousPeriod.TotalImpressions) * 100
		p.ImpressionsChange = &change
	}

	// Clicks change
	if p.PreviousPeriod.TotalClicks > 0 {
		change := float64(p.CurrentPeriod.TotalClicks-p.PreviousPeriod.TotalClicks) /
			float64(p.PreviousPeriod.TotalClicks) * 100
		p.ClicksChange = &change
	}

	// Conversions change
	if p.PreviousPeriod.TotalConversions > 0 {
		change := float64(p.CurrentPeriod.TotalConversions-p.PreviousPeriod.TotalConversions) /
			float64(p.PreviousPeriod.TotalConversions) * 100
		p.ConversionsChange = &change
	}

	// Revenue change
	if !p.PreviousPeriod.TotalRevenue.IsZero() {
		change, _ := p.CurrentPeriod.TotalRevenue.Sub(p.PreviousPeriod.TotalRevenue).
			Div(p.PreviousPeriod.TotalRevenue).Mul(decimal.NewFromInt(100)).Float64()
		p.RevenueChange = &change
	}

	// ROAS change
	if p.CurrentPeriod.ROAS != nil && p.PreviousPeriod.ROAS != nil && *p.PreviousPeriod.ROAS > 0 {
		change := (*p.CurrentPeriod.ROAS - *p.PreviousPeriod.ROAS) / *p.PreviousPeriod.ROAS * 100
		p.ROASChange = &change
	}

	// CTR change
	if p.CurrentPeriod.CTR != nil && p.PreviousPeriod.CTR != nil && *p.PreviousPeriod.CTR > 0 {
		change := (*p.CurrentPeriod.CTR - *p.PreviousPeriod.CTR) / *p.PreviousPeriod.CTR * 100
		p.CTRChange = &change
	}
}

// PlatformComparisonResult represents comparison between two platforms
type PlatformComparisonResult struct {
	PlatformA AggregatedResult `json:"platform_a"`
	PlatformB AggregatedResult `json:"platform_b"`

	// Winner for each metric (platform name)
	BetterROAS       string `json:"better_roas"`
	BetterCTR        string `json:"better_ctr"`
	LowerCPC         string `json:"lower_cpc"`
	LowerCPA         string `json:"lower_cpa"`
	HigherConversions string `json:"higher_conversions"`
	HigherRevenue    string `json:"higher_revenue"`

	// Insights
	Insights []string `json:"insights"`
}

// DetermineBetterPlatform determines which platform performs better for each metric
func (p *PlatformComparisonResult) DetermineBetterPlatform() {
	platformAName := "Platform A"
	platformBName := "Platform B"
	if p.PlatformA.Platform != nil {
		platformAName = string(*p.PlatformA.Platform)
	}
	if p.PlatformB.Platform != nil {
		platformBName = string(*p.PlatformB.Platform)
	}

	// Better ROAS (higher is better)
	if p.PlatformA.ROAS != nil && p.PlatformB.ROAS != nil {
		if *p.PlatformA.ROAS > *p.PlatformB.ROAS {
			p.BetterROAS = platformAName
			diff := *p.PlatformA.ROAS - *p.PlatformB.ROAS
			p.Insights = append(p.Insights, fmt.Sprintf("%s has %.2fx higher ROAS", platformAName, diff))
		} else if *p.PlatformB.ROAS > *p.PlatformA.ROAS {
			p.BetterROAS = platformBName
			diff := *p.PlatformB.ROAS - *p.PlatformA.ROAS
			p.Insights = append(p.Insights, fmt.Sprintf("%s has %.2fx higher ROAS", platformBName, diff))
		}
	}

	// Better CTR (higher is better)
	if p.PlatformA.CTR != nil && p.PlatformB.CTR != nil {
		if *p.PlatformA.CTR > *p.PlatformB.CTR {
			p.BetterCTR = platformAName
		} else {
			p.BetterCTR = platformBName
		}
	}

	// Lower CPC (lower is better)
	if p.PlatformA.CPC != nil && p.PlatformB.CPC != nil {
		if p.PlatformA.CPC.LessThan(*p.PlatformB.CPC) {
			p.LowerCPC = platformAName
			diff, _ := p.PlatformB.CPC.Sub(*p.PlatformA.CPC).Float64()
			p.Insights = append(p.Insights, fmt.Sprintf("%s has %.2f lower CPC", platformAName, diff))
		} else {
			p.LowerCPC = platformBName
		}
	}

	// Lower CPA (lower is better)
	if p.PlatformA.CPA != nil && p.PlatformB.CPA != nil {
		if p.PlatformA.CPA.LessThan(*p.PlatformB.CPA) {
			p.LowerCPA = platformAName
		} else {
			p.LowerCPA = platformBName
		}
	}

	// Higher Conversions
	if p.PlatformA.TotalConversions > p.PlatformB.TotalConversions {
		p.HigherConversions = platformAName
	} else if p.PlatformB.TotalConversions > p.PlatformA.TotalConversions {
		p.HigherConversions = platformBName
	}

	// Higher Revenue
	if p.PlatformA.TotalRevenue.GreaterThan(p.PlatformB.TotalRevenue) {
		p.HigherRevenue = platformAName
	} else if p.PlatformB.TotalRevenue.GreaterThan(p.PlatformA.TotalRevenue) {
		p.HigherRevenue = platformBName
	}
}

// ============================================================================
// Aggregation Service
// ============================================================================

// AggregationService handles multi-platform data aggregation
type AggregationService struct {
	metricsRepo       repository.MetricsRepository
	campaignRepo      repository.CampaignRepository
	connectedAcctRepo repository.ConnectedAccountRepository
	currencyConverter *currency.Converter
	cache             CacheProvider
	mu                sync.RWMutex

	// Pre-calculated summaries
	dailySummaries map[string]*AggregatedResult
}

// CacheProvider interface for caching
type CacheProvider interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(
	metricsRepo repository.MetricsRepository,
	campaignRepo repository.CampaignRepository,
	connectedAcctRepo repository.ConnectedAccountRepository,
	cache CacheProvider,
) *AggregationService {
	return &AggregationService{
		metricsRepo:       metricsRepo,
		campaignRepo:      campaignRepo,
		connectedAcctRepo: connectedAcctRepo,
		currencyConverter: currency.NewDefaultConverter(),
		cache:             cache,
		dailySummaries:    make(map[string]*AggregatedResult),
	}
}

// ============================================================================
// Core Aggregation Methods
// ============================================================================

// AggregationRequest represents a request for aggregated data
type AggregationRequest struct {
	OrganizationID uuid.UUID          `json:"organization_id"`
	DateRange      entity.DateRange   `json:"date_range"`
	Platforms      []entity.Platform  `json:"platforms,omitempty"`
	CampaignIDs    []uuid.UUID        `json:"campaign_ids,omitempty"`
	Level          AggregationLevel   `json:"level"`
	TargetCurrency string             `json:"target_currency"`
	UseCache       bool               `json:"use_cache"`
}

// AggregationResponse represents the response with aggregated data
type AggregationResponse struct {
	Request    AggregationRequest `json:"request"`
	Results    []AggregatedResult `json:"results"`
	Total      *AggregatedResult  `json:"total,omitempty"`
	GeneratedAt time.Time         `json:"generated_at"`
	FromCache  bool               `json:"from_cache"`
}

// GetUnifiedMetrics retrieves and normalizes metrics from all platforms
func (s *AggregationService) GetUnifiedMetrics(ctx context.Context, req AggregationRequest) ([]UnifiedMetrics, error) {
	// Get all campaigns for the organization
	campaignFilter := entity.CampaignFilter{
		OrganizationID: req.OrganizationID,
		Platforms:      req.Platforms,
	}

	if len(req.CampaignIDs) > 0 {
		// Filter to specific campaigns
		campaignFilter.CampaignIDs = req.CampaignIDs
	}

	campaigns, _, err := s.campaignRepo.List(ctx, campaignFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	// Build campaign lookup map
	campaignMap := make(map[uuid.UUID]entity.Campaign)
	for _, c := range campaigns {
		campaignMap[c.ID] = c
	}

	var unifiedMetrics []UnifiedMetrics

	// Get metrics for each campaign
	for _, campaign := range campaigns {
		dailyMetrics, err := s.metricsRepo.GetCampaignMetrics(ctx, campaign.ID, req.DateRange)
		if err != nil {
			continue // Skip failed campaigns
		}

		for _, m := range dailyMetrics {
			unified := UnifiedMetrics{
				ID:           uuid.New(),
				Date:         m.MetricDate,
				Platform:     m.Platform,
				CampaignID:   campaign.ID,
				CampaignName: campaign.PlatformCampaignName,
				AdAccountID:  campaign.AdAccountID,
				Spend:        s.convertCurrency(m.Spend, m.Currency, req.TargetCurrency),
				Impressions:  m.Impressions,
				Clicks:       m.Clicks,
				Conversions:  m.Conversions,
				Revenue:      s.convertCurrency(m.ConversionValue, m.Currency, req.TargetCurrency),
				Currency:     req.TargetCurrency,
				LastSyncedAt: m.LastSyncedAt,
			}
			unified.CalculateDerived()
			unifiedMetrics = append(unifiedMetrics, unified)
		}
	}

	return unifiedMetrics, nil
}

// Aggregate aggregates unified metrics based on the specified level
func (s *AggregationService) Aggregate(ctx context.Context, req AggregationRequest) (*AggregationResponse, error) {
	// Check cache first
	if req.UseCache && s.cache != nil {
		cacheKey := s.buildCacheKey(req)
		if cached, err := s.getFromCache(ctx, cacheKey); err == nil {
			return cached, nil
		}
	}

	// Get unified metrics
	metrics, err := s.GetUnifiedMetrics(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform aggregation based on level
	var results []AggregatedResult
	switch req.Level {
	case AggregationByPlatform:
		results = s.aggregateByPlatform(metrics, req.TargetCurrency)
	case AggregationByDate:
		results = s.aggregateByDate(metrics, req.TargetCurrency)
	case AggregationByCampaign:
		results = s.aggregateByCampaign(metrics, req.TargetCurrency)
	case AggregationTotal:
		total := s.aggregateTotal(metrics, req.TargetCurrency)
		results = []AggregatedResult{total}
	default:
		results = s.aggregateByPlatform(metrics, req.TargetCurrency)
	}

	// Calculate total
	total := s.aggregateTotal(metrics, req.TargetCurrency)

	response := &AggregationResponse{
		Request:     req,
		Results:     results,
		Total:       &total,
		GeneratedAt: time.Now(),
		FromCache:   false,
	}

	// Cache the result
	if req.UseCache && s.cache != nil {
		cacheKey := s.buildCacheKey(req)
		_ = s.setToCache(ctx, cacheKey, response, 5*time.Minute)
	}

	return response, nil
}

// aggregateByPlatform groups metrics by platform
func (s *AggregationService) aggregateByPlatform(metrics []UnifiedMetrics, currency string) []AggregatedResult {
	grouped := make(map[entity.Platform]*AggregatedResult)

	for _, m := range metrics {
		if _, exists := grouped[m.Platform]; !exists {
			p := m.Platform
			grouped[m.Platform] = &AggregatedResult{
				Key:      string(m.Platform),
				Level:    AggregationByPlatform,
				Platform: &p,
				Currency: currency,
			}
		}

		agg := grouped[m.Platform]
		agg.TotalSpend = agg.TotalSpend.Add(m.Spend)
		agg.TotalImpressions += m.Impressions
		agg.TotalClicks += m.Clicks
		agg.TotalConversions += m.Conversions
		agg.TotalRevenue = agg.TotalRevenue.Add(m.Revenue)
	}

	// Calculate derived metrics and count unique campaigns/days
	results := make([]AggregatedResult, 0, len(grouped))
	for _, agg := range grouped {
		agg.CalculateDerived()
		agg.CampaignCount = s.countUniqueCampaigns(metrics, agg.Platform)
		agg.DayCount = s.countUniqueDays(metrics, agg.Platform)
		results = append(results, *agg)
	}

	// Sort by spend descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalSpend.GreaterThan(results[j].TotalSpend)
	})

	return results
}

// aggregateByDate groups metrics by date
func (s *AggregationService) aggregateByDate(metrics []UnifiedMetrics, currency string) []AggregatedResult {
	grouped := make(map[string]*AggregatedResult)

	for _, m := range metrics {
		dateKey := m.Date.Format("2006-01-02")
		if _, exists := grouped[dateKey]; !exists {
			d := m.Date
			grouped[dateKey] = &AggregatedResult{
				Key:      dateKey,
				Level:    AggregationByDate,
				Date:     &d,
				Currency: currency,
			}
		}

		agg := grouped[dateKey]
		agg.TotalSpend = agg.TotalSpend.Add(m.Spend)
		agg.TotalImpressions += m.Impressions
		agg.TotalClicks += m.Clicks
		agg.TotalConversions += m.Conversions
		agg.TotalRevenue = agg.TotalRevenue.Add(m.Revenue)
	}

	// Calculate derived metrics
	results := make([]AggregatedResult, 0, len(grouped))
	for _, agg := range grouped {
		agg.CalculateDerived()
		results = append(results, *agg)
	}

	// Sort by date ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Date == nil || results[j].Date == nil {
			return false
		}
		return results[i].Date.Before(*results[j].Date)
	})

	return results
}

// aggregateByCampaign groups metrics by campaign
func (s *AggregationService) aggregateByCampaign(metrics []UnifiedMetrics, currency string) []AggregatedResult {
	grouped := make(map[uuid.UUID]*AggregatedResult)

	for _, m := range metrics {
		if _, exists := grouped[m.CampaignID]; !exists {
			cid := m.CampaignID
			p := m.Platform
			grouped[m.CampaignID] = &AggregatedResult{
				Key:        m.CampaignID.String(),
				Level:      AggregationByCampaign,
				CampaignID: &cid,
				Platform:   &p,
				Currency:   currency,
			}
		}

		agg := grouped[m.CampaignID]
		agg.TotalSpend = agg.TotalSpend.Add(m.Spend)
		agg.TotalImpressions += m.Impressions
		agg.TotalClicks += m.Clicks
		agg.TotalConversions += m.Conversions
		agg.TotalRevenue = agg.TotalRevenue.Add(m.Revenue)
		agg.CampaignCount = 1
	}

	// Calculate derived metrics
	results := make([]AggregatedResult, 0, len(grouped))
	for _, agg := range grouped {
		agg.CalculateDerived()
		results = append(results, *agg)
	}

	// Sort by spend descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalSpend.GreaterThan(results[j].TotalSpend)
	})

	return results
}

// aggregateTotal aggregates all metrics into a single total
func (s *AggregationService) aggregateTotal(metrics []UnifiedMetrics, currency string) AggregatedResult {
	agg := AggregatedResult{
		Key:      "total",
		Level:    AggregationTotal,
		Currency: currency,
	}

	campaignsSeen := make(map[uuid.UUID]bool)
	daysSeen := make(map[string]bool)

	for _, m := range metrics {
		agg.TotalSpend = agg.TotalSpend.Add(m.Spend)
		agg.TotalImpressions += m.Impressions
		agg.TotalClicks += m.Clicks
		agg.TotalConversions += m.Conversions
		agg.TotalRevenue = agg.TotalRevenue.Add(m.Revenue)

		campaignsSeen[m.CampaignID] = true
		daysSeen[m.Date.Format("2006-01-02")] = true
	}

	agg.CampaignCount = len(campaignsSeen)
	agg.DayCount = len(daysSeen)
	agg.CalculateDerived()

	return agg
}

// ============================================================================
// Comparison Methods
// ============================================================================

// ComparePeriods compares current period with previous period
func (s *AggregationService) ComparePeriods(ctx context.Context, req AggregationRequest) (*PeriodComparisonResult, error) {
	// Calculate previous period (same duration, immediately before)
	duration := req.DateRange.EndDate.Sub(req.DateRange.StartDate)
	previousRange := entity.DateRange{
		StartDate: req.DateRange.StartDate.Add(-duration - 24*time.Hour),
		EndDate:   req.DateRange.StartDate.Add(-24 * time.Hour),
	}

	// Get current period metrics
	currentReq := req
	currentReq.Level = AggregationTotal
	currentResponse, err := s.Aggregate(ctx, currentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period metrics: %w", err)
	}

	// Get previous period metrics
	previousReq := req
	previousReq.DateRange = previousRange
	previousReq.Level = AggregationTotal
	previousResponse, err := s.Aggregate(ctx, previousReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get previous period metrics: %w", err)
	}

	comparison := &PeriodComparisonResult{}
	if currentResponse.Total != nil {
		comparison.CurrentPeriod = *currentResponse.Total
	}
	if previousResponse.Total != nil {
		comparison.PreviousPeriod = *previousResponse.Total
	}

	comparison.CalculateChanges()

	return comparison, nil
}

// ComparePlatforms compares two platforms
func (s *AggregationService) ComparePlatforms(ctx context.Context, req AggregationRequest, platformA, platformB entity.Platform) (*PlatformComparisonResult, error) {
	// Get metrics for platform A
	reqA := req
	reqA.Platforms = []entity.Platform{platformA}
	reqA.Level = AggregationTotal
	responseA, err := s.Aggregate(ctx, reqA)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform A metrics: %w", err)
	}

	// Get metrics for platform B
	reqB := req
	reqB.Platforms = []entity.Platform{platformB}
	reqB.Level = AggregationTotal
	responseB, err := s.Aggregate(ctx, reqB)
	if err != nil {
		return nil, fmt.Errorf("failed to get platform B metrics: %w", err)
	}

	comparison := &PlatformComparisonResult{}
	if responseA.Total != nil {
		comparison.PlatformA = *responseA.Total
		comparison.PlatformA.Platform = &platformA
	}
	if responseB.Total != nil {
		comparison.PlatformB = *responseB.Total
		comparison.PlatformB.Platform = &platformB
	}

	comparison.DetermineBetterPlatform()

	return comparison, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (s *AggregationService) convertCurrency(amount decimal.Decimal, from, to string) decimal.Decimal {
	if from == to || from == "" || to == "" {
		return amount
	}
	if amount.IsZero() {
		return amount
	}
	converted, err := s.currencyConverter.Convert(amount, from, to)
	if err != nil {
		return amount
	}
	return converted
}

func (s *AggregationService) countUniqueCampaigns(metrics []UnifiedMetrics, platform *entity.Platform) int {
	seen := make(map[uuid.UUID]bool)
	for _, m := range metrics {
		if platform == nil || m.Platform == *platform {
			seen[m.CampaignID] = true
		}
	}
	return len(seen)
}

func (s *AggregationService) countUniqueDays(metrics []UnifiedMetrics, platform *entity.Platform) int {
	seen := make(map[string]bool)
	for _, m := range metrics {
		if platform == nil || m.Platform == *platform {
			seen[m.Date.Format("2006-01-02")] = true
		}
	}
	return len(seen)
}

// ============================================================================
// Caching Methods
// ============================================================================

func (s *AggregationService) buildCacheKey(req AggregationRequest) string {
	return fmt.Sprintf("agg:%s:%s:%s:%s:%s",
		req.OrganizationID.String(),
		req.DateRange.StartDate.Format("2006-01-02"),
		req.DateRange.EndDate.Format("2006-01-02"),
		string(req.Level),
		req.TargetCurrency,
	)
}

func (s *AggregationService) getFromCache(ctx context.Context, key string) (*AggregationResponse, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache not available")
	}

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var response AggregationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	response.FromCache = true
	return &response, nil
}

func (s *AggregationService) setToCache(ctx context.Context, key string, response *AggregationResponse, ttl time.Duration) error {
	if s.cache == nil {
		return nil
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.cache.Set(ctx, key, data, ttl)
}

// InvalidateCache invalidates cache for an organization
func (s *AggregationService) InvalidateCache(ctx context.Context, orgID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}
	// In a real implementation, you'd use pattern matching to delete all keys
	// for the organization. This is simplified.
	return nil
}

// ============================================================================
// Pre-calculation / Background Job Methods
// ============================================================================

// DailySummaryJob represents a background job for pre-calculating daily summaries
type DailySummaryJob struct {
	service       *AggregationService
	orgRepo       repository.OrganizationRepository
	isRunning     bool
	mu            sync.Mutex
	stopChan      chan struct{}
	lastRunTime   time.Time
	runInterval   time.Duration
}

// NewDailySummaryJob creates a new daily summary background job
func NewDailySummaryJob(service *AggregationService, orgRepo repository.OrganizationRepository, interval time.Duration) *DailySummaryJob {
	return &DailySummaryJob{
		service:     service,
		orgRepo:     orgRepo,
		runInterval: interval,
		stopChan:    make(chan struct{}),
	}
}

// Start starts the background job
func (j *DailySummaryJob) Start(ctx context.Context) error {
	j.mu.Lock()
	if j.isRunning {
		j.mu.Unlock()
		return fmt.Errorf("job already running")
	}
	j.isRunning = true
	j.mu.Unlock()

	go j.run(ctx)
	return nil
}

// Stop stops the background job
func (j *DailySummaryJob) Stop() {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.isRunning {
		close(j.stopChan)
		j.isRunning = false
	}
}

func (j *DailySummaryJob) run(ctx context.Context) {
	ticker := time.NewTicker(j.runInterval)
	defer ticker.Stop()

	// Run immediately on start
	j.calculateAllSummaries(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-j.stopChan:
			return
		case <-ticker.C:
			j.calculateAllSummaries(ctx)
		}
	}
}

// calculateAllSummaries calculates summaries for all organizations
func (j *DailySummaryJob) calculateAllSummaries(ctx context.Context) {
	j.lastRunTime = time.Now()

	// Get all organizations
	orgs, err := j.orgRepo.List(ctx, &entity.Pagination{Page: 1, PageSize: 1000})
	if err != nil {
		return
	}

	// Calculate summaries for each organization in parallel (with limit)
	const maxWorkers = 5
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, org := range orgs {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(orgID uuid.UUID) {
			defer wg.Done()
			defer func() { <-semaphore }()

			j.calculateOrgSummaries(ctx, orgID)
		}(org.ID)
	}

	wg.Wait()
}

// calculateOrgSummaries calculates and caches summaries for an organization
func (j *DailySummaryJob) calculateOrgSummaries(ctx context.Context, orgID uuid.UUID) {
	// Calculate for common date ranges
	dateRanges := []struct {
		name  string
		start time.Time
		end   time.Time
	}{
		{"today", time.Now().Truncate(24 * time.Hour), time.Now()},
		{"yesterday", time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour), time.Now().Truncate(24 * time.Hour)},
		{"last_7_days", time.Now().AddDate(0, 0, -7), time.Now()},
		{"last_30_days", time.Now().AddDate(0, 0, -30), time.Now()},
		{"this_month", time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC), time.Now()},
	}

	for _, dr := range dateRanges {
		// Calculate total aggregation
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: dr.start, EndDate: dr.end},
			Level:          AggregationTotal,
			TargetCurrency: "MYR",
			UseCache:       false,
		}

		response, err := j.service.Aggregate(ctx, req)
		if err != nil {
			continue
		}

		// Cache with longer TTL for pre-calculated data
		cacheKey := fmt.Sprintf("precalc:%s:%s", orgID.String(), dr.name)
		if j.service.cache != nil {
			data, _ := json.Marshal(response)
			_ = j.service.cache.Set(ctx, cacheKey, data, 1*time.Hour)
		}

		// Also calculate by platform
		req.Level = AggregationByPlatform
		platformResponse, err := j.service.Aggregate(ctx, req)
		if err == nil && j.service.cache != nil {
			platformKey := fmt.Sprintf("precalc:%s:%s:platform", orgID.String(), dr.name)
			data, _ := json.Marshal(platformResponse)
			_ = j.service.cache.Set(ctx, platformKey, data, 1*time.Hour)
		}
	}
}

// GetLastRunTime returns the last run time of the job
func (j *DailySummaryJob) GetLastRunTime() time.Time {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.lastRunTime
}

// IsRunning returns whether the job is currently running
func (j *DailySummaryJob) IsRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.isRunning
}

// ============================================================================
// Dashboard Data Methods (Optimized with Caching)
// ============================================================================

// DashboardData represents pre-aggregated dashboard data
type DashboardData struct {
	OrganizationID uuid.UUID           `json:"organization_id"`
	GeneratedAt    time.Time           `json:"generated_at"`
	FromCache      bool                `json:"from_cache"`

	// Period summaries
	Today      *AggregatedResult `json:"today,omitempty"`
	Yesterday  *AggregatedResult `json:"yesterday,omitempty"`
	Last7Days  *AggregatedResult `json:"last_7_days,omitempty"`
	Last30Days *AggregatedResult `json:"last_30_days,omitempty"`
	ThisMonth  *AggregatedResult `json:"this_month,omitempty"`

	// Platform breakdown (last 30 days)
	PlatformBreakdown []AggregatedResult `json:"platform_breakdown,omitempty"`

	// Daily trend (last 30 days)
	DailyTrend []AggregatedResult `json:"daily_trend,omitempty"`

	// Period comparison
	Comparison *PeriodComparisonResult `json:"comparison,omitempty"`
}

// GetDashboardData returns optimized dashboard data with caching
func (s *AggregationService) GetDashboardData(ctx context.Context, orgID uuid.UUID, targetCurrency string) (*DashboardData, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("dashboard:%s:%s", orgID.String(), targetCurrency)
	if s.cache != nil {
		if data, err := s.cache.Get(ctx, cacheKey); err == nil {
			var dashboard DashboardData
			if err := json.Unmarshal(data, &dashboard); err == nil {
				dashboard.FromCache = true
				return &dashboard, nil
			}
		}
	}

	dashboard := &DashboardData{
		OrganizationID: orgID,
		GeneratedAt:    time.Now(),
		FromCache:      false,
	}

	now := time.Now()

	// Calculate various time ranges in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Today
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: now.Truncate(24 * time.Hour), EndDate: now},
			Level:          AggregationTotal,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, req); err == nil && resp.Total != nil {
			mu.Lock()
			dashboard.Today = resp.Total
			mu.Unlock()
		}
	}()

	// Yesterday
	wg.Add(1)
	go func() {
		defer wg.Done()
		yesterdayStart := now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		yesterdayEnd := now.Truncate(24 * time.Hour)
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: yesterdayStart, EndDate: yesterdayEnd},
			Level:          AggregationTotal,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, req); err == nil && resp.Total != nil {
			mu.Lock()
			dashboard.Yesterday = resp.Total
			mu.Unlock()
		}
	}()

	// Last 7 days
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: now.AddDate(0, 0, -7), EndDate: now},
			Level:          AggregationTotal,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, req); err == nil && resp.Total != nil {
			mu.Lock()
			dashboard.Last7Days = resp.Total
			mu.Unlock()
		}
	}()

	// Last 30 days + platform breakdown + daily trend
	wg.Add(1)
	go func() {
		defer wg.Done()
		dateRange := entity.DateRange{StartDate: now.AddDate(0, 0, -30), EndDate: now}

		// Total
		reqTotal := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      dateRange,
			Level:          AggregationTotal,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, reqTotal); err == nil && resp.Total != nil {
			mu.Lock()
			dashboard.Last30Days = resp.Total
			mu.Unlock()
		}

		// Platform breakdown
		reqPlatform := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      dateRange,
			Level:          AggregationByPlatform,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, reqPlatform); err == nil {
			mu.Lock()
			dashboard.PlatformBreakdown = resp.Results
			mu.Unlock()
		}

		// Daily trend
		reqDaily := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      dateRange,
			Level:          AggregationByDate,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, reqDaily); err == nil {
			mu.Lock()
			dashboard.DailyTrend = resp.Results
			mu.Unlock()
		}
	}()

	// This month
	wg.Add(1)
	go func() {
		defer wg.Done()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: monthStart, EndDate: now},
			Level:          AggregationTotal,
			TargetCurrency: targetCurrency,
		}
		if resp, err := s.Aggregate(ctx, req); err == nil && resp.Total != nil {
			mu.Lock()
			dashboard.ThisMonth = resp.Total
			mu.Unlock()
		}
	}()

	// Period comparison (last 7 days vs previous 7 days)
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := AggregationRequest{
			OrganizationID: orgID,
			DateRange:      entity.DateRange{StartDate: now.AddDate(0, 0, -7), EndDate: now},
			TargetCurrency: targetCurrency,
		}
		if comparison, err := s.ComparePeriods(ctx, req); err == nil {
			mu.Lock()
			dashboard.Comparison = comparison
			mu.Unlock()
		}
	}()

	wg.Wait()

	// Cache the result
	if s.cache != nil {
		data, _ := json.Marshal(dashboard)
		_ = s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return dashboard, nil
}
