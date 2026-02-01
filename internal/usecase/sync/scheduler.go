package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/google/uuid"
)

// ============================================================================
// Scheduler Service
// ============================================================================

// Scheduler manages scheduled sync jobs with cron-like functionality
type Scheduler struct {
	syncStateRepo   repository.SyncStateRepository
	syncJobRepo     repository.SyncJobRepository
	connAccountRepo repository.ConnectedAccountRepository

	// Configuration
	configs map[entity.Platform]entity.SyncScheduleConfig

	// Control
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	isRunning  bool
	mu         sync.RWMutex

	// Tick interval for checking scheduled jobs
	tickInterval time.Duration

	// Worker ID for job claiming
	workerID string

	// Callback for job execution
	jobExecutor JobExecutor
}

// JobExecutor is the interface for executing sync jobs
type JobExecutor interface {
	ExecuteJob(ctx context.Context, job *entity.SyncJob) error
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	TickInterval time.Duration
	WorkerID     string
	Configs      map[entity.Platform]entity.SyncScheduleConfig
}

// DefaultSchedulerConfig returns default scheduler configuration
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		TickInterval: 1 * time.Minute,
		WorkerID:     fmt.Sprintf("worker-%s", uuid.New().String()[:8]),
		Configs:      entity.DefaultSyncScheduleConfigs(),
	}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(
	syncStateRepo repository.SyncStateRepository,
	syncJobRepo repository.SyncJobRepository,
	connAccountRepo repository.ConnectedAccountRepository,
	config *SchedulerConfig,
) *Scheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	return &Scheduler{
		syncStateRepo:   syncStateRepo,
		syncJobRepo:     syncJobRepo,
		connAccountRepo: connAccountRepo,
		configs:         config.Configs,
		tickInterval:    config.TickInterval,
		workerID:        config.WorkerID,
	}
}

// SetJobExecutor sets the job executor callback
func (s *Scheduler) SetJobExecutor(executor JobExecutor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobExecutor = executor
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.isRunning = true
	s.mu.Unlock()

	log.Printf("[Scheduler] Starting with worker ID: %s", s.workerID)

	// Start the main scheduler loop
	s.wg.Add(1)
	go s.run()

	// Start hourly sync checker for each platform
	for platform := range s.configs {
		s.wg.Add(1)
		go s.runHourlySyncChecker(platform)
	}

	// Start daily sync checker
	s.wg.Add(1)
	go s.runDailySyncChecker()

	// Start job processor
	s.wg.Add(1)
	go s.runJobProcessor()

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	log.Printf("[Scheduler] Stopping...")
	s.cancel()
	s.wg.Wait()

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	log.Printf("[Scheduler] Stopped")
	return nil
}

// IsRunning returns whether the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// ============================================================================
// Main Scheduler Loop
// ============================================================================

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	// This is a lightweight tick to check system health
	// Actual sync scheduling is done in dedicated goroutines
	log.Printf("[Scheduler] Tick at %s", time.Now().Format(time.RFC3339))
}

// ============================================================================
// Hourly Sync Checker
// ============================================================================

func (s *Scheduler) runHourlySyncChecker(platform entity.Platform) {
	defer s.wg.Done()

	config, ok := s.configs[platform]
	if !ok || !config.HourlySyncEnabled {
		log.Printf("[Scheduler] Hourly sync disabled for %s", platform)
		return
	}

	log.Printf("[Scheduler] Starting hourly sync checker for %s at minute %d", platform, config.HourlySyncMinuteOffset)

	for {
		// Calculate next run time
		nextRun := s.calculateNextHourlyRun(config.HourlySyncMinuteOffset)
		waitDuration := time.Until(nextRun)

		log.Printf("[Scheduler] Next hourly sync for %s at %s (waiting %s)",
			platform, nextRun.Format(time.RFC3339), waitDuration)

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(waitDuration):
			s.triggerHourlySync(platform)
		}
	}
}

func (s *Scheduler) calculateNextHourlyRun(minuteOffset int) time.Time {
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minuteOffset, 0, 0, now.Location())

	// If the minute offset has already passed this hour, schedule for next hour
	if nextRun.Before(now) || nextRun.Equal(now) {
		nextRun = nextRun.Add(time.Hour)
	}

	return nextRun
}

