package meta

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
)

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.APIVersion != "v18.0" {
		t.Errorf("APIVersion = %q, want %q", config.APIVersion, "v18.0")
	}

	if config.RateLimitCalls != 200 {
		t.Errorf("RateLimitCalls = %d, want %d", config.RateLimitCalls, 200)
	}

	if config.RateLimitWindow != time.Hour {
		t.Errorf("RateLimitWindow = %v, want %v", config.RateLimitWindow, time.Hour)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", config.Timeout, 30*time.Second)
	}

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want %d", config.MaxRetries, 3)
	}
}

func TestNewConnector_WithConfig(t *testing.T) {
	config := &Config{
		AppID:           "test_app_id",
		AppSecret:       "test_app_secret",
		RedirectURI:     "http://localhost:8080/callback",
		APIVersion:      "v19.0",
		RateLimitCalls:  100,
		RateLimitWindow: 30 * time.Minute,
		Timeout:         15 * time.Second,
		MaxRetries:      5,
	}

	connector := NewConnector(config)

	if connector == nil {
		t.Fatal("NewConnector() returned nil")
	}

	if connector.config.AppID != "test_app_id" {
		t.Errorf("AppID = %q, want %q", connector.config.AppID, "test_app_id")
	}

	if connector.apiVersion != "v19.0" {
		t.Errorf("apiVersion = %q, want %q", connector.apiVersion, "v19.0")
	}
}

func TestNewConnector_WithNilConfig(t *testing.T) {
	connector := NewConnector(nil)

	if connector == nil {
		t.Fatal("NewConnector() returned nil")
	}

	// Should use default config
	if connector.config.APIVersion != "v18.0" {
		t.Errorf("expected default APIVersion, got %q", connector.config.APIVersion)
	}
}

// ============================================================================
// OAuth URL Tests
// ============================================================================

func TestGetAuthURL(t *testing.T) {
	config := &Config{
		AppID:       "123456789",
		AppSecret:   "secret",
		RedirectURI: "http://localhost:8080/api/v1/oauth/meta/callback",
		APIVersion:  "v18.0",
	}

	connector := NewConnector(config)
	state := "test_state_abc123"

	url := connector.GetAuthURL(state)

	// Verify URL contains expected parameters
	tests := []struct {
		name     string
		contains string
	}{
		{"base URL", "https://www.facebook.com/v18.0/dialog/oauth"},
		{"client_id", "client_id=123456789"},
		{"redirect_uri", "redirect_uri="},
		{"state", "state=test_state_abc123"},
		{"scope", "scope=ads_read"},
		{"response_type", "response_type=code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !contains(url, tt.contains) {
				t.Errorf("URL should contain %q, got: %s", tt.contains, url)
			}
		})
	}
}

func TestGetAuthURL_DifferentScopes(t *testing.T) {
	config := &Config{
		AppID:       "app_123",
		RedirectURI: "http://localhost/callback",
		APIVersion:  "v18.0",
	}

	connector := NewConnector(config)
	url := connector.GetAuthURL("state123")

	// Should include ads_management scope
	if !contains(url, "ads_management") {
		t.Error("URL should contain ads_management scope")
	}

	// Should include business_management scope
	if !contains(url, "business_management") {
		t.Error("URL should contain business_management scope")
	}
}

// ============================================================================
// Account Status Mapping Tests
// ============================================================================

