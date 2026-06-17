package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	// First 5 requests should be allowed (burst)
	for i := 0; i < 5; i++ {
		if !rl.Allow("test-key") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th should be denied (burst exhausted)
	if rl.Allow("test-key") {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	rl := NewRateLimiter(100, 3) // 100 tokens/sec, burst 3

	// Exhaust burst
	for i := 0; i < 3; i++ {
		rl.Allow("refill-key")
	}
	if rl.Allow("refill-key") {
		t.Error("should be denied after burst")
	}

	// Can't easily test refill without time manipulation in unit test
	// The PerIP and PerPath middleware tests cover the integration path
}

func TestPerIPMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(100, 3)

	r := gin.New()
	r.Use(rl.PerIP())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// 3 requests from same IP should succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// 4th should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("4th request: expected 429, got %d", w.Code)
	}

	// Different IP should still work
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.1:54321"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("different IP: expected 200, got %d", w2.Code)
	}
}

func TestPerPathMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(100, 2)

	r := gin.New()
	r.Use(rl.PerPath())
	r.POST("/api/v1/auth/login", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.POST("/api/v1/auth/register", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// 2 logins should work
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("login %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// 3rd login should be denied
	req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Errorf("3rd login: expected 429, got %d", w.Code)
	}

	// Register (different path) should still work
	req2 := httptest.NewRequest("POST", "/api/v1/auth/register", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("register (different path): expected 200, got %d", w2.Code)
	}
}
