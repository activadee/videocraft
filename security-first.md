# VideoCraft Security-First Implementation

## 🔒 Critical Security Requirements

VideoCraft implements a **security-first architecture** that prioritizes protection against web vulnerabilities, command injection, and unauthorized access. This document outlines the mandatory security requirements for production deployment.

## 🚫 Zero-Tolerance Security Policies

### 1. CORS Wildcard Prohibition
- **CRITICAL**: `AllowOrigins: ["*"]` is **NEVER** permitted
- **Enforcement**: Strict domain allowlisting only
- **Impact**: Prevents cross-origin attacks and data exfiltration
- **Configuration Required**: `VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS` must be set

### 2. Domain Allowlisting Requirement
- **CRITICAL**: All domains must be explicitly approved
- **Format**: `trusted.example.com,api.trusted.org`
- **Validation**: Exact domain matching (no subdomain wildcards)
- **Monitoring**: All rejected origins are logged with threat levels

### 3. CSRF Protection for State Changes
- **CRITICAL**: All POST/PUT/DELETE requests require CSRF tokens
- **Implementation**: Token-based validation with format checking
- **Safe Methods**: GET/HEAD/OPTIONS bypass token validation
- **Token Source**: `/api/v1/csrf-token` endpoint

## 🛡️ Mandatory Security Controls

### HTTP Security Layer
```yaml
# REQUIRED: Production security configuration
security:
  allowed_domains:
    - "yourdomain.com"          # Replace with actual production domain
    - "api.yourdomain.com"      # Replace with actual API domain
  enable_csrf: true             # REQUIRED in production
  csrf_secret: "${CSRF_SECRET}" # REQUIRED: Secure random secret
  enable_auth: true             # REQUIRED: API authentication
  rate_limit: 100               # REQUIRED: Request rate limiting
```

### FFmpeg Command Injection Prevention
- **URL Validation**: Protocol allowlist (HTTP/HTTPS only)
- **Character Filtering**: Blocks shell metacharacters `;&|$(){}`
- **Path Traversal Protection**: Prevents `../` directory traversal
- **Domain Validation**: Optional domain allowlist for external resources

### Input Sanitization
- **JSON Schema Validation**: Strict schema enforcement
- **File Type Validation**: Allowed media types only
- **Size Limits**: Maximum file size enforcement
- **Content-Type Verification**: MIME type validation

## 📊 Security Monitoring Requirements

### Mandatory Logging
All security events MUST be logged with structured data:
```json
{
  "violation_type": "CORS_ORIGIN_REJECTED",
  "threat_level": "MEDIUM|HIGH",
  "origin": "blocked-domain",
  "client_ip": "source-ip",
  "timestamp": "ISO-8601"
}
```

### Alert Conditions
- **HIGH Threat Level**: Immediate alert required
- **Repeated Violations**: Rate-based alerting
- **Suspicious Patterns**: Pattern-based detection
- **Token Injection**: Malformed CSRF token attempts of injection

## ❌ Prohibited Configurations

### NEVER Allow
- `AllowOrigins: ["*"]` - Universal CORS access
- `localhost` domains in production
- Unencrypted CSRF secrets
- Disabled authentication in production
- Missing rate-limiting
- Debug mode in production

### Dangerous Patterns
- Protocol injection: `javascript:`, `data:`, `file:`
- Command injection: Shell metacharacters in URLs
- Path traversal: `../` sequences
- Cross-site scripting: HTML/JS in user input

## 🌍 Environment-Specific Requirements

### Development Environment
```bash
# Minimal security for development
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=false
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
```

### Staging Environment
```bash
# Production-like security for staging
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="staging.yourcompany.com"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="staging-secure-secret"
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
```

### Production Environment
```bash
# Maximum security for production
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="app.yourcompany.com,api.yourcompany.com"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="production-secure-secret"
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
export VIDEOCRAFT_SECURITY_RATE_LIMIT=100
```

## 🔑 Authentication Requirements

### API Key Security
- **Generation**: Cryptographically secure 256-bit keys
- **Storage**: Environment variables or secure vaults only
- **Rotation**: Regular key rotation (quarterly recommended)
- **Transmission**: HTTPS only, never in logs

### Bearer Token Format
```http
Authorization: Bearer your-secure-api-key
```

## 🚨 Incident Response Procedures

### Security Violation Response
1. **Immediate**: Log violation with full context
2. **Assessment**: Determine threat level and scope
3. **Containment**: Block malicious sources if needed
4. **Investigation**: Analyze attack patterns
5. **Remediation**: Implement additional protections

### CORS Attack Response
1. **Detection**: Monitor for `CORS_ORIGIN_REJECTED` events
2. **Analysis**: Check for coordinated attack patterns
3. **Blocking**: Consider IP-based blocking for persistent attacks
4. **Review**: Validate current domain allowlist

### CSRF Attack Response
1. **Detection**: Monitor for `CSRF_TOKEN_INVALID` events
2. **Analysis**: Correlate with other suspicious activities
3. **Protection**: Implement additional token validation
4. **Communication**: Notify affected users if needed

## 🧪 Security Testing Requirements

### Mandatory Tests
- **CORS Wildcard Prevention**: Ensure `*` origins rejected
- **Domain Allowlisting**: Verify strict domain enforcement
- **CSRF Token Validation**: Test token requirement and format
- **Command Injection Prevention**: Validate URL sanitization
- **Authentication Bypass**: Ensure auth cannot be bypassed

### Security Test Commands
```bash
# Test CORS rejection
curl -H "Origin: https://malicious.com" -X OPTIONS http://api/endpoint

# Test CSRF requirement
curl -X POST -H "Content-Type: application/json" http://api/endpoint

# Test authentication requirement
curl -X POST http://api/protected-endpoint

# Test command injection prevention
curl -X POST -d '{"url":"http://test.com; rm -rf /"}' http://api/endpoint
```

## 🔧 Security Maintenance

### Regular Security Tasks
- **Weekly**: Review security logs for anomalies
- **Monthly**: Update domain allowlists as needed
- **Quarterly**: Rotate API keys and CSRF secrets
- **Annually**: Full security architecture review

### Security Updates
- **Dependency Updates**: Regular security patch application
- **Configuration Review**: Validate all security settings
- **Test Updates**: Maintain comprehensive security test coverage
- **Documentation**: Keep security docs current

## 📋 Compliance Requirements

### Standards Alignment
- **OWASP Top 10**: Addresses CSRF and cross-origin attacks
- **NIST Cybersecurity Framework**: Implements Identify, Protect, Detect
- **SOC 2 Type II**: Provides security controls and logging
- **ISO 27001**: Security management system alignment

### Audit Requirements
- **Security Logs**: Retain structured security event logs
- **Configuration Audits**: Regular security config validation
- **Access Reviews**: Periodic API key and domain reviews
- **Incident Documentation**: Complete security incident records

---

## ✅ CRITICAL DEPLOYMENT CHECKLIST

Before production deployment, verify:

- [ ] No wildcard CORS origins configured
- [ ] Production domains added to allowlist
- [ ] CSRF protection enabled with secure secret
- [ ] Authentication enabled with strong API keys
- [ ] Rate limiting configured appropriately
- [ ] Security logging enabled and monitored
- [ ] All security tests passing
- [ ] Security configurations validated per environment

**Failure to implement these security requirements may result in:**
- Data breaches through CORS attacks
- Unauthorized actions via CSRF attacks  
- Command injection vulnerabilities
- Compliance violations
- Security audit failures

🔒 **Security is not optional - it's mandatory for all VideoCraft deployments.**