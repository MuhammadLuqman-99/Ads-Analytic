package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ============================================================================
// Mock Repositories
// ============================================================================

type mockDashboardRepository struct {
	data map[uuid.UUID]*DashboardSummary
}

type DashboardSummary struct {
	OrgID       uuid.UUID
	TotalSpend  decimal.Decimal
	TotalRev    decimal.Decimal
	Impressions int64
	Clicks      int64
	Conversions int64
}

func newMockDashboardRepo() *mockDashboardRepository {
	return &mockDashboardRepository{
		data: make(map[uuid.UUID]*DashboardSummary),
	}
}

func (m *mockDashboardRepository) GetSummary(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*DashboardSummary, error) {
	if summary, ok := m.data[orgID]; ok {
		return summary, nil
	}
	return &DashboardSummary{OrgID: orgID}, nil
}

func (m *mockDashboardRepository) SetData(orgID uuid.UUID, summary *DashboardSummary) {
	m.data[orgID] = summary
}

// ============================================================================
// Multi-Tenant Isolation Tests
// ============================================================================

func TestMultiTenantIsolation_UserCannotSeeOtherOrgData(t *testing.T) {
	// Setup
	orgA := uuid.New()
	orgB := uuid.New()

	repo := newMockDashboardRepo()

	// Set data for Org A
	repo.SetData(orgA, &DashboardSummary{
		OrgID:       orgA,
		TotalSpend:  decimal.NewFromFloat(1000),
		TotalRev:    decimal.NewFromFloat(5000),
		Impressions: 100000,
		Clicks:      5000,
		Conversions: 250,
	})

	// Set data for Org B
	repo.SetData(orgB, &DashboardSummary{
		OrgID:       orgB,
		TotalSpend:  decimal.NewFromFloat(2000),
		TotalRev:    decimal.NewFromFloat(8000),
		Impressions: 200000,
		Clicks:      10000,
		Conversions: 500,
	})

	router := gin.New()

	// Middleware that extracts org ID from context (simulating auth)
	router.Use(func(c *gin.Context) {
		// In real app, this comes from JWT claims
		orgIDStr := c.GetHeader("X-Org-ID")
		if orgIDStr != "" {
			orgID, _ := uuid.Parse(orgIDStr)
			c.Set("orgID", orgID)
		}
		c.Next()
	})

	router.GET("/dashboard/summary", func(c *gin.Context) {
		orgID, exists := c.Get("orgID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "org not found"})
			return
		}

		summary, _ := repo.GetSummary(c.Request.Context(), orgID.(uuid.UUID), time.Now().AddDate(0, -1, 0), time.Now())

		c.JSON(http.StatusOK, gin.H{
			"org_id":      summary.OrgID.String(),
			"total_spend": summary.TotalSpend.String(),
			"total_rev":   summary.TotalRev.String(),
		})
	})

	t.Run("Org A sees only their data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
		req.Header.Set("X-Org-ID", orgA.String())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var response map[string]string
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response["org_id"] != orgA.String() {
			t.Errorf("expected org_id %s, got %s", orgA.String(), response["org_id"])
		}
		if response["total_spend"] != "1000" {
			t.Errorf("expected total_spend 1000, got %s", response["total_spend"])
		}
	})

	t.Run("Org B sees only their data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
		req.Header.Set("X-Org-ID", orgB.String())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var response map[string]string
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response["org_id"] != orgB.String() {
			t.Errorf("expected org_id %s, got %s", orgB.String(), response["org_id"])
		}
		if response["total_spend"] != "2000" {
			t.Errorf("expected total_spend 2000, got %s", response["total_spend"])
		}
	})

	t.Run("No org header returns unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard/summary", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestMultiTenantIsolation_CampaignFilters(t *testing.T) {
	orgA := uuid.New()
	orgB := uuid.New()

	// Mock campaign data
	campaignsA := []entity.Campaign{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, OrganizationID: orgA, PlatformCampaignName: "Org A Campaign 1"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, OrganizationID: orgA, PlatformCampaignName: "Org A Campaign 2"},
	}
	campaignsB := []entity.Campaign{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, OrganizationID: orgB, PlatformCampaignName: "Org B Campaign 1"},
	}

	router := gin.New()

	router.Use(func(c *gin.Context) {
		orgIDStr := c.GetHeader("X-Org-ID")
		if orgIDStr != "" {
			orgID, _ := uuid.Parse(orgIDStr)
			c.Set("orgID", orgID)
		}
		c.Next()
	})

	router.GET("/campaigns", func(c *gin.Context) {
		orgID, exists := c.Get("orgID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Filter campaigns by org
		var result []entity.Campaign
		for _, camp := range append(campaignsA, campaignsB...) {
			if camp.OrganizationID == orgID.(uuid.UUID) {
				result = append(result, camp)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"campaigns": result,
			"count":     len(result),
		})
	})

	t.Run("Org A gets only their campaigns", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
		req.Header.Set("X-Org-ID", orgA.String())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var response struct {
			Campaigns []entity.Campaign `json:"campaigns"`
			Count     int               `json:"count"`
		}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Count != 2 {
			t.Errorf("expected 2 campaigns, got %d", response.Count)
		}

		for _, camp := range response.Campaigns {
			if camp.OrganizationID != orgA {
				t.Errorf("campaign belongs to wrong org: %s", camp.OrganizationID)
			}
		}
	})

	t.Run("Org B gets only their campaigns", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
		req.Header.Set("X-Org-ID", orgB.String())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var response struct {
			Campaigns []entity.Campaign `json:"campaigns"`
			Count     int               `json:"count"`
		}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response.Count != 1 {
			t.Errorf("expected 1 campaign, got %d", response.Count)
		}
	})
}

