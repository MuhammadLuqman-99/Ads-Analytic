package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for analytics data access
type Repository interface {
	// Event operations
	TrackEvent(ctx context.Context, event *Event) error
	TrackEvents(ctx context.Context, events []*Event) error
	GetEvents(ctx context.Context, filter *EventFilter) ([]*Event, error)
	GetEventCount(ctx context.Context, filter *EventFilter) (int64, error)

	// User profile operations
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *UserProfile) error

	// Metrics operations
	GetMetrics(ctx context.Context) (*Metrics, error)
	GetMetricsHistory(ctx context.Context, metric string, from, to time.Time) (*TimeSeries, error)

	// Active users
	GetDAU(ctx context.Context, date time.Time) (int64, error)
	GetWAU(ctx context.Context, date time.Time) (int64, error)
	GetMAU(ctx context.Context, date time.Time) (int64, error)
	GetActiveUsersTimeSeries(ctx context.Context, from, to time.Time, granularity string) (*TimeSeries, error)

	// Funnel analysis
	GetFunnel(ctx context.Context, name string, from, to time.Time) (*Funnel, error)

	// Cohort analysis
	GetCohortAnalysis(ctx context.Context, from, to time.Time, period string) (*CohortAnalysis, error)

	// Revenue metrics
	GetMRR(ctx context.Context) (float64, error)
	GetChurnRate(ctx context.Context, from, to time.Time) (float64, error)
	GetRevenueTimeSeries(ctx context.Context, from, to time.Time) (*TimeSeries, error)

	// Platform metrics
	GetPlatformBreakdown(ctx context.Context) (map[string]int64, error)

	// Feature usage
	GetFeatureUsage(ctx context.Context, from, to time.Time) (map[string]int64, error)
	GetFeatureUsageTimeSeries(ctx context.Context, feature string, from, to time.Time) (*TimeSeries, error)

	// Top users
	GetTopUsers(ctx context.Context, metric string, limit int) ([]*UserProfile, error)

	// Churned users
	GetChurnedUsers(ctx context.Context, days int) ([]*UserProfile, error)

	// Event aggregations
	GetEventsByType(ctx context.Context, from, to time.Time) (map[EventType]int64, error)
	GetEventsTimeSeries(ctx context.Context, eventType EventType, from, to time.Time) (*TimeSeries, error)
}

// EventFilter contains filter options for querying events
type EventFilter struct {
	UserID         *uuid.UUID
	OrganizationID *uuid.UUID
	SessionID      string
	Types          []EventType
	From           *time.Time
	To             *time.Time
	Properties     map[string]interface{}
	Limit          int
	Offset         int
	OrderBy        string
	OrderDir       string
}

// NewEventFilter creates a new event filter with defaults
func NewEventFilter() *EventFilter {
	return &EventFilter{
		Limit:    100,
		Offset:   0,
		OrderBy:  "timestamp",
		OrderDir: "DESC",
	}
}

// WithUserID filters by user ID
func (f *EventFilter) WithUserID(userID uuid.UUID) *EventFilter {
	f.UserID = &userID
	return f
}

// WithTypes filters by event types
func (f *EventFilter) WithTypes(types ...EventType) *EventFilter {
	f.Types = types
	return f
}

// WithTimeRange filters by time range
func (f *EventFilter) WithTimeRange(from, to time.Time) *EventFilter {
	f.From = &from
	f.To = &to
	return f
}

// WithPagination sets pagination options
func (f *EventFilter) WithPagination(limit, offset int) *EventFilter {
	f.Limit = limit
	f.Offset = offset
	return f
}
