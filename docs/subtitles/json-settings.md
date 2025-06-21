# ⚙️ JSON SubtitleSettings Configuration

VideoCraft v1.1+ supports per-request subtitle customization through JSON configuration, allowing you to override global subtitle settings for individual video generation requests.

## 🎯 Overview

### What's New in v1.1+
- **Per-Request Customization**: Override global config for specific videos
- **Complete Field Support**: All 11 SubtitleSettings fields available
- **Intelligent Fallback**: Missing fields use global configuration
- **Backward Compatibility**: Existing APIs unchanged

### Benefits
- **Design Flexibility**: Different subtitle styles per video
- **Brand Consistency**: Customize per client or project
- **A/B Testing**: Test different subtitle configurations
- **Multi-Language Support**: Different styles per language

## 📋 SubtitleSettings Fields Reference

| Field | Type | Description | Example | Default Fallback |
|-------|------|-------------|---------|------------------|
| `style` | string | Subtitle style: "progressive" or "classic" | `"progressive"` | Global config |
| `font-family` | string | Font family name | `"Arial"` | Global config |
| `font-size` | integer | Font size in points (6-300) | `24` | Global config |
| `word-color` | string | Text color in hex format | `"#FFFFFF"` | Global config |
| `line-color` | string | Line color in hex format | `"#FF4444"` | Global config |
| `shadow-color` | string | Shadow color in hex format | `"#808080"` | `"#808080"` |
| `shadow-offset` | integer | Shadow offset in pixels (0-20) | `2` | `1` |
| `box-color` | string | Background box color in hex | `"#000080"` | `"#000000"` |
| `position` | string | Subtitle position (see below) | `"center-bottom"` | Global config |
| `outline-color` | string | Outline color in hex format | `"#000000"` | Global config |
| `outline-width` | integer | Outline width in pixels (0-20) | `2` | `2` |

## 📍 Position Values

| Position | Description | Use Case |
|----------|-------------|----------|
| `left-bottom` | Bottom-left corner | Credits, notes |
| `center-bottom` | Bottom center (default) | Standard subtitles |
| `right-bottom` | Bottom-right corner | Language indicators |
| `left-center` | Middle-left | Side content |
| `center-center` | Center of screen | Dramatic emphasis |
| `right-center` | Middle-right | UI elements |
| `left-top` | Top-left corner | Watermarks |
| `center-top` | Top center | Titles, headers |
| `right-top` | Top-right corner | Status indicators |

## 🎨 Configuration Examples

### Basic Override
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
        "font-size": 28
      }
    }
  ]
}
```

### Complete Configuration
```json
{
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

## 🎭 Style Templates

### Corporate Presentation
```json
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
```

### Gaming/Entertainment
```json
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
```

### Educational Content
```json
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
```

### Accessibility High-Contrast
```json
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
```

### Karaoke Style
```json
{
  "type": "subtitles",
  "settings": {
    "font-family": "Arial",
    "font-size": 42,
    "word-color": "#FFFF00",
    "outline-color": "#FF0000",
    "outline-width": 3,
    "shadow-color": "#000000",
    "shadow-offset": 3,
    "position": "center-center",
    "style": "progressive"
  }
}
```

## 🔄 Fallback System

### How Fallback Works
When JSON fields are not provided, the system uses global configuration:

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
        "font-family": "Times New Roman",
        "font-size": 30
      }
    }
  ]
}
```

**Result**: Uses custom font and size, but inherits global values for:
- `word-color` → Global config value
- `outline-color` → Global config value  
- `position` → Global config value
- All other fields → Global config values

## ✅ Validation Rules

VideoCraft automatically validates all JSON subtitle settings during job creation. Invalid settings will cause the job creation to fail with descriptive error messages.

### Automatic Validation
- **When**: During `POST /api/v1/videos` request processing
- **Scope**: All projects in the video configuration array
- **Result**: Job creation fails immediately if any project has invalid settings
- **Benefit**: Prevents video processing with invalid configurations

### Font Size
- **Range**: 6-300 pixels
- **Invalid Examples**: `-5`, `0`, `500`
- **Error**: `"font-size must be between 6 and 300"`

### Outline Width
- **Range**: 0-20 pixels
- **Invalid Examples**: `-1`, `25`
- **Error**: `"outline-width must be between 0 and 20"`

### Shadow Offset
- **Range**: 0-20 pixels
- **Invalid Examples**: `-2`, `30`
- **Error**: `"shadow-offset must be between 0 and 20"`

### Color Format
- **Valid**: `#RRGGBB` hex format
- **Valid Examples**: `#FF0000`, `#00ff00`, `#0000FF`
- **Invalid Examples**: `#GGG`, `#12345`, `red`, `rgb(255,0,0)`
- **Error**: `"invalid color format, use #RRGGBB"`

