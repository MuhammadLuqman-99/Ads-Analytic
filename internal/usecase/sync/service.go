package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	domainService "github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service handles data synchronization from ad platforms
type Service struct {
	connectedAccRepo  repository.ConnectedAccountRepository
	adAccountRepo     repository.AdAccountRepository
	campaignRepo      repository.CampaignRepository
	adSetRepo         repository.AdSetRepository
	adRepo            repository.AdRepository
	metricsRepo       repository.MetricsRepository
	connectorRegistry ConnectorRegistry
	logger            zerolog.Logger
	mu                sync.Mutex
	syncing           map[uuid.UUID]bool // Track accounts being synced
}

// ConnectorRegistry provides access to platform connectors
type ConnectorRegistry interface {
	Get(platform entity.Platform) (domainService.PlatformConnector, bool)
}

// NewService creates a new sync service
func NewService(
	connectedAccRepo repository.ConnectedAccountRepository,
	adAccountRepo repository.AdAccountRepository,
	campaignRepo repository.CampaignRepository,
	adSetRepo repository.AdSetRepository,
	adRepo repository.AdRepository,
	metricsRepo repository.MetricsRepository,
	connectorRegistry ConnectorRegistry,
	logger zerolog.Logger,
) *Service {
	return &Service{
		connectedAccRepo:  connectedAccRepo,
		adAccountRepo:     adAccountRepo,
		campaignRepo:      campaignRepo,
		adSetRepo:         adSetRepo,
		adRepo:            adRepo,
		metricsRepo:       metricsRepo,
		connectorRegistry: connectorRegistry,
		logger:            logger.With().Str("service", "sync").Logger(),
		syncing:           make(map[uuid.UUID]bool),
	}
}

// SyncAccount syncs all data for a single connected account
func (s *Service) SyncAccount(ctx context.Context, accountID uuid.UUID) (*domainService.SyncResult, error) {
	// Check if already syncing
	if s.isSyncing(accountID) {
		return nil, errors.ErrConflict("Account is already being synced")
	}
	s.setSyncing(accountID, true)
	defer s.setSyncing(accountID, false)

	// Get connected account
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, errors.ErrNotFound("Connected account")
	}

	// Check account status
	if account.Status != entity.AccountStatusActive {
		return nil, errors.ErrBadRequest("Account is not active")
	}

	// Get connector
	connector, ok := s.connectorRegistry.Get(account.Platform)
	if !ok {
		return nil, errors.ErrBadRequest(fmt.Sprintf("Unsupported platform: %s", account.Platform))
	}

	result := &domainService.SyncResult{
		Platform:  account.Platform,
		AccountID: accountID,
		StartedAt: time.Now().Unix(),
		Errors:    make([]error, 0),
	}

	s.logger.Info().
		Str("account_id", accountID.String()).
		Str("platform", string(account.Platform)).
		Msg("Starting account sync")

	// Sync ad accounts
	adAccountsSynced, err := s.syncAdAccounts(ctx, account, connector)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to sync ad accounts")
		result.Errors = append(result.Errors, err)
	}

	// Get all ad accounts for this connected account
	adAccounts, err := s.adAccountRepo.ListByConnectedAccount(ctx, accountID)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list ad accounts")
		result.Errors = append(result.Errors, err)
	}

	// Sync campaigns for each ad account
	for _, adAccount := range adAccounts {
		campaignsSynced, adSetsSynced, adsSynced, syncErr := s.syncCampaignsForAccount(ctx, account, adAccount, connector)
		result.CampaignsSynced += campaignsSynced
		result.AdSetsSynced += adSetsSynced
		result.AdsSynced += adsSynced
		if syncErr != nil {
			result.Errors = append(result.Errors, syncErr)
		}
	}

	result.CampaignsSynced += adAccountsSynced
	result.CompletedAt = time.Now().Unix()

	// Update last synced timestamp
	if err := s.connectedAccRepo.UpdateLastSynced(ctx, accountID); err != nil {
		s.logger.Error().Err(err).Msg("Failed to update last synced timestamp")
	}

	s.logger.Info().
		Str("account_id", accountID.String()).
		Int("campaigns_synced", result.CampaignsSynced).
		Int("ad_sets_synced", result.AdSetsSynced).
		Int("ads_synced", result.AdsSynced).
		Int("errors", len(result.Errors)).
		Msg("Account sync completed")

	return result, nil
}

