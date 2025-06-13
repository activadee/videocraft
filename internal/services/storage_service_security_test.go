package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService_PathTraversalPrevention(t *testing.T) {
	// Create temp directories for testing
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "outputs")
	sensitiveDir := filepath.Join(tempDir, "sensitive")

	// Create directories
	require.NoError(t, os.MkdirAll(outputDir, 0755))
	require.NoError(t, os.MkdirAll(sensitiveDir, 0755))

	// Create sensitive file outside allowed directory
	sensitiveFile := filepath.Join(sensitiveDir, "secret.txt")
	require.NoError(t, os.WriteFile(sensitiveFile, []byte("SECRET DATA"), 0644))

	cfg := &config.Config{
		Storage: config.StorageConfig{
			OutputDir: outputDir,
		},
	}

	service := &storageService{
		cfg: cfg,
		log: logger.New("debug"),
	}

	// Test cases for path traversal attacks
	pathTraversalTests := []struct {
		name          string
		videoID       string
		expectedError string
		shouldFail    bool
	}{
		{
			name:          "Basic path traversal with ../",
			videoID:       "../sensitive/secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Multiple path traversal attempts",
			videoID:       "../../sensitive/secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Windows path traversal with ..\\ ",
			videoID:       "..\\sensitive\\secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Encoded path traversal %2e%2e/",
			videoID:       "%2e%2e/sensitive/secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Double encoded path traversal",
			videoID:       "%252e%252e/sensitive/secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Absolute path attempt",
			videoID:       "/etc/passwd",
			expectedError: "absolute path not allowed",
			shouldFail:    true,
		},
		{
			name:          "Hidden file access attempt",
			videoID:       "../.env",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Nested path traversal",
			videoID:       "safe/../../../sensitive/secret",
			expectedError: "path traversal detected",
			shouldFail:    true,
		},
		{
			name:          "Valid video ID should pass",
			videoID:       "valid-video-123",
			expectedError: "",
			shouldFail:    false,
		},
		{
			name:          "UUID format should pass",
			videoID:       "550e8400-e29b-41d4-a716-446655440000",
			expectedError: "",
			shouldFail:    false,
		},
	}

	for _, tt := range pathTraversalTests {
		t.Run(tt.name, func(t *testing.T) {
			// Test GetVideo method
			_, err := service.GetVideo(tt.videoID)

			if tt.shouldFail {
				require.Error(t, err, "Expected GetVideo to fail for: %s", tt.videoID)
				assert.Contains(t, err.Error(), tt.expectedError, "Error should contain expected message")
			} else {
				// For valid IDs, we expect "file not found" not "path traversal"
				if err != nil {
					assert.NotContains(t, err.Error(), "path traversal", "Valid ID should not trigger path traversal error")
				}
			}
		})
	}
}

func TestStorageService_DirectoryTraversalBoundaryCheck(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "outputs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.Config{
		Storage: config.StorageConfig{
			OutputDir: outputDir,
		},
	}

	service := &storageService{
		cfg: cfg,
		log: logger.New("debug"),
	}

	// Test that service operations are confined to designated directories
	t.Run("Should reject access outside output directory", func(t *testing.T) {
		// Attempt to access parent directory
		videoID := "../../../etc/passwd"

		_, err := service.GetVideo(videoID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal", "Should detect path traversal attempt")
	})

	t.Run("Should reject symbolic link traversal", func(t *testing.T) {
		// Create a symbolic link pointing outside the allowed directory
		linkPath := filepath.Join(outputDir, "malicious-link.mp4")
		targetPath := filepath.Join(tempDir, "outside.txt")

		// Create target file outside allowed directory
		require.NoError(t, os.WriteFile(targetPath, []byte("outside data"), 0644))

		// Create symbolic link (if supported)
		err := os.Symlink(targetPath, linkPath)
		if err != nil {
			t.Skipf("Symbolic links not supported on this system: %v", err)
			return
		}

		// Test accessing via symlink
		videoID := "malicious-link"
		_, err = service.GetVideo(videoID)
		require.Error(t, err)

		// Either symlink detection OR file not found is acceptable security behavior
		errorMsg := err.Error()
		isSecure := strings.Contains(errorMsg, "symbolic link") ||
			strings.Contains(errorMsg, "File not found") ||
			strings.Contains(errorMsg, "SECURITY_VIOLATION")
		assert.True(t, isSecure, "Should detect and reject symbolic link traversal or not find file, got: %s", errorMsg)
	})
}

