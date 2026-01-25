package entity

import (
	"time"

	"github.com/google/uuid"
)

// Platform represents the advertising platform type
type Platform string

const (
	PlatformMeta   Platform = "meta"
	PlatformTikTok Platform = "tiktok"
	PlatformShopee Platform = "shopee"
)

// String returns the string representation of the platform
func (p Platform) String() string {
	return string(p)
}

// IsValid checks if the platform is valid
func (p Platform) IsValid() bool {
	switch p {
	case PlatformMeta, PlatformTikTok, PlatformShopee:
		return true
	default:
		return false
	}
}

// AccountStatus represents the status of a connected account
type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusExpired  AccountStatus = "expired"
	AccountStatusRevoked  AccountStatus = "revoked"
)

// CampaignStatus represents the status of a campaign
type CampaignStatus string

const (
	CampaignStatusActive   CampaignStatus = "active"
	CampaignStatusPaused   CampaignStatus = "paused"
	CampaignStatusDeleted  CampaignStatus = "deleted"
	CampaignStatusArchived CampaignStatus = "archived"
	CampaignStatusDraft    CampaignStatus = "draft"
)

// CampaignObjective represents the objective of a campaign
type CampaignObjective string

const (
	ObjectiveAwareness    CampaignObjective = "awareness"
	ObjectiveTraffic      CampaignObjective = "traffic"
	ObjectiveEngagement   CampaignObjective = "engagement"
	ObjectiveLeads        CampaignObjective = "leads"
	ObjectiveAppPromotion CampaignObjective = "app_promotion"
	ObjectiveSales        CampaignObjective = "sales"
	ObjectiveConversions  CampaignObjective = "conversions"
	ObjectiveVideoViews   CampaignObjective = "video_views"
	ObjectiveMessages     CampaignObjective = "messages"
	ObjectiveStoreTraffic CampaignObjective = "store_traffic"
)

// UserRole represents the role of a user in an organization
type UserRole string

const (
	RoleOwner   UserRole = "owner"
	RoleAdmin   UserRole = "admin"
	RoleAnalyst UserRole = "analyst"
	RoleViewer  UserRole = "viewer"
)

// SubscriptionPlan represents the subscription tier
type SubscriptionPlan string

const (
	PlanFree         SubscriptionPlan = "free"
	PlanStarter      SubscriptionPlan = "starter"
	PlanProfessional SubscriptionPlan = "professional"
	PlanEnterprise   SubscriptionPlan = "enterprise"
)

// DateRange represents a date range for queries
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// NewDateRange creates a new date range
func NewDateRange(start, end time.Time) DateRange {
	return DateRange{
		StartDate: start,
		EndDate:   end,
	}
}

// Last7Days returns a date range for the last 7 days
func Last7Days() DateRange {
	now := time.Now()
	return DateRange{
		StartDate: now.AddDate(0, 0, -7),
		EndDate:   now,
	}
}

// Last30Days returns a date range for the last 30 days
func Last30Days() DateRange {
	now := time.Now()
	return DateRange{
		StartDate: now.AddDate(0, 0, -30),
		EndDate:   now,
	}
}

// BaseEntity contains common fields for all entities
type BaseEntity struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewBaseEntity creates a new base entity with a generated UUID
func NewBaseEntity() BaseEntity {
	return BaseEntity{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Organization represents a tenant in the multi-tenant system
type Organization struct {
	BaseEntity
	Name                  string           `json:"name" gorm:"size:255;not null"`
	Slug                  string           `json:"slug" gorm:"size:100;unique;not null"`
	LogoURL               string           `json:"logo_url,omitempty" gorm:"size:500"`
	SubscriptionPlan      SubscriptionPlan `json:"subscription_plan" gorm:"type:subscription_plan;default:'free'"`
	SubscriptionExpiresAt *time.Time       `json:"subscription_expires_at,omitempty"`
	Settings              JSONMap          `json:"settings,omitempty" gorm:"type:jsonb;default:'{}'"`
	Metadata              JSONMap          `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
	IsActive              bool             `json:"is_active" gorm:"default:true"`
}

// User represents a user account
type User struct {
	BaseEntity
	Email           string     `json:"email" gorm:"size:255;unique;not null"`
	PasswordHash    string     `json:"-" gorm:"size:255"`
	FirstName       string     `json:"first_name,omitempty" gorm:"size:100"`
	LastName        string     `json:"last_name,omitempty" gorm:"size:100"`
	AvatarURL       string     `json:"avatar_url,omitempty" gorm:"size:500"`
	Phone           string     `json:"phone,omitempty" gorm:"size:20"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	Metadata        JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	BaseEntity
	OrganizationID uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Role           UserRole   `json:"role" gorm:"type:user_role;default:'viewer'"`
	InvitedBy      *uuid.UUID `json:"invited_by,omitempty" gorm:"type:uuid"`
	InvitedAt      *time.Time `json:"invited_at,omitempty"`
	JoinedAt       *time.Time `json:"joined_at,omitempty"`
	IsActive       bool       `json:"is_active" gorm:"default:true"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// JSONMap is a helper type for JSONB columns
type JSONMap map[string]interface{}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// Offset returns the offset for the pagination
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// TotalPages returns the total number of pages
func (p *Pagination) TotalPages() int {
	if p.Total == 0 {
		return 0
	}
	pages := int(p.Total) / p.PageSize
	if int(p.Total)%p.PageSize > 0 {
		pages++
	}
	return pages
}

// NewPagination creates a new pagination with defaults
func NewPagination(page, pageSize int) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}
