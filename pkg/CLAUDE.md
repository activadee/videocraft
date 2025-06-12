# PKG Package - Shared Utilities and Libraries

## Overview
The `pkg` package contains reusable utilities and libraries that can be shared across different parts of the VideoCraft application or potentially with external projects. This package follows Go's convention for publicly available, stable APIs that other projects can import and use.

## Architecture

```
pkg/
├── logger/
│   └── logger.go         # Structured logging interface and implementation
├── subtitle/
│   └── ass_generator.go  # ASS subtitle file generation utilities
├── metrics/              # Metrics collection (placeholder for future implementation)
└── CLAUDE.md            # This documentation
```

## Core Packages

### Logger Package

**Location**: `pkg/logger/logger.go`

Provides a clean, structured logging interface built on top of logrus, with support for contextual logging and multiple output formats.

#### Interface Definition

```go
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
    
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})
    
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
}
```

#### Key Features

1. **Structured Logging**: Support for key-value pair logging
2. **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal
3. **Contextual Logging**: Add context with WithField/WithFields
4. **Configurable Output**: Support for different formatters
5. **Interface-Based**: Easy to mock for testing

#### Implementation

```go
type logger struct {
    log *logrus.Logger
}

func New(level string) Logger {
    log := logrus.New()
    log.SetOutput(os.Stdout)
    
    // Set log level
    switch level {
    case "debug":
        log.SetLevel(logrus.DebugLevel)
    case "info":
        log.SetLevel(logrus.InfoLevel)
    case "warn":
        log.SetLevel(logrus.WarnLevel)
    case "error":
        log.SetLevel(logrus.ErrorLevel)
    default:
        log.SetLevel(logrus.InfoLevel)
    }
    
    // Set formatter
    log.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
        ForceColors:   true,
    })
    
    return &logger{log: log}
}
```

#### Usage Examples

**Basic Logging**:
```go
package main

import "github.com/activadee/videocraft/pkg/logger"

func main() {
    log := logger.New("info")
    
    log.Info("Application starting")
    log.Debugf("Debug message with data: %v", data)
    log.Error("Something went wrong")
}
```

**Contextual Logging**:
```go
func processVideo(videoID string, log logger.Logger) {
    videoLog := log.WithFields(map[string]interface{}{
        "video_id": videoID,
        "function": "processVideo",
    })
    
    videoLog.Info("Starting video processing")
    
    // Process video...
    
    videoLog.WithField("duration", "45s").Info("Video processing completed")
}
```

**Service Integration**:
```go
type VideoService struct {
    log logger.Logger
}

func NewVideoService(log logger.Logger) *VideoService {
    return &VideoService{
        log: log.WithField("service", "video"),
    }
}

func (vs *VideoService) GenerateVideo(ctx context.Context, config *models.VideoConfig) error {
    requestLog := vs.log.WithFields(map[string]interface{}{
        "request_id": ctx.Value("request_id"),
        "video_width": config.Width,
        "video_height": config.Height,
    })
    
    requestLog.Info("Generating video")
    
    // Generate video...
    
    requestLog.Info("Video generation completed")
    return nil
}
```

#### Advanced Configuration

**JSON Formatter for Production**:
```go
func NewProductionLogger(level string) logger.Logger {
    log := logrus.New()
    
    // JSON formatter for structured logs
    log.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    
    // Set level
    logLevel, _ := logrus.ParseLevel(level)
    log.SetLevel(logLevel)
    
    return &logger{log: log}
}
```

**File Output**:
```go
func NewFileLogger(level, filename string) (logger.Logger, error) {
    log := logrus.New()
    
    file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, err
    }
    
    log.SetOutput(file)
    
    // Configure level and formatter
    logLevel, _ := logrus.ParseLevel(level)
    log.SetLevel(logLevel)
    
    return &logger{log: log}, nil
}
```

### Subtitle Package

**Location**: `pkg/subtitle/ass_generator.go`

Provides utilities for generating Advanced SubStation Alpha (ASS) subtitle files, supporting both progressive (word-by-word) and classic subtitle styles.

#### Core Types

```go
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

// SubtitleEvent represents a single subtitle display event
type SubtitleEvent struct {
    StartTime time.Duration
    EndTime   time.Duration
    Text      string
    Layer     int
}

// WordTimestamp represents a word with timing information
type WordTimestamp struct {
    Word  string  `json:"word"`
    Start float64 `json:"start"`
    End   float64 `json:"end"`
}
```

#### Key Functions

