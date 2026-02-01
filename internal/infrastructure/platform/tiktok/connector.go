package tiktok

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
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
	// TikTok API endpoints - Production
	productionBaseURL = "https://business-api.tiktok.com/open_api"
	productionAuthURL = "https://ads.tiktok.com/marketing_api/auth"

	// TikTok API endpoints - Sandbox
	sandboxBaseURL = "https://sandbox-ads.tiktok.com/open_api"
	sandboxAuthURL = "https://sandbox-ads.tiktok.com/marketing_api/auth"

	// Token endpoints (same for both environments)
	tokenURL   = "https://business-api.tiktok.com/open_api/v1.3/oauth2/access_token/"
	refreshURL = "https://business-api.tiktok.com/open_api/v1.3/oauth2/refresh_token/"

	// API version
	apiVersion = "v1.3"

	// TikTok error codes
	TikTokErrorCodeSuccess            = 0
	TikTokErrorCodeInvalidToken       = 40001
	TikTokErrorCodeTokenExpired       = 40002
	TikTokErrorCodePermissionDenied   = 40003
	TikTokErrorCodeRateLimit          = 40100
	TikTokErrorCodeInvalidParam       = 40002
	TikTokErrorCodeResourceNotFound   = 40004
	TikTokErrorCodeServerError        = 50000
	TikTokErrorCodeServiceUnavailable = 50001

	// Data levels for reports
	DataLevelAuctionCampaign   = "AUCTION_CAMPAIGN"
	DataLevelAuctionAdGroup    = "AUCTION_ADGROUP"
	DataLevelAuctionAd         = "AUCTION_AD"
	DataLevelReservationCampaign = "RESERVATION_CAMPAIGN"
	DataLevelReservationAdGroup  = "RESERVATION_ADGROUP"
	DataLevelReservationAd       = "RESERVATION_AD"
)

// Environment represents the TikTok API environment
type Environment string

const (
	EnvironmentProduction Environment = "production"
	EnvironmentSandbox    Environment = "sandbox"
)

// Connector implements the PlatformConnector interface for TikTok Ads
type Connector struct {
	*platform.BaseConnector
	config      *Config
	environment Environment
	baseURL     string
	authURL     string
}

// Config holds TikTok-specific configuration
type Config struct {
	AppID           string
	AppSecret       string
	RedirectURI     string
	Environment     Environment // "production" or "sandbox"
	RateLimitCalls  int
	RateLimitWindow time.Duration
	Timeout         time.Duration
	MaxRetries      int
}

// DefaultConfig returns default TikTok connector configuration
func DefaultConfig() *Config {
	return &Config{
		Environment:     EnvironmentProduction,
		RateLimitCalls:  10,
		RateLimitWindow: time.Second,
		Timeout:         30 * time.Second,
		MaxRetries:      3,
	}
}

// NewConnector creates a new TikTok Ads connector
func NewConnector(config *Config) *Connector {
	if config == nil {
		config = DefaultConfig()
	}

	// Determine URLs based on environment
	baseURL := productionBaseURL
	authURL := productionAuthURL
	if config.Environment == EnvironmentSandbox {
		baseURL = sandboxBaseURL
		authURL = sandboxAuthURL
	}

	baseConfig := &platform.ConnectorConfig{
		AppID:           config.AppID,
		AppSecret:       config.AppSecret,
		RedirectURI:     config.RedirectURI,
		BaseURL:         baseURL,
		RateLimitCalls:  config.RateLimitCalls,
		RateLimitWindow: config.RateLimitWindow,
		Timeout:         config.Timeout,
		MaxRetries:      config.MaxRetries,
	}

	return &Connector{
		BaseConnector: platform.NewBaseConnector(entity.PlatformTikTok, baseConfig),
		config:        config,
		environment:   config.Environment,
		baseURL:       baseURL,
		authURL:       authURL,
	}
}

// IsSandbox returns true if using sandbox environment
func (c *Connector) IsSandbox() bool {
	return c.environment == EnvironmentSandbox
}

// ============================================================================
// OAuth Methods
// ============================================================================

