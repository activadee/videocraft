# 🔐 API Authentication

VideoCraft uses API key-based authentication to secure all endpoints. Starting with v1.1.0, authentication is **enabled by default** for enhanced security.

## 🚀 Quick Start

### 1. Get Your API Key
VideoCraft automatically generates a secure API key on first startup:

```bash
# Check logs for auto-generated API key
docker-compose logs videocraft | grep "API key"
# or
docker logs videocraft-container | grep "Generated API key"
```

### 2. Set Custom API Key (Optional)
```bash
export VIDEOCRAFT_SECURITY_API_KEY="your-secure-api-key-here"
```

### 3. Make Authenticated Request
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/health
```

## 🔑 API Key Management

### Auto-Generated Keys
When no API key is provided, VideoCraft generates a cryptographically secure key:

```
Example: bf4c7bc7d9187f50a68fd6466a39c424e75d0ed4510a4041bac7d2aa3515c883
```

**Key Properties**:
- **Length**: 64 hexadecimal characters
- **Entropy**: 256-bit security
- **Format**: Hexadecimal string
- **Uniqueness**: Cryptographically random

### Custom API Keys
For production environments, set your own API key:

```bash
# Environment variable
export VIDEOCRAFT_SECURITY_API_KEY="your-custom-secure-key"

# YAML configuration
security:
  api_key: "${VIDEOCRAFT_SECURITY_API_KEY}"
```

**Requirements**:
- **Minimum Length**: 32 characters
- **Recommended**: 64+ characters
- **Character Set**: Alphanumeric (avoid special characters)
- **Secrecy**: Never commit to version control

## 🔐 Authentication Methods

### Bearer Token (Recommended)
Include API key in Authorization header:

```bash
curl -H "Authorization: Bearer your-api-key" \
     http://localhost:3002/api/v1/videos
```

```javascript
// JavaScript example
fetch('/api/v1/videos', {
  headers: {
    'Authorization': 'Bearer your-api-key',
    'Content-Type': 'application/json'
  }
});
```

### Direct Header (Alternative)
Include API key directly in Authorization header:

```bash
curl -H "Authorization: your-api-key" \
     http://localhost:3002/api/v1/videos
```

### Query Parameter (Fallback)
Include API key as URL parameter (less secure):

```bash
curl "http://localhost:3002/api/v1/videos?api_key=your-api-key"
```

> **⚠️ Warning**: Query parameter method exposes API key in URLs and logs. Use only for testing.

## 🛡️ Security Configuration

### Enable/Disable Authentication
```bash
# Enable authentication (default in v1.1.0+)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true

# Disable authentication (development only)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
```

### YAML Configuration
```yaml
security:
  enable_auth: true
  api_key: "${VIDEOCRAFT_SECURITY_API_KEY}"
  rate_limit: 100
