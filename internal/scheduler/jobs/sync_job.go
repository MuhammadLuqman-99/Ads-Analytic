package jobs

import (
	"context"
	"sync"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	appErrors "github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SyncJob handles data synchronization from ad platforms
type SyncJob struct {
	connectedAccRepo  repository.ConnectedAccountRepository
	connectorRegistry ConnectorRegistry
	syncService       SyncService
	logger            zerolog.Logger
	concurrency       int
}

// ConnectorRegistry provides access to platform connectors
type ConnectorRegistry interface {
	Get(platform entity.Platform) (service.PlatformConnector, bool)
}

// SyncService interface for sync operations
type SyncService interface {
	SyncAccount(ctx context.Context, accountID uuid.UUID) (*service.SyncResult, error)
	SyncMetrics(ctx context.Context, accountID uuid.UUID, dateRange entity.DateRange) (int, error)
}

// NewSyncJob creates a new sync job
func NewSyncJob(
	connectedAccRepo repository.ConnectedAccountRepository,
	connectorRegistry ConnectorRegistry,
	syncService SyncService,
	logger zerolog.Logger,
	concurrency int,
) *SyncJob {
	if concurrency <= 0 {
		concurrency = 3
	}
	return &SyncJob{
		connectedAccRepo:  connectedAccRepo,
		connectorRegistry: connectorRegistry,
		syncService:       syncService,
		logger:            logger.With().Str("job", "sync").Logger(),
		concurrency:       concurrency,
	}
}

// SyncJobResult represents the result of a sync job
type SyncJobResult struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalAccounts int
	Successful    int
	Failed        int
	Partial       int
	Errors        []SyncJobError
}

// SyncJobError represents an error during sync
type SyncJobError struct {
	AccountID uuid.UUID
	Platform  entity.Platform
	Error     string
	Timestamp time.Time
}

// Run executes the sync job for all active accounts
func (j *SyncJob) Run(ctx context.Context) (*SyncJobResult, error) {
	result := &SyncJobResult{
		StartTime: time.Now(),
		Errors:    make([]SyncJobError, 0),
	}

	// Get all active connected accounts
	accounts, err := j.connectedAccRepo.ListActive(ctx)
	if err != nil {
		return nil, appErrors.Wrap(err, appErrors.ErrCodeInternal, "Failed to list active accounts", 500)
	}

	result.TotalAccounts = len(accounts)

	if len(accounts) == 0 {
		j.logger.Info().Msg("No active accounts to sync")
		result.EndTime = time.Now()
		return result, nil
	}

	j.logger.Info().Int("accounts", len(accounts)).Msg("Starting sync job")

	// Create work channel
	accountChan := make(chan entity.ConnectedAccount, len(accounts))
	resultChan := make(chan syncResult, len(accounts))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < j.concurrency; i++ {
		wg.Add(1)
		go j.syncWorker(ctx, &wg, accountChan, resultChan)
	}

	// Send accounts to channel
	for _, account := range accounts {
		accountChan <- account
	}
	close(accountChan)

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for r := range resultChan {
		if r.err != nil {
			result.Failed++
			result.Errors = append(result.Errors, SyncJobError{
				AccountID: r.accountID,
				Platform:  r.platform,
				Error:     r.err.Error(),
				Timestamp: time.Now(),
			})
		} else if r.partial {
			result.Partial++
		} else {
			result.Successful++
		}
	}

	result.EndTime = time.Now()

	j.logger.Info().
		Int("successful", result.Successful).
		Int("failed", result.Failed).
		Int("partial", result.Partial).
		Dur("duration", result.EndTime.Sub(result.StartTime)).
		Msg("Sync job completed")

	return result, nil
}

// syncResult internal result for a single sync
type syncResult struct {
	accountID uuid.UUID
	platform  entity.Platform
	partial   bool
	err       error
}

