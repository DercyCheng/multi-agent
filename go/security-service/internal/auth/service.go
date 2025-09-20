package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/multi-agent/go/security-service/internal/config"
	"github.com/multi-agent/go/security-service/internal/models"
)

// DatabaseClient interface for database operations
type DatabaseClient interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUserLastLogin(ctx context.Context, userID string, loginTime time.Time) error
	CreateSession(ctx context.Context, session *models.Session) error
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateSession(ctx context.Context, session *models.Session) error
	InvalidateSession(ctx context.Context, sessionID string) error
	GetActiveSessions(ctx context.Context, userID string) ([]models.Session, error)
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
	CreateAuditEvent(ctx context.Context, event *models.AuditEvent) error
	GetAuditEvents(ctx context.Context, tenantID string, page, limit int, eventType, userID string) ([]models.AuditEvent, int, error)
}

// RBACService interface for RBAC operations
type RBACService interface {
	CheckRole(ctx context.Context, userID, tenantID, roleName string) (bool, error)
	CheckPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error)
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error)
}

// Service handles authentication operations
type Service struct {
	db          DatabaseClient
	rbacService RBACService
	jwtConfig   config.JWTConfig
	logger      *zap.Logger
}

// NewService creates a new authentication service
func NewService(db DatabaseClient, rbacService RBACService, jwtConfig config.JWTConfig, logger *zap.Logger) *Service {
	return &Service{
		db:          db,
		rbacService: rbacService,
		jwtConfig:   jwtConfig,
		logger:      logger,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	SessionID   string   `json:"session_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns tokens
func (s *Service) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by username
	user, err := s.db.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		s.logAuditEvent(c.Request.Context(), "", req.TenantID, "", "login", "user", "authenticate", "failed", 
			map[string]interface{}{"username": req.Username, "reason": "user_not_found"}, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logAuditEvent(c.Request.Context(), user.ID, req.TenantID, "", "login", "user", "authenticate", "failed", 
			map[string]interface{}{"username": req.Username, "reason": "invalid_password"}, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user has access to tenant
	hasAccess, err := s.rbacService.CheckRole(c.Request.Context(), user.ID, req.TenantID, "member")
	if err != nil || !hasAccess {
		s.logAuditEvent(c.Request.Context(), user.ID, req.TenantID, "", "login", "tenant", "access", "denied", 
			map[string]interface{}{"username": req.Username, "reason": "no_tenant_access"}, c.ClientIP(), c.Request.UserAgent())
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to tenant"})
		return
	}

	// Get user roles and permissions
	roles, err := s.rbacService.GetUserRoles(c.Request.Context(), user.ID, req.TenantID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.Error(err))
		roles = []rbac.Role{}
	}

	permissions, err := s.db.GetUserPermissions(c.Request.Context(), user.ID, req.TenantID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.Error(err))
		permissions = []rbac.Permission{}
	}

	// Create session
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TenantID:  req.TenantID,
		ExpiresAt: time.Now().Add(s.jwtConfig.RefreshTokenTTL),
		CreatedAt: time.Now(),
		LastUsedAt: time.Now(),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Status:    "active",
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, req.TenantID, session.ID, roles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	session.AccessToken = accessToken
	session.RefreshToken = refreshToken

	// Store session
	if err := s.db.CreateSession(c.Request.Context(), session); err != nil {
		s.logger.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Update last login time
	user.LastLoginAt = &session.CreatedAt
	if err := s.db.UpdateUserLastLogin(c.Request.Context(), user.ID, session.CreatedAt); err != nil {
		s.logger.Warn("Failed to update last login time", zap.Error(err))
	}

	// Log successful login
	s.logAuditEvent(c.Request.Context(), user.ID, req.TenantID, session.ID, "login", "user", "authenticate", "success", 
		map[string]interface{}{"username": req.Username}, c.ClientIP(), c.Request.UserAgent())

	s.logger.Info("User logged in successfully",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("tenant_id", req.TenantID),
		zap.String("session_id", session.ID),
	)

	response := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		User:         *user,
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken refreshes an access token
func (s *Service) RefreshToken(c *gin.Context) {
	var req models.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session by refresh token
	session, err := s.db.GetSessionByRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// Get user
	user, err := s.db.GetUser(c.Request.Context(), session.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get user roles and permissions
	roles, err := s.rbacService.GetUserRoles(c.Request.Context(), user.ID, session.TenantID)
	if err != nil {
		s.logger.Error("Failed to get user roles", zap.Error(err))
		roles = []rbac.Role{}
	}

	permissions, err := s.db.GetUserPermissions(c.Request.Context(), user.ID, session.TenantID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.Error(err))
		permissions = []rbac.Permission{}
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user.ID, session.TenantID, session.ID, roles, permissions)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update session
	session.AccessToken = accessToken
	session.LastUsedAt = time.Now()
	if err := s.db.UpdateSession(c.Request.Context(), session); err != nil {
		s.logger.Error("Failed to update session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	// Log token refresh
	s.logAuditEvent(c.Request.Context(), user.ID, session.TenantID, session.ID, "token_refresh", "session", "refresh", "success", 
		map[string]interface{}{}, c.ClientIP(), c.Request.UserAgent())

	response := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt,
		User:         *user,
	}

	c.JSON(http.StatusOK, response)
}

// Logout logs out a user
func (s *Service) Logout(c *gin.Context) {
	sessionID, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session not found"})
		return
	}

	userID, _ := c.Get("user_id")
	tenantID, _ := c.Get("tenant_id")

	// Invalidate session
	if err := s.db.InvalidateSession(c.Request.Context(), sessionID.(string)); err != nil {
		s.logger.Error("Failed to invalidate session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Log logout
	s.logAuditEvent(c.Request.Context(), getString(userID), getString(tenantID), sessionID.(string), "logout", "session", "invalidate", "success", 
		map[string]interface{}{}, c.ClientIP(), c.Request.UserAgent())

	s.logger.Info("User logged out successfully",
		zap.String("user_id", getString(userID)),
		zap.String("session_id", sessionID.(string)),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetProfile gets the current user's profile
func (s *Service) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := s.db.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		s.logger.Error("Failed to get user profile", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAuditEvents gets audit events with pagination
func (s *Service) GetAuditEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	eventType := c.Query("event_type")
	userID := c.Query("user_id")
	
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID required"})
		return
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	events, total, err := s.db.GetAuditEvents(c.Request.Context(), tenantID.(string), page, limit, eventType, userID)
	if err != nil {
		s.logger.Error("Failed to get audit events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetActiveSessions gets active sessions for the current user
func (s *Service) GetActiveSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessions, err := s.db.GetActiveSessions(c.Request.Context(), userID.(string))
	if err != nil {
		s.logger.Error("Failed to get active sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// RevokeSession revokes a specific session
func (s *Service) RevokeSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verify session belongs to user (or user is admin)
	session, err := s.db.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	if session.UserID != userID.(string) {
		// Check if user has admin permissions
		tenantID, _ := c.Get("tenant_id")
		hasPermission, err := s.rbacService.CheckPermission(c.Request.Context(), userID.(string), getString(tenantID), "session", "revoke")
		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}
	}

	// Revoke session
	if err := s.db.InvalidateSession(c.Request.Context(), sessionID); err != nil {
		s.logger.Error("Failed to revoke session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	// Log session revocation
	tenantID, _ := c.Get("tenant_id")
	currentSessionID, _ := c.Get("session_id")
	s.logAuditEvent(c.Request.Context(), userID.(string), getString(tenantID), getString(currentSessionID), "session_revoke", "session", "revoke", "success", 
		map[string]interface{}{"revoked_session_id": sessionID}, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateAccessToken generates a JWT access token
func (s *Service) generateAccessToken(userID, tenantID, sessionID string, roles []models.Role, permissions []models.Permission) (string, error) {
	now := time.Now()
	
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	
	permissionNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permissionNames[i] = perm.Name
	}

	claims := Claims{
		UserID:      userID,
		TenantID:    tenantID,
		SessionID:   sessionID,
		Roles:       roleNames,
		Permissions: permissionNames,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwtConfig.Issuer,
			Audience:  []string{s.jwtConfig.Audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtConfig.AccessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.SecretKey))
}

// generateRefreshToken generates a random refresh token
func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// logAuditEvent logs a security audit event
func (s *Service) logAuditEvent(ctx context.Context, userID, tenantID, sessionID, eventType, resource, action, result string, details map[string]interface{}, ipAddress, userAgent string) {
	event := &models.AuditEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		TenantID:  tenantID,
		SessionID: sessionID,
		EventType: eventType,
		Resource:  resource,
		Action:    action,
		Result:    result,
		Details:   details,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	if err := s.db.CreateAuditEvent(ctx, event); err != nil {
		s.logger.Error("Failed to log audit event", zap.Error(err))
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