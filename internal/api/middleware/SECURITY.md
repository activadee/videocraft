# API Security Middleware Documentation

## Overview

This package implements comprehensive security middleware for the VideoCraft API, focusing on CORS and CSRF protection to prevent cross-site attacks and unauthorized access.

## Security Components

### 1. Secure CORS Implementation (`cors.go`)

#### Key Security Features
- **Wildcard Origin Removal**: Eliminates `AllowOrigins: ["*"]` vulnerability
- **Strict Domain Allowlisting**: Only explicitly configured domains are permitted
- **Origin Validation Caching**: Performance optimization with thread-safe cache
- **Suspicious Pattern Detection**: Blocks malicious origin patterns
- **Comprehensive Security Logging**: Structured logging for all security events

#### Configuration
```go
// Secure configuration example
config.Security{
    AllowedDomains: []string{
        "trusted.example.com",
        "api.trusted.org",
    },
}
```

#### Security Enhancements
- **Protocol Enforcement**: Supports both HTTP and HTTPS for allowed domains
- **Subdomain Protection**: Rejects subdomains unless explicitly allowed
- **Injection Prevention**: Detects and blocks suspicious origin patterns
- **Audit Trail**: All validation attempts are logged with structured data

### 2. CSRF Protection (`csrf.go`)

#### Key Security Features
- **Token-Based Protection**: Validates CSRF tokens for state-changing requests
- **Enhanced Token Validation**: Format validation prevents injection attacks
- **Origin Correlation**: Logs origin information for attack attribution
- **Safe Method Exemption**: GET, HEAD, OPTIONS requests bypass CSRF checks

#### Token Security
- **Secure Generation**: Cryptographically secure random token generation
- **Format Validation**: Prevents malformed tokens and injection attempts
- **Hexadecimal Encoding**: Ensures tokens contain only safe characters
- **Length Requirements**: Minimum 32 characters for adequate entropy

#### Request Monitoring
- **Comprehensive Logging**: All CSRF events logged with context
- **Client Attribution**: IP, User-Agent, Origin tracking
- **Threat Classification**: High/Medium threat level assignment
- **Audit Trail**: Complete request context for security analysis

## Security Configuration

### Environment Variables
```bash
# Domain allowlist (required for CORS)
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="trusted.example.com,api.trusted.org"

# CSRF protection (optional)
VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
VIDEOCRAFT_SECURITY_CSRF_SECRET="your-secure-secret-key"
```

### YAML Configuration
```yaml
security:
  allowed_domains:
    - "trusted.example.com"
    - "api.trusted.org"
  enable_csrf: true
  csrf_secret: "your-secure-secret-key"
```

## Security Policies

### CORS Security Policy
1. **Zero Wildcard Tolerance**: No `*` origins allowed under any circumstances
2. **Explicit Domain Approval**: Every domain must be explicitly configured
3. **Protocol Awareness**: Both HTTP and HTTPS variants must be handled
4. **Subdomain Restriction**: Subdomains are blocked unless explicitly allowed
5. **Credential Limitation**: Credentials only allowed with single domain configs

### CSRF Security Policy
1. **State-Change Protection**: All POST, PUT, DELETE, PATCH requests require tokens
2. **Safe Method Exemption**: GET, HEAD, OPTIONS bypass CSRF validation
3. **Token Format Enforcement**: Strict token format prevents injection
4. **Origin Validation**: Cross-reference with CORS-allowed domains
5. **Comprehensive Logging**: All attempts logged for security monitoring

## Attack Prevention

### Prevented Attack Vectors
- **Cross-Site Request Forgery (CSRF)**: Token validation prevents unauthorized actions
- **Cross-Origin Attacks**: Strict domain allowlisting blocks malicious origins
- **Subdomain Takeover**: Explicit domain matching prevents subdomain attacks
- **Protocol Injection**: Pattern detection blocks `javascript:`, `data:` schemes
- **Token Injection**: Format validation prevents malformed token attacks

### Detection Capabilities
- **Suspicious Origins**: Automatic detection of malicious patterns
- **Invalid Tokens**: Malformed CSRF token detection and logging
- **Unauthorized Domains**: Real-time blocking and logging of non-allowlisted origins
- **Attack Attribution**: IP and User-Agent correlation for threat analysis

## Security Monitoring

