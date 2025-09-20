package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/multi-agent/go/orchestrator/internal/budget"
	"github.com/multi-agent/go/orchestrator/internal/database"
	"github.com/multi-agent/go/orchestrator/internal/redis"
	"github.com/multi-agent/go/orchestrator/internal/temporal"
)

func main() {
	// Initialize logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Multi-Agent Orchestrator")

	// Load configuration
	config := loadConfig()

	// Initialize database
	dbClient, err := database.NewClient(config.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer dbClient.Close()

	// Initialize Redis
	redisClient, err := redis.NewClient(config.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize budget manager
	budgetManager := budget.NewManager(logger, dbClient, redisClient)

	// Initialize Temporal worker
	temporalWorker, err := temporal.NewWorker(config.Temporal, logger, budgetManager)
	if err != nil {
		logger.Fatal("Failed to initialize Temporal worker", zap.Error(err))
	}

	// Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Temporal worker
	if err := temporalWorker.Start(ctx); err != nil {
		logger.Fatal("Failed to start Temporal worker", zap.Error(err))
	}
	defer temporalWorker.Stop()

	logger.Info("Multi-Agent Orchestrator started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutdown signal received, gracefully shutting down...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel() // Cancel main context

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded")
	case <-time.After(5 * time.Second):
		logger.Info("Graceful shutdown completed")
	}
}

// Config holds all configuration
type Config struct {
	Database database.Config `yaml:"database"`
	Redis    redis.Config    `yaml:"redis"`
	Temporal temporal.Config `yaml:"temporal"`
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	return Config{
		Database: database.Config{
			Host:         getEnv("POSTGRES_HOST", "localhost"),
			Port:         getEnvInt("POSTGRES_PORT", 5432),
			Database:     getEnv("POSTGRES_DB", "multi_agent"),
			Username:     getEnv("POSTGRES_USER", "multi_agent"),
			Password:     getEnv("POSTGRES_PASSWORD", "multi_agent_password"),
			SSLMode:      getEnv("POSTGRES_SSL_MODE", "disable"),
			MaxOpenConns: getEnvInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("POSTGRES_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnv("POSTGRES_MAX_LIFETIME", "5m"),
		},
		Redis: redis.Config{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getEnvInt("REDIS_DB", 0),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
			DialTimeout:  getEnv("REDIS_DIAL_TIMEOUT", "5s"),
			ReadTimeout:  getEnv("REDIS_READ_TIMEOUT", "3s"),
			WriteTimeout: getEnv("REDIS_WRITE_TIMEOUT", "3s"),
		},
		Temporal: temporal.Config{
			HostPort:   getEnv("TEMPORAL_HOST_PORT", "localhost:7233"),
			Namespace:  getEnv("TEMPORAL_NAMESPACE", "default"),
			TaskQueue:  getEnv("TEMPORAL_TASK_QUEUE", "multi-agent-task-queue"),
			WorkerName: getEnv("TEMPORAL_WORKER_NAME", "multi-agent-worker"),
		},
	}
}

// initLogger initializes the logger
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	return logger
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}