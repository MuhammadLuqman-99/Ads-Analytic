package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/httpclient"
	"github.com/ads-aggregator/ads-aggregator/pkg/ratelimit"
)

// BaseConnector provides common functionality for all platform connectors
type BaseConnector struct {
	platform    entity.Platform
	httpClient  *httpclient.Client
	rateLimiter *ratelimit.Limiter
	config      *ConnectorConfig
	mu          sync.RWMutex

	// Rate limit tracking
	rateLimitRemaining int
	rateLimitReset     time.Time
}

// ConnectorConfig holds configuration for a platform connector
type ConnectorConfig struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	APIVersion      string
	BaseURL         string
	RateLimitCalls  int
	RateLimitWindow time.Duration
	Timeout         time.Duration
	MaxRetries      int
}

// NewBaseConnector creates a new base connector
func NewBaseConnector(platform entity.Platform, config *ConnectorConfig) *BaseConnector {
	httpConfig := httpclient.DefaultConfig()
	httpConfig.Timeout = config.Timeout
	httpConfig.MaxRetries = config.MaxRetries
	httpConfig.RateLimitCalls = config.RateLimitCalls
	httpConfig.RateLimitWindow = config.RateLimitWindow

	return &BaseConnector{
		platform:           platform,
		httpClient:         httpclient.NewClient(httpConfig),
		rateLimiter:        ratelimit.NewLimiter(config.RateLimitCalls, config.RateLimitWindow),
		config:             config,
		rateLimitRemaining: config.RateLimitCalls,
	}
}

// Platform returns the platform type
func (b *BaseConnector) Platform() entity.Platform {
	return b.platform
}

// GetRateLimit returns the current rate limit status
func (b *BaseConnector) GetRateLimit() service.RateLimitStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return service.RateLimitStatus{
		Platform:  b.platform,
		Limit:     b.config.RateLimitCalls,
		Remaining: b.rateLimitRemaining,
		ResetAt:   b.rateLimitReset.Unix(),
		IsLimited: b.rateLimitRemaining <= 0 && time.Now().Before(b.rateLimitReset),
	}
}

// updateRateLimit updates the rate limit tracking from response headers
func (b *BaseConnector) updateRateLimit(remaining int, resetAt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.rateLimitRemaining = remaining
	b.rateLimitReset = resetAt
}

// DoRequest performs an HTTP request with rate limiting and error handling
func (b *BaseConnector) DoRequest(ctx context.Context, req *httpclient.Request) (*httpclient.Response, error) {
	// Wait for rate limiter
	if err := b.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.NewRateLimitError(b.platform.String(), b.config.RateLimitWindow)
	}

	resp, err := b.httpClient.Do(ctx, req)
	if err != nil {
		return nil, b.wrapError(err)
	}

	// Update rate limit from headers if available
	b.parseRateLimitHeaders(resp.Headers)

	// Check for errors
	if resp.StatusCode >= 400 {
		return resp, b.parseErrorResponse(resp)
	}

	return resp, nil
}

// DoGet performs a GET request
func (b *BaseConnector) DoGet(ctx context.Context, url string, headers map[string]string, params map[string]string) (*httpclient.Response, error) {
	return b.DoRequest(ctx, &httpclient.Request{
		Method:      http.MethodGet,
		URL:         url,
		Headers:     headers,
		QueryParams: params,
	})
}