// GetAuthURL generates the OAuth authorization URL
func (c *Connector) GetAuthURL(state string) string {
	params := url.Values{
		"app_id":       {c.config.AppID},
		"redirect_uri": {c.config.RedirectURI},
		"state":        {state},
	}
	return c.authURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for OAuth tokens
func (c *Connector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	body := map[string]string{
		"app_id":     c.config.AppID,
		"secret":     c.config.AppSecret,
		"auth_code":  code,
		"grant_type": "authorization_code",
	}

	resp, err := c.DoPost(ctx, tokenURL, nil, body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken           string `json:"access_token"`
			RefreshToken          string `json:"refresh_token"`
			AccessTokenExpiresIn  int    `json:"access_token_expires_in"`
			RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
			Scope                 string `json:"scope"`
		} `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Code != 0 {
		return nil, errors.NewPlatformAPIError(
			entity.PlatformTikTok.String(),
			tokenResp.Code,
			fmt.Sprintf("%d", tokenResp.Code),
			tokenResp.Message,
		)
	}

	var scopes []string
	if tokenResp.Data.Scope != "" {
		scopes = strings.Split(tokenResp.Data.Scope, ",")
	}

	return &entity.OAuthToken{
		AccessToken:  tokenResp.Data.AccessToken,
		RefreshToken: tokenResp.Data.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.Data.AccessTokenExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.Data.AccessTokenExpiresIn) * time.Second),
		Scopes:       scopes,
	}, nil
}

// RefreshToken refreshes an expired access token
func (c *Connector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	body := map[string]string{
		"app_id":        c.config.AppID,
		"secret":        c.config.AppSecret,
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	resp, err := c.DoPost(ctx, refreshURL, nil, body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			AccessToken           string `json:"access_token"`
			RefreshToken          string `json:"refresh_token"`
			AccessTokenExpiresIn  int    `json:"access_token_expires_in"`
			RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		} `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Code != 0 {
		return nil, errors.NewPlatformAPIError(
			entity.PlatformTikTok.String(),
			tokenResp.Code,
			fmt.Sprintf("%d", tokenResp.Code),
			tokenResp.Message,
		)
	}

	return &entity.OAuthToken{
		AccessToken:  tokenResp.Data.AccessToken,
		RefreshToken: tokenResp.Data.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.Data.AccessTokenExpiresIn,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResp.Data.AccessTokenExpiresIn) * time.Second),
	}, nil
}

// RevokeToken revokes an access token
func (c *Connector) RevokeToken(ctx context.Context, accessToken string) error {
	// TikTok doesn't have a standard token revocation endpoint
	// Tokens expire automatically
	return nil
}

// ============================================================================
// User & Account Methods
// ============================================================================

