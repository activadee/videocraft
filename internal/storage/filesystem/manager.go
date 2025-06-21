package services

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/activadee/videocraft/internal/api/models"
	"github.com/activadee/videocraft/internal/app"
	domainErrors "github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// Service provides file storage capabilities
type Service interface {
	StoreVideo(videoPath string) (string, error)
	GetVideo(videoID string) (string, error)
	DeleteVideo(videoID string) error
	ListVideos() ([]models.VideoInfo, error)
	CleanupOldFiles() error
}

type storageService struct {
	cfg *app.Config
	log logger.Logger
}

// NewService creates a new storage service
func NewService(cfg *app.Config, log logger.Logger) Service {
	return &storageService{
		cfg: cfg,
		log: log,
	}
}

// Security patterns for path validation
var (
	// Storage-specific path traversal patterns (different name to avoid conflict)
	storagePathTraversalRegex = regexp.MustCompile(`\.\.\/|\.\.\\|\.\.\\\/`)
	// Null byte injection
	nullByteRegex = regexp.MustCompile(`\x00`)
	// Control characters
	controlCharRegex = regexp.MustCompile(`[\x00-\x1f\x7f]`)
	// Valid video ID pattern (alphanumeric, hyphens, underscores)
	validVideoIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
)

func (s *storageService) StoreVideo(videoPath string) (string, error) {
	s.log.Debugf("Storing video: %s", videoPath)

	// Generate unique video ID
	videoID := uuid.New().String()

	// Ensure output directory exists
	if err := os.MkdirAll(s.cfg.Storage.OutputDir, 0755); err != nil {
		return "", domainErrors.StorageFailed(err)
	}

	// Get file extension
	ext := filepath.Ext(videoPath)
	if ext == "" {
		ext = ".mp4"
	}

	// Create destination path
	destPath := filepath.Join(s.cfg.Storage.OutputDir, fmt.Sprintf("%s%s", videoID, ext))

	// Copy file to destination
	if err := s.copyFile(videoPath, destPath); err != nil {
		return "", domainErrors.StorageFailed(err)
	}

	// Remove original temp file
	if err := os.Remove(videoPath); err != nil {
		s.log.Warnf("Failed to remove temp file %s: %v", videoPath, err)
	}

	s.log.Infof("Video stored with ID: %s", videoID)
	return videoID, nil
}

func (s *storageService) GetVideo(videoID string) (string, error) {
	s.log.Debugf("Getting video: %s", videoID)

	// Security validation
	if err := s.validateVideoID(videoID); err != nil {
		s.logSecurityViolation("Invalid video ID provided", map[string]interface{}{
			"video_id": videoID,
			"error":    err.Error(),
		})
		return "", err
	}

	// Sanitize and canonicalize the video ID
	sanitizedID, err := s.sanitizeVideoID(videoID)
	if err != nil {
		return "", err
	}

	// Build safe pattern within output directory
	pattern := filepath.Join(s.cfg.Storage.OutputDir, sanitizedID+".*")

	// Additional security check: ensure pattern is within output directory
	if validateErr := s.validatePathWithinBounds(pattern, s.cfg.Storage.OutputDir); validateErr != nil {
		s.logSecurityViolation("Path outside allowed directory", map[string]interface{}{
			"pattern":    pattern,
			"output_dir": s.cfg.Storage.OutputDir,
		})
		return "", errors.New("path traversal detected")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", domainErrors.StorageFailed(err)
	}

	if len(matches) == 0 {
		return "", domainErrors.FileNotFound(videoID)
	}

	// Security check: verify all matches are within allowed directory
	for _, match := range matches {
		if matchErr := s.validatePathWithinBounds(match, s.cfg.Storage.OutputDir); matchErr != nil {
			s.logSecurityViolation("Match outside allowed directory", map[string]interface{}{
				"match":      match,
				"output_dir": s.cfg.Storage.OutputDir,
			})
			continue
		}
	}

	// Return first valid match
	videoPath := matches[0]

	// Final security check on result path
	if finalErr := s.validatePathWithinBounds(videoPath, s.cfg.Storage.OutputDir); finalErr != nil {
		s.logSecurityViolation("Result path outside allowed directory", map[string]interface{}{
			"video_path": videoPath,
			"output_dir": s.cfg.Storage.OutputDir,
		})
		return "", errors.New("path traversal detected")
	}

	// Verify file exists and is not a symlink
	fileInfo, err := os.Lstat(videoPath)
	if err != nil {
		return "", domainErrors.FileNotFound(videoID)
	}

	// Reject symbolic links to prevent traversal
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		s.logSecurityViolation("Symbolic link access attempt", map[string]interface{}{
			"video_path": videoPath,
			"video_id":   videoID,
		})
		return "", errors.New("symbolic link access not allowed")
	}

	return videoPath, nil
}

func (s *storageService) DeleteVideo(videoID string) error {
	s.log.Debugf("Deleting video: %s", videoID)

	videoPath, err := s.GetVideo(videoID)
	if err != nil {
		return err
	}

	if err := os.Remove(videoPath); err != nil {
		return domainErrors.StorageFailed(err)
	}

	s.log.Infof("Video deleted: %s", videoID)
	return nil
}

