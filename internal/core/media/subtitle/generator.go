package subtitle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/activadee/videocraft/internal/api/models"
	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/core/media/audio"
	"github.com/activadee/videocraft/internal/core/services/transcription"
	"github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

const (
	subtitleStyleProgressive = "progressive"
)

// Service provides subtitle generation capabilities
type Service interface {
	GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error)
	ValidateSubtitleConfig(project models.VideoProject) error
	ValidateJSONSubtitleSettings(project models.VideoProject) error
	CleanupTempFiles(filePath string) error
}

type service struct {
	cfg           *app.Config
	log           logger.Logger
	transcription TranscriptionService
	audio         AudioService
}

// TranscriptionService interface for dependency injection
type TranscriptionService = transcription.Service

// AudioService interface for dependency injection
type AudioService = audio.Service

// SubtitleResult holds the result of subtitle generation
type SubtitleResult struct {
	FilePath           string        `json:"file_path"`
	EventCount         int           `json:"event_count"`
	TotalDuration      time.Duration `json:"total_duration"`
	TranscriptionCount int           `json:"transcription_count"`
	Style              string        `json:"style"`
}

// NewService creates a new subtitle service
func NewService(cfg *app.Config, log logger.Logger, transcription TranscriptionService, audio AudioService) Service {
	return &service{
		cfg:           cfg,
		log:           log,
		transcription: transcription,
		audio:         audio,
	}
}

// Deprecated: Use NewService instead
func newSubtitleService(cfg *app.Config, log logger.Logger) Service {
	return NewService(cfg, log, nil, nil)
}

