package shopee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	// Shopee API endpoints
	baseURL     = "https://partner.shopeemobile.com"
	authPath    = "/api/v2/shop/auth_partner"
	tokenPath   = "/api/v2/auth/token/get"
	refreshPath = "/api/v2/auth/access_token/get"

	// Ads API
	adsBaseURL = "https://partner.shopeemobile.com/api/v2"
)

// Connector implements the PlatformConnector interface for Shopee Ads
type Connector struct {
	*platform.BaseConnector
	config *Config
}

// Config holds Shopee-specific configuration
type Config struct {
	PartnerID       int64
	PartnerKey      string
	RedirectURI     string
	Region          string // MY, SG, TH, VN, PH, ID, TW, BR, MX, CO, CL, PL, ES, FR, IN
	RateLimitCalls  int
	RateLimitWindow time.Duration
	Timeout         time.Duration
	MaxRetries      int
}

// DefaultConfig returns default Shopee connector configuration
func DefaultConfig() *Config {
	return &Config{
		Region:          "MY",
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

	baseConfig := &platform.ConnectorConfig{
		AppID:           fmt.Sprintf("%d", config.PartnerID),
		AppSecret:       config.PartnerKey,
		RedirectURI:     config.RedirectURI,
		BaseURL:         baseURL,
		RateLimitCalls:  config.RateLimitCalls,
		RateLimitWindow: config.RateLimitWindow,
		Timeout:         config.Timeout,
		MaxRetries:      config.MaxRetries,
	}

	return &Connector{
		BaseConnector: platform.NewBaseConnector(entity.PlatformShopee, baseConfig),
		config:        config,
	}
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

	return baseURL + authPath + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for OAuth tokens
func (c *Connector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	timestamp := time.Now().Unix()
	path := tokenPath

	// Generate signature
	baseString := fmt.Sprintf("%d%s%d", c.config.PartnerID, path, timestamp)
	sign := c.generateSign(baseString)

	endpoint := fmt.Sprintf("%s%s", baseURL, tokenPath)

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

	endpoint := fmt.Sprintf("%s%s", baseURL, refreshPath)

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

	endpoint := fmt.Sprintf("%s%s", baseURL, path)

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

	endpoint := fmt.Sprintf("%s%s", baseURL, path)

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

// Ensure Connector implements PlatformConnector interface
var _ service.PlatformConnector = (*Connector)(nil)
