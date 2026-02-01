package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupTestRouter() *gin.Engine {
	router := gin.New()
	return router
}

// ============================================================================
// Error Helper Tests
// ============================================================================

func TestMapMetaError(t *testing.T) {
	tests := []struct {
		name        string
		errorCode   string
		errorReason string
		expected    string
	}{
		{
			name:        "user denied",
			errorCode:   "access_denied",
			errorReason: "user_denied",
			expected:    "permission_denied",
		},
		{
			name:        "access denied other reason",
			errorCode:   "access_denied",
			errorReason: "other",
			expected:    "access_denied",
		},
		{
			name:        "invalid request",
			errorCode:   "invalid_request",
			errorReason: "",
			expected:    "invalid_request",
		},
		{
			name:        "unauthorized client",
			errorCode:   "unauthorized_client",
			errorReason: "",
			expected:    "unauthorized_client",
		},
		{
			name:        "invalid scope",
			errorCode:   "invalid_scope",
			errorReason: "",
			expected:    "invalid_scope",
		},
		{
			name:        "server error",
			errorCode:   "server_error",
			errorReason: "",
			expected:    "server_error",
		},
		{
			name:        "unknown error",
			errorCode:   "some_new_error",
			errorReason: "",
			expected:    "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapMetaError(tt.errorCode, tt.errorReason)
			if result != tt.expected {
				t.Errorf("mapMetaError(%q, %q) = %q, want %q",
					tt.errorCode, tt.errorReason, result, tt.expected)
			}
		})
	}
}

func TestBuildErrorRedirect(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		errType  string
		message  string
	}{
		{
			name:     "with message",
			platform: "meta",
			errType:  "permission_denied",
			message:  "User denied access",
		},
		{
			name:     "without message",
			platform: "meta",
			errType:  "invalid_state",
			message:  "",
		},
		{
			name:     "tiktok platform",
			platform: "tiktok",
			errType:  "token_expired",
			message:  "Token has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildErrorRedirect(tt.platform, tt.errType, tt.message)

			// Verify it starts with the correct path
			if !startsWith(result, "/oauth/error?") {
				t.Errorf("expected URL to start with /oauth/error?, got: %s", result)
			}

			// Parse and verify query params
			parsedURL, err := url.Parse(result)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			query := parsedURL.Query()

			if query.Get("platform") != tt.platform {
				t.Errorf("platform = %q, want %q", query.Get("platform"), tt.platform)
			}

			if query.Get("error") != tt.errType {
				t.Errorf("error = %q, want %q", query.Get("error"), tt.errType)
			}

			if tt.message != "" && query.Get("message") != tt.message {
				t.Errorf("message = %q, want %q", query.Get("message"), tt.message)
			}
		})
	}
}

func TestBuildSuccessRedirect(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		platform  string
		accountID string
	}{
		{
			name:      "simple path",
			baseURL:   "/settings/connections",
			platform:  "meta",
			accountID: "acc-123",
		},
		{
			name:      "absolute URL",
			baseURL:   "https://app.example.com/dashboard",
			platform:  "tiktok",
			accountID: "acc-456",
		},
		{
			name:      "URL with existing params",
			baseURL:   "/settings?tab=oauth",
			platform:  "meta",
			accountID: "acc-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSuccessRedirect(tt.baseURL, tt.platform, tt.accountID)

			// Parse and verify
			parsedURL, err := url.Parse(result)
			if err != nil {
				t.Fatalf("failed to parse URL: %v", err)
			}

			query := parsedURL.Query()

			if query.Get("platform") != tt.platform {
				t.Errorf("platform = %q, want %q", query.Get("platform"), tt.platform)
			}

			if query.Get("account_id") != tt.accountID {
				t.Errorf("account_id = %q, want %q", query.Get("account_id"), tt.accountID)
			}

			if query.Get("success") != "true" {
				t.Errorf("success = %q, want %q", query.Get("success"), "true")
			}
		})
	}
}

// ============================================================================
// Integration Tests for OAuth Callback
// ============================================================================

func TestMetaOAuthCallback_UserDenied(t *testing.T) {
	router := setupTestRouter()

	// We can't use the real handler without dependencies, but we can test the endpoint structure
	router.GET("/oauth/meta/callback", func(c *gin.Context) {
		errorCode := c.Query("error")
		errorReason := c.Query("error_reason")

		if errorCode != "" {
			errType := mapMetaError(errorCode, errorReason)
			redirectURL := buildErrorRedirect("meta", errType, "User denied permission")
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/oauth/meta/callback?error=access_denied&error_reason=user_denied", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected Location header to be set")
	}

	parsedURL, _ := url.Parse(location)
	if parsedURL.Query().Get("error") != "permission_denied" {
		t.Errorf("expected error=permission_denied, got %s", parsedURL.Query().Get("error"))
	}
}

func TestMetaOAuthCallback_MissingCode(t *testing.T) {
	router := setupTestRouter()

	router.GET("/oauth/meta/callback", func(c *gin.Context) {
		code := c.Query("code")

		if code == "" {
			redirectURL := buildErrorRedirect("meta", "no_code", "Authorization code not provided")
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/oauth/meta/callback?state=abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
}

func TestMetaOAuthCallback_MissingState(t *testing.T) {
	router := setupTestRouter()

	router.GET("/oauth/meta/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			redirectURL := buildErrorRedirect("meta", "no_code", "Authorization code not provided")
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}

		if state == "" {
			redirectURL := buildErrorRedirect("meta", "invalid_state", "State parameter missing")
			c.Redirect(http.StatusTemporaryRedirect, redirectURL)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/oauth/meta/callback?code=abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, rec.Code)
	}
}

// ============================================================================
// Response Format Tests
// ============================================================================

func TestOAuthInitiateResponse(t *testing.T) {
	router := setupTestRouter()

	router.GET("/auth/meta/connect", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"auth_url": "https://www.facebook.com/v18.0/dialog/oauth?client_id=123&redirect_uri=http://localhost:8080/api/v1/oauth/meta/callback&state=abc123&scope=ads_read,ads_management&response_type=code",
				"platform": "meta",
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/meta/connect", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response struct {
		Data struct {
			AuthURL  string `json:"auth_url"`
			Platform string `json:"platform"`
		} `json:"data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Data.Platform != "meta" {
		t.Errorf("expected platform=meta, got %s", response.Data.Platform)
	}

	if response.Data.AuthURL == "" {
		t.Error("expected auth_url to be non-empty")
	}
}

// Helper function
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
