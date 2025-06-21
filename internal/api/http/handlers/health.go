package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/core/video/composition"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// HealthHandler handles health check and system status requests
type HealthHandler struct {
	services  *composition.Services
	logger    logger.Logger
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(services *composition.Services, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		services:  services,
		logger:    logger,
		startTime: time.Now(),
	}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

// HealthDetailed handles GET /health/detailed
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(h.startTime).String(),
		"system": gin.H{
			"goroutines": runtime.NumGoroutine(),
			"memory": gin.H{
				"alloc_mb":       bToMb(m.Alloc),
				"total_alloc_mb": bToMb(m.TotalAlloc),
				"sys_mb":         bToMb(m.Sys),
				"gc_runs":        m.NumGC,
			},
		},
		"services": gin.H{
			"ffmpeg":        "healthy", // TODO: Add actual health checks in Phase 2
			"transcription": "healthy",
			"storage":       "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}

// Metrics handles GET /metrics
func (h *HealthHandler) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := gin.H{
		"timestamp":                time.Now().UTC(),
		"uptime_seconds":           time.Since(h.startTime).Seconds(),
		"goroutines":               runtime.NumGoroutine(),
		"memory_alloc_bytes":       m.Alloc,
		"memory_total_alloc_bytes": m.TotalAlloc,
		"memory_sys_bytes":         m.Sys,
		"gc_runs":                  m.NumGC,
		"jobs_total":               0, // TODO: Add actual metrics in Phase 2
		"jobs_active":              0,
		"jobs_completed":           0,
		"jobs_failed":              0,
	}

	c.JSON(http.StatusOK, metrics)
}

// Ready handles GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: Add actual readiness checks in Phase 2
	// Check if all required services are available
	ready := true
	checks := gin.H{
		"database": "ok",
		"storage":  "ok",
		"ffmpeg":   "ok",
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ready":     ready,
		"checks":    checks,
		"timestamp": time.Now().UTC(),
	})
}

// Live handles GET /live
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": time.Now().UTC(),
	})
}

// Helper function to convert bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
