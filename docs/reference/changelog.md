# VideoCraft Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2024-12-16

### 🚨 BREAKING CHANGES

#### Legacy API Endpoints Removed (Epic #64)
- **Complete Legacy Removal**: All 9 legacy endpoints permanently removed
- **Versioned API Only**: Only `/api/v1/` prefixed endpoints available
- **Migration Required**: All clients must update to versioned endpoints

**Removed Legacy Endpoints:**
- `POST /generate-video` → Use `POST /api/v1/generate-video`
- `GET /videos` → Use `GET /api/v1/videos`
- `GET /download/:video_id` → Use `GET /api/v1/download/:video_id`
- `GET /status/:video_id` → Use `GET /api/v1/status/:video_id`
- `DELETE /videos/:video_id` → Use `DELETE /api/v1/videos/:video_id`
- `GET /jobs` → Use `GET /api/v1/jobs`
- `GET /jobs/:job_id` → Use `GET /api/v1/jobs/:job_id`
- `GET /jobs/:job_id/status` → Use `GET /api/v1/jobs/:job_id/status`
- `POST /jobs/:job_id/cancel` → Use `POST /api/v1/jobs/:job_id/cancel`

**Migration Resources:**
- **Migration Guide**: [`docs/migration/legacy-to-v1.md`](../migration/legacy-to-v1.md)
- **Breaking Changes**: [`docs/api/breaking-changes-v2.md`](../api/breaking-changes-v2.md)
- **Code Examples**: Python, JavaScript, Go, and cURL examples provided

### 🔒 Security Improvements
- **Reduced Attack Surface**: 9 fewer endpoints to secure and monitor
- **Simplified Security Model**: Consistent protection across all endpoints
- **Enhanced Security Audit**: Cleaner security review surface

### 🏗️ Architecture Improvements
- **Cleaner Codebase**: Simplified routing and endpoint management
- **Better Maintainability**: Single API version to maintain
- **Improved Documentation**: Focused on unified endpoint set
- **Enhanced Testing**: Reduced test surface area

### 📊 Performance Benefits
- **Optimized Routing**: Fewer routes to evaluate per request
- **Simplified Middleware**: No duplicate middleware application
- **Improved Monitoring**: Cleaner metrics and logging

**Migration Effort**: 🟢 Low (URL prefix changes only)  
**Functional Impact**: ✅ Zero (identical request/response formats)  
**Security Impact**: ✅ Improved (reduced attack surface)

---

## [1.2.0] - 2025-06-16

### <� New Features

#### JSON SubtitleSettings Integration (Epic #26)
- **Per-Request Subtitle Customization**: JSON-based subtitle settings override global configuration
- **Enhanced Progressive Subtitles**: Improved word-level timing precision
- **Dynamic Styling**: Runtime subtitle appearance customization
- **Backward Compatibility**: Existing configurations continue to work unchanged

**Implementation Details:**
- Added JSON SubtitleSettings parsing and validation
- Implemented fallback hierarchy: Request � Global � Defaults
- Enhanced subtitle service with dynamic configuration
- Added comprehensive test coverage for all scenarios

**Impact**: Enables per-video subtitle customization while maintaining system defaults

### =' Improvements
- Enhanced error messages for subtitle configuration validation
- Improved documentation for progressive subtitle system
- Optimized subtitle rendering performance

### =� Documentation
- Updated API documentation with JSON SubtitleSettings examples
- Added comprehensive subtitle service documentation
- Improved progressive subtitles technical guide

### >� Testing
- Added 15+ test scenarios for JSON SubtitleSettings
- Enhanced integration test coverage
- Added edge case validation tests

---

## [1.1.0] - 2025-06-16

### = Security Enhancements

#### Comprehensive Input Validation (Issue #14)
- **Request Validation**: Enhanced JSON schema validation for all API endpoints
- **URL Security**: Comprehensive URL validation with scheme and domain filtering
- **File Path Security**: Robust path traversal prevention with allowlist approach
- **Error Sanitization**: Secure error handling prevents information disclosure

