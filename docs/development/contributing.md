# Contributing to VideoCraft

We welcome contributions to VideoCraft! This guide will help you understand our development process, coding standards, and how to submit your contributions effectively.

## =€ Quick Start for Contributors

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/videocraft.git
cd videocraft

# Add upstream remote
git remote add upstream https://github.com/activadee/videocraft.git
```

### 2. Set Up Development Environment
```bash
# Install dependencies and tools
make dev-setup

# Verify setup
make test
make lint
```

### 3. Create Feature Branch
```bash
# Create and switch to feature branch
git checkout -b feature/your-feature-name

# Keep your branch updated
git fetch upstream
git rebase upstream/main
```

## =Ë Development Process

### Workflow Overview

1. **Issue Discussion** - Discuss feature/bug in GitHub Issues
2. **Branch Creation** - Create feature branch from main
3. **Development** - Implement changes following our guidelines
4. **Testing** - Write tests and ensure all tests pass
5. **Code Review** - Submit PR and address review feedback
6. **Merge** - Maintainer merges after approval

### Branch Naming Convention

```bash
# Feature branches
feature/add-progressive-subtitles
feature/improve-error-handling

# Bug fix branches
fix/cors-configuration-issue
fix/memory-leak-in-transcription

# Documentation branches
docs/update-api-documentation
docs/add-security-guide

# Refactoring branches
refactor/service-layer-cleanup
refactor/improve-test-coverage
```

## =» Coding Standards

### Go Code Standards

#### Package Organization
```go
// Good: Clear package structure
package services

import (
    "context"
    "fmt"
    
    "github.com/activadee/videocraft/internal/domain/models"
    "github.com/activadee/videocraft/pkg/logger"
)

// Service interface defined in consuming package
type VideoService interface {
    GenerateVideo(ctx context.Context, config *models.VideoConfig) (*models.Job, error)
}
```

#### Error Handling
```go
// Good: Wrapped errors with context
func (vs *VideoService) GenerateVideo(ctx context.Context, config *models.VideoConfig) (*models.Job, error) {
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid video configuration: %w", err)
    }
    
    job, err := vs.jobService.CreateJob(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create video generation job: %w", err)
    }
    
    return job, nil
}

// Bad: Generic errors without context
func (vs *VideoService) GenerateVideo(ctx context.Context, config *models.VideoConfig) (*models.Job, error) {
    if err := config.Validate(); err != nil {
        return nil, err
    }
    // ... no error wrapping
}
```

#### Struct Design
```go
// Good: Composed structs with clear dependencies
type VideoService struct {
    cfg           *config.Config
    logger        logger.Logger
    jobService    JobService
    storageService StorageService
    ffmpegService FFmpegService
}

func NewVideoService(
    cfg *config.Config,
    logger logger.Logger,
    jobService JobService,
    storageService StorageService,
    ffmpegService FFmpegService,
) *VideoService {
    return &VideoService{
        cfg:            cfg,
        logger:         logger,
        jobService:     jobService,
        storageService: storageService,
        ffmpegService:  ffmpegService,
    }
}
```

#### Interface Design
```go
// Good: Focused, single-responsibility interfaces
type JobService interface {
    CreateJob(ctx context.Context, config *models.VideoConfig) (*models.Job, error)
    GetJob(ctx context.Context, id string) (*models.Job, error)
    UpdateJobProgress(ctx context.Context, id string, progress int) error
}

// Good: Composable interfaces
type JobManager interface {
    JobService
    JobScheduler
    JobMonitor
}

// Bad: Large, multi-responsibility interfaces
type VideoProcessingService interface {
    CreateJob(ctx context.Context, config *models.VideoConfig) (*models.Job, error)
    ProcessVideo(ctx context.Context, job *models.Job) error
    TranscribeAudio(ctx context.Context, audioURL string) (*models.TranscriptionResult, error)
    GenerateSubtitles(ctx context.Context, transcription *models.TranscriptionResult) error
    EncodeVideo(ctx context.Context, scenes []models.Scene) error
    // ... too many responsibilities
}
```

### Code Formatting

#### Use Go Tools
```bash
# Format code
go fmt ./...
goimports -w .

# Or use make target
make fmt
```

#### Naming Conventions
```go
// Good: Clear, descriptive names
type VideoGenerationService struct {}
func (vgs *VideoGenerationService) CreateVideoFromScenes(scenes []Scene) error {}

var ErrInvalidVideoConfiguration = errors.New("invalid video configuration")

const (
    MaxVideoSizeMB        = 1024
    DefaultFrameRate      = 30
    ProgressiveSubtitles  = "progressive"
)

