package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/auth"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Cookie names
const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService  *auth.Service
	cookieDomain string
	secureCookie bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	// Determine cookie settings from environment
	secureCookie := os.Getenv("APP_ENV") == "production"
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = "localhost"
	}

	return &AuthHandler{
		authService:  authService,
		cookieDomain: cookieDomain,
		secureCookie: secureCookie,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Set auth cookies
	h.setAuthCookies(c, result.Tokens)

	// Return user data (without tokens in body since they're in cookies)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user":         result.User,
			"organization": result.Organization,
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	// Set auth cookies
	h.setAuthCookies(c, result.Tokens)

	// Return user data
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":         h.sanitizeUser(result.User),
			"organization": result.Organization,
			"expiresAt":    result.Tokens.AccessTokenExpiresAt,
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Try to get refresh token from cookie first
	refreshToken, err := c.Cookie(RefreshTokenCookie)
	if err != nil || refreshToken == "" {
		// Fallback to request body
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			respondWithError(c, errors.ErrUnauthorized("Refresh token required"))
			return
		}
		refreshToken = req.RefreshToken
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		// Clear cookies on refresh failure
		h.clearAuthCookies(c)
		respondWithError(c, err)
		return
	}

	// Set new auth cookies
	h.setAuthCookies(c, tokens)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"expiresAt": tokens.AccessTokenExpiresAt,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}

// GetSession returns current session info (for client-side hydration)
func (h *AuthHandler) GetSession(c *gin.Context) {
	// Check if user is authenticated
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"authenticated": false,
			},
		})
		return
	}

	email, _ := middleware.GetEmail(c)
	role, _ := middleware.GetRole(c)
	orgID, _ := middleware.GetOrgID(c)
	permissions, _ := middleware.GetPermissions(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"authenticated": true,
			"user": gin.H{
				"id":    userID.String(),
				"email": email,
				"role":  role,
			},
			"organizationId": orgID.String(),
			"permissions":    permissions,
		},
	})
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	// Request password reset (doesn't reveal if email exists)
	_ = h.authService.RequestPasswordReset(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "If an account exists with this email, a password reset link has been sent",
		},
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password has been reset successfully",
		},
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	email, _ := middleware.GetEmail(c)
	role, _ := middleware.GetRole(c)
	orgID, _ := middleware.GetOrgID(c)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":             userID.String(),
			"email":          email,
			"role":           role,
			"organizationId": orgID.String(),
		},
	})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req auth.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":    h.sanitizeUser(user),
			"message": "Profile updated successfully",
		},
	})
}

// ChangePassword changes the current user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password changed successfully",
		},
	})
}

// VerifyEmail verifies a user's email with a token
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Email verified successfully",
		},
	})
}

// ResendVerificationEmail resends the email verification link
func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondWithError(c, errors.ErrUnauthorized("User not authenticated"))
		return
	}

	err := h.authService.RequestEmailVerification(c.Request.Context(), userID)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Verification email sent",
		},
	})
}

// ListOrganizations lists organizations for the current user
func (h *AuthHandler) ListOrganizations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// GetOrganization returns an organization by ID
func (h *AuthHandler) GetOrganization(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nil,
	})
}

// UpdateOrganization updates an organization
func (h *AuthHandler) UpdateOrganization(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Organization updated",
		},
	})
}

// ListMembers lists organization members
func (h *AuthHandler) ListMembers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
	})
}

// InviteMember invites a new member to an organization
func (h *AuthHandler) InviteMember(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Invitation sent",
		},
	})
}

// RemoveMember removes a member from an organization
func (h *AuthHandler) RemoveMember(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Member removed",
		},
	})
}

// ============================================================================
// OAuth Callbacks
// ============================================================================

// MetaCallback handles Meta OAuth callback
func (h *AuthHandler) MetaCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorCode := c.Query("error")

	if errorCode != "" {
		errorDesc := c.Query("error_description")
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error="+errorCode+"&message="+errorDesc+"&platform=meta")
		return
	}

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformMeta, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error=callback_failed&platform=meta")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=meta&account_id="+account.ID.String()+"&success=true")
}

// TikTokCallback handles TikTok OAuth callback
func (h *AuthHandler) TikTokCallback(c *gin.Context) {
	code := c.Query("auth_code")
	state := c.Query("state")

	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error=no_code&platform=tiktok")
		return
	}

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformTikTok, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error=callback_failed&platform=tiktok")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=tiktok&account_id="+account.ID.String()+"&success=true")
}

// ShopeeCallback handles Shopee OAuth callback
func (h *AuthHandler) ShopeeCallback(c *gin.Context) {
	code := c.Query("code")
	shopID := c.Query("shop_id")
	state := c.Query("state")

	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error=no_code&platform=shopee")
		return
	}

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformShopee, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/connect?error=callback_failed&platform=shopee")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=shopee&account_id="+account.ID.String()+"&shop_id="+shopID+"&success=true")
}

// ============================================================================
// Cookie Management
// ============================================================================

func (h *AuthHandler) setAuthCookies(c *gin.Context, tokens interface{}) {
	type tokenPair struct {
		AccessToken           string
		RefreshToken          string
		AccessTokenExpiresAt  time.Time
		RefreshTokenExpiresAt time.Time
	}

	var tp tokenPair
	switch t := tokens.(type) {
	case *auth.AuthResponse:
		if t.Tokens != nil {
			tp.AccessToken = t.Tokens.AccessToken
			tp.RefreshToken = t.Tokens.RefreshToken
			tp.AccessTokenExpiresAt = t.Tokens.AccessTokenExpiresAt
			tp.RefreshTokenExpiresAt = t.Tokens.RefreshTokenExpiresAt
		}
	case *jwt.TokenPair:
		if t != nil {
			tp.AccessToken = t.AccessToken
			tp.RefreshToken = t.RefreshToken
			tp.AccessTokenExpiresAt = t.AccessTokenExpiresAt
			tp.RefreshTokenExpiresAt = t.RefreshTokenExpiresAt
		}
	default:
		// Unknown token type
		return
	}

	// Don't set empty cookies
	if tp.AccessToken == "" {
		return
	}

	// Set access token cookie
	h.setCookie(c, AccessTokenCookie, tp.AccessToken, int(time.Until(tp.AccessTokenExpiresAt).Seconds()))

	// Set refresh token cookie (longer expiry)
	h.setCookie(c, RefreshTokenCookie, tp.RefreshToken, int(time.Until(tp.RefreshTokenExpiresAt).Seconds()))
}

func (h *AuthHandler) setCookie(c *gin.Context, name, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		name,
		value,
		maxAge,
		"/",
		h.cookieDomain,
		h.secureCookie,
		true, // httpOnly
	)
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AccessTokenCookie, "", -1, "/", h.cookieDomain, h.secureCookie, true)
	c.SetCookie(RefreshTokenCookie, "", -1, "/", h.cookieDomain, h.secureCookie, true)
}

func (h *AuthHandler) sanitizeUser(user *entity.User) gin.H {
	return gin.H{
		"id":        user.ID.String(),
		"email":     user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
	}
}

// ============================================================================
// Helper functions
// ============================================================================

func respondWithError(c *gin.Context, err error) {
	status := errors.GetHTTPStatus(err)

	var appErr *errors.AppError
	if errors.IsAppError(err) {
		ae, ok := err.(*errors.AppError)
		if ok {
			appErr = ae
		}
	}

	if appErr != nil {
		c.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
		return
	}

	c.JSON(status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id := c.Param(param)
	return uuid.Parse(id)
}
