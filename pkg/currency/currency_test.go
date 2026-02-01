package currency

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestStaticRateProvider_GetRate(t *testing.T) {
	provider := NewStaticRateProvider()

	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{
			name:    "same currency",
			from:    "MYR",
			to:      "MYR",
			wantErr: false,
		},
		{
			name:    "USD to MYR",
			from:    "USD",
			to:      "MYR",
			wantErr: false,
		},
		{
			name:    "MYR to USD",
			from:    "MYR",
			to:      "USD",
			wantErr: false,
		},
		{
			name:    "USD to EUR (cross rate)",
			from:    "USD",
			to:      "EUR",
			wantErr: false,
		},
		{
			name:    "unsupported source currency",
			from:    "XYZ",
			to:      "MYR",
			wantErr: true,
		},
		{
			name:    "unsupported target currency",
			from:    "MYR",
			to:      "XYZ",
			wantErr: true,
		},
		{
			name:    "case insensitive - lowercase",
			from:    "usd",
			to:      "myr",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := provider.GetRate(tt.from, tt.to)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rate.IsZero() && tt.from != tt.to {
				t.Error("expected non-zero rate")
			}
		})
	}
}

func TestStaticRateProvider_SameCurrency(t *testing.T) {
	provider := NewStaticRateProvider()

	rate, err := provider.GetRate("USD", "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(decimal.NewFromInt(1)) {
		t.Errorf("same currency rate should be 1, got %s", rate.String())
	}
}

func TestConverter_Convert(t *testing.T) {
	converter := NewDefaultConverter()

	tests := []struct {
		name           string
		amount         decimal.Decimal
		from           string
		to             string
		wantErr        bool
		expectPositive bool
	}{
		{
			name:           "zero amount",
			amount:         decimal.Zero,
			from:           "USD",
			to:             "MYR",
			wantErr:        false,
			expectPositive: false,
		},
		{
			name:           "USD to MYR positive",
			amount:         decimal.NewFromInt(100),
			from:           "USD",
			to:             "MYR",
			wantErr:        false,
			expectPositive: true,
		},
		{
			name:           "MYR to USD positive",
			amount:         decimal.NewFromInt(447),
			from:           "MYR",
			to:             "USD",
			wantErr:        false,
			expectPositive: true,
		},
		{
			name:           "invalid currency",
			amount:         decimal.NewFromInt(100),
			from:           "INVALID",
			to:             "MYR",
			wantErr:        true,
			expectPositive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.amount, tt.from, tt.to)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectPositive && !result.IsPositive() {
				t.Errorf("expected positive result, got %s", result.String())
			}
		})
	}
}

func TestConverter_ConvertWithDefault(t *testing.T) {
	converter := NewDefaultConverter()
	defaultValue := decimal.NewFromInt(-1)

	// Valid conversion
	result := converter.ConvertWithDefault(decimal.NewFromInt(100), "USD", "MYR", defaultValue)
	if result.Equal(defaultValue) {
		t.Error("expected valid conversion, not default value")
	}

	// Invalid conversion should return default
	result = converter.ConvertWithDefault(decimal.NewFromInt(100), "INVALID", "MYR", defaultValue)
	if !result.Equal(defaultValue) {
		t.Errorf("expected default value for invalid currency, got %s", result.String())
	}
}

func TestConverter_ConvertWithDetails(t *testing.T) {
	converter := NewDefaultConverter()
	amount := decimal.NewFromInt(100)

	result, err := converter.ConvertWithDetails(amount, "USD", "MYR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.OriginalAmount.Equal(amount) {
		t.Errorf("original amount mismatch: got %s, want %s", result.OriginalAmount.String(), amount.String())
	}

	if result.OriginalCurrency != "USD" {
		t.Errorf("original currency mismatch: got %s, want USD", result.OriginalCurrency)
	}

	if result.TargetCurrency != "MYR" {
		t.Errorf("target currency mismatch: got %s, want MYR", result.TargetCurrency)
	}

	if !result.ConvertedAmount.IsPositive() {
		t.Error("converted amount should be positive")
	}

	if !result.ExchangeRate.IsPositive() {
		t.Error("exchange rate should be positive")
	}
}

