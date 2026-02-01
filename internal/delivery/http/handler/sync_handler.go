package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	syncpkg "github.com/ads-aggregator/ads-aggregator/internal/usecase/sync"
	"github.com/google/uuid"
)

// ============================================================================
// Sync Handler - API endpoints for sync operations
// ============================================================================

// SyncHandler handles sync-related HTTP endpoints
type SyncHandler struct {
	syncStateRepo   repository.SyncStateRepository
	syncJobRepo     repository.SyncJobRepository
	rateLimitRepo   repository.ManualSyncRateLimitRepository
	connAccountRepo repository.ConnectedAccountRepository
	scheduler       *syncpkg.Scheduler
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(
	syncStateRepo repository.SyncStateRepository,
	syncJobRepo repository.SyncJobRepository,
	rateLimitRepo repository.ManualSyncRateLimitRepository,
	connAccountRepo repository.ConnectedAccountRepository,
	scheduler *syncpkg.Scheduler,
) *SyncHandler {
	return &SyncHandler{
		syncStateRepo:   syncStateRepo,
		syncJobRepo:     syncJobRepo,
		rateLimitRepo:   rateLimitRepo,
		connAccountRepo: connAccountRepo,
		scheduler:       scheduler,
	}
}

// ============================================================================
// Manual Sync Trigger
// ============================================================================

// TriggerManualSyncRequest represents the request body for manual sync
type TriggerManualSyncRequest struct {
	ConnectedAccountID string  `json:"connected_account_id"`
	Scope              string  `json:"scope,omitempty"` // "account", "campaign", "metrics"
	CampaignID         *string `json:"campaign_id,omitempty"`
	DateRangeStart     *string `json:"date_range_start,omitempty"` // Format: "2006-01-02"
	DateRangeEnd       *string `json:"date_range_end,omitempty"`   // Format: "2006-01-02"
}

// TriggerManualSyncResponse represents the response for manual sync
type TriggerManualSyncResponse struct {
	Success           bool       `json:"success"`
	JobID             string     `json:"job_id,omitempty"`
	Message           string     `json:"message,omitempty"`
	RemainingManual   int        `json:"remaining_manual_syncs"`
	NextResetAt       time.Time  `json:"next_reset_at"`
}

// HandleTriggerManualSync handles POST /api/v1/sync/trigger
func (h *SyncHandler) HandleTriggerManualSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	// Parse request
	var req TriggerManualSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate connected account ID
	connAccountID, err := uuid.Parse(req.ConnectedAccountID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid connected_account_id")
		return
	}

	// Check rate limit
	rateLimit, err := h.rateLimitRepo.GetOrCreate(ctx, userID, orgID)
	if err != nil {
		log.Printf("[SyncHandler] Error getting rate limit: %v", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Check if user can trigger manual sync
	if !rateLimit.CanTriggerManualSync() {
		currentHour := time.Now().Truncate(time.Hour)
		nextReset := currentHour.Add(time.Hour)

		respondJSON(w, http.StatusTooManyRequests, TriggerManualSyncResponse{
			Success:         false,
			Message:         "rate limit exceeded: max 5 manual syncs per hour",
			RemainingManual: 0,
			NextResetAt:     nextReset,
		})
		return
	}

	// Parse scope
	scope := entity.SyncScopeMetrics // Default
	if req.Scope != "" {
		switch req.Scope {
		case "account":
			scope = entity.SyncScopeAccount
		case "campaign":
			scope = entity.SyncScopeCampaign
		case "metrics":
			scope = entity.SyncScopeMetrics
		case "structure":
			scope = entity.SyncScopeStructure
		default:
			respondError(w, http.StatusBadRequest, "invalid scope")
			return
		}
	}

	// Parse campaign ID if provided
	var campaignID *uuid.UUID
	if req.CampaignID != nil {
		id, err := uuid.Parse(*req.CampaignID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid campaign_id")
			return
		}
		campaignID = &id
	}

	// Parse date range
	var dateStart, dateEnd *time.Time
	if req.DateRangeStart != nil {
		t, err := time.Parse("2006-01-02", *req.DateRangeStart)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid date_range_start format, use YYYY-MM-DD")
			return
		}
		dateStart = &t
	}
	if req.DateRangeEnd != nil {
		t, err := time.Parse("2006-01-02", *req.DateRangeEnd)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid date_range_end format, use YYYY-MM-DD")
			return
		}
		dateEnd = &t
	}

	// Trigger manual sync
	syncReq := syncpkg.ManualSyncRequest{
		ConnectedAccountID: connAccountID,
		UserID:             userID,
		Scope:              scope,
		CampaignID:         campaignID,
		DateRangeStart:     dateStart,
		DateRangeEnd:       dateEnd,
	}

	job, err := h.scheduler.TriggerManualSync(ctx, syncReq)
	if err != nil {
		log.Printf("[SyncHandler] Error triggering manual sync: %v", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Increment rate limit counter
	if err := h.rateLimitRepo.IncrementCount(ctx, rateLimit.ID); err != nil {
		log.Printf("[SyncHandler] Error incrementing rate limit: %v", err)
	}

	// Calculate remaining syncs
	remaining := rateLimit.RemainingManualSyncs() - 1
	if remaining < 0 {
		remaining = 0
	}

	currentHour := time.Now().Truncate(time.Hour)
	nextReset := currentHour.Add(time.Hour)

	respondJSON(w, http.StatusAccepted, TriggerManualSyncResponse{
		Success:         true,
		JobID:           job.ID.String(),
		Message:         "sync job created successfully",
		RemainingManual: remaining,
		NextResetAt:     nextReset,
	})
}

// ============================================================================
// Data Freshness Endpoint
// ============================================================================

// DataFreshnessResponse represents the response for data freshness
type DataFreshnessResponse struct {
	Accounts      []entity.DataFreshnessInfo `json:"accounts"`
	OverallStatus string                      `json:"overall_status"` // "fresh", "stale", "outdated"
	LastUpdated   time.Time                   `json:"last_updated"`
}

// HandleGetDataFreshness handles GET /api/v1/sync/freshness
func (h *SyncHandler) HandleGetDataFreshness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	// Get freshness info for all accounts
	freshnessInfo, err := h.syncStateRepo.GetDataFreshness(ctx, orgID)
	if err != nil {
		log.Printf("[SyncHandler] Error getting data freshness: %v", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Calculate overall status
	overallStatus := "fresh"
	for _, info := range freshnessInfo {
		switch info.FreshnessStatus {
		case "outdated", "never_synced":
			overallStatus = "outdated"
		case "stale":
			if overallStatus != "outdated" {
				overallStatus = "stale"
			}
		case "recent":
			if overallStatus == "fresh" {
				overallStatus = "recent"
			}
		}
	}

	respondJSON(w, http.StatusOK, DataFreshnessResponse{
		Accounts:      freshnessInfo,
		OverallStatus: overallStatus,
		LastUpdated:   time.Now(),
	})
}

// ============================================================================
// Sync Job Status
// ============================================================================

// SyncJobStatusResponse represents a sync job status
type SyncJobStatusResponse struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	SyncType        string     `json:"sync_type"`
	Platform        string     `json:"platform"`
	ProgressPercent int        `json:"progress_percent"`
	ProgressMessage string     `json:"progress_message,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	RecordsProcessed int       `json:"records_processed"`
	RecordsFailed   int        `json:"records_failed"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// HandleGetSyncJobStatus handles GET /api/v1/sync/jobs/{job_id}
func (h *SyncHandler) HandleGetSyncJobStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get job ID from URL
	jobIDStr := r.PathValue("job_id")
	if jobIDStr == "" {
		respondError(w, http.StatusBadRequest, "job_id required")
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid job_id")
		return
	}

	job, err := h.syncJobRepo.GetByID(ctx, jobID)
	if err != nil {
		respondError(w, http.StatusNotFound, "job not found")
		return
	}

	respondJSON(w, http.StatusOK, SyncJobStatusResponse{
		ID:               job.ID.String(),
		Status:           string(job.Status),
		SyncType:         string(job.SyncType),
		Platform:         string(job.Platform),
		ProgressPercent:  job.ProgressPercent,
		ProgressMessage:  job.ProgressMessage,
		StartedAt:        job.StartedAt,
		CompletedAt:      job.CompletedAt,
		RecordsProcessed: job.RecordsProcessed,
		RecordsFailed:    job.RecordsFailed,
		ErrorMessage:     job.ErrorMessage,
	})
}

// HandleListRecentJobs handles GET /api/v1/sync/jobs
func (h *SyncHandler) HandleListRecentJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		respondError(w, http.StatusBadRequest, "organization_id required")
		return
	}

	jobs, err := h.syncJobRepo.ListRecent(ctx, orgID, 20)
	if err != nil {
		log.Printf("[SyncHandler] Error listing recent jobs: %v", err)
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var response []SyncJobStatusResponse
	for _, job := range jobs {
		response = append(response, SyncJobStatusResponse{
			ID:               job.ID.String(),
			Status:           string(job.Status),
			SyncType:         string(job.SyncType),
			Platform:         string(job.Platform),
			ProgressPercent:  job.ProgressPercent,
			ProgressMessage:  job.ProgressMessage,
			StartedAt:        job.StartedAt,
			CompletedAt:      job.CompletedAt,
			RecordsProcessed: job.RecordsProcessed,
			RecordsFailed:    job.RecordsFailed,
			ErrorMessage:     job.ErrorMessage,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  response,
		"count": len(response),
	})
}

// ============================================================================
// Sync State Endpoints
// ============================================================================

// HandleGetSyncState handles GET /api/v1/sync/state/{account_id}
func (h *SyncHandler) HandleGetSyncState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	accountIDStr := r.PathValue("account_id")
	if accountIDStr == "" {
		respondError(w, http.StatusBadRequest, "account_id required")
		return
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid account_id")
		return
	}

	state, err := h.syncStateRepo.GetByConnectedAccount(ctx, accountID)
	if err != nil {
		respondError(w, http.StatusNotFound, "sync state not found")
		return
	}

	// Calculate freshness
	state.DataFreshnessMinutes = state.CalculateDataFreshness()

	respondJSON(w, http.StatusOK, state)
}

// ============================================================================
// Helper Functions
// ============================================================================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
