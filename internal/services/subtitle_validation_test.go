package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/subtitle"
)

// Test the refactored validateMergedConfig method
func TestSubtitleService_ValidateMergedConfig(t *testing.T) {
	service := setupTestSubtitleService()

	tests := []struct {
		name      string
		config    subtitle.ASSConfig
		shouldErr bool
		errMsg    string
	}{
		{
			name: "Valid config",
			config: subtitle.ASSConfig{
				FontFamily:   "Arial",
				FontSize:     24,
				Position:     "center-bottom",
				WordColor:    "#FFFFFF",
				OutlineColor: "#000000",
				OutlineWidth: 2,
				ShadowOffset: 1,
			},
			shouldErr: false,
		},
		{
			name: "Font size too small",
			config: subtitle.ASSConfig{
				FontSize: 5, // Too small
			},
			shouldErr: true,
			errMsg:    "merged font size must be between 6 and 300",
		},
		{
			name: "Font size too large",
			config: subtitle.ASSConfig{
				FontSize: 400, // Too large
			},
			shouldErr: true,
			errMsg:    "merged font size must be between 6 and 300",
		},
		{
			name: "Outline width negative",
			config: subtitle.ASSConfig{
				FontSize:     24,
				OutlineWidth: -1, // Negative
			},
			shouldErr: true,
			errMsg:    "outline width must be between 0 and 20",
		},
		{
			name: "Shadow offset too large",
			config: subtitle.ASSConfig{
				FontSize:     24,
				OutlineWidth: 2,
				ShadowOffset: 25, // Too large
			},
			shouldErr: true,
			errMsg:    "shadow offset must be between 0 and 20",
		},
		{
			name: "Invalid hex color",
			config: subtitle.ASSConfig{
				FontSize:     24,
				OutlineWidth: 2,
				ShadowOffset: 1,
				WordColor:    "#GGGGGG", // Invalid hex
			},
			shouldErr: true,
			errMsg:    "invalid word_color format",
		},
		{
			name: "Valid non-hex color",
			config: subtitle.ASSConfig{
				FontSize:     24,
				OutlineWidth: 2,
				ShadowOffset: 1,
				WordColor:    "red", // Non-hex color (valid)
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateMergedConfig(tt.config)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test the refactored applyJSONSettingsOverrides method
func TestSubtitleService_ApplyJSONSettingsOverrides(t *testing.T) {
	service := setupTestSubtitleService()

	baseConfig := subtitle.ASSConfig{
		FontFamily:   "Arial",
		FontSize:     20,
		Position:     "center-bottom",
		WordColor:    "#FFFFFF",
		OutlineColor: "#000000",
		OutlineWidth: 2,
		ShadowOffset: 1,
		Style:        "progressive",
		LineColor:    "#FFFFFF",
		ShadowColor:  "#808080",
		BoxColor:     "#000000",
	}

	tests := []struct {
		name         string
		jsonSettings models.SubtitleSettings
		expectField  string
		expectValue  interface{}
	}{
		{
			name: "Override font family",
			jsonSettings: models.SubtitleSettings{
				FontFamily: "Custom Font",
			},
			expectField: "FontFamily",
			expectValue: "Custom Font",
		},
		{
			name: "Override font size",
			jsonSettings: models.SubtitleSettings{
				FontSize: 30,
			},
			expectField: "FontSize",
			expectValue: 30,
		},
		{
			name: "Override colors",
			jsonSettings: models.SubtitleSettings{
				WordColor:    "#FF0000",
				OutlineColor: "#00FF00",
				LineColor:    "#0000FF",
			},
			expectField: "WordColor",
			expectValue: "#FF0000",
		},
		{
			name: "Zero values are ignored",
			jsonSettings: models.SubtitleSettings{
				FontSize:     0, // Should be ignored
				OutlineWidth: 0, // Should be ignored
			},
			expectField: "FontSize",
			expectValue: 20, // Should remain base config value
		},
		{
			name: "Empty strings are ignored",
			jsonSettings: models.SubtitleSettings{
				FontFamily: "", // Should be ignored
				Position:   "", // Should be ignored
			},
			expectField: "FontFamily",
			expectValue: "Arial", // Should remain base config value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.applyJSONSettingsOverrides(baseConfig, tt.jsonSettings)

			switch tt.expectField {
			case "FontFamily":
				assert.Equal(t, tt.expectValue, result.FontFamily)
			case "FontSize":
				assert.Equal(t, tt.expectValue, result.FontSize)
			case "WordColor":
				assert.Equal(t, tt.expectValue, result.WordColor)
				// Also check that other colors were applied if provided
				if tt.jsonSettings.OutlineColor != "" {
					assert.Equal(t, tt.jsonSettings.OutlineColor, result.OutlineColor)
				}
				if tt.jsonSettings.LineColor != "" {
					assert.Equal(t, tt.jsonSettings.LineColor, result.LineColor)
				}
			}
		})
	}
}

// Test backward compatibility
func TestSubtitleService_BackwardCompatibility(t *testing.T) {
	service := setupTestSubtitleService()

	events := []subtitle.SubtitleEvent{
		{
			StartTime: 0,
			EndTime:   1000000000, // 1 second in nanoseconds
			Text:      "Test subtitle",
			Layer:     0,
		},
	}

	// Test deprecated createASSFile method still works
	filePath, err := service.createASSFile(events)
	
	require.NoError(t, err)
	require.NotEmpty(t, filePath)

	// Clean up
	defer service.CleanupTempFiles(filePath)

	// Verify file exists and contains expected content
	assContent, err := readFileContent(filePath)
	require.NoError(t, err)
	
	// Should use global config values (Arial, 20, etc.)
	assert.Contains(t, assContent, "Arial", "Should use global config FontFamily")
	assert.Contains(t, assContent, "20", "Should use global config FontSize") 
	assert.Contains(t, assContent, "Test subtitle", "Should contain subtitle text")
}

// Test error handling in createASSFileWithSettings
func TestSubtitleService_CreateASSFileWithSettings_ErrorHandling(t *testing.T) {
	cfg := &config.Config{
		Subtitles: config.SubtitlesConfig{
			FontFamily: "Arial",
			FontSize:   20,
			Position:   "center-bottom",
			Colors: config.ColorConfig{
				Word:    "#FFFFFF",
				Outline: "#000000",
			},
		},
		Storage: config.StorageConfig{
			TempDir: "/invalid/path/that/does/not/exist",
		},
	}

	service := &subtitleService{
		cfg: cfg,
		log: &NoopLogger{},
	}

	events := []subtitle.SubtitleEvent{
		{StartTime: 0, EndTime: 1000000000, Text: "Test", Layer: 0},
	}

	// Should fail due to invalid temp directory
	_, err := service.createASSFileWithSettings(events, models.SubtitleSettings{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create temp directory")
}

// Test mergeSettingsWithGlobalConfig error handling
func TestSubtitleService_MergeSettings_ErrorHandling(t *testing.T) {
	service := setupTestSubtitleService()

	// Test with invalid settings that should fail validation
	invalidSettings := models.SubtitleSettings{
		FontSize: 500, // Too large, should fail validation
	}

	_, err := service.mergeSettingsWithGlobalConfig(invalidSettings)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid merged subtitle configuration")
}