# JSON SubtitleSettings Configuration Examples

This document provides comprehensive examples of how to use the new JSON SubtitleSettings feature to control subtitle appearance on a per-request basis.

## Overview

As of v1.1, VideoCraft supports per-request subtitle customization through the `SubtitleSettings` field in video configuration JSON. This allows you to override global subtitle configuration for individual video generation requests.

## Basic Usage

### Minimal JSON Configuration

```json
{
  "scenes": [
    {
      "elements": [
        {
          "type": "audio",
          "src": "https://example.com/audio.mp3"
        }
      ]
    }
  ],
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Arial",
        "font-size": 24
      }
    }
  ]
}
```

### Complete JSON Configuration

```json
{
  "scenes": [
    {
      "elements": [
        {
          "type": "audio",
          "src": "https://example.com/audio.mp3"
        }
      ]
    }
  ],
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "style": "progressive",
        "font-family": "Comic Sans MS",
        "font-size": 48,
        "word-color": "#FF0000",
        "line-color": "#FF4444",
        "shadow-color": "#800000",
        "shadow-offset": 3,
        "box-color": "#000080",
        "position": "center-top",
        "outline-color": "#00FF00",
        "outline-width": 5
      }
    }
  ]
}
```

## SubtitleSettings Fields Reference

| Field | Type | Description | Default Fallback |
|-------|------|-------------|------------------|
| `style` | string | Subtitle style: "progressive" or "classic" | Global config |
| `font-family` | string | Font family name (e.g., "Arial", "Times New Roman") | Global config |
| `font-size` | integer | Font size in points (6-300) | Global config |
| `word-color` | string | Text color in hex format (#RRGGBB) | Global config |
| `line-color` | string | Line color in hex format (#RRGGBB) | Global config |
| `shadow-color` | string | Shadow color in hex format (#RRGGBB) | "#808080" |
| `shadow-offset` | integer | Shadow offset in pixels (0-20) | 1 |
| `box-color` | string | Background box color in hex format (#RRGGBB) | "#000000" |
| `position` | string | Subtitle position (see Position Values below) | Global config |
| `outline-color` | string | Outline color in hex format (#RRGGBB) | Global config |
| `outline-width` | integer | Outline width in pixels (0-20) | 2 |

### Position Values

| Position | Description |
|----------|-------------|
| `left-bottom` | Bottom-left corner |
| `center-bottom` | Bottom center (default) |
| `right-bottom` | Bottom-right corner |
| `left-center` | Middle-left |
| `center-center` | Center of screen |
| `right-center` | Middle-right |
| `left-top` | Top-left corner |
| `center-top` | Top center |
| `right-top` | Top-right corner |

## Configuration Examples by Use Case

### 1. Corporate Presentation Style

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Calibri",
        "font-size": 28,
        "word-color": "#2E2E2E",
        "outline-color": "#FFFFFF",
        "outline-width": 2,
        "position": "center-bottom",
        "style": "classic"
      }
    }
  ]
}
```

### 2. Gaming/Entertainment Style

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Impact",
        "font-size": 36,
        "word-color": "#FFFF00",
        "outline-color": "#000000",
        "outline-width": 4,
        "shadow-color": "#800080",
        "shadow-offset": 2,
        "position": "center-top",
        "style": "progressive"
      }
    }
  ]
}
```

### 3. Educational Content Style

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Georgia",
        "font-size": 22,
        "word-color": "#1F1F1F",
        "outline-color": "#F0F0F0",
        "outline-width": 1,
        "box-color": "#FFFFFF88",
        "position": "center-bottom",
        "style": "classic"
      }
    }
  ]
}
```

### 4. Dramatic/Cinematic Style

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Times New Roman",
        "font-size": 32,
        "word-color": "#F5F5DC",
        "outline-color": "#1C1C1C",
        "outline-width": 3,
        "shadow-color": "#000000",
        "shadow-offset": 4,
        "position": "center-bottom",
        "style": "progressive"
      }
    }
  ]
}
```

### 5. Accessibility High-Contrast Style

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Arial",
        "font-size": 32,
        "word-color": "#FFFFFF",
        "outline-color": "#000000",
        "outline-width": 6,
        "box-color": "#000000CC",
        "position": "center-bottom",
        "style": "classic"
      }
    }
  ]
}
```

## Progressive vs Classic Styles

### Progressive Style
- **Character-by-character reveal**: Words appear character by character synchronized with speech
- **Best for**: Karaoke, educational content, language learning
- **Performance**: Higher processing requirements
- **JSON Configuration**: `"style": "progressive"`

### Classic Style
- **Word or sentence-level**: Traditional subtitle display
- **Best for**: Movies, general content, faster processing
- **Performance**: Lower processing requirements
- **JSON Configuration**: `"style": "classic"`

## Color Format Guidelines

### Hex Colors
All color fields accept hex color format:
- **Format**: `#RRGGBB` (6 hexadecimal digits)
- **Examples**: `#FF0000` (red), `#00FF00` (green), `#0000FF` (blue)
- **Case**: Both uppercase and lowercase are supported

### Common Color Values
```json
{
  "word-color": "#FFFFFF",    // White
  "outline-color": "#000000", // Black
  "shadow-color": "#808080",  // Gray
  "box-color": "#000080"      // Navy blue
}
```

## Global Configuration Fallback

When JSON SubtitleSettings fields are not provided, the system falls back to global configuration values:

```yaml
# config.yaml (global configuration)
subtitles:
  enabled: true
  style: "progressive"
  font_family: "Arial"
  font_size: 24
  position: "center-bottom"
  colors:
    word: "#FFFFFF"
    outline: "#000000"
```