#### Secure Error Handling (Green Phase)
- **Sanitized Responses**: Generic error messages for external clients
- **Internal Logging**: Detailed error information for debugging
- **Security Event Tracking**: Comprehensive audit trail for security violations
- **Structured Error Codes**: Machine-readable error classification

### =� Technical Improvements
- Enhanced validation middleware with detailed error context
- Improved logging structure for security monitoring
- Optimized error handling performance

### =� Breaking Changes
- Error response format standardized across all endpoints
- Enhanced validation may reject previously accepted malformed requests

---

## [1.0.3] - 2025-06-16

### = Bug Fixes
- **Command Injection Protection**: Corrected regex for prohibited characters in FFmpeg command validation
- **Security Hardening**: Enhanced pattern detection for malicious command sequences

### = Security Improvements
- Strengthened command injection prevention
- Improved input sanitization for FFmpeg parameters
- Enhanced security testing coverage

---

## [1.0.2] - 2025-06-16

### = Bug Fixes
- **API Key Configuration**: Updated security settings to properly handle API key configuration
- **Environment Variables**: Fixed environment variable binding for security configuration

### � Configuration Improvements
- Improved configuration validation for security settings
- Enhanced error messages for missing security configuration

---

## [1.0.1] - 2025-06-16

### = Critical Security Update

#### Secure CORS Configuration (Issue #17)
- **=� BREAKING**: Removed wildcard CORS origins (`AllowOrigins: ["*"]`)
- **Domain Allowlisting**: Implemented strict domain-based access control
- **Performance Optimization**: Added origin validation caching
- **Security Logging**: Comprehensive CORS violation audit trail

#### CSRF Protection Implementation
- **Token-Based Validation**: Cryptographic CSRF tokens for state-changing requests
- **Enhanced Security**: HMAC-based token generation and validation
- **Safe Method Exemption**: GET/HEAD/OPTIONS bypass CSRF validation
- **Developer-Friendly**: Clear error messages and token retrieval endpoint

### =� Breaking Changes
**Required Migration Actions:**
```bash
# 1. Configure allowed domains (REQUIRED)
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,api.yourdomain.com"

# 2. Enable CSRF protection (recommended)
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="your-secure-secret"

# 3. Include CSRF tokens in API requests
curl -X POST -H "X-CSRF-Token: your-token" /api/v1/generate-video
```

### =� Security Enhancements
- Multi-layer CORS security with origin validation
- Suspicious pattern detection for malicious origins
- Rate limiting improvements
- Enhanced security headers

---

## [1.0.0] - 2025-06-15

### <� Initial Release

#### Core Features
- **Scene-Based Video Composition**: Multi-scene architecture with individual timing
- **Progressive Subtitle System**: Word-level timing with Whisper AI integration
- **Audio Synchronization**: Automatic timing calculation based on audio duration
- **RESTful API**: Comprehensive HTTP API for video generation
- **Job Management**: Asynchronous video processing with status tracking

#### Security Features
- **Multi-Layer Security Architecture**: Defense-in-depth security implementation
- **Command Injection Prevention**: FFmpeg parameter sanitization and validation
- **Input Validation**: Comprehensive request validation with security focus
- **Error Handling**: Secure error messages without information disclosure

#### Architecture
- **Clean Architecture**: Layered architecture with dependency injection
- **Go-Python Integration**: Efficient Whisper daemon communication
- **Docker Support**: Complete containerization with Docker Compose
- **Configuration Management**: Flexible configuration with environment variables

### = Critical Security Fixes

#### Path Traversal Prevention (Issue #11)
- **Directory Traversal Protection**: Comprehensive path sanitization
- **Allowlist Approach**: Restricted file access to designated directories
- **Security Validation**: Enhanced path validation with security logging

#### SSRF Prevention (Issue #10)
- **URL Validation**: Strict URL scheme and domain validation
- **Private Network Protection**: Blocked access to private IP ranges
- **Download Security**: Secure file download with size and type validation

