package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/activadee/videocraft/internal/api/models"
	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/core/media/audio"
	"github.com/activadee/videocraft/internal/core/media/subtitle"
	"github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// Service provides job queue management
type Service interface {
	CreateJob(config *models.VideoConfigArray) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
	ListJobs() ([]*models.Job, error)
	ProcessJob(ctx context.Context, job *models.Job) error
	CancelJob(jobID string) error
	UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error
	UpdateJobProgress(id string, progress int) error
	Start() error
	Stop() error
}

// Forward declaration - these will be injected
type FFmpegService interface {
	GenerateVideo(ctx context.Context, config *models.VideoConfigArray, progressChan chan<- int) (string, error)
	GenerateVideoWithSubtitles(ctx context.Context, config *models.VideoConfigArray, subtitleFilePath string, progressChan chan<- int) (string, error)
}

type SubtitleService interface {
	ValidateJSONSubtitleSettings(project models.VideoProject) error
	GenerateSubtitles(ctx context.Context, project models.VideoProject) (*subtitle.SubtitleResult, error)
	CleanupTempFiles(filePath string) error
}

type StorageService interface {
	StoreVideo(videoPath string) (string, error)
}

// Media service interfaces for URL analysis
type AudioService interface {
	AnalyzeAudio(ctx context.Context, url string) (*audio.AudioInfo, error)
}

type VideoService interface {
	AnalyzeVideo(ctx context.Context, videoURL string) (*models.VideoInfo, error)
}

type ImageService interface {
	ValidateImage(imageURL string) error
}

type service struct {
	cfg *app.Config
	log logger.Logger

	jobs     map[string]*models.Job
	mu       sync.RWMutex
	jobQueue chan *models.Job
	workers  int

	// Service dependencies
	ffmpeg   FFmpegService
	subtitle SubtitleService
	storage  StorageService

	// Media service dependencies
	audio AudioService
	video VideoService
	image ImageService
}

// NewService creates a new job service
func NewService(cfg *app.Config, log logger.Logger, ffmpeg FFmpegService, subtitle SubtitleService, storage StorageService, audio AudioService, video VideoService, image ImageService) Service {
	return &service{
		cfg:      cfg,
		log:      log,
		jobs:     make(map[string]*models.Job),
		jobQueue: make(chan *models.Job, cfg.Job.QueueSize),
		workers:  cfg.Job.Workers,
		ffmpeg:   ffmpeg,
		subtitle: subtitle,
		storage:  storage,
		audio:    audio,
		video:    video,
		image:    image,
	}
}

func (js *service) CreateJob(config *models.VideoConfigArray) (*models.Job, error) {
	js.log.Debug("Creating new job")

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, errors.InvalidInput(err.Error())
	}

	// Validate JSON subtitle settings for each project
	for _, project := range *config {
		if err := js.subtitle.ValidateJSONSubtitleSettings(project); err != nil {
			return nil, errors.InvalidInput(fmt.Sprintf("subtitle validation failed: %v", err))
		}
	}

	job := &models.Job{
		ID:        uuid.New().String(),
		Status:    models.JobStatusPending,
		Config:    *config,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store job
	js.mu.Lock()
	js.jobs[job.ID] = job
	js.mu.Unlock()

	// Queue job for processing
	select {
	case js.jobQueue <- job:
		js.log.Infof("Job created and queued: %s", job.ID)
	default:
		return nil, errors.InternalError(fmt.Errorf("job queue is full"))
	}

	return job, nil
}

func (js *service) GetJob(id string) (*models.Job, error) {
	js.mu.RLock()
	job, exists := js.jobs[id]
	js.mu.RUnlock()

	if !exists {
		return nil, errors.JobNotFound(id)
	}

	// Return a copy to prevent external modifications
	jobCopy := *job
	return &jobCopy, nil
}

func (js *service) ListJobs() ([]*models.Job, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	jobs := make([]*models.Job, 0, len(js.jobs))
	for _, job := range js.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	return jobs, nil
}

func (js *service) CancelJob(id string) error {
	js.mu.Lock()
	job, exists := js.jobs[id]
	if !exists {
		js.mu.Unlock()
		return errors.JobNotFound(id)
	}

	if job.Status == models.JobStatusCompleted || job.Status == models.JobStatusFailed {
		js.mu.Unlock()
		return errors.InvalidInput("cannot cancel completed or failed job")
	}

	job.Status = models.JobStatusCancelled
	job.UpdatedAt = time.Now()
	js.mu.Unlock()

	js.log.Infof("Job cancelled: %s", id)
	return nil
}

