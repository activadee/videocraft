package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/activadee/videocraft/internal/pkg/logger"
)

func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log fields
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"status":     c.Writer.Status(),
			"latency":    latency.String(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		if raw != "" {
			fields["query"] = raw
		}

		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		status := c.Writer.Status()
		message := "Request completed"

		switch {
		case status >= 500:
			log.WithFields(fields).Error(message)
		case status >= 400:
			log.WithFields(fields).Warn(message)
		default:
			log.WithFields(fields).Info(message)
		}
	}
}
