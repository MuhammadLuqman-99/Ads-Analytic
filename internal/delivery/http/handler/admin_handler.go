package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domain "github.com/ads-aggregator/ads-aggregator/internal/domain/analytics"
)

// AdminHandler handles admin panel API requests
type AdminHandler struct {
	repo domain.Repository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo domain.Repository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

// RegisterRoutes registers the admin routes
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Dashboard
	r.GET("/dashboard", h.GetDashboard)
	r.GET("/metrics", h.GetMetrics)

	// Active users
	r.GET("/users/active", h.GetActiveUsers)
	r.GET("/users/churned", h.GetChurnedUsers)
	r.GET("/users/top", h.GetTopUsers)
	r.GET("/users/:id", h.GetUserProfile)

	// Funnels
	r.GET("/funnels/:name", h.GetFunnel)

	// Time series
	r.GET("/timeseries/users", h.GetUsersTimeSeries)
	r.GET("/timeseries/revenue", h.GetRevenueTimeSeries)
	r.GET("/timeseries/events", h.GetEventsTimeSeries)
	r.GET("/timeseries/features", h.GetFeatureUsageTimeSeries)

	// Platform analytics
	r.GET("/platforms", h.GetPlatformBreakdown)

	// Feature usage
	r.GET("/features", h.GetFeatureUsage)

	// Events
	r.GET("/events", h.GetEvents)
	r.GET("/events/types", h.GetEventsByType)

	// Cohorts
	r.GET("/cohorts", h.GetCohortAnalysis)

	// Revenue
	r.GET("/revenue", h.GetRevenue)
}

// GetDashboard returns the main admin dashboard data
func (h *AdminHandler) GetDashboard(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := h.repo.GetMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	dau, _ := h.repo.GetDAU(ctx, now)
	wau, _ := h.repo.GetWAU(ctx, now)
	mau, _ := h.repo.GetMAU(ctx, now)
	platformBreakdown, _ := h.repo.GetPlatformBreakdown(ctx)

	c.JSON(http.StatusOK, gin.H{
		"metrics":           metrics,
		"dau":               dau,
		"wau":               wau,
		"mau":               mau,
		"platformBreakdown": platformBreakdown,
		"generatedAt":       time.Now(),
	})
}

// GetMetrics returns current metrics
func (h *AdminHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	metrics, err := h.repo.GetMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetActiveUsers returns active users data
func (h *AdminHandler) GetActiveUsers(c *gin.Context) {
	ctx := c.Request.Context()

	now := time.Now()
	dau, _ := h.repo.GetDAU(ctx, now)
	wau, _ := h.repo.GetWAU(ctx, now)
	mau, _ := h.repo.GetMAU(ctx, now)

	c.JSON(http.StatusOK, gin.H{
		"dau":  dau,
		"wau":  wau,
		"mau":  mau,
		"date": now,
	})
}

// GetChurnedUsers returns users who haven't logged in for N days
func (h *AdminHandler) GetChurnedUsers(c *gin.Context) {
	ctx := c.Request.Context()

	days, _ := strconv.Atoi(c.Query("days"))
	if days == 0 {
		days = 30
	}

	users, err := h.repo.GetChurnedUsers(ctx, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
		"days":  days,
	})
}

// GetTopUsers returns top users by a metric
func (h *AdminHandler) GetTopUsers(c *gin.Context) {
	ctx := c.Request.Context()

	metric := c.DefaultQuery("metric", "events")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, err := h.repo.GetTopUsers(ctx, metric, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"metric": metric,
	})
}

// GetUserProfile returns a user's analytics profile
func (h *AdminHandler) GetUserProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	profile, err := h.repo.GetUserProfile(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get user's recent events
	filter := domain.NewEventFilter().
		WithUserID(userID).
		WithPagination(50, 0)
	events, _ := h.repo.GetEvents(ctx, filter)

	c.JSON(http.StatusOK, gin.H{
		"profile":      profile,
		"recentEvents": events,
	})
}

