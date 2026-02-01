package auth

import (
	"context"
	"testing"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/google/uuid"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockUserRepository struct {
	users map[uuid.UUID]*entity.User
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[uuid.UUID]*entity.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.ErrNotFound("User")
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.ErrNotFound("User")
}

func (m *mockUserRepository) Update(ctx context.Context, user *entity.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockOrgRepository struct {
	orgs map[uuid.UUID]*entity.Organization
}

func newMockOrgRepo() *mockOrgRepository {
	return &mockOrgRepository{
		orgs: make(map[uuid.UUID]*entity.Organization),
	}
}

func (m *mockOrgRepository) Create(ctx context.Context, org *entity.Organization) error {
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Organization, error) {
	if org, ok := m.orgs[id]; ok {
		return org, nil
	}
	return nil, errors.ErrNotFound("Organization")
}

func (m *mockOrgRepository) GetBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	return nil, errors.ErrNotFound("Organization")
}

func (m *mockOrgRepository) Update(ctx context.Context, org *entity.Organization) error {
	m.orgs[org.ID] = org
	return nil
}

func (m *mockOrgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.orgs, id)
	return nil
}

func (m *mockOrgRepository) List(ctx context.Context, pagination *entity.Pagination) ([]entity.Organization, error) {
	return nil, nil
}

func (m *mockOrgRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Organization, error) {
	return nil, nil
}

type mockOrgMemberRepository struct {
	members map[uuid.UUID]*entity.OrganizationMember
}

func newMockOrgMemberRepo() *mockOrgMemberRepository {
	return &mockOrgMemberRepository{
		members: make(map[uuid.UUID]*entity.OrganizationMember),
	}
}

func (m *mockOrgMemberRepository) Create(ctx context.Context, member *entity.OrganizationMember) error {
	m.members[member.ID] = member
	return nil
}

func (m *mockOrgMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.OrganizationMember, error) {
	if member, ok := m.members[id]; ok {
		return member, nil
	}
	return nil, errors.ErrNotFound("Member")
}

func (m *mockOrgMemberRepository) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*entity.OrganizationMember, error) {
	for _, member := range m.members {
		if member.OrganizationID == orgID && member.UserID == userID {
			return member, nil
		}
	}
	return nil, errors.ErrNotFound("Member")
}

func (m *mockOrgMemberRepository) Update(ctx context.Context, member *entity.OrganizationMember) error {
	m.members[member.ID] = member
	return nil
}

func (m *mockOrgMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.members, id)
	return nil
}

func (m *mockOrgMemberRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, pagination *entity.Pagination) ([]entity.OrganizationMember, error) {
	return nil, nil
}

func (m *mockOrgMemberRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]entity.OrganizationMember, error) {
	var result []entity.OrganizationMember
	for _, member := range m.members {
		if member.UserID == userID {
			result = append(result, *member)
		}
	}
	return result, nil
}

func (m *mockOrgMemberRepository) UpdateRole(ctx context.Context, id uuid.UUID, role entity.UserRole) error {
	return nil
}

type mockConnectedAccountRepository struct {
	accounts map[uuid.UUID]*entity.ConnectedAccount
}

func newMockConnectedAccountRepo() *mockConnectedAccountRepository {
	return &mockConnectedAccountRepository{
		accounts: make(map[uuid.UUID]*entity.ConnectedAccount),
	}
}

func (m *mockConnectedAccountRepository) Create(ctx context.Context, account *entity.ConnectedAccount) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockConnectedAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ConnectedAccount, error) {
	if account, ok := m.accounts[id]; ok {
		return account, nil
	}
	return nil, errors.ErrNotFound("Connected account")
}

func (m *mockConnectedAccountRepository) GetByPlatformAccountID(ctx context.Context, orgID uuid.UUID, platform entity.Platform, platformAccountID string) (*entity.ConnectedAccount, error) {
	for _, acc := range m.accounts {
		if acc.OrganizationID == orgID && acc.Platform == platform && acc.PlatformAccountID == platformAccountID {
			return acc, nil
		}
	}
	return nil, errors.ErrNotFound("Connected account")
}

func (m *mockConnectedAccountRepository) Update(ctx context.Context, account *entity.ConnectedAccount) error {
	m.accounts[account.ID] = account
	return nil
}

func (m *mockConnectedAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.accounts, id)
	return nil
}

