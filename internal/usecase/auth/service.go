package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ads-aggregator/ads-aggregator/internal/domain/entity"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/repository"
	"github.com/ads-aggregator/ads-aggregator/internal/domain/service"
	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication and authorization logic
type Service struct {
	userRepo              repository.UserRepository
	orgRepo               repository.OrganizationRepository
	orgMemberRepo         repository.OrganizationMemberRepository
	connectedAccRepo      repository.ConnectedAccountRepository
	verificationTokenRepo repository.VerificationTokenRepository
	jwtManager            *jwt.Manager
	connectorRegistry     ConnectorRegistry
	stateStore            StateStore
	emailSender           EmailSender
	emailConfig           EmailConfig
}

// EmailSender interface for sending emails
type EmailSender interface {
	SendHTML(ctx context.Context, to, toName, subject, htmlBody string) error
}

// EmailConfig holds email-related configuration
type EmailConfig struct {
	FrontendURL             string
	EmailVerificationExpiry time.Duration
	PasswordResetExpiry     time.Duration
	AppName                 string
	SupportEmail            string
}

// StateStore stores OAuth state for verification
type StateStore interface {
	Save(ctx context.Context, state *entity.OAuthState) error
	Get(ctx context.Context, stateID string) (*entity.OAuthState, error)
	Delete(ctx context.Context, stateID string) error
}

// ConnectorRegistry provides access to platform connectors
type ConnectorRegistry interface {
	Get(platform entity.Platform) (service.PlatformConnector, bool)
}

// NewService creates a new auth service
func NewService(
	userRepo repository.UserRepository,
	orgRepo repository.OrganizationRepository,
	orgMemberRepo repository.OrganizationMemberRepository,
	connectedAccRepo repository.ConnectedAccountRepository,
	verificationTokenRepo repository.VerificationTokenRepository,
	jwtManager *jwt.Manager,
	connectorRegistry ConnectorRegistry,
	stateStore StateStore,
	emailSender EmailSender,
	emailConfig EmailConfig,
) *Service {
	return &Service{
		userRepo:              userRepo,
		orgRepo:               orgRepo,
		orgMemberRepo:         orgMemberRepo,
		connectedAccRepo:      connectedAccRepo,
		verificationTokenRepo: verificationTokenRepo,
		jwtManager:            jwtManager,
		connectorRegistry:     connectorRegistry,
		stateStore:            stateStore,
		emailSender:           emailSender,
		emailConfig:           emailConfig,
	}
}

// ============================================================================
// User Authentication
// ============================================================================

// RegisterRequest represents user registration input
type RegisterRequest struct {
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	FirstName        string `json:"first_name" binding:"required"`
	LastName         string `json:"last_name"`
	OrganizationName string `json:"organization_name" binding:"required"`
}

// LoginRequest represents user login input
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *entity.User         `json:"user"`
	Organization *entity.Organization `json:"organization,omitempty"`
	Tokens       *jwt.TokenPair       `json:"tokens"`
}

// Register creates a new user and organization
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.ErrConflict("Email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.ErrInternal("Failed to hash password")
	}

	// Create user
	user := &entity.User{
		BaseEntity:   entity.NewBaseEntity(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to create user", 500)
	}

	// Create organization
	org := &entity.Organization{
		BaseEntity:       entity.NewBaseEntity(),
		Name:             req.OrganizationName,
		Slug:             generateSlug(req.OrganizationName),
		SubscriptionPlan: entity.PlanFree,
		IsActive:         true,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to create organization", 500)
	}

	// Create organization membership (owner)
	now := time.Now()
	member := &entity.OrganizationMember{
		BaseEntity:     entity.NewBaseEntity(),
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           entity.RoleOwner,
		JoinedAt:       &now,
		IsActive:       true,
	}

	if err := s.orgMemberRepo.Create(ctx, member); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to create membership", 500)
	}

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		org.ID.String(),
		user.Email,
		string(entity.RoleOwner),
		[]string{"*"}, // Full permissions for owner
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to generate tokens", 500)
	}

	return &AuthResponse{
		User:         user,
		Organization: org,
		Tokens:       tokens,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, errors.ErrUnauthorized("Invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.ErrUnauthorized("Account is disabled")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.ErrUnauthorized("Invalid email or password")
	}

	// Get user's organizations
	memberships, err := s.orgMemberRepo.ListByUser(ctx, user.ID)
	if err != nil || len(memberships) == 0 {
		return nil, errors.ErrInternal("No organization found for user")
	}

	// Use first organization (or could let user choose)
	membership := memberships[0]
	org, err := s.orgRepo.GetByID(ctx, membership.OrganizationID)
	if err != nil {
		return nil, errors.ErrInternal("Failed to load organization")
	}

	// Update last login
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	tokens, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		org.ID.String(),
		user.Email,
		string(membership.Role),
		s.getPermissionsForRole(membership.Role),
	)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "Failed to generate tokens", 500)
	}

	return &AuthResponse{
		User:         user,
		Organization: org,
		Tokens:       tokens,
	}, nil
}

