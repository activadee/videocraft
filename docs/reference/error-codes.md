# Error Codes Reference

This document provides a comprehensive reference for all error codes used in VideoCraft, including their meanings, causes, and recommended solutions.

## =Ë Error Response Format

All API errors follow a standardized format:

```json
{
  "error": "Human-readable error message",
  "code": "MACHINE_READABLE_ERROR_CODE",
  "request_id": "req_123456789",
  "timestamp": "2024-01-15T10:30:00Z",
  "details": "Additional error details (optional)"
}
```

## =' Application Error Codes

### Video Processing Errors

#### INVALID_INPUT
**HTTP Status**: `400 Bad Request`
**Description**: Request validation failed or malformed input provided
**Causes**:
- Invalid JSON format in request body
- Missing required fields
- Invalid data types or values
- Malformed URLs or file paths

**Example Response**:
```json
{
  "error": "Invalid request format",
  "code": "INVALID_INPUT",
  "details": "Scene 'intro' is missing required 'elements' field"
}
```

**Solutions**:
- Validate JSON format before sending
- Ensure all required fields are present
- Check data types match API specification
- Verify URL formats are correct

#### FILE_NOT_FOUND
**HTTP Status**: `404 Not Found`
**Description**: Requested file or resource does not exist
**Causes**:
- Audio/video file URL returns 404
- Local file has been deleted or moved
- Incorrect file path in configuration
- File permissions prevent access

**Example Response**:
```json
{
  "error": "The requested file could not be found. Please verify the file exists.",
  "code": "FILE_NOT_FOUND"
}
```

**Solutions**:
- Verify file URLs are accessible
- Check file permissions and paths
- Ensure files haven't been moved or deleted
- Use absolute URLs for external resources

#### FFMPEG_FAILED
**HTTP Status**: `500 Internal Server Error`
**Description**: Video encoding or processing failed
**Causes**:
- Corrupted or invalid video/audio files
- Unsupported file formats
- FFmpeg binary not found or outdated
- Insufficient system resources

**Example Response**:
```json
{
  "error": "Video processing failed. Please check your input files and try again.",
  "code": "FFMPEG_FAILED"
}
```

**Solutions**:
- Verify input file formats are supported
- Check FFmpeg installation and version
- Ensure sufficient disk space and memory
- Try with smaller or different format files

#### TRANSCRIPTION_FAILED
**HTTP Status**: `500 Internal Server Error`
**Description**: Audio transcription process failed
**Causes**:
- Whisper AI model not available
- Python dependencies missing
- Audio file format not supported
- Insufficient memory for transcription

**Example Response**:
```json
{
  "error": "Audio transcription failed. Please ensure the audio file is valid.",
  "code": "TRANSCRIPTION_FAILED"
}
```

**Solutions**:
- Check Whisper AI installation
- Verify Python dependencies are installed
- Ensure audio file is in supported format
- Check available system memory

#### JOB_NOT_FOUND
**HTTP Status**: `404 Not Found`
**Description**: Requested job does not exist
**Causes**:
- Invalid job ID provided
- Job has been completed and cleaned up
- Job was cancelled or failed
- Database inconsistency

**Example Response**:
```json
{
  "error": "The requested job could not be found. It may have been completed or removed.",
  "code": "JOB_NOT_FOUND"
}
```

**Solutions**:
- Verify job ID is correct
- Check job status before making requests
- Handle job lifecycle appropriately
- Implement proper error handling

#### STORAGE_FAILED
**HTTP Status**: `500 Internal Server Error`
**Description**: File storage operation failed
**Causes**:
- Insufficient disk space
- File system permissions
- Storage service unavailable
- Corrupted storage system

**Example Response**:
```json
{
  "error": "Storage operation failed. Please try again later.",
  "code": "STORAGE_FAILED"
}
```

**Solutions**:
- Check available disk space
- Verify file system permissions
- Monitor storage service health
- Implement retry logic with backoff

