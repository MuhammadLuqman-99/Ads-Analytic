package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CampaignMetricsDaily represents daily metrics for a campaign
type CampaignMetricsDaily struct {
	BaseEntity
	CampaignID     uuid.UUID `json:"campaign_id" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	Platform       Platform  `json:"platform" gorm:"type:platform_type;not null"`
	MetricDate     time.Time `json:"metric_date" gorm:"type:date;not null"`

	// Core metrics
	Impressions  int64 `json:"impressions" gorm:"default:0"`
	Reach        int64 `json:"reach" gorm:"default:0"`
	Clicks       int64 `json:"clicks" gorm:"default:0"`
	UniqueClicks int64 `json:"unique_clicks" gorm:"default:0"`

	// Cost metrics
	Spend    decimal.Decimal `json:"spend" gorm:"type:decimal(15,4);default:0"`
	Currency string          `json:"currency" gorm:"size:3;default:'MYR'"`

	// Engagement metrics
	Likes          int64 `json:"likes" gorm:"default:0"`
	Comments       int64 `json:"comments" gorm:"default:0"`
	Shares         int64 `json:"shares" gorm:"default:0"`
	Saves          int64 `json:"saves" gorm:"default:0"`
	VideoViews     int64 `json:"video_views" gorm:"default:0"`
	VideoViewsP25  int64 `json:"video_views_p25" gorm:"default:0"`
	VideoViewsP50  int64 `json:"video_views_p50" gorm:"default:0"`
	VideoViewsP75  int64 `json:"video_views_p75" gorm:"default:0"`
	VideoViewsP100 int64 `json:"video_views_p100" gorm:"default:0"`

	// Conversion metrics
	Conversions       int64           `json:"conversions" gorm:"default:0"`
	ConversionValue   decimal.Decimal `json:"conversion_value" gorm:"type:decimal(15,4);default:0"`
	AddToCart         int64           `json:"add_to_cart" gorm:"default:0"`
	CheckoutInitiated int64           `json:"checkout_initiated" gorm:"default:0"`
	Purchases         int64           `json:"purchases" gorm:"default:0"`
	PurchaseValue     decimal.Decimal `json:"purchase_value" gorm:"type:decimal(15,4);default:0"`

	// Calculated metrics
	CTR  *float64         `json:"ctr,omitempty" gorm:"type:decimal(10,6)"`
	CPC  *decimal.Decimal `json:"cpc,omitempty" gorm:"type:decimal(15,4)"`
	CPM  *decimal.Decimal `json:"cpm,omitempty" gorm:"type:decimal(15,4)"`
	CPA  *decimal.Decimal `json:"cpa,omitempty" gorm:"type:decimal(15,4)"`
	ROAS *float64         `json:"roas,omitempty" gorm:"type:decimal(10,4)"`

	// Platform-specific metrics
	PlatformMetrics JSONMap    `json:"platform_metrics,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastSyncedAt    *time.Time `json:"last_synced_at,omitempty"`
}

// CalculateDerivedMetrics calculates CTR, CPC, CPM, CPA, and ROAS
func (m *CampaignMetricsDaily) CalculateDerivedMetrics() {
	// CTR = Clicks / Impressions * 100
	if m.Impressions > 0 {
		ctr := float64(m.Clicks) / float64(m.Impressions) * 100
		m.CTR = &ctr
	}

	// CPC = Spend / Clicks
	if m.Clicks > 0 {
		cpc := m.Spend.Div(decimal.NewFromInt(m.Clicks))
		m.CPC = &cpc
	}

	// CPM = Spend / Impressions * 1000
	if m.Impressions > 0 {
		cpm := m.Spend.Div(decimal.NewFromInt(m.Impressions)).Mul(decimal.NewFromInt(1000))
		m.CPM = &cpm
	}

	// CPA = Spend / Conversions
	if m.Conversions > 0 {
		cpa := m.Spend.Div(decimal.NewFromInt(m.Conversions))
		m.CPA = &cpa
	}

	// ROAS = ConversionValue / Spend
	if !m.Spend.IsZero() {
		roas, _ := m.ConversionValue.Div(m.Spend).Float64()
		m.ROAS = &roas
	}
}

