package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/pkg/logger"
)

// ErrorHandler provides backward compatibility - now uses SecureErrorHandler
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return SecureErrorHandler(log)
}

// Legacy error handler - deprecated, use SecureErrorHandler instead
func LegacyErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors if any occurred
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Extract request context for security logging
			requestContext := map[string]interface{}{
				"client_ip":   c.ClientIP(),
				"user_agent":  c.Request.UserAgent(),
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"request_id":  c.GetHeader("X-Request-ID"),
				"timestamp":   time.Now(),
			}

			// Check if this is a security-sensitive error
			if errors.IsSecuritySensitive(err.Err) {
				// Log security event
				securityLogEntry := errors.LogSecurityEvent(err.Err)
				for key, value := range requestContext {
					securityLogEntry[key] = value
				}
				log.WithFields(securityLogEntry).Error("SECURITY_VIOLATION: Security-sensitive error detected")
			} else {
				// Normal logging
				log.WithField("error", err.Error()).Error("Request error")
			}

			// Check if it's our custom error type
			if vpe, ok := err.Err.(*errors.VideoProcessingError); ok {
				status := getStatusFromErrorCode(vpe.Code)
				
				// Create secure response
				response := gin.H{
					"error":      errors.SanitizeForClient(err.Err),
					"code":       vpe.Code,
					"request_id": c.GetHeader("X-Request-ID"),
					"timestamp":  time.Now().Format(time.RFC3339),
				}
				
				c.JSON(status, response)
				return
			}

			// Generic error handling with secure response
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "Internal server error occurred",
				"code":       "INTERNAL_ERROR",
				"request_id": c.GetHeader("X-Request-ID"),
				"timestamp":  time.Now().Format(time.RFC3339),
			})
		}
	}
}

func getStatusFromErrorCode(code string) int {
	switch code {
	case errors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case errors.ErrCodeFileNotFound:
		return http.StatusNotFound
	case errors.ErrCodeJobNotFound:
		return http.StatusNotFound
	case errors.ErrCodeTimeout:
		return http.StatusRequestTimeout
	case errors.ErrCodeFFmpegFailed:
		return http.StatusUnprocessableEntity
	case errors.ErrCodeTranscriptionFailed:
		return http.StatusUnprocessableEntity
	case errors.ErrCodeDownloadFailed:
		return http.StatusBadGateway
	case errors.ErrCodeStorageFailed:
		return http.StatusInsufficientStorage
	default:
		return http.StatusInternalServerError
	}
}
