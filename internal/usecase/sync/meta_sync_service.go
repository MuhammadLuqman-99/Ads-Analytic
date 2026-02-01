package sync

import (
	"context"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/internal/worker"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MetaSyncService provides high-level sync operations for Meta ads data
type MetaSyncService struct {
	worker           *worker.MetaSyncWorker
	connector        service.PlatformConnector
	connRegistry     service.ConnectorFactory
	connectedAccRepo repository.ConnectedAccountRepository
	adAccountRepo    repository.AdAccountRepository
	campaignRepo     repository.CampaignRepository
	adSetRepo        repository.AdSetRepository
	adRepo           repository.AdRepository
	metricsRepo      repository.MetricsRepository
	logger           zerolog.Logger
	mu               sync.RWMutex
}

// MetaSyncServiceConfig holds configuration for the sync service
type MetaSyncServiceConfig struct {
	WorkerConfig *worker.MetaSyncWorkerConfig
}

// DefaultMetaSyncServiceConfig returns default configuration
func DefaultMetaSyncServiceConfig() *MetaSyncServiceConfig {
	return &MetaSyncServiceConfig{
		WorkerConfig: worker.DefaultMetaSyncWorkerConfig(),
	}
}

// NewMetaSyncService creates a new Meta sync service
func NewMetaSyncService(
	config *MetaSyncServiceConfig,
	connector service.PlatformConnector,
	connRegistry service.ConnectorFactory,
	connectedAccRepo repository.ConnectedAccountRepository,
	adAccountRepo repository.AdAccountRepository,
	campaignRepo repository.CampaignRepository,
	adSetRepo repository.AdSetRepository,
	adRepo repository.AdRepository,
	metricsRepo repository.MetricsRepository,
	logger zerolog.Logger,
) *MetaSyncService {
	if config == nil {
		config = DefaultMetaSyncServiceConfig()
	}

	svc := &MetaSyncService{
		connector:        connector,
		connRegistry:     connRegistry,
		connectedAccRepo: connectedAccRepo,
		adAccountRepo:    adAccountRepo,
		campaignRepo:     campaignRepo,
		adSetRepo:        adSetRepo,
		adRepo:           adRepo,
		metricsRepo:      metricsRepo,
		logger:           logger.With().Str("service", "meta_sync").Logger(),
	}

	// Create the worker
	svc.worker = worker.NewMetaSyncWorker(
		config.WorkerConfig,
		connector,
		connectedAccRepo,
		adAccountRepo,
		campaignRepo,
		adSetRepo,
		adRepo,
		metricsRepo,
		logger,
	)

	return svc
}

// Start starts the sync service and its worker pool
func (s *MetaSyncService) Start(ctx context.Context) error {
	s.logger.Info().Msg("Starting Meta sync service")
	return s.worker.Start(ctx)
}

// Stop stops the sync service
func (s *MetaSyncService) Stop() {
	s.logger.Info().Msg("Stopping Meta sync service")
	s.worker.Stop()
}

// SyncAccount synchronizes all data for a connected Meta account
func (s *MetaSyncService) SyncAccount(ctx context.Context, accountID uuid.UUID) (*service.SyncResult, error) {
	startTime := time.Now()

	s.logger.Info().
		Str("account_id", accountID.String()).
		Msg("Starting single account sync")

	result, err := s.worker.SyncAccount(ctx, accountID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("account_id", accountID.String()).
			Msg("Failed to sync account")
		return nil, err
	}

	syncResult := &service.SyncResult{
		Platform:        entity.PlatformMeta,
		AccountID:       accountID,
		CampaignsSynced: result.CampaignsSynced,
		AdSetsSynced:    result.AdSetsSynced,
		AdsSynced:       result.AdsSynced,
		MetricsSynced:   result.MetricsSynced,
		StartedAt:       startTime.Unix(),
		CompletedAt:     time.Now().Unix(),
	}

	if result.Error != nil {
		syncResult.Errors = []error{result.Error}
	}

	return syncResult, nil
}

// SyncCampaigns synchronizes only campaigns for an account
func (s *MetaSyncService) SyncCampaigns(ctx context.Context, accountID uuid.UUID) (int, error) {
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, errors.ErrNotFound("Connected account")
	}

	if account.Platform != entity.PlatformMeta {
		return 0, errors.ErrBadRequest("Account is not a Meta account")
	}

	if account.IsTokenExpired() {
		return 0, errors.NewTokenExpiredError(account.Platform.String(), "access", *account.TokenExpiresAt)
	}

	// Get ad accounts
	adAccounts, err := s.adAccountRepo.ListByConnectedAccount(ctx, accountID)
	if err != nil {
		return 0, err
	}

	totalCampaigns := 0
	retryer := worker.NewRetryer(worker.DefaultRetryConfig(), s.logger)

	for _, adAccount := range adAccounts {
		var campaigns []entity.Campaign
		retryResult := retryer.Execute(ctx, "GetCampaigns", func(ctx context.Context) error {
			var err error
			campaigns, err = s.connector.GetCampaigns(ctx, account.AccessToken, adAccount.PlatformAdAccountID)
			return err
		})

		if !retryResult.Success {
			s.logger.Error().
				Err(retryResult.LastError).
				Str("ad_account_id", adAccount.PlatformAdAccountID).
				Msg("Failed to fetch campaigns")
			continue
		}

		// Set organization and ad account IDs
		now := time.Now()
		for i := range campaigns {
			campaigns[i].OrganizationID = account.OrganizationID
			campaigns[i].AdAccountID = adAccount.ID
			campaigns[i].LastSyncedAt = &now
		}

		// Bulk upsert
		if err := s.campaignRepo.BulkUpsert(ctx, campaigns); err != nil {
			s.logger.Error().
				Err(err).
				Str("ad_account_id", adAccount.PlatformAdAccountID).
				Msg("Failed to bulk upsert campaigns")
			continue
		}

		totalCampaigns += len(campaigns)
	}

	s.logger.Info().
		Str("account_id", accountID.String()).
		Int("total_campaigns", totalCampaigns).
		Msg("Campaigns sync completed")

	return totalCampaigns, nil
}

