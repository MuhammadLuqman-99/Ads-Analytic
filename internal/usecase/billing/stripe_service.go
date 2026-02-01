package billing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/subscription"
)

// ============================================================================
// Stripe Billing Service
// ============================================================================

// StripeService handles all Stripe billing operations
type StripeService struct {
	subscriptionRepo repository.SubscriptionRepository
	paymentRepo      repository.PaymentHistoryRepository
	usageRepo        repository.UsageRepository
	orgRepo          repository.OrganizationRepository

	// Stripe configuration
	secretKey        string
	webhookSecret    string
	successURL       string
	cancelURL        string
	portalReturnURL  string

	// Price IDs per plan
	priceIDs map[entity.PlanTier]map[entity.BillingCycle]string
}

// StripeConfig holds Stripe configuration
type StripeConfig struct {
	SecretKey       string
	WebhookSecret   string
	SuccessURL      string
	CancelURL       string
	PortalReturnURL string
	PriceIDs        map[entity.PlanTier]map[entity.BillingCycle]string
}

// NewStripeService creates a new Stripe billing service
func NewStripeService(
	subscriptionRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentHistoryRepository,
	usageRepo repository.UsageRepository,
	orgRepo repository.OrganizationRepository,
	config *StripeConfig,
) *StripeService {
	stripe.Key = config.SecretKey

	return &StripeService{
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		usageRepo:        usageRepo,
		orgRepo:          orgRepo,
		secretKey:        config.SecretKey,
		webhookSecret:    config.WebhookSecret,
		successURL:       config.SuccessURL,
		cancelURL:        config.CancelURL,
		portalReturnURL:  config.PortalReturnURL,
		priceIDs:         config.PriceIDs,
	}
}

// ============================================================================
// Customer Management
// ============================================================================

// CreateCustomer creates a Stripe customer for an organization
func (s *StripeService) CreateCustomer(ctx context.Context, orgID uuid.UUID) (*stripe.Customer, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	params := &stripe.CustomerParams{
		Name:  stripe.String(org.Name),
		Email: stripe.String(""), // Will be set from admin user
		Metadata: map[string]string{
			"organization_id": orgID.String(),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Update subscription with customer ID
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err == nil && sub != nil {
		sub.StripeCustomerID = cust.ID
		if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
			log.Printf("[Stripe] Error updating subscription with customer ID: %v", err)
		}
	}

	return cust, nil
}

// GetOrCreateCustomer gets existing customer or creates new one
func (s *StripeService) GetOrCreateCustomer(ctx context.Context, orgID uuid.UUID) (*stripe.Customer, error) {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if sub.StripeCustomerID != "" {
		cust, err := customer.Get(sub.StripeCustomerID, nil)
		if err == nil {
			return cust, nil
		}
	}

	return s.CreateCustomer(ctx, orgID)
}

// ============================================================================
// Checkout Sessions
// ============================================================================

// CreateCheckoutSessionRequest represents checkout session request
type CreateCheckoutSessionRequest struct {
	OrganizationID uuid.UUID
	PlanTier       entity.PlanTier
	BillingCycle   entity.BillingCycle
	CouponCode     string
	CustomerEmail  string
}

// CreateCheckoutSession creates a Stripe checkout session for subscription
func (s *StripeService) CreateCheckoutSession(ctx context.Context, req CreateCheckoutSessionRequest) (*stripe.CheckoutSession, error) {
	// Get or create customer
	cust, err := s.GetOrCreateCustomer(ctx, req.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Get price ID
	priceID, err := s.getPriceID(req.PlanTier, req.BillingCycle)
	if err != nil {
		return nil, err
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(cust.ID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.successURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.cancelURL),
		Metadata: map[string]string{
			"organization_id": req.OrganizationID.String(),
			"plan_tier":       string(req.PlanTier),
			"billing_cycle":   string(req.BillingCycle),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"organization_id": req.OrganizationID.String(),
			},
		},
		// Allow promotion codes
		AllowPromotionCodes: stripe.Bool(true),
	}

	// Add customer email if provided
	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}

	// Add coupon if provided
	if req.CouponCode != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{
				Coupon: stripe.String(req.CouponCode),
			},
		}
		params.AllowPromotionCodes = nil // Can't use both
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess, nil
}

// ============================================================================
// Subscription Management
// ============================================================================

