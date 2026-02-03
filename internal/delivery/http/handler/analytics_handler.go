package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/cache"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/analytics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService *analytics.Service
	cache            *cache.Cache
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService *analytics.Service, cache *cache.Cache) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		cache:            cache,
	}
}

// CalculateMetricsRequest is the API request for metrics calculation
type CalculateMetricsRequest struct {
	DateRange      DateRangeRequest `json:"date_range" binding:"required"`
	Platforms      []string         `json:"platforms,omitempty"`
	CampaignIDs    []string         `json:"campaign_ids,omitempty"`
	AdAccountIDs   []string         `json:"ad_account_ids,omitempty"`
	TargetCurrency string           `json:"target_currency,omitempty"`
	IncludeDetails bool             `json:"include_details,omitempty"`
}

// DateRangeRequest represents the date range in API request
type DateRangeRequest struct {
	StartDate string `json:"start_date" binding:"required"` // ISO format: 2006-01-02 or RFC3339
	EndDate   string `json:"end_date" binding:"required"`   // ISO format: 2006-01-02 or RFC3339
}

// CalculateMetrics handles POST /api/v1/analytics/calculate
// @Summary Calculate comprehensive analytics metrics
// @Description Calculates ROAS, CPA, CTR per platform and combined with zero-division protection
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body CalculateMetricsRequest true "Analytics calculation request"
// @Success 200 {object} entity.AnalyticsResponse
// @Router /api/v1/analytics/calculate [post]
func (h *AnalyticsHandler) CalculateMetrics(c *gin.Context) {
	var req CalculateMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Parse date range
	startDate, err := parseDate(req.DateRange.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_date format",
			"details": "Use ISO format: 2006-01-02 or RFC3339",
		})
		return
	}
	endDate, err := parseDate(req.DateRange.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_date format",
			"details": "Use ISO format: 2006-01-02 or RFC3339",
		})
		return
	}

	// Get organization ID from auth context
	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	// Convert platforms
	platforms := make([]entity.Platform, 0, len(req.Platforms))
	for _, p := range req.Platforms {
		platform := entity.Platform(p)
		if !platform.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid platform",
				"details": "Valid platforms: meta, tiktok, shopee",
			})
			return
		}
		platforms = append(platforms, platform)
	}

	// Convert campaign IDs
	campaignIDs := make([]uuid.UUID, 0, len(req.CampaignIDs))
	for _, id := range req.CampaignIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid campaign_id format",
				"details": id + " is not a valid UUID",
			})
			return
		}
		campaignIDs = append(campaignIDs, uid)
	}

	// Convert ad account IDs
	adAccountIDs := make([]uuid.UUID, 0, len(req.AdAccountIDs))
	for _, id := range req.AdAccountIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid ad_account_id format",
				"details": id + " is not a valid UUID",
			})
			return
		}
		adAccountIDs = append(adAccountIDs, uid)
	}

	// Build analytics request
	analyticsReq := entity.AnalyticsRequest{
		OrganizationID: orgID,
		DateRange: entity.DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
		Platforms:      platforms,
		CampaignIDs:    campaignIDs,
		AdAccountIDs:   adAccountIDs,
		TargetCurrency: req.TargetCurrency,
		IncludeDetails: req.IncludeDetails,
	}

	// Default currency
	if analyticsReq.TargetCurrency == "" {
		analyticsReq.TargetCurrency = "MYR"
	}

	// Call service
	result, err := h.analyticsService.CalculateAnalytics(c.Request.Context(), analyticsReq)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetDashboard returns dashboard metrics with Redis caching
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	dateRange := parseDateRangeFromQuery(c)

	// Return mock data if analytics service is not available (local dev mode)
	if h.analyticsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    getMockDashboardMetrics(),
			"cached":  false,
			"mock":    true,
		})
		return
	}

	// Generate cache key
	cacheKey := cache.DashboardCacheKey(orgID, dateRange.StartDate, dateRange.EndDate)

	// Try to get from cache first
	if h.cache != nil {
		var cachedDashboard analytics.DashboardMetrics
		if err := h.cache.Get(c.Request.Context(), cacheKey, &cachedDashboard); err == nil {
			// Cache hit - return cached data
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    cachedDashboard,
				"cached":  true,
			})
			return
		}
	}

	// Cache miss - fetch from service
	dashboard, err := h.analyticsService.GetDashboardMetrics(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Store in cache (non-blocking, ignore errors)
	if h.cache != nil {
		go func() {
			_ = h.cache.Set(c.Request.Context(), cacheKey, dashboard, cache.DashboardTTL)
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
		"cached":  false,
	})
}

