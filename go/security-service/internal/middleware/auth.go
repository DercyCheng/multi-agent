package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/multi-agent/go/security-service/internal/auth"
	"github.com/multi-agent/go/security-service/internal/rbac"
)

// RequireAuth middleware that requires authentication
func RequireAuth(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("session_id", claims.SessionID)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequireRole middleware that requires a specific role
func RequireRole(rbacService *rbac.Service, requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not specified"})
			c.Abort()
			return
		}

		// Check if user has the required role
		hasRole, err := rbacService.CheckRole(c.Request.Context(), userID.(string), tenantID.(string), requiredRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check role"})
			c.Abort()
			return
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission middleware that requires a specific permission
func RequirePermission(rbacService *rbac.Service, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not specified"})
			c.Abort()
			return
		}

		// Parse permission (format: resource:action)
		parts := strings.SplitN(permission, ":", 2)
		if len(parts) != 2 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permission format"})
			c.Abort()
			return
		}

		resource := parts[0]
		action := parts[1]

		// Check if user has the required permission
		hasPermission, err := rbacService.CheckPermission(c.Request.Context(), userID.(string), tenantID.(string), resource, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permission"})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenantAccess middleware that ensures user has access to the specified tenant
func RequireTenantAccess(rbacService *rbac.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get tenant ID from URL parameter or query
		tenantID := c.Param("tenant_id")
		if tenantID == "" {
			tenantID = c.Query("tenant_id")
		}
		if tenantID == "" {
			// Try to get from context (set by previous middleware)
			if contextTenantID, exists := c.Get("tenant_id"); exists {
				tenantID = contextTenantID.(string)
			}
		}

		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
			c.Abort()
			return
		}

		// Check tenant access through RBAC service
		hasAccess, err := rbacService.CheckRole(c.Request.Context(), userID.(string), tenantID, "member")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tenant access"})
			c.Abort()
			return
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to tenant"})
			c.Abort()
			return
		}

		// Set tenant ID in context for downstream handlers
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

// Logger middleware for structured logging
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Security middleware for security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimiting middleware (basic implementation)
func RateLimiting() gin.HandlerFunc {
	// This is a basic implementation. In production, use Redis-based rate limiting
	return func(c *gin.Context) {
		// TODO: Implement proper rate limiting with Redis
		c.Next()
	}
}

// AuditLog middleware for security audit logging
func AuditLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log security-relevant events
		userID, _ := c.Get("user_id")
		tenantID, _ := c.Get("tenant_id")
		sessionID, _ := c.Get("session_id")

		logger.Info("Security Audit",
			zap.String("event", "api_access"),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("user_id", getString(userID)),
			zap.String("tenant_id", getString(tenantID)),
			zap.String("session_id", getString(sessionID)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()

		// Log response status for security events
		if c.Writer.Status() >= 400 {
			logger.Warn("Security Event",
				zap.String("event", "access_denied"),
				zap.Int("status_code", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("user_id", getString(userID)),
				zap.String("tenant_id", getString(tenantID)),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}

// Helper function to safely convert interface{} to string
func getString(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}