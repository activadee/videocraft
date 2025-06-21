package errors

import "fmt"

// Custom error types for the application

type VideoProcessingError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e VideoProcessingError) Error() string {
	return e.Message
}

func NewVideoProcessingError(code, message string, details map[string]interface{}) *VideoProcessingError {
	return &VideoProcessingError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Predefined error codes
const (
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeFileNotFound        = "FILE_NOT_FOUND"
	ErrCodeFFmpegFailed        = "FFMPEG_FAILED"
	ErrCodeTranscriptionFailed = "TRANSCRIPTION_FAILED"
	ErrCodeJobNotFound         = "JOB_NOT_FOUND"
	ErrCodeStorageFailed       = "STORAGE_FAILED"
	ErrCodeDownloadFailed      = "DOWNLOAD_FAILED"
	ErrCodeTimeout             = "TIMEOUT"
	ErrCodeInternalError       = "INTERNAL_ERROR"
)

// Error constructors
func InvalidInput(message string) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeInvalidInput, message, nil)
}

func FileNotFound(filename string) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeFileNotFound,
		fmt.Sprintf("File not found: %s", filename),
		map[string]interface{}{"filename": filename})
}

func FFmpegFailed(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeFFmpegFailed,
		fmt.Sprintf("FFmpeg execution failed: %v", err),
		map[string]interface{}{"original_error": err.Error()})
}

func TranscriptionFailed(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeTranscriptionFailed,
		fmt.Sprintf("Audio transcription failed: %v", err),
		map[string]interface{}{"original_error": err.Error()})
}

func JobNotFound(jobID string) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeJobNotFound,
		fmt.Sprintf("Job not found: %s", jobID),
		map[string]interface{}{"job_id": jobID})
}

func StorageFailed(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeStorageFailed,
		fmt.Sprintf("Storage operation failed: %v", err),
		map[string]interface{}{"original_error": err.Error()})
}

func DownloadFailed(url string, err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeDownloadFailed,
		fmt.Sprintf("Failed to download from %s: %v", url, err),
		map[string]interface{}{
			"url":            url,
			"original_error": err.Error(),
		})
}

func Timeout(operation string, timeout string) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeTimeout,
		fmt.Sprintf("Operation %s timed out after %s", operation, timeout),
		map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		})
}

func InternalError(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeInternalError,
		fmt.Sprintf("Internal server error: %v", err),
		map[string]interface{}{"original_error": err.Error()})
}

func ProcessingFailed(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeInternalError,
		fmt.Sprintf("Processing failed: %v", err),
		map[string]interface{}{"original_error": err.Error()})
}

// Secure error handling functions

// Client-safe error messages with helpful context
var clientErrorMessages = map[string]string{
	ErrCodeFFmpegFailed:        "Video processing failed. Please check your input files and try again.",
	ErrCodeFileNotFound:        "The requested file could not be found. Please verify the file exists.",
	ErrCodeDownloadFailed:      "Failed to download the specified resource. Please check the URL and try again.",
	ErrCodeTranscriptionFailed: "Audio transcription failed. Please ensure the audio file is valid.",
	ErrCodeStorageFailed:       "Storage operation failed. Please try again later.",
	ErrCodeTimeout:             "The request timed out. Please try again with a smaller file or shorter duration.",
	ErrCodeInvalidInput:        "Invalid request format",
	ErrCodeJobNotFound:         "The requested job could not be found. It may have been completed or removed.",
	ErrCodeInternalError:       "An internal error occurred. Please try again later or contact support.",
}

// SanitizeForClient returns a user-friendly error message safe for client consumption
func SanitizeForClient(err error) string {
	if vpe, ok := err.(*VideoProcessingError); ok {
		if message, exists := clientErrorMessages[vpe.Code]; exists {
			return message
		}
		return "An unexpected error occurred. Please try again later."
	}

	// For non-domain errors, provide generic but helpful message
	return "An error occurred while processing your request. Please try again later."
}

// GetServerDetails returns detailed error information for server-side logging
func GetServerDetails(err error) string {
	if vpe, ok := err.(*VideoProcessingError); ok {
		return vpe.Message
	}
	return err.Error()
}

// GetErrorCode returns the standardized error code
func GetErrorCode(err *VideoProcessingError) string {
	return err.Code
}

