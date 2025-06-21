package middleware

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// originCache provides thread-safe caching for origin validation results
type originCache struct {
	mu    sync.RWMutex
	cache map[string]bool
}

// newOriginCache creates a new origin validation cache
func newOriginCache() *originCache {
	return &originCache{
		cache: make(map[string]bool),
	}
}

// get retrieves a cached validation result
func (oc *originCache) get(origin string) (bool, bool) {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	result, exists := oc.cache[origin]
	return result, exists
}

// set stores a validation result in cache
func (oc *originCache) set(origin string, valid bool) {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	oc.cache[origin] = valid
}

// SecureCORS creates a secure CORS middleware that:
// 1. Removes wildcard origins (SECURITY FIX)
// 2. Implements strict domain allowlisting
// 3. Logs security violations with structured data
// 4. Configures secure CORS headers
// 5. Caches origin validation for performance
func SecureCORS(cfg *app.Config, log logger.Logger) gin.HandlerFunc {
	// If no allowed domains configured, reject all cross-origin requests
	if len(cfg.Security.AllowedDomains) == 0 {
		log.WithFields(map[string]interface{}{
			"security_policy": "CORS_STRICT_MODE",
			"allowed_domains": 0,
		}).Warn("No allowed domains configured for CORS - rejecting all cross-origin requests")
		return rejectAllCORS(log)
	}

	// Create origin validation cache for performance
	cache := newOriginCache()

	log.WithFields(map[string]interface{}{
		"security_policy":   "CORS_DOMAIN_ALLOWLIST",
		"allowed_domains":   cfg.Security.AllowedDomains,
		"domains_count":     len(cfg.Security.AllowedDomains),
		"allow_credentials": len(cfg.Security.AllowedDomains) == 1,
	}).Info("Secure CORS middleware initialized with domain allowlist")

	// Prepare allowed origins with proper protocol prefixes
	allowedOrigins := make([]string, 0, len(cfg.Security.AllowedDomains)*2)
	for _, domain := range cfg.Security.AllowedDomains {
		// Add both HTTP and HTTPS versions
		if !strings.HasPrefix(domain, "http") {
			allowedOrigins = append(allowedOrigins, "https://"+domain)
			allowedOrigins = append(allowedOrigins, "http://"+domain)
		} else {
			allowedOrigins = append(allowedOrigins, domain)
		}
	}

	// Create CORS config with secure defaults
	corsConfig := cors.Config{
		AllowOrigins: allowedOrigins, // NO WILDCARDS - only specific domains
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
			// Explicitly exclude dangerous methods like TRACE, CONNECT
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token", // Include CSRF token header
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-CSRF-Token",
		},
		// SECURITY: Don't allow credentials with multiple domains
		AllowCredentials: len(cfg.Security.AllowedDomains) == 1,
		MaxAge:           43200, // 12 hours preflight cache (12 * 3600 seconds)

		// Custom origin validator with caching for performance
		AllowOriginFunc: func(origin string) bool {
			return validateOriginWithCache(origin, cfg.Security.AllowedDomains, cache, log)
		},
	}

	return cors.New(corsConfig)
}

// validateOriginWithCache performs cached origin validation for performance
func validateOriginWithCache(origin string, allowedDomains []string, cache *originCache, log logger.Logger) bool {
	// Check cache first for performance
	if cached, exists := cache.get(origin); exists {
		return cached
	}

	// Validate and cache result
	valid := validateOrigin(origin, allowedDomains, log)
	cache.set(origin, valid)
	return valid
}

// validateOrigin performs strict origin validation with enhanced security logging
func validateOrigin(origin string, allowedDomains []string, log logger.Logger) bool {
	// Empty origin is allowed (same-origin requests)
	if origin == "" {
		return true
	}

	// First check if origin is explicitly allowed (before suspicious pattern check)
	for _, allowedDomain := range allowedDomains {
		if isExactDomainMatch(origin, allowedDomain) {
			log.WithFields(map[string]interface{}{
				"origin":         origin,
				"matched_domain": allowedDomain,
				"action":         "CORS_ALLOW",
			}).Debug("CORS origin validation: allowed")
			return true
		}
	}

	// Only check for suspicious patterns if the origin is NOT in allowed domains
	if containsSuspiciousPatterns(origin) {
		log.WithFields(map[string]interface{}{
			"origin":         origin,
			"violation_type": "CORS_SUSPICIOUS_ORIGIN",
			"threat_level":   "HIGH",
		}).Errorf("CORS_SECURITY_VIOLATION: Suspicious origin pattern detected: %s", origin)
		return false
	}

	// Log security violation with enhanced context
	log.WithFields(map[string]interface{}{
		"origin":          origin,
		"allowed_domains": allowedDomains,
		"violation_type":  "CORS_ORIGIN_REJECTED",
		"client_ip":       extractClientIP(origin),
		"threat_level":    "MEDIUM",
	}).Warnf("CORS_SECURITY_VIOLATION: Origin not in allowlist: %s", origin)

	return false
}

// isExactDomainMatch checks if origin exactly matches an allowed domain
func isExactDomainMatch(origin, allowedDomain string) bool {
	// Handle domains that already include protocol
	if strings.HasPrefix(allowedDomain, "http") {
		return origin == allowedDomain
	}
	// Handle bare domains - check both HTTP and HTTPS
	return origin == "https://"+allowedDomain || origin == "http://"+allowedDomain
}

// containsSuspiciousPatterns detects potentially malicious origin patterns
func containsSuspiciousPatterns(origin string) bool {
	suspiciousPatterns := []string{
		"javascript:", "data:", "file:", "ftp:",
		"localhost", "127.0.0.1", "0.0.0.0",
		"//", "\\", "..", "@",
		"<script", "</script>", "eval(",
		"%3cscript", "%3c/script%3e",
	}

	originLower := strings.ToLower(origin)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(originLower, pattern) {
			return true
		}
	}
	return false
}

// extractClientIP attempts to extract IP from origin for logging (simplified)
func extractClientIP(origin string) string {
	// Simple regex would be better, but for logging purposes this is sufficient
	if strings.Contains(origin, "://") {
		parts := strings.Split(origin, "://")
		if len(parts) > 1 {
			hostPart := strings.Split(parts[1], "/")[0]
			return strings.Split(hostPart, ":")[0]
		}
	}
	return "unknown"
}

// rejectAllCORS creates middleware that rejects all cross-origin requests
func rejectAllCORS(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// If there's an Origin header, it's a cross-origin request
		if origin != "" {
			log.WithFields(map[string]interface{}{
				"origin":         origin,
				"method":         c.Request.Method,
				"path":           c.Request.URL.Path,
				"violation_type": "CORS_NO_DOMAINS_CONFIGURED",
			}).Warnf("CORS_SECURITY_VIOLATION: Cross-origin request rejected - no domains configured")

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Cross-origin requests not allowed",
				"code":  "CORS_FORBIDDEN",
			})
			c.Abort()
			return
		}

		// Same-origin requests are allowed
		c.Next()
	}
}
