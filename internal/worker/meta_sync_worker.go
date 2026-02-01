package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/ratelimit"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MetaSyncWorkerConfig holds configuration for the Meta sync worker
type MetaSyncWorkerConfig struct {
	// MaxConcurrency is the maximum number of concurrent account syncs
	MaxConcurrency int
	// RateLimitCalls is the number of API calls allowed per window (Meta: 200/hour)
	RateLimitCalls int
	// RateLimitWindow is the time window for rate limiting
	RateLimitWindow time.Duration
	// SyncTimeout is the timeout for syncing a single account
	SyncTimeout time.Duration
	// BatchTimeout is the timeout for the entire batch sync
	BatchTimeout time.Duration
	// RetryConfig for failed API calls
	RetryConfig *RetryConfig
	// DateRange for metrics sync
	DefaultDateRange entity.DateRange
}

// DefaultMetaSyncWorkerConfig returns production-ready defaults for Meta API
func DefaultMetaSyncWorkerConfig() *MetaSyncWorkerConfig {
	return &MetaSyncWorkerConfig{
		MaxConcurrency:  5,
		RateLimitCalls:  200,
		RateLimitWindow: time.Hour,
		SyncTimeout:     10 * time.Minute,
		BatchTimeout:    60 * time.Minute,
		RetryConfig:     DefaultRetryConfig(),
		DefaultDateRange: entity.DateRange{
			StartDate: time.Now().AddDate(0, 0, -7), // Last 7 days
			EndDate:   time.Now(),
		},
	}
}

// SyncTask represents a single sync task for an account
type SyncTask struct {
	AccountID      uuid.UUID
	OrganizationID uuid.UUID
	Platform       entity.Platform
	AccessToken    string
	AdAccountID    string // Platform-specific ad account ID
	DateRange      entity.DateRange
	SyncCampaigns  bool
	SyncAdSets     bool
	SyncAds        bool
	SyncMetrics    bool
}

// SyncTaskResult represents the result of a sync task
type SyncTaskResult struct {
	Task            SyncTask
	CampaignsSynced int
	AdSetsSynced    int
	AdsSynced       int
	MetricsSynced   int
	Duration        time.Duration
	Error           error
	Retries         int
}

// IsSuccess returns true if the sync completed successfully
func (r *SyncTaskResult) IsSuccess() bool {
	return r.Error == nil
}

// HasPartialSuccess returns true if some data was synced despite errors
func (r *SyncTaskResult) HasPartialSuccess() bool {
	return r.Error != nil && (r.CampaignsSynced > 0 || r.AdSetsSynced > 0 || r.AdsSynced > 0 || r.MetricsSynced > 0)
}

// BatchSyncResult represents the result of syncing multiple accounts
type BatchSyncResult struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalTasks    int
	Successful    int
	Failed        int
	PartialFailed int
	Results       []SyncTaskResult
}

// MetaSyncWorker handles background synchronization of Meta ads data
type MetaSyncWorker struct {
	config           *MetaSyncWorkerConfig
	connector        service.PlatformConnector
	connectedAccRepo repository.ConnectedAccountRepository
	adAccountRepo    repository.AdAccountRepository
	campaignRepo     repository.CampaignRepository
	adSetRepo        repository.AdSetRepository
	adRepo           repository.AdRepository
	metricsRepo      repository.MetricsRepository
	rateLimiter      *ratelimit.MultiLimiter
	retryer          *Retryer
	logger           zerolog.Logger
	isRunning        atomic.Bool
	runningTasks     int32
	mu               sync.RWMutex
	taskQueue        chan SyncTask
	resultQueue      chan SyncTaskResult
	stopChan         chan struct{}
}

