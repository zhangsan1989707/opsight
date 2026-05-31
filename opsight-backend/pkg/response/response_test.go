package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupResponseRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestSuccess_Response(t *testing.T) {
	r := setupResponseRouter()
	r.GET("/test", func(c *gin.Context) {
		Success(c, gin.H{"key": "value"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["code"] != float64(200) {
		t.Errorf("expected code 200, got %v", body["code"])
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok || data["key"] != "value" {
		t.Errorf("expected data.key = value, got %v", body["data"])
	}
}

func TestError_Response(t *testing.T) {
	r := setupResponseRouter()
	r.GET("/test", func(c *gin.Context) {
		Error(c, http.StatusNotFound, ErrNotFound, "not found")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["message"] != "not found" {
		t.Errorf("expected message 'not found', got %v", body["message"])
	}
}

func TestPaginated_Response(t *testing.T) {
	r := setupResponseRouter()
	r.GET("/test", func(c *gin.Context) {
		Paginated(c, []string{"a", "b"}, 100, 2, 20)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["total"] != float64(100) {
		t.Errorf("expected total 100, got %v", body["total"])
	}
	if body["page"] != float64(2) {
		t.Errorf("expected page 2, got %v", body["page"])
	}
	if body["page_size"] != float64(20) {
		t.Errorf("expected page_size 20, got %v", body["page_size"])
	}
}
