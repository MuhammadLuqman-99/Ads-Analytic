package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockConnectedAccountRepository struct {
	accounts map[uuid.UUID]*entity.ConnectedAccount
	expiring []entity.ConnectedAccount
}

func newMockConnectedAccountRepo() *mockConnectedAccountRepository {
	return &mockConnectedAccountRepository{
		accounts: make(map[uuid.UUID]*entity.ConnectedAccount),
		expiring: make([]entity.ConnectedAccount, 0),
	}
}

func (m *mockConnectedAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ConnectedAccount, error) {
	if account, ok := m.accounts[id]; ok {
		// Return a copy to avoid mutation
		copy := *account
		return &copy, nil
	}
	return nil, errors.ErrNotFound("Connected account")
}

func (m *mockConnectedAccountRepository) Update(ctx context.Context, account *entity.ConnectedAccount) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockConnectedAccountRepository) ListExpiring(ctx context.Context, withinMinutes int) ([]entity.ConnectedAccount, error) {
	return m.expiring, nil
}

func (m *mockConnectedAccountRepository) Create(ctx context.Context, account *entity.ConnectedAccount) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockConnectedAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.accounts, id)
	return nil
}

func (m *mockConnectedAccountRepository) GetByPlatformAccountID(ctx context.Context, orgID uuid.UUID, platform entity.Platform, platformAccountID string) (*entity.ConnectedAccount, error) {
	return nil, errors.ErrNotFound("Connected account")
}

func (m *mockConnectedAccountRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepository) ListByPlatform(ctx context.Context, orgID uuid.UUID, platform entity.Platform) ([]entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepository) ListActive(ctx context.Context) ([]entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepository) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *interface{}) error {
	return nil
}

func (m *mockConnectedAccountRepository) UpdateSyncStatus(ctx context.Context, id uuid.UUID, status entity.AccountStatus, syncError string) error {
	return nil
}

func (m *mockConnectedAccountRepository) UpdateLastSynced(ctx context.Context, id uuid.UUID) error {
	return nil
}

// Add account helper
func (m *mockConnectedAccountRepository) addAccount(account *entity.ConnectedAccount) {
	m.accounts[account.ID] = account
}

func (m *mockConnectedAccountRepository) setExpiring(accounts []entity.ConnectedAccount) {
	m.expiring = accounts
}

type mockTokenRefreshLogRepository struct {
	logs []*entity.TokenRefreshLog
}

func newMockTokenRefreshLogRepo() *mockTokenRefreshLogRepository {
	return &mockTokenRefreshLogRepository{
		logs: make([]*entity.TokenRefreshLog, 0),
	}
}

func (m *mockTokenRefreshLogRepository) Create(ctx context.Context, log *entity.TokenRefreshLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockTokenRefreshLogRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]entity.TokenRefreshLog, error) {
	return nil, nil
}

func (m *mockTokenRefreshLogRepository) DeleteOld(ctx context.Context, olderThanDays int) error {
	return nil
}

type mockPlatformConnector struct {
	refreshErr   error
	refreshToken *entity.OAuthToken
}

func newMockPlatformConnector() *mockPlatformConnector {
	return &mockPlatformConnector{
		refreshToken: &entity.OAuthToken{
			AccessToken: "new_access_token",
			TokenType:   "Bearer",
			ExpiresIn:   5184000, // 60 days
			ExpiresAt:   time.Now().Add(60 * 24 * time.Hour),
		},
	}
}

func (m *mockPlatformConnector) GetAuthURL(state string) string {
	return "https://facebook.com/oauth?state=" + state
}

func (m *mockPlatformConnector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	return m.refreshToken, nil
}

func (m *mockPlatformConnector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshToken, nil
}

func (m *mockPlatformConnector) RevokeToken(ctx context.Context, accessToken string) error {
	return nil
}

func (m *mockPlatformConnector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	return &entity.PlatformUser{
		ID:   "12345",
		Name: "Test User",
	}, nil
}

func (m *mockPlatformConnector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	return nil, nil
}

func (m *mockPlatformConnector) HealthCheck(ctx context.Context) error {
	return nil
}

type mockConnectorRegistry struct {
	connectors map[entity.Platform]*mockPlatformConnector
}

func newMockConnectorRegistry() *mockConnectorRegistry {
	return &mockConnectorRegistry{
		connectors: map[entity.Platform]*mockPlatformConnector{
			entity.PlatformMeta: newMockPlatformConnector(),
		},
	}
}

func (m *mockConnectorRegistry) Get(platform entity.Platform) (interface{}, bool) {
	if connector, ok := m.connectors[platform]; ok {
		return connector, true
	}
	return nil, false
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTestAccount(id uuid.UUID, status entity.AccountStatus, expiresAt *time.Time) *entity.ConnectedAccount {
	return &entity.ConnectedAccount{
		BaseEntity: entity.BaseEntity{
			ID: id,
		},
		OrganizationID: uuid.New(),
		Platform:       entity.PlatformMeta,
		AccessToken:    "test_access_token",
		RefreshToken:   "test_refresh_token",
		Status:         status,
		TokenExpiresAt: expiresAt,
	}
}

// ============================================================================
// Tests
// ============================================================================

func TestTokenManager_ValidateToken_ActiveAccount(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	// Don't use encryption for this test
	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create active account with non-expired token
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusActive, &expiresAt)
	accountRepo.addAccount(account)

	// Validate token
	result, err := tm.ValidateToken(context.Background(), accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != accountID {
		t.Errorf("returned wrong account ID: got %v, want %v", result.ID, accountID)
	}
}

func TestTokenManager_ValidateToken_RevokedAccount(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create revoked account
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusRevoked, &expiresAt)
	accountRepo.addAccount(account)

	// Validate token should fail
	_, err = tm.ValidateToken(context.Background(), accountID)
	if err == nil {
		t.Fatal("expected error for revoked account")
	}

	appErr, ok := err.(*errors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got: %T", err)
	}

	if appErr.Code != errors.ErrCodeOAuthFailed {
		t.Errorf("expected ErrCodeOAuthFailed, got: %v", appErr.Code)
	}
}

