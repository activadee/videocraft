# Migration Guide: Legacy to v1 API Endpoints

## 🚨 Important Notice

**Legacy API endpoints have been removed in VideoCraft v2.0.0**

All legacy endpoints are now deprecated and will return `404 Not Found`. Please migrate to the versioned `/api/v1/` endpoints immediately.

## 📋 Quick Migration Overview

| ❌ Legacy Endpoint (REMOVED) | ✅ New v1 Endpoint | Migration Status |
|-------------------------------|-------------------|------------------|
| `POST /generate-video` | `POST /api/v1/generate-video` | ✅ Direct replacement |
| `GET /videos` | `GET /api/v1/videos` | ✅ Direct replacement |
| `GET /download/:video_id` | `GET /api/v1/download/:video_id` | ✅ Direct replacement |
| `GET /status/:video_id` | `GET /api/v1/status/:video_id` | ✅ Direct replacement |
| `DELETE /videos/:video_id` | `DELETE /api/v1/videos/:video_id` | ✅ Direct replacement |
| `GET /jobs` | `GET /api/v1/jobs` | ✅ Direct replacement |
| `GET /jobs/:job_id` | `GET /api/v1/jobs/:job_id` | ✅ Direct replacement |
| `GET /jobs/:job_id/status` | `GET /api/v1/jobs/:job_id/status` | ✅ Direct replacement |
| `POST /jobs/:job_id/cancel` | `POST /api/v1/jobs/:job_id/cancel` | ✅ Direct replacement |

## 🔄 Migration Steps

### Step 1: Update Base URL
Simply add `/api/v1` prefix to all your existing endpoint calls:

```diff
- POST https://your-domain.com/generate-video
+ POST https://your-domain.com/api/v1/generate-video
```

### Step 2: Test Your Changes
Verify all endpoints work correctly with the new URLs.

### Step 3: Deploy Updates
Deploy your updated client code to production.

## 💻 Code Examples

### Python Migration
```python
import requests

# ❌ OLD (v1.x) - Will return 404 in v2.0+
response = requests.post(
    "https://api.videocraft.com/generate-video",
    json={
        "background_video": "https://example.com/background.mp4",
        "audio_files": [
            {
                "url": "https://example.com/audio1.mp3",
                "label": "Speaker 1"
            }
        ],
        "subtitle_settings": {
            "enabled": True,
            "style": "progressive"
        }
    },
    headers={"Authorization": "Bearer YOUR_API_KEY"}
)

# ✅ NEW (v2.0+) - Correct endpoint
response = requests.post(
    "https://api.videocraft.com/api/v1/generate-video",  # Added /api/v1 prefix
    json={
        "background_video": "https://example.com/background.mp4",
        "audio_files": [
            {
                "url": "https://example.com/audio1.mp3",
                "label": "Speaker 1"
            }
        ],
        "subtitle_settings": {
            "enabled": True,
            "style": "progressive"
        }
    },
    headers={"Authorization": "Bearer YOUR_API_KEY"}
)
```

### JavaScript Migration
```javascript
// ❌ OLD (v1.x) - Will return 404 in v2.0+
const response = await fetch('/generate-video', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer YOUR_API_KEY'
    },
    body: JSON.stringify({
        background_video: 'https://example.com/background.mp4',
        audio_files: [
            {
                url: 'https://example.com/audio1.mp3',
                label: 'Speaker 1'
            }
        ],
        subtitle_settings: {
            enabled: true,
            style: 'progressive'
        }
    })
});

// ✅ NEW (v2.0+) - Correct endpoint  
const response = await fetch('/api/v1/generate-video', {  // Added /api/v1 prefix
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer YOUR_API_KEY'
    },
    body: JSON.stringify({
        background_video: 'https://example.com/background.mp4',
        audio_files: [
            {
                url: 'https://example.com/audio1.mp3',
                label: 'Speaker 1'
            }
        ],
        subtitle_settings: {
            enabled: true,
            style: 'progressive'
        }
    })
});
```

### Go Migration
```go
// ❌ OLD (v1.x) - Will return 404 in v2.0+
resp, err := http.Post(
    "https://api.videocraft.com/generate-video",
    "application/json",
    bytes.NewBuffer(jsonData),
)

// ✅ NEW (v2.0+) - Correct endpoint
resp, err := http.Post(
    "https://api.videocraft.com/api/v1/generate-video",  // Added /api/v1 prefix
    "application/json", 
    bytes.NewBuffer(jsonData),
)
```

