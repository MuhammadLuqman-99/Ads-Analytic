package shopee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Region Configuration
// ============================================================================

// Region represents a Shopee region/market
type Region string

const (
	RegionMalaysia    Region = "MY"
	RegionSingapore   Region = "SG"
	RegionThailand    Region = "TH"
	RegionVietnam     Region = "VN"
	RegionPhilippines Region = "PH"
	RegionIndonesia   Region = "ID"
	RegionTaiwan      Region = "TW"
	RegionBrazil      Region = "BR"
	RegionMexico      Region = "MX"
	RegionColombia    Region = "CO"
	RegionChile       Region = "CL"
	RegionPoland      Region = "PL"
	RegionSpain       Region = "ES"
	RegionFrance      Region = "FR"
	RegionIndia       Region = "IN"
)

// RegionBaseURL returns the base URL for a specific region
func RegionBaseURL(region Region) string {
	switch region {
	case RegionMalaysia:
		return "https://partner.shopeemobile.com"
	case RegionSingapore:
		return "https://partner.shopeemobile.com"
	case RegionThailand:
		return "https://partner.shopeemobile.com"
	case RegionVietnam:
		return "https://partner.shopeemobile.com"
	case RegionPhilippines:
		return "https://partner.shopeemobile.com"
	case RegionIndonesia:
		return "https://partner.shopeemobile.com"
	case RegionTaiwan:
		return "https://partner.shopeemobile.com"
	case RegionBrazil:
		return "https://openapi.shopee.com.br"
	case RegionMexico:
		return "https://openapi.shopee.com.mx"
	case RegionColombia:
		return "https://openapi.shopee.com.co"
	case RegionChile:
		return "https://openapi.shopee.cl"
	case RegionPoland:
		return "https://openapi.shopee.pl"
	case RegionSpain:
		return "https://openapi.shopee.es"
	case RegionFrance:
		return "https://openapi.shopee.fr"
	case RegionIndia:
		return "https://openapi.shopee.in"
	default:
		return "https://partner.shopeemobile.com"
	}
}

// RegionCurrency returns the default currency for a region
func RegionCurrency(region Region) string {
	switch region {
	case RegionMalaysia:
		return "MYR"
	case RegionSingapore:
		return "SGD"
	case RegionThailand:
		return "THB"
	case RegionVietnam:
		return "VND"
	case RegionPhilippines:
		return "PHP"
	case RegionIndonesia:
		return "IDR"
	case RegionTaiwan:
		return "TWD"
	case RegionBrazil:
		return "BRL"
	case RegionMexico:
		return "MXN"
	case RegionColombia:
		return "COP"
	case RegionChile:
		return "CLP"
	case RegionPoland:
		return "PLN"
	case RegionSpain, RegionFrance:
		return "EUR"
	case RegionIndia:
		return "INR"
	default:
		return "MYR"
	}
}

// ============================================================================
// Constants
// ============================================================================

const (
	// API paths
	authPath    = "/api/v2/shop/auth_partner"
	tokenPath   = "/api/v2/auth/token/get"
	refreshPath = "/api/v2/auth/access_token/get"

	// Ads API paths
	getAllAdsPath         = "/api/v2/ads/get_all_ads"
	getDailyPerformance   = "/api/v2/ads/get_daily_performance"
	getAdsPerformancePath = "/api/v2/ads/get_ads_daily_performance"

	// Order API paths
	getOrderListPath   = "/api/v2/order/get_order_list"
	getOrderDetailPath = "/api/v2/order/get_order_detail"

	// Token expiry (Shopee tokens expire in 4 hours)
	TokenExpiryDuration = 4 * time.Hour

	// Default pagination
	DefaultPageSize = 100
)

// ============================================================================
// Error Codes
// ============================================================================

const (
	// Common Shopee error codes
	ShopeeErrorInvalidPartner       = "error_invalid_partner"
	ShopeeErrorInvalidSign          = "error_invalid_sign"
	ShopeeErrorInvalidTimestamp     = "error_invalid_timestamp"
	ShopeeErrorInvalidAccessToken   = "error_invalid_access_token"
	ShopeeErrorAccessTokenExpired   = "error_access_token_expired"
	ShopeeErrorRefreshTokenExpired  = "error_refresh_token_expired"
	ShopeeErrorRateLimit            = "error_rate_limit"
	ShopeeErrorPermissionDenied     = "error_permission_denied"
	ShopeeErrorShopNotAuthorized    = "error_shop_not_authorized"
	ShopeeErrorInvalidShopID        = "error_invalid_shop_id"
	ShopeeErrorBannedShop           = "error_banned_shop"
	ShopeeErrorServerError          = "error_server"
	ShopeeErrorServiceUnavailable   = "error_service_unavailable"
	ShopeeErrorInvalidParam         = "error_invalid_param"
	ShopeeErrorInvalidCampaignState = "error_invalid_campaign_state"
)

// ============================================================================
// Ad Types
// ============================================================================

const (
	AdTypeKeywordAds    = "keyword_ads"
	AdTypeDiscoveryAds  = "discovery_ads"
	AdTypeShopAds       = "shop_ads"
	AdTypeTargetingAds  = "targeting_ads"
	AdTypeProductAds    = "product_ads"
	AdTypeCollectionAds = "collection_ads"
)

// ============================================================================
// Ad States
// ============================================================================

const (
	AdStateOngoing   = "ongoing"
	AdStatePaused    = "paused"
	AdStateEnded     = "ended"
	AdStateScheduled = "scheduled"
	AdStateSuspended = "suspended"
	AdStateDeleted   = "deleted"
)

// Connector implements the PlatformConnector interface for Shopee Ads
type Connector struct {
	*platform.BaseConnector
	config  *Config
	baseURL string
}

// Config holds Shopee-specific configuration
type Config struct {
	PartnerID       int64
	PartnerKey      string
	RedirectURI     string
	Region          Region // MY, SG, TH, VN, PH, ID, TW, BR, MX, CO, CL, PL, ES, FR, IN
	ShopID          int64  // Default shop ID (can be overridden per request)
	RateLimitCalls  int
	RateLimitWindow time.Duration
	Timeout         time.Duration
	MaxRetries      int
}