// AdSetMetricsDaily represents daily metrics for an ad set
type AdSetMetricsDaily struct {
	BaseEntity
	AdSetID        uuid.UUID `json:"ad_set_id" gorm:"type:uuid;not null"`
	CampaignID     uuid.UUID `json:"campaign_id" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	Platform       Platform  `json:"platform" gorm:"type:platform_type;not null"`
	MetricDate     time.Time `json:"metric_date" gorm:"type:date;not null"`

	// Core metrics
	Impressions  int64 `json:"impressions" gorm:"default:0"`
	Reach        int64 `json:"reach" gorm:"default:0"`
	Clicks       int64 `json:"clicks" gorm:"default:0"`
	UniqueClicks int64 `json:"unique_clicks" gorm:"default:0"`

	// Cost metrics
	Spend    decimal.Decimal `json:"spend" gorm:"type:decimal(15,4);default:0"`
	Currency string          `json:"currency" gorm:"size:3;default:'MYR'"`

	// Engagement metrics
	Likes      int64 `json:"likes" gorm:"default:0"`
	Comments   int64 `json:"comments" gorm:"default:0"`
	Shares     int64 `json:"shares" gorm:"default:0"`
	Saves      int64 `json:"saves" gorm:"default:0"`
	VideoViews int64 `json:"video_views" gorm:"default:0"`

	// Conversion metrics
	Conversions     int64           `json:"conversions" gorm:"default:0"`
	ConversionValue decimal.Decimal `json:"conversion_value" gorm:"type:decimal(15,4);default:0"`
	Purchases       int64           `json:"purchases" gorm:"default:0"`
	PurchaseValue   decimal.Decimal `json:"purchase_value" gorm:"type:decimal(15,4);default:0"`

	// Calculated metrics
	CTR  *float64         `json:"ctr,omitempty" gorm:"type:decimal(10,6)"`
	CPC  *decimal.Decimal `json:"cpc,omitempty" gorm:"type:decimal(15,4)"`
	CPM  *decimal.Decimal `json:"cpm,omitempty" gorm:"type:decimal(15,4)"`
	CPA  *decimal.Decimal `json:"cpa,omitempty" gorm:"type:decimal(15,4)"`
	ROAS *float64         `json:"roas,omitempty" gorm:"type:decimal(10,4)"`

	PlatformMetrics JSONMap    `json:"platform_metrics,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastSyncedAt    *time.Time `json:"last_synced_at,omitempty"`
}

