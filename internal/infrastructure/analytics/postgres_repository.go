package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	domain "github.com/ads-aggregator/ads-aggregator/internal/domain/analytics"
)

// PostgresRepository implements the analytics repository using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL analytics repository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// TrackEvent stores a single event
func (r *PostgresRepository) TrackEvent(ctx context.Context, event *domain.Event) error {
	properties, err := json.Marshal(event.Properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	err = r.db.WithContext(ctx).Exec(`
		INSERT INTO analytics_events (
			id, user_id, organization_id, session_id, event_type, properties,
			timestamp, ip_address, user_agent, referrer, country, city,
			device_type, browser, os
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		event.ID, event.UserID, event.OrganizationID, event.SessionID,
		event.Type, properties, event.Timestamp, event.IPAddress,
		event.UserAgent, event.Referrer, event.Country, event.City,
		event.DeviceType, event.Browser, event.OS,
	).Error

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Update user last seen
	if event.UserID != nil {
		r.updateUserLastSeen(ctx, *event.UserID, event.Timestamp)
	}

	return nil
}

// TrackEvents stores multiple events in a batch
func (r *PostgresRepository) TrackEvents(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			properties, _ := json.Marshal(event.Properties)

			err := tx.Exec(`
				INSERT INTO analytics_events (
					id, user_id, organization_id, session_id, event_type, properties,
					timestamp, ip_address, user_agent, referrer, country, city,
					device_type, browser, os
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				event.ID, event.UserID, event.OrganizationID, event.SessionID,
				event.Type, properties, event.Timestamp, event.IPAddress,
				event.UserAgent, event.Referrer, event.Country, event.City,
				event.DeviceType, event.Browser, event.OS,
			).Error

			if err != nil {
				return fmt.Errorf("failed to insert event: %w", err)
			}
		}
		return nil
	})
}

// GetEvents retrieves events based on filter
func (r *PostgresRepository) GetEvents(ctx context.Context, filter *domain.EventFilter) ([]*domain.Event, error) {
	query, args := r.buildEventQuery(filter, false)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventCount returns the count of events matching the filter
func (r *PostgresRepository) GetEventCount(ctx context.Context, filter *domain.EventFilter) (int64, error) {
	query, args := r.buildEventQuery(filter, true)

	var count int64
	err := r.db.WithContext(ctx).Raw(query, args...).Scan(&count).Error
	return count, err
}

// buildEventQuery builds a SQL query from the filter
func (r *PostgresRepository) buildEventQuery(filter *domain.EventFilter, countOnly bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.OrganizationID != nil {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, *filter.OrganizationID)
	}

	if filter.SessionID != "" {
		conditions = append(conditions, "session_id = ?")
		args = append(args, filter.SessionID)
	}

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = "?"
			args = append(args, t)
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.From != nil {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, *filter.From)
	}

	if filter.To != nil {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, *filter.To)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	if countOnly {
		return fmt.Sprintf("SELECT COUNT(*) FROM analytics_events %s", whereClause), args
	}

	orderBy := "timestamp"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}
	orderDir := "DESC"
	if filter.OrderDir != "" {
		orderDir = filter.OrderDir
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, organization_id, session_id, event_type, properties,
			   timestamp, ip_address, user_agent, referrer, country, city,
			   device_type, browser, os
		FROM analytics_events %s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, whereClause, orderBy, orderDir, filter.Limit, filter.Offset)

	return query, args
}

// scanEvent scans a row into an Event
func (r *PostgresRepository) scanEvent(rows *sql.Rows) (*domain.Event, error) {
	var event domain.Event
	var properties []byte
	var userID, orgID sql.NullString

	err := rows.Scan(
		&event.ID, &userID, &orgID, &event.SessionID,
		&event.Type, &properties, &event.Timestamp,
		&event.IPAddress, &event.UserAgent, &event.Referrer,
		&event.Country, &event.City, &event.DeviceType,
		&event.Browser, &event.OS,
	)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		uid, _ := uuid.Parse(userID.String)
		event.UserID = &uid
	}
	if orgID.Valid {
		oid, _ := uuid.Parse(orgID.String)
		event.OrganizationID = &oid
	}

	if len(properties) > 0 {
		json.Unmarshal(properties, &event.Properties)
	}

	return &event, nil
}

