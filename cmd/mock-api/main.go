// =============================================================================
// Mock Platform API Server
// Simulates Meta, TikTok, and Shopee Marketing APIs for local testing
// =============================================================================

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MockConfig holds the mock server configuration
type MockConfig struct {
	Mode     string // "normal", "error", "slow", "rate_limited"
	DelayMs  int    // Artificial delay in milliseconds
	ErrorPct int    // Percentage of requests to fail (0-100)
	mu       sync.RWMutex
}

var config = &MockConfig{
	Mode:     "normal",
	DelayMs:  0,
	ErrorPct: 0,
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Set config from environment
	if mode := os.Getenv("MOCK_MODE"); mode != "" {
		config.Mode = mode
	}
	if delay := os.Getenv("MOCK_DELAY_MS"); delay != "" {
		if d, err := strconv.Atoi(delay); err == nil {
			config.DelayMs = d
		}
	}

	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mockMiddleware())

	// Health check
	r.GET("/health", healthCheck)

	// Mock control endpoints
	r.POST("/control/mode", setMode)
	r.GET("/control/mode", getMode)
	r.POST("/control/reset", resetConfig)

	// Meta Marketing API mocks
	meta := r.Group("/meta")
	{
		meta.GET("/v18.0/me/adaccounts", metaListAdAccounts)
		meta.GET("/v18.0/act_:account_id/campaigns", metaListCampaigns)
		meta.GET("/v18.0/act_:account_id/insights", metaGetInsights)
		meta.GET("/v18.0/campaigns/:campaign_id/insights", metaCampaignInsights)
		meta.POST("/oauth/access_token", metaRefreshToken)
	}

	// TikTok Marketing API mocks
	tiktok := r.Group("/tiktok")
	{
		tiktok.GET("/open_api/v1.3/oauth2/advertiser/get", tiktokListAdvertisers)
		tiktok.GET("/open_api/v1.3/campaign/get", tiktokListCampaigns)
		tiktok.GET("/open_api/v1.3/report/integrated/get", tiktokGetReport)
		tiktok.POST("/open_api/v1.3/oauth2/access_token/get", tiktokRefreshToken)
	}

	// Shopee API mocks
	shopee := r.Group("/shopee")
	{
		shopee.GET("/api/v2/shop/get_shop_info", shopeeGetShopInfo)
		shopee.GET("/api/v2/product/get_item_list", shopeeListProducts)
		shopee.GET("/api/v2/shop/performance", shopeeGetPerformance)
		shopee.POST("/api/v2/auth/access_token/get", shopeeRefreshToken)
	}

	port := os.Getenv("MOCK_PORT")
	if port == "" {
		port = "9090"
	}

	log.Printf("Mock Platform API Server starting on port %s", port)
	log.Printf("Mode: %s, Delay: %dms", config.Mode, config.DelayMs)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// =============================================================================
// Middleware
// =============================================================================

func mockMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		config.mu.RLock()
		mode := config.Mode
		delay := config.DelayMs
		errorPct := config.ErrorPct
		config.mu.RUnlock()

		// Skip middleware for control endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path[:8] == "/control" {
			c.Next()
			return
		}

		// Apply artificial delay
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// Random slow response (2-5 seconds) in slow mode
		if mode == "slow" {
			slowDelay := rand.Intn(3000) + 2000
			time.Sleep(time.Duration(slowDelay) * time.Millisecond)
		}

		// Random errors based on error percentage
		if errorPct > 0 && rand.Intn(100) < errorPct {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": "Random mock error",
					"code":    500,
				},
			})
			c.Abort()
			return
		}

		// Mode-specific behaviors
		switch mode {
		case "error":
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": "Mock server in error mode",
					"code":    500,
				},
			})
			c.Abort()
			return

		case "rate_limited":
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "(#17) User request limit reached",
					"code":    17,
					"type":    "OAuthException",
				},
			})
			c.Abort()
			return

		case "token_expired":
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"message": "Error validating access token: Session has expired",
					"code":    190,
					"type":    "OAuthException",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// =============================================================================
