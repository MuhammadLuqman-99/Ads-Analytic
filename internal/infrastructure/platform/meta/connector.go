package meta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/internal/infrastructure/platform"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/httpclient"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	// Meta API endpoints
	baseURL  = "https://graph.facebook.com"
	authURL  = "https://www.facebook.com/v18.0/dialog/oauth"
	tokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"

	// Scopes required for ads management
	defaultScopes = "ads_read,ads_management,business_management,read_insights"

	// API fields
	campaignFields = "id,name,objective,status,daily_budget,lifetime_budget,start_time,end_time,created_time,updated_time"
	adSetFields    = "id,name,status,daily_budget,lifetime_budget,bid_amount,billing_event,optimization_goal,targeting,start_time,end_time"
	adFields       = "id,name,status,creative{id,name,title,body,call_to_action_type,image_url,video_id,thumbnail_url,link_url}"
	insightFields  = "impressions,reach,clicks,unique_clicks,spend,actions,conversions,purchase_roas,cost_per_action_type,video_p25_watched_actions,video_p50_watched_actions,video_p75_watched_actions,video_p100_watched_actions"

	// Extended insight fields for comprehensive metrics
	extendedInsightFields = "impressions,reach,frequency,clicks,unique_clicks,ctr,cpc,cpm,cpp,spend,actions,action_values,conversions,conversion_values,cost_per_action_type,cost_per_conversion,purchase_roas,website_purchase_roas,video_p25_watched_actions,video_p50_watched_actions,video_p75_watched_actions,video_p100_watched_actions,video_avg_time_watched_actions"

	// Meta error codes
	MetaErrorCodeTokenExpired   = 190
	MetaErrorCodeRateLimit      = 17
	MetaErrorCodePermission     = 10
	MetaErrorCodeInvalidParam   = 100
	MetaErrorCodeUserThrottled  = 4
	MetaErrorCodeAppThrottled   = 613
	MetaErrorCodeReportTimeout  = 2601
	MetaErrorCodeAsyncJobFailed = 2602

	// Async report polling
	asyncReportPollInterval = 5 * time.Second
	asyncReportMaxWait      = 10 * time.Minute
	largeDateRangeThreshold = 93 // Days - use async for >93 days
)

// Connector implements the PlatformConnector interface for Meta (Facebook) Ads
type Connector struct {
	*platform.BaseConnector
	config     *Config
	apiVersion string

	// Rate limit tracking from X-Business-Use-Case-Usage header
	rateLimitInfo *MetaRateLimitInfo
}

// MetaRateLimitInfo tracks rate limit from X-Business-Use-Case-Usage header
type MetaRateLimitInfo struct {
	CallCount                int     `json:"call_count"`
	TotalCPUTime             int     `json:"total_cputime"`
	TotalTime                int     `json:"total_time"`
	EstimatedTimeToRegain    int     `json:"estimated_time_to_regain_access"`
	Type                     string  `json:"type"`
	AccIDUtilPct             float64 `json:"acc_id_util_pct"`
	AdsInsightsThrottlePct   float64 `json:"ads_insights_throttle_pct"`
	AdsManagementThrottlePct float64 `json:"ads_management_throttle_pct"`
}

