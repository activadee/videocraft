package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/activadee/videocraft/internal/api"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

// Build information (set via ldflags)
var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	// Parse command line flags
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.Log.Level)

	// Initialize services
	services := initializeServices(cfg, appLogger)

	// Setup router
	router := api.NewRouter(cfg, services, appLogger)

	// Start server
	srv := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server:", err)
		}
	}()

	appLogger.Info("Server started on ", cfg.Server.Address())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown services first (this will stop the daemon)
	if services != nil {
		services.Shutdown()
	}

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown:", err)
	}

	appLogger.Info("Server exited")
}

func initializeServices(cfg *config.Config, appLogger logger.Logger) *services.Services {
	// Initialize all services with dependency injection
	audioSvc := services.NewAudioService(cfg, appLogger)
	transcriptionSvc := services.NewTranscriptionService(cfg, appLogger)
	subtitleSvc := services.NewSubtitleService(cfg, appLogger, transcriptionSvc, audioSvc)
	ffmpegSvc := services.NewFFmpegService(cfg, appLogger, transcriptionSvc, subtitleSvc, audioSvc)
	storageSvc := services.NewStorageService(cfg, appLogger)
	jobSvc := services.NewJobService(cfg, appLogger, ffmpegSvc, audioSvc, transcriptionSvc, storageSvc)

	return &services.Services{
		FFmpeg:        ffmpegSvc,
		Audio:         audioSvc,
		Transcription: transcriptionSvc,
		Subtitle:      subtitleSvc,
		Storage:       storageSvc,
		Job:           jobSvc,
	}
}

func printVersion() {
	fmt.Printf("VideoCraft %s\n", version)
	fmt.Printf("Git Commit: %s\n", gitCommit)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func printHelp() {
	fmt.Println("VideoCraft - Advanced Video Generation Platform")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  videocraft [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  -help      Show help information")
	fmt.Println("  -version   Show version information")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES:")
	fmt.Println("  Configuration can be set via environment variables with VIDEOCRAFT_ prefix")
	fmt.Println("  Example: VIDEOCRAFT_SERVER_PORT=8080")
	fmt.Println()
	fmt.Println("CONFIGURATION:")
	fmt.Println("  Configuration files are searched in:")
	fmt.Println("  - ./config.yaml")
	fmt.Println("  - ./config/config.yaml")
	fmt.Println("  - /etc/videocraft/config.yaml")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/activadee/videocraft")
}