// DoPost performs a POST request
func (b *BaseConnector) DoPost(ctx context.Context, url string, headers map[string]string, body interface{}) (*httpclient.Response, error) {
	return b.DoRequest(ctx, &httpclient.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// parseRateLimitHeaders parses rate limit information from response headers
func (b *BaseConnector) parseRateLimitHeaders(headers http.Header) {
	// This should be overridden by specific platform connectors
	// Common headers include X-RateLimit-Remaining, X-RateLimit-Reset, etc.
}

// parseErrorResponse parses an error response from the platform API
func (b *BaseConnector) parseErrorResponse(resp *httpclient.Response) error {
	// Try to parse as JSON error
	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	if err := json.Unmarshal(resp.Body, &errResp); err == nil && errResp.Error.Message != "" {
		platformErr := errors.NewPlatformAPIError(
			b.platform.String(),
			resp.StatusCode,
			errResp.Error.Code,
			errResp.Error.Message,
		)
		platformErr.WithRawResponse(resp.Body)
		return platformErr
	}

	// Handle rate limit errors
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := b.parseRetryAfter(resp.Headers)
		return errors.NewRateLimitError(b.platform.String(), retryAfter)
	}

	// Generic error
	return errors.NewPlatformAPIError(
		b.platform.String(),
		resp.StatusCode,
		"UNKNOWN",
		fmt.Sprintf("API request failed with status %d", resp.StatusCode),
	).WithRawResponse(resp.Body)
}

// parseRetryAfter parses the Retry-After header
func (b *BaseConnector) parseRetryAfter(headers http.Header) time.Duration {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return 60 * time.Second // Default to 60 seconds
	}

	// Try parsing as seconds
	var seconds int
	if _, err := fmt.Sscanf(retryAfter, "%d", &seconds); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as date
	if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(t)
	}

	return 60 * time.Second
}

// wrapError wraps an error with platform context
func (b *BaseConnector) wrapError(err error) error {
	if errors.IsAppError(err) {
		return err
	}
	return errors.Wrap(err, errors.ErrCodePlatformAPI, fmt.Sprintf("%s API error", b.platform), http.StatusBadGateway)
}

// ParseJSON parses JSON response body
func (b *BaseConnector) ParseJSON(body []byte, v interface{}) error {
	if err := json.Unmarshal(body, v); err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "Failed to parse response", http.StatusInternalServerError)
	}
	return nil
}

// BuildAuthHeader builds the authorization header with bearer token
func (b *BaseConnector) BuildAuthHeader(accessToken string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + accessToken,
	}
}

// HealthCheck performs a basic health check
func (b *BaseConnector) HealthCheck(ctx context.Context) error {
	// Override in specific connectors
	return nil
}

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(1<<uint(i-1)) * time.Second
			if waitTime > 30*time.Second {
				waitTime = 30 * time.Second
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !errors.IsRetryable(lastErr) {
			return lastErr
		}

		// Check for rate limit and respect retry-after
		if errors.IsRateLimit(lastErr) {
			retryAfter := errors.GetRetryAfter(lastErr)
			if retryAfter > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(retryAfter):
				}
			}
		}
	}
	return lastErr
}

// PaginatedRequest represents a paginated API request
type PaginatedRequest struct {
	Cursor  string
	Limit   int
	HasMore bool
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data    json.RawMessage
	Paging  PagingInfo
	HasNext bool
}

// PagingInfo holds pagination information
type PagingInfo struct {
	Cursors struct {
		Before string `json:"before"`
		After  string `json:"after"`
	} `json:"cursors"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

// FetchAllPages fetches all pages of a paginated endpoint
func FetchAllPages[T any](ctx context.Context, fetcher func(cursor string) ([]T, string, error)) ([]T, error) {
	var allItems []T
	cursor := ""

	for {
		items, nextCursor, err := fetcher(cursor)
		if err != nil {
			return allItems, err
		}

		allItems = append(allItems, items...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor

		// Respect context cancellation
		select {
		case <-ctx.Done():
			return allItems, ctx.Err()
		default:
		}
	}

	return allItems, nil
}

// ConnectorRegistry manages platform connectors
type ConnectorRegistry struct {
	mu         sync.RWMutex
	connectors map[entity.Platform]service.PlatformConnector
}

// NewConnectorRegistry creates a new connector registry
func NewConnectorRegistry() *ConnectorRegistry {
	return &ConnectorRegistry{
		connectors: make(map[entity.Platform]service.PlatformConnector),
	}
}

// Register registers a connector for a platform
func (r *ConnectorRegistry) Register(platform entity.Platform, connector service.PlatformConnector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors[platform] = connector
}

// Get retrieves a connector for a platform
func (r *ConnectorRegistry) Get(platform entity.Platform) (service.PlatformConnector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	connector, ok := r.connectors[platform]
	return connector, ok
}

// List returns all registered connectors
func (r *ConnectorRegistry) List() []service.PlatformConnector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connectors := make([]service.PlatformConnector, 0, len(r.connectors))
	for _, c := range r.connectors {
		connectors = append(connectors, c)
	}
	return connectors
}
