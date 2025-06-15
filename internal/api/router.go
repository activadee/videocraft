package api

import (
	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/api/handlers"
	"github.com/activadee/videocraft/internal/api/middleware"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

func NewRouter(cfg *config.Config, services *services.Services, log logger.Logger) *gin.Engine {
	// Set Gin mode
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	setupMiddleware(router, cfg, log)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(cfg, services, log)
	videoHandler := handlers.NewVideoHandler(cfg, services, log)
	jobHandler := handlers.NewJobHandler(cfg, services, log)

	// Setup routes
	setupRoutes(router, cfg, log, healthHandler, videoHandler, jobHandler)

	return router
}

func setupMiddleware(router *gin.Engine, cfg *config.Config, log logger.Logger) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Custom logging middleware
	router.Use(middleware.Logger(log))

	// Secure CORS middleware - NO WILDCARDS
	router.Use(middleware.SecureCORS(cfg, log))

	// CSRF protection middleware
	router.Use(middleware.CSRFProtection(cfg, log))

	// Error handling middleware
	router.Use(middleware.ErrorHandler(log))

	// Rate limiting middleware (if enabled)
	if cfg.Security.RateLimit > 0 {
		router.Use(middleware.RateLimit(cfg.Security.RateLimit))
	}

	// Authentication middleware will be applied per route group, not globally
}

func setupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	log logger.Logger,
	healthHandler *handlers.HealthHandler,
	videoHandler *handlers.VideoHandler,
	jobHandler *handlers.JobHandler,
) {
	// Health endpoints
	router.GET("/health", healthHandler.Health)
	router.GET("/health/detailed", healthHandler.HealthDetailed)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)
	router.GET("/metrics", healthHandler.Metrics)

	// CSRF token endpoint (no auth required) - must be outside authenticated groups
	router.GET("/api/v1/csrf-token", middleware.CSRFTokenEndpoint(cfg, log))

	// API v1 routes with authentication
	v1 := router.Group("/api/v1")
	if cfg.Security.EnableAuth {
		v1.Use(middleware.Auth(cfg.Security.APIKey))
	}

	// Video generation
	v1.POST("/generate-video", videoHandler.GenerateVideo)

	// Video management
	v1.GET("/download/:video_id", videoHandler.DownloadVideo)
	v1.GET("/status/:video_id", videoHandler.GetVideoStatus)
	v1.GET("/videos", videoHandler.ListVideos)
	v1.DELETE("/videos/:video_id", videoHandler.DeleteVideo)

	// Job management
	v1.GET("/jobs", jobHandler.ListJobs)
	v1.GET("/jobs/:job_id", jobHandler.GetJob)
	v1.GET("/jobs/:job_id/status", jobHandler.GetJobStatus)
	v1.POST("/jobs/:job_id/cancel", jobHandler.CancelJob)

	// Legacy routes (for backward compatibility with Python version)
	router.POST("/generate-video", videoHandler.GenerateVideo)
	router.GET("/download/:video_id", videoHandler.DownloadVideo)
	router.GET("/status/:video_id", videoHandler.GetVideoStatus)
	router.GET("/videos", videoHandler.ListVideos)
	router.DELETE("/videos/:video_id", videoHandler.DeleteVideo)
	router.GET("/jobs", jobHandler.ListJobs)
	router.GET("/jobs/:job_id", jobHandler.GetJob)
	router.GET("/jobs/:job_id/status", jobHandler.GetJobStatus)
	router.POST("/jobs/:job_id/cancel", jobHandler.CancelJob)

	// Documentation endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "VideoCraft API",
			"description": "FFmpeg Dialogue Video Generation Service",
			"version":     "1.0.0-go",
			"endpoints": gin.H{
				"health": gin.H{
					"GET /health":          "Basic health check",
					"GET /health/detailed": "Detailed health information",
					"GET /ready":           "Kubernetes readiness probe",
					"GET /live":            "Kubernetes liveness probe",
					"GET /metrics":         "System metrics",
				},
				"video_generation": gin.H{
					"POST /generate-video":        "Start video generation job",
					"POST /api/v1/generate-video": "Start video generation job (v1)",
				},
				"video_management": gin.H{
					"GET /download/:video_id":  "Download generated video",
					"GET /status/:video_id":    "Get video status",
					"GET /videos":              "List all videos",
					"DELETE /videos/:video_id": "Delete video",
				},
				"job_management": gin.H{
					"GET /jobs":                 "List all jobs",
					"GET /jobs/:job_id":         "Get job details",
					"GET /jobs/:job_id/status":  "Get job status",
					"POST /jobs/:job_id/cancel": "Cancel job",
				},
			},
			"examples": gin.H{
				"generate_video": gin.H{
					"url": "POST /generate-video",
					"body": gin.H{
						"background_video": "https://example.com/background.mp4",
						"audio_files": []gin.H{
							{
								"url":   "https://example.com/audio1.mp3",
								"label": "Speaker 1",
							},
						},
						"subtitle_settings": gin.H{
							"enabled": true,
							"style":   "progressive",
						},
					},
				},
			},
		})
	})
}