##### `NewASSGenerator(config ASSConfig) *ASSGenerator`
Creates a new ASS generator with specified styling configuration.

**Configuration Example**:
```go
config := ASSConfig{
    FontFamily:   "Arial",
    FontSize:     24,
    Position:     "center-bottom",
    WordColor:    "#FFFFFF",
    OutlineColor: "#000000",
    OutlineWidth: 2,
    ShadowOffset: 1,
}

generator := NewASSGenerator(config)
```

##### `GenerateASS(events []SubtitleEvent) string`
Generates complete ASS file content from subtitle events.

**Usage Example**:
```go
events := []SubtitleEvent{
    {
        StartTime: time.Duration(0 * time.Second),
        EndTime:   time.Duration(2 * time.Second),
        Text:      "Hello",
        Layer:     0,
    },
    {
        StartTime: time.Duration(2 * time.Second),
        EndTime:   time.Duration(4 * time.Second),
        Text:      "World",
        Layer:     0,
    },
}

assContent := generator.GenerateASS(events)
```

**Generated ASS Output**:
```
[Script Info]
Title: Generated Progressive Subtitles
ScriptType: v4.00+
WrapStyle: 0
ScaledBorderAndShadow: yes
YCbCr Matrix: TV.709

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,24,&H00FFFFFF,&H00FFFFFF,&H00000000,&H00000000,1,0,0,0,100,100,0,0,1,2,1,2,10,10,20,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,0:00:02.00,Default,,0,0,0,,Hello
Dialogue: 0,0:00:02.00,0:00:04.00,Default,,0,0,0,,World
```

##### Progressive Subtitle Functions

**`CreateProgressiveEventsWithSceneTiming(words []WordTimestamp, sceneTiming models.TimingSegment) []SubtitleEvent`**

Creates word-by-word subtitle events with proper scene timing alignment.

```go
// Word timestamps from Whisper (relative to audio file)
words := []WordTimestamp{
    {Word: "Hello", Start: 0.0, End: 0.8},
    {Word: "world", Start: 1.2, End: 1.8},
}

// Scene timing in video timeline
sceneTiming := models.TimingSegment{
    StartTime: 10.0,  // Scene starts at 10 seconds
    EndTime:   15.0,  // Scene ends at 15 seconds
}

// Generate progressive events
events := CreateProgressiveEventsWithSceneTiming(words, sceneTiming)

// Result: 
// Event 1: "Hello" from 10.0s to 11.2s
// Event 2: "world" from 11.2s to 11.8s
```

**`CreateProgressiveEvents(words []WordTimestamp, sceneStartTime time.Duration) []SubtitleEvent`**

Creates progressive events with normalized timing.

```go
words := []WordTimestamp{
    {Word: "Hello", Start: 0.0, End: 0.8},
    {Word: "world", Start: 1.2, End: 1.8},
}

sceneStart := time.Duration(5 * time.Second)
events := CreateProgressiveEvents(words, sceneStart)
```

**`CreateClassicEvents(text string, sceneStartTime, sceneDuration time.Duration) []SubtitleEvent`**

Creates traditional subtitle events (non-progressive).

```go
text := "Hello world"
startTime := time.Duration(10 * time.Second)
duration := time.Duration(3 * time.Second)

events := CreateClassicEvents(text, startTime, duration)
// Result: Single event "Hello world" from 10s to 13s
```

#### Color and Position Handling

**Color Conversion**:
```go
// parseColorToASS converts hex color (#RRGGBB) to ASS format (&HBBGGRR)
func (g *ASSGenerator) parseColorToASS(hexColor string) string {
    // Input: "#FFFFFF" (white)
    // Output: "&H00FFFFFF" (ASS white with alpha)
    
    // Input: "#FF0000" (red)
    // Output: "&H000000FF" (ASS red in BGR format)
}
```

**Position Mapping**:
```go
func (g *ASSGenerator) getAlignment(position string) int {
    alignmentMap := map[string]int{
        "left-bottom":    1,
        "center-bottom":  2,  // Default
        "right-bottom":   3,
        "left-center":    4,
        "center-center":  5,
        "right-center":   6,
        "left-top":       7,
        "center-top":     8,
        "right-top":      9,
    }
    
    return alignmentMap[position]
}
```

#### Time Formatting

