package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Error codes for the application
const (
	// General errors
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeBadRequest   = "BAD_REQUEST"

	// Platform-specific errors
	ErrCodePlatformAPI         = "PLATFORM_API_ERROR"
	ErrCodeRateLimit           = "RATE_LIMIT_EXCEEDED"
	ErrCodeTokenExpired        = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid        = "TOKEN_INVALID"
	ErrCodeOAuthFailed         = "OAUTH_FAILED"
	ErrCodePlatformTimeout     = "PLATFORM_TIMEOUT"
	ErrCodePlatformUnavailable = "PLATFORM_UNAVAILABLE"

	// Sync errors
	ErrCodeSyncFailed   = "SYNC_FAILED"
	ErrCodePartialSync  = "PARTIAL_SYNC"
	ErrCodeSyncConflict = "SYNC_CONFLICT"
)

// AppError represents an application error with additional context
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	HTTPStatus int               `json:"-"`
	Err        error             `json:"-"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithError adds the underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// WithDetails adds additional details
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key, value string) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// ToJSON converts the error to JSON
func (e *AppError) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// New creates a new AppError
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now().UTC(),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
		Timestamp:  time.Now().UTC(),
	}
}

// Common error constructors

// ErrInternal creates an internal server error
func ErrInternal(message string) *AppError {
	return New(ErrCodeInternal, message, http.StatusInternalServerError)
}

// ErrValidation creates a validation error
func ErrValidation(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// ErrUnauthorized creates an unauthorized error
func ErrUnauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

// ErrForbidden creates a forbidden error
func ErrForbidden(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

// ErrConflict creates a conflict error
func ErrConflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

// ErrBadRequest creates a bad request error
func ErrBadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	*AppError
	RetryAfter  time.Duration
	RetryCount  int
	MaxRetries  int
	ShouldRetry bool
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err *AppError, retryAfter time.Duration, maxRetries int) *RetryableError {
	return &RetryableError{
		AppError:    err,
		RetryAfter:  retryAfter,
		RetryCount:  0,
		MaxRetries:  maxRetries,
		ShouldRetry: true,
	}
}

// CanRetry checks if the error can be retried
func (e *RetryableError) CanRetry() bool {
	return e.ShouldRetry && e.RetryCount < e.MaxRetries
}

// IncrementRetry increments the retry count
func (e *RetryableError) IncrementRetry() {
	e.RetryCount++
}

// RateLimitError represents a rate limit error from a platform API
type RateLimitError struct {
	*AppError
	Platform   string
	RetryAfter time.Duration
	Limit      int
	Remaining  int
	ResetAt    time.Time
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(platform string, retryAfter time.Duration) *RateLimitError {
	return &RateLimitError{
		AppError: New(
			ErrCodeRateLimit,
			fmt.Sprintf("Rate limit exceeded for %s", platform),
			http.StatusTooManyRequests,
		),
		Platform:   platform,
		RetryAfter: retryAfter,
		ResetAt:    time.Now().Add(retryAfter),
	}
}

// WithLimitInfo adds rate limit info
func (e *RateLimitError) WithLimitInfo(limit, remaining int, resetAt time.Time) *RateLimitError {
	e.Limit = limit
	e.Remaining = remaining
	e.ResetAt = resetAt
	return e
}

// PlatformAPIError represents an error from a platform API
type PlatformAPIError struct {
	*AppError
	Platform     string
	StatusCode   int
	PlatformCode string
	PlatformMsg  string
	RequestID    string
	RawResponse  []byte
}

// NewPlatformAPIError creates a new platform API error
func NewPlatformAPIError(platform string, statusCode int, platformCode, platformMsg string) *PlatformAPIError {
	return &PlatformAPIError{
		AppError: New(
			ErrCodePlatformAPI,
			fmt.Sprintf("%s API error: %s", platform, platformMsg),
			http.StatusBadGateway,
		),
		Platform:     platform,
		StatusCode:   statusCode,
		PlatformCode: platformCode,
		PlatformMsg:  platformMsg,
	}
}

// WithRequestID adds the request ID
func (e *PlatformAPIError) WithRequestID(requestID string) *PlatformAPIError {
	e.RequestID = requestID
	e.AppError.WithMetadata("request_id", requestID)
	return e
}

// WithRawResponse adds the raw response body
func (e *PlatformAPIError) WithRawResponse(body []byte) *PlatformAPIError {
	e.RawResponse = body
	return e
}

// IsRetryable checks if the platform error is retryable
func (e *PlatformAPIError) IsRetryable() bool {
	// Server errors and rate limits are retryable
	return e.StatusCode >= 500 || e.StatusCode == 429
}

// TokenError represents a token-related error
type TokenError struct {
	*AppError
	Platform  string
	TokenType string // "access" or "refresh"
	ExpiredAt time.Time
}

// NewTokenExpiredError creates a new token expired error
func NewTokenExpiredError(platform, tokenType string, expiredAt time.Time) *TokenError {
	return &TokenError{
		AppError: New(
			ErrCodeTokenExpired,
			fmt.Sprintf("%s %s token has expired", platform, tokenType),
			http.StatusUnauthorized,
		),
		Platform:  platform,
		TokenType: tokenType,
		ExpiredAt: expiredAt,
	}
}

// NewTokenInvalidError creates a new token invalid error
func NewTokenInvalidError(platform, tokenType string) *TokenError {
	return &TokenError{
		AppError: New(
			ErrCodeTokenInvalid,
			fmt.Sprintf("%s %s token is invalid", platform, tokenType),
			http.StatusUnauthorized,
		),
		Platform:  platform,
		TokenType: tokenType,
	}
}

// OAuthError represents an OAuth flow error
type OAuthError struct {
	*AppError
	Platform  string
	OAuthCode string
	OAuthDesc string
	State     string
}

// NewOAuthError creates a new OAuth error
func NewOAuthError(platform, oauthCode, oauthDesc string) *OAuthError {
	return &OAuthError{
		AppError: New(
			ErrCodeOAuthFailed,
			fmt.Sprintf("OAuth failed for %s: %s", platform, oauthDesc),
			http.StatusBadRequest,
		),
		Platform:  platform,
		OAuthCode: oauthCode,
		OAuthDesc: oauthDesc,
	}
}

// SyncError represents a synchronization error
type SyncError struct {
	*AppError
	Platform     string
	AccountID    string
	FailedItems  int
	SuccessItems int
	Errors       []error
}

// NewSyncError creates a new sync error
func NewSyncError(platform, accountID string) *SyncError {
	return &SyncError{
		AppError: New(
			ErrCodeSyncFailed,
			fmt.Sprintf("Sync failed for %s account %s", platform, accountID),
			http.StatusInternalServerError,
		),
		Platform:  platform,
		AccountID: accountID,
		Errors:    make([]error, 0),
	}
}

// AddError adds an error to the sync error
func (e *SyncError) AddError(err error) {
	e.Errors = append(e.Errors, err)
	e.FailedItems++
}

// IsPartial checks if the sync was partially successful
func (e *SyncError) IsPartial() bool {
	return e.SuccessItems > 0 && e.FailedItems > 0
}

// Error type checking helpers

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	var retryErr *RetryableError
	if errors.As(err, &retryErr) {
		return retryErr.CanRetry()
	}

	var platformErr *PlatformAPIError
	if errors.As(err, &platformErr) {
		return platformErr.IsRetryable()
	}

	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

// IsRateLimit checks if an error is a rate limit error
func IsRateLimit(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

// IsTokenExpired checks if an error is a token expired error
func IsTokenExpired(err error) bool {
	var tokenErr *TokenError
	if errors.As(err, &tokenErr) {
		return tokenErr.Code == ErrCodeTokenExpired
	}
	return false
}

// GetRetryAfter gets the retry-after duration from an error
func GetRetryAfter(err error) time.Duration {
	var retryErr *RetryableError
	if errors.As(err, &retryErr) {
		return retryErr.RetryAfter
	}

	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return rateLimitErr.RetryAfter
	}

	return 0
}

// GetHTTPStatus gets the HTTP status code from an error
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
