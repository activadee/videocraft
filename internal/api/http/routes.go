package http

import (
	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/api/http/handlers"
	"github.com/activadee/videocraft/internal/api/http/middleware"
	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/core/video/composition"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

func NewRouter(cfg *app.Config, services *composition.Services, log logger.Logger) *gin.Engine {
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
	healthHandler := handlers.NewHealthHandler(services, log)
	videoHandler := handlers.NewVideoHandler(services, log)
	jobHandler := handlers.NewJobHandler(services, log)

	// Setup routes
	setupRoutes(router, cfg, log, healthHandler, videoHandler, jobHandler)

	return router
}

func setupMiddleware(router *gin.Engine, cfg *app.Config, log logger.Logger) {
	// Recovery middleware
	router.Use(gin.Recovery())

	// Custom logging middleware
	router.Use(middleware.Logger(log))

	// Secure CORS middleware - NO WILDCARDS
	router.Use(middleware.SecureCORS(cfg, log))

	// CSRF protection middleware
	router.Use(middleware.CSRFProtection(cfg, log))

	// Error handling middleware
	router.Use(middleware.SecureErrorHandler(log))

	// Rate limiting middleware (if enabled)
	if cfg.Security.RateLimit > 0 {
		router.Use(middleware.RateLimit(cfg.Security.RateLimit))
	}

	// Authentication middleware (if enabled) - BEFORE validation
	if cfg.Security.EnableAuth {
		router.Use(middleware.Auth(cfg.Security.APIKey))
	}

	// Request size limiting (1MB max)
	router.Use(middleware.RequestSizeLimit(1024 * 1024))

	// Input validation middleware - AFTER authentication
	router.Use(middleware.ValidationMiddleware(log))
}

func setupRoutes(
	router *gin.Engine,
	cfg *app.Config,
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

	// REST-compliant Video API
	v1.POST("/videos", videoHandler.CreateVideo) // Create video job
	v1.GET("/videos/:id", videoHandler.GetVideo) // Get video or status

	// REST-compliant Job API
	v1.GET("/jobs/:id", jobHandler.GetJob)       // Get job status
	v1.DELETE("/jobs/:id", jobHandler.DeleteJob) // Cancel job

	// Documentation endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "VideoCraft API",
			"description": "FFmpeg Dialogue Video Generation Service",
			"version":     "0.0.1",
			"features": gin.H{
				"v0.0.1":        "Initial release with security-first design",
				"documentation": "/docs/README.md",
			},
			"endpoints": gin.H{
				"health": gin.H{
					"GET /health":          "Basic health check",
					"GET /health/detailed": "Detailed health information",
					"GET /ready":           "Kubernetes readiness probe",
					"GET /live":            "Kubernetes liveness probe",
					"GET /metrics":         "System metrics",
				},
				"video_generation": gin.H{
					"POST /api/v1/generate-video": "Start video generation job",
				},
				"video_management": gin.H{
					"GET /api/v1/download/:video_id":  "Download generated video",
					"GET /api/v1/status/:video_id":    "Get video status",
					"GET /api/v1/videos":              "List all videos",
					"DELETE /api/v1/videos/:video_id": "Delete video",
				},
				"job_management": gin.H{
					"GET /api/v1/jobs":                 "List all jobs",
					"GET /api/v1/jobs/:job_id":         "Get job details",
					"GET /api/v1/jobs/:job_id/status":  "Get job status",
					"POST /api/v1/jobs/:job_id/cancel": "Cancel job",
				},
				"authentication": gin.H{
					"GET /api/v1/csrf-token": "Get CSRF token for authenticated requests",
				},
			},
			"examples": gin.H{
				"generate_video": gin.H{
					"url": "POST /api/v1/generate-video",
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
