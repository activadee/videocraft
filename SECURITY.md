# Security Configuration Guide

## Authentication by Default

As of version 1.0, VideoCraft enables authentication by default for security. This document explains the authentication system and migration from previous versions.

## Default Security Settings

### Authentication Status
- **Default**: `enable_auth: true` (ENABLED)
- **API Key**: Auto-generated 64-character hexadecimal string (256-bit entropy)
- **Rate Limiting**: 100 requests per minute per client IP

### Configuration
```yaml
security:
  rate_limit: 100
  enable_auth: true
  # api_key: "auto-generated-if-not-provided"
```

## API Key Management

### Automatic Generation
When authentication is enabled but no API key is provided, VideoCraft automatically generates a cryptographically secure API key:

```bash
# Example auto-generated key (64 hex characters)
bf4c7bc7d9187f50a68fd6466a39c424e75d0ed4510a4041bac7d2aa3515c883
```

### Custom API Key
To use your own API key, set it via environment variable:

```bash
export VIDEOCRAFT_SECURITY_API_KEY="your-custom-api-key-here"
```

### API Key Requirements
- Minimum length: 32 characters (recommended: 64+ characters)
- Should contain alphanumeric characters
- Must be kept secret and secure

## Authentication Methods

### Bearer Token (Recommended)
```bash
curl -H "Authorization: Bearer your-api-key" http://localhost:3002/api/v1/videos
```

### Query Parameter (Fallback)
```bash
curl "http://localhost:3002/api/v1/videos?api_key=your-api-key"
```

### Direct Header
```bash
curl -H "Authorization: your-api-key" http://localhost:3002/api/v1/videos
```

## Endpoint Protection

### Protected Endpoints
All API endpoints require authentication when `enable_auth: true`:
- `/api/v1/*` - All v1 API endpoints
- `/generate-video` - Legacy video generation
- `/videos` - Legacy video management
- `/jobs/*` - Job management

### Unprotected Endpoints
Health and monitoring endpoints are always accessible:
- `/health` - Basic health check
- `/ready` - Kubernetes readiness probe
- `/live` - Kubernetes liveness probe
- `/metrics` - System metrics

## Migration from Previous Versions

### For Existing Deployments

If you're upgrading from a version where authentication was disabled by default:

1. **Immediate Action Required**: Set an API key before restarting
   ```bash
   export VIDEOCRAFT_SECURITY_API_KEY="your-secure-api-key"
   ```

2. **Update client applications** to include authentication headers

3. **For temporary compatibility** (NOT RECOMMENDED for production):
   ```bash
   export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
   ```

### Migration Checklist

- [ ] Generate or choose a secure API key
- [ ] Set `VIDEOCRAFT_SECURITY_API_KEY` environment variable
- [ ] Update all client applications to use authentication
- [ ] Test authentication with all endpoints
- [ ] Remove any temporary auth bypass settings
- [ ] Monitor logs for authentication errors

## Error Responses

### Missing API Key
```json
{
  "error": "API key is required",
  "code": "MISSING_API_KEY"
}
```

### Invalid API Key
```json
{
  "error": "Invalid API key",
  "code": "INVALID_API_KEY"
}
```

## Security Best Practices

### API Key Security
1. **Never commit API keys** to version control
2. **Use environment variables** for API key storage
3. **Rotate API keys regularly** in production
4. **Use different keys** for different environments
5. **Monitor API key usage** through logs

### Network Security
1. **Use HTTPS** in production environments
2. **Configure proper CORS** settings
3. **Implement rate limiting** (enabled by default)
4. **Monitor for suspicious activity**

### Infrastructure Security
1. **Restrict container permissions**
2. **Use non-root users** in containers
3. **Implement network segmentation**
4. **Regular security updates**

## Troubleshooting

### Authentication Not Working

1. **Check API key configuration**:
   ```bash
   echo $VIDEOCRAFT_SECURITY_API_KEY
   ```

2. **Verify authentication is enabled**:
   ```bash
   curl -v http://localhost:3002/api/v1/videos
   # Should return 401 if auth is enabled
   ```

3. **Test with correct authentication**:
   ```bash
   curl -H "Authorization: Bearer $VIDEOCRAFT_SECURITY_API_KEY" http://localhost:3002/api/v1/videos
   ```

### Common Issues

**Issue**: 401 Unauthorized despite correct API key
- **Solution**: Check for extra spaces or characters in the API key

**Issue**: Health endpoints returning 401
- **Solution**: Health endpoints should never require auth; check middleware configuration

**Issue**: Auto-generated API key not working
- **Solution**: Ensure no explicit API key is set in config file or environment

## Logging and Monitoring

### Authentication Events
VideoCraft logs all authentication events:

```
INFO  Authentication successful for endpoint /api/v1/videos
WARN  Authentication failed: Invalid API key for endpoint /api/v1/generate-video  
ERROR Authentication failed: Missing API key for endpoint /api/v1/jobs
```

### Security Monitoring
Monitor these metrics for security incidents:
- Failed authentication attempts
- Unusual request patterns
- Rate limit violations
- Access to sensitive endpoints

## Development vs Production

### Development Environment
For development, you can temporarily disable authentication:

```bash
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
```

**⚠️ WARNING**: Never disable authentication in production environments.

### Production Environment
Production deployments should always:
1. Enable authentication (`enable_auth: true`)
2. Use strong, unique API keys
3. Enable rate limiting
4. Use HTTPS
5. Monitor authentication logs

## Compliance and Auditing

### Security Logging
All authentication attempts are logged with:
- Timestamp
- Client IP address
- Endpoint accessed
- Authentication result
- API key hint (last 4 characters only)

### Audit Trail
For compliance requirements:
1. Enable detailed logging
2. Store logs securely
3. Implement log rotation
4. Monitor access patterns
5. Regular security reviews

## Emergency Procedures

### Compromised API Key
If an API key is compromised:

1. **Immediately change the API key**:
   ```bash
   export VIDEOCRAFT_SECURITY_API_KEY="new-secure-api-key"
   ```

2. **Restart the service**:
   ```bash
   systemctl restart videocraft
   ```

3. **Monitor logs** for suspicious activity

4. **Update all client applications** with new API key

5. **Review access logs** for unauthorized usage

### Service Lockout
If you're locked out due to authentication issues:

1. **Check environment variables**:
   ```bash
   env | grep VIDEOCRAFT_SECURITY
   ```

2. **Temporarily disable auth** for emergency access:
   ```bash
   export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
   systemctl restart videocraft
   ```

3. **Fix authentication configuration**

4. **Re-enable authentication**:
   ```bash
   export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
   systemctl restart videocraft
   ```

## Contact and Support

For security-related issues or questions:
- Create an issue: https://github.com/activadee/videocraft/issues
- Security vulnerabilities: Please report privately to the maintainers

## Version History

- **v1.0+**: Authentication enabled by default
- **v0.x**: Authentication disabled by default (deprecated)

---

**Remember**: Security is everyone's responsibility. Keep your API keys secure and monitor your deployments regularly.