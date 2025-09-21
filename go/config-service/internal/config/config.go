package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	Environment string         `yaml:"environment" env:"ENVIRONMENT" default:"development"`
	Server      ServerConfig   `yaml:"server"`
	Database    DatabaseConfig `yaml:"database"`
	Redis       RedisConfig    `yaml:"redis"`
	Auth        AuthConfig     `yaml:"auth"`
	Metrics     MetricsConfig  `yaml:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int           `yaml:"port" env:"PORT" default:"8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"10s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" default:"60s"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver" env:"DB_DRIVER" default:"postgres"`
	Host     string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port     int    `yaml:"port" env:"DB_PORT" default:"5432"`
	Database string `yaml:"database" env:"DB_NAME" default:"config_service"`
	Username string `yaml:"username" env:"DB_USER" default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" default:""`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSLMODE" default:"disable"`
	MaxConns int    `yaml:"max_conns" env:"DB_MAX_CONNS" default:"10"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string        `yaml:"host" env:"REDIS_HOST" default:"localhost"`
	Port     int           `yaml:"port" env:"REDIS_PORT" default:"6379"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD" default:""`
	DB       int           `yaml:"db" env:"REDIS_DB" default:"0"`
	PoolSize int           `yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	Timeout  time.Duration `yaml:"timeout" default:"5s"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string        `yaml:"jwt_secret" env:"JWT_SECRET"`
	TokenExpiry   time.Duration `yaml:"token_expiry" default:"24h"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry" default:"168h"` // 7 days
	Issuer        string        `yaml:"issuer" default:"config-service"`
	EnableAPIKeys bool          `yaml:"enable_api_keys" default:"true"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED" default:"true"`
	Port    int    `yaml:"port" env:"METRICS_PORT" default:"9090"`
	Path    string `yaml:"path" default:"/metrics"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	config := &Config{}

	// Load from YAML file if exists
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		data, err := os.ReadFile(configFile)
		if err == nil {
			yaml.Unmarshal(data, config)
		}
	}

	// Apply defaults and environment variables
	applyDefaults(config)
	applyEnvVars(config)

	return config, nil
}

// applyDefaults applies default values
func applyDefaults(config *Config) {
	if config.Environment == "" {
		config.Environment = "development"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 10 * time.Second
	}

	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 10 * time.Second
	}

	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

	if config.Database.Driver == "" {
		config.Database.Driver = "postgres"
	}

	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}

	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}

	if config.Database.Database == "" {
		config.Database.Database = "config_service"
	}

	if config.Database.Username == "" {
		config.Database.Username = "postgres"
	}

	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}

	if config.Database.MaxConns == 0 {
		config.Database.MaxConns = 10
	}

	if config.Redis.Host == "" {
		config.Redis.Host = "localhost"
	}

	if config.Redis.Port == 0 {
		config.Redis.Port = 6379
	}

	if config.Redis.PoolSize == 0 {
		config.Redis.PoolSize = 10
	}

	if config.Redis.Timeout == 0 {
		config.Redis.Timeout = 5 * time.Second
	}

	if config.Auth.TokenExpiry == 0 {
		config.Auth.TokenExpiry = 24 * time.Hour
	}

	if config.Auth.RefreshExpiry == 0 {
		config.Auth.RefreshExpiry = 168 * time.Hour
	}

	if config.Auth.Issuer == "" {
		config.Auth.Issuer = "config-service"
	}

	if config.Metrics.Port == 0 {
		config.Metrics.Port = 9090
	}

	if config.Metrics.Path == "" {
		config.Metrics.Path = "/metrics"
	}
}

// applyEnvVars applies environment variable overrides
func applyEnvVars(config *Config) {
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}

	if db := os.Getenv("DB_NAME"); db != "" {
		config.Database.Database = db
	}

	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.Username = user
	}

	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		config.Database.Password = pass
	}

	if host := os.Getenv("REDIS_HOST"); host != "" {
		config.Redis.Host = host
	}

	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Redis.Port = p
		}
	}

	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
		config.Redis.Password = pass
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		config.Auth.JWTSecret = secret
	}
}