// GetUserProfile retrieves a user's analytics profile
func (r *PostgresRepository) GetUserProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	var profile domain.UserProfile

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			u.id as user_id,
			u.email,
			u.name,
			u.created_at,
			COALESCE(ap.first_seen_at, u.created_at) as first_seen_at,
			COALESCE(ap.last_seen_at, u.created_at) as last_seen_at,
			COALESCE(ap.total_events, 0) as total_events,
			COALESCE(ap.total_sessions, 0) as total_sessions,
			COALESCE(s.plan_name, 'free') as plan_name
		FROM users u
		LEFT JOIN analytics_profiles ap ON u.id = ap.user_id
		LEFT JOIN subscriptions s ON u.id = s.user_id AND s.status = 'active'
		WHERE u.id = ?
	`, userID).Scan(&profile).Error

	if err != nil {
		return nil, err
	}

	// Get connected platforms
	platforms, _ := r.getUserPlatforms(ctx, userID)
	profile.PlatformsConnected = platforms

	return &profile, nil
}

// getUserPlatforms gets the platforms connected by a user
func (r *PostgresRepository) getUserPlatforms(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var platforms []string
	err := r.db.WithContext(ctx).Raw(
		"SELECT DISTINCT platform FROM platform_connections WHERE user_id = ? AND status = 'active'",
		userID,
	).Scan(&platforms).Error
	return platforms, err
}

// UpdateUserProfile updates a user's analytics profile
func (r *PostgresRepository) UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error {
	properties, _ := json.Marshal(profile.Properties)
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO analytics_profiles (user_id, first_seen_at, last_seen_at, total_events, total_sessions, properties)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			total_events = EXCLUDED.total_events,
			total_sessions = EXCLUDED.total_sessions,
			properties = EXCLUDED.properties
	`, profile.UserID, profile.FirstSeenAt, profile.LastSeenAt,
		profile.TotalEvents, profile.TotalSessions, properties,
	).Error
}

// updateUserLastSeen updates the last seen timestamp for a user
func (r *PostgresRepository) updateUserLastSeen(ctx context.Context, userID uuid.UUID, timestamp time.Time) {
	r.db.WithContext(ctx).Exec(`
		INSERT INTO analytics_profiles (user_id, first_seen_at, last_seen_at, total_events)
		VALUES (?, ?, ?, 1)
		ON CONFLICT (user_id) DO UPDATE SET
			last_seen_at = GREATEST(analytics_profiles.last_seen_at, EXCLUDED.last_seen_at),
			total_events = analytics_profiles.total_events + 1
	`, userID, timestamp, timestamp)
}

