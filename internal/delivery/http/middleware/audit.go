package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================================
// Audit Logging
// ============================================================================

// AuditAction represents the type of auditable action
type AuditAction string

const (
	// Authentication actions
	AuditActionLogin         AuditAction = "auth.login"
	AuditActionLoginFailed   AuditAction = "auth.login_failed"
	AuditActionLogout        AuditAction = "auth.logout"
	AuditActionRegister      AuditAction = "auth.register"
	AuditActionPasswordReset AuditAction = "auth.password_reset"
	AuditActionPasswordChange AuditAction = "auth.password_change"
	AuditActionTokenRefresh  AuditAction = "auth.token_refresh"
	AuditActionMFAEnable     AuditAction = "auth.mfa_enable"
	AuditActionMFADisable    AuditAction = "auth.mfa_disable"

	// Platform connection actions
	AuditActionPlatformConnect    AuditAction = "platform.connect"
	AuditActionPlatformDisconnect AuditAction = "platform.disconnect"
	AuditActionPlatformSync       AuditAction = "platform.sync"
	AuditActionPlatformSyncFailed AuditAction = "platform.sync_failed"
	AuditActionTokenRefreshed     AuditAction = "platform.token_refreshed"

	// Data actions
	AuditActionDataExport   AuditAction = "data.export"
	AuditActionDataDelete   AuditAction = "data.delete"
	AuditActionReportGenerate AuditAction = "data.report_generate"

	// Settings actions
	AuditActionSettingsUpdate AuditAction = "settings.update"
	AuditActionAPIKeyCreate   AuditAction = "settings.api_key_create"
	AuditActionAPIKeyDelete   AuditAction = "settings.api_key_delete"

	// User management
	AuditActionUserInvite AuditAction = "user.invite"
	AuditActionUserRemove AuditAction = "user.remove"
	AuditActionRoleChange AuditAction = "user.role_change"

	// Billing actions
	AuditActionSubscriptionCreate AuditAction = "billing.subscription_create"
	AuditActionSubscriptionCancel AuditAction = "billing.subscription_cancel"
	AuditActionPaymentSuccess     AuditAction = "billing.payment_success"
	AuditActionPaymentFailed      AuditAction = "billing.payment_failed"

	// Security events
	AuditActionSuspiciousActivity AuditAction = "security.suspicious_activity"
	AuditActionRateLimited        AuditAction = "security.rate_limited"
	AuditActionBlockedRequest     AuditAction = "security.blocked_request"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID             string                 `json:"id"`
	Timestamp      time.Time              `json:"timestamp"`
	Action         AuditAction            `json:"action"`
	Severity       AuditSeverity          `json:"severity"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	OrganizationID *uuid.UUID             `json:"organization_id,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	RequestID      string                 `json:"request_id,omitempty"`
	Resource       string                 `json:"resource,omitempty"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	Details        map[string]interface{} `json:"details,omitempty"`
	Success        bool                   `json:"success"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	Duration       int64                  `json:"duration_ms,omitempty"`
}

// AuditLogger interface for audit logging implementations
type AuditLogger interface {
	Log(ctx context.Context, event *AuditEvent) error
	Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
}

// AuditFilter for querying audit logs
type AuditFilter struct {
	UserID         *uuid.UUID
	OrganizationID *uuid.UUID
	Action         *AuditAction
	Severity       *AuditSeverity
	FromTime       *time.Time
	ToTime         *time.Time
	IPAddress      string
	Success        *bool
	Limit          int
	Offset         int
}

