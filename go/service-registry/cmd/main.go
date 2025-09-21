package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger (mock)
	logger := &MockLogger{}

	// Load configuration (mock)
	cfg := &Config{
		Port:        8081,
		Environment: "development",
	}

	// Initialize service registry
	registry := NewServiceRegistry(logger)

	// Initialize health checker
	healthChecker := NewHealthChecker(registry, logger)
	go healthChecker.Start()

	// Initialize load balancer
	loadBalancer := NewLoadBalancer(registry, logger)

	// Initialize API
	apiHandler := NewAPIHandler(registry, healthChecker, loadBalancer, logger)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	apiHandler.SetupRoutes(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Service Registry",
			"port", cfg.Port,
			"environment", cfg.Environment,
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

// Config holds the application configuration
type Config struct {
	Port        int    `yaml:"port" env:"PORT" default:"8081"`
	Environment string `yaml:"environment" env:"ENVIRONMENT" default:"development"`
}

// MockLogger is a mock logger implementation
type MockLogger struct{}

func (l *MockLogger) Info(msg string, fields ...interface{}) {
	fmt.Printf("INFO: %s %v\n", msg, fields)
}

func (l *MockLogger) Error(msg string, fields ...interface{}) {
	fmt.Printf("ERROR: %s %v\n", msg, fields)
}

func (l *MockLogger) Fatal(msg string, fields ...interface{}) {
	fmt.Printf("FATAL: %s %v\n", msg, fields)
	os.Exit(1)
}