### cURL Migration
```bash
# ❌ OLD (v1.x) - Will return 404 in v2.0+
curl -X POST https://api.videocraft.com/generate-video \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"background_video": "https://example.com/background.mp4"}'

# ✅ NEW (v2.0+) - Correct endpoint  
curl -X POST https://api.videocraft.com/api/v1/generate-video \  # Added /api/v1 prefix
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{"background_video": "https://example.com/background.mp4"}'
```

## 🔍 Detailed Endpoint Migration

### Video Generation
```diff
- POST /generate-video
+ POST /api/v1/generate-video
```
**No changes** to request/response format, headers, or authentication.

### Video Management
```diff
- GET /videos
+ GET /api/v1/videos

- GET /download/:video_id  
+ GET /api/v1/download/:video_id

- GET /status/:video_id
+ GET /api/v1/status/:video_id

- DELETE /videos/:video_id
+ DELETE /api/v1/videos/:video_id
```
**No changes** to functionality, parameters, or response formats.

### Job Management  
```diff
- GET /jobs
+ GET /api/v1/jobs

- GET /jobs/:job_id
+ GET /api/v1/jobs/:job_id

- GET /jobs/:job_id/status
+ GET /api/v1/jobs/:job_id/status

- POST /jobs/:job_id/cancel
+ POST /api/v1/jobs/:job_id/cancel
```
**No changes** to job status values, response formats, or polling behavior.

## 🚨 Error Handling for Legacy Endpoints

If you accidentally call a legacy endpoint in v2.0+, you'll receive:

```json
{
    "error": "Not Found",
    "code": "ENDPOINT_NOT_FOUND", 
    "message": "The requested endpoint does not exist",
    "request_id": "req_123456789",
    "timestamp": "2024-12-16T10:30:00Z"
}
```

**HTTP Status**: `404 Not Found`

## ✅ Migration Checklist

### Pre-Migration
- [ ] Review all API calls in your codebase
- [ ] Identify all legacy endpoint usage
- [ ] Plan deployment timeline
- [ ] Test in development environment

### During Migration
- [ ] Update all endpoint URLs to include `/api/v1` prefix
- [ ] Update any hardcoded URLs in configuration
- [ ] Update documentation and examples  
- [ ] Test all API functionality

### Post-Migration Validation
- [ ] Verify all API calls return successful responses
- [ ] Check error logs for any 404 errors
- [ ] Monitor application performance
- [ ] Validate all features work correctly

## 🛠️ Troubleshooting

### Common Issues

**Issue**: Getting 404 errors after upgrading to v2.0.0
**Solution**: Ensure all API calls use `/api/v1/` prefix

**Issue**: Authentication not working
**Solution**: Verify your API key is correct and included in the `Authorization` header

**Issue**: Responses look different
**Solution**: Response formats are identical between legacy and v1 endpoints

### Getting Help

If you encounter issues during migration:

1. **Check the migration guide**: Ensure you've followed all steps correctly
2. **Review our API documentation**: [docs/api/overview.md](../api/overview.md)
3. **Contact support**: Include error messages and request IDs
4. **Check our troubleshooting guide**: [docs/troubleshooting/overview.md](../troubleshooting/overview.md)

## 📚 Additional Resources

- [API Overview](../api/overview.md) - Complete API documentation
- [Authentication Guide](../api/authentication.md) - Security implementation details  
- [Breaking Changes v2.0.0](../api/breaking-changes-v2.md) - Complete list of v2.0.0 changes
- [Examples Repository](https://github.com/activadee/videocraft-examples) - Sample implementations

## 🕐 Migration Timeline

- **v1.x**: Legacy endpoints supported alongside `/api/v1/` endpoints
- **v2.0.0**: Legacy endpoints removed, only `/api/v1/` endpoints available
- **Support**: Extended support available through normal channels

---

**Migration Required**: ✅ Required for v2.0.0 compatibility  
**Effort Level**: 🟢 Low (URL prefix changes only)  
**Breaking Changes**: 🔴 Yes (legacy endpoints removed)  
**Timeline**: 📅 Immediate (v2.0.0 release)

For technical support during migration, please contact our support team with specific error messages and request IDs.