// SyncMetrics synchronizes only metrics for an account within the given date range
func (s *MetaSyncService) SyncMetrics(ctx context.Context, accountID uuid.UUID, dateRange entity.DateRange) (int, error) {
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, errors.ErrNotFound("Connected account")
	}

	if account.Platform != entity.PlatformMeta {
		return 0, errors.ErrBadRequest("Account is not a Meta account")
	}

	if account.IsTokenExpired() {
		return 0, errors.NewTokenExpiredError(account.Platform.String(), "access", *account.TokenExpiresAt)
	}

	// Get ad accounts
	adAccounts, err := s.adAccountRepo.ListByConnectedAccount(ctx, accountID)
	if err != nil {
		return 0, err
	}

	totalMetrics := 0
	retryer := worker.NewRetryer(worker.DefaultRetryConfig(), s.logger)

	for _, adAccount := range adAccounts {
		// Get campaigns for this ad account
		campaigns, err := s.campaignRepo.ListByAdAccount(ctx, adAccount.ID)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("ad_account_id", adAccount.PlatformAdAccountID).
				Msg("Failed to list campaigns")
			continue
		}

		for _, campaign := range campaigns {
			var metrics []entity.CampaignMetricsDaily
			retryResult := retryer.Execute(ctx, "GetCampaignInsights", func(ctx context.Context) error {
				var err error
				metrics, err = s.connector.GetCampaignInsights(ctx, account.AccessToken, campaign.PlatformCampaignID, dateRange)
				return err
			})

			if !retryResult.Success {
				s.logger.Error().
					Err(retryResult.LastError).
					Str("campaign_id", campaign.PlatformCampaignID).
					Msg("Failed to fetch campaign insights")
				continue
			}

			if len(metrics) == 0 {
				continue
			}

			// Set organization and campaign IDs, calculate derived metrics
			now := time.Now()
			for i := range metrics {
				metrics[i].OrganizationID = account.OrganizationID
				metrics[i].CampaignID = campaign.ID
				metrics[i].Platform = account.Platform
				metrics[i].LastSyncedAt = &now
				metrics[i].CalculateDerivedMetrics()
			}

			// Bulk upsert metrics
			if err := s.metricsRepo.BulkUpsertCampaignMetrics(ctx, metrics); err != nil {
				s.logger.Error().
					Err(err).
					Str("campaign_id", campaign.PlatformCampaignID).
					Msg("Failed to bulk upsert campaign metrics")
				continue
			}

			totalMetrics += len(metrics)
		}
	}

	s.logger.Info().
		Str("account_id", accountID.String()).
		Int("total_metrics", totalMetrics).
		Time("start_date", dateRange.StartDate).
		Time("end_date", dateRange.EndDate).
		Msg("Metrics sync completed")

	return totalMetrics, nil
}

