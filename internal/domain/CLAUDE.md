# Domain Package - Core Business Models

## Overview
The `internal/domain` package contains the core business entities, value objects, and domain models that represent the fundamental concepts of VideoCraft. This package defines the ubiquitous language of the video generation domain and encapsulates business rules and validation logic.

## Architecture

```text
internal/domain/
├── models/
│   └── video.go           # Core video domain models
├── errors/
│   └── errors.go          # Domain-specific error types
└── CLAUDE.md             # This documentation
```

## Core Domain Models

### Video Configuration Models

The domain models represent the complete video generation specification:

```go
// VideoConfigArray represents the top-level video configuration
type VideoConfigArray struct {
    Video  VideoConfig `json:"video" validate:"required"`
    Scenes []Scene     `json:"scenes" validate:"required,min=1,dive"`
}

// VideoConfig defines the overall video properties
type VideoConfig struct {
    Width      int     `json:"width" validate:"required,min=1"`
    Height     int     `json:"height" validate:"required,min=1"`
    FPS        float64 `json:"fps" validate:"required,min=1,max=60"`
    Duration   float64 `json:"duration" validate:"required,min=0.1"`
    Background string  `json:"background" validate:"required"`
}

// Scene represents a single scene in the video
type Scene struct {
    Elements []Element `json:"elements" validate:"required,min=1,dive"`
}

// Element represents a single element within a scene
type Element struct {
    Type     string                 `json:"type" validate:"required,oneof=audio image video subtitle"`
    Src      string                 `json:"src" validate:"required"`
    Position *Position              `json:"position,omitempty"`
    Style    map[string]interface{} `json:"style,omitempty"`
    Timing   *Timing                `json:"timing,omitempty"`
}

// Position defines spatial positioning for visual elements
type Position struct {
    X      float64 `json:"x" validate:"min=0"`
    Y      float64 `json:"y" validate:"min=0"`
    Width  float64 `json:"width,omitempty" validate:"min=0"`
    Height float64 `json:"height,omitempty" validate:"min=0"`
    Z      int     `json:"z,omitempty"` // Layer index for overlays
}

// Timing defines temporal properties for elements
type Timing struct {
    Start    float64 `json:"start" validate:"min=0"`
    End      float64 `json:"end" validate:"gtfield=Start"`
    FadeIn   float64 `json:"fade_in,omitempty" validate:"min=0"`
    FadeOut  float64 `json:"fade_out,omitempty" validate:"min=0"`
}
```

### Job Management Models

Models for tracking asynchronous video generation jobs:

```go
// Job represents a video generation job
type Job struct {
    ID          string          `json:"id"`
    Status      JobStatus       `json:"status"`
    Progress    int             `json:"progress"`
    Config      *VideoConfigArray `json:"config"`
    OutputPath  string          `json:"output_path,omitempty"`
    Error       string          `json:"error,omitempty"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// JobStatus represents the current state of a job
type JobStatus string

const (
    JobStatusPending    JobStatus = "pending"
    JobStatusProcessing JobStatus = "processing"
    JobStatusCompleted  JobStatus = "completed"
    JobStatusFailed     JobStatus = "failed"
)

// JobMetrics provides job statistics
type JobMetrics struct {
    TotalJobs     int64         `json:"total_jobs"`
    PendingJobs   int64         `json:"pending_jobs"`
    ProcessingJobs int64        `json:"processing_jobs"`
    CompletedJobs int64         `json:"completed_jobs"`
    FailedJobs    int64         `json:"failed_jobs"`
    AverageTime   time.Duration `json:"average_time"`
}
```

### Audio and Transcription Models

Models for audio processing and speech recognition:

```go
// AudioInfo contains metadata about audio files
type AudioInfo struct {
    URL         string        `json:"url"`
    Duration    float64       `json:"duration"`
    SampleRate  int           `json:"sample_rate,omitempty"`
    Channels    int           `json:"channels,omitempty"`
    Format      string        `json:"format,omitempty"`
    Size        int64         `json:"size,omitempty"`
    Bitrate     int           `json:"bitrate,omitempty"`
}

