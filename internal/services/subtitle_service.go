package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
	"github.com/activadee/videocraft/pkg/subtitle"
)

type subtitleService struct {
	cfg             *config.Config
	log             logger.Logger
	transcription   TranscriptionService
	audio           AudioService
}

type SubtitleResult struct {
	FilePath           string        `json:"file_path"`
	EventCount         int           `json:"event_count"`
	TotalDuration      time.Duration `json:"total_duration"`
	TranscriptionCount int           `json:"transcription_count"`
	Style              string        `json:"style"`
}

func newSubtitleService(cfg *config.Config, log logger.Logger, transcription TranscriptionService, audio AudioService) SubtitleService {
	return &subtitleService{
		cfg:           cfg,
		log:           log,
		transcription: transcription,
		audio:         audio,
	}
}

func (ss *subtitleService) GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error) {
	if !ss.cfg.Subtitles.Enabled {
		ss.log.Debug("Subtitles disabled in configuration")
		return nil, nil
	}

	// Check if project has subtitle element
	var subtitleElement *models.Element
	for _, element := range project.Elements {
		if element.Type == "subtitles" {
			subtitleElement = &element
			break
		}
	}

	if subtitleElement == nil {
		ss.log.Debug("No subtitle element found in project")
		return nil, nil
	}

	ss.log.Info("Generating subtitles for project")

	// Collect audio elements from scenes
	audioElements := ss.collectAudioElements(project)
	if len(audioElements) == 0 {
		ss.log.Debug("No audio elements found for transcription")
		return nil, nil
	}

	// Transcribe audio elements
	transcriptionResults, err := ss.transcribeAudioElements(ctx, audioElements)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Generate subtitle events
	events, err := ss.generateSubtitleEvents(project, transcriptionResults, audioElements)
	if err != nil {
		return nil, fmt.Errorf("failed to generate subtitle events: %w", err)
	}

	if len(events) == 0 {
		ss.log.Debug("No subtitle events generated")
		return nil, nil
	}

	// Create ASS file
	filePath, err := ss.createASSFile(events)
	if err != nil {
		return nil, fmt.Errorf("failed to create ASS file: %w", err)
	}

	// Calculate total duration
	var totalDuration time.Duration
	for _, event := range events {
		if event.EndTime > totalDuration {
			totalDuration = event.EndTime
		}
	}

	result := &SubtitleResult{
		FilePath:           filePath,
		EventCount:         len(events),
		TotalDuration:      totalDuration,
		TranscriptionCount: len(transcriptionResults),
		Style:              ss.cfg.Subtitles.Style,
	}

	ss.log.Infof("Subtitles generated successfully: %d events, %s style, file: %s", 
		len(events), ss.cfg.Subtitles.Style, filePath)

	return result, nil
}

func (ss *subtitleService) collectAudioElements(project models.VideoProject) []models.Element {
	var audioElements []models.Element
	
	// Collect from scenes in order
	for _, scene := range project.Scenes {
		for _, element := range scene.Elements {
			if element.Type == "audio" {
				audioElements = append(audioElements, element)
			}
		}
	}
	
	return audioElements
}

