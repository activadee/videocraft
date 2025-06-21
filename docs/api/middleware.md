# HTTP Middleware Documentation

This document provides comprehensive documentation for VideoCraft's HTTP middleware layer, focusing on security, logging, and request processing middleware components.

## =' Middleware Architecture

VideoCraft uses a layered middleware architecture built on Gin framework:

```go
router.Use(middleware.Logger(logger))
router.Use(middleware.Recovery())
router.Use(middleware.SecureCORS(cfg, logger))
router.Use(middleware.CSRFProtection(cfg, logger))
router.Use(middleware.RequestID())
router.Use(middleware.RateLimit(rateLimiter))
```

## =� Security Middleware

### CORS Protection (`middleware/cors.go`)

**Critical Security Features:**
- Zero Wildcard Policy (eliminates `AllowOrigins: ["*"]` vulnerability)
- Strict Domain Allowlisting
- Suspicious Pattern Detection
- Performance Optimization with Caching
- Comprehensive Security Logging

#### Configuration

```go
func SecureCORS(cfg *config.Config, log logger.Logger) gin.HandlerFunc {
    corsConfig := cors.Config{
        AllowOrigins: prepareAllowedOrigins(cfg.Security.AllowedDomains),
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{
            "Origin", "Content-Type", "Authorization", 
            "X-Requested-With", "X-CSRF-Token",
        },
        AllowCredentials: len(cfg.Security.AllowedDomains) == 1,
        AllowOriginFunc: func(origin string) bool {
            return validateOriginWithCache(origin, cfg.Security.AllowedDomains, cache, log)
        },
    }
    return cors.New(corsConfig)
}
```

#### Origin Validation Process

```go
func validateOrigin(origin string, allowedDomains []string, log logger.Logger) bool {
    // 1. Empty origin allowed (same-origin requests)
    if origin == "" {
        return true
    }
    
    // 2. Check against allowed domains
    for _, allowedDomain := range allowedDomains {
        if isExactDomainMatch(origin, allowedDomain) {
            log.Debug("CORS origin validation: allowed")
            return true
        }
    }
    
    // 3. Check for suspicious patterns
    if containsSuspiciousPatterns(origin) {
        log.Error("CORS_SECURITY_VIOLATION: Suspicious origin pattern detected")
        return false
    }
    
    // 4. Log security violation
    log.Warn("CORS_SECURITY_VIOLATION: Origin not in allowlist")
    return false
}
```

#### Suspicious Pattern Detection

```go
func containsSuspiciousPatterns(origin string) bool {
    suspiciousPatterns := []string{
        "javascript:", "data:", "file:", "ftp:",
        "localhost", "127.0.0.1", "0.0.0.0",
        "//", "\\", "..", "@",
        "<script", "</script>", "eval(",
        "%3cscript", "%3c/script%3e",
    }
    
    originLower := strings.ToLower(origin)
    for _, pattern := range suspiciousPatterns {
        if strings.Contains(originLower, pattern) {
            return true
        }
    }
    return false
}
```

### CSRF Protection (`middleware/csrf.go`)

**Key Security Features:**
- Token-Based Validation for State-Changing Requests
- Enhanced Token Format Validation
- Cryptographic HMAC Verification
- Safe Method Exemption

#### CSRF Token Generation

```go
func GenerateCSRFToken(secret string) (string, error) {
    // Generate random bytes
    randomBytes := make([]byte, csrfTokenLength)
    if _, err := rand.Read(randomBytes); err != nil {
        return "", err
    }
    
    // Create HMAC: random_bytes + HMAC(secret, random_bytes)
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(randomBytes)
    signature := mac.Sum(nil)
    
    // Combine: hex(random_bytes) + hex(signature)
    token := hex.EncodeToString(randomBytes) + hex.EncodeToString(signature)
    return token, nil
}
```

#### CSRF Token Validation

