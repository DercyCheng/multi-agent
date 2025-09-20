package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/multi-agent/api-gateway/internal/config"
)

var (
	// Request metrics
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests processed by the gateway",
		},
		[]string{"service", "method", "status", "status_code"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	activeConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Number of active connections",
		},
		[]string{"service"},
	)

	// Service health metrics
	serviceHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_service_health",
			Help: "Health status of backend services (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service"},
	)

	// Rate limiting metrics
	rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"user_id", "endpoint"},
	)

	// Authentication metrics
	authAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)
)

// Init initializes metrics
func Init(cfg config.MetricsConfig) {
	if !cfg.Enabled {
		return
	}

	// Register metrics
	prometheus.MustRegister(
		requestsTotal,
		requestDuration,
		activeConnections,
		serviceHealth,
		rateLimitHits,
		authAttempts,
	)
}

// RecordRequest records request metrics
func RecordRequest(service, method, status string, statusCode int, duration time.Duration) {
	requestsTotal.WithLabelValues(service, method, status, strconv.Itoa(statusCode)).Inc()
	requestDuration.WithLabelValues(service, method, status).Observe(duration.Seconds())
}

// RecordServiceHealth records service health status
func RecordServiceHealth(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	serviceHealth.WithLabelValues(service).Set(value)
}

// RecordRateLimitHit records rate limit hit
func RecordRateLimitHit(userID, endpoint string) {
	rateLimitHits.WithLabelValues(userID, endpoint).Inc()
}

// RecordAuthAttempt records authentication attempt
func RecordAuthAttempt(method, status string) {
	authAttempts.WithLabelValues(method, status).Inc()
}

// IncActiveConnections increments active connections
func IncActiveConnections(service string) {
	activeConnections.WithLabelValues(service).Inc()
}

// DecActiveConnections decrements active connections
func DecActiveConnections(service string) {
	activeConnections.WithLabelValues(service).Dec()
}

// Handler returns Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// GetStats returns current metrics statistics
func GetStats() map[string]interface{} {
	// This would typically gather current metric values
	// For now, return a simple structure
	return map[string]interface{}{
		"metrics_enabled": true,
		"collectors": []string{
			"requests_total",
			"request_duration",
			"active_connections",
			"service_health",
			"rate_limit_hits",
			"auth_attempts",
		},
	}
}