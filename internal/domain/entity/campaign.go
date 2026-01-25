package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Campaign represents an advertising campaign from any platform
type Campaign struct {
	BaseEntity
	AdAccountID          uuid.UUID         `json:"ad_account_id" gorm:"type:uuid;not null"`
	OrganizationID       uuid.UUID         `json:"organization_id" gorm:"type:uuid;not null"`
	Platform             Platform          `json:"platform" gorm:"type:platform_type;not null"`
	PlatformCampaignID   string            `json:"platform_campaign_id" gorm:"size:255;not null"`
	PlatformCampaignName string            `json:"platform_campaign_name,omitempty" gorm:"size:500"`
	Objective            CampaignObjective `json:"objective,omitempty" gorm:"type:campaign_objective"`
	Status               CampaignStatus    `json:"status" gorm:"type:campaign_status;default:'active'"`
	DailyBudget          *decimal.Decimal  `json:"daily_budget,omitempty" gorm:"type:decimal(15,2)"`
	LifetimeBudget       *decimal.Decimal  `json:"lifetime_budget,omitempty" gorm:"type:decimal(15,2)"`
	BudgetCurrency       string            `json:"budget_currency" gorm:"size:3;default:'MYR'"`
	StartDate            *time.Time        `json:"start_date,omitempty" gorm:"type:date"`
	EndDate              *time.Time        `json:"end_date,omitempty" gorm:"type:date"`
	PlatformData         JSONMap           `json:"platform_data,omitempty" gorm:"type:jsonb;default:'{}'"`
	PlatformCreatedAt    *time.Time        `json:"platform_created_at,omitempty"`
	PlatformUpdatedAt    *time.Time        `json:"platform_updated_at,omitempty"`
	LastSyncedAt         *time.Time        `json:"last_synced_at,omitempty"`

	// Relations
	AdAccount *AdAccount `json:"ad_account,omitempty" gorm:"foreignKey:AdAccountID"`
	AdSets    []AdSet    `json:"ad_sets,omitempty" gorm:"foreignKey:CampaignID"`
	Ads       []Ad       `json:"ads,omitempty" gorm:"foreignKey:CampaignID"`
}

// AdSet represents an ad set/ad group within a campaign
type AdSet struct {
	BaseEntity
	CampaignID        uuid.UUID        `json:"campaign_id" gorm:"type:uuid;not null"`
	OrganizationID    uuid.UUID        `json:"organization_id" gorm:"type:uuid;not null"`
	Platform          Platform         `json:"platform" gorm:"type:platform_type;not null"`
	PlatformAdSetID   string           `json:"platform_ad_set_id" gorm:"size:255;not null"`
	PlatformAdSetName string           `json:"platform_ad_set_name,omitempty" gorm:"size:500"`
	Status            CampaignStatus   `json:"status" gorm:"type:campaign_status;default:'active'"`
	DailyBudget       *decimal.Decimal `json:"daily_budget,omitempty" gorm:"type:decimal(15,2)"`
	LifetimeBudget    *decimal.Decimal `json:"lifetime_budget,omitempty" gorm:"type:decimal(15,2)"`
	BidAmount         *decimal.Decimal `json:"bid_amount,omitempty" gorm:"type:decimal(15,4)"`
	BidStrategy       string           `json:"bid_strategy,omitempty" gorm:"size:100"`
	Targeting         JSONMap          `json:"targeting,omitempty" gorm:"type:jsonb;default:'{}'"`
	StartDate         *time.Time       `json:"start_date,omitempty" gorm:"type:date"`
	EndDate           *time.Time       `json:"end_date,omitempty" gorm:"type:date"`
	PlatformData      JSONMap          `json:"platform_data,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastSyncedAt      *time.Time       `json:"last_synced_at,omitempty"`

	// Relations
	Campaign *Campaign `json:"campaign,omitempty" gorm:"foreignKey:CampaignID"`
	Ads      []Ad      `json:"ads,omitempty" gorm:"foreignKey:AdSetID"`
}

// Ad represents an individual ad with creative details
type Ad struct {
	BaseEntity
	AdSetID        uuid.UUID      `json:"ad_set_id" gorm:"type:uuid;not null"`
	CampaignID     uuid.UUID      `json:"campaign_id" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null"`
	Platform       Platform       `json:"platform" gorm:"type:platform_type;not null"`
	PlatformAdID   string         `json:"platform_ad_id" gorm:"size:255;not null"`
	PlatformAdName string         `json:"platform_ad_name,omitempty" gorm:"size:500"`
	Status         CampaignStatus `json:"status" gorm:"type:campaign_status;default:'active'"`
	Headline       string         `json:"headline,omitempty" gorm:"size:500"`
	Description    string         `json:"description,omitempty" gorm:"type:text"`
	CallToAction   string         `json:"call_to_action,omitempty" gorm:"size:100"`
	DestinationURL string         `json:"destination_url,omitempty" gorm:"type:text"`
	DisplayURL     string         `json:"display_url,omitempty" gorm:"size:255"`
	ImageURL       string         `json:"image_url,omitempty" gorm:"type:text"`
	VideoURL       string         `json:"video_url,omitempty" gorm:"type:text"`
	ThumbnailURL   string         `json:"thumbnail_url,omitempty" gorm:"type:text"`
	CreativeData   JSONMap        `json:"creative_data,omitempty" gorm:"type:jsonb;default:'{}'"`
	PlatformData   JSONMap        `json:"platform_data,omitempty" gorm:"type:jsonb;default:'{}'"`
	LastSyncedAt   *time.Time     `json:"last_synced_at,omitempty"`

	// Relations
	AdSet    *AdSet    `json:"ad_set,omitempty" gorm:"foreignKey:AdSetID"`
	Campaign *Campaign `json:"campaign,omitempty" gorm:"foreignKey:CampaignID"`
}