// CancelSubscription cancels a subscription at period end
func (s *StripeService) CancelSubscription(ctx context.Context, orgID uuid.UUID, immediate bool) error {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	if sub.StripeSubscriptionID == "" {
		return fmt.Errorf("no active Stripe subscription")
	}

	if immediate {
		// Cancel immediately
		_, err = subscription.Cancel(sub.StripeSubscriptionID, nil)
		if err != nil {
			return fmt.Errorf("failed to cancel subscription: %w", err)
		}

		now := time.Now()
		sub.Status = entity.SubscriptionStatusCanceled
		sub.CanceledAt = &now
	} else {
		// Cancel at period end
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		_, err = subscription.Update(sub.StripeSubscriptionID, params)
		if err != nil {
			return fmt.Errorf("failed to schedule cancellation: %w", err)
		}

		sub.CancelAtPeriodEnd = true
	}

	return s.subscriptionRepo.Update(ctx, sub)
}

// ReactivateSubscription reactivates a subscription scheduled for cancellation
func (s *StripeService) ReactivateSubscription(ctx context.Context, orgID uuid.UUID) error {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	if sub.StripeSubscriptionID == "" {
		return fmt.Errorf("no Stripe subscription found")
	}

	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}
	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	sub.CancelAtPeriodEnd = false
	return s.subscriptionRepo.Update(ctx, sub)
}

// ChangePlan changes the subscription plan
func (s *StripeService) ChangePlan(ctx context.Context, orgID uuid.UUID, newPlan entity.PlanTier, newCycle entity.BillingCycle) error {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	if sub.StripeSubscriptionID == "" {
		return fmt.Errorf("no active Stripe subscription")
	}

	// Get new price ID
	newPriceID, err := s.getPriceID(newPlan, newCycle)
	if err != nil {
		return err
	}

	// Get current subscription from Stripe
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get Stripe subscription: %w", err)
	}

	if len(stripeSub.Items.Data) == 0 {
		return fmt.Errorf("no subscription items found")
	}

	// Update subscription
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(stripeSub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String(string(stripe.SubscriptionSchedulePhaseProrationBehaviorCreateProrations)),
	}

	_, err = subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update local subscription
	sub.PlanTier = newPlan
	sub.BillingCycle = newCycle
	sub.StripePriceID = newPriceID

	return s.subscriptionRepo.Update(ctx, sub)
}

// ============================================================================
// Payment Methods
// ============================================================================

// GetPaymentMethods gets payment methods for a customer
func (s *StripeService) GetPaymentMethods(ctx context.Context, orgID uuid.UUID) ([]*stripe.PaymentMethod, error) {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if sub.StripeCustomerID == "" {
		return nil, nil
	}

	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(sub.StripeCustomerID),
		Type:     stripe.String("card"),
	}

	var methods []*stripe.PaymentMethod
	iter := paymentmethod.List(params)
	for iter.Next() {
		methods = append(methods, iter.PaymentMethod())
	}

	return methods, iter.Err()
}

// SetDefaultPaymentMethod sets the default payment method
func (s *StripeService) SetDefaultPaymentMethod(ctx context.Context, orgID uuid.UUID, paymentMethodID string) error {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	if sub.StripeCustomerID == "" {
		return fmt.Errorf("no Stripe customer found")
	}

	// Attach payment method to customer
	attachParams := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(sub.StripeCustomerID),
	}
	_, err = paymentmethod.Attach(paymentMethodID, attachParams)
	if err != nil {
		return fmt.Errorf("failed to attach payment method: %w", err)
	}

	// Set as default
	customerParams := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}
	_, err = customer.Update(sub.StripeCustomerID, customerParams)
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}

	return nil
}

// ============================================================================
// Invoices
// ============================================================================

// GetInvoices gets invoices for an organization
func (s *StripeService) GetInvoices(ctx context.Context, orgID uuid.UUID, limit int) ([]*stripe.Invoice, error) {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if sub.StripeCustomerID == "" {
		return nil, nil
	}

	params := &stripe.InvoiceListParams{
		Customer: stripe.String(sub.StripeCustomerID),
	}
	params.Limit = stripe.Int64(int64(limit))

	var invoices []*stripe.Invoice
	iter := invoice.List(params)
	for iter.Next() {
		invoices = append(invoices, iter.Invoice())
	}

	return invoices, iter.Err()
}

// GetUpcomingInvoice gets the upcoming invoice
func (s *StripeService) GetUpcomingInvoice(ctx context.Context, orgID uuid.UUID) (*stripe.Invoice, error) {
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if sub.StripeCustomerID == "" {
		return nil, nil
	}

	params := &stripe.InvoiceUpcomingParams{
		Customer: stripe.String(sub.StripeCustomerID),
	}

	inv, err := invoice.Upcoming(params)
	if err != nil {
		return nil, err
	}

	return inv, nil
}

// ============================================================================
// Webhook Event Handling
// ============================================================================

