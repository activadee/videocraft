package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainErrors "github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/pkg/logger"
)

func TestSecureErrorHandler_NoStackTraceExposure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		setupError     func(*gin.Context)
		expectedStatus int
		shouldNotContain []string
	}{
		{
			name: "should not expose stack trace in panic recovery",
			setupError: func(c *gin.Context) {
				panic("runtime error: invalid memory address or nil pointer dereference\n\tgoroutine 1 [running]:\n\tmain.processVideo()\n\t\t/app/internal/services/video.go:123 +0x1a4")
			},
			expectedStatus: http.StatusInternalServerError,
			shouldNotContain: []string{
				"goroutine",
				"runtime error",
				"/app/internal",
				"panic:",
				"processVideo",
				"+0x1a4",
			},
		},
		{
			name: "should not expose sensitive file paths",
			setupError: func(c *gin.Context) {
				err := domainErrors.FileNotFound("/etc/passwd")
				c.Error(err)
			},
			expectedStatus: http.StatusNotFound,
			shouldNotContain: []string{
				"/etc/passwd",
				"/etc/",
				"passwd",
			},
		},
		{
			name: "should not expose internal URLs",
			setupError: func(c *gin.Context) {
				err := domainErrors.DownloadFailed("http://internal.database:3306/admin", errors.New("connection refused"))
				c.Error(err)
			},
			expectedStatus: http.StatusBadGateway,
			shouldNotContain: []string{
				"internal.database",
				"3306",
				"admin",
				"connection refused",
			},
		},
		{
			name: "should not expose database credentials",
			setupError: func(c *gin.Context) {
				err := domainErrors.InternalError(errors.New("database connection failed: mysql://user:password@localhost/db"))
				c.Error(err)
			},
			expectedStatus: http.StatusInternalServerError,
			shouldNotContain: []string{
				"mysql://",
				"user:password",
				"localhost/db",
				"database connection failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with secure error handler
			router := gin.New()
			
			// Use SecureErrorHandler with noop logger
			router.Use(SecureErrorHandler(newNoopLogger()))
			
			// Add test route that triggers error
			router.GET("/test", func(c *gin.Context) {
				tt.setupError(c)
				// Don't send success response if there's an error
				if len(c.Errors) > 0 {
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Execute request
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Verify response body doesn't contain sensitive information
			body := w.Body.String()
			for _, sensitiveInfo := range tt.shouldNotContain {
				assert.NotContains(t, body, sensitiveInfo, 
					"Response should not contain sensitive information: %s", sensitiveInfo)
			}

			// Verify response has standardized format
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Response should be valid JSON")
			
			assert.Contains(t, response, "error", "Response should contain error field")
			assert.Contains(t, response, "code", "Response should contain code field")
			assert.NotContains(t, response, "details", "Response should not contain details field")
			assert.NotContains(t, response, "stack_trace", "Response should not contain stack_trace field")
		})
	}
}

func TestSecureErrorHandler_StandardizedErrorCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
		expectedError  string
	}{
		{
			name:           "FFmpeg error should be sanitized",
			error:          domainErrors.FFmpegFailed(errors.New("ffmpeg failed with sensitive path /etc/passwd")),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "FFMPEG_FAILED",
			expectedError:  "Video processing failed. Please check your input files and try again.",
		},
		{
			name:           "File not found should be sanitized",
			error:          domainErrors.FileNotFound("/home/user/.ssh/id_rsa"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "FILE_NOT_FOUND", 
			expectedError:  "The requested file could not be found. Please verify the file exists.",
		},
		{
			name:           "Download error should be sanitized",
			error:          domainErrors.DownloadFailed("http://admin.internal/secrets", errors.New("forbidden")),
			expectedStatus: http.StatusBadGateway,
			expectedCode:   "DOWNLOAD_FAILED",
			expectedError:  "Failed to download the specified resource. Please check the URL and try again.",
		},
		{
			name:           "Transcription error should be sanitized",
			error:          domainErrors.TranscriptionFailed(errors.New("whisper model not found: /internal/models/whisper.bin")),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "TRANSCRIPTION_FAILED",
			expectedError:  "Audio transcription failed. Please ensure the audio file is valid.",
		},
		{
			name:           "Storage error should be sanitized",
			error:          domainErrors.StorageFailed(errors.New("permission denied: /var/lib/mysql/data")),
			expectedStatus: http.StatusInsufficientStorage,
			expectedCode:   "STORAGE_FAILED",
			expectedError:  "Storage operation failed. Please try again later.",
		},
		{
			name:           "Internal error should be sanitized",
			error:          domainErrors.InternalError(errors.New("database panic: connection string mysql://user:pass@host/db")),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedError:  "An internal error occurred. Please try again later or contact support.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create router with secure error handler
			router := gin.New()
			
			// Use SecureErrorHandler with noop logger
			router.Use(SecureErrorHandler(newNoopLogger()))
			
			// Add test route that triggers specific error
			router.GET("/test", func(c *gin.Context) {
				c.Error(tt.error)
			})

			// Execute request
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedCode, response["code"])
			assert.Equal(t, tt.expectedError, response["error"])
		})
	}
}