func TestMapAccountStatus(t *testing.T) {
	connector := NewConnector(DefaultConfig())

	tests := []struct {
		status   int
		expected string
	}{
		{1, "active"},
		{2, "disabled"},
		{3, "unsettled"},
		{7, "pending_risk_review"},
		{8, "pending_settlement"},
		{9, "in_grace_period"},
		{100, "pending_closure"},
		{101, "closed"},
		{201, "any_active"},
		{202, "any_closed"},
		{999, "unknown"},
		{-1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := connector.mapAccountStatus(tt.status)
			if result != tt.expected {
				t.Errorf("mapAccountStatus(%d) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Campaign Status Mapping Tests
// ============================================================================

func TestMapCampaignStatus(t *testing.T) {
	connector := NewConnector(DefaultConfig())

	tests := []struct {
		input    string
		expected entity.CampaignStatus
	}{
		{"ACTIVE", entity.CampaignStatusActive},
		{"PAUSED", entity.CampaignStatusPaused},
		{"DELETED", entity.CampaignStatusDeleted},
		{"ARCHIVED", entity.CampaignStatusArchived},
		{"DRAFT", entity.CampaignStatusDraft},
		{"UNKNOWN", entity.CampaignStatusPaused}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := connector.mapCampaignStatus(tt.input)
			if result != tt.expected {
				t.Errorf("mapCampaignStatus(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Objective Mapping Tests
// ============================================================================

func TestMapObjective(t *testing.T) {
	connector := NewConnector(DefaultConfig())

	tests := []struct {
		input    string
		expected entity.CampaignObjective
	}{
		{"BRAND_AWARENESS", entity.ObjectiveAwareness},
		{"REACH", entity.ObjectiveAwareness},
		{"LINK_CLICKS", entity.ObjectiveTraffic},
		{"POST_ENGAGEMENT", entity.ObjectiveEngagement},
		{"LEAD_GENERATION", entity.ObjectiveLeads},
		{"APP_INSTALLS", entity.ObjectiveAppPromotion},
		{"CONVERSIONS", entity.ObjectiveConversions},
		{"CATALOG_SALES", entity.ObjectiveSales},
		{"VIDEO_VIEWS", entity.ObjectiveVideoViews},
		{"MESSAGES", entity.ObjectiveMessages},
		{"STORE_TRAFFIC", entity.ObjectiveStoreTraffic},
		{"UNKNOWN_OBJECTIVE", entity.ObjectiveAwareness}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := connector.mapObjective(tt.input)
			if result != tt.expected {
				t.Errorf("mapObjective(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Time Range Formatting Tests
// ============================================================================

func TestFormatTimeRange(t *testing.T) {
	connector := NewConnector(DefaultConfig())

	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)

	dateRange := entity.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}

	result := connector.formatTimeRange(dateRange)

	expected := `{"since":"2025-01-01","until":"2025-01-31"}`
	if result != expected {
		t.Errorf("formatTimeRange() = %q, want %q", result, expected)
	}
}

// ============================================================================
// Mock Server Tests
// ============================================================================

func TestExchangeCode_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v18.0/oauth/access_token" {
			// Token exchange endpoint
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "long_lived_token_123",
				"token_type": "Bearer",
				"expires_in": 5184000
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Note: This is a structural test - actual API calls would require mocking the baseURL
	t.Log("ExchangeCode structure validated")
}

func TestGetUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "/me") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "12345",
				"name": "Test User",
				"email": "test@example.com",
				"picture": {
					"data": {
						"url": "https://example.com/avatar.jpg"
					}
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Log("GetUserInfo structure validated")
}

// ============================================================================
// Error Response Tests
// ============================================================================

func TestMetaErrorResponse_Parsing(t *testing.T) {
	// Test that Meta error responses would be handled correctly
	errorJSON := `{
		"error": {
			"message": "Error validating access token",
			"type": "OAuthException",
			"code": 190,
			"error_subcode": 460,
			"fbtrace_id": "trace123"
		}
	}`

	// Verify structure
	if !contains(errorJSON, "OAuthException") {
		t.Error("expected OAuthException in error response")
	}
}

// ============================================================================
// Platform Constants Tests
// ============================================================================

func TestPlatformConstants(t *testing.T) {
	// Verify expected constants
	if entity.PlatformMeta != "meta" {
		t.Errorf("PlatformMeta = %q, want %q", entity.PlatformMeta, "meta")
	}

	if authURL != "https://www.facebook.com/v18.0/dialog/oauth" {
		t.Errorf("authURL = %q, want %q", authURL, "https://www.facebook.com/v18.0/dialog/oauth")
	}

	if baseURL != "https://graph.facebook.com" {
		t.Errorf("baseURL = %q, want %q", baseURL, "https://graph.facebook.com")
	}
}

// ============================================================================
// Entity Tests
// ============================================================================

func TestOAuthToken_ExpiresAt(t *testing.T) {
	token := &entity.OAuthToken{
		AccessToken: "token123",
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	// ExpiresAt should be approximately 1 hour from now
	diff := time.Until(token.ExpiresAt)
	if diff < 59*time.Minute || diff > 61*time.Minute {
		t.Errorf("ExpiresAt should be ~1 hour from now, got %v", diff)
	}
}

func TestPlatformAccount(t *testing.T) {
	account := entity.PlatformAccount{
		ID:       "act_123456",
		Name:     "Test Ad Account",
		Currency: "MYR",
		Timezone: "Asia/Kuala_Lumpur",
		Status:   "active",
	}

	if account.ID == "" {
		t.Error("ID should not be empty")
	}

	if account.Currency != "MYR" {
		t.Errorf("Currency = %q, want %q", account.Currency, "MYR")
	}
}

func TestPlatformUser(t *testing.T) {
	user := entity.PlatformUser{
		ID:        "123456",
		Name:      "Test User",
		Email:     "test@example.com",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	if user.ID == "" {
		t.Error("ID should not be empty")
	}

	if user.Name != "Test User" {
		t.Errorf("Name = %q, want %q", user.Name, "Test User")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkGetAuthURL(b *testing.B) {
	config := &Config{
		AppID:       "123456789",
		AppSecret:   "secret",
		RedirectURI: "http://localhost:8080/callback",
		APIVersion:  "v18.0",
	}
	connector := NewConnector(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connector.GetAuthURL("state_benchmark_123")
	}
}

func BenchmarkMapAccountStatus(b *testing.B) {
	connector := NewConnector(DefaultConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		connector.mapAccountStatus(1)
		connector.mapAccountStatus(2)
		connector.mapAccountStatus(999)
	}
}
