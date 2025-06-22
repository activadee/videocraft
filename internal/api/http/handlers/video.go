package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/api/models"
	"github.com/activadee/videocraft/internal/core/video/composition"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

type VideoHandler struct {
	services *composition.Services
	log      logger.Logger
}

func NewVideoHandler(services *composition.Services, log logger.Logger) *VideoHandler {
	return &VideoHandler{
		services: services,
		log:      log,
	}
}

// CreateVideo handles POST /videos - REST-compliant video creation
func (h *VideoHandler) CreateVideo(c *gin.Context) {
	h.log.Info("Generate video request received")

	// Parse request body
	var config models.VideoConfigArray
	if err := c.ShouldBindJSON(&config); err != nil {
		h.log.Errorf("Failed to parse video config: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON format",
			"details": err.Error(),
		})
		return
	}

	// Validate configuration
	if len(config) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No video projects provided",
		})
		return
	}

	// Quick URL validation without downloading
	if err := h.validateMediaURLs(&config); err != nil {
		h.log.Errorf("Media URL validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid media URLs",
			"details": err.Error(),
		})
		return
	}

	// Create job for async processing
	job, err := h.services.Job.CreateJob(&config)
	if err != nil {
		h.log.Errorf("Failed to create job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create video generation job",
		})
		return
	}

	// Start background processing
	go func() {
		ctx := context.Background()
		if err := h.services.Job.ProcessJob(ctx, job); err != nil {
			h.log.Errorf("Background job processing failed: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"job_id": job.ID,
		"video_id": job.VideoID,
		"status": job.Status,
		"message": "Video generation started",
		"status_url": fmt.Sprintf("/api/v1/jobs/%s", job.ID),
	})
}

// GetVideo handles GET /videos/:id - Returns video file or status
func (h *VideoHandler) GetVideo(c *gin.Context) {
	videoID := c.Param("id")
	h.log.Debugf("Download video request for ID: %s", videoID)

	// Validate video ID
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Video ID is required",
		})
		return
	}

	// Get video file path from storage
	filePath, err := h.services.Storage.GetVideo(videoID)
	if err != nil {
		h.log.Errorf("Failed to get video %s: %v", videoID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Video not found",
			"video_id": videoID,
		})
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.log.Errorf("Video file not found on disk: %s", filePath)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Video file not found",
			"video_id": videoID,
		})
		return
	}

	// Set appropriate headers for video download
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="video_%s.mp4"`, videoID))
	c.Header("Cache-Control", "no-cache")

	// Stream the file
	c.File(filePath)
	h.log.Infof("Video %s downloaded successfully", videoID)
}


// validateMediaURLs performs lightweight URL validation without downloading
func (h *VideoHandler) validateMediaURLs(config *models.VideoConfigArray) error {
	for _, project := range *config {
		// Validate background video URLs
		for _, element := range project.Elements {
			if element.Type == "video" {
				if err := h.services.Video.ValidateVideo(element.Src); err != nil {
					return fmt.Errorf("invalid background video URL '%s': %w", element.Src, err)
				}
			}
		}
		
		// Validate scene element URLs
		for _, scene := range project.Scenes {
			for _, element := range scene.Elements {
				switch element.Type {
				case "audio":
					// Just validate URL format, don't download
					if element.Src == "" {
						return fmt.Errorf("audio URL cannot be empty")
					}
					if err := h.validateURL(element.Src); err != nil {
						return fmt.Errorf("invalid audio URL '%s': %w", element.Src, err)
					}
					
				case "image":
					if err := h.services.Image.ValidateImage(element.Src); err != nil {
						return fmt.Errorf("invalid image URL '%s': %w", element.Src, err)
					}
				}
			}
		}
	}
	
	return nil
}

// validateURL performs basic URL validation
func (h *VideoHandler) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	// Check protocol
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS protocols are allowed")
	}
	
	return nil
}