// HandleCheckoutCompleted handles checkout.session.completed webhook
func (s *StripeService) HandleCheckoutCompleted(ctx context.Context, sess *stripe.CheckoutSession) error {
	orgIDStr := sess.Metadata["organization_id"]
	if orgIDStr == "" {
		return fmt.Errorf("organization_id not found in metadata")
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return fmt.Errorf("invalid organization_id: %w", err)
	}

	planTier := entity.PlanTier(sess.Metadata["plan_tier"])
	billingCycle := entity.BillingCycle(sess.Metadata["billing_cycle"])

	// Update subscription
	sub, err := s.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		return err
	}

	sub.StripeCustomerID = sess.Customer.ID
	sub.StripeSubscriptionID = sess.Subscription.ID
	sub.PlanTier = planTier
	sub.BillingCycle = billingCycle
	sub.Status = entity.SubscriptionStatusActive

	// Set period from Stripe subscription
	stripeSub, err := subscription.Get(sess.Subscription.ID, nil)
	if err == nil {
		periodStart := time.Unix(stripeSub.CurrentPeriodStart, 0)
		periodEnd := time.Unix(stripeSub.CurrentPeriodEnd, 0)
		sub.CurrentPeriodStart = &periodStart
		sub.CurrentPeriodEnd = &periodEnd
	}

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return err
	}

	// Update usage limits
	limits := entity.GetPlanLimits(planTier)
	if err := s.usageRepo.UpdateLimits(ctx, orgID, limits); err != nil {
		log.Printf("[Stripe] Error updating usage limits: %v", err)
	}

	log.Printf("[Stripe] Checkout completed for org %s, plan: %s", orgID, planTier)

	return nil
}

// HandleSubscriptionUpdated handles customer.subscription.updated webhook
func (s *StripeService) HandleSubscriptionUpdated(ctx context.Context, stripeSub *stripe.Subscription) error {
	sub, err := s.subscriptionRepo.GetByStripeSubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}

	// Update status
	sub.Status = mapStripeStatus(stripeSub.Status)

	// Update period
	periodStart := time.Unix(stripeSub.CurrentPeriodStart, 0)
	periodEnd := time.Unix(stripeSub.CurrentPeriodEnd, 0)
	sub.CurrentPeriodStart = &periodStart
	sub.CurrentPeriodEnd = &periodEnd

	// Update cancel status
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd
	if stripeSub.CanceledAt > 0 {
		canceledAt := time.Unix(stripeSub.CanceledAt, 0)
		sub.CanceledAt = &canceledAt
	}

	return s.subscriptionRepo.Update(ctx, sub)
}

// HandleSubscriptionDeleted handles customer.subscription.deleted webhook
func (s *StripeService) HandleSubscriptionDeleted(ctx context.Context, stripeSub *stripe.Subscription) error {
	sub, err := s.subscriptionRepo.GetByStripeSubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	sub.Status = entity.SubscriptionStatusCanceled
	sub.CanceledAt = &now

	// Downgrade to free plan
	sub.PlanTier = entity.PlanTierFree
	sub.StripeSubscriptionID = ""
	sub.StripePriceID = ""

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return err
	}

	// Update usage limits to free tier
	limits := entity.GetPlanLimits(entity.PlanTierFree)
	if err := s.usageRepo.UpdateLimits(ctx, sub.OrganizationID, limits); err != nil {
		log.Printf("[Stripe] Error downgrading usage limits: %v", err)
	}

	log.Printf("[Stripe] Subscription deleted, org %s downgraded to free", sub.OrganizationID)

	return nil
}

// HandleInvoicePaid handles invoice.paid webhook
func (s *StripeService) HandleInvoicePaid(ctx context.Context, inv *stripe.Invoice) error {
	sub, err := s.subscriptionRepo.GetByStripeCustomerID(ctx, inv.Customer.ID)
	if err != nil {
		log.Printf("[Stripe] Invoice paid but no subscription found for customer %s", inv.Customer.ID)
		return nil
	}

	// Record payment
	payment := &entity.PaymentHistory{
		BaseEntity:            entity.NewBaseEntity(),
		OrganizationID:        sub.OrganizationID,
		SubscriptionID:        sub.ID,
		Amount:                decimal.NewFromInt(inv.AmountPaid).Div(decimal.NewFromInt(100)), // Convert from cents
		Currency:              string(inv.Currency),
		Status:                entity.PaymentStatusSucceeded,
		Description:           inv.Description,
		StripeInvoiceID:       inv.ID,
		StripePaymentIntentID: inv.PaymentIntent.ID,
		InvoiceNumber:         inv.Number,
		InvoiceURL:            inv.HostedInvoiceURL,
		InvoicePDF:            inv.InvoicePDF,
	}

	now := time.Now()
	payment.PaidAt = &now

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		log.Printf("[Stripe] Error recording payment: %v", err)
	}

	// Update subscription payment info
	sub.LastPaymentAt = &now
	sub.LastPaymentAmount = payment.Amount
	sub.PaymentFailCount = 0 // Reset failures

	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		log.Printf("[Stripe] Error updating subscription: %v", err)
	}

	log.Printf("[Stripe] Invoice paid for org %s, amount: %s %s",
		sub.OrganizationID, payment.Amount, payment.Currency)

	return nil
}

