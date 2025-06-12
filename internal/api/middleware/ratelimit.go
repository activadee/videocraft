package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	cleanup  *time.Ticker
}

type visitor struct {
	limiter  *tokenBucket
	lastSeen time.Time
}

type tokenBucket struct {
	tokens   int
	capacity int
	refill   time.Time
	mu       sync.Mutex
}

func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		cleanup:  time.NewTicker(time.Minute),
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return func(c *gin.Context) {
		// Skip rate limiting for health endpoints
		if isHealthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		ip := c.ClientIP()
		
		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			limiter: &tokenBucket{
				tokens:   rl.rate,
				capacity: rl.rate,
				refill:   time.Now(),
			},
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	v.lastSeen = time.Now()
	return v.limiter.allow()
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	
	// Refill tokens based on time elapsed
	if now.After(tb.refill) {
		elapsed := now.Sub(tb.refill)
		tokensToAdd := int(elapsed.Minutes())
		
		if tokensToAdd > 0 {
			tb.tokens += tokensToAdd
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			tb.refill = now
		}
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (rl *rateLimiter) cleanupVisitors() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		
		cutoff := time.Now().Add(-5 * time.Minute)
		for ip, v := range rl.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(rl.visitors, ip)
			}
		}
		
		rl.mu.Unlock()
	}
}