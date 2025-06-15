package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

// mockServices provides a minimal services implementation for testing
func createMockServices() *services.Services {
	return &services.Services{
		FFmpeg:        &mockFFmpegService{},
		Audio:         &mockAudioService{},
		Transcription: &mockTranscriptionService{},
		Subtitle:      &mockSubtitleService{},
		Storage:       &mockStorageService{},
		Job:           &mockJobService{},
	}
}

// Mock service implementations for testing
type mockStorageService struct{}

func (m *mockStorageService) StoreVideo(videoPath string) (string, error) { return "", nil }
func (m *mockStorageService) GetVideo(videoID string) (string, error)     { return "", nil }
func (m *mockStorageService) DeleteVideo(videoID string) error            { return nil }
func (m *mockStorageService) ListVideos() ([]services.VideoInfo, error) {
	return []services.VideoInfo{}, nil
}
func (m *mockStorageService) CleanupOldFiles() error { return nil }

type mockJobService struct{}

func (m *mockJobService) CreateJob(config *models.VideoConfigArray) (*models.Job, error) {
	return nil, nil
}
func (m *mockJobService) ProcessJob(ctx context.Context, job *models.Job) error { return nil }
func (m *mockJobService) GetJob(id string) (*models.Job, error)                 { return nil, nil }
func (m *mockJobService) UpdateJobProgress(id string, progress int) error       { return nil }
func (m *mockJobService) UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error {
	return nil
}
func (m *mockJobService) ListJobs() ([]*models.Job, error) { return []*models.Job{}, nil }
func (m *mockJobService) CancelJob(id string) error        { return nil }

type mockFFmpegService struct{}

func (m *mockFFmpegService) GenerateVideo(ctx context.Context, config *models.VideoConfigArray, progressChan chan<- int) (string, error) {
	return "", nil
}
func (m *mockFFmpegService) BuildCommand(config *models.VideoConfigArray) (*services.FFmpegCommand, error) {
	return nil, nil
}
func (m *mockFFmpegService) Execute(ctx context.Context, cmd *services.FFmpegCommand) error {
	return nil
}

type mockAudioService struct{}

func (m *mockAudioService) AnalyzeAudio(ctx context.Context, url string) (*services.AudioInfo, error) {
	return nil, nil
}
func (m *mockAudioService) CalculateSceneTiming(elements []models.Element) ([]models.TimingSegment, error) {
	return nil, nil
}
func (m *mockAudioService) DownloadAudio(ctx context.Context, url string) (string, error) {
	return "", nil
}

type mockTranscriptionService struct{}

func (m *mockTranscriptionService) TranscribeAudio(ctx context.Context, url string) (*services.TranscriptionResult, error) {
	return nil, nil
}
func (m *mockTranscriptionService) Shutdown() {}

type mockSubtitleService struct{}

func (m *mockSubtitleService) GenerateSubtitles(ctx context.Context, project models.VideoProject) (*services.SubtitleResult, error) {
	return nil, nil
}
func (m *mockSubtitleService) ValidateSubtitleConfig(project models.VideoProject) error { return nil }
func (m *mockSubtitleService) CleanupTempFiles(filePath string) error                   { return nil }

func TestRouter_AuthenticationEnforcementByDefault(t *testing.T) {
	t.Run("router should enforce authentication by default", func(t *testing.T) {
		// Create config with auth enabled (this should be the default)
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth: true,
				APIKey:     "test-api-key-12345678901234567890123456789012",
				RateLimit:  100,
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test protected endpoint without authentication
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/generate-video", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should be rejected with 401
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "API key is required")
	})

	t.Run("health endpoints should not require authentication", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth: true,
				APIKey:     "test-api-key-12345678901234567890123456789012",
				RateLimit:  100,
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test health endpoints
		healthEndpoints := []string{"/health", "/ready", "/live", "/metrics"}

		for _, endpoint := range healthEndpoints {
			req, _ := http.NewRequest(http.MethodGet, endpoint, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Health endpoints should not return 401 Unauthorized
			assert.NotEqual(t, http.StatusUnauthorized, w.Code,
				"Health endpoint %s should not require authentication", endpoint)
		}
	})

	t.Run("authenticated requests should be allowed", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth: true,
				APIKey:     "test-api-key-12345678901234567890123456789012",
				RateLimit:  100,
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test with correct authentication
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/videos", http.NoBody)
		req.Header.Set("Authorization", "Bearer test-api-key-12345678901234567890123456789012")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not be rejected with 401 (may be other errors due to mock services)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRouter_LegacyEndpointsRequireAuth(t *testing.T) {
	t.Run("legacy endpoints should also require authentication", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth: true,
				APIKey:     "test-api-key-12345678901234567890123456789012",
				RateLimit:  100,
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test legacy endpoints without auth
		legacyEndpoints := []struct {
			method string
			path   string
		}{
			{"POST", "/generate-video"},
			{"GET", "/videos"},
		}

		for _, endpoint := range legacyEndpoints {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, http.NoBody)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Legacy endpoints should also require authentication
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"Legacy endpoint %s %s should require authentication", endpoint.method, endpoint.path)
		}
	})
}

// noopLogger implements logger.Logger for testing
type noopLogger struct{}

func createNoopLogger() logger.Logger {
	return &noopLogger{}
}

func (n *noopLogger) Debug(args ...interface{})                              {}
func (n *noopLogger) Info(args ...interface{})                               {}
func (n *noopLogger) Warn(args ...interface{})                               {}
func (n *noopLogger) Error(args ...interface{})                              {}
func (n *noopLogger) Fatal(args ...interface{})                              {}
func (n *noopLogger) Debugf(format string, args ...interface{})              {}
func (n *noopLogger) Infof(format string, args ...interface{})               {}
func (n *noopLogger) Warnf(format string, args ...interface{})               {}
func (n *noopLogger) Errorf(format string, args ...interface{})              {}
func (n *noopLogger) Fatalf(format string, args ...interface{})              {}
func (n *noopLogger) WithField(key string, value interface{}) logger.Logger  { return n }
func (n *noopLogger) WithFields(fields map[string]interface{}) logger.Logger { return n }
