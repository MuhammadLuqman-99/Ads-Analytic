package middleware

import (
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/oauth"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// ContextKeyConnectedAccount stores the validated connected account
	ContextKeyConnectedAccount = "connected_account"
	// ContextKeyAccessToken stores the decrypted access token
	ContextKeyAccessToken = "platform_access_token"
)

// PlatformAuthMiddleware handles platform token validation
type PlatformAuthMiddleware struct {
	tokenManager *oauth.TokenManager
}

// NewPlatformAuthMiddleware creates a new platform auth middleware
func NewPlatformAuthMiddleware(tokenManager *oauth.TokenManager) *PlatformAuthMiddleware {
	return &PlatformAuthMiddleware{
		tokenManager: tokenManager,
	}
}

// RequirePlatformToken validates that the request has a valid platform token
// This middleware extracts the connected account ID from the URL parameter and validates the token
func (m *PlatformAuthMiddleware) RequirePlatformToken(accountIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		accountIDStr := c.Param(accountIDParam)
		if accountIDStr == "" {
			m.abortWithError(c, errors.ErrBadRequest("Account ID is required"))
			return
		}

		accountID, err := uuid.Parse(accountIDStr)
		if err != nil {
			m.abortWithError(c, errors.ErrBadRequest("Invalid account ID format"))
			return
		}

		// Validate and get the account with decrypted token
		account, err := m.tokenManager.ValidateToken(c.Request.Context(), accountID)
		if err != nil {
			m.abortWithError(c, err)
			return
		}

		// Verify the account belongs to the authenticated organization
		orgID, ok := GetOrgID(c)
		if !ok {
			m.abortWithError(c, errors.ErrUnauthorized("Organization not found in context"))
			return
		}

		if account.OrganizationID != orgID {
			m.abortWithError(c, errors.ErrForbidden("Account does not belong to this organization"))
			return
		}

		// Set account and token in context
		c.Set(ContextKeyConnectedAccount, account)
		c.Set(ContextKeyAccessToken, account.AccessToken)

		c.Next()
	}
}

// RequirePlatform validates that the request is for a specific platform
func (m *PlatformAuthMiddleware) RequirePlatform(platform entity.Platform) gin.HandlerFunc {
	return func(c *gin.Context) {
		account, ok := GetConnectedAccount(c)
		if !ok {
			m.abortWithError(c, errors.ErrInternal("Connected account not found in context"))
			return
		}

		if account.Platform != platform {
			m.abortWithError(c, errors.ErrBadRequest("This endpoint requires a "+string(platform)+" account"))
			return
		}

		c.Next()
	}
}

// RequireActiveToken validates that the token is active and not expired
func (m *PlatformAuthMiddleware) RequireActiveToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		account, ok := GetConnectedAccount(c)
		if !ok {
			m.abortWithError(c, errors.ErrInternal("Connected account not found in context"))
			return
		}

		if account.Status != entity.AccountStatusActive {
			m.abortWithError(c, errors.NewAppError(
				errors.ErrCodeOAuthFailed,
				"ACCOUNT_INACTIVE",
				"Connected account is not active. Current status: "+string(account.Status),
				401,
			))
			return
		}

		c.Next()
	}
}

// abortWithError aborts the request with an error response
func (m *PlatformAuthMiddleware) abortWithError(c *gin.Context, err error) {
	status := errors.GetHTTPStatus(err)

	var appErr *errors.AppError
	if errors.IsAppError(err) {
		ae, ok := err.(*errors.AppError)
		if ok {
			appErr = ae
		}
	}

	if appErr != nil {
		c.AbortWithStatusJSON(status, gin.H{
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}

	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}

// Helper functions

// GetConnectedAccount extracts the connected account from the gin context
func GetConnectedAccount(c *gin.Context) (*entity.ConnectedAccount, bool) {
	account, exists := c.Get(ContextKeyConnectedAccount)
	if !exists {
		return nil, false
	}
	return account.(*entity.ConnectedAccount), true
}

// GetPlatformAccessToken extracts the decrypted access token from the gin context
func GetPlatformAccessToken(c *gin.Context) (string, bool) {
	token, exists := c.Get(ContextKeyAccessToken)
	if !exists {
		return "", false
	}
	return token.(string), true
}