```

### Environment-Specific Setup

#### Development Environment
```bash
# Minimal security for local development
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
```

#### Production Environment
```bash
# Maximum security for production
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
export VIDEOCRAFT_SECURITY_API_KEY="production-secure-key"
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="app.yourcompany.com"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
```

## 🌐 Cross-Origin Authentication

### Required for CORS Requests
When making cross-origin requests, both authentication and CORS configuration are required:

```bash
# Configure allowed domains
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,api.yourdomain.com"
```

```javascript
// Frontend JavaScript
fetch('http://api.videocraft.com/api/v1/videos', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer your-api-key',
    'Content-Type': 'application/json',
    'Origin': 'https://yourdomain.com'  // Must be in allowed domains
  },
  body: JSON.stringify(videoConfig)
});
```

## 🔄 CSRF Protection Integration

### When CSRF is Enabled
For state-changing requests (POST, PUT, DELETE), both API key and CSRF token are required:

```bash
# 1. Get CSRF token
CSRF_TOKEN=$(curl -s http://localhost:3002/api/v1/csrf-token | jq -r '.csrf_token')

# 2. Make authenticated request with CSRF token
curl -X POST http://localhost:3002/api/v1/videos \
  -H "Authorization: Bearer $API_KEY" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d @config.json
```

### CSRF Configuration
```bash
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="your-csrf-secret"
```

## 📋 Protected Endpoints

### API Endpoints (Require Authentication)
- `POST /api/v1/videos` - Create video generation job
- `GET /api/v1/videos` - List videos
- `GET /api/v1/videos/{id}` - Get video details
- `GET /api/v1/videos/{id}/download` - Download video
- `DELETE /api/v1/videos/{id}` - Delete video
- `GET /api/v1/jobs/*` - Job management endpoints

### Public Endpoints (No Authentication)
- `GET /health` - Basic health check
- `GET /ready` - Kubernetes readiness probe
- `GET /live` - Kubernetes liveness probe
- `GET /metrics` - System metrics
- `GET /api/v1/csrf-token` - Get CSRF token

## ⚠️ Error Responses

### Missing API Key
```json
{
  "error": "Authorization header required",
  "code": "MISSING_AUTH_HEADER",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Invalid API Key
```json
{
  "error": "Invalid API key", 
  "code": "INVALID_API_KEY",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Invalid Format
```json
{
  "error": "Invalid authorization format",
  "code": "INVALID_AUTH_FORMAT", 
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 🔄 Migration Guide

### From v1.0.x to v1.1.0+

#### Before Upgrade
1. **Identify API Usage**: Find all client applications using the API
2. **Plan API Key Distribution**: Decide on key management strategy
3. **Test in Staging**: Verify authentication works with your setup

#### Required Changes
```bash
# 1. Enable authentication (now default)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true

# 2. Set API key (or use auto-generated)
export VIDEOCRAFT_SECURITY_API_KEY="your-secure-key"

# 3. Configure allowed domains for CORS
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com"
```

#### Update Client Code
```javascript
// Before v1.1.0 - No authentication
fetch('/api/v1/videos', {
  method: 'POST',
  body: JSON.stringify(config)
});

// v1.1.0+ - Authentication required
fetch('/api/v1/videos', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer your-api-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(config)
});
```

### Emergency Rollback
If authentication causes issues, temporarily disable it:

```bash
# Emergency bypass (NOT for production)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
systemctl restart videocraft
```

## 🔧 Client Implementation Examples

### JavaScript/TypeScript
```typescript
class VideoCraftAPI {
  private apiKey: string;
  private baseURL: string;

  constructor(apiKey: string, baseURL = 'http://localhost:3002/api/v1') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  private getHeaders(): Record<string, string> {
    return {
      'Authorization': `Bearer ${this.apiKey}`,
      'Content-Type': 'application/json'
    };
  }

  async createVideo(config: VideoConfig): Promise<JobResponse> {
    const response = await fetch(`${this.baseURL}/videos`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(config)
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  async getJobStatus(jobId: string): Promise<JobStatus> {
    const response = await fetch(`${this.baseURL}/jobs/${jobId}/status`, {
      headers: this.getHeaders()
    });

    return response.json();
  }
}
```

### Python
```python
import requests

class VideoCraftAPI:
    def __init__(self, api_key, base_url="http://localhost:3002/api/v1"):
        self.api_key = api_key
        self.base_url = base_url
    
    def get_headers(self):
        return {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
    
    def create_video(self, config):
        response = requests.post(
            f"{self.base_url}/videos",
            json=config,
            headers=self.get_headers()
        )
        response.raise_for_status()
        return response.json()
    
    def get_job_status(self, job_id):
        response = requests.get(
            f"{self.base_url}/jobs/{job_id}/status",
            headers=self.get_headers()
        )
        return response.json()
```

### cURL Scripts
```bash
#!/bin/bash
# api_client.sh

API_KEY="your-api-key"
BASE_URL="http://localhost:3002/api/v1"

# Create video
create_video() {
    local config_file=$1
    curl -X POST "${BASE_URL}/videos" \
        -H "Authorization: Bearer ${API_KEY}" \
        -H "Content-Type: application/json" \
        -d @"${config_file}"
}

# Get job status
get_job_status() {
    local job_id=$1
    curl "${BASE_URL}/jobs/${job_id}/status" \
        -H "Authorization: Bearer ${API_KEY}"
}

# Download video
download_video() {
    local video_id=$1
    local output_file=$2
    curl "${BASE_URL}/videos/${video_id}/download" \
        -H "Authorization: Bearer ${API_KEY}" \
        -o "${output_file}"
}
```

## 🔍 Troubleshooting

### Common Issues

#### Authentication Not Working
1. **Check API Key**: Verify key is correctly set
   ```bash
   echo $VIDEOCRAFT_SECURITY_API_KEY
   ```

2. **Check Format**: Ensure proper Bearer format
   ```bash
   curl -v -H "Authorization: Bearer $API_KEY" http://localhost:3002/api/v1/health
   ```

3. **Check Logs**: Look for authentication errors
   ```bash
   docker logs videocraft | grep -i auth
   ```

#### CORS Authentication Issues
1. **Check Allowed Domains**: Verify domain is configured
   ```bash
   echo $VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS
   ```

2. **Test Origin Header**: Include proper origin
   ```bash
   curl -H "Origin: https://yourdomain.com" \
        -H "Authorization: Bearer $API_KEY" \
        -X OPTIONS http://localhost:3002/api/v1/videos
   ```

### Debug Commands
```bash
# Test basic authentication
curl -v -H "Authorization: Bearer $API_KEY" http://localhost:3002/api/v1/health

# Test with invalid key
curl -v -H "Authorization: Bearer invalid-key" http://localhost:3002/api/v1/health

# Check authentication config
curl -I http://localhost:3002/api/v1/videos  # Should return 401
```

## 📚 Related Topics

### Security
- **[Security Overview](overview.md)** - Complete security architecture
- **[CORS & CSRF Protection](cors-csrf.md)** - Cross-origin security
- **[Security Configuration](../configuration/security-configuration.md)** - Environment setup

### API Usage
- **[API Overview](../api/overview.md)** - API introduction
- **[API Endpoints](../api/endpoints.md)** - Complete endpoint reference
- **[Error Codes](../reference/error-codes.md)** - Error handling reference

### Configuration
- **[Environment Variables](../configuration/environment-variables.md)** - All environment options
- **[Production Setup](../deployment/production-setup.md)** - Production configuration

---

**🔗 Next Steps**: [Configure CORS/CSRF](cors-csrf.md) | [Set up Production](../deployment/production-setup.md) | [API Reference](../api/endpoints.md)