func (m *mockConnectedAccountRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error) {
	var result []entity.ConnectedAccount
	for _, acc := range m.accounts {
		if acc.OrganizationID == orgID {
			result = append(result, *acc)
		}
	}
	return result, nil
}

func (m *mockConnectedAccountRepository) ListByPlatform(ctx context.Context, orgID uuid.UUID, platform entity.Platform) ([]entity.ConnectedAccount, error) {
	return nil, nil
}

func (m *mockConnectedAccountRepository) ListExpiring(ctx context.Context, withinMinutes int) ([]entity.ConnectedAccount, error) {
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

type mockStateStore struct {
	states map[string]*entity.OAuthState
}

func newMockStateStore() *mockStateStore {
	return &mockStateStore{
		states: make(map[string]*entity.OAuthState),
	}
}

func (m *mockStateStore) Save(ctx context.Context, state *entity.OAuthState) error {
	m.states[state.State] = state
	return nil
}

func (m *mockStateStore) Get(ctx context.Context, stateID string) (*entity.OAuthState, error) {
	if state, ok := m.states[stateID]; ok {
		return state, nil
	}
	return nil, errors.ErrNotFound("OAuth state")
}

func (m *mockStateStore) Delete(ctx context.Context, stateID string) error {
	delete(m.states, stateID)
	return nil
}

type mockConnectorRegistry struct{}

func newMockConnectorRegistry() *mockConnectorRegistry {
	return &mockConnectorRegistry{}
}

type mockPlatformConnector struct {
	authURL string
}

func (c *mockPlatformConnector) GetAuthURL(state string) string {
	return "https://facebook.com/oauth?state=" + state
}

func (c *mockPlatformConnector) ExchangeCode(ctx context.Context, code string) (*entity.OAuthToken, error) {
	return &entity.OAuthToken{
		AccessToken: "access_token_123",
		TokenType:   "Bearer",
		ExpiresIn:   5184000,
		ExpiresAt:   time.Now().Add(60 * 24 * time.Hour),
	}, nil
}

func (c *mockPlatformConnector) RefreshToken(ctx context.Context, refreshToken string) (*entity.OAuthToken, error) {
	return nil, nil
}

func (c *mockPlatformConnector) RevokeToken(ctx context.Context, accessToken string) error {
	return nil
}

func (c *mockPlatformConnector) GetUserInfo(ctx context.Context, accessToken string) (*entity.PlatformUser, error) {
	return &entity.PlatformUser{
		ID:   "fb_user_123",
		Name: "Test User",
	}, nil
}

func (c *mockPlatformConnector) GetAdAccounts(ctx context.Context, accessToken string) ([]entity.PlatformAccount, error) {
	return nil, nil
}

func (c *mockPlatformConnector) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockConnectorRegistry) Get(platform entity.Platform) (interface{}, bool) {
	if platform == entity.PlatformMeta {
		return &mockPlatformConnector{}, true
	}
	return nil, false
}

// ============================================================================
// Test Helper Functions
// ============================================================================

func TestGenerateState(t *testing.T) {
	// Test state generation produces unique values
	states := make(map[string]bool)

	for i := 0; i < 100; i++ {
		state, err := generateState()
		if err != nil {
			t.Fatalf("generateState() failed: %v", err)
		}

		if state == "" {
			t.Error("generated state should not be empty")
		}

		if states[state] {
			t.Errorf("duplicate state generated at iteration %d", i)
		}
		states[state] = true
	}
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(string) bool
	}{
		{
			name:  "simple name",
			input: "My Company",
			validate: func(s string) bool {
				return len(s) > 0 && !hasUpperCase(s)
			},
		},
		{
			name:  "special characters",
			input: "Company!@#$%",
			validate: func(s string) bool {
				return len(s) > 0
			},
		},
		{
			name:  "unicode",
			input: "マレーシア Company",
			validate: func(s string) bool {
				return len(s) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.input)
			if !tt.validate(result) {
				t.Errorf("generateSlug(%q) = %q, validation failed", tt.input, result)
			}
		})
	}
}

func hasUpperCase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

// ============================================================================
// OAuth State Tests
// ============================================================================

