# API Authentication

VideoCraft requires authentication for all API endpoints.

## Authentication Method

API access uses Bearer token authentication.

```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3002/api/v1/health
```

## Security Features

- Required authentication
- CSRF protection
- Enhanced rate limiting with user-based limits

### Rate Limiting

VideoCraft implements user-based rate limiting that:
- Uses API keys from Bearer tokens to identify users
- Falls back to client IP for unauthenticated requests
- Bypasses rate limiting for health monitoring endpoints (`/health`, `/ready`, `/live`, `/metrics`)
- Logs violations with hashed user identifiers for security
- Returns professional HTTP 429 responses with standard headers

## Related Documentation

- [API Overview](overview.md) - API introduction
- [Security Overview](../security/overview.md) - Security architecture