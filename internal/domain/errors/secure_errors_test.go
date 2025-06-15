package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureErrorHandling_SanitizeError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedClient string
		expectedServer string
		expectedCode   string
	}{
		{
			name:           "should sanitize stack trace from FFmpeg error",
			err:            FFmpegFailed(errors.New("exit status 1: ffmpeg: error while opening input file '/etc/passwd'")),
			expectedClient: "Video processing failed. Please check your input files and try again.",
			expectedServer: "FFmpeg execution failed: exit status 1: ffmpeg: error while opening input file '/etc/passwd'",
			expectedCode:   ErrCodeFFmpegFailed,
		},
		{
			name:           "should sanitize file path from file not found error",
			err:            FileNotFound("/home/user/.ssh/id_rsa"),
			expectedClient: "The requested file could not be found. Please verify the file exists.",
			expectedServer: "File not found: /home/user/.ssh/id_rsa",
			expectedCode:   ErrCodeFileNotFound,
		},
		{
			name:           "should sanitize download URL from download error",
			err:            DownloadFailed("http://internal.server/admin/secrets", errors.New("connection refused")),
			expectedClient: "Failed to download the specified resource. Please check the URL and try again.",
			expectedServer: "Failed to download from http://internal.server/admin/secrets: connection refused",
			expectedCode:   ErrCodeDownloadFailed,
		},
		{
			name:           "should sanitize transcription error details",
			err:            TranscriptionFailed(errors.New("whisper model failed: /internal/models/whisper.bin not found")),
			expectedClient: "Audio transcription failed. Please ensure the audio file is valid.",
			expectedServer: "Audio transcription failed: whisper model failed: /internal/models/whisper.bin not found",
			expectedCode:   ErrCodeTranscriptionFailed,
		},
		{
			name:           "should sanitize storage error with sensitive paths",
			err:            StorageFailed(errors.New("permission denied: /var/lib/mysql/data")),
			expectedClient: "Storage operation failed. Please try again later.",
			expectedServer: "Storage operation failed: permission denied: /var/lib/mysql/data",
			expectedCode:   ErrCodeStorageFailed,
		},
		{
			name:           "should sanitize internal error with stack trace",
			err:            InternalError(errors.New("database connection failed: mysql://user:password@localhost/db")),
			expectedClient: "An internal error occurred. Please try again later or contact support.",
			expectedServer: "Internal server error: database connection failed: mysql://user:password@localhost/db",
			expectedCode:   ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should fail - SanitizeForClient doesn't exist yet
			clientMsg := SanitizeForClient(tt.err)
			assert.Equal(t, tt.expectedClient, clientMsg)

			// This should fail - GetServerDetails doesn't exist yet
			serverMsg := GetServerDetails(tt.err)
			assert.Equal(t, tt.expectedServer, serverMsg)

			// This should fail - GetErrorCode doesn't exist yet
			if vpe, ok := tt.err.(*VideoProcessingError); ok {
				code := GetErrorCode(vpe)
				assert.Equal(t, tt.expectedCode, code)
			}
		})
	}
}

func TestSecureErrorHandling_NoStackTraceInClientResponse(t *testing.T) {
	// Create error with potential stack trace information
	originalErr := errors.New("panic: runtime error: invalid memory address or nil pointer dereference\n\tgoroutine 1 [running]:\n\tmain.processVideo()\n\t\t/app/internal/services/video.go:123 +0x1a4")
	err := InternalError(originalErr)

	// This should fail - SanitizeForClient doesn't exist yet
	clientMsg := SanitizeForClient(err)
	
	// Client message should not contain stack trace
	assert.NotContains(t, clientMsg, "goroutine")
	assert.NotContains(t, clientMsg, "runtime error")
	assert.NotContains(t, clientMsg, "/app/internal")
	assert.NotContains(t, clientMsg, "panic:")
	assert.Equal(t, "An internal error occurred. Please try again later or contact support.", clientMsg)
}