// GetMetrics retrieves current metrics
func (r *PostgresRepository) GetMetrics(ctx context.Context) (*domain.Metrics, error) {
	metrics := &domain.Metrics{
		CalculatedAt: time.Now(),
	}

	// Total users
	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM users").Scan(&metrics.TotalUsers)

	// DAU, WAU, MAU
	now := time.Now()
	metrics.ActiveUsersDAU, _ = r.GetDAU(ctx, now)
	metrics.ActiveUsersWAU, _ = r.GetWAU(ctx, now)
	metrics.ActiveUsersMAU, _ = r.GetMAU(ctx, now)

	// New users
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM users WHERE created_at >= ?", today).Scan(&metrics.NewUsersToday)
	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM users WHERE created_at >= ?", weekAgo).Scan(&metrics.NewUsersThisWeek)
	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM users WHERE created_at >= ?", monthAgo).Scan(&metrics.NewUsersThisMonth)

	// Funnel metrics
	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM users").Scan(&metrics.RegisteredUsers)
	r.db.WithContext(ctx).Raw("SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE status = 'active'").Scan(&metrics.ConnectedUsers)
	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= NOW() - INTERVAL '30 days'
	`).Scan(&metrics.ActiveUsers)
	r.db.WithContext(ctx).Raw("SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active' AND plan_name != 'free'").Scan(&metrics.PaidUsers)

	// Revenue
	metrics.MRR, _ = r.GetMRR(ctx)
	metrics.ARR = metrics.MRR * 12
	if metrics.PaidUsers > 0 {
		metrics.ARPU = metrics.MRR / float64(metrics.PaidUsers)
	}
	metrics.ChurnRate, _ = r.GetChurnRate(ctx, monthAgo, now)

	// Platform breakdown
	metrics.PlatformBreakdown, _ = r.GetPlatformBreakdown(ctx)

	// Feature usage
	metrics.FeatureUsage, _ = r.GetFeatureUsage(ctx, weekAgo, now)

	return metrics, nil
}

// GetDAU returns Daily Active Users
func (r *PostgresRepository) GetDAU(ctx context.Context, date time.Time) (int64, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= ? AND timestamp < ? AND user_id IS NOT NULL
	`, start, end).Scan(&count).Error
	return count, err
}

// GetWAU returns Weekly Active Users
func (r *PostgresRepository) GetWAU(ctx context.Context, date time.Time) (int64, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	start := end.AddDate(0, 0, -7)

	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= ? AND timestamp < ? AND user_id IS NOT NULL
	`, start, end).Scan(&count).Error
	return count, err
}

// GetMAU returns Monthly Active Users
func (r *PostgresRepository) GetMAU(ctx context.Context, date time.Time) (int64, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	start := end.AddDate(0, -1, 0)

	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= ? AND timestamp < ? AND user_id IS NOT NULL
	`, start, end).Scan(&count).Error
	return count, err
}

