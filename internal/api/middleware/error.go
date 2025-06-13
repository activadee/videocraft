package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/pkg/logger"
)

func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle errors if any occurred
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			log.WithField("error", err.Error()).Error("Request error")

			// Check if it's our custom error type
			if vpe, ok := err.Err.(*errors.VideoProcessingError); ok {
				status := getStatusFromErrorCode(vpe.Code)
				c.JSON(status, gin.H{
					"error": vpe.Message,
					"code":  vpe.Code,
				})
				return
			}

			// Generic error handling
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INTERNAL_ERROR",
			})
		}
	}
}

func getStatusFromErrorCode(code string) int {
	switch code {
	case errors.ErrCodeInvalidInput:
		return http.StatusBadRequest
	case errors.ErrCodeFileNotFound:
		return http.StatusNotFound
	case errors.ErrCodeJobNotFound:
		return http.StatusNotFound
	case errors.ErrCodeTimeout:
		return http.StatusRequestTimeout
	case errors.ErrCodeFFmpegFailed:
		return http.StatusUnprocessableEntity
	case errors.ErrCodeTranscriptionFailed:
		return http.StatusUnprocessableEntity
	case errors.ErrCodeDownloadFailed:
		return http.StatusBadGateway
	case errors.ErrCodeStorageFailed:
		return http.StatusInsufficientStorage
	default:
		return http.StatusInternalServerError
	}
}
