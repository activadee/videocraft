# Core Components

This document details the core components of VideoCraft's architecture.

## 1. HTTP API Layer (`internal/api/`)

### Handlers
Process HTTP requests and coordinate service calls:

- **`video.go`**: Video generation endpoints
- **`job.go`**: Job management endpoints  
- **`health.go`**: Health check and metrics

### Middleware
Cross-cutting concerns that apply to all requests:

- **`auth.go`**: Bearer token authentication
- **`cors.go`**: Secure CORS configuration with domain allowlisting
- **`csrf.go`**: CSRF protection with token validation
- **`logger.go`**: Request/response logging with correlation IDs
- **`error.go`**: Centralized error handling and formatting
- **`secure_error.go`**: Secure error handling with information disclosure prevention
- **`ratelimit.go`**: Rate limiting protection

### Router
Route configuration and middleware setup.

## 2. Service Layer (`internal/services/`)

### Job Service
Orchestrates the complete video generation workflow:

```go
type JobService interface {
    CreateJob(config *models.VideoConfigArray) (*models.Job, error)
    ProcessJob(ctx context.Context, job *models.Job) error
    GetJob(id string) (*models.Job, error)
    UpdateJobProgress(id string, progress int) error
}
```

**Key Responsibilities**:
- Video generation workflow orchestration
- Job status tracking and progress updates
- Error handling and recovery
- Resource management

### Audio Service
Audio file analysis and timing calculation:

```go
type AudioService interface {
    AnalyzeAudio(ctx context.Context, url string) (*AudioInfo, error)
    CalculateSceneTiming(elements []models.Element) ([]models.TimingSegment, error)
}
```

**Key Responsibilities**:
- Real audio duration analysis using FFprobe
- Scene timing calculation for video synchronization
- Concurrent audio processing
- Audio format validation

### Transcription Service
Python Whisper daemon communication:

```go
type TranscriptionService interface {
    TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error)
    Shutdown()
}
```

**Key Responsibilities**:
- Python Whisper daemon lifecycle management
- Word-level transcription with timestamps
- Go-Python communication via stdin/stdout
- Error recovery and daemon restart

### Subtitle Service
ASS subtitle generation with progressive timing:

```go
type SubtitleService interface {
    GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error)
    ValidateSubtitleConfig(project models.VideoProject) error
}
```

**Key Responsibilities**:
- Progressive subtitle timing calculation
- ASS file generation with styling
- JSON SubtitleSettings support (v1.1+)
- Word-by-word timing mapping

### FFmpeg Service
Video encoding and command generation:

```go
type FFmpegService interface {
    GenerateVideo(ctx context.Context, config *models.VideoConfigArray, progressChan chan<- int) (string, error)
    BuildCommand(config *models.VideoConfigArray) (*FFmpegCommand, error)
}
```

**Key Responsibilities**:
- Complex FFmpeg command generation
- Filter complex building for overlays
- Progress monitoring and reporting
- Security validation and command injection prevention

### Storage Service
File management and cleanup:

**Key Responsibilities**:
- Temporary file management
- Output file storage
- Automatic cleanup
- Storage quota management

## 3. Domain Layer (`internal/domain/`)

### Models
Core business entities:

- **`VideoProject`**: Complete video configuration
- **`Scene`**: Individual scene with elements
- **`Element`**: Audio, image, video, or subtitle element
- **`Job`**: Async processing job with status tracking
- **`TranscriptionResult`**: Whisper output with word-level timing

### Validation
Business rule enforcement and input validation:

- JSON schema validation
- Business rule enforcement
- Input sanitization
- Security validation

## Component Interactions

### Request Flow
1. **HTTP Layer** receives and validates requests
2. **Job Service** creates and orchestrates processing
3. **Audio Service** analyzes audio files in parallel
4. **Transcription Service** generates word-level transcripts
5. **Subtitle Service** creates progressive subtitles
6. **FFmpeg Service** generates final video
7. **Storage Service** manages file lifecycle

### Error Propagation
- Errors bubble up through layers
- Security-sensitive errors are sanitized
- Comprehensive logging at each layer
- Graceful degradation where possible

### Concurrency Model
- Goroutines for parallel audio analysis
- Worker pools for concurrent processing
- Resource limiting with semaphores
- Context-based cancellation

## Security Integration

Each component implements security best practices:

- **Input Validation**: At every layer boundary
- **Error Sanitization**: Security-aware error handling
- **Resource Limits**: Prevent resource exhaustion
- **Audit Logging**: Comprehensive security logging

## Related Documentation

- [Architecture Overview](overview.md) - High-level system design
- [Service Layer Details](../services/overview.md) - Service implementation
- [Security Implementation](../security/overview.md) - Security architecture
- [API Reference](../api/overview.md) - HTTP API documentation