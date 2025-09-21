package flags

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Rules       []Rule                 `json:"rules" db:"rules"`
	Rollout     *RolloutConfig         `json:"rollout,omitempty" db:"rollout"`
	Environment string                 `json:"environment" db:"environment"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// Rule represents a targeting rule
type Rule struct {
	ID         string    `json:"id"`
	Attribute  string    `json:"attribute"` // user_id, tenant_id, group, etc.
	Operator   string    `json:"operator"`  // equals, in, contains, regex, etc.
	Values     []string  `json:"values"`
	Percentage *int      `json:"percentage,omitempty"` // 0-100
	Enabled    bool      `json:"enabled"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

// RolloutConfig represents gradual rollout configuration
type RolloutConfig struct {
	Strategy   string     `json:"strategy"`   // percentage, user_group, time_based
	Percentage int        `json:"percentage"` // 0-100
	UserGroups []string   `json:"user_groups,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// EvaluationContext provides context for flag evaluation
type EvaluationContext struct {
	UserID    string                 `json:"user_id"`
	TenantID  string                 `json:"tenant_id"`
	Groups    []string               `json:"groups"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// EvaluationResult represents the result of flag evaluation
type EvaluationResult struct {
	FlagID      string                 `json:"flag_id"`
	Enabled     bool                   `json:"enabled"`
	Reason      string                 `json:"reason"`
	RuleMatched *string                `json:"rule_matched,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	EvaluatedAt time.Time              `json:"evaluated_at"`
}

// Storage interface for flag persistence
type Storage interface {
	// Flag management
	CreateFlag(ctx context.Context, flag *FeatureFlag) error
	GetFlag(ctx context.Context, id string) (*FeatureFlag, error)
	GetFlagByName(ctx context.Context, name, environment, tenantID string) (*FeatureFlag, error)
	UpdateFlag(ctx context.Context, flag *FeatureFlag) error
	DeleteFlag(ctx context.Context, id string) error
	ListFlags(ctx context.Context, environment, tenantID string) ([]*FeatureFlag, error)

	// Audit logging
	LogEvaluation(ctx context.Context, result *EvaluationResult, context *EvaluationContext) error
	LogFlagChange(ctx context.Context, flagID, action, changedBy string, before, after interface{}) error
}

// Manager manages feature flags
type Manager struct {
	storage     Storage
	logger      *zap.Logger
	cache       map[string]*FeatureFlag
	cacheMux    sync.RWMutex
	callbacks   map[string][]func(*FeatureFlag)
	callbackMux sync.RWMutex
}

// NewManager creates a new flags manager
func NewManager(storage Storage, logger *zap.Logger) *Manager {
	return &Manager{
		storage:   storage,
		logger:    logger,
		cache:     make(map[string]*FeatureFlag),
		callbacks: make(map[string][]func(*FeatureFlag)),
	}
}

// CreateFlag creates a new feature flag
func (m *Manager) CreateFlag(ctx context.Context, flag *FeatureFlag) error {
	flag.ID = generateID()
	flag.CreatedAt = time.Now()
	flag.UpdatedAt = time.Now()

	if err := m.storage.CreateFlag(ctx, flag); err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}

	// Update cache
	m.cacheMux.Lock()
	m.cache[flag.ID] = flag
	m.cacheMux.Unlock()

	// Trigger callbacks
	m.triggerCallbacks(flag.ID, flag)

	// Log the creation
	m.storage.LogFlagChange(ctx, flag.ID, "create", flag.CreatedBy, nil, flag)

	m.logger.Info("Feature flag created",
		zap.String("flag_id", flag.ID),
		zap.String("name", flag.Name),
		zap.String("created_by", flag.CreatedBy),
	)

	return nil
}

// GetFlag retrieves a feature flag by ID
func (m *Manager) GetFlag(ctx context.Context, id string) (*FeatureFlag, error) {
	// Check cache first
	m.cacheMux.RLock()
	if flag, exists := m.cache[id]; exists {
		m.cacheMux.RUnlock()
		return flag, nil
	}
	m.cacheMux.RUnlock()

	// Load from storage
	flag, err := m.storage.GetFlag(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get flag: %w", err)
	}

	// Update cache
	m.cacheMux.Lock()
	m.cache[id] = flag
	m.cacheMux.Unlock()

	return flag, nil
}

// UpdateFlag updates a feature flag
func (m *Manager) UpdateFlag(ctx context.Context, id string, updates *FeatureFlag) error {
	// Get current flag for audit
	currentFlag, err := m.GetFlag(ctx, id)
	if err != nil {
		return fmt.Errorf("flag not found: %w", err)
	}

	// Update timestamp
	updates.ID = id
	updates.UpdatedAt = time.Now()
	updates.CreatedAt = currentFlag.CreatedAt

	if err := m.storage.UpdateFlag(ctx, updates); err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	// Update cache
	m.cacheMux.Lock()
	m.cache[id] = updates
	m.cacheMux.Unlock()

	// Trigger callbacks
	m.triggerCallbacks(id, updates)

	// Log the change
	m.storage.LogFlagChange(ctx, id, "update", updates.CreatedBy, currentFlag, updates)

	m.logger.Info("Feature flag updated",
		zap.String("flag_id", id),
		zap.String("name", updates.Name),
	)

	return nil
}

// EvaluateFlag evaluates a feature flag for given context
func (m *Manager) EvaluateFlag(ctx context.Context, flagName, environment, tenantID string, evalCtx *EvaluationContext) (*EvaluationResult, error) {
	// Get flag
	flag, err := m.storage.GetFlagByName(ctx, flagName, environment, tenantID)
	if err != nil {
		result := &EvaluationResult{
			FlagID:      "unknown",
			Enabled:     false,
			Reason:      "flag_not_found",
			EvaluatedAt: time.Now(),
		}
		return result, nil
	}

	result := &EvaluationResult{
		FlagID:      flag.ID,
		Enabled:     false,
		Reason:      "default_disabled",
		EvaluatedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Check if flag is globally enabled
	if !flag.Enabled {
		result.Reason = "flag_disabled"
		m.logEvaluation(ctx, result, evalCtx)
		return result, nil
	}

	// Evaluate rules
	for _, rule := range flag.Rules {
		if !rule.Enabled {
			continue
		}

		if m.evaluateRule(&rule, evalCtx) {
			result.Enabled = true
			result.Reason = "rule_matched"
			ruleID := rule.ID
			result.RuleMatched = &ruleID
			result.Metadata["matched_rule"] = rule
			break
		}
	}

	// Check rollout configuration
	if result.Enabled && flag.Rollout != nil {
		if !m.evaluateRollout(flag.Rollout, evalCtx) {
			result.Enabled = false
			result.Reason = "rollout_excluded"
		} else {
			result.Metadata["rollout"] = flag.Rollout
		}
	}

	// If no rules matched but flag is enabled, use default behavior
	if !result.Enabled && flag.Enabled && len(flag.Rules) == 0 {
		result.Enabled = true
		result.Reason = "default_enabled"
	}

	// Log evaluation
	m.logEvaluation(ctx, result, evalCtx)

	return result, nil
}

// evaluateRule evaluates a single rule against context
func (m *Manager) evaluateRule(rule *Rule, ctx *EvaluationContext) bool {
	var value string

	switch rule.Attribute {
	case "user_id":
		value = ctx.UserID
	case "tenant_id":
		value = ctx.TenantID
	case "group":
		return m.evaluateGroups(rule, ctx.Groups)
	default:
		if v, exists := ctx.Metadata[rule.Attribute]; exists {
			if s, ok := v.(string); ok {
				value = s
			}
		}
	}

	switch rule.Operator {
	case "equals":
		return contains(rule.Values, value)
	case "not_equals":
		return !contains(rule.Values, value)
	case "in":
		return contains(rule.Values, value)
	case "not_in":
		return !contains(rule.Values, value)
	case "contains":
		for _, v := range rule.Values {
			if contains([]string{value}, v) {
				return true
			}
		}
		return false
	case "percentage":
		if rule.Percentage != nil {
			return m.evaluatePercentage(*rule.Percentage, ctx.UserID)
		}
	}

	return false
}

// evaluateGroups evaluates group membership
func (m *Manager) evaluateGroups(rule *Rule, userGroups []string) bool {
	for _, ruleGroup := range rule.Values {
		if contains(userGroups, ruleGroup) {
			return true
		}
	}
	return false
}

// evaluateRollout evaluates rollout configuration
func (m *Manager) evaluateRollout(rollout *RolloutConfig, ctx *EvaluationContext) bool {
	switch rollout.Strategy {
	case "percentage":
		return m.evaluatePercentage(rollout.Percentage, ctx.UserID)
	case "user_group":
		return m.evaluateUserGroups(rollout.UserGroups, ctx.Groups)
	case "time_based":
		return m.evaluateTimeWindow(rollout.StartTime, rollout.EndTime)
	}
	return true
}

// evaluatePercentage evaluates percentage-based rollout
func (m *Manager) evaluatePercentage(percentage int, userID string) bool {
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	// Use hash of userID to determine percentage bucket
	hash := simpleHash(userID)
	return (hash % 100) < percentage
}

// evaluateUserGroups evaluates user group rollout
func (m *Manager) evaluateUserGroups(rolloutGroups, userGroups []string) bool {
	for _, group := range rolloutGroups {
		if contains(userGroups, group) {
			return true
		}
	}
	return false
}

// evaluateTimeWindow evaluates time-based rollout
func (m *Manager) evaluateTimeWindow(start, end *time.Time) bool {
	now := time.Now()

	if start != nil && now.Before(*start) {
		return false
	}

	if end != nil && now.After(*end) {
		return false
	}

	return true
}

// RegisterCallback registers a callback for flag changes
func (m *Manager) RegisterCallback(flagID string, callback func(*FeatureFlag)) {
	m.callbackMux.Lock()
	defer m.callbackMux.Unlock()

	m.callbacks[flagID] = append(m.callbacks[flagID], callback)
}

// triggerCallbacks triggers callbacks for flag changes
func (m *Manager) triggerCallbacks(flagID string, flag *FeatureFlag) {
	m.callbackMux.RLock()
	callbacks := m.callbacks[flagID]
	m.callbackMux.RUnlock()

	for _, callback := range callbacks {
		go callback(flag)
	}
}

// logEvaluation logs flag evaluation
func (m *Manager) logEvaluation(ctx context.Context, result *EvaluationResult, evalCtx *EvaluationContext) {
	if err := m.storage.LogEvaluation(ctx, result, evalCtx); err != nil {
		m.logger.Error("Failed to log evaluation", zap.Error(err))
	}
}

// Helper functions

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("flag_%d", time.Now().UnixNano())
}

// contains checks if slice contains value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// simpleHash creates a simple hash for percentage evaluation
func simpleHash(s string) int {
	hash := 0
	for _, c := range s {
		hash = (hash*31 + int(c)) % 1000000
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}