// RefreshToken refreshes the access token using refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	tokens, err := s.jwtManager.RefreshTokenPair(refreshToken)
	if err != nil {
		return nil, errors.ErrUnauthorized("Invalid refresh token")
	}
	return tokens, nil
}

// ValidateToken validates an access token and returns the claims
func (s *Service) ValidateToken(ctx context.Context, accessToken string) (*jwt.Claims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(accessToken)
	if err != nil {
		if jwt.IsTokenExpired(err) {
			return nil, errors.ErrUnauthorized("Token has expired")
		}
		return nil, errors.ErrUnauthorized("Invalid token")
	}
	return claims, nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrNotFound("User")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.ErrUnauthorized("Invalid current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrInternal("Failed to hash password")
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// ============================================================================
// Email Verification
// ============================================================================

// RequestEmailVerification sends an email verification link to the user
func (s *Service) RequestEmailVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrNotFound("User")
	}

	// Check if already verified
	if user.EmailVerifiedAt != nil {
		return errors.ErrBadRequest("Email already verified")
	}

	// Delete any existing tokens for this user
	_ = s.verificationTokenRepo.DeleteByUserAndType(ctx, userID, entity.TokenTypeEmailVerification)

	// Generate token
	tokenStr, err := generateSecureToken()
	if err != nil {
		return errors.ErrInternal("Failed to generate verification token")
	}

	// Create verification token
	token := entity.NewVerificationToken(
		userID,
		tokenStr,
		entity.TokenTypeEmailVerification,
		s.emailConfig.EmailVerificationExpiry,
	)

	if err := s.verificationTokenRepo.Create(ctx, token); err != nil {
		return errors.ErrInternal("Failed to create verification token")
	}

	// Send verification email
	if s.emailSender != nil {
		verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.emailConfig.FrontendURL, tokenStr)
		htmlBody := s.buildEmailVerificationHTML(user.FullName(), verificationURL)

		if err := s.emailSender.SendHTML(ctx, user.Email, user.FullName(), "Verify Your Email", htmlBody); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to send verification email: %v\n", err)
		}
	}

	return nil
}

// VerifyEmail verifies a user's email with the provided token
func (s *Service) VerifyEmail(ctx context.Context, tokenStr string) error {
	token, err := s.verificationTokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return errors.ErrBadRequest("Invalid verification token")
	}

	// Check if token is valid
	if !token.IsValid() {
		if token.IsUsed {
			return errors.ErrBadRequest("Token has already been used")
		}
		return errors.ErrBadRequest("Token has expired")
	}

	// Verify token type
	if token.TokenType != entity.TokenTypeEmailVerification {
		return errors.ErrBadRequest("Invalid token type")
	}

	// Mark token as used
	if err := s.verificationTokenRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return errors.ErrInternal("Failed to mark token as used")
	}

	// Verify user's email
	if err := s.userRepo.VerifyEmail(ctx, token.UserID); err != nil {
		return errors.ErrInternal("Failed to verify email")
	}

	return nil
}

// ============================================================================
// Password Reset
// ============================================================================

// RequestPasswordReset sends a password reset email to the user
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Delete any existing password reset tokens for this user
	_ = s.verificationTokenRepo.DeleteByUserAndType(ctx, user.ID, entity.TokenTypePasswordReset)

	// Generate token
	tokenStr, err := generateSecureToken()
	if err != nil {
		return errors.ErrInternal("Failed to generate reset token")
	}

	// Create password reset token
	token := entity.NewVerificationToken(
		user.ID,
		tokenStr,
		entity.TokenTypePasswordReset,
		s.emailConfig.PasswordResetExpiry,
	)

	if err := s.verificationTokenRepo.Create(ctx, token); err != nil {
		return errors.ErrInternal("Failed to create reset token")
	}

	// Send password reset email
	if s.emailSender != nil {
		resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.emailConfig.FrontendURL, tokenStr)
		htmlBody := s.buildPasswordResetHTML(user.FullName(), resetURL)

		if err := s.emailSender.SendHTML(ctx, user.Email, user.FullName(), "Reset Your Password", htmlBody); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to send password reset email: %v\n", err)
		}
	}

	return nil
}