func (js *service) UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, exists := js.jobs[id]
	if !exists {
		return errors.JobNotFound(id)
	}

	job.Status = status
	job.UpdatedAt = time.Now()

	if errorMsg != "" {
		job.Error = errorMsg
	}

	if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		now := time.Now()
		job.CompletedAt = &now
	}

	return nil
}

func (js *service) UpdateJobProgress(id string, progress int) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, exists := js.jobs[id]
	if !exists {
		return errors.JobNotFound(id)
	}

	job.Progress = progress
	job.UpdatedAt = time.Now()

	return nil
}

func (js *service) ProcessJob(ctx context.Context, job *models.Job) error {
	js.log.Infof("Processing job: %s", job.ID)

	// Update status to processing
	if err := js.UpdateJobStatus(job.ID, models.JobStatusProcessing, ""); err != nil {
		return err
	}

	// Create progress channel
	progressChan := make(chan int, 10)
	go func() {
		for progress := range progressChan {
			if err := js.UpdateJobProgress(job.ID, progress); err != nil {
				js.log.Errorf("Failed to update job progress: %v", err)
			}
		}
	}()

	// Step 1: Analyze media URLs to get durations using media services
	js.log.Info("Analyzing media URLs for metadata")
	if err := js.analyzeMediaWithServices(ctx, &job.Config); err != nil {
		js.log.Errorf("Media analysis failed: %v", err)
		if updateErr := js.UpdateJobStatus(job.ID, models.JobStatusFailed, fmt.Sprintf("media analysis failed: %v", err)); updateErr != nil {
			js.log.Errorf("Failed to update job status: %v", updateErr)
		}
		return err
	}

	// Step 2: Generate subtitles if needed
	var subtitleFilePath string
	for _, project := range job.Config {
		if js.needsSubtitles(project) {
			js.log.Info("Generating subtitles for project")
			subtitleResult, err := js.subtitle.GenerateSubtitles(ctx, project)
			if err != nil {
				js.log.Errorf("Failed to generate subtitles: %v", err)
				if updateErr := js.UpdateJobStatus(job.ID, models.JobStatusFailed, fmt.Sprintf("subtitle generation failed: %v", err)); updateErr != nil {
					js.log.Errorf("Failed to update job status: %v", updateErr)
				}
				return err
			}
			subtitleFilePath = subtitleResult.FilePath
			js.log.Infof("Subtitles generated: %s (%d events)", subtitleFilePath, subtitleResult.EventCount)
			break // Only generate subtitles for the first project that needs them
		}
	}

	// Process the video generation
	var videoPath string
	var err error
	if subtitleFilePath != "" {
		videoPath, err = js.ffmpeg.GenerateVideoWithSubtitles(ctx, &job.Config, subtitleFilePath, progressChan)
	} else {
		videoPath, err = js.ffmpeg.GenerateVideo(ctx, &job.Config, progressChan)
	}
	// Note: progressChan is closed by the FFmpeg service

	if err != nil {
		if updateErr := js.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error()); updateErr != nil {
			js.log.Errorf("Failed to update job status to failed: %v", updateErr)
		}
		return err
	}

	// Store the generated video
	videoID, err := js.storage.StoreVideo(videoPath)
	if err != nil {
		if updateErr := js.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error()); updateErr != nil {
			js.log.Errorf("Failed to update job status to failed: %v", updateErr)
		}
		return err
	}

	// Update job with video ID and completion status
	js.mu.Lock()
	if jobPtr, exists := js.jobs[job.ID]; exists {
		jobPtr.VideoID = videoID
		jobPtr.Progress = 100
	}
	js.mu.Unlock()

	if err := js.UpdateJobStatus(job.ID, models.JobStatusCompleted, ""); err != nil {
		return err
	}

	// Cleanup subtitle files if any were generated
	if subtitleFilePath != "" {
		if err := js.subtitle.CleanupTempFiles(subtitleFilePath); err != nil {
			js.log.Warnf("Failed to cleanup subtitle file %s: %v", subtitleFilePath, err)
		}
	}

	js.log.Infof("Job completed successfully: %s, video ID: %s", job.ID, videoID)
	return nil
}