// AdMetricsDaily represents daily metrics for an individual ad
type AdMetricsDaily struct {
	BaseEntity
	AdID           uuid.UUID `json:"ad_id" gorm:"type:uuid;not null"`
	AdSetID        uuid.UUID `json:"ad_set_id" gorm:"type:uuid;not null"`
	CampaignID     uuid.UUID `json:"campaign_id" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null"`
	Platform       Platform  `json:"platform" gorm:"type:platform_type;not null"`
	MetricDate     time.Time `json:"metric_date" gorm:"type:date;not null"`

	// Core metrics
	Impressions  int64 `json:"impressions" gorm:"default:0"`
	Reach        int64 `json:"reach" gorm:"default:0"`
	Clicks       int64 `json:"clicks" gorm:"default:0"`
	UniqueClicks int64 `json:"unique_clicks" gorm:"default:0"`

	// Cost metrics
	Spend    decimal.Decimal `json:"spend" gorm:"type:decimal(15,4);default:0"`
	Currency string          `json:"currency" gorm:"size:3;default:'MYR'"`

	// Engagement metrics
	Likes      int64 `json:"likes" gorm:"default:0"`
	Comments   int64 `json:"comments" gorm:"default:0"`
	Shares     int64 `json:"shares" gorm:"default:0"`
	Saves      int64 `json:"saves" gorm:"default:0"`
	VideoViews int64 `json:"video_views" gorm:"default:0"`

	// Conversion metrics
	Conversions     int64           `json:"conversions" gorm:"default:0"`
	ConversionValue decimal.Decimal `json:"conversion_value" gorm:"type:decimal(15,4);default:0"`
	Purchases       int64           `json:"purchases" gorm:"default:0"`
	PurchaseValue   decimal.Decimal `json:"purchase_value" gorm:"type:decimal(15,4);default:0"`

	// Calculated metrics
	CTR  *float64         `json:"ctr,omitempty" gorm:"type:decimal(10,6)"`
	CPC  *decimal.Decimal `json:"cpc,omitempty" gorm:"type:decimal(15,4)"`
	CPM  *decimal.Decimal `json:"cpm,omitempty" gorm:"type:decimal(15,4)"`
	CPA  *decimal.Decimal `json:"cpa,omitempty" gorm:"type:decimal(15,4)"`
	ROAS *float64         `json:"roas,omitempty" gorm:"type:decimal(10,4)"`

	PlatformMetrics JSONMap    `json:"platform_metrics,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastSyncedAt    *time.Time `json:"last_synced_at,omitempty"`
}

// AggregatedMetrics represents aggregated metrics across campaigns/platforms
type AggregatedMetrics struct {
	TotalSpend       decimal.Decimal `json:"total_spend"`
	TotalImpressions int64           `json:"total_impressions"`
	TotalClicks      int64           `json:"total_clicks"`
	TotalConversions int64           `json:"total_conversions"`
	TotalRevenue     decimal.Decimal `json:"total_revenue"`
	AverageCTR       float64         `json:"average_ctr"`
	AverageCPC       decimal.Decimal `json:"average_cpc"`
	AverageCPM       decimal.Decimal `json:"average_cpm"`
	AverageCPA       decimal.Decimal `json:"average_cpa"`
	OverallROAS      float64         `json:"overall_roas"`
	Currency         string          `json:"currency"`
}

// PlatformMetricsSummary represents metrics summary by platform
type PlatformMetricsSummary struct {
	Platform      Platform        `json:"platform"`
	Spend         decimal.Decimal `json:"spend"`
	Impressions   int64           `json:"impressions"`
	Clicks        int64           `json:"clicks"`
	Conversions   int64           `json:"conversions"`
	CTR           float64         `json:"ctr"`
	CPC           decimal.Decimal `json:"cpc"`
	ROAS          float64         `json:"roas"`
	CampaignCount int             `json:"campaign_count"`
}

// DailyMetricsTrend represents daily metrics trend data
type DailyMetricsTrend struct {
	Date        time.Time       `json:"date"`
	Spend       decimal.Decimal `json:"spend"`
	Impressions int64           `json:"impressions"`
	Clicks      int64           `json:"clicks"`
	Conversions int64           `json:"conversions"`
	CTR         float64         `json:"ctr"`
	CPC         decimal.Decimal `json:"cpc"`
	ROAS        float64         `json:"roas"`
}

// MetricsFilter represents filters for querying metrics
type MetricsFilter struct {
	OrganizationID uuid.UUID   `json:"organization_id"`
	CampaignIDs    []uuid.UUID `json:"campaign_ids,omitempty"`
	AdSetIDs       []uuid.UUID `json:"ad_set_ids,omitempty"`
	AdIDs          []uuid.UUID `json:"ad_ids,omitempty"`
	Platforms      []Platform  `json:"platforms,omitempty"`
	DateRange      DateRange   `json:"date_range"`
	GroupBy        string      `json:"group_by,omitempty"` // "day", "week", "month"
}

// MetricsComparison represents a comparison between two periods
type MetricsComparison struct {
	CurrentPeriod     AggregatedMetrics `json:"current_period"`
	PreviousPeriod    AggregatedMetrics `json:"previous_period"`
	SpendChange       float64           `json:"spend_change_percent"`
	ClicksChange      float64           `json:"clicks_change_percent"`
	ConversionsChange float64           `json:"conversions_change_percent"`
	ROASChange        float64           `json:"roas_change_percent"`
}

// CalculateComparison calculates the percentage change between periods
func (m *MetricsComparison) CalculateComparison() {
	if !m.PreviousPeriod.TotalSpend.IsZero() {
		change, _ := m.CurrentPeriod.TotalSpend.Sub(m.PreviousPeriod.TotalSpend).
			Div(m.PreviousPeriod.TotalSpend).Mul(decimal.NewFromInt(100)).Float64()
		m.SpendChange = change
	}

	if m.PreviousPeriod.TotalClicks > 0 {
		m.ClicksChange = float64(m.CurrentPeriod.TotalClicks-m.PreviousPeriod.TotalClicks) /
			float64(m.PreviousPeriod.TotalClicks) * 100
	}

	if m.PreviousPeriod.TotalConversions > 0 {
		m.ConversionsChange = float64(m.CurrentPeriod.TotalConversions-m.PreviousPeriod.TotalConversions) /
			float64(m.PreviousPeriod.TotalConversions) * 100
	}

	if m.PreviousPeriod.OverallROAS > 0 {
		m.ROASChange = (m.CurrentPeriod.OverallROAS - m.PreviousPeriod.OverallROAS) /
			m.PreviousPeriod.OverallROAS * 100
	}
}

// TopPerformer represents a top performing entity
type TopPerformer struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Platform    Platform        `json:"platform"`
	Spend       decimal.Decimal `json:"spend"`
	Impressions int64           `json:"impressions"`
	Clicks      int64           `json:"clicks"`
	Conversions int64           `json:"conversions"`
	ROAS        float64         `json:"roas"`
	CTR         float64         `json:"ctr"`
}
