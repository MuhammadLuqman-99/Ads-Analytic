package handler

import (
	"net/http"
	"net/url"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/auth"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth-related HTTP requests
type OAuthHandler struct {
	authService *auth.Service
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(authService *auth.Service) *OAuthHandler {
	return &OAuthHandler{
		authService: authService,
	}
}

// InitiateMetaOAuthRequest is the request for initiating Meta OAuth
type InitiateMetaOAuthRequest struct {
	RedirectURL string `json:"redirect_url" form:"redirect_url"` // Where to redirect after OAuth
}

// InitiateMetaOAuth initiates the OAuth flow for Meta (Facebook) Ads
// @Summary Initiate Meta OAuth flow
// @Description Generates OAuth URL for connecting Meta Ads account
// @Tags OAuth
// @Accept json
// @Produce json
// @Param redirect_url query string false "Post-OAuth redirect URL"
// @Success 200 {object} map[string]interface{} "OAuth URL"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/meta/connect [get]
func (h *OAuthHandler) InitiateMetaOAuth(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("User not authenticated"))
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("Organization not found"))
		return
	}

	// Get redirect URL from query params, default to frontend callback page
	redirectURL := c.DefaultQuery("redirect_url", "/settings/connections")

	// Generate OAuth URL
	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), userID, orgID, entity.PlatformMeta, redirectURL)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"auth_url": authURL,
			"platform": "meta",
		},
	})
}

// MetaOAuthCallback handles the OAuth callback from Meta
// @Summary Handle Meta OAuth callback
// @Description Processes OAuth callback from Meta after user authorization
// @Tags OAuth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "OAuth state"
// @Param error query string false "Error code if user denied"
// @Param error_description query string false "Error description"
// @Success 302 "Redirect to success page"
// @Failure 302 "Redirect to error page"
// @Router /oauth/meta/callback [get]
func (h *OAuthHandler) MetaOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorCode := c.Query("error")
	errorDesc := c.Query("error_description")
	errorReason := c.Query("error_reason")

	// Handle user denied permission
	if errorCode != "" {
		errType := mapMetaError(errorCode, errorReason)
		redirectURL := buildErrorRedirect("meta", errType, errorDesc)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// Validate required parameters
	if code == "" {
		redirectURL := buildErrorRedirect("meta", "no_code", "Authorization code not provided")
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	if state == "" {
		redirectURL := buildErrorRedirect("meta", "invalid_state", "State parameter missing")
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// Process the callback
	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformMeta, code, state)
	if err != nil {
		errType := "callback_failed"
		errMsg := "Failed to complete authorization"

		// Check for specific error types
		if errors.IsAppError(err) {
			appErr := err.(*errors.AppError)
			switch appErr.Code {
			case errors.ErrCodeBadRequest:
				if appErr.Message == "Invalid OAuth state" {
					errType = "invalid_state"
					errMsg = "Authorization session expired or invalid. Please try again."
				} else if appErr.Message == "OAuth state expired" {
					errType = "state_expired"
					errMsg = "Authorization session has expired. Please try again."
				} else if appErr.Message == "Platform mismatch" {
					errType = "platform_mismatch"
					errMsg = "Platform mismatch detected. Please try again."
				}
			case errors.ErrCodeOAuthFailed:
				errType = "token_exchange_failed"
				errMsg = "Failed to exchange authorization code for access token."
			}
		}

		redirectURL := buildErrorRedirect("meta", errType, errMsg)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
		return
	}

	// Success - redirect with account info
	successURL := buildSuccessRedirect(redirectURL, "meta", account.ID.String())
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// InitiateTikTokOAuth initiates the OAuth flow for TikTok Ads
func (h *OAuthHandler) InitiateTikTokOAuth(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("User not authenticated"))
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("Organization not found"))
		return
	}

	redirectURL := c.DefaultQuery("redirect_url", "/settings/connections")

	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), userID, orgID, entity.PlatformTikTok, redirectURL)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"auth_url": authURL,
			"platform": "tiktok",
		},
	})
}

// InitiateShopeeOAuth initiates the OAuth flow for Shopee Ads
func (h *OAuthHandler) InitiateShopeeOAuth(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("User not authenticated"))
		return
	}

	orgID, ok := middleware.GetOrgID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("Organization not found"))
		return
	}

	redirectURL := c.DefaultQuery("redirect_url", "/settings/connections")

	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), userID, orgID, entity.PlatformShopee, redirectURL)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"auth_url": authURL,
			"platform": "shopee",
		},
	})
}

// Helper functions

// mapMetaError maps Meta OAuth error codes to user-friendly types
func mapMetaError(errorCode, errorReason string) string {
	switch errorCode {
	case "access_denied":
		if errorReason == "user_denied" {
			return "permission_denied"
		}
		return "access_denied"
	case "invalid_request":
		return "invalid_request"
	case "unauthorized_client":
		return "unauthorized_client"
	case "invalid_scope":
		return "invalid_scope"
	case "server_error":
		return "server_error"
	default:
		return "unknown_error"
	}
}

// buildErrorRedirect builds the error redirect URL
func buildErrorRedirect(platform, errType, message string) string {
	params := url.Values{}
	params.Set("platform", platform)
	params.Set("error", errType)
	if message != "" {
		params.Set("message", message)
	}
	return "/oauth/error?" + params.Encode()
}

// buildSuccessRedirect builds the success redirect URL
func buildSuccessRedirect(baseURL, platform, accountID string) string {
	// Parse the base URL
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		// Fallback to simple concatenation
		return baseURL + "?platform=" + platform + "&account_id=" + accountID + "&success=true"
	}

	// Add query parameters
	q := parsedURL.Query()
	q.Set("platform", platform)
	q.Set("account_id", accountID)
	q.Set("success", "true")
	parsedURL.RawQuery = q.Encode()

	return parsedURL.String()
}