// ResetPassword resets a user's password using the provided token
func (s *Service) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	token, err := s.verificationTokenRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		return errors.ErrBadRequest("Invalid reset token")
	}

	// Check if token is valid
	if !token.IsValid() {
		if token.IsUsed {
			return errors.ErrBadRequest("Token has already been used")
		}
		return errors.ErrBadRequest("Token has expired")
	}

	// Verify token type
	if token.TokenType != entity.TokenTypePasswordReset {
		return errors.ErrBadRequest("Invalid token type")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.ErrInternal("Failed to hash password")
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, token.UserID, string(hashedPassword)); err != nil {
		return errors.ErrInternal("Failed to update password")
	}

	// Mark token as used
	if err := s.verificationTokenRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return errors.ErrInternal("Failed to mark token as used")
	}

	// Delete all password reset tokens for this user
	_ = s.verificationTokenRepo.DeleteByUserAndType(ctx, token.UserID, entity.TokenTypePasswordReset)

	return nil
}

// ============================================================================
// Profile Management
// ============================================================================

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
}

// UpdateProfile updates a user's profile information
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrNotFound("User")
	}

	// Update fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.ErrInternal("Failed to update profile")
	}

	return user, nil
}

// GetUserProfile returns a user's profile by ID
func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.ErrNotFound("User")
	}
	return user, nil
}

// ============================================================================
// Email Templates (inline for simplicity)
// ============================================================================

