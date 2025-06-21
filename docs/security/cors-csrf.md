# CORS & CSRF Protection

VideoCraft implements comprehensive HTTP-level security through specialized middleware components.

## CORS Security (`internal/api/middleware/cors.go`)

### Key Features
- **Zero Wildcard Policy**: Eliminates `AllowOrigins: ["*"]` vulnerability
- **Strict Domain Allowlisting**: Only explicitly configured domains permitted
- **Origin Validation Caching**: Thread-safe performance optimization
- **Suspicious Pattern Detection**: Blocks malicious origin patterns
- **Comprehensive Security Logging**: Structured audit trail

### Configuration
```bash
# Environment variable
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="trusted.example.com,api.trusted.org"
```

```yaml
# YAML configuration
security:
  allowed_domains:
    - "trusted.example.com"
    - "api.trusted.org"
```

## CSRF Protection (`internal/api/middleware/csrf.go`)

### Key Features
- **Token-Based Validation**: Cryptographically secure CSRF tokens
- **State-Change Protection**: POST, PUT, DELETE, PATCH requests require tokens
- **Enhanced Token Validation**: Format validation prevents injection attacks
- **Origin Correlation**: Cross-reference with CORS-allowed domains
- **Safe Method Exemption**: GET, HEAD, OPTIONS bypass CSRF checks

### Usage
```bash
# Get CSRF token
curl http://localhost:3002/api/v1/csrf-token

# Include token in request
curl -X POST -H "X-CSRF-Token: your-token" http://localhost:3002/api/v1/generate-video
```

For detailed HTTP security implementation, see: [`internal/api/middleware/SECURITY.md`](../../internal/api/middleware/SECURITY.md)