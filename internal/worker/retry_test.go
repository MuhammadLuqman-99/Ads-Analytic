package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	appErrors "github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/rs/zerolog"
)

func TestNewRetryer(t *testing.T) {
	tests := []struct {
		name   string
		config *RetryConfig
		want   *RetryConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			want:   DefaultRetryConfig(),
		},
		{
			name: "zero multiplier uses default",
			config: &RetryConfig{
				MaxRetries:   5,
				BaseDelay:    100 * time.Millisecond,
				Multiplier:   0,
				JitterFactor: 0,
			},
			want: &RetryConfig{
				MaxRetries:   5,
				BaseDelay:    100 * time.Millisecond,
				Multiplier:   2.0,
				JitterFactor: 0,
			},
		},
		{
			name: "invalid jitter factor is corrected",
			config: &RetryConfig{
				MaxRetries:   3,
				BaseDelay:    100 * time.Millisecond,
				Multiplier:   2.0,
				JitterFactor: 1.5,
			},
			want: &RetryConfig{
				MaxRetries:   3,
				BaseDelay:    100 * time.Millisecond,
				Multiplier:   2.0,
				JitterFactor: 0.1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRetryer(tt.config, zerolog.Nop())
			if r.config.MaxRetries != tt.want.MaxRetries {
				t.Errorf("MaxRetries = %d, want %d", r.config.MaxRetries, tt.want.MaxRetries)
			}
			if r.config.Multiplier != tt.want.Multiplier {
				t.Errorf("Multiplier = %f, want %f", r.config.Multiplier, tt.want.Multiplier)
			}
		})
	}
}

func TestRetryer_Execute_Success(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	callCount := 0
	result := r.Execute(context.Background(), "test", func(ctx context.Context) error {
		callCount++
		return nil
	})

	if !result.Success {
		t.Error("Expected success")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
	if result.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", result.Attempts)
	}
	if result.LastError != nil {
		t.Errorf("Expected no error, got %v", result.LastError)
	}
}

func TestRetryer_Execute_RetryThenSuccess(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	callCount := 0
	result := r.Execute(context.Background(), "test", func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return appErrors.ErrInternal("temporary error")
		}
		return nil
	})

	if !result.Success {
		t.Errorf("Expected success after retries, got error: %v", result.LastError)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", result.Attempts)
	}
}

func TestRetryer_Execute_AllRetrysFail(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   2,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	expectedErr := errors.New("persistent error")
	callCount := 0
	result := r.Execute(context.Background(), "test", func(ctx context.Context) error {
		callCount++
		return appErrors.Wrap(expectedErr, appErrors.ErrCodeInternal, "wrapped error", 500)
	})

	if result.Success {
		t.Error("Expected failure")
	}
	if callCount != 3 { // Initial + 2 retries
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	if result.LastError == nil {
		t.Error("Expected error")
	}
}

func TestRetryer_Execute_ContextCancelled(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   5,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result := r.Execute(ctx, "test", func(ctx context.Context) error {
		callCount++
		return appErrors.ErrInternal("error")
	})

	if result.Success {
		t.Error("Expected failure due to context cancellation")
	}
	if !errors.Is(result.LastError, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", result.LastError)
	}
}

func TestRetryer_Execute_NonRetryableError(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   3,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	// 404 errors should not be retried
	callCount := 0
	result := r.Execute(context.Background(), "test", func(ctx context.Context) error {
		callCount++
		return appErrors.Wrap(errors.New("not found"), appErrors.ErrCodeNotFound, "not found", 404)
	})

	if result.Success {
		t.Error("Expected failure")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retry for 404), got %d", callCount)
	}
}

func TestRetryer_CalculateDelay(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   5,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0, // No jitter for predictable tests
	}
	r := NewRetryer(config, zerolog.Nop())

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},  // 100ms * 2^0 = 100ms
		{1, 200 * time.Millisecond},  // 100ms * 2^1 = 200ms
		{2, 400 * time.Millisecond},  // 100ms * 2^2 = 400ms
		{3, 800 * time.Millisecond},  // 100ms * 2^3 = 800ms
		{4, 1600 * time.Millisecond}, // 100ms * 2^4 = 1.6s
	}

	for _, tt := range tests {
		got := r.calculateDelay(tt.attempt)
		if got != tt.want {
			t.Errorf("calculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestRetryer_CalculateDelay_MaxCap(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   10,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	// After several attempts, should be capped at MaxDelay
	got := r.calculateDelay(10)
	if got != config.MaxDelay {
		t.Errorf("calculateDelay(10) = %v, want %v (MaxDelay)", got, config.MaxDelay)
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		status    int
		retryable bool
	}{
		{429, true},  // Too Many Requests
		{500, true},  // Internal Server Error
		{502, true},  // Bad Gateway
		{503, true},  // Service Unavailable
		{504, true},  // Gateway Timeout
		{400, false}, // Bad Request
		{401, false}, // Unauthorized
		{403, false}, // Forbidden
		{404, false}, // Not Found
		{409, false}, // Conflict
		{422, false}, // Unprocessable Entity
	}

	for _, tt := range tests {
		got := isRetryableHTTPStatus(tt.status)
		if got != tt.retryable {
			t.Errorf("isRetryableHTTPStatus(%d) = %v, want %v", tt.status, got, tt.retryable)
		}
	}
}

func TestRetryWithBackoff(t *testing.T) {
	callCount := 0
	err := RetryWithBackoff(context.Background(), 2, func(ctx context.Context) error {
		callCount++
		if callCount < 2 {
			return appErrors.ErrInternal("temp error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected success after retry, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestIsRateLimitError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "rate limit error",
			err:      appErrors.Wrap(errors.New("rate limited"), appErrors.ErrCodePlatformUnavailable, "rate limit", 429),
			expected: true,
		},
		{
			name:     "internal error",
			err:      appErrors.ErrInternal("internal error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimitError(tt.err)
			if got != tt.expected {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExecuteWithResult(t *testing.T) {
	config := &RetryConfig{
		MaxRetries:   2,
		BaseDelay:    10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		JitterFactor: 0,
	}
	r := NewRetryer(config, zerolog.Nop())

	callCount := 0
	result, retryResult := ExecuteWithResult(context.Background(), r, "test", func(ctx context.Context) (string, error) {
		callCount++
		if callCount < 2 {
			return "", appErrors.ErrInternal("temp error")
		}
		return "success", nil
	})

	if !retryResult.Success {
		t.Errorf("Expected success, got error: %v", retryResult.LastError)
	}
	if result != "success" {
		t.Errorf("Expected 'success', got '%s'", result)
	}
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("BaseDelay = %v, want 1s", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("MaxDelay = %v, want 30s", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", config.Multiplier)
	}
	if config.JitterFactor != 0.1 {
		t.Errorf("JitterFactor = %f, want 0.1", config.JitterFactor)
	}
}
