package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/MuhammadLuqman-99/ads-analytics/internal/domain/analytics"
)

// PostgresRepository implements the analytics repository using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL analytics repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// TrackEvent stores a single event
func (r *PostgresRepository) TrackEvent(ctx context.Context, event *domain.Event) error {
	properties, err := json.Marshal(event.Properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	query := `
		INSERT INTO analytics_events (
			id, user_id, organization_id, session_id, event_type, properties,
			timestamp, ip_address, user_agent, referrer, country, city,
			device_type, browser, os
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID, event.UserID, event.OrganizationID, event.SessionID,
		event.Type, properties, event.Timestamp, event.IPAddress,
		event.UserAgent, event.Referrer, event.Country, event.City,
		event.DeviceType, event.Browser, event.OS,
	)

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

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		properties, _ := json.Marshal(event.Properties)

		_, err := tx.ExecContext(ctx, `
			INSERT INTO analytics_events (
				id, user_id, organization_id, session_id, event_type, properties,
				timestamp, ip_address, user_agent, referrer, country, city,
				device_type, browser, os
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
			event.ID, event.UserID, event.OrganizationID, event.SessionID,
			event.Type, properties, event.Timestamp, event.IPAddress,
			event.UserAgent, event.Referrer, event.Country, event.City,
			event.DeviceType, event.Browser, event.OS,
		)

		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}

	return tx.Commit()
}

// GetEvents retrieves events based on filter
func (r *PostgresRepository) GetEvents(ctx context.Context, filter *domain.EventFilter) ([]*domain.Event, error) {
	query, args := r.buildEventQuery(filter, false)

	rows, err := r.db.QueryxContext(ctx, query, args...)
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
	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// buildEventQuery builds a SQL query from the filter
func (r *PostgresRepository) buildEventQuery(filter *domain.EventFilter, countOnly bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.OrganizationID != nil {
		conditions = append(conditions, fmt.Sprintf("organization_id = $%d", argIndex))
		args = append(args, *filter.OrganizationID)
		argIndex++
	}

	if filter.SessionID != "" {
		conditions = append(conditions, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, filter.SessionID)
		argIndex++
	}

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, t)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIndex))
		args = append(args, *filter.From)
		argIndex++
	}

	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIndex))
		args = append(args, *filter.To)
		argIndex++
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
func (r *PostgresRepository) scanEvent(rows *sqlx.Rows) (*domain.Event, error) {
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
	query := `
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
		WHERE u.id = $1
	`

	var profile domain.UserProfile
	err := r.db.GetContext(ctx, &profile, query, userID)
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
	query := `SELECT DISTINCT platform FROM platform_connections WHERE user_id = $1 AND status = 'active'`
	var platforms []string
	err := r.db.SelectContext(ctx, &platforms, query, userID)
	return platforms, err
}

// UpdateUserProfile updates a user's analytics profile
func (r *PostgresRepository) UpdateUserProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		INSERT INTO analytics_profiles (user_id, first_seen_at, last_seen_at, total_events, total_sessions, properties)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			total_events = EXCLUDED.total_events,
			total_sessions = EXCLUDED.total_sessions,
			properties = EXCLUDED.properties
	`

	properties, _ := json.Marshal(profile.Properties)
	_, err := r.db.ExecContext(ctx, query,
		profile.UserID, profile.FirstSeenAt, profile.LastSeenAt,
		profile.TotalEvents, profile.TotalSessions, properties,
	)
	return err
}

// updateUserLastSeen updates the last seen timestamp for a user
func (r *PostgresRepository) updateUserLastSeen(ctx context.Context, userID uuid.UUID, timestamp time.Time) {
	query := `
		INSERT INTO analytics_profiles (user_id, first_seen_at, last_seen_at, total_events)
		VALUES ($1, $2, $2, 1)
		ON CONFLICT (user_id) DO UPDATE SET
			last_seen_at = GREATEST(analytics_profiles.last_seen_at, EXCLUDED.last_seen_at),
			total_events = analytics_profiles.total_events + 1
	`
	r.db.ExecContext(ctx, query, userID, timestamp)
}

// GetMetrics retrieves current metrics
func (r *PostgresRepository) GetMetrics(ctx context.Context) (*domain.Metrics, error) {
	metrics := &domain.Metrics{
		CalculatedAt: time.Now(),
	}

	// Total users
	r.db.GetContext(ctx, &metrics.TotalUsers, "SELECT COUNT(*) FROM users")

	// DAU, WAU, MAU
	now := time.Now()
	metrics.ActiveUsersDAU, _ = r.GetDAU(ctx, now)
	metrics.ActiveUsersWAU, _ = r.GetWAU(ctx, now)
	metrics.ActiveUsersMAU, _ = r.GetMAU(ctx, now)

	// New users
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	r.db.GetContext(ctx, &metrics.NewUsersToday, "SELECT COUNT(*) FROM users WHERE created_at >= $1", today)
	r.db.GetContext(ctx, &metrics.NewUsersThisWeek, "SELECT COUNT(*) FROM users WHERE created_at >= $1", weekAgo)
	r.db.GetContext(ctx, &metrics.NewUsersThisMonth, "SELECT COUNT(*) FROM users WHERE created_at >= $1", monthAgo)

	// Funnel metrics
	r.db.GetContext(ctx, &metrics.RegisteredUsers, "SELECT COUNT(*) FROM users")
	r.db.GetContext(ctx, &metrics.ConnectedUsers, "SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE status = 'active'")
	r.db.GetContext(ctx, &metrics.ActiveUsers, `
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= NOW() - INTERVAL '30 days'
	`)
	r.db.GetContext(ctx, &metrics.PaidUsers, "SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active' AND plan_name != 'free'")

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
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= $1 AND timestamp < $2 AND user_id IS NOT NULL
	`, start, end)
	return count, err
}