// GetFunnel returns funnel analysis data
func (h *AdminHandler) GetFunnel(c *gin.Context) {
	ctx := c.Request.Context()

	name := c.Param("name")
	from, to := parseTimeRangeGin(c)

	funnel, err := h.repo.GetFunnel(ctx, name, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, funnel)
}

// GetUsersTimeSeries returns users time series data
func (h *AdminHandler) GetUsersTimeSeries(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)
	granularity := c.DefaultQuery("granularity", "day")

	series, err := h.repo.GetActiveUsersTimeSeries(ctx, from, to, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

// GetRevenueTimeSeries returns revenue time series data
func (h *AdminHandler) GetRevenueTimeSeries(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)

	series, err := h.repo.GetRevenueTimeSeries(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

// GetEventsTimeSeries returns events time series data
func (h *AdminHandler) GetEventsTimeSeries(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)
	eventType := domain.EventType(c.Query("type"))

	series, err := h.repo.GetEventsTimeSeries(ctx, eventType, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

// GetFeatureUsageTimeSeries returns feature usage time series
func (h *AdminHandler) GetFeatureUsageTimeSeries(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)
	feature := c.Query("feature")

	series, err := h.repo.GetFeatureUsageTimeSeries(ctx, feature, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, series)
}

// GetPlatformBreakdown returns platform usage breakdown
func (h *AdminHandler) GetPlatformBreakdown(c *gin.Context) {
	ctx := c.Request.Context()

	breakdown, err := h.repo.GetPlatformBreakdown(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate percentages
	var total int64
	for _, count := range breakdown {
		total += count
	}

	c.JSON(http.StatusOK, gin.H{
		"platforms": breakdown,
		"total":     total,
	})
}

// GetFeatureUsage returns feature usage data
func (h *AdminHandler) GetFeatureUsage(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)

	usage, err := h.repo.GetFeatureUsage(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"features": usage,
	})
}

// GetEvents returns filtered events
func (h *AdminHandler) GetEvents(c *gin.Context) {
	ctx := c.Request.Context()

	filter := domain.NewEventFilter()

	// Parse filters
	if userID := c.Query("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filter = filter.WithUserID(uid)
		}
	}

	if eventType := c.Query("type"); eventType != "" {
		filter = filter.WithTypes(domain.EventType(eventType))
	}

	from, to := parseTimeRangeGin(c)
	filter = filter.WithTimeRange(from, to)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter = filter.WithPagination(limit, offset)

	events, err := h.repo.GetEvents(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, _ := h.repo.GetEventCount(ctx, filter)

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventsByType returns event counts by type
func (h *AdminHandler) GetEventsByType(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)

	counts, err := h.repo.GetEventsByType(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": counts,
	})
}

// GetCohortAnalysis returns cohort analysis data
func (h *AdminHandler) GetCohortAnalysis(c *gin.Context) {
	ctx := c.Request.Context()

	from, to := parseTimeRangeGin(c)
	period := c.DefaultQuery("period", "weekly")

	analysis, err := h.repo.GetCohortAnalysis(ctx, from, to, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetRevenue returns revenue metrics
func (h *AdminHandler) GetRevenue(c *gin.Context) {
	ctx := c.Request.Context()

	mrr, _ := h.repo.GetMRR(ctx)
	from, to := parseTimeRangeGin(c)
	churnRate, _ := h.repo.GetChurnRate(ctx, from, to)

	c.JSON(http.StatusOK, gin.H{
		"mrr":       mrr,
		"churnRate": churnRate,
	})
}

// Helper functions

func parseTimeRangeGin(c *gin.Context) (from, to time.Time) {
	now := time.Now()
	to = now

	// Default to last 30 days
	from = now.AddDate(0, 0, -30)

	if fromStr := c.Query("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsed
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsed.Add(24*time.Hour - time.Second) // End of day
		}
	}

	return from, to
}
