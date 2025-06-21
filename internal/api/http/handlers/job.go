package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/core/video/composition"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// JobHandler handles job-related HTTP requests
type JobHandler struct {
	services *composition.Services
	logger   logger.Logger
}

// NewJobHandler creates a new job handler
func NewJobHandler(services *composition.Services, logger logger.Logger) *JobHandler {
	return &JobHandler{
		services: services,
		logger:   logger,
	}
}

// GetJob handles GET /jobs/:id - REST-compliant job status
func (h *JobHandler) GetJob(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Debugf("Get job request for ID: %s", jobID)

	// Validate job ID
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	// Get job from service
	job, err := h.services.Job.GetJob(jobID)
	if err != nil {
		h.logger.Errorf("Failed to get job %s: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"job_id": jobID,
		})
		return
	}

	// Build response
	response := gin.H{
		"job_id": job.ID,
		"video_id": job.VideoID,
		"status": job.Status,
		"progress": job.Progress,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}

	// Add conditional fields
	if job.CompletedAt != nil {
		response["completed_at"] = job.CompletedAt
		response["duration_seconds"] = job.CompletedAt.Sub(job.CreatedAt).Seconds()
	}

	if job.Error != "" {
		response["error"] = job.Error
	}

	// Add video URL if completed
	if job.Status == "completed" && job.VideoID != "" {
		response["video_url"] = fmt.Sprintf("/api/v1/videos/%s", job.VideoID)
	}

	c.JSON(http.StatusOK, response)
}

// DeleteJob handles DELETE /jobs/:id - REST-compliant job cancellation
func (h *JobHandler) DeleteJob(c *gin.Context) {
	jobID := c.Param("id")
	h.logger.Debugf("Job status request for ID: %s", jobID)

	// Validate job ID
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	// Get job from service
	job, err := h.services.Job.GetJob(jobID)
	if err != nil {
		h.logger.Errorf("Failed to get job %s: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
			"job_id": jobID,
		})
		return
	}

	// Build response
	response := gin.H{
		"job_id": job.ID,
		"video_id": job.VideoID,
		"status": job.Status,
		"progress": job.Progress,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}

	// Add conditional fields
	if job.CompletedAt != nil {
		response["completed_at"] = job.CompletedAt
		response["duration_seconds"] = job.CompletedAt.Sub(job.CreatedAt).Seconds()
	}

	if job.Error != "" {
		response["error"] = job.Error
	}

	// TODO: Implement job cancellation logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancellation not yet implemented",
		"job_id": jobID,
	})
}

// getIntQueryParam gets an integer query parameter with a default value
func (h *JobHandler) getIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
