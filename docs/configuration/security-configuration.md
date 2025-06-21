# Security Configuration Guide

This document provides comprehensive security configuration guidance for VideoCraft, focusing on production-ready security settings and best practices.

## = Security Configuration Overview

VideoCraft implements a multi-layered security architecture with configurable security controls:

```go
type SecurityConfig struct {
    APIKey         string   `mapstructure:"api_key"`
    RateLimit      int      `mapstructure:"rate_limit"`
    EnableAuth     bool     `mapstructure:"enable_auth"`
    AllowedDomains []string `mapstructure:"allowed_domains"`
    EnableCSRF     bool     `mapstructure:"enable_csrf"`
    CSRFSecret     string   `mapstructure:"csrf_secret"`
}
```

## =á Core Security Settings

### Authentication Configuration

```yaml
security:
  enable_auth: true
  api_key: "${VIDEOCRAFT_SECURITY_API_KEY}"
```

**Environment Variable:**
```bash
export VIDEOCRAFT_SECURITY_API_KEY="your-256-bit-hex-key-here"
```

**Key Generation:**
```bash
# Generate cryptographically secure API key
openssl rand -hex 32
```

**Production Requirements:**
-  **REQUIRED**: `enable_auth: true`
-  **REQUIRED**: Strong API key (256-bit minimum)
- L **NEVER**: Commit API keys to version control
-  **RECOMMENDED**: Rotate API keys regularly

### CORS Security Configuration

```yaml
security:
  # CRITICAL: Domain allowlisting (NO WILDCARDS)
  allowed_domains:
    - "yourdomain.com"
    - "api.yourdomain.com"
    - "admin.yourdomain.com"
```

**Environment Variable:**
```bash
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,api.yourdomain.com"
```

**Security Features:**
- =« **Zero Wildcard Policy**: No `*` origins permitted
-  **Strict Domain Allowlisting**: Only configured domains
- = **Suspicious Pattern Detection**: Blocks malicious origins
- =Ê **Security Violation Logging**: Comprehensive audit trail

### CSRF Protection

```yaml
security:
  enable_csrf: true
  csrf_secret: "${VIDEOCRAFT_SECURITY_CSRF_SECRET}"
```

**Environment Variable:**
```bash
export VIDEOCRAFT_SECURITY_CSRF_SECRET="your-256-bit-csrf-secret"
```

**CSRF Secret Generation:**
```bash
# Generate CSRF secret (64 characters minimum)
openssl rand -hex 32
```

**Protection Scope:**
-  **Protected**: POST, PUT, DELETE, PATCH
- ª **Exempt**: GET, HEAD, OPTIONS
- <¯ **Token Validation**: Cryptographic HMAC verification

### Rate Limiting

```yaml
security:
  rate_limit: 100  # requests per minute per client IP
```

**Rate Limiting Features:**
- < **Per-Client IP**: Individual rate limits
- ñ **Token Bucket Algorithm**: Burst handling
- =Ê **Configurable Limits**: Adjustable per environment
- =« **Automatic Blocking**: Exceeding limits results in 429

## <í Production Security Configuration

### Complete Production Config

```yaml
security:
  # Authentication (REQUIRED in production)
  enable_auth: true
  api_key: "${VIDEOCRAFT_SECURITY_API_KEY}"
  
  # CORS Security (CRITICAL)
  allowed_domains:
    - "yourdomain.com"
    - "app.yourdomain.com"
    - "api.yourdomain.com"
  
  # CSRF Protection (RECOMMENDED)
  enable_csrf: true
  csrf_secret: "${VIDEOCRAFT_SECURITY_CSRF_SECRET}"
  
  # Rate Limiting (REQUIRED)
  rate_limit: 500  # Higher for production load
```

### Environment Variables Template

