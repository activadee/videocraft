package subtitle

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/activadee/videocraft/internal/domain/models"
)

// ASSGenerator handles ASS subtitle file generation
type ASSGenerator struct {
	config ASSConfig
}

// ASSConfig defines styling configuration for ASS subtitles
type ASSConfig struct {
	FontFamily   string
	FontSize     int
	Position     string
	WordColor    string
	OutlineColor string
	OutlineWidth int
	ShadowOffset int
}

// SubtitleEvent represents a single subtitle event
type SubtitleEvent struct {
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
	Layer     int
}

// NewASSGenerator creates a new ASS generator with configuration
func NewASSGenerator(config ASSConfig) *ASSGenerator {
	return &ASSGenerator{config: config}
}

// GenerateASS creates complete ASS file content from subtitle events
func (g *ASSGenerator) GenerateASS(events []SubtitleEvent) string {
	var builder strings.Builder
	
	// Write header
	builder.WriteString(g.generateHeader())
	builder.WriteString("\n")
	
	// Write events
	builder.WriteString(g.generateEvents(events))
	
	return builder.String()
}

// generateHeader creates the ASS file header with styling
func (g *ASSGenerator) generateHeader() string {
	wordColor := g.parseColorToASS(g.config.WordColor)
	outlineColor := g.parseColorToASS(g.config.OutlineColor)
	alignment := g.getAlignment(g.config.Position)
	
	return fmt.Sprintf(`[Script Info]
Title: Generated Progressive Subtitles
ScriptType: v4.00+
WrapStyle: 0
ScaledBorderAndShadow: yes
YCbCr Matrix: TV.709

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,%s,%d,%s,%s,%s,&H00000000,1,0,0,0,100,100,0,0,1,%d,%d,%d,10,10,20,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text`,
		g.config.FontFamily,
		g.config.FontSize,
		wordColor,
		wordColor, // Secondary color same as primary for progressive
		outlineColor,
		g.config.OutlineWidth,
		g.config.ShadowOffset,
		alignment,
	)
}

// generateEvents creates ASS dialogue events from subtitle events
func (g *ASSGenerator) generateEvents(events []SubtitleEvent) string {
	var builder strings.Builder
	
	for _, event := range events {
		startTime := g.formatASSTime(event.StartTime)
		endTime := g.formatASSTime(event.EndTime)
		cleanText := g.cleanTextForASS(event.Text)
		
		line := fmt.Sprintf("Dialogue: %d,%s,%s,Default,,0,0,0,,%s\n",
			event.Layer,
			startTime,
			endTime,
			cleanText,
		)
		
		builder.WriteString(line)
	}
	
	return builder.String()
}

// formatASSTime converts time.Duration to ASS time format (H:MM:SS.CC)
func (g *ASSGenerator) formatASSTime(duration time.Duration) string {
	totalSeconds := duration.Seconds()
	hours := int(totalSeconds) / 3600
	minutes := (int(totalSeconds) % 3600) / 60
	seconds := int(totalSeconds) % 60
	centiseconds := int((totalSeconds - float64(int(totalSeconds))) * 100)
	
	return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
}

// parseColorToASS converts hex color (#RRGGBB) to ASS format (&HBBGGRR)
func (g *ASSGenerator) parseColorToASS(hexColor string) string {
	// Remove # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")
	
	// Ensure we have 6 characters
	if len(hexColor) != 6 {
		return "&H00FFFFFF" // Default white
	}
	
	// Extract RGB components
	r := hexColor[0:2]
	gComponent := hexColor[2:4]
	b := hexColor[4:6]
	
	// Convert to BGR format for ASS (with alpha channel)
	return fmt.Sprintf("&H00%s%s%s", b, gComponent, r)
}

// getAlignment maps position string to ASS alignment number
func (g *ASSGenerator) getAlignment(position string) int {
	alignmentMap := map[string]int{
		"left-bottom":    1,
		"center-bottom":  2,
		"right-bottom":   3,
		"left-center":    4,
		"center-center":  5,
		"right-center":   6,
		"left-top":       7,
		"center-top":     8,
		"right-top":      9,
	}
	
	if alignment, exists := alignmentMap[position]; exists {
		return alignment
	}
	
	return 2 // Default to center-bottom
}

