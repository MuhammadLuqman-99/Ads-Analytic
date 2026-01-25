package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	domainService "github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

// Scheduler manages scheduled jobs for data synchronization
type Scheduler struct {
	cron             *cron.Cron
	syncService      SyncService
	tokenManager     TokenManager
	connectedAccRepo repository.ConnectedAccountRepository
	logger           zerolog.Logger
	mu               sync.RWMutex
	running          bool
	jobs             map[string]cron.EntryID
}

// SyncService interface for sync operations
type SyncService interface {
	SyncAccount(ctx context.Context, accountID uuid.UUID) (*domainService.SyncResult, error)
	SyncMetrics(ctx context.Context, accountID uuid.UUID, dateRange entity.DateRange) (int, error)
	SyncAllActive(ctx context.Context) (*domainService.BatchSyncResult, error)
}

// TokenManager interface for token refresh operations
type TokenManager interface {
	RefreshTokenIfNeeded(ctx context.Context, accountID uuid.UUID) error
	RefreshAllExpiring(ctx context.Context) (int, error)
}

// Config holds scheduler configuration
type Config struct {
	Enabled                    bool
	SyncCronSchedule           string // e.g., "0 * * * *" for every hour
	TokenRefreshCronSchedule   string // e.g., "*/30 * * * *" for every 30 minutes
	MetricsAggregationSchedule string // e.g., "0 */6 * * *" for every 6 hours
	ConcurrentSyncs            int
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:                    true,
		SyncCronSchedule:           "0 * * * *",    // Every hour
		TokenRefreshCronSchedule:   "*/30 * * * *", // Every 30 minutes
		MetricsAggregationSchedule: "0 */6 * * *",  // Every 6 hours
		ConcurrentSyncs:            3,
	}
}

// NewScheduler creates a new scheduler
func NewScheduler(
	syncService SyncService,
	tokenManager TokenManager,
	connectedAccRepo repository.ConnectedAccountRepository,
	logger zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		cron:             cron.New(cron.WithSeconds()),
		syncService:      syncService,
		tokenManager:     tokenManager,
		connectedAccRepo: connectedAccRepo,
		logger:           logger.With().Str("component", "scheduler").Logger(),
		jobs:             make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler with the given configuration
func (s *Scheduler) Start(config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	if config == nil {
		config = DefaultConfig()
	}

	if !config.Enabled {
		s.logger.Info().Msg("Scheduler is disabled")
		return nil
	}

	// Schedule data sync job
	if config.SyncCronSchedule != "" {
		id, err := s.cron.AddFunc(config.SyncCronSchedule, s.runDataSync)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to schedule data sync job")
			return err
		}
		s.jobs["data_sync"] = id
		s.logger.Info().Str("schedule", config.SyncCronSchedule).Msg("Scheduled data sync job")
	}

	// Schedule token refresh job
	if config.TokenRefreshCronSchedule != "" {
		id, err := s.cron.AddFunc(config.TokenRefreshCronSchedule, s.runTokenRefresh)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to schedule token refresh job")
			return err
		}
		s.jobs["token_refresh"] = id
		s.logger.Info().Str("schedule", config.TokenRefreshCronSchedule).Msg("Scheduled token refresh job")
	}

	// Schedule metrics aggregation job
	if config.MetricsAggregationSchedule != "" {
		id, err := s.cron.AddFunc(config.MetricsAggregationSchedule, s.runMetricsAggregation)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to schedule metrics aggregation job")
			return err
		}
		s.jobs["metrics_aggregation"] = id
		s.logger.Info().Str("schedule", config.MetricsAggregationSchedule).Msg("Scheduled metrics aggregation job")
	}

	s.cron.Start()
	s.running = true
	s.logger.Info().Msg("Scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	ctx := s.cron.Stop()
	<-ctx.Done()
	s.running = false
	s.logger.Info().Msg("Scheduler stopped")
}

