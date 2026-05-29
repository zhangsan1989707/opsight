package middleware

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"opsight-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds common security-related HTTP headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// maxBodySize returns the maximum allowed request body size in bytes.
// Reads MAX_BODY_SIZE env var, defaults to 1 MB (1048576 bytes).
func maxBodySize() int64 {
	val := os.Getenv("MAX_BODY_SIZE")
	if val == "" {
		return 1048576 // 1 MB
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 1048576
	}
	return n
}

// BodySizeLimit limits the request body size. Returns 413 if exceeded.
func BodySizeLimit() gin.HandlerFunc {
	limit := maxBodySize()
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		// Read at most limit+1 bytes to detect overflow
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, limit+1))
		if err != nil {
			response.Error(c, http.StatusRequestEntityTooLarge, 413, "request body too large")
			c.Abort()
			return
		}
		if int64(len(body)) > limit {
			response.Error(c, http.StatusRequestEntityTooLarge, 413, "request body too large")
			c.Abort()
			return
		}

		// Restore the body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		c.Next()
	}
}

// suspiciousPatterns returns a list of suspicious substrings to check in query params.
func suspiciousPatterns() []string {
	return []string{
		"../", "..\\", // path traversal
		"<script",      // XSS
		"onerror=",     // XSS
		"onload=",      // XSS
		"javascript:",  // script injection
		"data:",        // data URI injection
		"union select", // SQL injection
		"<img",         // HTML injection
		"<svg",         // SVG injection
	}
}

// InputValidation checks query parameters for suspicious patterns.
func InputValidation() gin.HandlerFunc {
	patterns := suspiciousPatterns()
	return func(c *gin.Context) {
		raw := c.Request.URL.RawQuery
		if raw == "" {
			c.Next()
			return
		}
		lower := strings.ToLower(raw)
		for _, pattern := range patterns {
			if strings.Contains(lower, pattern) {
				response.Error(c, http.StatusBadRequest, 400, "request contains invalid characters")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
