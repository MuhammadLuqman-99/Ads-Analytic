package router

import (
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/handler"
	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Config holds router configuration
type Config struct {
	Mode           string   // "debug", "release", "test"
	AllowedOrigins []string // CORS allowed origins
	RateLimitRPS   int      // Requests per second
}

// DefaultConfig returns default router configuration
func DefaultConfig() *Config {
	return &Config{
		Mode: gin.ReleaseMode,
		AllowedOrigins: []string{
			"http://localhost:3000",      // Next.js dev
			"http://localhost:3001",      // Alternative dev port
			"https://yourdomain.com",     // Production
			"https://app.yourdomain.com", // Production app subdomain
		},
		RateLimitRPS: 100,
	}
}

// Router wraps gin.Engine with additional functionality
type Router struct {
	engine *gin.Engine
	config *Config

	// Handlers
	authHandler      *handler.AuthHandler
	platformHandler  *handler.PlatformHandler
	analyticsHandler *handler.AnalyticsHandler
	eventsHandler    *handler.EventsHandler

	// Middleware
	authMiddleware      *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewRouter creates a new router
func NewRouter(
	config *Config,
	authHandler *handler.AuthHandler,
	platformHandler *handler.PlatformHandler,
	analyticsHandler *handler.AnalyticsHandler,
	eventsHandler *handler.EventsHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) *Router {
	if config == nil {
		config = DefaultConfig()
	}

	gin.SetMode(config.Mode)

	return &Router{
		engine:              gin.New(),
		config:              config,
		authHandler:         authHandler,
		platformHandler:     platformHandler,
		analyticsHandler:    analyticsHandler,
		eventsHandler:       eventsHandler,
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// Setup configures all routes and middleware
// Middleware chain: Request → Recovery → RequestID → Logger → CORS → SecureHeaders → (RateLimit → Auth) → Handler
func (r *Router) Setup() *gin.Engine {
	// Global middleware
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.RequestLogger())
	r.engine.Use(r.corsMiddleware())
	r.engine.Use(middleware.SecureHeaders())

	// Health check endpoints (no rate limit, no auth)
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readinessCheck)

	// API v1
	v1 := r.engine.Group("/api/v1")
	{
		// Apply rate limiting to all API routes
		if r.rateLimitMiddleware != nil {
			v1.Use(r.rateLimitMiddleware.Handle())
		}

		// Public routes (no auth required)
		r.setupPublicRoutes(v1)

		// OAuth callbacks (no auth but validated by state)
		r.setupOAuthCallbacks(v1)

		// Protected routes (auth required)
		r.setupProtectedRoutes(v1)
	}

	return r.engine
}

// setupPublicRoutes configures public API routes
func (r *Router) setupPublicRoutes(rg *gin.RouterGroup) {
	// Auth routes - no authentication required
	auth := rg.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)
		auth.POST("/forgot-password", r.authHandler.ForgotPassword)
		auth.POST("/reset-password", r.authHandler.ResetPassword)
		auth.POST("/verify-email", r.authHandler.VerifyEmail)

		// Session check - optionally authenticated (returns user if token valid)
		auth.GET("/session", r.authMiddleware.OptionalAuth(), r.authHandler.GetSession)

		// Logout - can work with or without valid token (clears cookies)
		auth.POST("/logout", r.authHandler.Logout)
	}
}

// setupOAuthCallbacks configures OAuth callback routes
func (r *Router) setupOAuthCallbacks(rg *gin.RouterGroup) {
	// OAuth callbacks - validated by state parameter
	oauth := rg.Group("/auth")
	{
		oauth.GET("/meta/callback", r.authHandler.MetaCallback)
		oauth.GET("/tiktok/callback", r.authHandler.TikTokCallback)
		oauth.GET("/shopee/callback", r.authHandler.ShopeeCallback)
	}
}

