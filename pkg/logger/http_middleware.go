package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GinMiddleware returns a Gin middleware for structured logging
func GinMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// Add request ID to context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Get request body for logging (only for non-GET requests)
		var requestBody string
		if c.Request.Method != "GET" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Truncate if too long
			if len(requestBody) > 1000 {
				requestBody = requestBody[:1000] + "...[truncated]"
			}
		}

		// Create response writer wrapper to capture status and size
		rw := &responseWriter{ResponseWriter: c.Writer, statusCode: 200}
		c.Writer = rw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user info from context if available
		userID, _ := c.Get("user_id")
		orgID, _ := c.Get("org_id")

		// Build log event
		event := logger.Info()

		// Add fields
		event.Str("request_id", requestID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", rw.statusCode).
			Int("size", rw.size).
			Dur("duration", duration).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent())

		// Add optional fields
		if userID != nil {
			event.Str("user_id", userID.(string))
		}
		if orgID != nil {
			event.Str("org_id", orgID.(string))
		}

		// Log errors if any
		if len(c.Errors) > 0 {
			event.Str("errors", c.Errors.String())
		}

		// Determine log level based on status code
		if rw.statusCode >= 500 {
			logger.Error().
				Str("request_id", requestID).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", rw.statusCode).
				Dur("duration", duration).
				Str("request_body", requestBody).
				Msg("Server error")
		} else if rw.statusCode >= 400 {
			logger.Warn().
				Str("request_id", requestID).
				Str("method", c.Request.Method).
				Str("path", c.Request.URL.Path).
				Int("status", rw.statusCode).
				Dur("duration", duration).
				Msg("Client error")
		} else {
			event.Msg("Request completed")
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response details
type responseWriter struct {
	gin.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// HTTPRequestLog represents a structured HTTP request log
type HTTPRequestLog struct {
	Timestamp   time.Time     `json:"timestamp"`
	RequestID   string        `json:"request_id"`
	Method      string        `json:"method"`
	Path        string        `json:"path"`
	Query       string        `json:"query,omitempty"`
	Status      int           `json:"status"`
	Size        int           `json:"size"`
	Duration    time.Duration `json:"duration"`
	DurationMs  float64       `json:"duration_ms"`
	ClientIP    string        `json:"client_ip"`
	UserAgent   string        `json:"user_agent,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
	OrgID       string        `json:"org_id,omitempty"`
	Error       string        `json:"error,omitempty"`
	RequestBody string        `json:"request_body,omitempty"`
}
