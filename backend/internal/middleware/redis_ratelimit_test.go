package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func setupRedisLimiter(t *testing.T) *RedisRateLimiter {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	return NewRedisRateLimiter(rdb, 100, 5, time.Second) // 5 burst per second
}

func TestRedisRateLimiterAllow(t *testing.T) {
	rl := setupRedisLimiter(t)
	ctx := context.Background()
	key := "test:rl:allow"

	// Cleanup
	rl.rdb.Del(ctx, key)

	// First 5 should pass (burst)
	for i := 0; i < 5; i++ {
		ok, err := rl.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Allow %d: %v", i, err)
		}
		if !ok {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th should be denied
	ok, err := rl.Allow(ctx, key)
	if err == nil && ok {
		t.Error("6th request should be denied")
	}
}

func TestRedisRateLimiterPerIP(t *testing.T) {
	rl := setupRedisLimiter(t)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(rl.PerIP())
	r.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// 5 requests from same IP should pass
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.1.1.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("request %d: %d", i+1, w.Code)
		}
	}

	// 6th should fail
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.1.1.1:9999"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Errorf("expected 429, got %d", w.Code)
	}

	// Different IP should work
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.2.2.2:9999"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("different IP: expected 200, got %d", w2.Code)
	}
}

func TestRedisRateLimiterPerPath(t *testing.T) {
	rl := setupRedisLimiter(t)
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(rl.PerPath())
	r.POST("/login", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	r.POST("/register", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// Exhaust login path
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "10.1.1.1:9999"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// Login should be denied
	req := httptest.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "10.1.1.1:9999"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 429 {
		t.Error("login should be rate limited")
	}

	// Register (different path) should still work
	req2 := httptest.NewRequest("POST", "/register", nil)
	req2.RemoteAddr = "10.1.1.1:9999"
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code == 429 {
		t.Error("register should not be rate limited")
	}
}
