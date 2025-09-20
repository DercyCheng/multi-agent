package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multi-agent/api-gateway/pkg/logger"
)

// Logger creates logging middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log request details
		logger.Info("Request processed",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)
		
		return ""
	})
}

// RequestLogger creates detailed request logging middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get user info if available
		userID := c.GetString("user_id")
		tenantID := c.GetString("tenant_id")

		// Build log fields
		fields := []interface{}{
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency,
			"client_ip", clientIP,
			"user_agent", c.Request.UserAgent(),
		}

		if raw != "" {
			fields = append(fields, "query", raw)
		}

		if userID != "" {
			fields = append(fields, "user_id", userID)
		}

		if tenantID != "" {
			fields = append(fields, "tenant_id", tenantID)
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, "errors", c.Errors.String())
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("Server error", fields...)
		case status >= 400:
			logger.Warn("Client error", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}