// syncAdAccounts syncs ad accounts from the platform
func (s *Service) syncAdAccounts(ctx context.Context, account *entity.ConnectedAccount, connector domainService.PlatformConnector) (int, error) {
	platformAccounts, err := connector.GetAdAccounts(ctx, account.AccessToken)
	if err != nil {
		return 0, err
	}

	synced := 0
	for _, pa := range platformAccounts {
		adAccount := &entity.AdAccount{
			BaseEntity:            entity.NewBaseEntity(),
			ConnectedAccountID:    account.ID,
			OrganizationID:        account.OrganizationID,
			Platform:              account.Platform,
			PlatformAdAccountID:   pa.ID,
			PlatformAdAccountName: pa.Name,
			Currency:              pa.Currency,
			Timezone:              pa.Timezone,
			IsActive:              pa.Status == "active",
		}

		if err := s.adAccountRepo.Upsert(ctx, adAccount); err != nil {
			s.logger.Error().Err(err).Str("ad_account_id", pa.ID).Msg("Failed to upsert ad account")
			continue
		}
		synced++
	}

	return synced, nil
}

// syncCampaignsForAccount syncs campaigns, ad sets, and ads for an ad account
func (s *Service) syncCampaignsForAccount(
	ctx context.Context,
	account *entity.ConnectedAccount,
	adAccount entity.AdAccount,
	connector domainService.PlatformConnector,
) (int, int, int, error) {
	campaignsSynced := 0
	adSetsSynced := 0
	adsSynced := 0

	// Get campaigns
	campaigns, err := connector.GetCampaigns(ctx, account.AccessToken, adAccount.PlatformAdAccountID)
	if err != nil {
		return 0, 0, 0, err
	}

	for _, campaign := range campaigns {
		campaign.AdAccountID = adAccount.ID
		campaign.OrganizationID = account.OrganizationID

		// Upsert campaign
		if err := s.campaignRepo.Upsert(ctx, &campaign); err != nil {
			s.logger.Error().Err(err).Str("campaign_id", campaign.PlatformCampaignID).Msg("Failed to upsert campaign")
			continue
		}
		campaignsSynced++

		// Sync ad sets for this campaign
		adSets, err := connector.GetAdSets(ctx, account.AccessToken, campaign.PlatformCampaignID)
		if err != nil {
			s.logger.Error().Err(err).Str("campaign_id", campaign.PlatformCampaignID).Msg("Failed to get ad sets")
			continue
		}

		for _, adSet := range adSets {
			adSet.CampaignID = campaign.ID
			adSet.OrganizationID = account.OrganizationID

			if err := s.adSetRepo.Upsert(ctx, &adSet); err != nil {
				s.logger.Error().Err(err).Str("ad_set_id", adSet.PlatformAdSetID).Msg("Failed to upsert ad set")
				continue
			}
			adSetsSynced++

			// Sync ads for this ad set
			ads, err := connector.GetAds(ctx, account.AccessToken, adSet.PlatformAdSetID)
			if err != nil {
				s.logger.Error().Err(err).Str("ad_set_id", adSet.PlatformAdSetID).Msg("Failed to get ads")
				continue
			}

			for _, ad := range ads {
				ad.AdSetID = adSet.ID
				ad.CampaignID = campaign.ID
				ad.OrganizationID = account.OrganizationID

				if err := s.adRepo.Upsert(ctx, &ad); err != nil {
					s.logger.Error().Err(err).Str("ad_id", ad.PlatformAdID).Msg("Failed to upsert ad")
					continue
				}
				adsSynced++
			}
		}
	}

	return campaignsSynced, adSetsSynced, adsSynced, nil
}

