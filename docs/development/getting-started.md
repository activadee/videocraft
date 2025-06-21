# Getting Started with VideoCraft Development

This guide will help you set up VideoCraft for local development, including prerequisites, installation, and your first development workflow.

## =€ Quick Start

### Prerequisites

**Required Software:**
- **Go 1.23.0+** - [Download Go](https://golang.org/dl/)
- **Python 3.8+** - Required for Whisper AI transcription
- **FFmpeg** - For video processing
- **Git** - For version control

**System Requirements:**
- **Memory**: 4GB RAM minimum (8GB+ recommended)
- **Storage**: 2GB free space for dependencies
- **OS**: Linux, macOS, or Windows

### One-Command Setup

```bash
# Clone and setup development environment
git clone https://github.com/activadee/videocraft.git
cd videocraft
make dev-setup
```

This command will:
- Install development tools (linters, security scanners)
- Download Go dependencies
- Install Python requirements for Whisper
- Create required directories
- Verify the installation

## =Ë Detailed Setup Guide

### 1. System Dependencies

#### Install Go
```bash
# Check if Go is installed
go version

# If not installed, download from https://golang.org/dl/
# Or use a package manager:

# macOS
brew install go

# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Windows
# Download installer from https://golang.org/dl/
```

#### Install Python & pip
```bash
# Check Python version
python3 --version

# Install pip if not available
# macOS
brew install python

# Ubuntu/Debian
sudo apt install python3 python3-pip

# Windows
# Download from https://www.python.org/downloads/
```

#### Install FFmpeg
```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update
sudo apt install ffmpeg

# Windows
# Download from https://ffmpeg.org/download.html
# Or use chocolatey: choco install ffmpeg

# Verify installation
ffmpeg -version
```

### 2. Project Setup

#### Clone Repository
```bash
git clone https://github.com/activadee/videocraft.git
cd videocraft
```

#### Install Go Dependencies
```bash
# Download and verify modules
go mod download
go mod verify

# Install development tools
make install-tools
```

#### Install Python Dependencies
```bash
# Install Whisper AI dependencies
pip3 install -r scripts/requirements.txt

# Verify installation
python3 -c "import whisper; print('Whisper installed successfully')"
```

#### Create Required Directories
```bash
# Create working directories
mkdir -p generated_videos temp whisper_cache

# Verify directory structure
ls -la
```

### 3. Configuration

#### Environment Setup
```bash
# Copy example configuration
cp config/config.example.yaml config.yaml

# Set environment variables
export VIDEOCRAFT_SERVER_HOST="localhost"
export VIDEOCRAFT_SERVER_PORT="3002"
export VIDEOCRAFT_FFMPEG_BINARY_PATH="ffmpeg"
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_PATH="python3"

# For development (security relaxed)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="false"
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
```

#### Configuration File (config.yaml)
```yaml
server:
  host: "localhost"
  port: 3002

ffmpeg:
  binary_path: "ffmpeg"
  timeout: "30m"
  quality: 28  # Faster encoding for development
  preset: "fast"

transcription:
  enabled: true
  python:
    path: "python3"
    script_path: "./scripts"
    model: "tiny"  # Fastest model for development
    device: "cpu"

subtitles:
  enabled: true
  style: "progressive"
  font_family: "Arial"
  font_size: 24

storage:
  output_dir: "./generated_videos"
  temp_dir: "./temp"
  retention_days: 1  # Clean up quickly in development

security:
  enable_auth: false  # Disabled for development
  enable_csrf: false  # Disabled for development
  rate_limit: 50

log:
  level: "debug"
  format: "text"
```

### 4. Build and Run

#### Build Application
```bash
# Build binary
make build

# Or build for multiple platforms
make build-all

# Check binary
./videocraft --version
```

#### Run Development Server
```bash
# Option 1: Run with auto-reload (recommended)
make dev

# Option 2: Run directly
make run

# Option 3: Run with custom config
go run cmd/server/main.go --config=config/development.yaml

# Check if server is running
curl http://localhost:3002/health
```

#### Run Tests
```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run integration tests
make test-integration

# Run benchmarks
make benchmark
```

## =' Development Workflow

### Daily Development Commands

```bash
# Start development server with auto-reload
make dev

# Run tests continuously
make test

# Format code before committing
make fmt

# Run quality checks
make quality-check

# Build and test everything
make build test lint
```

### Code Quality Tools

#### Formatting
```bash
# Format all Go files
make fmt

# Check imports
goimports -w .
```

#### Linting
```bash
# Run all linters
make lint

# Individual linters
golangci-lint run
staticcheck ./...
go vet ./...
```

#### Security Scanning
```bash
# Run security scan
make security

# Individual security tools
govulncheck ./...
gosec ./...
```

### Testing Strategies

#### Unit Tests
```bash
# Run unit tests
go test ./internal/...

# Test specific package
go test ./internal/services/

# Test with verbose output
go test -v ./internal/services/subtitle_service_test.go
```

#### Integration Tests
```bash
# Run integration tests
go test -tags=integration ./tests/integration/

# Test specific integration
go test -tags=integration ./tests/integration/api_test.go
```

#### Manual Testing
```bash
# Test API endpoints
curl http://localhost:3002/health

# Generate test video
curl -X POST http://localhost:3002/api/v1/generate-video \
  -H "Content-Type: application/json" \
  -d '{
    "scenes": [
      {
        "id": "test",
        "elements": [
          {
            "type": "audio",
            "src": "https://example.com/test.mp3"
          }
        ]
      }
    ]
  }'
```

## =3 Docker Development

### Using Docker Compose
```bash
# Start development environment
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f videocraft

# Stop environment
docker-compose down
```

### Building Docker Image
```bash
# Build image
make docker-build

# Run container
make docker-run

# Or manually
docker build -t videocraft:dev .
docker run -p 3002:3002 videocraft:dev
```

## =à IDE Setup

### VS Code Configuration

#### Recommended Extensions
- **Go** - Rich Go language support
- **Docker** - Docker integration
- **REST Client** - API testing
- **GitLens** - Git integration
- **Prettier** - Code formatting

#### Settings (.vscode/settings.json)
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.vetFlags": ["-composites=false"],
  "go.buildOnSave": "package",
  "go.testOnSave": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### Launch Configuration (.vscode/launch.json)
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch VideoCache",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/server/main.go",
      "env": {
        "VIDEOCRAFT_LOG_LEVEL": "debug",
        "VIDEOCRAFT_SECURITY_ENABLE_AUTH": "false"
      },
      "args": ["--config", "config/development.yaml"]
    }
  ]
}
```

### GoLand/IntelliJ Configuration

#### Go Settings
- Enable Go modules support
- Set GOROOT and GOPATH correctly
- Configure code style to use gofmt

#### Run Configurations
- **Build**: `go build cmd/server/main.go`
- **Run**: `go run cmd/server/main.go`
- **Test**: `go test ./...`

## =Ý Environment Variables Reference

### Core Settings
```bash
# Server Configuration
VIDEOCRAFT_SERVER_HOST="localhost"
VIDEOCRAFT_SERVER_PORT="3002"

