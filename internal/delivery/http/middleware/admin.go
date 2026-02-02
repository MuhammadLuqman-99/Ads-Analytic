package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminRole represents admin role levels
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleViewer     AdminRole = "viewer"
)

// AdminUser represents an admin user
type AdminUser struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Role            AdminRole `json:"role" gorm:"type:varchar(50);not null;default:'admin'"`
	Permissions     []string  `json:"permissions" gorm:"-"`
	PermissionsJSON string    `json:"-" gorm:"column:permissions;type:jsonb;default:'[]'"`
}

// TableName returns the table name for AdminUser
func (AdminUser) TableName() string {
	return "admin_users"
}

// AdminContextKey is the context key for admin user
const AdminContextKey = "admin_user"

// AdminMiddleware handles admin authentication and authorization
type AdminMiddleware struct {
	db *gorm.DB
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(db *gorm.DB) *AdminMiddleware {
	return &AdminMiddleware{db: db}
}

// RequireAdmin requires the user to be an admin
func (m *AdminMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from auth context (assumes auth middleware ran first)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			// Try string conversion
			if userIDStr, ok := userIDValue.(string); ok {
				var err error
				userID, err = uuid.Parse(userIDStr)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		}

		// Check if user is an admin
		admin, err := m.getAdminUser(userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to check admin status"})
			return
		}

		// Add admin to context
		c.Set(AdminContextKey, admin)
		c.Next()
	}
}

// RequireRole requires a specific admin role
func (m *AdminMiddleware) RequireRole(roles ...AdminRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminValue, exists := c.Get(AdminContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		admin, ok := adminValue.(*AdminUser)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		hasRole := false
		for _, role := range roles {
			if admin.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

// RequirePermission requires a specific permission
func (m *AdminMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminValue, exists := c.Get(AdminContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		admin, ok := adminValue.(*AdminUser)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		// Super admin has all permissions
		if admin.Role == AdminRoleSuperAdmin {
			c.Next()
			return
		}

		hasPermission := false
		for _, p := range admin.Permissions {
			if p == permission || p == "*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}

		c.Next()
	}
}

// getAdminUser retrieves the admin user from the database
func (m *AdminMiddleware) getAdminUser(userID uuid.UUID) (*AdminUser, error) {
	var admin AdminUser

	err := m.db.Where("user_id = ?", userID).First(&admin).Error
	if err != nil {
		return nil, err
	}

	// Parse permissions JSON
	if admin.PermissionsJSON != "" {
		json.Unmarshal([]byte(admin.PermissionsJSON), &admin.Permissions)
	}

	return &admin, nil
}

// AuditLog logs admin actions
func (m *AdminMiddleware) AuditLog(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log the action after the request completes
		c.Next()

		adminValue, exists := c.Get(AdminContextKey)
		if !exists {
			return
		}

		admin, ok := adminValue.(*AdminUser)
		if !ok {
			return
		}

		m.logAdminAction(admin.ID, action, c)
	}
}

// logAdminAction logs an admin action to the database
func (m *AdminMiddleware) logAdminAction(adminID uuid.UUID, action string, c *gin.Context) {
	details := map[string]interface{}{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"query":  c.Request.URL.Query(),
		"status": c.Writer.Status(),
	}
	detailsJSON, _ := json.Marshal(details)

	m.db.Exec(`
		INSERT INTO admin_audit_log (admin_user_id, action, ip_address, details)
		VALUES (?, ?, ?, ?)
	`, adminID, action, getClientIP(c.Request), detailsJSON)
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// GetAdminFromContext retrieves the admin user from Gin context
func GetAdminFromContext(c *gin.Context) *AdminUser {
	adminValue, exists := c.Get(AdminContextKey)
	if !exists {
		return nil
	}
	admin, _ := adminValue.(*AdminUser)
	return admin
}
