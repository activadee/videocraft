package composition

import (
	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/core/media/audio"
	"github.com/activadee/videocraft/internal/core/media/image"
	"github.com/activadee/videocraft/internal/core/media/subtitle"
	"github.com/activadee/videocraft/internal/core/media/video"
	"github.com/activadee/videocraft/internal/core/services/job/queue"
	"github.com/activadee/videocraft/internal/core/services/transcription"
	"github.com/activadee/videocraft/internal/core/video/engine"
	"github.com/activadee/videocraft/internal/pkg/logger"
	storageServices "github.com/activadee/videocraft/internal/storage/filesystem"
)

// Services container
type Services struct {
	FFmpeg        FFmpegService
	Audio         AudioService
	Video         VideoService
	Image         ImageService
	Transcription TranscriptionService
	Subtitle      SubtitleService
	Storage       StorageService
	Job           JobService
}

// Shutdown gracefully shuts down all services
func (s *Services) Shutdown() {
	if s.Transcription != nil {
		s.Transcription.Shutdown()
	}
}

// FFmpegService handles video generation with FFmpeg
type FFmpegService = engine.Service

// AudioService handles audio file analysis and processing
type AudioService = audio.Service

// VideoService handles video file analysis and processing
type VideoService = video.Service

// ImageService handles image file processing and validation
type ImageService = image.Service

// TranscriptionService handles audio transcription
type TranscriptionService = transcription.Service

// SubtitleService handles subtitle generation
type SubtitleService = subtitle.Service

// StorageService handles file storage and management
type StorageService = storageServices.Service

// JobService handles job management and processing
type JobService = queue.Service

// Supporting types that are specific to this package

type FFmpegCommand struct {
	Args       []string
	OutputPath string
}

// NewServices creates a new services container with all implementations
func NewServices(cfg *app.Config, log logger.Logger) *Services {
	// Initialize core services without dependencies first
	audioService := audio.NewService(cfg, log)
	videoService := video.NewService(cfg, log)
	imageService := image.NewService(cfg, log)
	transcriptionService := transcription.NewService(cfg, log)
	ffmpegService := engine.NewService(cfg, log)
	storageService := storageServices.NewService(cfg, log)

	// Initialize services with dependencies
	subtitleService := subtitle.NewService(cfg, log, transcriptionService, audioService)

	// Initialize job service with all dependencies including media services
	jobService := queue.NewService(cfg, log, ffmpegService, subtitleService, storageService, audioService, videoService, imageService)

	return &Services{
		FFmpeg:        ffmpegService,
		Audio:         audioService,
		Video:         videoService,
		Image:         imageService,
		Transcription: transcriptionService,
		Subtitle:      subtitleService,
		Storage:       storageService,
		Job:           jobService,
	}
}
