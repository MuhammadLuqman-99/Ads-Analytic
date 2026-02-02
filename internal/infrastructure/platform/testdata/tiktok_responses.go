package testdata

// TikTokResponses contains sample TikTok Ads API responses for testing
var TikTokResponses = struct {
	// Campaigns
	CampaignsSuccess   string
	CampaignsEmpty     string
	CampaignsPaginated string

	// Reports/Insights
	ReportSuccess     string
	ReportEmpty       string
	ReportWithMetrics string

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
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789ABCDEF",
		"data": {
			"list": [
				{
					"campaign_id": "1780000000000001",
					"campaign_name": "TikTok Shop Promo",
					"advertiser_id": "7000000000000001",
					"objective_type": "PRODUCT_SALES",
					"status": "CAMPAIGN_STATUS_ENABLE",
					"budget_mode": "BUDGET_MODE_DAY",
					"budget": 500.00,
					"create_time": "2024-01-10 10:00:00",
					"modify_time": "2024-01-15 14:30:00"
				},
				{
					"campaign_id": "1780000000000002",
					"campaign_name": "Brand Video Campaign",
					"advertiser_id": "7000000000000001",
					"objective_type": "VIDEO_VIEWS",
					"status": "CAMPAIGN_STATUS_DISABLE",
					"budget_mode": "BUDGET_MODE_DAY",
					"budget": 300.00,
					"create_time": "2024-01-08 09:00:00",
					"modify_time": "2024-01-12 11:00:00"
				}
			],
			"page_info": {
				"page": 1,
				"page_size": 20,
				"total_number": 2,
				"total_page": 1
			}
		}
	}`,

	// Campaigns - empty response
	CampaignsEmpty: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789EMPTY",
		"data": {
			"list": [],
			"page_info": {
				"page": 1,
				"page_size": 20,
				"total_number": 0,
				"total_page": 0
			}
		}
	}`,

	// Campaigns - paginated (has more pages)
	CampaignsPaginated: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789PAGED",
		"data": {
			"list": [
				{
					"campaign_id": "1780000000000001",
					"campaign_name": "Page 1 Campaign",
					"status": "CAMPAIGN_STATUS_ENABLE"
				}
			],
			"page_info": {
				"page": 1,
				"page_size": 1,
				"total_number": 2,
				"total_page": 2
			}
		}
	}`,

	// Report - successful with metrics
	ReportSuccess: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789REPORT",
		"data": {
			"list": [
				{
					"dimensions": {
						"campaign_id": "1780000000000001",
						"stat_time_day": "2024-01-20"
					},
					"metrics": {
						"spend": "125.50",
						"impressions": "85000",
						"clicks": "2550",
						"conversions": "45",
						"conversion_rate": "1.76",
						"ctr": "3.0",
						"cpc": "0.049",
						"cpm": "1.48",
						"cost_per_conversion": "2.79",
						"total_complete_payment_rate": "0.85",
						"result": "45",
						"result_rate": "1.76",
						"cost_per_result": "2.79",
						"video_play_actions": "42500",
						"video_watched_2s": "38000",
						"video_watched_6s": "25000",
						"average_video_play": "8.5"
					}
				},
				{
					"dimensions": {
						"campaign_id": "1780000000000002",
						"stat_time_day": "2024-01-20"
					},
					"metrics": {
						"spend": "80.00",
						"impressions": "120000",
						"clicks": "1200",
						"conversions": "0",
						"conversion_rate": "0",
						"ctr": "1.0",
						"cpc": "0.067",
						"cpm": "0.67",
						"video_play_actions": "95000",
						"video_watched_2s": "85000",
						"video_watched_6s": "55000",
						"average_video_play": "12.3"
					}
				}
			],
			"page_info": {
				"page": 1,
				"page_size": 20,
				"total_number": 2,
				"total_page": 1
			}
		}
	}`,

	// Report - empty response
	ReportEmpty: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789EMPTY",
		"data": {
			"list": [],
			"page_info": {
				"page": 1,
				"page_size": 20,
				"total_number": 0,
				"total_page": 0
			}
		}
	}`,

	// Report - with extended metrics
	ReportWithMetrics: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789METRICS",
		"data": {
			"list": [
				{
					"dimensions": {
						"campaign_id": "1780000000000001",
						"adgroup_id": "1790000000000001",
						"ad_id": "1800000000000001",
						"stat_time_day": "2024-01-20"
					},
					"metrics": {
						"spend": "50.25",
						"impressions": "35000",
						"clicks": "1050",
						"conversions": "18",
						"conversion_rate": "1.71",
						"ctr": "3.0",
						"cpc": "0.048",
						"cpm": "1.44",
						"cost_per_conversion": "2.79",
						"reach": "28000",
						"frequency": "1.25",
						"total_onsite_shopping_value": "450.00",
						"onsite_shopping_roas": "8.96"
					}
				}
			],
			"page_info": {
				"page": 1,
				"page_size": 20,
				"total_number": 1,
				"total_page": 1
			}
		}
	}`,

	// OAuth - token exchange success
	TokenExchange: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789TOKEN",
		"data": {
			"access_token": "tiktok_access_token_abc123xyz789",
			"advertiser_ids": ["7000000000000001", "7000000000000002"],
			"scope": ["ad.read", "ad.write", "report.read"],
			"token_type": "Bearer",
			"expires_in": 86400,
			"refresh_token": "tiktok_refresh_token_def456uvw",
			"refresh_token_expires_in": 2592000
		}
	}`,

	// OAuth - token refresh success
	TokenRefresh: `{
		"code": 0,
		"message": "OK",
		"request_id": "20240121123456789REFRESH",
		"data": {
			"access_token": "tiktok_new_access_token_newtoken123",
			"advertiser_ids": ["7000000000000001"],
			"scope": ["ad.read", "ad.write", "report.read"],
			"token_type": "Bearer",
			"expires_in": 86400,
			"refresh_token": "tiktok_new_refresh_token_newrefresh",
			"refresh_token_expires_in": 2592000
		}
	}`,

	// OAuth - invalid token
	TokenInvalid: `{
		"code": 40105,
		"message": "Access token is invalid",
		"request_id": "20240121123456789INVALID"
	}`,

	// Error - rate limit
	RateLimitError: `{
		"code": 40100,
		"message": "Too Many Requests: API rate limit exceeded",
		"request_id": "20240121123456789RATELIMIT",
		"data": {
			"retry_after": 60
		}
	}`,

	// Error - authentication error
	AuthError: `{
		"code": 40001,
		"message": "Unauthorized: Access token expired or invalid",
		"request_id": "20240121123456789AUTH"
	}`,

	// Error - invalid parameter
	InvalidParamError: `{
		"code": 40002,
		"message": "Invalid Parameter: campaign_id is required",
		"request_id": "20240121123456789PARAM"
	}`,

	// Error - server error
	ServerError: `{
		"code": 50000,
		"message": "Internal Server Error: Please try again later",
		"request_id": "20240121123456789SERVER"
	}`,
}