func (ss *subtitleService) transcribeAudioElements(ctx context.Context, audioElements []models.Element) ([]*TranscriptionResult, error) {
	var results []*TranscriptionResult
	
	for i, audio := range audioElements {
		ss.log.Debugf("Transcribing audio %d/%d: %s", i+1, len(audioElements), audio.Src)
		
		result, err := ss.transcription.TranscribeAudio(ctx, audio.Src)
		if err != nil {
			ss.log.Warnf("Failed to transcribe audio %d: %v", i, err)
			// Create failed result
			result = &TranscriptionResult{
				Text:    "",
				Success: false,
			}
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

func (ss *subtitleService) generateSubtitleEvents(
	project models.VideoProject, 
	transcriptionResults []*TranscriptionResult, 
	audioElements []models.Element,
) ([]subtitle.SubtitleEvent, error) {
	var allEvents []subtitle.SubtitleEvent
	
	// Calculate scene timings based on actual audio durations (like Python implementation)
	sceneTimings, err := ss.calculateSceneTimings(transcriptionResults, audioElements)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate scene timings: %w", err)
	}
	
	for i, transcriptionResult := range transcriptionResults {
		if !transcriptionResult.Success || transcriptionResult.Text == "" {
			ss.log.Debugf("Skipping failed transcription %d", i)
			continue
		}
		
		// Get scene timing for this transcription
		var sceneTiming models.TimingSegment
		if i < len(sceneTimings) {
			sceneTiming = sceneTimings[i]
		} else {
			// Fallback if no timing available
			sceneTiming = models.TimingSegment{
				StartTime: float64(i * 5),
				EndTime:   float64((i + 1) * 5),
				AudioFile: "",
			}
		}
		
		var events []subtitle.SubtitleEvent
		
		// Generate events based on style
		if ss.cfg.Subtitles.Style == "progressive" && len(transcriptionResult.WordTimestamps) > 0 {
			// Progressive style - word by word with scene timing
			words := make([]subtitle.WordTimestamp, len(transcriptionResult.WordTimestamps))
			for j, wt := range transcriptionResult.WordTimestamps {
				words[j] = subtitle.WordTimestamp{
					Word:  wt.Word,
					Start: wt.Start,
					End:   wt.End,
				}
			}
			events = subtitle.CreateProgressiveEventsWithSceneTiming(words, sceneTiming)
		} else {
			// Classic style - full text at once
			sceneStartTime := time.Duration(sceneTiming.StartTime * float64(time.Second))
			sceneDuration := time.Duration((sceneTiming.EndTime - sceneTiming.StartTime) * float64(time.Second))
			events = subtitle.CreateClassicEvents(transcriptionResult.Text, sceneStartTime, sceneDuration)
		}
		
		allEvents = append(allEvents, events...)
	}
	
	return allEvents, nil
}

func (ss *subtitleService) calculateSceneTimings(transcriptionResults []*TranscriptionResult, audioElements []models.Element) ([]models.TimingSegment, error) {
	ss.log.Debug("Calculating scene timings based on actual audio file durations (like Python ffprobe)")
	
	var timings []models.TimingSegment
	currentTime := 0.0
	
	for i := range transcriptionResults {
		// Get REAL audio file duration using AudioService (like Python ffprobe)
		var duration float64
		
		if i < len(audioElements) {
			// Use AudioService to analyze actual audio file duration
			ctx := context.Background()
			audioInfo, err := ss.getAudioDuration(ctx, audioElements[i].Src)
			if err != nil {
				ss.log.Warnf("Failed to get audio duration for %s: %v, using fallback", audioElements[i].Src, err)
				duration = 30.0 // Fallback to reasonable default
			} else {
				duration = audioInfo.Duration
				ss.log.Debugf("Real audio duration for scene %d: %.2fs", i, duration)
			}
		} else {
			duration = 30.0 // Default fallback
		}
		
		timing := models.TimingSegment{
			StartTime: currentTime,
			EndTime:   currentTime + duration,
			AudioFile: audioElements[i].Src,
		}
		
		timings = append(timings, timing)
		currentTime += duration
		
		ss.log.Debugf("Scene %d timing: %.2fs - %.2fs (real file duration: %.2fs)", 
			i, timing.StartTime, timing.EndTime, duration)
	}
	
	ss.log.Debugf("Calculated %d scene timings with total duration: %.2fs", len(timings), currentTime)
	return timings, nil
}

func (ss *subtitleService) getAudioDuration(ctx context.Context, audioURL string) (*AudioInfo, error) {
	// Use the existing audio service to get real file duration
	return ss.audio.AnalyzeAudio(ctx, audioURL)
}

func (ss *subtitleService) createASSFile(events []subtitle.SubtitleEvent) (string, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll(ss.cfg.Storage.TempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Generate unique filename
	filename := fmt.Sprintf("subtitles_%s.ass", uuid.New().String()[:8])
	filePath := filepath.Join(ss.cfg.Storage.TempDir, filename)
	
	// Create ASS generator with configuration
	assConfig := subtitle.ASSConfig{
		FontFamily:   ss.cfg.Subtitles.FontFamily,
		FontSize:     ss.cfg.Subtitles.FontSize,
		Position:     ss.cfg.Subtitles.Position,
		WordColor:    ss.cfg.Subtitles.Colors.Word,
		OutlineColor: ss.cfg.Subtitles.Colors.Outline,
		OutlineWidth: 2, // Default outline width
		ShadowOffset: 1, // Default shadow offset
	}
	
	generator := subtitle.NewASSGenerator(assConfig)
	
	// Generate ASS content
	assContent := generator.GenerateASS(events)
	
	// Write to file
	if err := os.WriteFile(filePath, []byte(assContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write ASS file: %w", err)
	}
	
	ss.log.Debugf("ASS file created: %s", filePath)
	return filePath, nil
}

func (ss *subtitleService) ValidateSubtitleConfig(project models.VideoProject) error {
	if !ss.cfg.Subtitles.Enabled {
		return nil
	}
	
	// Check font size
	if ss.cfg.Subtitles.FontSize < 10 || ss.cfg.Subtitles.FontSize > 200 {
		return errors.InvalidInput("font size must be between 10 and 200")
	}
	
	// Validate colors
	if !ss.isValidHexColor(ss.cfg.Subtitles.Colors.Word) {
		return errors.InvalidInput("invalid word color format")
	}
	
	if !ss.isValidHexColor(ss.cfg.Subtitles.Colors.Outline) {
		return errors.InvalidInput("invalid outline color format")
	}
	
	// Validate style
	if ss.cfg.Subtitles.Style != "progressive" && ss.cfg.Subtitles.Style != "classic" {
		return errors.InvalidInput("subtitle style must be 'progressive' or 'classic'")
	}
	
	return nil
}

func (ss *subtitleService) isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	
	for _, c := range color[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	
	return true
}

func (ss *subtitleService) CleanupTempFiles(filePath string) error {
	if filePath == "" {
		return nil
	}
	
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		ss.log.Warnf("Failed to cleanup subtitle file %s: %v", filePath, err)
		return err
	}
	
	ss.log.Debugf("Cleaned up subtitle file: %s", filePath)
	return nil
}