// NewMetaSyncWorker creates a new Meta sync worker
func NewMetaSyncWorker(
	config *MetaSyncWorkerConfig,
	connector service.PlatformConnector,
	connectedAccRepo repository.ConnectedAccountRepository,
	adAccountRepo repository.AdAccountRepository,
	campaignRepo repository.CampaignRepository,
	adSetRepo repository.AdSetRepository,
	adRepo repository.AdRepository,
	metricsRepo repository.MetricsRepository,
	logger zerolog.Logger,
) *MetaSyncWorker {
	if config == nil {
		config = DefaultMetaSyncWorkerConfig()
	}

	return &MetaSyncWorker{
		config:           config,
		connector:        connector,
		connectedAccRepo: connectedAccRepo,
		adAccountRepo:    adAccountRepo,
		campaignRepo:     campaignRepo,
		adSetRepo:        adSetRepo,
		adRepo:           adRepo,
		metricsRepo:      metricsRepo,
		rateLimiter:      ratelimit.NewMultiLimiter(),
		retryer:          NewRetryer(config.RetryConfig, logger),
		logger:           logger.With().Str("worker", "meta_sync").Logger(),
		taskQueue:        make(chan SyncTask, 1000),
		resultQueue:      make(chan SyncTaskResult, 1000),
		stopChan:         make(chan struct{}),
	}
}

// Start starts the worker pool
func (w *MetaSyncWorker) Start(ctx context.Context) error {
	if w.isRunning.Load() {
		return errors.ErrBadRequest("Worker is already running")
	}

	w.isRunning.Store(true)
	w.logger.Info().
		Int("concurrency", w.config.MaxConcurrency).
		Int("rate_limit_calls", w.config.RateLimitCalls).
		Dur("rate_limit_window", w.config.RateLimitWindow).
		Msg("Starting Meta sync worker pool")

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < w.config.MaxConcurrency; i++ {
		wg.Add(1)
		go w.worker(ctx, &wg, i)
	}

	// Wait for workers to complete in a separate goroutine
	go func() {
		wg.Wait()
		w.isRunning.Store(false)
		w.logger.Info().Msg("Meta sync worker pool stopped")
	}()

	return nil
}

// Stop gracefully stops the worker pool
func (w *MetaSyncWorker) Stop() {
	if !w.isRunning.Load() {
		return
	}

	w.logger.Info().Msg("Stopping Meta sync worker pool")
	close(w.stopChan)
	close(w.taskQueue)
}

// IsRunning returns true if the worker is currently running
func (w *MetaSyncWorker) IsRunning() bool {
	return w.isRunning.Load()
}

// GetRunningTaskCount returns the number of currently running tasks
func (w *MetaSyncWorker) GetRunningTaskCount() int32 {
	return atomic.LoadInt32(&w.runningTasks)
}

// SubmitTask submits a sync task to the worker queue
func (w *MetaSyncWorker) SubmitTask(task SyncTask) error {
	if !w.isRunning.Load() {
		return errors.ErrBadRequest("Worker is not running")
	}

	select {
	case w.taskQueue <- task:
		w.logger.Debug().
			Str("account_id", task.AccountID.String()).
			Str("ad_account_id", task.AdAccountID).
			Msg("Task submitted to queue")
		return nil
	default:
		return errors.ErrInternal("Task queue is full")
	}
}

