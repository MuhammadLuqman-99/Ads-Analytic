package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domain "github.com/MuhammadLuqman-99/ads-analytics/internal/domain/analytics"
)

// AdminHandler handles admin panel API requests
type AdminHandler struct {
	repo domain.Repository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo domain.Repository) *AdminHandler {
	return &AdminHandler{repo: repo}
}

// Routes returns the admin routes
func (h *AdminHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Dashboard
	r.Get("/dashboard", h.GetDashboard)
	r.Get("/metrics", h.GetMetrics)

	// Active users
	r.Get("/users/active", h.GetActiveUsers)
	r.Get("/users/churned", h.GetChurnedUsers)
	r.Get("/users/top", h.GetTopUsers)
	r.Get("/users/{id}", h.GetUserProfile)

	// Funnels
	r.Get("/funnels/{name}", h.GetFunnel)

	// Time series
	r.Get("/timeseries/users", h.GetUsersTimeSeries)
	r.Get("/timeseries/revenue", h.GetRevenueTimeSeries)
	r.Get("/timeseries/events", h.GetEventsTimeSeries)
	r.Get("/timeseries/features", h.GetFeatureUsageTimeSeries)

	// Platform analytics
	r.Get("/platforms", h.GetPlatformBreakdown)

	// Events
	r.Get("/events", h.GetEvents)
	r.Get("/events/types", h.GetEventsByType)

	// Cohorts
	r.Get("/cohorts", h.GetCohortAnalysis)

	return r
}

// GetDashboard returns the main admin dashboard data
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.repo.GetMetrics(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get additional data
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, -1, 0)

	// Conversion funnel
	funnel, _ := h.repo.GetFunnel(ctx, "activation", monthAgo, now)

	// Active users time series
	activeUsersSeries, _ := h.repo.GetActiveUsersTimeSeries(ctx, weekAgo, now, "day")

	// Events by type
	eventsByType, _ := h.repo.GetEventsByType(ctx, weekAgo, now)

	dashboard := map[string]interface{}{
		"metrics":            metrics,
		"funnel":             funnel,
		"active_users_chart": activeUsersSeries,
		"events_by_type":     eventsByType,
		"generated_at":       time.Now(),
	}

	respondJSON(w, dashboard)
}

// GetMetrics returns current metrics
func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.repo.GetMetrics(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, metrics)
}

// GetActiveUsers returns active users data
func (h *AdminHandler) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	now := time.Now()
	period := r.URL.Query().Get("period")

	var count int64
	var err error

	switch period {
	case "week":
		count, err = h.repo.GetWAU(ctx, now)
	case "month":
		count, err = h.repo.GetMAU(ctx, now)
	default:
		count, err = h.repo.GetDAU(ctx, now)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"count":  count,
		"period": period,
		"date":   now,
	})
}

// GetChurnedUsers returns users who haven't logged in for N days
func (h *AdminHandler) GetChurnedUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days == 0 {
		days = 30
	}

	users, err := h.repo.GetChurnedUsers(ctx, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"users": users,
		"count": len(users),
		"days":  days,
	})
}

// GetTopUsers returns top users by a metric
func (h *AdminHandler) GetTopUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "events"
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	users, err := h.repo.GetTopUsers(ctx, metric, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"users":  users,
		"metric": metric,
	})
}

// GetUserProfile returns a user's analytics profile
func (h *AdminHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	profile, err := h.repo.GetUserProfile(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user's recent events
	filter := domain.NewEventFilter().
		WithUserID(userID).
		WithPagination(50, 0)
	events, _ := h.repo.GetEvents(ctx, filter)

	respondJSON(w, map[string]interface{}{
		"profile":       profile,
		"recent_events": events,
	})
}

// GetFunnel returns funnel analysis data
func (h *AdminHandler) GetFunnel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	name := chi.URLParam(r, "name")
	from, to := parseTimeRange(r)

	funnel, err := h.repo.GetFunnel(ctx, name, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, funnel)
}

// GetUsersTimeSeries returns users time series data
func (h *AdminHandler) GetUsersTimeSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "day"
	}

	series, err := h.repo.GetActiveUsersTimeSeries(ctx, from, to, granularity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, series)
}

// GetRevenueTimeSeries returns revenue time series data
func (h *AdminHandler) GetRevenueTimeSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)

	series, err := h.repo.GetRevenueTimeSeries(ctx, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, series)
}

// GetEventsTimeSeries returns events time series data
func (h *AdminHandler) GetEventsTimeSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)
	eventType := domain.EventType(r.URL.Query().Get("type"))

	series, err := h.repo.GetEventsTimeSeries(ctx, eventType, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, series)
}

// GetFeatureUsageTimeSeries returns feature usage time series
func (h *AdminHandler) GetFeatureUsageTimeSeries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)
	feature := r.URL.Query().Get("feature")

	series, err := h.repo.GetFeatureUsageTimeSeries(ctx, feature, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, series)
}

// GetPlatformBreakdown returns platform usage breakdown
func (h *AdminHandler) GetPlatformBreakdown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	breakdown, err := h.repo.GetPlatformBreakdown(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate percentages
	var total int64
	for _, count := range breakdown {
		total += count
	}

	result := make([]map[string]interface{}, 0, len(breakdown))
	for platform, count := range breakdown {
		percentage := float64(0)
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		result = append(result, map[string]interface{}{
			"platform":   platform,
			"count":      count,
			"percentage": percentage,
		})
	}

	respondJSON(w, map[string]interface{}{
		"platforms": result,
		"total":     total,
	})
}

// GetEvents returns filtered events
func (h *AdminHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := domain.NewEventFilter()

	// Parse filters
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if uid, err := uuid.Parse(userID); err == nil {
			filter = filter.WithUserID(uid)
		}
	}

	if eventType := r.URL.Query().Get("type"); eventType != "" {
		filter = filter.WithTypes(domain.EventType(eventType))
	}

	from, to := parseTimeRange(r)
	filter = filter.WithTimeRange(from, to)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 {
		limit = 100
	}
	filter = filter.WithPagination(limit, offset)

	events, err := h.repo.GetEvents(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := h.repo.GetEventCount(ctx, filter)

	respondJSON(w, map[string]interface{}{
		"events": events,
		"total":  count,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventsByType returns event counts by type
func (h *AdminHandler) GetEventsByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)

	counts, err := h.repo.GetEventsByType(ctx, from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, counts)
}

// GetCohortAnalysis returns cohort analysis data
func (h *AdminHandler) GetCohortAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	from, to := parseTimeRange(r)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "weekly"
	}

	analysis, err := h.repo.GetCohortAnalysis(ctx, from, to, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, analysis)
}

// Helper functions

func parseTimeRange(r *http.Request) (from, to time.Time) {
	now := time.Now()
	to = now

	// Default to last 30 days
	from = now.AddDate(0, 0, -30)

	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = parsed
		}
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if parsed, err := time.Parse("2006-01-02", toStr); err == nil {
			to = parsed.Add(24*time.Hour - time.Second) // End of day
		}
	}

	return from, to
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
