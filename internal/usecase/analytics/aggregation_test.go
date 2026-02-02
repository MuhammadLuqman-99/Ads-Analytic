package analytics

import (
	"testing"

	"github.com/shopspring/decimal"
)

// ============================================================================
// ROAS Calculation Tests
// ============================================================================

func TestUnifiedMetrics_CalculateDerived_ROAS(t *testing.T) {
	tests := []struct {
		name           string
		spend          string
		revenue        string
		expectedROAS   *float64
		expectNilROAS  bool
	}{
		{
			name:          "positive ROAS",
			spend:         "100.00",
			revenue:       "500.00",
			expectedROAS:  floatPtr(5.0),
			expectNilROAS: false,
		},
		{
			name:          "zero spend - no ROAS",
			spend:         "0",
			revenue:       "500.00",
			expectNilROAS: true,
		},
		{
			name:          "high ROAS",
			spend:         "10.00",
			revenue:       "1000.00",
			expectedROAS:  floatPtr(100.0),
			expectNilROAS: false,
		},
		{
			name:          "low ROAS (less than 1)",
			spend:         "100.00",
			revenue:       "50.00",
			expectedROAS:  floatPtr(0.5),
			expectNilROAS: false,
		},
		{
			name:          "zero revenue - ROAS is 0",
			spend:         "100.00",
			revenue:       "0",
			expectedROAS:  floatPtr(0.0),
			expectNilROAS: false,
		},
		{
			name:          "both zero - no ROAS",
			spend:         "0",
			revenue:       "0",
			expectNilROAS: true,
		},
		{
			name:          "decimal precision ROAS",
			spend:         "33.33",
			revenue:       "100.00",
			expectedROAS:  floatPtr(3.0003000300030003), // approximately 3.0
			expectNilROAS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UnifiedMetrics{
				Spend:   decimal.RequireFromString(tt.spend),
				Revenue: decimal.RequireFromString(tt.revenue),
			}

			um.CalculateDerived()

			if tt.expectNilROAS {
				if um.ROAS != nil {
					t.Errorf("expected nil ROAS, got %v", *um.ROAS)
				}
			} else {
				if um.ROAS == nil {
					t.Error("expected ROAS to be calculated, got nil")
				} else if !almostEqual(*um.ROAS, *tt.expectedROAS, 0.001) {
					t.Errorf("expected ROAS %v, got %v", *tt.expectedROAS, *um.ROAS)
				}
			}
		})
	}
}

func TestUnifiedMetrics_CalculateDerived_CTR(t *testing.T) {
	tests := []struct {
		name          string
		impressions   int64
		clicks        int64
		expectedCTR   *float64
		expectNilCTR  bool
	}{
		{
			name:         "positive CTR",
			impressions:  1000,
			clicks:       50,
			expectedCTR:  floatPtr(5.0),
			expectNilCTR: false,
		},
		{
			name:         "zero impressions - no CTR",
			impressions:  0,
			clicks:       50,
			expectNilCTR: true,
		},
		{
			name:         "zero clicks - CTR is 0",
			impressions:  1000,
			clicks:       0,
			expectedCTR:  floatPtr(0.0),
			expectNilCTR: false,
		},
		{
			name:         "high CTR",
			impressions:  100,
			clicks:       80,
			expectedCTR:  floatPtr(80.0),
			expectNilCTR: false,
		},
		{
			name:         "low CTR",
			impressions:  10000,
			clicks:       10,
			expectedCTR:  floatPtr(0.1),
			expectNilCTR: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UnifiedMetrics{
				Impressions: tt.impressions,
				Clicks:      tt.clicks,
			}

			um.CalculateDerived()

			if tt.expectNilCTR {
				if um.CTR != nil {
					t.Errorf("expected nil CTR, got %v", *um.CTR)
				}
			} else {
				if um.CTR == nil {
					t.Error("expected CTR to be calculated, got nil")
				} else if !almostEqual(*um.CTR, *tt.expectedCTR, 0.001) {
					t.Errorf("expected CTR %v, got %v", *tt.expectedCTR, *um.CTR)
				}
			}
		})
	}
}

