package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/api/handlers"
	"github.com/activadee/videocraft/internal/api/middleware"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
	domainErrors "github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

func TestSecureErrorHandling_EndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup test configuration
	cfg := &config.Config{
		Auth: config.AuthConfig{
			APIKey: "test-api-key",
		},
	}
	
	// Create mock services that will return errors
	mockServices := &services.Services{
		Job: &mockJobService{},
	}
	
	// Create router with secure error handling
	router := gin.New()
	
	// Use SecureErrorHandler with noop logger
	router.Use(middleware.SecureErrorHandler(newNoopLogger()))
	router.Use(middleware.Auth(cfg.Auth.APIKey))
	
	// Setup handlers
	videoHandler := handlers.NewVideoHandler(cfg, mockServices, logger.NewNoop())
	
	// Register routes
	v1 := router.Group("/api/v1")
	v1.POST("/videos", videoHandler.GenerateVideo)

	tests := []struct {
		name             string
		requestBody      interface{}
		setupMockError   func(*mockJobService)
		expectedStatus   int
		expectedCode     string
		expectedError    string
		shouldNotContain []string
	}{
		{
			name: "should sanitize FFmpeg error responses",
			requestBody: models.VideoConfigArray{
				Video: models.VideoConfig{
					Width:      1920,
					Height:     1080,
					FPS:        30,
					Duration:   10,
					Background: "test.mp4",
				},
				Scenes: []models.Scene{
					{
						Elements: []models.Element{
							{Type: "audio", Src: "http://example.com/audio.mp3"},
						},
					},
				},
			},
			setupMockError: func(mjs *mockJobService) {
				mjs.shouldError = true
				mjs.errorMessage = "FFmpeg failed: cannot access /etc/passwd file"
				mjs.errorCode = "FFMPEG_FAILED"
			},
			expectedStatus:   http.StatusUnprocessableEntity,
			expectedCode:     "FFMPEG_FAILED",
			expectedError:    "Video processing failed",
			shouldNotContain: []string{"/etc/passwd", "cannot access"},
		},
		{
			name: "should sanitize storage error responses",
			requestBody: models.VideoConfigArray{
				Video: models.VideoConfig{
					Width:      1920,
					Height:     1080,
					FPS:        30,
					Duration:   10,
					Background: "test.mp4",
				},
				Scenes: []models.Scene{
					{
						Elements: []models.Element{
							{Type: "audio", Src: "http://example.com/audio.mp3"},
						},
					},
				},
			},
			setupMockError: func(mjs *mockJobService) {
				mjs.shouldError = true
				mjs.errorMessage = "Storage failed: permission denied accessing /var/lib/mysql/data"
				mjs.errorCode = "STORAGE_FAILED"
			},
			expectedStatus:   http.StatusInsufficientStorage,
			expectedCode:     "STORAGE_FAILED",
			expectedError:    "Storage operation failed",
			shouldNotContain: []string{"/var/lib/mysql", "permission denied"},
		},
		{
			name: "should sanitize internal error responses",
			requestBody: models.VideoConfigArray{
				Video: models.VideoConfig{
					Width:      1920,
					Height:     1080,
					FPS:        30,
					Duration:   10,
					Background: "test.mp4",
				},
				Scenes: []models.Scene{
					{
						Elements: []models.Element{
							{Type: "audio", Src: "http://example.com/audio.mp3"},
						},
					},
				},
			},
			setupMockError: func(mjs *mockJobService) {
				mjs.shouldError = true
				mjs.errorMessage = "Database panic: mysql://user:password@localhost/db connection failed\ngoroutine 1 [running]:\nmain.connectDB()"
				mjs.errorCode = "INTERNAL_ERROR"
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedError:  "Internal server error occurred",
			shouldNotContain: []string{
				"mysql://",
				"user:password",
				"Database panic",
				"goroutine",
				"connectDB",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock error
			mockService := mockServices.Job.(*mockJobService)
			tt.setupMockError(mockService)
			
			// Create request
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			
			req, err := http.NewRequest("POST", "/api/v1/videos", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-api-key")
			
			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// Verify response structure
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Response should be valid JSON")
			
			// Verify standardized error format
			assert.Equal(t, tt.expectedCode, response["code"])
			assert.Equal(t, tt.expectedError, response["error"])
			
			// Verify no sensitive information is exposed
			responseBody := w.Body.String()
			for _, sensitive := range tt.shouldNotContain {
				assert.NotContains(t, responseBody, sensitive,
					"Response should not contain sensitive information: %s", sensitive)
			}
			
			// Verify response doesn't contain prohibited fields
			assert.NotContains(t, response, "details")
			assert.NotContains(t, response, "original_error")
			assert.NotContains(t, response, "stack_trace")
			assert.NotContains(t, response, "internal_message")
		})
	}
}

func TestSecureErrorHandling_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Setup router
	router := gin.New()
	
	// Use SecureErrorHandler with noop logger
	router.Use(middleware.SecureErrorHandler(newNoopLogger()))
	
	// Add test endpoint that validates JSON
	router.POST("/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedCode   string
		expectedError  string
	}{
		{
			name:           "should sanitize JSON parsing errors",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_INPUT",
			expectedError:  "Invalid request format",
		},
		{
			name:           "should sanitize malformed JSON errors",
			requestBody:    `{"unclosed": "string`,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_INPUT",
			expectedError:  "Invalid request format",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)
			
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedCode, response["code"])
			assert.Equal(t, tt.expectedError, response["error"])
			
			// Should not contain parsing details
			responseBody := w.Body.String()
			assert.NotContains(t, responseBody, "json:")
			assert.NotContains(t, responseBody, "syntax error")
			assert.NotContains(t, responseBody, "unexpected")
		})
	}
}

