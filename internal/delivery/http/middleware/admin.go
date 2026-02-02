package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Role        AdminRole `json:"role" db:"role"`
	Permissions []string  `json:"permissions" db:"permissions"`
}

// AdminContextKey is the context key for admin user
type AdminContextKey struct{}

// AdminMiddleware handles admin authentication and authorization
type AdminMiddleware struct {
	db *sqlx.DB
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(db *sqlx.DB) *AdminMiddleware {
	return &AdminMiddleware{db: db}
}

// RequireAdmin requires the user to be an admin
func (m *AdminMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from auth context (assumes auth middleware ran first)
		userID, ok := r.Context().Value(UserIDKey{}).(uuid.UUID)
		if !ok {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Check if user is an admin
		admin, err := m.getAdminUser(r.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondError(w, http.StatusForbidden, "admin access required")
				return
			}
			respondError(w, http.StatusInternalServerError, "failed to check admin status")
			return
		}

		// Add admin to context
		ctx := context.WithValue(r.Context(), AdminContextKey{}, admin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole requires a specific admin role
func (m *AdminMiddleware) RequireRole(roles ...AdminRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			admin, ok := r.Context().Value(AdminContextKey{}).(*AdminUser)
			if !ok {
				respondError(w, http.StatusForbidden, "admin access required")
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
				respondError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission requires a specific permission
func (m *AdminMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			admin, ok := r.Context().Value(AdminContextKey{}).(*AdminUser)
			if !ok {
				respondError(w, http.StatusForbidden, "admin access required")
				return
			}

			// Super admin has all permissions
			if admin.Role == AdminRoleSuperAdmin {
				next.ServeHTTP(w, r)
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
				respondError(w, http.StatusForbidden, "permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getAdminUser retrieves the admin user from the database
func (m *AdminMiddleware) getAdminUser(ctx context.Context, userID uuid.UUID) (*AdminUser, error) {
	query := `
		SELECT id, user_id, role, permissions
		FROM admin_users
		WHERE user_id = $1
	`

	var admin AdminUser
	var permissions []byte

	err := m.db.QueryRowContext(ctx, query, userID).Scan(
		&admin.ID, &admin.UserID, &admin.Role, &permissions,
	)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		json.Unmarshal(permissions, &admin.Permissions)
	}

	return &admin, nil
}

// AuditLog logs admin actions
func (m *AdminMiddleware) AuditLog(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			admin, _ := r.Context().Value(AdminContextKey{}).(*AdminUser)

			// Log the action after the request completes
			defer func() {
				if admin != nil {
					m.logAdminAction(r.Context(), admin.ID, action, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// logAdminAction logs an admin action to the database
func (m *AdminMiddleware) logAdminAction(ctx context.Context, adminID uuid.UUID, action string, r *http.Request) {
	query := `
		INSERT INTO admin_audit_log (admin_user_id, action, ip_address, details)
		VALUES ($1, $2, $3, $4)
	`

	details := map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.Query(),
	}
	detailsJSON, _ := json.Marshal(details)

	m.db.ExecContext(ctx, query, adminID, action, getClientIP(r), detailsJSON)
}

// UserIDKey is used for user ID context key
type UserIDKey struct{}

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

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// GetAdminFromContext retrieves the admin user from context
func GetAdminFromContext(ctx context.Context) *AdminUser {
	admin, _ := ctx.Value(AdminContextKey{}).(*AdminUser)
	return admin
}