// GetWAU returns Weekly Active Users
func (r *PostgresRepository) GetWAU(ctx context.Context, date time.Time) (int64, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	start := end.AddDate(0, 0, -7)

	var count int64
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= $1 AND timestamp < $2 AND user_id IS NOT NULL
	`, start, end)
	return count, err
}

// GetMAU returns Monthly Active Users
func (r *PostgresRepository) GetMAU(ctx context.Context, date time.Time) (int64, error) {
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
	start := end.AddDate(0, -1, 0)

	var count int64
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(DISTINCT user_id) FROM analytics_events
		WHERE timestamp >= $1 AND timestamp < $2 AND user_id IS NOT NULL
	`, start, end)
	return count, err
}

// GetActiveUsersTimeSeries returns active users over time
func (r *PostgresRepository) GetActiveUsersTimeSeries(ctx context.Context, from, to time.Time, granularity string) (*domain.TimeSeries, error) {
	interval := "1 day"
	truncate := "day"
	switch granularity {
	case "hour":
		interval = "1 hour"
		truncate = "hour"
	case "week":
		interval = "1 week"
		truncate = "week"
	case "month":
		interval = "1 month"
		truncate = "month"
	}

	query := fmt.Sprintf(`
		SELECT date_trunc('%s', timestamp) as ts, COUNT(DISTINCT user_id) as value
		FROM analytics_events
		WHERE timestamp >= $1 AND timestamp <= $2 AND user_id IS NOT NULL
		GROUP BY ts
		ORDER BY ts
	`, truncate)

	rows, err := r.db.QueryxContext(ctx, query, from, to)
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
	// Implementation depends on metric type
	return r.GetActiveUsersTimeSeries(ctx, from, to, "day")
}

