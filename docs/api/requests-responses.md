# API Requests and Responses Documentation

This document provides comprehensive documentation for VideoCraft's API request and response formats, including detailed schemas, examples, and validation rules.

## =Ë API Overview

VideoCraft provides a RESTful API for video generation with the following key characteristics:

- **Content Type**: `application/json` for all requests/responses
- **Authentication**: Bearer token authentication (configurable)
- **Rate Limiting**: Configurable per-client rate limiting
- **CSRF Protection**: Token-based CSRF protection for state-changing requests
- **Error Handling**: Standardized error responses with detailed information

## <¬ Video Generation API

### POST /api/v1/generate-video

Creates a new video generation job.

#### Request Format

```json
{
  "scenes": [
    {
      "id": "intro",
      "background-color": "#000000",
      "elements": [
        {
          "type": "audio",
          "src": "https://example.com/audio/intro.mp3",
          "volume": 1.0,
          "duration": 10.5
        },
        {
          "type": "image", 
          "src": "https://example.com/images/background.jpg",
          "x": 0,
          "y": 0,
          "resize": "cover",
          "duration": 10.5
        },
        {
          "type": "subtitles",
          "language": "en",
          "settings": {
            "style": "progressive",
            "font-family": "Arial",
            "font-size": 24,
            "word-color": "#FFFFFF",
            "outline-color": "#000000",
            "outline-width": 2,
            "position": "center-bottom"
          }
        }
      ]
    }
  ]
}
```

#### Request Schema

**VideoConfigArray** (Root):
```typescript
interface VideoConfigArray {
  scenes: Scene[];
}
```

**Scene**:
```typescript
interface Scene {
  id: string;                    // Required: Unique scene identifier
  "background-color"?: string;   // Optional: Scene background color (hex)
  elements: Element[];           // Required: Array of scene elements
}
```

**Element**:
```typescript
interface Element {
  type: "audio" | "image" | "video" | "subtitles";  // Required: Element type
  src?: string;                  // Required for audio/image/video
  id?: string;                   // Optional: Element identifier
  
  // Positioning (for visual elements)
  x?: number;                    // Optional: X position in pixels
  y?: number;                    // Optional: Y position in pixels
  "z-index"?: number;           // Optional: Layer order (higher = front)
  
  // Audio properties
  volume?: number;              // Optional: Audio volume (0.0-1.0)
  
  // Visual properties  
  resize?: "cover" | "contain" | "stretch" | "center";  // Optional: Resize behavior
  duration?: number;            // Optional: Element duration in seconds
  
  // Subtitle properties
  language?: string;            // Optional: Language code (e.g., "en", "de")
  settings?: SubtitleSettings;  // Optional: Subtitle styling
}
```

**SubtitleSettings**:
```typescript
interface SubtitleSettings {
  style?: "progressive" | "classic";     // Optional: Subtitle style
  "font-family"?: string;               // Optional: Font family name
  "font-size"?: number;                 // Optional: Font size in points
  "word-color"?: string;                // Optional: Text color (hex)
  "line-color"?: string;                // Optional: Line color (hex)
  "shadow-color"?: string;              // Optional: Shadow color (hex)
  "shadow-offset"?: number;             // Optional: Shadow offset in pixels
  "box-color"?: string;                 // Optional: Background box color (hex)
  position?: string;                    // Optional: Position ("center-bottom", etc.)
  "outline-color"?: string;             // Optional: Text outline color (hex)
  "outline-width"?: number;             // Optional: Outline width in pixels
}
```

#### Validation Rules

**Scene Validation:**
- `id`: Required, must be non-empty string
- `elements`: Required, must contain at least one element
- Must contain at least one audio element per scene

**Element Validation:**
- `type`: Required, must be one of: "audio", "image", "video", "subtitles"
- `src`: Required for audio/image/video elements
- `duration`: Must be positive number if specified
- `volume`: Must be between 0.0 and 1.0 if specified
- `x`, `y`: Must be non-negative if specified

**SubtitleSettings Validation:**
- `font-size`: Must be positive integer if specified
- `outline-width`: Must be non-negative if specified
- `shadow-offset`: Must be non-negative if specified
- Colors: Must be valid hex colors (e.g., "#FFFFFF") if specified

