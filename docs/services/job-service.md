# Job Service

The Job Service orchestrates the complete video generation workflow with comprehensive validation.

## Service Interface

```go
type JobService interface {
    CreateJob(config *models.VideoConfigArray) (*models.Job, error)
    GetJob(id string) (*models.Job, error)
    ListJobs() ([]*models.Job, error)
    CancelJob(id string) error
    UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error
    UpdateJobProgress(id string, progress int) error
    ProcessJob(ctx context.Context, job *models.Job) error
}
```

## CreateJob Workflow

The `CreateJob` method performs comprehensive validation before job creation:

1. **Configuration Validation**: Validates the video configuration structure
2. **Subtitle Settings Validation**: Validates JSON subtitle settings for each project using `SubtitleService.ValidateJSONSubtitleSettings()`
3. **Job Creation**: Creates job with unique ID and queues for processing
4. **Queue Management**: Handles job queue capacity and worker distribution

### Validation Process
```go
// Validate configuration structure
if err := config.Validate(); err != nil {
    return nil, errors.InvalidInput(err.Error())
}

// Validate JSON subtitle settings for each project
for _, project := range *config {
    if err := js.subtitle.ValidateJSONSubtitleSettings(project); err != nil {
        return nil, errors.InvalidInput(fmt.Sprintf("subtitle validation failed: %v", err))
    }
}
```

## Responsibilities

- **Pre-Processing Validation**: Ensures all configurations are valid before job creation
- **Video Generation Orchestration**: Coordinates all services for video production
- **Job Lifecycle Management**: Handles job creation, status tracking, and completion
- **Error Handling and Recovery**: Provides detailed error messages and graceful failure handling
- **Resource Management**: Manages worker pools and queue capacity
- **Progress Reporting**: Real-time job progress updates