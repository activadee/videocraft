package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/pkg/logger"
)

// SecureCORS creates a secure CORS middleware that:
// 1. Removes wildcard origins
// 2. Implements domain allowlisting
// 3. Logs security violations
// 4. Configures proper CORS headers
func SecureCORS(cfg *config.Config, log logger.Logger) gin.HandlerFunc {
	// If no allowed domains configured, reject all cross-origin requests
	if len(cfg.Security.AllowedDomains) == 0 {
		log.Warn("No allowed domains configured for CORS - rejecting all cross-origin requests")
		return rejectAllCORS(log)
	}

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
		MaxAge:          43200, // 12 hours preflight cache (12 * 3600 seconds)
		
		// Custom origin validator for additional security
		AllowOriginFunc: func(origin string) bool {
			return validateOrigin(origin, cfg.Security.AllowedDomains, log)
		},
	}

	return cors.New(corsConfig)
}

// validateOrigin performs strict origin validation with security logging
func validateOrigin(origin string, allowedDomains []string, log logger.Logger) bool {
	// Empty origin is allowed (same-origin requests)
	if origin == "" {
		return true
	}

	// Strict domain matching - no wildcards, no subdomains unless explicitly allowed
	for _, allowedDomain := range allowedDomains {
		if origin == "https://"+allowedDomain || origin == "http://"+allowedDomain {
			return true
		}
	}

	// Log security violation
	log.WithFields(map[string]interface{}{
		"origin":          origin,
		"allowed_domains": allowedDomains,
		"violation_type":  "CORS_ORIGIN_REJECTED",
	}).Warnf("CORS_SECURITY_VIOLATION: Origin not in allowlist: %s", origin)

	return false
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