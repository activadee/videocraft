# Environment Variables

Complete reference for all VideoCraft environment variables.

## Security Variables

```bash
# Domain allowlist (comma-separated)
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS=trusted.example.com,cdn.trusted.org

# Rate limiting (requests per minute per user)
VIDEOCRAFT_SECURITY_RATE_LIMIT=100

# Authentication
VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
```

## Server Configuration

```bash
# Server settings
VIDEOCRAFT_SERVER_PORT=3002
VIDEOCRAFT_SERVER_HOST=0.0.0.0
```

## Related Documentation

- [Configuration Overview](overview.md) - Configuration guide