// GetUserInfo retrieves the authenticated user's information
func (c *Connector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	endpoint := fmt.Sprintf("%s/%s/user/info/", c.baseURL, apiVersion)

	resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), nil)
	if err != nil {
		return nil, err
	}

	var userResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			Email       string `json:"email"`
		} `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &userResp); err != nil {
		return nil, err
	}

	if userResp.Code != 0 {
		return nil, errors.NewPlatformAPIError(
			entity.PlatformTikTok.String(),
			userResp.Code,
			fmt.Sprintf("%d", userResp.Code),
			userResp.Message,
		)
	}

	return &entity.PlatformUser{
		ID:    userResp.Data.ID,
		Name:  userResp.Data.DisplayName,
		Email: userResp.Data.Email,
	}, nil
}

// GetAdAccounts retrieves all ad accounts accessible by the token
func (c *Connector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	endpoint := fmt.Sprintf("%s/%s/advertiser/info/", c.baseURL, apiVersion)

	var allAccounts []entity.PlatformAccount
	page := 1
	pageSize := 100

	for {
		params := map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": fmt.Sprintf("%d", pageSize),
		}

		resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), params)
		if err != nil {
			return allAccounts, err
		}

		var accResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				List []struct {
					AdvertiserID   string `json:"advertiser_id"`
					AdvertiserName string `json:"advertiser_name"`
					Currency       string `json:"currency"`
					Timezone       string `json:"timezone"`
					Status         string `json:"status"`
				} `json:"list"`
				PageInfo struct {
					Page       int `json:"page"`
					PageSize   int `json:"page_size"`
					TotalPage  int `json:"total_page"`
					TotalCount int `json:"total_count"`
				} `json:"page_info"`
			} `json:"data"`
		}

		if err := c.ParseJSON(resp.Body, &accResp); err != nil {
			return allAccounts, err
		}

		if accResp.Code != 0 {
			return allAccounts, errors.NewPlatformAPIError(
				entity.PlatformTikTok.String(),
				accResp.Code,
				fmt.Sprintf("%d", accResp.Code),
				accResp.Message,
			)
		}

		for _, acc := range accResp.Data.List {
			allAccounts = append(allAccounts, entity.PlatformAccount{
				ID:       acc.AdvertiserID,
				Name:     acc.AdvertiserName,
				Currency: acc.Currency,
				Timezone: acc.Timezone,
				Status:   acc.Status,
			})
		}

		if page >= accResp.Data.PageInfo.TotalPage {
			break
		}
		page++
	}

	return allAccounts, nil
}

// ============================================================================
// Campaign Methods
// ============================================================================

// GetCampaigns retrieves all campaigns for an ad account
func (c *Connector) GetCampaigns(ctx context.Context, accessToken string, adAccountID string) ([]entity.Campaign, error) {
	endpoint := fmt.Sprintf("%s/%s/campaign/get/", c.baseURL, apiVersion)

	var allCampaigns []entity.Campaign
	page := 1
	pageSize := 100

	for {
		params := map[string]string{
			"advertiser_id": adAccountID,
			"page":          fmt.Sprintf("%d", page),
			"page_size":     fmt.Sprintf("%d", pageSize),
		}

		resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), params)
		if err != nil {
			return allCampaigns, err
		}

		var campResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				List []struct {
					CampaignID    string `json:"campaign_id"`
					CampaignName  string `json:"campaign_name"`
					ObjectiveType string `json:"objective_type"`
					Status        string `json:"status"`
					Budget        string `json:"budget"`
					BudgetMode    string `json:"budget_mode"`
					CreateTime    string `json:"create_time"`
					ModifyTime    string `json:"modify_time"`
				} `json:"list"`
				PageInfo struct {
					Page       int `json:"page"`
					PageSize   int `json:"page_size"`
					TotalPage  int `json:"total_page"`
					TotalCount int `json:"total_count"`
				} `json:"page_info"`
			} `json:"data"`
		}

		if err := c.ParseJSON(resp.Body, &campResp); err != nil {
			return allCampaigns, err
		}

		if campResp.Code != 0 {
			return allCampaigns, errors.NewPlatformAPIError(
				entity.PlatformTikTok.String(),
				campResp.Code,
				fmt.Sprintf("%d", campResp.Code),
				campResp.Message,
			)
		}

		for _, camp := range campResp.Data.List {
			campaign := entity.Campaign{
				Platform:             entity.PlatformTikTok,
				PlatformCampaignID:   camp.CampaignID,
				PlatformCampaignName: camp.CampaignName,
				Status:               c.mapCampaignStatus(camp.Status),
				Objective:            c.mapObjective(camp.ObjectiveType),
			}

			if camp.Budget != "" {
				budget, _ := decimal.NewFromString(camp.Budget)
				if camp.BudgetMode == "BUDGET_MODE_DAY" {
					campaign.DailyBudget = &budget
				} else {
					campaign.LifetimeBudget = &budget
				}
			}

			if camp.CreateTime != "" {
				if t, err := time.Parse("2006-01-02 15:04:05", camp.CreateTime); err == nil {
					campaign.PlatformCreatedAt = &t
				}
			}

			if camp.ModifyTime != "" {
				if t, err := time.Parse("2006-01-02 15:04:05", camp.ModifyTime); err == nil {
					campaign.PlatformUpdatedAt = &t
				}
			}

			allCampaigns = append(allCampaigns, campaign)
		}

		if page >= campResp.Data.PageInfo.TotalPage {
			break
		}
		page++
	}

	return allCampaigns, nil
}

// GetCampaign retrieves a single campaign by ID
func (c *Connector) GetCampaign(ctx context.Context, accessToken string, campaignID string) (*entity.Campaign, error) {
	campaigns, err := c.GetCampaigns(ctx, accessToken, campaignID)
	if err != nil {
		return nil, err
	}

	for _, camp := range campaigns {
		if camp.PlatformCampaignID == campaignID {
			return &camp, nil
		}
	}

	return nil, errors.ErrNotFound("Campaign")
}

// ============================================================================
// AdSet (AdGroup) Methods
// ============================================================================

// GetAdSets retrieves all ad sets for a campaign
func (c *Connector) GetAdSets(ctx context.Context, accessToken string, campaignID string) ([]entity.AdSet, error) {
	endpoint := fmt.Sprintf("%s/%s/adgroup/get/", c.baseURL, apiVersion)

	var allAdSets []entity.AdSet
	page := 1
	pageSize := 100

	// Note: TikTok requires advertiser_id, not campaign_id directly
	// This is simplified - in production, you'd need to look up the advertiser_id
	for {
		filteringJSON, _ := json.Marshal(map[string]interface{}{
			"campaign_ids": []string{campaignID},
		})

		params := map[string]string{
			"advertiser_id": "", // Need to be passed from caller
			"page":          fmt.Sprintf("%d", page),
			"page_size":     fmt.Sprintf("%d", pageSize),
			"filtering":     string(filteringJSON),
		}

		resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), params)
		if err != nil {
			return allAdSets, err
		}

		var adgroupResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				List []struct {
					AdgroupID        string  `json:"adgroup_id"`
					AdgroupName      string  `json:"adgroup_name"`
					Status           string  `json:"status"`
					Budget           float64 `json:"budget"`
					BudgetMode       string  `json:"budget_mode"`
					BidType          string  `json:"bid_type"`
					Bid              float64 `json:"bid"`
					OptimizationGoal string  `json:"optimization_goal"`
				} `json:"list"`
				PageInfo struct {
					Page      int `json:"page"`
					PageSize  int `json:"page_size"`
					TotalPage int `json:"total_page"`
				} `json:"page_info"`
			} `json:"data"`
		}

		if err := c.ParseJSON(resp.Body, &adgroupResp); err != nil {
			return allAdSets, err
		}

		if adgroupResp.Code != 0 {
			break // No more data or error
		}

		for _, ag := range adgroupResp.Data.List {
			adSet := entity.AdSet{
				Platform:          entity.PlatformTikTok,
				PlatformAdSetID:   ag.AdgroupID,
				PlatformAdSetName: ag.AdgroupName,
				Status:            c.mapCampaignStatus(ag.Status),
				BidStrategy:       ag.OptimizationGoal,
			}

			budget := decimal.NewFromFloat(ag.Budget)
			if ag.BudgetMode == "BUDGET_MODE_DAY" {
				adSet.DailyBudget = &budget
			} else {
				adSet.LifetimeBudget = &budget
			}

			bid := decimal.NewFromFloat(ag.Bid)
			adSet.BidAmount = &bid

			allAdSets = append(allAdSets, adSet)
		}

		if page >= adgroupResp.Data.PageInfo.TotalPage {
			break
		}
		page++
	}

	return allAdSets, nil
}

// GetAdSet retrieves a single ad set by ID
func (c *Connector) GetAdSet(ctx context.Context, accessToken string, adSetID string) (*entity.AdSet, error) {
	adSets, err := c.GetAdSets(ctx, accessToken, adSetID)
	if err != nil {
		return nil, err
	}

	for _, as := range adSets {
		if as.PlatformAdSetID == adSetID {
			return &as, nil
		}
	}

	return nil, errors.ErrNotFound("AdSet")
}

// ============================================================================
// Ad Methods
// ============================================================================

// GetAds retrieves all ads for an ad set
func (c *Connector) GetAds(ctx context.Context, accessToken string, adSetID string) ([]entity.Ad, error) {
	endpoint := fmt.Sprintf("%s/%s/ad/get/", c.baseURL, apiVersion)

	var allAds []entity.Ad
	page := 1
	pageSize := 100

	for {
		filteringJSON, _ := json.Marshal(map[string]interface{}{
			"adgroup_ids": []string{adSetID},
		})

		params := map[string]string{
			"advertiser_id": "",
			"page":          fmt.Sprintf("%d", page),
			"page_size":     fmt.Sprintf("%d", pageSize),
			"filtering":     string(filteringJSON),
		}

		resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), params)
		if err != nil {
			return allAds, err
		}

		var adResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				List []struct {
					AdID           string   `json:"ad_id"`
					AdName         string   `json:"ad_name"`
					Status         string   `json:"status"`
					AdText         string   `json:"ad_text"`
					CallToAction   string   `json:"call_to_action"`
					LandingPageURL string   `json:"landing_page_url"`
					ImageIDs       []string `json:"image_ids"`
					VideoID        string   `json:"video_id"`
				} `json:"list"`
				PageInfo struct {
					Page      int `json:"page"`
					PageSize  int `json:"page_size"`
					TotalPage int `json:"total_page"`
				} `json:"page_info"`
			} `json:"data"`
		}

		if err := c.ParseJSON(resp.Body, &adResp); err != nil {
			return allAds, err
		}

		if adResp.Code != 0 {
			break
		}

		for _, a := range adResp.Data.List {
			ad := entity.Ad{
				Platform:       entity.PlatformTikTok,
				PlatformAdID:   a.AdID,
				PlatformAdName: a.AdName,
				Status:         c.mapCampaignStatus(a.Status),
				Description:    a.AdText,
				CallToAction:   a.CallToAction,
				DestinationURL: a.LandingPageURL,
				CreativeData: entity.JSONMap{
					"video_id":  a.VideoID,
					"image_ids": a.ImageIDs,
				},
			}
			allAds = append(allAds, ad)
		}

		if page >= adResp.Data.PageInfo.TotalPage {
			break
		}
		page++
	}

	return allAds, nil
}

// GetAd retrieves a single ad by ID
func (c *Connector) GetAd(ctx context.Context, accessToken string, adID string) (*entity.Ad, error) {
	ads, err := c.GetAds(ctx, accessToken, adID)
	if err != nil {
		return nil, err
	}

	for _, a := range ads {
		if a.PlatformAdID == adID {
			return &a, nil
		}
	}

	return nil, errors.ErrNotFound("Ad")
}

// ============================================================================
// Insights Methods
// ============================================================================

// GetCampaignInsights retrieves performance metrics for a campaign
func (c *Connector) GetCampaignInsights(ctx context.Context, accessToken string, campaignID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	endpoint := fmt.Sprintf("%s/%s/report/integrated/get/", c.baseURL, apiVersion)

	dataLevel := "AUCTION_CAMPAIGN"
	dimensions := []string{"campaign_id", "stat_time_day"}
	metrics := []string{
		"spend", "impressions", "reach", "clicks",
		"conversion", "cost_per_conversion",
		"ctr", "cpc", "cpm",
		"video_views", "video_play_actions",
	}

	dimensionsJSON, _ := json.Marshal(dimensions)
	metricsJSON, _ := json.Marshal(metrics)

	params := map[string]string{
		"advertiser_id": "",
		"data_level":    dataLevel,
		"dimensions":    string(dimensionsJSON),
		"metrics":       string(metricsJSON),
		"start_date":    dateRange.StartDate.Format("2006-01-02"),
		"end_date":      dateRange.EndDate.Format("2006-01-02"),
		"page":          "1",
		"page_size":     "1000",
	}

	resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), params)
	if err != nil {
		return nil, err
	}

	var reportResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				Dimensions struct {
					CampaignID  string `json:"campaign_id"`
					StatTimeDay string `json:"stat_time_day"`
				} `json:"dimensions"`
				Metrics struct {
					Spend             string `json:"spend"`
					Impressions       string `json:"impressions"`
					Reach             string `json:"reach"`
					Clicks            string `json:"clicks"`
					Conversion        string `json:"conversion"`
					CostPerConversion string `json:"cost_per_conversion"`
					CTR               string `json:"ctr"`
					CPC               string `json:"cpc"`
					CPM               string `json:"cpm"`
					VideoViews        string `json:"video_views"`
					VideoPlayActions  string `json:"video_play_actions"`
				} `json:"metrics"`
			} `json:"list"`
		} `json:"data"`
	}

	if err := c.ParseJSON(resp.Body, &reportResp); err != nil {
		return nil, err
	}

	var metrics_list []entity.CampaignMetricsDaily
	for _, item := range reportResp.Data.List {
		metricDate, _ := time.Parse("2006-01-02", item.Dimensions.StatTimeDay)

		metric := entity.CampaignMetricsDaily{
			BaseEntity: entity.BaseEntity{ID: uuid.New()},
			Platform:   entity.PlatformTikTok,
			MetricDate: metricDate,
			Currency:   "USD",
		}

		metric.Spend = c.parseDecimal(item.Metrics.Spend)
		metric.Impressions = c.parseInt64(item.Metrics.Impressions)
		metric.Reach = c.parseInt64(item.Metrics.Reach)
		metric.Clicks = c.parseInt64(item.Metrics.Clicks)
		metric.Conversions = c.parseInt64(item.Metrics.Conversion)
		metric.VideoViews = c.parseInt64(item.Metrics.VideoViews)

		ctr := c.parseFloat(item.Metrics.CTR)
		metric.CTR = &ctr

		cpc := c.parseDecimal(item.Metrics.CPC)
		metric.CPC = &cpc

		cpm := c.parseDecimal(item.Metrics.CPM)
		metric.CPM = &cpm

		cpa := c.parseDecimal(item.Metrics.CostPerConversion)
		metric.CPA = &cpa

		metrics_list = append(metrics_list, metric)
	}

	return metrics_list, nil
}

// GetAdSetInsights retrieves performance metrics for an ad set
func (c *Connector) GetAdSetInsights(ctx context.Context, accessToken string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	// Similar implementation to GetCampaignInsights but with adgroup level
	return nil, nil // Placeholder
}

// GetAdInsights retrieves performance metrics for an ad
func (c *Connector) GetAdInsights(ctx context.Context, accessToken string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	// Similar implementation to GetCampaignInsights but with ad level
	return nil, nil // Placeholder
}

// GetAccountInsights retrieves aggregated insights for an ad account
func (c *Connector) GetAccountInsights(ctx context.Context, accessToken string, adAccountID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	// Aggregate campaign insights
	return nil, nil // Placeholder
}

// HealthCheck verifies the connector can connect to TikTok API
func (c *Connector) HealthCheck(ctx context.Context) error {
	// TikTok doesn't have a simple health check endpoint
	// We could try to fetch user info with an invalid token to check connectivity
	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (c *Connector) buildTikTokHeaders(accessToken string) map[string]string {
	return map[string]string{
		"Access-Token": accessToken,
		"Content-Type": "application/json",
	}
}

// ============================================================================
// Request Signing Methods
// ============================================================================

// generateSignature generates the HMAC-SHA256 signature for TikTok API
// TikTok requires: timestamp + access_token + secret
func (c *Connector) generateSignature(params map[string]string) string {
	// Sort parameters by key
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build signature string
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString(params[k])
	}

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.config.AppSecret))
	h.Write([]byte(signStr.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// SignRequest signs a request with timestamp, access token, and secret
func (c *Connector) SignRequest(accessToken string, params map[string]string) map[string]string {
	// Add timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Create params map for signing
	signParams := make(map[string]string)
	for k, v := range params {
		signParams[k] = v
	}
	signParams["timestamp"] = timestamp
	signParams["access_token"] = accessToken

	// Generate signature
	signature := c.generateSignature(signParams)

	// Add signature and timestamp to original params
	params["timestamp"] = timestamp
	params["sign"] = signature

	return params
}

// BuildSignedHeaders builds headers with signature for TikTok API
func (c *Connector) BuildSignedHeaders(accessToken string, timestamp string) map[string]string {
	return map[string]string{
		"Access-Token": accessToken,
		"Content-Type": "application/json",
		"X-Timestamp":  timestamp,
	}
}

func (c *Connector) mapCampaignStatus(status string) entity.CampaignStatus {
	switch strings.ToUpper(status) {
	case "ENABLE", "STATUS_ENABLE":
		return entity.CampaignStatusActive
	case "DISABLE", "STATUS_DISABLE":
		return entity.CampaignStatusPaused
	case "DELETE", "STATUS_DELETE":
		return entity.CampaignStatusDeleted
	default:
		return entity.CampaignStatusDraft
	}
}

func (c *Connector) mapObjective(objective string) entity.CampaignObjective {
	switch strings.ToUpper(objective) {
	case "REACH":
		return entity.ObjectiveAwareness
	case "TRAFFIC":
		return entity.ObjectiveTraffic
	case "VIDEO_VIEWS":
		return entity.ObjectiveVideoViews
	case "LEAD_GENERATION":
		return entity.ObjectiveLeads
	case "APP_PROMOTION", "APP_INSTALL":
		return entity.ObjectiveAppPromotion
	case "CONVERSIONS", "PRODUCT_SALES":
		return entity.ObjectiveSales
	default:
		return entity.ObjectiveConversions
	}
}

func (c *Connector) parseInt64(s string) int64 {
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val
}

func (c *Connector) parseFloat(s string) float64 {
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val
}

func (c *Connector) parseDecimal(s string) decimal.Decimal {
	val, _ := decimal.NewFromString(s)
	return val
}

// ============================================================================
// Error Handling
// ============================================================================

// TikTokError represents a TikTok API error response
type TikTokError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// HandleTikTokError handles TikTok-specific error codes and returns appropriate error
func (c *Connector) HandleTikTokError(code int, message string) error {
	switch code {
	case TikTokErrorCodeSuccess:
		return nil

	case TikTokErrorCodeInvalidToken:
		return errors.NewTokenInvalidError("tiktok", "access")

	case TikTokErrorCodeTokenExpired:
		return errors.NewTokenExpiredError("tiktok", "access", time.Now())

	case TikTokErrorCodePermissionDenied:
		return errors.NewPlatformAPIError("tiktok", 403, fmt.Sprintf("%d", code), "Permission denied: "+message)

	case TikTokErrorCodeRateLimit:
		return errors.NewRateLimitError("tiktok", 60*time.Second)

	case TikTokErrorCodeResourceNotFound:
		return errors.ErrNotFound("resource")

	case TikTokErrorCodeServerError, TikTokErrorCodeServiceUnavailable:
		return errors.NewPlatformAPIError("tiktok", 500, fmt.Sprintf("%d", code), "Server error: "+message)

	default:
		return errors.NewPlatformAPIError("tiktok", 400, fmt.Sprintf("%d", code), message)
	}
}

// ============================================================================
// Enhanced Report Methods with AUCTION vs RESERVATION
// ============================================================================

// ReportParams holds parameters for fetching TikTok reports
type ReportParams struct {
	AdvertiserID  string
	DateRange     entity.DateRange
	DataLevel     string   // AUCTION_CAMPAIGN, AUCTION_ADGROUP, etc.
	Dimensions    []string
	Metrics       []string
	Filtering     map[string]interface{}
	OrderField    string
	OrderType     string // ASC, DESC
	Page          int
	PageSize      int
}

// DefaultReportMetrics returns the default metrics for TikTok reports
func DefaultReportMetrics() []string {
	return []string{
		"spend", "impressions", "reach", "clicks",
		"conversion", "cost_per_conversion", "conversion_rate",
		"ctr", "cpc", "cpm",
		"video_views", "video_play_actions",
		"average_video_play", "average_video_play_per_user",
		"likes", "comments", "shares", "follows",
		"profile_visits", "profile_visits_rate",
	}
}

// GetIntegratedReport fetches reports with full parameter control
func (c *Connector) GetIntegratedReport(ctx context.Context, accessToken string, params ReportParams) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/%s/report/integrated/get/", c.baseURL, apiVersion)

	// Build dimensions and metrics JSON
	dimensionsJSON, _ := json.Marshal(params.Dimensions)
	metricsJSON, _ := json.Marshal(params.Metrics)

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 1000
	}

	queryParams := map[string]string{
		"advertiser_id": params.AdvertiserID,
		"data_level":    params.DataLevel,
		"dimensions":    string(dimensionsJSON),
		"metrics":       string(metricsJSON),
		"start_date":    params.DateRange.StartDate.Format("2006-01-02"),
		"end_date":      params.DateRange.EndDate.Format("2006-01-02"),
		"page":          fmt.Sprintf("%d", params.Page),
		"page_size":     fmt.Sprintf("%d", params.PageSize),
	}

	// Add filtering if provided
	if len(params.Filtering) > 0 {
		filteringJSON, _ := json.Marshal(params.Filtering)
		queryParams["filtering"] = string(filteringJSON)
	}

	// Add ordering if provided
	if params.OrderField != "" {
		queryParams["order_field"] = params.OrderField
		if params.OrderType != "" {
			queryParams["order_type"] = params.OrderType
		}
	}

	var allResults []map[string]interface{}
	page := params.Page

	for {
		queryParams["page"] = fmt.Sprintf("%d", page)

		resp, err := c.DoGet(ctx, endpoint, c.buildTikTokHeaders(accessToken), queryParams)
		if err != nil {
			return allResults, err
		}

		var reportResp struct {
			Code      int    `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
			Data      struct {
				List     []map[string]interface{} `json:"list"`
				PageInfo struct {
					Page       int `json:"page"`
					PageSize   int `json:"page_size"`
					TotalPage  int `json:"total_page"`
					TotalCount int `json:"total_count"`
				} `json:"page_info"`
			} `json:"data"`
		}

		if err := c.ParseJSON(resp.Body, &reportResp); err != nil {
			return allResults, err
		}

		if reportResp.Code != TikTokErrorCodeSuccess {
			return allResults, c.HandleTikTokError(reportResp.Code, reportResp.Message)
		}

		allResults = append(allResults, reportResp.Data.List...)

		if page >= reportResp.Data.PageInfo.TotalPage {
			break
		}
		page++
	}

	return allResults, nil
}

