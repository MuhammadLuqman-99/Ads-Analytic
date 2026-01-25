package entity

import (
	"time"

	"github.com/google/uuid"
)

// ConnectedAccount represents an OAuth connection to an ad platform
type ConnectedAccount struct {
	BaseEntity
	OrganizationID      uuid.UUID     `json:"organization_id" gorm:"type:uuid;not null"`
	Platform            Platform      `json:"platform" gorm:"type:platform_type;not null"`
	PlatformAccountID   string        `json:"platform_account_id" gorm:"size:255;not null"`
	PlatformAccountName string        `json:"platform_account_name,omitempty" gorm:"size:255"`
	PlatformUserID      string        `json:"platform_user_id,omitempty" gorm:"size:255"`
	Status              AccountStatus `json:"status" gorm:"type:account_status;default:'active'"`
	LastSyncedAt        *time.Time    `json:"last_synced_at,omitempty"`
	SyncError           string        `json:"sync_error,omitempty"`
	AccountTimezone     string        `json:"account_timezone,omitempty" gorm:"size:50"`
	AccountCurrency     string        `json:"account_currency" gorm:"size:3;default:'MYR'"`
	Metadata            JSONMap       `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`

	// OAuth tokens (encrypted in storage)
	AccessToken    string     `json:"-" gorm:"type:text;not null"`
	RefreshToken   string     `json:"-" gorm:"type:text"`
	TokenType      string     `json:"-" gorm:"size:50;default:'Bearer'"`
	TokenExpiresAt *time.Time `json:"-"`
	TokenScopes    []string   `json:"-" gorm:"type:text[]"`

	// Relations
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	AdAccounts   []AdAccount   `json:"ad_accounts,omitempty" gorm:"foreignKey:ConnectedAccountID"`
}

// IsTokenExpired checks if the access token is expired
func (c *ConnectedAccount) IsTokenExpired() bool {
	if c.TokenExpiresAt == nil {
		return false
	}
	// Consider token expired 5 minutes before actual expiry
	return time.Now().Add(5 * time.Minute).After(*c.TokenExpiresAt)
}

// NeedsRefresh checks if the token needs to be refreshed
func (c *ConnectedAccount) NeedsRefresh() bool {
	if c.TokenExpiresAt == nil {
		return false
	}
	// Refresh if expires in less than 30 minutes
	return time.Now().Add(30 * time.Minute).After(*c.TokenExpiresAt)
}

// OAuthToken represents OAuth token data for token exchange/refresh
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scopes       []string  `json:"scopes,omitempty"`
}

// OAuthState represents the state for OAuth flow
type OAuthState struct {
	State          string    `json:"state"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Platform       Platform  `json:"platform"`
	RedirectURL    string    `json:"redirect_url"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// IsExpired checks if the OAuth state is expired
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// TokenRefreshLog represents a log entry for token refresh operations
type TokenRefreshLog struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ConnectedAccountID uuid.UUID  `json:"connected_account_id" gorm:"type:uuid;not null"`
	RefreshStatus      string     `json:"refresh_status" gorm:"size:20;not null"` // 'success', 'failed'
	ErrorMessage       string     `json:"error_message,omitempty"`
	OldExpiresAt       *time.Time `json:"old_expires_at,omitempty"`
	NewExpiresAt       *time.Time `json:"new_expires_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// AdAccount represents an ad/business account within a connected platform account
type AdAccount struct {
	BaseEntity
	ConnectedAccountID    uuid.UUID  `json:"connected_account_id" gorm:"type:uuid;not null"`
	OrganizationID        uuid.UUID  `json:"organization_id" gorm:"type:uuid;not null"`
	Platform              Platform   `json:"platform" gorm:"type:platform_type;not null"`
	PlatformAdAccountID   string     `json:"platform_ad_account_id" gorm:"size:255;not null"`
	PlatformAdAccountName string     `json:"platform_ad_account_name,omitempty" gorm:"size:255"`
	Currency              string     `json:"currency" gorm:"size:3;default:'MYR'"`
	Timezone              string     `json:"timezone,omitempty" gorm:"size:50"`
	IsActive              bool       `json:"is_active" gorm:"default:true"`
	LastSyncedAt          *time.Time `json:"last_synced_at,omitempty"`
	Metadata              JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`

	// Relations
	ConnectedAccount *ConnectedAccount `json:"connected_account,omitempty" gorm:"foreignKey:ConnectedAccountID"`
	Campaigns        []Campaign        `json:"campaigns,omitempty" gorm:"foreignKey:AdAccountID"`
}

// PlatformAccount represents a minimal account info returned from platforms
type PlatformAccount struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Currency string `json:"currency,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	Status   string `json:"status,omitempty"`
}

// PlatformUser represents user info from platform OAuth
type PlatformUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}
