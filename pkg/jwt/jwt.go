package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims structure
type Claims struct {
	jwt.RegisteredClaims
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"org_id,omitempty"`
	Email          string    `json:"email,omitempty"`
	Role           string    `json:"role,omitempty"`
	TokenType      TokenType `json:"token_type"`
	Permissions    []string  `json:"permissions,omitempty"`
}

// Manager handles JWT token operations
type Manager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

// NewManager creates a new JWT manager
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		issuer:             "ads-aggregator",
	}
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"`
}

// GenerateTokenPair generates both access and refresh tokens
func (m *Manager) GenerateTokenPair(userID, orgID, email, role string, permissions []string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessExpiry := now.Add(m.accessTokenExpiry)
	accessToken, err := m.generateToken(userID, orgID, email, role, permissions, AccessToken, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshExpiry := now.Add(m.refreshTokenExpiry)
	refreshToken, err := m.generateToken(userID, orgID, email, role, nil, RefreshToken, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiry,
		RefreshTokenExpiresAt: refreshExpiry,
		TokenType:             "Bearer",
	}, nil
}

// GenerateAccessToken generates only an access token
func (m *Manager) GenerateAccessToken(userID, orgID, email, role string, permissions []string) (string, time.Time, error) {
	expiry := time.Now().Add(m.accessTokenExpiry)
	token, err := m.generateToken(userID, orgID, email, role, permissions, AccessToken, expiry)
	return token, expiry, err
}

// GenerateRefreshToken generates only a refresh token
func (m *Manager) GenerateRefreshToken(userID, orgID, email, role string) (string, time.Time, error) {
	expiry := time.Now().Add(m.refreshTokenExpiry)
	token, err := m.generateToken(userID, orgID, email, role, nil, RefreshToken, expiry)
	return token, expiry, err
}

// generateToken generates a JWT token with the given claims
func (m *Manager) generateToken(userID, orgID, email, role string, permissions []string, tokenType TokenType, expiry time.Time) (string, error) {
	// Generate a unique token ID
	jti, err := generateTokenID()
	if err != nil {
		return "", err
	}

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
		UserID:         userID,
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		TokenType:      tokenType,
		Permissions:    permissions,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// RefreshTokenPair refreshes an access token using a refresh token
func (m *Manager) RefreshTokenPair(refreshTokenString string) (*TokenPair, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Generate new token pair with the same user info
	return m.GenerateTokenPair(
		claims.UserID,
		claims.OrganizationID,
		claims.Email,
		claims.Role,
		nil, // Permissions will be fetched fresh for access token
	)
}

// ExtractTokenFromHeader extracts the JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrMissingToken
	}

	// Check for "Bearer " prefix
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", ErrInvalidAuthHeader
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// generateTokenID generates a unique token ID
func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// JWT Errors
var (
	ErrTokenExpired      = errors.New("token has expired")
	ErrTokenMalformed    = errors.New("token is malformed")
	ErrTokenNotValidYet  = errors.New("token is not valid yet")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrMissingToken      = errors.New("missing authentication token")
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
)

// IsTokenExpired checks if the error is a token expired error
func IsTokenExpired(err error) bool {
	return errors.Is(err, ErrTokenExpired)
}

// IsTokenInvalid checks if the error indicates an invalid token
func IsTokenInvalid(err error) bool {
	return errors.Is(err, ErrTokenInvalid) ||
		errors.Is(err, ErrTokenMalformed) ||
		errors.Is(err, ErrInvalidTokenType)
}
