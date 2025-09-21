package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIHandler handles HTTP requests for service registry
type APIHandler struct {
	registry      *ServiceRegistry
	healthChecker *HealthChecker
	loadBalancer  *LoadBalancer
	logger        *MockLogger
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(registry *ServiceRegistry, healthChecker *HealthChecker, loadBalancer *LoadBalancer, logger *MockLogger) *APIHandler {
	return &APIHandler{
		registry:      registry,
		healthChecker: healthChecker,
		loadBalancer:  loadBalancer,
		logger:        logger,
	}
}

// SetupRoutes sets up the API routes
func (h *APIHandler) SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", h.healthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Service registration and discovery
		services := v1.Group("/services")
		{
			services.GET("", h.listServices)
			services.POST("/register", h.registerService)
			services.DELETE("/:id", h.deregisterService)
			services.POST("/:id/heartbeat", h.serviceHeartbeat)
			services.POST("/:id/health", h.updateServiceHealth)
			services.GET("/discover/:name", h.discoverService)
			services.GET("/select/:name", h.selectService)
		}

		// Health monitoring
		health := v1.Group("/health")
		{
			health.GET("/services", h.getServicesHealth)
			health.GET("/services/:id", h.getServiceHealth)
		}

		// Load balancing
		lb := v1.Group("/loadbalancer")
		{
			lb.GET("/strategies", h.getLoadBalancingStrategies)
			lb.POST("/strategy", h.setLoadBalancingStrategy)
		}

		// Registry management
		registry := v1.Group("/registry")
		{
			registry.GET("/stats", h.getRegistryStats)
			registry.POST("/cleanup", h.cleanupExpiredServices)
		}
	}

	// Static files for Web UI
	router.Static("/ui", "./web/dist")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/ui")
	})
}

// Health check endpoint
func (h *APIHandler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "service-registry",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    "24h",
	})
}

// Service registration and discovery handlers

func (h *APIHandler) listServices(c *gin.Context) {
	services := h.registry.ListServices()

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

func (h *APIHandler) registerService(c *gin.Context) {
	var request struct {
		Name    string            `json:"name" binding:"required"`
		Address string            `json:"address" binding:"required"`
		Port    int               `json:"port" binding:"required"`
		Tags    []string          `json:"tags"`
		Meta    map[string]string `json:"meta"`
		TTL     int               `json:"ttl"` // seconds
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := &Service{
		ID:      generateServiceID(),
		Name:    request.Name,
		Address: request.Address,
		Port:    request.Port,
		Tags:    request.Tags,
		Meta:    request.Meta,
		TTL:     time.Duration(request.TTL) * time.Second,
	}

	if service.TTL == 0 {
		service.TTL = 30 * time.Second
	}

	if err := h.registry.Register(service); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"service": service,
		"message": "Service registered successfully",
	})
}

func (h *APIHandler) deregisterService(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.registry.Deregister(serviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service deregistered successfully",
	})
}

func (h *APIHandler) serviceHeartbeat(c *gin.Context) {
	serviceID := c.Param("id")

	if err := h.registry.Heartbeat(serviceID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Heartbeat received",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *APIHandler) updateServiceHealth(c *gin.Context) {
	serviceID := c.Param("id")

	var request struct {
		Status  string `json:"status" binding:"required"`
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	healthStatus := HealthStatus{
		Status:      request.Status,
		Message:     request.Message,
		LastChecked: time.Now(),
	}

	if err := h.registry.UpdateHealth(serviceID, healthStatus); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Health status updated successfully",
		"health":  healthStatus,
	})
}

func (h *APIHandler) discoverService(c *gin.Context) {
	serviceName := c.Param("name")

	services, err := h.registry.Discover(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"services":     services,
		"count":        len(services),
	})
}

func (h *APIHandler) selectService(c *gin.Context) {
	serviceName := c.Param("name")

	service, err := h.loadBalancer.SelectService(serviceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"selected":     service,
		"strategy":     h.loadBalancer.strategy,
	})
}

// Health monitoring handlers

func (h *APIHandler) getServicesHealth(c *gin.Context) {
	services := h.registry.ListServices()

	healthySrvs := 0
	unhealthySrvs := 0
	criticalSrvs := 0

	for _, service := range services {
		switch service.Health.Status {
		case "healthy":
			healthySrvs++
		case "unhealthy":
			unhealthySrvs++
		case "critical":
			criticalSrvs++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"total":     len(services),
			"healthy":   healthySrvs,
			"unhealthy": unhealthySrvs,
			"critical":  criticalSrvs,
		},
		"services": services,
	})
}

func (h *APIHandler) getServiceHealth(c *gin.Context) {
	serviceID := c.Param("id")

	service, err := h.registry.GetService(serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_id": serviceID,
		"health":     service.Health,
		"heartbeat":  service.LastHeartbeat,
	})
}

// Load balancing handlers

func (h *APIHandler) getLoadBalancingStrategies(c *gin.Context) {
	strategies := []string{
		"round_robin",
		"random",
		"least_connections",
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies":       strategies,
		"current_strategy": h.loadBalancer.strategy,
	})
}

func (h *APIHandler) setLoadBalancingStrategy(c *gin.Context) {
	var request struct {
		Strategy string `json:"strategy" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validStrategies := map[string]bool{
		"round_robin":       true,
		"random":            true,
		"least_connections": true,
	}

	if !validStrategies[request.Strategy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid strategy. Valid strategies: round_robin, random, least_connections",
		})
		return
	}

	h.loadBalancer.strategy = request.Strategy

	c.JSON(http.StatusOK, gin.H{
		"message":  "Load balancing strategy updated",
		"strategy": request.Strategy,
	})
}

// Registry management handlers

func (h *APIHandler) getRegistryStats(c *gin.Context) {
	services := h.registry.ListServices()

	stats := gin.H{
		"total_services": len(services),
		"by_status":      make(map[string]int),
		"by_name":        make(map[string]int),
		"by_tags":        make(map[string]int),
	}

	statusMap := stats["by_status"].(map[string]int)
	nameMap := stats["by_name"].(map[string]int)
	tagMap := stats["by_tags"].(map[string]int)

	for _, service := range services {
		// Count by status
		statusMap[service.Health.Status]++

		// Count by name
		nameMap[service.Name]++

		// Count by tags
		for _, tag := range service.Tags {
			tagMap[tag]++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *APIHandler) cleanupExpiredServices(c *gin.Context) {
	h.registry.CleanupExpiredServices()

	c.JSON(http.StatusOK, gin.H{
		"message":   "Expired services cleanup completed",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Helper functions

func generateServiceID() string {
	return fmt.Sprintf("svc_%d", time.Now().UnixNano())
}