### Partial Override Example

```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Custom Font",
        "font-size": 30
      }
    }
  ]
}
```

**Note**: When only some fields are provided, missing fields (word-color, outline-color, position) will use global config values.

## Validation Rules

The system validates JSON SubtitleSettings and will reject invalid configurations:

### Font Size
- **Range**: 6-300 pixels
- **Invalid**: Values outside this range will cause validation errors

### Outline Width
- **Range**: 0-20 pixels
- **Invalid**: Negative values or values > 20

### Shadow Offset
- **Range**: 0-20 pixels
- **Invalid**: Negative values or values > 20

### Colors
- **Format**: Must be valid hex format if starting with `#`
- **Invalid**: `#GGGGGG`, `#12345`, malformed hex codes

### Position
- **Valid**: Must be one of the position values listed above
- **Invalid**: Custom position strings not in the allowed list

### Style
- **Valid**: "progressive" or "classic"
- **Invalid**: Any other string values

## API Integration Examples

### cURL Example

```bash
curl -X POST http://localhost:3002/api/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <API_TOKEN>" \
  -d '{
    "scenes": [
      {
        "elements": [
          {
            "type": "audio",
            "src": "https://example.com/audio.mp3"
          }
        ]
      }
    ],
    "elements": [
      {
        "type": "subtitles",
        "settings": {
          "font-family": "Arial",
          "font-size": 28,
          "word-color": "#FFFFFF",
          "outline-color": "#000000",
          "position": "center-bottom",
          "style": "progressive"
        }
      }
    ]
  }'
```

### JavaScript/TypeScript Example

```typescript
interface SubtitleSettings {
  style?: 'progressive' | 'classic';
  'font-family'?: string;
  'font-size'?: number;
  'word-color'?: string;
  'line-color'?: string;
  'shadow-color'?: string;
  'shadow-offset'?: number;
  'box-color'?: string;
  position?: string;
  'outline-color'?: string;
  'outline-width'?: number;
}

interface VideoRequest {
  scenes: Array<{
    elements: Array<{
      type: string;
      src?: string;
      settings?: SubtitleSettings;
    }>;
  }>;
  elements: Array<{
    type: string;
    settings?: SubtitleSettings;
  }>;
}

const videoRequest: VideoRequest = {
  scenes: [
    {
      elements: [
        {
          type: "audio",
          src: "https://example.com/audio.mp3"
        }
      ]
    }
  ],
  elements: [
    {
      type: "subtitles",
      settings: {
        "font-family": "Arial",
        "font-size": 24,
        "word-color": "#FFFFFF",
        "outline-color": "#000000",
        style: "progressive"
      }
    }
  ]
};

fetch('/api/v1/videos', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer <API_TOKEN>'
  },
  body: JSON.stringify(videoRequest)
});
```

## Migration Guide

### From Global Config Only (v2.0 and earlier)

**Before** (global config only):
```yaml
# config.yaml
subtitles:
  font_family: "Arial"
  font_size: 24
```

**After** (per-request control):
```json
{
  "elements": [
    {
      "type": "subtitles",
      "settings": {
        "font-family": "Arial",
        "font-size": 24
      }
    }
  ]
}
```

### Gradual Migration Strategy

1. **Phase 1**: Keep global config, add JSON overrides for specific requests
2. **Phase 2**: Gradually move common settings to JSON
3. **Phase 3**: Use global config only for system-wide defaults

## Performance Considerations

### Progressive Style Performance
- **Higher CPU usage**: Character-by-character rendering
- **Recommended**: Use for shorter content or when effect is essential

### Classic Style Performance
- **Lower CPU usage**: Traditional word-level rendering
- **Recommended**: Use for longer content or when performance is critical

### Optimal Settings
```json
{
  "settings": {
    "style": "classic",        // Better performance
    "font-size": 24,          // Reasonable size
    "outline-width": 2,       // Minimal processing
    "shadow-offset": 1        // Light shadow effect
  }
}
```

## Troubleshooting

### Common Issues

1. **SubtitleSettings Ignored**
   - **Cause**: Missing `type: "subtitles"` element
   - **Solution**: Ensure subtitle element is present in `elements` array

2. **Validation Errors**
   - **Cause**: Invalid field values (font size, colors, position)
   - **Solution**: Check validation rules above

3. **Global Config Still Used**
   - **Cause**: Empty SubtitleSettings object or missing fields
   - **Solution**: Specify the fields you want to override

4. **Styling Not Applied**
   - **Cause**: Invalid hex color format
   - **Solution**: Use proper #RRGGBB format for colors

### Debug API Response

Check the job status API for validation errors:
```bash
curl -H "Authorization: Bearer <API_TOKEN>" \
  http://localhost:3002/api/v1/jobs/{job-id}
```

Look for error messages in the response that indicate SubtitleSettings validation failures.

## Best Practices

1. **Use Fallback Strategy**: Provide global config for defaults, JSON for customization
2. **Validate Settings Before Sending**: Check color formats and value ranges client-side
3. **Test Combinations**: Verify color contrast and readability
4. **Performance First**: Use classic style for production unless progressive is required
5. **Consistent Styling**: Define reusable subtitle style templates
6. **Accessibility**: Ensure sufficient contrast for subtitle visibility

## Future Enhancements

Planned features for future releases:
- Animation effects for progressive subtitles
- Multiple subtitle tracks per video
- Font weight and style controls
- Advanced positioning with pixel-perfect control
- Subtitle timing offset controls