// CampaignSummary represents a summary view of a campaign
type CampaignSummary struct {
	ID           uuid.UUID         `json:"id"`
	Platform     Platform          `json:"platform"`
	Name         string            `json:"name"`
	Status       CampaignStatus    `json:"status"`
	Objective    CampaignObjective `json:"objective"`
	DailyBudget  *decimal.Decimal  `json:"daily_budget,omitempty"`
	TotalSpend   decimal.Decimal   `json:"total_spend"`
	Impressions  int64             `json:"impressions"`
	Clicks       int64             `json:"clicks"`
	Conversions  int64             `json:"conversions"`
	CTR          float64           `json:"ctr"`
	CPC          decimal.Decimal   `json:"cpc"`
	ROAS         float64           `json:"roas"`
	LastSyncedAt *time.Time        `json:"last_synced_at,omitempty"`
}

// CampaignFilter represents filters for querying campaigns
type CampaignFilter struct {
	OrganizationID uuid.UUID           `json:"organization_id"`
	AdAccountIDs   []uuid.UUID         `json:"ad_account_ids,omitempty"`
	Platforms      []Platform          `json:"platforms,omitempty"`
	Statuses       []CampaignStatus    `json:"statuses,omitempty"`
	Objectives     []CampaignObjective `json:"objectives,omitempty"`
	DateRange      *DateRange          `json:"date_range,omitempty"`
	SearchTerm     string              `json:"search_term,omitempty"`
	Pagination     *Pagination         `json:"pagination,omitempty"`
}

// AdSetFilter represents filters for querying ad sets
type AdSetFilter struct {
	OrganizationID uuid.UUID        `json:"organization_id"`
	CampaignIDs    []uuid.UUID      `json:"campaign_ids,omitempty"`
	Platforms      []Platform       `json:"platforms,omitempty"`
	Statuses       []CampaignStatus `json:"statuses,omitempty"`
	SearchTerm     string           `json:"search_term,omitempty"`
	Pagination     *Pagination      `json:"pagination,omitempty"`
}

// AdFilter represents filters for querying ads
type AdFilter struct {
	OrganizationID uuid.UUID        `json:"organization_id"`
	AdSetIDs       []uuid.UUID      `json:"ad_set_ids,omitempty"`
	CampaignIDs    []uuid.UUID      `json:"campaign_ids,omitempty"`
	Platforms      []Platform       `json:"platforms,omitempty"`
	Statuses       []CampaignStatus `json:"statuses,omitempty"`
	SearchTerm     string           `json:"search_term,omitempty"`
	Pagination     *Pagination      `json:"pagination,omitempty"`
}
