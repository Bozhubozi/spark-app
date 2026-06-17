package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter uses Redis for distributed rate limiting (token bucket).
type RedisRateLimiter struct {
	rdb   *redis.Client
	rate  int // max requests per window
	burst int // allow bursts up to this
	ttl   time.Duration
}

func NewRedisRateLimiter(rdb *redis.Client, rate, burst int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{rdb: rdb, rate: rate, burst: burst, ttl: window}
}

func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMicro()
	windowStart := now - rl.ttl.Microseconds()

	pipe := rl.rdb.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(windowStart))

	// Count current
	countCmd := pipe.ZCard(ctx, key)

	// Add current request with nanosecond randomness for ordering
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})

	// Set TTL on the key
	pipe.Expire(ctx, key, rl.ttl*2)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, nil // fail open on Redis error
	}

	count, _ := countCmd.Result()
	allowed := count+1 <= int64(rl.burst)
	return allowed, nil
}

// PerIP returns middleware that rate-limits by client IP using Redis.
func (rl *RedisRateLimiter) PerIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ok, _ := rl.Allow(c.Request.Context(), "rl:ip:"+c.ClientIP())
		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// PerPath returns middleware that rate-limits by IP + path using Redis.
func (rl *RedisRateLimiter) PerPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rl:path:" + c.ClientIP() + ":" + c.FullPath()
		ok, _ := rl.Allow(c.Request.Context(), key)
		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
