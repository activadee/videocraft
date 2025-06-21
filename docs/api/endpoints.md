# =� API Endpoints Reference

Complete reference for all VideoCraft API endpoints with request/response examples, authentication requirements, and error handling.

## =� Base Information

- **Base URL**: `http://localhost:3002/api/v1`
- **Content-Type**: `application/json`
- **Authentication**: Bearer token required (v1.1.0+)
- **CSRF Protection**: Required for POST/PUT/DELETE/PATCH requests

## <� Video Generation Endpoints

### Create Video
Submit a video generation job for async processing.

**Endpoint**: `POST /generate-video`

**Authentication**: Required
**CSRF Token**: Required

#### Request
```bash
curl -X POST http://localhost:3002/api/v1/generate-video \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-CSRF-Token: CSRF_TOKEN" \
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
          "style": "progressive",
          "font-family": "Arial",
          "font-size": 24
        }
      }
    ]
  }'
```

#### Response (202 Accepted)
```json
{
  "success": true,
  "data": {
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "progress": 0,
    "created_at": "2024-01-01T12:00:00Z",
    "status_url": "/api/v1/jobs/550e8400-e29b-41d4-a716-446655440000/status"
  },
  "message": "Video generation job created successfully",
  "request_id": "req_123456789"
}
```

#### Request Schema
```typescript
interface VideoCreateRequest {
  scenes: Scene[];           // Array of scenes
  elements?: Element[];      // Global elements (subtitles, etc.)
  quality?: string;          // "low" | "medium" | "high"
  resolution?: string;       // "1920x1080" | "1280x720" | etc.
  width?: number;            // Custom width
  height?: number;           // Custom height
}

interface Scene {
  id?: string;               // Optional scene identifier
  elements: Element[];       // Scene elements
}

interface Element {
  type: string;              // "audio" | "image" | "video" | "subtitles"
  src?: string;              // Source URL (for audio/image/video)
  settings?: Record<string, any>; // Element-specific settings
  x?: number;                // X position (for images)
  y?: number;                // Y position (for images)
  width?: number;            // Width (for images)
  height?: number;           // Height (for images)
  volume?: number;           // Volume (0.0-1.0, for audio/video)
  language?: string;         // Language code (for subtitles)
}
```

### List Videos
Retrieve a list of generated videos with pagination.

**Endpoint**: `GET /videos`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     "http://localhost:3002/api/v1/videos?limit=10&offset=0&status=completed"
```

#### Query Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `limit` | integer | Number of results (1-100) | 20 |
| `offset` | integer | Results offset | 0 |
| `status` | string | Filter by status | all |
| `created_after` | string | ISO 8601 timestamp | - |
| `created_before` | string | ISO 8601 timestamp | - |

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "videos": [
      {
        "id": "video_123",
        "job_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "completed",
        "file_size": 15728640,
        "duration": 45.2,
        "resolution": "1920x1080",
        "created_at": "2024-01-01T12:00:00Z",
        "completed_at": "2024-01-01T12:02:30Z",
        "download_url": "/api/v1/videos/video_123/download"
      }
    ],
    "pagination": {
      "limit": 10,
      "offset": 0,
      "total": 42,
      "has_more": true
    }
  },
  "request_id": "req_987654321"
}
```

### Get Video Details
Retrieve detailed information about a specific video.

**Endpoint**: `GET /videos/{id}`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/videos/video_123
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "id": "video_123",
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "completed",
    "file_size": 15728640,
    "file_format": "mp4",
    "duration": 45.2,
    "resolution": "1920x1080",
    "framerate": 30,
    "audio_codec": "aac",
    "video_codec": "h264",
    "subtitle_tracks": [
      {
        "language": "en",
        "style": "progressive",
        "word_count": 127
      }
    ],
    "processing_stats": {
      "audio_analysis_time": 2.1,
      "transcription_time": 15.8,
      "subtitle_generation_time": 3.2,
      "video_encoding_time": 24.5
    },
    "created_at": "2024-01-01T12:00:00Z",
    "completed_at": "2024-01-01T12:02:30Z",
    "download_url": "/api/v1/videos/video_123/download",
    "preview_url": "/api/v1/videos/video_123/preview"
  },
  "request_id": "req_456789123"
}
```

### Download Video
Download the generated video file.

**Endpoint**: `GET /videos/{id}/download`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     -o "video.mp4" \
     http://localhost:3002/api/v1/videos/video_123/download
```

#### Response (200 OK)
- **Content-Type**: `video/mp4`
- **Content-Disposition**: `attachment; filename="video_123.mp4"`
- **Content-Length**: File size in bytes
- **Body**: Binary video data

### Delete Video
Delete a video and its associated files.

**Endpoint**: `DELETE /videos/{id}`

**Authentication**: Required
**CSRF Token**: Required