// GetAuctionCampaignReport fetches AUCTION campaign reports
func (c *Connector) GetAuctionCampaignReport(ctx context.Context, accessToken string, advertiserID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	params := ReportParams{
		AdvertiserID: advertiserID,
		DateRange:    dateRange,
		DataLevel:    DataLevelAuctionCampaign,
		Dimensions:   []string{"campaign_id", "stat_time_day"},
		Metrics:      DefaultReportMetrics(),
	}

	results, err := c.GetIntegratedReport(ctx, accessToken, params)
	if err != nil {
		return nil, err
	}

	return c.mapCampaignReportResults(results), nil
}

// GetReservationCampaignReport fetches RESERVATION campaign reports
func (c *Connector) GetReservationCampaignReport(ctx context.Context, accessToken string, advertiserID string, dateRange entity.DateRange) ([]entity.CampaignMetricsDaily, error) {
	params := ReportParams{
		AdvertiserID: advertiserID,
		DateRange:    dateRange,
		DataLevel:    DataLevelReservationCampaign,
		Dimensions:   []string{"campaign_id", "stat_time_day"},
		Metrics:      DefaultReportMetrics(),
	}

	results, err := c.GetIntegratedReport(ctx, accessToken, params)
	if err != nil {
		return nil, err
	}

	return c.mapCampaignReportResults(results), nil
}

