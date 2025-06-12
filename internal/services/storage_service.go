package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/pkg/logger"
)

type storageService struct {
	cfg *config.Config
	log logger.Logger
}

func (s *storageService) StoreVideo(videoPath string) (string, error) {
	s.log.Debugf("Storing video: %s", videoPath)

	// Generate unique video ID
	videoID := uuid.New().String()

	// Ensure output directory exists
	if err := os.MkdirAll(s.cfg.Storage.OutputDir, 0755); err != nil {
		return "", errors.StorageFailed(err)
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
		return "", errors.StorageFailed(err)
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

	// Find file with matching ID (regardless of extension)
	pattern := filepath.Join(s.cfg.Storage.OutputDir, videoID+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", errors.StorageFailed(err)
	}

	if len(matches) == 0 {
		return "", errors.FileNotFound(videoID)
	}

	// Return first match
	videoPath := matches[0]
	
	// Verify file exists
	if _, err := os.Stat(videoPath); err != nil {
		return "", errors.FileNotFound(videoID)
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
		return errors.StorageFailed(err)
	}

	s.log.Infof("Video deleted: %s", videoID)
	return nil
}

func (s *storageService) ListVideos() ([]VideoInfo, error) {
	s.log.Debug("Listing videos")

	pattern := filepath.Join(s.cfg.Storage.OutputDir, "*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.StorageFailed(err)
	}

	videos := make([]VideoInfo, 0, len(matches))

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

		video := VideoInfo{
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
		return errors.StorageFailed(err)
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