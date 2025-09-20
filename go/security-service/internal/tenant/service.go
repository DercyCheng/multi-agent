package tenant

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
)

// Service handles tenant management operations
type Service struct {
	db     *database.Client
	logger *zap.Logger
}

// NewService creates a new tenant service
func NewService(db *database.Client, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Domain      string                 `json:"domain" db:"domain"`
	Status      string                 `json:"status" db:"status"`
	Settings    map[string]interface{} `json:"settings" db:"settings"`
	Limits      TenantLimits          `json:"limits" db:"limits"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at" db:"updated_at"`
	CreatedBy   string                `json:"created_by" db:"created_by"`
}

// TenantLimits defines resource limits for a tenant
type TenantLimits struct {
	MaxUsers           int     `json:"max_users"`
	MaxAgents          int     `json:"max_agents"`
	MaxWorkflows       int     `json:"max_workflows"`
	DailyTokenLimit    int     `json:"daily_token_limit"`
	MonthlyTokenLimit  int     `json:"monthly_token_limit"`
	DailyBudgetUSD     float64 `json:"daily_budget_usd"`
	MonthlyBudgetUSD   float64 `json:"monthly_budget_usd"`
	StorageQuotaGB     int     `json:"storage_quota_gb"`
	APIRateLimit       int     `json:"api_rate_limit"`
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Domain   string                 `json:"domain" binding:"required"`
	Settings map[string]interface{} `json:"settings"`
	Limits   TenantLimits          `json:"limits"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name     string                 `json:"name"`
	Domain   string                 `json:"domain"`
	Status   string                 `json:"status"`
	Settings map[string]interface{} `json:"settings"`
	Limits   TenantLimits          `json:"limits"`
}

// CreateTenant creates a new tenant
func (s *Service) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
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

	// Set default limits if not provided
	if req.Limits.MaxUsers == 0 {
		req.Limits.MaxUsers = 100
	}
	if req.Limits.MaxAgents == 0 {
		req.Limits.MaxAgents = 50
	}
	if req.Limits.MaxWorkflows == 0 {
		req.Limits.MaxWorkflows = 1000
	}
	if req.Limits.DailyTokenLimit == 0 {
		req.Limits.DailyTokenLimit = 100000
	}
	if req.Limits.MonthlyTokenLimit == 0 {
		req.Limits.MonthlyTokenLimit = 3000000
	}
	if req.Limits.DailyBudgetUSD == 0 {
		req.Limits.DailyBudgetUSD = 10.0
	}
	if req.Limits.MonthlyBudgetUSD == 0 {
		req.Limits.MonthlyBudgetUSD = 300.0
	}
	if req.Limits.StorageQuotaGB == 0 {
		req.Limits.StorageQuotaGB = 10
	}
	if req.Limits.APIRateLimit == 0 {
		req.Limits.APIRateLimit = 1000
	}

	tenant := &Tenant{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Domain:    req.Domain,
		Status:    "active",
		Settings:  req.Settings,
		Limits:    req.Limits,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: userID.(string),
	}

	if err := s.db.CreateTenant(c.Request.Context(), tenant); err != nil {
		s.logger.Error("Failed to create tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	s.logger.Info("Tenant created successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name),
		zap.String("created_by", tenant.CreatedBy),
	)

	c.JSON(http.StatusCreated, tenant)
}

// ListTenants lists all tenants with pagination
func (s *Service) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tenants, total, err := s.db.ListTenants(c.Request.Context(), page, limit, status)
	if err != nil {
		s.logger.Error("Failed to list tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTenant retrieves a specific tenant by ID
func (s *Service) GetTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	tenant, err := s.db.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		s.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant updates a tenant
func (s *Service) UpdateTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing tenant
	tenant, err := s.db.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		s.logger.Error("Failed to get tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tenant"})
		return
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Domain != "" {
		tenant.Domain = req.Domain
	}
	if req.Status != "" {
		tenant.Status = req.Status
	}
	if req.Settings != nil {
		tenant.Settings = req.Settings
	}
	if req.Limits.MaxUsers > 0 {
		tenant.Limits = req.Limits
	}
	tenant.UpdatedAt = time.Now()

	if err := s.db.UpdateTenant(c.Request.Context(), tenant); err != nil {
		s.logger.Error("Failed to update tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	s.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", tenant.ID),
		zap.String("name", tenant.Name),
	)

	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant deletes a tenant (soft delete)
func (s *Service) DeleteTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := s.db.DeleteTenant(c.Request.Context(), tenantID); err != nil {
		if err.Error() == "tenant not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		s.logger.Error("Failed to delete tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tenant"})
		return
	}

	s.logger.Info("Tenant deleted successfully",
		zap.String("tenant_id", tenantID),
		zap.String("deleted_by", userID.(string)),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Tenant deleted successfully"})
}

// AddUserToTenant adds a user to a tenant
func (s *Service) AddUserToTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID is required"})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.AddUserToTenant(c.Request.Context(), tenantID, req.UserID, req.Role); err != nil {
		s.logger.Error("Failed to add user to tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to tenant"})
		return
	}

	s.logger.Info("User added to tenant successfully",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", req.UserID),
		zap.String("role", req.Role),
	)

	c.JSON(http.StatusOK, gin.H{"message": "User added to tenant successfully"})
}

// RemoveUserFromTenant removes a user from a tenant
func (s *Service) RemoveUserFromTenant(c *gin.Context) {
	tenantID := c.Param("id")
	userID := c.Param("user_id")
	
	if tenantID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID and User ID are required"})
		return
	}

	if err := s.db.RemoveUserFromTenant(c.Request.Context(), tenantID, userID); err != nil {
		s.logger.Error("Failed to remove user from tenant", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from tenant"})
		return
	}

	s.logger.Info("User removed from tenant successfully",
		zap.String("tenant_id", tenantID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "User removed from tenant successfully"})
}