// TranscriptionResult contains the output from speech recognition
type TranscriptionResult struct {
    Text           string          `json:"text"`
    Language       string          `json:"language"`
    Duration       float64         `json:"duration"`
    Segments       []Segment       `json:"segments"`
    WordTimestamps []WordTimestamp `json:"word_timestamps"`
    Confidence     float64         `json:"confidence,omitempty"`
}

// Segment represents a text segment with timing
type Segment struct {
    ID     int     `json:"id"`
    Start  float64 `json:"start"`
    End    float64 `json:"end"`
    Text   string  `json:"text"`
    Words  []Word  `json:"words,omitempty"`
}

// Word represents a single word with precise timing
type Word struct {
    Word       string  `json:"word"`
    Start      float64 `json:"start"`
    End        float64 `json:"end"`
    Confidence float64 `json:"confidence,omitempty"`
}

// WordTimestamp provides simplified word timing for subtitles
type WordTimestamp struct {
    Word  string  `json:"word"`
    Start float64 `json:"start"`
    End   float64 `json:"end"`
}
```

### Subtitle Models

Models for progressive subtitle generation:

```go
// SubtitleConfig defines subtitle generation parameters
type SubtitleConfig struct {
    Enabled          bool                   `json:"enabled"`
    Language         string                 `json:"language"`
    FontFamily       string                 `json:"font_family"`
    FontSize         int                    `json:"font_size"`
    FontColor        string                 `json:"font_color"`
    OutlineColor     string                 `json:"outline_color"`
    OutlineWidth     int                    `json:"outline_width"`
    BackgroundColor  string                 `json:"background_color,omitempty"`
    Position         SubtitlePosition       `json:"position"`
    Progressive      bool                   `json:"progressive"`
    Style            map[string]interface{} `json:"style,omitempty"`
}

// SubtitlePosition defines subtitle placement
type SubtitlePosition struct {
    Horizontal string  `json:"horizontal" validate:"oneof=left center right"`
    Vertical   string  `json:"vertical" validate:"oneof=top middle bottom"`
    MarginX    float64 `json:"margin_x"`
    MarginY    float64 `json:"margin_y"`
}

// SubtitleTrack contains all subtitle events for a video
type SubtitleTrack struct {
    Events   []SubtitleEvent `json:"events"`
    Language string          `json:"language"`
    Format   string          `json:"format"` // ASS, SRT, WebVTT
}

// SubtitleEvent represents a single subtitle display event
type SubtitleEvent struct {
    StartTime time.Duration `json:"start_time"`
    EndTime   time.Duration `json:"end_time"`
    Text      string        `json:"text"`
    Style     *SubtitleStyle `json:"style,omitempty"`
    Type      string        `json:"type"` // "progressive", "static"
}

// SubtitleStyle defines styling for individual subtitle events
type SubtitleStyle struct {
    FontName     string `json:"font_name,omitempty"`
    FontSize     int    `json:"font_size,omitempty"`
    PrimaryColor string `json:"primary_color,omitempty"`
    OutlineColor string `json:"outline_color,omitempty"`
    Alignment    int    `json:"alignment,omitempty"`
}
```

### Timing and Synchronization Models

Models for precise timing calculations:

```go
// TimingSegment represents a time period for scene-based content
type TimingSegment struct {
    StartTime    float64 `json:"start_time"`
    EndTime      float64 `json:"end_time"`
    Duration     float64 `json:"duration"`
    AudioFile    string  `json:"audio_file,omitempty"`
    SceneIndex   int     `json:"scene_index"`
}

// VideoProject combines all aspects of a video generation project
type VideoProject struct {
    Config       *VideoConfigArray    `json:"config"`
    AudioInfo    []AudioInfo          `json:"audio_info"`
    Transcription *TranscriptionResult `json:"transcription,omitempty"`
    Timings      []TimingSegment      `json:"timings"`
    SubtitleTrack *SubtitleTrack      `json:"subtitle_track,omitempty"`
}

