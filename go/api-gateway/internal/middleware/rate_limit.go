package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"golang.org/x/time/rate"
	"github.com/multi-agent/api-gateway/internal/config"
	"github.com/multi-agent/api-gateway/pkg/logger"
	"github.com/multi-agent/api-gateway/pkg/metrics"
)

// RateLimiter handles rate limiting
type RateLimiter struct {
	config      *config.RateLimitConfig
	redisClient *redis.Client
	limiters    map[string]*rate.Limiter
}

// RateLimit creates rate limiting middleware
func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Initialize Redis client for distributed rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379", // This should come from config
		DB:   3,            // Use a separate DB for rate limiting
	})

	limiter := &RateLimiter{
		config:      &cfg,
		redisClient: redisClient,
		limiters:    make(map[string]*rate.Limiter),
	}

	return limiter.Handle
}

// Handle processes rate limiting
func (rl *RateLimiter) Handle(c *gin.Context) {
	// Get client identifier (user ID or IP)
	clientID := rl.getClientID(c)
	
	// Check rate limit
	allowed, err := rl.checkRateLimit(c.Request.Context(), clientID, c.Request.URL.Path)
	if err != nil {
		logger.Error("Rate limit check failed", "client", clientID, "error", err)
		c.Next() // Allow request on error
		return
	}

	if !allowed {
		// Record rate limit hit
		metrics.RecordRateLimitHit(clientID, c.Request.URL.Path)
		
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "Rate limit exceeded",
			"code":  "RATE_LIMIT_EXCEEDED",
			"retry_after": 60,
		})
		c.Abort()
		return
	}

	c.Next()
}

// getClientID extracts client identifier
func (rl *RateLimiter) getClientID(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok && uid != "" {
			return fmt.Sprintf("user:%s", uid)
		}
	}

	// Fall back to IP address
	clientIP := c.ClientIP()
	return fmt.Sprintf("ip:%s", clientIP)
}

// checkRateLimit checks if request is within rate limit
func (rl *RateLimiter) checkRateLimit(ctx context.Context, clientID, endpoint string) (bool, error) {
	// Use Redis for distributed rate limiting
	key := fmt.Sprintf("ratelimit:%s:%s", clientID, endpoint)
	
	// Use sliding window rate limiting
	now := time.Now()
	window := time.Minute
	
	// Remove old entries
	rl.redisClient.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.Add(-window).Unix(), 10))
	
	// Count current requests in window
	count, err := rl.redisClient.ZCard(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Check if limit exceeded
	if int(count) >= rl.config.RequestsPerMin {
		return false, nil
	}

	// Add current request
	score := float64(now.Unix())
	member := fmt.Sprintf("%d-%d", now.UnixNano(), count)
	
	pipe := rl.redisClient.Pipeline()
	pipe.ZAdd(ctx, key, &redis.Z{Score: score, Member: member})
	pipe.Expire(ctx, key, window)
	
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TokenBucket implements token bucket rate limiting for burst handling
type TokenBucket struct {
	limiter *rate.Limiter
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(requestsPerMin, burstSize int) *TokenBucket {
	// Convert requests per minute to requests per second
	rps := rate.Limit(float64(requestsPerMin) / 60.0)
	
	return &TokenBucket{
		limiter: rate.NewLimiter(rps, burstSize),
	}
}

// Allow checks if request is allowed
func (tb *TokenBucket) Allow() bool {
	return tb.limiter.Allow()
}

// Wait waits until request can be processed
func (tb *TokenBucket) Wait(ctx context.Context) error {
	return tb.limiter.Wait(ctx)
}