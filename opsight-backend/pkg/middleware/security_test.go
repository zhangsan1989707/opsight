package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestRouter(mw gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func TestSecurityHeaders(t *testing.T) {
	r := setupTestRouter(SecurityHeaders())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, val := range expected {
		if got := w.Header().Get(header); got != val {
			t.Errorf("header %s: expected %q, got %q", header, val, got)
		}
	}
}

func TestInputValidation_BlocksPathTraversal(t *testing.T) {
	r := setupTestRouter(InputValidation())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?file=../../etc/passwd", nil)
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400 for path traversal, got %d", w.Code)
	}
}

func TestInputValidation_BlocksXSS(t *testing.T) {
	r := setupTestRouter(InputValidation())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?q=<script>alert(1)</script>", nil)
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400 for XSS attempt, got %d", w.Code)
	}
}

func TestInputValidation_BlocksSQLInjection(t *testing.T) {
	r := setupTestRouter(InputValidation())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?q=1 UNION SELECT * FROM users", nil)
	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400 for SQL injection, got %d", w.Code)
	}
}

func TestInputValidation_AllowsNormal(t *testing.T) {
	r := setupTestRouter(InputValidation())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?q=hello+world&page=2", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200 for normal request, got %d", w.Code)
	}
}
