package errortracker

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// Config holds Sentry configuration
type Config struct {
	DSN              string
	Environment      string
	Release          string
	SampleRate       float64
	TracesSampleRate float64
	Debug            bool
	ServerName       string
}

// ErrorTracker wraps Sentry functionality
type ErrorTracker struct {
	config Config
}

var defaultTracker *ErrorTracker

// Init initializes Sentry
func Init(cfg Config) (*ErrorTracker, error) {
	if cfg.DSN == "" {
		// Sentry is optional, return nil tracker if DSN not set
		return nil, nil
	}

	if cfg.SampleRate == 0 {
		cfg.SampleRate = 1.0
	}
	if cfg.TracesSampleRate == 0 {
		cfg.TracesSampleRate = 0.1
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		SampleRate:       cfg.SampleRate,
		TracesSampleRate: cfg.TracesSampleRate,
		Debug:            cfg.Debug,
		ServerName:       cfg.ServerName,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Customize event before sending
			return event
		},
		AttachStacktrace: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	defaultTracker = &ErrorTracker{config: cfg}
	return defaultTracker, nil
}

// Default returns the default error tracker
func Default() *ErrorTracker {
	return defaultTracker
}

// Close flushes pending events before shutdown
func Close() {
	sentry.Flush(2 * time.Second)
}

// GinMiddleware returns Sentry middleware for Gin
func GinMiddleware() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
	})
}

// GinRecovery returns a Gin recovery middleware that reports panics to Sentry
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				// Get hub from context
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetLevel(sentry.LevelFatal)
						scope.SetTag("panic", "true")
						scope.SetExtra("stack_trace", stackTrace)
						scope.SetRequest(c.Request)

						// Add context values
						if requestID, exists := c.Get("request_id"); exists {
							scope.SetTag("request_id", requestID.(string))
						}
						if userID, exists := c.Get("user_id"); exists {
							scope.SetUser(sentry.User{ID: userID.(string)})
						}
						if orgID, exists := c.Get("org_id"); exists {
							scope.SetTag("org_id", orgID.(string))
						}

						hub.CaptureMessage(fmt.Sprintf("Panic recovered: %v", err))
					})
				} else {
					// Fallback if no hub in context
					sentry.CaptureMessage(fmt.Sprintf("Panic recovered: %v", err))
				}

				// Return 500 error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}

// CaptureError sends an error to Sentry
func CaptureError(ctx context.Context, err error) {
	if defaultTracker == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.CaptureException(err)
}

// CaptureMessage sends a message to Sentry
func CaptureMessage(ctx context.Context, message string) {
	if defaultTracker == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.CaptureMessage(message)
}

// CaptureErrorWithContext sends an error with additional context
func CaptureErrorWithContext(ctx context.Context, err error, extra map[string]interface{}) {
	if defaultTracker == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.WithScope(func(scope *sentry.Scope) {
		for key, value := range extra {
			scope.SetExtra(key, value)
		}
		hub.CaptureException(err)
	})
}

// SetUser sets user information for the current scope
func SetUser(ctx context.Context, userID, email, orgID string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.Scope().SetUser(sentry.User{
		ID:    userID,
		Email: email,
	})
	hub.Scope().SetTag("org_id", orgID)
}

// SetTag sets a tag for the current scope
func SetTag(ctx context.Context, key, value string) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.Scope().SetTag(key, value)
}

// SetExtra sets extra data for the current scope
func SetExtra(ctx context.Context, key string, value interface{}) {
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.Scope().SetExtra(key, value)
}

// StartSpan starts a new Sentry span
func StartSpan(ctx context.Context, operation string) (*sentry.Span, context.Context) {
	span := sentry.StartSpan(ctx, operation)
	return span, span.Context()
}

// ============================================================================
// Error severity levels for alerting
// ============================================================================

// ErrorSeverity represents error severity levels
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical" // Requires immediate attention
	SeverityError    ErrorSeverity = "error"    // Significant error
	SeverityWarning  ErrorSeverity = "warning"  // Potential issue
	SeverityInfo     ErrorSeverity = "info"     // Informational
)

// CaptureWithSeverity sends an error with a specific severity level
func CaptureWithSeverity(ctx context.Context, err error, severity ErrorSeverity) {
	if defaultTracker == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	var level sentry.Level
	switch severity {
	case SeverityCritical:
		level = sentry.LevelFatal
	case SeverityError:
		level = sentry.LevelError
	case SeverityWarning:
		level = sentry.LevelWarning
	case SeverityInfo:
		level = sentry.LevelInfo
	default:
		level = sentry.LevelError
	}

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		scope.SetTag("severity", string(severity))
		hub.CaptureException(err)
	})
}

// ============================================================================
// Error categories for grouping
// ============================================================================

// ErrorCategory represents error categories
type ErrorCategory string

const (
	CategoryDatabase   ErrorCategory = "database"
	CategoryAPI        ErrorCategory = "api"
	CategoryPlatform   ErrorCategory = "platform"
	CategoryBilling    ErrorCategory = "billing"
	CategoryAuth       ErrorCategory = "auth"
	CategorySync       ErrorCategory = "sync"
	CategoryValidation ErrorCategory = "validation"
)

// CaptureWithCategory sends an error with a specific category
func CaptureWithCategory(ctx context.Context, err error, category ErrorCategory, extra map[string]interface{}) {
	if defaultTracker == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("category", string(category))
		for key, value := range extra {
			scope.SetExtra(key, value)
		}
		hub.CaptureException(err)
	})
}

// ============================================================================
// Alert trigger helper
// ============================================================================

// AlertConfig defines when to trigger alerts
type AlertConfig struct {
	CriticalKeywords []string
	AlertWebhookURL  string
}

// ShouldAlert determines if an error should trigger an alert
func ShouldAlert(err error, severity ErrorSeverity) bool {
	if severity == SeverityCritical {
		return true
	}

	// Check for critical keywords
	errStr := err.Error()
	criticalKeywords := []string{
		"payment failed",
		"database connection",
		"redis connection",
		"rate limit exceeded",
		"authentication failed",
		"token expired",
		"out of memory",
	}

	for _, keyword := range criticalKeywords {
		if contains(errStr, keyword) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
