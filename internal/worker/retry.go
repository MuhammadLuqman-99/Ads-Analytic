package worker

import (
	"context"
	stderrors "errors"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/rs/zerolog"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// BaseDelay is the initial delay between retries
	BaseDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the factor by which delay increases each retry (default 2)
	Multiplier float64
	// JitterFactor adds randomness to delays (0.0 to 1.0, default 0.1)
	JitterFactor float64
}

// DefaultRetryConfig returns the default retry configuration
// Suitable for Meta API which has 200 calls/hour rate limit
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   3,
		BaseDelay:    1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}
}

// RetryResult contains information about the retry execution
type RetryResult struct {
	Attempts     int
	TotalLatency time.Duration
	LastError    error
	Success      bool
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// Retryer handles retry logic with exponential backoff
type Retryer struct {
	config *RetryConfig
	logger zerolog.Logger
}

// NewRetryer creates a new Retryer with the given configuration
func NewRetryer(config *RetryConfig, logger zerolog.Logger) *Retryer {
	if config == nil {
		config = DefaultRetryConfig()
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}
	if config.JitterFactor < 0 || config.JitterFactor > 1 {
		config.JitterFactor = 0.1
	}
	return &Retryer{
		config: config,
		logger: logger.With().Str("component", "retryer").Logger(),
	}
}

// Execute runs the given function with retry logic
func (r *Retryer) Execute(ctx context.Context, operation string, fn RetryableFunc) *RetryResult {
	result := &RetryResult{
		Attempts: 0,
	}
	startTime := time.Now()

	var lastErr error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		result.Attempts = attempt + 1

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			result.LastError = ctx.Err()
			result.TotalLatency = time.Since(startTime)
			return result
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			result.Success = true
			result.TotalLatency = time.Since(startTime)
			if attempt > 0 {
				r.logger.Info().
					Str("operation", operation).
					Int("attempts", attempt+1).
					Dur("total_latency", result.TotalLatency).
					Msg("Operation succeeded after retry")
			}
			return result
		}

		lastErr = err
		result.LastError = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			r.logger.Warn().
				Err(err).
				Str("operation", operation).
				Int("attempt", attempt+1).
				Msg("Non-retryable error encountered")
			result.TotalLatency = time.Since(startTime)
			return result
		}

		// Don't sleep after the last attempt
		if attempt < r.config.MaxRetries {
			delay := r.calculateDelay(attempt)
			r.logger.Warn().
				Err(err).
				Str("operation", operation).
				Int("attempt", attempt+1).
				Int("max_retries", r.config.MaxRetries).
				Dur("next_delay", delay).
				Msg("Retrying after error")

			select {
			case <-ctx.Done():
				result.LastError = ctx.Err()
				result.TotalLatency = time.Since(startTime)
				return result
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	result.TotalLatency = time.Since(startTime)
	r.logger.Error().
		Err(lastErr).
		Str("operation", operation).
		Int("attempts", result.Attempts).
		Dur("total_latency", result.TotalLatency).
		Msg("All retry attempts exhausted")

	return result
}

// ExecuteWithResult runs a function that returns a value with retry logic
func ExecuteWithResult[T any](ctx context.Context, r *Retryer, operation string, fn func(ctx context.Context) (T, error)) (T, *RetryResult) {
	var result T
	var lastResult T

	retryResult := r.Execute(ctx, operation, func(ctx context.Context) error {
		var err error
		lastResult, err = fn(ctx)
		if err == nil {
			result = lastResult
		}
		return err
	})

	if retryResult.Success {
		return result, retryResult
	}
	return lastResult, retryResult
}

// calculateDelay calculates the delay for the given attempt using exponential backoff with jitter
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Calculate base exponential delay
	delay := float64(r.config.BaseDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	// Add jitter: ±JitterFactor% of the delay
	if r.config.JitterFactor > 0 {
		jitter := delay * r.config.JitterFactor * (2*rand.Float64() - 1)
		delay += jitter
	}

	// Ensure delay is non-negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// isRetryable determines if an error is retryable
func (r *Retryer) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors - not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Check for our custom AppError types
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		return isRetryableHTTPStatus(appErr.HTTPStatus)
	}

	// Default to retrying unknown errors
	return true
}

// isRetryableHTTPStatus determines if an HTTP status code indicates a retryable error
func isRetryableHTTPStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests, // 429 - Rate limited
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
		http.StatusInternalServerError: // 500
		return true
	default:
		return false
	}
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		return appErr.HTTPStatus == http.StatusTooManyRequests
	}
	return false
}

// GetRetryAfter extracts the retry-after duration from a rate limit error
// Returns 0 if not a rate limit error or duration cannot be determined
func GetRetryAfter(err error) time.Duration {
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		if appErr.HTTPStatus == http.StatusTooManyRequests {
			// Meta API typically requires waiting 10-60 seconds for rate limit recovery
			// Default to 1 minute if not specified
			return 1 * time.Minute
		}
	}
	return 0
}

// RetryWithBackoff is a convenience function for simple retry cases
func RetryWithBackoff(ctx context.Context, maxRetries int, fn RetryableFunc) error {
	config := &RetryConfig{
		MaxRetries:   maxRetries,
		BaseDelay:    100 * time.Millisecond, // Faster for tests, still reasonable for production
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}
	retryer := NewRetryer(config, zerolog.Nop())
	result := retryer.Execute(ctx, "operation", fn)
	if result.Success {
		return nil
	}
	return result.LastError
}

// RetryForever retries indefinitely until success or context cancellation
func RetryForever(ctx context.Context, baseDelay, maxDelay time.Duration, logger zerolog.Logger, fn RetryableFunc) error {
	attempt := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		if delay > maxDelay {
			delay = maxDelay
		}

		// Add jitter
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitter

		logger.Warn().
			Err(err).
			Int("attempt", attempt+1).
			Dur("next_delay", delay).
			Msg("Operation failed, retrying indefinitely")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			attempt++
		}
	}
}