func TestSecureErrorHandling_PanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create router with secure error handling
	router := gin.New()
	
	// Use SecureErrorHandler with noop logger
	router.Use(middleware.SecureErrorHandler(newNoopLogger()))
	
	// Add route that panics
	router.GET("/panic", func(c *gin.Context) {
		panic("database connection string: mysql://user:password@localhost/db\ngoroutine 1 [running]:\nmain.processRequest()\n\t/app/internal/handlers/video.go:123 +0x1a4")
	})
	
	// Execute request
	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should recover and return sanitized error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "INTERNAL_ERROR", response["code"])
	assert.Equal(t, "Internal server error occurred", response["error"])
	
	// Should not contain panic details
	responseBody := w.Body.String()
	assert.NotContains(t, responseBody, "mysql://")
	assert.NotContains(t, responseBody, "user:password")
	assert.NotContains(t, responseBody, "goroutine")
	assert.NotContains(t, responseBody, "/app/internal")
	assert.NotContains(t, responseBody, "processRequest")
}

// mockJobService implements a mock job service for testing
type mockJobService struct {
	shouldError  bool
	errorMessage string
	errorCode    string
}

func (mjs *mockJobService) CreateJob(config *models.VideoConfigArray) (*models.Job, error) {
	if mjs.shouldError {
		// This should fail - secure error functions don't exist yet
		switch mjs.errorCode {
		case "FFMPEG_FAILED":
			return nil, domainErrors.FFmpegFailed(errors.New(mjs.errorMessage))
		case "STORAGE_FAILED":
			return nil, domainErrors.StorageFailed(errors.New(mjs.errorMessage))
		case "INTERNAL_ERROR":
			return nil, domainErrors.InternalError(errors.New(mjs.errorMessage))
		default:
			return nil, errors.New(mjs.errorMessage)
		}
	}
	
	return &models.Job{
		ID:     "test-job-id",
		Status: models.JobStatusPending,
	}, nil
}

func (mjs *mockJobService) GetJob(id string) (*models.Job, error) {
	return &models.Job{
		ID:     id,
		Status: models.JobStatusPending,
	}, nil
}

func (mjs *mockJobService) ProcessJob(ctx context.Context, job *models.Job) error {
	return nil
}

func (mjs *mockJobService) UpdateJobProgress(id string, progress int) error {
	return nil
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