```bash
# Security Configuration
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="true"
export VIDEOCRAFT_SECURITY_API_KEY="$(openssl rand -hex 32)"
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,app.yourdomain.com"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF="true"
export VIDEOCRAFT_SECURITY_CSRF_SECRET="$(openssl rand -hex 32)"
export VIDEOCRAFT_SECURITY_RATE_LIMIT="500"
```

## =' Development Security Configuration

### Development Config

```yaml
security:
  # Relaxed for development (NOT for production)
  enable_auth: false
  
  # Local development domains
  allowed_domains:
    - "localhost:3000"
    - "127.0.0.1:3000"
    - "dev.local"
  
  # CSRF disabled for easier development
  enable_csrf: false
  
  # Lower rate limits for testing
  rate_limit: 50
```

### Development Environment

```bash
# Development Security (DEVELOPMENT ONLY)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="false"
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF="false"
export VIDEOCRAFT_SECURITY_RATE_LIMIT="50"
```

## =¨ Security Validation

### Configuration Validation

VideoCraft automatically validates security configuration at startup:

```go
func (c *SecurityConfig) Validate() error {
    // Production authentication requirement
    if c.EnableAuth && c.APIKey == "" {
        return errors.New("API key required when authentication is enabled")
    }
    
    // CSRF secret requirement
    if c.EnableCSRF && len(c.CSRFSecret) < 32 {
        return errors.New("CSRF secret must be at least 32 characters")
    }
    
    // Domain validation
    for _, domain := range c.AllowedDomains {
        if strings.Contains(domain, "*") {
            return errors.New("wildcard domains not permitted for security")
        }
    }
    
    return nil
}
```

### Security Checklist

**Pre-Production Security Checklist:**

- [ ] **Authentication enabled** (`enable_auth: true`)
- [ ] **Strong API key** (256-bit minimum)
- [ ] **Domain allowlisting configured** (no wildcards)
- [ ] **CSRF protection enabled** (`enable_csrf: true`)
- [ ] **CSRF secret configured** (256-bit minimum)
- [ ] **Rate limiting configured** (appropriate for load)
- [ ] **Secrets via environment variables** (not in config files)
- [ ] **HTTPS enforced** (reverse proxy configuration)
- [ ] **Security logging enabled** (audit trail)

## = Advanced Security Features

### API Key Auto-Generation

VideoCraft automatically generates secure API keys when authentication is enabled but no key is provided:

```go
// Auto-generate API key if authentication is enabled but no key is provided
if config.Security.EnableAuth && config.Security.APIKey == "" && !viper.IsSet("security.api_key") {
    generatedKey, err := generateSecureAPIKey()
    if err != nil {
        return nil, fmt.Errorf("failed to generate API key: %w", err)
    }
    config.Security.APIKey = generatedKey
}
```

### CSRF Secret Auto-Generation

Similar auto-generation for CSRF secrets:

```go
// Auto-generate strong CSRF secret if CSRF is enabled but none supplied
if config.Security.EnableCSRF &&
    config.Security.CSRFSecret == "" &&
    !viper.IsSet("security.csrf_secret") {
    secret, err := generateSecureAPIKey() // 256-bit hex == 64 chars
    if err != nil {
        return nil, fmt.Errorf("failed to generate CSRF secret: %w", err)
    }
    config.Security.CSRFSecret = secret
}
```

### Security Logging Configuration

```yaml
log:
  level: "info"  # Security events at INFO level and above
  format: "json"  # Structured logging for security analysis
```

**Security Log Fields:**
- `violation_type`: Type of security violation
- `threat_level`: HIGH, MEDIUM, LOW
- `client_ip`: Source IP address
- `origin`: Request origin
- `user_agent`: Client user agent
- `timestamp`: Event timestamp

## < Network Security

### HTTPS Configuration

VideoCraft should always run behind HTTPS in production. Configure your reverse proxy:

**Nginx Example:**
```nginx
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    location / {
        proxy_pass http://127.0.0.1:3002;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Firewall Configuration

**Recommended Firewall Rules:**
```bash
# Allow HTTPS traffic
ufw allow 443/tcp

