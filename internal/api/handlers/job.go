package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/services"
	"github.com/activadee/videocraft/pkg/logger"
)

const (
	jobStatusCompleted = "completed"
	jobStatusFailed    = "failed"
)

type JobHandler struct {
	cfg      *config.Config
	services *services.Services
	log      logger.Logger
}

func NewJobHandler(cfg *config.Config, svcContainer *services.Services, log logger.Logger) *JobHandler {
	return &JobHandler{
		cfg:      cfg,
		services: svcContainer,
		log:      log,
	}
}

// GetJobStatus handles GET /jobs/:job_id/status
func (h *JobHandler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	job, err := h.services.Job.GetJob(jobID)
	if err != nil {
		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeJobNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
		}
		return
	}

	response := gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"progress":   job.Progress,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}

	if job.Status == jobStatusCompleted {
		response["video_id"] = job.VideoID
		response["download_url"] = "/download/" + job.VideoID
		response["completed_at"] = job.CompletedAt
	}

	if job.Status == jobStatusFailed {
		response["error"] = job.Error
	}

	c.JSON(http.StatusOK, response)
}

// ListJobs handles GET /jobs
func (h *JobHandler) ListJobs(c *gin.Context) {
	jobs, err := h.services.Job.ListJobs()
	if err != nil {
		h.log.Errorf("Failed to list jobs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list jobs",
		})
		return
	}

	// Transform jobs for response
	jobResponses := make([]gin.H, len(jobs))
	for i, job := range jobs {
		jobResponse := gin.H{
			"job_id":     job.ID,
			"status":     job.Status,
			"progress":   job.Progress,
			"created_at": job.CreatedAt,
			"updated_at": job.UpdatedAt,
		}

		if job.Status == jobStatusCompleted {
			jobResponse["video_id"] = job.VideoID
			jobResponse["download_url"] = "/download/" + job.VideoID
			jobResponse["completed_at"] = job.CompletedAt
		}

		if job.Status == jobStatusFailed {
			jobResponse["error"] = job.Error
		}

		jobResponses[i] = jobResponse
	}

	c.JSON(http.StatusOK, gin.H{
		"jobs":  jobResponses,
		"count": len(jobs),
	})
}

// CancelJob handles POST /jobs/:job_id/cancel
func (h *JobHandler) CancelJob(c *gin.Context) {
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	err := h.services.Job.CancelJob(jobID)
	if err != nil {
		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			switch vpe.Code {
			case errors.ErrCodeJobNotFound:
				status = http.StatusNotFound
			case errors.ErrCodeInvalidInput:
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to cancel job",
			})
		}
		return
	}

	h.log.Infof("Job cancelled: %s", jobID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	})
}

// GetJob handles GET /jobs/:job_id (detailed job information)
func (h *JobHandler) GetJob(c *gin.Context) {
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	job, err := h.services.Job.GetJob(jobID)
	if err != nil {
		if vpe, ok := err.(*errors.VideoProcessingError); ok {
			status := http.StatusInternalServerError
			if vpe.Code == errors.ErrCodeJobNotFound {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{
				"error": vpe.Message,
				"code":  vpe.Code,
			})
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
		}
		return
	}

	response := gin.H{
		"job_id":     job.ID,
		"status":     job.Status,
		"progress":   job.Progress,
		"config":     job.Config,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}

	if job.Status == jobStatusCompleted {
		response["video_id"] = job.VideoID
		response["download_url"] = "/download/" + job.VideoID
		response["completed_at"] = job.CompletedAt
	}

	if job.Status == jobStatusFailed {
		response["error"] = job.Error
	}

	c.JSON(http.StatusOK, response)
}