// Control Endpoints
// =============================================================================

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "mock-platform-api",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func setMode(c *gin.Context) {
	var req struct {
		Mode     string `json:"mode"`
		DelayMs  int    `json:"delay_ms"`
		ErrorPct int    `json:"error_pct"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.mu.Lock()
	if req.Mode != "" {
		config.Mode = req.Mode
	}
	if req.DelayMs >= 0 {
		config.DelayMs = req.DelayMs
	}
	if req.ErrorPct >= 0 && req.ErrorPct <= 100 {
		config.ErrorPct = req.ErrorPct
	}
	config.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message":   "Mock configuration updated",
		"mode":      config.Mode,
		"delay_ms":  config.DelayMs,
		"error_pct": config.ErrorPct,
	})
}

func getMode(c *gin.Context) {
	config.mu.RLock()
	defer config.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"mode":      config.Mode,
		"delay_ms":  config.DelayMs,
		"error_pct": config.ErrorPct,
	})
}

func resetConfig(c *gin.Context) {
	config.mu.Lock()
	config.Mode = "normal"
	config.DelayMs = 0
	config.ErrorPct = 0
	config.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "Mock configuration reset to normal",
	})
}

// =============================================================================
// Meta Marketing API Mocks
// =============================================================================

func metaListAdAccounts(c *gin.Context) {
	// Pagination
	afterCursor := c.Query("after")
	limit := 25
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	accounts := []map[string]interface{}{
		{
			"id":                   "act_123456789",
			"name":                 "Test Ad Account 1 - Business",
			"account_status":       1,
			"currency":             "MYR",
			"timezone_name":        "Asia/Kuala_Lumpur",
			"amount_spent":         "150000",
			"business_name":        "Test Business Sdn Bhd",
			"disable_reason":       0,
			"created_time":         "2023-01-15T08:30:00+0800",
			"funding_source_details": map[string]interface{}{
				"type": 1,
			},
		},
		{
			"id":                   "act_987654321",
			"name":                 "Test Ad Account 2 - Pro",
			"account_status":       1,
			"currency":             "MYR",
			"timezone_name":        "Asia/Kuala_Lumpur",
			"amount_spent":         "75000",
			"business_name":        "Pro Marketing Agency",
			"disable_reason":       0,
			"created_time":         "2023-03-20T10:00:00+0800",
			"funding_source_details": map[string]interface{}{
				"type": 1,
			},
		},
	}

	response := map[string]interface{}{
		"data": accounts,
	}

	// Add paging if needed
	if afterCursor == "" && len(accounts) >= limit {
		response["paging"] = map[string]interface{}{
			"cursors": map[string]string{
				"before": "QVFXYZ123",
				"after":  "QVFXYZ456",
			},
			"next": fmt.Sprintf("https://graph.facebook.com/v18.0/me/adaccounts?limit=%d&after=QVFXYZ456", limit),
		}
	}

	c.JSON(http.StatusOK, response)
}

func metaListCampaigns(c *gin.Context) {
	accountID := c.Param("account_id")

	campaigns := generateMetaCampaigns(accountID, 25)

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": campaigns,
		"paging": map[string]interface{}{
			"cursors": map[string]string{
				"before": "CAMP_BEFORE_123",
				"after":  "CAMP_AFTER_456",
			},
		},
	})
}

func metaGetInsights(c *gin.Context) {
	// Get date range
	datePreset := c.Query("date_preset")
	if datePreset == "" {
		datePreset = "last_30d"
	}

	insights := generateMetaInsights(30)

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": insights,
		"paging": map[string]interface{}{
			"cursors": map[string]string{
				"before": "INSIGHT_BEFORE",
				"after":  "INSIGHT_AFTER",
			},
		},
	})
}

func metaCampaignInsights(c *gin.Context) {
	campaignID := c.Param("campaign_id")

	// Generate daily insights for the campaign
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	insights := make([]map[string]interface{}, 0, days)
	baseDate := time.Now().AddDate(0, 0, -days)

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)

		// Skip some days randomly (simulate data gaps)
		if rand.Float32() < 0.05 {
			continue
		}

		spend := rand.Float64()*490 + 10 // RM10 - RM500
		impressions := rand.Intn(49000) + 1000
		ctr := rand.Float64()*4 + 1 // 1% - 5%
		clicks := int(float64(impressions) * ctr / 100)
		conversions := int(float64(clicks) * (rand.Float64()*7 + 1) / 100) // 1% - 8% conversion rate
		revenue := float64(conversions) * (rand.Float64()*100 + 50)       // RM50-150 per conversion

		insights = append(insights, map[string]interface{}{
			"campaign_id":      campaignID,
			"date_start":       date.Format("2006-01-02"),
			"date_stop":        date.Format("2006-01-02"),
			"impressions":      strconv.Itoa(impressions),
			"reach":            strconv.Itoa(int(float64(impressions) * 0.7)),
			"clicks":           strconv.Itoa(clicks),
			"spend":            fmt.Sprintf("%.2f", spend),
			"ctr":              fmt.Sprintf("%.4f", ctr),
			"cpc":              fmt.Sprintf("%.4f", spend/float64(clicks)),
			"cpm":              fmt.Sprintf("%.4f", spend/float64(impressions)*1000),
			"actions": []map[string]interface{}{
				{"action_type": "purchase", "value": strconv.Itoa(conversions)},
				{"action_type": "add_to_cart", "value": strconv.Itoa(conversions * 3)},
				{"action_type": "link_click", "value": strconv.Itoa(clicks)},
			},
			"action_values": []map[string]interface{}{
				{"action_type": "purchase", "value": fmt.Sprintf("%.2f", revenue)},
			},
		})
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": insights,
	})
}

func metaRefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": fmt.Sprintf("mock_meta_token_%d", time.Now().Unix()),
		"token_type":   "bearer",
		"expires_in":   3600,
	})
}

// =============================================================================
// TikTok Marketing API Mocks
// =============================================================================

func tiktokListAdvertisers(c *gin.Context) {
	advertisers := []map[string]interface{}{
		{
			"advertiser_id":   "7123456789012345678",
			"advertiser_name": "TikTok Test Advertiser 1",
			"currency":        "MYR",
			"timezone":        "Asia/Kuala_Lumpur",
			"status":          "STATUS_ENABLE",
			"role":            "ROLE_ADVERTISER",
			"create_time":     "2023-06-01 10:00:00",
		},
		{
			"advertiser_id":   "7987654321098765432",
			"advertiser_name": "TikTok Test Advertiser 2",
			"currency":        "MYR",
			"timezone":        "Asia/Kuala_Lumpur",
			"status":          "STATUS_ENABLE",
			"role":            "ROLE_ADVERTISER",
			"create_time":     "2023-08-15 14:30:00",
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "OK",
		"data": map[string]interface{}{
			"list": advertisers,
		},
		"request_id": fmt.Sprintf("tiktok_req_%d", time.Now().UnixNano()),
	})
}

func tiktokListCampaigns(c *gin.Context) {
	campaigns := generateTikTokCampaigns(20)

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "OK",
		"data": map[string]interface{}{
			"list": campaigns,
			"page_info": map[string]interface{}{
				"page":       1,
				"page_size":  20,
				"total_page": 1,
				"total_num":  len(campaigns),
			},
		},
		"request_id": fmt.Sprintf("tiktok_req_%d", time.Now().UnixNano()),
	})
}

func tiktokGetReport(c *gin.Context) {
	// Generate report data
	days := 30
	reports := make([]map[string]interface{}, 0, days)
	baseDate := time.Now().AddDate(0, 0, -days)

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)

		// Skip some days
		if rand.Float32() < 0.03 {
			continue
		}

		spend := rand.Float64()*400 + 20
		impressions := rand.Intn(40000) + 2000
		clicks := int(float64(impressions) * (rand.Float64()*3.5 + 1.5) / 100)
		conversions := int(float64(clicks) * (rand.Float64()*6 + 1) / 100)
		revenue := float64(conversions) * (rand.Float64()*80 + 40)

		reports = append(reports, map[string]interface{}{
			"stat_time_day": date.Format("2006-01-02 00:00:00"),
			"metrics": map[string]interface{}{
				"spend":              fmt.Sprintf("%.2f", spend),
				"impressions":        strconv.Itoa(impressions),
				"clicks":             strconv.Itoa(clicks),
				"ctr":                fmt.Sprintf("%.4f", float64(clicks)/float64(impressions)*100),
				"cpc":                fmt.Sprintf("%.4f", spend/float64(clicks)),
				"cpm":                fmt.Sprintf("%.4f", spend/float64(impressions)*1000),
				"conversion":         strconv.Itoa(conversions),
				"cost_per_conversion": fmt.Sprintf("%.4f", spend/float64(conversions+1)),
				"total_purchase_value": fmt.Sprintf("%.2f", revenue),
				"video_views_p25":    strconv.Itoa(int(float64(impressions) * 0.6)),
				"video_views_p50":    strconv.Itoa(int(float64(impressions) * 0.4)),
				"video_views_p75":    strconv.Itoa(int(float64(impressions) * 0.25)),
				"video_views_p100":   strconv.Itoa(int(float64(impressions) * 0.15)),
			},
		})
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "OK",
		"data": map[string]interface{}{
			"list": reports,
			"page_info": map[string]interface{}{
				"page":       1,
				"page_size":  days,
				"total_page": 1,
				"total_num":  len(reports),
			},
		},
		"request_id": fmt.Sprintf("tiktok_req_%d", time.Now().UnixNano()),
	})
}

func tiktokRefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "OK",
		"data": map[string]interface{}{
			"access_token":  fmt.Sprintf("mock_tiktok_token_%d", time.Now().Unix()),
			"refresh_token": fmt.Sprintf("mock_tiktok_refresh_%d", time.Now().Unix()),
			"token_type":    "Bearer",
			"expires_in":    86400,
			"scope":         "user.info.basic,campaign.read,advertiser.read",
		},
		"request_id": fmt.Sprintf("tiktok_req_%d", time.Now().UnixNano()),
	})
}

// =============================================================================
// Shopee API Mocks
// =============================================================================

func shopeeGetShopInfo(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"shop_id":     12345678,
		"shop_name":   "Test Shopee Store",
		"region":      "MY",
		"status":      "NORMAL",
		"is_cb":       false,
		"is_cnsc":     false,
		"description": "A test Shopee store for local development",
		"videos":      []interface{}{},
		"images":      []interface{}{},
	})
}

func shopeeListProducts(c *gin.Context) {
	products := []map[string]interface{}{
		{
			"item_id":      100001,
			"item_name":    "Product A - Best Seller",
			"item_status":  "NORMAL",
			"update_time":  time.Now().Unix(),
			"create_time":  time.Now().AddDate(0, -3, 0).Unix(),
			"price_info":   []map[string]interface{}{{"original_price": 99.90}},
			"stock_info_v2": map[string]interface{}{"summary_info": map[string]interface{}{"total_reserved_stock": 150}},
		},
		{
			"item_id":      100002,
			"item_name":    "Product B - New Arrival",
			"item_status":  "NORMAL",
			"update_time":  time.Now().Unix(),
			"create_time":  time.Now().AddDate(0, -1, 0).Unix(),
			"price_info":   []map[string]interface{}{{"original_price": 149.90}},
			"stock_info_v2": map[string]interface{}{"summary_info": map[string]interface{}{"total_reserved_stock": 80}},
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"item":       products,
			"total":      len(products),
			"has_more":   false,
			"next_offset": "",
		},
	})
}

func shopeeGetPerformance(c *gin.Context) {
	// Generate performance data
	days := 30
	performance := make([]map[string]interface{}, 0, days)
	baseDate := time.Now().AddDate(0, 0, -days)

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)

		visits := rand.Intn(500) + 50
		orders := int(float64(visits) * (rand.Float64()*3 + 1) / 100) // 1-4% conversion
		revenue := float64(orders) * (rand.Float64()*100 + 30)
		ads_spend := rand.Float64()*100 + 10

		performance = append(performance, map[string]interface{}{
			"date":              date.Format("2006-01-02"),
			"page_views":        visits * 2,
			"unique_visitors":   visits,
			"orders":            orders,
			"units_sold":        orders + rand.Intn(orders+1),
			"gross_merchandise_value": revenue,
			"ads_spend":         ads_spend,
			"ads_impressions":   rand.Intn(10000) + 1000,
			"ads_clicks":        rand.Intn(500) + 50,
		})
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"response": map[string]interface{}{
			"data": performance,
		},
	})
}

func shopeeRefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  fmt.Sprintf("mock_shopee_token_%d", time.Now().Unix()),
		"refresh_token": fmt.Sprintf("mock_shopee_refresh_%d", time.Now().Unix()),
		"expire_in":     14400,
		"request_id":    fmt.Sprintf("shopee_%d", time.Now().UnixNano()),
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

func generateMetaCampaigns(accountID string, count int) []map[string]interface{} {
	objectives := []string{"CONVERSIONS", "LINK_CLICKS", "REACH", "VIDEO_VIEWS", "LEAD_GENERATION"}
	statuses := []string{"ACTIVE", "PAUSED", "ACTIVE", "ACTIVE", "PAUSED"} // More active than paused

	campaigns := make([]map[string]interface{}, 0, count)

	for i := 0; i < count; i++ {
		budget := (rand.Float64()*490 + 10) * 100 // RM10-500 in cents
		status := statuses[rand.Intn(len(statuses))]
		objective := objectives[rand.Intn(len(objectives))]

		campaigns = append(campaigns, map[string]interface{}{
			"id":                  fmt.Sprintf("%s_camp_%d", accountID, 1000+i),
			"name":                fmt.Sprintf("Meta Campaign %d - %s", i+1, objective),
			"account_id":          accountID,
			"objective":           objective,
			"status":              status,
			"effective_status":    status,
			"configured_status":   status,
			"daily_budget":        fmt.Sprintf("%.0f", budget),
			"lifetime_budget":     "0",
			"budget_remaining":    fmt.Sprintf("%.0f", budget*0.7),
			"buying_type":         "AUCTION",
			"special_ad_categories": []interface{}{},
			"created_time":        time.Now().AddDate(0, 0, -rand.Intn(90)).Format(time.RFC3339),
			"updated_time":        time.Now().AddDate(0, 0, -rand.Intn(7)).Format(time.RFC3339),
			"start_time":          time.Now().AddDate(0, 0, -rand.Intn(60)).Format(time.RFC3339),
		})
	}

	return campaigns
}

func generateTikTokCampaigns(count int) []map[string]interface{} {
	objectives := []string{"CONVERSIONS", "TRAFFIC", "REACH", "VIDEO_VIEWS", "LEAD_GENERATION"}
	statuses := []string{"CAMPAIGN_STATUS_ENABLE", "CAMPAIGN_STATUS_DISABLE", "CAMPAIGN_STATUS_ENABLE"}

	campaigns := make([]map[string]interface{}, 0, count)

	for i := 0; i < count; i++ {
		budget := rand.Float64()*490 + 10
		status := statuses[rand.Intn(len(statuses))]
		objective := objectives[rand.Intn(len(objectives))]

		campaigns = append(campaigns, map[string]interface{}{
			"campaign_id":     fmt.Sprintf("7%017d", rand.Int63n(99999999999999999)),
			"campaign_name":   fmt.Sprintf("TikTok Campaign %d - %s", i+1, objective),
			"advertiser_id":   "7123456789012345678",
			"objective":       objective,
			"objective_type":  objective,
			"campaign_type":   "REGULAR_CAMPAIGN",
			"status":          status,
			"operation_status": status,
			"budget_mode":     "BUDGET_MODE_DAY",
			"budget":          fmt.Sprintf("%.2f", budget),
			"create_time":     time.Now().AddDate(0, 0, -rand.Intn(90)).Format("2006-01-02 15:04:05"),
			"modify_time":     time.Now().AddDate(0, 0, -rand.Intn(7)).Format("2006-01-02 15:04:05"),
		})
	}

	return campaigns
}

func generateMetaInsights(days int) []map[string]interface{} {
	insights := make([]map[string]interface{}, 0, days)
	baseDate := time.Now().AddDate(0, 0, -days)

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)

		spend := rand.Float64()*490 + 10
		impressions := rand.Intn(49000) + 1000
		clicks := int(float64(impressions) * (rand.Float64()*4 + 1) / 100)
		conversions := int(float64(clicks) * (rand.Float64()*7 + 1) / 100)
		revenue := float64(conversions) * (rand.Float64()*100 + 50)

		insights = append(insights, map[string]interface{}{
			"date_start":  date.Format("2006-01-02"),
			"date_stop":   date.Format("2006-01-02"),
			"impressions": strconv.Itoa(impressions),
			"reach":       strconv.Itoa(int(float64(impressions) * 0.7)),
			"clicks":      strconv.Itoa(clicks),
			"spend":       fmt.Sprintf("%.2f", spend),
			"actions": []map[string]interface{}{
				{"action_type": "purchase", "value": strconv.Itoa(conversions)},
				{"action_type": "link_click", "value": strconv.Itoa(clicks)},
			},
			"action_values": []map[string]interface{}{
				{"action_type": "purchase", "value": fmt.Sprintf("%.2f", revenue)},
			},
		})
	}

	return insights
}
