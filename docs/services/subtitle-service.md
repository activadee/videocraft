# Subtitle Service

The Subtitle Service handles the generation of progressive subtitles with precise timing synchronization.

## Overview

VideoCraft's subtitle system implements progressive timing to eliminate traditional subtitle gaps, providing seamless word-by-word subtitle display synchronized with audio.

## Core Features

### Progressive Subtitles
- **Zero-Gap Timing**: Eliminates empty spaces between subtitles
- **Word-Level Precision**: Individual word timing based on Whisper transcription
- **Real Duration Mapping**: Synchronizes with actual audio duration analysis
- **Configurable Styling**: ASS format with customizable appearance

### JSON SubtitleSettings (v1.1+)
Per-request subtitle customization with:
- Font styling and colors
- Position and alignment
- Animation effects
- Timing adjustments

## Service Interface

```go
type SubtitleService interface {
    GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error)
    ValidateSubtitleConfig(project models.VideoProject) error
    ValidateJSONSubtitleSettings(project models.VideoProject) error
    CleanupTempFiles(filePath string) error
}
```

## Method Details

### ValidateJSONSubtitleSettings
Validates JSON subtitle settings for a video project:
- Validates font size ranges (6-300 pixels)
- Validates color formats (#RRGGBB)
- Validates position values (9 supported positions)
- Validates outline width (0-20 pixels)
- Validates shadow offset (0-20 pixels)
- Validates style values (progressive/classic)

This method is automatically called during job creation to ensure all subtitle settings are valid before video processing begins.

### ValidateSubtitleConfig
Validates general subtitle configuration and project compatibility.

### CleanupTempFiles
Removes temporary subtitle files after processing to prevent disk space issues.

## Implementation Details

### Progressive Timing Algorithm
1. **Whisper Integration**: Receives word-level timestamps from transcription service
2. **Real Duration Analysis**: Uses audio service for actual duration calculation
3. **Gap Elimination**: Calculates continuous timing without gaps
4. **Synchronization**: Maps subtitle timing to video scenes

### ASS Generation
- Advanced SubStation Alpha format
- Styling with fonts, colors, and effects
- Precise timing control
- Multi-line subtitle support

### JSON Settings Integration
Supports dynamic subtitle configuration:
```json
{
  "subtitleSettings": {
    "font_family": "Arial",
    "font_size": 24,
    "primary_color": "#FFFFFF",
    "position": "bottom"
  }
}
```

## Security Features

- Input validation for subtitle content
- Path traversal prevention for file operations
- Resource limits for subtitle generation
- Sanitization of subtitle text content

## Performance Optimizations

- Concurrent subtitle processing
- Memory-efficient ASS generation
- Cached font metrics
- Streaming subtitle output

## Error Handling

- Graceful degradation when transcription fails
- Fallback to basic timing when progressive fails
- Comprehensive error logging
- Recovery from partial failures

## Testing

The service includes comprehensive tests:
- Unit tests for timing calculations
- Integration tests with transcription service
- Security tests for input validation
- Performance tests for large subtitle sets

## Related Documentation

- [Progressive Subtitles Architecture](../architecture/progressive-subtitles.md)
- [Transcription Service](transcription-service.md)
- [Audio Service](../services/audio-service.md)
- [JSON SubtitleSettings Examples](../subtitle-settings-json-examples.md)