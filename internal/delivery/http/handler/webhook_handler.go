package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
)

// ============================================================================
// Webhook Handler
// ============================================================================

// WebhookHandler handles incoming webhooks from ad platforms
type WebhookHandler struct {
	webhookRepo     repository.WebhookEventRepository
	syncStateRepo   repository.SyncStateRepository
	syncJobRepo     repository.SyncJobRepository
	connAccountRepo repository.ConnectedAccountRepository

	// Platform secrets for signature verification
	metaAppSecret   string
	tiktokAppSecret string
	shopeePartnerKey string

	// Webhook processor
	processor WebhookProcessor
}

// WebhookProcessor processes webhook events
type WebhookProcessor interface {
	ProcessEvent(ctx context.Context, event *entity.WebhookEvent) error
}

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	MetaAppSecret    string
	TikTokAppSecret  string
	ShopeePartnerKey string
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	webhookRepo repository.WebhookEventRepository,
	syncStateRepo repository.SyncStateRepository,
	syncJobRepo repository.SyncJobRepository,
	connAccountRepo repository.ConnectedAccountRepository,
	config *WebhookConfig,
) *WebhookHandler {
	return &WebhookHandler{
		webhookRepo:      webhookRepo,
		syncStateRepo:    syncStateRepo,
		syncJobRepo:      syncJobRepo,
		connAccountRepo:  connAccountRepo,
		metaAppSecret:    config.MetaAppSecret,
		tiktokAppSecret:  config.TikTokAppSecret,
		shopeePartnerKey: config.ShopeePartnerKey,
	}
}

// SetProcessor sets the webhook processor
func (h *WebhookHandler) SetProcessor(processor WebhookProcessor) {
	h.processor = processor
}

// ============================================================================
// Meta (Facebook) Webhook Handler
// ============================================================================

// HandleMetaWebhook handles incoming Meta webhooks
func (h *WebhookHandler) HandleMetaWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Handle verification challenge (GET request)
	if r.Method == http.MethodGet {
		h.handleMetaVerification(w, r)
		return
	}

	// Handle webhook event (POST request)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Webhook] Error reading body: %v", err)
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get signature from header
	signature := r.Header.Get("X-Hub-Signature-256")

	// Create webhook event record
	event := &entity.WebhookEvent{
		ID:           uuid.New(),
		Platform:     entity.PlatformMeta,
		ReceivedAt:   time.Now(),
		Signature:    signature,
		RawPayload:   string(body),
	}

	// Verify signature
	if h.metaAppSecret != "" {
		event.SignatureValid = h.verifyMetaSignature(body, signature)
		if !event.SignatureValid {
			log.Printf("[Webhook] Invalid Meta signature")
			event.ProcessingStatus = "failed"
			event.ProcessingError = "invalid signature"
			h.webhookRepo.Create(ctx, event)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse payload
	var payload MetaWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[Webhook] Error parsing Meta payload: %v", err)
		event.ProcessingStatus = "failed"
		event.ProcessingError = "invalid JSON payload"
		h.webhookRepo.Create(ctx, event)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	event.EventType = payload.Object
	event.ParsedPayload = entity.JSONMap{
		"object": payload.Object,
		"entry":  payload.Entry,
	}

	// Save event
	if err := h.webhookRepo.Create(ctx, event); err != nil {
		log.Printf("[Webhook] Error saving webhook event: %v", err)
	}

	// Process entries
	for _, entry := range payload.Entry {
		h.processMetaEntry(ctx, event, entry)
	}

	// Respond with 200 OK immediately
	w.WriteHeader(http.StatusOK)
}

// handleMetaVerification handles Meta webhook verification challenge
func (h *WebhookHandler) handleMetaVerification(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	// Verify token matches your configured verify token
	verifyToken := h.metaAppSecret[:16] // Use first 16 chars of app secret as verify token

	if mode == "subscribe" && token == verifyToken {
		log.Printf("[Webhook] Meta verification successful")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Printf("[Webhook] Meta verification failed: mode=%s", mode)
	http.Error(w, "Verification failed", http.StatusForbidden)
}

// verifyMetaSignature verifies the Meta webhook signature
func (h *WebhookHandler) verifyMetaSignature(payload []byte, signature string) bool {
	if signature == "" || h.metaAppSecret == "" {
		return false
	}

	// Signature format: sha256=<hash>
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha256" {
		return false
	}

	expectedSig := parts[1]

	// Calculate HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(h.metaAppSecret))
	mac.Write(payload)
	actualSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(actualSig))
}

