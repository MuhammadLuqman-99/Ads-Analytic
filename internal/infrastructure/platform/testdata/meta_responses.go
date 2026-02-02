// Package testdata provides mock API responses for testing platform connectors
package testdata

// MetaResponses contains sample Meta (Facebook) API responses for testing
var MetaResponses = struct {
	// Campaigns
	CampaignsSuccess     string
	CampaignsEmpty       string
	CampaignsPaginated   string
	CampaignsNextPage    string

	// Insights
	InsightsSuccess      string
	InsightsEmpty        string
	InsightsWithBreakdown string

	// AdSets
	AdSetsSuccess        string

	// Ads
	AdsSuccess           string

	// OAuth
	TokenExchange        string
	TokenRefresh         string
	TokenInvalid         string

	// Errors
	RateLimitError       string
	AuthError            string
	InvalidRequestError  string
	ServerError          string
}{
	// Campaigns - successful response with multiple campaigns
	CampaignsSuccess: `{
		"data": [
			{
				"id": "120330000000000001",
				"name": "Summer Sale 2024",
				"status": "ACTIVE",
				"objective": "CONVERSIONS",
				"daily_budget": "5000",
				"lifetime_budget": "0",
				"created_time": "2024-01-15T10:30:00+0000",
				"updated_time": "2024-01-20T14:00:00+0000",
				"effective_status": "ACTIVE",
				"configured_status": "ACTIVE"
			},
			{
				"id": "120330000000000002",
				"name": "Brand Awareness Q1",
				"status": "PAUSED",
				"objective": "BRAND_AWARENESS",
				"daily_budget": "3000",
				"lifetime_budget": "0",
				"created_time": "2024-01-10T08:00:00+0000",
				"updated_time": "2024-01-18T12:00:00+0000",
				"effective_status": "PAUSED",
				"configured_status": "PAUSED"
			}
		],
		"paging": {
			"cursors": {
				"before": "MAZDZD",
				"after": "MjQZD"
			}
		}
	}`,

	// Campaigns - empty response
	CampaignsEmpty: `{
		"data": [],
		"paging": {}
	}`,

	// Campaigns - paginated response (has next page)
	CampaignsPaginated: `{
		"data": [
			{
				"id": "120330000000000001",
				"name": "Campaign Page 1",
				"status": "ACTIVE",
				"objective": "CONVERSIONS"
			}
		],
		"paging": {
			"cursors": {
				"before": "MAZDZD",
				"after": "MjQZD"
			},
			"next": "https://graph.facebook.com/v18.0/act_123/campaigns?after=MjQZD"
		}
	}`,

	// Campaigns - second page
	CampaignsNextPage: `{
		"data": [
			{
				"id": "120330000000000002",
				"name": "Campaign Page 2",
				"status": "ACTIVE",
				"objective": "CONVERSIONS"
			}
		],
		"paging": {
			"cursors": {
				"before": "MjQZD",
				"after": "MzYZD"
			}
		}
	}`,

	// Insights - successful response with metrics
	InsightsSuccess: `{
		"data": [
			{
				"campaign_id": "120330000000000001",
				"campaign_name": "Summer Sale 2024",
				"impressions": "150000",
				"clicks": "4500",
				"spend": "250.50",
				"reach": "85000",
				"ctr": "3.0",
				"cpc": "0.056",
				"cpm": "1.67",
				"actions": [
					{"action_type": "purchase", "value": "120"},
					{"action_type": "add_to_cart", "value": "450"},
					{"action_type": "link_click", "value": "4500"}
				],
				"action_values": [
					{"action_type": "purchase", "value": "12500.00"}
				],
				"date_start": "2024-01-15",
				"date_stop": "2024-01-21"
			},
			{
				"campaign_id": "120330000000000002",
				"campaign_name": "Brand Awareness Q1",
				"impressions": "250000",
				"clicks": "2500",
				"spend": "180.00",
				"reach": "150000",
				"ctr": "1.0",
				"cpc": "0.072",
				"cpm": "0.72",
				"actions": [
					{"action_type": "link_click", "value": "2500"}
				],
				"date_start": "2024-01-15",
				"date_stop": "2024-01-21"
			}
		],
		"paging": {}
	}`,

	// Insights - empty response
	InsightsEmpty: `{
		"data": [],
		"paging": {}
	}`,

	// Insights - with age/gender breakdown
	InsightsWithBreakdown: `{
		"data": [
			{
				"campaign_id": "120330000000000001",
				"impressions": "50000",
				"clicks": "1500",
				"spend": "83.50",
				"age": "25-34",
				"gender": "female",
				"date_start": "2024-01-15",
				"date_stop": "2024-01-21"
			},
			{
				"campaign_id": "120330000000000001",
				"impressions": "45000",
				"clicks": "1350",
				"spend": "75.00",
				"age": "25-34",
				"gender": "male",
				"date_start": "2024-01-15",
				"date_stop": "2024-01-21"
			}
		],
		"paging": {}
	}`,

	// AdSets - successful response
	AdSetsSuccess: `{
		"data": [
			{
				"id": "120330000000000101",
				"name": "Interest - Fashion",
				"campaign_id": "120330000000000001",
				"status": "ACTIVE",
				"targeting": {
					"age_min": 25,
					"age_max": 45,
					"genders": [2],
					"geo_locations": {
						"countries": ["MY", "SG"]
					}
				},
				"optimization_goal": "OFFSITE_CONVERSIONS",
				"billing_event": "IMPRESSIONS",
				"bid_amount": 500,
				"daily_budget": "2500"
			}
		],
		"paging": {}
	}`,

	// Ads - successful response
	AdsSuccess: `{
		"data": [
			{
				"id": "120330000000000201",
				"name": "Video Ad - Product Showcase",
				"adset_id": "120330000000000101",
				"campaign_id": "120330000000000001",
				"status": "ACTIVE",
				"effective_status": "ACTIVE",
				"creative": {
					"id": "120330000000000301",
					"name": "Summer Creative v1",
					"title": "Summer Sale - Up to 50% Off!",
					"body": "Shop now and save big on summer essentials.",
					"image_url": "https://example.com/image.jpg"
				}
			}
		],
		"paging": {}
	}`,

	// OAuth - token exchange success
	TokenExchange: `{
		"access_token": "EAAGm0PX4ZAAAAABkZB5sampletoken123456789",
		"token_type": "bearer",
		"expires_in": 5183944
	}`,

	// OAuth - token refresh success
	TokenRefresh: `{
		"access_token": "EAAGm0PX4ZAAAAABkZB5newtoken987654321",
		"token_type": "bearer",
		"expires_in": 5183944
	}`,

	// OAuth - invalid token
	TokenInvalid: `{
		"error": {
			"message": "Invalid OAuth access token.",
			"type": "OAuthException",
			"code": 190,
			"error_subcode": 460,
			"fbtrace_id": "AbCdEfGhIjK123"
		}
	}`,

	// Error - rate limit
	RateLimitError: `{
		"error": {
			"message": "(#17) User request limit reached",
			"type": "OAuthException",
			"code": 17,
			"error_subcode": 2446079,
			"is_transient": true,
			"error_user_title": "Rate Limit Exceeded",
			"error_user_msg": "You've made too many requests. Please wait before trying again.",
			"fbtrace_id": "AbCdEfGhIjK456"
		}
	}`,

	// Error - authentication error
	AuthError: `{
		"error": {
			"message": "Error validating access token: Session has expired",
			"type": "OAuthException",
			"code": 190,
			"error_subcode": 463,
			"fbtrace_id": "AbCdEfGhIjK789"
		}
	}`,

	// Error - invalid request
	InvalidRequestError: `{
		"error": {
			"message": "(#100) Invalid parameter",
			"type": "OAuthException",
			"code": 100,
			"error_subcode": 33,
			"fbtrace_id": "AbCdEfGhIjKabc"
		}
	}`,

	// Error - server error
	ServerError: `{
		"error": {
			"message": "An unexpected error has occurred. Please retry your request later.",
			"type": "OAuthException",
			"code": 2,
			"is_transient": true,
			"fbtrace_id": "AbCdEfGhIjKdef"
		}
	}`,
}
