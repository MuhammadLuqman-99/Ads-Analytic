package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Subscription Plans
// ============================================================================

// PlanTier represents the subscription tier
type PlanTier string

const (
	PlanTierFree     PlanTier = "free"
	PlanTierPro      PlanTier = "pro"
	PlanTierBusiness PlanTier = "business"
)

// PlanLimits defines the limits for each plan
type PlanLimits struct {
	MaxAdAccounts      int   `json:"max_ad_accounts"`
	DataRetentionDays  int   `json:"data_retention_days"`
	MaxAPICallsPerDay  int64 `json:"max_api_calls_per_day"`
	MaxUsersPerOrg     int   `json:"max_users_per_org"`
	MaxStorageMB       int64 `json:"max_storage_mb"`
	AdvancedAnalytics  bool  `json:"advanced_analytics"`
	CustomReports      bool  `json:"custom_reports"`
	PrioritySupport    bool  `json:"priority_support"`
	WebhooksEnabled    bool  `json:"webhooks_enabled"`
	APIAccessEnabled   bool  `json:"api_access_enabled"`
	WhiteLabelEnabled  bool  `json:"white_label_enabled"`
}

// SubscriptionPlanInfo contains plan details
type SubscriptionPlanInfo struct {
	Tier              PlanTier        `json:"tier"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	PriceMonthlyMYR   decimal.Decimal `json:"price_monthly_myr"`
	PriceYearlyMYR    decimal.Decimal `json:"price_yearly_myr"`
	StripePriceIDMon  string          `json:"stripe_price_id_monthly,omitempty"`
	StripePriceIDYear string          `json:"stripe_price_id_yearly,omitempty"`
	Limits            PlanLimits      `json:"limits"`
	Features          []string        `json:"features"`
	IsPopular         bool            `json:"is_popular"`
}

// GetPlanLimits returns the limits for a given plan tier
func GetPlanLimits(tier PlanTier) PlanLimits {
	switch tier {
	case PlanTierFree:
		return PlanLimits{
			MaxAdAccounts:      1,
			DataRetentionDays:  7,
			MaxAPICallsPerDay:  100,
			MaxUsersPerOrg:     1,
			MaxStorageMB:       100,
			AdvancedAnalytics:  false,
			CustomReports:      false,
			PrioritySupport:    false,
			WebhooksEnabled:    false,
			APIAccessEnabled:   false,
			WhiteLabelEnabled:  false,
		}
	case PlanTierPro:
		return PlanLimits{
			MaxAdAccounts:      5,
			DataRetentionDays:  30,
			MaxAPICallsPerDay:  10000,
			MaxUsersPerOrg:     5,
			MaxStorageMB:       1024, // 1GB
			AdvancedAnalytics:  true,
			CustomReports:      true,
			PrioritySupport:    false,
			WebhooksEnabled:    true,
			APIAccessEnabled:   true,
			WhiteLabelEnabled:  false,
		}
	case PlanTierBusiness:
		return PlanLimits{
			MaxAdAccounts:      -1, // Unlimited
			DataRetentionDays:  90,
			MaxAPICallsPerDay:  100000,
			MaxUsersPerOrg:     -1, // Unlimited
			MaxStorageMB:       10240, // 10GB
			AdvancedAnalytics:  true,
			CustomReports:      true,
			PrioritySupport:    true,
			WebhooksEnabled:    true,
			APIAccessEnabled:   true,
			WhiteLabelEnabled:  true,
		}
	default:
		return GetPlanLimits(PlanTierFree)
	}
}

// GetAllPlans returns all available subscription plans
func GetAllPlans() []SubscriptionPlanInfo {
	return []SubscriptionPlanInfo{
		{
			Tier:            PlanTierFree,
			Name:            "Free",
			Description:     "Perfect for getting started",
			PriceMonthlyMYR: decimal.Zero,
			PriceYearlyMYR:  decimal.Zero,
			Limits:          GetPlanLimits(PlanTierFree),
			Features: []string{
				"1 ad account",
				"7 days data retention",
				"Basic dashboard",
				"Email support",
			},
			IsPopular: false,
		},
		{
			Tier:            PlanTierPro,
			Name:            "Pro",
			Description:     "For growing businesses",
			PriceMonthlyMYR: decimal.NewFromInt(99),
			PriceYearlyMYR:  decimal.NewFromInt(990), // 2 months free
			Limits:          GetPlanLimits(PlanTierPro),
			Features: []string{
				"5 ad accounts",
				"30 days data retention",
				"Advanced analytics",
				"Custom reports",
				"API access",
				"Webhook integrations",
			},
			IsPopular: true,
		},
		{
			Tier:            PlanTierBusiness,
			Name:            "Business",
			Description:     "For agencies and enterprises",
			PriceMonthlyMYR: decimal.NewFromInt(299),
			PriceYearlyMYR:  decimal.NewFromInt(2990), // 2 months free
			Limits:          GetPlanLimits(PlanTierBusiness),
			Features: []string{
				"Unlimited ad accounts",
				"90 days data retention",
				"All Pro features",
				"Priority support",
				"White-label reports",
				"Dedicated account manager",
			},
			IsPopular: false,
		},
	}
}

// ============================================================================
// Subscription Entity
// ============================================================================

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusPastDue   SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusTrialing  SubscriptionStatus = "trialing"
	SubscriptionStatusPaused    SubscriptionStatus = "paused"
	SubscriptionStatusUnpaid    SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

// BillingCycle represents the billing frequency
type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

// Subscription represents an organization's subscription
type Subscription struct {
	BaseEntity
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex"`

	// Plan info
	PlanTier     PlanTier           `json:"plan_tier" gorm:"type:plan_tier;not null;default:'free'"`
	Status       SubscriptionStatus `json:"status" gorm:"type:subscription_status;not null;default:'active'"`
	BillingCycle BillingCycle       `json:"billing_cycle" gorm:"type:billing_cycle"`

	// Stripe integration
	StripeCustomerID     string `json:"stripe_customer_id,omitempty" gorm:"size:255;index"`
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty" gorm:"size:255;index"`
	StripePriceID        string `json:"stripe_price_id,omitempty" gorm:"size:255"`

	// Billing period
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"default:false"`

	// Payment info
	LastPaymentAt     *time.Time      `json:"last_payment_at,omitempty"`
	LastPaymentAmount decimal.Decimal `json:"last_payment_amount,omitempty" gorm:"type:decimal(10,2)"`
	PaymentFailCount  int             `json:"payment_fail_count" gorm:"default:0"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// IsActive returns true if subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// IsPaid returns true if it's a paid subscription
func (s *Subscription) IsPaid() bool {
	return s.PlanTier != PlanTierFree
}

// GetLimits returns the plan limits for this subscription
func (s *Subscription) GetLimits() PlanLimits {
	return GetPlanLimits(s.PlanTier)
}

// DaysUntilExpiry returns days until subscription expires
func (s *Subscription) DaysUntilExpiry() int {
	if s.CurrentPeriodEnd == nil {
		return -1
	}
	days := int(time.Until(*s.CurrentPeriodEnd).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// ============================================================================
// Payment History
// ============================================================================

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
	PaymentStatusDisputed  PaymentStatus = "disputed"
)

// PaymentHistory records all payment transactions
type PaymentHistory struct {
	BaseEntity
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	SubscriptionID uuid.UUID       `json:"subscription_id" gorm:"type:uuid;not null;index"`

	// Payment details
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency    string          `json:"currency" gorm:"size:3;default:'MYR'"`
	Status      PaymentStatus   `json:"status" gorm:"type:payment_status;not null"`
	Description string          `json:"description,omitempty" gorm:"size:500"`

	// Stripe references
	StripePaymentIntentID string `json:"stripe_payment_intent_id,omitempty" gorm:"size:255;index"`
	StripeInvoiceID       string `json:"stripe_invoice_id,omitempty" gorm:"size:255"`
	StripeChargeID        string `json:"stripe_charge_id,omitempty" gorm:"size:255"`

	// Invoice
	InvoiceNumber string     `json:"invoice_number,omitempty" gorm:"size:50"`
	InvoiceURL    string     `json:"invoice_url,omitempty" gorm:"size:500"`
	InvoicePDF    string     `json:"invoice_pdf,omitempty" gorm:"size:500"`

	// Payment method
	PaymentMethod     string `json:"payment_method,omitempty" gorm:"size:50"` // card, fpx, etc.
	PaymentMethodLast4 string `json:"payment_method_last4,omitempty" gorm:"size:4"`
	PaymentMethodBrand string `json:"payment_method_brand,omitempty" gorm:"size:20"`

	// Timestamps
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
	FailedAt    *time.Time `json:"failed_at,omitempty"`
	FailReason  string     `json:"fail_reason,omitempty" gorm:"size:500"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// ============================================================================
// Plan Change Request
// ============================================================================

// PlanChangeType represents the type of plan change
type PlanChangeType string

const (
	PlanChangeUpgrade   PlanChangeType = "upgrade"
	PlanChangeDowngrade PlanChangeType = "downgrade"
	PlanChangeCycle     PlanChangeType = "cycle_change"
)

// PlanChangeRequest represents a pending plan change
type PlanChangeRequest struct {
	BaseEntity
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index"`
	SubscriptionID uuid.UUID `json:"subscription_id" gorm:"type:uuid;not null"`

	// Change details
	ChangeType     PlanChangeType `json:"change_type" gorm:"type:plan_change_type;not null"`
	FromPlan       PlanTier       `json:"from_plan" gorm:"type:plan_tier;not null"`
	ToPlan         PlanTier       `json:"to_plan" gorm:"type:plan_tier;not null"`
	FromCycle      BillingCycle   `json:"from_cycle,omitempty" gorm:"type:billing_cycle"`
	ToCycle        BillingCycle   `json:"to_cycle,omitempty" gorm:"type:billing_cycle"`

	// Scheduling
	EffectiveAt    time.Time  `json:"effective_at" gorm:"not null"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	Status         string     `json:"status" gorm:"size:20;default:'pending'"` // pending, processed, canceled

	// Proration
	ProratedAmount decimal.Decimal `json:"prorated_amount,omitempty" gorm:"type:decimal(10,2)"`

	// Metadata
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
}

// ============================================================================
// Coupon/Discount
// ============================================================================

// Coupon represents a discount coupon
type Coupon struct {
	BaseEntity
	Code           string          `json:"code" gorm:"size:50;unique;not null"`
	Name           string          `json:"name" gorm:"size:100;not null"`
	Description    string          `json:"description,omitempty" gorm:"size:500"`

	// Discount
	DiscountType   string          `json:"discount_type" gorm:"size:20;not null"` // percentage, fixed
	DiscountValue  decimal.Decimal `json:"discount_value" gorm:"type:decimal(10,2);not null"`
	Currency       string          `json:"currency,omitempty" gorm:"size:3"`

	// Limits
	MaxRedemptions   int        `json:"max_redemptions,omitempty"`
	CurrentRedempts  int        `json:"current_redemptions" gorm:"default:0"`
	ValidFrom        *time.Time `json:"valid_from,omitempty"`
	ValidUntil       *time.Time `json:"valid_until,omitempty"`
	ApplicablePlans  []PlanTier `json:"applicable_plans,omitempty" gorm:"type:text[]"`
	DurationMonths   int        `json:"duration_months,omitempty"` // 0 = forever

	// Stripe
	StripeCouponID string `json:"stripe_coupon_id,omitempty" gorm:"size:255"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`
}

// IsValid checks if coupon is valid for use
func (c *Coupon) IsValid() bool {
	if !c.IsActive {
		return false
	}
	now := time.Now()
	if c.ValidFrom != nil && now.Before(*c.ValidFrom) {
		return false
	}
	if c.ValidUntil != nil && now.After(*c.ValidUntil) {
		return false
	}
	if c.MaxRedemptions > 0 && c.CurrentRedempts >= c.MaxRedemptions {
		return false
	}
	return true
}

// ============================================================================
// Credit Balance (for prepaid/credits model)
// ============================================================================

// CreditTransaction represents a credit balance transaction
type CreditTransaction struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`

	// Transaction
	Type           string          `json:"type" gorm:"size:20;not null"` // purchase, usage, refund, bonus
	Amount         decimal.Decimal `json:"amount" gorm:"type:decimal(10,2);not null"` // positive or negative
	BalanceAfter   decimal.Decimal `json:"balance_after" gorm:"type:decimal(10,2);not null"`
	Description    string          `json:"description,omitempty" gorm:"size:500"`

	// Reference
	ReferenceType  string    `json:"reference_type,omitempty" gorm:"size:50"` // payment, api_call, etc.
	ReferenceID    string    `json:"reference_id,omitempty" gorm:"size:255"`

	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}
