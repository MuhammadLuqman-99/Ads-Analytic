package oauth

import (
	"context"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/pkg/crypto"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// TokenManager handles OAuth token refresh and validation
type TokenManager struct {
	connectedAccRepo  repository.ConnectedAccountRepository
	tokenRefreshRepo  repository.TokenRefreshLogRepository
	connectorRegistry ConnectorRegistry
	encryptor         *crypto.TokenEncryptor
	logger            zerolog.Logger
}

// ConnectorRegistry provides access to platform connectors
type ConnectorRegistry interface {
	Get(platform entity.Platform) (service.PlatformConnector, bool)
}

// NewTokenManager creates a new token manager
func NewTokenManager(
	connectedAccRepo repository.ConnectedAccountRepository,
	tokenRefreshRepo repository.TokenRefreshLogRepository,
	connectorRegistry ConnectorRegistry,
	encryptionKey string,
	logger zerolog.Logger,
) (*TokenManager, error) {
	var encryptor *crypto.TokenEncryptor
	if encryptionKey != "" {
		var err error
		encryptor, err = crypto.NewTokenEncryptor(encryptionKey)
		if err != nil {
			return nil, err
		}
	}

	return &TokenManager{
		connectedAccRepo:  connectedAccRepo,
		tokenRefreshRepo:  tokenRefreshRepo,
		connectorRegistry: connectorRegistry,
		encryptor:         encryptor,
		logger:            logger.With().Str("component", "token_manager").Logger(),
	}, nil
}

// RefreshTokenIfNeeded refreshes the token if it's expiring within 30 minutes
func (m *TokenManager) RefreshTokenIfNeeded(ctx context.Context, accountID uuid.UUID) error {
	account, err := m.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return errors.ErrNotFound("Connected account")
	}

	// Check if token needs refresh
	if !account.NeedsRefresh() {
		return nil
	}

	return m.refreshToken(ctx, account)
}

// RefreshAllExpiring refreshes all tokens expiring within the specified minutes
func (m *TokenManager) RefreshAllExpiring(ctx context.Context) (int, error) {
	// Get accounts with tokens expiring in 60 minutes
	accounts, err := m.connectedAccRepo.ListExpiring(ctx, 60)
	if err != nil {
		return 0, err
	}

	refreshed := 0
	for _, account := range accounts {
		if err := m.refreshToken(ctx, &account); err != nil {
			m.logger.Error().
				Err(err).
				Str("account_id", account.ID.String()).
				Str("platform", string(account.Platform)).
				Msg("Failed to refresh token")
			continue
		}
		refreshed++
	}

	return refreshed, nil
}

// ValidateToken checks if a token is valid and refreshes if needed
func (m *TokenManager) ValidateToken(ctx context.Context, accountID uuid.UUID) (*entity.ConnectedAccount, error) {
	account, err := m.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, errors.ErrNotFound("Connected account")
	}

	// Check account status
	switch account.Status {
	case entity.AccountStatusRevoked:
		return nil, errors.NewAppError(
			errors.ErrCodeOAuthFailed,
			"TOKEN_REVOKED",
			"Access has been revoked. Please reconnect your account.",
			401,
		)
	case entity.AccountStatusExpired:
		return nil, errors.NewAppError(
			errors.ErrCodeOAuthFailed,
			"TOKEN_EXPIRED",
			"Token has expired. Please reconnect your account.",
			401,
		)
	case entity.AccountStatusInactive:
		return nil, errors.NewAppError(
			errors.ErrCodeOAuthFailed,
			"ACCOUNT_INACTIVE",
			"Account is inactive.",
			401,
		)
	}

	// Check if token is expired
	if account.IsTokenExpired() {
		// Try to refresh
		if err := m.refreshToken(ctx, account); err != nil {
			// Update status to expired
			account.Status = entity.AccountStatusExpired
			_ = m.connectedAccRepo.Update(ctx, account)

			return nil, errors.NewAppError(
				errors.ErrCodeOAuthFailed,
				"TOKEN_EXPIRED",
				"Token has expired and could not be refreshed. Please reconnect your account.",
				401,
			)
		}
		// Re-fetch after refresh
		account, _ = m.connectedAccRepo.GetByID(ctx, accountID)
	}

	// Decrypt token if encryptor is available
	if m.encryptor != nil && account.AccessToken != "" {
		decrypted, err := m.encryptor.Decrypt(account.AccessToken)
		if err != nil {
			// Token might not be encrypted (legacy), use as-is
			m.logger.Debug().
				Str("account_id", account.ID.String()).
				Msg("Token decryption failed, using raw token")
		} else {
			account.AccessToken = decrypted
		}
	}

	return account, nil
}