# FFmpeg Configuration  
VIDEOCRAFT_FFMPEG_BINARY_PATH="ffmpeg"
VIDEOCRAFT_FFMPEG_TIMEOUT="30m"
VIDEOCRAFT_FFMPEG_QUALITY="28"
VIDEOCRAFT_FFMPEG_PRESET="fast"

# Transcription Configuration
VIDEOCRAFT_TRANSCRIPTION_ENABLED="true"
VIDEOCRAFT_TRANSCRIPTION_PYTHON_PATH="python3"
VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="tiny"
VIDEOCRAFT_TRANSCRIPTION_DAEMON_ENABLED="true"

# Storage Configuration
VIDEOCRAFT_STORAGE_OUTPUT_DIR="./generated_videos"
VIDEOCRAFT_STORAGE_TEMP_DIR="./temp"
VIDEOCRAFT_STORAGE_RETENTION_DAYS="1"

# Security Configuration (Development)
VIDEOCRAFT_SECURITY_ENABLE_AUTH="false"
VIDEOCRAFT_SECURITY_ENABLE_CSRF="false"
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
VIDEOCRAFT_SECURITY_RATE_LIMIT="50"

# Logging
VIDEOCRAFT_LOG_LEVEL="debug"
VIDEOCRAFT_LOG_FORMAT="text"
```

### Production Settings
```bash
# Security (Production)
VIDEOCRAFT_SECURITY_ENABLE_AUTH="true"
VIDEOCRAFT_SECURITY_API_KEY="your-secure-api-key"
VIDEOCRAFT_SECURITY_ENABLE_CSRF="true"
VIDEOCRAFT_SECURITY_CSRF_SECRET="your-csrf-secret"
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="yourdomain.com,api.yourdomain.com"
VIDEOCRAFT_SECURITY_RATE_LIMIT="500"

