package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/activadee/videocraft/internal/api"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Log.Level)

	// Initialize services
	services := initializeServices(cfg, logger)

	// Setup router
	router := api.NewRouter(cfg, services, logger)

	// Start server
	srv := &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server:", err)
		}
	}()

	logger.Info("Server started on ", cfg.Server.Address())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown services first (this will stop the daemon)
	if services != nil {
		services.Shutdown()
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}

func initializeServices(cfg *config.Config, log logger.Logger) *services.Services {
	// Initialize all services with dependency injection
	audioSvc := services.NewAudioService(cfg, log)
	transcriptionSvc := services.NewTranscriptionService(cfg, log)
	subtitleSvc := services.NewSubtitleService(cfg, log, transcriptionSvc, audioSvc)
	ffmpegSvc := services.NewFFmpegService(cfg, log, transcriptionSvc, subtitleSvc, audioSvc)
	storageSvc := services.NewStorageService(cfg, log)
	jobSvc := services.NewJobService(cfg, log, ffmpegSvc, audioSvc, transcriptionSvc, storageSvc)

	return &services.Services{
		FFmpeg:        ffmpegSvc,
		Audio:         audioSvc,
		Transcription: transcriptionSvc,
		Subtitle:      subtitleSvc,
		Storage:       storageSvc,
		Job:           jobSvc,
	}
}