// processMetaEntry processes a single Meta webhook entry
func (h *WebhookHandler) processMetaEntry(ctx context.Context, event *entity.WebhookEvent, entry MetaWebhookEntry) {
	log.Printf("[Webhook] Processing Meta entry for ad account: %s", entry.ID)

	// Find connected account by platform ID
	// Note: entry.ID is the ad account ID in format "act_123456"
	adAccountID := strings.TrimPrefix(entry.ID, "act_")

	for _, change := range entry.Changes {
		h.processMetaChange(ctx, event, adAccountID, change)
	}
}

// processMetaChange processes a single Meta webhook change
func (h *WebhookHandler) processMetaChange(ctx context.Context, event *entity.WebhookEvent, adAccountID string, change MetaWebhookChange) {
	log.Printf("[Webhook] Meta change: field=%s", change.Field)

	switch change.Field {
	case "campaigns":
		h.triggerCampaignSync(ctx, entity.PlatformMeta, adAccountID, change.Value)
	case "ads":
		h.triggerAdSync(ctx, entity.PlatformMeta, adAccountID, change.Value)
	case "adsets":
		h.triggerAdSetSync(ctx, entity.PlatformMeta, adAccountID, change.Value)
	case "insights":
		h.triggerMetricsSync(ctx, entity.PlatformMeta, adAccountID, change.Value)
	default:
		log.Printf("[Webhook] Unhandled Meta change field: %s", change.Field)
	}
}

// ============================================================================
// TikTok Webhook Handler
// ============================================================================