// GetActiveUsersTimeSeries returns active users over time
func (r *PostgresRepository) GetActiveUsersTimeSeries(ctx context.Context, from, to time.Time, granularity string) (*domain.TimeSeries, error) {
	truncate := "day"
	switch granularity {
	case "hour":
		truncate = "hour"
	case "week":
		truncate = "week"
	case "month":
		truncate = "month"
	}

	query := fmt.Sprintf(`
		SELECT date_trunc('%s', timestamp) as ts, COUNT(DISTINCT user_id) as value
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ? AND user_id IS NOT NULL
		GROUP BY ts
		ORDER BY ts
	`, truncate)

	rows, err := r.db.WithContext(ctx).Raw(query, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	series := &domain.TimeSeries{Name: "active_users"}
	for rows.Next() {
		var point domain.TimeSeriesPoint
		rows.Scan(&point.Timestamp, &point.Value)
		series.Points = append(series.Points, point)
	}

	return series, nil
}

// GetMetricsHistory returns historical metrics
func (r *PostgresRepository) GetMetricsHistory(ctx context.Context, metric string, from, to time.Time) (*domain.TimeSeries, error) {
	return r.GetActiveUsersTimeSeries(ctx, from, to, "day")
}

// GetFunnel returns conversion funnel data
func (r *PostgresRepository) GetFunnel(ctx context.Context, name string, from, to time.Time) (*domain.Funnel, error) {
	funnel := &domain.Funnel{Name: name}

	type stepDef struct {
		name  string
		query string
	}

	var steps []stepDef

	switch name {
	case "activation":
		steps = []stepDef{
			{"Registered", "SELECT COUNT(*) FROM users WHERE created_at >= ? AND created_at <= ?"},
			{"Email Verified", "SELECT COUNT(*) FROM users WHERE email_verified = true AND created_at >= ? AND created_at <= ?"},
			{"Platform Connected", "SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE created_at >= ? AND created_at <= ?"},
			{"Dashboard Viewed", "SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE event_type = 'dashboard_viewed' AND timestamp >= ? AND timestamp <= ?"},
			{"Paid Conversion", "SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active' AND plan_name != 'free' AND created_at >= ? AND created_at <= ?"},
		}
	default:
		steps = []stepDef{
			{"Registered", "SELECT COUNT(*) FROM users WHERE created_at >= ? AND created_at <= ?"},
			{"Connected", "SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE created_at >= ? AND created_at <= ?"},
			{"Active", "SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE timestamp >= ? AND timestamp <= ?"},
		}
	}

	var totalCount int64
	for i, step := range steps {
		var count int64
		r.db.WithContext(ctx).Raw(step.query, from, to).Scan(&count)

		funnelStep := domain.FunnelStep{
			Name:  step.name,
			Count: count,
		}

		if i == 0 {
			totalCount = count
			funnelStep.Percentage = 100
			funnelStep.DropOff = 0
		} else if totalCount > 0 {
			funnelStep.Percentage = float64(count) / float64(totalCount) * 100
			prevCount := funnel.Steps[i-1].Count
			if prevCount > 0 {
				funnelStep.DropOff = float64(prevCount-count) / float64(prevCount) * 100
			}
		}

		funnel.Steps = append(funnel.Steps, funnelStep)
	}

	return funnel, nil
}

// GetCohortAnalysis returns cohort retention analysis
func (r *PostgresRepository) GetCohortAnalysis(ctx context.Context, from, to time.Time, period string) (*domain.CohortAnalysis, error) {
	analysis := &domain.CohortAnalysis{Period: period}
	return analysis, nil
}

// GetMRR returns Monthly Recurring Revenue
func (r *PostgresRepository) GetMRR(ctx context.Context) (float64, error) {
	var mrr float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(
			CASE plan_name
				WHEN 'pro' THEN 99
				WHEN 'business' THEN 299
				ELSE 0
			END
		), 0)
		FROM subscriptions
		WHERE status = 'active' AND plan_name != 'free'
	`).Scan(&mrr).Error
	return mrr, err
}

// GetChurnRate returns the churn rate for a period
func (r *PostgresRepository) GetChurnRate(ctx context.Context, from, to time.Time) (float64, error) {
	var startCount, churned int64

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM subscriptions
		WHERE status = 'active' AND created_at < ? AND plan_name != 'free'
	`, from).Scan(&startCount)

	r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM subscriptions
		WHERE status = 'cancelled' AND cancelled_at >= ? AND cancelled_at <= ? AND plan_name != 'free'
	`, from, to).Scan(&churned)

	if startCount == 0 {
		return 0, nil
	}

	return float64(churned) / float64(startCount) * 100, nil
}

// GetRevenueTimeSeries returns revenue over time
func (r *PostgresRepository) GetRevenueTimeSeries(ctx context.Context, from, to time.Time) (*domain.TimeSeries, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT date_trunc('day', created_at) as ts, SUM(amount) as value
		FROM payments
		WHERE created_at >= ? AND created_at <= ? AND status = 'succeeded'
		GROUP BY ts
		ORDER BY ts
	`, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	series := &domain.TimeSeries{Name: "revenue"}
	for rows.Next() {
		var point domain.TimeSeriesPoint
		rows.Scan(&point.Timestamp, &point.Value)
		series.Points = append(series.Points, point)
	}

	return series, nil
}

// GetPlatformBreakdown returns user count per platform
func (r *PostgresRepository) GetPlatformBreakdown(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT platform, COUNT(DISTINCT user_id) as count
		FROM platform_connections
		WHERE status = 'active'
		GROUP BY platform
	`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdown := make(map[string]int64)
	for rows.Next() {
		var platform string
		var count int64
		rows.Scan(&platform, &count)
		breakdown[platform] = count
	}

	return breakdown, nil
}

// GetFeatureUsage returns feature usage counts
func (r *PostgresRepository) GetFeatureUsage(ctx context.Context, from, to time.Time) (map[string]int64, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ?
		AND event_type IN ('dashboard_viewed', 'campaign_viewed', 'campaign_exported',
						   'report_generated', 'analytics_viewed', 'settings_updated')
		GROUP BY event_type
	`, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := make(map[string]int64)
	for rows.Next() {
		var eventType string
		var count int64
		rows.Scan(&eventType, &count)
		usage[eventType] = count
	}

	return usage, nil
}