// syncWorker processes accounts from the channel
func (j *SyncJob) syncWorker(ctx context.Context, wg *sync.WaitGroup, accounts <-chan entity.ConnectedAccount, results chan<- syncResult) {
	defer wg.Done()

	for account := range accounts {
		result := syncResult{
			accountID: account.ID,
			platform:  account.Platform,
		}

		// Check if account token needs refresh
		if account.IsTokenExpired() {
			result.err = appErrors.NewTokenExpiredError(
				account.Platform.String(),
				"access",
				*account.TokenExpiresAt,
			)
			results <- result
			continue
		}

		// Run sync
		syncResult, err := j.syncService.SyncAccount(ctx, account.ID)
		if err != nil {
			result.err = err
		} else if !syncResult.IsSuccess() {
			if syncResult.HasPartialSuccess() {
				result.partial = true
			} else {
				result.err = appErrors.ErrInternal("Sync failed with no items synced")
			}
		}

		results <- result
	}
}

// RunForAccount executes sync for a single account
func (j *SyncJob) RunForAccount(ctx context.Context, accountID uuid.UUID) (*service.SyncResult, error) {
	account, err := j.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, appErrors.ErrNotFound("Connected account")
	}

	if account.Status != entity.AccountStatusActive {
		return nil, appErrors.ErrBadRequest("Account is not active")
	}

	if account.IsTokenExpired() {
		return nil, appErrors.NewTokenExpiredError(
			account.Platform.String(),
			"access",
			*account.TokenExpiresAt,
		)
	}

	return j.syncService.SyncAccount(ctx, accountID)
}

// RunMetricsSync syncs metrics for all accounts
func (j *SyncJob) RunMetricsSync(ctx context.Context, dateRange entity.DateRange) (*SyncJobResult, error) {
	result := &SyncJobResult{
		StartTime: time.Now(),
		Errors:    make([]SyncJobError, 0),
	}

	accounts, err := j.connectedAccRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	result.TotalAccounts = len(accounts)

	for _, account := range accounts {
		if account.IsTokenExpired() {
			result.Failed++
			result.Errors = append(result.Errors, SyncJobError{
				AccountID: account.ID,
				Platform:  account.Platform,
				Error:     "Token expired",
				Timestamp: time.Now(),
			})
			continue
		}

		_, err := j.syncService.SyncMetrics(ctx, account.ID, dateRange)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, SyncJobError{
				AccountID: account.ID,
				Platform:  account.Platform,
				Error:     err.Error(),
				Timestamp: time.Now(),
			})
			continue
		}
		result.Successful++
	}

	result.EndTime = time.Now()
	return result, nil
}

// TokenRefreshJob handles OAuth token refresh
type TokenRefreshJob struct {
	connectedAccRepo  repository.ConnectedAccountRepository
	tokenRefreshRepo  repository.TokenRefreshLogRepository
	connectorRegistry ConnectorRegistry
	logger            zerolog.Logger
}

// NewTokenRefreshJob creates a new token refresh job
func NewTokenRefreshJob(
	connectedAccRepo repository.ConnectedAccountRepository,
	tokenRefreshRepo repository.TokenRefreshLogRepository,
	connectorRegistry ConnectorRegistry,
	logger zerolog.Logger,
) *TokenRefreshJob {
	return &TokenRefreshJob{
		connectedAccRepo:  connectedAccRepo,
		tokenRefreshRepo:  tokenRefreshRepo,
		connectorRegistry: connectorRegistry,
		logger:            logger.With().Str("job", "token_refresh").Logger(),
	}
}

// TokenRefreshResult represents the result of token refresh job
type TokenRefreshResult struct {
	StartTime time.Time
	EndTime   time.Time
	Total     int
	Refreshed int
	Failed    int
	Skipped   int
}

