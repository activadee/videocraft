package image

import (
	"context"
	"fmt"
	"image"
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

// Service provides image file processing and validation capabilities
type Service interface {
	ProcessImage(ctx context.Context, imageURL string, targetWidth, targetHeight int) (*models.ImageInfo, error)
	DownloadImage(ctx context.Context, imageURL string) (string, error)
	ValidateImage(imageURL string) error
	ResizeImage(inputPath, outputPath string, width, height int) error
	GetImageInfo(filePath string) (*models.ImageInfo, error)
}

type service struct {
	cfg *app.Config
	log logger.Logger
}

// NewService creates a new image processing service
func NewService(cfg *app.Config, log logger.Logger) Service {
	return &service{
		cfg: cfg,
		log: log,
	}
}

// ProcessImage downloads, validates and processes an image for video composition
func (s *service) ProcessImage(ctx context.Context, imageURL string, targetWidth, targetHeight int) (*models.ImageInfo, error) {
	s.log.Debugf("Processing image: %s", imageURL)

	// Validate URL first
	if err := s.ValidateImage(imageURL); err != nil {
		return nil, errors.InvalidInput(fmt.Sprintf("invalid image URL: %v", err))
	}

	// Download image to temporary location
	tempPath, err := s.DownloadImage(ctx, imageURL)
	if err != nil {
		return nil, err
	}

	// Clean up temporary file when done
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			s.log.Warnf("Failed to cleanup temporary image file %s: %v", tempPath, err)
		}
	}()

	// Get image information
	imageInfo, err := s.GetImageInfo(tempPath)
	if err != nil {
		return nil, errors.ProcessingFailed(fmt.Errorf("failed to get image info: %w", err))
	}

	// Set original URL
	imageInfo.URL = imageURL

	// Resize image if target dimensions are specified
	if targetWidth > 0 && targetHeight > 0 {
		processedPath := s.generateProcessedPath(tempPath)
		if err := s.ResizeImage(tempPath, processedPath, targetWidth, targetHeight); err != nil {
			return nil, errors.ProcessingFailed(fmt.Errorf("failed to resize image: %w", err))
		}

		// Update image info with new dimensions
		imageInfo.Width = targetWidth
		imageInfo.Height = targetHeight
		imageInfo.ProcessedPath = processedPath

		// Clean up processed file
		defer func() {
			if err := os.Remove(processedPath); err != nil {
				s.log.Warnf("Failed to cleanup processed image file %s: %v", processedPath, err)
			}
		}()
	}

	s.log.Infof("Image processing completed: %dx%d, %s",
		imageInfo.Width, imageInfo.Height, imageInfo.Format)

	return imageInfo, nil
}

// DownloadImage downloads an image file to temporary storage
func (s *service) DownloadImage(ctx context.Context, imageURL string) (string, error) {
	s.log.Debugf("Downloading image: %s", imageURL)

	// Create temporary file
	tempDir := s.cfg.Storage.TempDir
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to create temp directory: %w", err))
	}

	// Detect file extension from URL
	extension := s.detectImageExtension(imageURL)
	if extension == "" {
		extension = ".jpg" // Default fallback
	}

	// Generate unique filename with proper extension
	filename := fmt.Sprintf("image_%d%s", s.generateTempID(), extension)
	tempPath := filepath.Join(tempDir, filename)

	// Download with timeout and follow redirects
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to create request: %w", err))
	}

	// Add user agent to avoid blocking
	req.Header.Set("User-Agent", "VideoCraft/1.0 (Image Processor)")

	client := &http.Client{
		Timeout: s.cfg.FFmpeg.Timeout,
		// Allow redirects (default behavior)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to download image: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.ProcessingFailed(fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode))
	}

	// Check Content-Type to ensure we're getting an image
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		s.log.Warnf("Unexpected content type for image URL %s: %s", imageURL, contentType)
		// Continue anyway, some servers don't set proper content types
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
		return "", errors.ProcessingFailed(fmt.Errorf("failed to write image data: %w", err))
	}

	s.log.Debugf("Image downloaded to: %s", tempPath)
	return tempPath, nil
}

