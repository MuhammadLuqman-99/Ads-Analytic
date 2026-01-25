package handler

import (
	"net/http"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/auth"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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

	c.JSON(http.StatusCreated, gin.H{
		"data": result,
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

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	tokens, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		respondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tokens,
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

	// TODO: Implement forgot password logic
	c.JSON(http.StatusOK, gin.H{
		"message": "If an account exists with this email, a password reset link has been sent",
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

	// TODO: Implement reset password logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Password has been reset successfully",
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	email, _ := middleware.GetEmail(c)
	role, _ := middleware.GetRole(c)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":    userID,
			"email": email,
			"role":  role,
		},
	})
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, errors.ErrValidation(err.Error()))
		return
	}

	// TODO: Implement profile update
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// ChangePassword changes the current user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
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
		"message": "Password changed successfully",
	})
}

// ListOrganizations lists organizations for the current user
func (h *AuthHandler) ListOrganizations(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{
		"data": []interface{}{},
	})
}

// GetOrganization returns an organization by ID
func (h *AuthHandler) GetOrganization(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{
		"data": nil,
	})
}

// UpdateOrganization updates an organization
func (h *AuthHandler) UpdateOrganization(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{
		"message": "Organization updated",
	})
}

// ListMembers lists organization members
func (h *AuthHandler) ListMembers(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{
		"data": []interface{}{},
	})
}

// InviteMember invites a new member to an organization
func (h *AuthHandler) InviteMember(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusCreated, gin.H{
		"message": "Invitation sent",
	})
}

// RemoveMember removes a member from an organization
func (h *AuthHandler) RemoveMember(c *gin.Context) {
	// TODO: Implement
	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed",
	})
}

// OAuth Callbacks

// MetaCallback handles Meta OAuth callback
func (h *AuthHandler) MetaCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorCode := c.Query("error")

	if errorCode != "" {
		errorDesc := c.Query("error_description")
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=meta&error="+errorCode+"&message="+errorDesc)
		return
	}

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformMeta, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=meta&error=callback_failed")
		return
	}

	// Redirect to success page with account ID
	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=meta&account_id="+account.ID.String())
}

// TikTokCallback handles TikTok OAuth callback
func (h *AuthHandler) TikTokCallback(c *gin.Context) {
	code := c.Query("auth_code")
	state := c.Query("state")

	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=tiktok&error=no_code")
		return
	}

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformTikTok, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=tiktok&error=callback_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=tiktok&account_id="+account.ID.String())
}

// ShopeeCallback handles Shopee OAuth callback
func (h *AuthHandler) ShopeeCallback(c *gin.Context) {
	code := c.Query("code")
	shopID := c.Query("shop_id")

	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=shopee&error=no_code")
		return
	}

	// For Shopee, the state might be encoded differently
	state := c.Query("state")

	account, redirectURL, err := h.authService.HandleOAuthCallback(c.Request.Context(), entity.PlatformShopee, code, state)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/oauth/error?platform=shopee&error=callback_failed")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectURL+"?platform=shopee&account_id="+account.ID.String()+"&shop_id="+shopID)
}

// Helper functions

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
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
		return
	}

	c.JSON(status, gin.H{
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