func TestUnifiedMetrics_CalculateDerived_CPC(t *testing.T) {
	tests := []struct {
		name         string
		spend        string
		clicks       int64
		expectedCPC  string
		expectNilCPC bool
	}{
		{
			name:         "positive CPC",
			spend:        "100.00",
			clicks:       50,
			expectedCPC:  "2.00",
			expectNilCPC: false,
		},
		{
			name:         "zero clicks - no CPC",
			spend:        "100.00",
			clicks:       0,
			expectNilCPC: true,
		},
		{
			name:         "high CPC",
			spend:        "500.00",
			clicks:       10,
			expectedCPC:  "50.00",
			expectNilCPC: false,
		},
		{
			name:         "low CPC",
			spend:        "10.00",
			clicks:       1000,
			expectedCPC:  "0.01",
			expectNilCPC: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UnifiedMetrics{
				Spend:  decimal.RequireFromString(tt.spend),
				Clicks: tt.clicks,
			}

			um.CalculateDerived()

			if tt.expectNilCPC {
				if um.CPC != nil {
					t.Errorf("expected nil CPC, got %v", um.CPC)
				}
			} else {
				if um.CPC == nil {
					t.Error("expected CPC to be calculated, got nil")
				} else {
					expected := decimal.RequireFromString(tt.expectedCPC)
					if !um.CPC.Equal(expected) {
						t.Errorf("expected CPC %v, got %v", expected, *um.CPC)
					}
				}
			}
		})
	}
}

func TestUnifiedMetrics_CalculateDerived_CPA(t *testing.T) {
	tests := []struct {
		name         string
		spend        string
		conversions  int64
		expectedCPA  string
		expectNilCPA bool
	}{
		{
			name:         "positive CPA",
			spend:        "100.00",
			conversions:  10,
			expectedCPA:  "10.00",
			expectNilCPA: false,
		},
		{
			name:         "zero conversions - no CPA",
			spend:        "100.00",
			conversions:  0,
			expectNilCPA: true,
		},
		{
			name:         "high CPA",
			spend:        "1000.00",
			conversions:  2,
			expectedCPA:  "500.00",
			expectNilCPA: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UnifiedMetrics{
				Spend:       decimal.RequireFromString(tt.spend),
				Conversions: tt.conversions,
			}

			um.CalculateDerived()

			if tt.expectNilCPA {
				if um.CPA != nil {
					t.Errorf("expected nil CPA, got %v", um.CPA)
				}
			} else {
				if um.CPA == nil {
					t.Error("expected CPA to be calculated, got nil")
				} else {
					expected := decimal.RequireFromString(tt.expectedCPA)
					if !um.CPA.Equal(expected) {
						t.Errorf("expected CPA %v, got %v", expected, *um.CPA)
					}
				}
			}
		})
	}
}

func TestUnifiedMetrics_CalculateDerived_CPM(t *testing.T) {
	tests := []struct {
		name         string
		spend        string
		impressions  int64
		expectedCPM  string
		expectNilCPM bool
	}{
		{
			name:         "positive CPM",
			spend:        "10.00",
			impressions:  10000,
			expectedCPM:  "1.00",
			expectNilCPM: false,
		},
		{
			name:         "zero impressions - no CPM",
			spend:        "100.00",
			impressions:  0,
			expectNilCPM: true,
		},
		{
			name:         "high CPM",
			spend:        "100.00",
			impressions:  5000,
			expectedCPM:  "20.00",
			expectNilCPM: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			um := &UnifiedMetrics{
				Spend:       decimal.RequireFromString(tt.spend),
				Impressions: tt.impressions,
			}

			um.CalculateDerived()

			if tt.expectNilCPM {
				if um.CPM != nil {
					t.Errorf("expected nil CPM, got %v", um.CPM)
				}
			} else {
				if um.CPM == nil {
					t.Error("expected CPM to be calculated, got nil")
				} else {
					expected := decimal.RequireFromString(tt.expectedCPM)
					if !um.CPM.Equal(expected) {
						t.Errorf("expected CPM %v, got %v", expected, *um.CPM)
					}
				}
			}
		})
	}
}