// ValidateImage validates an image URL for security and format
func (s *service) ValidateImage(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("image URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check protocol
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS protocols are allowed")
	}

	// Check for suspicious patterns
	lowerURL := strings.ToLower(imageURL)
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

// ResizeImage resizes an image using FFmpeg
func (s *service) ResizeImage(inputPath, outputPath string, width, height int) error {
	s.log.Debugf("Resizing image %s to %dx%d", inputPath, width, height)

	// Determine output format based on extension
	outputExt := strings.ToLower(filepath.Ext(outputPath))
	if outputExt == ".tmp" || outputExt == "" {
		// Default to JPEG if no proper extension
		outputPath = strings.TrimSuffix(outputPath, outputExt) + ".jpg"
		outputExt = ".jpg"
	}

	// Build FFmpeg command for image scaling
	args := []string{
		"-y", // Overwrite output
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
	}

	// Add format-specific options
	switch outputExt {
	case ".jpg", ".jpeg":
		args = append(args, "-q:v", "2") // High quality JPEG
	case ".png":
		args = append(args, "-compression_level", "6") // PNG compression
	case ".webp":
		args = append(args, "-quality", "90") // WebP quality
	default:
		// Default to high quality
		args = append(args, "-q:v", "2")
	}

	args = append(args, outputPath)

	cmd := exec.Command(s.cfg.FFmpeg.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg resize failed: %w, output: %s", err, string(output))
	}

	s.log.Debugf("Image resized successfully: %s", outputPath)
	return nil
}

// GetImageInfo extracts image information
func (s *service) GetImageInfo(filePath string) (*models.ImageInfo, error) {
	s.log.Debugf("Getting image info for: %s", filePath)

	// Get file size first
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if file is too small to be a valid image
	if fileInfo.Size() < 100 {
		return nil, fmt.Errorf("file too small to be a valid image: %d bytes", fileInfo.Size())
	}

	// Try to use FFprobe first for more reliable image detection
	imageInfo, err := s.getImageInfoWithFFprobe(filePath)
	if err == nil {
		return imageInfo, nil
	}

	s.log.Debugf("FFprobe failed, falling back to Go image library: %v", err)

	// Fallback to Go's image library
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode image to get dimensions
	imgConfig, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	imageInfo = &models.ImageInfo{
		Width:  imgConfig.Width,
		Height: imgConfig.Height,
		Format: format,
		Size:   fileInfo.Size(),
		Path:   filePath,
	}

	s.log.Debugf("Image info: %dx%d, %s, %d bytes",
		imageInfo.Width, imageInfo.Height, imageInfo.Format, imageInfo.Size)

	return imageInfo, nil
}

// generateProcessedPath generates a path for processed image files
func (s *service) generateProcessedPath(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// If original has .tmp extension, use .jpg for processed file
	if ext == ".tmp" {
		ext = ".jpg"
	}

	return filepath.Join(dir, fmt.Sprintf("%s_processed%s", name, ext))
}

// getImageInfoWithFFprobe uses FFprobe to get image information
func (s *service) getImageInfoWithFFprobe(filePath string) (*models.ImageInfo, error) {
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

	// Parse FFprobe output for image info
	imageInfo, err := s.parseFFprobeImageOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		imageInfo.Size = fileInfo.Size()
	}
	imageInfo.Path = filePath

	return imageInfo, nil
}

// parseFFprobeImageOutput parses FFprobe JSON output for image information
func (s *service) parseFFprobeImageOutput(output string) (*models.ImageInfo, error) {
	// Simple parsing - would use proper JSON parsing in production
	imageInfo := &models.ImageInfo{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse width
		if strings.Contains(line, `"width"`) {
			if width, err := s.extractJSONValueInt(line, "width"); err == nil {
				imageInfo.Width = width
			}
		}

		// Parse height
		if strings.Contains(line, `"height"`) {
			if height, err := s.extractJSONValueInt(line, "height"); err == nil {
				imageInfo.Height = height
			}
		}

		// Parse codec (format)
		if strings.Contains(line, `"codec_name"`) {
			if codec, err := s.extractJSONValueString(line, "codec_name"); err == nil {
				imageInfo.Format = codec
			}
		}
	}

	// Validate required fields
	if imageInfo.Width <= 0 || imageInfo.Height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %dx%d", imageInfo.Width, imageInfo.Height)
	}

	return imageInfo, nil
}

// extractJSONValueInt extracts an integer value from a JSON line
func (s *service) extractJSONValueInt(line, key string) (int, error) {
	value, err := s.extractJSONValueString(line, key)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(value)
}

// extractJSONValueString extracts a string value from a JSON line
func (s *service) extractJSONValueString(line, key string) (string, error) {
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

// detectImageExtension tries to detect the image file extension from URL
func (s *service) detectImageExtension(imageURL string) string {
	// Parse URL to get path
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return ""
	}

	// Get extension from path
	ext := strings.ToLower(filepath.Ext(parsedURL.Path))

	// Validate it's a supported image extension
	supportedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			return ext
		}
	}

	// Check for common image query parameters (like Google Drive)
	if strings.Contains(imageURL, "image") || strings.Contains(imageURL, "photo") {
		return ".jpg"
	}

	return ""
}

// generateTempID generates a unique ID for temporary files
func (s *service) generateTempID() int64 {
	// Simple implementation - in production would use more robust ID generation
	return int64(os.Getpid())*1000000 + int64(len(s.cfg.Storage.TempDir))
}