// MetaError represents a Meta API error response
type MetaError struct {
	Code       int    `json:"code"`
	Subcode    int    `json:"error_subcode"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	FBTraceID  string `json:"fbtrace_id"`
	IsTransient bool  `json:"is_transient"`
}

// AsyncReportStatus represents the status of an async insights report
type AsyncReportStatus struct {
	ID                 string  `json:"id"`
	AccountID          string  `json:"account_id"`
	TimeRef            int64   `json:"time_ref"`
	AsyncStatus        string  `json:"async_status"`
	AsyncPercentComplete int   `json:"async_percent_completion"`
	ResultURL          string  `json:"result_url,omitempty"`
}

// Config holds Meta-specific configuration
type Config struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	APIVersion      string
	RateLimitCalls  int
	RateLimitWindow time.Duration
	Timeout         time.Duration
	MaxRetries      int
}

// DefaultConfig returns default Meta connector configuration
func DefaultConfig() *Config {
	return &Config{
		APIVersion:      "v18.0",
		RateLimitCalls:  200,
		RateLimitWindow: time.Hour,
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}
}

// NewConnector creates a new Meta Ads connector
func NewConnector(config *Config) *Connector {
	if config == nil {
		config = DefaultConfig()
	}

	baseConfig := &platform.ConnectorConfig{
		AppID:           config.AppID,
		AppSecret:       config.AppSecret,
		RedirectURI:     config.RedirectURI,
		APIVersion:      config.APIVersion,
		BaseURL:         baseURL,
		RateLimitCalls:  config.RateLimitCalls,
		RateLimitWindow: config.RateLimitWindow,
		Timeout:         config.Timeout,
		MaxRetries:      config.MaxRetries,
	}

	return &Connector{
		BaseConnector: platform.NewBaseConnector(entity.PlatformMeta, baseConfig),
		config:        config,
		apiVersion:    config.APIVersion,
	}
}

// ============================================================================
// OAuth Methods
// ============================================================================

// GetAuthURL generates the OAuth authorization URL
func (c *Connector) GetAuthURL(state string) string {
	params := url.Values{
		"client_id":     {c.config.AppID},
		"redirect_uri":  {c.config.RedirectURI},
		"state":         {state},
		"scope":         {defaultScopes},
		"response_type": {"code"},
	}
	return authURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for OAuth tokens
func (c *Connector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	resp, err := c.DoGet(ctx, tokenURL, nil, map[string]string{
		"client_id":     c.config.AppID,
		"client_secret": c.config.AppSecret,
		"redirect_uri":  c.config.RedirectURI,
		"code":          code,
	})
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	// Exchange for long-lived token
	longLivedToken, err := c.exchangeForLongLivedToken(ctx, tokenResp.AccessToken)
	if err != nil {
		// Use short-lived token if long-lived exchange fails
		return &entity.OAuthToken{
			AccessToken: tokenResp.AccessToken,
			TokenType:   tokenResp.TokenType,
			ExpiresIn:   tokenResp.ExpiresIn,
			ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		}, nil
	}

	return longLivedToken, nil
}

// exchangeForLongLivedToken exchanges a short-lived token for a long-lived one
func (c *Connector) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (*entity.OAuthToken, error) {
	endpoint := fmt.Sprintf("%s/%s/oauth/access_token", baseURL, c.apiVersion)

	resp, err := c.DoGet(ctx, endpoint, nil, map[string]string{
		"grant_type":        "fb_exchange_token",
		"client_id":         c.config.AppID,
		"client_secret":     c.config.AppSecret,
		"fb_exchange_token": shortLivedToken,
	})
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	return &entity.OAuthToken{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresIn:   tokenResp.ExpiresIn,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken refreshes an expired access token
// Note: Meta uses long-lived tokens that need to be exchanged before expiry
func (c *Connector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	// Meta doesn't use refresh tokens in the traditional sense
	// Long-lived tokens need to be exchanged before expiry
	return c.exchangeForLongLivedToken(ctx, refreshToken)
}

// RevokeToken revokes an access token
func (c *Connector) RevokeToken(ctx context.Context, accessToken string) error {
	endpoint := fmt.Sprintf("%s/%s/me/permissions", baseURL, c.apiVersion)

	_, err := c.DoRequest(ctx, &httpclient.Request{
		Method:  http.MethodDelete,
		URL:     endpoint,
		Headers: c.BuildAuthHeader(accessToken),
	})
	return err
}

// ============================================================================
// User & Account Methods
// ============================================================================

// GetUserInfo retrieves the authenticated user's information
func (c *Connector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	endpoint := fmt.Sprintf("%s/%s/me", baseURL, c.apiVersion)

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), map[string]string{
		"fields": "id,name,email,picture",
	})
	if err != nil {
		return nil, err
	}

	var userResp struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := c.ParseJSON(resp.Body, &userResp); err != nil {
		return nil, err
	}

	return &entity.PlatformUser{
		ID:        userResp.ID,
		Name:      userResp.Name,
		Email:     userResp.Email,
		AvatarURL: userResp.Picture.Data.URL,
	}, nil
}

// GetAdAccounts retrieves all ad accounts accessible by the token
func (c *Connector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	endpoint := fmt.Sprintf("%s/%s/me/adaccounts", baseURL, c.apiVersion)

	var allAccounts []entity.PlatformAccount

	// Fetch all pages
	cursor := ""
	for {
		params := map[string]string{
			"fields": "id,name,currency,timezone_name,account_status",
			"limit":  "100",
		}
		if cursor != "" {
			params["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
		if err != nil {
			return allAccounts, err
		}

		var pageResp struct {
			Data []struct {
				ID            string `json:"id"`
				Name          string `json:"name"`
				Currency      string `json:"currency"`
				TimezoneName  string `json:"timezone_name"`
				AccountStatus int    `json:"account_status"`
			} `json:"data"`
			Paging struct {
				Cursors struct {
					After string `json:"after"`
				} `json:"cursors"`
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allAccounts, err
		}

		for _, acc := range pageResp.Data {
			allAccounts = append(allAccounts, entity.PlatformAccount{
				ID:       strings.TrimPrefix(acc.ID, "act_"),
				Name:     acc.Name,
				Currency: acc.Currency,
				Timezone: acc.TimezoneName,
				Status:   c.mapAccountStatus(acc.AccountStatus),
			})
		}

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After
	}

	return allAccounts, nil
}

// mapAccountStatus maps Meta account status to string
func (c *Connector) mapAccountStatus(status int) string {
	switch status {
	case 1:
		return "active"
	case 2:
		return "disabled"
	case 3:
		return "unsettled"
	case 7:
		return "pending_risk_review"
	case 8:
		return "pending_settlement"
	case 9:
		return "in_grace_period"
	case 100:
		return "pending_closure"
	case 101:
		return "closed"
	case 201:
		return "any_active"
	case 202:
		return "any_closed"
	default:
		return "unknown"
	}
}

// ============================================================================
// Campaign Methods
// ============================================================================

// GetCampaigns retrieves all campaigns for an ad account
func (c *Connector) GetCampaigns(ctx context.Context, accessToken string, adAccountID string) ([]entity.Campaign, error) {
	endpoint := fmt.Sprintf("%s/%s/act_%s/campaigns", baseURL, c.apiVersion, adAccountID)

	var allCampaigns []entity.Campaign

	cursor := ""
	for {
		params := map[string]string{
			"fields": campaignFields,
			"limit":  "100",
		}
		if cursor != "" {
			params["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
		if err != nil {
			return allCampaigns, err
		}

		var pageResp struct {
			Data   []metaCampaign      `json:"data"`
			Paging platform.PagingInfo `json:"paging"`
		}

		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allCampaigns, err
		}

		for _, mc := range pageResp.Data {
			campaign := c.mapCampaign(mc)
			allCampaigns = append(allCampaigns, campaign)
		}

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After
	}

	return allCampaigns, nil
}

// GetCampaign retrieves a single campaign by ID
func (c *Connector) GetCampaign(ctx context.Context, accessToken string, campaignID string) (*entity.Campaign, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, c.apiVersion, campaignID)

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), map[string]string{
		"fields": campaignFields,
	})
	if err != nil {
		return nil, err
	}

	var mc metaCampaign
	if err := c.ParseJSON(resp.Body, &mc); err != nil {
		return nil, err
	}

	campaign := c.mapCampaign(mc)
	return &campaign, nil
}

// ============================================================================
// AdSet Methods
// ============================================================================

// GetAdSets retrieves all ad sets for a campaign
func (c *Connector) GetAdSets(ctx context.Context, accessToken string, campaignID string) ([]entity.AdSet, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/adsets", baseURL, c.apiVersion, campaignID)

	var allAdSets []entity.AdSet

	cursor := ""
	for {
		params := map[string]string{
			"fields": adSetFields,
			"limit":  "100",
		}
		if cursor != "" {
			params["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
		if err != nil {
			return allAdSets, err
		}

		var pageResp struct {
			Data   []metaAdSet         `json:"data"`
			Paging platform.PagingInfo `json:"paging"`
		}

		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allAdSets, err
		}

		for _, mas := range pageResp.Data {
			adSet := c.mapAdSet(mas)
			allAdSets = append(allAdSets, adSet)
		}

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After
	}

	return allAdSets, nil
}

// GetAdSet retrieves a single ad set by ID
func (c *Connector) GetAdSet(ctx context.Context, accessToken string, adSetID string) (*entity.AdSet, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, c.apiVersion, adSetID)

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), map[string]string{
		"fields": adSetFields,
	})
	if err != nil {
		return nil, err
	}

	var mas metaAdSet
	if err := c.ParseJSON(resp.Body, &mas); err != nil {
		return nil, err
	}

	adSet := c.mapAdSet(mas)
	return &adSet, nil
}

// ============================================================================
// Ad Methods
// ============================================================================

// GetAds retrieves all ads for an ad set
func (c *Connector) GetAds(ctx context.Context, accessToken string, adSetID string) ([]entity.Ad, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/ads", baseURL, c.apiVersion, adSetID)

	var allAds []entity.Ad

	cursor := ""
	for {
		params := map[string]string{
			"fields": adFields,
			"limit":  "100",
		}
		if cursor != "" {
			params["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
		if err != nil {
			return allAds, err
		}

		var pageResp struct {
			Data   []metaAd            `json:"data"`
			Paging platform.PagingInfo `json:"paging"`
		}

		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allAds, err
		}

		for _, ma := range pageResp.Data {
			ad := c.mapAd(ma)
			allAds = append(allAds, ad)
		}

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After
	}

	return allAds, nil
}

// GetAd retrieves a single ad by ID
func (c *Connector) GetAd(ctx context.Context, accessToken string, adID string) (*entity.Ad, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, c.apiVersion, adID)

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), map[string]string{
		"fields": adFields,
	})
	if err != nil {
		return nil, err
	}

	var ma metaAd
	if err := c.ParseJSON(resp.Body, &ma); err != nil {
		return nil, err
	}

	ad := c.mapAd(ma)
	return &ad, nil
}

// ============================================================================
// Insights Methods
// ============================================================================

// GetCampaignInsights retrieves performance metrics for a campaign
func (c *Connector) GetCampaignInsights(ctx context.Context, accessToken string, campaignID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/insights", baseURL, c.apiVersion, campaignID)

	params := map[string]string{
		"fields":         insightFields,
		"time_range":     c.formatTimeRange(dateRange),
		"time_increment": "1", // Daily breakdown
		"level":          "campaign",
	}

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var insightsResp struct {
		Data []metaInsight `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &insightsResp); err != nil {
		return nil, err
	}

	var metrics []entity.CampaignMetricsDaily
	for _, insight := range insightsResp.Data {
		metric := c.mapCampaignInsight(insight, campaignID)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAdSetInsights retrieves performance metrics for an ad set
func (c *Connector) GetAdSetInsights(ctx context.Context, accessToken string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/insights", baseURL, c.apiVersion, adSetID)

	params := map[string]string{
		"fields":         insightFields,
		"time_range":     c.formatTimeRange(dateRange),
		"time_increment": "1",
		"level":          "adset",
	}

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var insightsResp struct {
		Data []metaInsight `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &insightsResp); err != nil {
		return nil, err
	}

	var metrics []entity.AdSetMetricsDaily
	for _, insight := range insightsResp.Data {
		metric := c.mapAdSetInsight(insight, adSetID)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAdInsights retrieves performance metrics for an ad
func (c *Connector) GetAdInsights(ctx context.Context, accessToken string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/insights", baseURL, c.apiVersion, adID)

	params := map[string]string{
		"fields":         insightFields,
		"time_range":     c.formatTimeRange(dateRange),
		"time_increment": "1",
		"level":          "ad",
	}

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var insightsResp struct {
		Data []metaInsight `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &insightsResp); err != nil {
		return nil, err
	}

	var metrics []entity.AdMetricsDaily
	for _, insight := range insightsResp.Data {
		metric := c.mapAdInsight(insight, adID)
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetAccountInsights retrieves aggregated insights for an ad account
func (c *Connector) GetAccountInsights(ctx context.Context, accessToken string, adAccountID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	endpoint := fmt.Sprintf("%s/%s/act_%s/insights", baseURL, c.apiVersion, adAccountID)

	params := map[string]string{
		"fields":     insightFields,
		"time_range": c.formatTimeRange(dateRange),
		"level":      "account",
	}

	resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var insightsResp struct {
		Data []metaInsight `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &insightsResp); err != nil {
		return nil, err
	}

	if len(insightsResp.Data) == 0 {
		return &entity.AggregatedMetrics{}, nil
	}

	return c.mapAggregatedInsight(insightsResp.Data[0]), nil
}

// HealthCheck verifies the connector can connect to Meta API
func (c *Connector) HealthCheck(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/%s/me", baseURL, c.apiVersion)

	// Use app token for health check
	appToken := fmt.Sprintf("%s|%s", c.config.AppID, c.config.AppSecret)

	_, err := c.DoGet(ctx, endpoint, nil, map[string]string{
		"access_token": appToken,
	})
	if err != nil {
		return errors.Wrap(err, errors.ErrCodePlatformUnavailable, "Meta API health check failed", http.StatusServiceUnavailable)
	}

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (c *Connector) formatTimeRange(dateRange entity.DateRange) string {
	return fmt.Sprintf(`{"since":"%s","until":"%s"}`,
		dateRange.StartDate.Format("2006-01-02"),
		dateRange.EndDate.Format("2006-01-02"),
	)
}

// ============================================================================
// Meta API Response Types
// ============================================================================

type metaCampaign struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Objective      string `json:"objective"`
	Status         string `json:"status"`
	DailyBudget    string `json:"daily_budget"`
	LifetimeBudget string `json:"lifetime_budget"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	CreatedTime    string `json:"created_time"`
	UpdatedTime    string `json:"updated_time"`
}

type metaAdSet struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Status           string          `json:"status"`
	DailyBudget      string          `json:"daily_budget"`
	LifetimeBudget   string          `json:"lifetime_budget"`
	BidAmount        string          `json:"bid_amount"`
	BillingEvent     string          `json:"billing_event"`
	OptimizationGoal string          `json:"optimization_goal"`
	Targeting        json.RawMessage `json:"targeting"`
	StartTime        string          `json:"start_time"`
	EndTime          string          `json:"end_time"`
}

type metaAd struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Creative struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Title        string `json:"title"`
		Body         string `json:"body"`
		CallToAction string `json:"call_to_action_type"`
		ImageURL     string `json:"image_url"`
		VideoID      string `json:"video_id"`
		ThumbnailURL string `json:"thumbnail_url"`
		LinkURL      string `json:"link_url"`
	} `json:"creative"`
}

type metaInsight struct {
	// Object identifiers (when fetched at account level with level=campaign/adset/ad)
	CampaignID   string `json:"campaign_id,omitempty"`
	CampaignName string `json:"campaign_name,omitempty"`
	AdSetID      string `json:"adset_id,omitempty"`
	AdSetName    string `json:"adset_name,omitempty"`
	AdID         string `json:"ad_id,omitempty"`
	AdName       string `json:"ad_name,omitempty"`
	AccountID    string `json:"account_id,omitempty"`

	// Date range
	DateStart string `json:"date_start"`
	DateStop  string `json:"date_stop"`

	// Core metrics
	Impressions  string `json:"impressions"`
	Reach        string `json:"reach"`
	Frequency    string `json:"frequency,omitempty"`
	Clicks       string `json:"clicks"`
	UniqueClicks string `json:"unique_clicks"`
	Spend        string `json:"spend"`

	// Pre-calculated metrics from API
	CTR string `json:"ctr,omitempty"`
	CPC string `json:"cpc,omitempty"`
	CPM string `json:"cpm,omitempty"`
	CPP string `json:"cpp,omitempty"` // Cost per 1000 people reached

	// Actions (engagement, conversions, etc.)
	Actions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"actions"`
	ActionValues []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"action_values,omitempty"`

	// Conversions
	Conversions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"conversions"`
	ConversionValues []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"conversion_values,omitempty"`

	// Cost per action
	CostPerActionType []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_action_type,omitempty"`
	CostPerConversion []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"cost_per_conversion,omitempty"`

	// ROAS
	PurchaseROAS []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"purchase_roas"`
	WebsitePurchaseROAS []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"website_purchase_roas,omitempty"`

	// Video metrics
	VideoP25WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p25_watched_actions"`
	VideoP50WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p50_watched_actions"`
	VideoP75WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p75_watched_actions"`
	VideoP100WatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_p100_watched_actions"`
	VideoAvgTimeWatchedActions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"video_avg_time_watched_actions,omitempty"`
}

// ============================================================================
// Mapping Methods
// ============================================================================

func (c *Connector) mapCampaign(mc metaCampaign) entity.Campaign {
	campaign := entity.Campaign{
		Platform:             entity.PlatformMeta,
		PlatformCampaignID:   mc.ID,
		PlatformCampaignName: mc.Name,
		Status:               c.mapCampaignStatus(mc.Status),
		Objective:            c.mapObjective(mc.Objective),
	}

	if mc.DailyBudget != "" {
		budget, _ := decimal.NewFromString(mc.DailyBudget)
		// Meta returns budget in cents, convert to dollars
		budget = budget.Div(decimal.NewFromInt(100))
		campaign.DailyBudget = &budget
	}

	if mc.LifetimeBudget != "" {
		budget, _ := decimal.NewFromString(mc.LifetimeBudget)
		budget = budget.Div(decimal.NewFromInt(100))
		campaign.LifetimeBudget = &budget
	}

	if mc.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, mc.StartTime); err == nil {
			campaign.StartDate = &t
		}
	}

	if mc.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, mc.EndTime); err == nil {
			campaign.EndDate = &t
		}
	}

	if mc.CreatedTime != "" {
		if t, err := time.Parse(time.RFC3339, mc.CreatedTime); err == nil {
			campaign.PlatformCreatedAt = &t
		}
	}

	if mc.UpdatedTime != "" {
		if t, err := time.Parse(time.RFC3339, mc.UpdatedTime); err == nil {
			campaign.PlatformUpdatedAt = &t
		}
	}

	return campaign
}

func (c *Connector) mapAdSet(mas metaAdSet) entity.AdSet {
	adSet := entity.AdSet{
		Platform:          entity.PlatformMeta,
		PlatformAdSetID:   mas.ID,
		PlatformAdSetName: mas.Name,
		Status:            c.mapCampaignStatus(mas.Status),
		BidStrategy:       mas.OptimizationGoal,
	}

	if mas.DailyBudget != "" {
		budget, _ := decimal.NewFromString(mas.DailyBudget)
		budget = budget.Div(decimal.NewFromInt(100))
		adSet.DailyBudget = &budget
	}

	if mas.LifetimeBudget != "" {
		budget, _ := decimal.NewFromString(mas.LifetimeBudget)
		budget = budget.Div(decimal.NewFromInt(100))
		adSet.LifetimeBudget = &budget
	}

	if mas.BidAmount != "" {
		bid, _ := decimal.NewFromString(mas.BidAmount)
		bid = bid.Div(decimal.NewFromInt(100))
		adSet.BidAmount = &bid
	}

	if len(mas.Targeting) > 0 {
		var targeting map[string]interface{}
		if err := json.Unmarshal(mas.Targeting, &targeting); err == nil {
			adSet.Targeting = targeting
		}
	}

	if mas.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, mas.StartTime); err == nil {
			adSet.StartDate = &t
		}
	}

	if mas.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, mas.EndTime); err == nil {
			adSet.EndDate = &t
		}
	}

	return adSet
}

func (c *Connector) mapAd(ma metaAd) entity.Ad {
	return entity.Ad{
		Platform:       entity.PlatformMeta,
		PlatformAdID:   ma.ID,
		PlatformAdName: ma.Name,
		Status:         c.mapCampaignStatus(ma.Status),
		Headline:       ma.Creative.Title,
		Description:    ma.Creative.Body,
		CallToAction:   ma.Creative.CallToAction,
		DestinationURL: ma.Creative.LinkURL,
		ImageURL:       ma.Creative.ImageURL,
		ThumbnailURL:   ma.Creative.ThumbnailURL,
		CreativeData: entity.JSONMap{
			"creative_id": ma.Creative.ID,
			"video_id":    ma.Creative.VideoID,
		},
	}
}

func (c *Connector) mapCampaignInsight(mi metaInsight, campaignID string) entity.CampaignMetricsDaily {
	metricDate, _ := time.Parse("2006-01-02", mi.DateStart)

	metrics := entity.CampaignMetricsDaily{
		BaseEntity: entity.BaseEntity{ID: uuid.New()},
		Platform:   entity.PlatformMeta,
		MetricDate: metricDate,
		Currency:   "USD", // Meta reports in account currency
	}

	metrics.Impressions = c.parseInt64(mi.Impressions)
	metrics.Reach = c.parseInt64(mi.Reach)
	metrics.Clicks = c.parseInt64(mi.Clicks)
	metrics.UniqueClicks = c.parseInt64(mi.UniqueClicks)
	metrics.Spend = c.parseDecimal(mi.Spend)

	// Parse video metrics
	metrics.VideoViewsP25 = c.sumActionValues(mi.VideoP25WatchedActions)
	metrics.VideoViewsP50 = c.sumActionValues(mi.VideoP50WatchedActions)
	metrics.VideoViewsP75 = c.sumActionValues(mi.VideoP75WatchedActions)
	metrics.VideoViewsP100 = c.sumActionValues(mi.VideoP100WatchedActions)
	metrics.VideoViews = metrics.VideoViewsP25 // Use P25 as video views

	// Parse action metrics
	for _, action := range mi.Actions {
		switch action.ActionType {
		case "like":
			metrics.Likes = c.parseInt64(action.Value)
		case "comment":
			metrics.Comments = c.parseInt64(action.Value)
		case "post":
			metrics.Shares = c.parseInt64(action.Value)
		case "onsite_conversion.post_save":
			metrics.Saves = c.parseInt64(action.Value)
		case "omni_add_to_cart":
			metrics.AddToCart = c.parseInt64(action.Value)
		case "omni_initiated_checkout":
			metrics.CheckoutInitiated = c.parseInt64(action.Value)
		case "omni_purchase":
			metrics.Purchases = c.parseInt64(action.Value)
		}
	}

	// Parse conversion metrics
	for _, conv := range mi.Conversions {
		if conv.ActionType == "omni_purchase" {
			metrics.Conversions = c.parseInt64(conv.Value)
		}
	}

	// Parse ROAS
	for _, roas := range mi.PurchaseROAS {
		if roas.ActionType == "omni_purchase" {
			roasVal, _ := strconv.ParseFloat(roas.Value, 64)
			metrics.ROAS = &roasVal
		}
	}

	// Calculate derived metrics
	metrics.CalculateDerivedMetrics()

	return metrics
}

func (c *Connector) mapAdSetInsight(mi metaInsight, adSetID string) entity.AdSetMetricsDaily {
	metricDate, _ := time.Parse("2006-01-02", mi.DateStart)

	metrics := entity.AdSetMetricsDaily{
		BaseEntity: entity.BaseEntity{ID: uuid.New()},
		Platform:   entity.PlatformMeta,
		MetricDate: metricDate,
		Currency:   "USD",
	}

	metrics.Impressions = c.parseInt64(mi.Impressions)
	metrics.Reach = c.parseInt64(mi.Reach)
	metrics.Clicks = c.parseInt64(mi.Clicks)
	metrics.UniqueClicks = c.parseInt64(mi.UniqueClicks)
	metrics.Spend = c.parseDecimal(mi.Spend)

	return metrics
}

func (c *Connector) mapAdInsight(mi metaInsight, adID string) entity.AdMetricsDaily {
	metricDate, _ := time.Parse("2006-01-02", mi.DateStart)

	metrics := entity.AdMetricsDaily{
		BaseEntity: entity.BaseEntity{ID: uuid.New()},
		Platform:   entity.PlatformMeta,
		MetricDate: metricDate,
		Currency:   "USD",
	}

	metrics.Impressions = c.parseInt64(mi.Impressions)
	metrics.Reach = c.parseInt64(mi.Reach)
	metrics.Clicks = c.parseInt64(mi.Clicks)
	metrics.UniqueClicks = c.parseInt64(mi.UniqueClicks)
	metrics.Spend = c.parseDecimal(mi.Spend)

	return metrics
}

func (c *Connector) mapAggregatedInsight(mi metaInsight) *entity.AggregatedMetrics {
	metrics := &entity.AggregatedMetrics{
		TotalSpend:       c.parseDecimal(mi.Spend),
		TotalImpressions: c.parseInt64(mi.Impressions),
		TotalClicks:      c.parseInt64(mi.Clicks),
		Currency:         "USD",
	}

	// Parse conversions and purchase value
	for _, action := range mi.Actions {
		if action.ActionType == "omni_purchase" {
			metrics.TotalConversions = c.parseInt64(action.Value)
		}
	}

	// Calculate derived metrics
	if metrics.TotalImpressions > 0 {
		metrics.AverageCTR = float64(metrics.TotalClicks) / float64(metrics.TotalImpressions) * 100
	}

	if metrics.TotalClicks > 0 {
		metrics.AverageCPC = metrics.TotalSpend.Div(decimal.NewFromInt(metrics.TotalClicks))
	}

	if metrics.TotalImpressions > 0 {
		metrics.AverageCPM = metrics.TotalSpend.Div(decimal.NewFromInt(metrics.TotalImpressions)).Mul(decimal.NewFromInt(1000))
	}

	if metrics.TotalConversions > 0 {
		metrics.AverageCPA = metrics.TotalSpend.Div(decimal.NewFromInt(metrics.TotalConversions))
	}

	// Parse ROAS
	for _, roas := range mi.PurchaseROAS {
		if roas.ActionType == "omni_purchase" {
			metrics.OverallROAS, _ = strconv.ParseFloat(roas.Value, 64)
		}
	}

	return metrics
}

func (c *Connector) mapCampaignStatus(status string) entity.CampaignStatus {
	switch strings.ToUpper(status) {
	case "ACTIVE":
		return entity.CampaignStatusActive
	case "PAUSED":
		return entity.CampaignStatusPaused
	case "DELETED":
		return entity.CampaignStatusDeleted
	case "ARCHIVED":
		return entity.CampaignStatusArchived
	case "DRAFT":
		return entity.CampaignStatusDraft
	default:
		// Unknown statuses default to paused for safety
		return entity.CampaignStatusPaused
	}
}

func (c *Connector) mapObjective(objective string) entity.CampaignObjective {
	switch strings.ToUpper(objective) {
	// New Meta OUTCOME_* objectives (API v18+)
	case "OUTCOME_AWARENESS":
		return entity.ObjectiveAwareness
	case "OUTCOME_TRAFFIC":
		return entity.ObjectiveTraffic
	case "OUTCOME_ENGAGEMENT":
		return entity.ObjectiveEngagement
	case "OUTCOME_LEADS":
		return entity.ObjectiveLeads
	case "OUTCOME_APP_PROMOTION":
		return entity.ObjectiveAppPromotion
	case "OUTCOME_SALES":
		return entity.ObjectiveSales

	// Legacy Meta objectives (still returned by some campaigns)
	case "BRAND_AWARENESS", "REACH":
		return entity.ObjectiveAwareness
	case "LINK_CLICKS":
		return entity.ObjectiveTraffic
	case "POST_ENGAGEMENT", "PAGE_LIKES", "EVENT_RESPONSES":
		return entity.ObjectiveEngagement
	case "LEAD_GENERATION":
		return entity.ObjectiveLeads
	case "APP_INSTALLS":
		return entity.ObjectiveAppPromotion
	case "CONVERSIONS":
		return entity.ObjectiveConversions
	case "CATALOG_SALES", "PRODUCT_CATALOG_SALES":
		return entity.ObjectiveSales
	case "VIDEO_VIEWS":
		return entity.ObjectiveVideoViews
	case "MESSAGES":
		return entity.ObjectiveMessages
	case "STORE_TRAFFIC", "STORE_VISITS":
		return entity.ObjectiveStoreTraffic

	default:
		// Default to awareness for unknown objectives
		return entity.ObjectiveAwareness
	}
}

func (c *Connector) parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func (c *Connector) parseDecimal(s string) decimal.Decimal {
	val, _ := decimal.NewFromString(s)
	return val
}

func (c *Connector) sumActionValues(actions []struct {
	ActionType string `json:"action_type"`
	Value      string `json:"value"`
}) int64 {
	var sum int64
	for _, a := range actions {
		sum += c.parseInt64(a.Value)
	}
	return sum
}

// ============================================================================
// Rate Limit Methods
// ============================================================================

// ParseRateLimitHeaders parses Meta's X-Business-Use-Case-Usage header
func (c *Connector) ParseRateLimitHeaders(headers map[string][]string) *MetaRateLimitInfo {
	usageHeader := ""
	for k, v := range headers {
		if strings.EqualFold(k, "X-Business-Use-Case-Usage") && len(v) > 0 {
			usageHeader = v[0]
			break
		}
	}

	if usageHeader == "" {
		return nil
	}

	// X-Business-Use-Case-Usage format: {"app_id":{"call_count":..., "total_cputime":...}}
	var usageMap map[string]map[string]MetaRateLimitInfo
	if err := json.Unmarshal([]byte(usageHeader), &usageMap); err != nil {
		return nil
	}

	// Get the first (and usually only) entry
	for _, appUsage := range usageMap {
		for _, info := range appUsage {
			c.rateLimitInfo = &info
			return &info
		}
	}

	return nil
}

// IsApproachingRateLimit checks if we're approaching rate limit threshold
func (c *Connector) IsApproachingRateLimit() bool {
	if c.rateLimitInfo == nil {
		return false
	}

	// Meta throttles at 80% usage - warn at 70%
	return c.rateLimitInfo.CallCount >= 70 ||
		c.rateLimitInfo.TotalCPUTime >= 70 ||
		c.rateLimitInfo.TotalTime >= 70
}

// GetEstimatedWaitTime returns estimated wait time if rate limited
func (c *Connector) GetEstimatedWaitTime() time.Duration {
	if c.rateLimitInfo == nil || c.rateLimitInfo.EstimatedTimeToRegain == 0 {
		return 0
	}
	return time.Duration(c.rateLimitInfo.EstimatedTimeToRegain) * time.Minute
}

// ============================================================================
// Enhanced Error Handling
// ============================================================================

// parseMetaError parses Meta API error from response
func (c *Connector) parseMetaError(body []byte) (*MetaError, error) {
	var errResp struct {
		Error MetaError `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return nil, err
	}
	return &errResp.Error, nil
}

