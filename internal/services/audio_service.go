package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
	"github.com/google/uuid"
)

type audioService struct {
	cfg *config.Config
	log logger.Logger
}

func (s *audioService) AnalyzeAudio(ctx context.Context, url string) (*AudioInfo, error) {
	s.log.Debugf("Analyzing audio: %s", url)

	// Download audio file temporarily
	tempPath, err := s.DownloadAudio(ctx, url)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempPath)

	// Use FFprobe to get comprehensive audio information
	audioInfo, err := s.getAudioInfo(tempPath)
	if err != nil {
		return nil, errors.InternalError(fmt.Errorf("failed to get audio info: %w", err))
	}

	// Update URL to original URL (not temp path)
	audioInfo.URL = url

	s.log.Debugf("Audio analysis complete: duration=%.2fs, format=%s, bitrate=%d",
		audioInfo.Duration, audioInfo.Format, audioInfo.Bitrate)

	return audioInfo, nil
}

func (s *audioService) CalculateSceneTiming(elements []models.Element) ([]models.TimingSegment, error) {
	s.log.Debug("Calculating scene timing from elements")

	// Extract audio elements and group by scene
	audioElements := make([]models.Element, 0)
	for _, element := range elements {
		if element.Type == "audio" {
			audioElements = append(audioElements, element)
		}
	}

	segments := make([]models.TimingSegment, 0, len(audioElements))
	currentTime := 0.0

	for i, audio := range audioElements {
		// Analyze audio to get duration
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		audioInfo, err := s.AnalyzeAudio(ctx, audio.Src)
		cancel()

		if err != nil {
			return nil, fmt.Errorf("failed to analyze audio element %d: %w", i, err)
		}

		duration := audio.Duration
		if duration <= 0 {
			duration = audioInfo.Duration
		}

		segment := models.TimingSegment{
			StartTime: currentTime,
			EndTime:   currentTime + duration,
			AudioFile: audio.Src,
		}

		segments = append(segments, segment)
		currentTime += duration
	}

	s.log.Debugf("Calculated %d timing segments with total duration: %.2f seconds", len(segments), currentTime)
	return segments, nil
}

func (s *audioService) DownloadAudio(ctx context.Context, url string) (string, error) {
	s.log.Debugf("Downloading audio: %s", url)

	// Resolve Google Drive URLs
	downloadURL := s.resolveGoogleDriveURL(url)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", errors.DownloadFailed(url, err)
	}

	// Execute request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.DownloadFailed(url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.DownloadFailed(url, fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	// Determine file extension
	ext := s.getFileExtension(resp.Header.Get("Content-Type"), url)

	// Create temporary file
	tempFile := filepath.Join(s.cfg.Storage.TempDir, fmt.Sprintf("audio_%s%s", uuid.New().String()[:8], ext))

	// Ensure temp directory exists
	if err := os.MkdirAll(s.cfg.Storage.TempDir, 0755); err != nil {
		return "", errors.StorageFailed(err)
	}

	// Create output file
	out, err := os.Create(tempFile)
	if err != nil {
		return "", errors.StorageFailed(err)
	}
	defer out.Close()

	// Copy data
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return "", errors.DownloadFailed(url, err)
	}

	s.log.Debugf("Audio downloaded to: %s", tempFile)
	return tempFile, nil
}

func (s *audioService) resolveGoogleDriveURL(url string) string {
	if !strings.Contains(url, "drive.google.com") {
		return url
	}

	// Extract file ID from various Google Drive URL formats
	patterns := []string{
		"/file/d/",
		"id=",
		"/d/",
	}

	var fileID string
	for _, pattern := range patterns {
		if idx := strings.Index(url, pattern); idx != -1 {
			start := idx + len(pattern)
			end := start
			for end < len(url) && url[end] != '/' && url[end] != '&' && url[end] != '?' {
				end++
			}
			fileID = url[start:end]
			break
		}
	}

	if fileID != "" {
		return fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", fileID)
	}

	return url
}

func (s *audioService) getFileExtension(contentType, url string) string {
	// Try to get extension from Content-Type
	if strings.Contains(contentType, "audio") {
		if strings.Contains(contentType, "wav") {
			return ".wav"
		} else if strings.Contains(contentType, "mp3") {
			return ".mp3"
		} else if strings.Contains(contentType, "ogg") {
			return ".ogg"
		}
	}

	// Try to get extension from URL
	if ext := filepath.Ext(url); ext != "" {
		return ext
	}

	// Default
	return ".mp3"
}

func (s *audioService) getAudioDuration(filePath string) (float64, error) {
	s.log.Debugf("Getting audio duration for: %s", filePath)

	// Use FFprobe to get duration
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	return parseDuration(string(output))
}

func (s *audioService) getAudioInfo(filePath string) (*AudioInfo, error) {
	s.log.Debugf("Getting audio info for: %s", filePath)

	// Use FFprobe to get comprehensive audio information
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	return s.parseAudioInfo(string(output), filePath)
}

func (s *audioService) parseAudioInfo(jsonOutput, filePath string) (*AudioInfo, error) {
	var probe FFProbeOutput
	if err := json.Unmarshal([]byte(jsonOutput), &probe); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Parse duration
	duration, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	// Parse bitrate
	bitrate, _ := strconv.Atoi(probe.Format.BitRate)

	// Parse size
	size, _ := strconv.ParseInt(probe.Format.Size, 10, 64)

	// Get audio stream info
	var format string
	for _, stream := range probe.Streams {
		if stream.CodecType == "audio" {
			format = stream.CodecName
			break
		}
	}

	return &AudioInfo{
		URL:      filePath,
		Duration: duration,
		Format:   format,
		Bitrate:  bitrate,
		Size:     size,
	}, nil
}

// FFProbe output structures
type FFProbeOutput struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	CodecName  string `json:"codec_name"`
	CodecType  string `json:"codec_type"`
	BitRate    string `json:"bit_rate"`
	SampleRate string `json:"sample_rate"`
	Channels   int    `json:"channels"`
}

type Format struct {
	Filename   string `json:"filename"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	FormatName string `json:"format_name"`
}

// Helper function to parse duration from FFprobe output
func parseDuration(output string) (float64, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return 0, fmt.Errorf("empty duration output")
	}

	duration, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}