// SyncAllActiveAccounts syncs all active Meta accounts
func (w *MetaSyncWorker) SyncAllActiveAccounts(ctx context.Context) (*BatchSyncResult, error) {
	ctx, cancel := context.WithTimeout(ctx, w.config.BatchTimeout)
	defer cancel()

	result := &BatchSyncResult{
		StartTime: time.Now(),
		Results:   make([]SyncTaskResult, 0),
	}

	// Get all active Meta accounts
	accounts, err := w.connectedAccRepo.ListActive(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to list active accounts", 500)
	}

	// Filter only Meta accounts
	var metaAccounts []entity.ConnectedAccount
	for _, acc := range accounts {
		if acc.Platform == entity.PlatformMeta {
			metaAccounts = append(metaAccounts, acc)
		}
	}

	if len(metaAccounts) == 0 {
		w.logger.Info().Msg("No active Meta accounts to sync")
		result.EndTime = time.Now()
		return result, nil
	}

	result.TotalTasks = len(metaAccounts)

	w.logger.Info().
		Int("total_accounts", len(metaAccounts)).
		Msg("Starting batch sync for Meta accounts")

	// Create result collector
	resultChan := make(chan SyncTaskResult, len(metaAccounts))

	// Use semaphore for concurrency control
	sem := make(chan struct{}, w.config.MaxConcurrency)
	var wg sync.WaitGroup

	for _, account := range metaAccounts {
		wg.Add(1)
		go func(acc entity.ConnectedAccount) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultChan <- SyncTaskResult{
					Task: SyncTask{
						AccountID: acc.ID,
						Platform:  acc.Platform,
					},
					Error: ctx.Err(),
				}
				return
			}

			// Check if token is expired
			if acc.IsTokenExpired() {
				resultChan <- SyncTaskResult{
					Task: SyncTask{
						AccountID: acc.ID,
						Platform:  acc.Platform,
					},
					Error: errors.NewTokenExpiredError(acc.Platform.String(), "access", *acc.TokenExpiresAt),
				}
				return
			}

			// Create sync task
			task := SyncTask{
				AccountID:      acc.ID,
				OrganizationID: acc.OrganizationID,
				Platform:       acc.Platform,
				AccessToken:    acc.AccessToken,
				AdAccountID:    acc.PlatformAccountID,
				DateRange:      w.config.DefaultDateRange,
				SyncCampaigns:  true,
				SyncAdSets:     true,
				SyncAds:        true,
				SyncMetrics:    true,
			}

			// Execute sync
			syncResult := w.executeSync(ctx, task)
			resultChan <- syncResult
		}(account)
	}

	// Close result channel when all tasks complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for syncResult := range resultChan {
		result.Results = append(result.Results, syncResult)
		if syncResult.IsSuccess() {
			result.Successful++
		} else if syncResult.HasPartialSuccess() {
			result.PartialFailed++
		} else {
			result.Failed++
		}
	}

	result.EndTime = time.Now()

	w.logger.Info().
		Int("total", result.TotalTasks).
		Int("successful", result.Successful).
		Int("failed", result.Failed).
		Int("partial", result.PartialFailed).
		Dur("duration", result.EndTime.Sub(result.StartTime)).
		Msg("Batch sync completed")

	return result, nil
}

// SyncAccount syncs a single connected account
func (w *MetaSyncWorker) SyncAccount(ctx context.Context, accountID uuid.UUID) (*SyncTaskResult, error) {
	account, err := w.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, errors.ErrNotFound("Connected account")
	}

	if account.Platform != entity.PlatformMeta {
		return nil, errors.ErrBadRequest("Account is not a Meta account")
	}

	if account.IsTokenExpired() {
		return nil, errors.NewTokenExpiredError(account.Platform.String(), "access", *account.TokenExpiresAt)
	}

	task := SyncTask{
		AccountID:      account.ID,
		OrganizationID: account.OrganizationID,
		Platform:       account.Platform,
		AccessToken:    account.AccessToken,
		AdAccountID:    account.PlatformAccountID,
		DateRange:      w.config.DefaultDateRange,
		SyncCampaigns:  true,
		SyncAdSets:     true,
		SyncAds:        true,
		SyncMetrics:    true,
	}

	result := w.executeSync(ctx, task)
	return &result, nil
}

// worker is the main worker loop
func (w *MetaSyncWorker) worker(ctx context.Context, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()

	w.logger.Debug().Int("worker_id", workerID).Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug().Int("worker_id", workerID).Msg("Worker stopped due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Debug().Int("worker_id", workerID).Msg("Worker stopped")
			return
		case task, ok := <-w.taskQueue:
			if !ok {
				w.logger.Debug().Int("worker_id", workerID).Msg("Task queue closed, worker exiting")
				return
			}

			atomic.AddInt32(&w.runningTasks, 1)
			result := w.executeSync(ctx, task)
			atomic.AddInt32(&w.runningTasks, -1)

			select {
			case w.resultQueue <- result:
			default:
				w.logger.Warn().Msg("Result queue is full, dropping result")
			}
		}
	}
}

