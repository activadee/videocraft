package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuth_MiddlewareBlocksUnauthorizedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should reject requests without API key", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-api-key"))
		router.GET("/api/v1/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "API key is required")
		assert.Contains(t, w.Body.String(), "MISSING_API_KEY")
	})

	t.Run("should reject requests with invalid API key", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("correct-api-key"))
		router.GET("/api/v1/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer wrong-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid API key")
		assert.Contains(t, w.Body.String(), "INVALID_API_KEY")
	})

	t.Run("should allow requests with correct API key", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-api-key"))
		router.GET("/api/v1/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer test-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("should allow health endpoints without authentication", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-api-key"))
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/health", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
	})
}

func TestAuth_SecurityErrorMessages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should provide proper error structure for missing auth", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-key"))
		router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Response should have proper structure for API clients
		body := w.Body.String()
		assert.Contains(t, body, "\"error\":")
		assert.Contains(t, body, "\"code\":")
		assert.Contains(t, body, "MISSING_API_KEY")
	})

	t.Run("should provide proper error structure for invalid auth", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("correct-key"))
		router.GET("/test", func(c *gin.Context) { c.JSON(200, gin.H{}) })

		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer invalid-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		body := w.Body.String()
		assert.Contains(t, body, "\"error\":")
		assert.Contains(t, body, "\"code\":")
		assert.Contains(t, body, "INVALID_API_KEY")
	})
}

func TestAuth_AlternativeAuthMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should accept API key via query parameter", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-api-key"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/test?api_key=test-api-key", http.NoBody)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("should accept API key via Authorization header without Bearer prefix", func(t *testing.T) {
		router := gin.New()
		router.Use(Auth("test-api-key"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "test-api-key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}