// GetOverview returns overview metrics
func (h *AnalyticsHandler) GetOverview(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRangeFromQuery(c)

	report, err := h.analyticsService.GetCrossPlatformReport(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

// GetPlatformComparison returns platform comparison
func (h *AnalyticsHandler) GetPlatformComparison(c *gin.Context) {
	// Return mock data if analytics service is not available
	if h.analyticsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    getMockPlatformData(),
			"mock":    true,
		})
		return
	}

	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRangeFromQuery(c)

	report, err := h.analyticsService.GetCrossPlatformReport(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": report.ByPlatform})
}

// TimeSeriesResponse represents the time series API response format
type TimeSeriesResponse struct {
	DateRange   entity.DateRange           `json:"date_range"`
	Granularity string                     `json:"granularity"`
	Data        []entity.DailyMetricsTrend `json:"data"`
	Totals      *TotalsResponse            `json:"totals,omitempty"`
}

// TotalsResponse contains summary totals for the time series
type TotalsResponse struct {
	Spend       float64 `json:"spend"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
}

// GetTrends returns metric trends (time series data)
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	// Return mock data if analytics service is not available
	if h.analyticsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    getMockTimeseriesData(),
			"mock":    true,
		})
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	dateRange := parseDateRangeFromQuery(c)
	platform := c.Query("platforms")
	granularity := c.DefaultQuery("granularity", "day")

	// Get time series metrics
	trend, err := h.analyticsService.GetTimeSeriesMetrics(c.Request.Context(), orgID, dateRange, platform, granularity)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Calculate totals from the trend data
	var totals TotalsResponse
	for _, t := range trend {
		spendFloat, _ := t.Spend.Float64()
		totals.Spend += spendFloat
		totals.Impressions += t.Impressions
		totals.Clicks += t.Clicks
		totals.Conversions += t.Conversions
		// Revenue is calculated from ROAS * Spend
		if t.ROAS > 0 {
			totals.Revenue += spendFloat * t.ROAS
		}
	}

	response := TimeSeriesResponse{
		DateRange:   dateRange,
		Granularity: granularity,
		Data:        trend,
		Totals:      &totals,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// TopPerformersResponse represents the top performers API response
type TopPerformersResponse struct {
	Campaigns []CampaignPerformanceResponse `json:"campaigns"`
	Metric    string                        `json:"metric"`
	Limit     int                           `json:"limit"`
}

// CampaignPerformanceResponse represents a single campaign performance entry
type CampaignPerformanceResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Platform string  `json:"platform"`
	Status   string  `json:"status"`
	Spend    float64 `json:"spend"`
	ROAS     float64 `json:"roas"`
	CTR      float64 `json:"ctr"`
	CPA      float64 `json:"cpa"`
	Change   float64 `json:"change"`
	Trend    string  `json:"trend"` // "up", "down", "stable"
}

// GetTopPerformers returns top performing campaigns
func (h *AnalyticsHandler) GetTopPerformers(c *gin.Context) {
	// Return mock data if analytics service is not available
	if h.analyticsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"campaigns": getMockTopCampaigns(),
				"metric":    c.DefaultQuery("metric", "roas"),
				"limit":     5,
			},
			"mock": true,
		})
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization not found in context"})
		return
	}

	dateRange := parseDateRangeFromQuery(c)
	limitStr := c.DefaultQuery("limit", "5")
	metric := c.DefaultQuery("metric", "roas")

	limit := 5
	if l, err := parseInt(limitStr); err == nil && l > 0 {
		limit = l
	}

	// Get top campaigns from service
	topCampaigns, err := h.analyticsService.GetTopCampaigns(c.Request.Context(), orgID, dateRange, limit)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Convert to response format
	campaigns := make([]CampaignPerformanceResponse, 0, len(topCampaigns))
	for _, tc := range topCampaigns {
		spendFloat, _ := tc.Spend.Float64()
		cpaFloat := float64(0)
		if tc.Conversions > 0 {
			cpaFloat = spendFloat / float64(tc.Conversions)
		}

		// Determine trend based on ROAS performance
		trend := "stable"
		if tc.ROAS > 3.0 {
			trend = "up"
		} else if tc.ROAS < 1.0 {
			trend = "down"
		}

		campaigns = append(campaigns, CampaignPerformanceResponse{
			ID:       tc.ID.String(),
			Name:     tc.Name,
			Platform: string(tc.Platform),
			Status:   "active", // Default status since TopPerformer doesn't have it
			Spend:    spendFloat,
			ROAS:     tc.ROAS,
			CTR:      tc.CTR,
			CPA:      cpaFloat,
			Change:   0, // Change calculation would require previous period data
			Trend:    trend,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": TopPerformersResponse{
			Campaigns: campaigns,
			Metric:    metric,
			Limit:     limit,
		},
	})
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// GenerateReport generates a downloadable report
func (h *AnalyticsHandler) GenerateReport(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRangeFromQuery(c)

	report, err := h.analyticsService.GenerateReport(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

// GetReport retrieves a generated report
func (h *AnalyticsHandler) GetReport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": nil})
}

// ExportAnalytics handles GET /api/v1/analytics/export
// @Summary Export analytics data
// @Description Exports analytics data in CSV or Excel format
// @Tags Analytics
// @Accept json
// @Produce text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param format query string false "Export format (csv, xlsx)" default(csv)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {file} file
// @Router /api/v1/analytics/export [get]
func (h *AnalyticsHandler) ExportAnalytics(c *gin.Context) {
	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Organization not found in context",
			},
		})
		return
	}

	format := c.DefaultQuery("format", "csv")
	dateRange := parseDateRangeFromQuery(c)

	// Get analytics data for export
	data, err := h.analyticsService.ExportAnalytics(c.Request.Context(), orgID, dateRange, format)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Set appropriate headers based on format
	filename := "analytics_export_" + dateRange.StartDate.Format("2006-01-02") + "_" + dateRange.EndDate.Format("2006-01-02")
	switch format {
	case "xlsx":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename="+filename+".xlsx")
	default:
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename="+filename+".csv")
	}

	c.Data(http.StatusOK, c.GetHeader("Content-Type"), data)
}

// Campaign handlers
func (h *AnalyticsHandler) ListCampaigns(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
func (h *AnalyticsHandler) GetCampaign(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": nil}) }
func (h *AnalyticsHandler) GetCampaignMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": nil})
}
func (h *AnalyticsHandler) ListAdSets(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
func (h *AnalyticsHandler) ListAds(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
func (h *AnalyticsHandler) GetAdSet(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"data": nil}) }
func (h *AnalyticsHandler) GetAdSetMetrics(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": nil}) }
func (h *AnalyticsHandler) ListAdsByAdSet(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}
func (h *AnalyticsHandler) GetAd(c *gin.Context)        { c.JSON(http.StatusOK, gin.H{"data": nil}) }
func (h *AnalyticsHandler) GetAdMetrics(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"data": nil}) }

// parseDate parses a date string in various formats
func parseDate(s string) (time.Time, error) {
	// Try RFC3339 first (full timestamp)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try ISO date format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	// Try with time but no timezone
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t, nil
	}

	return time.Time{}, errors.New("invalid date format")
}

// parseDateRangeFromQuery parses date range from query parameters
func parseDateRangeFromQuery(c *gin.Context) entity.DateRange {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	// Default to last 30 days if not provided
	if startStr == "" || endStr == "" {
		return entity.Last30Days()
	}

	startDate, err := parseDate(startStr)
	if err != nil {
		return entity.Last30Days()
	}

	endDate, err := parseDate(endStr)
	if err != nil {
		return entity.Last30Days()
	}

	// Validate range
	if startDate.After(endDate) {
		return entity.Last30Days()
	}

	return entity.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}
}

// parseDateRange is an alias for parseDateRangeFromQuery for backwards compatibility
func parseDateRange(c *gin.Context) entity.DateRange {
	return parseDateRangeFromQuery(c)
}

func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	id := c.Param(param)
	return uuid.Parse(id)
}

// ============================================================================
// Mock data functions for local development/testing
// ============================================================================

// getMockDashboardMetrics returns mock dashboard data for local testing
func getMockDashboardMetrics() gin.H {
	return gin.H{
		"totals": gin.H{
			"spend":       12500.00,
			"revenue":     45000.00,
			"roas":        3.6,
			"conversions": 1250,
			"clicks":      25000,
			"impressions": 500000,
			"ctr":         5.0,
			"cpc":         0.50,
			"cpa":         10.00,
		},
		"changes": gin.H{
			"spend":       12.5,
			"revenue":     18.3,
			"roas":        5.2,
			"conversions": 15.0,
		},
		"activeCampaigns": 12,
		"lastSyncedAt":    time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
	}
}

// getMockPlatformData returns mock platform comparison data
func getMockPlatformData() []gin.H {
	return []gin.H{
		{
			"platform":    "meta",
			"spend":       5000.00,
			"revenue":     18000.00,
			"roas":        3.6,
			"conversions": 500,
			"impressions": 200000,
			"clicks":      10000,
			"ctr":         5.0,
		},
		{
			"platform":    "tiktok",
			"spend":       4500.00,
			"revenue":     16200.00,
			"roas":        3.6,
			"conversions": 450,
			"impressions": 180000,
			"clicks":      9000,
			"ctr":         5.0,
		},
		{
			"platform":    "shopee",
			"spend":       3000.00,
			"revenue":     10800.00,
			"roas":        3.6,
			"conversions": 300,
			"impressions": 120000,
			"clicks":      6000,
			"ctr":         5.0,
		},
	}
}

// getMockTimeseriesData returns mock timeseries data
func getMockTimeseriesData() []gin.H {
	data := make([]gin.H, 0, 30)
	baseSpend := 400.0
	baseRevenue := 1500.0

	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		// Add some variation
		multiplier := 1.0 + (float64(i%7) * 0.1)
		data = append(data, gin.H{
			"date":        date.Format("2006-01-02"),
			"spend":       baseSpend * multiplier,
			"revenue":     baseRevenue * multiplier,
			"conversions": int(40 * multiplier),
			"clicks":      int(800 * multiplier),
			"impressions": int(16000 * multiplier),
		})
	}
	return data
}

// getMockTopCampaigns returns mock top campaigns data
func getMockTopCampaigns() []gin.H {
	return []gin.H{
		{
			"id":          "camp-001",
			"name":        "Summer Sale 2026",
			"platform":    "meta",
			"status":      "active",
			"spend":       2500.00,
			"revenue":     10000.00,
			"roas":        4.0,
			"conversions": 250,
		},
		{
			"id":          "camp-002",
			"name":        "Brand Awareness",
			"platform":    "tiktok",
			"status":      "active",
			"spend":       2000.00,
			"revenue":     7200.00,
			"roas":        3.6,
			"conversions": 180,
		},
		{
			"id":          "camp-003",
			"name":        "Product Launch",
			"platform":    "shopee",
			"status":      "active",
			"spend":       1500.00,
			"revenue":     6000.00,
			"roas":        4.0,
			"conversions": 150,
		},
		{
			"id":          "camp-004",
			"name":        "Retargeting",
			"platform":    "meta",
			"status":      "active",
			"spend":       1000.00,
			"revenue":     4500.00,
			"roas":        4.5,
			"conversions": 100,
		},
		{
			"id":          "camp-005",
			"name":        "Flash Sale",
			"platform":    "shopee",
			"status":      "active",
			"spend":       800.00,
			"revenue":     2800.00,
			"roas":        3.5,
			"conversions": 80,
		},
	}
}