# Allow SSH (restrict to management IPs)
ufw allow from MANAGEMENT_IP to any port 22

# Block direct access to application port
ufw deny 3002/tcp

# Enable firewall
ufw enable
```

## =Ê Security Monitoring

### Security Metrics

Monitor these security-related metrics:

- **Authentication failures per minute**
- **CORS violations per minute** 
- **CSRF token validation failures**
- **Rate limit exceeded events**
- **Suspicious origin detections**

### Log Analysis

**Security Event Queries:**
```bash
# CORS violations
grep "CORS_SECURITY_VIOLATION" logs/app.log | jq .

# Authentication failures
grep "INVALID_API_KEY" logs/app.log | jq .

# CSRF violations
grep "CSRF_SECURITY_VIOLATION" logs/app.log | jq .

# Suspicious patterns
grep "SUSPICIOUS_ORIGIN" logs/app.log | jq .
```

##   Security Warnings

### Critical Security Warnings

**=« NEVER DO THIS:**
```yaml
security:
  allowed_domains: ["*"]  # CRITICAL VULNERABILITY
```

**=« NEVER DO THIS:**
```yaml
security:
  enable_auth: false      # ONLY for development
  api_key: "weak-key"     # Use strong keys only
  csrf_secret: "short"    # Must be 32+ characters
```

### Common Security Misconfigurations

1. **Wildcard Origins**: Using `*` in allowed domains
2. **Weak API Keys**: Short or predictable keys
3. **Disabled Authentication**: In production environments
4. **HTTP in Production**: Not using HTTPS
5. **Hardcoded Secrets**: Secrets in configuration files
6. **Permissive CORS**: Allowing all origins
7. **Disabled CSRF**: Without proper justification

## = Security Maintenance

### Regular Security Tasks

**Monthly:**
- [ ] Review security logs for anomalies
- [ ] Update allowed domains list
- [ ] Check for security updates

**Quarterly:**
- [ ] Rotate API keys
- [ ] Rotate CSRF secrets
- [ ] Security configuration audit
- [ ] Penetration testing

**Annually:**
- [ ] Comprehensive security review
- [ ] Update security documentation
- [ ] Security training for team

### Key Management

**API Key Rotation Process:**
1. Generate new API key
2. Update environment variable
3. Restart application
4. Update client configurations
5. Revoke old key after transition

**Automation Example:**
```bash
#!/bin/bash
# api-key-rotation.sh

# Generate new key
NEW_KEY=$(openssl rand -hex 32)

# Update environment
export VIDEOCRAFT_SECURITY_API_KEY="$NEW_KEY"

# Restart service
systemctl restart videocraft

# Log rotation
echo "$(date): API key rotated" >> /var/log/videocraft-security.log
```

## =Ë Security Configuration Reference

### Complete Security Configuration Schema

```yaml
security:
  # Authentication
  enable_auth: true                    # Enable API key authentication
  api_key: "${API_KEY}"               # API key (256-bit recommended)
  
  # CORS Security
  allowed_domains:                     # Allowed domains (NO wildcards)
    - "yourdomain.com"
    - "api.yourdomain.com"
  
  # CSRF Protection
  enable_csrf: true                    # Enable CSRF protection
  csrf_secret: "${CSRF_SECRET}"       # CSRF secret (256-bit minimum)
  
  # Rate Limiting
  rate_limit: 100                      # Requests per minute per IP
```

### Environment Variables Reference

```bash
# Authentication
VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
VIDEOCRAFT_SECURITY_API_KEY=your-256-bit-api-key

# CORS
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS=domain1.com,domain2.com

# CSRF
VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
VIDEOCRAFT_SECURITY_CSRF_SECRET=your-256-bit-csrf-secret

# Rate Limiting
VIDEOCRAFT_SECURITY_RATE_LIMIT=100
```

This comprehensive security configuration guide ensures VideoCraft is deployed with production-ready security controls and follows security best practices.