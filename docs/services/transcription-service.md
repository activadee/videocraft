# Transcription Service

The Transcription Service handles Python Whisper daemon communication for AI-powered transcription.

## Service Interface

```go
type TranscriptionService interface {
    TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error)
    Shutdown()
}
```

## Key Responsibilities

- Python Whisper daemon lifecycle management
- Word-level transcription with timestamps
- Go-Python communication via stdin/stdout
- Error recovery and daemon restart