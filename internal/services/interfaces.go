package services

import (
	"context"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
)

// Services container
type Services struct {
	FFmpeg        FFmpegService
	Audio         AudioService
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
type FFmpegService interface {
	GenerateVideo(ctx context.Context, config *models.VideoConfigArray, progressChan chan<- int) (string, error)
	BuildCommand(config *models.VideoConfigArray) (*FFmpegCommand, error)
	Execute(ctx context.Context, cmd *FFmpegCommand) error
}

// AudioService handles audio file analysis and processing
type AudioService interface {
	AnalyzeAudio(ctx context.Context, url string) (*AudioInfo, error)
	CalculateSceneTiming(elements []models.Element) ([]models.TimingSegment, error)
	DownloadAudio(ctx context.Context, url string) (string, error)
}

// TranscriptionService handles audio transcription
type TranscriptionService interface {
	TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error)
	Shutdown()
}

// SubtitleService handles subtitle generation
type SubtitleService interface {
	GenerateSubtitles(ctx context.Context, project models.VideoProject) (*SubtitleResult, error)
	ValidateSubtitleConfig(project models.VideoProject) error
	CleanupTempFiles(filePath string) error
}

// StorageService handles file storage and management
type StorageService interface {
	StoreVideo(videoPath string) (string, error)
	GetVideo(videoID string) (string, error)
	DeleteVideo(videoID string) error
	ListVideos() ([]VideoInfo, error)
	CleanupOldFiles() error
}

// JobService handles job management and processing
type JobService interface {
	CreateJob(config *models.VideoConfigArray) (*models.Job, error)
	GetJob(id string) (*models.Job, error)
	ListJobs() ([]*models.Job, error)
	CancelJob(id string) error
	UpdateJobStatus(id string, status models.JobStatus, errorMsg string) error
	UpdateJobProgress(id string, progress int) error
	ProcessJob(ctx context.Context, job *models.Job) error
}

// Supporting types

type FFmpegCommand struct {
	Args       []string
	OutputPath string
}

type AudioInfo struct {
	URL      string  `json:"url"`
	Duration float64 `json:"duration"`
	Format   string  `json:"format"`
	Bitrate  int     `json:"bitrate"`
	Size     int64   `json:"size"`
}

type VideoInfo struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

// Constructor functions
func NewFFmpegService(cfg *config.Config, log logger.Logger, transcription TranscriptionService, subtitle SubtitleService, audio AudioService) FFmpegService {
	return &ffmpegService{
		cfg:           cfg,
		log:           log,
		transcription: transcription,
		subtitle:      subtitle,
		audio:         audio,
	}
}

func NewAudioService(cfg *config.Config, log logger.Logger) AudioService {
	return &audioService{cfg: cfg, log: log}
}

func NewTranscriptionService(cfg *config.Config, log logger.Logger) TranscriptionService {
	return newTranscriptionService(cfg, log)
}

func NewSubtitleService(cfg *config.Config, log logger.Logger, transcription TranscriptionService, audio AudioService) SubtitleService {
	return newSubtitleService(cfg, log, transcription, audio)
}

func NewStorageService(cfg *config.Config, log logger.Logger) StorageService {
	return &storageService{cfg: cfg, log: log}
}

func NewJobService(
	cfg *config.Config, 
	log logger.Logger,
	ffmpeg FFmpegService,
	audio AudioService,
	transcription TranscriptionService,
	storage StorageService,
) JobService {
	return newJobService(cfg, log, ffmpeg, audio, transcription, storage)
}