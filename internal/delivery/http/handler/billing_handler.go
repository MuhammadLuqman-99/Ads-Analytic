package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/usecase/billing"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// ============================================================================
// Billing Handler
// ============================================================================

// BillingHandler handles billing-related HTTP endpoints
type BillingHandler struct {
	stripeService    *billing.StripeService
	subscriptionRepo repository.SubscriptionRepository
	paymentRepo      repository.PaymentHistoryRepository
	usageRepo        repository.UsageRepository
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(
	stripeService *billing.StripeService,
	subscriptionRepo repository.SubscriptionRepository,
	paymentRepo repository.PaymentHistoryRepository,
	usageRepo repository.UsageRepository,
) *BillingHandler {
	return &BillingHandler{
		stripeService:    stripeService,
		subscriptionRepo: subscriptionRepo,
		paymentRepo:      paymentRepo,
		usageRepo:        usageRepo,
	}
}

// ============================================================================
// Subscription Endpoints
// ============================================================================

// HandleGetPlans returns available subscription plans
func (h *BillingHandler) HandleGetPlans(w http.ResponseWriter, r *http.Request) {
	plans := entity.GetAllPlans()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"plans": plans,
	})
}

// HandleGetSubscription returns the current subscription
func (h *BillingHandler) HandleGetSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	sub, err := h.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "subscription not found")
		return
	}

	// Get limits
	limits := sub.GetLimits()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscription": sub,
		"limits":       limits,
		"is_active":    sub.IsActive(),
		"is_paid":      sub.IsPaid(),
		"days_until_expiry": sub.DaysUntilExpiry(),
	})
}

// CreateCheckoutRequest represents checkout session request
type CreateCheckoutRequest struct {
	PlanTier     string `json:"plan_tier"`
	BillingCycle string `json:"billing_cycle"`
	CouponCode   string `json:"coupon_code,omitempty"`
}

// HandleCreateCheckout creates a Stripe checkout session
func (h *BillingHandler) HandleCreateCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	userEmail, _ := ctx.Value("user_email").(string)

	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate plan tier
	planTier := entity.PlanTier(req.PlanTier)
	if planTier != entity.PlanTierPro && planTier != entity.PlanTierBusiness {
		respondError(w, http.StatusBadRequest, "invalid plan tier")
		return
	}

	// Validate billing cycle
	billingCycle := entity.BillingCycle(req.BillingCycle)
	if billingCycle != entity.BillingCycleMonthly && billingCycle != entity.BillingCycleYearly {
		respondError(w, http.StatusBadRequest, "invalid billing cycle")
		return
	}

	checkoutReq := billing.CreateCheckoutSessionRequest{
		OrganizationID: orgID,
		PlanTier:       planTier,
		BillingCycle:   billingCycle,
		CouponCode:     req.CouponCode,
		CustomerEmail:  userEmail,
	}

	session, err := h.stripeService.CreateCheckoutSession(ctx, checkoutReq)
	if err != nil {
		log.Printf("[Billing] Error creating checkout: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"checkout_url": session.URL,
		"session_id":   session.ID,
	})
}

// HandleCancelSubscription cancels the subscription
func (h *BillingHandler) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	// Parse request for immediate cancellation flag
	var req struct {
		Immediate bool `json:"immediate"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.stripeService.CancelSubscription(ctx, orgID, req.Immediate); err != nil {
		log.Printf("[Billing] Error canceling subscription: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	message := "Subscription will be canceled at end of billing period"
	if req.Immediate {
		message = "Subscription has been canceled"
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": message,
	})
}

// HandleReactivateSubscription reactivates a canceled subscription
func (h *BillingHandler) HandleReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	if err := h.stripeService.ReactivateSubscription(ctx, orgID); err != nil {
		log.Printf("[Billing] Error reactivating subscription: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription reactivated",
	})
}

// ChangePlanRequest represents a plan change request
type ChangePlanRequest struct {
	PlanTier     string `json:"plan_tier"`
	BillingCycle string `json:"billing_cycle"`
}

// HandleChangePlan changes the subscription plan
func (h *BillingHandler) HandleChangePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	var req ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	planTier := entity.PlanTier(req.PlanTier)
	billingCycle := entity.BillingCycle(req.BillingCycle)

	if err := h.stripeService.ChangePlan(ctx, orgID, planTier, billingCycle); err != nil {
		log.Printf("[Billing] Error changing plan: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Plan changed successfully",
	})
}

// ============================================================================
// Payment History
// ============================================================================

// HandleGetPaymentHistory returns payment history
func (h *BillingHandler) HandleGetPaymentHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	payments, err := h.paymentRepo.ListByOrganization(ctx, orgID, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get payment history")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payments": payments,
	})
}

// HandleGetInvoices returns Stripe invoices
func (h *BillingHandler) HandleGetInvoices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	invoices, err := h.stripeService.GetInvoices(ctx, orgID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get invoices")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"invoices": invoices,
	})
}

// HandleGetUpcomingInvoice returns the upcoming invoice
func (h *BillingHandler) HandleGetUpcomingInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	invoice, err := h.stripeService.GetUpcomingInvoice(ctx, orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no upcoming invoice")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"upcoming_invoice": invoice,
	})
}

// ============================================================================
// Usage
// ============================================================================

// HandleGetUsage returns usage summary
func (h *BillingHandler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	// Get current period usage
	sub, err := h.subscriptionRepo.GetByOrganization(ctx, orgID)
	if err != nil {
		respondError(w, http.StatusNotFound, "subscription not found")
		return
	}

	startDate := sub.CurrentPeriodStart
	endDate := sub.CurrentPeriodEnd

	if startDate == nil || endDate == nil {
		respondError(w, http.StatusBadRequest, "no billing period defined")
		return
	}

	summary, err := h.usageRepo.GetSummary(ctx, orgID, *startDate, *endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get usage summary")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"usage":   summary,
		"limits":  sub.GetLimits(),
		"plan":    sub.PlanTier,
	})
}

// ============================================================================
// Stripe Webhook Handler
// ============================================================================

// HandleStripeWebhook handles Stripe webhook events
func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Stripe Webhook] Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature
	signature := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, signature, h.stripeService.GetWebhookSecret())
	if err != nil {
		log.Printf("[Stripe Webhook] Signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	log.Printf("[Stripe Webhook] Received event: %s", event.Type)

	// Handle event
	var handleErr error
	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			handleErr = err
		} else {
			handleErr = h.stripeService.HandleCheckoutCompleted(ctx, &sess)
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			handleErr = err
		} else {
			handleErr = h.stripeService.HandleSubscriptionUpdated(ctx, &sub)
		}

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			handleErr = err
		} else {
			handleErr = h.stripeService.HandleSubscriptionDeleted(ctx, &sub)
		}

	case "invoice.paid":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			handleErr = err
		} else {
			handleErr = h.stripeService.HandleInvoicePaid(ctx, &inv)
		}

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			handleErr = err
		} else {
			handleErr = h.stripeService.HandleInvoicePaymentFailed(ctx, &inv)
		}

	default:
		log.Printf("[Stripe Webhook] Unhandled event type: %s", event.Type)
	}

	if handleErr != nil {
		log.Printf("[Stripe Webhook] Error handling %s: %v", event.Type, handleErr)
		http.Error(w, "Webhook handler error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}
