package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		// Get user identifier (API key or IP fallback)
		userID := getUserIdentifier(c)

		allowed, remaining := rl.allow(userID)

		if !allowed {
			// Add rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.rate))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"details": gin.H{
					"limit":       rl.rate,
					"window":      "1 minute",
					"retry_after": 60,
				},
			})

			// Log rate limit violation (hash user ID for security)
			logrus.WithFields(logrus.Fields{
				"user_id":  hashUserIDForLogging(userID),
				"endpoint": c.Request.URL.Path,
				"method":   c.Request.Method,
				"ip":       c.ClientIP(),
			}).Warn("Rate limit exceeded")

			c.Abort()
			return
		}

		// Add rate limit headers for successful requests
		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.rate))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

func (rl *rateLimiter) allow(userID string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[userID]
	if !exists {
		v = &visitor{
			limiter: &tokenBucket{
				tokens:   rl.rate,
				capacity: rl.rate,
				refill:   time.Now(),
			},
			lastSeen: time.Now(),
		}
		rl.visitors[userID] = v
	}

	v.lastSeen = time.Now()
	allowed := v.limiter.allow()
	remaining := v.limiter.tokens
	return allowed, remaining
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
		for userID, v := range rl.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(rl.visitors, userID)
			}
		}

		rl.mu.Unlock()
	}
}

// getUserIdentifier extracts user identifier from the request
func getUserIdentifier(c *gin.Context) string {
	// Try to get API key from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			apiKey := strings.TrimPrefix(authHeader, "Bearer ")
			if apiKey != "" {
				return apiKey // Use API key as user identifier
			}
		}
	}

	// Fallback to client IP for unauthenticated requests
	return c.ClientIP()
}

// hashUserIDForLogging creates a safe hash of the user ID for logging
// This prevents API keys from being logged in plaintext while maintaining traceability
func hashUserIDForLogging(userID string) string {
	// For IP addresses, log them directly as they're not sensitive
	if strings.Contains(userID, ".") || strings.Contains(userID, ":") {
		return userID
	}

	// For API keys, create a SHA-256 hash with a short prefix for identification
	h := sha256.Sum256([]byte(userID))
	hash := hex.EncodeToString(h[:])

	// Return first 8 characters of hash with a prefix for easier identification
	return "hash:" + hash[:8]
}