#### DOWNLOAD_FAILED
**HTTP Status**: `500 Internal Server Error`
**Description**: Failed to download external resource
**Causes**:
- Network connectivity issues
- External server unavailable
- Invalid or expired URLs
- File size exceeds limits

**Example Response**:
```json
{
  "error": "Failed to download the specified resource. Please check the URL and try again.",
  "code": "DOWNLOAD_FAILED"
}
```

**Solutions**:
- Verify network connectivity
- Check external service availability
- Validate URL accessibility
- Ensure file size is within limits

#### TIMEOUT
**HTTP Status**: `408 Request Timeout`
**Description**: Operation exceeded timeout limit
**Causes**:
- Large file processing
- Slow network connections
- Overloaded system resources
- Configuration timeout too short

**Example Response**:
```json
{
  "error": "The request timed out. Please try again with a smaller file or shorter duration.",
  "code": "TIMEOUT"
}
```

**Solutions**:
- Reduce file size or duration
- Increase timeout configuration
- Optimize system resources
- Implement asynchronous processing

#### INTERNAL_ERROR
**HTTP Status**: `500 Internal Server Error`
**Description**: Unexpected internal system error
**Causes**:
- Unhandled exceptions
- Service dependencies failure
- Configuration errors
- System resource exhaustion

**Example Response**:
```json
{
  "error": "An internal error occurred. Please try again later or contact support.",
  "code": "INTERNAL_ERROR"
}
```

**Solutions**:
- Check server logs for details
- Verify service dependencies
- Review configuration settings
- Contact support if persistent

## = Security Error Codes

### Authentication Errors

#### MISSING_AUTH_HEADER
**HTTP Status**: `401 Unauthorized`
**Description**: Authorization header is required but not provided
**Causes**:
- Client not sending Authorization header
- Header name misspelled
- API key authentication enabled but not configured

**Example Response**:
```json
{
  "error": "Authorization header required",
  "code": "MISSING_AUTH_HEADER"
}
```

**Solutions**:
- Include `Authorization: Bearer {api_key}` header
- Verify header spelling and format
- Check API configuration settings

#### INVALID_AUTH_FORMAT
**HTTP Status**: `401 Unauthorized`
**Description**: Authorization header format is invalid
**Causes**:
- Missing "Bearer" prefix
- Malformed header value
- Extra spaces or characters

**Example Response**:
```json
{
  "error": "Invalid authorization format",
  "code": "INVALID_AUTH_FORMAT"
}
```

**Solutions**:
- Use format: `Authorization: Bearer {api_key}`
- Remove extra spaces or characters
- Verify header value construction

#### INVALID_API_KEY
**HTTP Status**: `401 Unauthorized`
**Description**: API key is invalid or expired
**Causes**:
- Incorrect API key value
- API key has been revoked
- API key has expired
- API key not configured on server

**Example Response**:
```json
{
  "error": "Invalid API key",
  "code": "INVALID_API_KEY"
}
```

**Solutions**:
- Verify API key value is correct
- Check if API key needs renewal
- Contact administrator for valid key
- Verify server-side configuration

### CSRF Protection Errors

#### CSRF_TOKEN_MISSING
**HTTP Status**: `403 Forbidden`
**Description**: CSRF token required but not provided
**Causes**:
- POST/PUT/DELETE request without CSRF token
- Missing X-CSRF-Token header
- CSRF protection enabled but token not retrieved

**Example Response**:
```json
{
  "error": "CSRF token required for state-changing requests",
  "code": "CSRF_TOKEN_MISSING"
}
```

**Solutions**:
- Get CSRF token from `/api/v1/csrf-token`
- Include `X-CSRF-Token: {token}` header
- Only required for POST/PUT/DELETE requests

#### CSRF_TOKEN_INVALID
**HTTP Status**: `403 Forbidden`
**Description**: CSRF token validation failed
**Causes**:
- Token has expired
- Token format is invalid
- Token was tampered with
- Wrong secret used for validation