#### Request
```bash
curl -X DELETE \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-CSRF-Token: CSRF_TOKEN" \
  http://localhost:3002/api/v1/videos/video_123
```

#### Response (200 OK)
```json
{
  "success": true,
  "message": "Video deleted successfully",
  "request_id": "req_789123456"
}
```

## =� Job Management Endpoints

### List Jobs
Retrieve a list of processing jobs.

**Endpoint**: `GET /jobs`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     "http://localhost:3002/api/v1/jobs?status=processing&limit=20"
```

#### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by job status |
| `limit` | integer | Number of results (1-100) |
| `offset` | integer | Results offset |

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "jobs": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "processing",
        "progress": 75,
        "current_step": "Generating video with FFmpeg",
        "created_at": "2024-01-01T12:00:00Z",
        "updated_at": "2024-01-01T12:01:45Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 5
    }
  }
}
```

### Get Job Details
Retrieve detailed information about a specific job.

**Endpoint**: `GET /jobs/{id}`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/jobs/550e8400-e29b-41d4-a716-446655440000
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "processing",
    "progress": 75,
    "current_step": "Generating video with FFmpeg",
    "estimated_completion": "2024-01-01T12:03:00Z",
    "video_config": {
      "scenes": [ /* original configuration */ ],
      "elements": [ /* original configuration */ ]
    },
    "processing_log": [
      {
        "timestamp": "2024-01-01T12:00:05Z",
        "step": "audio_analysis",
        "status": "completed",
        "duration": 2.1
      },
      {
        "timestamp": "2024-01-01T12:00:20Z",
        "step": "transcription",
        "status": "completed",
        "duration": 15.8
      },
      {
        "timestamp": "2024-01-01T12:01:35Z",
        "step": "video_generation",
        "status": "in_progress",
        "progress": 75
      }
    ],
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:01:45Z"
  }
}
```

### Get Job Status
Get quick status information for a job (optimized for polling).

**Endpoint**: `GET /jobs/{id}/status`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/jobs/550e8400-e29b-41d4-a716-446655440000/status
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "completed",
    "progress": 100,
    "current_step": "Video generation completed",
    "result": {
      "video_id": "video_123",
      "download_url": "/api/v1/videos/video_123/download",
      "file_size": 15728640,
      "duration": 45.2
    },
    "updated_at": "2024-01-01T12:02:30Z"
  }
}
```

#### Job Status Values
- `pending` - Job queued, not yet started
- `processing` - Video generation in progress
- `completed` - Video generation successful
- `failed` - Processing failed
- `cancelled` - Job cancelled by user

### Cancel Job
Cancel a running job.

**Endpoint**: `POST /jobs/{id}/cancel`

**Authentication**: Required
**CSRF Token**: Required

#### Request
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "X-CSRF-Token: CSRF_TOKEN" \
  http://localhost:3002/api/v1/jobs/550e8400-e29b-41d4-a716-446655440000/cancel
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "job_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "cancelled",
    "cancelled_at": "2024-01-01T12:01:30Z"
  },
  "message": "Job cancelled successfully"
}
```

## = Security Endpoints

### Get CSRF Token
Retrieve a CSRF token for protected requests.

**Endpoint**: `GET /csrf-token`

**Authentication**: Not required

#### Request
```bash
curl http://localhost:3002/api/v1/csrf-token
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "csrf_token": "gAAAAABhZ3K4xvF7yY8sN9m2...",
    "expires_at": "2024-01-01T13:00:00Z"
  }
}
```

## =� System Health Endpoints

### Basic Health Check
Quick health status for load balancers.

**Endpoint**: `GET /health`

**Authentication**: Not required

#### Request
```bash
curl http://localhost:3002/health
```

#### Response (200 OK)
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.2.0"
}
```

### Detailed Health Check
Comprehensive system information.

**Endpoint**: `GET /health/detailed`

**Authentication**: Required

#### Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/health/detailed
```

#### Response (200 OK)
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.2.0",
    "uptime": "72h35m12s",
    "services": {
      "database": "healthy",
      "whisper_daemon": "healthy",
      "ffmpeg": "healthy",
      "storage": "healthy"
    },
    "system": {
      "memory": {
        "used": 1073741824,
        "total": 8589934592,
        "usage_percent": 12.5
      },
      "disk": {
        "used": 5368709120,
        "available": 21474836480,
        "usage_percent": 20.0
      },
      "cpu": {
        "usage_percent": 15.2,
        "load_average": [0.8, 0.9, 1.1]
      },
      "goroutines": 42
    },
    "jobs": {
      "active": 2,
      "pending": 0,
      "completed_today": 145,
      "failed_today": 3
    }
  }
}
```

### System Metrics
Prometheus-compatible metrics.

**Endpoint**: `GET /metrics`

**Authentication**: Not required

#### Request
```bash
curl http://localhost:3002/metrics
```

