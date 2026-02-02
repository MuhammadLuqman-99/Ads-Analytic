package analytics

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents different types of analytics events
type EventType string

const (
	// User lifecycle events
	EventUserRegistered    EventType = "user_registered"
	EventUserVerified      EventType = "user_verified"
	EventUserLoggedIn      EventType = "user_logged_in"
	EventUserLoggedOut     EventType = "user_logged_out"
	EventUserChurned       EventType = "user_churned"
	EventUserReactivated   EventType = "user_reactivated"
	EventUserDeleted       EventType = "user_deleted"

	// Platform connection events
	EventPlatformConnected    EventType = "platform_connected"
	EventPlatformDisconnected EventType = "platform_disconnected"
	EventPlatformSyncStarted  EventType = "platform_sync_started"
	EventPlatformSyncComplete EventType = "platform_sync_complete"
	EventPlatformSyncFailed   EventType = "platform_sync_failed"

	// Feature usage events
	EventDashboardViewed   EventType = "dashboard_viewed"
	EventCampaignViewed    EventType = "campaign_viewed"
	EventCampaignExported  EventType = "campaign_exported"
	EventReportGenerated   EventType = "report_generated"
	EventAnalyticsViewed   EventType = "analytics_viewed"
	EventSettingsUpdated   EventType = "settings_updated"

	// Subscription events
	EventPlanUpgraded      EventType = "plan_upgraded"
	EventPlanDowngraded    EventType = "plan_downgraded"
	EventTrialStarted      EventType = "trial_started"
	EventTrialEnded        EventType = "trial_ended"
	EventSubscriptionCreated EventType = "subscription_created"
	EventSubscriptionCancelled EventType = "subscription_cancelled"
	EventPaymentSucceeded  EventType = "payment_succeeded"
	EventPaymentFailed     EventType = "payment_failed"

	// Engagement events
	EventFeatureUsed       EventType = "feature_used"
	EventPageViewed        EventType = "page_viewed"
	EventButtonClicked     EventType = "button_clicked"
)

// Event represents an analytics event
type Event struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty" db:"user_id"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty" db:"organization_id"`
	SessionID      string                 `json:"session_id,omitempty" db:"session_id"`
	Type           EventType              `json:"type" db:"event_type"`
	Properties     map[string]interface{} `json:"properties" db:"properties"`
	Timestamp      time.Time              `json:"timestamp" db:"timestamp"`

	// Context
	IPAddress      string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      string                 `json:"user_agent,omitempty" db:"user_agent"`
	Referrer       string                 `json:"referrer,omitempty" db:"referrer"`
	Country        string                 `json:"country,omitempty" db:"country"`
	City           string                 `json:"city,omitempty" db:"city"`
	DeviceType     string                 `json:"device_type,omitempty" db:"device_type"`
	Browser        string                 `json:"browser,omitempty" db:"browser"`
	OS             string                 `json:"os,omitempty" db:"os"`
}

// NewEvent creates a new analytics event
func NewEvent(eventType EventType, userID *uuid.UUID, properties map[string]interface{}) *Event {
	if properties == nil {
		properties = make(map[string]interface{})
	}

	return &Event{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       eventType,
		Properties: properties,
		Timestamp:  time.Now().UTC(),
	}
}

// WithOrganization adds organization context
func (e *Event) WithOrganization(orgID uuid.UUID) *Event {
	e.OrganizationID = &orgID
	return e
}

// WithSession adds session context
func (e *Event) WithSession(sessionID string) *Event {
	e.SessionID = sessionID
	return e
}

// WithContext adds request context (IP, user agent, etc.)
func (e *Event) WithContext(ctx *EventContext) *Event {
	if ctx == nil {
		return e
	}
	e.IPAddress = ctx.IPAddress
	e.UserAgent = ctx.UserAgent
	e.Referrer = ctx.Referrer
	e.Country = ctx.Country
	e.City = ctx.City
	e.DeviceType = ctx.DeviceType
	e.Browser = ctx.Browser
	e.OS = ctx.OS
	return e
}

// WithProperty adds a single property
func (e *Event) WithProperty(key string, value interface{}) *Event {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	e.Properties[key] = value
	return e
}

// EventContext contains request context information
type EventContext struct {
	IPAddress  string
	UserAgent  string
	Referrer   string
	Country    string
	City       string
	DeviceType string
	Browser    string
	OS         string
}

// UserProfile represents aggregated user data for analytics
type UserProfile struct {
	UserID           uuid.UUID              `json:"user_id" db:"user_id"`
	Email            string                 `json:"email" db:"email"`
	Name             string                 `json:"name" db:"name"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	FirstSeenAt      time.Time              `json:"first_seen_at" db:"first_seen_at"`
	LastSeenAt       time.Time              `json:"last_seen_at" db:"last_seen_at"`
	TotalEvents      int                    `json:"total_events" db:"total_events"`
	TotalSessions    int                    `json:"total_sessions" db:"total_sessions"`
	PlanName         string                 `json:"plan_name" db:"plan_name"`
	PlatformsConnected []string             `json:"platforms_connected" db:"platforms_connected"`
	Properties       map[string]interface{} `json:"properties" db:"properties"`
}

// Metrics represents aggregated metrics
type Metrics struct {
	// User metrics
	TotalUsers       int64   `json:"total_users"`
	ActiveUsersDAU   int64   `json:"active_users_dau"`
	ActiveUsersWAU   int64   `json:"active_users_wau"`
	ActiveUsersMAU   int64   `json:"active_users_mau"`
	NewUsersToday    int64   `json:"new_users_today"`
	NewUsersThisWeek int64   `json:"new_users_this_week"`
	NewUsersThisMonth int64  `json:"new_users_this_month"`

	// Revenue metrics
	MRR              float64 `json:"mrr"`
	ARR              float64 `json:"arr"`
	ChurnRate        float64 `json:"churn_rate"`
	ARPU             float64 `json:"arpu"`
	LTV              float64 `json:"ltv"`

	// Conversion funnel
	RegisteredUsers  int64   `json:"registered_users"`
	ConnectedUsers   int64   `json:"connected_users"`
	ActiveUsers      int64   `json:"active_users"`
	PaidUsers        int64   `json:"paid_users"`

	// Platform metrics
	PlatformBreakdown map[string]int64 `json:"platform_breakdown"`

	// Feature usage
	FeatureUsage     map[string]int64 `json:"feature_usage"`

	// Timestamp
	CalculatedAt     time.Time `json:"calculated_at"`
}

// FunnelStep represents a step in the conversion funnel
type FunnelStep struct {
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
	DropOff    float64 `json:"drop_off"`
}

// Funnel represents a conversion funnel
type Funnel struct {
	Name  string       `json:"name"`
	Steps []FunnelStep `json:"steps"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TimeSeries represents time series data
type TimeSeries struct {
	Name   string            `json:"name"`
	Points []TimeSeriesPoint `json:"points"`
}

// Cohort represents a user cohort for analysis
type Cohort struct {
	Date       time.Time   `json:"date"`
	Size       int64       `json:"size"`
	Retention  []float64   `json:"retention"` // Retention percentages by period
}

// CohortAnalysis represents cohort analysis data
type CohortAnalysis struct {
	Period  string   `json:"period"` // "daily", "weekly", "monthly"
	Cohorts []Cohort `json:"cohorts"`
}
