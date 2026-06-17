package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDNew(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid, _ := c.Get("request_id")
		c.JSON(200, gin.H{"request_id": rid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status: %d", w.Code)
	}

	// Response header should contain X-Request-ID
	rid := w.Header().Get(HeaderRequestID)
	if rid == "" {
		t.Fatal("X-Request-ID header is empty")
	}
	if len(rid) != 8 {
		t.Errorf("request ID length: got %d, want 8", len(rid))
	}
}

func TestRequestIDPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid, _ := c.Get("request_id")
		c.String(200, rid.(string))
	})

	// Send a custom request ID
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderRequestID, "custom-abc")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "custom-abc" {
		t.Errorf("request ID not passed through: got %q", w.Body.String())
	}
	if w.Header().Get(HeaderRequestID) != "custom-abc" {
		t.Error("response should echo the custom request ID")
	}
}

func TestRequestIDsAreUnique(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		rid, _ := c.Get("request_id")
		c.String(200, rid.(string))
	})

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		ids[w.Body.String()] = true
	}

	if len(ids) != 100 {
		t.Errorf("expected 100 unique IDs, got %d", len(ids))
	}
}