func (s *storageService) ListVideos() ([]models.VideoInfo, error) {
	s.log.Debug("Listing videos")

	pattern := filepath.Join(s.cfg.Storage.OutputDir, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, domainErrors.StorageFailed(err)
	}

	videos := make([]models.VideoInfo, 0, len(matches))

	for _, match := range matches {
		// Skip directories
		if info, err := os.Stat(match); err != nil || info.IsDir() {
			continue
		}

		// Extract video ID from filename
		filename := filepath.Base(match)
		ext := filepath.Ext(filename)
		videoID := strings.TrimSuffix(filename, ext)

		// Get file info
		fileInfo, err := os.Stat(match)
		if err != nil {
			s.log.Warnf("Failed to get file info for %s: %v", match, err)
			continue
		}

		video := models.VideoInfo{
			ID:        videoID,
			Filename:  filename,
			Size:      fileInfo.Size(),
			CreatedAt: fileInfo.ModTime().Format(time.RFC3339),
		}

		videos = append(videos, video)
	}

	s.log.Debugf("Found %d videos", len(videos))
	return videos, nil
}

func (s *storageService) CleanupOldFiles() error {
	s.log.Debug("Starting cleanup of old files")

	cutoffTime := time.Now().AddDate(0, 0, -s.cfg.Storage.RetentionDays)

	// Cleanup output directory
	if err := s.cleanupDirectory(s.cfg.Storage.OutputDir, cutoffTime); err != nil {
		return err
	}

	// Cleanup temp directory
	if err := s.cleanupDirectory(s.cfg.Storage.TempDir, cutoffTime); err != nil {
		return err
	}

	s.log.Info("File cleanup completed")
	return nil
}

func (s *storageService) cleanupDirectory(dir string, cutoffTime time.Time) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to clean
	}

	pattern := filepath.Join(dir, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return domainErrors.StorageFailed(err)
	}

	deletedCount := 0

	for _, match := range matches {
		fileInfo, err := os.Stat(match)
		if err != nil {
			continue
		}

		// Skip directories
		if fileInfo.IsDir() {
			continue
		}

		// Delete files older than cutoff time
		if fileInfo.ModTime().Before(cutoffTime) {
			if err := os.Remove(match); err != nil {
				s.log.Warnf("Failed to delete old file %s: %v", match, err)
			} else {
				deletedCount++
				s.log.Debugf("Deleted old file: %s", match)
			}
		}
	}

	if deletedCount > 0 {
		s.log.Infof("Deleted %d old files from %s", deletedCount, dir)
	}

	return nil
}

// validateVideoID checks if video ID is safe and valid
func (s *storageService) validateVideoID(videoID string) error {
	// Check for empty or whitespace-only ID
	if strings.TrimSpace(videoID) == "" {
		return errors.New("empty video ID not allowed")
	}

	// Check for path traversal patterns
	if storagePathTraversalRegex.MatchString(videoID) {
		return errors.New("path traversal detected")
	}

	// Check for null bytes
	if nullByteRegex.MatchString(videoID) {
		return errors.New("null byte injection detected")
	}

	// Check for control characters
	if controlCharRegex.MatchString(videoID) {
		return errors.New("control characters not allowed")
	}

	// Check for absolute paths
	if filepath.IsAbs(videoID) {
		return errors.New("absolute path not allowed")
	}

	// URL decode to catch encoded attacks
	decodedID, err := url.QueryUnescape(videoID)
	if err == nil && decodedID != videoID {
		// Check decoded version for path traversal
		if storagePathTraversalRegex.MatchString(decodedID) {
			return errors.New("path traversal detected")
		}
	}

	// Double decode to catch double-encoded attacks
	doubleDecodedID, err := url.QueryUnescape(decodedID)
	if err == nil && doubleDecodedID != decodedID {
		if storagePathTraversalRegex.MatchString(doubleDecodedID) {
			return errors.New("path traversal detected")
		}
	}

	return nil
}

// sanitizeVideoID cleans and normalizes the video ID
func (s *storageService) sanitizeVideoID(videoID string) (string, error) {
	// Trim whitespace
	sanitized := strings.TrimSpace(videoID)

	// Remove any null bytes
	sanitized = nullByteRegex.ReplaceAllString(sanitized, "")

	// Remove control characters
	sanitized = controlCharRegex.ReplaceAllString(sanitized, "")

	// Ensure it's not empty after sanitization
	if sanitized == "" {
		return "", errors.New("video ID becomes empty after sanitization")
	}

	// Check if it matches valid pattern
	if !validVideoIDRegex.MatchString(sanitized) {
		return "", errors.New("invalid video ID format")
	}

	return sanitized, nil
}

// validatePathWithinBounds ensures path is within allowed directory
func (s *storageService) validatePathWithinBounds(targetPath, allowedDir string) error {
	// Clean and resolve paths
	cleanTarget, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	cleanAllowed, err := filepath.Abs(filepath.Clean(allowedDir))
	if err != nil {
		return fmt.Errorf("failed to resolve allowed directory: %w", err)
	}

	// Check if target is within allowed directory
	relPath, err := filepath.Rel(cleanAllowed, cleanTarget)
	if err != nil {
		return fmt.Errorf("failed to determine relative path: %w", err)
	}

	// If relative path starts with ".." it's outside the allowed directory
	if strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, "/") {
		return errors.New("path traversal detected")
	}

	return nil
}

// logSecurityViolation logs security-related events
func (s *storageService) logSecurityViolation(message string, fields map[string]interface{}) {
	fields["security_event"] = true
	fields["component"] = "storage_service"
	s.log.WithFields(fields).Errorf("SECURITY_VIOLATION: %s", message)
}

func (s *storageService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
