package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
)

// MockTranscriptionService for testing
type MockTranscriptionService struct {
	results map[string]*TranscriptionResult
}

func (mts *MockTranscriptionService) TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error) {
	if result, exists := mts.results[url]; exists {
		return result, nil
	}
	return &TranscriptionResult{
		Text:    "Test transcription",
		Success: true,
		WordTimestamps: []WhisperWordTimestamp{
			{Word: "Test", Start: 0.0, End: 0.5},
			{Word: "transcription", Start: 0.6, End: 1.2},
		},
	}, nil
}

func (mts *MockTranscriptionService) Shutdown() {}

// MockAudioService for testing
type MockAudioService struct {
	durations map[string]float64
}

func (mas *MockAudioService) AnalyzeAudio(ctx context.Context, url string) (*AudioInfo, error) {
	duration := 30.0 // Default duration
	if d, exists := mas.durations[url]; exists {
		duration = d
	}
	return &AudioInfo{
		URL:      url,
		Duration: duration,
	}, nil
}

func (mas *MockAudioService) CalculateSceneTiming(elements []models.Element) ([]models.TimingSegment, error) {
	return nil, nil
}

func (mas *MockAudioService) DownloadAudio(ctx context.Context, url string) (string, error) {
	return "/tmp/test_audio.mp3", nil
}

// Using existing NoopLogger from ffmpeg_service_security_test.go

// setupTestSubtitleService creates a subtitle service for testing
func setupTestSubtitleService() *subtitleService {
	cfg := &config.Config{
		Subtitles: config.SubtitlesConfig{
			Enabled:    true,
			Style:      "progressive",
			FontFamily: "Arial",
			FontSize:   20,
			Position:   "center-bottom",
			Colors: config.ColorConfig{
				Word:    "#FFFFFF",
				Outline: "#000000",
			},
		},
		Storage: config.StorageConfig{
			TempDir: "/tmp/videocraft",
		},
	}

	mockTranscription := &MockTranscriptionService{
		results: make(map[string]*TranscriptionResult),
	}

	mockAudio := &MockAudioService{
		durations: make(map[string]float64),
	}

	return &subtitleService{
		cfg:           cfg,
		log:           &NoopLogger{},
		transcription: mockTranscription,
		audio:         mockAudio,
	}
}

// Test that current implementation ignores JSON SubtitleSettings (THIS SHOULD FAIL)
func TestSubtitleService_JSONSettingsIgnored_ShouldUseGlobalConfig(t *testing.T) {
	service := setupTestSubtitleService()

	// Create project with custom SubtitleSettings in JSON
	project := models.VideoProject{
		Scenes: []models.Scene{
			{
				Elements: []models.Element{
					{
						Type: "audio",
						Src:  "http://example.com/audio1.mp3",
					},
				},
			},
		},
		Elements: []models.Element{
			{
				Type: "subtitles",
				Settings: models.SubtitleSettings{
					FontFamily:   "Comic Sans MS",  // Different from global config
					FontSize:     48,               // Different from global config  
					WordColor:    "#FF0000",        // Different from global config
					OutlineColor: "#00FF00",        // Different from global config
					OutlineWidth: 5,                // Different from global config
					Position:     "center-top",     // Different from global config
				},
			},
		},
	}

	ctx := context.Background()
	result, err := service.GenerateSubtitles(ctx, project)

	require.NoError(t, err)
	require.NotNil(t, result)

	// Read the generated ASS file to verify settings
	assContent, err := readFileContent(result.FilePath)
	require.NoError(t, err)

	// THIS TEST SHOULD FAIL because current implementation ignores JSON settings
	// and uses only global config. We expect JSON settings to be used instead.
	
	// Check that JSON settings are used (these assertions will FAIL with current code)
	assert.Contains(t, assContent, "Comic Sans MS", "Should use JSON FontFamily, not global config")
	assert.Contains(t, assContent, "48", "Should use JSON FontSize, not global config")
	assert.Contains(t, assContent, "&H000000FF", "Should use JSON WordColor (#FF0000), not global config")
	assert.Contains(t, assContent, "&H0000FF00", "Should use JSON OutlineColor (#00FF00), not global config")
	assert.Contains(t, assContent, ",5,", "Should use JSON OutlineWidth (5), not global config")
	assert.Contains(t, assContent, ",8,", "Should use JSON Position alignment (center-top=8), not global config")
}