// refreshToken performs the actual token refresh
func (m *TokenManager) refreshToken(ctx context.Context, account *entity.ConnectedAccount) error {
	connector, ok := m.connectorRegistry.Get(account.Platform)
	if !ok {
		return errors.ErrBadRequest("Unsupported platform: " + string(account.Platform))
	}

	// For Meta, we need the current access token (not refresh token)
	// Meta uses long-lived token exchange, not traditional refresh tokens
	tokenToRefresh := account.AccessToken
	if account.RefreshToken != "" {
		tokenToRefresh = account.RefreshToken
	}

	// Decrypt if needed
	if m.encryptor != nil && tokenToRefresh != "" {
		decrypted, err := m.encryptor.Decrypt(tokenToRefresh)
		if err == nil {
			tokenToRefresh = decrypted
		}
	}

	// Record old expiry for logging
	oldExpiresAt := account.TokenExpiresAt

	// Attempt refresh
	newToken, err := connector.RefreshToken(ctx, tokenToRefresh)

	// Log the refresh attempt
	refreshLog := &entity.TokenRefreshLog{
		ConnectedAccountID: account.ID,
		OldExpiresAt:       oldExpiresAt,
	}

	if err != nil {
		refreshLog.RefreshStatus = "failed"
		refreshLog.ErrorMessage = err.Error()
		_ = m.tokenRefreshRepo.Create(ctx, refreshLog)

		m.logger.Error().
			Err(err).
			Str("account_id", account.ID.String()).
			Str("platform", string(account.Platform)).
			Msg("Token refresh failed")

		return err
	}

	// Encrypt new token if encryptor is available
	accessToken := newToken.AccessToken
	refreshToken := newToken.RefreshToken
	if m.encryptor != nil {
		if encrypted, err := m.encryptor.Encrypt(accessToken); err == nil {
			accessToken = encrypted
		}
		if refreshToken != "" {
			if encrypted, err := m.encryptor.Encrypt(refreshToken); err == nil {
				refreshToken = encrypted
			}
		}
	}

	// Update account with new token
	account.AccessToken = accessToken
	account.RefreshToken = refreshToken
	account.TokenExpiresAt = &newToken.ExpiresAt
	account.TokenScopes = newToken.Scopes
	account.Status = entity.AccountStatusActive

	if err := m.connectedAccRepo.Update(ctx, account); err != nil {
		return errors.ErrInternal("Failed to update account with new token")
	}

	// Log success
	refreshLog.RefreshStatus = "success"
	refreshLog.NewExpiresAt = &newToken.ExpiresAt
	_ = m.tokenRefreshRepo.Create(ctx, refreshLog)

	m.logger.Info().
		Str("account_id", account.ID.String()).
		Str("platform", string(account.Platform)).
		Time("new_expires_at", newToken.ExpiresAt).
		Msg("Token refreshed successfully")

	return nil
}

// EncryptToken encrypts a token for storage
func (m *TokenManager) EncryptToken(token string) (string, error) {
	if m.encryptor == nil || token == "" {
		return token, nil
	}
	return m.encryptor.Encrypt(token)
}

// DecryptToken decrypts a token from storage
func (m *TokenManager) DecryptToken(encryptedToken string) (string, error) {
	if m.encryptor == nil || encryptedToken == "" {
		return encryptedToken, nil
	}
	return m.encryptor.Decrypt(encryptedToken)
}

// GetDecryptedAccessToken retrieves and decrypts an access token for an account
func (m *TokenManager) GetDecryptedAccessToken(ctx context.Context, accountID uuid.UUID) (string, error) {
	account, err := m.ValidateToken(ctx, accountID)
	if err != nil {
		return "", err
	}
	return account.AccessToken, nil
}