// IsRunning returns true if the scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetNextRun returns the next scheduled run time for a job
func (s *Scheduler) GetNextRun(jobName string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, ok := s.jobs[jobName]; ok {
		entry := s.cron.Entry(id)
		return entry.Next, true
	}
	return time.Time{}, false
}

// RunNow manually triggers a job
func (s *Scheduler) RunNow(jobName string) error {
	switch jobName {
	case "data_sync":
		go s.runDataSync()
	case "token_refresh":
		go s.runTokenRefresh()
	case "metrics_aggregation":
		go s.runMetricsAggregation()
	default:
		return ErrUnknownJob
	}
	return nil
}

// Job implementations

func (s *Scheduler) runDataSync() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	s.logger.Info().Msg("Starting scheduled data sync")
	startTime := time.Now()

	result, err := s.syncService.SyncAllActive(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Scheduled data sync failed")
		return
	}

	duration := time.Since(startTime)
	s.logger.Info().
		Int("total_accounts", result.TotalAccounts).
		Int("success", result.SuccessCount).
		Int("partial", result.PartialCount).
		Int("failed", result.FailureCount).
		Dur("duration", duration).
		Msg("Scheduled data sync completed")
}

func (s *Scheduler) runTokenRefresh() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	s.logger.Info().Msg("Starting scheduled token refresh")
	startTime := time.Now()

	// Get accounts with expiring tokens
	accounts, err := s.connectedAccRepo.ListExpiring(ctx, 60) // Expiring in 60 minutes
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get expiring accounts")
		return
	}

	refreshed := 0
	failed := 0

	for _, account := range accounts {
		if err := s.tokenManager.RefreshTokenIfNeeded(ctx, account.ID); err != nil {
			s.logger.Error().
				Err(err).
				Str("account_id", account.ID.String()).
				Str("platform", string(account.Platform)).
				Msg("Failed to refresh token")
			failed++
			continue
		}
		refreshed++
	}

	duration := time.Since(startTime)
	s.logger.Info().
		Int("refreshed", refreshed).
		Int("failed", failed).
		Dur("duration", duration).
		Msg("Scheduled token refresh completed")
}

func (s *Scheduler) runMetricsAggregation() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	s.logger.Info().Msg("Starting scheduled metrics aggregation")
	startTime := time.Now()

	// Get all active accounts
	accounts, err := s.connectedAccRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get active accounts")
		return
	}

	// Define date range for metrics sync (last 7 days)
	dateRange := entity.DateRange{
		StartDate: time.Now().AddDate(0, 0, -7),
		EndDate:   time.Now(),
	}

	totalMetrics := 0
	for _, account := range accounts {
		count, err := s.syncService.SyncMetrics(ctx, account.ID, dateRange)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("account_id", account.ID.String()).
				Msg("Failed to sync metrics for account")
			continue
		}
		totalMetrics += count
	}

	duration := time.Since(startTime)
	s.logger.Info().
		Int("accounts_processed", len(accounts)).
		Int("metrics_synced", totalMetrics).
		Dur("duration", duration).
		Msg("Scheduled metrics aggregation completed")
}

// JobStatus represents the status of a scheduled job
type JobStatus struct {
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`
	LastRun   time.Time `json:"last_run,omitempty"`
	NextRun   time.Time `json:"next_run"`
	IsRunning bool      `json:"is_running"`
}

// GetJobStatuses returns the status of all scheduled jobs
func (s *Scheduler) GetJobStatuses() []JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]JobStatus, 0, len(s.jobs))
	for name, id := range s.jobs {
		entry := s.cron.Entry(id)
		statuses = append(statuses, JobStatus{
			Name:    name,
			NextRun: entry.Next,
			LastRun: entry.Prev,
		})
	}
	return statuses
}

// Errors
var (
	ErrUnknownJob = &SchedulerError{Message: "unknown job name"}
)

// SchedulerError represents a scheduler error
type SchedulerError struct {
	Message string
}

func (e *SchedulerError) Error() string {
	return e.Message
}