// Test that JSON settings override global config (THIS SHOULD FAIL)
func TestSubtitleService_JSONSettingsOverrideGlobalConfig(t *testing.T) {
	service := setupTestSubtitleService()

	// Test cases with different JSON settings
	testCases := []struct {
		name             string
		jsonSettings     models.SubtitleSettings
		expectedInASS    []string
		notExpectedInASS []string
	}{
		{
			name: "Custom font family and size",
			jsonSettings: models.SubtitleSettings{
				FontFamily: "Times New Roman",
				FontSize:   36,
			},
			expectedInASS:    []string{"Times New Roman", "36"},
			notExpectedInASS: []string{"Arial"}, // Global config values (removed "20" as it appears in margins)
		},
		{
			name: "Custom colors",
			jsonSettings: models.SubtitleSettings{
				WordColor:    "#FFFF00", // Yellow
				OutlineColor: "#800080", // Purple
			},
			expectedInASS:    []string{"&H0000FFFF", "&H00800080"}, // ASS color format
			notExpectedInASS: []string{"&H00FFFFFF"}, // Global config colors (removed &H00000000 as it's BackColour default)
		},
		{
			name: "Custom outline and position",
			jsonSettings: models.SubtitleSettings{
				OutlineWidth: 3,
				Position:     "left-top",
			},
			expectedInASS:    []string{",3,", ",7,"}, // OutlineWidth=3, Position=left-top=7
			notExpectedInASS: []string{",2,", ",2,"}, // Global config values
		},
		{
			name: "All JSON settings",
			jsonSettings: models.SubtitleSettings{
				FontFamily:   "Verdana",
				FontSize:     28,
				WordColor:    "#00FFFF", // Cyan
				OutlineColor: "#FF00FF", // Magenta
				OutlineWidth: 4,
				Position:     "right-bottom",
			},
			expectedInASS: []string{
				"Verdana", "28", "&H00FFFF00", "&H00FF00FF", ",4,", ",3,",
			},
			notExpectedInASS: []string{
				"Arial", "&H00FFFFFF", // Removed "20", "&H00000000" and ",2," as they appear in other contexts
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project := models.VideoProject{
				Scenes: []models.Scene{
					{
						Elements: []models.Element{
							{Type: "audio", Src: "http://example.com/audio.mp3"},
						},
					},
				},
				Elements: []models.Element{
					{
						Type:     "subtitles",
						Settings: tc.jsonSettings,
					},
				},
			}

			ctx := context.Background()
			result, err := service.GenerateSubtitles(ctx, project)

			require.NoError(t, err)
			require.NotNil(t, result)

			assContent, err := readFileContent(result.FilePath)
			require.NoError(t, err)

			// These assertions will FAIL because current implementation ignores JSON settings
			for _, expected := range tc.expectedInASS {
				assert.Contains(t, assContent, expected, 
					"ASS file should contain JSON setting value: %s", expected)
			}

			for _, notExpected := range tc.notExpectedInASS {
				assert.NotContains(t, assContent, notExpected,
					"ASS file should NOT contain global config value: %s", notExpected)
			}
		})
	}
}

// Test that missing JSON settings fall back to global config (THIS SHOULD FAIL)
func TestSubtitleService_PartialJSONSettings_FallbackToGlobalConfig(t *testing.T) {
	service := setupTestSubtitleService()

	project := models.VideoProject{
		Scenes: []models.Scene{
			{
				Elements: []models.Element{
					{Type: "audio", Src: "http://example.com/audio.mp3"},
				},
			},
		},
		Elements: []models.Element{
			{
				Type: "subtitles",
				Settings: models.SubtitleSettings{
					// Only specify some settings, others should fall back to global config
					FontFamily: "Custom Font",
					FontSize:   42,
					// WordColor, OutlineColor, OutlineWidth, Position should use global config
				},
			},
		},
	}

	ctx := context.Background()
	result, err := service.GenerateSubtitles(ctx, project)

	require.NoError(t, err)
	require.NotNil(t, result)

	assContent, err := readFileContent(result.FilePath)
	require.NoError(t, err)

	// THIS WILL FAIL: Should use JSON settings where provided
	assert.Contains(t, assContent, "Custom Font", "Should use JSON FontFamily")
	assert.Contains(t, assContent, "42", "Should use JSON FontSize")

	// THIS WILL FAIL: Should fall back to global config for missing fields
	assert.Contains(t, assContent, "&H00FFFFFF", "Should use global WordColor for missing JSON field")
	assert.Contains(t, assContent, "&H00000000", "Should use global OutlineColor for missing JSON field")
	assert.Contains(t, assContent, ",2,", "Should use global OutlineWidth for missing JSON field")
	assert.Contains(t, assContent, ",2,", "Should use global Position for missing JSON field")
}

// NOTE: Tests for createASSFileWithSettings and extended ASSConfig moved to separate test files
// because they cause compilation errors (methods/fields don't exist yet)

// NOTE: ValidateJSONSubtitleSettings test moved to separate file - method doesn't exist yet

// Helper function to read file content
func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// NOTE: Tests for extractSubtitleSettings and mergeSettingsWithGlobalConfig moved to separate file
// because these methods don't exist yet and cause compilation errors