**Example Response**:
```json
{
  "error": "Invalid CSRF token",
  "code": "CSRF_TOKEN_INVALID"
}
```

**Solutions**:
- Get fresh CSRF token
- Verify token format is correct
- Check token hasn't been modified
- Ensure proper token transmission

#### CSRF_TOKEN_MALFORMED
**HTTP Status**: `403 Forbidden`
**Description**: CSRF token format is malformed
**Causes**:
- Token contains invalid characters
- Token length is incorrect
- Token encoding is wrong
- Potential injection attempt

**Example Response**:
```json
{
  "error": "Invalid CSRF token format",
  "code": "CSRF_TOKEN_MALFORMED"
}
```

**Solutions**:
- Get new CSRF token from endpoint
- Don't modify token value
- Verify token transmission method
- Check for character encoding issues

### CORS Security Errors

#### CORS_ORIGIN_REJECTED
**HTTP Status**: `403 Forbidden`
**Description**: Request origin not in allowed domains list
**Causes**:
- Domain not in ALLOWED_DOMAINS configuration
- Cross-origin request from unauthorized domain
- Missing or incorrect Origin header

**Solutions**:
- Add domain to ALLOWED_DOMAINS configuration
- Use allowed domain for requests
- Configure proper CORS settings

#### CORS_SUSPICIOUS_ORIGIN
**HTTP Status**: `403 Forbidden`
**Description**: Origin contains suspicious patterns
**Causes**:
- Potential security attack
- Malformed origin header
- Suspicious URL schemes or patterns

**Solutions**:
- Use legitimate domain origins
- Check for malformed headers
- Verify request source legitimacy

### Rate Limiting Errors

#### RATE_LIMIT_EXCEEDED
**HTTP Status**: `429 Too Many Requests`
**Description**: Request rate limit exceeded
**Causes**:
- Too many requests in time window
- Client not implementing proper throttling
- Aggressive retry behavior

**Example Response**:
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

**Solutions**:
- Wait before retrying (see retry_after)
- Implement exponential backoff
- Reduce request frequency
- Check rate limit headers

## < HTTP Status Code Mapping

| HTTP Status | Description | Common Error Codes |
|-------------|-------------|--------------------|
| **400** | Bad Request | `INVALID_INPUT`, `CSRF_TOKEN_MALFORMED` |
| **401** | Unauthorized | `MISSING_AUTH_HEADER`, `INVALID_API_KEY` |
| **403** | Forbidden | `CSRF_TOKEN_MISSING`, `CORS_ORIGIN_REJECTED` |
| **404** | Not Found | `FILE_NOT_FOUND`, `JOB_NOT_FOUND` |
| **408** | Request Timeout | `TIMEOUT` |
| **429** | Too Many Requests | `RATE_LIMIT_EXCEEDED` |
| **500** | Internal Server Error | `FFMPEG_FAILED`, `INTERNAL_ERROR` |

## =Ê Error Response Headers

Rate limiting responses include additional headers:

```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705318260
Content-Type: application/json

{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 60
}
```

## = Debugging Error Codes

### Client-Side Debugging

1. **Check HTTP Status Code**
   ```javascript
   fetch('/api/v1/generate-video')
     .then(response => {
       if (!response.ok) {
         console.log('HTTP Status:', response.status);
         return response.json().then(error => {
           console.log('Error Code:', error.code);
           console.log('Error Message:', error.error);
         });
       }
     });
   ```

2. **Handle Specific Error Codes**
   ```javascript
   function handleError(error) {
     switch (error.code) {
       case 'INVALID_INPUT':
         // Show validation errors to user
         break;
       case 'CSRF_TOKEN_MISSING':
         // Get new CSRF token and retry
         break;
       case 'RATE_LIMIT_EXCEEDED':
         // Wait and retry with backoff
         break;
       default:
         // Generic error handling
     }
   }
   ```

### Server-Side Debugging