// GetFeatureUsageTimeSeries returns feature usage over time
func (r *PostgresRepository) GetFeatureUsageTimeSeries(ctx context.Context, feature string, from, to time.Time) (*domain.TimeSeries, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT date_trunc('day', timestamp) as ts, COUNT(*) as value
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ? AND event_type = ?
		GROUP BY ts
		ORDER BY ts
	`, from, to, feature).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	series := &domain.TimeSeries{Name: feature}
	for rows.Next() {
		var point domain.TimeSeriesPoint
		rows.Scan(&point.Timestamp, &point.Value)
		series.Points = append(series.Points, point)
	}

	return series, nil
}

// GetTopUsers returns top users by a metric
func (r *PostgresRepository) GetTopUsers(ctx context.Context, metric string, limit int) ([]*domain.UserProfile, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.email, u.name, u.created_at,
			   COALESCE(ap.total_events, 0) as total_events
		FROM users u
		LEFT JOIN analytics_profiles ap ON u.id = ap.user_id
		ORDER BY ap.total_events DESC NULLS LAST
		LIMIT ?
	`, limit).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*domain.UserProfile
	for rows.Next() {
		var p domain.UserProfile
		rows.Scan(&p.UserID, &p.Email, &p.Name, &p.CreatedAt, &p.TotalEvents)
		profiles = append(profiles, &p)
	}

	return profiles, nil
}

// GetChurnedUsers returns users who haven't logged in for N days
func (r *PostgresRepository) GetChurnedUsers(ctx context.Context, days int) ([]*domain.UserProfile, error) {
	query := fmt.Sprintf(`
		SELECT u.id, u.email, u.name, u.created_at,
			   ap.last_seen_at
		FROM users u
		LEFT JOIN analytics_profiles ap ON u.id = ap.user_id
		WHERE ap.last_seen_at < NOW() - INTERVAL '%d days'
		   OR (ap.last_seen_at IS NULL AND u.created_at < NOW() - INTERVAL '%d days')
		ORDER BY COALESCE(ap.last_seen_at, u.created_at) ASC
	`, days, days)

	rows, err := r.db.WithContext(ctx).Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*domain.UserProfile
	for rows.Next() {
		var p domain.UserProfile
		var lastSeen sql.NullTime
		rows.Scan(&p.UserID, &p.Email, &p.Name, &p.CreatedAt, &lastSeen)
		if lastSeen.Valid {
			p.LastSeenAt = lastSeen.Time
		}
		profiles = append(profiles, &p)
	}

	return profiles, nil
}

// GetEventsByType returns event counts by type
func (r *PostgresRepository) GetEventsByType(ctx context.Context, from, to time.Time) (map[domain.EventType]int64, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ?
		GROUP BY event_type
	`, from, to).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.EventType]int64)
	for rows.Next() {
		var eventType domain.EventType
		var count int64
		rows.Scan(&eventType, &count)
		counts[eventType] = count
	}

	return counts, nil
}

// GetEventsTimeSeries returns events over time for a specific type
func (r *PostgresRepository) GetEventsTimeSeries(ctx context.Context, eventType domain.EventType, from, to time.Time) (*domain.TimeSeries, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT date_trunc('day', timestamp) as ts, COUNT(*) as value
		FROM analytics_events
		WHERE timestamp >= ? AND timestamp <= ? AND event_type = ?
		GROUP BY ts
		ORDER BY ts
	`, from, to, eventType).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	series := &domain.TimeSeries{Name: string(eventType)}
	for rows.Next() {
		var point domain.TimeSeriesPoint
		rows.Scan(&point.Timestamp, &point.Value)
		series.Points = append(series.Points, point)
	}

	return series, nil
}
