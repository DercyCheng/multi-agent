package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multi-agent/api-gateway/pkg/metrics"
)

// Metrics creates metrics collection middleware
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment active connections
		metrics.IncActiveConnections("gateway")
		defer metrics.DecActiveConnections("gateway")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "error"
		}

		// Record metrics
		metrics.RecordRequest(
			"gateway",
			c.Request.Method,
			status,
			c.Writer.Status(),
			duration,
		)
	}
}

// ProxyMetrics creates metrics middleware for proxy requests
func ProxyMetrics(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		status := "success"
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			status = "error"
		}

		// Record proxy metrics
		metrics.RecordRequest(
			serviceName,
			c.Request.Method,
			status,
			statusCode,
			duration,
		)
	}
}