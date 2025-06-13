package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

type HealthHandler struct {
	cfg       *config.Config
	services  *services.Services
	log       logger.Logger
	startTime time.Time
}

func NewHealthHandler(cfg *config.Config, svcContainer *services.Services, log logger.Logger) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		services:  svcContainer,
		log:       log,
		startTime: time.Now(),
	}
}

// Basic health check
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

// Detailed health check with system information
// GET /health/detailed
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)

	response := gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
		"uptime": uptime.String(),
		"system": gin.H{
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"memory": gin.H{
				"allocated":   m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"heap_alloc":  m.HeapAlloc,
				"heap_sys":    m.HeapSys,
				"gc_cycles":   m.NumGC,
			},
		},
		"config": gin.H{
			"workers":     h.cfg.Job.Workers,
			"queue_size":  h.cfg.Job.QueueSize,
			"ffmpeg_path": h.cfg.FFmpeg.BinaryPath,
			"output_dir":  h.cfg.Storage.OutputDir,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Kubernetes readiness probe
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if services are ready
	// For now, just return OK
	// TODO: Add actual readiness checks (database, external services, etc.)

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"checks": gin.H{
			"storage": h.checkStorageHealth(),
			"ffmpeg":  h.checkFFmpegHealth(),
		},
	})
}

// Kubernetes liveness probe
// GET /live
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"time":   time.Now().UTC(),
	})
}

// System metrics endpoint
// GET /metrics
func (h *HealthHandler) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get job statistics
	jobs, _ := h.services.Job.ListJobs()

	jobStats := make(map[string]int)
	for _, job := range jobs {
		jobStats[string(job.Status)]++
	}

	// Get video statistics
	videos, _ := h.services.Storage.ListVideos()

	var totalVideoSize int64
	for _, video := range videos {
		totalVideoSize += video.Size
	}

	metrics := gin.H{
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(h.startTime).Seconds(),
		"memory": gin.H{
			"allocated_mb": float64(m.Alloc) / 1024 / 1024,
			"heap_mb":      float64(m.HeapAlloc) / 1024 / 1024,
			"gc_cycles":    m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"jobs": gin.H{
			"total":     len(jobs),
			"by_status": jobStats,
		},
		"storage": gin.H{
			"videos_count":  len(videos),
			"total_size_mb": float64(totalVideoSize) / 1024 / 1024,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func (h *HealthHandler) checkStorageHealth() string {
	// Check if output directory is writable
	// This is a simple check - in production you might want more thorough checks
	videos, err := h.services.Storage.ListVideos()
	if err != nil {
		return "unhealthy: " + err.Error()
	}

	return "healthy (" + string(rune(len(videos))) + " videos)"
}

func (h *HealthHandler) checkFFmpegHealth() string {
	// TODO: Implement FFmpeg binary check
	// This could run a simple FFmpeg command to verify it's working
	return "unknown"
}
