package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Service represents a registered service
type Service struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Address       string            `json:"address"`
	Port          int               `json:"port"`
	Tags          []string          `json:"tags"`
	Meta          map[string]string `json:"meta"`
	Health        HealthStatus      `json:"health"`
	RegisteredAt  time.Time         `json:"registered_at"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	TTL           time.Duration     `json:"ttl"`
}

// HealthStatus represents service health status
type HealthStatus struct {
	Status      string    `json:"status"` // healthy, unhealthy, critical
	Message     string    `json:"message"`
	LastChecked time.Time `json:"last_checked"`
	CheckCount  int       `json:"check_count"`
	FailCount   int       `json:"fail_count"`
}

// ServiceRegistry manages service registration and discovery
type ServiceRegistry struct {
	services map[string]*Service
	mu       sync.RWMutex
	logger   *MockLogger
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(logger *MockLogger) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]*Service),
		logger:   logger,
	}
}

// Register registers a new service
func (sr *ServiceRegistry) Register(service *Service) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	service.RegisteredAt = time.Now()
	service.LastHeartbeat = time.Now()
	service.Health = HealthStatus{
		Status:      "healthy",
		Message:     "Service registered",
		LastChecked: time.Now(),
		CheckCount:  0,
		FailCount:   0,
	}

	if service.TTL == 0 {
		service.TTL = 30 * time.Second
	}

	sr.services[service.ID] = service
	sr.logger.Info("Service registered",
		"service_id", service.ID,
		"name", service.Name,
		"address", service.Address,
		"port", service.Port,
	)

	return nil
}

// Deregister removes a service from the registry
func (sr *ServiceRegistry) Deregister(serviceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service, exists := sr.services[serviceID]; exists {
		delete(sr.services, serviceID)
		sr.logger.Info("Service deregistered",
			"service_id", serviceID,
			"name", service.Name,
		)
	}

	return nil
}

// Discover finds services by name
func (sr *ServiceRegistry) Discover(name string) ([]*Service, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var services []*Service
	for _, service := range sr.services {
		if service.Name == name && service.Health.Status == "healthy" {
			services = append(services, service)
		}
	}

	return services, nil
}

// GetService gets a service by ID
func (sr *ServiceRegistry) GetService(serviceID string) (*Service, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	if service, exists := sr.services[serviceID]; exists {
		return service, nil
	}

	return nil, fmt.Errorf("service not found: %s", serviceID)
}

// ListServices lists all registered services
func (sr *ServiceRegistry) ListServices() []*Service {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var services []*Service
	for _, service := range sr.services {
		services = append(services, service)
	}

	return services
}

// UpdateHealth updates service health status
func (sr *ServiceRegistry) UpdateHealth(serviceID string, status HealthStatus) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service, exists := sr.services[serviceID]; exists {
		service.Health = status
		service.LastHeartbeat = time.Now()
		return nil
	}

	return fmt.Errorf("service not found: %s", serviceID)
}

// Heartbeat updates service heartbeat
func (sr *ServiceRegistry) Heartbeat(serviceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if service, exists := sr.services[serviceID]; exists {
		service.LastHeartbeat = time.Now()
		if service.Health.Status != "healthy" {
			service.Health.Status = "healthy"
			service.Health.Message = "Service recovered"
			service.Health.LastChecked = time.Now()
		}
		return nil
	}

	return fmt.Errorf("service not found: %s", serviceID)
}

// CleanupExpiredServices removes services that haven't sent heartbeats
func (sr *ServiceRegistry) CleanupExpiredServices() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	now := time.Now()
	var expiredServices []string

	for id, service := range sr.services {
		if now.Sub(service.LastHeartbeat) > service.TTL {
			expiredServices = append(expiredServices, id)
		}
	}

	for _, id := range expiredServices {
		service := sr.services[id]
		delete(sr.services, id)
		sr.logger.Info("Service expired and removed",
			"service_id", id,
			"name", service.Name,
			"last_heartbeat", service.LastHeartbeat,
		)
	}
}

// HealthChecker performs health checks on registered services
type HealthChecker struct {
	registry *ServiceRegistry
	logger   *MockLogger
	interval time.Duration
	timeout  time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(registry *ServiceRegistry, logger *MockLogger) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		logger:   logger,
		interval: 30 * time.Second,
		timeout:  5 * time.Second,
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllServices()
			hc.registry.CleanupExpiredServices()
		}
	}
}

// checkAllServices checks health of all registered services
func (hc *HealthChecker) checkAllServices() {
	services := hc.registry.ListServices()

	for _, service := range services {
		go hc.checkServiceHealth(service)
	}
}

// checkServiceHealth checks health of a single service
func (hc *HealthChecker) checkServiceHealth(service *Service) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	// Mock health check - in real implementation, this would make HTTP requests
	// to the service's health endpoint

	status := HealthStatus{
		Status:      "healthy",
		Message:     "Health check passed",
		LastChecked: time.Now(),
		CheckCount:  service.Health.CheckCount + 1,
		FailCount:   service.Health.FailCount,
	}

	// Simulate occasional failures for testing
	if time.Now().Unix()%10 == 0 {
		status.Status = "unhealthy"
		status.Message = "Health check failed"
		status.FailCount = service.Health.FailCount + 1
	}

	hc.registry.UpdateHealth(service.ID, status)

	// Cancel context
	_ = ctx
}

// LoadBalancer provides load balancing for service discovery
type LoadBalancer struct {
	registry *ServiceRegistry
	logger   *MockLogger
	strategy string // round_robin, least_connections, random
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(registry *ServiceRegistry, logger *MockLogger) *LoadBalancer {
	return &LoadBalancer{
		registry: registry,
		logger:   logger,
		strategy: "round_robin",
	}
}

// SelectService selects a service instance based on load balancing strategy
func (lb *LoadBalancer) SelectService(serviceName string) (*Service, error) {
	services, err := lb.registry.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy services found for: %s", serviceName)
	}

	switch lb.strategy {
	case "round_robin":
		return lb.roundRobin(services), nil
	case "random":
		return lb.random(services), nil
	case "least_connections":
		return lb.leastConnections(services), nil
	default:
		return services[0], nil
	}
}

// roundRobin implements round-robin load balancing
func (lb *LoadBalancer) roundRobin(services []*Service) *Service {
	// Simple round-robin based on current time
	index := int(time.Now().Unix()) % len(services)
	return services[index]
}

// random implements random load balancing
func (lb *LoadBalancer) random(services []*Service) *Service {
	index := int(time.Now().UnixNano()) % len(services)
	return services[index]
}

// leastConnections implements least connections load balancing
func (lb *LoadBalancer) leastConnections(services []*Service) *Service {
	// Mock implementation - in real scenario, track active connections
	return services[0]
}
