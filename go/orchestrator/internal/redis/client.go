package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config holds Redis configuration
type Config struct {
	Host         string `yaml:"host" env:"REDIS_HOST" default:"localhost"`
	Port         int    `yaml:"port" env:"REDIS_PORT" default:"6379"`
	Password     string `yaml:"password" env:"REDIS_PASSWORD" default:""`
	Database     int    `yaml:"database" env:"REDIS_DB" default:"0"`
	PoolSize     int    `yaml:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int    `yaml:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	MaxRetries   int    `yaml:"max_retries" env:"REDIS_MAX_RETRIES" default:"3"`
	DialTimeout  string `yaml:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  string `yaml:"read_timeout" env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout string `yaml:"write_timeout" env:"REDIS_WRITE_TIMEOUT" default:"3s"`
}

// Client wraps Redis operations with tenant isolation
type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
}

// NewClient creates a new Redis client
func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	dialTimeout, _ := time.ParseDuration(config.DialTimeout)
	readTimeout, _ := time.ParseDuration(config.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(config.WriteTimeout)

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.Int("database", config.Database),
	)

	return &Client{
		rdb:    rdb,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Health checks Redis connectivity
func (c *Client) Health(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// GetClient returns the underlying Redis client
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// Session management with tenant isolation

// SessionData represents cached session data
type SessionData struct {
	UserID      string                 `json:"user_id"`
	TenantID    string                 `json:"tenant_id"`
	Context     map[string]interface{} `json:"context"`
	TokenBudget int                    `json:"token_budget"`
	TokensUsed  int                    `json:"tokens_used"`
	CostUSD     float64                `json:"cost_usd"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
}

// SetSession stores session data with tenant isolation
func (c *Client) SetSession(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error {
	key := c.sessionKey(sessionID, data.TenantID)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	err = c.rdb.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}
	
	// Also set a tenant-specific index for cleanup
	tenantKey := c.tenantSessionsKey(data.TenantID)
	c.rdb.SAdd(ctx, tenantKey, sessionID)
	c.rdb.Expire(ctx, tenantKey, ttl+time.Hour) // Keep index slightly longer
	
	return nil
}

// GetSession retrieves session data with tenant validation
func (c *Client) GetSession(ctx context.Context, sessionID, tenantID string) (*SessionData, error) {
	key := c.sessionKey(sessionID, tenantID)
	
	jsonData, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	var data SessionData
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	
	// Validate tenant ID
	if data.TenantID != tenantID {
		return nil, fmt.Errorf("session access denied")
	}
	
	return &data, nil
}

// UpdateSession updates session data
func (c *Client) UpdateSession(ctx context.Context, sessionID string, data *SessionData) error {
	key := c.sessionKey(sessionID, data.TenantID)
	
	// Check if session exists
	exists, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("session not found")
	}
	
	data.UpdatedAt = time.Now()
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	// Keep existing TTL
	ttl, err := c.rdb.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get session TTL: %w", err)
	}
	
	err = c.rdb.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	return nil
}

// DeleteSession removes session data
func (c *Client) DeleteSession(ctx context.Context, sessionID, tenantID string) error {
	key := c.sessionKey(sessionID, tenantID)
	
	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	// Remove from tenant index
	tenantKey := c.tenantSessionsKey(tenantID)
	c.rdb.SRem(ctx, tenantKey, sessionID)
	
	return nil
}

// Cache operations for tool results and patterns

// CacheData represents cached data with metadata
type CacheData struct {
	Data      interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time   `json:"created_at"`
	HitCount  int         `json:"hit_count"`
}

// SetCache stores data in cache with tenant isolation
func (c *Client) SetCache(ctx context.Context, key string, tenantID string, data interface{}, ttl time.Duration) error {
	cacheKey := c.cacheKey(key, tenantID)
	
	cacheData := &CacheData{
		Data:      data,
		CreatedAt: time.Now(),
		HitCount:  0,
	}
	
	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}
	
	err = c.rdb.Set(ctx, cacheKey, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	
	return nil
}

// GetCache retrieves data from cache with hit counting
func (c *Client) GetCache(ctx context.Context, key string, tenantID string) (interface{}, error) {
	cacheKey := c.cacheKey(key, tenantID)
	
	jsonData, err := c.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache miss")
		}
		return nil, fmt.Errorf("failed to get cache: %w", err)
	}
	
	var cacheData CacheData
	err = json.Unmarshal([]byte(jsonData), &cacheData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	
	// Increment hit count asynchronously
	go func() {
		cacheData.HitCount++
		if updatedData, err := json.Marshal(cacheData); err == nil {
			ttl, _ := c.rdb.TTL(context.Background(), cacheKey).Result()
			c.rdb.Set(context.Background(), cacheKey, updatedData, ttl)
		}
	}()
	
	return cacheData.Data, nil
}

// Workspace operations for P2P coordination

// PublishToWorkspace publishes data to a topic for P2P coordination
func (c *Client) PublishToWorkspace(ctx context.Context, topic string, tenantID string, data interface{}) error {
	workspaceKey := c.workspaceKey(topic, tenantID)
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace data: %w", err)
	}
	
	// Store data with expiration
	err = c.rdb.Set(ctx, workspaceKey, jsonData, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to workspace: %w", err)
	}
	
	// Notify subscribers
	channelKey := c.workspaceChannelKey(topic, tenantID)
	err = c.rdb.Publish(ctx, channelKey, "data_available").Err()
	if err != nil {
		c.logger.Warn("Failed to notify workspace subscribers", zap.Error(err))
	}
	
	return nil
}

// SubscribeToWorkspace subscribes to workspace topic updates
func (c *Client) SubscribeToWorkspace(ctx context.Context, topic string, tenantID string) *redis.PubSub {
	channelKey := c.workspaceChannelKey(topic, tenantID)
	return c.rdb.Subscribe(ctx, channelKey)
}

// GetFromWorkspace retrieves data from workspace topic
func (c *Client) GetFromWorkspace(ctx context.Context, topic string, tenantID string) (interface{}, error) {
	workspaceKey := c.workspaceKey(topic, tenantID)
	
	jsonData, err := c.rdb.Get(ctx, workspaceKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("workspace data not available")
		}
		return nil, fmt.Errorf("failed to get workspace data: %w", err)
	}
	
	var data interface{}
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace data: %w", err)
	}
	
	return data, nil
}

// Metrics and monitoring

// Budget interface methods

// IncrementCounter increments a counter by value (for budget interface)
func (c *Client) IncrementCounter(ctx context.Context, key string, value int) error {
	return c.rdb.IncrBy(ctx, key, int64(value)).Err()
}

// IncrementMetricCounter increments a counter metric
func (c *Client) IncrementMetricCounter(ctx context.Context, metric string, labels map[string]string) error {
	key := c.metricsKey(metric, labels)
	return c.rdb.Incr(ctx, key).Err()
}

// GetCounter gets counter value (for budget interface)
func (c *Client) GetCounter(ctx context.Context, key string) (int, error) {
	val, err := c.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// SetWithExpiry sets value with expiry (for budget interface)
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiry).Err()
}

// Get gets value (for budget interface)
func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

// SetGauge sets a gauge metric value
func (c *Client) SetGauge(ctx context.Context, metric string, value float64, labels map[string]string) error {
	key := c.metricsKey(metric, labels)
	return c.rdb.Set(ctx, key, value, time.Hour).Err()
}

// RecordHistogram records a histogram value
func (c *Client) RecordHistogram(ctx context.Context, metric string, value float64, labels map[string]string) error {
	key := c.metricsKey(metric, labels)
	timestamp := time.Now().Unix()
	
	// Use sorted set to store histogram values with timestamps
	return c.rdb.ZAdd(ctx, key, redis.Z{
		Score:  float64(timestamp),
		Member: value,
	}).Err()
}

// Cleanup operations

// CleanupExpiredKeys removes expired keys and performs maintenance
func (c *Client) CleanupExpiredKeys(ctx context.Context, tenantID string) (int64, error) {
	var totalDeleted int64
	
	// Cleanup expired sessions
	tenantKey := c.tenantSessionsKey(tenantID)
	sessionIDs, err := c.rdb.SMembers(ctx, tenantKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get tenant sessions: %w", err)
	}
	
	for _, sessionID := range sessionIDs {
		sessionKey := c.sessionKey(sessionID, tenantID)
		exists, err := c.rdb.Exists(ctx, sessionKey).Result()
		if err != nil {
			continue
		}
		if exists == 0 {
			// Session expired, remove from index
			c.rdb.SRem(ctx, tenantKey, sessionID)
			totalDeleted++
		}
	}
	
	return totalDeleted, nil
}

// Key generation helpers

func (c *Client) sessionKey(sessionID, tenantID string) string {
	return fmt.Sprintf("session:%s:%s", tenantID, sessionID)
}

func (c *Client) tenantSessionsKey(tenantID string) string {
	return fmt.Sprintf("tenant_sessions:%s", tenantID)
}

func (c *Client) cacheKey(key, tenantID string) string {
	return fmt.Sprintf("cache:%s:%s", tenantID, key)
}

func (c *Client) workspaceKey(topic, tenantID string) string {
	return fmt.Sprintf("workspace:%s:%s", tenantID, topic)
}

func (c *Client) workspaceChannelKey(topic, tenantID string) string {
	return fmt.Sprintf("workspace_channel:%s:%s", tenantID, topic)
}

func (c *Client) metricsKey(metric string, labels map[string]string) string {
	key := fmt.Sprintf("metrics:%s", metric)
	for k, v := range labels {
		key += fmt.Sprintf(":%s=%s", k, v)
	}
	return key
}