// executeSync performs the actual sync operation for a task
func (w *MetaSyncWorker) executeSync(ctx context.Context, task SyncTask) SyncTaskResult {
	ctx, cancel := context.WithTimeout(ctx, w.config.SyncTimeout)
	defer cancel()

	startTime := time.Now()
	result := SyncTaskResult{
		Task: task,
	}

	// Get rate limiter for this ad account
	limiterKey := fmt.Sprintf("meta_%s", task.AdAccountID)
	limiter := w.rateLimiter.GetOrCreate(limiterKey, w.config.RateLimitCalls, w.config.RateLimitWindow)

	w.logger.Info().
		Str("account_id", task.AccountID.String()).
		Str("ad_account_id", task.AdAccountID).
		Bool("sync_campaigns", task.SyncCampaigns).
		Bool("sync_adsets", task.SyncAdSets).
		Bool("sync_ads", task.SyncAds).
		Bool("sync_metrics", task.SyncMetrics).
		Msg("Starting account sync")

	// Get ad accounts first
	adAccounts, err := w.syncAdAccounts(ctx, task, limiter)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		w.logSyncError(task, err, "Failed to sync ad accounts")
		return result
	}

	// Sync campaigns, ad sets, ads, and metrics for each ad account
	for _, adAccount := range adAccounts {
		if task.SyncCampaigns {
			campaigns, err := w.syncCampaigns(ctx, task, adAccount.ID, adAccount.PlatformAdAccountID, limiter)
			if err != nil {
				w.logSyncError(task, err, "Failed to sync campaigns")
				result.Error = err
			} else {
				result.CampaignsSynced += len(campaigns)

				// Sync ad sets for each campaign
				if task.SyncAdSets {
					for _, campaign := range campaigns {
						adSets, err := w.syncAdSets(ctx, task, campaign.ID, campaign.PlatformCampaignID, limiter)
						if err != nil {
							w.logSyncError(task, err, "Failed to sync ad sets")
							if result.Error == nil {
								result.Error = err
							}
						} else {
							result.AdSetsSynced += len(adSets)

							// Sync ads for each ad set
							if task.SyncAds {
								for _, adSet := range adSets {
									ads, err := w.syncAds(ctx, task, campaign.ID, adSet.ID, adSet.PlatformAdSetID, limiter)
									if err != nil {
										w.logSyncError(task, err, "Failed to sync ads")
										if result.Error == nil {
											result.Error = err
										}
									} else {
										result.AdsSynced += len(ads)
									}
								}
							}
						}
					}
				}

				// Sync metrics for campaigns
				if task.SyncMetrics {
					for _, campaign := range campaigns {
						count, err := w.syncCampaignMetrics(ctx, task, campaign.ID, campaign.PlatformCampaignID, limiter)
						if err != nil {
							w.logSyncError(task, err, "Failed to sync campaign metrics")
							if result.Error == nil {
								result.Error = err
							}
						} else {
							result.MetricsSynced += count
						}
					}
				}
			}
		}
	}

	result.Duration = time.Since(startTime)

	// Update last synced timestamp
	if err := w.connectedAccRepo.UpdateLastSynced(ctx, task.AccountID); err != nil {
		w.logSyncError(task, err, "Failed to update last synced timestamp")
	}

	if result.Error == nil {
		w.logger.Info().
			Str("account_id", task.AccountID.String()).
			Int("campaigns", result.CampaignsSynced).
			Int("ad_sets", result.AdSetsSynced).
			Int("ads", result.AdsSynced).
			Int("metrics", result.MetricsSynced).
			Dur("duration", result.Duration).
			Msg("Account sync completed successfully")
	} else {
		w.logger.Warn().
			Str("account_id", task.AccountID.String()).
			Int("campaigns", result.CampaignsSynced).
			Int("ad_sets", result.AdSetsSynced).
			Int("ads", result.AdsSynced).
			Int("metrics", result.MetricsSynced).
			Dur("duration", result.Duration).
			Err(result.Error).
			Msg("Account sync completed with errors")
	}

	return result
}