// DefaultConfig returns default Shopee connector configuration
func DefaultConfig() *Config {
	return &Config{
		Region:          RegionMalaysia,
		RateLimitCalls:  1000,
		RateLimitWindow: time.Minute,
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}
}

// NewConnector creates a new Shopee Ads connector
func NewConnector(config *Config) *Connector {
	if config == nil {
		config = DefaultConfig()
	}

	regionURL := RegionBaseURL(config.Region)

	baseConfig := &platform.ConnectorConfig{
		AppID:           fmt.Sprintf("%d", config.PartnerID),
		AppSecret:       config.PartnerKey,
		RedirectURI:     config.RedirectURI,
		BaseURL:         regionURL,
		RateLimitCalls:  config.RateLimitCalls,
		RateLimitWindow: config.RateLimitWindow,
		Timeout:         config.Timeout,
		MaxRetries:      config.MaxRetries,
	}

	return &Connector{
		BaseConnector: platform.NewBaseConnector(entity.PlatformShopee, baseConfig),
		config:        config,
		baseURL:       regionURL,
	}
}

// NewConnectorWithRegion creates a new Shopee connector for a specific region
func NewConnectorWithRegion(partnerID int64, partnerKey string, region Region) *Connector {
	config := &Config{
		PartnerID:       partnerID,
		PartnerKey:      partnerKey,
		Region:          region,
		RateLimitCalls:  1000,
		RateLimitWindow: time.Minute,
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}
	return NewConnector(config)
}

// GetRegion returns the configured region
func (c *Connector) GetRegion() Region {
	return c.config.Region
}

// GetCurrency returns the default currency for the configured region
func (c *Connector) GetCurrency() string {
	return RegionCurrency(c.config.Region)
}

// ============================================================================
// OAuth Methods
// ============================================================================