// SyncMetrics syncs metrics for an account within a date range
func (s *Service) SyncMetrics(ctx context.Context, accountID uuid.UUID, dateRange entity.DateRange) (int, error) {
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, errors.ErrNotFound("Connected account")
	}

	connector, ok := s.connectorRegistry.Get(account.Platform)
	if !ok {
		return 0, errors.ErrBadRequest(fmt.Sprintf("Unsupported platform: %s", account.Platform))
	}

	metricsSynced := 0

	// Get campaigns for this account
	campaigns, err := s.campaignRepo.ListByOrganization(ctx, account.OrganizationID, nil)
	if err != nil {
		return 0, err
	}

	for _, campaign := range campaigns {
		if campaign.Platform != account.Platform {
			continue
		}

		// Get insights for campaign
		insights, err := connector.GetCampaignInsights(ctx, account.AccessToken, campaign.PlatformCampaignID, dateRange)
		if err != nil {
			s.logger.Error().Err(err).Str("campaign_id", campaign.PlatformCampaignID).Msg("Failed to get campaign insights")
			continue
		}

		for _, insight := range insights {
			insight.CampaignID = campaign.ID
			insight.OrganizationID = account.OrganizationID

			if err := s.metricsRepo.UpsertCampaignMetrics(ctx, &insight); err != nil {
				s.logger.Error().Err(err).Str("campaign_id", campaign.PlatformCampaignID).Msg("Failed to upsert campaign metrics")
				continue
			}
			metricsSynced++
		}
	}

	return metricsSynced, nil
}

// BatchSync syncs multiple accounts
func (s *Service) BatchSync(ctx context.Context, request domainService.BatchSyncRequest) (*domainService.BatchSyncResult, error) {
	result := &domainService.BatchSyncResult{
		OrganizationID: request.OrganizationID,
		Results:        make([]domainService.SyncResult, 0),
		StartedAt:      time.Now().Unix(),
	}

	var wg sync.WaitGroup
	resultsChan := make(chan domainService.SyncResult, len(request.AccountIDs))
	semaphore := make(chan struct{}, 3) // Limit concurrent syncs

	for _, accountID := range request.AccountIDs {
		wg.Add(1)
		go func(accID uuid.UUID) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			syncResult, err := s.SyncAccount(ctx, accID)
			if err != nil {
				resultsChan <- domainService.SyncResult{
					AccountID: accID,
					Errors:    []error{err},
				}
				return
			}
			resultsChan <- *syncResult
		}(accountID)
	}

	// Wait for all syncs to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for syncResult := range resultsChan {
		result.Results = append(result.Results, syncResult)
		result.TotalAccounts++
		if syncResult.IsSuccess() {
			result.SuccessCount++
		} else if syncResult.HasPartialSuccess() {
			result.PartialCount++
		} else {
			result.FailureCount++
		}
	}

	result.CompletedAt = time.Now().Unix()
	return result, nil
}

// SyncAllActive syncs all active accounts
func (s *Service) SyncAllActive(ctx context.Context) (*domainService.BatchSyncResult, error) {
	// Get all active accounts
	accounts, err := s.connectedAccRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	accountIDs := make([]uuid.UUID, len(accounts))
	for i, acc := range accounts {
		accountIDs[i] = acc.ID
	}

	return s.BatchSync(ctx, domainService.BatchSyncRequest{
		AccountIDs:    accountIDs,
		SyncCampaigns: true,
		SyncAdSets:    true,
		SyncAds:       true,
		SyncMetrics:   true,
		DateRange:     entity.Last7Days(),
	})
}

// Helper methods

func (s *Service) isSyncing(accountID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncing[accountID]
}

func (s *Service) setSyncing(accountID uuid.UUID, syncing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if syncing {
		s.syncing[accountID] = true
	} else {
		delete(s.syncing, accountID)
	}
}
