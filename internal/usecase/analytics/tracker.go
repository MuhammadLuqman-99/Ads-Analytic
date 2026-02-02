package analytics

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	domain "github.com/MuhammadLuqman-99/ads-analytics/internal/domain/analytics"
)

// Tracker handles analytics event tracking
type Tracker struct {
	repo         domain.Repository
	buffer       []*domain.Event
	bufferMu     sync.Mutex
	bufferSize   int
	flushTicker  *time.Ticker
	stopCh       chan struct{}
	enabled      bool
}

// TrackerConfig contains tracker configuration
type TrackerConfig struct {
	BufferSize    int           // Number of events to buffer before flushing
	FlushInterval time.Duration // How often to flush the buffer
	Enabled       bool          // Whether tracking is enabled
}

// DefaultTrackerConfig returns default configuration
func DefaultTrackerConfig() *TrackerConfig {
	return &TrackerConfig{
		BufferSize:    100,
		FlushInterval: 10 * time.Second,
		Enabled:       true,
	}
}

// NewTracker creates a new analytics tracker
func NewTracker(repo domain.Repository, config *TrackerConfig) *Tracker {
	if config == nil {
		config = DefaultTrackerConfig()
	}

	t := &Tracker{
		repo:       repo,
		buffer:     make([]*domain.Event, 0, config.BufferSize),
		bufferSize: config.BufferSize,
		stopCh:     make(chan struct{}),
		enabled:    config.Enabled,
	}

	// Start flush ticker
	if config.Enabled {
		t.flushTicker = time.NewTicker(config.FlushInterval)
		go t.flushLoop()
	}

	return t
}

// Track records an analytics event
func (t *Tracker) Track(event *domain.Event) {
	if !t.enabled {
		return
	}

	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	t.buffer = append(t.buffer, event)

	// Flush if buffer is full
	if len(t.buffer) >= t.bufferSize {
		go t.flush()
	}
}

// TrackEvent is a convenience method to track an event
func (t *Tracker) TrackEvent(eventType domain.EventType, userID *uuid.UUID, properties map[string]interface{}) {
	event := domain.NewEvent(eventType, userID, properties)
	t.Track(event)
}

// flushLoop periodically flushes the buffer
func (t *Tracker) flushLoop() {
	for {
		select {
		case <-t.flushTicker.C:
			t.flush()
		case <-t.stopCh:
			t.flush() // Final flush
			return
		}
	}
}

// flush writes buffered events to the repository
func (t *Tracker) flush() {
	t.bufferMu.Lock()
	if len(t.buffer) == 0 {
		t.bufferMu.Unlock()
		return
	}

	events := t.buffer
	t.buffer = make([]*domain.Event, 0, t.bufferSize)
	t.bufferMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := t.repo.TrackEvents(ctx, events); err != nil {
		log.Printf("[Analytics] Failed to flush events: %v", err)
		// Re-add events to buffer on failure
		t.bufferMu.Lock()
		t.buffer = append(events, t.buffer...)
		t.bufferMu.Unlock()
	} else {
		log.Printf("[Analytics] Flushed %d events", len(events))
	}
}

// Stop stops the tracker and flushes remaining events
func (t *Tracker) Stop() {
	if t.flushTicker != nil {
		t.flushTicker.Stop()
	}
	close(t.stopCh)
}

// --- Convenience methods for common events ---

// TrackUserRegistered tracks a user registration event
func (t *Tracker) TrackUserRegistered(userID uuid.UUID, email, name, source string) {
	t.TrackEvent(domain.EventUserRegistered, &userID, map[string]interface{}{
		"email":  email,
		"name":   name,
		"source": source,
	})
}

// TrackUserLoggedIn tracks a user login event
func (t *Tracker) TrackUserLoggedIn(userID uuid.UUID, method string) {
	t.TrackEvent(domain.EventUserLoggedIn, &userID, map[string]interface{}{
		"method": method,
	})
}

// TrackPlatformConnected tracks a platform connection event
func (t *Tracker) TrackPlatformConnected(userID uuid.UUID, platform string, isFirst bool) {
	t.TrackEvent(domain.EventPlatformConnected, &userID, map[string]interface{}{
		"platform": platform,
		"is_first": isFirst,
	})
}

// TrackPlatformDisconnected tracks a platform disconnection event
func (t *Tracker) TrackPlatformDisconnected(userID uuid.UUID, platform, reason string) {
	t.TrackEvent(domain.EventPlatformDisconnected, &userID, map[string]interface{}{
		"platform": platform,
		"reason":   reason,
	})
}

// TrackDashboardViewed tracks a dashboard view event
func (t *Tracker) TrackDashboardViewed(userID uuid.UUID, dashboardType string) {
	t.TrackEvent(domain.EventDashboardViewed, &userID, map[string]interface{}{
		"dashboard_type": dashboardType,
	})
}

// TrackCampaignExported tracks a campaign export event
func (t *Tracker) TrackCampaignExported(userID uuid.UUID, format string, count int) {
	t.TrackEvent(domain.EventCampaignExported, &userID, map[string]interface{}{
		"format": format,
		"count":  count,
	})
}

// TrackPlanUpgraded tracks a plan upgrade event
func (t *Tracker) TrackPlanUpgraded(userID uuid.UUID, fromPlan, toPlan string, amount float64) {
	t.TrackEvent(domain.EventPlanUpgraded, &userID, map[string]interface{}{
		"from_plan": fromPlan,
		"to_plan":   toPlan,
		"amount":    amount,
	})
}

// TrackPlanDowngraded tracks a plan downgrade event
func (t *Tracker) TrackPlanDowngraded(userID uuid.UUID, fromPlan, toPlan, reason string) {
	t.TrackEvent(domain.EventPlanDowngraded, &userID, map[string]interface{}{
		"from_plan": fromPlan,
		"to_plan":   toPlan,
		"reason":    reason,
	})
}

// TrackUserChurned tracks a user churn event
func (t *Tracker) TrackUserChurned(userID uuid.UUID, lastSeenDays int, reason string) {
	t.TrackEvent(domain.EventUserChurned, &userID, map[string]interface{}{
		"last_seen_days": lastSeenDays,
		"reason":         reason,
	})
}

// TrackFeatureUsed tracks generic feature usage
func (t *Tracker) TrackFeatureUsed(userID uuid.UUID, feature string, metadata map[string]interface{}) {
	props := map[string]interface{}{
		"feature": feature,
	}
	for k, v := range metadata {
		props[k] = v
	}
	t.TrackEvent(domain.EventFeatureUsed, &userID, props)
}

// TrackPageViewed tracks a page view
func (t *Tracker) TrackPageViewed(userID *uuid.UUID, path, referrer string) {
	t.TrackEvent(domain.EventPageViewed, userID, map[string]interface{}{
		"path":     path,
		"referrer": referrer,
	})
}

// TrackPaymentSucceeded tracks a successful payment
func (t *Tracker) TrackPaymentSucceeded(userID uuid.UUID, amount float64, plan string) {
	t.TrackEvent(domain.EventPaymentSucceeded, &userID, map[string]interface{}{
		"amount": amount,
		"plan":   plan,
	})
}

// TrackPaymentFailed tracks a failed payment
func (t *Tracker) TrackPaymentFailed(userID uuid.UUID, amount float64, reason string) {
	t.TrackEvent(domain.EventPaymentFailed, &userID, map[string]interface{}{
		"amount": amount,
		"reason": reason,
	})
}
