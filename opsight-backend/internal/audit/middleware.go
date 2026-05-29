package audit

import (
	"strings"

	"opsight-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware returns a Gin middleware that automatically logs all
// POST, PUT, PATCH, and DELETE requests after they complete.
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip read-only methods
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		// Process the request
		c.Next()

		// Extract user info from context (set by auth middleware)
		userID, email, _ := auth.GetCurrentUser(c)
		userName := email
		if userName == "" {
			userName = "anonymous"
		}

		// Determine action from method
		action := strings.ToLower(method)
		resource := strings.TrimPrefix(c.FullPath(), "/api/v1/")
		if resource == "" {
			resource = c.Request.URL.Path
		}

		// Determine status from response code
		status := "success"
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			status = "failure"
		}

		Log(
			userID,
			userName,
			action,
			resource,
			"",
			"",
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			status,
		)
	}
}