// HandleInvoicePaymentFailed handles invoice.payment_failed webhook
func (s *StripeService) HandleInvoicePaymentFailed(ctx context.Context, inv *stripe.Invoice) error {
	sub, err := s.subscriptionRepo.GetByStripeCustomerID(ctx, inv.Customer.ID)
	if err != nil {
		return nil
	}

	// Record failed payment
	payment := &entity.PaymentHistory{
		BaseEntity:            entity.NewBaseEntity(),
		OrganizationID:        sub.OrganizationID,
		SubscriptionID:        sub.ID,
		Amount:                decimal.NewFromInt(inv.AmountDue).Div(decimal.NewFromInt(100)),
		Currency:              string(inv.Currency),
		Status:                entity.PaymentStatusFailed,
		Description:           inv.Description,
		StripeInvoiceID:       inv.ID,
		StripePaymentIntentID: inv.PaymentIntent.ID,
	}

	now := time.Now()
	payment.FailedAt = &now

	if inv.PaymentIntent != nil && inv.PaymentIntent.LastPaymentError != nil {
		payment.FailReason = inv.PaymentIntent.LastPaymentError.Msg
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		log.Printf("[Stripe] Error recording failed payment: %v", err)
	}

	// Increment payment failures
	if err := s.subscriptionRepo.IncrementPaymentFails(ctx, sub.ID); err != nil {
		log.Printf("[Stripe] Error incrementing payment fails: %v", err)
	}

	// Update status to past due
	if err := s.subscriptionRepo.UpdateStatus(ctx, sub.ID, entity.SubscriptionStatusPastDue); err != nil {
		log.Printf("[Stripe] Error updating subscription status: %v", err)
	}

	// Auto-downgrade after 3 failed attempts
	sub, _ = s.subscriptionRepo.GetByID(ctx, sub.ID)
	if sub.PaymentFailCount >= 3 {
		log.Printf("[Stripe] Auto-downgrading org %s due to payment failures", sub.OrganizationID)

		// Cancel subscription
		if sub.StripeSubscriptionID != "" {
			subscription.Cancel(sub.StripeSubscriptionID, nil)
		}

		// Downgrade to free
		sub.Status = entity.SubscriptionStatusCanceled
		sub.PlanTier = entity.PlanTierFree
		sub.CanceledAt = &now
		s.subscriptionRepo.Update(ctx, sub)

		limits := entity.GetPlanLimits(entity.PlanTierFree)
		s.usageRepo.UpdateLimits(ctx, sub.OrganizationID, limits)
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *StripeService) getPriceID(tier entity.PlanTier, cycle entity.BillingCycle) (string, error) {
	if s.priceIDs == nil {
		return "", fmt.Errorf("price IDs not configured")
	}

	cyclePrices, ok := s.priceIDs[tier]
	if !ok {
		return "", fmt.Errorf("no prices configured for plan: %s", tier)
	}

	priceID, ok := cyclePrices[cycle]
	if !ok {
		return "", fmt.Errorf("no price configured for %s %s", tier, cycle)
	}

	return priceID, nil
}

func mapStripeStatus(status stripe.SubscriptionStatus) entity.SubscriptionStatus {
	switch status {
	case stripe.SubscriptionStatusActive:
		return entity.SubscriptionStatusActive
	case stripe.SubscriptionStatusPastDue:
		return entity.SubscriptionStatusPastDue
	case stripe.SubscriptionStatusCanceled:
		return entity.SubscriptionStatusCanceled
	case stripe.SubscriptionStatusTrialing:
		return entity.SubscriptionStatusTrialing
	case stripe.SubscriptionStatusPaused:
		return entity.SubscriptionStatusPaused
	case stripe.SubscriptionStatusUnpaid:
		return entity.SubscriptionStatusUnpaid
	case stripe.SubscriptionStatusIncomplete, stripe.SubscriptionStatusIncompleteExpired:
		return entity.SubscriptionStatusIncomplete
	default:
		return entity.SubscriptionStatusActive
	}
}

// GetWebhookSecret returns the webhook secret for signature verification
func (s *StripeService) GetWebhookSecret() string {
	return s.webhookSecret
}