// HandleMetaError handles Meta-specific error codes and returns appropriate error
func (c *Connector) HandleMetaError(statusCode int, body []byte) error {
	metaErr, err := c.parseMetaError(body)
	if err != nil {
		return errors.NewPlatformAPIError("meta", statusCode, "UNKNOWN", "Failed to parse error response")
	}

	switch metaErr.Code {
	case MetaErrorCodeTokenExpired:
		tokenErr := errors.NewTokenExpiredError("meta", "access", time.Now())
		tokenErr.AppError.WithMetadata("fb_trace_id", metaErr.FBTraceID)
		tokenErr.AppError.WithMetadata("error_type", metaErr.Type)
		return tokenErr

	case MetaErrorCodeRateLimit, MetaErrorCodeUserThrottled, MetaErrorCodeAppThrottled:
		waitTime := c.GetEstimatedWaitTime()
		if waitTime == 0 {
			waitTime = 60 * time.Second
		}
		return errors.NewRateLimitError("meta", waitTime)

	case MetaErrorCodePermission:
		return errors.NewPlatformAPIError("meta", statusCode, strconv.Itoa(metaErr.Code), "Permission denied: "+metaErr.Message)

	case MetaErrorCodeInvalidParam:
		return errors.ErrValidation(metaErr.Message).WithMetadata("platform", "meta")

	case MetaErrorCodeReportTimeout:
		return errors.NewPlatformAPIError("meta", statusCode, strconv.Itoa(metaErr.Code), "Report generation timeout - try async report")

	case MetaErrorCodeAsyncJobFailed:
		return errors.NewPlatformAPIError("meta", statusCode, strconv.Itoa(metaErr.Code), "Async job failed: "+metaErr.Message)

	default:
		platformErr := errors.NewPlatformAPIError("meta", statusCode, strconv.Itoa(metaErr.Code), metaErr.Message)
		if metaErr.IsTransient {
			platformErr.WithMetadata("is_transient", "true")
		}
		return platformErr
	}
}

