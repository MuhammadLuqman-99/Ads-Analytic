// Package currency provides currency conversion utilities for the ads analytics platform.
package currency

import (
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
)

// ExchangeRateProvider defines the interface for exchange rate providers
type ExchangeRateProvider interface {
	GetRate(from, to string) (decimal.Decimal, error)
	GetSupportedCurrencies() []string
}

// StaticRateProvider provides static exchange rates
// In production, this should be replaced with a dynamic provider (API-based)
type StaticRateProvider struct {
	mu           sync.RWMutex
	baseCurrency string
	rates        map[string]decimal.Decimal // Rates relative to base currency
}

// DefaultRates returns default exchange rates relative to MYR
// These rates are for development/testing; production should use live rates
func DefaultRates() map[string]decimal.Decimal {
	return map[string]decimal.Decimal{
		"MYR": decimal.NewFromInt(1),         // Base currency
		"USD": decimal.NewFromFloat(4.47),    // 1 USD = 4.47 MYR
		"SGD": decimal.NewFromFloat(3.30),    // 1 SGD = 3.30 MYR
		"EUR": decimal.NewFromFloat(4.82),    // 1 EUR = 4.82 MYR
		"GBP": decimal.NewFromFloat(5.65),    // 1 GBP = 5.65 MYR
		"THB": decimal.NewFromFloat(0.13),    // 1 THB = 0.13 MYR
		"IDR": decimal.NewFromFloat(0.00028), // 1 IDR = 0.00028 MYR
		"PHP": decimal.NewFromFloat(0.077),   // 1 PHP = 0.077 MYR
		"VND": decimal.NewFromFloat(0.00018), // 1 VND = 0.00018 MYR
		"CNY": decimal.NewFromFloat(0.62),    // 1 CNY = 0.62 MYR
		"JPY": decimal.NewFromFloat(0.030),   // 1 JPY = 0.030 MYR
		"KRW": decimal.NewFromFloat(0.0033),  // 1 KRW = 0.0033 MYR
		"INR": decimal.NewFromFloat(0.053),   // 1 INR = 0.053 MYR
		"AUD": decimal.NewFromFloat(2.92),    // 1 AUD = 2.92 MYR
		"HKD": decimal.NewFromFloat(0.57),    // 1 HKD = 0.57 MYR
		"TWD": decimal.NewFromFloat(0.14),    // 1 TWD = 0.14 MYR
	}
}

// NewStaticRateProvider creates a new static rate provider with default rates
func NewStaticRateProvider() *StaticRateProvider {
	return &StaticRateProvider{
		baseCurrency: "MYR",
		rates:        DefaultRates(),
	}
}

// NewStaticRateProviderWithRates creates a provider with custom rates
func NewStaticRateProviderWithRates(baseCurrency string, rates map[string]decimal.Decimal) *StaticRateProvider {
	return &StaticRateProvider{
		baseCurrency: baseCurrency,
		rates:        rates,
	}
}

// GetRate returns the exchange rate from one currency to another
func (p *StaticRateProvider) GetRate(from, to string) (decimal.Decimal, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	// Same currency - rate is 1
	if from == to {
		return decimal.NewFromInt(1), nil
	}

	fromRate, fromOK := p.rates[from]
	toRate, toOK := p.rates[to]

	if !fromOK {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", from)
	}
	if !toOK {
		return decimal.Zero, fmt.Errorf("unsupported currency: %s", to)
	}

	// Convert: first to base currency (MYR), then to target
	// Rate = toRate / fromRate (how many 'to' units per 'from' unit)
	// Actually: fromRate is how many MYR per 1 unit of 'from'
	// toRate is how many MYR per 1 unit of 'to'
	// So to convert X 'from' to 'to': X * fromRate / toRate
	// The rate multiplier is: fromRate / toRate
	if toRate.IsZero() {
		return decimal.Zero, fmt.Errorf("invalid rate for currency: %s", to)
	}

	rate := fromRate.Div(toRate)
	return rate, nil
}