```go
func CSRFProtection(cfg *config.Config, log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip for safe methods
        if isSafeMethod(c.Request.Method) {
            c.Next()
            return
        }
        
        // Get and validate token
        token := c.GetHeader("X-CSRF-Token")
        if token == "" {
            log.Warn("CSRF_SECURITY_VIOLATION: Missing CSRF token")
            c.JSON(http.StatusForbidden, gin.H{
                "error": "CSRF token required",
                "code":  "CSRF_TOKEN_MISSING",
            })
            c.Abort()
            return
        }
        
        // Validate token format
        if !isValidTokenFormat(token) {
            log.Error("CSRF_SECURITY_VIOLATION: Malformed CSRF token")
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Invalid CSRF token format",
                "code":  "CSRF_TOKEN_MALFORMED",
            })
            c.Abort()
            return
        }
        
        // Cryptographic validation
        if !isValidCSRFToken(token, cfg.Security.CSRFSecret) {
            log.Error("CSRF_SECURITY_VIOLATION: Invalid CSRF token")
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Invalid CSRF token",
                "code":  "CSRF_TOKEN_INVALID",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

#### Safe Method Detection

```go
func isSafeMethod(method string) bool {
    switch strings.ToUpper(method) {
    case "GET", "HEAD", "OPTIONS":
        return true
    default:
        return false
    }
}
```

### Authentication Middleware (`middleware/auth.go`)

#### Bearer Token Authentication

```go
func Auth(cfg *config.Config, log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Skip auth if disabled (development mode)
        if !cfg.Security.EnableAuth {
            c.Next()
            return
        }
        
        // Get authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            log.Warn("Authentication failed: missing authorization header")
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header required",
                "code":  "MISSING_AUTH_HEADER",
            })
            c.Abort()
            return
        }
        
        // Parse bearer token
        const bearerPrefix = "Bearer "
        if !strings.HasPrefix(authHeader, bearerPrefix) {
            log.Warn("Authentication failed: invalid format")
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid authorization format",
                "code":  "INVALID_AUTH_FORMAT",
            })
            c.Abort()
            return
        }
        
        token := strings.TrimPrefix(authHeader, bearerPrefix)
        
        // Constant-time comparison to prevent timing attacks
        if !secureCompare(token, cfg.Security.APIKey) {
            log.Warn("Authentication failed: invalid API key")
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid API key",
                "code":  "INVALID_API_KEY",
            })
            c.Abort()
            return
        }
        
        log.Debug("Authentication successful")
        c.Next()
    }
}
```

#### Secure Token Comparison

```go
func secureCompare(a, b string) bool {
    // Prevent timing attacks with constant-time comparison
    if len(a) != len(b) {
        return false
    }
    
    var result byte
    for i := 0; i < len(a); i++ {
        result |= a[i] ^ b[i]
    }
    
    return result == 0
}
```

## � Rate Limiting Middleware

### User-Based Rate Limiting with Token Bucket Algorithm

VideoCraft implements enhanced rate limiting with user-based limits using API keys from Bearer tokens, with IP fallback for unauthenticated requests.

```go
type rateLimiter struct {
    visitors map[string]*visitor
    mu       sync.RWMutex
    rate     int
    cleanup  *time.Ticker
}

type visitor struct {
    limiter  *tokenBucket
    lastSeen time.Time
}

