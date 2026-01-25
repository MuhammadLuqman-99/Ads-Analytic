package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/ratelimit"
)

// Client is a production-ready HTTP client with retry logic, circuit breaker, and rate limiting
type Client struct {
	httpClient     *http.Client
	rateLimiter    *ratelimit.Limiter
	circuitBreaker *CircuitBreaker
	config         *ClientConfig
	mu             sync.RWMutex
	requestLogger  RequestLogger
}

// ClientConfig holds the HTTP client configuration
type ClientConfig struct {
	Timeout          time.Duration
	MaxRetries       int
	RetryWaitMin     time.Duration
	RetryWaitMax     time.Duration
	RetryableStatus  []int
	RateLimitCalls   int
	RateLimitWindow  time.Duration
	CircuitThreshold int
	CircuitTimeout   time.Duration
	UserAgent        string
}

// DefaultConfig returns a default client configuration
func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
		RetryableStatus: []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		RateLimitCalls:   100,
		RateLimitWindow:  time.Minute,
		CircuitThreshold: 5,
		CircuitTimeout:   30 * time.Second,
		UserAgent:        "AdsAggregator/1.0",
	}
}

// RequestLogger is an interface for logging HTTP requests
type RequestLogger interface {
	LogRequest(method, url string, statusCode int, duration time.Duration, err error)
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(config *ClientConfig) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		rateLimiter:    ratelimit.NewLimiter(config.RateLimitCalls, config.RateLimitWindow),
		circuitBreaker: NewCircuitBreaker(config.CircuitThreshold, config.CircuitTimeout),
		config:         config,
	}
}

// WithRateLimiter sets a custom rate limiter
func (c *Client) WithRateLimiter(limiter *ratelimit.Limiter) *Client {
	c.rateLimiter = limiter
	return c
}

// WithLogger sets the request logger
func (c *Client) WithLogger(logger RequestLogger) *Client {
	c.requestLogger = logger
	return c
}

// Request represents an HTTP request
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        interface{}
	BodyReader  io.Reader
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// Do executes an HTTP request with retry logic and circuit breaker
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	// Check circuit breaker
	if !c.circuitBreaker.Allow() {
		return nil, errors.New(
			errors.ErrCodePlatformUnavailable,
			"Circuit breaker is open, service temporarily unavailable",
			http.StatusServiceUnavailable,
		)
	}

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.NewRateLimitError("client", c.config.RateLimitWindow)
	}

	var lastErr error
	var resp *Response

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			waitTime := c.calculateBackoff(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		resp, lastErr = c.doRequest(ctx, req)
		if lastErr == nil && !c.isRetryableStatus(resp.StatusCode) {
			c.circuitBreaker.Success()
			return resp, nil
		}

		// Check if error is retryable
		if lastErr != nil && !c.isRetryableError(lastErr) {
			c.circuitBreaker.Failure()
			return nil, lastErr
		}

		// Check if status is retryable
		if resp != nil && !c.isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		c.circuitBreaker.Failure()
	}

	return resp, lastErr
}

// doRequest executes a single HTTP request
func (c *Client) doRequest(ctx context.Context, req *Request) (*Response, error) {
	start := time.Now()

	// Build URL with query params
	reqURL, err := c.buildURL(req.URL, req.QueryParams)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to build URL", http.StatusInternalServerError)
	}

	// Prepare body
	var bodyReader io.Reader
	if req.BodyReader != nil {
		bodyReader = req.BodyReader
	} else if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to marshal request body", http.StatusInternalServerError)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, reqURL, bodyReader)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to create HTTP request", http.StatusInternalServerError)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", c.config.UserAgent)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		duration := time.Since(start)
		c.logRequest(req.Method, reqURL, 0, duration, err)
		return nil, errors.Wrap(err, errors.ErrCodePlatformTimeout, "HTTP request failed", http.StatusGatewayTimeout)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to read response body", http.StatusInternalServerError)
	}

	duration := time.Since(start)
	c.logRequest(req.Method, reqURL, httpResp.StatusCode, duration, nil)

	return &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       body,
		Duration:   duration,
	}, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string, headers map[string]string, queryParams map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:      http.MethodGet,
		URL:         url,
		Headers:     headers,
		QueryParams: queryParams,
	})
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, url string, headers map[string]string, body interface{}) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headers,
		Body:    body,
	})
}

// PostForm performs a POST request with form data
func (c *Client) PostForm(ctx context.Context, urlStr string, headers map[string]string, formData url.Values) (*Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	return c.Do(ctx, &Request{
		Method:     http.MethodPost,
		URL:        urlStr,
		Headers:    headers,
		BodyReader: bytes.NewReader([]byte(formData.Encode())),
	})
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, url string, headers map[string]string) (*Response, error) {
	return c.Do(ctx, &Request{
		Method:  http.MethodDelete,
		URL:     url,
		Headers: headers,
	})
}

// Helper methods

func (c *Client) buildURL(baseURL string, queryParams map[string]string) (string, error) {
	if len(queryParams) == 0 {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	backoff := float64(c.config.RetryWaitMin) * math.Pow(2, float64(attempt-1))
	if backoff > float64(c.config.RetryWaitMax) {
		backoff = float64(c.config.RetryWaitMax)
	}

	// Add jitter (±25%)
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
	return time.Duration(backoff + jitter)
}

func (c *Client) isRetryableStatus(statusCode int) bool {
	for _, code := range c.config.RetryableStatus {
		if statusCode == code {
			return true
		}
	}
	return false
}

func (c *Client) isRetryableError(err error) bool {
	// Network errors and timeouts are retryable
	return errors.IsRetryable(err)
}

func (c *Client) logRequest(method, url string, statusCode int, duration time.Duration, err error) {
	if c.requestLogger != nil {
		c.requestLogger.LogRequest(method, url, statusCode, duration, err)
	}
}

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	mu           sync.RWMutex
	failures     int
	threshold    int
	timeout      time.Duration
	state        CircuitState
	lastFailure  time.Time
	halfOpenChan chan struct{}
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:    threshold,
		timeout:      timeout,
		state:        CircuitClosed,
		halfOpenChan: make(chan struct{}, 1),
	}
}

// Allow checks if a request is allowed
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			// Transition to half-open
			select {
			case cb.halfOpenChan <- struct{}{}:
				cb.mu.RUnlock()
				cb.mu.Lock()
				cb.state = CircuitHalfOpen
				cb.mu.Unlock()
				cb.mu.RLock()
				return true
			default:
				return false
			}
		}
		return false
	case CircuitHalfOpen:
		return false
	}
	return false
}

// Success records a successful request
func (cb *CircuitBreaker) Success() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		select {
		case <-cb.halfOpenChan:
		default:
		}
	}
}

// Failure records a failed request
func (cb *CircuitBreaker) Failure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		select {
		case <-cb.halfOpenChan:
		default:
		}
	}
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}
