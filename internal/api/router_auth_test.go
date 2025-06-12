package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

// mockServices provides a minimal services implementation for testing
func createMockServices() *services.Services {
	return &services.Services{
		FFmpeg:        nil,
		Audio:         nil,
		Transcription: nil,
		Subtitle:      nil,
		Storage:       nil,
		Job:           nil,
	}
}

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
		logger := createNoopLogger()
		
		router := NewRouter(cfg, services, logger)
		
		// Test protected endpoint without authentication
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/generate-video", nil)
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
		logger := createNoopLogger()
		
		router := NewRouter(cfg, services, logger)
		
		// Test health endpoints
		healthEndpoints := []string{"/health", "/ready", "/live", "/metrics"}
		
		for _, endpoint := range healthEndpoints {
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
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
		logger := createNoopLogger()
		
		router := NewRouter(cfg, services, logger)
		
		// Test with correct authentication
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/videos", nil)
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
		logger := createNoopLogger()
		
		router := NewRouter(cfg, services, logger)
		
		// Test legacy endpoints without auth
		legacyEndpoints := []string{
			"/generate-video",
			"/videos",
		}
		
		for _, endpoint := range legacyEndpoints {
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			// Legacy endpoints should also require authentication
			assert.Equal(t, http.StatusUnauthorized, w.Code,
				"Legacy endpoint %s should require authentication", endpoint)
		}
	})
}

// noopLogger implements logger.Logger for testing
type noopLogger struct{}

func createNoopLogger() logger.Logger {
	return &noopLogger{}
}

func (n *noopLogger) Debug(args ...interface{})                   {}
func (n *noopLogger) Info(args ...interface{})                    {}
func (n *noopLogger) Warn(args ...interface{})                    {}
func (n *noopLogger) Error(args ...interface{})                   {}
func (n *noopLogger) Fatal(args ...interface{})                   {}
func (n *noopLogger) Debugf(format string, args ...interface{})  {}
func (n *noopLogger) Infof(format string, args ...interface{})   {}
func (n *noopLogger) Warnf(format string, args ...interface{})   {}
func (n *noopLogger) Errorf(format string, args ...interface{})  {}
func (n *noopLogger) Fatalf(format string, args ...interface{})  {}
func (n *noopLogger) WithField(key string, value interface{}) logger.Logger { return n }
func (n *noopLogger) WithFields(fields map[string]interface{}) logger.Logger { return n }