package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `json:"server"`
	Auth      AuthConfig      `json:"auth"`
	Services  ServicesConfig  `json:"services"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	CORS      CORSConfig      `json:"cors"`
	Logging   LoggingConfig   `json:"logging"`
	Metrics   MetricsConfig   `json:"metrics"`
	Security  SecurityConfig  `json:"security"`
}

type ServerConfig struct {
	Port         int    `json:"port"`
	Mode         string `json:"mode"`
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
	IdleTimeout  int    `json:"idle_timeout"`
}

type AuthConfig struct {
	JWTSecret     string        `json:"jwt_secret"`
	JWTExpiration time.Duration `json:"jwt_expiration"`
	APIKeyHeader  string        `json:"api_key_header"`
	RedisURL      string        `json:"redis_url"`
}

type ServicesConfig struct {
	Orchestrator ServiceEndpoint `json:"orchestrator"`
	LLMService   ServiceEndpoint `json:"llm_service"`
	AgentCore    ServiceEndpoint `json:"agent_core"`
}

type ServiceEndpoint struct {
	URL     string        `json:"url"`
	Timeout time.Duration `json:"timeout"`
	Retries int           `json:"retries"`
}

type RateLimitConfig struct {
	Enabled        bool          `json:"enabled"`
	RequestsPerMin int           `json:"requests_per_min"`
	BurstSize      int           `json:"burst_size"`
	CleanupPeriod  time.Duration `json:"cleanup_period"`
}

type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	Output string `json:"output"`
}

type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port"`
	Path    string `json:"path"`
}

type SecurityConfig struct {
	EnableHTTPS       bool     `json:"enable_https"`
	CertFile          string   `json:"cert_file"`
	KeyFile           string   `json:"key_file"`
	TrustedProxies    []string `json:"trusted_proxies"`
	MaxRequestSize    int64    `json:"max_request_size"`
	RequestTimeout    int      `json:"request_timeout"`
	EnableCSRF        bool     `json:"enable_csrf"`
	CSRFSecret        string   `json:"csrf_secret"`
	ContentTypeNoSniff bool    `json:"content_type_no_sniff"`
	FrameOptions      string   `json:"frame_options"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnvInt("GATEWAY_PORT", 8080),
			Mode:         getEnv("GATEWAY_MODE", "development"),
			ReadTimeout:  getEnvInt("GATEWAY_READ_TIMEOUT", 30),
			WriteTimeout: getEnvInt("GATEWAY_WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvInt("GATEWAY_IDLE_TIMEOUT", 120),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpiration: time.Duration(getEnvInt("JWT_EXPIRATION", 3600)) * time.Second,
			APIKeyHeader:  getEnv("API_KEY_HEADER", "X-API-Key"),
			RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/4"),
		},
		Services: ServicesConfig{
			Orchestrator: ServiceEndpoint{
				URL:     getEnv("ORCHESTRATOR_URL", "http://localhost:8081"),
				Timeout: time.Duration(getEnvInt("ORCHESTRATOR_TIMEOUT", 30)) * time.Second,
				Retries: getEnvInt("ORCHESTRATOR_RETRIES", 3),
			},
			LLMService: ServiceEndpoint{
				URL:     getEnv("LLM_SERVICE_URL", "http://localhost:8000"),
				Timeout: time.Duration(getEnvInt("LLM_SERVICE_TIMEOUT", 60)) * time.Second,
				Retries: getEnvInt("LLM_SERVICE_RETRIES", 3),
			},
			AgentCore: ServiceEndpoint{
				URL:     getEnv("AGENT_CORE_URL", "http://localhost:8082"),
				Timeout: time.Duration(getEnvInt("AGENT_CORE_TIMEOUT", 30)) * time.Second,
				Retries: getEnvInt("AGENT_CORE_RETRIES", 3),
			},
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin: getEnvInt("RATE_LIMIT_REQUESTS_PER_MIN", 100),
			BurstSize:      getEnvInt("RATE_LIMIT_BURST_SIZE", 10),
			CleanupPeriod:  time.Duration(getEnvInt("RATE_LIMIT_CLEANUP_PERIOD", 300)) * time.Second,
		},
		CORS: CORSConfig{
			AllowOrigins:     getEnvSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
			AllowMethods:     getEnvSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowHeaders:     getEnvSlice("CORS_ALLOW_HEADERS", []string{"*"}),
			ExposeHeaders:    getEnvSlice("CORS_EXPOSE_HEADERS", []string{}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:           getEnvInt("CORS_MAX_AGE", 86400),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("METRICS_ENABLED", true),
			Port:    getEnvInt("METRICS_PORT", 9090),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		Security: SecurityConfig{
			EnableHTTPS:       getEnvBool("ENABLE_HTTPS", false),
			CertFile:          getEnv("CERT_FILE", ""),
			KeyFile:           getEnv("KEY_FILE", ""),
			TrustedProxies:    getEnvSlice("TRUSTED_PROXIES", []string{}),
			MaxRequestSize:    int64(getEnvInt("MAX_REQUEST_SIZE", 10485760)), // 10MB
			RequestTimeout:    getEnvInt("REQUEST_TIMEOUT", 30),
			EnableCSRF:        getEnvBool("ENABLE_CSRF", false),
			CSRFSecret:        getEnv("CSRF_SECRET", ""),
			ContentTypeNoSniff: getEnvBool("CONTENT_TYPE_NO_SNIFF", true),
			FrameOptions:      getEnv("FRAME_OPTIONS", "DENY"),
		},
	}

	return cfg, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
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

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}