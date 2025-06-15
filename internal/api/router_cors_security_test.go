package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/pkg/logger"
)

// TestCORS_WildcardOriginsRemoved tests that wildcard CORS origins are removed
func TestCORS_WildcardOriginsRemoved(t *testing.T) {
	t.Run("should reject wildcard CORS origins", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false, // Disable auth for CORS testing
				AllowedDomains: []string{"trusted.example.com", "api.trusted.org"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test preflight request from disallowed origin
		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
		req.Header.Set("Origin", "https://malicious.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not allow all origins (wildcard should be removed)
		corsHeader := w.Header().Get("Access-Control-Allow-Origin")
		assert.NotEqual(t, "*", corsHeader, "Wildcard CORS origins should be removed")
		assert.NotEqual(t, "https://malicious.example.com", corsHeader, "Malicious origin should be rejected")
	})
}

// TestCORS_DomainAllowlisting tests that only approved domains can access the API
func TestCORS_DomainAllowlisting(t *testing.T) {
	t.Run("should only allow approved domains", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com", "api.trusted.org"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		testCases := []struct {
			name           string
			origin         string
			shouldBeAllowed bool
		}{
			{
				name:           "trusted domain should be allowed",
				origin:         "https://trusted.example.com",
				shouldBeAllowed: true,
			},
			{
				name:           "second trusted domain should be allowed",
				origin:         "https://api.trusted.org",
				shouldBeAllowed: true,
			},
			{
				name:           "untrusted domain should be rejected",
				origin:         "https://malicious.example.com",
				shouldBeAllowed: false,
			},
			{
				name:           "subdomain of trusted domain should be rejected",
				origin:         "https://sub.trusted.example.com",
				shouldBeAllowed: false,
			},
			{
				name:           "similar domain should be rejected",
				origin:         "https://trusted.example.com.evil.org",
				shouldBeAllowed: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test preflight request
				req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
				req.Header.Set("Origin", tc.origin)
				req.Header.Set("Access-Control-Request-Method", "POST")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				corsHeader := w.Header().Get("Access-Control-Allow-Origin")

				if tc.shouldBeAllowed {
					assert.Equal(t, tc.origin, corsHeader, "Trusted origin should be allowed")
				} else {
					assert.NotEqual(t, tc.origin, corsHeader, "Untrusted origin should not be allowed")
				}
			})
		}
	})
}

// TestCORS_CSRFProtection tests that CSRF protection is implemented
func TestCORS_CSRFProtection(t *testing.T) {
	t.Run("should implement CSRF token validation", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				EnableCSRF:      true,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test POST request without CSRF token
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/generate-video", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://trusted.example.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should be rejected due to missing CSRF token
		assert.Equal(t, http.StatusForbidden, w.Code, "Request without CSRF token should be rejected")
		assert.Contains(t, w.Body.String(), "CSRF", "Response should mention CSRF protection")
	})

	t.Run("should accept requests with valid CSRF token", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				EnableCSRF:      true,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// First, get CSRF token
		tokenReq, _ := http.NewRequest(http.MethodGet, "/api/v1/csrf-token", nil)
		tokenReq.Header.Set("Origin", "https://trusted.example.com")

		tokenW := httptest.NewRecorder()
		router.ServeHTTP(tokenW, tokenReq)

		// Extract CSRF token from response
		require.Equal(t, http.StatusOK, tokenW.Code, "Should be able to get CSRF token")
		
		// Test POST request with CSRF token
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/generate-video", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://trusted.example.com")
		req.Header.Set("X-CSRF-Token", "valid-csrf-token") // In real implementation, extract from tokenW

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not be rejected due to CSRF (may fail for other reasons)
		assert.NotEqual(t, http.StatusForbidden, w.Code, "Request with valid CSRF token should not be rejected for CSRF reasons")
	})
}

