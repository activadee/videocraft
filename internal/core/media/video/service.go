package video

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/activadee/videocraft/internal/api/models"
	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// Service provides video file analysis and processing capabilities
type Service interface {
	AnalyzeVideo(ctx context.Context, videoURL string) (*models.VideoInfo, error)
	DownloadVideo(ctx context.Context, videoURL string) (string, error)
	ValidateVideo(videoURL string) error
	GetVideoMetadata(filePath string) (*models.VideoInfo, error)
}

type service struct {
	cfg *app.Config
	log logger.Logger
}

// NewService creates a new video processing service
func NewService(cfg *app.Config, log logger.Logger) Service {
	return &service{
		cfg: cfg,
		log: log,
	}
}

// AnalyzeVideo analyzes a video file directly from URL using FFprobe
func (s *service) AnalyzeVideo(ctx context.Context, videoURL string) (*models.VideoInfo, error) {
	s.log.Debugf("Analyzing video URL with FFprobe: %s", videoURL)

	// Validate URL first
	if err := s.ValidateVideo(videoURL); err != nil {
		return nil, errors.InvalidInput(fmt.Sprintf("invalid video URL: %v", err))
	}

	// Get metadata using FFprobe directly from URL
	videoInfo, err := s.GetVideoMetadataFromURL(ctx, videoURL)
	if err != nil {
		return nil, errors.ProcessingFailed(fmt.Errorf("failed to get video metadata from URL: %w", err))
	}

	// Set original URL
	videoInfo.URL = videoURL

	s.log.Infof("Video analysis completed: %dx%d, %.2fs, %s",
		videoInfo.Width, videoInfo.Height, videoInfo.Duration, videoInfo.Format)

	return videoInfo, nil
}

// DownloadVideo downloads a video file to temporary storage
func (s *service) DownloadVideo(ctx context.Context, videoURL string) (string, error) {
	s.log.Debugf("Downloading video: %s", videoURL)

	// Create temporary file
	tempDir := s.cfg.Storage.TempDir
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to create temp directory: %w", err))
	}

	// Generate unique filename
	filename := fmt.Sprintf("video_%d.tmp", s.generateTempID())
	tempPath := filepath.Join(tempDir, filename)

	// Download with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to create request: %w", err))
	}

	client := &http.Client{
		Timeout: s.cfg.FFmpeg.Timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to download video: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to download video: HTTP %d", resp.StatusCode))
	}

	// Create output file
	outFile, err := os.Create(tempPath)
	if err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to create temp file: %w", err))
	}
	defer outFile.Close()

	// Copy data
	_, err = outFile.ReadFrom(resp.Body)
	if err != nil {
		os.Remove(tempPath) // Cleanup on error
		return "", errors.ProcessingFailed(fmt.Errorf("failed to write video data: %w", err))
	}

	s.log.Debugf("Video downloaded to: %s", tempPath)
	return tempPath, nil
}

// ValidateVideo validates a video URL for security and format
func (s *service) ValidateVideo(videoURL string) error {
	if videoURL == "" {
		return fmt.Errorf("video URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(videoURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check protocol
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS protocols are allowed")
	}

	// Check for suspicious patterns
	lowerURL := strings.ToLower(videoURL)
	suspiciousPatterns := []string{
		"javascript:", "data:", "file:", "ftp:",
		"../", "..\\", ";", "|", "`",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerURL, pattern) {
			return fmt.Errorf("URL contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// GetVideoMetadata extracts video metadata using FFprobe
func (s *service) GetVideoMetadata(filePath string) (*models.VideoInfo, error) {
	s.log.Debugf("Getting video metadata for: %s", filePath)

	// Build FFprobe command
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	cmd := exec.Command(s.cfg.FFmpeg.FFprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse FFprobe output
	videoInfo, err := s.parseFFprobeOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	return videoInfo, nil
}

// GetVideoMetadataFromURL extracts video metadata directly from URL using FFprobe
func (s *service) GetVideoMetadataFromURL(ctx context.Context, videoURL string) (*models.VideoInfo, error) {
	s.log.Debugf("Getting video metadata from URL: %s", videoURL)

	// Build FFprobe command for URL
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoURL,
	}

	cmd := exec.CommandContext(ctx, s.cfg.FFmpeg.FFprobePath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed for URL %s: %w", videoURL, err)
	}

	// Parse FFprobe output
	videoInfo, err := s.parseFFprobeOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	return videoInfo, nil
}

// parseFFprobeOutput parses FFprobe JSON output
func (s *service) parseFFprobeOutput(output string) (*models.VideoInfo, error) {
	// Simple parsing - in production, would use proper JSON parsing
	videoInfo := &models.VideoInfo{
		Format: "mp4", // default
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse duration
		if strings.Contains(line, `"duration"`) {
			if duration, err := s.extractJSONValue(line, "duration"); err == nil {
				if dur, err := strconv.ParseFloat(duration, 64); err == nil {
					videoInfo.Duration = dur
				}
			}
		}

		// Parse width
		if strings.Contains(line, `"width"`) {
			if width, err := s.extractJSONValue(line, "width"); err == nil {
				if w, err := strconv.Atoi(width); err == nil {
					videoInfo.Width = w
				}
			}
		}

		// Parse height
		if strings.Contains(line, `"height"`) {
			if height, err := s.extractJSONValue(line, "height"); err == nil {
				if h, err := strconv.Atoi(height); err == nil {
					videoInfo.Height = h
				}
			}
		}

		// Parse codec
		if strings.Contains(line, `"codec_name"`) {
			if codec, err := s.extractJSONValue(line, "codec_name"); err == nil {
				videoInfo.Codec = codec
			}
		}
	}

	// Validate required fields
	if videoInfo.Duration <= 0 {
		return nil, fmt.Errorf("invalid duration: %f", videoInfo.Duration)
	}

	if videoInfo.Width <= 0 || videoInfo.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", videoInfo.Width, videoInfo.Height)
	}

	return videoInfo, nil
}

// extractJSONValue extracts a value from a JSON line
func (s *service) extractJSONValue(line, key string) (string, error) {
	// Find the key
	keyPattern := fmt.Sprintf(`"%s":`, key)
	keyIndex := strings.Index(line, keyPattern)
	if keyIndex == -1 {
		return "", fmt.Errorf("key not found")
	}

	// Find the value start
	valueStart := keyIndex + len(keyPattern)
	line = strings.TrimSpace(line[valueStart:])

	// Extract value (handle quoted and unquoted values)
	if strings.HasPrefix(line, `"`) {
		// Quoted string value
		endQuote := strings.Index(line[1:], `"`)
		if endQuote == -1 {
			return "", fmt.Errorf("unterminated string")
		}
		return line[1 : endQuote+1], nil
	} else {
		// Numeric value
		endIndex := strings.IndexAny(line, ",}")
		if endIndex == -1 {
			endIndex = len(line)
		}
		return strings.TrimSpace(line[:endIndex]), nil
	}
}

// generateTempID generates a unique ID for temporary files
func (s *service) generateTempID() int64 {
	// Simple implementation - in production would use more robust ID generation
	return int64(os.Getpid())*1000000 + int64(len(s.cfg.Storage.TempDir))
}
