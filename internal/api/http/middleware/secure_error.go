package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// Threat level constants
const (
	ThreatLevelHigh    = "HIGH"
	ThreatLevelMedium  = "MEDIUM"
	ThreatLevelLow     = "LOW"
	ThreatLevelUnknown = "UNKNOWN"
)

// SecureErrorHandler provides secure error handling middleware
func SecureErrorHandler(log logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Set up panic recovery
		defer func() {
			if recovered := recover(); recovered != nil {
				handlePanicRecovery(c, recovered, log)
			}
		}()

		// Process the request
		c.Next()

		// Handle errors if any occurred
		if len(c.Errors) > 0 {
			handleRequestErrors(c, c.Errors, log)
		}
	})
}

// handlePanicRecovery handles panic situations with secure error responses
func handlePanicRecovery(c *gin.Context, recovered interface{}, log logger.Logger) {
	// Extract request context for logging
	requestContext := extractRequestContext(c)

	// Create internal error from panic
	var err error
	switch x := recovered.(type) {
	case string:
		err = domainErrors.InternalError(errors.New(x))
	case error:
		err = domainErrors.InternalError(x)
	default:
		err = domainErrors.InternalError(errors.New("unknown panic occurred"))
	}

	// Log the panic with full details server-side
	logSecuritySensitiveError(err, requestContext, log)

	// Return sanitized error to client
	response := createSecureErrorResponse(err, c)
	c.JSON(http.StatusInternalServerError, response)
	c.Abort()
}

// handleRequestErrors processes request errors with security considerations
func handleRequestErrors(c *gin.Context, ginErrors []*gin.Error, log logger.Logger) {
	if len(ginErrors) == 0 {
		return
	}

	// Check if response was already written
	if c.Writer.Written() {
		return
	}

	// Get the last error
	lastError := ginErrors[len(ginErrors)-1]
	err := lastError.Err

	// Convert JSON errors to domain errors
	if isJSONError(err) {
		err = domainErrors.InvalidInput("Invalid request format")
	}

	// Extract request context
	requestContext := extractRequestContext(c)

	// Check if this is a security-sensitive error
	if domainErrors.IsSecuritySensitive(err) {
		logSecuritySensitiveError(err, requestContext, log)
	} else {
		// Log normally for non-sensitive errors
		logContext := domainErrors.GetLogContext(err)
		for key, value := range requestContext {
			logContext[key] = value
		}
		log.WithFields(logContext).Error("Request error occurred")
	}

	// Determine HTTP status code
	statusCode := getStatusCodeFromError(err)

	// Create secure response
	response := createSecureErrorResponse(err, c)

	c.JSON(statusCode, response)
	c.Abort()
}

// extractRequestContext gathers request information for logging
func extractRequestContext(c *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"request_id": c.GetHeader("X-Request-ID"),
		"timestamp":  time.Now(),
	}
}

// logSecuritySensitiveError logs security-sensitive errors with enhanced detail
func logSecuritySensitiveError(err error, requestContext map[string]interface{}, log logger.Logger) {
	// Get security event log entry
	securityLogEntry := domainErrors.LogSecurityEvent(err)

	// Add request context efficiently
	for key, value := range requestContext {
		securityLogEntry[key] = value
	}

	// Add server-side error details
	securityLogEntry["server_error_details"] = domainErrors.GetServerDetails(err)

	// Add stack trace if available (server-side only)
	// Limit stack trace size to prevent log bloat
	if stack := debug.Stack(); stack != nil {
		stackStr := string(stack)
		if len(stackStr) > 2048 { // Limit to 2KB
			stackStr = stackStr[:2048] + "...[truncated]"
		}
		securityLogEntry["stack_trace"] = stackStr
	}

	// Add threat assessment
	securityLogEntry["threat_level"] = assessThreatLevel(err)
	securityLogEntry["recommended_action"] = getRecommendedAction(err)

	log.WithFields(securityLogEntry).Error("SECURITY_VIOLATION: Security-sensitive error detected")
}

