package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ads-aggregator/ads-aggregator/config"
	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/handler"
	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/middleware"
	"github.com/ads-aggregator/ads-aggregator/internal/delivery/http/router"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/cache"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/events"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/persistence"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/persistence/postgres"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform/meta"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform/shopee"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform/tiktok"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/auth"
	"github.com/ads-aggregator/ads-aggregator/pkg/jwt"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("app", cfg.App.Name).
		Str("env", cfg.App.Env).
		Int("port", cfg.App.Port).
		Msg("Starting application")

	// Initialize database
	db, err := postgres.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()
	log.Info().Str("host", cfg.Database.Host).Msg("Database connected")

	// Initialize platform connectors
	connectorRegistry := initConnectors(cfg)
	log.Info().Int("connectors", len(connectorRegistry.List())).Msg("Platform connectors initialized")

	// Initialize Redis client for caching
	var redisClient *redis.Client
	var appCache *cache.Cache
	if cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Warn().Err(err).Msg("Failed to connect to Redis, caching disabled")
			redisClient = nil
		} else {
			log.Info().Str("addr", cfg.Redis.Addr()).Msg("Redis connected")
			appCache = cache.NewCache(redisClient)
		}
	}

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		cfg.RateLimit.Requests,
		cfg.RateLimit.Requests*2, // burst
	)

	// Initialize event broadcaster for SSE
	broadcaster := events.NewBroadcaster()
	log.Info().Msg("Event broadcaster initialized")

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	orgRepo := postgres.NewOrganizationRepository(db)
	orgMemberRepo := postgres.NewOrganizationMemberRepository(db)
	connectedAccRepo := postgres.NewConnectedAccountRepository(db)
	verificationTokenRepo := postgres.NewVerificationTokenRepository(db)

	// Initialize state store for OAuth
	stateStore := persistence.NewInMemoryStateStore(15 * time.Minute)

	// Initialize auth service
	frontendURL := os.Getenv("APP_URL")
	if frontendURL == "" {
		frontendURL = fmt.Sprintf("http://localhost:%d", cfg.App.Port)
	}
	emailConfig := auth.EmailConfig{
		FrontendURL:             frontendURL,
		EmailVerificationExpiry: 24 * time.Hour,
		PasswordResetExpiry:     1 * time.Hour,
		AppName:                 cfg.App.Name,
		SupportEmail:            "support@example.com",
	}
	authService := auth.NewService(
		userRepo,
		orgRepo,
		orgMemberRepo,
		connectedAccRepo,
		verificationTokenRepo,
		jwtManager,
		connectorRegistry,
		stateStore,
		nil, // emailSender - not required for local testing
		emailConfig,
	)
	log.Info().Msg("Auth service initialized")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	platformHandler := handler.NewPlatformHandler(nil, nil)
	analyticsHandler := handler.NewAnalyticsHandler(nil, appCache)
	eventsHandler := handler.NewEventsHandler(broadcaster)

	// Initialize router
	routerConfig := &router.Config{
		Mode:           "release",
		AllowedOrigins: []string{"*"},
		RateLimitRPS:   cfg.RateLimit.Requests,
	}
	if cfg.IsDevelopment() {
		routerConfig.Mode = "debug"
	}

	r := router.NewRouter(
		routerConfig,
		authHandler,
		platformHandler,
		analyticsHandler,
		eventsHandler,
		authMiddleware,
		rateLimitMiddleware,
	)
	engine := r.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Msgf("Server listening on port %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	// Close Redis connection
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Warn().Err(err).Msg("Error closing Redis connection")
		}
	}

	log.Info().Msg("Server exited")
}

// initConnectors initializes platform connectors
func initConnectors(cfg *config.Config) *platform.ConnectorRegistry {
	registry := platform.NewConnectorRegistry()

	// Meta connector
	metaConnector := meta.NewConnector(&meta.Config{
		AppID:           cfg.Meta.AppID,
		AppSecret:       cfg.Meta.AppSecret,
		RedirectURI:     cfg.Meta.RedirectURI,
		APIVersion:      cfg.Meta.APIVersion,
		RateLimitCalls:  cfg.Meta.RateLimitCalls,
		RateLimitWindow: cfg.Meta.RateLimitWindow,
		Timeout:         cfg.HTTP.Timeout,
		MaxRetries:      cfg.HTTP.MaxRetries,
	})
	registry.Register(entity.PlatformMeta, metaConnector)

	// TikTok connector
	tiktokConnector := tiktok.NewConnector(&tiktok.Config{
		AppID:           cfg.TikTok.AppID,
		AppSecret:       cfg.TikTok.AppSecret,
		RedirectURI:     cfg.TikTok.RedirectURI,
		RateLimitCalls:  cfg.TikTok.RateLimitCalls,
		RateLimitWindow: cfg.TikTok.RateLimitWindow,
		Timeout:         cfg.HTTP.Timeout,
		MaxRetries:      cfg.HTTP.MaxRetries,
	})
	registry.Register(entity.PlatformTikTok, tiktokConnector)

	// Shopee connector
	shopeePartnerID := int64(0)
	fmt.Sscanf(cfg.Shopee.PartnerID, "%d", &shopeePartnerID)
	shopeeConnector := shopee.NewConnector(&shopee.Config{
		PartnerID:       shopeePartnerID,
		PartnerKey:      cfg.Shopee.PartnerKey,
		RedirectURI:     cfg.Shopee.RedirectURI,
		RateLimitCalls:  cfg.Shopee.RateLimitCalls,
		RateLimitWindow: cfg.Shopee.RateLimitWindow,
		Timeout:         cfg.HTTP.Timeout,
		MaxRetries:      cfg.HTTP.MaxRetries,
	})
	registry.Register(entity.PlatformShopee, shopeeConnector)

	return registry
}
