package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
)

func TestFFmpegService_URLValidation_CommandInjectionPrevention(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
	}

	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	tests := []struct {
		name          string
		maliciousURL  string
		expectedError string
		shouldFail    bool
	}{
		{
			name:          "Command injection with semicolon",
			maliciousURL:  "http://example.com/video.mp4; rm -rf /",
			expectedError: "URL contains prohibited characters",
			shouldFail:    true,
		},
		{
			name:          "Command injection with pipe",
			maliciousURL:  "http://example.com/video.mp4 | cat /etc/passwd",
			expectedError: "URL contains prohibited characters",
			shouldFail:    true,
		},
		{
			name:          "Command injection with backticks",
			maliciousURL:  "http://example.com/video.mp4`whoami`",
			expectedError: "URL contains prohibited characters",
			shouldFail:    true,
		},
		{
			name:          "Command injection with dollar substitution",
			maliciousURL:  "http://example.com/video.mp4$(rm -rf /)",
			expectedError: "URL contains prohibited characters",
			shouldFail:    true,
		},
		{
			name:          "Path traversal attempt",
			maliciousURL:  "http://example.com/../../../etc/passwd",
			expectedError: "URL contains path traversal sequences",
			shouldFail:    true,
		},
		{
			name:          "File protocol injection",
			maliciousURL:  "file:///etc/passwd",
			expectedError: "protocol not allowed",
			shouldFail:    true,
		},
		{
			name:          "Data URI injection",
			maliciousURL:  "data:text/plain;base64,SGVsbG8gV29ybGQ=",
			expectedError: "protocol not allowed",
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test URL validation directly
			err := service.ValidateURL(tt.maliciousURL)

			if tt.shouldFail {
				require.Error(t, err, "Expected validation to fail for malicious URL")
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err, "Expected validation to pass for valid URL")
			}
		})
	}
}

func TestFFmpegService_URLValidation_ValidURLs(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
	}

	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	validURLs := []string{
		"https://example.com/video.mp4",
		"http://example.com/audio.wav",
		"https://cdn.example.com/media/video.avi",
		"http://streaming.example.org/live.m3u8",
	}

	for _, url := range validURLs {
		t.Run("Valid URL: "+url, func(t *testing.T) {
			err := service.ValidateURL(url)
			assert.NoError(t, err, "Valid URL should pass validation")
		})
	}
}

func TestFFmpegService_InputSanitization(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
	}

	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	tests := []struct {
		name           string
		input          string
		expectedOutput string
		shouldFail     bool
	}{
		{
			name:           "Remove shell metacharacters",
			input:          "video.mp4; rm -rf /",
			expectedOutput: "video.mp4",
			shouldFail:     false,
		},
		{
			name:           "Remove pipe characters",
			input:          "video.mp4 | cat /etc/passwd",
			expectedOutput: "video.mp4",
			shouldFail:     false,
		},
		{
			name:           "Clean path traversal",
			input:          "../../../etc/passwd",
			expectedOutput: "etc/passwd",
			shouldFail:     false,
		},
		{
			name:       "Reject if only malicious content",
			input:      "; rm -rf /",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, err := service.SanitizeInput(tt.input)

			if tt.shouldFail {
				require.Error(t, err, "Expected sanitization to fail for malicious input")
			} else {
				require.NoError(t, err, "Expected sanitization to succeed")
				assert.Equal(t, tt.expectedOutput, sanitized)
			}
		})
	}
}

func TestFFmpegService_BuildCommand_WithMaliciousURLs(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
	}

	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	// Create a config with malicious URLs
	videoConfig := models.VideoConfigArray{
		{
			Scenes: []models.Scene{
				{
					Elements: []models.Element{
						{
							Type: "audio",
							Src:  "http://example.com/audio.wav | cat /etc/passwd",
						},
						{
							Type: "image",
							Src:  "http://example.com/img.jpg`whoami`",
						},
					},
				},
			},
		},
	}

	// This should fail due to malicious URLs
	_, err := service.BuildCommand(&videoConfig)
	require.Error(t, err, "Expected command building to fail with malicious URLs")
	assert.Contains(t, err.Error(), "security validation failed")
}

func TestFFmpegService_URLAllowlist(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
		Security: config.SecurityConfig{
			AllowedDomains: []string{
				"trusted.example.com",
				"cdn.trusted.org",
			},
		},
	}

	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	tests := []struct {
		name       string
		url        string
		shouldPass bool
	}{
		{
			name:       "Allowed domain should pass",
			url:        "https://trusted.example.com/video.mp4",
			shouldPass: true,
		},
		{
			name:       "Another allowed domain should pass",
			url:        "https://cdn.trusted.org/audio.wav",
			shouldPass: true,
		},
		{
			name:       "Disallowed domain should fail",
			url:        "https://malicious.example.com/video.mp4",
			shouldPass: false,
		},
		{
			name:       "Unknown domain should fail",
			url:        "https://unknown.com/video.mp4",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateURLAllowlist(tt.url)

			if tt.shouldPass {
				assert.NoError(t, err, "URL from allowed domain should pass")
			} else {
				assert.Error(t, err, "URL from disallowed domain should fail")
				assert.Contains(t, err.Error(), "domain not in allowlist")
			}
		})
	}
}

// NoopLogger implements logger.Logger for testing
type NoopLogger struct{}

func (n *NoopLogger) Debug(args ...interface{}) {}
func (n *NoopLogger) Info(args ...interface{})  {}
func (n *NoopLogger) Warn(args ...interface{})  {}
func (n *NoopLogger) Error(args ...interface{}) {}
func (n *NoopLogger) Fatal(args ...interface{}) {}

func (n *NoopLogger) Debugf(format string, args ...interface{}) {}
func (n *NoopLogger) Infof(format string, args ...interface{})  {}
func (n *NoopLogger) Warnf(format string, args ...interface{})  {}
func (n *NoopLogger) Errorf(format string, args ...interface{}) {}
func (n *NoopLogger) Fatalf(format string, args ...interface{}) {}

func (n *NoopLogger) WithField(key string, value interface{}) logger.Logger {
	return n
}

func (n *NoopLogger) WithFields(fields map[string]interface{}) logger.Logger {
	return n
}

func (n *NoopLogger) WithError(err error) logger.Logger {
	return n
}