// ProgressiveSubtitleMap maps timing to character reveals
type ProgressiveSubtitleMap struct {
    SceneIndex    int                    `json:"scene_index"`
    AudioFile     string                 `json:"audio_file"`
    StartTime     float64                `json:"start_time"`
    EndTime       float64                `json:"end_time"`
    Words         []WordTimestamp        `json:"words"`
    CharacterMap  []CharacterReveal      `json:"character_map"`
}

// CharacterReveal defines when each character should be revealed
type CharacterReveal struct {
    Character string        `json:"character"`
    Timestamp time.Duration `json:"timestamp"`
    Position  int           `json:"position"`
}
```

## Domain Validation

### Business Rules

The domain models enforce business rules through validation:

```go
// ValidateVideoConfig ensures video configuration is valid
func (vc *VideoConfig) Validate() error {
    if vc.Width <= 0 || vc.Height <= 0 {
        return errors.New("video dimensions must be positive")
    }
    
    if vc.FPS <= 0 || vc.FPS > 60 {
        return errors.New("fps must be between 1 and 60")
    }
    
    if vc.Duration <= 0 {
        return errors.New("duration must be positive")
    }
    
    return nil
}

// ValidateScene ensures scene has valid elements
func (s *Scene) Validate() error {
    if len(s.Elements) == 0 {
        return errors.New("scene must have at least one element")
    }
    
    audioCount := 0
    for _, element := range s.Elements {
        if err := element.Validate(); err != nil {
            return fmt.Errorf("invalid element: %w", err)
        }
        
        if element.Type == "audio" {
            audioCount++
        }
    }
    
    if audioCount == 0 {
        return errors.New("scene must have at least one audio element")
    }
    
    return nil
}

// ValidateElement ensures element configuration is valid
func (e *Element) Validate() error {
    validTypes := map[string]bool{
        "audio":    true,
        "image":    true,
        "video":    true,
        "subtitle": true,
    }
    
    if !validTypes[e.Type] {
        return fmt.Errorf("invalid element type: %s", e.Type)
    }
    
    if e.Src == "" {
        return errors.New("element source is required")
    }
    
    // URL validation for external resources
    if strings.HasPrefix(e.Src, "http") {
        if _, err := url.Parse(e.Src); err != nil {
            return fmt.Errorf("invalid URL: %w", err)
        }
    }
    
    // Timing validation
    if e.Timing != nil {
        if e.Timing.Start < 0 {
            return errors.New("timing start must be non-negative")
        }
        if e.Timing.End <= e.Timing.Start {
            return errors.New("timing end must be after start")
        }
    }
    
    return nil
}
```

### Progressive Subtitle Validation

```go
// ValidateProgressiveSubtitles ensures subtitle configuration is valid
func (sc *SubtitleConfig) ValidateProgressive() error {
    if !sc.Enabled {
        return nil // No validation needed if disabled
    }
    
    if !sc.Progressive {
        return nil // Standard subtitle validation
    }
    
    // Progressive subtitle specific validation
    if sc.FontSize <= 0 {
        return errors.New("font size must be positive for progressive subtitles")
    }
    
    validPositions := map[string]bool{
        "left": true, "center": true, "right": true,
    }
    if !validPositions[sc.Position.Horizontal] {
        return errors.New("invalid horizontal position")
    }
    
    validVertical := map[string]bool{
        "top": true, "middle": true, "bottom": true,
    }
    if !validVertical[sc.Position.Vertical] {
        return errors.New("invalid vertical position")
    }
    
    return nil
}
```

## Domain Events

### Job Lifecycle Events

```go
// JobEvent represents a domain event in the job lifecycle
type JobEvent struct {
    Type      JobEventType `json:"type"`
    JobID     string       `json:"job_id"`
    Timestamp time.Time    `json:"timestamp"`
    Data      interface{}  `json:"data,omitempty"`
}

