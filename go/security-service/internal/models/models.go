package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Status       string     `json:"status" db:"status"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at" db:"last_used_at"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Status       string    `json:"status" db:"status"`
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Domain    string                 `json:"domain" db:"domain"`
	Status    string                 `json:"status" db:"status"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	Limits    map[string]interface{} `json:"limits" db:"limits"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy string                 `json:"created_by" db:"created_by"`
}

// Role represents a role in the RBAC system
type Role struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
}

// Permission represents a permission in the RBAC system
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	RoleID     string     `json:"role_id" db:"role_id"`
	TenantID   string     `json:"tenant_id" db:"tenant_id"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	AssignedBy string     `json:"assigned_by" db:"assigned_by"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"`
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	SessionID string                 `json:"session_id" db:"session_id"`
	EventType string                 `json:"event_type" db:"event_type"`
	Resource  string                 `json:"resource" db:"resource"`
	Action    string                 `json:"action" db:"action"`
	Result    string                 `json:"result" db:"result"`
	Details   map[string]interface{} `json:"details" db:"details"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TenantID string `json:"tenant_id"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
	Tenant       Tenant    `json:"tenant"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateTenantRequest represents a tenant creation request
type CreateTenantRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Domain   string                 `json:"domain" binding:"required"`
	Settings map[string]interface{} `json:"settings"`
	Limits   map[string]interface{} `json:"limits"`
}

// UpdateTenantRequest represents a tenant update request
type UpdateTenantRequest struct {
	Name     string                 `json:"name"`
	Domain   string                 `json:"domain"`
	Status   string                 `json:"status"`
	Settings map[string]interface{} `json:"settings"`
	Limits   map[string]interface{} `json:"limits"`
}

// CreateRoleRequest represents a role creation request
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TenantID    string `json:"tenant_id" binding:"required"`
}

// UpdateRoleRequest represents a role update request
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AssignRoleRequest represents a role assignment request
type AssignRoleRequest struct {
	UserID    string     `json:"user_id" binding:"required"`
	RoleID    string     `json:"role_id" binding:"required"`
	TenantID  string     `json:"tenant_id" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// UpdateRolePermissionsRequest represents a role permissions update request
type UpdateRolePermissionsRequest struct {
	PermissionNames []string `json:"permission_names" binding:"required"`
}

// PolicyEvaluationRequest represents a policy evaluation request
type PolicyEvaluationRequest struct {
	UserID     string                 `json:"user_id" binding:"required"`
	TenantID   string                 `json:"tenant_id" binding:"required"`
	Resource   string                 `json:"resource" binding:"required"`
	Action     string                 `json:"action" binding:"required"`
	Context    map[string]interface{} `json:"context"`
	ClientIP   string                 `json:"client_ip"`
	UserAgent  string                 `json:"user_agent"`
	WorkflowID string                 `json:"workflow_id"`
	AgentID    string                 `json:"agent_id"`
}

// PolicyEvaluationResponse represents a policy evaluation response
type PolicyEvaluationResponse struct {
	Allow  bool   `json:"allow"`
	Reason string `json:"reason,omitempty"`
}

// ListResponse represents a generic list response
type ListResponse struct {
	Data  interface{} `json:"data"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}