#### Response (200 OK)
```
# HELP videocraft_jobs_total Total number of jobs processed
# TYPE videocraft_jobs_total counter
videocraft_jobs_total{status="completed"} 1247
videocraft_jobs_total{status="failed"} 23

# HELP videocraft_job_duration_seconds Job processing duration
# TYPE videocraft_job_duration_seconds histogram
videocraft_job_duration_seconds_bucket{le="10"} 45
videocraft_job_duration_seconds_bucket{le="30"} 234
videocraft_job_duration_seconds_bucket{le="60"} 456

# HELP videocraft_api_requests_total Total API requests
# TYPE videocraft_api_requests_total counter
videocraft_api_requests_total{method="GET",status="200"} 5623
videocraft_api_requests_total{method="POST",status="202"} 1247
```

### Kubernetes Probes
Liveness and readiness probes for Kubernetes.

**Endpoints**: 
- `GET /live` - Liveness probe
- `GET /ready` - Readiness probe

**Authentication**: Not required

#### Request
```bash
curl http://localhost:3002/live
curl http://localhost:3002/ready
```

#### Response (200 OK)
```json
{"status": "ok"}
```

## � Error Responses

### Authentication Errors

#### Missing API Key (401)
```json
{
  "error": "Authorization header required",
  "code": "MISSING_AUTH_HEADER",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789",
  "help_url": "https://docs.videocraft.io/errors/authentication"
}
```

#### Invalid API Key (401)
```json
{
  "error": "Invalid API key",
  "code": "INVALID_API_KEY",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### CSRF Errors

#### Missing CSRF Token (403)
```json
{
  "error": "CSRF token required for this request",
  "code": "CSRF_TOKEN_REQUIRED",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789",
  "help_url": "https://docs.videocraft.io/errors/csrf"
}
```

### Validation Errors

#### Invalid Request Format (400)
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "scenes": "At least one scene is required",
    "elements[0].src": "URL is required for audio elements"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789",
  "help_url": "https://docs.videocraft.io/errors/validation"
}
```

### Rate Limiting (429)
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60,
  "limit": 100,
  "remaining": 0,
  "reset_at": "2024-01-01T12:01:00Z",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### Resource Not Found (404)
```json
{
  "error": "Video not found",
  "code": "VIDEO_NOT_FOUND",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### Server Error (500)
```json
{
  "error": "An internal error occurred. Please try again later or contact support.",
  "code": "INTERNAL_ERROR",
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789",
  "help_url": "https://docs.videocraft.io/errors/general"
}
```

## =� Client Libraries

### JavaScript/TypeScript SDK
```typescript
import { VideoCraftAPI } from '@videocraft/sdk';

const api = new VideoCraftAPI({
  baseURL: 'http://localhost:3002/api/v1',
  apiKey: 'your-api-key',
  enableCSRF: true
});

// Create video
const job = await api.videos.create({
  scenes: [{
    elements: [{
      type: 'audio',
      src: 'https://example.com/audio.mp3'
    }]
  }],
  elements: [{
    type: 'subtitles',
    settings: { style: 'progressive' }
  }]
});

// Poll for completion
const result = await api.jobs.waitForCompletion(job.job_id);
console.log('Video ready:', result.download_url);
```

### Python SDK
```python
from videocraft import VideoCraftAPI

api = VideoCraftAPI(
    base_url='http://localhost:3002/api/v1',
    api_key='your-api-key'
)

# Create video
job = api.videos.create({
    'scenes': [{
        'elements': [{
            'type': 'audio',
            'src': 'https://example.com/audio.mp3'
        }]
    }],
    'elements': [{
        'type': 'subtitles',
        'settings': {'style': 'progressive'}
    }]
})

# Wait for completion
result = api.jobs.wait_for_completion(job['job_id'])
print(f"Video ready: {result['download_url']}")
```

## =� Related Topics

### API Integration
- **[API Overview](overview.md)** - API introduction and concepts
- **[Authentication](authentication.md)** - API key setup and security
- **[Requests & Responses](requests-responses.md)** - Detailed format documentation

### Configuration
- **[Video Configuration](../video-generation/configuration.md)** - Complete video config reference
- **[Subtitle Settings](../subtitles/json-settings.md)** - Subtitle configuration (v1.1+)

### Security
- **[CORS & CSRF Protection](../security/cors-csrf.md)** - Cross-origin security
- **[Security Overview](../security/overview.md)** - Complete security architecture

### Troubleshooting
- **[Error Codes](../reference/error-codes.md)** - Complete error reference
- **[Troubleshooting Guide](../reference/troubleshooting.md)** - Common issues and solutions

---

**= Quick Links**: [Video Configuration](../video-generation/configuration.md) | [Authentication Setup](authentication.md) | [Error Reference](../reference/error-codes.md)