// JobEventType defines the type of job event
type JobEventType string

const (
    JobCreated    JobEventType = "job_created"
    JobStarted    JobEventType = "job_started"
    JobProgress   JobEventType = "job_progress"
    JobCompleted  JobEventType = "job_completed"
    JobFailed     JobEventType = "job_failed"
)

// ProgressEvent contains progress update data
type ProgressEvent struct {
    Progress    int    `json:"progress"`
    Stage       string `json:"stage"`
    Description string `json:"description"`
}
```

## Value Objects

### Color Value Object

```go
// Color represents a color value with validation
type Color struct {
    Value string `json:"value"`
}

// NewColor creates a validated color value
func NewColor(value string) (*Color, error) {
    if !isValidColor(value) {
        return nil, errors.New("invalid color format")
    }
    return &Color{Value: value}, nil
}

// String returns the color value
func (c Color) String() string {
    return c.Value
}

// isValidColor validates color formats (hex, rgb, rgba, named)
func isValidColor(value string) bool {
    // Hex colors
    if matched, _ := regexp.MatchString(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`, value); matched {
        return true
    }
    
    // RGB/RGBA colors
    if matched, _ := regexp.MatchString(`^rgba?\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*(,\s*[\d.]+)?\s*\)$`, value); matched {
        return true
    }
    
    // Named colors
    namedColors := map[string]bool{
        "white": true, "black": true, "red": true, "green": true, "blue": true,
        "yellow": true, "cyan": true, "magenta": true, "transparent": true,
    }
    return namedColors[strings.ToLower(value)]
}
```

### Duration Value Object

```go
// VideoDuration represents a video duration with validation
type VideoDuration struct {
    Seconds float64 `json:"seconds"`
}

// NewVideoDuration creates a validated duration
func NewVideoDuration(seconds float64) (*VideoDuration, error) {
    if seconds <= 0 {
        return nil, errors.New("duration must be positive")
    }
    if seconds > 3600 { // 1 hour max
        return nil, errors.New("duration cannot exceed 1 hour")
    }
    return &VideoDuration{Seconds: seconds}, nil
}

// String returns formatted duration
func (d VideoDuration) String() string {
    minutes := int(d.Seconds) / 60
    seconds := int(d.Seconds) % 60
    return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Milliseconds returns duration in milliseconds
func (d VideoDuration) Milliseconds() int64 {
    return int64(d.Seconds * 1000)
}
```

## Domain Services

### Timing Calculator

```go
// TimingCalculator handles complex timing calculations
type TimingCalculator struct{}

// CalculateSceneTimings computes timing for scenes based on audio duration
func (tc *TimingCalculator) CalculateSceneTimings(audioInfos []AudioInfo) ([]TimingSegment, error) {
    if len(audioInfos) == 0 {
        return nil, errors.New("no audio information provided")
    }
    
    var segments []TimingSegment
    currentTime := 0.0
    
    for i, audio := range audioInfos {
        segment := TimingSegment{
            StartTime:  currentTime,
            EndTime:    currentTime + audio.Duration,
            Duration:   audio.Duration,
            AudioFile:  audio.URL,
            SceneIndex: i,
        }
        
        segments = append(segments, segment)
        currentTime += audio.Duration
    }
    
    return segments, nil
}

// ValidateTimingConsistency ensures timing segments are consistent
func (tc *TimingCalculator) ValidateTimingConsistency(segments []TimingSegment) error {
    for i, segment := range segments {
        if segment.StartTime < 0 {
            return fmt.Errorf("segment %d has negative start time", i)
        }
        
        if segment.EndTime <= segment.StartTime {
            return fmt.Errorf("segment %d has invalid end time", i)
        }
        
        if i > 0 {
            prevSegment := segments[i-1]
            if segment.StartTime < prevSegment.EndTime {
                return fmt.Errorf("segment %d overlaps with previous segment", i)
            }
        }
    }
    
    return nil
}
```

### Progressive Subtitle Calculator

```go
// ProgressiveCalculator handles progressive subtitle timing
type ProgressiveCalculator struct{}

// CalculateCharacterReveal computes character reveal timing
func (pc *ProgressiveCalculator) CalculateCharacterReveal(
    words []WordTimestamp, 
    sceneStart float64,
) ([]CharacterReveal, error) {
    var reveals []CharacterReveal
    position := 0
    
    for _, word := range words {
        wordChars := []rune(word.Word)
        wordDuration := word.End - word.Start
        charDuration := wordDuration / float64(len(wordChars))
        
        for i, char := range wordChars {
            timestamp := time.Duration((sceneStart + word.Start + float64(i)*charDuration) * float64(time.Second))
            
            reveal := CharacterReveal{
                Character: string(char),
                Timestamp: timestamp,
                Position:  position,
            }
            
            reveals = append(reveals, reveal)
            position++
        }
        
        // Add space after word (except last word)
        if word != words[len(words)-1] {
            spaceTimestamp := time.Duration((sceneStart + word.End) * float64(time.Second))
            reveals = append(reveals, CharacterReveal{
                Character: " ",
                Timestamp: spaceTimestamp,
                Position:  position,
            })
            position++
        }
    }
    
    return reveals, nil
}
```

## Repository Interfaces

The domain defines interfaces for data persistence:

```go
// JobRepository defines job persistence operations
type JobRepository interface {
    Create(job *Job) error
    GetByID(id string) (*Job, error)
    Update(job *Job) error
    Delete(id string) error
    List(limit, offset int) ([]*Job, error)
    GetByStatus(status JobStatus) ([]*Job, error)
}

// VideoConfigRepository defines video configuration operations
type VideoConfigRepository interface {
    Store(id string, config *VideoConfigArray) error
    Retrieve(id string) (*VideoConfigArray, error)
    Delete(id string) error
}

// TranscriptionRepository defines transcription data operations
type TranscriptionRepository interface {
    Store(audioURL string, result *TranscriptionResult) error
    Retrieve(audioURL string) (*TranscriptionResult, error)
    Exists(audioURL string) bool
    Clear() error
}
```

## Error Handling

### Domain Errors

```go
// DomainError represents business logic errors
type DomainError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Field   string                 `json:"field,omitempty"`
    Value   interface{}            `json:"value,omitempty"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (e DomainError) Error() string {
    return e.Message
}

// Common domain errors
var (
    ErrInvalidVideoConfig     = DomainError{Code: "INVALID_VIDEO_CONFIG", Message: "Invalid video configuration"}
    ErrInvalidScene          = DomainError{Code: "INVALID_SCENE", Message: "Invalid scene configuration"}
    ErrInvalidElement        = DomainError{Code: "INVALID_ELEMENT", Message: "Invalid element configuration"}
    ErrInvalidTiming         = DomainError{Code: "INVALID_TIMING", Message: "Invalid timing configuration"}
    ErrJobNotFound           = DomainError{Code: "JOB_NOT_FOUND", Message: "Job not found"}
    ErrJobAlreadyProcessing  = DomainError{Code: "JOB_ALREADY_PROCESSING", Message: "Job is already being processed"}
    ErrInvalidJobStatus      = DomainError{Code: "INVALID_JOB_STATUS", Message: "Invalid job status transition"}
)

// NewValidationError creates a validation error with field context
func NewValidationError(field string, value interface{}, message string) DomainError {
    return DomainError{
        Code:    "VALIDATION_ERROR",
        Message: message,
        Field:   field,
        Value:   value,
    }
}
```

## JSON Serialization

### Custom JSON Marshaling

```go
// MarshalJSON provides custom JSON serialization for JobStatus
func (js JobStatus) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(js))
}