// mapCampaignReportResults maps raw report results to CampaignMetricsDaily
func (c *Connector) mapCampaignReportResults(results []map[string]interface{}) []entity.CampaignMetricsDaily {
	var metrics []entity.CampaignMetricsDaily

	for _, item := range results {
		dims, ok := item["dimensions"].(map[string]interface{})
		if !ok {
			continue
		}
		mets, ok := item["metrics"].(map[string]interface{})
		if !ok {
			continue
		}

		dateStr, _ := dims["stat_time_day"].(string)
		metricDate, _ := time.Parse("2006-01-02", dateStr)

		metric := entity.CampaignMetricsDaily{
			BaseEntity: entity.BaseEntity{ID: uuid.New()},
			Platform:   entity.PlatformTikTok,
			MetricDate: metricDate,
			Currency:   "USD",
		}

		// Map metrics
		metric.Spend = c.parseDecimalFromInterface(mets["spend"])
		metric.Impressions = c.parseInt64FromInterface(mets["impressions"])
		metric.Reach = c.parseInt64FromInterface(mets["reach"])
		metric.Clicks = c.parseInt64FromInterface(mets["clicks"])
		metric.Conversions = c.parseInt64FromInterface(mets["conversion"])
		metric.VideoViews = c.parseInt64FromInterface(mets["video_views"])
		metric.Likes = c.parseInt64FromInterface(mets["likes"])
		metric.Comments = c.parseInt64FromInterface(mets["comments"])
		metric.Shares = c.parseInt64FromInterface(mets["shares"])

		// Map calculated metrics
		ctr := c.parseFloatFromInterface(mets["ctr"])
		metric.CTR = &ctr

		cpc := c.parseDecimalFromInterface(mets["cpc"])
		metric.CPC = &cpc

		cpm := c.parseDecimalFromInterface(mets["cpm"])
		metric.CPM = &cpm

		cpa := c.parseDecimalFromInterface(mets["cost_per_conversion"])
		metric.CPA = &cpa

		metrics = append(metrics, metric)
	}

	return metrics
}

