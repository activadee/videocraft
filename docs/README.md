# VideoCraft Documentation Hub

Welcome to the VideoCraft documentation center. This documentation is organized by topic for easy navigation and discovery.

## 🚀 Quick Start Paths

### For Developers Getting Started
1. [Architecture Overview](architecture/overview.md) - Understand the system design
2. [Development Guidelines](development/guidelines.md) - Setup and contribution workflow  
3. [Configuration](configuration/overview.md) - Environment setup
4. [API Reference](api/overview.md) - Integration guide

### For Security Teams
1. [Security Overview](security/overview.md) - Comprehensive security architecture
2. [CORS & CSRF Protection](security/cors-csrf.md) - HTTP security layer
3. [Vulnerability Management](security/vulnerability-management.md) - Security monitoring

### For Operations Teams
1. [Deployment Guide](deployment/docker.md) - Docker and production setup
2. [Configuration Management](configuration/overview.md) - Environment variables
3. [Troubleshooting](troubleshooting/overview.md) - Common issues and debugging

## 📚 Documentation Sections

### 🏗️ Architecture & Design
- [Architecture Overview](architecture/overview.md) - System design patterns and principles
- [Core Components](architecture/components.md) - HTTP, Service, and Domain layers
- [Data Flow & Processing](architecture/data-flow.md) - Video generation workflow
- [Progressive Subtitles System](architecture/progressive-subtitles.md) - Innovative timing solution

### 🔒 Security & Compliance
- [Security Overview](security/overview.md) - Multi-layered security architecture
- [CORS & CSRF Protection](security/cors-csrf.md) - HTTP security implementation
- [Command Injection Prevention](security/ffmpeg-security.md) - FFmpeg security validation
- [Error Handling Security](security/secure-error-handling.md) - Information disclosure prevention
- [Vulnerability Management](security/vulnerability-management.md) - Security monitoring
- [Best Practices](security/best-practices.md) - Security guidelines

### 🌐 API Reference
- [API Overview](api/overview.md) - RESTful API design and endpoints
- [Authentication & Authorization](api/authentication.md) - Security implementation
- [Endpoints Reference](api/endpoints.md) - Complete API documentation
- [Error Responses](api/error-responses.md) - Error handling and codes

### ⚙️ Services & Components
- [Service Layer Overview](services/overview.md) - Service architecture
- [Job Service](services/job-service.md) - Video generation orchestration
- [Audio Service](services/audio-service.md) - Audio analysis and timing
- [Transcription Service](services/transcription-service.md) - Python-Go integration
- [Subtitle Service](services/subtitle-service.md) - ASS subtitle generation
- [FFmpeg Service](services/ffmpeg-service.md) - Video encoding integration

### 🔧 Configuration & Setup
- [Configuration Overview](configuration/overview.md) - Complete configuration guide
- [Environment Variables](configuration/environment-variables.md) - All environment settings
- [Security Configuration](configuration/security.md) - Security-specific settings

### 👨‍💻 Development & Contributing
- [Development Guidelines](development/guidelines.md) - Code standards and workflow
- [Testing Strategy](development/testing.md) - Unit, integration, and security testing
- [Performance Optimization](development/performance.md) - Concurrent processing and memory management
- [Debugging Guide](development/debugging.md) - Logging and troubleshooting

### 🚀 Deployment & Operations
- [Docker Deployment](deployment/docker.md) - Container setup and orchestration
- [Production Setup](deployment/production.md) - Scaling and monitoring
- [CI/CD Pipeline](deployment/ci-cd.md) - GitHub Actions workflow

### 🔍 Troubleshooting & Support
- [Common Issues](troubleshooting/overview.md) - Frequent problems and solutions
- [Debugging Tools](troubleshooting/debugging.md) - Logging and monitoring
- [Performance Issues](troubleshooting/performance.md) - Optimization troubleshooting

## 🎯 Key Features Documentation

### Progressive Subtitles Innovation
VideoCraft's most innovative feature solves traditional subtitle timing gaps:
- [Progressive Subtitles System](architecture/progressive-subtitles.md) - Core innovation explained
- [Real Duration Timing](services/audio-service.md#real-duration-analysis) - Technical implementation
- [JSON SubtitleSettings (v1.1+)](services/subtitle-service.md#json-subtitle-settings) - Per-request customization

### Security-First Architecture
Comprehensive protection against modern threats:
- [Multi-Layer Security](security/overview.md#multi-layered-architecture) - Defense in depth
- [Command Injection Prevention](security/ffmpeg-security.md) - FFmpeg protection
- [40+ Security Patterns](security/secure-error-handling.md#security-patterns) - Pattern detection

### Python-Go Integration
Efficient Whisper AI integration:
- [Whisper Daemon Architecture](services/transcription-service.md#daemon-architecture) - Long-running process
- [Go-Python Communication](services/transcription-service.md#communication) - stdin/stdout protocol
- [Lifecycle Management](services/transcription-service.md#lifecycle) - Automatic startup/shutdown

## 🔗 External References

- [GitHub Repository](https://github.com/activadee/videocraft)
- [API Documentation](api/overview.md)
- [Security Best Practices](security/best-practices.md)
- [Contributing Guidelines](development/guidelines.md)

---

**Last Updated**: Documentation restructured for topic-based organization (Issue #71)  
**Version**: v1.2.0+  
**Architecture**: Clean Architecture with Security-First Design