func TestConverter_IsSupportedCurrency(t *testing.T) {
	converter := NewDefaultConverter()

	supportedCurrencies := []string{"MYR", "USD", "SGD", "EUR", "GBP", "THB", "IDR"}
	for _, currency := range supportedCurrencies {
		if !converter.IsSupportedCurrency(currency) {
			t.Errorf("%s should be supported", currency)
		}
	}

	unsupportedCurrencies := []string{"XYZ", "ABC", "FOO"}
	for _, currency := range unsupportedCurrencies {
		if converter.IsSupportedCurrency(currency) {
			t.Errorf("%s should not be supported", currency)
		}
	}
}

func TestStaticRateProvider_UpdateRate(t *testing.T) {
	provider := NewStaticRateProvider()

	// Update an existing rate
	newRate := decimal.NewFromFloat(5.0)
	provider.UpdateRate("USD", newRate)

	rate, err := provider.GetRate("USD", "MYR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Rate should be updated (1 USD = 5 MYR now)
	// When converting to MYR, the rate is fromRate/toRate = 5/1 = 5
	if !rate.Equal(newRate) {
		t.Errorf("rate not updated correctly: got %s, want %s", rate.String(), newRate.String())
	}
}

func TestStaticRateProvider_UpdateRates(t *testing.T) {
	provider := NewStaticRateProvider()

	newRates := map[string]decimal.Decimal{
		"USD": decimal.NewFromFloat(5.0),
		"EUR": decimal.NewFromFloat(6.0),
	}
	provider.UpdateRates(newRates)

	usdRate, _ := provider.GetRate("USD", "MYR")
	if !usdRate.Equal(decimal.NewFromFloat(5.0)) {
		t.Errorf("USD rate not updated: got %s", usdRate.String())
	}

	eurRate, _ := provider.GetRate("EUR", "MYR")
	if !eurRate.Equal(decimal.NewFromFloat(6.0)) {
		t.Errorf("EUR rate not updated: got %s", eurRate.String())
	}
}

func TestGlobalConverterFunctions(t *testing.T) {
	// Test global Convert function
	result, err := Convert(decimal.NewFromInt(100), "USD", "MYR")
	if err != nil {
		t.Errorf("global Convert failed: %v", err)
	}
	if !result.IsPositive() {
		t.Error("expected positive result from global Convert")
	}

	// Test global GetRate function
	rate, err := GetRate("USD", "MYR")
	if err != nil {
		t.Errorf("global GetRate failed: %v", err)
	}
	if !rate.IsPositive() {
		t.Error("expected positive rate from global GetRate")
	}

	// Test global GetSupportedCurrencies
	currencies := GetSupportedCurrencies()
	if len(currencies) == 0 {
		t.Error("expected supported currencies from global function")
	}

	// Test global IsSupportedCurrency
	if !IsSupportedCurrency("MYR") {
		t.Error("MYR should be supported")
	}
}

func TestCrossRateConversion(t *testing.T) {
	converter := NewDefaultConverter()

	// Convert USD to EUR (cross rate through MYR)
	amount := decimal.NewFromInt(100)
	result, err := converter.Convert(amount, "USD", "EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get rates for verification
	usdToMYR, _ := converter.GetRate("USD", "MYR")
	eurToMYR, _ := converter.GetRate("EUR", "MYR")

	// 100 USD -> MYR -> EUR
	// Expected: 100 * usdToMYR / eurToMYR
	expected := amount.Mul(usdToMYR).Div(eurToMYR)

	// Should be approximately equal (using 4 decimal places)
	if !result.Round(4).Equal(expected.Round(4)) {
		t.Errorf("cross rate conversion mismatch: got %s, expected approx %s", result.String(), expected.String())
	}
}

func TestRoundTripConversion(t *testing.T) {
	converter := NewDefaultConverter()

	// Start with 100 USD
	original := decimal.NewFromInt(100)

	// Convert USD -> MYR
	inMYR, err := converter.Convert(original, "USD", "MYR")
	if err != nil {
		t.Fatalf("USD to MYR failed: %v", err)
	}

	// Convert MYR -> USD
	backToUSD, err := converter.Convert(inMYR, "MYR", "USD")
	if err != nil {
		t.Fatalf("MYR to USD failed: %v", err)
	}

	// Should get back approximately the same amount (within rounding)
	if !original.Round(2).Equal(backToUSD.Round(2)) {
		t.Errorf("round trip conversion failed: started with %s, ended with %s", original.String(), backToUSD.String())
	}
}