1. **Check Server Logs**
   ```bash
   # Find errors in logs
   grep "ERROR" /var/log/videocraft/app.log
   
   # Search for specific error codes
   grep "FFMPEG_FAILED" /var/log/videocraft/app.log
   ```

2. **Monitor Error Patterns**
   ```bash
   # Count error occurrences
   grep -o '"code":"[^"]*"' app.log | sort | uniq -c
   
   # Recent errors
   tail -f /var/log/videocraft/app.log | grep ERROR
   ```

## =à Error Handling Best Practices

### Client Implementation

```javascript
class VideoCraftClient {
  async makeRequest(endpoint, options = {}) {
    try {
      const response = await fetch(endpoint, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${this.apiKey}`,
          'X-CSRF-Token': await this.getCSRFToken(),
          ...options.headers
        }
      });

      if (!response.ok) {
        const error = await response.json();
        throw new VideoCraftError(error.code, error.error, response.status);
      }

      return await response.json();
    } catch (error) {
      return this.handleError(error);
    }
  }

  handleError(error) {
    if (error instanceof VideoCraftError) {
      switch (error.code) {
        case 'CSRF_TOKEN_MISSING':
        case 'CSRF_TOKEN_INVALID':
          // Refresh CSRF token and retry
          return this.refreshCSRFAndRetry();
        
        case 'RATE_LIMIT_EXCEEDED':
          // Implement exponential backoff
          return this.retryWithBackoff(error.retryAfter);
        
        case 'INVALID_INPUT':
          // Show validation errors to user
          this.showValidationErrors(error.message);
          break;
        
        default:
          // Generic error handling
          this.showGenericError(error.message);
      }
    }
    
    throw error;
  }
}
```

### Server Implementation

```go
func handleError(c *gin.Context, err error) {
    logEntry := map[string]interface{}{
        "error": err.Error(),
        "path": c.Request.URL.Path,
        "method": c.Request.Method,
        "client_ip": c.ClientIP(),
        "request_id": c.GetHeader("X-Request-ID"),
    }

    if vpe, ok := err.(*errors.VideoProcessingError); ok {
        logEntry["error_code"] = vpe.Code
        
        // Log security-sensitive errors differently
        if errors.IsSecuritySensitive(err) {
            logger.WithFields(errors.LogSecurityEvent(err)).Error("Security-sensitive error")
        } else {
            logger.WithFields(logEntry).Error("Video processing error")
        }
        
        c.JSON(getHTTPStatus(vpe.Code), errors.ToClientResponse(err))
    } else {
        logger.WithFields(logEntry).Error("Unexpected error")
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "An unexpected error occurred",
            "code": "INTERNAL_ERROR",
        })
    }
}
```

## =Þ Support and Troubleshooting

### Common Issues and Solutions

1. **Authentication Problems**
   - Verify API key configuration
   - Check authorization header format
   - Ensure HTTPS is used in production

2. **CSRF Issues**
   - Get fresh tokens for each session
   - Include tokens in all state-changing requests
   - Verify CSRF configuration is correct

3. **Rate Limiting**
   - Implement proper retry logic
   - Monitor request patterns
   - Consider request batching

4. **File Processing Errors**
   - Check file formats and sizes
   - Verify external URLs are accessible
   - Monitor system resources

### Getting Help

- **Documentation**: [Complete API Documentation](../api/overview.md)
- **Issues**: [GitHub Issues](https://github.com/activadee/videocraft/issues)
- **Security**: Report security issues to security@activadee.com
- **Support**: Contact support with error codes and request IDs

### Error Monitoring

Set up monitoring for these high-priority error codes:
- `INTERNAL_ERROR` - System health issues
- `RATE_LIMIT_EXCEEDED` - Traffic patterns
- `INVALID_API_KEY` - Authentication problems
- `FFMPEG_FAILED` - Processing capabilities

This comprehensive error code reference helps developers understand, debug, and handle all error conditions in VideoCraft effectively.