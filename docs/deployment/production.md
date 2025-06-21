# Production Setup

Guide for deploying VideoCraft in production environments.

## Production Configuration

```bash
# Production security settings
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="app.yourcompany.com"
VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
```

## Deployment Best Practices

1. Use HTTPS in production
2. Enable authentication and CSRF protection
3. Configure proper domain allowlisting
4. Set up monitoring and logging
5. Implement backup strategies

## Related Documentation

- [Docker Deployment](docker.md) - Container deployment
- [Security Configuration](../security/overview.md) - Security setup