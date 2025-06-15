package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/pkg/logger"
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
		"client_ip":   c.ClientIP(),
		"user_agent":  c.Request.UserAgent(),
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"request_id":  c.GetHeader("X-Request-ID"),
		"timestamp":   time.Now(),
	}
}

// logSecuritySensitiveError logs security-sensitive errors with enhanced detail
func logSecuritySensitiveError(err error, requestContext map[string]interface{}, log logger.Logger) {
	// Get security event log entry
	securityLogEntry := domainErrors.LogSecurityEvent(err)
	
	// Add request context
	for key, value := range requestContext {
		securityLogEntry[key] = value
	}
	
	// Add server-side error details
	securityLogEntry["server_error_details"] = domainErrors.GetServerDetails(err)
	
	// Add stack trace if available (server-side only)
	if stack := debug.Stack(); stack != nil {
		securityLogEntry["stack_trace"] = string(stack)
	}
	
	log.WithFields(securityLogEntry).Error("SECURITY_VIOLATION: Security-sensitive error detected")
}

// createSecureErrorResponse creates a sanitized error response for clients
func createSecureErrorResponse(err error, c *gin.Context) map[string]interface{} {
	// Start with basic secure response
	response := domainErrors.ToClientResponse(err)
	
	// Add request metadata (safe information only)
	response["request_id"] = c.GetHeader("X-Request-ID")
	response["timestamp"] = time.Now().Format(time.RFC3339)
	
	// Ensure no sensitive fields are included
	delete(response, "details")
	delete(response, "original_error")
	delete(response, "stack_trace")
	delete(response, "internal_message")
	delete(response, "server_details")
	
	return response
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