// parseDecimalFromInterface parses decimal from interface{}
func (c *Connector) parseDecimalFromInterface(v interface{}) decimal.Decimal {
	switch val := v.(type) {
	case string:
		d, _ := decimal.NewFromString(val)
		return d
	case float64:
		return decimal.NewFromFloat(val)
	case int:
		return decimal.NewFromInt(int64(val))
	default:
		return decimal.Zero
	}
}

// parseInt64FromInterface parses int64 from interface{}
func (c *Connector) parseInt64FromInterface(v interface{}) int64 {
	switch val := v.(type) {
	case string:
		var i int64
		fmt.Sscanf(val, "%d", &i)
		return i
	case float64:
		return int64(val)
	case int:
		return int64(val)
	default:
		return 0
	}
}

// parseFloatFromInterface parses float64 from interface{}
func (c *Connector) parseFloatFromInterface(v interface{}) float64 {
	switch val := v.(type) {
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case float64:
		return val
	case int:
		return float64(val)
	default:
		return 0
	}
}

// ============================================================================
// Complete AdSet and Ad Insights Methods
// ============================================================================

// GetAdSetInsightsComplete retrieves complete performance metrics for an ad set
func (c *Connector) GetAdSetInsightsComplete(ctx context.Context, accessToken string, advertiserID string, adSetID string, dateRange entity.DateRange) ([]entity.AdSetMetricsDaily, error) {
	params := ReportParams{
		AdvertiserID: advertiserID,
		DateRange:    dateRange,
		DataLevel:    DataLevelAuctionAdGroup,
		Dimensions:   []string{"adgroup_id", "stat_time_day"},
		Metrics:      DefaultReportMetrics(),
		Filtering: map[string]interface{}{
			"adgroup_ids": []string{adSetID},
		},
	}

	results, err := c.GetIntegratedReport(ctx, accessToken, params)
	if err != nil {
		return nil, err
	}

	return c.mapAdSetReportResults(results), nil
}

