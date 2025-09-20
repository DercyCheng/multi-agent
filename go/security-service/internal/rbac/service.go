package rbac

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/multi-agent/go/security-service/internal/database"
	"github.com/multi-agent/go/security-service/internal/opa"
)

// Service handles RBAC operations
type Service struct {
	db        *database.Client
	opaEngine *opa.Engine
	logger    *zap.Logger
}

// NewService creates a new RBAC service
func NewService(db *database.Client, opaEngine *opa.Engine, logger *zap.Logger) *Service {
	return &Service{
		db:        db,
		opaEngine: opaEngine,
		logger:    logger,
	}
}

// Role represents a role in the system
type Role struct {
	ID          string      `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	TenantID    string      `json:"tenant_id" db:"tenant_id"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	CreatedBy   string      `json:"created_by" db:"created_by"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents a user-role assignment
type UserRole struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RoleID    string    `json:"role_id" db:"role_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy string    `json:"assigned_by" db:"assigned_by"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
}

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	TenantID    string   `json:"tenant_id" binding:"required"`
	Permissions []string `json:"permissions"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// AssignRoleRequest represents a request to assign a role to a user
type AssignRoleRequest struct {
	RoleID    string     `json:"role_id" binding:"required"`
	TenantID  string     `json:"tenant_id" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateRole creates a new role