type tokenBucket struct {
    tokens   int
    capacity int
    refill   time.Time
    mu       sync.Mutex
}
```

### Enhanced Rate Limiting Features

**Key Features:**
- **User-Based Limits**: Per-user limits using API keys from Bearer tokens
- **IP Fallback**: Uses client IP for unauthenticated requests
- **Health Endpoint Bypass**: System monitoring endpoints skip rate limiting
- **Security Logging**: Rate limit violations logged with hashed user IDs
- **Professional HTTP Responses**: Standard X-RateLimit headers and structured JSON

### Rate Limiting Implementation

```go
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
    rl := &rateLimiter{
        visitors: make(map[string]*visitor),
        rate:     requestsPerMinute,
        cleanup:  time.NewTicker(time.Minute),
    }

    // Start cleanup goroutine
    go rl.cleanupVisitors()

    return func(c *gin.Context) {
        // Skip rate limiting for health endpoints
        if isHealthEndpoint(c.Request.URL.Path) {
            c.Next()
            return
        }

        // Get user identifier (API key or IP fallback)
        userID := getUserIdentifier(c)

        allowed, remaining := rl.allow(userID)

        if !allowed {
            // Add standard rate limit headers
            c.Header("X-RateLimit-Limit", strconv.Itoa(rl.rate))
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("Retry-After", "60")

            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "code":  "RATE_LIMIT_EXCEEDED",
                "details": gin.H{
                    "limit":       rl.rate,
                    "window":      "1 minute",
                    "retry_after": 60,
                },
            })

            // Log rate limit violation (hash user ID for security)
            logrus.WithFields(logrus.Fields{
                "user_id":  hashUserIDForLogging(userID),
                "endpoint": c.Request.URL.Path,
                "method":   c.Request.Method,
                "ip":       c.ClientIP(),
            }).Warn("Rate limit exceeded")

            c.Abort()
            return
        }

        // Add rate limit headers for successful requests
        c.Header("X-RateLimit-Limit", strconv.Itoa(rl.rate))
        c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

        c.Next()
    }
}
```

### User Identification Logic

```go
// getUserIdentifier extracts user identifier from the request
func getUserIdentifier(c *gin.Context) string {
    // Try to get API key from Authorization header
    authHeader := c.GetHeader("Authorization")
    if authHeader != "" {
        if strings.HasPrefix(authHeader, "Bearer ") {
            apiKey := strings.TrimPrefix(authHeader, "Bearer ")
            if apiKey != "" {
                return apiKey // Use API key as user identifier
            }
        }
    }

    // Fallback to client IP for unauthenticated requests
    return c.ClientIP()
}
```

### Security-Compliant Logging

```go
// hashUserIDForLogging creates a safe hash of the user ID for logging
// This prevents API keys from being logged in plaintext while maintaining traceability
func hashUserIDForLogging(userID string) string {
    // For IP addresses, log them directly as they're not sensitive
    if strings.Contains(userID, ".") || strings.Contains(userID, ":") {
        return userID
    }
    
    // For API keys, create a SHA-256 hash with a short prefix for identification
    h := sha256.Sum256([]byte(userID))
    hash := hex.EncodeToString(h[:])
    
    // Return first 8 characters of hash with a prefix for easier identification
    return "hash:" + hash[:8]
}
```

### Health Endpoint Bypass

```go
// isHealthEndpoint checks if the request path is a health monitoring endpoint
func isHealthEndpoint(path string) bool {
    healthEndpoints := []string{"/health", "/ready", "/live", "/metrics"}
    for _, endpoint := range healthEndpoints {
        if path == endpoint {
            return true
        }
    }
    return false
}
```

## =� Logging Middleware

### Structured Request Logging

```go
func Logger(logger logger.Logger) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        // Create structured log entry
        fields := map[string]interface{}{
            "timestamp":   param.TimeStamp,
            "method":      param.Method,
            "path":        param.Path,
            "status":      param.StatusCode,
            "latency":     param.Latency,
            "client_ip":   param.ClientIP,
            "user_agent":  param.Request.UserAgent(),
            "request_id":  param.Request.Header.Get("X-Request-ID"),
            "body_size":   param.BodySize,
        }
        
        // Add error info if present
        if param.ErrorMessage != "" {
            fields["error"] = param.ErrorMessage
        }
        
        // Log based on status code
        if param.StatusCode >= 500 {
            logger.WithFields(fields).Error("Request completed with server error")
        } else if param.StatusCode >= 400 {
            logger.WithFields(fields).Warn("Request completed with client error")
        } else {
            logger.WithFields(fields).Info("Request completed successfully")
        }
        
        return ""
    })
}
```

### Security Event Logging

```go
type SecurityLogger struct {
    logger logger.Logger
}