func (s *Scheduler) triggerHourlySync(platform entity.Platform) {
	log.Printf("[Scheduler] Triggering hourly sync for %s", platform)

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	// Get accounts due for hourly sync
	states, err := s.syncStateRepo.ListDueForHourlySync(ctx, platform)
	if err != nil {
		log.Printf("[Scheduler] Error getting accounts for hourly sync: %v", err)
		return
	}

	log.Printf("[Scheduler] Found %d accounts for hourly sync on %s", len(states), platform)

	for _, state := range states {
		if !state.CanSync() {
			log.Printf("[Scheduler] Skipping account %s - cannot sync (syncing: %t, rate limited: %t)",
				state.ConnectedAccountID, state.IsSyncing, state.IsRateLimited())
			continue
		}

		// Create sync job
		job := &entity.SyncJob{
			BaseEntity:         entity.NewBaseEntity(),
			OrganizationID:     state.OrganizationID,
			ConnectedAccountID: state.ConnectedAccountID,
			Platform:           platform,
			SyncType:           entity.SyncTypeHourly,
			SyncScope:          entity.SyncScopeMetrics,
			Status:             entity.SyncStatusPending,
			Priority:           1,
			ScheduledAt:        time.Now(),
			MaxRetries:         s.configs[platform].MaxRetries,
			TriggeredBy:        "scheduler",
		}

		// Set date range for hourly sync (last 2 days to catch any delays)
		now := time.Now()
		start := now.AddDate(0, 0, -2)
		job.DateRangeStart = &start
		job.DateRangeEnd = &now

		if err := s.syncJobRepo.Create(ctx, job); err != nil {
			log.Printf("[Scheduler] Error creating hourly sync job: %v", err)
			continue
		}

		log.Printf("[Scheduler] Created hourly sync job %s for account %s",
			job.ID, state.ConnectedAccountID)
	}
}

// ============================================================================
// Daily Sync Checker
// ============================================================================

func (s *Scheduler) runDailySyncChecker() {
	defer s.wg.Done()

	log.Printf("[Scheduler] Starting daily sync checker")

	for {
		// Calculate next daily run (check every hour if it's time)
		nextCheck := time.Now().Truncate(time.Hour).Add(time.Hour)
		waitDuration := time.Until(nextCheck)

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(waitDuration):
			currentHour := time.Now().Hour()
			s.checkDailySync(currentHour)
		}
	}
}

func (s *Scheduler) checkDailySync(currentHour int) {
	for platform, config := range s.configs {
		if !config.DailySyncEnabled {
			continue
		}

		if config.DailySyncHour != currentHour {
			continue
		}

		log.Printf("[Scheduler] Triggering daily sync for %s", platform)
		s.triggerDailySync(platform, config)
	}
}

func (s *Scheduler) triggerDailySync(platform entity.Platform, config entity.SyncScheduleConfig) {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Minute)
	defer cancel()

	// Get accounts due for daily sync
	states, err := s.syncStateRepo.ListDueForDailySync(ctx, platform, config.DailySyncHour)
	if err != nil {
		log.Printf("[Scheduler] Error getting accounts for daily sync: %v", err)
		return
	}

	log.Printf("[Scheduler] Found %d accounts for daily sync on %s", len(states), platform)

	for _, state := range states {
		if !state.CanSync() {
			continue
		}

		// Create sync job
		job := &entity.SyncJob{
			BaseEntity:         entity.NewBaseEntity(),
			OrganizationID:     state.OrganizationID,
			ConnectedAccountID: state.ConnectedAccountID,
			Platform:           platform,
			SyncType:           entity.SyncTypeDaily,
			SyncScope:          entity.SyncScopeAccount, // Full account sync
			Status:             entity.SyncStatusPending,
			Priority:           0, // Lower priority than hourly
			ScheduledAt:        time.Now(),
			MaxRetries:         config.MaxRetries,
			TriggeredBy:        "scheduler",
		}

		// Set date range for daily sync (lookback days)
		now := time.Now()
		start := now.AddDate(0, 0, -config.DailySyncLookbackDays)
		job.DateRangeStart = &start
		job.DateRangeEnd = &now

		if err := s.syncJobRepo.Create(ctx, job); err != nil {
			log.Printf("[Scheduler] Error creating daily sync job: %v", err)
			continue
		}

		log.Printf("[Scheduler] Created daily sync job %s for account %s",
			job.ID, state.ConnectedAccountID)
	}
}

// ============================================================================
// Job Processor
// ============================================================================

func (s *Scheduler) runJobProcessor() {
	defer s.wg.Done()

	log.Printf("[Scheduler] Starting job processor")

	ticker := time.NewTicker(10 * time.Second) // Check for jobs every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processNextJob()
		}
	}
}