// GetLogContext returns structured context for logging
func GetLogContext(err error) map[string]interface{} {
	logContext := make(map[string]interface{})

	if vpe, ok := err.(*VideoProcessingError); ok {
		logContext["error_type"] = "VideoProcessingError"
		logContext["error_code"] = vpe.Code
		logContext["original_error"] = vpe.Message

		// Add details if available
		if vpe.Details != nil {
			logContext["error_details"] = vpe.Details
		}
	} else {
		logContext["error_type"] = "UnknownError"
		logContext["error_code"] = "UNKNOWN"
		logContext["original_error"] = err.Error()
	}

	return logContext
}

// ToClientResponse returns a safe client response structure
func ToClientResponse(err error) map[string]interface{} {
	response := make(map[string]interface{})

	if vpe, ok := err.(*VideoProcessingError); ok {
		response["error"] = SanitizeForClient(err)
		response["code"] = vpe.Code
	} else {
		response["error"] = "An error occurred"
		response["code"] = "UNKNOWN_ERROR"
	}

	return response
}

// Comprehensive security-sensitive patterns grouped by category
var (
	sensitiveFilePaths = []string{
		"/etc/", "/root/", "/home/", "/var/lib/", "/var/log/", "/var/run/",
		"/usr/local/", "/opt/", "/boot/", "/sys/", "/proc/", "/dev/",
		".ssh/", ".config/", ".env", ".git/", ".svn/", ".htaccess",
		"passwd", "shadow", "sudoers", "hosts", "fstab", "crontab",
	}

	sensitiveURLSchemes = []string{
		"mysql://", "postgres://", "postgresql://", "mongodb://", "redis://",
		"ldap://", "ldaps://", "ftp://", "sftp://", "file://", "jdbc:",
		"data:", "javascript:", "vbscript:",
	}

	sensitiveNetworkTargets = []string{
		"localhost:", "127.0.0.1:", "0.0.0.0:", "::1:",
		"internal.", "corp.", "intranet.", "local.",
		"admin.", "test.", "staging.", "dev.",
	}

	sensitiveKeywords = []string{
		"password", "passwd", "secret", "token", "key", "credential",
		"private", "confidential", "auth", "session", "cookie",
		"api_key", "access_token", "refresh_token", "jwt",
	}
)

// IsSecuritySensitive determines if an error contains security-sensitive information
func IsSecuritySensitive(err error) bool {
	if vpe, ok := err.(*VideoProcessingError); ok {
		message := vpe.Message

		// Check all pattern categories
		if containsAnyPattern(message, sensitiveFilePaths) ||
			containsAnyPattern(message, sensitiveURLSchemes) ||
			containsAnyPattern(message, sensitiveNetworkTargets) ||
			containsAnyPattern(message, sensitiveKeywords) {
			return true
		}

		// Check error details for sensitive information
		if vpe.Details != nil {
			for key, value := range vpe.Details {
				// Check if key itself is sensitive
				if containsAnyPattern(key, sensitiveKeywords) {
					return true
				}

				// Check string values for sensitive patterns
				if str, ok := value.(string); ok {
					if containsAnyPattern(str, sensitiveFilePaths) ||
						containsAnyPattern(str, sensitiveURLSchemes) ||
						containsAnyPattern(str, sensitiveNetworkTargets) ||
						containsAnyPattern(str, sensitiveKeywords) {
						return true
					}
				}
			}
		}
	}

	return false
}

// Helper function to check if text contains any pattern from a list
func containsAnyPattern(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if contains(text, pattern) {
			return true
		}
	}
	return false
}

// LogSecurityEvent returns structured logging information for security-sensitive errors
func LogSecurityEvent(err error) map[string]interface{} {
	logEntry := GetLogContext(err)
	logEntry["SECURITY_SENSITIVE"] = true
	logEntry["alert_level"] = "HIGH"

	if vpe, ok := err.(*VideoProcessingError); ok {
		logEntry["error_type"] = vpe.Code
	}

	return logEntry
}

// Helper function for efficient case-insensitive string contains
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	// Convert to lowercase for case-insensitive comparison
	sLower := toLower(s)
	substrLower := toLower(substr)

	// Use optimized Boyer-Moore-style search for better performance
	return indexOfSubstring(sLower, substrLower) >= 0
}

// Optimized substring search using simple but efficient algorithm
func indexOfSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(s) < len(substr) {
		return -1
	}

	// Simple but efficient substring search
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Simple lowercase conversion (ASCII only for performance)
func toLower(s string) string {
	result := make([]byte, len(s))
	for i, b := range []byte(s) {
		if b >= 'A' && b <= 'Z' {
			result[i] = b + 32
		} else {
			result[i] = b
		}
	}
	return string(result)
}
