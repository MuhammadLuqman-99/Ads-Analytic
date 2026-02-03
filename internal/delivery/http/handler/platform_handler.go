package handler

import (
	"net/http"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/auth"
	syncUsecase "github.com/ads-aggregator/ads-aggregator/internal/usecase/sync"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/gin-gonic/gin"
)

// PlatformHandler handles platform-related HTTP requests
type PlatformHandler struct {
	authService *auth.Service
	syncService *syncUsecase.Service
}

// NewPlatformHandler creates a new platform handler
func NewPlatformHandler(authService *auth.Service, syncService *syncUsecase.Service) *PlatformHandler {
	return &PlatformHandler{authService: authService, syncService: syncService}
}

// ListConnectedAccounts lists all connected platform accounts
func (h *PlatformHandler) ListConnectedAccounts(c *gin.Context) {
	// Return mock data if auth service is not available (local dev mode)
	if h.authService == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    getMockConnections(),
			"total":   3,
			"mock":    true,
		})
		return
	}

	orgID, _ := middleware.GetOrgID(c)
	accounts, err := h.authService.GetConnectedAccounts(c.Request.Context(), orgID)
	if err != nil {
		respondWithError(c, err)
		return
	}

	response := make([]gin.H, len(accounts))
	for i, acc := range accounts {
		response[i] = gin.H{
			"id": acc.ID, "platform": acc.Platform, "status": acc.Status,
			"platform_account_name": acc.PlatformAccountName, "last_synced_at": acc.LastSyncedAt,
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": response, "total": len(response)})
}

// GetAuthURL returns the OAuth authorization URL for a platform
func (h *PlatformHandler) GetAuthURL(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	orgID, _ := middleware.GetOrgID(c)
	platform := entity.Platform(c.Param("platform"))
	if !platform.IsValid() {
		respondWithError(c, errors.ErrBadRequest("Invalid platform"))
		return
	}

	authURL, err := h.authService.GetOAuthURL(c.Request.Context(), userID, orgID, platform, c.Query("redirect_url"))
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"auth_url": authURL, "platform": platform})
}

// DisconnectAccount disconnects a platform account
func (h *PlatformHandler) DisconnectAccount(c *gin.Context) {
	accountID, err := parseUUID(c, "accountId")
	if err != nil {
		respondWithError(c, errors.ErrBadRequest("Invalid account ID"))
		return
	}
	if err := h.authService.DisconnectPlatform(c.Request.Context(), accountID); err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account disconnected"})
}

// TriggerSync triggers a sync for a platform account
func (h *PlatformHandler) TriggerSync(c *gin.Context) {
	accountID, err := parseUUID(c, "accountId")
	if err != nil {
		respondWithError(c, errors.ErrBadRequest("Invalid account ID"))
		return
	}
	result, err := h.syncService.SyncAccount(c.Request.Context(), accountID)
	if err != nil {
		respondWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"campaigns_synced": result.CampaignsSynced, "ad_sets_synced": result.AdSetsSynced,
		"ads_synced": result.AdsSynced, "errors": len(result.Errors),
	}})
}

// GetSyncStatus gets the sync status
func (h *PlatformHandler) GetSyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "idle"}})
}

// ListAdAccounts lists all ad accounts
func (h *PlatformHandler) ListAdAccounts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
}

// GetAdAccount gets an ad account by ID
func (h *PlatformHandler) GetAdAccount(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": nil})
}

// Webhooks
func (h *PlatformHandler) MetaWebhook(c *gin.Context)   { c.Status(http.StatusOK) }
func (h *PlatformHandler) TikTokWebhook(c *gin.Context) { c.Status(http.StatusOK) }
func (h *PlatformHandler) ShopeeWebhook(c *gin.Context) { c.Status(http.StatusOK) }

// getMockConnections returns mock connected accounts for local testing
func getMockConnections() []gin.H {
	return []gin.H{
		{
			"id":                    "conn-meta-001",
			"platform":              "meta",
			"status":                "active",
			"platform_account_name": "My Business Page",
			"platform_account_id":   "123456789",
			"last_synced_at":        "2026-02-03T02:30:00Z",
		},
		{
			"id":                    "conn-tiktok-001",
			"platform":              "tiktok",
			"status":                "active",
			"platform_account_name": "TikTok Ads Account",
			"platform_account_id":   "987654321",
			"last_synced_at":        "2026-02-03T02:15:00Z",
		},
		{
			"id":                    "conn-shopee-001",
			"platform":              "shopee",
			"status":                "active",
			"platform_account_name": "Shopee Seller Center",
			"platform_account_id":   "456789123",
			"last_synced_at":        "2026-02-03T02:00:00Z",
		},
	}
}
