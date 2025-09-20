package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/multi-agent/go/security-service/internal/config"
)

// Engine handles Open Policy Agent integration
type Engine struct {
	config     config.OPAConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewEngine creates a new OPA engine
func NewEngine(config config.OPAConfig, logger *zap.Logger) (*Engine, error) {
	engine := &Engine{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}

	// Test OPA connection
	if err := engine.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to OPA: %w", err)
	}

	logger.Info("OPA engine initialized successfully",
		zap.String("server_url", config.ServerURL),
		zap.String("policy_path", config.PolicyPath),
	)

	return engine, nil
}

// Policy represents an OPA policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Rules       string                 `json:"rules"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	CreatedBy   string                `json:"created_by"`
	Status      string                `json:"status"`
}

// EvaluationRequest represents a policy evaluation request
type EvaluationRequest struct {
	Input  map[string]interface{} `json:"input"`
	Policy string                 `json:"policy,omitempty"`
	Query  string                 `json:"query,omitempty"`
}

// EvaluationResult represents a policy evaluation result
type EvaluationResult struct {
	Result   interface{}            `json:"result"`
	Decision bool                   `json:"decision"`
	Reason   string                 `json:"reason,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// testConnection tests the connection to OPA server
func (e *Engine) testConnection() error {
	resp, err := e.httpClient.Get(e.config.ServerURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OPA server returned status %d", resp.StatusCode)
	}

	return nil
}

// EvaluatePolicy evaluates a policy against input data
func (e *Engine) EvaluatePolicy(c *gin.Context) {
	var req EvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := e.evaluate(c.Request.Context(), req)
	if err != nil {
		e.logger.Error("Failed to evaluate policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate policy"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// evaluate performs the actual policy evaluation
func (e *Engine) evaluate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	// Prepare OPA query
	opaRequest := map[string]interface{}{
		"input": req.Input,
	}

	jsonData, err := json.Marshal(opaRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to OPA
	url := e.config.ServerURL + e.config.PolicyPath
	if req.Query != "" {
		url += "/" + req.Query
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	var opaResponse map[string]interface{}
	if err := json.Unmarshal(body, &opaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract decision from result
	result := &EvaluationResult{
		Result: opaResponse["result"],
	}

	// Determine decision based on result
	if resultData, ok := opaResponse["result"].(map[string]interface{}); ok {
		if allow, exists := resultData["allow"]; exists {
			if allowBool, ok := allow.(bool); ok {
				result.Decision = allowBool
			}
		}
		if reason, exists := resultData["reason"]; exists {
			if reasonStr, ok := reason.(string); ok {
				result.Reason = reasonStr
			}
		}
	}

	return result, nil
}

// CheckPermission checks if a user has permission for a specific action
func (e *Engine) CheckPermission(ctx context.Context, userID, tenantID, resource, action string) (bool, error) {
	req := EvaluationRequest{
		Input: map[string]interface{}{
			"user_id":   userID,
			"tenant_id": tenantID,
			"resource":  resource,
			"action":    action,
		},
		Query: "allow",
	}

	result, err := e.evaluate(ctx, req)
	if err != nil {
		return false, err
	}

	return result.Decision, nil
}

// CheckTenantAccess checks if a user has access to a specific tenant
func (e *Engine) CheckTenantAccess(ctx context.Context, userID, tenantID string) (bool, error) {
	req := EvaluationRequest{
		Input: map[string]interface{}{
			"user_id":   userID,
			"tenant_id": tenantID,
			"action":    "access",
		},
		Query: "tenant_access",
	}

	result, err := e.evaluate(ctx, req)
	if err != nil {
		return false, err
	}

	return result.Decision, nil
}

// CreatePolicy creates a new policy
func (e *Engine) CreatePolicy(c *gin.Context) {
	var policy Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	policy.CreatedBy = userID.(string)
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()
	policy.Status = "active"

	// Store policy in OPA
	if err := e.storePolicy(c.Request.Context(), &policy); err != nil {
		e.logger.Error("Failed to store policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create policy"})
		return
	}

	c.JSON(http.StatusCreated, policy)
}

// ListPolicies lists all policies
func (e *Engine) ListPolicies(c *gin.Context) {
	policies, err := e.listPolicies(c.Request.Context())
	if err != nil {
		e.logger.Error("Failed to list policies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list policies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

// GetPolicy retrieves a specific policy
func (e *Engine) GetPolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	policy, err := e.getPolicy(c.Request.Context(), policyID)
	if err != nil {
		e.logger.Error("Failed to get policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get policy"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// UpdatePolicy updates a policy
func (e *Engine) UpdatePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	var policy Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy.ID = policyID
	policy.UpdatedAt = time.Now()

	if err := e.updatePolicy(c.Request.Context(), &policy); err != nil {
		e.logger.Error("Failed to update policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update policy"})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// DeletePolicy deletes a policy
func (e *Engine) DeletePolicy(c *gin.Context) {
	policyID := c.Param("id")
	if policyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Policy ID is required"})
		return
	}

	if err := e.deletePolicy(c.Request.Context(), policyID); err != nil {
		e.logger.Error("Failed to delete policy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Policy deleted successfully"})
}

// storePolicy stores a policy in OPA
func (e *Engine) storePolicy(ctx context.Context, policy *Policy) error {
	url := e.config.ServerURL + "/v1/policies/" + policy.ID
	
	jsonData, err := json.Marshal(map[string]string{
		"raw": policy.Rules,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store policy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// listPolicies lists all policies from OPA
func (e *Engine) listPolicies(ctx context.Context) ([]Policy, error) {
	url := e.config.ServerURL + "/v1/policies"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert OPA response to Policy structs
	var policies []Policy
	if result, ok := response["result"].([]interface{}); ok {
		for _, item := range result {
			if policyData, ok := item.(map[string]interface{}); ok {
				policy := Policy{
					ID:   policyData["id"].(string),
					Name: policyData["id"].(string), // Use ID as name for now
				}
				policies = append(policies, policy)
			}
		}
	}

	return policies, nil
}

// getPolicy retrieves a specific policy from OPA
func (e *Engine) getPolicy(ctx context.Context, policyID string) (*Policy, error) {
	url := e.config.ServerURL + "/v1/policies/" + policyID
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("policy not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	policy := &Policy{
		ID:   policyID,
		Name: policyID,
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if raw, exists := result["raw"]; exists {
			policy.Rules = raw.(string)
		}
	}

	return policy, nil
}

// updatePolicy updates a policy in OPA
func (e *Engine) updatePolicy(ctx context.Context, policy *Policy) error {
	return e.storePolicy(ctx, policy)
}

// deletePolicy deletes a policy from OPA
func (e *Engine) deletePolicy(ctx context.Context, policyID string) error {
	url := e.config.ServerURL + "/v1/policies/" + policyID
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("OPA returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}