// GetAuthURL generates the OAuth authorization URL
func (c *Connector) GetAuthURL(state string) string {
	timestamp := time.Now().Unix()
	path := authPath

	// Generate signature
	baseString := fmt.Sprintf("%d%s%d", c.config.PartnerID, path, timestamp)
	sign := c.generateSign(baseString)

	params := url.Values{
		"partner_id": {fmt.Sprintf("%d", c.config.PartnerID)},
		"timestamp":  {fmt.Sprintf("%d", timestamp)},
		"sign":       {sign},
		"redirect":   {c.config.RedirectURI},
	}

	return c.baseURL + authPath + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for OAuth tokens
func (c *Connector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	timestamp := time.Now().Unix()
	path := tokenPath

	// Generate signature
	baseString := fmt.Sprintf("%d%s%d", c.config.PartnerID, path, timestamp)
	sign := c.generateSign(baseString)

	endpoint := fmt.Sprintf("%s%s", c.baseURL, tokenPath)

	body := map[string]interface{}{
		"code":       code,
		"partner_id": c.config.PartnerID,
		"timestamp":  timestamp,
		"sign":       sign,
	}

	resp, err := c.DoPost(ctx, endpoint, nil, body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		Error        string  `json:"error"`
		Message      string  `json:"message"`
		AccessToken  string  `json:"access_token"`
		RefreshToken string  `json:"refresh_token"`
		ExpireIn     int     `json:"expire_in"`
		ShopIDList   []int64 `json:"shop_id_list"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, errors.NewOAuthError(
			entity.PlatformShopee.String(),
			tokenResp.Error,
			tokenResp.Message,
		)
	}

	return &entity.OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpireIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpireIn) * time.Second),
	}, nil
}

// RefreshToken refreshes an expired access token
func (c *Connector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	timestamp := time.Now().Unix()
	path := refreshPath

	// Generate signature
	baseString := fmt.Sprintf("%d%s%d", c.config.PartnerID, path, timestamp)
	sign := c.generateSign(baseString)

	endpoint := fmt.Sprintf("%s%s", c.baseURL, refreshPath)

	body := map[string]interface{}{
		"refresh_token": refreshToken,
		"partner_id":    c.config.PartnerID,
		"timestamp":     timestamp,
		"sign":          sign,
	}

	resp, err := c.DoPost(ctx, endpoint, nil, body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		Error        string `json:"error"`
		Message      string `json:"message"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpireIn     int    `json:"expire_in"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, errors.NewOAuthError(
			entity.PlatformShopee.String(),
			tokenResp.Error,
			tokenResp.Message,
		)
	}

	return &entity.OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpireIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.ExpireIn) * time.Second),
	}, nil
}

// RevokeToken revokes an access token
func (c *Connector) RevokeToken(ctx context.Context, accessToken string) error {
	// Shopee doesn't have a token revocation endpoint
	return nil
}

// ============================================================================
// User & Account Methods
// ============================================================================

// GetUserInfo retrieves the authenticated user's information
func (c *Connector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	// Shopee doesn't have a user info endpoint in the same way
	// Return a placeholder
	return &entity.PlatformUser{
		ID:   "shopee_user",
		Name: "Shopee Seller",
	}, nil
}

// GetAdAccounts retrieves all ad accounts (shops) accessible by the token
func (c *Connector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	// For Shopee, ad accounts are essentially shops
	// This would typically come from the token exchange response
	return []entity.PlatformAccount{}, nil
}

// ============================================================================
// Campaign Methods
// ============================================================================

// GetCampaigns retrieves all campaigns for a shop
func (c *Connector) GetCampaigns(ctx context.Context, accessToken string, shopID string) ([]entity.Campaign, error) {
	timestamp := time.Now().Unix()
	path := "/api/v2/ads/get_all_ads"

	shopIDInt, _ := strconv.ParseInt(shopID, 10, 64)

	// Generate signature
	baseString := fmt.Sprintf("%d%s%d%s%d", c.config.PartnerID, path, timestamp, accessToken, shopIDInt)
	sign := c.generateSign(baseString)

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	params := map[string]string{
		"partner_id":   fmt.Sprintf("%d", c.config.PartnerID),
		"timestamp":    fmt.Sprintf("%d", timestamp),
		"access_token": accessToken,
		"shop_id":      shopID,
		"sign":         sign,
	}

	resp, err := c.DoGet(ctx, endpoint, nil, params)
	if err != nil {
		return nil, err
	}

	var adsResp struct {
		Error    string `json:"error"`
		Message  string `json:"message"`
		Response struct {
			Ads []struct {
				CampaignID uint64 `json:"campaign_id"`
				Title      string `json:"title"`
				State      string `json:"state"`
				Type       string `json:"type"`
				Budget     struct {
					DailyBudget float64 `json:"daily_budget"`
					TotalBudget float64 `json:"total_budget"`
				} `json:"budget"`
				CreateTime int64 `json:"create_time"`
				UpdateTime int64 `json:"update_time"`
			} `json:"ads"`
		} `json:"response"`
	}

	if err := c.ParseJSON(resp.Body, &adsResp); err != nil {
		return nil, err
	}

	if adsResp.Error != "" {
		return nil, errors.NewPlatformAPIError(
			entity.PlatformShopee.String(),
			0,
			adsResp.Error,
			adsResp.Message,
		)
	}

	var campaigns []entity.Campaign
	for _, ad := range adsResp.Response.Ads {
		campaign := entity.Campaign{
			Platform:             entity.PlatformShopee,
			PlatformCampaignID:   fmt.Sprintf("%d", ad.CampaignID),
			PlatformCampaignName: ad.Title,
			Status:               c.mapCampaignStatus(ad.State),
			Objective:            c.mapObjective(ad.Type),
		}

		if ad.Budget.DailyBudget > 0 {
			budget := decimal.NewFromFloat(ad.Budget.DailyBudget)
			campaign.DailyBudget = &budget
		}

		if ad.Budget.TotalBudget > 0 {
			budget := decimal.NewFromFloat(ad.Budget.TotalBudget)
			campaign.LifetimeBudget = &budget
		}

		if ad.CreateTime > 0 {
			t := time.Unix(ad.CreateTime, 0)
			campaign.PlatformCreatedAt = &t
		}

		if ad.UpdateTime > 0 {
			t := time.Unix(ad.UpdateTime, 0)
			campaign.PlatformUpdatedAt = &t
		}

		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

// GetCampaign retrieves a single campaign by ID
func (c *Connector) GetCampaign(ctx context.Context, accessToken string, campaignID string) (*entity.Campaign, error) {
	// Shopee doesn't have a single campaign endpoint
	// Would need to filter from GetCampaigns
	return nil, errors.ErrNotFound("Campaign")
}

// ============================================================================
// AdSet Methods (Shopee doesn't have AdSets, uses flat structure)
// ============================================================================

// GetAdSets returns empty as Shopee uses flat ad structure
func (c *Connector) GetAdSets(ctx context.Context, accessToken string, campaignID string) ([]entity.AdSet, error) {
	// Shopee uses a flat structure without ad sets
	return []entity.AdSet{}, nil
}

// GetAdSet returns nil as Shopee doesn't have ad sets
func (c *Connector) GetAdSet(ctx context.Context, accessToken string, adSetID string) (*entity.AdSet, error) {
	return nil, errors.ErrNotFound("AdSet")
}

// ============================================================================
// Ad Methods
// ============================================================================

// GetAds retrieves ads for a campaign
func (c *Connector) GetAds(ctx context.Context, accessToken string, adSetID string) ([]entity.Ad, error) {
	// Shopee ads are at campaign level
	return []entity.Ad{}, nil
}

// GetAd retrieves a single ad by ID
func (c *Connector) GetAd(ctx context.Context, accessToken string, adID string) (*entity.Ad, error) {
	return nil, errors.ErrNotFound("Ad")
}

// ============================================================================
// Insights Methods
// ============================================================================

// GetCampaignInsights retrieves performance metrics for a campaign
func (c *Connector) GetCampaignInsights(ctx context.Context, accessToken string, campaignID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	timestamp := time.Now().Unix()
	path := "/api/v2/ads/get_daily_performance"

	// Note: This is a simplified implementation
	// Real implementation would need proper shop_id handling

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	params := map[string]string{
		"partner_id":   fmt.Sprintf("%d", c.config.PartnerID),
		"timestamp":    fmt.Sprintf("%d", timestamp),
		"access_token": accessToken,
		"campaign_id":  campaignID,
		"start_date":   dateRange.StartDate.Format("2006-01-02"),
		"end_date":     dateRange.EndDate.Format("2006-01-02"),
	}

	resp, err := c.DoGet(ctx, endpoint, nil, params)
	if err != nil {
		return nil, err
	}

	var perfResp struct {
		Error    string `json:"error"`
		Message  string `json:"message"`
		Response struct {
			DailyData []struct {
				Date        string  `json:"date"`
				Impressions int64   `json:"impression"`
				Clicks      int64   `json:"click"`
				Cost        float64 `json:"cost"`
				Orders      int64   `json:"direct_order"`
				GMV         float64 `json:"direct_gmv"`
				ROAS        float64 `json:"direct_roas"`
			} `json:"daily_data"`
		} `json:"response"`
	}

	if err := c.ParseJSON(resp.Body, &perfResp); err != nil {
		return nil, err
	}

	var metrics []entity.CampaignMetricsDaily
	for _, day := range perfResp.Response.DailyData {
		metricDate, _ := time.Parse("2006-01-02", day.Date)

		metric := entity.CampaignMetricsDaily{
			BaseEntity:    entity.BaseEntity{ID: uuid.New()},
			Platform:      entity.PlatformShopee,
			MetricDate:    metricDate,
			Currency:      "MYR", // Default to MYR, should be from account
			Impressions:   day.Impressions,
			Clicks:        day.Clicks,
			Spend:         decimal.NewFromFloat(day.Cost),
			Purchases:     day.Orders,
			PurchaseValue: decimal.NewFromFloat(day.GMV),
		}

		// Calculate derived metrics
		if day.Impressions > 0 {
			ctr := float64(day.Clicks) / float64(day.Impressions) * 100
			metric.CTR = &ctr
		}

		if day.Clicks > 0 {
			cpc := decimal.NewFromFloat(day.Cost).Div(decimal.NewFromInt(day.Clicks))
			metric.CPC = &cpc
		}

		if day.Impressions > 0 {
			cpm := decimal.NewFromFloat(day.Cost).Div(decimal.NewFromInt(day.Impressions)).Mul(decimal.NewFromInt(1000))
			metric.CPM = &cpm
		}

		roas := day.ROAS
		metric.ROAS = &roas

		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAdSetInsights returns nil as Shopee doesn't have ad sets
func (c *Connector) GetAdSetInsights(ctx context.Context, accessToken string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	return nil, nil
}

// GetAdInsights returns nil as Shopee uses campaign-level metrics
func (c *Connector) GetAdInsights(ctx context.Context, accessToken string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	return nil, nil
}

// GetAccountInsights retrieves aggregated insights for a shop
func (c *Connector) GetAccountInsights(ctx context.Context, accessToken string, shopID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	// Aggregate from campaign insights
	return nil, nil
}

// HealthCheck verifies the connector can connect to Shopee API
func (c *Connector) HealthCheck(ctx context.Context) error {
	// Shopee health check would require valid credentials
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// generateSign generates HMAC-SHA256 signature for Shopee API
func (c *Connector) generateSign(baseString string) string {
	h := hmac.New(sha256.New, []byte(c.config.PartnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// SignRequest generates a complete signed request for Shopee API
// Signature format: partner_id + path + timestamp + access_token + shop_id
func (c *Connector) SignRequest(path string, accessToken string, shopID int64) (timestamp int64, sign string) {
	timestamp = time.Now().Unix()
	baseString := fmt.Sprintf("%d%s%d%s%d", c.config.PartnerID, path, timestamp, accessToken, shopID)
	sign = c.generateSign(baseString)
	return
}

// SignRequestWithParams generates signed request parameters
func (c *Connector) SignRequestWithParams(path string, accessToken string, shopID int64) map[string]string {
	timestamp, sign := c.SignRequest(path, accessToken, shopID)
	return map[string]string{
		"partner_id":   fmt.Sprintf("%d", c.config.PartnerID),
		"timestamp":    fmt.Sprintf("%d", timestamp),
		"access_token": accessToken,
		"shop_id":      fmt.Sprintf("%d", shopID),
		"sign":         sign,
	}
}

func (c *Connector) mapCampaignStatus(state string) entity.CampaignStatus {
	switch strings.ToLower(state) {
	case "ongoing", "active":
		return entity.CampaignStatusActive
	case "paused", "suspended":
		return entity.CampaignStatusPaused
	case "ended", "expired":
		return entity.CampaignStatusArchived
	case "deleted":
		return entity.CampaignStatusDeleted
	default:
		return entity.CampaignStatusDraft
	}
}

func (c *Connector) mapObjective(adType string) entity.CampaignObjective {
	switch strings.ToLower(adType) {
	case "discovery_ads":
		return entity.ObjectiveAwareness
	case "keyword_ads", "shop_ads":
		return entity.ObjectiveTraffic
	case "targeting_ads":
		return entity.ObjectiveSales
	default:
		return entity.ObjectiveSales
	}
}

// ============================================================================
// Error Handling
// ============================================================================

// HandleShopeeError converts Shopee error codes to appropriate error types
func (c *Connector) HandleShopeeError(errorCode string, message string, statusCode int) error {
	switch errorCode {
	case ShopeeErrorAccessTokenExpired:
		return errors.NewTokenExpiredError(entity.PlatformShopee.String(), "access", time.Now())
	case ShopeeErrorRefreshTokenExpired:
		return errors.NewTokenExpiredError(entity.PlatformShopee.String(), "refresh", time.Now())
	case ShopeeErrorInvalidAccessToken:
		return errors.NewTokenInvalidError(entity.PlatformShopee.String(), "access")
	case ShopeeErrorRateLimit:
		return errors.NewRateLimitError(entity.PlatformShopee.String(), 60*time.Second)
	case ShopeeErrorInvalidSign:
		return errors.NewPlatformAPIError(entity.PlatformShopee.String(), http.StatusUnauthorized, errorCode, "Invalid signature - check partner_key and signing order")
	case ShopeeErrorInvalidTimestamp:
		return errors.NewPlatformAPIError(entity.PlatformShopee.String(), http.StatusBadRequest, errorCode, "Invalid timestamp - ensure server time is synchronized")
	case ShopeeErrorShopNotAuthorized:
		return errors.NewPlatformAPIError(entity.PlatformShopee.String(), http.StatusForbidden, errorCode, "Shop not authorized - re-authorization required")
	case ShopeeErrorBannedShop:
		return errors.NewPlatformAPIError(entity.PlatformShopee.String(), http.StatusForbidden, errorCode, "Shop is banned")
	case ShopeeErrorServerError, ShopeeErrorServiceUnavailable:
		platformErr := errors.NewPlatformAPIError(entity.PlatformShopee.String(), http.StatusServiceUnavailable, errorCode, message)
		platformErr.WithMetadata("is_transient", "true")
		return platformErr
	default:
		return errors.NewPlatformAPIError(entity.PlatformShopee.String(), statusCode, errorCode, message)
	}
}

// IsRetryableError checks if a Shopee error is retryable
func (c *Connector) IsRetryableError(errorCode string) bool {
	switch errorCode {
	case ShopeeErrorRateLimit, ShopeeErrorServerError, ShopeeErrorServiceUnavailable:
		return true
	default:
		return false
	}
}

// ============================================================================
// Token Management
// ============================================================================

// TokenInfo holds token information with expiry tracking
type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	ShopID       int64
	mu           sync.RWMutex
}

// IsExpired checks if the token is expired or about to expire (within 5 minutes)
func (t *TokenInfo) IsExpired() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}

// NeedsRefresh checks if the token needs refresh (within 30 minutes of expiry)
func (t *TokenInfo) NeedsRefresh() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Now().Add(30 * time.Minute).After(t.ExpiresAt)
}

// Update updates the token info thread-safely
func (t *TokenInfo) Update(accessToken, refreshToken string, expiresIn int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AccessToken = accessToken
	t.RefreshToken = refreshToken
	t.ExpiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
}

// RefreshTokenIfNeeded refreshes the token if it's about to expire
func (c *Connector) RefreshTokenIfNeeded(ctx context.Context, tokenInfo *TokenInfo) error {
	if !tokenInfo.NeedsRefresh() {
		return nil
	}

	newToken, err := c.RefreshToken(ctx, tokenInfo.RefreshToken)
	if err != nil {
		return err
	}

	tokenInfo.Update(newToken.AccessToken, newToken.RefreshToken, newToken.ExpiresIn)
	return nil
}

// ============================================================================
// Enhanced Ads API Methods
// ============================================================================

// GetAllAdsParams holds parameters for GetAllAds request
type GetAllAdsParams struct {
	ShopID     int64
	AdType     string   // keyword_ads, discovery_ads, shop_ads, targeting_ads
	AdState    string   // ongoing, paused, ended
	PageSize   int      // Default 100, max 100
	PageNumber int      // 0-indexed
	CampaignID []uint64 // Optional: filter by campaign IDs
}

// ShopeeAd represents a Shopee ad from the API
type ShopeeAd struct {
	CampaignID  uint64  `json:"campaign_id"`
	Title       string  `json:"title"`
	State       string  `json:"state"`
	Type        string  `json:"type"`
	DailyBudget float64 `json:"daily_budget"`
	TotalBudget float64 `json:"total_budget"`
	StartTime   int64   `json:"start_time"`
	EndTime     int64   `json:"end_time"`
	CreateTime  int64   `json:"create_time"`
	UpdateTime  int64   `json:"update_time"`
	// Product-level ads
	ItemID int64 `json:"item_id,omitempty"`
	// Keyword ads specific
	Keywords []struct {
		Keyword     string  `json:"keyword"`
		MatchType   string  `json:"match_type"` // broad, exact
		BidPrice    float64 `json:"bid_price"`
		QualityScore int    `json:"quality_score"`
	} `json:"keywords,omitempty"`
}

// GetAllAdsResponse represents the response from get_all_ads API
type GetAllAdsResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		Ads      []ShopeeAd `json:"ads"`
		More     bool       `json:"more"`
		Total    int        `json:"total"`
		PageInfo struct {
			PageSize   int `json:"page_size"`
			PageNumber int `json:"page_number"`
			TotalCount int `json:"total_count"`
		} `json:"page_info"`
	} `json:"response"`
}

// GetAllAdsComplete retrieves all ads with pagination support
func (c *Connector) GetAllAdsComplete(ctx context.Context, accessToken string, params GetAllAdsParams) ([]ShopeeAd, error) {
	var allAds []ShopeeAd

	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = DefaultPageSize
	}

	for {
		ads, hasMore, err := c.fetchAdsPage(ctx, accessToken, params)
		if err != nil {
			return allAds, err
		}

		allAds = append(allAds, ads...)

		if !hasMore {
			break
		}

		params.PageNumber++

		// Check context cancellation
		select {
		case <-ctx.Done():
			return allAds, ctx.Err()
		default:
		}
	}

	return allAds, nil
}

func (c *Connector) fetchAdsPage(ctx context.Context, accessToken string, params GetAllAdsParams) ([]ShopeeAd, bool, error) {
	path := getAllAdsPath
	signParams := c.SignRequestWithParams(path, accessToken, params.ShopID)

	// Add filter parameters
	if params.AdType != "" {
		signParams["type"] = params.AdType
	}
	if params.AdState != "" {
		signParams["state"] = params.AdState
	}
	signParams["page_size"] = fmt.Sprintf("%d", params.PageSize)
	signParams["page_number"] = fmt.Sprintf("%d", params.PageNumber)

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	resp, err := c.DoGet(ctx, endpoint, nil, signParams)
	if err != nil {
		return nil, false, err
	}

	var adsResp GetAllAdsResponse
	if err := c.ParseJSON(resp.Body, &adsResp); err != nil {
		return nil, false, err
	}

	if adsResp.Error != "" {
		return nil, false, c.HandleShopeeError(adsResp.Error, adsResp.Message, 0)
	}

	return adsResp.Response.Ads, adsResp.Response.More, nil
}

// ============================================================================
// Enhanced Performance Metrics Methods
// ============================================================================

// ShopeePerformanceMetrics represents detailed Shopee ad performance
type ShopeePerformanceMetrics struct {
	Date string `json:"date"`

	// Impressions & Clicks
	Impression int64 `json:"impression"`
	Click      int64 `json:"click"`

	// Cost breakdown (Shopee-specific: broad_match + exact_match)
	BroadMatchCost float64 `json:"broad_match_cost"`
	ExactMatchCost float64 `json:"exact_match_cost"`
	Cost           float64 `json:"cost"` // Total cost if provided directly

	// Conversion metrics
	DirectOrder   int64   `json:"direct_order"`
	DirectGMV     float64 `json:"direct_gmv"`
	DirectROAS    float64 `json:"direct_roas"`
	IndirectOrder int64   `json:"indirect_order"`
	IndirectGMV   float64 `json:"indirect_gmv"`

	// Engagement
	ProductClick  int64 `json:"product_click"`
	AddToCart     int64 `json:"add_to_cart"`
	CheckoutStart int64 `json:"checkout_start"`

	// Keyword ads specific
	AverageRank float64 `json:"average_rank"`

	// Campaign/Ad identifiers
	CampaignID uint64 `json:"campaign_id,omitempty"`
	ItemID     int64  `json:"item_id,omitempty"`
}

// GetTotalSpend returns total spend (broad_match_cost + exact_match_cost or cost)
func (m *ShopeePerformanceMetrics) GetTotalSpend() float64 {
	// If individual costs are available, sum them
	if m.BroadMatchCost > 0 || m.ExactMatchCost > 0 {
		return m.BroadMatchCost + m.ExactMatchCost
	}
	// Otherwise use the total cost field
	return m.Cost
}

// GetTotalOrders returns total orders (direct + indirect)
func (m *ShopeePerformanceMetrics) GetTotalOrders() int64 {
	return m.DirectOrder + m.IndirectOrder
}

// GetTotalGMV returns total GMV (direct + indirect)
func (m *ShopeePerformanceMetrics) GetTotalGMV() float64 {
	return m.DirectGMV + m.IndirectGMV
}

// DailyPerformanceParams holds parameters for daily performance request
type DailyPerformanceParams struct {
	ShopID     int64
	CampaignID uint64
	StartDate  time.Time
	EndDate    time.Time
	AdType     string // Optional: filter by ad type
}

// GetAdsDailyPerformanceComplete retrieves detailed daily performance metrics
func (c *Connector) GetAdsDailyPerformanceComplete(ctx context.Context, accessToken string, params DailyPerformanceParams) ([]ShopeePerformanceMetrics, error) {
	path := getAdsPerformancePath
	signParams := c.SignRequestWithParams(path, accessToken, params.ShopID)

	signParams["start_date"] = params.StartDate.Format("2006-01-02")
	signParams["end_date"] = params.EndDate.Format("2006-01-02")

	if params.CampaignID > 0 {
		signParams["campaign_id"] = fmt.Sprintf("%d", params.CampaignID)
	}
	if params.AdType != "" {
		signParams["type"] = params.AdType
	}

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	resp, err := c.DoGet(ctx, endpoint, nil, signParams)
	if err != nil {
		return nil, err
	}

	var perfResp struct {
		Error    string `json:"error"`
		Message  string `json:"message"`
		Response struct {
			DailyData []ShopeePerformanceMetrics `json:"daily_data"`
		} `json:"response"`
	}

	if err := c.ParseJSON(resp.Body, &perfResp); err != nil {
		return nil, err
	}

	if perfResp.Error != "" {
		return nil, c.HandleShopeeError(perfResp.Error, perfResp.Message, 0)
	}

	return perfResp.Response.DailyData, nil
}

// MapToNormalizedMetrics converts Shopee metrics to normalized CampaignMetricsDaily
func (c *Connector) MapToNormalizedMetrics(shopeeMetrics []ShopeePerformanceMetrics, campaignID uuid.UUID, orgID uuid.UUID) []entity.CampaignMetricsDaily {
	var metrics []entity.CampaignMetricsDaily

	for _, sm := range shopeeMetrics {
		metricDate, _ := time.Parse("2006-01-02", sm.Date)

		// Calculate total spend from broad + exact match costs
		totalSpend := sm.GetTotalSpend()

		metric := entity.CampaignMetricsDaily{
			BaseEntity:     entity.NewBaseEntity(),
			CampaignID:     campaignID,
			OrganizationID: orgID,
			Platform:       entity.PlatformShopee,
			MetricDate:     metricDate,
			Currency:       c.GetCurrency(),
			Impressions:    sm.Impression,
			Clicks:         sm.Click,
			Spend:          decimal.NewFromFloat(totalSpend),
			Purchases:      sm.DirectOrder,
			PurchaseValue:  decimal.NewFromFloat(sm.DirectGMV),
			AddToCart:      sm.AddToCart,
			Conversions:    sm.GetTotalOrders(),
			ConversionValue: decimal.NewFromFloat(sm.GetTotalGMV()),
			PlatformMetrics: entity.JSONMap{
				"broad_match_cost":  sm.BroadMatchCost,
				"exact_match_cost":  sm.ExactMatchCost,
				"direct_order":      sm.DirectOrder,
				"direct_gmv":        sm.DirectGMV,
				"indirect_order":    sm.IndirectOrder,
				"indirect_gmv":      sm.IndirectGMV,
				"product_click":     sm.ProductClick,
				"checkout_start":    sm.CheckoutStart,
				"average_rank":      sm.AverageRank,
			},
		}

		// Calculate derived metrics with zero-division protection
		if sm.Impression > 0 {
			ctr := float64(sm.Click) / float64(sm.Impression) * 100
			metric.CTR = &ctr
			cpm := decimal.NewFromFloat(totalSpend).Div(decimal.NewFromInt(sm.Impression)).Mul(decimal.NewFromInt(1000))
			metric.CPM = &cpm
		}

		if sm.Click > 0 {
			cpc := decimal.NewFromFloat(totalSpend).Div(decimal.NewFromInt(sm.Click))
			metric.CPC = &cpc
		}

		totalOrders := sm.GetTotalOrders()
		if totalOrders > 0 {
			cpa := decimal.NewFromFloat(totalSpend).Div(decimal.NewFromInt(totalOrders))
			metric.CPA = &cpa
		}

		if totalSpend > 0 {
			roas := sm.GetTotalGMV() / totalSpend
			metric.ROAS = &roas
		}

		now := time.Now()
		metric.LastSyncedAt = &now

		metrics = append(metrics, metric)
	}

	return metrics
}

// ============================================================================
// Order API Methods (for Revenue Calculation)
// ============================================================================

// OrderStatus represents Shopee order status
type OrderStatus string

const (
	OrderStatusUnpaid      OrderStatus = "UNPAID"
	OrderStatusReady       OrderStatus = "READY_TO_SHIP"
	OrderStatusProcessed   OrderStatus = "PROCESSED"
	OrderStatusShipped     OrderStatus = "SHIPPED"
	OrderStatusCompleted   OrderStatus = "COMPLETED"
	OrderStatusInCancel    OrderStatus = "IN_CANCEL"
	OrderStatusCancelled   OrderStatus = "CANCELLED"
	OrderStatusInvoicePend OrderStatus = "INVOICE_PENDING"
)

// ShopeeOrder represents a Shopee order
type ShopeeOrder struct {
	OrderSN        string      `json:"order_sn"`
	OrderStatus    OrderStatus `json:"order_status"`
	TotalAmount    float64     `json:"total_amount"`
	Currency       string      `json:"currency"`
	CreateTime     int64       `json:"create_time"`
	UpdateTime     int64       `json:"update_time"`
	PayTime        int64       `json:"pay_time,omitempty"`
	ShipByDate     int64       `json:"ship_by_date,omitempty"`
	BuyerUserID    int64       `json:"buyer_user_id"`
	MessageToSeller string     `json:"message_to_seller,omitempty"`
	ItemList       []struct {
		ItemID         int64   `json:"item_id"`
		ItemName       string  `json:"item_name"`
		ItemSKU        string  `json:"item_sku"`
		ModelID        int64   `json:"model_id"`
		ModelName      string  `json:"model_name"`
		ModelSKU       string  `json:"model_sku"`
		ModelQuantity  int     `json:"model_quantity_purchased"`
		ModelPrice     float64 `json:"model_discounted_price"`
		ModelOrigPrice float64 `json:"model_original_price"`
	} `json:"item_list"`
	// Ad attribution (if available)
	AdCampaignID uint64 `json:"ad_campaign_id,omitempty"`
}

// GetOrderListParams holds parameters for order list request
type GetOrderListParams struct {
	ShopID      int64
	TimeRangeField string // create_time, update_time
	TimeFrom    int64
	TimeTo      int64
	PageSize    int
	Cursor      string
	OrderStatus OrderStatus // Optional filter
}

// GetOrderListResponse represents the order list API response
type GetOrderListResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		More       bool   `json:"more"`
		NextCursor string `json:"next_cursor"`
		OrderList  []struct {
			OrderSN     string `json:"order_sn"`
			OrderStatus string `json:"order_status"`
		} `json:"order_list"`
	} `json:"response"`
}

// GetOrderList retrieves orders within a time range
func (c *Connector) GetOrderList(ctx context.Context, accessToken string, params GetOrderListParams) ([]string, string, bool, error) {
	path := getOrderListPath
	signParams := c.SignRequestWithParams(path, accessToken, params.ShopID)

	if params.TimeRangeField == "" {
		params.TimeRangeField = "create_time"
	}
	signParams["time_range_field"] = params.TimeRangeField
	signParams["time_from"] = fmt.Sprintf("%d", params.TimeFrom)
	signParams["time_to"] = fmt.Sprintf("%d", params.TimeTo)

	if params.PageSize <= 0 {
		params.PageSize = 100
	}
	signParams["page_size"] = fmt.Sprintf("%d", params.PageSize)

	if params.Cursor != "" {
		signParams["cursor"] = params.Cursor
	}

	if params.OrderStatus != "" {
		signParams["order_status"] = string(params.OrderStatus)
	}

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	resp, err := c.DoGet(ctx, endpoint, nil, signParams)
	if err != nil {
		return nil, "", false, err
	}

	var orderResp GetOrderListResponse
	if err := c.ParseJSON(resp.Body, &orderResp); err != nil {
		return nil, "", false, err
	}

	if orderResp.Error != "" {
		return nil, "", false, c.HandleShopeeError(orderResp.Error, orderResp.Message, 0)
	}

	orderSNs := make([]string, len(orderResp.Response.OrderList))
	for i, o := range orderResp.Response.OrderList {
		orderSNs[i] = o.OrderSN
	}

	return orderSNs, orderResp.Response.NextCursor, orderResp.Response.More, nil
}

// GetAllOrders retrieves all orders within a time range with pagination
func (c *Connector) GetAllOrders(ctx context.Context, accessToken string, shopID int64, startTime, endTime time.Time) ([]string, error) {
	var allOrderSNs []string
	cursor := ""

	params := GetOrderListParams{
		ShopID:         shopID,
		TimeRangeField: "create_time",
		TimeFrom:       startTime.Unix(),
		TimeTo:         endTime.Unix(),
		PageSize:       100,
	}

	for {
		params.Cursor = cursor

		orderSNs, nextCursor, hasMore, err := c.GetOrderList(ctx, accessToken, params)
		if err != nil {
			return allOrderSNs, err
		}

		allOrderSNs = append(allOrderSNs, orderSNs...)

		if !hasMore || nextCursor == "" {
			break
		}
		cursor = nextCursor

		select {
		case <-ctx.Done():
			return allOrderSNs, ctx.Err()
		default:
		}
	}

	return allOrderSNs, nil
}

// GetOrderDetails retrieves detailed information for specific orders
func (c *Connector) GetOrderDetails(ctx context.Context, accessToken string, shopID int64, orderSNs []string) ([]ShopeeOrder, error) {
	if len(orderSNs) == 0 {
		return nil, nil
	}

	// Shopee allows max 50 orders per request
	const batchSize = 50
	var allOrders []ShopeeOrder

	for i := 0; i < len(orderSNs); i += batchSize {
		end := i + batchSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batch := orderSNs[i:end]

		orders, err := c.fetchOrderDetailsBatch(ctx, accessToken, shopID, batch)
		if err != nil {
			return allOrders, err
		}

		allOrders = append(allOrders, orders...)

		select {
		case <-ctx.Done():
			return allOrders, ctx.Err()
		default:
		}
	}

	return allOrders, nil
}

func (c *Connector) fetchOrderDetailsBatch(ctx context.Context, accessToken string, shopID int64, orderSNs []string) ([]ShopeeOrder, error) {
	path := getOrderDetailPath
	signParams := c.SignRequestWithParams(path, accessToken, shopID)
	signParams["order_sn_list"] = strings.Join(orderSNs, ",")
	signParams["response_optional_fields"] = "item_list,buyer_user_id"

	endpoint := fmt.Sprintf("%s%s", c.baseURL, path)

	resp, err := c.DoGet(ctx, endpoint, nil, signParams)
	if err != nil {
		return nil, err
	}

	var detailResp struct {
		Error    string `json:"error"`
		Message  string `json:"message"`
		Response struct {
			OrderList []ShopeeOrder `json:"order_list"`
		} `json:"response"`
	}

	if err := c.ParseJSON(resp.Body, &detailResp); err != nil {
		return nil, err
	}

	if detailResp.Error != "" {
		return nil, c.HandleShopeeError(detailResp.Error, detailResp.Message, 0)
	}

	return detailResp.Response.OrderList, nil
}

// CalculateRevenueFromOrders calculates total revenue from completed orders
func (c *Connector) CalculateRevenueFromOrders(orders []ShopeeOrder) (totalRevenue decimal.Decimal, completedCount int) {
	for _, order := range orders {
		if order.OrderStatus == OrderStatusCompleted {
			totalRevenue = totalRevenue.Add(decimal.NewFromFloat(order.TotalAmount))
			completedCount++
		}
	}
	return
}

// CalculateDailyRevenue groups orders by date and calculates daily revenue
func (c *Connector) CalculateDailyRevenue(orders []ShopeeOrder) map[string]decimal.Decimal {
	dailyRevenue := make(map[string]decimal.Decimal)

	for _, order := range orders {
		if order.OrderStatus == OrderStatusCompleted && order.PayTime > 0 {
			date := time.Unix(order.PayTime, 0).Format("2006-01-02")
			dailyRevenue[date] = dailyRevenue[date].Add(decimal.NewFromFloat(order.TotalAmount))
		}
	}

	return dailyRevenue
}

// ============================================================================
// Aggregated Insights
// ============================================================================

// GetAccountInsightsComplete retrieves complete account insights
func (c *Connector) GetAccountInsightsComplete(ctx context.Context, accessToken string, shopID int64, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	// Get all campaigns
	campaigns, err := c.GetCampaigns(ctx, accessToken, fmt.Sprintf("%d", shopID))
	if err != nil {
		return nil, err
	}

	var totalSpend, totalRevenue decimal.Decimal
	var totalImpressions, totalClicks, totalConversions int64

	// Get metrics for each campaign
	for _, campaign := range campaigns {
		params := DailyPerformanceParams{
			ShopID:    shopID,
			StartDate: dateRange.StartDate,
			EndDate:   dateRange.EndDate,
		}

		// Parse campaign ID
		if campaignID, err := strconv.ParseUint(campaign.PlatformCampaignID, 10, 64); err == nil {
			params.CampaignID = campaignID
		}

		metrics, err := c.GetAdsDailyPerformanceComplete(ctx, accessToken, params)
		if err != nil {
			continue // Skip failed campaigns
		}

		for _, m := range metrics {
			totalSpend = totalSpend.Add(decimal.NewFromFloat(m.GetTotalSpend()))
			totalRevenue = totalRevenue.Add(decimal.NewFromFloat(m.GetTotalGMV()))
			totalImpressions += m.Impression
			totalClicks += m.Click
			totalConversions += m.GetTotalOrders()
		}
	}

	// Calculate aggregated metrics
	result := &entity.AggregatedMetrics{
		TotalSpend:       totalSpend,
		TotalImpressions: totalImpressions,
		TotalClicks:      totalClicks,
		TotalConversions: totalConversions,
		TotalRevenue:     totalRevenue,
		Currency:         c.GetCurrency(),
	}

	// Calculate derived metrics with zero-division protection
	if totalImpressions > 0 {
		result.AverageCTR = float64(totalClicks) / float64(totalImpressions) * 100
		result.AverageCPM = totalSpend.Div(decimal.NewFromInt(totalImpressions)).Mul(decimal.NewFromInt(1000))
	}

	if totalClicks > 0 {
		result.AverageCPC = totalSpend.Div(decimal.NewFromInt(totalClicks))
	}

	if totalConversions > 0 {
		result.AverageCPA = totalSpend.Div(decimal.NewFromInt(totalConversions))
	}

	if !totalSpend.IsZero() {
		result.OverallROAS, _ = totalRevenue.Div(totalSpend).Float64()
	}

	return result, nil
}

// ============================================================================
// Rate Limit Handling
// ============================================================================

// ShopeeRateLimitInfo holds rate limit information
type ShopeeRateLimitInfo struct {
	RequestQuota    int
	RequestRemaining int
	ResetTimestamp  int64
}

// ParseRateLimitHeaders parses rate limit headers from Shopee response
func (c *Connector) ParseRateLimitHeaders(headers map[string][]string) *ShopeeRateLimitInfo {
	info := &ShopeeRateLimitInfo{}

	if quota, ok := headers["X-Shopee-Ratelimit-Limit"]; ok && len(quota) > 0 {
		info.RequestQuota, _ = strconv.Atoi(quota[0])
	}

	if remaining, ok := headers["X-Shopee-Ratelimit-Remaining"]; ok && len(remaining) > 0 {
		info.RequestRemaining, _ = strconv.Atoi(remaining[0])
	}

	if reset, ok := headers["X-Shopee-Ratelimit-Reset"]; ok && len(reset) > 0 {
		info.ResetTimestamp, _ = strconv.ParseInt(reset[0], 10, 64)
	}

	return info
}

// ============================================================================
// Utility Functions
// ============================================================================

// SortAdsByPerformance sorts ads by a performance metric
func SortAdsByPerformance(metrics []ShopeePerformanceMetrics, sortBy string, descending bool) {
	sort.Slice(metrics, func(i, j int) bool {
		var valI, valJ float64

		switch sortBy {
		case "spend":
			valI, valJ = metrics[i].GetTotalSpend(), metrics[j].GetTotalSpend()
		case "gmv":
			valI, valJ = metrics[i].GetTotalGMV(), metrics[j].GetTotalGMV()
		case "roas":
			valI, valJ = metrics[i].DirectROAS, metrics[j].DirectROAS
		case "clicks":
			valI, valJ = float64(metrics[i].Click), float64(metrics[j].Click)
		case "impressions":
			valI, valJ = float64(metrics[i].Impression), float64(metrics[j].Impression)
		case "orders":
			valI, valJ = float64(metrics[i].GetTotalOrders()), float64(metrics[j].GetTotalOrders())
		default:
			valI, valJ = metrics[i].GetTotalSpend(), metrics[j].GetTotalSpend()
		}

		if descending {
			return valI > valJ
		}
		return valI < valJ
	})
}

// Ensure Connector implements PlatformConnector interface
var _ service.PlatformConnector = (*Connector)(nil)
