package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Meta      MetaConfig
	TikTok    TikTokConfig
	Shopee    ShopeeConfig
	Scheduler SchedulerConfig
	Log       LogConfig
	RateLimit RateLimitConfig
	HTTP      HTTPClientConfig
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name          string
	Env           string
	Port          int
	Debug         bool
	EncryptionKey string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DSN returns the database connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret              string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
}

// MetaConfig holds Meta (Facebook) API configuration
type MetaConfig struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	APIVersion      string
	RateLimitCalls  int
	RateLimitWindow time.Duration
}

// TikTokConfig holds TikTok API configuration
type TikTokConfig struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	RateLimitCalls  int
	RateLimitWindow time.Duration
}

// ShopeeConfig holds Shopee API configuration
type ShopeeConfig struct {
	PartnerID       string
	PartnerKey      string
	RedirectURI     string
	RateLimitCalls  int
	RateLimitWindow time.Duration
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	Enabled                    bool
	SyncCronSchedule           string
	TokenRefreshCronSchedule   string
	MetricsAggregationSchedule string
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string
	Format string
	Output string
}

// RateLimitConfig holds API rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	Timeout      time.Duration
	MaxRetries   int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:          getEnv("APP_NAME", "ads-aggregator"),
			Env:           getEnv("APP_ENV", "development"),
			Port:          getEnvAsInt("APP_PORT", 8080),
			Debug:         getEnvAsBool("APP_DEBUG", true),
			EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "ads_aggregator"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", ""),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		Meta: MetaConfig{
			AppID:           getEnv("META_APP_ID", ""),
			AppSecret:       getEnv("META_APP_SECRET", ""),
			RedirectURI:     getEnv("META_REDIRECT_URI", "http://localhost:8080/api/v1/oauth/meta/callback"),
			APIVersion:      getEnv("META_API_VERSION", "v18.0"),
			RateLimitCalls:  getEnvAsInt("META_RATE_LIMIT_CALLS", 200),
			RateLimitWindow: getEnvAsDuration("META_RATE_LIMIT_WINDOW", time.Hour),
		},
		TikTok: TikTokConfig{
			AppID:           getEnv("TIKTOK_APP_ID", ""),
			AppSecret:       getEnv("TIKTOK_APP_SECRET", ""),
			RedirectURI:     getEnv("TIKTOK_REDIRECT_URI", "http://localhost:8080/api/v1/oauth/tiktok/callback"),
			RateLimitCalls:  getEnvAsInt("TIKTOK_RATE_LIMIT_CALLS", 10),
			RateLimitWindow: getEnvAsDuration("TIKTOK_RATE_LIMIT_WINDOW", time.Second),
		},
		Shopee: ShopeeConfig{
			PartnerID:       getEnv("SHOPEE_PARTNER_ID", ""),
			PartnerKey:      getEnv("SHOPEE_PARTNER_KEY", ""),
			RedirectURI:     getEnv("SHOPEE_REDIRECT_URI", "http://localhost:8080/api/v1/oauth/shopee/callback"),
			RateLimitCalls:  getEnvAsInt("SHOPEE_RATE_LIMIT_CALLS", 1000),
			RateLimitWindow: getEnvAsDuration("SHOPEE_RATE_LIMIT_WINDOW", time.Minute),
		},
		Scheduler: SchedulerConfig{
			Enabled:                    getEnvAsBool("SCHEDULER_ENABLED", true),
			SyncCronSchedule:           getEnv("SYNC_CRON_SCHEDULE", "0 * * * *"),
			TokenRefreshCronSchedule:   getEnv("TOKEN_REFRESH_CRON_SCHEDULE", "*/30 * * * *"),
			MetricsAggregationSchedule: getEnv("METRICS_AGGREGATION_CRON_SCHEDULE", "0 */6 * * *"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		RateLimit: RateLimitConfig{
			Requests: getEnvAsInt("API_RATE_LIMIT_REQUESTS", 100),
			Window:   getEnvAsDuration("API_RATE_LIMIT_WINDOW", time.Minute),
		},
		HTTP: HTTPClientConfig{
			Timeout:      getEnvAsDuration("HTTP_CLIENT_TIMEOUT", 30*time.Second),
			MaxRetries:   getEnvAsInt("HTTP_CLIENT_MAX_RETRIES", 3),
			RetryWaitMin: getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MIN", time.Second),
			RetryWaitMax: getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MAX", 30*time.Second),
		},
	}

	// Validate required configurations
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "" && c.App.Env == "production" {
		return fmt.Errorf("JWT_SECRET is required in production")
	}

	if c.App.EncryptionKey == "" && c.App.Env == "production" {
		return fmt.Errorf("ENCRYPTION_KEY is required in production")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
