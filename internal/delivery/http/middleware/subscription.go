package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys for subscription data
const (
	ContextKeySubscription = "subscription"
	ContextKeyPlanLimits   = "plan_limits"
	ContextKeyTenantCtx    = "tenant_context"
)

// SubscriptionMiddleware handles subscription-based access control
type SubscriptionMiddleware struct {
	subscriptionRepo repository.SubscriptionRepository
	usageRepo        repository.UsageRepository

	// Cache for subscriptions (simple in-memory cache)
	cache      map[uuid.UUID]*cachedSubscription
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
}

type cachedSubscription struct {
	subscription *entity.Subscription
	limits       entity.PlanLimits
	cachedAt     time.Time
}

// NewSubscriptionMiddleware creates a new subscription middleware
func NewSubscriptionMiddleware(
	subscriptionRepo repository.SubscriptionRepository,
	usageRepo repository.UsageRepository,
) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		subscriptionRepo: subscriptionRepo,
		usageRepo:        usageRepo,
		cache:            make(map[uuid.UUID]*cachedSubscription),
		cacheTTL:         5 * time.Minute,
	}
}

// RequireActiveSubscription ensures the organization has an active subscription
func (m *SubscriptionMiddleware) RequireActiveSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get(ContextKeyOrgID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Organization not found in context",
				},
			})
			return
		}

		orgUUID := orgID.(uuid.UUID)

		// Get subscription (with caching)
		sub, limits, err := m.getSubscriptionWithCache(c, orgUUID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "NO_SUBSCRIPTION",
					"message": "No active subscription found",
				},
			})
			return
		}

		// Check if subscription is active
		if !sub.IsActive() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "SUBSCRIPTION_INACTIVE",
					"message": "Subscription is not active",
					"status":  sub.Status,
				},
			})
			return
		}

		// Build tenant context
		userID, _ := GetUserID(c)
		role, _ := GetRole(c)

		tenantCtx := &entity.TenantContext{
			OrganizationID:   orgUUID,
			UserID:           userID,
			UserRole:         entity.UserRole(role),
			SubscriptionTier: sub.PlanTier,
			Limits:           limits,
			IsActive:         sub.IsActive(),
		}

		// Set context values
		c.Set(ContextKeySubscription, sub)
		c.Set(ContextKeyPlanLimits, limits)
		c.Set(ContextKeyTenantCtx, tenantCtx)

		c.Next()
	}
}

// TrackAPIUsage tracks API calls and enforces rate limits
func (m *SubscriptionMiddleware) TrackAPIUsage() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get(ContextKeyOrgID)
		if !exists {
			c.Next()
			return
		}

		orgUUID := orgID.(uuid.UUID)

		// Check API limit before processing
		withinLimit, currentCount, limit, err := m.usageRepo.CheckAPILimit(c.Request.Context(), orgUUID)
		if err != nil {
			log.Printf("[Subscription] Error checking API limit: %v", err)
			// Don't block on error, just log
			c.Next()
			return
		}

		if !withinLimit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "API_LIMIT_EXCEEDED",
					"message": "Daily API call limit exceeded",
					"current": currentCount,
					"limit":   limit,
				},
			})
			return
		}

		// Increment API call count
		if err := m.usageRepo.IncrementAPICallCount(c.Request.Context(), orgUUID); err != nil {
			log.Printf("[Subscription] Error incrementing API count: %v", err)
		}

		c.Next()
	}
}

// RequireFeature checks if the organization has access to a specific feature
func (m *SubscriptionMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get(ContextKeyTenantCtx)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "NO_SUBSCRIPTION",
					"message": "Subscription context not found",
				},
			})
			return
		}

		tenant := tenantCtx.(*entity.TenantContext)

		if !tenant.HasFeature(feature) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FEATURE_NOT_AVAILABLE",
					"message": "This feature is not available on your current plan",
					"feature": feature,
					"plan":    tenant.SubscriptionTier,
				},
			})
			return
		}

		c.Next()
	}
}

// RequireMinPlan checks if the organization has at least the specified plan tier
func (m *SubscriptionMiddleware) RequireMinPlan(minTier entity.PlanTier) gin.HandlerFunc {
	tierOrder := map[entity.PlanTier]int{
		entity.PlanTierFree:     0,
		entity.PlanTierPro:      1,
		entity.PlanTierBusiness: 2,
	}

	return func(c *gin.Context) {
		tenantCtx, exists := c.Get(ContextKeyTenantCtx)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "NO_SUBSCRIPTION",
					"message": "Subscription context not found",
				},
			})
			return
		}

		tenant := tenantCtx.(*entity.TenantContext)

		if tierOrder[tenant.SubscriptionTier] < tierOrder[minTier] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "PLAN_UPGRADE_REQUIRED",
					"message": "Please upgrade your plan to access this feature",
					"current": tenant.SubscriptionTier,
					"required": minTier,
				},
			})
			return
		}

		c.Next()
	}
}

