package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/multi-agent/api-gateway/pkg/logger"
)

// Recovery creates recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log the panic
		logger.Error("Panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"stack", string(debug.Stack()),
		)

		// Return error response
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	})
}

// CustomRecovery creates custom recovery middleware with callback
func CustomRecovery(handler func(c *gin.Context, err interface{})) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Log the panic
		logger.Error("Panic recovered",
			"error", recovered,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"user_id", c.GetString("user_id"),
			"stack", string(debug.Stack()),
		)

		// Call custom handler if provided
		if handler != nil {
			handler(c, recovered)
			return
		}

		// Default error response
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
			"request_id": c.GetString("request_id"),
		})
	})
}