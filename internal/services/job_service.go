package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
)

type jobService struct {
	cfg           *config.Config
	log           logger.Logger
	ffmpeg        FFmpegService
	audio         AudioService
	transcription TranscriptionService
	storage       StorageService

	jobs      map[string]*models.Job
	mu        sync.RWMutex
	jobQueue  chan *models.Job
	workers   int
}

func newJobService(
	cfg *config.Config,
	log logger.Logger,
	ffmpeg FFmpegService,
	audio AudioService,
	transcription TranscriptionService,
	storage StorageService,
) JobService {
	js := &jobService{
		cfg:           cfg,
		log:           log,
		ffmpeg:        ffmpeg,
		audio:         audio,
		transcription: transcription,
		storage:       storage,
		jobs:          make(map[string]*models.Job),
		jobQueue:      make(chan *models.Job, cfg.Job.QueueSize),
		workers:       cfg.Job.Workers,
	}

	// Start worker pool
	js.startWorkers()

	return js
}

func (js *jobService) CreateJob(config *models.VideoConfigArray) (*models.Job, error) {
	js.log.Debug("Creating new job")

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, errors.InvalidInput(err.Error())
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

func (js *jobService) GetJob(id string) (*models.Job, error) {
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

func (js *jobService) ListJobs() ([]*models.Job, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	jobs := make([]*models.Job, 0, len(js.jobs))
	for _, job := range js.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	return jobs, nil
}

func (js *jobService) CancelJob(id string) error {
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

func (js *jobService) UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error {
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

func (js *jobService) UpdateJobProgress(id string, progress int) error {
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

func (js *jobService) ProcessJob(ctx context.Context, job *models.Job) error {
	js.log.Infof("Processing job: %s", job.ID)

	// Update status to processing
	if err := js.UpdateJobStatus(job.ID, models.JobStatusProcessing, ""); err != nil {
		return err
	}

	// Create progress channel
	progressChan := make(chan int, 10)
	go func() {
		for progress := range progressChan {
			js.UpdateJobProgress(job.ID, progress)
		}
	}()

	// Process the video generation
	videoPath, err := js.ffmpeg.GenerateVideo(ctx, &job.Config, progressChan)
	// Note: progressChan is closed by the FFmpeg service

	if err != nil {
		js.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error())
		return err
	}

	// Store the generated video
	videoID, err := js.storage.StoreVideo(videoPath)
	if err != nil {
		js.UpdateJobStatus(job.ID, models.JobStatusFailed, err.Error())
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

	js.log.Infof("Job completed successfully: %s, video ID: %s", job.ID, videoID)
	return nil
}

func (js *jobService) startWorkers() {
	for i := 0; i < js.workers; i++ {
		go js.worker(i)
	}
	js.log.Infof("Started %d job workers", js.workers)
}

func (js *jobService) worker(id int) {
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