// mapAdSetReportResults maps raw report results to AdSetMetricsDaily
func (c *Connector) mapAdSetReportResults(results []map[string]interface{}) []entity.AdSetMetricsDaily {
	var metrics []entity.AdSetMetricsDaily

	for _, item := range results {
		dims, ok := item["dimensions"].(map[string]interface{})
		if !ok {
			continue
		}
		mets, ok := item["metrics"].(map[string]interface{})
		if !ok {
			continue
		}

		dateStr, _ := dims["stat_time_day"].(string)
		metricDate, _ := time.Parse("2006-01-02", dateStr)

		metric := entity.AdSetMetricsDaily{
			BaseEntity: entity.BaseEntity{ID: uuid.New()},
			Platform:   entity.PlatformTikTok,
			MetricDate: metricDate,
			Currency:   "USD",
		}

		metric.Spend = c.parseDecimalFromInterface(mets["spend"])
		metric.Impressions = c.parseInt64FromInterface(mets["impressions"])
		metric.Reach = c.parseInt64FromInterface(mets["reach"])
		metric.Clicks = c.parseInt64FromInterface(mets["clicks"])
		metric.UniqueClicks = c.parseInt64FromInterface(mets["unique_clicks"])

		metrics = append(metrics, metric)
	}

	return metrics
}

