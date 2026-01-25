package middleware

import (
	"strings"

	"github.com/ads-aggregator/ads-aggregator/pkg/errors"
	"github.com/ads-aggregator/ads-aggregator/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys for storing authentication data
const (
	ContextKeyUserID      = "user_id"
	ContextKeyOrgID       = "org_id"
	ContextKeyEmail       = "email"
	ContextKeyRole        = "role"
	ContextKeyPermissions = "permissions"
	ContextKeyClaims      = "claims"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtManager *jwt.Manager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// Authenticate returns a middleware that validates JWT tokens
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.abortWithError(c, errors.ErrUnauthorized("Missing authorization header"))
			return
		}

		// Extract token from Bearer scheme
		token, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.abortWithError(c, errors.ErrUnauthorized("Invalid authorization header"))
			return
		}

		// Validate token
		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			if jwt.IsTokenExpired(err) {
				m.abortWithError(c, errors.ErrUnauthorized("Token has expired"))
				return
			}
			m.abortWithError(c, errors.ErrUnauthorized("Invalid token"))
			return
		}

		// Parse UUIDs
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			m.abortWithError(c, errors.ErrUnauthorized("Invalid user ID in token"))
			return
		}

		orgID, err := uuid.Parse(claims.OrganizationID)
		if err != nil {
			m.abortWithError(c, errors.ErrUnauthorized("Invalid organization ID in token"))
			return
		}

		// Set context values
		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyOrgID, orgID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyPermissions, claims.Permissions)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// RequireRole returns a middleware that requires a specific role
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			m.abortWithError(c, errors.ErrForbidden("Role not found"))
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		m.abortWithError(c, errors.ErrForbidden("Insufficient role permissions"))
	}
}

// RequirePermission returns a middleware that requires specific permissions
func (m *AuthMiddleware) RequirePermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perms, exists := c.Get(ContextKeyPermissions)
		if !exists {
			m.abortWithError(c, errors.ErrForbidden("Permissions not found"))
			return
		}

		userPermissions := perms.([]string)

		// Check for wildcard permission
		for _, p := range userPermissions {
			if p == "*" {
				c.Next()
				return
			}
		}

		// Check each required permission
		for _, required := range requiredPermissions {
			found := false
			for _, p := range userPermissions {
				if m.matchPermission(p, required) {
					found = true
					break
				}
			}
			if !found {
				m.abortWithError(c, errors.ErrForbidden("Missing required permission: "+required))
				return
			}
		}

		c.Next()
	}
}

// matchPermission checks if a user permission matches the required permission
func (m *AuthMiddleware) matchPermission(userPerm, required string) bool {
	// Exact match
	if userPerm == required {
		return true
	}

	// Wildcard match (e.g., "campaigns:*" matches "campaigns:read")
	if strings.HasSuffix(userPerm, ":*") {
		prefix := strings.TrimSuffix(userPerm, "*")
		if strings.HasPrefix(required, prefix) {
			return true
		}
	}

	return false
}

// RequireOrganization validates that the request is for the authenticated organization
func (m *AuthMiddleware) RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get org ID from URL parameter
		orgIDParam := c.Param("orgId")
		if orgIDParam == "" {
			c.Next()
			return
		}

		requestedOrgID, err := uuid.Parse(orgIDParam)
		if err != nil {
			m.abortWithError(c, errors.ErrBadRequest("Invalid organization ID"))
			return
		}

		// Get authenticated org ID
		authOrgID, exists := c.Get(ContextKeyOrgID)
		if !exists {
			m.abortWithError(c, errors.ErrForbidden("Organization not found in context"))
			return
		}

		// Verify match
		if authOrgID.(uuid.UUID) != requestedOrgID {
			m.abortWithError(c, errors.ErrForbidden("Access denied to this organization"))
			return
		}

		c.Next()
	}
}

// abortWithError aborts the request with an error response
func (m *AuthMiddleware) abortWithError(c *gin.Context, err *errors.AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	})
}

// Helper functions to extract context values

// GetUserID extracts the user ID from the gin context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// GetOrgID extracts the organization ID from the gin context
func GetOrgID(c *gin.Context) (uuid.UUID, bool) {
	orgID, exists := c.Get(ContextKeyOrgID)
	if !exists {
		return uuid.Nil, false
	}
	return orgID.(uuid.UUID), true
}

// GetEmail extracts the email from the gin context
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(ContextKeyEmail)
	if !exists {
		return "", false
	}
	return email.(string), true
}

// GetRole extracts the role from the gin context
func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	return role.(string), true
}

// GetPermissions extracts the permissions from the gin context
func GetPermissions(c *gin.Context) ([]string, bool) {
	perms, exists := c.Get(ContextKeyPermissions)
	if !exists {
		return nil, false
	}
	return perms.([]string), true
}

// GetClaims extracts the full JWT claims from the gin context
func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	return claims.(*jwt.Claims), true
}

// OptionalAuth returns a middleware that optionally authenticates
// It doesn't fail if no token is provided, but sets context if token is valid
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token, err := jwt.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := m.jwtManager.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		userID, _ := uuid.Parse(claims.UserID)
		orgID, _ := uuid.Parse(claims.OrganizationID)

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyOrgID, orgID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyPermissions, claims.Permissions)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}