// assessThreatLevel provides automated threat level assessment
func assessThreatLevel(err error) string {
	if vpe, ok := err.(*domainErrors.VideoProcessingError); ok {
		message := vpe.Message

		// High threat indicators
		highThreatPatterns := []string{
			"/etc/passwd", "/etc/shadow", ".ssh/id_rsa", "/root/",
			"admin", "secret", "password", "credential",
		}

		for _, pattern := range highThreatPatterns {
			if strings.Contains(strings.ToLower(message), pattern) {
				return ThreatLevelHigh
			}
		}

		// Medium threat indicators
		mediumThreatPatterns := []string{
			"/etc/", "/var/", "localhost:", "internal.",
		}

		for _, pattern := range mediumThreatPatterns {
			if strings.Contains(strings.ToLower(message), pattern) {
				return ThreatLevelMedium
			}
		}

		return ThreatLevelLow
	}

	return ThreatLevelUnknown
}

// getRecommendedAction provides automated response recommendations
func getRecommendedAction(err error) string {
	threatLevel := assessThreatLevel(err)

	switch threatLevel {
	case ThreatLevelHigh:
		return "IMMEDIATE_REVIEW_REQUIRED - Check for potential intrusion attempts"
	case ThreatLevelMedium:
		return "MONITOR_CLOSELY - Review access patterns and log retention"
	case ThreatLevelLow:
		return "LOG_AND_MONITOR - Standard security logging"
	default:
		return "INVESTIGATE - Unknown threat pattern detected"
	}
}

// createSecureErrorResponse creates a sanitized error response for clients
func createSecureErrorResponse(err error, c *gin.Context) map[string]interface{} {
	// Start with basic secure response
	response := domainErrors.ToClientResponse(err)

	// Generate or get request ID for tracking
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	// Add safe metadata for client debugging
	response["request_id"] = requestID
	response["timestamp"] = time.Now().Format(time.RFC3339)
	response["success"] = false

	// Add helpful links for common errors
	if errorCode, ok := response["code"].(string); ok {
		response["help_url"] = generateHelpURL(errorCode)
	}

	// Ensure no sensitive fields are included (defensive programming)
	sensitiveFields := []string{
		"details", "original_error", "stack_trace", "internal_message",
		"server_details", "file_path", "url", "credentials", "password",
		"secret", "token", "key", "session", "cookie",
	}

	for _, field := range sensitiveFields {
		delete(response, field)
	}

	return response
}

// generateRequestID creates a simple request ID for tracking
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// generateHelpURL provides helpful documentation links for error codes
func generateHelpURL(errorCode string) string {
	baseURL := "https://docs.videocraft.io/errors/"

	switch errorCode {
	case domainErrors.ErrCodeFFmpegFailed:
		return baseURL + "video-processing"
	case domainErrors.ErrCodeFileNotFound:
		return baseURL + "file-not-found"
	case domainErrors.ErrCodeDownloadFailed:
		return baseURL + "download-issues"
	case domainErrors.ErrCodeTranscriptionFailed:
		return baseURL + "transcription-issues"
	case domainErrors.ErrCodeInvalidInput:
		return baseURL + "input-validation"
	case domainErrors.ErrCodeTimeout:
		return baseURL + "timeouts"
	default:
		return baseURL + "general"
	}
}

// getStatusCodeFromError determines appropriate HTTP status code
func getStatusCodeFromError(err error) int {
	// Handle JSON binding errors (common client errors)
	if isJSONError(err) {
		return http.StatusBadRequest
	}

	// Handle domain errors
	if vpe, ok := err.(*domainErrors.VideoProcessingError); ok {
		return getSecureStatusFromErrorCode(vpe.Code)
	}

	// Default to internal server error
	return http.StatusInternalServerError
}

// isJSONError checks if error is from JSON parsing/binding
func isJSONError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "json:") ||
		strings.Contains(errStr, "cannot unmarshal") ||
		strings.Contains(errStr, "invalid character") ||
		strings.Contains(errStr, "unexpected end of JSON input")
}

// getSecureStatusFromErrorCode maps error codes to HTTP status codes
func getSecureStatusFromErrorCode(code string) int {
	switch code {
	case domainErrors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case domainErrors.ErrCodeFileNotFound:
		return http.StatusNotFound
	case domainErrors.ErrCodeJobNotFound:
		return http.StatusNotFound
	case domainErrors.ErrCodeTimeout:
		return http.StatusRequestTimeout
	case domainErrors.ErrCodeFFmpegFailed:
		return http.StatusUnprocessableEntity
	case domainErrors.ErrCodeTranscriptionFailed:
		return http.StatusUnprocessableEntity
	case domainErrors.ErrCodeDownloadFailed:
		return http.StatusBadGateway
	case domainErrors.ErrCodeStorageFailed:
		return http.StatusInsufficientStorage
	default:
		return http.StatusInternalServerError
	}
}