#### Response Format (Success)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "Video generation started",
  "created_at": "2024-01-15T10:30:00Z",
  "status_url": "/jobs/550e8400-e29b-41d4-a716-446655440000/status"
}
```

**HTTP Status**: `202 Accepted`

#### Response Format (Validation Error)

```json
{
  "error": "Invalid configuration",
  "details": "Scene 'intro' element 0: src is required for audio elements",
  "code": "VALIDATION_ERROR",
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**HTTP Status**: `400 Bad Request`

#### Example cURL Request

```bash
curl -X POST "https://api.yourdomain.com/api/v1/generate-video" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -H "X-CSRF-Token: your-csrf-token" \
  -d '{
    "scenes": [
      {
        "id": "intro",
        "background-color": "#1a1a1a",
        "elements": [
          {
            "type": "audio",
            "src": "https://example.com/audio/intro.mp3",
            "volume": 0.8,
            "duration": 15.0
          },
          {
            "type": "image",
            "src": "https://example.com/images/logo.png",
            "x": 100,
            "y": 50,
            "resize": "contain"
          },
          {
            "type": "subtitles",
            "language": "en",
            "settings": {
              "style": "progressive",
              "font-family": "Helvetica",
              "font-size": 28,
              "word-color": "#FFFFFF",
              "outline-color": "#000000",
              "outline-width": 3,
              "position": "center-bottom"
            }
          }
        ]
      }
    ]
  }'
```

## =Ê Job Management API

### GET /api/v1/jobs/{job_id}/status

Retrieves the current status of a video generation job.

#### Request Format

**URL Parameters:**
- `job_id`: UUID of the job

**Headers:**
- `Authorization: Bearer {api_key}` (if authentication enabled)

#### Response Format (In Progress)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": 75,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:32:30Z"
}
```

#### Response Format (Completed)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "progress": 100,
  "video_id": "550e8400-e29b-41d4-a716-446655440001",
  "download_url": "/download/550e8400-e29b-41d4-a716-446655440001",
  "completed_at": "2024-01-15T10:35:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

#### Response Format (Failed)

```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "failed",
  "progress": 45,
  "error": "Audio file not accessible: https://example.com/audio/intro.mp3",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:32:15Z"
}
```

**HTTP Status**: `200 OK`

#### Job Status Values

| Status | Description |
|--------|-------------|
| `pending` | Job created but not yet started |
| `processing` | Job is currently being processed |
| `completed` | Job completed successfully |
| `failed` | Job failed with error |
| `cancelled` | Job was cancelled |

### GET /api/v1/jobs

Lists all jobs with optional filtering.

#### Query Parameters

- `status`: Filter by job status (`pending`, `processing`, `completed`, `failed`)
- `limit`: Maximum number of jobs to return (default: 50, max: 100)
- `offset`: Number of jobs to skip for pagination (default: 0)

#### Response Format

```json
{
  "jobs": [
    {
      "job_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "completed",
      "progress": 100,
      "video_id": "video-001",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:35:00Z",
      "completed_at": "2024-01-15T10:35:00Z"
    },
    {
      "job_id": "550e8400-e29b-41d4-a716-446655440001",
      "status": "processing",
      "progress": 60,
      "created_at": "2024-01-15T10:40:00Z",
      "updated_at": "2024-01-15T10:42:00Z"
    }
  ],
  "total_count": 25,
  "pagination": {
    "limit": 50,
    "offset": 0,
    "has_more": false
  }
}
```

## =å Video Download API

### GET /api/v1/download/{video_id}

Downloads the generated video file.

#### Request Format

**URL Parameters:**
- `video_id`: UUID of the generated video

**Headers:**
- `Authorization: Bearer {api_key}` (if authentication enabled)

#### Response Format (Success)

**Headers:**
- `Content-Type: video/mp4`
- `Content-Disposition: attachment; filename="video_id.mp4"`
- `Content-Length: {file_size}`

**Body**: Binary video file data

**HTTP Status**: `200 OK`

#### Response Format (Not Found)

```json
{
  "error": "Video not found",
  "code": "VIDEO_NOT_FOUND",
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**HTTP Status**: `404 Not Found`

## = Security API

### GET /api/v1/csrf-token

Retrieves a CSRF token for subsequent requests.

#### Request Format

**Headers:**
- No authentication required for this endpoint

#### Response Format

```json
{
  "csrf_token": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6",
  "expires_in": 3600,
  "usage": "Include this token in the X-CSRF-Token header for state-changing requests"
}
```

**HTTP Status**: `200 OK`

#### Response Format (CSRF Disabled)

```json
{
  "error": "CSRF protection not enabled",
  "message": "CSRF token endpoint is disabled when CSRF protection is off"
}
```

**HTTP Status**: `404 Not Found`

## <å Health Check API

### GET /health

Basic health check endpoint.

#### Response Format

```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:00Z"
}
```

**HTTP Status**: `200 OK`

### GET /health/detailed

Detailed health check with system information.

#### Response Format

```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:00Z",
  "uptime": "2h30m45s",
  "system": {
    "go_version": "go1.21.5",
    "goroutines": 25,
    "memory": {
      "allocated": 15728640,
      "total_alloc": 157286400,
      "sys": 71303192,
      "heap_alloc": 15728640,
      "heap_sys": 67108864,
      "gc_cycles": 12
    }
  },
  "config": {
    "workers": 4,
    "queue_size": 100,
    "ffmpeg_path": "/usr/bin/ffmpeg",
    "output_dir": "./generated_videos"
  }
}
```

## L Error Response Format

### Standard Error Response

All API errors follow a consistent format:

```json
{
  "error": "Human-readable error message",
  "code": "MACHINE_READABLE_ERROR_CODE",
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": "Additional error details (optional)"
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `INVALID_REQUEST` | Malformed JSON or missing required fields | 400 |
| `VALIDATION_ERROR` | Request validation failed | 400 |
| `MISSING_AUTH_HEADER` | Authorization header missing | 401 |
| `INVALID_API_KEY` | Invalid or expired API key | 401 |
| `CSRF_TOKEN_MISSING` | CSRF token required but not provided | 403 |
| `CSRF_TOKEN_INVALID` | CSRF token validation failed | 403 |
| `JOB_NOT_FOUND` | Requested job does not exist | 404 |
| `VIDEO_NOT_FOUND` | Requested video does not exist | 404 |
| `RATE_LIMIT_EXCEEDED` | Too many requests | 429 |
| `INTERNAL_ERROR` | Server error occurred | 500 |

### Validation Error Details

For validation errors, additional context is provided:

```json
{
  "error": "Invalid configuration",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "scenes[0].elements[1].src",
    "value": "",
    "message": "src is required for image elements"
  },
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Rate Limiting Error

```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60,
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Headers:**
- `X-RateLimit-Limit: 100`
- `X-RateLimit-Remaining: 0`
- `X-RateLimit-Reset: 1705318260`

## = Authentication

### Bearer Token Authentication

Include API key in Authorization header:

```http
Authorization: Bearer your-api-key-here
```

### CSRF Protection

For state-changing requests (POST, PUT, DELETE), include CSRF token:

```http
X-CSRF-Token: your-csrf-token-here
```

## =Ê Request/Response Headers

### Common Request Headers

```http
Content-Type: application/json
Authorization: Bearer your-api-key
X-CSRF-Token: your-csrf-token
User-Agent: YourApp/1.0
```

### Common Response Headers

```http
Content-Type: application/json
X-Request-ID: req_123456789
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1705318260
```

## >ê Testing Examples

### JavaScript/Fetch Example

```javascript
// Get CSRF token
const csrfResponse = await fetch('/api/v1/csrf-token');
const { csrf_token } = await csrfResponse.json();

// Generate video
const response = await fetch('/api/v1/generate-video', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-api-key',
    'X-CSRF-Token': csrf_token
  },
  body: JSON.stringify({
    scenes: [
      {
        id: 'intro',
        elements: [
          {
            type: 'audio',
            src: 'https://example.com/audio.mp3',
            volume: 1.0
          },
          {
            type: 'subtitles',
            language: 'en',
            settings: {
              style: 'progressive',
              'font-size': 24,
              'word-color': '#FFFFFF'
            }
          }
        ]
      }
    ]
  })
});

