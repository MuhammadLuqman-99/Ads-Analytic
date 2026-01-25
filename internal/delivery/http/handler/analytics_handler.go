package handler

import (
	"net/http"

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

// GetDashboard returns dashboard metrics
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	dateRange := parseDateRange(c)

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
	dateRange := parseDateRange(c)

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
	dateRange := parseDateRange(c)

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
	dateRange := parseDateRange(c)

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

func parseDateRange(c *gin.Context) entity.DateRange {
	return entity.Last30Days()
}

func parseUUIDParam(c *gin.Context, param string) (uuid.UUID, error) {
	id := c.Param(param)
	return uuid.Parse(id)
}