func TestSecureErrorHandler_SecurityEventLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create a test logger that captures log entries
	var logEntries []map[string]interface{}
	testLogger := &testLogger{
		entries: &logEntries,
	}
	
	// Security-sensitive errors that should trigger special logging
	securityErrors := []struct {
		name  string
		error error
	}{
		{
			name:  "sensitive file access",
			error: domainErrors.FileNotFound("/etc/passwd"),
		},
		{
			name:  "SSH key access",
			error: domainErrors.FileNotFound("/root/.ssh/id_rsa"),
		},
		{
			name:  "internal service probe",
			error: domainErrors.DownloadFailed("http://localhost:22/", errors.New("connection refused")),
		},
		{
			name:  "database directory access",
			error: domainErrors.StorageFailed(errors.New("permission denied: /var/lib/mysql")),
		},
	}

	for _, tt := range securityErrors {
		t.Run(tt.name, func(t *testing.T) {
			// Reset log entries
			logEntries = []map[string]interface{}{}
			
			// Create router with secure error handler
			router := gin.New()
			
			// This should fail - SecureErrorHandler doesn't exist yet
			router.Use(SecureErrorHandler(testLogger))
			
			// Add test route
			router.GET("/test", func(c *gin.Context) {
				c.Error(tt.error)
			})

			// Execute request
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify security event was logged
			require.Greater(t, len(logEntries), 0, "Should have logged security event")
			
			logEntry := logEntries[len(logEntries)-1] // Get last log entry
			assert.Contains(t, logEntry, "SECURITY_SENSITIVE", "Should mark as security sensitive")
			assert.Contains(t, logEntry, "error_type", "Should include error type")
			assert.Contains(t, logEntry, "client_ip", "Should include client IP")
			assert.Contains(t, logEntry, "user_agent", "Should include user agent")
		})
	}
}

func TestSecureErrorHandler_RequestContextLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create test logger
	var logEntries []map[string]interface{}
	testLogger := &testLogger{
		entries: &logEntries,
	}
	
	// Create router
	router := gin.New()
	
	// This should fail - SecureErrorHandler doesn't exist yet
	router.Use(SecureErrorHandler(testLogger))
	
	router.GET("/test", func(c *gin.Context) {
		err := domainErrors.InternalError(errors.New("test error"))
		c.Error(err)
	})

	// Execute request with specific headers
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Request-ID", "test-request-123")
	req.RemoteAddr = "192.168.1.100:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify request context was logged
	require.Greater(t, len(logEntries), 0)
	
	logEntry := logEntries[len(logEntries)-1]
	assert.Equal(t, "TestAgent/1.0", logEntry["user_agent"])
	assert.Equal(t, "test-request-123", logEntry["request_id"])
	assert.Contains(t, logEntry["client_ip"], "192.168.1.100")
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/test", logEntry["path"])
}

func TestSecureErrorHandler_NoErrorDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create router
	router := gin.New()
	
	// Use SecureErrorHandler with noop logger
	router.Use(SecureErrorHandler(newNoopLogger()))
	
	router.GET("/test", func(c *gin.Context) {
		// Create error with sensitive details
		err := domainErrors.NewVideoProcessingError(
			"CUSTOM_ERROR",
			"Database connection failed: mysql://user:password@localhost/db",
			map[string]interface{}{
				"database_url": "mysql://user:password@localhost/db",
				"stack_trace":  "panic: runtime error\n\tgoroutine 1 [running]",
				"config_path":  "/etc/videocraft/secret.yaml",
			},
		)
		c.Error(err)
	})

	// Execute request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response doesn't contain sensitive details
	body := w.Body.String()
	var response map[string]interface{}
	err := json.Unmarshal([]byte(body), &response)
	require.NoError(t, err)
	
	// Should not contain details field or sensitive information
	assert.NotContains(t, response, "details")
	assert.NotContains(t, body, "mysql://")
	assert.NotContains(t, body, "password")
	assert.NotContains(t, body, "stack_trace")
	assert.NotContains(t, body, "config_path")
	assert.NotContains(t, body, "goroutine")
}

