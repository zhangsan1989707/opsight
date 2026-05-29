package auth

import (
	"net/http"
	"strings"

	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	// ContextUserID is the gin context key for the authenticated user ID.
	ContextUserID = "auth_user_id"
	// ContextUserEmail is the gin context key for the authenticated user email.
	ContextUserEmail = "auth_user_email"
	// ContextUserRole is the gin context key for the authenticated user role.
	ContextUserRole = "auth_user_role"
)

// AuthRequired extracts and validates the Bearer token from the Authorization
// header. On success it stores user info in the gin context. On failure it
// aborts with a 401 response.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := ValidateToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, response.ErrUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserEmail, claims.Email)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole returns middleware that checks the authenticated user's role
// against the allowed list. Must be used after AuthRequired.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextUserRole)
		if !exists {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, "authentication required")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Error(c, http.StatusForbidden, response.ErrForbidden, "invalid user role")
			c.Abort()
			return
		}

		for _, r := range roles {
			if strings.EqualFold(userRole, r) {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, response.ErrForbidden, "insufficient permissions")
		c.Abort()
	}
}

// GetCurrentUser returns user ID, email, and role from the gin context.
// Returns zero values if the context keys are missing.
func GetCurrentUser(c *gin.Context) (userID uint, email, role string) {
	if v, ok := c.Get(ContextUserID); ok {
		userID, _ = v.(uint)
	}
	if v, ok := c.Get(ContextUserEmail); ok {
		email, _ = v.(string)
	}
	if v, ok := c.Get(ContextUserRole); ok {
		role, _ = v.(string)
	}
	return
}

// GetCurrentUserID returns the authenticated user ID from the gin context.
func GetCurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get(ContextUserID); ok {
		id, _ := v.(uint)
		return id
	}
	return 0
}
