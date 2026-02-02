package middleware

import (
	"net/http"

	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/gin-gonic/gin"
	stderrors "errors"
)

// ErrorHandler middleware handles errors and returns consistent error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError converts an error to a consistent error response
func handleError(c *gin.Context, err error) {
	// Check for ValidationError first (has field-level details)
	var validationErr *errors.ValidationError
	if stderrors.As(err, &validationErr) {
		c.JSON(validationErr.HTTPStatus, validationErr.ToErrorResponse())
		return
	}

	// Check for RateLimitError
	var rateLimitErr *errors.RateLimitError
	if stderrors.As(err, &rateLimitErr) {
		retryAfter := int(rateLimitErr.RetryAfter.Seconds())
		c.Header("Retry-After", string(rune(retryAfter)))
		c.JSON(rateLimitErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:       rateLimitErr.Code,
				Message:    rateLimitErr.Message,
				RetryAfter: retryAfter,
				Platform:   rateLimitErr.Platform,
			},
		})
		return
	}

	// Check for PlatformAPIError
	var platformErr *errors.PlatformAPIError
	if stderrors.As(err, &platformErr) {
		c.JSON(platformErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:     platformErr.Code,
				Message:  platformErr.Message,
				Platform: platformErr.Platform,
			},
		})
		return
	}

	// Check for TokenError
	var tokenErr *errors.TokenError
	if stderrors.As(err, &tokenErr) {
		c.JSON(tokenErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:     tokenErr.Code,
				Message:  tokenErr.Message,
				Platform: tokenErr.Platform,
			},
		})
		return
	}

	// Check for generic AppError
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, appErr.ToErrorResponse())
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeInternal,
			Message: "An unexpected error occurred",
		},
	})
}

// RespondWithError is a helper function to respond with an error
// Use this in handlers instead of c.JSON for errors
func RespondWithError(c *gin.Context, err error) {
	// Check for ValidationError first (has field-level details)
	var validationErr *errors.ValidationError
	if stderrors.As(err, &validationErr) {
		c.JSON(validationErr.HTTPStatus, validationErr.ToErrorResponse())
		c.Abort()
		return
	}

	// Check for RateLimitError
	var rateLimitErr *errors.RateLimitError
	if stderrors.As(err, &rateLimitErr) {
		retryAfter := int(rateLimitErr.RetryAfter.Seconds())
		c.Header("Retry-After", string(rune(retryAfter)))
		c.JSON(rateLimitErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:       rateLimitErr.Code,
				Message:    rateLimitErr.Message,
				RetryAfter: retryAfter,
				Platform:   rateLimitErr.Platform,
			},
		})
		c.Abort()
		return
	}

	// Check for PlatformAPIError
	var platformErr *errors.PlatformAPIError
	if stderrors.As(err, &platformErr) {
		c.JSON(platformErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:     platformErr.Code,
				Message:  platformErr.Message,
				Platform: platformErr.Platform,
			},
		})
		c.Abort()
		return
	}

	// Check for TokenError
	var tokenErr *errors.TokenError
	if stderrors.As(err, &tokenErr) {
		c.JSON(tokenErr.HTTPStatus, &errors.ErrorResponse{
			Success: false,
			Error: &errors.ErrorDetail{
				Code:     tokenErr.Code,
				Message:  tokenErr.Message,
				Platform: tokenErr.Platform,
			},
		})
		c.Abort()
		return
	}

	// Check for generic AppError
	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, appErr.ToErrorResponse())
		c.Abort()
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeInternal,
			Message: "An unexpected error occurred",
		},
	})
	c.Abort()
}

// RespondWithValidationError is a helper for validation errors
func RespondWithValidationError(c *gin.Context, message string, fields ...errors.FieldError) {
	validationErr := errors.NewValidationError(message, fields...)
	c.JSON(http.StatusBadRequest, validationErr.ToErrorResponse())
	c.Abort()
}

// RespondWithUnauthorized is a helper for unauthorized errors
func RespondWithUnauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication required"
	}
	c.JSON(http.StatusUnauthorized, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeUnauthorized,
			Message: message,
		},
	})
	c.Abort()
}

// RespondWithForbidden is a helper for forbidden errors
func RespondWithForbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access denied"
	}
	c.JSON(http.StatusForbidden, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeForbidden,
			Message: message,
		},
	})
	c.Abort()
}

// RespondWithNotFound is a helper for not found errors
func RespondWithNotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeNotFound,
			Message: resource + " not found",
		},
	})
	c.Abort()
}

// RespondWithRateLimit is a helper for rate limit errors
func RespondWithRateLimit(c *gin.Context, retryAfterSeconds int) {
	c.Header("Retry-After", string(rune(retryAfterSeconds)))
	c.JSON(http.StatusTooManyRequests, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:       errors.ErrCodeRateLimit,
			Message:    "Too many requests. Please try again later.",
			RetryAfter: retryAfterSeconds,
		},
	})
	c.Abort()
}

// RespondWithPlatformError is a helper for platform API errors
func RespondWithPlatformError(c *gin.Context, platform, message string) {
	c.JSON(http.StatusBadGateway, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:     errors.ErrCodePlatformAPI,
			Message:  message,
			Platform: platform,
		},
	})
	c.Abort()
}

// RespondWithSubscriptionLimit is a helper for subscription limit errors
func RespondWithSubscriptionLimit(c *gin.Context, resource string, limit int) {
	c.JSON(http.StatusPaymentRequired, &errors.ErrorResponse{
		Success: false,
		Error: &errors.ErrorDetail{
			Code:    errors.ErrCodeSubscriptionLimit,
			Message: "Subscription limit reached for " + resource,
		},
	})
	c.Abort()
}