func NewSecurityLogger(logger logger.Logger) *SecurityLogger {
    return &SecurityLogger{logger: logger}
}

func (sl *SecurityLogger) LogSecurityViolation(violationType, threatLevel string, context map[string]interface{}) {
    fields := map[string]interface{}{
        "event_type":     "SECURITY_VIOLATION",
        "violation_type": violationType,
        "threat_level":   threatLevel,
        "timestamp":      time.Now(),
    }
    
    // Merge context fields
    for k, v := range context {
        fields[k] = v
    }
    
    switch threatLevel {
    case "HIGH":
        sl.logger.WithFields(fields).Error("High-severity security violation detected")
    case "MEDIUM":
        sl.logger.WithFields(fields).Warn("Medium-severity security violation detected")
    case "LOW":
        sl.logger.WithFields(fields).Info("Low-severity security event")
    default:
        sl.logger.WithFields(fields).Warn("Security violation detected")
    }
}
```

## = Recovery Middleware

### Panic Recovery with Logging

```go
func Recovery(logger logger.Logger) gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        var err error
        
        switch x := recovered.(type) {
        case string:
            err = errors.New(x)
        case error:
            err = x
        default:
            err = errors.New("unknown panic")
        }
        
        // Log the panic with stack trace
        fields := map[string]interface{}{
            "error":      err.Error(),
            "method":     c.Request.Method,
            "path":       c.Request.URL.Path,
            "client_ip":  c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "request_id": c.GetHeader("X-Request-ID"),
        }
        
        // Include stack trace in debug mode
        if gin.Mode() == gin.DebugMode {
            fields["stack_trace"] = string(debug.Stack())
        }
        
        logger.WithFields(fields).Error("Panic recovered")
        
        // Return error response
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":      "Internal server error",
            "code":       "INTERNAL_ERROR",
            "request_id": c.GetHeader("X-Request-ID"),
        })
        
        c.Abort()
    })
}
```

## <� Request ID Middleware

### Request Tracing

```go
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if request ID already exists
        requestID := c.GetHeader("X-Request-ID")
        
        // Generate new ID if not present
        if requestID == "" {
            requestID = generateRequestID()
        }
        
        // Set request ID in context and response header
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        c.Next()
    }
}

func generateRequestID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return fmt.Sprintf("%x", b)
}
```

## =' Custom Middleware Development

### Error Handling Middleware

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Process any errors that occurred during request handling
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            // Log the error
            fields := map[string]interface{}{
                "error":      err.Error(),
                "method":     c.Request.Method,
                "path":       c.Request.URL.Path,
                "request_id": c.GetHeader("X-Request-ID"),
            }
            
            // Determine error type and response
            var statusCode int
            var errorCode string
            
            switch err.Type {
            case gin.ErrorTypeBind:
                statusCode = http.StatusBadRequest
                errorCode = "INVALID_REQUEST"
            case gin.ErrorTypePublic:
                statusCode = http.StatusBadRequest
                errorCode = "BAD_REQUEST"
            default:
                statusCode = http.StatusInternalServerError
                errorCode = "INTERNAL_ERROR"
            }
            
            // Don't override if response already written
            if !c.Writer.Written() {
                c.JSON(statusCode, gin.H{
                    "error":      err.Error(),
                    "code":       errorCode,
                    "request_id": c.GetHeader("X-Request-ID"),
                })
            }
        }
    }
}
```

### Custom Security Headers

```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Security headers
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        // Remove server information
        c.Header("Server", "")
        
        c.Next()
    }
}
```

## =� Middleware Metrics

### Performance Monitoring