// GetAdInsightsComplete retrieves complete performance metrics for an ad
func (c *Connector) GetAdInsightsComplete(ctx context.Context, accessToken string, advertiserID string, adID string, dateRange entity.DateRange) ([]entity.AdMetricsDaily, error) {
	params := ReportParams{
		AdvertiserID: advertiserID,
		DateRange:    dateRange,
		DataLevel:    DataLevelAuctionAd,
		Dimensions:   []string{"ad_id", "stat_time_day"},
		Metrics:      DefaultReportMetrics(),
		Filtering: map[string]interface{}{
			"ad_ids": []string{adID},
		},
	}

	results, err := c.GetIntegratedReport(ctx, accessToken, params)
	if err != nil {
		return nil, err
	}

	return c.mapAdReportResults(results), nil
}

// mapAdReportResults maps raw report results to AdMetricsDaily
func (c *Connector) mapAdReportResults(results []map[string]interface{}) []entity.AdMetricsDaily {
	var metrics []entity.AdMetricsDaily

	for _, item := range results {
		dims, ok := item["dimensions"].(map[string]interface{})
		if !ok {
			continue
		}
		mets, ok := item["metrics"].(map[string]interface{})
		if !ok {
			continue
		}

		dateStr, _ := dims["stat_time_day"].(string)
		metricDate, _ := time.Parse("2006-01-02", dateStr)

		metric := entity.AdMetricsDaily{
			BaseEntity: entity.BaseEntity{ID: uuid.New()},
			Platform:   entity.PlatformTikTok,
			MetricDate: metricDate,
			Currency:   "USD",
		}

		metric.Spend = c.parseDecimalFromInterface(mets["spend"])
		metric.Impressions = c.parseInt64FromInterface(mets["impressions"])
		metric.Reach = c.parseInt64FromInterface(mets["reach"])
		metric.Clicks = c.parseInt64FromInterface(mets["clicks"])
		metric.UniqueClicks = c.parseInt64FromInterface(mets["unique_clicks"])

		metrics = append(metrics, metric)
	}

	return metrics
}

// GetAccountInsightsComplete retrieves aggregated insights for an ad account
func (c *Connector) GetAccountInsightsComplete(ctx context.Context, accessToken string, advertiserID string, dateRange entity.DateRange) (*entity.AggregatedMetrics, error) {
	params := ReportParams{
		AdvertiserID: advertiserID,
		DateRange:    dateRange,
		DataLevel:    DataLevelAuctionCampaign,
		Dimensions:   []string{"stat_time_day"},
		Metrics:      DefaultReportMetrics(),
	}

	results, err := c.GetIntegratedReport(ctx, accessToken, params)
	if err != nil {
		return nil, err
	}

	// Aggregate all results
	aggregated := &entity.AggregatedMetrics{
		Currency: "USD",
	}

	for _, item := range results {
		mets, ok := item["metrics"].(map[string]interface{})
		if !ok {
			continue
		}

		aggregated.TotalSpend = aggregated.TotalSpend.Add(c.parseDecimalFromInterface(mets["spend"]))
		aggregated.TotalImpressions += c.parseInt64FromInterface(mets["impressions"])
		aggregated.TotalClicks += c.parseInt64FromInterface(mets["clicks"])
		aggregated.TotalConversions += c.parseInt64FromInterface(mets["conversion"])
	}

	// Calculate derived metrics
	if aggregated.TotalImpressions > 0 {
		aggregated.AverageCTR = float64(aggregated.TotalClicks) / float64(aggregated.TotalImpressions) * 100
		aggregated.AverageCPM = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalImpressions)).Mul(decimal.NewFromInt(1000))
	}

	if aggregated.TotalClicks > 0 {
		aggregated.AverageCPC = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalClicks))
	}

	if aggregated.TotalConversions > 0 {
		aggregated.AverageCPA = aggregated.TotalSpend.Div(decimal.NewFromInt(aggregated.TotalConversions))
	}

	return aggregated, nil
}

// Ensure Connector implements PlatformConnector interface
var _ service.PlatformConnector = (*Connector)(nil)
