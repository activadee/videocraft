# 1.0.0 (2025-06-15)


### Bug Fixes

* 🔒 [CRITICAL] Enable SSL Certificate Verification in Whisper Daemon ([#20](https://github.com/activadee/videocraft/issues/20)) ([7d9354e](https://github.com/activadee/videocraft/commit/7d9354e1b6c84c2e3be9a369af075ad60c2720f0)), closes [#9](https://github.com/activadee/videocraft/issues/9)
* 🔒 [CRITICAL] Fix Path Traversal Vulnerability in Storage Service ([#25](https://github.com/activadee/videocraft/issues/25)) ([21afa87](https://github.com/activadee/videocraft/commit/21afa8780d8005201bed907a7261759ec9b6a65f)), closes [#11](https://github.com/activadee/videocraft/issues/11) [#11](https://github.com/activadee/videocraft/issues/11)
* 🔒 [CRITICAL] Prevent SSRF in Whisper Daemon URL Downloads ([2fa11e4](https://github.com/activadee/videocraft/commit/2fa11e49a9fc826d7de43c94b4311a4d84e3173a)), closes [#10](https://github.com/activadee/videocraft/issues/10)
* update claude-code-action to v0.0.19-oauth and add GITHUB_ACTOR environment variable ([98636ba](https://github.com/activadee/videocraft/commit/98636baab78833b76d6c55f6a733527c1cbdebce))


### Features

* add `/work-on-task` command with detailed TDD workflow and usage instructions ([51828be](https://github.com/activadee/videocraft/commit/51828beb6ae8abe8b3ccb076042d4f3c5c41e500))
* add comprehensive CI/CD workflows and development tooling ([b0b77bf](https://github.com/activadee/videocraft/commit/b0b77bf11c6ac9ff0006bafe2cdb9b10411d7fee))
* enable authentication by default for security ([072e428](https://github.com/activadee/videocraft/commit/072e42877087c928213df0373d0e2c9a05205d01)), closes [#12](https://github.com/activadee/videocraft/issues/12)
* FFmpeg Command Injection Prevention - Security Task 1 ([#1](https://github.com/activadee/videocraft/issues/1)) ([7d693a8](https://github.com/activadee/videocraft/commit/7d693a8afc8dcb4c7d23d7f42482a829b28e9eba))
* implement `/create-tasks` command for structured task generation and TDD compatibility ([fc172a8](https://github.com/activadee/videocraft/commit/fc172a8e09160c86fe6910728410cdb76716516b))
* initial VideoCraft implementation with comprehensive documentation ([14ad1bf](https://github.com/activadee/videocraft/commit/14ad1bf621cdf99a045ce58d4eb4735fcb020e17))
* modernize GitHub Actions workflow with 2025 best practices and fix tests ([#37](https://github.com/activadee/videocraft/issues/37)) ([66a036a](https://github.com/activadee/videocraft/commit/66a036a457b8f838b260e0fc87a3e8e56ab241cd)), closes [#38](https://github.com/activadee/videocraft/issues/38)


### BREAKING CHANGES

* Authentication is now enabled by default to prevent unauthorized access

- Enable auth by default in config.yaml and Go defaults
- Auto-generate secure 256-bit API keys when none provided
- Add comprehensive test coverage for auth middleware
- Add router tests for authentication enforcement
- Create SECURITY.md with migration guide
- Update README.md with authentication instructions

Security improvements:
- All API endpoints now require authentication by default
- Cryptographically secure API key generation
- Proper error responses for unauthorized requests
- Health endpoints remain unprotected for monitoring

Migration guide provided in SECURITY.md for existing deployments.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial VideoCraft implementation with comprehensive documentation
- Progressive subtitles system with word-level timing synchronization
- Python Whisper daemon integration for speech recognition
- FFmpeg-based video processing pipeline
- RESTful API with job management and status tracking
- Comprehensive configuration management with YAML and environment variables
- Docker and Docker Compose support
- GitHub Actions workflows for testing, linting, and releases
- Semantic release automation
- Security scanning and vulnerability checking
- Multi-platform binary builds (Linux, macOS, Windows, ARM64)
- Code quality tools (golangci-lint, staticcheck, gosec)
- Comprehensive documentation across all packages

### Features
- **Video Generation**: Scene-based video composition with timing control
- **Progressive Subtitles**: Character-by-character subtitle reveal synchronized with speech
- **Speech Recognition**: Integration with OpenAI Whisper for accurate transcription
- **Audio Synchronization**: Real audio duration-based scene timing (not just speech duration)
- **ASS Subtitle Generation**: Advanced SubStation Alpha subtitle files with rich styling
- **RESTful API**: HTTP API for video generation and job management
- **Job Queue**: Asynchronous job processing with status tracking
- **Configuration**: Flexible configuration via YAML files and environment variables
- **Logging**: Structured logging with configurable levels and formats
- **Health Checks**: HTTP health check endpoints for monitoring
- **Rate Limiting**: Request rate limiting for API protection
- **Authentication**: Optional API key authentication
- **Error Handling**: Comprehensive error handling and reporting

### Technical Implementation
- **Go**: Backend implemented in Go 1.24+ with clean architecture (CI uses Go 1.24.4)
- **Python Integration**: Python Whisper daemon for AI transcription
- **FFmpeg**: Video processing using FFmpeg with complex filter chains
- **Docker**: Containerized deployment with multi-stage builds
- **CI/CD**: Modern GitHub Actions workflow with 7 parallel jobs (lint, test, integration, security, coverage, benchmark, docker), Go 1.24.4, and 2025 best practices (~50% faster CI)
- **Security**: Comprehensive vulnerability scanning (gosec, govulncheck) and secure defaults
- **Performance**: Concurrent processing and resource optimization
- **Monitoring**: Health checks and metrics collection ready

### Documentation
- Complete README with setup and usage instructions
- Technical architecture documentation (CLAUDE.md)
- API documentation with examples
- Package-specific documentation for all modules
- Configuration reference with all options
- Development setup and contributing guidelines
- Docker deployment instructions

## [0.1.0] - 2024-01-XX

### Added
- Initial release of VideoCraft
- Core video generation functionality
- Progressive subtitles system
- Python Whisper integration
- RESTful API
- Docker support
- Comprehensive documentation