func (s *Scheduler) processNextJob() {
	s.mu.RLock()
	executor := s.jobExecutor
	s.mu.RUnlock()

	if executor == nil {
		return // No executor set
	}

	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// Get pending jobs
	jobs, err := s.syncJobRepo.ListPending(ctx, 1)
	if err != nil {
		log.Printf("[Scheduler] Error listing pending jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	job := jobs[0]

	// Try to claim the job
	if err := s.syncJobRepo.ClaimJob(ctx, job.ID, s.workerID); err != nil {
		log.Printf("[Scheduler] Failed to claim job %s: %v", job.ID, err)
		return
	}

	log.Printf("[Scheduler] Processing job %s for %s", job.ID, job.Platform)

	// Execute the job
	if err := executor.ExecuteJob(s.ctx, &job); err != nil {
		log.Printf("[Scheduler] Job %s failed: %v", job.ID, err)
		s.handleJobFailure(ctx, &job, err)
	} else {
		log.Printf("[Scheduler] Job %s completed successfully", job.ID)
	}
}

func (s *Scheduler) handleJobFailure(ctx context.Context, job *entity.SyncJob, err error) {
	config := s.configs[job.Platform]

	if job.CanRetry() {
		// Calculate backoff
		backoff := config.RetryBackoffSeconds * (job.RetryCount + 1)
		if backoff > config.MaxRetryBackoff {
			backoff = config.MaxRetryBackoff
		}

		retryAt := time.Now().Add(time.Duration(backoff) * time.Second)
		if schedErr := s.syncJobRepo.ScheduleRetry(ctx, job.ID, retryAt, err.Error()); schedErr != nil {
			log.Printf("[Scheduler] Failed to schedule retry for job %s: %v", job.ID, schedErr)
		} else {
			log.Printf("[Scheduler] Job %s scheduled for retry at %s", job.ID, retryAt.Format(time.RFC3339))
		}
	} else {
		// Mark as failed
		if markErr := s.syncJobRepo.MarkFailed(ctx, job.ID, err.Error(), "MAX_RETRIES_EXCEEDED"); markErr != nil {
			log.Printf("[Scheduler] Failed to mark job %s as failed: %v", job.ID, markErr)
		}

		// Update sync state
		if stateErr := s.syncStateRepo.UpdateSyncFailed(ctx, job.ConnectedAccountID, err.Error()); stateErr != nil {
			log.Printf("[Scheduler] Failed to update sync state: %v", stateErr)
		}
	}
}

// ============================================================================
// Manual Sync Trigger
// ============================================================================

// TriggerManualSync creates a manual sync job for an account
func (s *Scheduler) TriggerManualSync(ctx context.Context, req ManualSyncRequest) (*entity.SyncJob, error) {
	// Get the connected account
	account, err := s.connAccountRepo.GetByID(ctx, req.ConnectedAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected account: %w", err)
	}

	// Get sync state
	state, err := s.syncStateRepo.GetByConnectedAccount(ctx, req.ConnectedAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}

	// Check if sync is possible
	if state.IsSyncing {
		return nil, fmt.Errorf("sync already in progress for this account")
	}

	if state.IsRateLimited() {
		return nil, fmt.Errorf("account is rate limited until %s", state.RateLimitResetAt.Format(time.RFC3339))
	}

	// Create sync job
	job := &entity.SyncJob{
		BaseEntity:         entity.NewBaseEntity(),
		OrganizationID:     account.OrganizationID,
		ConnectedAccountID: req.ConnectedAccountID,
		Platform:           account.Platform,
		SyncType:           entity.SyncTypeManual,
		SyncScope:          req.Scope,
		Status:             entity.SyncStatusPending,
		Priority:           10, // High priority for manual syncs
		ScheduledAt:        time.Now(),
		MaxRetries:         1, // Less retries for manual syncs
		TriggeredBy:        fmt.Sprintf("user:%s", req.UserID),
		TriggeredByUser:    &req.UserID,
	}

	// Set date range
	if req.DateRangeStart != nil {
		job.DateRangeStart = req.DateRangeStart
	} else {
		start := time.Now().AddDate(0, 0, -7)
		job.DateRangeStart = &start
	}

	if req.DateRangeEnd != nil {
		job.DateRangeEnd = req.DateRangeEnd
	} else {
		end := time.Now()
		job.DateRangeEnd = &end
	}

	if req.CampaignID != nil {
		job.CampaignID = req.CampaignID
		job.SyncScope = entity.SyncScopeCampaign
	}

	if err := s.syncJobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create sync job: %w", err)
	}

	log.Printf("[Scheduler] Created manual sync job %s for account %s by user %s",
		job.ID, req.ConnectedAccountID, req.UserID)

	return job, nil
}

// ManualSyncRequest represents a request for manual sync
type ManualSyncRequest struct {
	ConnectedAccountID uuid.UUID         `json:"connected_account_id"`
	UserID             uuid.UUID         `json:"user_id"`
	Scope              entity.SyncScope  `json:"scope"`
	CampaignID         *uuid.UUID        `json:"campaign_id,omitempty"`
	DateRangeStart     *time.Time        `json:"date_range_start,omitempty"`
	DateRangeEnd       *time.Time        `json:"date_range_end,omitempty"`
}

// ============================================================================
// Utility Methods
// ============================================================================

// GetScheduleConfig returns the schedule config for a platform
func (s *Scheduler) GetScheduleConfig(platform entity.Platform) (entity.SyncScheduleConfig, bool) {
	config, ok := s.configs[platform]
	return config, ok
}

// UpdateScheduleConfig updates the schedule config for a platform
func (s *Scheduler) UpdateScheduleConfig(platform entity.Platform, config entity.SyncScheduleConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[platform] = config
}

// GetWorkerID returns the worker ID
func (s *Scheduler) GetWorkerID() string {
	return s.workerID
}