// cleanTextForASS escapes special characters for ASS format
func (g *ASSGenerator) cleanTextForASS(text string) string {
	// Replace newlines with ASS line breaks
	text = strings.ReplaceAll(text, "\n", "\\N")
	
	// Escape braces
	text = strings.ReplaceAll(text, "{", "\\{")
	text = strings.ReplaceAll(text, "}", "\\}")
	
	// Replace pipe with hard space
	text = strings.ReplaceAll(text, "|", "\\h")
	
	// Clean up extra whitespace
	text = strings.Join(strings.Fields(text), " ")
	
	return text
}

// CreateProgressiveEvents generates word-by-word subtitle events
func CreateProgressiveEvents(words []WordTimestamp, sceneStartTime time.Duration) []SubtitleEvent {
	var events []SubtitleEvent
	
	if len(words) == 0 {
		return events
	}
	
	// Find the actual audio duration from word timestamps
	var maxWordEnd float64
	for _, word := range words {
		if word.End > maxWordEnd {
			maxWordEnd = word.End
		}
	}
	
	// If all words start from the beginning, use them directly (relative timing)
	// If words have a significant offset, normalize them
	var minWordStart float64 = words[0].Start
	for _, word := range words {
		if word.Start < minWordStart {
			minWordStart = word.Start
		}
	}
	
	for i, word := range words {
		if strings.TrimSpace(word.Word) == "" {
			continue
		}
		
		// Normalize timestamps to start from scene beginning
		// This handles both relative (starting near 0) and absolute timestamps
		normalizedStart := word.Start - minWordStart
		normalizedEnd := word.End - minWordStart
		
		startTime := sceneStartTime + time.Duration(normalizedStart*float64(time.Second))
		
		// End time is either the start of the next word or word's end time
		var endTime time.Duration
		if i+1 < len(words) {
			nextWordStart := sceneStartTime + time.Duration((words[i+1].Start-minWordStart)*float64(time.Second))
			endTime = nextWordStart
		} else {
			endTime = sceneStartTime + time.Duration(normalizedEnd*float64(time.Second))
		}
		
		event := SubtitleEvent{
			StartTime: startTime,
			EndTime:   endTime,
			Text:      strings.TrimSpace(word.Word),
			Layer:     0,
		}
		
		events = append(events, event)
	}
	
	return events
}

// CreateProgressiveEventsWithSceneTiming generates word-by-word subtitle events with proper scene timing
func CreateProgressiveEventsWithSceneTiming(words []WordTimestamp, sceneTiming models.TimingSegment) []SubtitleEvent {
	var events []SubtitleEvent
	
	if len(words) == 0 {
		return events
	}
	
	sceneStartTime := time.Duration(sceneTiming.StartTime * float64(time.Second))
	sceneEndTime := time.Duration(sceneTiming.EndTime * float64(time.Second))
	
	for i, word := range words {
		if strings.TrimSpace(word.Word) == "" {
			continue
		}
		
		// Map Whisper timestamps (relative to audio file) to absolute video timeline
		// This is the key fix: use scene timing boundaries properly
		startTime := sceneStartTime + time.Duration(word.Start*float64(time.Second))
		
		// End time is either the start of the next word or word's end time
		var endTime time.Duration
		if i+1 < len(words) {
			nextWordStart := sceneStartTime + time.Duration(words[i+1].Start*float64(time.Second))
			endTime = nextWordStart
		} else {
			endTime = sceneStartTime + time.Duration(word.End*float64(time.Second))
		}
		
		// Ensure we don't exceed scene boundaries
		if startTime < sceneStartTime {
			startTime = sceneStartTime
		}
		if endTime > sceneEndTime {
			endTime = sceneEndTime
		}
		
		event := SubtitleEvent{
			StartTime: startTime,
			EndTime:   endTime,
			Text:      strings.TrimSpace(word.Word),
			Layer:     0,
		}
		
		events = append(events, event)
	}
	
	return events
}

// CreateClassicEvents generates scene-based subtitle events (non-progressive)
func CreateClassicEvents(text string, sceneStartTime, sceneDuration time.Duration) []SubtitleEvent {
	if strings.TrimSpace(text) == "" {
		return []SubtitleEvent{}
	}
	
	event := SubtitleEvent{
		StartTime: sceneStartTime,
		EndTime:   sceneStartTime + sceneDuration,
		Text:      strings.TrimSpace(text),
		Layer:     0,
	}
	
	return []SubtitleEvent{event}
}

// WordTimestamp represents a word with timing information
type WordTimestamp struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}