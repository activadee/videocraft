package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
)

// Test extractSubtitleSettings method
func TestSubtitleService_ExtractSubtitleSettings(t *testing.T) {
	service := setupTestSubtitleService()

	tests := []struct {
		name     string
		project  models.VideoProject
		expected models.SubtitleSettings
	}{
		{
			name: "Extract from global elements",
			project: models.VideoProject{
				Elements: []models.Element{
					{
						Type: "subtitles",
						Settings: models.SubtitleSettings{
							FontFamily: "Test Font",
							FontSize:   30,
							WordColor:  "#FF0000",
						},
					},
				},
			},
			expected: models.SubtitleSettings{
				FontFamily: "Test Font",
				FontSize:   30,
				WordColor:  "#FF0000",
			},
		},
		{
			name: "Extract from scene elements",
			project: models.VideoProject{
				Scenes: []models.Scene{
					{
						Elements: []models.Element{
							{
								Type: "subtitles",
								Settings: models.SubtitleSettings{
									FontFamily: "Scene Font",
									FontSize:   24,
								},
							},
						},
					},
				},
			},
			expected: models.SubtitleSettings{
				FontFamily: "Scene Font",
				FontSize:   24,
			},
		},
		{
			name:     "No subtitle element found",
			project:  models.VideoProject{},
			expected: models.SubtitleSettings{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractSubtitleSettings(tt.project)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test mergeSettingsWithGlobalConfig method
func TestSubtitleService_MergeSettingsWithGlobalConfig(t *testing.T) {
	cfg := &config.Config{
		Subtitles: config.SubtitlesConfig{
			FontFamily: "Arial",
			FontSize:   20,
			Position:   "center-bottom",
			Style:      "progressive",
			Colors: config.ColorConfig{
				Word:    "#FFFFFF",
				Outline: "#000000",
			},
		},
	}

	service := &subtitleService{
		cfg: cfg,
		log: &NoopLogger{},
	}

	tests := []struct {
		name         string
		jsonSettings models.SubtitleSettings
		expectFont   string
		expectSize   int
		expectWord   string
		expectPos    string
	}{
		{
			name:         "Empty JSON settings - use all global config",
			jsonSettings: models.SubtitleSettings{},
			expectFont:   "Arial",
			expectSize:   20,
			expectWord:   "#FFFFFF",
			expectPos:    "center-bottom",
		},
		{
			name: "Partial JSON settings - merge with global config",
			jsonSettings: models.SubtitleSettings{
				FontFamily: "Custom Font",
				FontSize:   30,
				// WordColor and Position should come from global config
			},
			expectFont: "Custom Font",
			expectSize: 30,
			expectWord: "#FFFFFF", // From global config
			expectPos:  "center-bottom", // From global config
		},
		{
			name: "Full JSON settings - override global config",
			jsonSettings: models.SubtitleSettings{
				FontFamily:   "Full Custom",
				FontSize:     40,
				WordColor:    "#FF0000",
				OutlineColor: "#00FF00",
				Position:     "center-top",
			},
			expectFont: "Full Custom",
			expectSize: 40,
			expectWord: "#FF0000",
			expectPos:  "center-top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.mergeSettingsWithGlobalConfig(tt.jsonSettings)
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectFont, result.FontFamily)
			assert.Equal(t, tt.expectSize, result.FontSize)
			assert.Equal(t, tt.expectWord, result.WordColor)
			assert.Equal(t, tt.expectPos, result.Position)
		})
	}
}

// Test ValidateJSONSubtitleSettings method
func TestSubtitleService_ValidateJSONSubtitleSettings(t *testing.T) {
	service := setupTestSubtitleService()

	tests := []struct {
		name      string
		project   models.VideoProject
		shouldErr bool
		errMsg    string
	}{
		{
			name: "Valid settings",
			project: models.VideoProject{
				Elements: []models.Element{
					{
						Type: "subtitles",
						Settings: models.SubtitleSettings{
							FontFamily:   "Arial",
							FontSize:     24,
							WordColor:    "#FFFFFF",
							OutlineColor: "#000000",
							Position:     "center-bottom",
							Style:        "progressive",
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "Invalid font size - too small",
			project: models.VideoProject{
				Elements: []models.Element{
					{
						Type: "subtitles",
						Settings: models.SubtitleSettings{
							FontSize: 5, // Too small
						},
					},
				},
			},
			shouldErr: true,
			errMsg:    "font size must be between 10 and 200",
		},
		{
			name: "Invalid color format",
			project: models.VideoProject{
				Elements: []models.Element{
					{
						Type: "subtitles",
						Settings: models.SubtitleSettings{
							WordColor: "not-a-color",
						},
					},
				},
			},
			shouldErr: true,
			errMsg:    "invalid word color format",
		},
		{
			name: "Invalid position",
			project: models.VideoProject{
				Elements: []models.Element{
					{
						Type: "subtitles",
						Settings: models.SubtitleSettings{
							Position: "invalid-position",
						},
					},
				},
			},
			shouldErr: true,
			errMsg:    "invalid position",
		},
		{
			name:      "No subtitle element - validation passes",
			project:   models.VideoProject{},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateJSONSubtitleSettings(tt.project)

			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}