// GetFunnel returns conversion funnel data
func (r *PostgresRepository) GetFunnel(ctx context.Context, name string, from, to time.Time) (*domain.Funnel, error) {
	funnel := &domain.Funnel{Name: name}

	// Define funnel steps based on name
	var steps []struct {
		name  string
		query string
	}

	switch name {
	case "activation":
		steps = []struct {
			name  string
			query string
		}{
			{"Registered", "SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at <= $2"},
			{"Email Verified", "SELECT COUNT(*) FROM users WHERE email_verified = true AND created_at >= $1 AND created_at <= $2"},
			{"Platform Connected", "SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE created_at >= $1 AND created_at <= $2"},
			{"Dashboard Viewed", "SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE event_type = 'dashboard_viewed' AND timestamp >= $1 AND timestamp <= $2"},
			{"Paid Conversion", "SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'active' AND plan_name != 'free' AND created_at >= $1 AND created_at <= $2"},
		}
	default:
		steps = []struct {
			name  string
			query string
		}{
			{"Registered", "SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at <= $2"},
			{"Connected", "SELECT COUNT(DISTINCT user_id) FROM platform_connections WHERE created_at >= $1 AND created_at <= $2"},
			{"Active", "SELECT COUNT(DISTINCT user_id) FROM analytics_events WHERE timestamp >= $1 AND timestamp <= $2"},
		}
	}

	var totalCount int64
	for i, step := range steps {
		var count int64
		r.db.GetContext(ctx, &count, step.query, from, to)

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
	// Simplified cohort analysis
	analysis := &domain.CohortAnalysis{Period: period}

	truncate := "week"
	if period == "monthly" {
		truncate = "month"
	}

	query := fmt.Sprintf(`
		WITH cohorts AS (
			SELECT
				user_id,
				date_trunc('%s', created_at) as cohort_date
			FROM users
			WHERE created_at >= $1 AND created_at <= $2
		),
		activity AS (
			SELECT
				user_id,
				date_trunc('%s', timestamp) as activity_date
			FROM analytics_events
			WHERE user_id IS NOT NULL AND timestamp >= $1
		)
		SELECT
			c.cohort_date,
			COUNT(DISTINCT c.user_id) as cohort_size,
			a.activity_date,
			COUNT(DISTINCT a.user_id) as active_users
		FROM cohorts c
		LEFT JOIN activity a ON c.user_id = a.user_id
		GROUP BY c.cohort_date, a.activity_date
		ORDER BY c.cohort_date, a.activity_date
	`, truncate, truncate)

	// Execute and process results
	// Simplified return for now
	return analysis, nil
}

// GetMRR returns Monthly Recurring Revenue
func (r *PostgresRepository) GetMRR(ctx context.Context) (float64, error) {
	var mrr float64
	err := r.db.GetContext(ctx, &mrr, `
		SELECT COALESCE(SUM(
			CASE plan_name
				WHEN 'pro' THEN 99
				WHEN 'business' THEN 299
				ELSE 0
			END
		), 0)
		FROM subscriptions
		WHERE status = 'active' AND plan_name != 'free'
	`)
	return mrr, err
}

// GetChurnRate returns the churn rate for a period
func (r *PostgresRepository) GetChurnRate(ctx context.Context, from, to time.Time) (float64, error) {
	var startCount, endCount, churned int64

	r.db.GetContext(ctx, &startCount, `
		SELECT COUNT(*) FROM subscriptions
		WHERE status = 'active' AND created_at < $1 AND plan_name != 'free'
	`, from)

	r.db.GetContext(ctx, &churned, `
		SELECT COUNT(*) FROM subscriptions
		WHERE status = 'cancelled' AND cancelled_at >= $1 AND cancelled_at <= $2 AND plan_name != 'free'
	`, from, to)

	if startCount == 0 {
		return 0, nil
	}

	return float64(churned) / float64(startCount) * 100, nil
}

// GetRevenueTimeSeries returns revenue over time
func (r *PostgresRepository) GetRevenueTimeSeries(ctx context.Context, from, to time.Time) (*domain.TimeSeries, error) {
	query := `
		SELECT date_trunc('day', created_at) as ts, SUM(amount) as value
		FROM payments
		WHERE created_at >= $1 AND created_at <= $2 AND status = 'succeeded'
		GROUP BY ts
		ORDER BY ts
	`

	rows, err := r.db.QueryxContext(ctx, query, from, to)
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
	query := `
		SELECT platform, COUNT(DISTINCT user_id) as count
		FROM platform_connections
		WHERE status = 'active'
		GROUP BY platform
	`

	rows, err := r.db.QueryxContext(ctx, query)
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
	query := `
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		WHERE timestamp >= $1 AND timestamp <= $2
		AND event_type IN ('dashboard_viewed', 'campaign_viewed', 'campaign_exported',
						   'report_generated', 'analytics_viewed', 'settings_updated')
		GROUP BY event_type
	`

	rows, err := r.db.QueryxContext(ctx, query, from, to)
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
	query := `
		SELECT date_trunc('day', timestamp) as ts, COUNT(*) as value
		FROM analytics_events
		WHERE timestamp >= $1 AND timestamp <= $2 AND event_type = $3
		GROUP BY ts
		ORDER BY ts
	`

	rows, err := r.db.QueryxContext(ctx, query, from, to, feature)
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
	query := `
		SELECT u.id, u.email, u.name, u.created_at,
			   COALESCE(ap.total_events, 0) as total_events
		FROM users u
		LEFT JOIN analytics_profiles ap ON u.id = ap.user_id
		ORDER BY ap.total_events DESC NULLS LAST
		LIMIT $1
	`

	rows, err := r.db.QueryxContext(ctx, query, limit)
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
	query := `
		SELECT u.id, u.email, u.name, u.created_at,
			   ap.last_seen_at
		FROM users u
		LEFT JOIN analytics_profiles ap ON u.id = ap.user_id
		WHERE ap.last_seen_at < NOW() - INTERVAL '%d days'
		   OR (ap.last_seen_at IS NULL AND u.created_at < NOW() - INTERVAL '%d days')
		ORDER BY COALESCE(ap.last_seen_at, u.created_at) ASC
	`

	rows, err := r.db.QueryxContext(ctx, fmt.Sprintf(query, days, days))
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
	query := `
		SELECT event_type, COUNT(*) as count
		FROM analytics_events
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY event_type
	`

	rows, err := r.db.QueryxContext(ctx, query, from, to)
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
	query := `
		SELECT date_trunc('day', timestamp) as ts, COUNT(*) as value
		FROM analytics_events
		WHERE timestamp >= $1 AND timestamp <= $2 AND event_type = $3
		GROUP BY ts
		ORDER BY ts
	`

	rows, err := r.db.QueryxContext(ctx, query, from, to, eventType)
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