// syncAdAccounts fetches and stores ad accounts
func (w *MetaSyncWorker) syncAdAccounts(ctx context.Context, task SyncTask, limiter *ratelimit.Limiter) ([]entity.AdAccount, error) {
	// Wait for rate limit
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	var platformAccounts []entity.PlatformAccount
	retryResult := w.retryer.Execute(ctx, "GetAdAccounts", func(ctx context.Context) error {
		var err error
		platformAccounts, err = w.connector.GetAdAccounts(ctx, task.AccessToken)
		return err
	})

	if !retryResult.Success {
		return nil, retryResult.LastError
	}

	// Transform and upsert ad accounts
	adAccounts := make([]entity.AdAccount, 0, len(platformAccounts))
	for _, pa := range platformAccounts {
		adAccount := entity.AdAccount{
			ConnectedAccountID:    task.AccountID,
			OrganizationID:        task.OrganizationID,
			Platform:              task.Platform,
			PlatformAdAccountID:   pa.ID,
			PlatformAdAccountName: pa.Name,
			Currency:              pa.Currency,
			Timezone:              pa.Timezone,
			IsActive:              pa.Status == "active" || pa.Status == "ACTIVE",
		}

		if err := w.adAccountRepo.Upsert(ctx, &adAccount); err != nil {
			w.logger.Error().Err(err).Str("platform_id", pa.ID).Msg("Failed to upsert ad account")
			continue
		}

		adAccounts = append(adAccounts, adAccount)
	}

	w.logger.Debug().
		Int("count", len(adAccounts)).
		Msg("Synced ad accounts")

	return adAccounts, nil
}

// syncCampaigns fetches and stores campaigns for an ad account
func (w *MetaSyncWorker) syncCampaigns(ctx context.Context, task SyncTask, adAccountID uuid.UUID, platformAdAccountID string, limiter *ratelimit.Limiter) ([]entity.Campaign, error) {
	// Wait for rate limit
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	var platformCampaigns []entity.Campaign
	retryResult := w.retryer.Execute(ctx, "GetCampaigns", func(ctx context.Context) error {
		var err error
		platformCampaigns, err = w.connector.GetCampaigns(ctx, task.AccessToken, platformAdAccountID)
		return err
	})

	if !retryResult.Success {
		return nil, retryResult.LastError
	}

	// Set organization and ad account IDs
	now := time.Now()
	for i := range platformCampaigns {
		platformCampaigns[i].OrganizationID = task.OrganizationID
		platformCampaigns[i].AdAccountID = adAccountID
		platformCampaigns[i].LastSyncedAt = &now
	}

	// Bulk upsert campaigns
	if err := w.campaignRepo.BulkUpsert(ctx, platformCampaigns); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to bulk upsert campaigns", 500)
	}

	w.logger.Debug().
		Int("count", len(platformCampaigns)).
		Str("ad_account_id", platformAdAccountID).
		Msg("Synced campaigns")

	// Fetch campaigns from DB to get their UUIDs
	campaigns, err := w.campaignRepo.ListByAdAccount(ctx, adAccountID)
	if err != nil {
		return nil, err
	}

	return campaigns, nil
}

// syncAdSets fetches and stores ad sets for a campaign
func (w *MetaSyncWorker) syncAdSets(ctx context.Context, task SyncTask, campaignID uuid.UUID, platformCampaignID string, limiter *ratelimit.Limiter) ([]entity.AdSet, error) {
	// Wait for rate limit
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	var platformAdSets []entity.AdSet
	retryResult := w.retryer.Execute(ctx, "GetAdSets", func(ctx context.Context) error {
		var err error
		platformAdSets, err = w.connector.GetAdSets(ctx, task.AccessToken, platformCampaignID)
		return err
	})

	if !retryResult.Success {
		return nil, retryResult.LastError
	}

	// Set organization and campaign IDs
	now := time.Now()
	for i := range platformAdSets {
		platformAdSets[i].OrganizationID = task.OrganizationID
		platformAdSets[i].CampaignID = campaignID
		platformAdSets[i].LastSyncedAt = &now
	}

	// Bulk upsert ad sets
	if err := w.adSetRepo.BulkUpsert(ctx, platformAdSets); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to bulk upsert ad sets", 500)
	}

	w.logger.Debug().
		Int("count", len(platformAdSets)).
		Str("campaign_id", platformCampaignID).
		Msg("Synced ad sets")

	// Fetch ad sets from DB to get their UUIDs
	adSets, err := w.adSetRepo.ListByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	return adSets, nil
}