// GetSupportedCurrencies returns a list of supported currency codes
func (p *StaticRateProvider) GetSupportedCurrencies() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	currencies := make([]string, 0, len(p.rates))
	for code := range p.rates {
		currencies = append(currencies, code)
	}
	return currencies
}

// UpdateRate updates a single exchange rate (thread-safe)
func (p *StaticRateProvider) UpdateRate(currency string, rate decimal.Decimal) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rates[strings.ToUpper(currency)] = rate
}

// UpdateRates updates multiple exchange rates (thread-safe)
func (p *StaticRateProvider) UpdateRates(rates map[string]decimal.Decimal) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for code, rate := range rates {
		p.rates[strings.ToUpper(code)] = rate
	}
}

// Converter handles currency conversion operations
type Converter struct {
	provider ExchangeRateProvider
}

// NewConverter creates a new currency converter with the given provider
func NewConverter(provider ExchangeRateProvider) *Converter {
	return &Converter{provider: provider}
}

// NewDefaultConverter creates a converter with the default static rate provider
func NewDefaultConverter() *Converter {
	return &Converter{provider: NewStaticRateProvider()}
}

// Convert converts an amount from one currency to another
func (c *Converter) Convert(amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if amount.IsZero() {
		return decimal.Zero, nil
	}

	rate, err := c.provider.GetRate(from, to)
	if err != nil {
		return decimal.Zero, err
	}

	return amount.Mul(rate), nil
}

// ConvertWithDefault converts an amount, returning a default value on error
func (c *Converter) ConvertWithDefault(amount decimal.Decimal, from, to string, defaultValue decimal.Decimal) decimal.Decimal {
	result, err := c.Convert(amount, from, to)
	if err != nil {
		return defaultValue
	}
	return result
}

// MustConvert converts an amount, panicking on error (use only when error is impossible)
func (c *Converter) MustConvert(amount decimal.Decimal, from, to string) decimal.Decimal {
	result, err := c.Convert(amount, from, to)
	if err != nil {
		panic(err)
	}
	return result
}

// GetRate returns the exchange rate between two currencies
func (c *Converter) GetRate(from, to string) (decimal.Decimal, error) {
	return c.provider.GetRate(from, to)
}

// GetSupportedCurrencies returns supported currency codes
func (c *Converter) GetSupportedCurrencies() []string {
	return c.provider.GetSupportedCurrencies()
}

// IsSupportedCurrency checks if a currency code is supported
func (c *Converter) IsSupportedCurrency(code string) bool {
	for _, supported := range c.GetSupportedCurrencies() {
		if strings.EqualFold(supported, code) {
			return true
		}
	}
	return false
}

// ConversionResult represents the result of a currency conversion
type ConversionResult struct {
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	OriginalCurrency string          `json:"original_currency"`
	ConvertedAmount  decimal.Decimal `json:"converted_amount"`
	TargetCurrency   string          `json:"target_currency"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate"`
}

// ConvertWithDetails returns detailed conversion information
func (c *Converter) ConvertWithDetails(amount decimal.Decimal, from, to string) (*ConversionResult, error) {
	rate, err := c.provider.GetRate(from, to)
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		OriginalAmount:   amount,
		OriginalCurrency: strings.ToUpper(from),
		ConvertedAmount:  amount.Mul(rate),
		TargetCurrency:   strings.ToUpper(to),
		ExchangeRate:     rate,
	}, nil
}

// Global default converter instance
var defaultConverter = NewDefaultConverter()

// Convert converts using the global default converter
func Convert(amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	return defaultConverter.Convert(amount, from, to)
}

// GetRate gets exchange rate using the global default converter
func GetRate(from, to string) (decimal.Decimal, error) {
	return defaultConverter.GetRate(from, to)
}

// GetSupportedCurrencies returns supported currencies from the global converter
func GetSupportedCurrencies() []string {
	return defaultConverter.GetSupportedCurrencies()
}

// IsSupportedCurrency checks currency support using the global converter
func IsSupportedCurrency(code string) bool {
	return defaultConverter.IsSupportedCurrency(code)
}
