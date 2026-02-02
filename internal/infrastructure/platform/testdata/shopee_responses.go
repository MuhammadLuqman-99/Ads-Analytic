package testdata

// ShopeeResponses contains sample Shopee Ads API responses for testing
var ShopeeResponses = struct {
	// Campaigns
	CampaignsSuccess   string
	CampaignsEmpty     string
	CampaignsPaginated string

	// Shop Metrics
	ShopMetricsSuccess string
	ShopMetricsEmpty   string

	// Product Ads
	ProductAdsSuccess string

	// OAuth
	TokenExchange     string
	TokenRefresh      string
	TokenInvalid      string

	// Errors
	RateLimitError    string
	AuthError         string
	InvalidParamError string
	ServerError       string
}{
	// Campaigns - successful response
	CampaignsSuccess: `{
		"error": "",
		"message": "",
		"response": {
			"campaigns": [
				{
					"campaign_id": 1001,
					"campaign_name": "Shopee 11.11 Sale",
					"shop_id": 500001,
					"campaign_type": "PRODUCT_SEARCH",
					"status": "ONGOING",
					"daily_budget": 100.00,
					"total_budget": 1000.00,
					"start_time": 1704067200,
					"end_time": 1706745600,
					"create_time": 1704067200,
					"update_time": 1705276800
				},
				{
					"campaign_id": 1002,
					"campaign_name": "Discovery Ads",
					"shop_id": 500001,
					"campaign_type": "DISCOVERY",
					"status": "PAUSED",
					"daily_budget": 50.00,
					"total_budget": 500.00,
					"start_time": 1703980800,
					"end_time": 1706572800,
					"create_time": 1703980800,
					"update_time": 1705190400
				}
			],
			"more": false,
			"total": 2
		},
		"request_id": "shopee_req_abc123"
	}`,

	// Campaigns - empty response
	CampaignsEmpty: `{
		"error": "",
		"message": "",
		"response": {
			"campaigns": [],
			"more": false,
			"total": 0
		},
		"request_id": "shopee_req_empty"
	}`,

	// Campaigns - paginated (has more)
	CampaignsPaginated: `{
		"error": "",
		"message": "",
		"response": {
			"campaigns": [
				{
					"campaign_id": 1001,
					"campaign_name": "Page 1 Campaign",
					"status": "ONGOING"
				}
			],
			"more": true,
			"total": 25
		},
		"request_id": "shopee_req_page1"
	}`,

	// Shop Metrics - successful response
	ShopMetricsSuccess: `{
		"error": "",
		"message": "",
		"response": {
			"data": [
				{
					"date": "2024-01-20",
					"campaign_id": 1001,
					"campaign_name": "Shopee 11.11 Sale",
					"impressions": 50000,
					"clicks": 2500,
					"orders": 125,
					"gmv": 12500.00,
					"cost": 75.00,
					"ctr": 5.0,
					"conversion_rate": 5.0,
					"cpc": 0.03,
					"roas": 166.67,
					"acos": 0.6
				},
				{
					"date": "2024-01-20",
					"campaign_id": 1002,
					"campaign_name": "Discovery Ads",
					"impressions": 30000,
					"clicks": 900,
					"orders": 27,
					"gmv": 2700.00,
					"cost": 40.00,
					"ctr": 3.0,
					"conversion_rate": 3.0,
					"cpc": 0.044,
					"roas": 67.5,
					"acos": 1.48
				}
			],
			"summary": {
				"total_impressions": 80000,
				"total_clicks": 3400,
				"total_orders": 152,
				"total_gmv": 15200.00,
				"total_cost": 115.00,
				"avg_ctr": 4.25,
				"avg_conversion_rate": 4.47,
				"overall_roas": 132.17
			}
		},
		"request_id": "shopee_req_metrics"
	}`,

	// Shop Metrics - empty response
	ShopMetricsEmpty: `{
		"error": "",
		"message": "",
		"response": {
			"data": [],
			"summary": {
				"total_impressions": 0,
				"total_clicks": 0,
				"total_orders": 0,
				"total_gmv": 0,
				"total_cost": 0,
				"avg_ctr": 0,
				"avg_conversion_rate": 0,
				"overall_roas": 0
			}
		},
		"request_id": "shopee_req_nometrics"
	}`,

	// Product Ads - successful response
	ProductAdsSuccess: `{
		"error": "",
		"message": "",
		"response": {
			"ads": [
				{
					"ad_id": 2001,
					"campaign_id": 1001,
					"item_id": 100001,
					"item_name": "Premium Wireless Earbuds",
					"status": "ONGOING",
					"bid_price": 0.50,
					"quality_score": 8.5,
					"impressions": 15000,
					"clicks": 750,
					"orders": 45,
					"gmv": 4500.00,
					"cost": 22.50,
					"ctr": 5.0,
					"roas": 200.0
				},
				{
					"ad_id": 2002,
					"campaign_id": 1001,
					"item_id": 100002,
					"item_name": "Smart Watch Pro",
					"status": "ONGOING",
					"bid_price": 0.75,
					"quality_score": 7.8,
					"impressions": 12000,
					"clicks": 480,
					"orders": 24,
					"gmv": 2880.00,
					"cost": 18.00,
					"ctr": 4.0,
					"roas": 160.0
				}
			],
			"more": false,
			"total": 2
		},
		"request_id": "shopee_req_products"
	}`,

	// OAuth - token exchange success
	TokenExchange: `{
		"error": "",
		"message": "",
		"access_token": "shopee_access_token_xyz789abc123",
		"refresh_token": "shopee_refresh_token_def456uvw",
		"expire_in": 14400,
		"request_id": "shopee_req_token"
	}`,

	// OAuth - token refresh success
	TokenRefresh: `{
		"error": "",
		"message": "",
		"access_token": "shopee_new_access_token_newtoken",
		"refresh_token": "shopee_new_refresh_token_newrefresh",
		"expire_in": 14400,
		"request_id": "shopee_req_refresh"
	}`,

	// OAuth - invalid token
	TokenInvalid: `{
		"error": "error_auth",
		"message": "Invalid or expired access token",
		"request_id": "shopee_req_invalid"
	}`,

	// Error - rate limit
	RateLimitError: `{
		"error": "error_too_many_request",
		"message": "Request frequency exceeds limit, please try again later",
		"request_id": "shopee_req_ratelimit"
	}`,

	// Error - authentication error
	AuthError: `{
		"error": "error_auth",
		"message": "Authentication failed. Please check your credentials",
		"request_id": "shopee_req_auth"
	}`,

	// Error - invalid parameter
	InvalidParamError: `{
		"error": "error_param",
		"message": "Invalid parameter: shop_id is required",
		"request_id": "shopee_req_param"
	}`,

	// Error - server error
	ServerError: `{
		"error": "error_server",
		"message": "Internal server error. Please try again later",
		"request_id": "shopee_req_server"
	}`,
}

// ShopeeRegions contains region-specific API URLs for testing
var ShopeeRegions = map[string]string{
	"MY": "https://partner.shopeemobile.com",    // Malaysia
	"SG": "https://partner.shopeemobile.com",    // Singapore
	"TH": "https://partner.shopeemobile.com",    // Thailand
	"ID": "https://partner.shopeemobile.com",    // Indonesia
	"VN": "https://partner.shopeemobile.com",    // Vietnam
	"PH": "https://partner.shopeemobile.com",    // Philippines
	"BR": "https://openapi.shopee.com.br",       // Brazil
	"MX": "https://openapi.shopee.com.mx",       // Mexico
	"CO": "https://openapi.shopee.com.co",       // Colombia
	"CL": "https://openapi.shopee.cl",           // Chile
}