// HandleTikTokWebhook handles incoming TikTok webhooks
func (h *WebhookHandler) HandleTikTokWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// TikTok signature verification
	signature := r.Header.Get("X-TT-Signature")
	timestamp := r.Header.Get("X-TT-Timestamp")

	event := &entity.WebhookEvent{
		ID:           uuid.New(),
		Platform:     entity.PlatformTikTok,
		ReceivedAt:   time.Now(),
		Signature:    signature,
		RawPayload:   string(body),
	}

	// Verify signature
	if h.tiktokAppSecret != "" {
		event.SignatureValid = h.verifyTikTokSignature(body, signature, timestamp)
		if !event.SignatureValid {
			log.Printf("[Webhook] Invalid TikTok signature")
			event.ProcessingStatus = "failed"
			event.ProcessingError = "invalid signature"
			h.webhookRepo.Create(ctx, event)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse payload
	var payload TikTokWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[Webhook] Error parsing TikTok payload: %v", err)
		event.ProcessingStatus = "failed"
		event.ProcessingError = "invalid JSON payload"
		h.webhookRepo.Create(ctx, event)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	event.EventType = payload.EventType
	event.ParsedPayload = entity.JSONMap{
		"event_type":    payload.EventType,
		"advertiser_id": payload.AdvertiserID,
		"data":          payload.Data,
	}

	if err := h.webhookRepo.Create(ctx, event); err != nil {
		log.Printf("[Webhook] Error saving webhook event: %v", err)
	}

	// Process based on event type
	h.processTikTokEvent(ctx, event, &payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

// verifyTikTokSignature verifies TikTok webhook signature
func (h *WebhookHandler) verifyTikTokSignature(payload []byte, signature, timestamp string) bool {
	if signature == "" || h.tiktokAppSecret == "" {
		return false
	}

	// TikTok signature: HMAC-SHA256(timestamp + body)
	message := timestamp + string(payload)
	mac := hmac.New(sha256.New, []byte(h.tiktokAppSecret))
	mac.Write([]byte(message))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// processTikTokEvent processes a TikTok webhook event
func (h *WebhookHandler) processTikTokEvent(ctx context.Context, event *entity.WebhookEvent, payload *TikTokWebhookPayload) {
	log.Printf("[Webhook] Processing TikTok event: %s for advertiser %s", payload.EventType, payload.AdvertiserID)

	switch payload.EventType {
	case "campaign_update", "campaign_create", "campaign_delete":
		h.triggerCampaignSync(ctx, entity.PlatformTikTok, payload.AdvertiserID, payload.Data)
	case "adgroup_update", "adgroup_create", "adgroup_delete":
		h.triggerAdSetSync(ctx, entity.PlatformTikTok, payload.AdvertiserID, payload.Data)
	case "ad_update", "ad_create", "ad_delete":
		h.triggerAdSync(ctx, entity.PlatformTikTok, payload.AdvertiserID, payload.Data)
	case "report_ready":
		h.triggerMetricsSync(ctx, entity.PlatformTikTok, payload.AdvertiserID, payload.Data)
	default:
		log.Printf("[Webhook] Unhandled TikTok event type: %s", payload.EventType)
	}
}

// ============================================================================
// Shopee Webhook Handler
// ============================================================================

// HandleShopeeWebhook handles incoming Shopee webhooks
func (h *WebhookHandler) HandleShopeeWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Shopee uses different signature method
	signature := r.Header.Get("Authorization")

	event := &entity.WebhookEvent{
		ID:           uuid.New(),
		Platform:     entity.PlatformShopee,
		ReceivedAt:   time.Now(),
		Signature:    signature,
		RawPayload:   string(body),
	}

	// Parse payload first to get required fields for signature verification
	var payload ShopeeWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[Webhook] Error parsing Shopee payload: %v", err)
		event.ProcessingStatus = "failed"
		event.ProcessingError = "invalid JSON payload"
		h.webhookRepo.Create(ctx, event)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Verify signature
	if h.shopeePartnerKey != "" {
		event.SignatureValid = h.verifyShopeeSignature(payload, signature)
		if !event.SignatureValid {
			log.Printf("[Webhook] Invalid Shopee signature")
			event.ProcessingStatus = "failed"
			event.ProcessingError = "invalid signature"
			h.webhookRepo.Create(ctx, event)
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	event.EventType = payload.PushType
	event.ParsedPayload = entity.JSONMap{
		"push_type": payload.PushType,
		"shop_id":   payload.ShopID,
		"data":      payload.Data,
	}

	if err := h.webhookRepo.Create(ctx, event); err != nil {
		log.Printf("[Webhook] Error saving webhook event: %v", err)
	}

	// Process based on push type
	h.processShopeeEvent(ctx, event, &payload)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"code": 0})
}

// verifyShopeeSignature verifies Shopee webhook signature
func (h *WebhookHandler) verifyShopeeSignature(payload ShopeeWebhookPayload, signature string) bool {
	if signature == "" || h.shopeePartnerKey == "" {
		return false
	}

	// Shopee signature: base_string = partner_id + push_type + timestamp
	baseString := fmt.Sprintf("%d%s%d", payload.PartnerID, payload.PushType, payload.Timestamp)
	mac := hmac.New(sha256.New, []byte(h.shopeePartnerKey))
	mac.Write([]byte(baseString))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// processShopeeEvent processes a Shopee webhook event
func (h *WebhookHandler) processShopeeEvent(ctx context.Context, event *entity.WebhookEvent, payload *ShopeeWebhookPayload) {
	log.Printf("[Webhook] Processing Shopee event: %s for shop %d", payload.PushType, payload.ShopID)

	shopIDStr := fmt.Sprintf("%d", payload.ShopID)

	switch payload.PushType {
	case "campaign_update", "campaign_status_change":
		h.triggerCampaignSync(ctx, entity.PlatformShopee, shopIDStr, payload.Data)
	case "order_status_update":
		h.triggerMetricsSync(ctx, entity.PlatformShopee, shopIDStr, payload.Data)
	case "ads_performance_update":
		h.triggerMetricsSync(ctx, entity.PlatformShopee, shopIDStr, payload.Data)
	default:
		log.Printf("[Webhook] Unhandled Shopee push type: %s", payload.PushType)
	}
}

// ============================================================================
// Sync Trigger Methods
// ============================================================================

func (h *WebhookHandler) triggerCampaignSync(ctx context.Context, platform entity.Platform, accountID string, data interface{}) {
	h.createSyncJob(ctx, platform, accountID, entity.SyncScopeStructure, data)
}

func (h *WebhookHandler) triggerAdSetSync(ctx context.Context, platform entity.Platform, accountID string, data interface{}) {
	h.createSyncJob(ctx, platform, accountID, entity.SyncScopeStructure, data)
}

func (h *WebhookHandler) triggerAdSync(ctx context.Context, platform entity.Platform, accountID string, data interface{}) {
	h.createSyncJob(ctx, platform, accountID, entity.SyncScopeStructure, data)
}

func (h *WebhookHandler) triggerMetricsSync(ctx context.Context, platform entity.Platform, accountID string, data interface{}) {
	h.createSyncJob(ctx, platform, accountID, entity.SyncScopeMetrics, data)
}

func (h *WebhookHandler) createSyncJob(ctx context.Context, platform entity.Platform, platformAccountID string, scope entity.SyncScope, data interface{}) {
	// Find connected accounts with this platform account ID
	accounts, err := h.connAccountRepo.ListByPlatform(ctx, uuid.Nil, platform) // This needs org filtering in real impl
	if err != nil {
		log.Printf("[Webhook] Error finding accounts: %v", err)
		return
	}

	for _, account := range accounts {
		if account.PlatformAccountID != platformAccountID {
			continue
		}

		// Check sync state
		state, err := h.syncStateRepo.GetByConnectedAccount(ctx, account.ID)
		if err != nil {
			log.Printf("[Webhook] Error getting sync state: %v", err)
			continue
		}

		if !state.CanSync() {
			log.Printf("[Webhook] Cannot sync account %s - busy or rate limited", account.ID)
			continue
		}

		// Create webhook-triggered sync job
		now := time.Now()
		start := now.AddDate(0, 0, -1) // Last day for webhook updates

		job := &entity.SyncJob{
			BaseEntity:         entity.NewBaseEntity(),
			OrganizationID:     account.OrganizationID,
			ConnectedAccountID: account.ID,
			Platform:           platform,
			SyncType:           entity.SyncTypeWebhook,
			SyncScope:          scope,
			Status:             entity.SyncStatusPending,
			Priority:           5, // Medium-high priority
			ScheduledAt:        now,
			DateRangeStart:     &start,
			DateRangeEnd:       &now,
			MaxRetries:         2,
			TriggeredBy:        "webhook",
			Metadata: entity.JSONMap{
				"webhook_data": data,
			},
		}

		if err := h.syncJobRepo.Create(ctx, job); err != nil {
			log.Printf("[Webhook] Error creating sync job: %v", err)
			continue
		}

		log.Printf("[Webhook] Created sync job %s for account %s", job.ID, account.ID)
	}
}

// ============================================================================
// Webhook Payload Types
// ============================================================================

// MetaWebhookPayload represents Meta webhook payload
type MetaWebhookPayload struct {
	Object string              `json:"object"`
	Entry  []MetaWebhookEntry  `json:"entry"`
}

// MetaWebhookEntry represents a single entry in Meta webhook
type MetaWebhookEntry struct {
	ID      string              `json:"id"`
	Time    int64               `json:"time"`
	Changes []MetaWebhookChange `json:"changes"`
}

// MetaWebhookChange represents a change in Meta webhook
type MetaWebhookChange struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

// TikTokWebhookPayload represents TikTok webhook payload
type TikTokWebhookPayload struct {
	EventType    string      `json:"event_type"`
	AdvertiserID string      `json:"advertiser_id"`
	Timestamp    int64       `json:"timestamp"`
	Data         interface{} `json:"data"`
}

// ShopeeWebhookPayload represents Shopee webhook payload
type ShopeeWebhookPayload struct {
	PushType  string      `json:"push_type"`
	PartnerID int64       `json:"partner_id"`
	ShopID    int64       `json:"shop_id"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}
