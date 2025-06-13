package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

type VideoHandler struct {
	cfg      *config.Config
	services *services.Services
	log      logger.Logger
}

func NewVideoHandler(cfg *config.Config, svcContainer *services.Services, log logger.Logger) *VideoHandler {
	return &VideoHandler{
		cfg:      cfg,
		services: svcContainer,
		log:      log,
	}
}

// GenerateVideo handles POST /generate-video
func (h *VideoHandler) GenerateVideo(c *gin.Context) {
	var videoConfig models.VideoConfigArray

	if err := c.ShouldBindJSON(&videoConfig); err != nil {
		h.log.Debugf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate configuration
	if err := videoConfig.Validate(); err != nil {
		h.log.Debugf("Invalid configuration: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid configuration",
			"details": err.Error(),
		})
		return
	}

	// Create job
	job, err := h.services.Job.CreateJob(&videoConfig)
	if err != nil {
		h.log.Errorf("Failed to create job: %v", err)

		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeInvalidInput {
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create job",
			})
		}
		return
	}

	h.log.Infof("Video generation job created: %s", job.ID)

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"message":    "Video generation started",
		"created_at": job.CreatedAt,
		"status_url": "/jobs/" + job.ID + "/status",
	})
}

// DownloadVideo handles GET /download/:video_id
func (h *VideoHandler) DownloadVideo(c *gin.Context) {
	videoID := c.Param("video_id")

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Video ID is required",
		})
		return
	}

	// Get video file path
	videoPath, err := h.services.Storage.GetVideo(videoID)
	if err != nil {
		h.log.Debugf("Video not found: %s", videoID)

		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeFileNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Video not found",
			})
		}
		return
	}

	// Set headers for file download
	filename := filepath.Base(videoPath)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "video/mp4")

	// Serve the file
	c.File(videoPath)
}

// GetVideoStatus handles GET /status/:video_id
func (h *VideoHandler) GetVideoStatus(c *gin.Context) {
	videoID := c.Param("video_id")

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Video ID is required",
		})
		return
	}

	// Check if video exists
	_, err := h.services.Storage.GetVideo(videoID)
	if err != nil {
		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeFileNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Video not found",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"video_id":     videoID,
		"status":       "available",
		"download_url": "/download/" + videoID,
	})
}

// ListVideos handles GET /videos
func (h *VideoHandler) ListVideos(c *gin.Context) {
	videos, err := h.services.Storage.ListVideos()
	if err != nil {
		h.log.Errorf("Failed to list videos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list videos",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos": videos,
		"count":  len(videos),
	})
}

// DeleteVideo handles DELETE /videos/:video_id
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	videoID := c.Param("video_id")

	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Video ID is required",
		})
		return
	}

	err := h.services.Storage.DeleteVideo(videoID)
	if err != nil {
		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeFileNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete video",
			})
		}
		return
	}

	h.log.Infof("Video deleted: %s", videoID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Video deleted successfully",
		"video_id": videoID,
	})
}