# Performance (Production)
VIDEOCRAFT_JOB_WORKERS="8"
VIDEOCRAFT_JOB_MAX_CONCURRENT="20"
VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="base"
VIDEOCRAFT_FFMPEG_QUALITY="23"
VIDEOCRAFT_FFMPEG_PRESET="medium"
```

## =¨ Troubleshooting

### Common Issues

#### Go Build Errors
```bash
# Update Go modules
go mod tidy
go mod download

# Clear module cache
go clean -modcache

# Verify Go installation
go env GOROOT GOPATH
```

#### Python/Whisper Issues
```bash
# Reinstall requirements
pip3 install -r scripts/requirements.txt --force-reinstall

# Check Python path
which python3
python3 --version

# Test Whisper import
python3 -c "import whisper; print('OK')"
```

#### FFmpeg Issues
```bash
# Check FFmpeg installation
ffmpeg -version
which ffmpeg

# Test FFmpeg functionality
ffmpeg -f lavfi -i testsrc=duration=1:size=320x240:rate=1 test.mp4
```

#### Port Already in Use
```bash
# Find process using port 3002
lsof -i :3002

# Kill process
kill -9 <PID>

# Or use different port
export VIDEOCRAFT_SERVER_PORT="3003"
```

#### Permission Issues
```bash
# Fix directory permissions
chmod 755 generated_videos temp whisper_cache

# Fix binary permissions
chmod +x videocraft
```

### Debug Mode
```bash
# Run with debug logging
export VIDEOCRAFT_LOG_LEVEL="debug"
make run

# Enable Go race detector
go run -race cmd/server/main.go

# Enable verbose testing
go test -v -race ./...
```

### Performance Debugging
```bash
# CPU profiling
go build -o videocraft cmd/server/main.go
./videocraft -cpuprofile=cpu.prof

# Memory profiling
go tool pprof http://localhost:3002/debug/pprof/heap

# Trace analysis
go tool trace trace.out
```

## =Ú Next Steps

### Development Resources
- [Development Guidelines](guidelines.md) - Code standards and practices
- [API Documentation](../api/overview.md) - API reference and examples
- [Architecture Guide](../architecture/overview.md) - System design overview
- [Security Guide](../security/overview.md) - Security implementation details

### Contributing
- [Contributing Guidelines](contributing.md) - How to contribute to VideoCraft
- [Code Review Process](guidelines.md#code-review) - Code review standards
- [Testing Strategy](../development/testing.md) - Testing best practices

### Advanced Topics
- [Docker Deployment](../deployment/docker.md) - Container deployment
- [Production Setup](../deployment/production.md) - Production deployment
- [Performance Tuning](../troubleshooting/performance.md) - Optimization guide

## <¯ Quick Development Tasks

### Create Your First Feature
```bash
# 1. Create feature branch
git checkout -b feature/my-new-feature

# 2. Make changes
# ... edit code ...

# 3. Test changes
make test
make lint

# 4. Build and run
make build
make run

# 5. Manual testing
curl http://localhost:3002/health

# 6. Commit changes
git add .
git commit -m "feat: add my new feature"
```

### Debug a Test Failure
```bash
# 1. Run specific test
go test -v ./internal/services/subtitle_service_test.go

# 2. Run with race detector
go test -race ./internal/services/

# 3. Debug with prints
go test -v ./internal/services/ -run TestSpecificFunction

# 4. Run integration test
go test -tags=integration ./tests/integration/
```

This getting started guide provides everything needed to begin developing with VideoCraft. Follow the setup steps, and you'll be ready to contribute to the project!