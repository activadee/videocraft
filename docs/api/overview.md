# 🌐 API Overview

VideoCraft provides a comprehensive RESTful API for video generation, job management, and system monitoring. The API is built with security-first principles and designed for both programmatic access and interactive use.

## 🚀 API Highlights

### Core Capabilities
- **Async Video Generation**: Submit jobs and track progress
- **Scene-Based Configuration**: Complex video composition support
- **Progressive Subtitles**: AI-powered word-level timing
- **Security-First Design**: Comprehensive authentication and protection
- **Real-Time Monitoring**: Job status and system health endpoints

### API Characteristics
- **RESTful Design**: Standard HTTP methods and status codes
- **JSON-Based**: Request and response payloads in JSON
- **Stateless**: No server-side session management
- **Versioned**: API versioning for backward compatibility
- **Rate Limited**: Protection against abuse and overload

## 🔐 Security Model

### Authentication Required (v1.1.0+)
All API endpoints require authentication by default:

```bash
# Include API key in Authorization header
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/videos
```

### CORS & CSRF Protection
- **Strict Domain Allowlisting**: No wildcard origins allowed
- **CSRF Token Validation**: Required for state-changing requests
- **Origin Validation**: Suspicious pattern detection and blocking

### Security Headers
```bash
# Required headers for cross-origin requests
Origin: https://yourdomain.com
Authorization: Bearer YOUR_API_KEY
X-CSRF-Token: csrf-token-from-endpoint  # For POST/PUT/DELETE
```

## 📊 API Structure

### Base URL
```
http://localhost:3002/api/v1
```

### Endpoint Categories

#### 🎬 Video Generation
- `POST /videos` - Create video generation job
- `GET /videos` - List generated videos
- `GET /videos/{id}` - Get video information
- `GET /videos/{id}/download` - Download video file
- `DELETE /videos/{id}` - Delete video

#### 📋 Job Management
- `GET /jobs` - List all jobs
- `GET /jobs/{id}` - Get job details
- `GET /jobs/{id}/status` - Get job status
- `POST /jobs/{id}/cancel` - Cancel job

#### 🔐 Security
- `GET /csrf-token` - Get CSRF token for protected requests

#### 💊 System Health
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed system information
- `GET /metrics` - System metrics
- `GET /ready` - Kubernetes readiness probe
- `GET /live` - Kubernetes liveness probe

## 🎯 Request/Response Patterns

### Standard Response Format

#### Success Response
```json
{
  "success": true,
  "data": { /* response data */ },
  "message": "Operation completed successfully",
  "request_id": "req_123456789"
}
```

#### Error Response
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": "Scene 'intro' is missing audio element",
  "request_id": "req_123456789",
  "timestamp": "2024-01-01T12:00:00Z",
  "help_url": "https://docs.videocraft.io/errors/validation"
}
```

### Async Job Pattern
Video generation follows an async pattern:

1. **Submit Job**: `POST /videos` returns job ID immediately
2. **Track Progress**: Poll `GET /jobs/{id}/status` for updates
3. **Download Result**: Use `GET /videos/{id}/download` when complete

```bash
# 1. Submit video generation job
curl -X POST http://localhost:3002/api/v1/videos \
  -H "Authorization: Bearer $API_KEY" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d @config.json

# Response: {"job_id": "550e8400-e29b...", "status": "pending"}

# 2. Check job status
curl http://localhost:3002/api/v1/jobs/550e8400-e29b.../status \
  -H "Authorization: Bearer $API_KEY"

# Response: {"status": "processing", "progress": 75}

# 3. Download when complete
curl http://localhost:3002/api/v1/videos/550e8400-e29b.../download \
  -H "Authorization: Bearer $API_KEY" \
  -o result.mp4
```

## 📋 Common Request Examples

### Create Video with Progressive Subtitles
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
          "style": "progressive",
          "font-family": "Arial",
          "font-size": 24,
          "word-color": "#FFFFFF",
          "outline-color": "#000000"
        }
      }
    ]
  }'
```

### Get System Health
```bash
curl http://localhost:3002/api/v1/health/detailed \
  -H "Authorization: Bearer $API_KEY"
```

### List Videos with Pagination
```bash
curl http://localhost:3002/api/v1/videos?limit=10&offset=0 \
  -H "Authorization: Bearer $API_KEY"
```

## 🔄 Rate Limiting

### Default Limits
- **Requests per minute**: 100 per client IP
- **Burst allowance**: 20 requests
- **Rate limit headers**: Included in responses

### Rate Limit Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1640995200
Retry-After: 60
```

### Handle Rate Limiting
```bash
# Check rate limit headers in response
curl -I http://localhost:3002/api/v1/health \
  -H "Authorization: Bearer $API_KEY"

