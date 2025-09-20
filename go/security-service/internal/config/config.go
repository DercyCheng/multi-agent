package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the security service
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	OPA      OPAConfig      `yaml:"opa"`
	Redis    RedisConfig    `yaml:"redis"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int    `yaml:"port" env:"PORT" default:"8083"`
	Mode         string `yaml:"mode" env:"GIN_MODE" default:"debug"`
	ReadTimeout  int    `yaml:"read_timeout" env:"READ_TIMEOUT" default:"30"`
	WriteTimeout int    `yaml:"write_timeout" env:"WRITE_TIMEOUT" default:"30"`
	IdleTimeout  int    `yaml:"idle_timeout" env:"IDLE_TIMEOUT" default:"120"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host         string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port         int    `yaml:"port" env:"DB_PORT" default:"5432"`
	Database     string `yaml:"database" env:"DB_NAME" default:"multiagent"`
	Username     string `yaml:"username" env:"DB_USER" default:"postgres"`
	Password     string `yaml:"password" env:"DB_PASSWORD" default:""`
	SSLMode      string `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"5"`
	MaxLifetime  string `yaml:"max_lifetime" env:"DB_MAX_LIFETIME" default:"5m"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey        string        `yaml:"secret_key" env:"JWT_SECRET_KEY"`
	AccessTokenTTL   time.Duration `yaml:"access_token_ttl" env:"JWT_ACCESS_TTL" default:"15m"`
	RefreshTokenTTL  time.Duration `yaml:"refresh_token_ttl" env:"JWT_REFRESH_TTL" default:"7d"`
	Issuer           string        `yaml:"issuer" env:"JWT_ISSUER" default:"multi-agent-platform"`
	Audience         string        `yaml:"audience" env:"JWT_AUDIENCE" default:"multi-agent-users"`
}

// OPAConfig holds Open Policy Agent configuration
type OPAConfig struct {
	ServerURL    string `yaml:"server_url" env:"OPA_SERVER_URL" default:"http://localhost:8181"`
	PolicyPath   string `yaml:"policy_path" env:"OPA_POLICY_PATH" default:"/v1/data/multiagent"`
	BundlePath   string `yaml:"bundle_path" env:"OPA_BUNDLE_PATH" default:"./policies"`
	EnableBundle bool   `yaml:"enable_bundle" env:"OPA_ENABLE_BUNDLE" default:"true"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `yaml:"host" env:"REDIS_HOST" default:"localhost"`
	Port         int    `yaml:"port" env:"REDIS_PORT" default:"6379"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD" default:""`
	Database     int    `yaml:"database" env:"REDIS_DB" default:"1"`
	PoolSize     int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int    `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PORT", 8083),
			Mode:         getEnvString("GIN_MODE", "debug"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:         getEnvString("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			Database:     getEnvString("DB_NAME", "multiagent"),
			Username:     getEnvString("DB_USER", "postgres"),
			Password:     getEnvString("DB_PASSWORD", ""),
			SSLMode:      getEnvString("DB_SSL_MODE", "disable"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvString("DB_MAX_LIFETIME", "5m"),
		},
		JWT: JWTConfig{
			SecretKey:       getEnvString("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTokenTTL: getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
			Issuer:          getEnvString("JWT_ISSUER", "multi-agent-platform"),
			Audience:        getEnvString("JWT_AUDIENCE", "multi-agent-users"),
		},
		OPA: OPAConfig{
			ServerURL:    getEnvString("OPA_SERVER_URL", "http://localhost:8181"),
			PolicyPath:   getEnvString("OPA_POLICY_PATH", "/v1/data/multiagent"),
			BundlePath:   getEnvString("OPA_BUNDLE_PATH", "./policies"),
			EnableBundle: getEnvBool("OPA_ENABLE_BUNDLE", true),
		},
		Redis: RedisConfig{
			Host:         getEnvString("REDIS_HOST", "localhost"),
			Port:         getEnvInt("REDIS_PORT", 6379),
			Password:     getEnvString("REDIS_PASSWORD", ""),
			Database:     getEnvInt("REDIS_DB", 1),
			PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 5),
		},
	}

	// Validate required fields
	if config.JWT.SecretKey == "your-secret-key-change-in-production" {
		return nil, fmt.Errorf("JWT_SECRET_KEY must be set in production")
	}

	return config, nil
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}