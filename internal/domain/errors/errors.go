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

// Secure error handling functions

// SanitizeForClient returns a generic error message safe for client consumption
func SanitizeForClient(err error) string {
	if vpe, ok := err.(*VideoProcessingError); ok {
		switch vpe.Code {
		case ErrCodeFFmpegFailed:
			return "Video processing failed"
		case ErrCodeFileNotFound:
			return "File not found"
		case ErrCodeDownloadFailed:
			return "Download failed"
		case ErrCodeTranscriptionFailed:
			return "Transcription failed"
		case ErrCodeStorageFailed:
			return "Storage operation failed"
		case ErrCodeTimeout:
			return "Request timeout"
		case ErrCodeInvalidInput:
			return "Invalid input provided"
		case ErrCodeJobNotFound:
			return "Job not found"
		case ErrCodeInternalError:
			return "Internal server error occurred"
		default:
			return "An error occurred"
		}
	}
	
	// For non-domain errors, return generic message
	return "An error occurred"
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

// IsSecuritySensitive determines if an error contains security-sensitive information
func IsSecuritySensitive(err error) bool {
	if vpe, ok := err.(*VideoProcessingError); ok {
		// Check for sensitive patterns in error message
		sensitivePatterns := []string{
			"/etc/",
			"/root/",
			"/home/",
			"/var/lib/",
			".ssh/",
			"passwd",
			"shadow",
			"mysql://",
			"postgres://",
			"mongodb://",
			"localhost:",
			"127.0.0.1:",
			"internal.",
			"admin",
			"secret",
			"private",
		}
		
		message := vpe.Message
		for _, pattern := range sensitivePatterns {
			if contains(message, pattern) {
				return true
			}
		}
		
		// Check error details for sensitive information
		if vpe.Details != nil {
			for key, value := range vpe.Details {
				if contains(key, "password") || contains(key, "secret") || contains(key, "token") {
					return true
				}
				if str, ok := value.(string); ok {
					for _, pattern := range sensitivePatterns {
						if contains(str, pattern) {
							return true
						}
					}
				}
			}
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

// Helper function for case-insensitive string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && 
		(s[:len(substr)] == substr || 
		 s[len(s)-len(substr):] == substr || 
		 indexOfSubstring(s, substr) >= 0))
}

// Simple substring search
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