// syncAds fetches and stores ads for an ad set
func (w *MetaSyncWorker) syncAds(ctx context.Context, task SyncTask, campaignID, adSetID uuid.UUID, platformAdSetID string, limiter *ratelimit.Limiter) ([]entity.Ad, error) {
	// Wait for rate limit
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	var platformAds []entity.Ad
	retryResult := w.retryer.Execute(ctx, "GetAds", func(ctx context.Context) error {
		var err error
		platformAds, err = w.connector.GetAds(ctx, task.AccessToken, platformAdSetID)
		return err
	})

	if !retryResult.Success {
		return nil, retryResult.LastError
	}

	// Set organization, campaign, and ad set IDs
	now := time.Now()
	for i := range platformAds {
		platformAds[i].OrganizationID = task.OrganizationID
		platformAds[i].CampaignID = campaignID
		platformAds[i].AdSetID = adSetID
		platformAds[i].LastSyncedAt = &now
	}

	// Bulk upsert ads
	if err := w.adRepo.BulkUpsert(ctx, platformAds); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to bulk upsert ads", 500)
	}

	w.logger.Debug().
		Int("count", len(platformAds)).
		Str("ad_set_id", platformAdSetID).
		Msg("Synced ads")

	// Fetch ads from DB to get their UUIDs
	ads, err := w.adRepo.ListByAdSet(ctx, adSetID)
	if err != nil {
		return nil, err
	}

	return ads, nil
}

// syncCampaignMetrics fetches and stores metrics for a campaign
func (w *MetaSyncWorker) syncCampaignMetrics(ctx context.Context, task SyncTask, campaignID uuid.UUID, platformCampaignID string, limiter *ratelimit.Limiter) (int, error) {
	// Wait for rate limit
	if err := limiter.Wait(ctx); err != nil {
		return 0, err
	}

	var metrics []entity.CampaignMetricsDaily
	retryResult := w.retryer.Execute(ctx, "GetCampaignInsights", func(ctx context.Context) error {
		var err error
		metrics, err = w.connector.GetCampaignInsights(ctx, task.AccessToken, platformCampaignID, task.DateRange)
		return err
	})

	if !retryResult.Success {
		return 0, retryResult.LastError
	}

	if len(metrics) == 0 {
		return 0, nil
	}

	// Set organization and campaign IDs, calculate derived metrics
	now := time.Now()
	for i := range metrics {
		metrics[i].OrganizationID = task.OrganizationID
		metrics[i].CampaignID = campaignID
		metrics[i].Platform = task.Platform
		metrics[i].LastSyncedAt = &now
		metrics[i].CalculateDerivedMetrics()
	}

	// Bulk upsert metrics
	if err := w.metricsRepo.BulkUpsertCampaignMetrics(ctx, metrics); err != nil {
		return 0, errors.Wrap(err, errors.ErrCodeInternal, "Failed to bulk upsert campaign metrics", 500)
	}

	w.logger.Debug().
		Int("count", len(metrics)).
		Str("campaign_id", platformCampaignID).
		Msg("Synced campaign metrics")

	return len(metrics), nil
}

// logSyncError logs a sync error with context
func (w *MetaSyncWorker) logSyncError(task SyncTask, err error, msg string) {
	w.logger.Error().
		Err(err).
		Str("account_id", task.AccountID.String()).
		Str("ad_account_id", task.AdAccountID).
		Str("platform", string(task.Platform)).
		Msg(msg)
}

// GetStats returns current worker statistics
func (w *MetaSyncWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"is_running":    w.isRunning.Load(),
		"running_tasks": atomic.LoadInt32(&w.runningTasks),
		"concurrency":   w.config.MaxConcurrency,
		"rate_limit":    fmt.Sprintf("%d calls/%s", w.config.RateLimitCalls, w.config.RateLimitWindow),
	}
}