### Log Structure
All security events include structured fields:
```json
{
  "level": "WARN",
  "message": "CORS_SECURITY_VIOLATION: Origin not in allowlist",
  "fields": {
    "origin": "https://malicious.example.com",
    "allowed_domains": ["trusted.example.com"],
    "violation_type": "CORS_ORIGIN_REJECTED",
    "client_ip": "192.168.1.100",
    "threat_level": "MEDIUM",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Security Event Types
- `CORS_ORIGIN_REJECTED`: Unauthorized origin blocked
- `CORS_SUSPICIOUS_ORIGIN`: Malicious pattern detected
- `CSRF_TOKEN_MISSING`: Missing CSRF token
- `CSRF_TOKEN_INVALID`: Invalid CSRF token
- `CSRF_TOKEN_MALFORMED`: Malformed token detected

## Performance Optimizations

### CORS Optimizations
- **Origin Validation Caching**: Thread-safe cache for repeated origins
- **Efficient Pattern Matching**: Optimized string operations
- **Early Rejection**: Fast-fail for obviously malicious patterns

### CSRF Optimizations
- **Safe Method Bypass**: Quick path for read-only requests
- **Format Pre-validation**: Early rejection of malformed tokens
- **Minimal Overhead**: Lightweight token validation process

## Testing Coverage

### Security Test Coverage
- **Wildcard Origin Prevention**: Ensures no `*` origins are accepted
- **Domain Allowlisting**: Validates strict domain enforcement
- **CSRF Token Validation**: Comprehensive token security testing
- **Suspicious Pattern Detection**: Malicious origin pattern blocking
- **Security Logging**: Verification of audit trail completeness

### Test Categories
- **Positive Security Tests**: Verify legitimate requests are allowed
- **Negative Security Tests**: Confirm malicious requests are blocked
- **Edge Case Testing**: Handle unusual but valid scenarios
- **Performance Testing**: Ensure security doesn't impact performance
- **Regression Testing**: Prevent security feature degradation

## Integration Guidelines

### Development Environment
```yaml
# Development-friendly configuration
security:
  allowed_domains:
    - "localhost:3000"
    - "127.0.0.1:3000"
  enable_csrf: false  # Disabled for easier development
```

### Production Environment
```yaml
# Production security configuration
security:
  allowed_domains:
    - "app.yourcompany.com"
    - "api.yourcompany.com"
  enable_csrf: true
  csrf_secret: "${CSRF_SECRET}"  # From environment
```

### Staging Environment
```yaml
# Staging environment configuration
security:
  allowed_domains:
    - "staging.yourcompany.com"
    - "staging-api.yourcompany.com"
  enable_csrf: true
  csrf_secret: "${STAGING_CSRF_SECRET}"
```

## Security Best Practices

### Configuration Management
1. **Environment-Specific Configs**: Use different settings per environment
2. **Secret Management**: Store CSRF secrets in secure vaults
3. **Domain Validation**: Verify domain ownership before allowlisting
4. **Regular Review**: Periodically audit allowed domains list

### Monitoring and Alerting
1. **Security Log Monitoring**: Monitor for security violation patterns
2. **Threat Detection**: Alert on high threat level events
3. **Rate Limiting**: Monitor for unusual request patterns
4. **Incident Response**: Establish procedures for security events

### Deployment Security
1. **Configuration Validation**: Validate security config at startup
2. **Health Checks**: Include security middleware in health monitoring
3. **Gradual Rollout**: Test security changes in staging first
4. **Rollback Plan**: Maintain ability to quickly disable security features

## Troubleshooting

### Common Issues

#### CORS Errors
- **Symptom**: "CORS policy" errors in browser console
- **Solution**: Add your domain to `allowed_domains` configuration
- **Debug**: Check server logs for `CORS_ORIGIN_REJECTED` events

#### CSRF Errors
- **Symptom**: "CSRF token required" errors on POST requests
- **Solution**: Get token from `/api/v1/csrf-token` and include in headers
- **Debug**: Check for `CSRF_TOKEN_MISSING` or `CSRF_TOKEN_INVALID` logs

#### Performance Issues
- **Symptom**: Slow origin validation
- **Solution**: Origin caching should resolve automatically
- **Debug**: Monitor cache hit rates in debug logs

### Debug Commands
```bash
# Check current CORS configuration
curl -H "Origin: https://yoursite.com" -X OPTIONS http://localhost:3002/api/v1/videos

# Get CSRF token
curl http://localhost:3002/api/v1/csrf-token

# Test CSRF protection
curl -X POST -H "Content-Type: application/json" \
     -H "X-CSRF-Token: your-token" \
     http://localhost:3002/api/v1/generate-video
```

## Security Compliance

This middleware implementation helps achieve compliance with:
- **OWASP Top 10**: Addresses CSRF and Cross-Origin vulnerabilities
- **NIST Cybersecurity Framework**: Implements Identify, Protect, Detect functions
- **SOC 2 Type II**: Provides security controls and audit logging
- **GDPR**: Ensures unauthorized cross-origin access is prevented

## Maintenance

### Regular Security Reviews
- Review allowed domains quarterly
- Update suspicious pattern detection rules
- Monitor security logs for new attack patterns
- Update CSRF secret rotation procedures

### Version Updates
- Test security middleware with framework updates
- Validate continued effectiveness after dependency updates
- Benchmark performance impact of security features
- Update documentation with any configuration changes