// Run executes the token refresh job
func (j *TokenRefreshJob) Run(ctx context.Context, expiringWithinMinutes int) (*TokenRefreshResult, error) {
	result := &TokenRefreshResult{
		StartTime: time.Now(),
	}

	// Get accounts with expiring tokens
	accounts, err := j.connectedAccRepo.ListExpiring(ctx, expiringWithinMinutes)
	if err != nil {
		return nil, err
	}

	result.Total = len(accounts)

	for _, account := range accounts {
		if account.RefreshToken == "" {
			result.Skipped++
			j.logger.Warn().
				Str("account_id", account.ID.String()).
				Str("platform", string(account.Platform)).
				Msg("No refresh token available")
			continue
		}

		connector, ok := j.connectorRegistry.Get(account.Platform)
		if !ok {
			result.Skipped++
			continue
		}

		// Attempt token refresh
		newToken, err := connector.RefreshToken(ctx, account.RefreshToken)
		if err != nil {
			result.Failed++
			j.logger.Error().
				Err(err).
				Str("account_id", account.ID.String()).
				Str("platform", string(account.Platform)).
				Msg("Failed to refresh token")

			// Log the failure
			j.logRefreshAttempt(ctx, account.ID, false, err.Error(), account.TokenExpiresAt, nil)
			continue
		}

		// Update account with new tokens
		oldExpiry := account.TokenExpiresAt
		account.AccessToken = newToken.AccessToken
		if newToken.RefreshToken != "" {
			account.RefreshToken = newToken.RefreshToken
		}
		account.TokenExpiresAt = &newToken.ExpiresAt

		if err := j.connectedAccRepo.Update(ctx, &account); err != nil {
			result.Failed++
			j.logger.Error().
				Err(err).
				Str("account_id", account.ID.String()).
				Msg("Failed to update account tokens")
			continue
		}

		result.Refreshed++
		j.logRefreshAttempt(ctx, account.ID, true, "", oldExpiry, &newToken.ExpiresAt)

		j.logger.Info().
			Str("account_id", account.ID.String()).
			Str("platform", string(account.Platform)).
			Time("new_expiry", newToken.ExpiresAt).
			Msg("Token refreshed successfully")
	}

	result.EndTime = time.Now()
	return result, nil
}

// logRefreshAttempt logs a token refresh attempt
func (j *TokenRefreshJob) logRefreshAttempt(ctx context.Context, accountID uuid.UUID, success bool, errorMsg string, oldExpiry, newExpiry *time.Time) {
	status := "success"
	if !success {
		status = "failed"
	}

	log := &entity.TokenRefreshLog{
		ID:                 uuid.New(),
		ConnectedAccountID: accountID,
		RefreshStatus:      status,
		ErrorMessage:       errorMsg,
		OldExpiresAt:       oldExpiry,
		NewExpiresAt:       newExpiry,
		CreatedAt:          time.Now(),
	}

	if err := j.tokenRefreshRepo.Create(ctx, log); err != nil {
		j.logger.Error().Err(err).Msg("Failed to log token refresh attempt")
	}
}

// RefreshTokenIfNeeded refreshes a token if it's about to expire
func (j *TokenRefreshJob) RefreshTokenIfNeeded(ctx context.Context, accountID uuid.UUID) error {
	account, err := j.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}

	if !account.NeedsRefresh() {
		return nil
	}

	connector, ok := j.connectorRegistry.Get(account.Platform)
	if !ok {
		return appErrors.ErrBadRequest("Unsupported platform: " + string(account.Platform))
	}

	newToken, err := connector.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return err
	}

	account.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		account.RefreshToken = newToken.RefreshToken
	}
	account.TokenExpiresAt = &newToken.ExpiresAt

	return j.connectedAccRepo.Update(ctx, account)
}

// RefreshAllExpiring refreshes all expiring tokens
func (j *TokenRefreshJob) RefreshAllExpiring(ctx context.Context) (int, error) {
	result, err := j.Run(ctx, 30) // Refresh tokens expiring in 30 minutes
	if err != nil {
		return 0, err
	}
	return result.Refreshed, nil
}
