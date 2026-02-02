package entity

import (
	"time"

	"github.com/google/uuid"
)

// TokenType represents the type of verification token
type TokenType string

const (
	TokenTypeEmailVerification TokenType = "email_verification"
	TokenTypePasswordReset     TokenType = "password_reset"
)

// VerificationToken represents an email verification or password reset token
type VerificationToken struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	Token     string     `json:"token" gorm:"size:255;unique;not null"`
	TokenType TokenType  `json:"token_type" gorm:"size:50;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	IsUsed    bool       `json:"is_used" gorm:"default:false"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name
func (VerificationToken) TableName() string {
	return "verification_tokens"
}

// NewVerificationToken creates a new verification token
func NewVerificationToken(userID uuid.UUID, token string, tokenType TokenType, expiresIn time.Duration) *VerificationToken {
	return &VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		TokenType: tokenType,
		ExpiresAt: time.Now().Add(expiresIn),
		IsUsed:    false,
		CreatedAt: time.Now(),
	}
}

// IsExpired checks if the token has expired
func (t *VerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the token is valid (not expired and not used)
func (t *VerificationToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed
}

// MarkUsed marks the token as used
func (t *VerificationToken) MarkUsed() {
	now := time.Now()
	t.UsedAt = &now
	t.IsUsed = true
}