func (s *Service) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role := &Role{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		TenantID:    req.TenantID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   userID.(string),
	}

	if err := s.db.CreateRole(c.Request.Context(), role); err != nil {
		s.logger.Error("Failed to create role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	// Assign permissions to role
	if len(req.Permissions) > 0 {
		if err := s.db.AssignPermissionsToRole(c.Request.Context(), role.ID, req.Permissions); err != nil {
			s.logger.Error("Failed to assign permissions to role", zap.Error(err))
			// Continue, role is created but permissions failed
		}
	}

	s.logger.Info("Role created successfully",
		zap.String("role_id", role.ID),
		zap.String("name", role.Name),
		zap.String("tenant_id", role.TenantID),
	)

	c.JSON(http.StatusCreated, role)
}

// ListRoles lists all roles with optional filtering
func (s *Service) ListRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	tenantID := c.Query("tenant_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	roles, total, err := s.db.ListRoles(c.Request.Context(), page, limit, tenantID)
	if err != nil {
		s.logger.Error("Failed to list roles", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// UpdateRole updates a role
func (s *Service) UpdateRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing role
	role, err := s.db.GetRole(c.Request.Context(), roleID)
	if err != nil {
		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		s.logger.Error("Failed to get role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get role"})
		return
	}

	// Update fields
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	role.UpdatedAt = time.Now()

	if err := s.db.UpdateRole(c.Request.Context(), role); err != nil {
		s.logger.Error("Failed to update role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	// Update permissions if provided
	if len(req.Permissions) > 0 {
		if err := s.db.UpdateRolePermissions(c.Request.Context(), roleID, req.Permissions); err != nil {
			s.logger.Error("Failed to update role permissions", zap.Error(err))
		}
	}

	s.logger.Info("Role updated successfully",
		zap.String("role_id", role.ID),
		zap.String("name", role.Name),
	)

	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a role
func (s *Service) DeleteRole(c *gin.Context) {
	roleID := c.Param("id")
	if roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role ID is required"})
		return
	}

	if err := s.db.DeleteRole(c.Request.Context(), roleID); err != nil {
		if err.Error() == "role not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		s.logger.Error("Failed to delete role", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	s.logger.Info("Role deleted successfully", zap.String("role_id", roleID))
	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// CreatePermission creates a new permission
func (s *Service) CreatePermission(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission := &Permission{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		CreatedAt:   time.Now(),
	}

	if err := s.db.CreatePermission(c.Request.Context(), permission); err != nil {
		s.logger.Error("Failed to create permission", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	s.logger.Info("Permission created successfully",
		zap.String("permission_id", permission.ID),
		zap.String("name", permission.Name),
		zap.String("resource", permission.Resource),
		zap.String("action", permission.Action),
	)

	c.JSON(http.StatusCreated, permission)
}

// ListPermissions lists all permissions
func (s *Service) ListPermissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	resource := c.Query("resource")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	permissions, total, err := s.db.ListPermissions(c.Request.Context(), page, limit, resource)
	if err != nil {
		s.logger.Error("Failed to list permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// AssignRole assigns a role to a user
func (s *Service) AssignRole(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	assignedBy, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userRole := &UserRole{
		ID:         uuid.New().String(),
		UserID:     userID,
		RoleID:     req.RoleID,
		TenantID:   req.TenantID,
		AssignedAt: time.Now(),
		AssignedBy: assignedBy.(string),
		ExpiresAt:  req.ExpiresAt,
	}

	if err := s.db.AssignRoleToUser(c.Request.Context(), userRole); err != nil {
		s.logger.Error("Failed to assign role to user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
		return
	}

	s.logger.Info("Role assigned to user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", req.RoleID),
		zap.String("tenant_id", req.TenantID),
	)

	c.JSON(http.StatusCreated, userRole)
}

// RevokeRole revokes a role from a user
func (s *Service) RevokeRole(c *gin.Context) {
	userID := c.Param("user_id")
	roleID := c.Param("role_id")
	
	if userID == "" || roleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID and Role ID are required"})
		return
	}

	if err := s.db.RevokeRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		s.logger.Error("Failed to revoke role from user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke role"})
		return
	}

	s.logger.Info("Role revoked from user successfully",
		zap.String("user_id", userID),
		zap.String("role_id", roleID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Role revoked successfully"})
}

// GetUserPermissions gets all permissions for a user
func (s *Service) GetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	permissions, err := s.db.GetUserPermissions(c.Request.Context(), userID, tenantID)
	if err != nil {
		s.logger.Error("Failed to get user permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// CheckPermission checks if a user has a specific permission
func (s *Service) CheckPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	// First check database for explicit permissions
	hasPermission, err := s.db.CheckUserPermission(ctx, userID, tenantID, resource, action)
	if err != nil {
		return false, fmt.Errorf("failed to check database permission: %w", err)
	}

	if hasPermission {
		return true, nil
	}

	// If not found in database, check OPA policies
	return s.opaEngine.CheckPermission(ctx, userID, tenantID, resource, action)
}

// CheckRole checks if a user has a specific role
func (s *Service) CheckRole(ctx context.Context, userID, tenantID, roleName string) (bool, error) {
	return s.db.CheckUserRole(ctx, userID, tenantID, roleName)
}

// GetUserRoles gets all roles for a user in a tenant
func (s *Service) GetUserRoles(ctx context.Context, userID, tenantID string) ([]Role, error) {
	return s.db.GetUserRoles(ctx, userID, tenantID)
}

// InitializeDefaultRoles creates default roles and permissions
func (s *Service) InitializeDefaultRoles(ctx context.Context, tenantID string) error {
	// Default permissions
	defaultPermissions := []Permission{
		{ID: uuid.New().String(), Name: "read_workflows", Resource: "workflow", Action: "read", Description: "Read workflow data"},
		{ID: uuid.New().String(), Name: "create_workflows", Resource: "workflow", Action: "create", Description: "Create new workflows"},
		{ID: uuid.New().String(), Name: "update_workflows", Resource: "workflow", Action: "update", Description: "Update existing workflows"},
		{ID: uuid.New().String(), Name: "delete_workflows", Resource: "workflow", Action: "delete", Description: "Delete workflows"},
		{ID: uuid.New().String(), Name: "read_agents", Resource: "agent", Action: "read", Description: "Read agent data"},
		{ID: uuid.New().String(), Name: "create_agents", Resource: "agent", Action: "create", Description: "Create new agents"},
		{ID: uuid.New().String(), Name: "update_agents", Resource: "agent", Action: "update", Description: "Update existing agents"},
		{ID: uuid.New().String(), Name: "delete_agents", Resource: "agent", Action: "delete", Description: "Delete agents"},
		{ID: uuid.New().String(), Name: "manage_users", Resource: "user", Action: "manage", Description: "Manage users"},
		{ID: uuid.New().String(), Name: "manage_rbac", Resource: "rbac", Action: "manage", Description: "Manage roles and permissions"},
		{ID: uuid.New().String(), Name: "manage_policies", Resource: "policy", Action: "manage", Description: "Manage security policies"},
		{ID: uuid.New().String(), Name: "read_audit", Resource: "audit", Action: "read", Description: "Read audit logs"},
	}

	// Create permissions
	for _, perm := range defaultPermissions {
		perm.CreatedAt = time.Now()
		if err := s.db.CreatePermission(ctx, &perm); err != nil {
			s.logger.Warn("Failed to create default permission", zap.String("name", perm.Name), zap.Error(err))
		}
	}

	// Default roles
	defaultRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        "admin",
			description: "Full system administrator",
			permissions: []string{"manage_users", "manage_rbac", "manage_policies", "read_audit", "create_workflows", "update_workflows", "delete_workflows", "read_workflows", "create_agents", "update_agents", "delete_agents", "read_agents"},
		},
		{
			name:        "developer",
			description: "Developer with workflow and agent management",
			permissions: []string{"create_workflows", "update_workflows", "read_workflows", "create_agents", "update_agents", "read_agents"},
		},
		{
			name:        "operator",
			description: "Operator with read and execute permissions",
			permissions: []string{"read_workflows", "read_agents"},
		},
		{
			name:        "viewer",
			description: "Read-only access",
			permissions: []string{"read_workflows", "read_agents"},
		},
	}

	// Create roles
	for _, roleData := range defaultRoles {
		role := &Role{
			ID:          uuid.New().String(),
			Name:        roleData.name,
			Description: roleData.description,
			TenantID:    tenantID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			CreatedBy:   "system",
		}

		if err := s.db.CreateRole(ctx, role); err != nil {
			s.logger.Warn("Failed to create default role", zap.String("name", role.Name), zap.Error(err))
			continue
		}

		// Assign permissions to role
		if err := s.db.AssignPermissionsToRole(ctx, role.ID, roleData.permissions); err != nil {
			s.logger.Warn("Failed to assign permissions to default role", zap.String("role", role.Name), zap.Error(err))
		}
	}

	s.logger.Info("Default roles and permissions initialized", zap.String("tenant_id", tenantID))
	return nil
}