// needsSubtitles checks if a project needs subtitle generation
func (js *service) needsSubtitles(project models.VideoProject) bool {
	// Check if there are any subtitle elements in the project
	for _, element := range project.Elements {
		if element.Type == "subtitles" {
			return true
		}
	}

	// Check if there are any subtitle elements in scenes
	for _, scene := range project.Scenes {
		for _, element := range scene.Elements {
			if element.Type == "subtitles" {
				return true
			}
		}
	}

	return false
}

// analyzeMediaWithServices uses media services to analyze URLs without downloading
func (js *service) analyzeMediaWithServices(ctx context.Context, config *models.VideoConfigArray) error {
	js.log.Info("Starting media URL analysis with media services")

	for projectIdx := range *config {
		project := &(*config)[projectIdx] // Get pointer to modify original

		// Analyze audio elements to get durations
		for sceneIdx := range project.Scenes {
			for elementIdx := range project.Scenes[sceneIdx].Elements {
				element := &project.Scenes[sceneIdx].Elements[elementIdx]

				switch element.Type {
				case "audio":
					js.log.Debugf("Analyzing audio URL: %s", element.Src)
					audioInfo, err := js.audio.AnalyzeAudio(ctx, element.Src)
					if err != nil {
						js.log.Warnf("Failed to analyze audio '%s': %v, using default duration", element.Src, err)
						element.Duration = 10.0 // Fallback duration
					} else {
						element.Duration = audioInfo.GetDuration()
						js.log.Debugf("Audio duration: %.2fs", element.Duration)
					}
				case "image":
					js.log.Debugf("Validating image URL: %s", element.Src)
					if err := js.image.ValidateImage(element.Src); err != nil {
						js.log.Errorf("Failed to validate image '%s': %v", element.Src, err)
						return fmt.Errorf("invalid image URL '%s': %w", element.Src, err)
					}
					js.log.Debugf("Image URL validated successfully")
				}
			}
		}

		// Analyze background elements (video, image, etc.)
		for elementIdx := range project.Elements {
			element := &project.Elements[elementIdx]
			switch element.Type {
			case "video":
				js.log.Debugf("Analyzing background video URL: %s", element.Src)
				videoInfo, err := js.video.AnalyzeVideo(ctx, element.Src)
				if err != nil {
					js.log.Warnf("Failed to analyze video '%s': %v, using default duration", element.Src, err)
					element.Duration = 30.0 // Fallback duration
				} else {
					element.Duration = videoInfo.GetDuration()
					js.log.Debugf("Video duration: %.2fs", element.Duration)
				}
			case "image":
				js.log.Debugf("Validating background image URL: %s", element.Src)
				if err := js.image.ValidateImage(element.Src); err != nil {
					js.log.Errorf("Failed to validate background image '%s': %v", element.Src, err)
					return fmt.Errorf("invalid background image URL '%s': %w", element.Src, err)
				}
				js.log.Debugf("Background image URL validated successfully")
			}
		}
	}

	js.log.Info("Media URL analysis completed")
	return nil
}

func (js *service) startWorkers() {
	for i := 0; i < js.workers; i++ {
		go js.worker(i)
	}
	js.log.Infof("Started %d job workers", js.workers)
}

func (js *service) worker(id int) {
	js.log.Debugf("Job worker %d started", id)

	for job := range js.jobQueue {
		// Check if job was cancelled
		js.mu.RLock()
		currentJob, exists := js.jobs[job.ID]
		if !exists || currentJob.Status == models.JobStatusCancelled {
			js.mu.RUnlock()
			js.log.Debugf("Skipping cancelled job: %s", job.ID)
			continue
		}
		js.mu.RUnlock()

		// Process the job with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

		workerLog := js.log.WithFields(map[string]interface{}{
			"worker": id,
			"job_id": job.ID,
		})

		workerLog.Info("Worker processing job")

		if err := js.ProcessJob(ctx, job); err != nil {
			workerLog.Errorf("Job processing failed: %v", err)
		} else {
			workerLog.Info("Job processing completed")
		}

		cancel()
	}

	js.log.Debugf("Job worker %d stopped", id)
}

func (js *service) Start() error {
	js.log.Info("Starting job service")
	js.startWorkers()
	return nil
}

func (js *service) Stop() error {
	js.log.Info("Stopping job service")
	close(js.jobQueue)
	return nil
}