// Bad: Unclear abbreviations
type VGS struct {}
func (v *VGS) CreateVid(s []Scene) error {}
var ErrInvConf = errors.New("inv conf")
```

### Testing Standards

#### Test Structure
```go
func TestVideoService_GenerateVideo(t *testing.T) {
    // Arrange
    cfg := &config.Config{/* test config */}
    logger := logger.NewNoop()
    mockJobService := &mocks.MockJobService{}
    
    service := NewVideoService(cfg, logger, mockJobService, nil, nil)
    
    testConfig := &models.VideoConfig{
        Scenes: []models.Scene{
            {ID: "test-scene", Elements: []models.Element{/* elements */}},
        },
    }
    
    // Act
    job, err := service.GenerateVideo(context.Background(), testConfig)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, job)
    assert.Equal(t, models.JobStatusPending, job.Status)
    
    // Verify mock interactions
    mockJobService.AssertCalled(t, "CreateJob", mock.Anything, testConfig)
}
```

#### Test Coverage
```bash
# Aim for 80%+ coverage
make coverage

# View coverage report
go tool cover -html=coverage.out
```

#### Test Types
```go
// Unit tests - isolated component testing
func TestSubtitleService_GenerateProgressiveSubtitles(t *testing.T) {
    // Test individual service methods
}

// Integration tests - component interaction testing
func TestAPI_VideoGeneration_EndToEnd(t *testing.T) {
    // Test full workflow through API
}

// Benchmark tests - performance testing
func BenchmarkVideoService_GenerateVideo(b *testing.B) {
    // Performance benchmarking
}
```

## = Security Guidelines

### Security Requirements

#### Input Validation
```go
// Good: Comprehensive validation
func (vs *VideoService) ValidateVideoConfig(config *models.VideoConfig) error {
    if config == nil {
        return errors.New("video configuration is required")
    }
    
    // Validate URL schemes
    for _, scene := range config.Scenes {
        for _, element := range scene.Elements {
            if element.Src != "" {
                if err := validateURL(element.Src); err != nil {
                    return fmt.Errorf("invalid element source URL: %w", err)
                }
            }
        }
    }
    
    return nil
}

func validateURL(urlStr string) error {
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return err
    }
    
    // Only allow specific schemes
    allowedSchemes := map[string]bool{
        "http": true, "https": true,
    }
    
    if !allowedSchemes[parsed.Scheme] {
        return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
    }
    
    return nil
}
```

#### Secure Error Handling
```go
// Good: Sanitized error messages
func (vs *VideoService) ProcessVideo(ctx context.Context, job *models.Job) error {
    if err := vs.validateJob(job); err != nil {
        // Log detailed error internally
        vs.logger.WithError(err).Error("Job validation failed")
        
        // Return generic error to client
        return errors.New("invalid job configuration")
    }
    
    return nil
}

// Bad: Exposing internal details
func (vs *VideoService) ProcessVideo(ctx context.Context, job *models.Job) error {
    if err := vs.database.GetJob(job.ID); err != nil {
        return fmt.Errorf("database error: %s, table: jobs, query: SELECT * FROM jobs WHERE id = %s", err, job.ID)
    }
    return nil
}
```

#### CORS and Security Headers
```go
// Follow existing security patterns
func setupSecureMiddleware(router *gin.Engine, cfg *config.Config) {
    // Use existing secure CORS implementation
    router.Use(middleware.SecureCORS(cfg, logger))
    
    // No wildcard origins allowed
    // Domain allowlisting required
    // CSRF protection for state-changing requests
}
```

## =Ý Documentation Standards

### Code Documentation
```go
// Good: Clear, comprehensive documentation
// VideoService provides video generation capabilities with scene-based composition,
// progressive subtitles, and automated audio synchronization.
//
// The service orchestrates multiple components:
//   - Audio transcription via Whisper AI
//   - Progressive subtitle generation with word-level timing
//   - Video encoding via FFmpeg with scene transitions
//   - Job management for asynchronous processing
type VideoService struct {
    // ...
}

// GenerateVideo creates a new video generation job from the provided configuration.
// It validates the configuration, creates timing segments based on audio duration,
// and initiates asynchronous video processing.
//
// Parameters:
//   - ctx: Request context for cancellation and timeouts
//   - config: Video configuration with scenes, elements, and settings
//
// Returns:
//   - *models.Job: Created job with pending status
//   - error: Validation or creation errors
//
// The method performs the following steps:
//   1. Validates video configuration and scene elements
//   2. Analyzes audio files for duration and metadata
//   3. Creates timing segments for scene synchronization
//   4. Initializes job with pending status
//   5. Queues job for background processing
func (vs *VideoService) GenerateVideo(ctx context.Context, config *models.VideoConfig) (*models.Job, error) {
    // Implementation...
}
```

### README and Documentation
- Update relevant documentation when adding features
- Include usage examples for new APIs
- Document breaking changes in CHANGELOG.md
- Add configuration options to documentation

### API Documentation
```go
// Document API endpoints with detailed examples
// POST /api/v1/generate-video
//
// Creates a new video generation job.
//
// Request Body:
//   {
//     "scenes": [
//       {
//         "id": "intro",
//         "elements": [
//           {
//             "type": "audio",
//             "src": "https://example.com/intro.mp3"
//           }
//         ]
//       }
//     ]
//   }
//
// Response (202 Accepted):
//   {
//     "job_id": "uuid",
//     "status": "pending",
//     "status_url": "/jobs/{job_id}/status"
//   }
func (vh *VideoHandler) GenerateVideo(c *gin.Context) {
    // Implementation...
}
```

##  Quality Assurance

### Pre-Commit Checklist

Before submitting a pull request, ensure:

- [ ] **Code compiles without warnings**
  ```bash
  make build
  ```

- [ ] **All tests pass**
  ```bash
  make test
  make test-integration
  ```

- [ ] **Code is properly formatted**
  ```bash
  make fmt
  ```

- [ ] **Linting passes**
  ```bash
  make lint
  ```

- [ ] **Security scan passes**
  ```bash
  make security
  ```

- [ ] **Documentation is updated**
  - Code comments added/updated
  - README.md updated if needed
  - API documentation updated

- [ ] **Tests added for new functionality**
  - Unit tests for business logic
  - Integration tests for workflows
  - Edge cases covered

### Quality Tools

#### Automated Quality Checks
```bash
# Run all quality checks
make quality-check

