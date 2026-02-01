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
	Mode           string // "debug", "release", "test"
	AllowedOrigins []string
	RateLimitRPS   int
}

// DefaultConfig returns default router configuration
func DefaultConfig() *Config {
	return &Config{
		Mode:           gin.ReleaseMode,
		AllowedOrigins: []string{"*"},
		RateLimitRPS:   100,
	}
}

// Router wraps gin.Engine with additional functionality
type Router struct {
	engine              *gin.Engine
	config              *Config
	authHandler         *handler.AuthHandler
	platformHandler     *handler.PlatformHandler
	analyticsHandler    *handler.AnalyticsHandler
	authMiddleware      *middleware.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewRouter creates a new router
func NewRouter(
	config *Config,
	authHandler *handler.AuthHandler,
	platformHandler *handler.PlatformHandler,
	analyticsHandler *handler.AnalyticsHandler,
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
		authMiddleware:      authMiddleware,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// Setup configures all routes and middleware
func (r *Router) Setup() *gin.Engine {
	// Global middleware
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.RequestLogger())
	r.engine.Use(middleware.RequestID())
	r.engine.Use(r.corsMiddleware())

	// Health check
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

		// Protected routes (auth required)
		r.setupProtectedRoutes(v1)
	}

	return r.engine
}

// setupPublicRoutes configures public API routes
func (r *Router) setupPublicRoutes(rg *gin.RouterGroup) {
	// Auth routes
	auth := rg.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)
		auth.POST("/forgot-password", r.authHandler.ForgotPassword)
		auth.POST("/reset-password", r.authHandler.ResetPassword)
	}

	// OAuth callbacks (public but validated by state)
	oauth := rg.Group("/oauth")
	{
		oauth.GET("/meta/callback", r.authHandler.MetaCallback)
		oauth.GET("/tiktok/callback", r.authHandler.TikTokCallback)
		oauth.GET("/shopee/callback", r.authHandler.ShopeeCallback)
	}
}

// setupProtectedRoutes configures protected API routes
func (r *Router) setupProtectedRoutes(rg *gin.RouterGroup) {
	// Apply auth middleware
	protected := rg.Group("")
	protected.Use(r.authMiddleware.Authenticate())

	// User routes
	user := protected.Group("/user")
	{
		user.GET("/me", r.authHandler.GetCurrentUser)
		user.PUT("/me", r.authHandler.UpdateProfile)
		user.POST("/change-password", r.authHandler.ChangePassword)
	}

	// Organization routes
	org := protected.Group("/organizations")
	{
		org.GET("", r.authHandler.ListOrganizations)
		org.GET("/:orgId", r.authHandler.GetOrganization)
		org.PUT("/:orgId", r.authHandler.UpdateOrganization)
		org.GET("/:orgId/members", r.authHandler.ListMembers)
		org.POST("/:orgId/members", r.authHandler.InviteMember)
		org.DELETE("/:orgId/members/:memberId", r.authHandler.RemoveMember)
	}

	// Platform connection routes
	platforms := protected.Group("/platforms")
	{
		platforms.GET("", r.platformHandler.ListConnectedAccounts)
		platforms.GET("/:platform/auth-url", r.platformHandler.GetAuthURL)
		platforms.DELETE("/:accountId", r.platformHandler.DisconnectAccount)
		platforms.POST("/:accountId/sync", r.platformHandler.TriggerSync)
		platforms.GET("/:accountId/sync-status", r.platformHandler.GetSyncStatus)
	}

	// Ad accounts routes
	adAccounts := protected.Group("/ad-accounts")
	{
		adAccounts.GET("", r.platformHandler.ListAdAccounts)
		adAccounts.GET("/:accountId", r.platformHandler.GetAdAccount)
	}

	// Campaign routes
	campaigns := protected.Group("/campaigns")
	{
		campaigns.GET("", r.analyticsHandler.ListCampaigns)
		campaigns.GET("/:campaignId", r.analyticsHandler.GetCampaign)
		campaigns.GET("/:campaignId/metrics", r.analyticsHandler.GetCampaignMetrics)
		campaigns.GET("/:campaignId/ad-sets", r.analyticsHandler.ListAdSets)
		campaigns.GET("/:campaignId/ads", r.analyticsHandler.ListAds)
	}

	// Ad set routes
	adSets := protected.Group("/ad-sets")
	{
		adSets.GET("/:adSetId", r.analyticsHandler.GetAdSet)
		adSets.GET("/:adSetId/metrics", r.analyticsHandler.GetAdSetMetrics)
		adSets.GET("/:adSetId/ads", r.analyticsHandler.ListAdsByAdSet)
	}

	// Ad routes
	ads := protected.Group("/ads")
	{
		ads.GET("/:adId", r.analyticsHandler.GetAd)
		ads.GET("/:adId/metrics", r.analyticsHandler.GetAdMetrics)
	}

	// Analytics routes
	analytics := protected.Group("/analytics")
	{
		analytics.GET("/dashboard", r.analyticsHandler.GetDashboard)
		analytics.GET("/overview", r.analyticsHandler.GetOverview)
		analytics.GET("/platform-comparison", r.analyticsHandler.GetPlatformComparison)
		analytics.GET("/trends", r.analyticsHandler.GetTrends)
		analytics.GET("/top-performers", r.analyticsHandler.GetTopPerformers)
		analytics.POST("/calculate", r.analyticsHandler.CalculateMetrics) // New analytics calculation endpoint
		analytics.POST("/reports/generate", r.analyticsHandler.GenerateReport)
		analytics.GET("/reports/:reportId", r.analyticsHandler.GetReport)
	}

	// Webhook routes (for platform integrations)
	webhooks := protected.Group("/webhooks")
	{
		webhooks.POST("/meta", r.platformHandler.MetaWebhook)
		webhooks.POST("/tiktok", r.platformHandler.TikTokWebhook)
		webhooks.POST("/shopee", r.platformHandler.ShopeeWebhook)
	}
}

// corsMiddleware returns CORS middleware
func (r *Router) corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     r.config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// healthCheck returns OK if the server is running
func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// readinessCheck returns OK if the server is ready to accept requests
func (r *Router) readinessCheck(c *gin.Context) {
	// TODO: Add database and cache connectivity checks
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
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
