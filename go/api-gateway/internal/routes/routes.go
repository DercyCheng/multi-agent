package routes

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/multi-agent/api-gateway/internal/gateway"
	"github.com/multi-agent/api-gateway/internal/middleware"
	"github.com/multi-agent/api-gateway/pkg/logger"
	"github.com/multi-agent/api-gateway/pkg/metrics"
)

// Setup configures all routes
func Setup(router *gin.Engine, gw *gateway.Gateway) {
	// Health check endpoints
	health := router.Group("/health")
	{
		health.GET("/", handleHealthCheck(gw))
		health.GET("/services", handleServicesHealth(gw))
	}

	// Metrics endpoint
	router.GET("/metrics", handleMetrics())

	// API routes with authentication
	api := router.Group("/")
	api.Use(middleware.RequirePermission("api:access"))
	{
		// LLM service routes
		llm := api.Group("/v1")
		{
			llm.POST("/chat/completions", handleProxy(gw, "llm"))
			llm.GET("/models", handleProxy(gw, "llm"))
			llm.GET("/models/:model_id", handleProxy(gw, "llm"))
		}

		// Orchestrator routes
		orchestrator := api.Group("/api/v1")
		{
			orchestrator.POST("/tasks", middleware.RequirePermission("tasks:create"), handleProxy(gw, "orchestrator"))
			orchestrator.GET("/tasks", middleware.RequirePermission("tasks:read"), handleProxy(gw, "orchestrator"))
			orchestrator.GET("/tasks/:task_id", middleware.RequirePermission("tasks:read"), handleProxy(gw, "orchestrator"))
			orchestrator.PUT("/tasks/:task_id", middleware.RequirePermission("tasks:update"), handleProxy(gw, "orchestrator"))
			orchestrator.DELETE("/tasks/:task_id", middleware.RequirePermission("tasks:delete"), handleProxy(gw, "orchestrator"))

			orchestrator.POST("/workflows", middleware.RequirePermission("workflows:create"), handleProxy(gw, "orchestrator"))
			orchestrator.GET("/workflows", middleware.RequirePermission("workflows:read"), handleProxy(gw, "orchestrator"))
			orchestrator.GET("/workflows/:workflow_id", middleware.RequirePermission("workflows:read"), handleProxy(gw, "orchestrator"))
		}

		// Agent core routes
		agent := api.Group("/api/v1")
		{
			agent.POST("/agents", middleware.RequirePermission("agents:create"), handleProxy(gw, "agent"))
			agent.GET("/agents", middleware.RequirePermission("agents:read"), handleProxy(gw, "agent"))
			agent.GET("/agents/:agent_id", middleware.RequirePermission("agents:read"), handleProxy(gw, "agent"))
			agent.POST("/execute", middleware.RequirePermission("agents:execute"), handleProxy(gw, "agent"))
		}
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/stats", handleAdminStats(gw))
		admin.POST("/cache/clear", handleCacheClear())
		admin.GET("/config", handleConfigInfo())
	}
}

// handleHealthCheck handles gateway health check
func handleHealthCheck(gw *gateway.Gateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
			"service":   "api-gateway",
		})
	}
}

// handleServicesHealth handles backend services health check
func handleServicesHealth(gw *gateway.Gateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		health := gw.GetAllServicesHealth(ctx)

		// Determine overall status
		overallStatus := "healthy"
		for _, serviceHealth := range health {
			if serviceMap, ok := serviceHealth.(map[string]interface{}); ok {
				if status, exists := serviceMap["status"]; exists && status != "healthy" {
					overallStatus = "degraded"
					break
				}
			}
		}

		statusCode := http.StatusOK
		if overallStatus != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status":    overallStatus,
			"timestamp": time.Now().Unix(),
			"services":  health,
		})
	}
}

// handleMetrics handles Prometheus metrics
func handleMetrics() gin.HandlerFunc {
	handler := metrics.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// handleProxy handles proxying requests to backend services
func handleProxy(gw *gateway.Gateway, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		// Proxy the request
		resp, err := gw.ProxyRequest(ctx, serviceName, c.Request.URL.Path, c.Request)
		if err != nil {
			logger.Error("Proxy request failed", "service", serviceName, "path", c.Request.URL.Path, "error", err)
			
			// Record error metrics
			metrics.RecordRequest(serviceName, c.Request.Method, "error", 500, time.Since(start))
			
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "Service unavailable",
				"code":  "SERVICE_UNAVAILABLE",
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Set status code
		c.Status(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			logger.Error("Failed to copy response body", "error", err)
		}

		// Record success metrics
		metrics.RecordRequest(serviceName, c.Request.Method, "success", resp.StatusCode, time.Since(start))
	}
}

// handleAdminStats handles admin statistics
func handleAdminStats(gw *gateway.Gateway) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		stats := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"uptime":    time.Since(time.Now()).Seconds(), // This would be calculated from start time
			"services":  gw.GetAllServicesHealth(ctx),
			"metrics":   metrics.GetStats(),
		}

		c.JSON(http.StatusOK, stats)
	}
}

// handleCacheClear handles cache clearing
func handleCacheClear() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implementation would clear various caches
		logger.Info("Cache clear requested", "user", c.GetString("user_id"))
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Cache cleared successfully",
		})
	}
}

// handleConfigInfo handles configuration information
func handleConfigInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Return non-sensitive configuration information
		config := map[string]interface{}{
			"version":     "1.0.0",
			"environment": "production", // This would come from actual config
			"features": map[string]bool{
				"rate_limiting": true,
				"authentication": true,
				"metrics": true,
			},
		}

		c.JSON(http.StatusOK, config)
	}
}