```go
// formatASSTime converts time.Duration to ASS time format (H:MM:SS.CC)
func (g *ASSGenerator) formatASSTime(duration time.Duration) string {
    // Input: 65.75 seconds
    // Output: "0:01:05.75"
    
    totalSeconds := duration.Seconds()
    hours := int(totalSeconds) / 3600
    minutes := (int(totalSeconds) % 3600) / 60
    seconds := int(totalSeconds) % 60
    centiseconds := int((totalSeconds - float64(int(totalSeconds))) * 100)
    
    return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
}
```

#### Text Sanitization

```go
// cleanTextForASS escapes special characters for ASS format
func (g *ASSGenerator) cleanTextForASS(text string) string {
    // Replace newlines with ASS line breaks
    text = strings.ReplaceAll(text, "\n", "\\N")
    
    // Escape braces (used for ASS commands)
    text = strings.ReplaceAll(text, "{", "\\{")
    text = strings.ReplaceAll(text, "}", "\\}")
    
    // Replace pipe with hard space
    text = strings.ReplaceAll(text, "|", "\\h")
    
    // Clean up extra whitespace
    text = strings.Join(strings.Fields(text), " ")
    
    return text
}
```

## Integration Examples

### Service Integration

```go
package services

import (
    "github.com/activadee/videocraft/pkg/logger"
    "github.com/activadee/videocraft/pkg/subtitle"
)

type SubtitleService struct {
    log       logger.Logger
    generator *subtitle.ASSGenerator
}

func NewSubtitleService(log logger.Logger) *SubtitleService {
    config := subtitle.ASSConfig{
        FontFamily:   "Arial",
        FontSize:     24,
        Position:     "center-bottom",
        WordColor:    "#FFFFFF",
        OutlineColor: "#000000",
        OutlineWidth: 2,
        ShadowOffset: 1,
    }
    
    return &SubtitleService{
        log:       log.WithField("service", "subtitle"),
        generator: subtitle.NewASSGenerator(config),
    }
}

func (ss *SubtitleService) GenerateSubtitles(words []subtitle.WordTimestamp, sceneTiming models.TimingSegment) (string, error) {
    ss.log.WithFields(map[string]interface{}{
        "word_count":    len(words),
        "scene_start":   sceneTiming.StartTime,
        "scene_end":     sceneTiming.EndTime,
    }).Info("Generating progressive subtitles")
    
    events := subtitle.CreateProgressiveEventsWithSceneTiming(words, sceneTiming)
    
    if len(events) == 0 {
        ss.log.Warn("No subtitle events generated")
        return "", nil
    }
    
    assContent := ss.generator.GenerateASS(events)
    
    ss.log.WithField("events_count", len(events)).Info("Subtitles generated successfully")
    
    return assContent, nil
}
```

### Configuration-Based Subtitle Generation

```go
func CreateConfigurableSubtitles(config *models.SubtitleConfig, words []subtitle.WordTimestamp, timing models.TimingSegment) (string, error) {
    // Create ASS config from domain config
    assConfig := subtitle.ASSConfig{
        FontFamily:   config.FontFamily,
        FontSize:     config.FontSize,
        Position:     fmt.Sprintf("%s-%s", config.Position.Horizontal, config.Position.Vertical),
        WordColor:    config.FontColor,
        OutlineColor: config.OutlineColor,
        OutlineWidth: config.OutlineWidth,
        ShadowOffset: 1,
    }
    
    generator := subtitle.NewASSGenerator(assConfig)
    
    var events []subtitle.SubtitleEvent
    
    if config.Progressive {
        events = subtitle.CreateProgressiveEventsWithSceneTiming(words, timing)
    } else {
        // Combine all words into single text
        var allText strings.Builder
        for i, word := range words {
            if i > 0 {
                allText.WriteString(" ")
            }
            allText.WriteString(word.Word)
        }
        
        duration := time.Duration((timing.EndTime - timing.StartTime) * float64(time.Second))
        startTime := time.Duration(timing.StartTime * float64(time.Second))
        
        events = subtitle.CreateClassicEvents(allText.String(), startTime, duration)
    }
    
    return generator.GenerateASS(events), nil
}
```

## Testing Support

### Logger Testing

```go
// MockLogger for testing
type MockLogger struct {
    entries []LogEntry
}

type LogEntry struct {
    Level   string
    Message string
    Fields  map[string]interface{}
}

func (ml *MockLogger) Info(args ...interface{}) {
    ml.entries = append(ml.entries, LogEntry{
        Level:   "info",
        Message: fmt.Sprint(args...),
    })
}

func (ml *MockLogger) WithField(key string, value interface{}) logger.Logger {
    // Return new mock with field context
    return &MockLogger{entries: ml.entries}
}

// Test example
func TestVideoService_GenerateVideo(t *testing.T) {
    mockLog := &MockLogger{}
    service := NewVideoService(mockLog)
    
    err := service.GenerateVideo(context.Background(), config)
    
    assert.NoError(t, err)
    assert.Len(t, mockLog.entries, 2) // Start and completion logs
    assert.Equal(t, "info", mockLog.entries[0].Level)
}
```

