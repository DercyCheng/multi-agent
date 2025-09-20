package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/multi-agent/api-gateway/internal/config"
	"github.com/multi-agent/api-gateway/pkg/logger"
	"github.com/multi-agent/api-gateway/pkg/metrics"
)

// Gateway represents the API gateway
type Gateway struct {
	config   *config.Config
	client   *http.Client
	services map[string]*ServiceProxy
}

// ServiceProxy represents a backend service proxy
type ServiceProxy struct {
	Name    string
	BaseURL *url.URL
	Client  *http.Client
	Retries int
}

// New creates a new gateway instance
func New(cfg *config.Config) (*Gateway, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Create service proxies
	services := make(map[string]*ServiceProxy)

	// Orchestrator service
	orchestratorURL, err := url.Parse(cfg.Services.Orchestrator.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid orchestrator URL: %w", err)
	}
	services["orchestrator"] = &ServiceProxy{
		Name:    "orchestrator",
		BaseURL: orchestratorURL,
		Client: &http.Client{
			Timeout: cfg.Services.Orchestrator.Timeout,
		},
		Retries: cfg.Services.Orchestrator.Retries,
	}

	// LLM service
	llmURL, err := url.Parse(cfg.Services.LLMService.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM service URL: %w", err)
	}
	services["llm"] = &ServiceProxy{
		Name:    "llm",
		BaseURL: llmURL,
		Client: &http.Client{
			Timeout: cfg.Services.LLMService.Timeout,
		},
		Retries: cfg.Services.LLMService.Retries,
	}

	// Agent core service
	agentURL, err := url.Parse(cfg.Services.AgentCore.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid agent core URL: %w", err)
	}
	services["agent"] = &ServiceProxy{
		Name:    "agent",
		BaseURL: agentURL,
		Client: &http.Client{
			Timeout: cfg.Services.AgentCore.Timeout,
		},
		Retries: cfg.Services.AgentCore.Retries,
	}

	return &Gateway{
		config:   cfg,
		client:   client,
		services: services,
	}, nil
}

// ProxyRequest proxies a request to the appropriate backend service
func (g *Gateway) ProxyRequest(ctx context.Context, serviceName, path string, req *http.Request) (*http.Response, error) {
	service, exists := g.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// Build target URL
	targetURL := *service.BaseURL
	targetURL.Path = strings.TrimPrefix(path, fmt.Sprintf("/%s", serviceName))
	targetURL.RawQuery = req.URL.RawQuery

	// Create new request
	var body io.Reader
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // Reset original body
	}

	proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy request: %w", err)
	}

	// Copy headers
	for key, values := range req.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Add gateway headers
	proxyReq.Header.Set("X-Forwarded-For", req.RemoteAddr)
	proxyReq.Header.Set("X-Gateway-Service", serviceName)

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= service.Retries; attempt++ {
		if attempt > 0 {
			logger.Warn("Retrying request", "service", serviceName, "attempt", attempt, "error", lastErr)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		start := time.Now()
		resp, lastErr = service.Client.Do(proxyReq)
		duration := time.Since(start)

		// Record metrics
		status := "success"
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		if lastErr != nil || (resp != nil && resp.StatusCode >= 500) {
			status = "error"
		}

		metrics.RecordRequest(serviceName, req.Method, status, statusCode, duration)

		if lastErr == nil && resp.StatusCode < 500 {
			break // Success or client error (don't retry)
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", service.Retries, lastErr)
	}

	return resp, nil
}

// GetServiceHealth checks the health of a backend service
func (g *Gateway) GetServiceHealth(ctx context.Context, serviceName string) (map[string]interface{}, error) {
	service, exists := g.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// Build health check URL
	healthURL := *service.BaseURL
	healthURL.Path = "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := service.Client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"error":   err.Error(),
			"service": serviceName,
		}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]interface{}{
			"status":  "unhealthy",
			"error":   "failed to read response",
			"service": serviceName,
		}, nil
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		return map[string]interface{}{
			"status":     "unhealthy",
			"error":      "invalid response format",
			"service":    serviceName,
			"raw_body":   string(body),
			"status_code": resp.StatusCode,
		}, nil
	}

	health["service"] = serviceName
	health["status_code"] = resp.StatusCode

	return health, nil
}

// GetAllServicesHealth checks the health of all backend services
func (g *Gateway) GetAllServicesHealth(ctx context.Context) map[string]interface{} {
	results := make(map[string]interface{})

	for serviceName := range g.services {
		health, err := g.GetServiceHealth(ctx, serviceName)
		if err != nil {
			results[serviceName] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			results[serviceName] = health
		}
	}

	return results
}

// RouteRequest determines which service should handle the request
func (g *Gateway) RouteRequest(path string) (string, error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	
	// Split path into segments
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return "", fmt.Errorf("invalid path")
	}

	// Route based on path prefix
	switch {
	case strings.HasPrefix(path, "v1/chat") || strings.HasPrefix(path, "v1/models"):
		return "llm", nil
	case strings.HasPrefix(path, "api/v1/tasks") || strings.HasPrefix(path, "api/v1/workflows"):
		return "orchestrator", nil
	case strings.HasPrefix(path, "api/v1/agents") || strings.HasPrefix(path, "api/v1/execute"):
		return "agent", nil
	case strings.HasPrefix(path, "health"):
		return "gateway", nil // Handle health checks locally
	case strings.HasPrefix(path, "metrics"):
		return "gateway", nil // Handle metrics locally
	default:
		return "", fmt.Errorf("no route found for path: %s", path)
	}
}

// Close closes the gateway and cleans up resources
func (g *Gateway) Close() error {
	// Close HTTP clients
	for _, service := range g.services {
		if transport, ok := service.Client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	if transport, ok := g.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}

	return nil
}