// TestCORS_ProperHeaders tests that CORS headers are properly configured
func TestCORS_ProperHeaders(t *testing.T) {
	t.Run("should set proper CORS headers", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test preflight request
		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
		req.Header.Set("Origin", "https://trusted.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Check required CORS headers
		assert.Equal(t, "https://trusted.example.com", w.Header().Get("Access-Control-Allow-Origin"), "Should set proper Allow-Origin")
		
		allowMethods := w.Header().Get("Access-Control-Allow-Methods")
		assert.Contains(t, allowMethods, "POST", "Should allow POST method")
		assert.Contains(t, allowMethods, "GET", "Should allow GET method")
		assert.NotContains(t, allowMethods, "TRACE", "Should not allow dangerous TRACE method")
		
		allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
		assert.Contains(t, allowHeaders, "Content-Type", "Should allow Content-Type header")
		assert.Contains(t, allowHeaders, "Authorization", "Should allow Authorization header")
		
		// Should not allow credentials with multiple origins
		credentials := w.Header().Get("Access-Control-Allow-Credentials")
		if len(cfg.Security.AllowedDomains) > 1 {
			assert.NotEqual(t, "true", credentials, "Should not allow credentials with multiple domains")
		}
	})

	t.Run("should reject dangerous methods", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		dangerousMethods := []string{"TRACE", "CONNECT", "PATCH"}

		for _, method := range dangerousMethods {
			req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
			req.Header.Set("Origin", "https://trusted.example.com")
			req.Header.Set("Access-Control-Request-Method", method)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			allowMethods := w.Header().Get("Access-Control-Allow-Methods")
			assert.NotContains(t, allowMethods, method, "Should not allow dangerous method: %s", method)
		}
	})
}

// TestCORS_PreflightHandling tests that preflight requests are properly handled
func TestCORS_PreflightHandling(t *testing.T) {
	t.Run("should handle preflight requests properly", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test preflight request
		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
		req.Header.Set("Origin", "https://trusted.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 200 or 204 for preflight
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNoContent, 
			"Preflight request should return 200 or 204, got %d", w.Code)

		// Should have CORS headers
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"), "Should have Allow-Origin header")
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"), "Should have Allow-Methods header")
		
		// Should have cache control for preflight
		maxAge := w.Header().Get("Access-Control-Max-Age")
		if maxAge != "" {
			assert.NotEqual(t, "0", maxAge, "Should cache preflight responses")
		}
	})

	t.Run("should reject preflight from unauthorized origin", func(t *testing.T) {
		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		testLogger := createNoopLogger()

		router := NewRouter(cfg, services, testLogger)

		// Test preflight request from unauthorized origin
		req, _ := http.NewRequest(http.MethodOptions, "/api/v1/videos", nil)
		req.Header.Set("Origin", "https://malicious.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should not set CORS headers for unauthorized origin
		corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
		assert.NotEqual(t, "https://malicious.example.com", corsOrigin, "Should not allow unauthorized origin")
		assert.NotEqual(t, "*", corsOrigin, "Should not use wildcard origin")
	})
}

// TestCORS_SecurityViolationLogging tests that security violations are logged
func TestCORS_SecurityViolationLogging(t *testing.T) {
	t.Run("should log CORS security violations", func(t *testing.T) {
		// Create a test logger that captures log messages
		testLogger := &testLogger{messages: make([]string, 0)}

		cfg := &config.Config{
			Security: config.SecurityConfig{
				EnableAuth:      false,
				AllowedDomains: []string{"trusted.example.com"},
			},
			Log: config.LogConfig{
				Level:  "info",
				Format: "text",
			},
		}

		services := createMockServices()
		router := NewRouter(cfg, services, testLogger)

		// Test request from unauthorized origin
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/videos", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "https://malicious.example.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should log security violation
		found := false
		for _, msg := range testLogger.messages {
			if strings.Contains(strings.ToUpper(msg), "CORS") && 
			   strings.Contains(strings.ToUpper(msg), "VIOLATION") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should log CORS security violations")
	})
}

// testLogger captures log messages for testing
type testLogger struct {
	messages []string
}

func (t *testLogger) Debug(args ...interface{})                              { t.messages = append(t.messages, "DEBUG") }
func (t *testLogger) Info(args ...interface{})                               { t.messages = append(t.messages, "INFO") }
func (t *testLogger) Warn(args ...interface{})                               { t.messages = append(t.messages, "WARN") }
func (t *testLogger) Error(args ...interface{})                              { t.messages = append(t.messages, "ERROR") }
func (t *testLogger) Fatal(args ...interface{})                              { t.messages = append(t.messages, "FATAL") }
func (t *testLogger) Debugf(format string, args ...interface{})              { t.messages = append(t.messages, "DEBUG") }
func (t *testLogger) Infof(format string, args ...interface{})               { t.messages = append(t.messages, "INFO") }
func (t *testLogger) Warnf(format string, args ...interface{})               { t.messages = append(t.messages, "WARN") }
func (t *testLogger) Errorf(format string, args ...interface{})              { t.messages = append(t.messages, "ERROR") }
func (t *testLogger) Fatalf(format string, args ...interface{})              { t.messages = append(t.messages, "FATAL") }
func (t *testLogger) WithField(key string, value interface{}) logger.Logger  { return t }
func (t *testLogger) WithFields(fields map[string]interface{}) logger.Logger { return t }