// noopLogger implements logger.Logger for testing with no-op behavior
type noopLogger struct{}

func newNoopLogger() logger.Logger {
	return &noopLogger{}
}

func (nl *noopLogger) Debug(args ...interface{})                         {}
func (nl *noopLogger) Info(args ...interface{})                          {}
func (nl *noopLogger) Warn(args ...interface{})                          {}
func (nl *noopLogger) Error(args ...interface{})                         {}
func (nl *noopLogger) Fatal(args ...interface{})                         {}
func (nl *noopLogger) Debugf(format string, args ...interface{})         {}
func (nl *noopLogger) Infof(format string, args ...interface{})          {}
func (nl *noopLogger) Warnf(format string, args ...interface{})          {}
func (nl *noopLogger) Errorf(format string, args ...interface{})         {}
func (nl *noopLogger) Fatalf(format string, args ...interface{})         {}
func (nl *noopLogger) WithField(key string, value interface{}) logger.Logger { return nl }
func (nl *noopLogger) WithFields(fields map[string]interface{}) logger.Logger { return nl }

// testLogger captures log entries for testing
type testLogger struct {
	entries *[]map[string]interface{}
}

func (tl *testLogger) Debug(args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "debug", "message": args})
}

func (tl *testLogger) Info(args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "info", "message": args})
}

func (tl *testLogger) Warn(args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "warn", "message": args})
}

func (tl *testLogger) Error(args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "error", "message": args})
}

func (tl *testLogger) Fatal(args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "fatal", "message": args})
}

func (tl *testLogger) Debugf(format string, args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "debug", "message": args})
}

func (tl *testLogger) Infof(format string, args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "info", "message": args})
}

func (tl *testLogger) Warnf(format string, args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "warn", "message": args})
}

func (tl *testLogger) Errorf(format string, args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "error", "message": args})
}

func (tl *testLogger) Fatalf(format string, args ...interface{}) {
	*tl.entries = append(*tl.entries, map[string]interface{}{"level": "fatal", "message": args})
}

func (tl *testLogger) WithField(key string, value interface{}) logger.Logger {
	return tl.WithFields(map[string]interface{}{key: value})
}

func (tl *testLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return &testLoggerWithFields{
		testLogger: tl,
		fields:     fields,
	}
}

func (tl *testLogger) WithError(err error) logger.Logger {
	return tl.WithField("error", err.Error())
}

// testLoggerWithFields implements logger.Logger with fields
type testLoggerWithFields struct {
	*testLogger
	fields map[string]interface{}
}

func (tlwf *testLoggerWithFields) Debug(args ...interface{}) {
	entry := map[string]interface{}{"level": "debug", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Info(args ...interface{}) {
	entry := map[string]interface{}{"level": "info", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Warn(args ...interface{}) {
	entry := map[string]interface{}{"level": "warn", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Error(args ...interface{}) {
	entry := map[string]interface{}{"level": "error", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Fatal(args ...interface{}) {
	entry := map[string]interface{}{"level": "fatal", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Debugf(format string, args ...interface{}) {
	entry := map[string]interface{}{"level": "debug", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Infof(format string, args ...interface{}) {
	entry := map[string]interface{}{"level": "info", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Warnf(format string, args ...interface{}) {
	entry := map[string]interface{}{"level": "warn", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Errorf(format string, args ...interface{}) {
	entry := map[string]interface{}{"level": "error", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) Fatalf(format string, args ...interface{}) {
	entry := map[string]interface{}{"level": "fatal", "message": args}
	for k, v := range tlwf.fields {
		entry[k] = v
	}
	*tlwf.entries = append(*tlwf.entries, entry)
}

func (tlwf *testLoggerWithFields) WithField(key string, value interface{}) logger.Logger {
	newFields := make(map[string]interface{})
	for k, v := range tlwf.fields {
		newFields[k] = v
	}
	newFields[key] = value
	return &testLoggerWithFields{
		testLogger: tlwf.testLogger,
		fields:     newFields,
	}
}

func (tlwf *testLoggerWithFields) WithFields(fields map[string]interface{}) logger.Logger {
	newFields := make(map[string]interface{})
	for k, v := range tlwf.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &testLoggerWithFields{
		testLogger: tlwf.testLogger,
		fields:     newFields,
	}
}

func (tlwf *testLoggerWithFields) WithError(err error) logger.Logger {
	return tlwf.WithField("error", err.Error())
}