// setupProtectedRoutes configures protected API routes
func (r *Router) setupProtectedRoutes(rg *gin.RouterGroup) {
	// Apply auth middleware to all protected routes
	protected := rg.Group("")
	protected.Use(r.authMiddleware.Authenticate())

	// ============================================
	// Dashboard routes (mapped to analytics)
	// ============================================
	dashboard := protected.Group("/dashboard")
	{
		dashboard.GET("/summary", r.analyticsHandler.GetDashboard)
		dashboard.GET("/timeseries", r.analyticsHandler.GetTrends)
		dashboard.GET("/platforms", r.analyticsHandler.GetPlatformComparison)
		dashboard.GET("/top-campaigns", r.analyticsHandler.GetTopPerformers)
	}

	// ============================================
	// Events routes (SSE for real-time updates)
	// ============================================
	if r.eventsHandler != nil {
		events := protected.Group("/events")
		{
			events.GET("/stream", r.eventsHandler.Stream)
			events.GET("/status", r.eventsHandler.GetStatus)
		}
	}

	// ============================================
	// Campaign routes
	// ============================================
	campaigns := protected.Group("/campaigns")
	{
		campaigns.GET("", r.analyticsHandler.ListCampaigns)
		campaigns.GET("/:id", r.analyticsHandler.GetCampaign)
		campaigns.GET("/:id/metrics", r.analyticsHandler.GetCampaignMetrics)
		campaigns.GET("/:id/ad-sets", r.analyticsHandler.ListAdSets)
		campaigns.GET("/:id/ads", r.analyticsHandler.ListAds)
	}

	// ============================================
	// Connections routes (mapped to platform handler)
	// ============================================
	connections := protected.Group("/connections")
	{
		connections.GET("", r.platformHandler.ListConnectedAccounts)
		connections.GET("/platforms", r.getAvailablePlatforms)
		connections.POST("/connect/:platform", r.platformHandler.GetAuthURL)
		connections.DELETE("/:id", r.platformHandler.DisconnectAccount)
		connections.POST("/:id/sync", r.platformHandler.TriggerSync)
		connections.GET("/:id/sync-status", r.platformHandler.GetSyncStatus)
	}

	// ============================================
	// Analytics routes
	// ============================================
	analytics := protected.Group("/analytics")
	{
		analytics.GET("/summary", r.analyticsHandler.GetOverview)
		analytics.GET("/comparison", r.analyticsHandler.GetPlatformComparison)
		analytics.GET("/trends", r.analyticsHandler.GetTrends)
		analytics.GET("/top-performers", r.analyticsHandler.GetTopPerformers)
		analytics.GET("/export", r.analyticsHandler.ExportAnalytics)
		analytics.POST("/calculate", r.analyticsHandler.CalculateMetrics)
		analytics.POST("/reports/generate", r.analyticsHandler.GenerateReport)
		analytics.GET("/reports/:reportId", r.analyticsHandler.GetReport)
	}

	// ============================================
	// Settings routes (mapped to auth handler)
	// ============================================
	settings := protected.Group("/settings")
	{
		// Profile
		settings.GET("/profile", r.authHandler.GetCurrentUser)
		settings.PUT("/profile", r.authHandler.UpdateProfile)

		// Organization
		settings.GET("/organization", r.getOrganizationSettings)
		settings.PUT("/organization", r.updateOrganizationSettings)

		// Team
		settings.GET("/team", r.authHandler.ListMembers)
		settings.POST("/team/invite", r.authHandler.InviteMember)
		settings.DELETE("/team/:memberId", r.authHandler.RemoveMember)

		// Billing (placeholder)
		settings.GET("/billing", r.getBillingInfo)
	}

	// ============================================
	// User routes
	// ============================================
	user := protected.Group("/user")
	{
		user.GET("/me", r.authHandler.GetCurrentUser)
		user.PUT("/me", r.authHandler.UpdateProfile)
		user.POST("/change-password", r.authHandler.ChangePassword)
		user.POST("/resend-verification", r.authHandler.ResendVerificationEmail)
	}

	// ============================================
	// Organization routes
	// ============================================
	org := protected.Group("/organizations")
	{
		org.GET("", r.authHandler.ListOrganizations)
		org.GET("/:orgId", r.authHandler.GetOrganization)
		org.PUT("/:orgId", r.authHandler.UpdateOrganization)
		org.GET("/:orgId/members", r.authHandler.ListMembers)
		org.POST("/:orgId/members", r.authHandler.InviteMember)
		org.DELETE("/:orgId/members/:memberId", r.authHandler.RemoveMember)
	}

	// ============================================
	// Platform routes (legacy)
	// ============================================
	platforms := protected.Group("/platforms")
	{
		platforms.GET("", r.platformHandler.ListConnectedAccounts)
		platforms.GET("/auth-url/:platform", r.platformHandler.GetAuthURL)
		platforms.DELETE("/:accountId", r.platformHandler.DisconnectAccount)
		platforms.POST("/:accountId/sync", r.platformHandler.TriggerSync)
		platforms.GET("/:accountId/sync-status", r.platformHandler.GetSyncStatus)
	}

	// ============================================
	// Ad accounts routes
	// ============================================
	adAccounts := protected.Group("/ad-accounts")
	{
		adAccounts.GET("", r.platformHandler.ListAdAccounts)
		adAccounts.GET("/:accountId", r.platformHandler.GetAdAccount)
	}

	// ============================================
	// Ad set routes
	// ============================================
	adSets := protected.Group("/ad-sets")
	{
		adSets.GET("/:adSetId", r.analyticsHandler.GetAdSet)
		adSets.GET("/:adSetId/metrics", r.analyticsHandler.GetAdSetMetrics)
		adSets.GET("/:adSetId/ads", r.analyticsHandler.ListAdsByAdSet)
	}

	// ============================================
	// Ad routes
	// ============================================
	ads := protected.Group("/ads")
	{
		ads.GET("/:adId", r.analyticsHandler.GetAd)
		ads.GET("/:adId/metrics", r.analyticsHandler.GetAdMetrics)
	}

	// ============================================
	// Webhook routes
	// ============================================
	webhooks := protected.Group("/webhooks")
	{
		webhooks.POST("/meta", r.platformHandler.MetaWebhook)
		webhooks.POST("/tiktok", r.platformHandler.TikTokWebhook)
		webhooks.POST("/shopee", r.platformHandler.ShopeeWebhook)
	}
}

// corsMiddleware returns CORS middleware with proper configuration
func (r *Router) corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     r.config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// healthCheck returns OK if the server is running
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// readinessCheck returns OK if the server is ready to accept requests
func (r *Router) readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// getAvailablePlatforms returns list of supported platforms
func (r *Router) getAvailablePlatforms(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": []gin.H{
			{"id": "meta", "name": "Meta Ads", "description": "Facebook & Instagram Ads", "icon": "meta"},
			{"id": "tiktok", "name": "TikTok Ads", "description": "TikTok for Business", "icon": "tiktok"},
			{"id": "shopee", "name": "Shopee Ads", "description": "Shopee Seller Ads", "icon": "shopee"},
		},
	})
}

// getOrganizationSettings returns organization settings
func (r *Router) getOrganizationSettings(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":       orgID,
			"currency": "MYR",
			"timezone": "Asia/Kuala_Lumpur",
		},
	})
}

// updateOrganizationSettings updates organization settings
func (r *Router) updateOrganizationSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "Settings updated"},
	})
}

// getBillingInfo returns billing information
func (r *Router) getBillingInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"plan":   "free",
			"status": "active",
		},
	})
}

// Engine returns the underlying gin.Engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