#### SSL Certificate Verification (Issue #9)
- **Certificate Validation**: Enforced SSL certificate verification
- **TLS Security**: Enhanced TLS configuration for external connections
- **Security Headers**: Comprehensive security header implementation

### <� Technical Implementation

#### Video Generation Pipeline
1. **Configuration Validation**: JSON schema validation with security checks
2. **Audio Analysis**: Duration extraction and metadata analysis
3. **Transcription Processing**: Whisper AI integration for speech recognition
4. **Subtitle Generation**: Progressive subtitle creation with timing
5. **Video Encoding**: FFmpeg-based video generation with scene composition

#### Progressive Subtitles Innovation
- **Zero-Gap Timing**: Eliminates traditional subtitle timing gaps
- **Word-Level Precision**: Character-by-character reveal synchronization
- **ASS Format Generation**: Rich styling with fonts, colors, and effects
- **Real Duration Analysis**: Precise audio duration calculation for timing

#### Service Architecture
- **Job Service**: Asynchronous job management and processing
- **Audio Service**: Audio file analysis and metadata extraction
- **Transcription Service**: Whisper AI daemon integration
- **Subtitle Service**: Progressive subtitle generation and styling
- **Storage Service**: Secure file management with cleanup policies
- **FFmpeg Service**: Video encoding with security validation

### =� Performance Features
- **Concurrent Processing**: Multi-worker job processing
- **Daemon Architecture**: Long-running Whisper process for efficiency
- **Caching Strategy**: Intelligent caching for transcription results
- **Resource Management**: Configurable resource limits and cleanup

### =� Monitoring and Observability
- **Health Checks**: Comprehensive health monitoring endpoints
- **Metrics**: Performance and usage metrics collection
- **Logging**: Structured logging with security event tracking
- **Error Tracking**: Detailed error reporting and analysis

### =' Configuration and Deployment
- **Environment-Based Configuration**: Flexible configuration management
- **Docker Support**: Complete containerization with multi-stage builds
- **Development Tools**: Comprehensive development and testing tools
- **CI/CD Integration**: Automated testing and deployment pipelines

### =� Documentation
- **Comprehensive Documentation**: Complete API and implementation guides
- **Security Guidelines**: Detailed security implementation documentation
- **Development Guide**: Step-by-step development setup instructions
- **Deployment Guide**: Production deployment best practices

---

## Version History Summary

| Version | Release Date | Type | Key Features |
|---------|--------------|------|--------------|
| **1.2.0** | 2025-06-16 | <� Feature | JSON SubtitleSettings Integration |
| **1.1.0** | 2025-06-16 | = Security | Input Validation & Error Handling |
| **1.0.3** | 2025-06-16 | = Bugfix | Command Injection Fix |
| **1.0.2** | 2025-06-16 | = Bugfix | API Key Configuration |
| **1.0.1** | 2025-06-16 | = Security | CORS & CSRF Protection |
| **1.0.0** | 2025-06-15 | <� Initial | Core Platform Release |

## Migration Guides

### From 1.1.x to 1.2.x
No breaking changes. JSON SubtitleSettings are additive and optional.

### From 1.0.x to 1.1.x
Enhanced validation may reject malformed requests. Review error handling.

### From 0.x to 1.0.x
**Critical Security Migration Required:**
- Configure `VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS`
- Enable CSRF protection for production
- Update API clients to include CSRF tokens

## Support

- **Documentation**: [Complete Documentation](../README.md)
- **Issues**: [GitHub Issues](https://github.com/activadee/videocraft/issues)
- **Security**: Report security issues to security@activadee.com
- **API Reference**: [API Documentation](../api/overview.md)

## Links

- [Repository](https://github.com/activadee/videocraft)
- [Documentation](../README.md)
- [Security Policy](../security/overview.md)
- [Contributing Guide](../development/contributing.md)