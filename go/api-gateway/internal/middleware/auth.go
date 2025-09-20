package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/go-redis/redis/v8"
	"github.com/multi-agent/api-gateway/internal/config"

)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	config      *config.AuthConfig
	redisClient *redis.Client
}

// UserClaims represents JWT claims
type UserClaims struct {
	UserID     string   `json:"user_id"`
	TenantID   string   `json:"tenant_id"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Auth creates authentication middleware
func Auth(cfg config.AuthConfig) gin.HandlerFunc {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: strings.TrimPrefix(cfg.RedisURL, "redis://"),
	})

	middleware := &AuthMiddleware{
		config:      &cfg,
		redisClient: redisClient,
	}

	return middleware.Handle
}

// Handle processes authentication
func (a *AuthMiddleware) Handle(c *gin.Context) {
	// Skip authentication for health and metrics endpoints
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") {
		c.Next()
		return
	}

	// Try JWT authentication first
	if token := a.extractJWTToken(c); token != "" {
		if user, err := a.validateJWTToken(token); err == nil {
			a.setUserContext(c, user)
			c.Next()
			return
		}
	}

	// Try API key authentication
	if apiKey := a.extractAPIKey(c); apiKey != "" {
		if user, err := a.validateAPIKey(c.Request.Context(), apiKey); err == nil {
			a.setUserContext(c, user)
			c.Next()
			return
		}
	}

	// No valid authentication found
	c.JSON(http.StatusUnauthorized, gin.H{
		"error": "Authentication required",
		"code":  "UNAUTHORIZED",
	})
	c.Abort()
}

// extractJWTToken extracts JWT token from Authorization header
func (a *AuthMiddleware) extractJWTToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// extractAPIKey extracts API key from header
func (a *AuthMiddleware) extractAPIKey(c *gin.Context) string {
	return c.GetHeader(a.config.APIKeyHeader)
}

// validateJWTToken validates JWT token and returns user claims
func (a *AuthMiddleware) validateJWTToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Check if token is blacklisted (optional)
	if a.redisClient != nil {
		blacklisted, err := a.redisClient.Get(context.Background(), fmt.Sprintf("blacklist:%s", tokenString)).Result()
		if err == nil && blacklisted == "true" {
			return nil, fmt.Errorf("token blacklisted")
		}
	}

	return claims, nil
}

// validateAPIKey validates API key and returns user information
func (a *AuthMiddleware) validateAPIKey(ctx context.Context, apiKey string) (*UserClaims, error) {
	if a.redisClient == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	// Get API key information from Redis
	keyInfo, err := a.redisClient.HGetAll(ctx, fmt.Sprintf("apikey:%s", apiKey)).Result()
	if err != nil || len(keyInfo) == 0 {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if API key is active
	if keyInfo["status"] != "active" {
		return nil, fmt.Errorf("API key inactive")
	}

	// Check expiration
	if expiresAt, exists := keyInfo["expires_at"]; exists && expiresAt != "" {
		expTime, err := time.Parse(time.RFC3339, expiresAt)
		if err == nil && expTime.Before(time.Now()) {
			return nil, fmt.Errorf("API key expired")
		}
	}

	// Create user claims from API key info
	claims := &UserClaims{
		UserID:   keyInfo["user_id"],
		TenantID: keyInfo["tenant_id"],
		Email:    keyInfo["email"],
		Roles:    strings.Split(keyInfo["roles"], ","),
		Permissions: strings.Split(keyInfo["permissions"], ","),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   keyInfo["user_id"],
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.JWTExpiration)),
		},
	}

	// Update last used timestamp
	a.redisClient.HSet(ctx, fmt.Sprintf("apikey:%s", apiKey), "last_used", time.Now().Format(time.RFC3339))

	return claims, nil
}

// setUserContext sets user information in request context
func (a *AuthMiddleware) setUserContext(c *gin.Context, user *UserClaims) {
	c.Set("user_id", user.UserID)
	c.Set("tenant_id", user.TenantID)
	c.Set("user_email", user.Email)
	c.Set("user_roles", user.Roles)
	c.Set("user_permissions", user.Permissions)
	c.Set("user_claims", user)

	// Add to request headers for downstream services
	c.Request.Header.Set("X-User-ID", user.UserID)
	c.Request.Header.Set("X-Tenant-ID", user.TenantID)
	c.Request.Header.Set("X-User-Email", user.Email)
	c.Request.Header.Set("X-User-Roles", strings.Join(user.Roles, ","))
}

// RequirePermission creates middleware that requires specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissions, exists := c.Get("user_permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		permissions, ok := userPermissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid permissions",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		// Check if user has required permission
		hasPermission := false
		for _, perm := range permissions {
			if perm == permission || perm == "*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Permission required: %s", permission),
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole creates middleware that requires specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		roles, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid roles",
				"code":  "FORBIDDEN",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, r := range roles {
			if r == role || r == "admin" {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("Role required: %s", role),
				"code":  "INSUFFICIENT_ROLE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}