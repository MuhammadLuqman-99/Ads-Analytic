package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RateLimitMiddleware handles request rate limiting
type RateLimitMiddleware struct {
	requestsPerSecond int
	burst             int
	limiters          map[string]*tokenBucket
	mu                sync.RWMutex
	cleanupInterval   time.Duration
}

// tokenBucket implements a token bucket rate limiter
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(requestsPerSecond, burst int) *RateLimitMiddleware {
	m := &RateLimitMiddleware{
		requestsPerSecond: requestsPerSecond,
		burst:             burst,
		limiters:          make(map[string]*tokenBucket),
		cleanupInterval:   5 * time.Minute,
	}

	// Start cleanup goroutine
	go m.cleanup()

	return m
}

// Handle returns the rate limiting middleware handler
func (m *RateLimitMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (prefer user ID, fallback to IP)
		key := m.getClientKey(c)

		// Get or create limiter for this client
		limiter := m.getLimiter(key)

		// Check if request is allowed
		if !limiter.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests, please try again later",
				},
				"retry_after": 1,
			})
			return
		}

		c.Next()
	}
}

// getClientKey returns a unique identifier for the client
func (m *RateLimitMiddleware) getClientKey(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get(ContextKeyUserID); exists {
		return "user:" + userID.(uuid.UUID).String()
	}

	// Fallback to IP address
	ip := c.ClientIP()
	return "ip:" + ip
}

// getLimiter gets or creates a rate limiter for the given key
func (m *RateLimitMiddleware) getLimiter(key string) *tokenBucket {
	m.mu.RLock()
	limiter, exists := m.limiters[key]
	m.mu.RUnlock()

	if exists {
		return limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = m.limiters[key]; exists {
		return limiter
	}

	limiter = &tokenBucket{
		tokens:     float64(m.burst),
		maxTokens:  float64(m.burst),
		refillRate: float64(m.requestsPerSecond),
		lastRefill: time.Now(),
	}
	m.limiters[key] = limiter

	return limiter
}

// allow checks if a request is allowed
func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	// Check if we have enough tokens
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

// cleanup periodically removes old limiters
func (m *RateLimitMiddleware) cleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		threshold := time.Now().Add(-m.cleanupInterval)
		for key, limiter := range m.limiters {
			limiter.mu.Lock()
			if limiter.lastRefill.Before(threshold) {
				delete(m.limiters, key)
			}
			limiter.mu.Unlock()
		}
		m.mu.Unlock()
	}
}

// RequestLogger returns a middleware that logs HTTP requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// Get request ID
		requestID, _ := c.Get(ContextKeyRequestID)

		// Log format
		logEntry := map[string]interface{}{
			"status":     status,
			"method":     method,
			"path":       path,
			"query":      query,
			"ip":         clientIP,
			"latency_ms": latency.Milliseconds(),
			"user_agent": c.Request.UserAgent(),
		}

		if requestID != nil {
			logEntry["request_id"] = requestID
		}

		// Get user ID if authenticated
		if userID, exists := c.Get(ContextKeyUserID); exists {
			logEntry["user_id"] = userID.(uuid.UUID).String()
		}

		// Get any errors
		if len(c.Errors) > 0 {
			logEntry["errors"] = c.Errors.String()
		}

		// TODO: Use structured logging (zerolog)
		// For now, we're just setting up the middleware
	}
}

// ContextKeyRequestID is the context key for request ID
const ContextKeyRequestID = "request_id"

// RequestID returns a middleware that adds a request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for existing request ID in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in context and response header
		c.Set(ContextKeyRequestID, requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// Timeout returns a middleware that limits request processing time
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a done channel
		done := make(chan struct{})

		// Run handler in goroutine
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Request completed normally
		case <-time.After(timeout):
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": gin.H{
					"code":    "TIMEOUT",
					"message": "Request timeout",
				},
			})
		}
	}
}

// Recovery returns a middleware that recovers from panics
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred",
					"details": err.Error(),
				},
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "An unexpected error occurred",
			},
		})
	})
}

// SecureHeaders returns a middleware that adds security headers
func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