# Implement exponential backoff for 429 responses
```

## 🚨 Error Handling

### HTTP Status Codes
- `200 OK` - Successful request
- `202 Accepted` - Async job created
- `400 Bad Request` - Invalid request format
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - CSRF token required/invalid
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Error Code Reference
| Code | Description | Solution |
|------|-------------|----------|
| `MISSING_AUTH_HEADER` | Authorization header required | Include `Authorization: Bearer API_KEY` |
| `INVALID_API_KEY` | API key is invalid | Check API key value |
| `CSRF_TOKEN_REQUIRED` | CSRF token missing | Get token from `/csrf-token` |
| `CORS_ORIGIN_REJECTED` | Origin not allowed | Configure allowed domains |
| `VALIDATION_ERROR` | Request validation failed | Check request format |
| `RATE_LIMIT_EXCEEDED` | Too many requests | Implement rate limiting |

## 🎛️ API Configuration

### Security Configuration
```bash
# Required for API access
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com"
export VIDEOCRAFT_SECURITY_API_KEY="your-api-key"
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true

# Optional CSRF protection
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="your-csrf-secret"
```

### CORS Configuration
The API requires explicit domain configuration for cross-origin requests:

```bash
# Development
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"

# Production
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="app.yourcompany.com,api.yourcompany.com"
```

## 📚 Client Integration

### JavaScript/TypeScript Example
```typescript
class VideoCraftAPI {
  private baseURL = 'http://localhost:3002/api/v1';
  private apiKey: string;
  private csrfToken?: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getCSRFToken(): Promise<string> {
    if (!this.csrfToken) {
      const response = await fetch(`${this.baseURL}/csrf-token`);
      const data = await response.json();
      this.csrfToken = data.csrf_token;
    }
    return this.csrfToken;
  }

  async createVideo(config: VideoConfig): Promise<JobResponse> {
    const token = await this.getCSRFToken();
    
    const response = await fetch(`${this.baseURL}/videos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.apiKey}`,
        'X-CSRF-Token': token,
      },
      body: JSON.stringify(config),
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    return response.json();
  }

  async getJobStatus(jobId: string): Promise<JobStatus> {
    const response = await fetch(`${this.baseURL}/jobs/${jobId}/status`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });

    return response.json();
  }
}
```

### Python Example
```python
import requests
import time

class VideoCraftAPI:
    def __init__(self, api_key, base_url="http://localhost:3002/api/v1"):
        self.api_key = api_key
        self.base_url = base_url
        self.csrf_token = None
    
    def get_headers(self, include_csrf=False):
        headers = {"Authorization": f"Bearer {self.api_key}"}
        if include_csrf:
            headers["X-CSRF-Token"] = self.get_csrf_token()
        return headers
    
    def get_csrf_token(self):
        if not self.csrf_token:
            response = requests.get(f"{self.base_url}/csrf-token")
            self.csrf_token = response.json()["csrf_token"]
        return self.csrf_token
    
    def create_video(self, config):
        response = requests.post(
            f"{self.base_url}/videos",
            json=config,
            headers=self.get_headers(include_csrf=True)
        )
        response.raise_for_status()
        return response.json()
    
    def wait_for_completion(self, job_id, timeout=300):
        start_time = time.time()
        while time.time() - start_time < timeout:
            status = self.get_job_status(job_id)
            if status["status"] == "completed":
                return status
            elif status["status"] == "failed":
                raise Exception(f"Job failed: {status.get('error')}")
            time.sleep(5)
        raise TimeoutError("Job did not complete within timeout")
```

## 📚 Related Topics

### Authentication & Security
- **[API Authentication](authentication.md)** - Detailed authentication setup
- **[Security Overview](../security/overview.md)** - Comprehensive security architecture
- **[CORS & CSRF Protection](../security/cors-csrf.md)** - Cross-origin security

### API Reference
- **[Endpoints](endpoints.md)** - Complete endpoint reference
- **[Requests & Responses](requests-responses.md)** - Detailed format documentation
- **[Middleware](middleware.md)** - Middleware functionality

### Configuration
- **[Security Configuration](../configuration/security-configuration.md)** - Security setup
- **[Environment Variables](../configuration/environment-variables.md)** - Environment configuration

### Troubleshooting
- **[Error Codes](../reference/error-codes.md)** - Complete error reference
- **[Troubleshooting](../reference/troubleshooting.md)** - Common issues and solutions

## 🚀 Getting Started

1. **[Set up Authentication](authentication.md)** - Get your API key and configure security
2. **[Make your first request](endpoints.md#create-video)** - Generate your first video
3. **[Explore endpoints](endpoints.md)** - Complete API reference
4. **[Handle errors](../reference/error-codes.md)** - Error handling best practices

---

**🔗 Quick Links**: [Authentication Setup](authentication.md) | [Complete Endpoints](endpoints.md) | [Error Reference](../reference/error-codes.md)