```go
func MetricsMiddleware(metrics *Metrics) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        // Record metrics
        duration := time.Since(start)
        
        metrics.RequestDuration.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
            fmt.Sprintf("%d", c.Writer.Status()),
        ).Observe(duration.Seconds())
        
        metrics.RequestTotal.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
            fmt.Sprintf("%d", c.Writer.Status()),
        ).Inc()
        
        if c.Writer.Status() >= 400 {
            metrics.ErrorTotal.WithLabelValues(
                c.Request.Method,
                c.Request.URL.Path,
                fmt.Sprintf("%d", c.Writer.Status()),
            ).Inc()
        }
    }
}
```

## >� Middleware Testing

### Unit Testing Example

```go
func TestCSRFMiddleware(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    cfg := &config.Config{
        Security: config.SecurityConfig{
            EnableCSRF:  true,
            CSRFSecret: "test-secret-32-characters-long",
        },
    }
    
    router.Use(middleware.CSRFProtection(cfg, logger.NewNoop()))
    router.POST("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"success": true})
    })
    
    t.Run("Missing CSRF token", func(t *testing.T) {
        req, _ := http.NewRequest(http.MethodPost, "/test", nil)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusForbidden, w.Code)
        
        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.Equal(t, "CSRF_TOKEN_MISSING", response["code"])
    })
    
    t.Run("Valid CSRF token", func(t *testing.T) {
        token, _ := middleware.GenerateCSRFToken(cfg.Security.CSRFSecret)
        
        req, _ := http.NewRequest(http.MethodPost, "/test", nil)
        req.Header.Set("X-CSRF-Token", token)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
    })
}
```

## =' Middleware Configuration

### Environment-Specific Middleware

```go
func SetupMiddleware(router *gin.Engine, cfg *config.Config, logger logger.Logger) {
    // Global middleware (always applied)
    router.Use(middleware.RequestID())
    router.Use(middleware.Logger(logger))
    router.Use(middleware.Recovery(logger))
    router.Use(middleware.SecurityHeaders())
    
    // Security middleware (configurable)
    router.Use(middleware.SecureCORS(cfg, logger))
    
    if cfg.Security.EnableCSRF {
        router.Use(middleware.CSRFProtection(cfg, logger))
    }
    
    if cfg.Security.EnableAuth {
        // Apply to protected routes only
        protected := router.Group("/api/v1")
        protected.Use(middleware.Auth(cfg, logger))
    }
    
    // Rate limiting (configurable)
    if cfg.Security.RateLimit > 0 {
        router.Use(middleware.RateLimit(cfg.Security.RateLimit))
    }
    
    // Development-only middleware
    if gin.Mode() == gin.DebugMode {
        router.Use(middleware.DebugHeaders())
    }
}
```

## =� Security Best Practices

### Middleware Security Checklist

- [ ] **CORS configured with strict domain allowlisting** (no wildcards)
- [ ] **CSRF protection enabled** for state-changing requests
- [ ] **Authentication required** for protected endpoints
- [ ] **Rate limiting implemented** per user (API key) with IP fallback
- [ ] **Security headers set** (X-Content-Type-Options, etc.)
- [ ] **Request logging enabled** for audit trail
- [ ] **Error handling** prevents information disclosure
- [ ] **Recovery middleware** handles panics gracefully

### Common Security Pitfalls

**L Avoid These Mistakes:**

1. **Wildcard CORS origins**: `AllowOrigins: ["*"]`
2. **Disabled CSRF in production**: `EnableCSRF: false`
3. **Weak API keys**: Short or predictable keys
4. **Missing rate limiting**: No request throttling
5. **Information disclosure**: Detailed error messages in production
6. **Missing security headers**: No XSS/CSRF protection headers

** Best Practices:**

1. **Strict CORS allowlisting**: Only specific domains
2. **Mandatory CSRF tokens**: For all state-changing requests
3. **Strong authentication**: Cryptographically secure keys
4. **Comprehensive rate limiting**: Per-client and global limits
5. **Sanitized error responses**: No sensitive information
6. **Security headers**: Complete protection suite

This middleware documentation ensures proper implementation of security, logging, and request processing middleware in VideoCraft's HTTP layer.