// UnmarshalJSON provides custom JSON deserialization for JobStatus
func (js *JobStatus) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    
    switch s {
    case "pending", "processing", "completed", "failed":
        *js = JobStatus(s)
    default:
        return fmt.Errorf("invalid job status: %s", s)
    }
    
    return nil
}

// MarshalJSON provides custom time formatting
func (j Job) MarshalJSON() ([]byte, error) {
    type Alias Job
    return json.Marshal(&struct {
        CreatedAt   string  `json:"created_at"`
        UpdatedAt   string  `json:"updated_at"`
        CompletedAt *string `json:"completed_at,omitempty"`
        *Alias
    }{
        CreatedAt:   j.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   j.UpdatedAt.Format(time.RFC3339),
        CompletedAt: func() *string {
            if j.CompletedAt != nil {
                s := j.CompletedAt.Format(time.RFC3339)
                return &s
            }
            return nil
        }(),
        Alias: (*Alias)(&j),
    })
}
```

## Testing Support

### Test Builders

```go
// JobBuilder provides a fluent interface for building test jobs
type JobBuilder struct {
    job *Job
}

// NewJobBuilder creates a new job builder
func NewJobBuilder() *JobBuilder {
    return &JobBuilder{
        job: &Job{
            ID:        "test-job-" + generateID(),
            Status:    JobStatusPending,
            Progress:  0,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
}

// WithStatus sets the job status
func (jb *JobBuilder) WithStatus(status JobStatus) *JobBuilder {
    jb.job.Status = status
    return jb
}

// WithProgress sets the job progress
func (jb *JobBuilder) WithProgress(progress int) *JobBuilder {
    jb.job.Progress = progress
    return jb
}

// WithConfig sets the video configuration
func (jb *JobBuilder) WithConfig(config *VideoConfigArray) *JobBuilder {
    jb.job.Config = config
    return jb
}

// Build returns the constructed job
func (jb *JobBuilder) Build() *Job {
    return jb.job
}

// Example usage in tests:
// job := NewJobBuilder().
//     WithStatus(JobStatusProcessing).
//     WithProgress(50).
//     Build()
```

### Mock Implementations

```go
// MockJobRepository provides a test implementation
type MockJobRepository struct {
    jobs map[string]*Job
    mu   sync.RWMutex
}

func NewMockJobRepository() *MockJobRepository {
    return &MockJobRepository{
        jobs: make(map[string]*Job),
    }
}

func (mjr *MockJobRepository) Create(job *Job) error {
    mjr.mu.Lock()
    defer mjr.mu.Unlock()
    
    if _, exists := mjr.jobs[job.ID]; exists {
        return errors.New("job already exists")
    }
    
    mjr.jobs[job.ID] = job
    return nil
}

func (mjr *MockJobRepository) GetByID(id string) (*Job, error) {
    mjr.mu.RLock()
    defer mjr.mu.RUnlock()
    
    job, exists := mjr.jobs[id]
    if !exists {
        return nil, ErrJobNotFound
    }
    
    return job, nil
}
```

## Best Practices

### Domain Design Principles

1. **Ubiquitous Language**: Use consistent terminology throughout the domain
2. **Encapsulation**: Keep business rules within domain objects
3. **Immutability**: Prefer immutable value objects where possible
4. **Validation**: Validate business rules at the domain level
5. **Self-Documenting**: Use clear naming and structure

### Performance Considerations

1. **Value Object Caching**: Cache frequently used value objects
2. **Lazy Loading**: Load complex associations only when needed
3. **Bulk Operations**: Provide efficient bulk operations for repositories
4. **Indexing**: Consider database indexing strategies for queries

### Testing Guidelines

1. **Unit Tests**: Test domain logic in isolation
2. **Property-Based Testing**: Use property-based tests for validation
3. **Test Builders**: Use builders for complex object construction
4. **Mock Repositories**: Use mocks for testing domain services
5. **Domain Events**: Test event generation and handling