### Subtitle Testing

```go
func TestASSGenerator_GenerateASS(t *testing.T) {
    config := subtitle.ASSConfig{
        FontFamily:   "Arial",
        FontSize:     24,
        Position:     "center-bottom",
        WordColor:    "#FFFFFF",
        OutlineColor: "#000000",
        OutlineWidth: 2,
        ShadowOffset: 1,
    }
    
    generator := subtitle.NewASSGenerator(config)
    
    events := []subtitle.SubtitleEvent{
        {
            StartTime: 0,
            EndTime:   time.Second,
            Text:      "Hello",
            Layer:     0,
        },
    }
    
    result := generator.GenerateASS(events)
    
    assert.Contains(t, result, "[Script Info]")
    assert.Contains(t, result, "[V4+ Styles]")
    assert.Contains(t, result, "[Events]")
    assert.Contains(t, result, "Dialogue: 0,0:00:00.00,0:00:01.00,Default,,0,0,0,,Hello")
}

func TestCreateProgressiveEvents(t *testing.T) {
    words := []subtitle.WordTimestamp{
        {Word: "Hello", Start: 0.0, End: 0.5},
        {Word: "world", Start: 0.7, End: 1.2},
    }
    
    sceneStart := time.Duration(10 * time.Second)
    events := subtitle.CreateProgressiveEvents(words, sceneStart)
    
    assert.Len(t, events, 2)
    
    // First word
    assert.Equal(t, "Hello", events[0].Text)
    assert.Equal(t, sceneStart, events[0].StartTime)
    assert.Equal(t, sceneStart+time.Duration(700*time.Millisecond), events[0].EndTime)
    
    // Second word
    assert.Equal(t, "world", events[1].Text)
    assert.Equal(t, sceneStart+time.Duration(700*time.Millisecond), events[1].StartTime)
}
```

## Future Enhancements

### Metrics Package (Planned)

The `pkg/metrics` directory is reserved for future metrics collection utilities:

```go
// Planned metrics package structure
package metrics

type Collector interface {
    Counter(name string) Counter
    Histogram(name string) Histogram
    Gauge(name string) Gauge
}

type Counter interface {
    Inc()
    Add(delta float64)
}

type Histogram interface {
    Observe(value float64)
}

type Gauge interface {
    Set(value float64)
    Inc()
    Dec()
}
```

### Additional Utilities (Planned)

1. **pkg/cache**: Caching utilities for transcription results
2. **pkg/validation**: Common validation utilities
3. **pkg/http**: HTTP client utilities with retry logic
4. **pkg/storage**: File storage abstractions
5. **pkg/crypto**: Encryption/decryption utilities

## Best Practices

### Package Design

1. **Public APIs**: Only expose stable, well-designed APIs
2. **Backward Compatibility**: Maintain backward compatibility
3. **Interface Segregation**: Use small, focused interfaces
4. **Documentation**: Comprehensive documentation for public APIs
5. **Testing**: High test coverage for all public functions

### Usage Guidelines

1. **Import Management**: Use specific imports, not wildcard
2. **Error Handling**: Return descriptive errors
3. **Context Propagation**: Support context.Context where appropriate
4. **Resource Management**: Proper cleanup of resources
5. **Performance**: Optimize for common use cases

### Development Workflow

1. **API Design**: Design interfaces before implementations
2. **Testing**: Write tests before implementation (TDD)
3. **Documentation**: Update documentation with API changes
4. **Versioning**: Use semantic versioning for breaking changes
5. **Code Review**: Review all changes to public APIs

## Contributing

### Adding New Packages

1. **Proposal**: Discuss new package proposals
2. **Design**: Create interface and API design
3. **Implementation**: Implement with comprehensive tests
4. **Documentation**: Add complete documentation
5. **Review**: Code review and approval process

### Modifying Existing Packages

1. **Backward Compatibility**: Ensure no breaking changes
2. **Testing**: Update tests for all changes
3. **Documentation**: Update documentation
4. **Deprecation**: Use deprecation for phased removals
5. **Migration**: Provide migration guides for major changes