// InMemoryAuditLogger is a simple in-memory audit logger for development
type InMemoryAuditLogger struct {
	events []AuditEvent
	maxSize int
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger(maxSize int) *InMemoryAuditLogger {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &InMemoryAuditLogger{
		events:  make([]AuditEvent, 0, maxSize),
		maxSize: maxSize,
	}
}

// Log logs an audit event
func (l *InMemoryAuditLogger) Log(ctx context.Context, event *AuditEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// In production, this should write to a database
	l.events = append(l.events, *event)

	// Trim if over max size
	if len(l.events) > l.maxSize {
		l.events = l.events[len(l.events)-l.maxSize:]
	}

	return nil
}

// Query queries audit logs
func (l *InMemoryAuditLogger) Query(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
	var results []AuditEvent

	for _, event := range l.events {
		if l.matchesFilter(&event, &filter) {
			results = append(results, event)
		}
	}

	// Apply limit and offset
	if filter.Offset > 0 && filter.Offset < len(results) {
		results = results[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

func (l *InMemoryAuditLogger) matchesFilter(event *AuditEvent, filter *AuditFilter) bool {
	if filter.UserID != nil && (event.UserID == nil || *event.UserID != *filter.UserID) {
		return false
	}
	if filter.OrganizationID != nil && (event.OrganizationID == nil || *event.OrganizationID != *filter.OrganizationID) {
		return false
	}
	if filter.Action != nil && event.Action != *filter.Action {
		return false
	}
	if filter.Severity != nil && event.Severity != *filter.Severity {
		return false
	}
	if filter.FromTime != nil && event.Timestamp.Before(*filter.FromTime) {
		return false
	}
	if filter.ToTime != nil && event.Timestamp.After(*filter.ToTime) {
		return false
	}
	if filter.IPAddress != "" && event.IPAddress != filter.IPAddress {
		return false
	}
	if filter.Success != nil && event.Success != *filter.Success {
		return false
	}
	return true
}

// AuditMiddleware provides audit logging for requests
type AuditMiddleware struct {
	logger   AuditLogger
	actions  map[string]AuditAction // path+method -> action
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(logger AuditLogger) *AuditMiddleware {
	return &AuditMiddleware{
		logger: logger,
		actions: map[string]AuditAction{
			"POST:/api/v1/auth/login":       AuditActionLogin,
			"POST:/api/v1/auth/logout":      AuditActionLogout,
			"POST:/api/v1/auth/register":    AuditActionRegister,
			"POST:/api/v1/auth/reset-password": AuditActionPasswordReset,
			"PUT:/api/v1/auth/password":     AuditActionPasswordChange,
			"POST:/api/v1/auth/refresh":     AuditActionTokenRefresh,

			"POST:/api/v1/oauth/connect":    AuditActionPlatformConnect,
			"DELETE:/api/v1/accounts":       AuditActionPlatformDisconnect,
			"POST:/api/v1/sync":             AuditActionPlatformSync,

			"POST:/api/v1/export":           AuditActionDataExport,
			"POST:/api/v1/reports":          AuditActionReportGenerate,

			"PUT:/api/v1/settings":          AuditActionSettingsUpdate,
			"POST:/api/v1/api-keys":         AuditActionAPIKeyCreate,
			"DELETE:/api/v1/api-keys":       AuditActionAPIKeyDelete,

			"POST:/api/v1/users/invite":     AuditActionUserInvite,
			"DELETE:/api/v1/users":          AuditActionUserRemove,
		},
	}
}

// Handle returns the audit middleware handler
func (m *AuditMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get action from path+method
		key := c.Request.Method + ":" + c.Request.URL.Path
		action, shouldAudit := m.actions[key]

		// Check for prefix matches (for parameterized routes)
		if !shouldAudit {
			for pattern, act := range m.actions {
				if matchesPattern(key, pattern) {
					action = act
					shouldAudit = true
					break
				}
			}
		}

		// Process request
		c.Next()

		// Only audit if this is an auditable action
		if !shouldAudit {
			return
		}

		// Build audit event
		event := &AuditEvent{
			ID:        uuid.New().String(),
			Timestamp: time.Now().UTC(),
			Action:    action,
			Severity:  AuditSeverityInfo,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Success:   c.Writer.Status() < 400,
			Duration:  time.Since(start).Milliseconds(),
		}

		// Get request ID
		if requestID, exists := c.Get(ContextKeyRequestID); exists {
			event.RequestID = requestID.(string)
		}

		// Get user ID
		if userID, exists := c.Get(ContextKeyUserID); exists {
			uid := userID.(uuid.UUID)
			event.UserID = &uid
		}

		// Get organization ID
		if orgID, exists := c.Get(ContextKeyOrgID); exists {
			oid := orgID.(uuid.UUID)
			event.OrganizationID = &oid
		}

		// Set severity based on action and success
		if !event.Success {
			event.Severity = AuditSeverityWarning
			if action == AuditActionLoginFailed || action == AuditActionBlockedRequest {
				event.Severity = AuditSeverityWarning
			}
		}

		// Add details
		event.Details = map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
		}

		// Get any errors
		if len(c.Errors) > 0 {
			event.ErrorMessage = c.Errors.String()
			event.Severity = AuditSeverityError
		}

		// Log the event (async)
		go func() {
			_ = m.logger.Log(context.Background(), event)
		}()
	}
}

// LogEvent logs a custom audit event
func (m *AuditMiddleware) LogEvent(c *gin.Context, action AuditAction, details map[string]interface{}, success bool, errMsg string) {
	event := &AuditEvent{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    action,
		Severity:  AuditSeverityInfo,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   details,
		Success:   success,
	}

	if errMsg != "" {
		event.ErrorMessage = errMsg
		event.Severity = AuditSeverityError
	}

	if requestID, exists := c.Get(ContextKeyRequestID); exists {
		event.RequestID = requestID.(string)
	}

	if userID, exists := c.Get(ContextKeyUserID); exists {
		uid := userID.(uuid.UUID)
		event.UserID = &uid
	}

	if orgID, exists := c.Get(ContextKeyOrgID); exists {
		oid := orgID.(uuid.UUID)
		event.OrganizationID = &oid
	}

	go func() {
		_ = m.logger.Log(context.Background(), event)
	}()
}

// matchesPattern checks if a key matches a pattern (supports * for path params)
func matchesPattern(key, pattern string) bool {
	// Simple prefix matching for now
	// In production, use proper path matching
	if len(key) >= len(pattern) {
		return key[:len(pattern)] == pattern
	}
	return false
}

// AuditEventJSON marshals audit event to JSON for logging
func (e *AuditEvent) JSON() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf(`{"id":"%s","action":"%s","error":"failed to marshal"}`, e.ID, e.Action)
	}
	return string(data)
}