// ============================================================================
// Async Report Methods for Large Date Ranges
// ============================================================================

// GetAccountInsightsAsync creates an async report for large date ranges
func (c *Connector) GetAccountInsightsAsync(ctx context.Context, accessToken string, adAccountID string, dateRange entity.DateRange, level string) ([]metaInsight, error) {
	// Step 1: Create async job
	endpoint := fmt.Sprintf("%s/%s/act_%s/insights", baseURL, c.apiVersion, adAccountID)

	params := map[string]string{
		"fields":         extendedInsightFields,
		"time_range":     c.formatTimeRange(dateRange),
		"time_increment": "1",
		"level":          level,
		"access_token":   accessToken,
	}

	// POST to create async job
	resp, err := c.DoPost(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var jobResp struct {
		ReportRunID string `json:"report_run_id"`
	}
	if err := c.ParseJSON(resp.Body, &jobResp); err != nil {
		return nil, err
	}

	// Step 2: Poll for completion
	insights, err := c.pollAsyncReport(ctx, accessToken, jobResp.ReportRunID)
	if err != nil {
		return nil, err
	}

	return insights, nil
}

// pollAsyncReport polls an async report until completion
func (c *Connector) pollAsyncReport(ctx context.Context, accessToken string, reportRunID string) ([]metaInsight, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", baseURL, c.apiVersion, reportRunID)
	deadline := time.Now().Add(asyncReportMaxWait)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(asyncReportPollInterval):
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), map[string]string{
			"fields": "async_status,async_percent_completion",
		})
		if err != nil {
			return nil, err
		}

		var status AsyncReportStatus
		if err := c.ParseJSON(resp.Body, &status); err != nil {
			return nil, err
		}

		switch status.AsyncStatus {
		case "Job Completed":
			// Fetch results
			return c.fetchAsyncReportResults(ctx, accessToken, reportRunID)

		case "Job Failed":
			return nil, errors.NewPlatformAPIError("meta", 500, "ASYNC_FAILED", "Async report job failed")

		case "Job Running", "Job Not Started":
			// Continue polling
			continue

		default:
			// Unknown status, continue polling
			continue
		}
	}

	return nil, errors.NewPlatformAPIError("meta", 504, "TIMEOUT", "Async report timed out")
}