### Position
- **Valid**: Must be one of the 9 position values listed above
- **Invalid Examples**: `custom-position`, `top-left`, `middle`
- **Error**: `"invalid position value"`

### Style
- **Valid**: `"progressive"` or `"classic"`
- **Invalid Examples**: `"animated"`, `"custom"`
- **Error**: `"style must be 'progressive' or 'classic'"`

## 🌐 API Integration

### cURL Example
```bash
curl -X POST http://localhost:3002/api/v1/videos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
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

### JavaScript Example
```javascript
const videoConfig = {
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

const response = await fetch('/api/v1/videos', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${apiKey}`,
    'X-CSRF-Token': csrfToken
  },
  body: JSON.stringify(videoConfig)
});
```

### TypeScript Interface
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
  position?: 'left-bottom' | 'center-bottom' | 'right-bottom' |
            'left-center' | 'center-center' | 'right-center' |
            'left-top' | 'center-top' | 'right-top';
  'outline-color'?: string;
  'outline-width'?: number;
}

interface SubtitleElement {
  type: 'subtitles';
  language?: string;
  settings?: SubtitleSettings;
}

interface VideoConfig {
  scenes: Array<{
    elements: Array<{
      type: string;
      src?: string;
    }>;
  }>;
  elements: SubtitleElement[];
}
```

## 🔄 Migration Guide

### From Global Config Only (v1.0.x)
**Before** (global config only):
```yaml
# config.yaml
subtitles:
  font_family: "Arial"
  font_size: 24
  style: "progressive"
```

**After** (per-request control):
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

### Gradual Migration Strategy
1. **Phase 1**: Keep global config, add JSON overrides for specific videos
2. **Phase 2**: Gradually move common settings to JSON templates
3. **Phase 3**: Use global config only for system-wide defaults

## 🔍 Troubleshooting

### Common Issues

#### SubtitleSettings Ignored
**Symptom**: Settings in JSON don't affect output
**Cause**: Missing `type: "subtitles"` element
**Solution**: Ensure subtitle element is present in `elements` array

```json
{
  "elements": [
    {
      "type": "subtitles",  // Required!
      "settings": { /* your settings */ }
    }
  ]
}
```

#### Validation Errors
**Symptom**: API returns validation error
**Cause**: Invalid field values
**Solution**: Check validation rules above

```bash
# Check job status for detailed error info
curl http://localhost:3002/api/v1/jobs/{job-id}/status \
  -H "Authorization: Bearer $API_KEY"
```

#### Global Config Still Used
**Symptom**: Custom settings not applied
**Cause**: Empty settings object or missing fields
**Solution**: Specify the fields you want to override

```json
{
  "type": "subtitles",
  "settings": {} // Empty - will use all global values
}

// Should be:
{
  "type": "subtitles", 
  "settings": {
    "font-size": 30 // Override just this field
  }
}
```

#### Color Format Errors
**Symptom**: "Invalid color format" error
**Cause**: Wrong color format
**Solution**: Use proper `#RRGGBB` format

```json
// Wrong
"word-color": "red"
"word-color": "rgb(255,0,0)"
"word-color": "#FFF"

// Correct
"word-color": "#FF0000"
"word-color": "#ff0000"  // lowercase ok
```

## 🚀 Best Practices

### Configuration Management
1. **Create Templates**: Define reusable subtitle styles
2. **Use Fallbacks**: Don't override everything, use intelligent defaults
3. **Validate Client-Side**: Check color formats and ranges before sending
4. **Test Combinations**: Verify readability and contrast
5. **Document Styles**: Maintain style guide for your application

### Performance Considerations
1. **Progressive vs Classic**: Use classic for longer content
2. **Font Size**: Larger fonts increase processing time
3. **Complex Styles**: More styling options = more processing
4. **Caching**: Cache common configurations

### Accessibility Guidelines
1. **High Contrast**: Ensure sufficient color contrast
2. **Readable Fonts**: Use clear, legible font families
3. **Appropriate Size**: 24-32px for most viewing distances
4. **Position Consistency**: Keep position predictable for users

## 📚 Related Topics

### Subtitle System
- **[Progressive Subtitles](progressive-subtitles.md)** - Word-level timing system
- **[Subtitles Overview](overview.md)** - Subtitle system introduction
- **[ASS Generation](ass-generation.md)** - Subtitle file format details

### Configuration
- **[Video Configuration](../video-generation/configuration.md)** - Complete config format
- **[Environment Variables](../configuration/environment-variables.md)** - Global configuration

### API Usage
- **[API Overview](../api/overview.md)** - API introduction
- **[API Endpoints](../api/endpoints.md)** - Complete endpoint reference

### Examples
- **[Video Generation Examples](../video-generation/configuration.md)** - Complete video examples

---

**🔗 Next Steps**: [Explore Progressive Subtitles](progressive-subtitles.md) | [Configure Video Generation](../video-generation/configuration.md) | [API Reference](../api/endpoints.md)