func TestOAuthState_Save_Get_Delete(t *testing.T) {
	store := newMockStateStore()
	ctx := context.Background()

	state := &entity.OAuthState{
		State:          "test_state_123",
		OrganizationID: uuid.New(),
		UserID:         uuid.New(),
		Platform:       entity.PlatformMeta,
		RedirectURL:    "/settings/connections",
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	// Save
	if err := store.Save(ctx, state); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Get
	retrieved, err := store.Get(ctx, state.State)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.State != state.State {
		t.Errorf("State mismatch: got %q, want %q", retrieved.State, state.State)
	}

	if retrieved.Platform != state.Platform {
		t.Errorf("Platform mismatch: got %q, want %q", retrieved.Platform, state.Platform)
	}

	// Delete
	if err := store.Delete(ctx, state.State); err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get(ctx, state.State)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestOAuthState_GetNonExistent(t *testing.T) {
	store := newMockStateStore()
	ctx := context.Background()

	_, err := store.Get(ctx, "non_existent_state")
	if err == nil {
		t.Error("expected error for non-existent state")
	}
}

// ============================================================================
// Service Tests
// ============================================================================

func TestGetConnectedAccounts(t *testing.T) {
	connectedAccRepo := newMockConnectedAccountRepo()

	orgID := uuid.New()

	// Add some accounts
	for i := 0; i < 3; i++ {
		account := &entity.ConnectedAccount{
			BaseEntity:     entity.BaseEntity{ID: uuid.New()},
			OrganizationID: orgID,
			Platform:       entity.PlatformMeta,
			Status:         entity.AccountStatusActive,
		}
		connectedAccRepo.Create(context.Background(), account)
	}

	// Different org
	otherOrgAccount := &entity.ConnectedAccount{
		BaseEntity:     entity.BaseEntity{ID: uuid.New()},
		OrganizationID: uuid.New(),
		Platform:       entity.PlatformMeta,
		Status:         entity.AccountStatusActive,
	}
	connectedAccRepo.Create(context.Background(), otherOrgAccount)

	// Get accounts
	accounts, err := connectedAccRepo.ListByOrganization(context.Background(), orgID)
	if err != nil {
		t.Fatalf("ListByOrganization() failed: %v", err)
	}

	if len(accounts) != 3 {
		t.Errorf("expected 3 accounts, got %d", len(accounts))
	}
}

func TestDisconnectPlatform(t *testing.T) {
	connectedAccRepo := newMockConnectedAccountRepo()

	accountID := uuid.New()
	account := &entity.ConnectedAccount{
		BaseEntity:     entity.BaseEntity{ID: accountID},
		OrganizationID: uuid.New(),
		Platform:       entity.PlatformMeta,
		AccessToken:    "test_token",
		Status:         entity.AccountStatusActive,
	}
	connectedAccRepo.Create(context.Background(), account)

	// Disconnect (simulated)
	retrievedAcc, err := connectedAccRepo.GetByID(context.Background(), accountID)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	retrievedAcc.Status = entity.AccountStatusRevoked
	if err := connectedAccRepo.Update(context.Background(), retrievedAcc); err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify
	updated, _ := connectedAccRepo.GetByID(context.Background(), accountID)
	if updated.Status != entity.AccountStatusRevoked {
		t.Errorf("expected status %q, got %q", entity.AccountStatusRevoked, updated.Status)
	}
}

// ============================================================================
// Role Permissions Tests
// ============================================================================

func TestGetPermissionsForRole(t *testing.T) {
	s := &Service{}

	tests := []struct {
		name        string
		role        entity.UserRole
		minExpected int
		containsAll bool
	}{
		{
			name:        "owner has all permissions",
			role:        entity.RoleOwner,
			minExpected: 1,
			containsAll: true,
		},
		{
			name:        "admin has multiple permissions",
			role:        entity.RoleAdmin,
			minExpected: 5,
			containsAll: false,
		},
		{
			name:        "analyst has read permissions",
			role:        entity.RoleAnalyst,
			minExpected: 3,
			containsAll: false,
		},
		{
			name:        "viewer has limited permissions",
			role:        entity.RoleViewer,
			minExpected: 2,
			containsAll: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := s.getPermissionsForRole(tt.role)

			if len(perms) < tt.minExpected {
				t.Errorf("expected at least %d permissions, got %d", tt.minExpected, len(perms))
			}

			if tt.containsAll {
				hasAll := false
				for _, p := range perms {
					if p == "*" {
						hasAll = true
						break
					}
				}
				if !hasAll {
					t.Error("owner should have '*' permission")
				}
			}
		})
	}
}