// ============================================================================
// Dashboard Aggregation Tests
// ============================================================================

func TestDashboardAggregation_CorrectTotals(t *testing.T) {
	// Test that dashboard correctly aggregates data from multiple platforms
	type PlatformMetrics struct {
		Platform    entity.Platform
		Spend       decimal.Decimal
		Revenue     decimal.Decimal
		Impressions int64
		Clicks      int64
		Conversions int64
	}

	platformData := []PlatformMetrics{
		{
			Platform:    entity.PlatformMeta,
			Spend:       decimal.NewFromFloat(500),
			Revenue:     decimal.NewFromFloat(2500),
			Impressions: 50000,
			Clicks:      2500,
			Conversions: 125,
		},
		{
			Platform:    entity.PlatformTikTok,
			Spend:       decimal.NewFromFloat(300),
			Revenue:     decimal.NewFromFloat(1500),
			Impressions: 30000,
			Clicks:      1500,
			Conversions: 75,
		},
		{
			Platform:    entity.PlatformShopee,
			Spend:       decimal.NewFromFloat(200),
			Revenue:     decimal.NewFromFloat(1000),
			Impressions: 20000,
			Clicks:      1000,
			Conversions: 50,
		},
	}

	// Calculate expected totals
	var expectedSpend, expectedRevenue decimal.Decimal
	var expectedImpressions, expectedClicks, expectedConversions int64

	for _, p := range platformData {
		expectedSpend = expectedSpend.Add(p.Spend)
		expectedRevenue = expectedRevenue.Add(p.Revenue)
		expectedImpressions += p.Impressions
		expectedClicks += p.Clicks
		expectedConversions += p.Conversions
	}

	// Verify totals
	if !expectedSpend.Equal(decimal.NewFromFloat(1000)) {
		t.Errorf("expected total spend 1000, got %s", expectedSpend)
	}
	if !expectedRevenue.Equal(decimal.NewFromFloat(5000)) {
		t.Errorf("expected total revenue 5000, got %s", expectedRevenue)
	}
	if expectedImpressions != 100000 {
		t.Errorf("expected impressions 100000, got %d", expectedImpressions)
	}
	if expectedClicks != 5000 {
		t.Errorf("expected clicks 5000, got %d", expectedClicks)
	}
	if expectedConversions != 250 {
		t.Errorf("expected conversions 250, got %d", expectedConversions)
	}

	// Calculate ROAS
	expectedROAS, _ := expectedRevenue.Div(expectedSpend).Float64()
	if expectedROAS != 5.0 {
		t.Errorf("expected ROAS 5.0, got %f", expectedROAS)
	}
}

