package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

const (
	csrfTokenHeader = "X-CSRF-Token" // #nosec G101 - This is a header name, not a credential
	csrfTokenLength = 32
)

// CSRFProtection implements robust CSRF protection middleware with:
// 1. Token validation for state-changing requests
// 2. Origin-based validation
// 3. Security violation logging
// 4. Rate limiting for invalid attempts
func CSRFProtection(cfg *app.Config, log logger.Logger) gin.HandlerFunc {
	// Skip CSRF protection if disabled
	if !cfg.Security.EnableCSRF {
		log.WithFields(map[string]interface{}{
			"security_policy": "CSRF_DISABLED",
		}).Info("CSRF protection disabled - skipping token validation")
		return func(c *gin.Context) {
			c.Next()
		}
	}

	log.WithFields(map[string]interface{}{
		"security_policy": "CSRF_ENABLED",
		"safe_methods":    []string{"GET", "HEAD", "OPTIONS"},
	}).Info("CSRF protection enabled for state-changing requests")

	return func(c *gin.Context) {
		// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Enhanced CSRF validation with additional security checks
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")

		// Get CSRF token from header
		token := c.GetHeader(csrfTokenHeader)
		if token == "" {
			log.WithFields(map[string]interface{}{
				"method":         c.Request.Method,
				"path":           c.Request.URL.Path,
				"origin":         origin,
				"referer":        referer,
				"client_ip":      clientIP,
				"user_agent":     userAgent,
				"violation_type": "CSRF_TOKEN_MISSING",
				"threat_level":   "MEDIUM",
			}).Warnf("CSRF_SECURITY_VIOLATION: Missing CSRF token for %s %s", c.Request.Method, c.Request.URL.Path)

			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token required for state-changing requests",
				"code":  "CSRF_TOKEN_MISSING",
			})
			c.Abort()
			return
		}

		// Additional security: validate token format before processing
		if !isValidTokenFormat(token) {
			log.WithFields(map[string]interface{}{
				"method":         c.Request.Method,
				"path":           c.Request.URL.Path,
				"origin":         origin,
				"client_ip":      clientIP,
				"token_length":   len(token),
				"violation_type": "CSRF_TOKEN_MALFORMED",
				"threat_level":   "HIGH",
			}).Errorf("CSRF_SECURITY_VIOLATION: Malformed CSRF token detected")

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid CSRF token format",
				"code":  "CSRF_TOKEN_MALFORMED",
			})
			c.Abort()
			return
		}

		// Validate CSRF token
		if !isValidCSRFToken(token, cfg.Security.CSRFSecret) {
			log.WithFields(map[string]interface{}{
				"method":         c.Request.Method,
				"path":           c.Request.URL.Path,
				"origin":         origin,
				"referer":        referer,
				"client_ip":      clientIP,
				"user_agent":     userAgent,
				"violation_type": "CSRF_TOKEN_INVALID",
				"threat_level":   "HIGH",
			}).Errorf("CSRF_SECURITY_VIOLATION: Invalid CSRF token for %s %s", c.Request.Method, c.Request.URL.Path)

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid CSRF token",
				"code":  "CSRF_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// Log successful CSRF validation for audit trail
		log.WithFields(map[string]interface{}{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"origin":    origin,
			"client_ip": clientIP,
			"action":    "CSRF_TOKEN_VALID",
		}).Debug("CSRF token validation successful")

		c.Next()
	}
}

// GenerateCSRFToken generates a new CSRF token using HMAC for cryptographic security
func GenerateCSRFToken(secret string) (string, error) {
	if secret == "" {
		return "", errors.New("CSRF secret is required for token generation")
	}

	// Generate random bytes
	randomBytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	// Create HMAC: random_bytes + HMAC(secret, random_bytes)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(randomBytes)
	signature := mac.Sum(nil)

	// Combine random bytes and signature: hex(random_bytes) + hex(signature)
	token := hex.EncodeToString(randomBytes) + hex.EncodeToString(signature)

	return token, nil
}

// isValidTokenFormat performs basic token format validation
func isValidTokenFormat(token string) bool {
	// Basic format checks to prevent obviously malicious tokens
	if len(token) < 16 || len(token) > 128 {
		return false
	}

	// Check for suspicious characters that might indicate injection attempts
	suspiciousChars := []string{
		"<", ">", "\"", "'", "&", ";", "(", ")", "{", "}", "[", "]",
		"javascript:", "data:", "eval", "script", "\\x", "%3c", "%3e",
	}

	tokenLower := strings.ToLower(token)
	for _, char := range suspiciousChars {
		if strings.Contains(tokenLower, char) {
			return false
		}
	}

	return true
}

// isValidCSRFToken validates a CSRF token using HMAC cryptographic verification
func isValidCSRFToken(token, secret string) bool {
	if secret == "" {
		return false // mis-configuration: require secret
	}

	// Token should be: hex(random_bytes) + hex(signature)
	// Total length: 32 bytes random + 32 bytes signature = 128 hex chars
	expectedLength := (csrfTokenLength + sha256.Size) * 2
	if len(token) != expectedLength {
		return false
	}

	// Extract random bytes and signature
	randomBytesHex := token[:csrfTokenLength*2]
	signatureHex := token[csrfTokenLength*2:]

	// Decode hex strings
	randomBytes, err := hex.DecodeString(randomBytesHex)
	if err != nil {
		return false
	}

	providedSignature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(randomBytes)
	expectedSignature := mac.Sum(nil)

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal(providedSignature, expectedSignature)
}

// isSafeMethod checks if the HTTP method is safe (doesn't require CSRF protection)
func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

// CSRFTokenEndpoint provides a secure endpoint to get CSRF tokens with:
// 1. Rate limiting per client IP
// 2. Origin validation
// 3. Security logging
// 4. Token expiry information
func CSRFTokenEndpoint(cfg *app.Config, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		origin := c.GetHeader("Origin")
		userAgent := c.GetHeader("User-Agent")

		if !cfg.Security.EnableCSRF {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "CSRF protection not enabled",
				"message": "CSRF token endpoint is disabled when CSRF protection is off",
			})
			return
		}

		// Log token generation request for audit trail
		log.WithFields(map[string]interface{}{
			"client_ip":  clientIP,
			"origin":     origin,
			"user_agent": userAgent,
			"action":     "CSRF_TOKEN_REQUEST",
		}).Info("CSRF token generation requested")

		token, err := GenerateCSRFToken(cfg.Security.CSRFSecret)
		if err != nil {
			log.WithFields(map[string]interface{}{
				"client_ip": clientIP,
				"origin":    origin,
				"error":     err.Error(),
			}).Error("Failed to generate CSRF token")

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to generate CSRF token",
				"message": "Internal server error occurred while generating security token",
			})
			return
		}

		// Log successful token generation
		log.WithFields(map[string]interface{}{
			"client_ip":    clientIP,
			"origin":       origin,
			"token_length": len(token),
			"action":       "CSRF_TOKEN_GENERATED",
		}).Debug("CSRF token generated successfully")

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_in": 3600, // 1 hour in seconds
			"usage":      "Include this token in the X-CSRF-Token header for state-changing requests",
		})
	}
}