func TestStorageService_PathCanonicalization(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "outputs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.Config{
		Storage: config.StorageConfig{
			OutputDir: outputDir,
		},
	}

	service := &storageService{
		cfg: cfg,
		log: logger.New("debug"),
	}

	// Test path canonicalization and normalization
	canonicalizationTests := []struct {
		name        string
		videoID     string
		shouldFail  bool
		description string
	}{
		{
			name:        "Mixed slashes should be normalized",
			videoID:     "video\\..\\sensitive/secret",
			shouldFail:  true,
			description: "Mixed path separators should be normalized and detected",
		},
		{
			name:        "Redundant path segments",
			videoID:     "video/./../../sensitive",
			shouldFail:  true,
			description: "Redundant path segments should be resolved and detected",
		},
		{
			name:        "Valid path with redundant segments",
			videoID:     "valid/./video",
			shouldFail:  false,
			description: "Valid paths with redundant segments should be allowed after normalization",
		},
	}

	for _, tt := range canonicalizationTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetVideo(tt.videoID)

			if tt.shouldFail {
				require.Error(t, err, tt.description)
			} else {
				// For valid paths, we might get "file not found" but not security errors
				if err != nil {
					assert.NotContains(t, err.Error(), "path traversal", tt.description)
				}
			}
		})
	}
}

func TestStorageService_InputSanitization(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "outputs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.Config{
		Storage: config.StorageConfig{
			OutputDir: outputDir,
		},
	}

	service := &storageService{
		cfg: cfg,
		log: logger.New("debug"),
	}

	// Test input sanitization for various attack vectors
	sanitizationTests := []struct {
		name        string
		videoID     string
		shouldFail  bool
		description string
	}{
		{
			name:        "Null byte injection",
			videoID:     "video\x00../../../etc/passwd",
			shouldFail:  true,
			description: "Null byte injection should be detected and rejected",
		},
		{
			name:        "Control characters",
			videoID:     "video\r\n../sensitive",
			shouldFail:  true,
			description: "Control characters should be sanitized",
		},
		{
			name:        "Unicode normalization attack",
			videoID:     "video\u002e\u002e/sensitive",
			shouldFail:  true,
			description: "Unicode path traversal should be detected",
		},
		{
			name:        "Empty string",
			videoID:     "",
			shouldFail:  true,
			description: "Empty video ID should be rejected",
		},
		{
			name:        "Whitespace only",
			videoID:     "   ",
			shouldFail:  true,
			description: "Whitespace-only video ID should be rejected",
		},
	}

	for _, tt := range sanitizationTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetVideo(tt.videoID)

			if tt.shouldFail {
				require.Error(t, err, tt.description)
			}
		})
	}
}

func TestStorageService_ErrorHandlingSecurity(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "outputs")
	require.NoError(t, os.MkdirAll(outputDir, 0755))

	cfg := &config.Config{
		Storage: config.StorageConfig{
			OutputDir: outputDir,
		},
	}

	service := &storageService{
		cfg: cfg,
		log: logger.New("debug"),
	}

	t.Run("Error messages should not leak filesystem information", func(t *testing.T) {
		// Test that error messages don't reveal internal paths or sensitive info
		videoID := "../../../etc/passwd"

		_, err := service.GetVideo(videoID)
		require.Error(t, err)

		// Error should not contain sensitive filesystem paths
		errorMsg := err.Error()
		assert.NotContains(t, errorMsg, "/etc/passwd", "Error should not leak attempted path")
		assert.NotContains(t, errorMsg, tempDir, "Error should not leak internal directory structure")

		// Error should be generic security message
		assert.Contains(t, errorMsg, "path traversal", "Error should indicate security violation")
	})

	t.Run("Valid but non-existent files should return appropriate error", func(t *testing.T) {
		videoID := "non-existent-but-valid-id"

		_, err := service.GetVideo(videoID)
		require.Error(t, err)

		// Should return "file not found" type error, not security error
		errorMsg := err.Error()
		assert.NotContains(t, errorMsg, "path traversal", "Valid ID should not trigger security error")
	})
}