// CheckAccountLimit validates if the organization can add more ad accounts
func (m *SubscriptionMiddleware) CheckAccountLimit(getCurrentCount func(c *gin.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get(ContextKeyTenantCtx)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "NO_SUBSCRIPTION",
					"message": "Subscription context not found",
				},
			})
			return
		}

		tenant := tenantCtx.(*entity.TenantContext)

		currentCount, err := getCurrentCount(c)
		if err != nil {
			log.Printf("[Subscription] Error getting account count: %v", err)
			c.Next()
			return
		}

		if !tenant.CanAddAccount(currentCount) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "ACCOUNT_LIMIT_REACHED",
					"message": "You have reached the maximum number of ad accounts for your plan",
					"current": currentCount,
					"limit":   tenant.Limits.MaxAdAccounts,
					"plan":    tenant.SubscriptionTier,
				},
			})
			return
		}

		c.Next()
	}
}

// CheckUserLimit validates if the organization can add more users
func (m *SubscriptionMiddleware) CheckUserLimit(getCurrentCount func(c *gin.Context) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCtx, exists := c.Get(ContextKeyTenantCtx)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "NO_SUBSCRIPTION",
					"message": "Subscription context not found",
				},
			})
			return
		}

		tenant := tenantCtx.(*entity.TenantContext)

		currentCount, err := getCurrentCount(c)
		if err != nil {
			log.Printf("[Subscription] Error getting user count: %v", err)
			c.Next()
			return
		}

		if !tenant.CanAddUser(currentCount) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "USER_LIMIT_REACHED",
					"message": "You have reached the maximum number of users for your plan",
					"current": currentCount,
					"limit":   tenant.Limits.MaxUsersPerOrg,
					"plan":    tenant.SubscriptionTier,
				},
			})
			return
		}

		c.Next()
	}
}

// CheckStorageLimit validates storage usage
func (m *SubscriptionMiddleware) CheckStorageLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID, exists := c.Get(ContextKeyOrgID)
		if !exists {
			c.Next()
			return
		}

		orgUUID := orgID.(uuid.UUID)

		usage, err := m.usageRepo.GetOrCreateDaily(c.Request.Context(), orgUUID, time.Now())
		if err != nil {
			log.Printf("[Subscription] Error getting usage: %v", err)
			c.Next()
			return
		}

		if usage.IsStorageLimitExceeded() {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "STORAGE_LIMIT_EXCEEDED",
					"message": "Storage limit exceeded for your plan",
					"used":    usage.StorageUsedBytes,
					"limit":   usage.StorageLimitBytes,
				},
			})
			return
		}

		c.Next()
	}
}

// getSubscriptionWithCache retrieves subscription with caching
func (m *SubscriptionMiddleware) getSubscriptionWithCache(c *gin.Context, orgID uuid.UUID) (*entity.Subscription, entity.PlanLimits, error) {
	// Check cache first
	m.cacheMutex.RLock()
	cached, exists := m.cache[orgID]
	m.cacheMutex.RUnlock()

	if exists && time.Since(cached.cachedAt) < m.cacheTTL {
		return cached.subscription, cached.limits, nil
	}

	// Fetch from database
	sub, err := m.subscriptionRepo.GetByOrganization(c.Request.Context(), orgID)
	if err != nil {
		return nil, entity.PlanLimits{}, err
	}

	limits := sub.GetLimits()

	// Update cache
	m.cacheMutex.Lock()
	m.cache[orgID] = &cachedSubscription{
		subscription: sub,
		limits:       limits,
		cachedAt:     time.Now(),
	}
	m.cacheMutex.Unlock()

	return sub, limits, nil
}

// InvalidateCache removes a subscription from the cache
func (m *SubscriptionMiddleware) InvalidateCache(orgID uuid.UUID) {
	m.cacheMutex.Lock()
	delete(m.cache, orgID)
	m.cacheMutex.Unlock()
}

// ClearCache clears all cached subscriptions
func (m *SubscriptionMiddleware) ClearCache() {
	m.cacheMutex.Lock()
	m.cache = make(map[uuid.UUID]*cachedSubscription)
	m.cacheMutex.Unlock()
}

// Helper functions to extract subscription context

// GetSubscription extracts the subscription from the gin context
func GetSubscription(c *gin.Context) (*entity.Subscription, bool) {
	sub, exists := c.Get(ContextKeySubscription)
	if !exists {
		return nil, false
	}
	return sub.(*entity.Subscription), true
}

// GetPlanLimits extracts the plan limits from the gin context
func GetPlanLimits(c *gin.Context) (entity.PlanLimits, bool) {
	limits, exists := c.Get(ContextKeyPlanLimits)
	if !exists {
		return entity.PlanLimits{}, false
	}
	return limits.(entity.PlanLimits), true
}

// GetTenantContext extracts the tenant context from the gin context
func GetTenantContext(c *gin.Context) (*entity.TenantContext, bool) {
	ctx, exists := c.Get(ContextKeyTenantCtx)
	if !exists {
		return nil, false
	}
	return ctx.(*entity.TenantContext), true
}