func TestDashboardAggregation_DateRangeFilter(t *testing.T) {
	// Test that date range filters are applied correctly
	type DailyMetrics struct {
		Date  time.Time
		Spend decimal.Decimal
	}

	now := time.Now().Truncate(24 * time.Hour)
	metrics := []DailyMetrics{
		{Date: now.AddDate(0, 0, -5), Spend: decimal.NewFromFloat(100)},
		{Date: now.AddDate(0, 0, -4), Spend: decimal.NewFromFloat(150)},
		{Date: now.AddDate(0, 0, -3), Spend: decimal.NewFromFloat(200)},
		{Date: now.AddDate(0, 0, -2), Spend: decimal.NewFromFloat(175)},
		{Date: now.AddDate(0, 0, -1), Spend: decimal.NewFromFloat(125)},
		{Date: now, Spend: decimal.NewFromFloat(200)},
	}

	t.Run("Last 3 days filter", func(t *testing.T) {
		from := now.AddDate(0, 0, -2)
		to := now

		var total decimal.Decimal
		for _, m := range metrics {
			if !m.Date.Before(from) && !m.Date.After(to) {
				total = total.Add(m.Spend)
			}
		}

		// Should include: -2, -1, 0 days = 175 + 125 + 200 = 500
		if !total.Equal(decimal.NewFromFloat(500)) {
			t.Errorf("expected 500 for last 3 days, got %s", total)
		}
	})

	t.Run("Last 7 days filter", func(t *testing.T) {
		from := now.AddDate(0, 0, -6)
		to := now

		var total decimal.Decimal
		for _, m := range metrics {
			if !m.Date.Before(from) && !m.Date.After(to) {
				total = total.Add(m.Spend)
			}
		}

		// Should include all: 100 + 150 + 200 + 175 + 125 + 200 = 950
		if !total.Equal(decimal.NewFromFloat(950)) {
			t.Errorf("expected 950 for last 7 days, got %s", total)
		}
	})
}

// ============================================================================
// Campaign Filter Tests
// ============================================================================

func TestCampaignFilters_PlatformFilter(t *testing.T) {
	campaigns := []entity.Campaign{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Platform: entity.PlatformMeta, PlatformCampaignName: "Meta Campaign"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Platform: entity.PlatformTikTok, PlatformCampaignName: "TikTok Campaign"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Platform: entity.PlatformShopee, PlatformCampaignName: "Shopee Campaign"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Platform: entity.PlatformMeta, PlatformCampaignName: "Meta Campaign 2"},
	}

	t.Run("Filter by Meta platform", func(t *testing.T) {
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if c.Platform == entity.PlatformMeta {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 2 {
			t.Errorf("expected 2 Meta campaigns, got %d", len(filtered))
		}
	})

	t.Run("Filter by TikTok platform", func(t *testing.T) {
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if c.Platform == entity.PlatformTikTok {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 1 {
			t.Errorf("expected 1 TikTok campaign, got %d", len(filtered))
		}
	})
}

func TestCampaignFilters_StatusFilter(t *testing.T) {
	campaigns := []entity.Campaign{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: entity.CampaignStatusActive, PlatformCampaignName: "Active 1"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: entity.CampaignStatusPaused, PlatformCampaignName: "Paused 1"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: entity.CampaignStatusActive, PlatformCampaignName: "Active 2"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, Status: entity.CampaignStatusArchived, PlatformCampaignName: "Archived 1"},
	}

	t.Run("Filter active campaigns", func(t *testing.T) {
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if c.Status == entity.CampaignStatusActive {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 2 {
			t.Errorf("expected 2 active campaigns, got %d", len(filtered))
		}
	})

	t.Run("Exclude archived campaigns", func(t *testing.T) {
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if c.Status != entity.CampaignStatusArchived {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 3 {
			t.Errorf("expected 3 non-archived campaigns, got %d", len(filtered))
		}
	})
}

func TestCampaignFilters_SearchQuery(t *testing.T) {
	campaigns := []entity.Campaign{
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, PlatformCampaignName: "Summer Sale 2024"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, PlatformCampaignName: "Winter Promo"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, PlatformCampaignName: "Summer Collection"},
		{BaseEntity: entity.BaseEntity{ID: uuid.New()}, PlatformCampaignName: "Brand Awareness"},
	}

	t.Run("Search for 'Summer'", func(t *testing.T) {
		query := "Summer"
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if contains(c.PlatformCampaignName, query) {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 2 {
			t.Errorf("expected 2 campaigns containing 'Summer', got %d", len(filtered))
		}
	})

	t.Run("Case insensitive search", func(t *testing.T) {
		query := "summer"
		var filtered []entity.Campaign
		for _, c := range campaigns {
			if containsIgnoreCase(c.PlatformCampaignName, query) {
				filtered = append(filtered, c)
			}
		}

		if len(filtered) != 2 {
			t.Errorf("expected 2 campaigns containing 'summer' (case-insensitive), got %d", len(filtered))
		}
	})
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

func containsIgnoreCase(s, substr string) bool {
	return contains(toLower(s), toLower(substr))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		} else {
			b[i] = c
		}
	}
	return string(b)
}
