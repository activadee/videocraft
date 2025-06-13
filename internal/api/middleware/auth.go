package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health endpoints
		if isHealthEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get API key from header or query parameter
		authHeader := c.GetHeader("Authorization")
		var providedKey string

		if authHeader != "" {
			// Handle Bearer token format
			if strings.HasPrefix(authHeader, "Bearer ") {
				providedKey = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				providedKey = authHeader
			}
		} else {
			// Fallback to query parameter
			providedKey = c.Query("api_key")
		}

		// Validate API key
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required",
				"code":  "MISSING_API_KEY",
			})
			c.Abort()
			return
		}

		if providedKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
				"code":  "INVALID_API_KEY",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isHealthEndpoint(path string) bool {
	healthPaths := []string{
		"/health",
		"/ready",
		"/live",
		"/metrics",
	}

	for _, healthPath := range healthPaths {
		if strings.HasPrefix(path, healthPath) {
			return true
		}
	}

	return false
}
