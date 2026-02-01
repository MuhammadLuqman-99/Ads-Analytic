package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/analytics"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService *analytics.Service
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
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

// GetDashboard returns dashboard metrics
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRangeFromQuery(c)

	dashboard, err := h.analyticsService.GetDashboardMetrics(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dashboard})
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
	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRangeFromQuery(c)

	report, err := h.analyticsService.GetCrossPlatformReport(c.Request.Context(), orgID, dateRange)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report.ByPlatform})
}

// GetTrends returns metric trends
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// GetTopPerformers returns top performing campaigns
func (h *AnalyticsHandler) GetTopPerformers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
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