func TestSecureErrorHandling_NoSensitivePathsInClientResponse(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		sensitivePaths []string
	}{
		{
			name: "should not expose system paths",
			err:  FileNotFound("/etc/passwd"),
			sensitivePaths: []string{"/etc/passwd", "/etc/", "passwd"},
		},
		{
			name: "should not expose user paths",
			err:  FileNotFound("/home/user/.ssh/id_rsa"),
			sensitivePaths: []string{"/home/user", ".ssh", "id_rsa"},
		},
		{
			name: "should not expose application paths",
			err:  StorageFailed(errors.New("failed to access /var/lib/videocraft/secrets")),
			sensitivePaths: []string{"/var/lib", "videocraft", "secrets"},
		},
		{
			name: "should not expose internal URLs",
			err:  DownloadFailed("http://internal.service/admin", errors.New("timeout")),
			sensitivePaths: []string{"internal.service", "admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should fail - SanitizeForClient doesn't exist yet
			clientMsg := SanitizeForClient(tt.err)
			
			for _, sensitivePath := range tt.sensitivePaths {
				assert.NotContains(t, clientMsg, sensitivePath, 
					"Client message should not contain sensitive path: %s", sensitivePath)
			}
		})
	}
}

func TestSecureErrorHandling_StandardizedErrorCodes(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "FFmpeg error should have standardized code",
			err:          FFmpegFailed(errors.New("any error")),
			expectedCode: ErrCodeFFmpegFailed,
		},
		{
			name:         "File not found should have standardized code",
			err:          FileNotFound("any-file"),
			expectedCode: ErrCodeFileNotFound,
		},
		{
			name:         "Download error should have standardized code",
			err:          DownloadFailed("any-url", errors.New("any error")),
			expectedCode: ErrCodeDownloadFailed,
		},
		{
			name:         "Transcription error should have standardized code",
			err:          TranscriptionFailed(errors.New("any error")),
			expectedCode: ErrCodeTranscriptionFailed,
		},
		{
			name:         "Storage error should have standardized code",
			err:          StorageFailed(errors.New("any error")),
			expectedCode: ErrCodeStorageFailed,
		},
		{
			name:         "Internal error should have standardized code",
			err:          InternalError(errors.New("any error")),
			expectedCode: ErrCodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should fail - GetErrorCode doesn't exist yet
			if vpe, ok := tt.err.(*VideoProcessingError); ok {
				code := GetErrorCode(vpe)
				assert.Equal(t, tt.expectedCode, code)
			} else {
				t.Errorf("Expected VideoProcessingError, got %T", tt.err)
			}
		})
	}
}

func TestSecureErrorHandling_DetailedServerLogging(t *testing.T) {
	// Test that server-side logging includes full details
	originalErr := errors.New("database connection failed: mysql://user:password@db:3306/videocraft")
	err := InternalError(originalErr)

	// This should fail - GetServerDetails doesn't exist yet
	serverDetails := GetServerDetails(err)
	
	// Server details should include full error information for debugging
	assert.Contains(t, serverDetails, "database connection failed")
	assert.Contains(t, serverDetails, "mysql://user:password@db:3306/videocraft")
	
	// This should fail - GetLogContext doesn't exist yet
	logContext := GetLogContext(err)
	require.NotNil(t, logContext)
	assert.Contains(t, logContext, "error_type")
	assert.Contains(t, logContext, "error_code")
	assert.Contains(t, logContext, "original_error")
}

func TestSecureErrorHandling_ErrorResponseStructure(t *testing.T) {
	err := FFmpegFailed(errors.New("ffmpeg execution failed"))
	
	// This should fail - ToClientResponse doesn't exist yet
	clientResponse := ToClientResponse(err)
	
	// Should have standardized structure
	assert.Contains(t, clientResponse, "error")
	assert.Contains(t, clientResponse, "code")
	assert.NotContains(t, clientResponse, "details") // No sensitive details
	assert.NotContains(t, clientResponse, "original_error")
	assert.NotContains(t, clientResponse, "stack_trace")
}

func TestSecureErrorHandling_SecurityViolationLogging(t *testing.T) {
	// Test errors that might indicate security issues
	securityErrors := []error{
		FileNotFound("/etc/passwd"),
		FileNotFound("/etc/shadow"),
		FileNotFound("/root/.ssh/id_rsa"),
		DownloadFailed("file:///etc/passwd", errors.New("access denied")),
		DownloadFailed("http://localhost:22/", errors.New("connection refused")),
		StorageFailed(errors.New("permission denied: /var/lib/mysql")),
	}

	for _, err := range securityErrors {
		// This should fail - IsSecuritySensitive doesn't exist yet
		isSensitive := IsSecuritySensitive(err)
		assert.True(t, isSensitive, "Error should be marked as security sensitive: %v", err)
		
		// This should fail - LogSecurityEvent doesn't exist yet
		logEntry := LogSecurityEvent(err)
		assert.Contains(t, logEntry, "SECURITY_SENSITIVE")
		assert.Contains(t, logEntry, "error_type")
	}
}