const result = await response.json();
console.log('Job ID:', result.job_id);
```

### Python/Requests Example

```python
import requests
import time

# Configuration
API_BASE = 'https://api.yourdomain.com/api/v1'
API_KEY = 'your-api-key'

headers = {
    'Content-Type': 'application/json',
    'Authorization': f'Bearer {API_KEY}'
}

# Get CSRF token
csrf_response = requests.get(f'{API_BASE}/csrf-token')
csrf_token = csrf_response.json()['csrf_token']
headers['X-CSRF-Token'] = csrf_token

# Generate video
video_config = {
    'scenes': [
        {
            'id': 'intro',
            'background-color': '#000000',
            'elements': [
                {
                    'type': 'audio',
                    'src': 'https://example.com/audio.mp3',
                    'volume': 0.8,
                    'duration': 10.0
                },
                {
                    'type': 'subtitles',
                    'language': 'en',
                    'settings': {
                        'style': 'progressive',
                        'font-family': 'Arial',
                        'font-size': 24,
                        'word-color': '#FFFFFF',
                        'outline-color': '#000000'
                    }
                }
            ]
        }
    ]
}

# Start job
response = requests.post(
    f'{API_BASE}/generate-video',
    headers=headers,
    json=video_config
)

if response.status_code == 202:
    job_data = response.json()
    job_id = job_data['job_id']
    print(f'Job started: {job_id}')
    
    # Poll job status
    while True:
        status_response = requests.get(
            f'{API_BASE}/jobs/{job_id}/status',
            headers={'Authorization': f'Bearer {API_KEY}'}
        )
        
        status_data = status_response.json()
        print(f'Status: {status_data["status"]} ({status_data["progress"]}%)')
        
        if status_data['status'] in ['completed', 'failed']:
            break
            
        time.sleep(5)
    
    # Download video if completed
    if status_data['status'] == 'completed':
        video_id = status_data['video_id']
        download_response = requests.get(
            f'{API_BASE}/download/{video_id}',
            headers={'Authorization': f'Bearer {API_KEY}'}
        )
        
        with open(f'video_{video_id}.mp4', 'wb') as f:
            f.write(download_response.content)
        
        print(f'Video downloaded: video_{video_id}.mp4')
else:
    print(f'Error: {response.json()}')
```

This comprehensive request/response documentation provides everything needed to integrate with VideoCraft's API, including detailed schemas, validation rules, error handling, and practical examples.