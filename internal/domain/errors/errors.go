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
	ErrCodeInvalidInput      = "INVALID_INPUT"
	ErrCodeFileNotFound      = "FILE_NOT_FOUND"
	ErrCodeFFmpegFailed      = "FFMPEG_FAILED"
	ErrCodeTranscriptionFailed = "TRANSCRIPTION_FAILED"
	ErrCodeJobNotFound       = "JOB_NOT_FOUND"
	ErrCodeStorageFailed     = "STORAGE_FAILED"
	ErrCodeDownloadFailed    = "DOWNLOAD_FAILED"
	ErrCodeTimeout           = "TIMEOUT"
	ErrCodeInternalError     = "INTERNAL_ERROR"
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
			"url": url,
			"original_error": err.Error(),
		})
}

func Timeout(operation string, timeout string) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeTimeout, 
		fmt.Sprintf("Operation %s timed out after %s", operation, timeout), 
		map[string]interface{}{
			"operation": operation,
			"timeout": timeout,
		})
}

func InternalError(err error) *VideoProcessingError {
	return NewVideoProcessingError(ErrCodeInternalError, 
		fmt.Sprintf("Internal server error: %v", err), 
		map[string]interface{}{"original_error": err.Error()})
}