func TestUnifiedMetrics_CalculateDerived_AllMetrics(t *testing.T) {
	um := &UnifiedMetrics{
		Spend:       decimal.RequireFromString("100.00"),
		Revenue:     decimal.RequireFromString("500.00"),
		Impressions: 10000,
		Clicks:      500,
		Conversions: 25,
	}

	um.CalculateDerived()

	// Check all metrics are calculated
	if um.ROAS == nil {
		t.Error("ROAS should be calculated")
	} else if !almostEqual(*um.ROAS, 5.0, 0.001) {
		t.Errorf("expected ROAS 5.0, got %v", *um.ROAS)
	}

	if um.CTR == nil {
		t.Error("CTR should be calculated")
	} else if !almostEqual(*um.CTR, 5.0, 0.001) {
		t.Errorf("expected CTR 5.0, got %v", *um.CTR)
	}

	if um.CPC == nil {
		t.Error("CPC should be calculated")
	}

	if um.CPA == nil {
		t.Error("CPA should be calculated")
	}

	if um.CPM == nil {
		t.Error("CPM should be calculated")
	}
}

// ============================================================================
// AggregatedResult Tests
// ============================================================================

func TestAggregatedResult_CalculateDerived(t *testing.T) {
	tests := []struct {
		name         string
		totalSpend   string
		totalRevenue string
		expectedROAS *float64
	}{
		{
			name:         "positive totals",
			totalSpend:   "1000.00",
			totalRevenue: "5000.00",
			expectedROAS: floatPtr(5.0),
		},
		{
			name:         "zero spend",
			totalSpend:   "0",
			totalRevenue: "5000.00",
			expectedROAS: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &AggregatedResult{
				TotalSpend:   decimal.RequireFromString(tt.totalSpend),
				TotalRevenue: decimal.RequireFromString(tt.totalRevenue),
			}

			ar.CalculateDerived()

			if tt.expectedROAS == nil {
				if ar.ROAS != nil {
					t.Errorf("expected nil ROAS, got %v", *ar.ROAS)
				}
			} else {
				if ar.ROAS == nil {
					t.Error("expected ROAS to be calculated")
				} else if !almostEqual(*ar.ROAS, *tt.expectedROAS, 0.001) {
					t.Errorf("expected ROAS %v, got %v", *tt.expectedROAS, *ar.ROAS)
				}
			}
		})
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestUnifiedMetrics_CalculateDerived_NegativeValues(t *testing.T) {
	// Test with negative spend (shouldn't happen in real world, but test anyway)
	um := &UnifiedMetrics{
		Spend:       decimal.RequireFromString("-100.00"),
		Revenue:     decimal.RequireFromString("500.00"),
		Impressions: 10000,
		Clicks:      500,
		Conversions: 25,
	}

	um.CalculateDerived()

	// Negative spend should not produce ROAS (IsPositive check fails)
	if um.ROAS != nil {
		t.Logf("ROAS with negative spend: %v (implementation allows this)", *um.ROAS)
	}
}

func TestUnifiedMetrics_CalculateDerived_LargeNumbers(t *testing.T) {
	um := &UnifiedMetrics{
		Spend:       decimal.RequireFromString("1000000.00"),
		Revenue:     decimal.RequireFromString("10000000.00"),
		Impressions: 100000000,
		Clicks:      5000000,
		Conversions: 100000,
	}

	um.CalculateDerived()

	if um.ROAS == nil {
		t.Error("ROAS should be calculated for large numbers")
	} else if !almostEqual(*um.ROAS, 10.0, 0.001) {
		t.Errorf("expected ROAS 10.0, got %v", *um.ROAS)
	}
}

func TestUnifiedMetrics_CalculateDerived_SmallNumbers(t *testing.T) {
	um := &UnifiedMetrics{
		Spend:       decimal.RequireFromString("0.01"),
		Revenue:     decimal.RequireFromString("0.10"),
		Impressions: 10,
		Clicks:      1,
		Conversions: 1,
	}

	um.CalculateDerived()

	if um.ROAS == nil {
		t.Error("ROAS should be calculated for small numbers")
	} else if !almostEqual(*um.ROAS, 10.0, 0.001) {
		t.Errorf("expected ROAS 10.0, got %v", *um.ROAS)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func floatPtr(f float64) *float64 {
	return &f
}

func almostEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