# Individual checks
make fmt      # Format code
make vet      # Go vet analysis
make lint     # Golangci-lint
make security # Security scan
make test     # All tests
```

#### Continuous Integration
The CI pipeline automatically runs:
- Code compilation
- Unit and integration tests
- Code coverage analysis
- Security vulnerability scanning
- Docker image building

## = Pull Request Process

### Creating a Pull Request

1. **Ensure branch is up to date**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run quality checks**
   ```bash
   make quality-check
   ```

3. **Commit with conventional commits**
   ```bash
   git commit -m "feat: add progressive subtitle generation"
   git commit -m "fix: resolve CORS configuration issue"
   git commit -m "docs: update API documentation"
   ```

4. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create pull request on GitHub**

### Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Screenshots (if applicable)

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Code is commented, particularly hard-to-understand areas
- [ ] Documentation updated
- [ ] Tests added that prove fix is effective or feature works
- [ ] No new warnings introduced
```

### Review Process

1. **Automated Checks** - CI pipeline runs automatically
2. **Code Review** - Maintainer reviews code quality and design
3. **Testing** - Reviewer tests functionality manually if needed
4. **Approval** - Maintainer approves if changes meet standards
5. **Merge** - Maintainer merges using squash merge

### Review Criteria

Reviewers check for:
- **Correctness**: Does the code do what it's supposed to do?
- **Security**: Are there any security vulnerabilities?
- **Performance**: Are there any performance implications?
- **Maintainability**: Is the code readable and maintainable?
- **Testing**: Are tests adequate and do they pass?
- **Documentation**: Is documentation updated appropriately?

## = Bug Reports

### Reporting Bugs

1. **Search existing issues** first
2. **Use bug report template**
3. **Provide reproduction steps**
4. **Include environment details**
5. **Add relevant logs/screenshots**

### Bug Report Template

```markdown
**Bug Description**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected Behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
 - OS: [e.g. macOS, Ubuntu 20.04]
 - Go Version: [e.g. 1.21.5]
 - VideoCraft Version: [e.g. v1.2.0]
 - Docker Version: [e.g. 20.10.17] (if using Docker)

**Additional Context**
Add any other context about the problem here.
```

## =¡ Feature Requests

### Proposing Features

1. **Search existing feature requests**
2. **Create detailed feature request**
3. **Discuss implementation approach**
4. **Wait for maintainer approval**
5. **Implement after approval**

### Feature Request Template

```markdown
**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is.

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions.

**Additional context**
Add any other context or screenshots about the feature request.

**Implementation Notes**
Any technical details about how this might be implemented.
```

## =€ Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Workflow

1. **Feature Development** - Features developed on feature branches
2. **Release Branch** - Create release branch from main
3. **Release Candidate** - Test release candidate thoroughly
4. **Release** - Tag and publish release
5. **Post-Release** - Merge back to main and update documentation

## > Community Guidelines

### Code of Conduct

- **Be Respectful**: Treat everyone with respect and kindness
- **Be Collaborative**: Work together constructively
- **Be Inclusive**: Welcome people of all backgrounds
- **Be Patient**: Help others learn and grow
- **Be Professional**: Maintain professional communication

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code reviews and implementation discussions

### Getting Help

- Check existing documentation first
- Search GitHub Issues for similar problems
- Create new issue with detailed information
- Tag maintainers if urgent

## =Ú Resources

### Development Resources
- [Getting Started Guide](getting-started.md)
- [Development Guidelines](guidelines.md)
- [API Documentation](../api/overview.md)
- [Architecture Overview](../architecture/overview.md)

### External Resources
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Conventional Commits](https://www.conventionalcommits.org/)

### Tools
- [golangci-lint](https://golangci-lint.run/) - Linting
- [gosec](https://github.com/securego/gosec) - Security scanning
- [govulncheck](https://golang.org/x/vuln/cmd/govulncheck) - Vulnerability checking
- [testify](https://github.com/stretchr/testify) - Testing framework

Thank you for contributing to VideoCraft! Your contributions help make video generation more accessible and powerful for everyone.