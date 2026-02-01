package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// ============================================================================
// Job Runner - Executes sync jobs with retry and rate limiting
// ============================================================================

// JobRunner executes sync jobs with proper error handling and retry logic
type JobRunner struct {
	// Repositories
	syncStateRepo   repository.SyncStateRepository
	syncJobRepo     repository.SyncJobRepository
	retryQueueRepo  repository.RetryQueueRepository
	errorLogRepo    repository.SyncErrorLogRepository
	connAccountRepo repository.ConnectedAccountRepository
	campaignRepo    repository.CampaignRepository
	metricsRepo     repository.MetricsRepository

	// Platform connectors
	connectors map[entity.Platform]service.PlatformConnector

	// Rate limiters per platform
	rateLimiters map[entity.Platform]*rate.Limiter
	rateLimiterMu sync.RWMutex

	// Configuration
	configs map[entity.Platform]entity.SyncScheduleConfig

	// Concurrency control
	maxConcurrentJobs int
	runningJobs       sync.WaitGroup
	jobSemaphore      chan struct{}

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// JobRunnerConfig holds configuration for the job runner
type JobRunnerConfig struct {
	MaxConcurrentJobs int
	Configs           map[entity.Platform]entity.SyncScheduleConfig
}

// DefaultJobRunnerConfig returns default job runner configuration
func DefaultJobRunnerConfig() *JobRunnerConfig {
	return &JobRunnerConfig{
		MaxConcurrentJobs: 5,
		Configs:           entity.DefaultSyncScheduleConfigs(),
	}
}

// NewJobRunner creates a new job runner
func NewJobRunner(
	syncStateRepo repository.SyncStateRepository,
	syncJobRepo repository.SyncJobRepository,
	retryQueueRepo repository.RetryQueueRepository,
	errorLogRepo repository.SyncErrorLogRepository,
	connAccountRepo repository.ConnectedAccountRepository,
	campaignRepo repository.CampaignRepository,
	metricsRepo repository.MetricsRepository,
	config *JobRunnerConfig,
) *JobRunner {
	if config == nil {
		config = DefaultJobRunnerConfig()
	}

	runner := &JobRunner{
		syncStateRepo:     syncStateRepo,
		syncJobRepo:       syncJobRepo,
		retryQueueRepo:    retryQueueRepo,
		errorLogRepo:      errorLogRepo,
		connAccountRepo:   connAccountRepo,
		campaignRepo:      campaignRepo,
		metricsRepo:       metricsRepo,
		connectors:        make(map[entity.Platform]service.PlatformConnector),
		rateLimiters:      make(map[entity.Platform]*rate.Limiter),
		configs:           config.Configs,
		maxConcurrentJobs: config.MaxConcurrentJobs,
		jobSemaphore:      make(chan struct{}, config.MaxConcurrentJobs),
	}

	// Initialize rate limiters
	for platform, cfg := range config.Configs {
		runner.rateLimiters[platform] = rate.NewLimiter(
			rate.Limit(cfg.MaxRequestsPerMinute)/60, // Convert to per-second
			cfg.RequestBurstSize,
		)
	}

	return runner
}

// RegisterConnector registers a platform connector
func (r *JobRunner) RegisterConnector(platform entity.Platform, connector service.PlatformConnector) {
	r.connectors[platform] = connector
}

// Start starts the job runner
func (r *JobRunner) Start(ctx context.Context) {
	r.ctx, r.cancel = context.WithCancel(ctx)
	log.Printf("[JobRunner] Started with max %d concurrent jobs", r.maxConcurrentJobs)
}

// Stop stops the job runner and waits for running jobs to complete
func (r *JobRunner) Stop() {
	log.Printf("[JobRunner] Stopping...")
	r.cancel()
	r.runningJobs.Wait()
	log.Printf("[JobRunner] Stopped")
}

// ============================================================================
// Job Execution
// ============================================================================

// ExecuteJob executes a single sync job (implements JobExecutor interface)
func (r *JobRunner) ExecuteJob(ctx context.Context, job *entity.SyncJob) error {
	// Acquire semaphore
	select {
	case r.jobSemaphore <- struct{}{}:
		defer func() { <-r.jobSemaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	r.runningJobs.Add(1)
	defer r.runningJobs.Done()

	log.Printf("[JobRunner] Executing job %s (type: %s, platform: %s)", job.ID, job.SyncType, job.Platform)

	// Update job status to running
	now := time.Now()
	job.StartedAt = &now
	job.Status = entity.SyncStatusRunning
	if err := r.syncJobRepo.Update(ctx, job); err != nil {
		log.Printf("[JobRunner] Error updating job status: %v", err)
	}

	// Update sync state to syncing
	if err := r.syncStateRepo.UpdateSyncStarted(ctx, job.ConnectedAccountID, job.ID); err != nil {
		log.Printf("[JobRunner] Error updating sync state: %v", err)
	}

	// Execute based on sync scope
	var err error
	switch job.SyncScope {
	case entity.SyncScopeAccount:
		err = r.syncFullAccount(ctx, job)
	case entity.SyncScopeCampaign:
		err = r.syncCampaign(ctx, job)
	case entity.SyncScopeMetrics:
		err = r.syncMetrics(ctx, job)
	case entity.SyncScopeStructure:
		err = r.syncStructure(ctx, job)
	default:
		err = fmt.Errorf("unknown sync scope: %s", job.SyncScope)
	}

	// Handle result
	if err != nil {
		r.handleJobError(ctx, job, err)
		return err
	}

	// Mark job as completed
	if err := r.syncJobRepo.MarkCompleted(ctx, job.ID, job.RecordsProcessed, job.RecordsFailed); err != nil {
		log.Printf("[JobRunner] Error marking job completed: %v", err)
	}

	// Update sync state
	if err := r.syncStateRepo.UpdateSyncCompleted(ctx, job.ConnectedAccountID, job.SyncType); err != nil {
		log.Printf("[JobRunner] Error updating sync state: %v", err)
	}

	log.Printf("[JobRunner] Job %s completed: %d records processed, %d failed",
		job.ID, job.RecordsProcessed, job.RecordsFailed)

	return nil
}

// ============================================================================
// Sync Operations
// ============================================================================

func (r *JobRunner) syncFullAccount(ctx context.Context, job *entity.SyncJob) error {
	log.Printf("[JobRunner] Full account sync for %s", job.ConnectedAccountID)

	// Sync structure first
	if err := r.syncStructure(ctx, job); err != nil {
		return fmt.Errorf("structure sync failed: %w", err)
	}

	// Then sync metrics
	if err := r.syncMetrics(ctx, job); err != nil {
		return fmt.Errorf("metrics sync failed: %w", err)
	}

	return nil
}

func (r *JobRunner) syncStructure(ctx context.Context, job *entity.SyncJob) error {
	connector, ok := r.connectors[job.Platform]
	if !ok {
		return fmt.Errorf("no connector for platform: %s", job.Platform)
	}

	// Get connected account for access token
	account, err := r.connAccountRepo.GetByID(ctx, job.ConnectedAccountID)
	if err != nil {
		return fmt.Errorf("failed to get connected account: %w", err)
	}

	// Check token expiry
	if account.IsTokenExpired() {
		return fmt.Errorf("access token expired")
	}

	// Rate limit
	if err := r.waitForRateLimit(ctx, job.Platform); err != nil {
		return err
	}

	// Update progress
	r.updateProgress(ctx, job.ID, 10, "Fetching campaigns...")

	// Get campaigns
	campaigns, err := connector.GetCampaigns(ctx, account.AccessToken, account.PlatformAccountID)
	if err != nil {
		return fmt.Errorf("failed to get campaigns: %w", err)
	}

	r.updateProgress(ctx, job.ID, 50, fmt.Sprintf("Processing %d campaigns...", len(campaigns)))

	// Upsert campaigns
	for _, campaign := range campaigns {
		campaign.OrganizationID = account.OrganizationID
		if err := r.campaignRepo.Upsert(ctx, &campaign); err != nil {
			log.Printf("[JobRunner] Error upserting campaign: %v", err)
			job.RecordsFailed++
		} else {
			job.RecordsProcessed++
		}
	}

	r.updateProgress(ctx, job.ID, 100, "Structure sync complete")

	return nil
}

func (r *JobRunner) syncMetrics(ctx context.Context, job *entity.SyncJob) error {
	connector, ok := r.connectors[job.Platform]
	if !ok {
		return fmt.Errorf("no connector for platform: %s", job.Platform)
	}

	account, err := r.connAccountRepo.GetByID(ctx, job.ConnectedAccountID)
	if err != nil {
		return fmt.Errorf("failed to get connected account: %w", err)
	}

	if account.IsTokenExpired() {
		return fmt.Errorf("access token expired")
	}

	// Get date range
	dateRange := entity.DateRange{}
	if job.DateRangeStart != nil {
		dateRange.StartDate = *job.DateRangeStart
	} else {
		dateRange.StartDate = time.Now().AddDate(0, 0, -7)
	}
	if job.DateRangeEnd != nil {
		dateRange.EndDate = *job.DateRangeEnd
	} else {
		dateRange.EndDate = time.Now()
	}

	r.updateProgress(ctx, job.ID, 10, "Fetching campaigns for metrics...")

	// Get campaigns
	campaigns, err := r.campaignRepo.ListByAdAccount(ctx, job.ConnectedAccountID)
	if err != nil {
		return fmt.Errorf("failed to list campaigns: %w", err)
	}

	totalCampaigns := len(campaigns)
	if totalCampaigns == 0 {
		r.updateProgress(ctx, job.ID, 100, "No campaigns to sync")
		return nil
	}

	// Sync metrics for each campaign
	for i, campaign := range campaigns {
		if err := r.waitForRateLimit(ctx, job.Platform); err != nil {
			return err
		}

		progress := 10 + (80 * (i + 1) / totalCampaigns)
		r.updateProgress(ctx, job.ID, progress, fmt.Sprintf("Syncing metrics for campaign %d/%d", i+1, totalCampaigns))

		metrics, err := connector.GetCampaignInsights(ctx, account.AccessToken, campaign.PlatformCampaignID, dateRange)
		if err != nil {
			log.Printf("[JobRunner] Error getting insights for campaign %s: %v", campaign.ID, err)
			job.RecordsFailed++
			continue
		}

		// Upsert metrics
		for j := range metrics {
			metrics[j].CampaignID = campaign.ID
			metrics[j].OrganizationID = account.OrganizationID
			if err := r.metricsRepo.UpsertCampaignMetrics(ctx, &metrics[j]); err != nil {
				log.Printf("[JobRunner] Error upserting metrics: %v", err)
				job.RecordsFailed++
			} else {
				job.RecordsProcessed++
			}
		}
	}

	r.updateProgress(ctx, job.ID, 100, "Metrics sync complete")

	return nil
}

func (r *JobRunner) syncCampaign(ctx context.Context, job *entity.SyncJob) error {
	if job.CampaignID == nil {
		return fmt.Errorf("campaign ID is required for campaign sync")
	}

	connector, ok := r.connectors[job.Platform]
	if !ok {
		return fmt.Errorf("no connector for platform: %s", job.Platform)
	}

	account, err := r.connAccountRepo.GetByID(ctx, job.ConnectedAccountID)
	if err != nil {
		return fmt.Errorf("failed to get connected account: %w", err)
	}

	if account.IsTokenExpired() {
		return fmt.Errorf("access token expired")
	}

	campaign, err := r.campaignRepo.GetByID(ctx, *job.CampaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Get date range
	dateRange := entity.DateRange{}
	if job.DateRangeStart != nil {
		dateRange.StartDate = *job.DateRangeStart
	} else {
		dateRange.StartDate = time.Now().AddDate(0, 0, -7)
	}
	if job.DateRangeEnd != nil {
		dateRange.EndDate = *job.DateRangeEnd
	} else {
		dateRange.EndDate = time.Now()
	}

	r.updateProgress(ctx, job.ID, 30, "Fetching campaign metrics...")

	if err := r.waitForRateLimit(ctx, job.Platform); err != nil {
		return err
	}

	metrics, err := connector.GetCampaignInsights(ctx, account.AccessToken, campaign.PlatformCampaignID, dateRange)
	if err != nil {
		return fmt.Errorf("failed to get campaign insights: %w", err)
	}

	r.updateProgress(ctx, job.ID, 70, "Saving metrics...")

	for i := range metrics {
		metrics[i].CampaignID = campaign.ID
		metrics[i].OrganizationID = account.OrganizationID
		if err := r.metricsRepo.UpsertCampaignMetrics(ctx, &metrics[i]); err != nil {
			log.Printf("[JobRunner] Error upserting metrics: %v", err)
			job.RecordsFailed++
		} else {
			job.RecordsProcessed++
		}
	}

	r.updateProgress(ctx, job.ID, 100, "Campaign sync complete")

	return nil
}

// ============================================================================
// Rate Limiting
// ============================================================================

func (r *JobRunner) waitForRateLimit(ctx context.Context, platform entity.Platform) error {
	r.rateLimiterMu.RLock()
	limiter, ok := r.rateLimiters[platform]
	r.rateLimiterMu.RUnlock()

	if !ok {
		return nil // No rate limiter for this platform
	}

	if err := limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	return nil
}

// SetRateLimitFromHeader updates rate limiter based on API response headers
func (r *JobRunner) SetRateLimitFromHeader(platform entity.Platform, remaining int, resetAt time.Time) {
	r.rateLimiterMu.Lock()
	defer r.rateLimiterMu.Unlock()

	if remaining <= 0 {
		// Set a very low limit until reset
		duration := time.Until(resetAt)
		if duration > 0 {
			r.rateLimiters[platform] = rate.NewLimiter(rate.Every(duration), 1)
		}
	}
}

// ============================================================================
// Error Handling
// ============================================================================

func (r *JobRunner) handleJobError(ctx context.Context, job *entity.SyncJob, err error) {
	log.Printf("[JobRunner] Job %s failed: %v", job.ID, err)

	// Log error
	errorLog := &entity.SyncErrorLog{
		ID:                 uuid.New(),
		SyncJobID:          job.ID,
		ConnectedAccountID: job.ConnectedAccountID,
		Platform:           job.Platform,
		ErrorType:          r.classifyError(err),
		ErrorMessage:       err.Error(),
		IsRetryable:        r.isRetryableError(err),
		CreatedAt:          time.Now(),
	}

	if logErr := r.errorLogRepo.Create(ctx, errorLog); logErr != nil {
		log.Printf("[JobRunner] Error logging sync error: %v", logErr)
	}

	// Check if retryable
	if job.CanRetry() && r.isRetryableError(err) {
		r.scheduleRetry(ctx, job, err)
	} else {
		// Mark as failed
		if markErr := r.syncJobRepo.MarkFailed(ctx, job.ID, err.Error(), r.classifyError(err)); markErr != nil {
			log.Printf("[JobRunner] Error marking job failed: %v", markErr)
		}

		// Update sync state
		if stateErr := r.syncStateRepo.UpdateSyncFailed(ctx, job.ConnectedAccountID, err.Error()); stateErr != nil {
			log.Printf("[JobRunner] Error updating sync state: %v", stateErr)
		}
	}
}

func (r *JobRunner) scheduleRetry(ctx context.Context, job *entity.SyncJob, err error) {
	config := r.configs[job.Platform]

	// Calculate exponential backoff
	backoff := config.RetryBackoffSeconds * (1 << job.RetryCount) // 2^retryCount
	if backoff > config.MaxRetryBackoff {
		backoff = config.MaxRetryBackoff
	}

	retryAt := time.Now().Add(time.Duration(backoff) * time.Second)

	// Add to retry queue
	entry := &entity.RetryQueueEntry{
		ID:                 uuid.New(),
		SyncJobID:          job.ID,
		ConnectedAccountID: job.ConnectedAccountID,
		Platform:           job.Platform,
		RetryAt:            retryAt,
		RetryCount:         job.RetryCount + 1,
		MaxRetries:         job.MaxRetries,
		LastError:          err.Error(),
		CreatedAt:          time.Now(),
	}

	if err := r.retryQueueRepo.Enqueue(ctx, entry); err != nil {
		log.Printf("[JobRunner] Error enqueueing retry: %v", err)
		return
	}

	// Update job status
	if err := r.syncJobRepo.ScheduleRetry(ctx, job.ID, retryAt, err.Error()); err != nil {
		log.Printf("[JobRunner] Error scheduling retry: %v", err)
	}

	log.Printf("[JobRunner] Job %s scheduled for retry at %s (attempt %d/%d)",
		job.ID, retryAt.Format(time.RFC3339), job.RetryCount+1, job.MaxRetries)
}

func (r *JobRunner) classifyError(err error) string {
	errStr := err.Error()

	switch {
	case contains(errStr, "rate limit", "too many requests", "throttle"):
		return "rate_limit"
	case contains(errStr, "token expired", "unauthorized", "invalid token"):
		return "auth"
	case contains(errStr, "network", "connection", "timeout"):
		return "network"
	case contains(errStr, "validation", "invalid"):
		return "validation"
	case contains(errStr, "not found"):
		return "not_found"
	default:
		return "api_error"
	}
}

func (r *JobRunner) isRetryableError(err error) bool {
	errType := r.classifyError(err)
	switch errType {
	case "rate_limit", "network":
		return true
	case "auth", "validation", "not_found":
		return false
	default:
		return true // Retry unknown errors
	}
}

// ============================================================================
// Retry Queue Processing
// ============================================================================

// ProcessRetryQueue processes entries from the retry queue
func (r *JobRunner) ProcessRetryQueue(ctx context.Context) error {
	entries, err := r.retryQueueRepo.Dequeue(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to dequeue retry entries: %w", err)
	}

	for _, entry := range entries {
		// Get the job
		job, err := r.syncJobRepo.GetByID(ctx, entry.SyncJobID)
		if err != nil {
			log.Printf("[JobRunner] Error getting job for retry: %v", err)
			continue
		}

		// Remove from retry queue
		if err := r.retryQueueRepo.Remove(ctx, entry.ID); err != nil {
			log.Printf("[JobRunner] Error removing retry entry: %v", err)
		}

		// Reset job for retry
		job.Status = entity.SyncStatusPending
		job.RetryCount = entry.RetryCount
		if err := r.syncJobRepo.Update(ctx, job); err != nil {
			log.Printf("[JobRunner] Error updating job for retry: %v", err)
			continue
		}

		log.Printf("[JobRunner] Job %s re-queued for retry (attempt %d)", job.ID, entry.RetryCount)
	}

	return nil
}

// ============================================================================
// Progress Updates
// ============================================================================

func (r *JobRunner) updateProgress(ctx context.Context, jobID uuid.UUID, percent int, message string) {
	if err := r.syncJobRepo.UpdateProgress(ctx, jobID, percent, message); err != nil {
		log.Printf("[JobRunner] Error updating progress: %v", err)
	}
}

// ============================================================================
// Utility Functions
// ============================================================================

func contains(s string, substrs ...string) bool {
	lower := toLower(s)
	for _, substr := range substrs {
		if containsStr(lower, toLower(substr)) {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && findSubstr(s, substr) >= 0))
}

func findSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