// fetchAsyncReportResults fetches results from completed async report
func (c *Connector) fetchAsyncReportResults(ctx context.Context, accessToken string, reportRunID string) ([]metaInsight, error) {
	endpoint := fmt.Sprintf("%s/%s/%s/insights", baseURL, c.apiVersion, reportRunID)

	var allInsights []metaInsight
	cursor := ""

	for {
		params := map[string]string{"limit": "500"}
		if cursor != "" {
			params["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
		if err != nil {
			return allInsights, err
		}

		var pageResp struct {
			Data   []metaInsight       `json:"data"`
			Paging platform.PagingInfo `json:"paging"`
		}
		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allInsights, err
		}

		allInsights = append(allInsights, pageResp.Data...)

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After
	}

	return allInsights, nil
}

// ============================================================================
// Enhanced Insights Methods
// ============================================================================

// InsightsParams holds parameters for fetching insights
type InsightsParams struct {
	DateRange      entity.DateRange
	Level          string   // "campaign", "adset", "ad", "account"
	DatePreset     string   // "today", "yesterday", "this_month", "last_7d", "last_30d", etc.
	TimeIncrement  string   // "1" (daily), "7" (weekly), "monthly", "all_days"
	Fields         []string // Custom fields to fetch
	Breakdowns     []string // "age", "gender", "country", "placement", etc.
	ActionBreakdowns []string // "action_type", "action_device"
	Filtering      []InsightFilter
	UseAsync       bool
}

// InsightFilter represents a filter for insights
type InsightFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// GetInsightsWithParams fetches insights with comprehensive parameters
func (c *Connector) GetInsightsWithParams(ctx context.Context, accessToken string, objectID string, params InsightsParams) ([]metaInsight, error) {
	// Determine if async is needed based on date range
	daysDiff := int(params.DateRange.EndDate.Sub(params.DateRange.StartDate).Hours() / 24)
	useAsync := params.UseAsync || daysDiff > largeDateRangeThreshold

	endpoint := fmt.Sprintf("%s/%s/%s/insights", baseURL, c.apiVersion, objectID)

	// Build query params
	queryParams := map[string]string{
		"level": params.Level,
	}

	// Date range or preset
	if params.DatePreset != "" {
		queryParams["date_preset"] = params.DatePreset
	} else {
		queryParams["time_range"] = c.formatTimeRange(params.DateRange)
	}

	// Time increment
	if params.TimeIncrement != "" {
		queryParams["time_increment"] = params.TimeIncrement
	} else {
		queryParams["time_increment"] = "1" // Default daily
	}

	// Fields
	if len(params.Fields) > 0 {
		queryParams["fields"] = strings.Join(params.Fields, ",")
	} else {
		queryParams["fields"] = extendedInsightFields
	}

	// Breakdowns
	if len(params.Breakdowns) > 0 {
		queryParams["breakdowns"] = strings.Join(params.Breakdowns, ",")
	}

	// Action breakdowns
	if len(params.ActionBreakdowns) > 0 {
		queryParams["action_breakdowns"] = strings.Join(params.ActionBreakdowns, ",")
	}

	// Filtering
	if len(params.Filtering) > 0 {
		filterJSON, _ := json.Marshal(params.Filtering)
		queryParams["filtering"] = string(filterJSON)
	}

	// Use async for large date ranges
	if useAsync {
		return c.getInsightsAsync(ctx, accessToken, endpoint, queryParams)
	}

	return c.getInsightsSync(ctx, accessToken, endpoint, queryParams)
}

// getInsightsSync fetches insights synchronously (for small date ranges)
func (c *Connector) getInsightsSync(ctx context.Context, accessToken string, endpoint string, params map[string]string) ([]metaInsight, error) {
	var allInsights []metaInsight
	cursor := ""

	for {
		pageParams := make(map[string]string)
		for k, v := range params {
			pageParams[k] = v
		}
		pageParams["limit"] = "500"
		if cursor != "" {
			pageParams["after"] = cursor
		}

		resp, err := c.DoGet(ctx, endpoint, c.BuildAuthHeader(accessToken), pageParams)
		if err != nil {
			// Check if it's a rate limit error
			if resp != nil {
				c.ParseRateLimitHeaders(resp.Headers)
			}
			return allInsights, err
		}

		// Parse rate limit headers
		c.ParseRateLimitHeaders(resp.Headers)

		var pageResp struct {
			Data   []metaInsight       `json:"data"`
			Paging platform.PagingInfo `json:"paging"`
		}
		if err := c.ParseJSON(resp.Body, &pageResp); err != nil {
			return allInsights, err
		}

		allInsights = append(allInsights, pageResp.Data...)

		if pageResp.Paging.Next == "" {
			break
		}
		cursor = pageResp.Paging.Cursors.After

		// Check rate limit before next request
		if c.IsApproachingRateLimit() {
			waitTime := c.GetEstimatedWaitTime()
			if waitTime > 0 {
				select {
				case <-ctx.Done():
					return allInsights, ctx.Err()
				case <-time.After(waitTime):
				}
			}
		}
	}

	return allInsights, nil
}

// getInsightsAsync fetches insights asynchronously (for large date ranges)
func (c *Connector) getInsightsAsync(ctx context.Context, accessToken string, endpoint string, params map[string]string) ([]metaInsight, error) {
	// Add access_token for POST
	params["access_token"] = accessToken

	// Create async job via POST
	resp, err := c.DoPost(ctx, endpoint, c.BuildAuthHeader(accessToken), params)
	if err != nil {
		return nil, err
	}

	var jobResp struct {
		ReportRunID string `json:"report_run_id"`
	}
	if err := c.ParseJSON(resp.Body, &jobResp); err != nil {
		return nil, err
	}

	// Poll for completion
	return c.pollAsyncReport(ctx, accessToken, jobResp.ReportRunID)
}

// ============================================================================
// Batch Insights for Multiple Campaigns
// ============================================================================

// GetBatchCampaignInsights fetches insights for multiple campaigns efficiently
func (c *Connector) GetBatchCampaignInsights(ctx context.Context, accessToken string, adAccountID string, campaignIDs []string, dateRange entity.DateRange) (map[string][]entity.CampaignMetricsDaily, error) {
	results := make(map[string][]entity.CampaignMetricsDaily)

	// Use ad account level with campaign filtering for efficiency
	endpoint := fmt.Sprintf("%s/%s/act_%s/insights", baseURL, c.apiVersion, adAccountID)

	params := map[string]string{
		"fields":         extendedInsightFields,
		"time_range":     c.formatTimeRange(dateRange),
		"time_increment": "1",
		"level":          "campaign",
		"limit":          "500",
	}

	// Add campaign filter if specific campaigns requested
	if len(campaignIDs) > 0 {
		filterJSON, _ := json.Marshal([]map[string]string{
			{
				"field":    "campaign.id",
				"operator": "IN",
				"value":    "[" + strings.Join(campaignIDs, ",") + "]",
			},
		})
		params["filtering"] = string(filterJSON)
	}

	insights, err := c.getInsightsSync(ctx, accessToken, endpoint, params)
	if err != nil {
		return results, err
	}

	// Group by campaign
	for _, insight := range insights {
		campaignID := insight.CampaignID
		if campaignID == "" {
			continue
		}
		metric := c.mapCampaignInsight(insight, campaignID)
		results[campaignID] = append(results[campaignID], metric)
	}

	return results, nil
}

// Ensure Connector implements PlatformConnector interface
var _ service.PlatformConnector = (*Connector)(nil)
