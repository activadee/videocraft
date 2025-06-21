# Subtitles Overview

VideoCraft provides a comprehensive subtitle system with progressive timing, JSON-based customization, and automatic validation.

## Key Features

### Progressive Subtitles
- **Zero-Gap Timing**: Eliminates empty spaces between subtitles
- **Word-Level Precision**: Individual word timing based on Whisper transcription
- **Real Duration Mapping**: Synchronizes with actual audio duration

### JSON Subtitle Settings (v1.1+)
- **Per-Request Customization**: Override global settings for individual videos
- **Complete Configuration**: All subtitle styling options available
- **Automatic Validation**: Settings validated during job creation

### Advanced Configuration
- **11 Configurable Fields**: Font, colors, position, outline, shadow, box styling
- **9 Position Options**: Full screen positioning control
- **Style Modes**: Progressive and classic subtitle styles
- **Fallback System**: Global configuration as intelligent defaults

## Architecture

### Service Integration
The subtitle system is integrated into the video generation workflow:

1. **Validation Phase**: `SubtitleService.ValidateJSONSubtitleSettings()` validates all subtitle configurations during job creation
2. **Generation Phase**: `SubtitleService.GenerateSubtitles()` creates progressive subtitles with precise timing
3. **Cleanup Phase**: `SubtitleService.CleanupTempFiles()` removes temporary files after processing

### Progressive Timing Algorithm
1. **Whisper Integration**: Receives word-level timestamps from transcription
2. **Real Duration Analysis**: Uses audio service for actual duration calculation
3. **Gap Elimination**: Calculates continuous timing without gaps
4. **Synchronization**: Maps subtitle timing to video scenes

## Usage

### Basic Configuration
```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Arial",
        "font-size": 24,
        "style": "progressive"
      }
    }
  ]
}
```

### Validation Workflow
All subtitle settings are automatically validated when creating video jobs:
- Invalid configurations prevent job creation
- Detailed error messages guide correction
- Validation covers all fields and ranges

## Related Documentation

- **[JSON Settings](json-settings.md)** - Complete configuration reference and examples
- **[Progressive Subtitles](progressive-subtitles.md)** - Technical implementation details
- **[ASS Generation](ass-generation.md)** - Subtitle file format specifics
- **[Subtitle Service](../services/subtitle-service.md)** - Service interface and methods

## Security & Performance

### Validation & Security
- Input validation prevents malformed configurations
- Resource limits prevent abuse
- Path traversal protection for file operations

### Performance Optimizations
- Concurrent subtitle processing
- Memory-efficient ASS generation
- Cached font metrics
- Streaming subtitle output

---

**Next Steps**: [Configure JSON Settings](json-settings.md) | [Learn Progressive Subtitles](progressive-subtitles.md) | [Explore Service API](../services/subtitle-service.md)