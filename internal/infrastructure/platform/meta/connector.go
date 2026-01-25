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
	insightFields  = "impressions,reach,clicks,unique_clicks,spend,actions,conversions,purchase_roas,video_p25_watched_actions,video_p50_watched_actions,video_p75_watched_actions,video_p100_watched_actions"
)

// Connector implements the PlatformConnector interface for Meta (Facebook) Ads
type Connector struct {
	*platform.BaseConnector
	config     *Config
	apiVersion string
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
	DateStart    string `json:"date_start"`
	DateStop     string `json:"date_stop"`
	Impressions  string `json:"impressions"`
	Reach        string `json:"reach"`
	Clicks       string `json:"clicks"`
	UniqueClicks string `json:"unique_clicks"`
	Spend        string `json:"spend"`
	Actions      []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"actions"`
	Conversions []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"conversions"`
	PurchaseROAS []struct {
		ActionType string `json:"action_type"`
		Value      string `json:"value"`
	} `json:"purchase_roas"`
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
	default:
		return entity.CampaignStatusDraft
	}
}

func (c *Connector) mapObjective(objective string) entity.CampaignObjective {
	switch strings.ToUpper(objective) {
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
	default:
		return entity.ObjectiveConversions
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

// Ensure Connector implements PlatformConnector interface
var _ service.PlatformConnector = (*Connector)(nil)