func TestTokenManager_ValidateToken_ExpiredAccount(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create expired account
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusExpired, &expiresAt)
	accountRepo.addAccount(account)

	// Validate token should fail
	_, err = tm.ValidateToken(context.Background(), accountID)
	if err == nil {
		t.Fatal("expected error for expired account")
	}
}

func TestTokenManager_ValidateToken_InactiveAccount(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create inactive account
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusInactive, &expiresAt)
	accountRepo.addAccount(account)

	// Validate token should fail
	_, err = tm.ValidateToken(context.Background(), accountID)
	if err == nil {
		t.Fatal("expected error for inactive account")
	}
}

func TestTokenManager_ValidateToken_NotFound(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Try to validate non-existent account
	_, err = tm.ValidateToken(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestTokenManager_EncryptDecryptToken(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	encryptionKey := "12345678901234567890123456789012"
	tm, err := NewTokenManager(accountRepo, logRepo, nil, encryptionKey, logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	originalToken := "test_access_token_12345"

	// Encrypt
	encrypted, err := tm.EncryptToken(originalToken)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if encrypted == originalToken {
		t.Error("encrypted token should be different from original")
	}

	// Decrypt
	decrypted, err := tm.DecryptToken(encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != originalToken {
		t.Errorf("decrypted token doesn't match: got %q, want %q", decrypted, originalToken)
	}
}

func TestTokenManager_EncryptToken_NoEncryptor(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	// Create without encryption key
	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	originalToken := "test_access_token"

	// Should return token as-is
	result, err := tm.EncryptToken(originalToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != originalToken {
		t.Errorf("expected token to be unchanged: got %q, want %q", result, originalToken)
	}
}

func TestTokenManager_RefreshTokenIfNeeded_NotNeeded(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create account with token expiring in 2 hours (not within 30 min window)
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusActive, &expiresAt)
	accountRepo.addAccount(account)

	// Should not refresh
	err = tm.RefreshTokenIfNeeded(context.Background(), accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTokenManager_RefreshTokenIfNeeded_NotFound(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Try to refresh non-existent account
	err = tm.RefreshTokenIfNeeded(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for non-existent account")
	}
}

func TestTokenManager_GetDecryptedAccessToken(t *testing.T) {
	accountRepo := newMockConnectedAccountRepo()
	logRepo := newMockTokenRefreshLogRepo()
	logger := zerolog.Nop()

	tm, err := NewTokenManager(accountRepo, logRepo, nil, "", logger)
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	// Create active account
	accountID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Hour)
	account := createTestAccount(accountID, entity.AccountStatusActive, &expiresAt)
	accountRepo.addAccount(account)

	// Should return the access token
	token, err := tm.GetDecryptedAccessToken(context.Background(), accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "test_access_token" {
		t.Errorf("unexpected token: got %q, want %q", token, "test_access_token")
	}
}

// ============================================================================
// Entity Tests
// ============================================================================

func TestConnectedAccount_IsTokenExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{
			name:      "nil expiry",
			expiresAt: nil,
			expected:  false,
		},
		{
			name:      "expired",
			expiresAt: timePtr(time.Now().Add(-1 * time.Hour)),
			expected:  true,
		},
		{
			name:      "expiring soon (within 5 min buffer)",
			expiresAt: timePtr(time.Now().Add(3 * time.Minute)),
			expected:  true,
		},
		{
			name:      "not expired",
			expiresAt: timePtr(time.Now().Add(2 * time.Hour)),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &entity.ConnectedAccount{
				TokenExpiresAt: tt.expiresAt,
			}

			if got := account.IsTokenExpired(); got != tt.expected {
				t.Errorf("IsTokenExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConnectedAccount_NeedsRefresh(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt *time.Time
		expected  bool
	}{
		{
			name:      "nil expiry",
			expiresAt: nil,
			expected:  false,
		},
		{
			name:      "needs refresh (within 30 min)",
			expiresAt: timePtr(time.Now().Add(20 * time.Minute)),
			expected:  true,
		},
		{
			name:      "does not need refresh",
			expiresAt: timePtr(time.Now().Add(2 * time.Hour)),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &entity.ConnectedAccount{
				TokenExpiresAt: tt.expiresAt,
			}

			if got := account.NeedsRefresh(); got != tt.expected {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOAuthState_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired",
			expiresAt: time.Now().Add(-1 * time.Minute),
			expected:  true,
		},
		{
			name:      "not expired",
			expiresAt: time.Now().Add(5 * time.Minute),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &entity.OAuthState{
				ExpiresAt: tt.expiresAt,
			}

			if got := state.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