// SyncAllActive synchronizes all active Meta accounts
func (s *MetaSyncService) SyncAllActive(ctx context.Context) (*service.BatchSyncResult, error) {
	startTime := time.Now()

	s.logger.Info().Msg("Starting sync for all active Meta accounts")

	result, err := s.worker.SyncAllActiveAccounts(ctx)
	if err != nil {
		return nil, err
	}

	// Convert worker result to service result
	batchResult := &service.BatchSyncResult{
		TotalAccounts: result.TotalTasks,
		SuccessCount:  result.Successful,
		PartialCount:  result.PartialFailed,
		FailureCount:  result.Failed,
		Results:       make([]service.SyncResult, 0, len(result.Results)),
		StartedAt:     startTime.Unix(),
		CompletedAt:   time.Now().Unix(),
	}

	for _, r := range result.Results {
		syncResult := service.SyncResult{
			Platform:        r.Task.Platform,
			AccountID:       r.Task.AccountID,
			CampaignsSynced: r.CampaignsSynced,
			AdSetsSynced:    r.AdSetsSynced,
			AdsSynced:       r.AdsSynced,
			MetricsSynced:   r.MetricsSynced,
		}
		if r.Error != nil {
			syncResult.Errors = []error{r.Error}
		}
		batchResult.Results = append(batchResult.Results, syncResult)
	}

	return batchResult, nil
}

// GetWorkerStats returns the worker statistics
func (s *MetaSyncService) GetWorkerStats() map[string]interface{} {
	if s.worker == nil {
		return map[string]interface{}{
			"is_running": false,
		}
	}
	return s.worker.GetStats()
}

// ScheduleSyncForAccount schedules a sync task for a specific account
func (s *MetaSyncService) ScheduleSyncForAccount(ctx context.Context, accountID uuid.UUID, options *SyncOptions) error {
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return errors.ErrNotFound("Connected account")
	}

	if account.Platform != entity.PlatformMeta {
		return errors.ErrBadRequest("Account is not a Meta account")
	}

	if options == nil {
		options = DefaultSyncOptions()
	}

	task := worker.SyncTask{
		AccountID:      account.ID,
		OrganizationID: account.OrganizationID,
		Platform:       account.Platform,
		AccessToken:    account.AccessToken,
		AdAccountID:    account.PlatformAccountID,
		DateRange:      options.DateRange,
		SyncCampaigns:  options.SyncCampaigns,
		SyncAdSets:     options.SyncAdSets,
		SyncAds:        options.SyncAds,
		SyncMetrics:    options.SyncMetrics,
	}

	return s.worker.SubmitTask(task)
}

// SyncOptions configures what to sync
type SyncOptions struct {
	DateRange     entity.DateRange
	SyncCampaigns bool
	SyncAdSets    bool
	SyncAds       bool
	SyncMetrics   bool
}

// DefaultSyncOptions returns default sync options
func DefaultSyncOptions() *SyncOptions {
	return &SyncOptions{
		DateRange: entity.DateRange{
			StartDate: time.Now().AddDate(0, 0, -7),
			EndDate:   time.Now(),
		},
		SyncCampaigns: true,
		SyncAdSets:    true,
		SyncAds:       true,
		SyncMetrics:   true,
	}
}

// Ensure MetaSyncService implements the DataSyncer interface
var _ service.DataSyncer = (*MetaSyncService)(nil)

// BatchSync implements service.DataSyncer
func (s *MetaSyncService) BatchSync(ctx context.Context, request service.BatchSyncRequest) (*service.BatchSyncResult, error) {
	// Filter for Meta platform only
	isMeta := false
	for _, p := range request.Platforms {
		if p == entity.PlatformMeta {
			isMeta = true
			break
		}
	}

	if !isMeta && len(request.Platforms) > 0 {
		return nil, errors.ErrBadRequest("This service only handles Meta platform")
	}

	return s.SyncAllActive(ctx)
}