func (s *Service) buildEmailVerificationHTML(userName, verificationURL string) string {
	appName := s.emailConfig.AppName
	if appName == "" {
		appName = "Ads Analytics"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #4F46E5; }
        .button { display: inline-block; background: #4F46E5; color: #fff !important; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">%s</div>
        <h2>Verify Your Email</h2>
        <p>Hi %s,</p>
        <p>Please verify your email by clicking the button below:</p>
        <p style="text-align: center;"><a href="%s" class="button">Verify Email</a></p>
        <p>This link expires in 24 hours.</p>
        <div class="footer"><p>&copy; %s</p></div>
    </div>
</body>
</html>`, appName, userName, verificationURL, appName)
}

func (s *Service) buildPasswordResetHTML(userName, resetURL string) string {
	appName := s.emailConfig.AppName
	if appName == "" {
		appName = "Ads Analytics"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .container { background: #fff; border-radius: 8px; padding: 40px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .logo { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #4F46E5; }
        .button { display: inline-block; background: #DC2626; color: #fff !important; text-decoration: none; padding: 14px 28px; border-radius: 6px; font-weight: 600; }
        .footer { margin-top: 30px; font-size: 12px; color: #666; text-align: center; }
        .warning { background: #FEF3C7; padding: 15px; border-radius: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">%s</div>
        <h2>Reset Your Password</h2>
        <p>Hi %s,</p>
        <p>Click the button below to reset your password:</p>
        <p style="text-align: center;"><a href="%s" class="button">Reset Password</a></p>
        <p>This link expires in 1 hour.</p>
        <div class="warning"><strong>Didn't request this?</strong> Please ignore this email.</div>
        <div class="footer"><p>&copy; %s</p></div>
    </div>
</body>
</html>`, appName, userName, resetURL, appName)
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ============================================================================
// OAuth Platform Connections
// ============================================================================

// GetOAuthURL generates OAuth URL for a platform
func (s *Service) GetOAuthURL(ctx context.Context, userID, orgID uuid.UUID, platform entity.Platform, redirectURL string) (string, error) {
	connector, ok := s.connectorRegistry.Get(platform)
	if !ok {
		return "", errors.ErrBadRequest(fmt.Sprintf("Unsupported platform: %s", platform))
	}

	// Generate state
	state, err := generateState()
	if err != nil {
		return "", errors.ErrInternal("Failed to generate OAuth state")
	}

	// Store state for verification
	oauthState := &entity.OAuthState{
		State:          state,
		OrganizationID: orgID,
		UserID:         userID,
		Platform:       platform,
		RedirectURL:    redirectURL,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}

	if err := s.stateStore.Save(ctx, oauthState); err != nil {
		return "", errors.ErrInternal("Failed to save OAuth state")
	}

	return connector.GetAuthURL(state), nil
}

// HandleOAuthCallback processes OAuth callback from platform
func (s *Service) HandleOAuthCallback(ctx context.Context, platform entity.Platform, code, state string) (*entity.ConnectedAccount, string, error) {
	// Verify state
	oauthState, err := s.stateStore.Get(ctx, state)
	if err != nil || oauthState == nil {
		return nil, "", errors.ErrBadRequest("Invalid OAuth state")
	}

	// Delete state after use
	defer s.stateStore.Delete(ctx, state)

	// Check expiry
	if oauthState.IsExpired() {
		return nil, "", errors.ErrBadRequest("OAuth state expired")
	}

	// Verify platform matches
	if oauthState.Platform != platform {
		return nil, "", errors.ErrBadRequest("Platform mismatch")
	}

	// Get connector
	connector, ok := s.connectorRegistry.Get(platform)
	if !ok {
		return nil, "", errors.ErrBadRequest(fmt.Sprintf("Unsupported platform: %s", platform))
	}

	// Exchange code for tokens
	token, err := connector.ExchangeCode(ctx, code)
	if err != nil {
		return nil, "", errors.Wrap(err, errors.ErrCodeOAuthFailed, "Failed to exchange OAuth code", 400)
	}

	// Get user info from platform
	userInfo, err := connector.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, "", errors.Wrap(err, errors.ErrCodeOAuthFailed, "Failed to get user info", 400)
	}

	// Check if account already connected
	existingAccount, _ := s.connectedAccRepo.GetByPlatformAccountID(
		ctx,
		oauthState.OrganizationID,
		platform,
		userInfo.ID,
	)

	if existingAccount != nil {
		// Update existing account
		existingAccount.AccessToken = token.AccessToken
		existingAccount.RefreshToken = token.RefreshToken
		existingAccount.TokenExpiresAt = &token.ExpiresAt
		existingAccount.TokenScopes = token.Scopes
		existingAccount.Status = entity.AccountStatusActive
		existingAccount.PlatformAccountName = userInfo.Name

		if err := s.connectedAccRepo.Update(ctx, existingAccount); err != nil {
			return nil, "", errors.ErrInternal("Failed to update connected account")
		}

		return existingAccount, oauthState.RedirectURL, nil
	}

	// Create new connected account
	account := &entity.ConnectedAccount{
		BaseEntity:          entity.NewBaseEntity(),
		OrganizationID:      oauthState.OrganizationID,
		Platform:            platform,
		PlatformAccountID:   userInfo.ID,
		PlatformAccountName: userInfo.Name,
		PlatformUserID:      userInfo.ID,
		AccessToken:         token.AccessToken,
		RefreshToken:        token.RefreshToken,
		TokenType:           token.TokenType,
		TokenExpiresAt:      &token.ExpiresAt,
		TokenScopes:         token.Scopes,
		Status:              entity.AccountStatusActive,
	}

	if err := s.connectedAccRepo.Create(ctx, account); err != nil {
		return nil, "", errors.ErrInternal("Failed to create connected account")
	}

	return account, oauthState.RedirectURL, nil
}

// DisconnectPlatform disconnects a platform account
func (s *Service) DisconnectPlatform(ctx context.Context, accountID uuid.UUID) error {
	account, err := s.connectedAccRepo.GetByID(ctx, accountID)
	if err != nil {
		return errors.ErrNotFound("Connected account")
	}

	// Revoke token if connector supports it
	connector, ok := s.connectorRegistry.Get(account.Platform)
	if ok {
		_ = connector.RevokeToken(ctx, account.AccessToken)
	}

	// Update status to revoked
	account.Status = entity.AccountStatusRevoked
	return s.connectedAccRepo.Update(ctx, account)
}

// GetConnectedAccounts returns all connected accounts for an organization
func (s *Service) GetConnectedAccounts(ctx context.Context, orgID uuid.UUID) ([]entity.ConnectedAccount, error) {
	return s.connectedAccRepo.ListByOrganization(ctx, orgID)
}

// ============================================================================
// Helper Methods
// ============================================================================

func (s *Service) getPermissionsForRole(role entity.UserRole) []string {
	switch role {
	case entity.RoleOwner:
		return []string{"*"}
	case entity.RoleAdmin:
		return []string{
			"campaigns:read", "campaigns:write",
			"analytics:read", "analytics:write",
			"accounts:read", "accounts:write",
			"users:read", "users:invite",
		}
	case entity.RoleAnalyst:
		return []string{
			"campaigns:read",
			"analytics:read",
			"accounts:read",
		}
	case entity.RoleViewer:
		return []string{
			"campaigns:read",
			"analytics:read",
		}
	default:
		return []string{}
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateSlug(name string) string {
	// Simple slug generation - should be more robust in production
	slug := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			slug += string(r)
		} else if r >= 'A' && r <= 'Z' {
			slug += string(r + 32) // lowercase
		} else if r == ' ' || r == '-' {
			slug += "-"
		}
	}
	// Add random suffix for uniqueness
	b := make([]byte, 4)
	rand.Read(b)
	return slug + "-" + base64.RawURLEncoding.EncodeToString(b)[:6]
}
