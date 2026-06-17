package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateBucket struct {
	tokens   float64
	lastTime time.Time
}

type RateLimiter struct {
	mu     sync.Mutex
	bursts map[string]*rateBucket
	rate   float64 // tokens per second
	burst  int     // max burst
}

func NewRateLimiter(ratePerSec float64, burst int) *RateLimiter {
	return &RateLimiter{
		bursts: make(map[string]*rateBucket),
		rate:   ratePerSec,
		burst:  burst,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.bursts[key]
	now := time.Now()
	if !ok {
		rl.bursts[key] = &rateBucket{tokens: float64(rl.burst) - 1, lastTime: now}
		return true
	}

	elapsed := now.Sub(b.lastTime).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastTime = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Cleanup expired entries periodically.
func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			rl.mu.Lock()
			cutoff := time.Now().Add(-interval)
			for k, b := range rl.bursts {
				if b.lastTime.Before(cutoff) {
					delete(rl.bursts, k)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// PerIP returns a Gin middleware that rate-limits by client IP.
func (rl *RateLimiter) PerIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// PerPath returns middleware that rate-limits by path (e.g. login endpoint).
func (rl *RateLimiter) PerPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP() + ":" + c.FullPath()
		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