func (ss *service) GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error) {
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

	// Extract subtitle settings from project
	subtitleSettings := ss.extractSubtitleSettings(project)

	// Create ASS file with settings
	filePath, err := ss.createASSFileWithSettings(events, subtitleSettings)
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

func (ss *service) collectAudioElements(project models.VideoProject) []models.Element {
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

func (ss *service) transcribeAudioElements(ctx context.Context, audioElements []models.Element) ([]*transcription.TranscriptionResult, error) {
	var results []*transcription.TranscriptionResult

	for i, audio := range audioElements {
		ss.log.Debugf("Transcribing audio %d/%d: %s", i+1, len(audioElements), audio.Src)

		result, err := ss.transcription.TranscribeAudio(ctx, audio.Src)
		if err != nil {
			ss.log.Warnf("Failed to transcribe audio %d: %v", i, err)
			// Create failed result
			result = &transcription.TranscriptionResult{
				Text:    "",
				Success: false,
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (ss *service) generateSubtitleEvents(
	project models.VideoProject,
	transcriptionResults []*transcription.TranscriptionResult,
	audioElements []models.Element,
) ([]SubtitleEvent, error) {
	var allEvents []SubtitleEvent

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

		var events []SubtitleEvent

		// Generate events based on style
		if ss.cfg.Subtitles.Style == subtitleStyleProgressive && len(transcriptionResult.WordTimestamps) > 0 {
			// Progressive style - word by word with scene timing
			words := make([]WordTimestamp, len(transcriptionResult.WordTimestamps))
			for j, wt := range transcriptionResult.WordTimestamps {
				words[j] = WordTimestamp{
					Word:  wt.Word,
					Start: wt.Start,
					End:   wt.End,
				}
			}
			events = CreateProgressiveEventsWithSceneTiming(words, sceneTiming)
		} else {
			// Classic style - full text at once
			sceneStartTime := time.Duration(sceneTiming.StartTime * float64(time.Second))
			sceneDuration := time.Duration((sceneTiming.EndTime - sceneTiming.StartTime) * float64(time.Second))
			events = CreateClassicEvents(transcriptionResult.Text, sceneStartTime, sceneDuration)
		}

		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

func (ss *service) calculateSceneTimings(transcriptionResults []*transcription.TranscriptionResult, audioElements []models.Element) ([]models.TimingSegment, error) {
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

func (ss *service) getAudioDuration(ctx context.Context, audioURL string) (*audio.AudioInfo, error) {
	// Use the existing audio service to get real file duration
	return ss.audio.AnalyzeAudio(ctx, audioURL)
}

// createASSFile provides backward compatibility by using global config only
// Deprecated: This method is maintained for backward compatibility but doesn't support JSON settings
// Use createASSFileWithSettings for new implementations that need JSON SubtitleSettings support
func (ss *service) createASSFile(events []SubtitleEvent) (string, error) {
	// For backward compatibility, delegate to new method with empty settings (uses global config)
	return ss.createASSFileWithSettings(events, models.SubtitleSettings{})
}

func (ss *service) ValidateSubtitleConfig(project models.VideoProject) error {
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

func (ss *service) isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}

	for _, c := range color[1:] {
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') && (c < 'a' || c > 'f') {
			return false
		}
	}

	return true
}

func (ss *service) CleanupTempFiles(filePath string) error {
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

// extractSubtitleSettings extracts SubtitleSettings from a VideoProject
// Looks for subtitle elements in both global elements and scene elements
// Returns empty SubtitleSettings if no subtitle element is found
func (ss *service) extractSubtitleSettings(project models.VideoProject) models.SubtitleSettings {
	// Look for subtitle element in project
	for _, element := range project.Elements {
		if element.Type == "subtitles" {
			return element.Settings
		}
	}

	// Check scenes for subtitle elements
	for _, scene := range project.Scenes {
		for _, element := range scene.Elements {
			if element.Type == "subtitles" {
				return element.Settings
			}
		}
	}

	// Return empty settings if no subtitle element found
	return models.SubtitleSettings{}
}

// createASSFileWithSettings creates ASS file using provided SubtitleSettings
// This method replaces the original createASSFile to support JSON subtitle configuration
// The provided settings are merged with global config before ASS generation
func (ss *service) createASSFileWithSettings(events []SubtitleEvent, settings models.SubtitleSettings) (string, error) {
	// Ensure temp directory exists
	if err := os.MkdirAll(ss.cfg.Storage.TempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate unique filename
	filename := fmt.Sprintf("subtitles_%s.ass", uuid.New().String()[:8])
	filePath := filepath.Join(ss.cfg.Storage.TempDir, filename)

	// Merge JSON settings with global config to create ASS config
	assConfig, err := ss.mergeSettingsWithGlobalConfig(settings)
	if err != nil {
		return "", fmt.Errorf("failed to merge settings: %w", err)
	}

	// Create ASS generator with merged configuration
	generator := NewASSGenerator(assConfig)

	// Generate ASS content
	assContent := generator.GenerateASS(events)

	// Write to file
	if err := os.WriteFile(filePath, []byte(assContent), 0600); err != nil {
		return "", fmt.Errorf("failed to write ASS file: %w", err)
	}

	ss.log.Debugf("ASS file created with custom settings: %s", filePath)
	return filePath, nil
}

// mergeSettingsWithGlobalConfig merges JSON SubtitleSettings with global config
// JSON settings take precedence over global config, with global config as fallback
func (ss *service) mergeSettingsWithGlobalConfig(jsonSettings models.SubtitleSettings) (ASSConfig, error) {
	// Check for nil configuration
	if ss.cfg == nil {
		return ASSConfig{}, fmt.Errorf("subtitle service configuration is nil")
	}

	// Start with global config as base, providing sensible defaults
	config := ASSConfig{
		FontFamily:   ss.cfg.Subtitles.FontFamily,
		FontSize:     ss.cfg.Subtitles.FontSize,
		Position:     ss.cfg.Subtitles.Position,
		WordColor:    ss.cfg.Subtitles.Colors.Word,
		OutlineColor: ss.cfg.Subtitles.Colors.Outline,
		OutlineWidth: 2, // TODO: Add OutlineWidth to global config to avoid hard-coded defaults
		ShadowOffset: 1, // TODO: Add ShadowOffset to global config to avoid hard-coded defaults
		Style:        ss.cfg.Subtitles.Style,
		LineColor:    ss.cfg.Subtitles.Colors.Word, // Default line color same as word color
		ShadowColor:  "#808080",                    // TODO: Add ShadowColor to global config to avoid hard-coded defaults
		BoxColor:     "#000000",                    // TODO: Add BoxColor to global config to avoid hard-coded defaults
	}

	// Use helper function to override with JSON settings where provided
	config = ss.applyJSONSettingsOverrides(config, jsonSettings)

	// Fix LineColor: if not explicitly set in JSON, use the final WordColor (after JSON override)
	if jsonSettings.LineColor == "" {
		config.LineColor = config.WordColor
	}

	// Validate merged configuration
	if err := ss.validateMergedConfig(config); err != nil {
		return config, fmt.Errorf("invalid merged subtitle configuration: %w", err)
	}

	ss.log.Debugf("Merged subtitle settings: JSON overrides applied to global config")
	return config, nil
}

// applyJSONSettingsOverrides applies non-empty JSON settings to the base config
func (ss *service) applyJSONSettingsOverrides(baseConfig ASSConfig, jsonSettings models.SubtitleSettings) ASSConfig {
	config := baseConfig

	// String fields: override if non-empty
	if jsonSettings.FontFamily != "" {
		config.FontFamily = jsonSettings.FontFamily
	}
	if jsonSettings.Position != "" {
		config.Position = jsonSettings.Position
	}
	if jsonSettings.WordColor != "" {
		config.WordColor = jsonSettings.WordColor
	}
	if jsonSettings.OutlineColor != "" {
		config.OutlineColor = jsonSettings.OutlineColor
	}
	if jsonSettings.Style != "" {
		config.Style = jsonSettings.Style
	}
	if jsonSettings.LineColor != "" {
		config.LineColor = jsonSettings.LineColor
	}
	if jsonSettings.ShadowColor != "" {
		config.ShadowColor = jsonSettings.ShadowColor
	}
	if jsonSettings.BoxColor != "" {
		config.BoxColor = jsonSettings.BoxColor
	}

	// Integer fields: override if non-zero
	if jsonSettings.FontSize != 0 {
		config.FontSize = jsonSettings.FontSize
	}
	if jsonSettings.OutlineWidth != 0 {
		config.OutlineWidth = jsonSettings.OutlineWidth
	}
	if jsonSettings.ShadowOffset != 0 {
		config.ShadowOffset = jsonSettings.ShadowOffset
	}

	return config
}

// validateMergedConfig validates the final merged configuration
func (ss *service) validateMergedConfig(config ASSConfig) error {
	// Validate font size
	if config.FontSize < 6 || config.FontSize > 300 {
		return errors.InvalidInput("merged font size must be between 6 and 300")
	}

	// Validate outline width
	if config.OutlineWidth < 0 || config.OutlineWidth > 20 {
		return errors.InvalidInput("outline width must be between 0 and 20")
	}

	// Validate shadow offset
	if config.ShadowOffset < 0 || config.ShadowOffset > 20 {
		return errors.InvalidInput("shadow offset must be between 0 and 20")
	}

	// Validate colors if they look like hex colors
	colorFields := map[string]string{
		"word_color":    config.WordColor,
		"outline_color": config.OutlineColor,
		"line_color":    config.LineColor,
		"shadow_color":  config.ShadowColor,
		"box_color":     config.BoxColor,
	}

	for fieldName, color := range colorFields {
		if color != "" && strings.HasPrefix(color, "#") && !ss.isValidHexColor(color) {
			return errors.InvalidInput(fmt.Sprintf("invalid %s format: %s", fieldName, color))
		}
	}

	return nil
}

// ValidateJSONSubtitleSettings validates SubtitleSettings from JSON
func (ss *service) ValidateJSONSubtitleSettings(project models.VideoProject) error {
	settings := ss.extractSubtitleSettings(project)

	// If no subtitle settings found, validation passes
	if settings == (models.SubtitleSettings{}) {
		return nil
	}

	// Validate font size
	if settings.FontSize != 0 && (settings.FontSize < 10 || settings.FontSize > 200) {
		return errors.InvalidInput("font size must be between 10 and 200")
	}

	// Validate colors (if provided)
	if settings.WordColor != "" && !ss.isValidHexColor(settings.WordColor) {
		return errors.InvalidInput("invalid word color format")
	}
	if settings.OutlineColor != "" && !ss.isValidHexColor(settings.OutlineColor) {
		return errors.InvalidInput("invalid outline color format")
	}
	if settings.LineColor != "" && !ss.isValidHexColor(settings.LineColor) {
		return errors.InvalidInput("invalid line color format")
	}
	if settings.ShadowColor != "" && !ss.isValidHexColor(settings.ShadowColor) {
		return errors.InvalidInput("invalid shadow color format")
	}
	if settings.BoxColor != "" && !ss.isValidHexColor(settings.BoxColor) {
		return errors.InvalidInput("invalid box color format")
	}

	// Validate position (if provided)
	if settings.Position != "" {
		validPositions := map[string]bool{
			"left-bottom": true, "center-bottom": true, "right-bottom": true,
			"left-center": true, "center-center": true, "right-center": true,
			"left-top": true, "center-top": true, "right-top": true,
		}
		if !validPositions[settings.Position] {
			return errors.InvalidInput("invalid position")
		}
	}

	// Validate style (if provided)
	if settings.Style != "" && settings.Style != "progressive" && settings.Style != "classic" {
		return errors.InvalidInput("subtitle style must be 'progressive' or 'classic'")
	}

	return nil
}
