package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================================
// Input Validation & Sanitization
// ============================================================================

// SQLInjectionPatterns contains regex patterns for SQL injection detection
var SQLInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union\s+(all\s+)?select)`),
	regexp.MustCompile(`(?i)(select\s+.+\s+from)`),
	regexp.MustCompile(`(?i)(insert\s+into)`),
	regexp.MustCompile(`(?i)(update\s+.+\s+set)`),
	regexp.MustCompile(`(?i)(delete\s+from)`),
	regexp.MustCompile(`(?i)(drop\s+(table|database))`),
	regexp.MustCompile(`(?i)(alter\s+table)`),
	regexp.MustCompile(`(?i)(exec(ute)?(\s|\+)+(s|x)p\w+)`),
	regexp.MustCompile(`(?i)(--)`),
	regexp.MustCompile(`(?i)(;.*--)`),
	regexp.MustCompile(`(?i)(/\*.*\*/)`),
	regexp.MustCompile(`(?i)(waitfor\s+delay)`),
	regexp.MustCompile(`(?i)(benchmark\s*\()`),
	regexp.MustCompile(`(?i)(sleep\s*\()`),
}

// XSSPatterns contains regex patterns for XSS detection
var XSSPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script[^>]*>`),
	regexp.MustCompile(`(?i)</script>`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
	regexp.MustCompile(`(?i)<iframe[^>]*>`),
	regexp.MustCompile(`(?i)<object[^>]*>`),
	regexp.MustCompile(`(?i)<embed[^>]*>`),
	regexp.MustCompile(`(?i)<svg[^>]*onload`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)vbscript:`),
	regexp.MustCompile(`(?i)data:\s*text/html`),
}

// InputValidator validates and sanitizes user input
type InputValidator struct {
	maxBodySize      int64
	maxQueryLen      int
	maxPathLen       int
	allowedMethods   map[string]bool
	blocklistPaths   []string
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxBodySize:    10 * 1024 * 1024, // 10MB
		maxQueryLen:    2048,
		maxPathLen:     1024,
		allowedMethods: map[string]bool{
			"GET": true, "POST": true, "PUT": true,
			"PATCH": true, "DELETE": true, "OPTIONS": true, "HEAD": true,
		},
		blocklistPaths: []string{
			"...", "..\\", "../", "/..", "\\..",
			"/etc/", "/proc/", "/sys/",
			".git", ".env", ".htaccess",
		},
	}
}

// ValidateRequest middleware validates incoming requests
func (v *InputValidator) ValidateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate HTTP method
		if !v.allowedMethods[c.Request.Method] {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
				"error": gin.H{
					"code":    "METHOD_NOT_ALLOWED",
					"message": "HTTP method not allowed",
				},
			})
			return
		}

		// Check path length
		if len(c.Request.URL.Path) > v.maxPathLen {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "PATH_TOO_LONG",
					"message": "Request path too long",
				},
			})
			return
		}

		// Check query string length
		if len(c.Request.URL.RawQuery) > v.maxQueryLen {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "QUERY_TOO_LONG",
					"message": "Query string too long",
				},
			})
			return
		}

		// Check for path traversal attempts
		path := c.Request.URL.Path
		for _, blocklist := range v.blocklistPaths {
			if strings.Contains(path, blocklist) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"code":    "INVALID_PATH",
						"message": "Invalid path detected",
					},
				})
				return
			}
		}

		// Check for SQL injection in query params
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if v.containsSQLInjection(value) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": gin.H{
							"code":    "MALICIOUS_INPUT",
							"message": fmt.Sprintf("Invalid input detected in parameter: %s", key),
						},
					})
					return
				}
			}
		}

		// Check for XSS in query params
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if v.containsXSS(value) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": gin.H{
							"code":    "MALICIOUS_INPUT",
							"message": fmt.Sprintf("Invalid input detected in parameter: %s", key),
						},
					})
					return
				}
			}
		}

		c.Next()
	}
}

// containsSQLInjection checks if input contains SQL injection patterns
func (v *InputValidator) containsSQLInjection(input string) bool {
	for _, pattern := range SQLInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// containsXSS checks if input contains XSS patterns
func (v *InputValidator) containsXSS(input string) bool {
	for _, pattern := range XSSPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeString sanitizes a string by escaping HTML and removing dangerous patterns
func SanitizeString(input string) string {
	// Escape HTML entities
	sanitized := html.EscapeString(input)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove control characters (except tab, newline, carriage return)
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ============================================================================
// CSRF Protection
// ============================================================================

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	TokenLength    int
	CookieName     string
	HeaderName     string
	CookiePath     string
	CookieDomain   string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
	MaxAge         int
	ExemptMethods  []string
	ExemptPaths    []string
}

// DefaultCSRFConfig returns default CSRF configuration
func DefaultCSRFConfig() *CSRFConfig {
	return &CSRFConfig{
		TokenLength:    32,
		CookieName:     "_csrf",
		HeaderName:     "X-CSRF-Token",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
		MaxAge:         3600, // 1 hour
		ExemptMethods:  []string{"GET", "HEAD", "OPTIONS"},
		ExemptPaths:    []string{"/api/v1/auth/login", "/api/v1/auth/register", "/health"},
	}
}

// CSRFMiddleware provides CSRF protection
type CSRFMiddleware struct {
	config *CSRFConfig
	tokens sync.Map // token -> expiry
}

// NewCSRFMiddleware creates a new CSRF middleware
func NewCSRFMiddleware(config *CSRFConfig) *CSRFMiddleware {
	if config == nil {
		config = DefaultCSRFConfig()
	}
	m := &CSRFMiddleware{config: config}

	// Start cleanup goroutine
	go m.cleanup()

	return m
}

// Handle returns the CSRF middleware handler
func (m *CSRFMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip exempt methods
		for _, method := range m.config.ExemptMethods {
			if c.Request.Method == method {
				// Still set token for GET requests
				m.ensureToken(c)
				c.Next()
				return
			}
		}

		// Skip exempt paths
		for _, path := range m.config.ExemptPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Validate CSRF token
		cookieToken, err := c.Cookie(m.config.CookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_TOKEN_MISSING",
					"message": "CSRF token not found",
				},
			})
			return
		}

		headerToken := c.GetHeader(m.config.HeaderName)
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_TOKEN_MISSING",
					"message": "CSRF token not found in header",
				},
			})
			return
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_TOKEN_INVALID",
					"message": "CSRF token mismatch",
				},
			})
			return
		}

		// Verify token is in our store (not expired)
		if _, ok := m.tokens.Load(cookieToken); !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "CSRF_TOKEN_EXPIRED",
					"message": "CSRF token expired",
				},
			})
			return
		}

		c.Next()
	}
}

// ensureToken ensures a CSRF token exists in the cookie
func (m *CSRFMiddleware) ensureToken(c *gin.Context) {
	// Check if token already exists
	if token, err := c.Cookie(m.config.CookieName); err == nil && token != "" {
		// Verify it's in our store
		if _, ok := m.tokens.Load(token); ok {
			return
		}
	}

	// Generate new token
	token := m.generateToken()
	expiry := time.Now().Add(time.Duration(m.config.MaxAge) * time.Second)
	m.tokens.Store(token, expiry)

	// Set cookie
	c.SetSameSite(m.config.CookieSameSite)
	c.SetCookie(
		m.config.CookieName,
		token,
		m.config.MaxAge,
		m.config.CookiePath,
		m.config.CookieDomain,
		m.config.CookieSecure,
		false, // Not HTTP-only so JavaScript can read it for header
	)
}

// generateToken generates a cryptographically secure random token
func (m *CSRFMiddleware) generateToken() string {
	b := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(b); err != nil {
		// Fallback to UUID if random fails
		return uuid.New().String()
	}
	return base64.URLEncoding.EncodeToString(b)
}

// cleanup periodically removes expired tokens
func (m *CSRFMiddleware) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.tokens.Range(func(key, value interface{}) bool {
			if expiry, ok := value.(time.Time); ok && now.After(expiry) {
				m.tokens.Delete(key)
			}
			return true
		})
	}
}

// ============================================================================
// Enhanced Rate Limiting (per-user and per-IP)
// ============================================================================

// DualRateLimiter applies different rate limits for users and IPs
type DualRateLimiter struct {
	userLimiter *RateLimitConfig
	ipLimiter   *RateLimitConfig
	limiters    map[string]*tokenBucketEnhanced
	mu          sync.RWMutex
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// tokenBucketEnhanced is an enhanced token bucket with minute-based limits
type tokenBucketEnhanced struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	lastAccess time.Time
	mu         sync.Mutex
}

// NewDualRateLimiter creates a rate limiter with separate user and IP limits
func NewDualRateLimiter(userReqPerMin, ipReqPerMin int) *DualRateLimiter {
	d := &DualRateLimiter{
		userLimiter: &RateLimitConfig{
			RequestsPerMinute: userReqPerMin, // 100 req/min for users
			BurstSize:         userReqPerMin / 2,
		},
		ipLimiter: &RateLimitConfig{
			RequestsPerMinute: ipReqPerMin, // 50 req/min for IPs
			BurstSize:         ipReqPerMin / 2,
		},
		limiters: make(map[string]*tokenBucketEnhanced),
	}

	go d.cleanup()
	return d
}

// Handle returns the rate limiting middleware handler
func (d *DualRateLimiter) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		var config *RateLimitConfig

		// Check if user is authenticated
		if userID, exists := c.Get(ContextKeyUserID); exists {
			key = "user:" + userID.(uuid.UUID).String()
			config = d.userLimiter
		} else {
			key = "ip:" + c.ClientIP()
			config = d.ipLimiter
		}

		limiter := d.getLimiter(key, config)

		if !limiter.allow() {
			retryAfter := 60 / config.RequestsPerMinute
			if retryAfter < 1 {
				retryAfter = 1
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":        "RATE_LIMITED",
					"message":     "Too many requests. Please slow down.",
					"retry_after": retryAfter,
				},
			})
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%.0f", limiter.tokens))

		c.Next()
	}
}

func (d *DualRateLimiter) getLimiter(key string, config *RateLimitConfig) *tokenBucketEnhanced {
	d.mu.RLock()
	limiter, exists := d.limiters[key]
	d.mu.RUnlock()

	if exists {
		return limiter
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if limiter, exists = d.limiters[key]; exists {
		return limiter
	}

	limiter = &tokenBucketEnhanced{
		tokens:     float64(config.BurstSize),
		maxTokens:  float64(config.BurstSize),
		refillRate: float64(config.RequestsPerMinute) / 60.0, // per second
		lastRefill: time.Now(),
		lastAccess: time.Now(),
	}
	d.limiters[key] = limiter
	return limiter
}

func (tb *tokenBucketEnhanced) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now
	tb.lastAccess = now

	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

func (d *DualRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		d.mu.Lock()
		threshold := time.Now().Add(-10 * time.Minute)
		for key, limiter := range d.limiters {
			limiter.mu.Lock()
			if limiter.lastAccess.Before(threshold) {
				delete(d.limiters, key)
			}
			limiter.mu.Unlock()
		}
		d.mu.Unlock()
	}
}

// ============================================================================
// Enhanced Security Headers
// ============================================================================

// SecurityHeadersConfig holds security header configuration
type SecurityHeadersConfig struct {
	ContentSecurityPolicy   string
	XContentTypeOptions     string
	XFrameOptions           string
	XXSSProtection          string
	StrictTransportSecurity string
	ReferrerPolicy          string
	PermissionsPolicy       string
	CacheControl            string
}

// DefaultSecurityHeadersConfig returns default security headers
func DefaultSecurityHeadersConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		ContentSecurityPolicy: strings.Join([]string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
			"style-src 'self' 'unsafe-inline'",
			"img-src 'self' data: https:",
			"font-src 'self' data:",
			"connect-src 'self' https:",
			"frame-ancestors 'self'",
			"form-action 'self'",
			"base-uri 'self'",
			"object-src 'none'",
			"upgrade-insecure-requests",
		}, "; "),
		XContentTypeOptions:     "nosniff",
		XFrameOptions:           "DENY",
		XXSSProtection:          "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), payment=()",
		CacheControl:            "no-store, no-cache, must-revalidate, proxy-revalidate",
	}
}

// EnhancedSecurityHeaders returns middleware with comprehensive security headers
func EnhancedSecurityHeaders(config *SecurityHeadersConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSecurityHeadersConfig()
	}

	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", config.ContentSecurityPolicy)
		c.Header("X-Content-Type-Options", config.XContentTypeOptions)
		c.Header("X-Frame-Options", config.XFrameOptions)
		c.Header("X-XSS-Protection", config.XXSSProtection)
		c.Header("Strict-Transport-Security", config.StrictTransportSecurity)
		c.Header("Referrer-Policy", config.ReferrerPolicy)
		c.Header("Permissions-Policy", config.PermissionsPolicy)
		c.Header("Cache-Control", config.CacheControl)
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Remove server identification
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
